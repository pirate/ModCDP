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

  async sendTimed(method: string, params: Record<string, unknown>, session_id: string | null, timeout_ms: number) {
    return await this.sendWithTimeout(method, params, session_id, timeout_ms);
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

test("ExtensionInjector keeps the ModCDP service worker alive through offscreen keepalive", async () => {
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
    const result = await injector.inject();
    assert.equal(result?.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf");
    const session_id = result?.session_id;
    assert.equal(typeof session_id, "string");

    let contexts: { type?: string; url?: string }[] = [];
    for (let attempt = 0; attempt < 50; attempt++) {
      const evaluated = await cdp.send(
        "Runtime.evaluate",
        {
          expression:
            "chrome.runtime.getContexts({}).then((contexts) => contexts.map((context) => ({ type: context.contextType, url: context.documentUrl || context.origin || '' })))",
          awaitPromise: true,
          returnByValue: true,
        },
        session_id,
      );
      contexts = (evaluated.result as { value?: { type?: string; url?: string }[] }).value ?? [];
      if (
        contexts.some(
          (context) =>
            context.type === "OFFSCREEN_DOCUMENT" &&
            context.url === "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/offscreen/keepalive.html",
        )
      ) {
        break;
      }
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
    assert.equal(
      contexts.some(
        (context) =>
          context.type === "OFFSCREEN_DOCUMENT" &&
          context.url === "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/offscreen/keepalive.html",
      ),
      true,
    );

    await new Promise((resolve) => setTimeout(resolve, 3_000));
    const targets = (await cdp.send("Target.getTargets")) as { targetInfos?: { type?: string; url?: string }[] };
    assert.equal(
      targets.targetInfos?.some(
        (target) =>
          target.type === "service_worker" &&
          target.url === "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js",
      ),
      true,
    );
    const version = await cdp.send(
      "Runtime.evaluate",
      { expression: "globalThis.ModCDP?.__ModCDPServerVersion", returnByValue: true },
      session_id,
    );
    assert.equal((version.result as { value?: unknown }).value, 1);
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
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
  }).launch();
  const cdp = await CdpSocket.connect(chrome.ws_url!);
  let target_id: string | null = null;
  const injector = new ProbeExtensionInjector({
    send: (method, params = {}, session_id = null) =>
      cdp.send(method, params as Record<string, unknown>, session_id ?? undefined),
  });

  try {
    const created = await cdp.send("Target.createTarget", { url: "about:blank#modcdp-timeout" });
    target_id = created.targetId as string;
    const attached = await cdp.send("Target.attachToTarget", { targetId: target_id, flatten: true });
    const session_id = attached.sessionId as string;
    await cdp.send("Runtime.enable", {}, session_id);
    await assert.rejects(
      () =>
        injector.sendTimed(
          "Runtime.evaluate",
          { expression: "new Promise(() => {})", awaitPromise: true },
          session_id,
          5,
        ),
      /Runtime\.evaluate timed out after 5ms/,
    );
  } finally {
    if (target_id) await cdp.send("Target.closeTarget", { targetId: target_id }).catch(() => ({}));
    await injector.close();
    await cdp.close();
    await chrome.close();
  }
});

test("ExtensionInjector wakes configured extension with a hidden background target", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
  const cdp = await CdpSocket.connect(chrome.ws_url!);
  const injector = new ProbeExtensionInjector({
    extension_id: "mdedooklbnfejodmnhmkdpkaedafkehf",
    send: (method, params = {}, session_id = null) =>
      cdp.send(method, params as Record<string, unknown>, session_id ?? undefined),
  });

  try {
    assert.equal(await injector.wake(), true);
    const targets = (await cdp.send("Target.getTargets")) as { targetInfos?: { url?: string }[] };
    assert.equal(
      targets.targetInfos?.some(
        (target) => target.url === "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/wake.html",
      ),
      true,
    );
  } finally {
    await injector.close();
    await cdp.close();
    await chrome.close();
  }
});

test("ExtensionInjector base inject reports the subclass name", async () => {
  await assert.rejects(() => new ExtensionInjector().inject(), /ExtensionInjector\.inject is not implemented/);
});
