import type { CdpCommandMessage } from "../types/modcdp.js";
import { resolveCdpWebSocketUrl } from "../launcher/BrowserLauncher.js";
import { UpstreamTransport, type UpstreamTransportConfig } from "./UpstreamTransport.js";

export class WebSocketUpstreamTransport extends UpstreamTransport {
  readonly mode = "ws" as const;
  readonly endpoint_kind = "raw_cdp" as const;
  ws: WebSocket | null = null;

  constructor({ cdp_url = null }: { cdp_url?: string | null } = {}) {
    super();
    this.url = cdp_url ?? "";
  }

  update(config: UpstreamTransportConfig = {}) {
    if (config.cdp_url) this.url = config.cdp_url;
    return this;
  }

  getServerConfig() {
    return this.url ? { server_loopback_cdp_url: this.url } : {};
  }

  async connect() {
    if (!this.url) throw new Error("upstream.upstream_mode=ws requires upstream_cdp_url or launcher-provided cdp_url.");
    // cdp_url may start as an HTTP discovery endpoint; from here on it is the resolved WebSocket CDP endpoint.
    this.url = await resolveCdpWebSocketUrl(this.url, "upstream_cdp_url");
    const ws = new WebSocket(this.url);
    this.ws = ws;
    ws.addEventListener("message", (event) => this.parseAndEmitRecv(event.data));
    ws.addEventListener("close", () => this.emitClose(new Error("CDP websocket closed")));
    ws.addEventListener("error", () => this.emitClose(new Error("CDP websocket error")));
    await new Promise<void>((resolve, reject) => {
      const cleanup = () => {
        ws.removeEventListener("open", onOpen);
        ws.removeEventListener("error", onError);
      };
      const onOpen = () => {
        cleanup();
        resolve();
      };
      const onError = () => {
        cleanup();
        reject(new Error("CDP websocket error"));
      };
      ws.addEventListener("open", onOpen);
      ws.addEventListener("error", onError);
    });
  }

  send(message: CdpCommandMessage) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) throw new Error("CDP websocket is not connected.");
    this.ws.send(JSON.stringify(message));
  }

  async close() {
    try {
      this.ws?.close();
    } catch {}
    this.ws = null;
  }
}
