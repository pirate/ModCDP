import { once } from "node:events";
import WebSocket from "ws";
import { expect, test } from "vitest";

import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { AutoSessionRouter } from "../src/router/AutoSessionRouter.js";

test("AutoSessionRouter rejects pending execution context waiters when a session detaches", async () => {
  const router = new AutoSessionRouter(
    async () => ({}),
    () => 5_000,
  );
  const wait = router.waitForExecutionContext("detached-session", { timeout_ms: 5_000 });

  router.recordProtocolEvent(
    "Target.attachedToTarget",
    {
      sessionId: "detached-session",
      targetInfo: { targetId: "target-1", type: "page" },
    },
    null,
  );
  router.recordProtocolEvent("Target.detachedFromTarget", { sessionId: "detached-session" }, null);
  router.recordProtocolEvent("Runtime.executionContextCreated", { context: { id: 42 } }, "detached-session");

  await expect(wait).rejects.toThrow(
    "Runtime execution context wait cancelled because session detached-session detached.",
  );
  expect(router.sessionIdForTarget("target-1")).toBeNull();
  expect(router.execution_contexts.get("detached-session")).toBeUndefined();
}, 5_000);

test("AutoSessionRouter tracks real target sessions and execution contexts", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
  }).launch();
  const ws = new WebSocket(chrome.cdp_url!);
  await once(ws, "open");
  let next_id = 1;
  const pending = new Map<number, (message: Record<string, unknown>) => void>();
  const router = new AutoSessionRouter(
    (method, params = {}, session_id = null) =>
      send(method, params as Record<string, unknown>, session_id) as Promise<Record<string, unknown>>,
    () => 5_000,
  );

  function send(method: string, params: Record<string, unknown> = {}, session_id: string | null = null) {
    const id = next_id++;
    ws.send(JSON.stringify({ id, method, params, ...(session_id ? { sessionId: session_id } : {}) }));
    return new Promise<Record<string, unknown>>((resolve, reject) => {
      pending.set(id, (message) => {
        if (message.error) reject(new Error(JSON.stringify(message.error)));
        else resolve((message.result ?? {}) as Record<string, unknown>);
      });
    });
  }

  ws.on("message", (data) => {
    const message = JSON.parse(data.toString()) as Record<string, unknown>;
    if (typeof message.id === "number") {
      pending.get(message.id)?.(message);
      pending.delete(message.id);
      return;
    }
    if (typeof message.method !== "string") return;
    router.recordProtocolEvent(
      message.method,
      message.params,
      typeof message.sessionId === "string" ? message.sessionId : null,
    );
  });

  try {
    await send("Target.setAutoAttach", { autoAttach: true, waitForDebuggerOnStart: false, flatten: true });
    await send("Target.setDiscoverTargets", { discover: true });
    const created = await send("Target.createTarget", { url: "about:blank#modcdp-auto-session-router" });
    const target_id = created.targetId as string;
    await expect.poll(() => router.sessionIdForTarget(target_id), { timeout: 5_000 }).toEqual(expect.any(String));
    const session_id = router.sessionIdForTarget(target_id)!;

    const context_promise = router.waitForExecutionContext(session_id, { timeout_ms: 5_000 });
    await send("Runtime.enable", {}, session_id);
    await expect(context_promise).resolves.toEqual(expect.any(Number));
    expect(router.execution_contexts.get(session_id)).toEqual(expect.any(Number));

    await send("Target.detachFromTarget", { sessionId: session_id });
    await expect.poll(() => router.sessionIdForTarget(target_id), { timeout: 5_000 }).toBeNull();
    await send("Target.closeTarget", { targetId: target_id }).catch(() => ({}));
  } finally {
    ws.close();
    await once(ws, "close").catch(() => {});
    await chrome.close();
  }
}, 60_000);
