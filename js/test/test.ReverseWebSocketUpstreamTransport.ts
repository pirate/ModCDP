import assert from "node:assert/strict";
import { once } from "node:events";
import path from "node:path";
import { fileURLToPath } from "node:url";
import WebSocket from "ws";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { ReverseWebSocketUpstreamTransport } from "../src/transport/ReverseWebSocketUpstreamTransport.js";
import { ModCDPClient } from "../src/client/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

test("reversews upstream config owns bind updates, wait timeout, and injector config", async () => {
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: "127.0.0.1:29292",
    upstream_reversews_wait_timeout_ms: 10,
  });
  assert.equal(transport.url, "ws://127.0.0.1:29292");
  assert.deepEqual(transport.getInjectorConfig(), { upstream_reversews_url: "ws://127.0.0.1:29292" });
  assert.equal(
    transport.update({
      upstream_reversews_bind: "127.0.0.1:29293",
      upstream_reversews_wait_timeout_ms: 5,
    }),
    transport,
  );
  assert.equal(transport.url, "ws://127.0.0.1:29293");
  assert.deepEqual(transport.getInjectorConfig(), { upstream_reversews_url: "ws://127.0.0.1:29293" });
  assert.throws(
    () => transport.send({ id: 1, method: "Browser.getVersion" }),
    /No reverse ModCDP extension peer is connected/,
  );
  await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms/);
});

test("reversews upstream close rejects pending peer waits", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: `127.0.0.1:${reverse_port}`,
    upstream_reversews_wait_timeout_ms: 5_000,
  });
  const pending = transport.waitForPeer();

  await transport.close();

  await assert.rejects(
    () => pending,
    new RegExp(`Reverse websocket transport at ws://127\\.0\\.0\\.1:${reverse_port} closed before a peer connected`),
  );
});

test("reversews upstream close resets peer wait state", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: `127.0.0.1:${reverse_port}`,
    upstream_reversews_wait_timeout_ms: 5,
  });
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

test("reversews upstream waits again after a peer disconnects", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: `127.0.0.1:${reverse_port}`,
    upstream_reversews_wait_timeout_ms: 5,
  });
  await transport.connect();
  const peer = new WebSocket(transport.url);
  await once(peer, "open");
  peer.send(JSON.stringify({ type: "modcdp.reverse.hello", role: "test-peer", version: 1 }));

  try {
    await transport.waitForPeer();
    peer.close();
    await waitFor(() => (transport as unknown as { socket: unknown | null }).socket === null);

    await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms/);
  } finally {
    peer.close();
    await transport.close();
  }
});

test("reversews upstream accepts a replacement peer after disconnect", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: `127.0.0.1:${reverse_port}`,
    upstream_reversews_wait_timeout_ms: 500,
  });
  await transport.connect();
  const first_peer = new WebSocket(transport.url);
  await once(first_peer, "open");
  first_peer.send(JSON.stringify({ type: "modcdp.reverse.hello", role: "first-peer", version: 1 }));

  try {
    await transport.waitForPeer();
    first_peer.close();
    await waitFor(() => (transport as unknown as { socket: unknown | null }).socket === null);

    const second_peer = new WebSocket(transport.url);
    await once(second_peer, "open");
    second_peer.send(JSON.stringify({ type: "modcdp.reverse.hello", role: "second-peer", version: 1 }));
    try {
      await transport.waitForPeer();
      assert.equal(transport.peer_info?.role, "second-peer");
    } finally {
      second_peer.close();
    }
  } finally {
    first_peer.close();
    await transport.close();
  }
});

test("reversews upstream accepts a real extension reverse connection and routes CDP through loopback", async () => {
  const reverse_bind = "127.0.0.1:29292";
  const reverse_url = `ws://${reverse_bind}`;
  const reverse = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: {
        headless: process.platform === "linux" && !process.env.DISPLAY,
        sandbox: process.platform !== "linux",
      },
    },
    upstream: { upstream_mode: "reversews", upstream_reversews_bind: reverse_bind },
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
    await new Promise((resolve) => setTimeout(resolve, 1_500));
    const second_version = (await reverse.send("Browser.getVersion")) as Record<string, unknown>;
    assert.equal(typeof second_version.product, "string");
  } finally {
    await reverse.close();
  }
}, 60_000);

async function waitFor(predicate: () => boolean, timeout_ms = 2_000) {
  const deadline = Date.now() + timeout_ms;
  while (Date.now() < deadline) {
    if (predicate()) return;
    await new Promise((resolve) => setTimeout(resolve, 20));
  }
  throw new Error("Timed out waiting for condition");
}
