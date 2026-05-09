import assert from "node:assert/strict";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";
import { z } from "zod";

import { ModCDPClient, type ModCDPClientInstance } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

test("custom commands install flat namespace methods through a real service worker", async () => {
  type CustomClient = ModCDPClientInstance<{
    "Custom.doSomething": { params: { id: string }; result: boolean };
  }>;
  const cdp = new ModCDPClient({
    launch: {
      mode: "local",
      options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { mode: "ws" },
    extension: {
      mode: "auto",
      path: EXTENSION_PATH,
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
    },
    client: { routes: { "Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp" } },
    server: { routes: { "*.*": "loopback_cdp" } },
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
    launch: {
      mode: "local",
      options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { mode: "ws" },
    extension: {
      mode: "auto",
      path: EXTENSION_PATH,
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
    },
    client: { routes: { "Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp" } },
    server: { routes: { "*.*": "loopback_cdp" } },
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
