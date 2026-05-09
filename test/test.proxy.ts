import assert from "node:assert/strict";
import { spawn, type ChildProcess } from "node:child_process";
import { once } from "node:events";
import { mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { getBinaryPath } from "@eplightning/nats-server";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { startProxy } from "../bridge/proxy.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";
import { CdpSocket } from "./helpers.BrowserLauncher.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");
const LOCAL_TEST_LAUNCH_OPTIONS = { headless: true, sandbox: process.platform !== "linux" };

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function waitForWebSocket(url: string, timeout_ms = 10_000) {
  const deadline = Date.now() + timeout_ms;
  let last_error: unknown = null;
  while (Date.now() < deadline) {
    try {
      const ws = new WebSocket(url);
      await new Promise<void>((resolve, reject) => {
        ws.addEventListener("open", () => resolve(), { once: true });
        ws.addEventListener("error", reject, { once: true });
      });
      ws.close();
      return;
    } catch (error) {
      last_error = error;
      await delay(50);
    }
  }
  throw last_error instanceof Error ? last_error : new Error(`Timed out waiting for ${url}`);
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

async function startNatsServer() {
  const websocket_port = await LocalBrowserLauncher.freePort();
  const client_port = await LocalBrowserLauncher.freePort();
  const dir = await mkdtemp(path.join(tmpdir(), "modcdp-proxy-nats-"));
  const config_path = path.join(dir, "nats.conf");
  await writeFile(
    config_path,
    [
      `host: "127.0.0.1"`,
      `port: ${client_port}`,
      `websocket {`,
      `  host: "127.0.0.1"`,
      `  port: ${websocket_port}`,
      `  no_tls: true`,
      `}`,
      ``,
    ].join("\n"),
  );
  const proc = spawn(await getBinaryPath(), ["-c", config_path], { stdio: "ignore" });
  const url = `ws://127.0.0.1:${websocket_port}`;
  try {
    await waitForWebSocket(url);
  } catch (error) {
    await closeProcess(proc);
    await rm(dir, { recursive: true, force: true });
    throw error;
  }
  return {
    url,
    close: async () => {
      await closeProcess(proc);
      await rm(dir, { recursive: true, force: true });
    },
  };
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

    const created = await cdp.send("Target.createTarget", { url: `about:blank#modcdp-proxy-${transport}` });
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
    launch: {
      mode: "local",
      options: LOCAL_TEST_LAUNCH_OPTIONS,
    },
    upstream: { mode: "ws" },
    extension: {
      mode: "auto",
      path: EXTENSION_PATH,
    },
    server: {
      routes: { "*.*": "loopback_cdp" },
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
    launch: {
      mode: "local",
      options: LOCAL_TEST_LAUNCH_OPTIONS,
    },
    upstream: { mode: "pipe" },
    extension: {
      mode: "auto",
      path: EXTENSION_PATH,
    },
    server: {
      routes: { "*.*": "loopback_cdp" },
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
  const proxy_script = path.resolve(HERE, "..", "dist", "bridge", "proxy.js");
  const proc = spawn(
    process.execPath,
    [
      proxy_script,
      "--port",
      String(proxy_port),
      "--launch=local",
      "--launch-options",
      JSON.stringify({ headless: true, sandbox: process.platform !== "linux" }),
      "--upstream=pipe",
      "--extension=auto",
      "--extension-path",
      EXTENSION_PATH,
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
  const proxy_script = path.resolve(HERE, "..", "dist", "bridge", "proxy.js");
  const proc = spawn(
    process.execPath,
    [
      proxy_script,
      "--port",
      String(proxy_port),
      "--launch=local",
      "--launch-options",
      JSON.stringify({ headless: true, sandbox: process.platform !== "linux" }),
      "--upstream=ws",
      "--extension=auto",
      "--extension-path",
      EXTENSION_PATH,
      "--server-routes",
      JSON.stringify({ "*.*": "loopback_cdp" }),
    ],
    { stdio: ["ignore", "pipe", "pipe"] },
  );

  try {
    await waitForHttpJsonVersion(`http://127.0.0.1:${proxy_port}/json/version`);
    await expectProxyCdpWorks(`ws://127.0.0.1:${proxy_port}/devtools/browser/proxy`, "cli-ws-local");
  } finally {
    await closeProcess(proc);
  }
}, 60_000);

test("proxy CLI maps ws upstream URL and route shorthands into an existing real browser", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy_script = path.resolve(HERE, "..", "dist", "bridge", "proxy.js");
  const proc = spawn(
    process.execPath,
    [
      proxy_script,
      "--port",
      String(proxy_port),
      "--launch=remote",
      "--upstream=ws",
      "--upstream-ws-url",
      chrome.cdp_url!,
      "--extension=discover",
      "--client-routes",
      JSON.stringify({ "Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp" }),
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
  const reverse_port = await LocalBrowserLauncher.freePort();
  const proxy_script = path.resolve(HERE, "..", "dist", "bridge", "proxy.js");
  const proc = spawn(
    process.execPath,
    [
      proxy_script,
      "--port",
      String(proxy_port),
      "--launch=local",
      "--launch-options",
      JSON.stringify({ headless: true, sandbox: process.platform !== "linux" }),
      "--upstream=reversews",
      "--upstream-reversews-bind",
      `127.0.0.1:${reverse_port}`,
      "--upstream-reversews-wait-timeout-ms",
      "10000",
      "--extension=auto",
      "--extension-path",
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

test("proxy CLI maps user-facing flags into a real nativemessaging local launch", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const manifest_dir = await mkdtemp(path.join(tmpdir(), "modcdp-proxy-native-"));
  const manifest_path = path.join(manifest_dir, "com.modcdp.bridge.json");
  const proxy_script = path.resolve(HERE, "..", "dist", "bridge", "proxy.js");
  const proc = spawn(
    process.execPath,
    [
      proxy_script,
      "--port",
      String(proxy_port),
      "--launch=local",
      "--launch-options",
      JSON.stringify({ headless: true, sandbox: process.platform !== "linux" }),
      "--upstream=nativemessaging",
      "--upstream-nativemessaging-manifest",
      manifest_path,
      "--extension=auto",
      "--extension-path",
      EXTENSION_PATH,
      "--server-routes",
      JSON.stringify({ "*.*": "loopback_cdp" }),
    ],
    { stdio: ["ignore", "pipe", "pipe"] },
  );

  try {
    await waitForHttpJsonVersion(`http://127.0.0.1:${proxy_port}/json/version`, 20_000);
    await expectProxyCdpWorks(`ws://127.0.0.1:${proxy_port}/devtools/browser/proxy`, "cli-nativemessaging");
  } finally {
    await closeProcess(proc);
    await rm(manifest_dir, { recursive: true, force: true });
  }
}, 90_000);

test("proxy upgrades a vanilla CDP websocket to ModCDP against a real browser over nats upstream", async () => {
  const nats = await startNatsServer();
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy = await startProxy({
    port: proxy_port,
    launch: {
      mode: "local",
      options: LOCAL_TEST_LAUNCH_OPTIONS,
    },
    upstream: {
      mode: "nats",
      nats_url: nats.url,
      nats_subject_prefix: `modcdp.proxy.${Date.now()}`,
    },
    extension: {
      mode: "auto",
      path: EXTENSION_PATH,
    },
    server: {
      routes: { "*.*": "loopback_cdp" },
    },
  });

  try {
    await expectProxyCdpWorks(proxy.url, "nats");
  } finally {
    await proxy.close();
    await nats.close();
  }
}, 90_000);

test("proxy upgrades a vanilla CDP websocket to ModCDP against a real browser over nativemessaging upstream", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const proxy = await startProxy({
    port: proxy_port,
    launch: {
      mode: "local",
      options: LOCAL_TEST_LAUNCH_OPTIONS,
    },
    upstream: { mode: "nativemessaging" },
    extension: {
      mode: "auto",
      path: EXTENSION_PATH,
    },
    server: {
      routes: { "*.*": "loopback_cdp" },
    },
  });

  try {
    await expectProxyCdpWorks(proxy.url, "nativemessaging");
  } finally {
    await proxy.close();
  }
}, 90_000);

test("proxy upgrades a vanilla CDP websocket to ModCDP against a real browser over reversews upstream", async () => {
  const proxy_port = await LocalBrowserLauncher.freePort();
  const reverse_port = await LocalBrowserLauncher.freePort();
  const reverse_bind = `127.0.0.1:${reverse_port}`;
  const reverse_url = `ws://${reverse_bind}`;
  const proxy = await startProxy({
    port: proxy_port,
    upstream: { mode: "reversews", reversews_bind: reverse_bind },
    server: {
      routes: { "*.*": "loopback_cdp" },
    },
  });
  const bootstrap = new ModCDPClient({
    launch: {
      mode: "local",
      options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { mode: "ws" },
    extension: {
      mode: "auto",
      path: EXTENSION_PATH,
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
    },
    server: {
      routes: { "*.*": "loopback_cdp" },
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
  const reverse_port = await LocalBrowserLauncher.freePort();
  const reverse_bind = `127.0.0.1:${reverse_port}`;
  const proxy = await startProxy({
    port: proxy_port,
    launch: {
      mode: "local",
      options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: {
      mode: "reversews",
      reversews_bind: reverse_bind,
      reversews_wait_timeout_ms: 10_000,
    },
    extension: {
      mode: "auto",
      path: EXTENSION_PATH,
    },
    server: {
      routes: { "*.*": "loopback_cdp" },
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
        launch: {
          mode: "local",
          options: {
            headless: true,
            sandbox: process.platform !== "linux",
            extra_args: [`--load-extension=${EXTENSION_PATH}`],
          },
        },
        upstream: {
          mode: "reversews",
          reversews_bind: `127.0.0.1:${reverse_port}`,
          reversews_wait_timeout_ms: 1_000,
        },
        extension: {
          mode: "discover",
          extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
          require_service_worker_target: true,
          service_worker_probe_timeout_ms: 200,
          service_worker_ready_timeout_ms: 200,
        },
        server: {
          routes: { "*.*": "loopback_cdp" },
        },
      }),
    /Timed out waiting 1000ms for reverse ModCDP extension connection/,
  );
}, 60_000);
