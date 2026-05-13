import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";
import { z } from "zod";

import { ModCDPClient, type ModCDPClientInstance } from "../src/client/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

test("custom commands install flat namespace methods through a real service worker", async () => {
  type CustomClient = ModCDPClientInstance<{
    "Custom.doSomething": { params: { id: string }; result: boolean };
  }>;
  const cdp = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
    client: { client_routes: { "Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp" } },
    server: { server_routes: { "*.*": "loopback_cdp" } },
  }) as CustomClient;

  try {
    await cdp.connect();
    await cdp.Mod.addCustomCommand("Custom.doSomething", {
      params_schema: z.object({ id: z.string() }),
      result_schema: z.object({ success: z.boolean() }),
      expression: "async ({ id }) => ({ success: id === 'abc' })",
    });

    const success: boolean = await cdp.Custom.doSomething({ id: "abc" });
    const rawSuccess: boolean = Boolean(await cdp.send("Custom.doSomething", { id: "abc" }));

    assert.equal(success, true);
    assert.equal(rawSuccess, true);
    // @ts-expect-error typed custom command params reject non-string ids statically.
    await assert.rejects(() => cdp.Custom.doSomething({ id: 123 }));
  } finally {
    await cdp.close();
  }
}, 60_000);

test("custom events validate raw string handlers through a real service worker", async () => {
  const EventSchema = z.object({ data: z.string() });
  type Event = z.infer<typeof EventSchema>;
  type CustomClient = ModCDPClientInstance<Record<never, never>, { "Custom.someEvent": Event }>;
  const cdp = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
    client: { client_routes: { "Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp" } },
    server: { server_routes: { "*.*": "loopback_cdp" } },
  }) as CustomClient;
  const seen: string[] = [];

  try {
    await cdp.connect();
    await cdp.Mod.addCustomEvent("Custom.someEvent", { event_schema: EventSchema });
    const received = new Promise<void>((resolve) => {
      cdp.on("Custom.someEvent", (event: Event) => {
        seen.push(event.data);
        resolve();
      });
    });

    await cdp.Mod.evaluate({
      expression: "async () => await globalThis.ModCDP.emit('Custom.someEvent', { data: 'ok' })",
    });
    await received;
    assert.deepEqual(seen, ["ok"]);
  } finally {
    await cdp.close();
  }
}, 60_000);

test("schema-only custom commands register without a websocket", async () => {
  const cdp = new ModCDPClient();

  const result = await cdp.send("Mod.addCustomCommand", {
    name: "Custom.echo",
    params_schema: {
      type: "object",
      properties: { text: { type: "string", minLength: 1 } },
      required: ["text"],
      additionalProperties: false,
    },
    result_schema: {
      type: "object",
      properties: { text: { type: "string" } },
      required: ["text"],
      additionalProperties: false,
    },
  });

  assert.deepEqual(result, { name: "Custom.echo", registered: true });
  const command_params_schemas = (cdp as unknown as { command_params_schemas: Map<string, z.ZodType> })
    .command_params_schemas;
  const command_result_schemas = (cdp as unknown as { command_result_schemas: Map<string, z.ZodType> })
    .command_result_schemas;
  assert.deepEqual(command_params_schemas.get("Custom.echo")?.parse({ text: "ok" }), { text: "ok" });
  assert.throws(() => command_params_schemas.get("Custom.echo")?.parse({ text: "" }));
  assert.throws(() => command_params_schemas.get("Custom.echo")?.parse({ text: "ok", extra: true }));
  assert.deepEqual(command_result_schemas.get("Custom.echo")?.parse({ text: "ok" }), { text: "ok" });
  assert.throws(() => command_result_schemas.get("Custom.echo")?.parse({ text: 123 }));
});

test("constructor custom command and event schemas validate nested payloads", () => {
  const cdp = new ModCDPClient({
    custom_commands: [
      {
        name: "Custom.collect",
        params_schema: {
          type: "object",
          properties: {
            items: {
              type: "array",
              minItems: 1,
              items: {
                type: "object",
                properties: {
                  id: { type: "string" },
                  count: { type: "integer", minimum: 1 },
                },
                required: ["id", "count"],
                additionalProperties: false,
              },
            },
          },
          required: ["items"],
          additionalProperties: false,
        },
      },
    ],
    custom_events: [
      {
        name: "Custom.ready",
        event_schema: {
          type: "object",
          properties: {
            url: { type: "string", pattern: "^https://" },
            ready: { type: "boolean" },
          },
          required: ["url", "ready"],
          additionalProperties: false,
        },
      },
      { name: "Custom.count", event_schema: { type: "integer", minimum: 1 } },
    ],
  });
  const command_params_schemas = (cdp as unknown as { command_params_schemas: Map<string, z.ZodType> })
    .command_params_schemas;
  const event_schemas = (cdp as unknown as { event_schemas: Map<string, z.ZodType> }).event_schemas;

  const valid_params = { items: [{ id: "a", count: 1 }] };
  assert.deepEqual(command_params_schemas.get("Custom.collect")?.parse(valid_params), valid_params);
  assert.throws(() => command_params_schemas.get("Custom.collect")?.parse({ items: [{ id: "a", count: 0 }] }));
  assert.throws(() => command_params_schemas.get("Custom.collect")?.parse({ items: [] }));
  assert.deepEqual(event_schemas.get("Custom.ready")?.parse({ url: "https://example.com", ready: true }), {
    url: "https://example.com",
    ready: true,
  });
  assert.throws(() => event_schemas.get("Custom.ready")?.parse({ url: "http://example.com", ready: true }));
  assert.deepEqual(event_schemas.get("Custom.count")?.parse({ value: 3 }), { value: 3 });
  assert.throws(() => event_schemas.get("Custom.count")?.parse({ value: 0 }));
});
