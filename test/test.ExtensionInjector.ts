import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { mkdtemp, rm } from "node:fs/promises";
import os from "node:os";
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

  matches(target: { type?: string; url?: string }) {
    return this.serviceWorkerTargetMatches(target);
  }

  writeRuntimeConfig(extension_path: string) {
    this.writeExtensionRuntimeConfig(extension_path);
  }

  async sendTimed(method: string, timeout_ms: number) {
    return await this.sendWithTimeout(method, {}, null, timeout_ms);
  }

  async wake() {
    return await this.wakeConfiguredExtension();
  }
}

test("ExtensionInjector probes a real extension service worker with shared base config", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
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

test("ExtensionInjector owns shared injector config and runtime transport config", async () => {
  const injector = new ProbeExtensionInjector({
    extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    service_worker_url_suffixes: ["/modcdp/service_worker.js"],
    reverse_proxy_url: "ws://127.0.0.1:29292",
    nats_url: "ws://127.0.0.1:4223",
  });
  injector.update({ native_host_name: "com.modcdp.bridge" });

  const runtime_config_dir = await mkdtemp(path.join(os.tmpdir(), "modcdp-extension-"));
  try {
    assert.deepEqual(injector.getTransportConfig(), { extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" });
    assert.deepEqual(injector.getLauncherConfig(), {});
    assert.equal(
      injector.matches({
        type: "service_worker",
        url: "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/modcdp/service_worker.js",
      }),
      true,
    );
    assert.equal(
      injector.matches({
        type: "service_worker",
        url: "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/background.js",
      }),
      false,
    );

    injector.writeRuntimeConfig(runtime_config_dir);
    assert.deepEqual(JSON.parse(readFileSync(path.join(runtime_config_dir, "modcdp", "config.json"), "utf8")), {
      reverse_proxy_url: "ws://127.0.0.1:29292",
      native_host_name: "com.modcdp.bridge",
      nats_url: "ws://127.0.0.1:4223",
    });
    assert.match(
      readFileSync(path.join(runtime_config_dir, "config.js"), "utf8"),
      /globalThis\.__MODCDP_RUNTIME_CONFIG__/,
    );
  } finally {
    await injector.close();
    await rm(runtime_config_dir, { recursive: true, force: true });
  }
});

test("ExtensionInjector sendWithTimeout enforces cdp send timeout", async () => {
  const injector = new ProbeExtensionInjector({
    send: async () => {
      await new Promise((resolve) => setTimeout(resolve, 50));
      return {};
    },
  });

  await assert.rejects(() => injector.sendTimed("Runtime.evaluate", 5), /Runtime\.evaluate timed out after 5ms/);
});

test("ExtensionInjector wakes configured extension with a hidden background target", async () => {
  const sent: { method: string; params: Record<string, unknown> }[] = [];
  const injector = new ProbeExtensionInjector({
    extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    send: async (method, params = {}) => {
      sent.push({ method, params: params as Record<string, unknown> });
      return { targetId: "wake-target" };
    },
  });

  assert.equal(await injector.wake(), true);
  assert.deepEqual(sent, [
    {
      method: "Target.createTarget",
      params: {
        url: "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/modcdp/wake.html",
        background: true,
        hidden: true,
        focus: false,
      },
    },
  ]);
});

test("ExtensionInjector base inject reports the subclass name", async () => {
  await assert.rejects(() => new ExtensionInjector().inject(), /ExtensionInjector\.inject is not implemented/);
});
