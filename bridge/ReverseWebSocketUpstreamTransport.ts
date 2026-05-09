import type { CdpCommandMessage } from "../types/modcdp.js";
import { parseHostPort, UpstreamTransport, type UpstreamTransportConfig } from "./UpstreamTransport.js";

export const DEFAULT_REVERSEWS_BIND = "127.0.0.1:29292";
export const DEFAULT_REVERSEWS_WAIT_TIMEOUT_MS = 10_000;

type ReverseHello = {
  type: "modcdp.reverse.hello";
  role?: string;
  version?: number;
  extension_id?: string | null;
};

export class ReverseWebSocketUpstreamTransport extends UpstreamTransport {
  readonly mode = "reversews" as const;
  readonly endpoint_kind = "modcdp_server" as const;
  declare url: string;
  private server: unknown = null;
  private socket: {
    readyState: number;
    OPEN: number;
    send: (data: string) => void;
    close: (...args: unknown[]) => void;
  } | null = null;
  private peer_waiters = new Set<{
    resolve: () => void;
    reject: (error: Error) => void;
    timeout: ReturnType<typeof setTimeout>;
  }>();
  peer_info: ReverseHello | null = null;

  constructor(
    bind = DEFAULT_REVERSEWS_BIND,
    private wait_timeout_ms = DEFAULT_REVERSEWS_WAIT_TIMEOUT_MS,
  ) {
    super();
    this.setBind(bind);
  }

  update(config: UpstreamTransportConfig = {}) {
    if (config.reversews_bind) this.setBind(config.reversews_bind);
    else if (config.url) this.setBind(config.url);
    if (typeof config.reversews_wait_timeout_ms === "number") this.wait_timeout_ms = config.reversews_wait_timeout_ms;
    return this;
  }

  getInjectorConfig() {
    return { reverse_proxy_url: this.url };
  }

  private setBind(bind: string) {
    const { host, port } = parseHostPort(bind, "127.0.0.1", 29292);
    this.url = `ws://${host}:${port}`;
  }

  async connect() {
    const { WebSocketServer } = await import("ws");
    const { host, port } = parseHostPort(this.url, "127.0.0.1", 29292);
    const server = new WebSocketServer({ host, port });
    this.server = server;
    server.on("connection", (socket) => this.accept(socket));
    await new Promise<void>((resolve, reject) => {
      server.once("listening", () => resolve());
      server.once("error", reject);
    });
  }

  send(message: CdpCommandMessage) {
    if (!this.socket || this.socket.readyState !== this.socket.OPEN) {
      throw new Error(`No reverse ModCDP extension peer is connected at ${this.url}.`);
    }
    this.socket.send(JSON.stringify(message));
  }

  async waitForPeer() {
    if (this.socket && this.socket.readyState === this.socket.OPEN) return;
    await new Promise<void>((resolve, reject) => {
      let waiter!: { resolve: () => void; reject: (error: Error) => void; timeout: ReturnType<typeof setTimeout> };
      const timeout = setTimeout(() => {
        this.peer_waiters.delete(waiter);
        reject(new Error(`Timed out waiting ${this.wait_timeout_ms}ms for reverse ModCDP extension connection.`));
      }, this.wait_timeout_ms);
      waiter = { resolve, reject, timeout };
      this.peer_waiters.add(waiter);
    });
  }

  async close() {
    try {
      this.socket?.close();
    } catch {}
    this.socket = null;
    const server = this.server as { close?: (callback: () => void) => void } | null;
    if (server?.close) await new Promise<void>((resolve) => server.close?.(() => resolve()));
    this.server = null;
    for (const waiter of this.peer_waiters) {
      clearTimeout(waiter.timeout);
      waiter.reject(new Error(`Reverse websocket transport at ${this.url} closed before a peer connected.`));
    }
    this.peer_waiters.clear();
  }

  private accept(socket: any) {
    const fail = (message: string) => {
      try {
        socket.close(1008, message.slice(0, 120));
      } catch {}
    };
    const timeout = setTimeout(() => fail("reverse hello timeout"), this.wait_timeout_ms);
    socket.once("message", (buf: unknown) => {
      clearTimeout(timeout);
      let hello: ReverseHello;
      try {
        const parsed = JSON.parse(String(buf));
        if (parsed?.type !== "modcdp.reverse.hello") throw new Error("missing hello type");
        hello = parsed;
      } catch (error) {
        fail(`invalid reverse hello: ${error instanceof Error ? error.message : String(error)}`);
        return;
      }
      if (this.socket && this.socket !== socket) {
        try {
          this.socket.close(1012, "reverse peer replaced");
        } catch {}
      }
      this.socket = socket;
      this.peer_info = hello;
      socket.on("message", (data: unknown) => this.parseAndEmitRecv(data));
      socket.on("close", () => {
        if (this.socket !== socket) return;
        this.socket = null;
        this.peer_info = null;
        this.emitClose(new Error("Reverse ModCDP websocket closed"));
      });
      socket.on("error", () => {
        if (this.socket !== socket) return;
        this.socket = null;
        this.peer_info = null;
        this.emitClose(new Error("Reverse ModCDP websocket error"));
      });
      for (const waiter of this.peer_waiters) {
        clearTimeout(waiter.timeout);
        waiter.resolve();
      }
      this.peer_waiters.clear();
    });
  }
}
