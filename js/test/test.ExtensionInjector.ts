import assert from "node:assert/strict";
import { test } from "vitest";

import { ExtensionInjector } from "../src/injector/ExtensionInjector.js";

class ProbeExtensionInjector extends ExtensionInjector {
  matches(target: { type?: string; url?: string }) {
    return this.serviceWorkerTargetMatches(target);
  }
}

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

test("ExtensionInjector base inject reports the subclass name", async () => {
  await assert.rejects(() => new ExtensionInjector().inject(), /ExtensionInjector\.inject is not implemented/);
});
