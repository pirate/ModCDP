import type { CdpCommandMessage, CdpEventMessage, CdpResponseMessage } from "../types/modcdp.js";
import type { ModCDPServerOptions } from "../types/modcdp.js";
import type { BrowserLaunchOptions } from "./BrowserLauncher.js";
import { CdpEventMessageSchema, CdpResponseMessageSchema } from "../types/modcdp.js";

export type UpstreamMode = "ws" | "pipe" | "nativemessaging" | "reversews" | "nats";
export type UpstreamEndpointKind = "raw_cdp" | "modcdp_server";
export type UpstreamTransportConfig = {
  cdp_url?: string | null;
  ws_url?: string | null;
  url?: string | null;
  user_data_dir?: string | null;
  pipe_read?: NodeJS.ReadableStream | null;
  pipe_write?: NodeJS.WritableStream | null;
  manifest_path?: string | null;
  manifest_paths?: string[] | null;
  extension_id?: string | null;
  nats_url?: string | null;
  nats_subject_prefix?: string | null;
  reversews_bind?: string | null;
};

export class UpstreamTransport {
  readonly mode: UpstreamMode = "ws";
  readonly endpoint_kind: UpstreamEndpointKind = "raw_cdp";
  url?: string;
  private recv_listeners = new Set<(message: CdpResponseMessage | CdpEventMessage) => void>();
  private close_listeners = new Set<(error: Error) => void>();

  async connect() {
    throw new Error(`${this.constructor.name}.connect is not implemented.`);
  }

  update(_config: UpstreamTransportConfig = {}) {
    return this;
  }

  getLauncherConfig(): BrowserLaunchOptions {
    return {};
  }

  getServerConfig(): Partial<ModCDPServerOptions> {
    return {};
  }

  async close() {}

  send(_message: CdpCommandMessage) {
    throw new Error(`${this.constructor.name}.send is not implemented.`);
  }

  onRecv(listener: (message: CdpResponseMessage | CdpEventMessage) => void) {
    this.recv_listeners.add(listener);
    return () => this.recv_listeners.delete(listener);
  }

  onClose(listener: (error: Error) => void) {
    this.close_listeners.add(listener);
    return () => this.close_listeners.delete(listener);
  }

  protected emitRecv(message: CdpResponseMessage | CdpEventMessage) {
    for (const listener of this.recv_listeners) listener(message);
  }

  protected emitClose(error: Error) {
    for (const listener of this.close_listeners) listener(error);
  }

  protected parseAndEmitRecv(data: unknown) {
    try {
      const parsed = JSON.parse(typeof data === "string" ? data : String(data));
      this.emitRecv("id" in parsed ? CdpResponseMessageSchema.parse(parsed) : CdpEventMessageSchema.parse(parsed));
    } catch {}
  }

  async waitForPeer() {}
}

export function parseHostPort(value: string, defaultHost: string, defaultPort: number) {
  const parsed = new URL(/^[a-z][a-z\d+\-.]*:\/\//i.test(value) ? value : `ws://${value}`);
  const host = parsed.hostname || defaultHost;
  const port = Number(parsed.port || defaultPort);
  if (!Number.isInteger(port) || port <= 0 || port > 65_535) throw new Error(`Invalid host:port ${value}`);
  return { host, port };
}

export function endpointKindForUpstream(mode: UpstreamMode): UpstreamEndpointKind {
  return mode === "ws" || mode === "pipe" ? "raw_cdp" : "modcdp_server";
}
