import assert from "node:assert/strict";
import { spawn, type ChildProcess } from "node:child_process";
import { once } from "node:events";
import { writeFile, mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { getBinaryPath } from "@eplightning/nats-server";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { NatsUpstreamTransport } from "../src/transport/NatsUpstreamTransport.js";

test("nats upstream config owns url, subject prefix, wait timeout, and injector config", async () => {
  const transport = new NatsUpstreamTransport({
    upstream_nats_url: "ws://127.0.0.1:4223",
    upstream_nats_subject_prefix: "modcdp.one",
  });
  assert.equal(transport.url, "ws://127.0.0.1:4223/");
  assert.equal(transport.upstream_nats_subject_prefix, "modcdp.one");
  assert.deepEqual(transport.getInjectorConfig(), {
    upstream_nats_url: "ws://127.0.0.1:4223/",
    upstream_nats_subject_prefix: "modcdp.one",
  });
  assert.equal(
    transport.update({
      upstream_nats_url: "nats://127.0.0.1:4222",
      upstream_nats_subject_prefix: "modcdp.two",
      upstream_nats_role: "browser",
      upstream_nats_wait_timeout_ms: 5,
    }),
    transport,
  );
  assert.equal(transport.url, "nats://127.0.0.1:4222");
  assert.equal(transport.upstream_nats_subject_prefix, "modcdp.two");
  await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms for NATS ModCDP peer/);
});

test("nats upstream close rejects pending peer waits", async () => {
  const transport = new NatsUpstreamTransport({
    upstream_nats_url: "ws://127.0.0.1:4223",
    upstream_nats_subject_prefix: "modcdp.close",
    upstream_nats_wait_timeout_ms: 5_000,
  });
  const pending = transport.waitForPeer();

  await transport.close();

  await assert.rejects(() => pending, /NATS transport for modcdp\.close closed before a peer connected/);
});

test("nats upstream close resets peer wait state", async () => {
  const transport = new NatsUpstreamTransport({
    upstream_nats_wait_timeout_ms: 5,
  });
  (transport as unknown as { handlePayload: (payload: string) => void }).handlePayload(
    `{"type":"modcdp.nats.hello","role":"browser","version":1}`,
  );

  await transport.waitForPeer();
  await transport.close();

  await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms for NATS ModCDP peer/);
});

test("nats upstream reconnects after close against a real NATS server", async () => {
  const nats = await startNatsServer();
  const transport = new NatsUpstreamTransport({
    upstream_nats_url: nats.url,
    upstream_nats_subject_prefix: `modcdp.reconnect.${Date.now()}`,
  });

  try {
    await transport.connect();
    assert.equal((transport as unknown as { connected: boolean }).connected, true);
    await transport.close();
    assert.equal((transport as unknown as { connected: boolean }).connected, false);
    await transport.connect();
    assert.equal((transport as unknown as { connected: boolean }).connected, true);
  } finally {
    await transport.close();
    await nats.close();
  }
}, 60_000);

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
  const dir = await mkdtemp(path.join(tmpdir(), "modcdp-nats-"));
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
  const proc = spawn(await getBinaryPath(), ["-c", config_path], {
    stdio: "ignore",
  });
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
