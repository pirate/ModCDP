import assert from "node:assert/strict";
import { spawn, type ChildProcess } from "node:child_process";
import { once } from "node:events";
import { writeFile, mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { getBinaryPath } from "@eplightning/nats-server";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { NatsUpstreamTransport } from "../bridge/NatsUpstreamTransport.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

test("nats upstream config owns url, subject prefix, wait timeout, and injector config", async () => {
  const transport = new NatsUpstreamTransport({ url: "ws://127.0.0.1:4223", subject_prefix: "modcdp.one" });
  assert.equal(transport.url, "ws://127.0.0.1:4223/");
  assert.equal(transport.subject_prefix, "modcdp.one");
  assert.deepEqual(transport.getInjectorConfig(), {
    nats_url: "ws://127.0.0.1:4223/",
    nats_subject_prefix: "modcdp.one",
  });
  assert.equal(
    transport.update({
      nats_url: "nats://127.0.0.1:4222",
      nats_subject_prefix: "modcdp.two",
      role: "browser",
      wait_timeout_ms: 5,
    }),
    transport,
  );
  assert.equal(transport.url, "nats://127.0.0.1:4222");
  assert.equal(transport.subject_prefix, "modcdp.two");
  await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms for NATS ModCDP peer/);
});

test("nats upstream close rejects pending peer waits", async () => {
  const transport = new NatsUpstreamTransport({
    url: "ws://127.0.0.1:4223",
    subject_prefix: "modcdp.close",
    wait_timeout_ms: 5_000,
  });
  const pending = transport.waitForPeer();

  await transport.close();

  await assert.rejects(() => pending, /NATS transport for modcdp\.close closed before a peer connected/);
});

test("nats upstream reconnects after close against a real NATS server", async () => {
  const nats = await start_nats_server();
  const transport = new NatsUpstreamTransport({
    url: nats.url,
    subject_prefix: `modcdp.reconnect.${Date.now()}`,
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

async function wait_for_websocket(url: string, timeout_ms = 10_000) {
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

async function close_process(proc: ChildProcess) {
  if (proc.exitCode != null || proc.signalCode != null) return;
  proc.kill("SIGTERM");
  await Promise.race([once(proc, "exit"), delay(2_000)]);
  if (proc.exitCode != null || proc.signalCode != null) return;
  proc.kill("SIGKILL");
  await Promise.race([once(proc, "exit"), delay(2_000)]);
}

async function start_nats_server() {
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
  const proc = spawn(await getBinaryPath(), ["-c", config_path], { stdio: "ignore" });
  const url = `ws://127.0.0.1:${websocket_port}`;
  try {
    await wait_for_websocket(url);
  } catch (error) {
    await close_process(proc);
    await rm(dir, { recursive: true, force: true });
    throw error;
  }
  return {
    url,
    close: async () => {
      await close_process(proc);
      await rm(dir, { recursive: true, force: true });
    },
  };
}

test("nats upstream relays CDP through a real extension over a real NATS server", async () => {
  const nats = await start_nats_server();
  const subject_prefix = `modcdp.test.${Date.now()}`;
  const nats_client = new ModCDPClient({
    launch: {
      mode: "local",
      options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: {
      mode: "nats",
      nats_url: nats.url,
      nats_subject_prefix: subject_prefix,
    },
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
    await nats_client.connect();
    assert.equal(nats_client.transport?.mode, "nats");
    assert.equal(nats_client.upstream_endpoint_kind, "modcdp_server");
    assert.equal(nats_client.transport?.url, `${nats.url}/`);
    assert.equal(nats_client.upstream.nats_subject_prefix, subject_prefix);
    const version = (await nats_client.send("Browser.getVersion")) as Record<string, unknown>;
    assert.equal(typeof version.product, "string");
  } finally {
    await nats_client.close();
    await nats.close();
  }
}, 90_000);
