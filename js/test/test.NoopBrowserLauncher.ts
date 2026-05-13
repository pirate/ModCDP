import assert from "node:assert/strict";
import { test } from "vitest";

import { NoopBrowserLauncher } from "../src/launcher/NoopBrowserLauncher.js";

test("NoopBrowserLauncher records an empty launched browser", async () => {
  const launcher = new NoopBrowserLauncher();
  const launched = await launcher.launch();

  assert.equal(launched.cdp_url, null);
  assert.equal(launcher.launched, launched);
  assert.deepEqual(launcher.getTransportConfig(), {
    cdp_url: null,
    user_data_dir: null,
    pipe_read: null,
    pipe_write: null,
  });
  await launched.close();
});
