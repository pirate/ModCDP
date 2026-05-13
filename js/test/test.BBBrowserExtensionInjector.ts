import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, it } from "vitest";

import { ModCDPClient } from "../src/client/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

describe("BBBrowserExtensionInjector", () => {
  it(
    "uploads the real extension and launches a Browserbase browser with it installed",
    { timeout: 180_000 },
    async () => {
      assert.ok(process.env.BROWSERBASE_API_KEY?.trim(), "BROWSERBASE_API_KEY is required for live Browserbase tests");
      const cdp = new ModCDPClient({
        launcher: {
          launcher_mode: "bb",
          launcher_options: {
            timeout: 120,
            ...(process.env.BROWSERBASE_REGION ? { region: process.env.BROWSERBASE_REGION } : {}),
          },
        },
        upstream: { upstream_mode: "ws" },
        injector: {
          injector_mode: "inject",
          injector_extension_path: EXTENSION_PATH,
          injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
          injector_trust_service_worker_target: true,
        },
      });

      try {
        await cdp.connect();
        assert.equal(cdp.connect_timing?.injector_source, "bb");
        assert.equal(typeof cdp.extension_id, "string");
        const service_worker_url = await cdp.Mod.evaluate({
          expression: "chrome.runtime.getURL('modcdp/service_worker.js')",
        });
        assert.match(String(service_worker_url), /^chrome-extension:\/\/[a-z]{32}\/modcdp\/service_worker\.js$/);
      } finally {
        await cdp.close();
      }
    },
  );
});
