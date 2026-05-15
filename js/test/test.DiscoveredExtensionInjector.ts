import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { ModCDPClient } from "../src/client/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

test("DiscoveredExtensionInjector attaches to an already-loaded real ModCDP extension", async () => {
  const owner = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true },
    },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
  });
  const cdp = new ModCDPClient({
    launcher: { launcher_mode: "remote" },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "discover",
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
  });

  try {
    await owner.connect();
    cdp.upstream.upstream_cdp_url = owner.cdp_url;
    await cdp.connect();
    assert.equal(cdp.connect_timing?.injector_source, "discovered");
    assert.equal(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf");
    assert.equal(
      await cdp.Mod.evaluate({ expression: "chrome.runtime.getURL('modcdp/service_worker.js')" }),
      "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js",
    );
  } finally {
    await cdp.close();
    await owner.close();
  }
}, 60_000);
