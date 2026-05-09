import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

test("ModCDPClient connects with nested launch/upstream/extension/client/server config", async () => {
  const cdp = new ModCDPClient({
    launch: {
      mode: "local",
      options: { headless: process.platform === "linux", sandbox: process.platform !== "linux" },
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
    assert.equal(cdp.extension.mode, "auto");
    assert.equal(cdp.client.routes["*.*"], "direct_cdp");
    assert.equal(cdp.upstream_endpoint_kind, "raw_cdp");
    assert.match(cdp.cdp_url ?? "", /^ws:\/\//);
    assert.equal(typeof (await cdp.Browser.getVersion()).product, "string");
  } finally {
    await cdp.close();
  }
}, 60_000);
