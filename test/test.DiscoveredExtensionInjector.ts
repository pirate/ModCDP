import assert from "node:assert/strict";
import { cp, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");
const CUSTOM_EXTENSION_ID = "hhklgmbgnbeghnjidampacgmgnhelifg";
const CUSTOM_EXTENSION_PUBLIC_KEY =
  "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzG1LUbtH0aHMKjTAUeT0saY8xfnRNENctFJme3C1qnsqT7PAXMxJC4nT7tBZy2gEGRirBb3zIZ3OyAu9a0QR8lTLupDp4qHWOhQ7dl9ZjxjQdYa4Gby0xuXLdQrJIxDbmuv+UVJvYa8vRTwQB8koygbzDDDP5/YiB6mc0hbh8XBb82Ossy7T280k8280o/rS0CXdioUraCHj58PDhfxbs18TBcYfOjuRqua9J2oddxobtGehSD0gDtbvn2IWDtRajOlgZZyuS1vLoSR7C1ulFzpRSYPEMhI2x+wphut7E3QImyJ577YeULVGpt988FcixOou7udjx3/IUWjpq8046wIDAQAB";

test("DiscoveredExtensionInjector attaches to an already-loaded real ModCDP extension", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
  const cdp = new ModCDPClient({
    launch: { mode: "remote" },
    upstream: { mode: "ws", ws_url: chrome.cdp_url },
    extension: {
      mode: "discover",
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
    },
  });

  try {
    await cdp.connect();
    assert.equal(cdp.connect_timing?.extension_source, "discovered");
    assert.equal(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf");
    const service_worker_url = await cdp.Mod.evaluate({
      expression: "chrome.runtime.getURL('modcdp/service_worker.js')",
    });
    assert.match(
      String(service_worker_url),
      /^chrome-extension:\/\/mdedooklbnfejodmnhmkdpkaedafkehf\/modcdp\/service_worker\.js$/,
    );
  } finally {
    await cdp.close();
    await chrome.close();
  }
}, 60_000);

test("DiscoveredExtensionInjector selects the configured extension when multiple ModCDP workers exist", async () => {
  const custom_extension_path = await mkdtemp(path.join(tmpdir(), "modcdp-custom-extension-"));
  await cp(EXTENSION_PATH, custom_extension_path, { recursive: true });
  const manifest_path = path.join(custom_extension_path, "manifest.json");
  const manifest = JSON.parse(await readFile(manifest_path, "utf8")) as Record<string, unknown>;
  manifest.key = CUSTOM_EXTENSION_PUBLIC_KEY;
  manifest.name = "ModCDP Bridge Custom Test";
  await writeFile(manifest_path, `${JSON.stringify(manifest, null, 2)}\n`);

  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH},${custom_extension_path}`],
  }).launch();
  const cdp = new ModCDPClient({
    launch: { mode: "remote" },
    upstream: { mode: "ws", ws_url: chrome.cdp_url },
    extension: {
      mode: "discover",
      extension_id: CUSTOM_EXTENSION_ID,
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
      require_service_worker_target: true,
    },
  });

  try {
    await cdp.connect();
    assert.equal(cdp.connect_timing?.extension_source, "discovered");
    assert.equal(cdp.extension_id, CUSTOM_EXTENSION_ID);
    assert.equal(await cdp.Mod.evaluate({ expression: "chrome.runtime.id" }), CUSTOM_EXTENSION_ID);

    const targets = (await cdp.sendRaw("Target.getTargets")) as {
      targetInfos: { type?: string; url?: string }[];
    };
    const modcdp_workers = targets.targetInfos.filter(
      (target) => target.type === "service_worker" && target.url?.endsWith("/modcdp/service_worker.js"),
    );
    assert.equal(
      modcdp_workers.some(
        (target) => target.url === `chrome-extension://${CUSTOM_EXTENSION_ID}/modcdp/service_worker.js`,
      ),
      true,
    );
    assert.equal(
      modcdp_workers.some(
        (target) => target.url === "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js",
      ),
      true,
    );
  } finally {
    await cdp.close();
    await chrome.close();
    await rm(custom_extension_path, { recursive: true, force: true });
  }
}, 60_000);
