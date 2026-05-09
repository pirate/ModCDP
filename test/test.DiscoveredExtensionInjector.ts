import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

test("DiscoveredExtensionInjector attaches to an already-loaded real ModCDP extension", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: process.platform === "linux",
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
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
    assert.equal(cdp.connect_timing?.extension_source, "discovered");
    assert.equal(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf");
    const service_worker_url = await cdp.Mod.evaluate({
      expression: "chrome.runtime.getURL('modcdp/service_worker.js')",
    });
    assert.match(
      String(service_worker_url),
      /^chrome-extension:\/\/mdedooklbnfejodmnhmkdpkaedafkehf\/modcdp\/service_worker\.js$/,
    );
  } finally {
    await cdp.close();
    await chrome.close();
  }
}, 60_000);
