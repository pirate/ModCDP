import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";
import { CdpSocket } from "./helpers.BrowserLauncher.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

test("ModCDPClient connects with nested launch/upstream/extension/client/server config", async () => {
  const cdp = new ModCDPClient({
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
    client: {
      routes: { "Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp" },
      hydrate_aliases: true,
      mirror_upstream_events: true,
      cdp_send_timeout_ms: 10_000,
      event_wait_timeout_ms: 10_000,
    },
    server: {
      routes: { "*.*": "loopback_cdp" },
      cdp_send_timeout_ms: 10_000,
      loopback_execution_context_timeout_ms: 10_000,
      ws_connect_error_settle_timeout_ms: 250,
    },
  });

  try {
    await cdp.connect();
    assert.equal(cdp.launch.mode, "local");
    assert.equal(cdp.upstream.mode, "ws");
    assert.equal(cdp.upstream.reversews_wait_timeout_ms, 10_000);
    assert.equal(cdp.extension.mode, "auto");
    assert.equal(cdp.client.routes["*.*"], "direct_cdp");
    assert.equal(cdp.upstream_endpoint_kind, "raw_cdp");
    assert.match(cdp.cdp_url ?? "", /^ws:\/\//);
    await delay(2_000);
    const targets = (await cdp.sendRaw("Target.getTargets")) as {
      targetInfos: { type?: string; url?: string }[];
    };
    assert.equal(
      targets.targetInfos.some(
        (target) =>
          target.type === "service_worker" &&
          target.url === `chrome-extension://${cdp.extension_id}/modcdp/service_worker.js`,
      ),
      true,
    );
    assert.equal(
      targets.targetInfos.some(
        (target) =>
          target.type === "background_page" &&
          target.url === `chrome-extension://${cdp.extension_id}/offscreen/keepalive.html`,
      ),
      true,
    );
    assert.equal(typeof (await cdp.Browser.getVersion()).product, "string");
    const sent_at = Date.now();
    const pong = new Promise<Record<string, unknown>>((resolve) => {
      const listener = (payload: Record<string, unknown>) => {
        if (payload.sent_at !== sent_at) return;
        cdp.off("Mod.pong", listener);
        resolve(payload);
      };
      cdp.on("Mod.pong", listener);
    });
    const ping_result = (await cdp.Mod.ping({ sent_at })) as Record<string, unknown>;
    const pong_payload = await pong;
    assert.equal(ping_result.ok, true);
    assert.equal(pong_payload.sent_at, sent_at);
    assert.equal(typeof pong_payload.received_at, "number");
    assert.equal(pong_payload.from, "extension-service-worker");
  } finally {
    await cdp.close();
  }
}, 60_000);

test("ModCDPClient preserves explicit empty service worker suffix config", async () => {
  const cdp = new ModCDPClient({
    extension: {
      mode: "borrow",
      service_worker_url_suffixes: [],
    },
  });

  assert.deepEqual(cdp.extension.service_worker_url_suffixes, []);
  assert.deepEqual((await cdp._baseExtensionInjectorConfig()).service_worker_url_suffixes, []);
});

test("ModCDPClient.close does not close a remote browser it did not launch", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
  const raw_cdp = await CdpSocket.connect(chrome.ws_url!);
  const cdp = new ModCDPClient({
    launch: { mode: "remote" },
    upstream: { mode: "ws", ws_url: chrome.cdp_url },
    extension: {
      mode: "discover",
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
    },
  });

  try {
    await cdp.connect();
    await cdp.close();
    await delay(500);
    const version = await raw_cdp.send("Browser.getVersion");
    assert.match(String(version.product), /Chrome|Chromium/);
  } finally {
    await raw_cdp.close();
    await cdp.close();
    await chrome.close();
  }
}, 60_000);

test("ModCDPClient.close keeps injector files until after launched browser shutdown", async () => {
  const order: string[] = [];
  const cdp = new ModCDPClient();
  cdp.transport = {
    close: async () => order.push("transport"),
  } as any;
  cdp._launched = {
    close: async () => order.push("browser"),
  } as any;
  cdp._extension_injectors = [
    {
      close: async () => order.push("injector"),
    } as any,
  ];

  await cdp.close();

  assert.deepEqual(order, ["transport", "browser", "injector"]);
  assert.equal(cdp.transport, null);
  assert.equal(cdp._launched, null);
  assert.deepEqual(cdp._extension_injectors, []);
});
