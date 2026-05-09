import assert from "node:assert/strict";
import { once } from "node:events";
import path from "node:path";
import { fileURLToPath } from "node:url";
import WebSocket from "ws";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { ReverseWebSocketUpstreamTransport } from "../bridge/ReverseWebSocketUpstreamTransport.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

test("reversews upstream config owns bind updates, wait timeout, and injector config", async () => {
  const transport = new ReverseWebSocketUpstreamTransport("127.0.0.1:29292", 10);
  assert.equal(transport.url, "ws://127.0.0.1:29292");
  assert.deepEqual(transport.getInjectorConfig(), { reverse_proxy_url: "ws://127.0.0.1:29292" });
  assert.equal(transport.update({ reversews_bind: "127.0.0.1:29293", reversews_wait_timeout_ms: 5 }), transport);
  assert.equal(transport.url, "ws://127.0.0.1:29293");
  assert.deepEqual(transport.getInjectorConfig(), { reverse_proxy_url: "ws://127.0.0.1:29293" });
  await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms/);
});

test("reversews upstream close rejects pending peer waits", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport(`127.0.0.1:${reverse_port}`, 5_000);
  const pending = transport.waitForPeer();

  await transport.close();

  await assert.rejects(
    () => pending,
    new RegExp(`Reverse websocket transport at ws://127\\.0\\.0\\.1:${reverse_port} closed before a peer connected`),
  );
});

test("reversews upstream close resets peer wait state", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport(`127.0.0.1:${reverse_port}`, 5);
  await transport.connect();
  const peer = new WebSocket(transport.url);
  await once(peer, "open");
  peer.send(JSON.stringify({ type: "modcdp.reverse.hello", role: "test-peer", version: 1 }));

  try {
    await transport.waitForPeer();
    assert.deepEqual(transport.peer_info, { type: "modcdp.reverse.hello", role: "test-peer", version: 1 });
    await transport.close();

    await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms/);
    assert.equal(transport.peer_info, null);
  } finally {
    peer.close();
    await transport.close();
  }
});

test("reversews upstream accepts a real extension reverse connection and routes CDP through loopback", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const reverse_bind = `127.0.0.1:${reverse_port}`;
  const reverse_url = `ws://${reverse_bind}`;
  const reverse = new ModCDPClient({
    launch: {
      mode: "local",
      options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { mode: "reversews", reversews_bind: reverse_bind },
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
    await reverse.connect();
    assert.equal(reverse.transport?.mode, "reversews");
    assert.equal(reverse.upstream_endpoint_kind, "modcdp_server");
    assert.equal(reverse.transport?.url, reverse_url);
    assert.equal(
      (reverse.transport as ReverseWebSocketUpstreamTransport).peer_info?.extension_id,
      "mdedooklbnfejodmnhmkdpkaedafkehf",
    );
    const version = (await reverse.send("Browser.getVersion")) as Record<string, unknown>;
    assert.equal(typeof version.product, "string");
  } finally {
    await reverse.close();
  }
}, 60_000);
