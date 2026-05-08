import assert from "node:assert/strict";
import test from "node:test";
import { z } from "zod";

import { ModCDPClient, type ModCDPClientInstance } from "../client/js/ModCDPClient.js";

test("schema-only custom commands install flat namespace methods", async () => {
  type CustomClient = ModCDPClientInstance<{
    "Custom.doSomething": { params: { id: string }; result: boolean };
  }>;
  const cdp = new ModCDPClient({ routes: { "Custom.*": "direct_cdp" } }) as CustomClient;
  cdp._sendRaw = async () => ({ success: true });

  await cdp.Mod.addCustomCommand("Custom.doSomething", {
    params_schema: z.object({ id: z.string() }),
    result_schema: z.object({ success: z.boolean() }),
  });

  const success: boolean = await cdp.Custom.doSomething({ id: "abc" });
  const rawSuccess: boolean = Boolean(await cdp.send("Custom.doSomething", { id: "abc" }));

  assert.equal(success, true);
  assert.equal(rawSuccess, true);
  // @ts-expect-error typed custom command params reject non-string ids statically.
  await assert.rejects(() => cdp.Custom.doSomething({ id: 123 }));
});

test("schema-only custom events validate raw string handlers", async () => {
  const cdp = new ModCDPClient();
  const EventSchema = z.object({ data: z.string() });
  type Event = z.infer<typeof EventSchema>;
  type CustomClient = ModCDPClientInstance<Record<never, never>, { "Custom.someEvent": Event }>;
  const typedCdp = cdp as CustomClient;
  const seen: string[] = [];

  await cdp.Mod.addCustomEvent("Custom.someEvent", { event_schema: EventSchema });
  typedCdp.on("Custom.someEvent", (event: Event) => {
    seen.push(event.data);
  });

  cdp.emit("Custom.someEvent", EventSchema.parse({ data: "ok" }));
  assert.deepEqual(seen, ["ok"]);
});
