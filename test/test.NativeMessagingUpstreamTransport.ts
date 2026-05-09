import assert from "node:assert/strict";
import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { NativeMessagingUpstreamTransport } from "../bridge/NativeMessagingUpstreamTransport.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");
const nativeHostName = (label: string) => `com.modcdp.test.${label}.${process.pid}`;

test("nativemessaging upstream config owns manifest, host, wait timeout, loopback, and injector config", async () => {
  const transport = new NativeMessagingUpstreamTransport({
    manifest_path: "/tmp/modcdp-native-host.json",
    manifest_paths: ["/tmp/modcdp-native-host-extra.json"],
    host_name: "com.modcdp.test",
    extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    wait_timeout_ms: 10,
  });
  assert.deepEqual(transport.getInjectorConfig(), { native_host_name: "com.modcdp.test" });
  assert.deepEqual(transport.getServerConfig(), {});
  assert.equal(
    transport.update({
      ws_url: "ws://127.0.0.1:9222/devtools/browser/test",
      manifest_paths: [],
      native_host_name: "com.modcdp.updated",
      wait_timeout_ms: 5,
    }),
    transport,
  );
  assert.deepEqual(transport.getServerConfig(), {
    loopback_cdp_url: "ws://127.0.0.1:9222/devtools/browser/test",
  });
  assert.deepEqual(transport.getInjectorConfig(), { native_host_name: "com.modcdp.updated" });
  assert.equal(
    (transport as unknown as { include_default_manifest_paths: boolean }).include_default_manifest_paths,
    false,
  );
  transport.update({ manifest_path: null });
  assert.equal(
    (transport as unknown as { include_default_manifest_paths: boolean }).include_default_manifest_paths,
    true,
  );
  transport.update({ user_data_dir: "/tmp/modcdp-profile-one" });
  transport.update({ user_data_dir: "/tmp/modcdp-profile-one" });
  transport.update({ user_data_dir: "/tmp/modcdp-profile-two" });
  assert.deepEqual((transport as unknown as { manifest_paths: string[] }).manifest_paths, [
    path.join("/tmp/modcdp-profile-two", "NativeMessagingHosts", "com.modcdp.updated.json"),
    path.join("/tmp/modcdp-profile-two", "Default", "NativeMessagingHosts", "com.modcdp.updated.json"),
  ]);
  await assert.rejects(
    () => transport.waitForPeer(),
    /Timed out waiting 5ms for native messaging host com\.modcdp\.updated/,
  );
});

test("nativemessaging upstream close rejects pending peer waits", async () => {
  const transport = new NativeMessagingUpstreamTransport({
    host_name: "com.modcdp.close",
    wait_timeout_ms: 5_000,
  });
  const pending = transport.waitForPeer();

  await transport.close();

  await assert.rejects(
    () => pending,
    /Native messaging transport for com\.modcdp\.close closed before a peer connected/,
  );
});

test("nativemessaging upstream close resets peer wait state", async () => {
  const transport = new NativeMessagingUpstreamTransport({ wait_timeout_ms: 5 });
  (transport as unknown as { socket: { destroyed: boolean; destroy: () => void } | null }).socket = {
    destroyed: false,
    destroy() {
      this.destroyed = true;
    },
  };

  await transport.waitForPeer();
  await transport.close();

  await assert.rejects(
    () => transport.waitForPeer(),
    /Timed out waiting 5ms for native messaging host com\.modcdp\.bridge/,
  );
});

test("nativemessaging upstream installs the launch-profile native host manifest and connects to a real extension", async () => {
  const host_name = nativeHostName("client");
  const native_client = new ModCDPClient({
    launch: {
      mode: "local",
      options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { mode: "nativemessaging", nativemessaging_host_name: host_name },
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
    await native_client.connect();
    assert.equal(native_client.transport?.mode, "nativemessaging");
    assert.equal(native_client.upstream_endpoint_kind, "modcdp_server");
    assert.match(native_client.transport?.url ?? "", /^native:\/\/.+@127\.0\.0\.1:\d+$/);
    assert.equal(native_client.transport?.url?.startsWith(`native://${host_name}@`), true);
    assert.equal(
      existsSync(path.join(native_client._launched?.profile_dir ?? "", "NativeMessagingHosts", `${host_name}.json`)),
      true,
    );
    const version = (await native_client.send("Browser.getVersion")) as Record<string, unknown>;
    assert.equal(typeof version.product, "string");
  } finally {
    await native_client.close();
  }
}, 90_000);
