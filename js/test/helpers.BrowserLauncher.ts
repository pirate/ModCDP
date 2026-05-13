import { once } from "node:events";
import WebSocket from "ws";
import { expect } from "vitest";

import { resolveCdpWebSocketUrl } from "../src/launcher/BrowserLauncher.js";

type Pending = {
  resolve: (value: Record<string, unknown>) => void;
  reject: (error: Error) => void;
};

export class CdpSocket {
  private nextId = 1;
  private pending = new Map<number, Pending>();
  private events = new Map<string, Record<string, unknown>[]>();

  constructor(readonly ws: WebSocket) {
    ws.on("message", (data) => {
      const message = JSON.parse(data.toString()) as Record<string, unknown>;
      if (typeof message.id === "number") {
        const pending = this.pending.get(message.id);
        if (!pending) return;
        this.pending.delete(message.id);
        if (message.error) pending.reject(new Error(JSON.stringify(message.error)));
        else pending.resolve((message.result ?? {}) as Record<string, unknown>);
        return;
      }
      if (typeof message.method === "string") {
        const events = this.events.get(message.method) ?? [];
        events.push(message);
        this.events.set(message.method, events);
      }
    });
    ws.on("close", () => {
      for (const [id, pending] of this.pending) {
        pending.reject(new Error(`CDP websocket closed while command ${id} was pending`));
      }
      this.pending.clear();
    });
  }

  static async connect(endpoint: string) {
    const cdp_url = await resolveCdpWebSocketUrl(endpoint, "test cdp endpoint");
    const ws = new WebSocket(cdp_url);
    await once(ws, "open");
    return new CdpSocket(ws);
  }

  send(method: string, params: Record<string, unknown> = {}, sessionId?: string) {
    const id = this.nextId++;
    this.ws.send(JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) }));
    return new Promise<Record<string, unknown>>((resolve, reject) => {
      this.pending.set(id, { resolve, reject });
    });
  }

  async close() {
    if (this.ws.readyState === WebSocket.CLOSED) return;
    const closed = once(this.ws, "close").catch(() => {});
    this.ws.close();
    await closed;
  }
}

export class PipeCdpSocket {
  private nextId = 100;
  private pending = new Map<number, Pending>();
  private buffer = "";

  constructor(
    readonly pipe_read: NodeJS.ReadableStream,
    readonly pipe_write: NodeJS.WritableStream,
  ) {
    pipe_read.on("data", (chunk: Buffer | string) => {
      this.buffer += chunk.toString();
      while (this.buffer.includes("\0")) {
        const [raw, ...rest] = this.buffer.split("\0");
        this.buffer = rest.join("\0");
        if (!raw) continue;
        const message = JSON.parse(raw) as Record<string, unknown>;
        if (typeof message.id !== "number") continue;
        const pending = this.pending.get(message.id);
        if (!pending) continue;
        this.pending.delete(message.id);
        if (message.error) pending.reject(new Error(JSON.stringify(message.error)));
        else pending.resolve((message.result ?? {}) as Record<string, unknown>);
      }
    });
  }

  send(method: string, params: Record<string, unknown> = {}, sessionId?: string) {
    const id = this.nextId++;
    this.pipe_write.write(`${JSON.stringify({ id, method, params, ...(sessionId ? { sessionId } : {}) })}\0`);
    return new Promise<Record<string, unknown>>((resolve, reject) => {
      this.pending.set(id, { resolve, reject });
    });
  }
}

export async function expectCdpBrowserSurface(cdp: Pick<CdpSocket, "send">) {
  const version = await cdp.send("Browser.getVersion");
  expect(version.product).toEqual(expect.stringMatching(/Chrome|Chromium/));
  expect(version.protocolVersion).toEqual(expect.any(String));

  const created = await cdp.send("Target.createTarget", { url: "about:blank#modcdp-launcher-test" });
  expect(created.targetId).toEqual(expect.any(String));
  const targetId = created.targetId as string;

  try {
    const attached = await cdp.send("Target.attachToTarget", { targetId, flatten: true });
    expect(attached.sessionId).toEqual(expect.any(String));
    const sessionId = attached.sessionId as string;
    await cdp.send("Runtime.enable", {}, sessionId);
    const evaluated = await cdp.send(
      "Runtime.evaluate",
      { expression: "(() => ({ ok: true, value: 42 }))()", returnByValue: true },
      sessionId,
    );
    expect(evaluated.result).toMatchObject({ type: "object", value: { ok: true, value: 42 } });
  } finally {
    await cdp.send("Target.closeTarget", { targetId }).catch(() => ({}));
  }
}

export async function expectHttpEndpointDown(url: string) {
  await expect
    .poll(
      async () => {
        try {
          await fetch(`${url}/json/version`);
          return false;
        } catch {
          return true;
        }
      },
      { timeout: 5_000, interval: 100 },
    )
    .toBe(true);
}
