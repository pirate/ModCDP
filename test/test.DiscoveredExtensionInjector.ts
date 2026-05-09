import assert from "node:assert/strict";
import crypto from "node:crypto";
import { cp, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../bridge/LocalBrowserLauncher.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

function extensionIdFromManifestKey(public_key_der: Buffer) {
  const digest = crypto.createHash("sha256").update(public_key_der).digest().subarray(0, 16);
  const alphabet = "abcdefghijklmnop";
  return [...digest].map((byte) => alphabet[byte >> 4] + alphabet[byte & 0x0f]).join("");
}

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
  const { publicKey } = crypto.generateKeyPairSync("rsa", {
    modulusLength: 2048,
    publicKeyEncoding: { type: "spki", format: "der" },
    privateKeyEncoding: { type: "pkcs8", format: "pem" },
  });
  const custom_extension_id = extensionIdFromManifestKey(publicKey);
  await cp(EXTENSION_PATH, custom_extension_path, { recursive: true });
  const manifest_path = path.join(custom_extension_path, "manifest.json");
  const manifest = JSON.parse(await readFile(manifest_path, "utf8")) as Record<string, unknown>;
  manifest.key = publicKey.toString("base64");
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
      extension_id: custom_extension_id,
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
      require_service_worker_target: true,
    },
  });

  try {
    await cdp.connect();
    assert.equal(cdp.connect_timing?.extension_source, "discovered");
    assert.equal(cdp.extension_id, custom_extension_id);
    assert.equal(await cdp.Mod.evaluate({ expression: "chrome.runtime.id" }), custom_extension_id);

    const targets = (await cdp.sendRaw("Target.getTargets")) as {
      targetInfos: { type?: string; url?: string }[];
    };
    const modcdp_workers = targets.targetInfos.filter(
      (target) => target.type === "service_worker" && target.url?.endsWith("/modcdp/service_worker.js"),
    );
    assert.equal(
      modcdp_workers.some(
        (target) => target.url === `chrome-extension://${custom_extension_id}/modcdp/service_worker.js`,
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
