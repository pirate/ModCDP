import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { ExtensionsLoadUnpackedInjector } from "../bridge/ExtensionsLoadUnpackedInjector.js";
import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { CdpSocket } from "./helpers.BrowserLauncher.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

test("ExtensionsLoadUnpackedInjector exercises the real CDP loadUnpacked path", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: process.platform === "linux",
    sandbox: process.platform !== "linux",
  }).launch();
  const cdp = await CdpSocket.connect(chrome.ws_url!);
  const injector = new ExtensionsLoadUnpackedInjector({
    send: (method, params = {}, session_id = null) =>
      cdp.send(method, params as Record<string, unknown>, session_id ?? undefined),
    extension_path: EXTENSION_PATH,
  });

  try {
    await injector.prepare();
    const result = await injector.inject();
    assert.equal(result, null);
    assert.match(injector.last_error?.message ?? "", /Method not available|Method.*not.*found|wasn't found/i);
  } finally {
    await cdp.close();
    await injector.close();
    await chrome.close();
  }
}, 60_000);
