import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

test("BorrowedExtensionInjector bootstraps ModCDP inside a live extension service worker", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: process.platform === "linux",
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
  const cdp = new ModCDPClient({
    launch: { mode: "remote" },
    upstream: { mode: "ws", ws_url: chrome.cdp_url },
    extension: {
      mode: "borrow",
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
    },
  });

  try {
    await cdp.connect();
    assert.equal(cdp.connect_timing?.extension_source, "borrowed");
    assert.equal(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf");
    const target_infos = ((await cdp.send("Target.getTargets")) as { targetInfos?: unknown[] }).targetInfos ?? [];
    assert.ok(target_infos.length > 0);
  } finally {
    await cdp.close();
    await chrome.close();
  }
}, 60_000);
