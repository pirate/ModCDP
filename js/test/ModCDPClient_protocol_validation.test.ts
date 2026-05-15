import assert from "node:assert/strict";
import { test } from "vitest";
import { z } from "zod";

import { ModCDPClient, type ModCDPClientInstance } from "../src/client/ModCDPClient.js";
import type { cdp } from "../src/types/generated/cdp.js";

test("protocol validation covers native methods, native events, custom methods, custom events, and native overrides", async () => {
  type CustomClient = ModCDPClientInstance<
    {
      "Custom.echo": { params: { text: string }; result: { ok: boolean } };
      "Target.getTargets": {
        params: cdp.types.ts.Target.GetTargetsParams;
        result: cdp.types.ts.Target.GetTargetsResult & {
          targetInfos: Array<cdp.types.ts.Target.TargetInfo & { tabId?: number }>;
        };
      };
    },
    { "Custom.ready": { ok: boolean } }
  >;

  const client = new ModCDPClient() as CustomClient;

  const runtimeParams: cdp.types.ts.Runtime.EvaluateParams = { expression: "1 + 1", returnByValue: true };
  const runtimeResult: cdp.types.ts.Runtime.EvaluateResult = {
    result: { type: "number", value: 2, description: "2" },
  };
  const nativeEvent: cdp.types.ts.Target.TargetCreatedEvent = {
    targetInfo: {
      targetId: "target-1",
      type: "page",
      title: "Example",
      url: "https://example.com",
      attached: false,
      canAccessOpener: false,
    },
  };
  const customParams: Parameters<CustomClient["Custom"]["echo"]>[0] = { text: "ok" };
  const customResult: Awaited<ReturnType<CustomClient["Custom"]["echo"]>> = { ok: true };

  if (false) {
    // @ts-expect-error Runtime.evaluate params require expression.
    const badRuntimeParams: cdp.types.ts.Runtime.EvaluateParams = {};
    void badRuntimeParams;
    const badNativeEvent: cdp.types.ts.Target.TargetCreatedEvent = {
      targetInfo: {
        ...nativeEvent.targetInfo,
        // @ts-expect-error Target.targetCreated targetInfo.targetId is a string.
        targetId: 1,
      },
    };
    void badNativeEvent;
    // @ts-expect-error Custom.echo requires text.
    client.Custom.echo({ id: "wrong" });
    // @ts-expect-error Custom.echo returns ok as boolean.
    const badCustomResult: Awaited<ReturnType<CustomClient["Custom"]["echo"]>> = { ok: "yes" };
    void badCustomResult;
    await client.Mod.addMiddleware({
      name: client.Target.getTargets,
      phase: client.RESPONSE,
      expression: "async (value, next) => next(value)",
    });
  }

  assert.deepEqual(client.command_params_schemas.get("Runtime.evaluate")?.parse(runtimeParams), runtimeParams);
  assert.deepEqual(client.command_result_schemas.get("Runtime.evaluate")?.parse(runtimeResult), runtimeResult);
  assert.deepEqual(client.event_schemas.get("Target.targetCreated")?.parse(nativeEvent), nativeEvent);
  assert.throws(() => client.command_params_schemas.get("Runtime.evaluate")?.parse({}));
  assert.throws(() => client.command_result_schemas.get("Runtime.evaluate")?.parse({}));
  assert.throws(() => client.event_schemas.get("Target.targetCreated")?.parse({}));

  await client.Mod.addCustomCommand("Custom.echo", {
    params_schema: z.object({ text: z.string() }),
    result_schema: z.object({ ok: z.boolean() }),
  });
  await client.Mod.addCustomEvent("Custom.ready", { event_schema: z.object({ ok: z.boolean() }) });

  assert.deepEqual(client.command_params_schemas.get("Custom.echo")?.parse(customParams), customParams);
  assert.deepEqual(client.command_result_schemas.get("Custom.echo")?.parse(customResult), customResult);
  assert.deepEqual(client.event_schemas.get("Custom.ready")?.parse({ ok: true }), { ok: true });
  assert.throws(() => client.command_params_schemas.get("Custom.echo")?.parse({ text: 1 }));
  assert.throws(() => client.command_result_schemas.get("Custom.echo")?.parse({ ok: "yes" }));
  assert.throws(() => client.event_schemas.get("Custom.ready")?.parse({ ok: "yes" }));

  await client.Mod.addCustomCommand({
    name: client.Target.getTargets,
    result_schema: z.object({
      targetInfos: z.array(
        z.object({
          targetId: z.string(),
          type: z.string(),
          title: z.string(),
          url: z.string(),
          attached: z.boolean(),
          canAccessOpener: z.boolean(),
          tabId: z.number().optional(),
        }),
      ),
    }),
  });
  await client.Mod.addCustomEvent({ name: client.Target.targetCreated });

  const extendedTargets: Awaited<ReturnType<CustomClient["Target"]["getTargets"]>> = {
    targetInfos: [{ ...nativeEvent.targetInfo, tabId: 7 }],
  };
  assert.deepEqual(client.command_result_schemas.get("Target.getTargets")?.parse(extendedTargets), extendedTargets);
  assert.deepEqual(
    client.event_schemas.get("Target.targetCreated")?.parse({ targetInfo: { ...nativeEvent.targetInfo, tabId: 7 } }),
    {
      targetInfo: { ...nativeEvent.targetInfo, tabId: 7 },
    },
  );
});
