import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { WebSocketUpstreamTransport } from "../bridge/WebSocketUpstreamTransport.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

test("ws upstream constructor, update, server config, and unconnected errors match the transport surface", async () => {
  const transport = new WebSocketUpstreamTransport();
  assert.equal(transport.url, "");
  assert.deepEqual(transport.getServerConfig(), {});
  assert.equal(transport.update({ ws_url: "ws://127.0.0.1:1/devtools/browser/test" }), transport);
  assert.equal(transport.url, "ws://127.0.0.1:1/devtools/browser/test");
  assert.deepEqual(transport.getServerConfig(), { loopback_cdp_url: "ws://127.0.0.1:1/devtools/browser/test" });
  const unconfigured = new WebSocketUpstreamTransport();
  await assert.rejects(() => unconfigured.connect(), /upstream\.mode=ws requires/);
  assert.throws(() => unconfigured.send({ id: 1, method: "Browser.getVersion" }), /CDP websocket is not connected/);
});

test("ws upstream launches a real browser and speaks raw CDP", async () => {
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
  });

  try {
    await cdp.connect();
    assert.equal(cdp.transport?.mode, "ws");
    assert.equal(cdp.upstream_endpoint_kind, "raw_cdp");
    assert.equal(cdp.connect_timing?.upstream_mode, "ws");
    assert.equal(cdp.connect_timing?.upstream_endpoint_kind, "raw_cdp");
    assert.equal(
      cdp.connect_timing?.transport_duration_ms,
      (cdp.connect_timing?.transport_connected_at ?? 0) - (cdp.connect_timing?.transport_started_at ?? 0),
    );
    assert.match(cdp.cdp_url ?? "", /^ws:\/\//);
    const version = (await cdp.sendRaw("Browser.getVersion")) as Record<string, unknown>;
    assert.equal(typeof version.product, "string");
    await new Promise((resolve) => setTimeout(resolve, 1_500));
    const target_infos = ((await cdp.sendRaw("Target.getTargets")) as { targetInfos?: { type?: string; url?: string }[] })
      .targetInfos;
    assert.equal(
      target_infos?.some(
        (target) => target.type === "service_worker" && target.url?.endsWith("/modcdp/service_worker.js"),
      ),
      true,
    );
    assert.equal(
      await cdp.Mod.evaluate({
        expression: "Boolean(globalThis.ModCDP?.handleCommand && chrome.runtime.getURL('modcdp/service_worker.js'))",
      }),
      true,
    );
  } finally {
    await cdp.close();
  }
}, 60_000);

test("ws upstream resolves a real HTTP CDP endpoint to the browser websocket", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
  }).launch();
  const transport = new WebSocketUpstreamTransport(chrome.cdp_url);

  try {
    const response = new Promise<Record<string, unknown>>((resolve) => {
      transport.onRecv((message) => {
        if ("id" in message && message.id === 1) resolve(message as Record<string, unknown>);
      });
    });
    await transport.connect();
    assert.match(transport.url ?? "", /^ws:\/\//);
    transport.send({ id: 1, method: "Browser.getVersion", params: {} });
    const message = await response;
    assert.equal(typeof (message.result as Record<string, unknown>).product, "string");
  } finally {
    await transport.close();
    await chrome.close();
  }
}, 60_000);
