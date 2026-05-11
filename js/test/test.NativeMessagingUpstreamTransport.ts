import assert from "node:assert/strict";
import { once } from "node:events";
import { existsSync } from "node:fs";
import net from "node:net";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { NativeMessagingUpstreamTransport } from "../src/transport/NativeMessagingUpstreamTransport.js";
import { ModCDPClient } from "../src/client/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");
const upstreamNativeMessagingHostName = (label: string) => `com.modcdp.test.${label}.${process.pid}`;

test("nativemessaging upstream config owns manifest, host, wait timeout, loopback, and injector config", async () => {
  const transport = new NativeMessagingUpstreamTransport({
    upstream_nativemessaging_manifest: "/tmp/modcdp-native-host.json",
    upstream_nativemessaging_manifests: ["/tmp/modcdp-native-host-extra.json"],
    upstream_nativemessaging_host_name: "com.modcdp.test",
    injector_extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    upstream_nativemessaging_wait_timeout_ms: 10,
  });
  assert.deepEqual(transport.getInjectorConfig(), {
    upstream_nativemessaging_host_name: "com.modcdp.test",
  });
  assert.deepEqual(transport.getServerConfig(), {});
  assert.equal(
    transport.update({
      cdp_url: "ws://127.0.0.1:9222/devtools/browser/test",
      upstream_nativemessaging_manifests: [],
      upstream_nativemessaging_host_name: "com.modcdp.updated",
      upstream_nativemessaging_wait_timeout_ms: 5,
    }),
    transport,
  );
  assert.deepEqual(transport.getServerConfig(), {
    server_loopback_cdp_url: "ws://127.0.0.1:9222/devtools/browser/test",
  });
  assert.deepEqual(transport.getInjectorConfig(), {
    upstream_nativemessaging_host_name: "com.modcdp.updated",
  });
  assert.equal(
    (transport as unknown as { include_default_manifest_paths: boolean }).include_default_manifest_paths,
    false,
  );
  transport.update({ upstream_nativemessaging_manifest: null });
  assert.equal(
    (transport as unknown as { include_default_manifest_paths: boolean }).include_default_manifest_paths,
    true,
  );
  transport.update({ user_data_dir: "/tmp/modcdp-profile-one" });
  transport.update({ user_data_dir: "/tmp/modcdp-profile-one" });
  transport.update({ user_data_dir: "/tmp/modcdp-profile-two" });
  assert.deepEqual(
    (transport as unknown as { upstream_nativemessaging_manifests: string[] }).upstream_nativemessaging_manifests,
    [
      path.join("/tmp/modcdp-profile-two", "NativeMessagingHosts", "com.modcdp.updated.json"),
      path.join("/tmp/modcdp-profile-two", "Default", "NativeMessagingHosts", "com.modcdp.updated.json"),
    ],
  );
  await assert.rejects(
    () => transport.waitForPeer(),
    /Timed out waiting 5ms for native messaging host com\.modcdp\.updated/,
  );
});

test("nativemessaging upstream close rejects pending peer waits", async () => {
  const transport = new NativeMessagingUpstreamTransport({
    upstream_nativemessaging_host_name: "com.modcdp.close",
    upstream_nativemessaging_wait_timeout_ms: 5_000,
  });
  const pending = transport.waitForPeer();

  await transport.close();

  await assert.rejects(
    () => pending,
    /Native messaging transport for com\.modcdp\.close closed before a peer connected/,
  );
});

test("nativemessaging upstream close resets peer wait state", async () => {
  const upstream_nativemessaging_host_name = upstreamNativeMessagingHostName("close-reset");
  const transport = new NativeMessagingUpstreamTransport({
    upstream_nativemessaging_host_name,
    upstream_nativemessaging_wait_timeout_ms: 5,
  });
  await transport.connect();
  const port = Number(new URL(transport.url).port);
  const peer = net.createConnection({ host: "127.0.0.1", port });
  await once(peer, "connect");

  try {
    await transport.waitForPeer();
    await transport.close();

    await assert.rejects(() => transport.waitForPeer(), {
      message: `Timed out waiting 5ms for native messaging host ${upstream_nativemessaging_host_name}.`,
    });
  } finally {
    peer.destroy();
    await transport.close();
  }
});

test("nativemessaging upstream waits again after a peer disconnects", async () => {
  const upstream_nativemessaging_host_name = upstreamNativeMessagingHostName("disconnect-reset");
  const transport = new NativeMessagingUpstreamTransport({
    upstream_nativemessaging_host_name,
    upstream_nativemessaging_wait_timeout_ms: 5,
  });
  await transport.connect();
  const port = Number(new URL(transport.url).port);
  const peer = net.createConnection({ host: "127.0.0.1", port });
  await once(peer, "connect");

  try {
    await transport.waitForPeer();
    peer.destroy();
    await waitFor(() => (transport as unknown as { socket: unknown | null }).socket === null);

    await assert.rejects(() => transport.waitForPeer(), {
      message: `Timed out waiting 5ms for native messaging host ${upstream_nativemessaging_host_name}.`,
    });
  } finally {
    peer.destroy();
    await transport.close();
  }
});

test("nativemessaging upstream accepts a replacement peer after disconnect", async () => {
  const upstream_nativemessaging_host_name = upstreamNativeMessagingHostName("replacement");
  const transport = new NativeMessagingUpstreamTransport({
    upstream_nativemessaging_host_name,
    upstream_nativemessaging_wait_timeout_ms: 500,
  });
  await transport.connect();
  const port = Number(new URL(transport.url).port);
  const first_peer = net.createConnection({ host: "127.0.0.1", port });
  await once(first_peer, "connect");

  try {
    await transport.waitForPeer();
    first_peer.destroy();
    await waitFor(() => (transport as unknown as { socket: unknown | null }).socket === null);

    const second_peer = net.createConnection({ host: "127.0.0.1", port });
    await once(second_peer, "connect");
    try {
      await transport.waitForPeer();
    } finally {
      second_peer.destroy();
    }
  } finally {
    first_peer.destroy();
    await transport.close();
  }
});

test("nativemessaging upstream installs the launch-profile native host manifest and connects to a real extension", async () => {
  const upstream_nativemessaging_host_name = "com.modcdp.bridge";
  const native_client = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: {
      upstream_mode: "nativemessaging",
      upstream_nativemessaging_host_name: upstream_nativemessaging_host_name,
    },
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
    await native_client.connect();
    assert.equal(native_client.transport?.mode, "nativemessaging");
    assert.equal(native_client.upstream_endpoint_kind, "modcdp_server");
    assert.match(native_client.transport?.url ?? "", /^native:\/\/.+@127\.0\.0\.1:\d+$/);
    assert.equal(native_client.transport?.url?.startsWith(`native://${upstream_nativemessaging_host_name}@`), true);
    assert.equal(
      existsSync(
        path.join(
          native_client._launched?.profile_dir ?? "",
          "NativeMessagingHosts",
          `${upstream_nativemessaging_host_name}.json`,
        ),
      ),
      true,
    );
    const version = (await native_client.send("Browser.getVersion")) as Record<string, unknown>;
    assert.equal(typeof version.product, "string");
    await new Promise((resolve) => setTimeout(resolve, 1_500));
    const second_version = (await native_client.send("Browser.getVersion")) as Record<string, unknown>;
    assert.equal(typeof second_version.product, "string");
  } finally {
    await native_client.close();
  }
}, 90_000);

async function waitFor(predicate: () => boolean, timeout_ms = 2_000) {
  const deadline = Date.now() + timeout_ms;
  while (Date.now() < deadline) {
    if (predicate()) return;
    await new Promise((resolve) => setTimeout(resolve, 20));
  }
  throw new Error("Timed out waiting for condition");
}
