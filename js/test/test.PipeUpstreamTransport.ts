import assert from "node:assert/strict";
import path from "node:path";
import { PassThrough } from "node:stream";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { PipeUpstreamTransport } from "../src/transport/PipeUpstreamTransport.js";
import { ModCDPClient } from "../src/client/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

test("pipe upstream constructor, update, launcher config, and unconnected errors match the transport surface", async () => {
  const transport = new PipeUpstreamTransport({ cdp_url: "pipe://constructor" });
  assert.equal(transport.mode, "pipe");
  assert.equal(transport.endpoint_kind, "raw_cdp");
  assert.equal(transport.url, "pipe://constructor");
  assert.deepEqual(transport.getLauncherConfig(), { remote_debugging: "pipe" });
  assert.equal(transport.update({ cdp_url: "pipe://1234" }), transport);
  assert.equal(transport.url, "pipe://1234");
  await assert.rejects(() => transport.connect(), /upstream\.upstream_mode=pipe requires/);
  assert.throws(() => transport.send({ id: 1, method: "Browser.getVersion" }), /CDP pipe is not connected/);
});

test("pipe upstream resets connection state after pipe end and errors", async () => {
  for (const event_name of ["end", "read_error", "write_error"] as const) {
    const pipe_read = new PassThrough();
    const pipe_write = new PassThrough();
    const transport = new PipeUpstreamTransport({ pipe_read, pipe_write, cdp_url: "pipe://test" });
    const closed: Error[] = [];
    transport.onClose((error) => closed.push(error));

    await transport.connect();
    transport.send({ id: 1, method: "Browser.getVersion", params: {} });

    if (event_name === "end") pipe_read.emit("end");
    else if (event_name === "read_error") pipe_read.emit("error", new Error("read failed"));
    else pipe_write.emit("error", new Error("write failed"));

    assert.equal(closed.length, 1);
    assert.throws(
      () => transport.send({ id: 2, method: "Browser.getVersion", params: {} }),
      /CDP pipe is not connected/,
    );
    await transport.close();
  }
});

test("pipe upstream launches a real browser and uses a pid-scoped pipe URL", async () => {
  const cdp = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { upstream_mode: "pipe" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
  });

  try {
    await cdp.connect();
    assert.equal(cdp.transport?.mode, "pipe");
    assert.equal(cdp.upstream_endpoint_kind, "raw_cdp");
    assert.match(cdp.cdp_url ?? "", /^pipe:\/\/\d+$/);
    assert.equal(cdp.transport?.url, cdp.cdp_url);
    const version = (await cdp.sendRaw("Browser.getVersion")) as Record<string, unknown>;
    assert.equal(typeof version.product, "string");
  } finally {
    await cdp.close();
  }
}, 60_000);
