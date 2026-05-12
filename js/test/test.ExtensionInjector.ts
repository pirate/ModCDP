import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { ExtensionInjector, type ExtensionInjectionResult } from "../src/injector/ExtensionInjector.js";
import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { CdpSocket } from "./helpers.BrowserLauncher.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

class ProbeExtensionInjector extends ExtensionInjector {
  async inject(): Promise<ExtensionInjectionResult | null> {
    return await this.waitForReadyServiceWorker(this.options.injector_service_worker_ready_timeout_ms ?? 60_000, {
      matched_only: true,
    });
  }

  matches(target: { type?: string; url?: string }) {
    return this.serviceWorkerTargetMatches(target);
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
  const cdp = await CdpSocket.connect(chrome.cdp_url!);
  const injector = new ProbeExtensionInjector({
    send: (method, params = {}, session_id = null) =>
      cdp.send(method, params as Record<string, unknown>, session_id ?? undefined),
    attachToTarget: async (target_id) => {
      const attached = await cdp.send("Target.attachToTarget", { targetId: target_id, flatten: true });
      return typeof attached.sessionId === "string" ? attached.sessionId : null;
    },
    injector_extension_id: "mdedooklbnfejodmnhmkdpkaedafkehf",
    injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
    injector_trust_service_worker_target: true,
  });

  try {
    assert.deepEqual(injector.getLauncherConfig(), {});
    assert.deepEqual(injector.getTransportConfig(), { injector_extension_id: "mdedooklbnfejodmnhmkdpkaedafkehf" });
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
  const cdp = await CdpSocket.connect(chrome.cdp_url!);
  const injector = new ProbeExtensionInjector({
    send: (method, params = {}, session_id = null) =>
      cdp.send(method, params as Record<string, unknown>, session_id ?? undefined),
    attachToTarget: async (target_id) => {
      const attached = await cdp.send("Target.attachToTarget", { targetId: target_id, flatten: true });
      return typeof attached.sessionId === "string" ? attached.sessionId : null;
    },
    injector_extension_id: "mdedooklbnfejodmnhmkdpkaedafkehf",
    injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
    injector_trust_service_worker_target: true,
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

test("ExtensionInjector owns shared injector config", async () => {
  const injector = new ProbeExtensionInjector({
    injector_extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
  });

  try {
    assert.deepEqual(injector.getTransportConfig(), { injector_extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" });
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
  } finally {
    await injector.close();
  }
});

test("ExtensionInjector sendWithTimeout enforces cdp send timeout", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
  }).launch();
  const cdp = await CdpSocket.connect(chrome.cdp_url!);
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
  const cdp = await CdpSocket.connect(chrome.cdp_url!);
  const injector = new ProbeExtensionInjector({
    injector_extension_id: "mdedooklbnfejodmnhmkdpkaedafkehf",
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
