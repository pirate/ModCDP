import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { ModCDPClient } from "../src/client/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

test("BorrowedExtensionInjector bootstraps ModCDP inside a live extension service worker", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
  const cdp = new ModCDPClient({
    launcher: { launcher_mode: "remote" },
    upstream: { upstream_mode: "ws", upstream_cdp_url: chrome.cdp_url },
    injector: {
      injector_mode: "borrow",
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
  });

  try {
    await cdp.connect();
    assert.equal(cdp.connect_timing?.injector_source, "borrowed");
    assert.equal(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf");
    const target_infos = ((await cdp.send("Target.getTargets")) as { targetInfos?: unknown[] }).targetInfos ?? [];
    assert.ok(target_infos.length > 0);
  } finally {
    await cdp.close();
    await chrome.close();
  }
}, 60_000);
