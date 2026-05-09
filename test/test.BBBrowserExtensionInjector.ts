import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, it } from "vitest";

import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");
const hasBrowserbaseEnv = Boolean(process.env.BROWSERBASE_API_KEY?.trim());

describe.skipIf(!hasBrowserbaseEnv)("BBBrowserExtensionInjector", () => {
  it(
    "uploads the real extension and launches a Browserbase browser with it installed",
    { timeout: 180_000 },
    async () => {
      const cdp = new ModCDPClient({
        launch: {
          mode: "bb",
          options: {
            project_id: process.env.BROWSERBASE_PROJECT_ID,
            timeout: 120,
            ...(process.env.BROWSERBASE_REGION ? { region: process.env.BROWSERBASE_REGION } : {}),
          },
        },
        upstream: { mode: "ws" },
        extension: {
          mode: "inject",
          path: EXTENSION_PATH,
          service_worker_url_suffixes: ["/modcdp/service_worker.js"],
          trust_service_worker_target: true,
        },
      });

      try {
        await cdp.connect();
        assert.equal(cdp.connect_timing?.extension_source, "bb");
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
