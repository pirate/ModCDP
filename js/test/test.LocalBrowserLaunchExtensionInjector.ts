import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { DEFAULT_MODCDP_EXTENSION_ID } from "../src/injector/ExtensionInjector.js";
import { LocalBrowserLaunchExtensionInjector } from "../src/injector/LocalBrowserLaunchExtensionInjector.js";
import { ModCDPClient } from "../src/client/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

test("LocalBrowserLaunchExtensionInjector loads the real extension during local launch", async () => {
  const cdp = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: {
        headless: process.platform === "linux" && !process.env.DISPLAY,
        sandbox: process.platform !== "linux",
      },
    },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "inject",
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
      injector_service_worker_probe_timeout_ms: 30_000,
    },
    client: {
      client_cdp_send_timeout_ms: 30_000,
    },
  });

  try {
    await cdp.connect();
    assert.equal(cdp.connect_timing?.injector_source, "local_launch");
    assert.equal(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf");
    assert.match(cdp.ext_session_id ?? "", /^.+$/);
    const service_worker_url = await cdp.Mod.evaluate({
      expression: "chrome.runtime.getURL('modcdp/service_worker.js')",
    });
    assert.match(
      String(service_worker_url),
      /^chrome-extension:\/\/mdedooklbnfejodmnhmkdpkaedafkehf\/modcdp\/service_worker\.js$/,
    );
  } finally {
    await cdp.close();
  }
}, 60_000);

test("LocalBrowserLaunchExtensionInjector prepares launcher config", async () => {
  const injector = new LocalBrowserLaunchExtensionInjector({ injector_extension_path: EXTENSION_PATH });

  try {
    await injector.prepare();
    const extra_args = injector.getLauncherConfig().extra_args ?? [];
    assert.equal(extra_args.length, 1);
    assert.match(extra_args[0], /^--load-extension=/);
    assert.equal(injector.options.injector_extension_id, DEFAULT_MODCDP_EXTENSION_ID);
  } finally {
    await injector.close();
  }
});

test("LocalBrowserLaunchExtensionInjector falls back to the default extension zip", async () => {
  const injector = new LocalBrowserLaunchExtensionInjector();

  try {
    await injector.prepare();
    const extra_args = injector.getLauncherConfig().extra_args ?? [];
    assert.equal(extra_args.length, 1);
    assert.match(extra_args[0], /^--load-extension=.*modcdp-extension-/);
    assert.equal(injector.options.injector_extension_id, DEFAULT_MODCDP_EXTENSION_ID);
  } finally {
    await injector.close();
  }
});
