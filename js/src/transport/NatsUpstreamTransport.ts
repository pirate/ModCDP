import net from "node:net";
import tls from "node:tls";
import type { CdpCommandMessage } from "../types/modcdp.js";
import { UpstreamTransport, type UpstreamTransportConfig } from "./UpstreamTransport.js";

export const DEFAULT_UPSTREAM_NATS_URL = "ws://127.0.0.1:4223";
export const DEFAULT_UPSTREAM_NATS_SUBJECT_PREFIX = "modcdp.default";
export const DEFAULT_UPSTREAM_NATS_WAIT_TIMEOUT_MS = 10_000;

type NatsRole = "client" | "browser";
type NatsOptions = {
  upstream_nats_url?: string | null;
  upstream_nats_subject_prefix?: string | null;
  upstream_nats_role?: NatsRole;
  upstream_nats_wait_timeout_ms?: number;
};

type NatsSocket = WebSocket | net.Socket | tls.TLSSocket;

export class NatsUpstreamTransport extends UpstreamTransport {
  readonly mode = "nats" as const;
  readonly endpoint_kind = "modcdp_server" as const;
  declare url: string;
  upstream_nats_subject_prefix: string;
  private upstream_nats_role: NatsRole;
  private wait_timeout_ms: number;
  private socket: NatsSocket | null = null;
  private tcp_buffer = Buffer.alloc(0);
  private ws_buffer = "";
  private sid = "1";
  private connected = false;
  private peer_seen = false;
  private peer_waiters = new Set<{
    resolve: () => void;
    reject: (error: Error) => void;
    timeout: ReturnType<typeof setTimeout>;
  }>();

  constructor(options: NatsOptions = {}) {
    super();
    const { url, upstream_nats_subject_prefix } = normalizeNatsUrl(
      options.upstream_nats_url ?? DEFAULT_UPSTREAM_NATS_URL,
      options.upstream_nats_subject_prefix,
    );
    this.url = url;
    this.upstream_nats_subject_prefix = upstream_nats_subject_prefix;
    this.upstream_nats_role = options.upstream_nats_role ?? "client";
    this.wait_timeout_ms = options.upstream_nats_wait_timeout_ms ?? DEFAULT_UPSTREAM_NATS_WAIT_TIMEOUT_MS;
  }

  update(config: UpstreamTransportConfig = {}) {
    if (config.upstream_nats_url || config.upstream_nats_subject_prefix) {
      const normalized = normalizeNatsUrl(
        config.upstream_nats_url ?? this.url,
        config.upstream_nats_subject_prefix ?? this.upstream_nats_subject_prefix,
      );
      this.url = normalized.url;
      this.upstream_nats_subject_prefix = normalized.upstream_nats_subject_prefix;
    }
    if (config.upstream_nats_role === "client" || config.upstream_nats_role === "browser")
      this.upstream_nats_role = config.upstream_nats_role;
    if (typeof config.upstream_nats_wait_timeout_ms === "number")
      this.wait_timeout_ms = config.upstream_nats_wait_timeout_ms;
    return this;
  }

  getInjectorConfig() {
    return {
      upstream_nats_url: this.url,
      upstream_nats_subject_prefix: this.upstream_nats_subject_prefix,
    };
  }

  async connect() {
    if (this.connected) return;
    const parsed = new URL(this.url);
    if (parsed.protocol === "ws:" || parsed.protocol === "wss:") await this.connectWebSocket(parsed);
    else if (parsed.protocol === "nats:" || parsed.protocol === "tls:") await this.connectTcp(parsed);
    else
      throw new Error(`upstream.upstream_mode=nats requires ws://, wss://, nats://, or tls:// URL, got ${this.url}.`);
    this.connected = true;
    this.subscribe();
    this.publish(this.outgoingSubject(), {
      type: "modcdp.nats.hello",
      role: this.upstream_nats_role,
      version: 1,
    });
  }

  send(message: CdpCommandMessage) {
    if (!this.connected || !this.socket) throw new Error("NATS transport is not connected.");
    this.publish(this.outgoingSubject(), {
      type: "modcdp.nats.message",
      message,
    });
  }

  async waitForPeer() {
    if (this.peer_seen) return;
    await new Promise<void>((resolve, reject) => {
      let waiter!: {
        resolve: () => void;
        reject: (error: Error) => void;
        timeout: ReturnType<typeof setTimeout>;
      };
      const timeout = setTimeout(() => {
        this.peer_waiters.delete(waiter);
        reject(new Error(`Timed out waiting ${this.wait_timeout_ms}ms for NATS ModCDP peer.`));
      }, this.wait_timeout_ms);
      waiter = { resolve, reject, timeout };
      this.peer_waiters.add(waiter);
    });
  }

  async close() {
    try {
      if (this.socket instanceof WebSocket) this.socket.close();
      else this.socket?.destroy();
    } catch {}
    this.socket = null;
    this.connected = false;
    this.peer_seen = false;
    for (const waiter of this.peer_waiters) {
      clearTimeout(waiter.timeout);
      waiter.reject(
        new Error(`NATS transport for ${this.upstream_nats_subject_prefix} closed before a peer connected.`),
      );
    }
    this.peer_waiters.clear();
  }

  private async connectWebSocket(url: URL) {
    const ws = new WebSocket(url);
    this.socket = ws;
    ws.addEventListener("message", (event) => {
      void this.readWebSocket(event.data);
    });
    ws.addEventListener("close", () => this.emitClose(new Error("NATS websocket closed")));
    ws.addEventListener("error", () => this.emitClose(new Error("NATS websocket error")));
    await new Promise<void>((resolve, reject) => {
      const cleanup = () => {
        ws.removeEventListener("open", onOpen);
        ws.removeEventListener("error", onError);
      };
      const onOpen = () => {
        cleanup();
        this.writeProtocol(`CONNECT ${JSON.stringify(connectOptions())}\r\nPING\r\n`);
        resolve();
      };
      const onError = () => {
        cleanup();
        reject(new Error(`NATS websocket connection failed for ${url.toString()}`));
      };
      ws.addEventListener("open", onOpen);
      ws.addEventListener("error", onError);
    });
  }

  private async connectTcp(url: URL) {
    const port = Number(url.port || (url.protocol === "tls:" ? 4222 : 4222));
    const host = url.hostname || "127.0.0.1";
    const socket = url.protocol === "tls:" ? tls.connect({ host, port }) : net.connect({ host, port });
    this.socket = socket;
    socket.on("data", (chunk) => this.readTcp(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk)));
    socket.on("close", () => this.emitClose(new Error("NATS socket closed")));
    socket.on("error", () => this.emitClose(new Error("NATS socket error")));
    await new Promise<void>((resolve, reject) => {
      socket.once("connect", () => {
        this.writeProtocol(`CONNECT ${JSON.stringify(connectOptions())}\r\nPING\r\n`);
        resolve();
      });
      socket.once("error", reject);
    });
  }

  private subscribe() {
    this.writeProtocol(`SUB ${this.incomingSubject()} ${this.sid}\r\n`);
  }

  private publish(subject: string, message: unknown) {
    const body = JSON.stringify(message);
    this.writeProtocol(`PUB ${subject} ${Buffer.byteLength(body)}\r\n${body}\r\n`);
  }

  private writeProtocol(data: string) {
    const socket = this.socket;
    if (!socket) throw new Error("NATS transport is not connected.");
    if (socket instanceof WebSocket) socket.send(data);
    else socket.write(data);
  }

  private incomingSubject() {
    return `${this.upstream_nats_subject_prefix}.${this.upstream_nats_role === "client" ? "browser_to_client" : "client_to_browser"}`;
  }

  private outgoingSubject() {
    return `${this.upstream_nats_subject_prefix}.${this.upstream_nats_role === "client" ? "client_to_browser" : "browser_to_client"}`;
  }

  private async readWebSocket(data: unknown) {
    if (data instanceof ArrayBuffer) this.ws_buffer += Buffer.from(data).toString("utf8");
    else if (ArrayBuffer.isView(data))
      this.ws_buffer += Buffer.from(data.buffer, data.byteOffset, data.byteLength).toString("utf8");
    else if (typeof Blob !== "undefined" && data instanceof Blob) this.ws_buffer += await data.text();
    else this.ws_buffer += String(data);
    this.ws_buffer = this.consumeProtocol(this.ws_buffer);
  }

  private readTcp(chunk: Buffer) {
    this.tcp_buffer = Buffer.concat([this.tcp_buffer, chunk]);
    const remaining = this.consumeProtocol(this.tcp_buffer.toString("utf8"));
    this.tcp_buffer = Buffer.from(remaining, "utf8");
  }

  private consumeProtocol(buffer: string) {
    for (;;) {
      const lineEnd = buffer.indexOf("\r\n");
      if (lineEnd < 0) return buffer;
      const line = buffer.slice(0, lineEnd);
      const upper = line.toUpperCase();
      if (upper.startsWith("MSG ")) {
        const parts = line.split(/\s+/);
        const size = Number(parts[parts.length - 1]);
        const payloadStart = lineEnd + 2;
        const payloadEnd = payloadStart + size;
        if (!Number.isInteger(size) || buffer.length < payloadEnd + 2) return buffer;
        const payload = buffer.slice(payloadStart, payloadEnd);
        buffer = buffer.slice(payloadEnd + 2);
        this.handlePayload(payload);
        continue;
      }
      buffer = buffer.slice(lineEnd + 2);
      if (upper === "PING") this.writeProtocol("PONG\r\n");
      else if (upper.startsWith("-ERR")) this.emitClose(new Error(`NATS error: ${line}`));
    }
  }

  private handlePayload(payload: string) {
    let parsed: unknown;
    try {
      parsed = JSON.parse(payload);
    } catch {
      return;
    }
    const record = parsed && typeof parsed === "object" ? (parsed as Record<string, unknown>) : null;
    if (record?.type === "modcdp.nats.hello") {
      this.peer_seen = true;
      for (const waiter of this.peer_waiters) {
        clearTimeout(waiter.timeout);
        waiter.resolve();
      }
      this.peer_waiters.clear();
      return;
    }
    const message = record?.type === "modcdp.nats.message" ? record.message : parsed;
    this.parseAndEmitRecv(JSON.stringify(message));
  }
}

function connectOptions() {
  return {
    verbose: false,
    pedantic: false,
    lang: "modcdp",
    version: "1",
    protocol: 1,
  };
}

function normalizeNatsUrl(url: string, upstream_nats_subject_prefix?: string | null) {
  const parsed = new URL(url);
  const subject = upstream_nats_subject_prefix || parsed.searchParams.get("upstream_nats_subject_prefix");
  parsed.searchParams.delete("upstream_nats_subject_prefix");
  return {
    url: parsed.toString(),
    upstream_nats_subject_prefix: sanitizeSubjectPrefix(subject || DEFAULT_UPSTREAM_NATS_SUBJECT_PREFIX),
  };
}

function sanitizeSubjectPrefix(value: string) {
  const subject = value.trim();
  if (!subject || /[\s*>]/.test(subject)) throw new Error(`Invalid NATS subject prefix ${value}`);
  return subject;
}
