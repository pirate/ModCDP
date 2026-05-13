import assert from "node:assert/strict";
import { spawn, type ChildProcess } from "node:child_process";
import { once } from "node:events";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { startProxy } from "../src/proxy/proxy.js";
import { ModCDPClient } from "../src/client/ModCDPClient.js";
import { CdpSocket } from "./helpers.BrowserLauncher.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");
const LOCAL_TEST_LAUNCH_OPTIONS = {
  headless: true,
  sandbox: process.platform !== "linux",
};

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
    const version = await cdp.send("Browser.getVersion");
    assert.equal(typeof version.product, "string");

    const evaluated = await cdp.send("Mod.evaluate", {
      expression: `({ ok: true, transport: ${JSON.stringify(transport)} })`,
    });
    assert.deepEqual(evaluated, { ok: true, transport });

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
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
    },
    server: {
      server_routes: { "*.*": "loopback_cdp" },
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
      JSON.stringify({ headless: true, sandbox: process.platform !== "linux" }),
      "--upstream-mode=pipe",
      "--injector-mode=auto",
      "--injector-service-worker-url-suffixes",
      JSON.stringify(["/modcdp/service_worker.js"]),
      "--injector-trust-service-worker-target",
      "true",
      "--injector-service-worker-probe-timeout-ms",
      "30000",
      "--server-routes",
      JSON.stringify({ "*.*": "loopback_cdp" }),
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
      JSON.stringify({ headless: true, sandbox: process.platform !== "linux" }),
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
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
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
      chrome.cdp_url!,
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
    await chrome.close();
  }
}, 60_000);

test("proxy CLI maps user-facing flags into a real reversews local launch", async () => {
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
      JSON.stringify({ headless: true, sandbox: process.platform !== "linux" }),
      "--upstream-mode=reversews",
      "--upstream-reversews-wait-timeout-ms",
      "10000",
      "--injector-mode=auto",
      "--injector-extension-path",
      EXTENSION_PATH,
      "--server-routes",
      JSON.stringify({ "*.*": "loopback_cdp" }),
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
  const reverse_port = await LocalBrowserLauncher.freePort();
  const reverse_bind = `127.0.0.1:${reverse_port}`;
  const reverse_url = `ws://${reverse_bind}`;
  const proxy = await startProxy({
    port: proxy_port,
    upstream: { upstream_mode: "reversews", upstream_reversews_bind: reverse_bind },
    server: {
      server_routes: { "*.*": "loopback_cdp" },
    },
  });
  const bootstrap = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
    server: {
      server_routes: { "*.*": "loopback_cdp" },
    },
  });

  try {
    await bootstrap.connect();
    await bootstrap.send("Mod.evaluate", {
      params: { reverse_url },
      expression: "async ({ reverse_url }) => ModCDP.startReverseBridge(reverse_url)",
    });
    await expectProxyCdpWorks(proxy.url, "reversews");
  } finally {
    await bootstrap.close();
    await proxy.close();
  }
}, 90_000);

test("proxy reversews local launch auto-injects the extension through the real client path", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy = await startProxy({
    port: proxy_port,
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: {
      upstream_mode: "reversews",
      upstream_reversews_wait_timeout_ms: 10_000,
    },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
    },
    server: {
      server_routes: { "*.*": "loopback_cdp" },
    },
  });

  try {
    await expectProxyCdpWorks(proxy.url, "reversews-local-launch");
  } finally {
    await proxy.close();
  }
}, 90_000);

test("proxy passes custom extension discovery config through to ModCDPClient", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const reverse_port = await LocalBrowserLauncher.freePort();
  await assert.rejects(
    () =>
      startProxy({
        port: proxy_port,
        launcher: {
          launcher_mode: "local",
          launcher_options: {
            headless: true,
            sandbox: process.platform !== "linux",
            extra_args: [`--load-extension=${EXTENSION_PATH}`],
          },
        },
        upstream: {
          upstream_mode: "reversews",
          upstream_reversews_bind: `127.0.0.1:${reverse_port}`,
          upstream_reversews_wait_timeout_ms: 1_000,
        },
        injector: {
          injector_mode: "discover",
          injector_extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
          injector_require_service_worker_target: true,
          injector_service_worker_probe_timeout_ms: 200,
          injector_service_worker_ready_timeout_ms: 200,
        },
        server: {
          server_routes: { "*.*": "loopback_cdp" },
        },
      }),
    /Timed out waiting 1000ms for reverse ModCDP extension connection/,
  );
}, 60_000);
