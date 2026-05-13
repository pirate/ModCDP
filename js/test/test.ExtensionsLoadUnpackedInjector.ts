import assert from "node:assert/strict";
import { test } from "vitest";

import { ExtensionsLoadUnpackedInjector } from "../src/injector/ExtensionsLoadUnpackedInjector.js";
import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { CdpSocket } from "./helpers.BrowserLauncher.js";

test("ExtensionsLoadUnpackedInjector exercises the real CDP loadUnpacked path", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
  }).launch();
  const cdp = await CdpSocket.connect(chrome.cdp_url!);
  const injector = new ExtensionsLoadUnpackedInjector({
    send: (method, params = {}, session_id = null) =>
      cdp.send(method, params as Record<string, unknown>, session_id ?? undefined),
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
