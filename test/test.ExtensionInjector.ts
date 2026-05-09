import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { ExtensionInjector, type ExtensionInjectionResult } from "../bridge/ExtensionInjector.js";
import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { CdpSocket } from "./helpers.BrowserLauncher.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

class ProbeExtensionInjector extends ExtensionInjector {
  async inject(): Promise<ExtensionInjectionResult | null> {
    return await this.waitForReadyServiceWorker(this.options.service_worker_ready_timeout_ms ?? 60_000, {
      matched_only: true,
    });
  }
}

test("ExtensionInjector probes a real extension service worker with shared base config", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: process.platform === "linux",
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
  const cdp = await CdpSocket.connect(chrome.ws_url!);
  const injector = new ProbeExtensionInjector({
    send: (method, params = {}, session_id = null) =>
      cdp.send(method, params as Record<string, unknown>, session_id ?? undefined),
    attachToTarget: async (target_id) => {
      const attached = await cdp.send("Target.attachToTarget", { targetId: target_id, flatten: true });
      return typeof attached.sessionId === "string" ? attached.sessionId : null;
    },
    extension_id: "mdedooklbnfejodmnhmkdpkaedafkehf",
    service_worker_url_suffixes: ["/modcdp/service_worker.js"],
    trust_matched_service_worker: true,
  });

  try {
    assert.deepEqual(injector.getLauncherConfig(), {});
    assert.deepEqual(injector.getTransportConfig(), { extension_id: "mdedooklbnfejodmnhmkdpkaedafkehf" });
    const result = await injector.inject();
    assert.equal(result?.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf");
    assert.equal(result?.url?.endsWith("/modcdp/service_worker.js"), true);
  } finally {
    await cdp.close();
    await injector.close();
    await chrome.close();
  }
}, 60_000);
