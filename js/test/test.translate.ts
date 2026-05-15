import assert from "node:assert/strict";
import { test } from "vitest";

import {
  CUSTOM_EVENT_BINDING_NAME,
  encodeBindingPayload,
  routeFor,
  unwrapEventIfNeeded,
  unwrapResponseIfNeeded,
  wrapCommandIfNeeded,
} from "../src/translate/translate.js";

test("translate routes, wraps, and unwraps ModCDP protocol messages deterministically", () => {
  assert.equal(routeFor("Browser.getVersion", { "Browser.*": "direct_cdp", "*.*": "service_worker" }), "direct_cdp");
  assert.equal(routeFor("Target.getTargets", { "Browser.*": "direct_cdp", "*.*": "service_worker" }), "service_worker");

  const direct = wrapCommandIfNeeded("Browser.getVersion", {}, { routes: { "*.*": "direct_cdp" } });
  assert.equal(direct.target, "direct_cdp");
  assert.deepEqual(direct.steps, [{ method: "Browser.getVersion", params: {} }]);

  const wrapped = wrapCommandIfNeeded(
    "Mod.evaluate",
    { expression: "({ ok: true })", params: { value: 1 } },
    { cdpSessionId: "session-1" },
  );
  assert.equal(wrapped.target, "service_worker");
  assert.equal(wrapped.steps[0]?.method, "Runtime.callFunctionOn");
  const wrapped_step_params = wrapped.steps[0]?.params as { functionDeclaration?: unknown } | undefined;
  assert.match(String(wrapped_step_params?.functionDeclaration), /attachToSession\("session-1"\)/);
  assert.equal(wrapped.steps[0]?.unwrap, "runtime");

  const configured = wrapCommandIfNeeded("Mod.configure", { server: { server_routes: { "*.*": "loopback_cdp" } } });
  assert.equal(configured.steps[0]?.unwrap, "runtime_json");

  assert.deepEqual(unwrapResponseIfNeeded({ result: { type: "object", value: { ok: true } } }, "runtime"), {
    ok: true,
  });
  assert.deepEqual(unwrapResponseIfNeeded({ product: "Chrome/1" }, null), { product: "Chrome/1" });

  const payload = encodeBindingPayload({
    event: "Custom.ready",
    data: { ready: true },
    cdpSessionId: "session-2",
  });
  assert.deepEqual(
    unwrapEventIfNeeded(
      "Runtime.bindingCalled",
      { name: CUSTOM_EVENT_BINDING_NAME, payload },
      "session-1",
      "session-1",
    ),
    { event: "Custom.ready", data: { ready: true }, sessionId: "session-2" },
  );
  assert.equal(unwrapEventIfNeeded("Runtime.consoleAPICalled", { name: CUSTOM_EVENT_BINDING_NAME, payload }), null);
});
