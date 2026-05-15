import assert from "node:assert/strict";
import { spawn, type ChildProcess } from "node:child_process";
import { once } from "node:events";
import { existsSync, readdirSync, statSync } from "node:fs";
import { mkdtemp, rm } from "node:fs/promises";
import { homedir, platform, tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { ModCDPClient } from "../src/client/ModCDPClient.js";
import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { startProxy } from "../src/proxy/proxy.js";
import { CdpSocket } from "./helpers.BrowserLauncher.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");
const LOCAL_TEST_LAUNCH_OPTIONS = {
  headless: true,
};
const REVERSEWS_TEST_BROWSER_PATH = reversewsTestBrowserPath();

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function waitForHttpJsonVersion(url: string, timeout_ms = 10_000) {
  const deadline = Date.now() + timeout_ms;
  let last_error: unknown = null;
  while (Date.now() < deadline) {
    try {
      const response = await fetch(url);
      if (response.ok) return;
      last_error = new Error(`${url} returned ${response.status}`);
    } catch (error) {
      last_error = error;
    }
    await delay(50);
  }
  throw last_error instanceof Error ? last_error : new Error(`Timed out waiting for ${url}`);
}

async function closeProcess(proc: ChildProcess) {
  if (proc.exitCode != null || proc.signalCode != null) return;
  proc.kill("SIGTERM");
  await Promise.race([once(proc, "exit"), delay(2_000)]);
  if (proc.exitCode != null || proc.signalCode != null) return;
  proc.kill("SIGKILL");
  await Promise.race([once(proc, "exit"), delay(2_000)]);
}

async function removeTree(pathname: string) {
  for (let attempt = 0; attempt < 10; attempt++) {
    try {
      await rm(pathname, { recursive: true, force: true });
      return;
    } catch (error) {
      if (attempt === 9) throw error;
      await delay(100 * (attempt + 1));
    }
  }
}

async function expectProxyCdpWorks(proxy_url: string, transport: string) {
  const cdp = await CdpSocket.connect(proxy_url);
  let target_id: string | null = null;
  try {
    const evaluated = await cdp.send("Mod.evaluate", {
      expression: `({ ok: true, transport: ${JSON.stringify(transport)} })`,
    });
    assert.deepEqual(evaluated, { ok: true, transport });

    if (transport.includes("reversews")) {
      const runtime = await cdp.send("Runtime.evaluate", {
        expression: "document.readyState",
        returnByValue: true,
      });
      assert.equal((runtime.result as { value?: unknown } | undefined)?.value, "complete");
      return;
    }

    if (transport.includes("pipe")) {
      await cdp.send("Mod.addCustomCommand", {
        name: "Custom.runtimeReadyState",
        expression:
          "async () => await cdp.send('Runtime.evaluate', { expression: 'document.readyState', returnByValue: true })",
      });
      const runtime = await cdp.send("Custom.runtimeReadyState");
      assert.equal((runtime.result as { value?: unknown } | undefined)?.value, "complete");
      return;
    }

    const version = await cdp.send("Browser.getVersion");
    assert.equal(typeof version.product, "string");

    const created = await cdp.send("Target.createTarget", {
      url: `about:blank#modcdp-proxy-${transport}`,
    });
    assert.equal(typeof created.targetId, "string");
    target_id = created.targetId as string;
  } finally {
    if (target_id) await cdp.send("Target.closeTarget", { targetId: target_id }).catch(() => ({}));
    await cdp.close();
  }
}

test("proxy upgrades a vanilla CDP websocket to ModCDP against a real browser over ws upstream", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy = await startProxy({
    port: proxy_port,
    launcher: {
      launcher_mode: "local",
      launcher_options: LOCAL_TEST_LAUNCH_OPTIONS,
    },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
    },
    server: {
      server_routes: { "*.*": "loopback_cdp" },
    },
  });

  try {
    await expectProxyCdpWorks(proxy.url, "ws");
  } finally {
    await proxy.close();
  }
}, 60_000);

test("proxy upgrades a vanilla CDP websocket to ModCDP against a real browser over pipe upstream", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy = await startProxy({
    port: proxy_port,
    launcher: {
      launcher_mode: "local",
      launcher_options: LOCAL_TEST_LAUNCH_OPTIONS,
    },
    upstream: { upstream_mode: "pipe" },
    injector: {
      injector_mode: "inject",
      injector_extension_path: EXTENSION_PATH,
    },
    server: {
      server_routes: { "*.*": "chrome_debugger" },
    },
  });

  try {
    await expectProxyCdpWorks(proxy.url, "pipe");
  } finally {
    await proxy.close();
  }
}, 60_000);

test("proxy CLI maps user-facing flags into a real pipe upstream browser session", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy_script = path.resolve(HERE, "..", "..", "dist", "js", "src", "proxy", "proxy.js");
  const proc = spawn(
    process.execPath,
    [
      proxy_script,
      "--port",
      String(proxy_port),
      "--launcher-mode=local",
      "--launcher-options",
      JSON.stringify({ headless: true }),
      "--upstream-mode=pipe",
      "--injector-mode=inject",
      "--injector-extension-path",
      EXTENSION_PATH,
      "--injector-service-worker-url-suffixes",
      JSON.stringify(["/modcdp/service_worker.js"]),
      "--injector-trust-service-worker-target",
      "true",
      "--server-routes",
      JSON.stringify({ "*.*": "chrome_debugger" }),
    ],
    { stdio: ["ignore", "pipe", "pipe"] },
  );

  try {
    await waitForHttpJsonVersion(`http://127.0.0.1:${proxy_port}/json/version`);
    await expectProxyCdpWorks(`ws://127.0.0.1:${proxy_port}/devtools/browser/proxy`, "cli-pipe");
  } finally {
    await closeProcess(proc);
  }
}, 60_000);

test("proxy CLI maps local ws launch without requiring upstream ws url", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy_script = path.resolve(HERE, "..", "..", "dist", "js", "src", "proxy", "proxy.js");
  const user_data_dir = await mkdtemp(path.join(tmpdir(), "modcdp-proxy-profile-"));
  const executable_path = LocalBrowserLauncher.findChromeBinary();
  const proc = spawn(
    process.execPath,
    [
      proxy_script,
      "--port",
      String(proxy_port),
      "--launcher-mode=local",
      "--launcher-executable-path",
      executable_path,
      "--launcher-user-data-dir",
      user_data_dir,
      "--launcher-options",
      JSON.stringify({ headless: true }),
      "--upstream-mode=ws",
      "--injector-mode=auto",
      "--injector-extension-path",
      EXTENSION_PATH,
      "--client",
      JSON.stringify({
        client_routes: {
          "Mod.*": "service_worker",
          "Custom.*": "service_worker",
          "*.*": "direct_cdp",
        },
      }),
      "--server",
      JSON.stringify({ server_routes: { "*.*": "loopback_cdp" } }),
    ],
    { stdio: ["ignore", "pipe", "pipe"] },
  );

  try {
    await waitForHttpJsonVersion(`http://127.0.0.1:${proxy_port}/json/version`);
    await expectProxyCdpWorks(`ws://127.0.0.1:${proxy_port}/devtools/browser/proxy`, "cli-ws-local");
  } finally {
    await closeProcess(proc);
    await removeTree(user_data_dir);
  }
}, 60_000);

test("proxy CLI maps ws upstream URL and route shorthands into an existing real browser", async () => {
  const owner = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: LOCAL_TEST_LAUNCH_OPTIONS,
    },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
  });
  await owner.connect();
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy_script = path.resolve(HERE, "..", "..", "dist", "js", "src", "proxy", "proxy.js");
  const proc = spawn(
    process.execPath,
    [
      proxy_script,
      "--port",
      String(proxy_port),
      "--launcher-mode=remote",
      "--upstream-mode=ws",
      "--upstream-cdp-url",
      owner.cdp_url!,
      "--injector-mode=discover",
      "--client-routes",
      JSON.stringify({
        "Mod.*": "service_worker",
        "Custom.*": "service_worker",
        "*.*": "direct_cdp",
      }),
      "--server-routes",
      JSON.stringify({ "*.*": "loopback_cdp" }),
    ],
    { stdio: ["ignore", "pipe", "pipe"] },
  );

  try {
    await waitForHttpJsonVersion(`http://127.0.0.1:${proxy_port}/json/version`);
    await expectProxyCdpWorks(`ws://127.0.0.1:${proxy_port}/devtools/browser/proxy`, "cli-ws");
  } finally {
    await closeProcess(proc);
    await owner.close();
  }
}, 60_000);

test("proxy CLI maps user-facing flags into a real reversews browser session", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy_script = path.resolve(HERE, "..", "..", "dist", "js", "src", "proxy", "proxy.js");
  const proc = spawn(
    process.execPath,
    [
      proxy_script,
      "--port",
      String(proxy_port),
      "--launcher-mode=local",
      "--launcher-options",
      JSON.stringify({
        ...LOCAL_TEST_LAUNCH_OPTIONS,
        // Reversews is browser -> client only. After explicit CHROME_PATH and
        // CI /usr/bin/chromium, these tests use Chrome for Testing because
        // Canary rejects --load-extension in this local test path.
        executable_path: REVERSEWS_TEST_BROWSER_PATH,
      }),
      "--upstream-mode=reversews",
      "--upstream-reversews-wait-timeout-ms",
      "10000",
      "--injector-mode=auto",
      "--injector-extension-path",
      EXTENSION_PATH,
      "--injector-service-worker-url-suffixes",
      JSON.stringify(["/modcdp/service_worker.js"]),
      "--injector-trust-service-worker-target",
      "true",
      "--injector-service-worker-probe-timeout-ms",
      "1000",
    ],
    { stdio: ["ignore", "pipe", "pipe"] },
  );

  try {
    await waitForHttpJsonVersion(`http://127.0.0.1:${proxy_port}/json/version`, 20_000);
    await expectProxyCdpWorks(`ws://127.0.0.1:${proxy_port}/devtools/browser/proxy`, "cli-reversews");
  } finally {
    await closeProcess(proc);
  }
}, 90_000);

test("proxy upgrades a vanilla CDP websocket to ModCDP against a real browser over reversews upstream", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy = await startProxy({
    port: proxy_port,
    launcher: {
      launcher_mode: "local",
      launcher_options: {
        ...LOCAL_TEST_LAUNCH_OPTIONS,
        // Reversews is browser -> client only. After explicit CHROME_PATH and
        // CI /usr/bin/chromium, these tests use Chrome for Testing because
        // Canary rejects --load-extension in this local test path.
        executable_path: REVERSEWS_TEST_BROWSER_PATH,
      },
    },
    upstream: { upstream_mode: "reversews", upstream_reversews_wait_timeout_ms: 10_000 },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
      injector_service_worker_probe_timeout_ms: 1_000,
    },
  });

  try {
    await expectProxyCdpWorks(proxy.url, "reversews");
  } finally {
    await proxy.close();
  }
}, 90_000);

function reversewsTestBrowserPath() {
  const explicit_candidates = [process.env.CHROME_PATH, platform() === "linux" ? "/usr/bin/chromium" : null].filter(
    (candidate): candidate is string => Boolean(candidate),
  );
  for (const candidate of explicit_candidates) {
    if (existsSync(candidate)) return candidate;
  }
  const home = homedir();
  const patterns =
    platform() === "darwin"
      ? [
          path.join(
            home,
            "Library/Caches/ms-playwright/chromium-*/chrome-mac*/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing",
          ),
          path.join(home, "Library/Caches/ms-playwright/chromium-*/chrome-mac*/Chromium.app/Contents/MacOS/Chromium"),
          path.join(
            home,
            "Library/Caches/puppeteer/chrome/mac*-*/chrome-mac*/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing",
          ),
        ]
      : platform() === "win32"
        ? [
            path.join(
              process.env.LOCALAPPDATA || path.join(home, "AppData/Local"),
              "ms-playwright/chromium-*/chrome-win*/chrome.exe",
            ),
            path.join(home, ".cache/puppeteer/chrome/win*-*/chrome-win*/chrome.exe"),
          ]
        : [
            path.join(home, ".cache/ms-playwright/chromium-*/chrome-linux*/chrome"),
            "/opt/pw-browsers/chromium-*/chrome-linux*/chrome",
            path.join(home, ".cache/puppeteer/chrome/linux-*/chrome-linux*/chrome"),
          ];
  const candidates = newestFirst(patterns.flatMap(expandGlob));
  if (candidates[0]) return candidates[0];
  throw new Error("Reversews tests require CHROME_PATH, /usr/bin/chromium, or Chrome for Testing.");
}

function expandGlob(pattern: string) {
  const normalized = path.normalize(pattern);
  const { root } = path.parse(normalized);
  const parts = normalized.slice(root.length).split(path.sep).filter(Boolean);
  let candidates = [root || "."];
  for (const part of parts) {
    const has_wildcard = part.includes("*");
    const matcher = has_wildcard ? wildcardToRegExp(part) : null;
    const next: string[] = [];
    for (const base of candidates) {
      if (!existsSync(base)) continue;
      if (!has_wildcard) {
        const candidate = path.join(base, part);
        if (existsSync(candidate)) next.push(candidate);
        continue;
      }
      for (const child of readdirSync(base)) {
        if (matcher!.test(child)) next.push(path.join(base, child));
      }
    }
    candidates = next;
  }
  return candidates.filter((candidate) => existsSync(candidate));
}

function wildcardToRegExp(value: string) {
  return new RegExp(`^${value.replace(/[.+^${}()|[\]\\]/g, "\\$&").replace(/\*/g, ".*")}$`);
}

function newestFirst(candidates: string[]) {
  return [...new Set(candidates)].sort((a, b) => {
    const left = scorePath(a);
    const right = scorePath(b);
    return right.version - left.version || right.mtime - left.mtime || a.localeCompare(b);
  });
}

function scorePath(candidate: string) {
  const numbers = candidate.match(/\d+/g)?.map(Number) ?? [];
  const version = numbers.length > 0 ? Math.max(...numbers) : 0;
  let mtime = 0;
  try {
    mtime = statSync(candidate).mtimeMs;
  } catch {}
  return { version, mtime };
}
