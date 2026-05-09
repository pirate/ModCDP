// proxy.js: a transparent local CDP proxy that "upgrades" any vanilla CDP
// client to speak Mod.* / Custom.*. By default listens on ws://127.0.0.1:9223
// and forwards to http://127.0.0.1:9222.
//
// Behavior on each client connection:
//   - Connect a ModCDPClient to the existing upstream so auto-attach,
//     extension discovery, and injection stay in the main client implementation.
//   - For websocket upstreams, reuse that client's upstream websocket + hidden
//     extension session to rewrite Mod.* / Custom.* outbound and
//     Runtime.bindingCalled inbound; forward everything else unchanged.
//   - For pipe, native messaging, and NATS upstreams, let ModCDPClient own the
//     selected transport and proxy downstream CDP-shaped messages through it.
//   - Keep mirrored upstream events private by default so vanilla CDP clients
//     only see native upstream CDP messages. Set forward_mirrored_upstream_events to
//     true when debugging the service-worker mirror path itself.
//
// Run as a CLI:
//   node proxy.js --port 9223 --upstream=ws --upstream-ws-url=http://127.0.0.1:9222
//   node proxy.js --port 9223 --launch=local --upstream=pipe
//   node proxy.js --port 9223 --launch=local --upstream=nats --upstream-nats-url=ws://127.0.0.1:4223
//
// Or import { startProxy } and embed.

import http from "node:http";
import path from "node:path";
import { fileURLToPath } from "node:url";
import type { RawData } from "ws";
import type { WebSocket } from "ws";

import { ModCDPClient } from "../client/js/ModCDPClient.js";
import {
  UPSTREAM_EVENT_BINDING_NAME,
  wrapModCDPEvaluate,
  wrapModCDPAddCustomCommand,
  wrapModCDPAddMiddleware,
  wrapModCDPAddCustomEvent,
  wrapCustomCommand,
  unwrapResponseIfNeeded,
  unwrapEventIfNeeded,
} from "./translate.js";
import type {
  CdpCommandMessage,
  CdpEventMessage,
  CdpResponseMessage,
  CdpMessage,
  ProxyConnectionState,
} from "../types/modcdp.js";
import {
  CdpCommandMessageSchema,
  CdpEventMessageSchema,
  CdpResponseMessageSchema,
  ModCDPAddCustomCommandParamsSchema,
  ModCDPAddCustomEventObjectParamsSchema,
  ModCDPAddMiddlewareParamsSchema,
  ModCDPEvaluateParamsSchema,
  normalizeModCDPName,
} from "../types/modcdp.js";
import { events as RuntimeEvents } from "../types/zod/Runtime.js";
import { events as TargetEvents } from "../types/zod/Target.js";

const ROOT = path.dirname(fileURLToPath(import.meta.url));
const DEFAULT_EXTENSION_PATH = path.resolve(ROOT, "..", "extension");
const DEFAULT_PORT = 9223;
const DEFAULT_HOST = "127.0.0.1";
const DEFAULT_UPSTREAM = "http://127.0.0.1:9222";
export const DEFAULT_UPSTREAM_MONITOR_INTERVAL_MS = 1_000;
export const DEFAULT_REVERSE_WAIT_TIMEOUT_MS = 2_000;

const DEBUG = process.env.PROXY_DEBUG === "1";
const log = (...args) => console.log("[proxy]", ...args);
const dbg = (...args) => {
  if (DEBUG) console.log("[proxy:dbg]", ...args);
};

const MAGIC_METHODS = new Set(["Mod.evaluate", "Mod.addCustomCommand", "Mod.addCustomEvent", "Mod.addMiddleware"]);
const ROUTE_TO_SW_RE = /^(Mod|Custom)\./;
const isWsUrl = (url) => typeof url === "string" && /^wss?:\/\//i.test(url);

// --- public API -------------------------------------------------------------

export async function startProxy({
  host = DEFAULT_HOST,
  port = DEFAULT_PORT,
  launch = { mode: "remote" },
  upstream = { mode: "ws", ws_url: DEFAULT_UPSTREAM },
  extension = { mode: "auto", path: DEFAULT_EXTENSION_PATH },
  client: clientOptions = {},
  server: serverOptions = {},
  forward_mirrored_upstream_events = false,
  upstream_monitor_interval_ms = DEFAULT_UPSTREAM_MONITOR_INTERVAL_MS,
}: {
  host?: string;
  port?: number;
  launch?: { mode?: string; executable_path?: string | null; user_data_dir?: string | null; options?: Record<string, unknown> };
  upstream?: {
    mode?: string;
    ws_url?: string | null;
    nats_url?: string | null;
    nats_subject_prefix?: string | null;
    reversews_bind?: string | null;
    reversews_wait_timeout_ms?: number | null;
    nativemessaging_manifest?: string | null;
  };
  extension?: { mode?: string; path?: string | null };
  client?: Record<string, unknown>;
  server?: Record<string, unknown> | null;
  forward_mirrored_upstream_events?: boolean;
  upstream_monitor_interval_ms?: number;
} = {}) {
  const { WebSocket, WebSocketServer } = await loadWsForProxy();
  const upstreamMode = upstream.mode ?? "ws";
  const upstreamWsUrl = upstream.ws_url ?? (launch.mode === "local" ? null : DEFAULT_UPSTREAM);
  const clientManagedUpstream = upstreamMode === "nativemessaging" || upstreamMode === "nats" || upstreamMode === "pipe";
  const managed_reverse_upstream =
    upstreamMode === "reversews" &&
    (launch.mode === "local" || launch.mode === "bb" || (launch.mode === "remote" && upstream.ws_url != null));
  const reverse_wait_timeout_ms = upstream.reversews_wait_timeout_ms ?? DEFAULT_REVERSE_WAIT_TIMEOUT_MS;
  const reverseOptions =
    upstreamMode === "reversews" && !managed_reverse_upstream
      ? parseHostPort(upstream.reversews_bind ?? "127.0.0.1:29292", DEFAULT_HOST, 29292)
      : null;
  const reversePeer = reverseOptions ? createReversePeerState(reverse_wait_timeout_ms) : null;
  const servesLocalDiscovery = Boolean(reversePeer) || managed_reverse_upstream || clientManagedUpstream;
  let managed_reverse_cdp: ModCDPClient | null = null;
  const httpServer = http.createServer(async (req, res) => {
    try {
      const requestUrl = req.url === "/json/version/" ? "/json/version" : req.url;
      if (requestUrl === "/json/version") {
        res.writeHead(200, { "content-type": "application/json" });
        res.end(JSON.stringify({ webSocketDebuggerUrl: `ws://${req.headers.host}/devtools/browser/proxy` }));
        return;
      }
      if (servesLocalDiscovery) {
        if (requestUrl === "/json" || requestUrl === "/json/list" || requestUrl === "/json/list/") {
          res.writeHead(200, { "content-type": "application/json" });
          res.end(
            JSON.stringify([
              {
                id: "proxy",
                type: "browser",
                title: upstreamMode === "nativemessaging" ? "ModCDP Native Messaging Proxy" : "ModCDP Reverse Proxy",
                webSocketDebuggerUrl: `ws://${req.headers.host}/devtools/browser/proxy`,
              },
            ]),
          );
          return;
        }
        res.writeHead(404);
        res.end("Not found.");
        return;
      }
      if (!upstreamWsUrl || isWsUrl(upstreamWsUrl)) {
        res.writeHead(404);
        res.end("HTTP discovery is unavailable for this upstream.");
        return;
      }
      const upstreamRes = await fetch(`${upstreamWsUrl}${requestUrl}`);
      const text = await upstreamRes.text();
      const contentType = upstreamRes.headers.get("content-type") || "";
      if (contentType.includes("application/json")) {
        const body = JSON.parse(text);
        rewriteWsUrls(body, req.headers.host);
        res.writeHead(upstreamRes.status, { "content-type": "application/json" });
        res.end(JSON.stringify(body));
      } else {
        res.writeHead(upstreamRes.status, Object.fromEntries(upstreamRes.headers));
        res.end(text);
      }
    } catch (error) {
      res.writeHead(502);
      res.end(error.message);
    }
  });

  let stopUpstreamMonitor: (() => void) | null = null;
  let reverseWss: InstanceType<typeof WebSocketServer> | null = null;
  const wss = new WebSocketServer({ server: httpServer });
  let closing = false;
  let closePromise: Promise<void> | null = null;
  const close = async () => {
    if (closePromise) return closePromise;
    closing = true;
    closePromise = (async () => {
      stopUpstreamMonitor?.();
      stopUpstreamMonitor = null;
      for (const socket of wss.clients) {
        try {
          socket.close();
        } catch {}
      }
      if (reversePeer?.socket) {
        try {
          reversePeer.socket.close();
        } catch {}
      }
      await managed_reverse_cdp?.close().catch(() => {});
      managed_reverse_cdp = null;
      await new Promise<void>((resolve) => wss.close(() => resolve()));
      if (reverseWss) await new Promise<void>((resolve) => reverseWss?.close(() => resolve()));
      if (httpServer.listening) await new Promise<void>((resolve) => httpServer.close(() => resolve()));
    })();
    return closePromise;
  };
  wss.on("connection", (client, req) => {
    log("client connected", req.url);
    // attach a synchronous early-buffer immediately so we don't lose messages
    // sent before bootstrap (e.g. Playwright's first commands).
    const earlyBuffer = [];
    const earlyHandler = (buf) => earlyBuffer.push(buf);
    client.on("message", earlyHandler);
    if (reversePeer) {
      handleReverseConnection(client, earlyBuffer, earlyHandler, reversePeer).catch((err) => {
        log("reverse connection failed:", err.message);
        try {
          client.close(1011, err.message.slice(0, 120));
        } catch {}
      });
      return;
    }
    if (managed_reverse_cdp) {
      wireClientManagedConnection(client, earlyBuffer, earlyHandler, managed_reverse_cdp, false);
      return;
    }
    if (clientManagedUpstream) {
      handleClientManagedConnection(client, earlyBuffer, earlyHandler, {
        launch,
        upstream: { ...upstream, mode: upstreamMode, ws_url: upstreamWsUrl },
        extension,
        client: clientOptions,
        server: serverOptions,
      }).catch((err) => {
        log("client-managed connection failed:", err.message);
        try {
          client.close(1011, err.message.slice(0, 120));
        } catch {}
      });
      return;
    }
    handleConnection(client, earlyBuffer, earlyHandler, {
      launch,
      upstream: { ...upstream, mode: upstreamMode, ws_url: upstreamWsUrl },
      extension,
      client: clientOptions,
      server: serverOptions,
      forward_mirrored_upstream_events,
      onUpstreamClosed: () => {
        void close().catch((error) => log("proxy close failed:", errorMessage(error)));
      },
    }).catch((err) => {
      log("connection failed:", err.message);
      try {
        client.close(1011, err.message.slice(0, 120));
      } catch {}
    });
  });

  if (reversePeer && reverseOptions) {
    reverseWss = new WebSocketServer({ host: reverseOptions.host, port: reverseOptions.port });
    reverseWss.on("connection", (socket, req) => {
      log("reverse candidate connected", req.socket.remoteAddress);
      acceptReversePeer(reversePeer, socket);
    });
    reverseWss.on("error", (error) => log("reverse listener error:", errorMessage(error)));
  }

  if (managed_reverse_upstream) {
    managed_reverse_cdp = new ModCDPClient({
      launch: {
        mode: launch.mode as "local" | "remote" | "bb" | "none",
        executable_path: launch.executable_path,
        user_data_dir: launch.user_data_dir,
        options: launch.options as any,
      },
      upstream: {
        mode: "reversews",
        reversews_bind: upstream.reversews_bind,
        reversews_wait_timeout_ms: reverse_wait_timeout_ms,
      },
      extension: {
        mode: (extension.mode ?? "auto") as "auto" | "discover" | "inject" | "borrow" | "none",
        path: extension.path ?? DEFAULT_EXTENSION_PATH,
        service_worker_url_suffixes: ["/modcdp/service_worker.js"],
        trust_service_worker_target: true,
      },
      client: { hydrate_aliases: false, ...clientOptions },
      server: serverOptions as any,
    });
    try {
      await managed_reverse_cdp.connect();
    } catch (error) {
      await managed_reverse_cdp.close().catch(() => {});
      managed_reverse_cdp = null;
      throw error;
    }
  }

  await new Promise<void>((resolve) => httpServer.listen(port, host, () => resolve()));
  if (!reversePeer && !managed_reverse_upstream && !clientManagedUpstream && upstreamWsUrl && !isWsUrl(upstreamWsUrl)) {
    stopUpstreamMonitor = monitorUpstream(
      upstreamWsUrl,
      upstream_monitor_interval_ms,
      () => {
        void close().catch((error) => log("proxy close failed:", errorMessage(error)));
      },
      WebSocket,
    );
  }
  log(
    reverseOptions
      ? `listening on ws://${host}:${port}/  (reverse: ws://${reverseOptions.host}:${reverseOptions.port})`
      : managed_reverse_upstream
        ? `listening on ws://${host}:${port}/  (upstream: reversews:${upstream.reversews_bind ?? "127.0.0.1:29292"})`
        : clientManagedUpstream
        ? `listening on ws://${host}:${port}/  (upstream: ${upstreamMode})`
        : `listening on ws://${host}:${port}/  (upstream: ${upstreamMode}:${upstreamWsUrl ?? "local-launch"})`,
  );

  return {
    url: `http://${host}:${port}`,
    ws_url: `ws://${host}:${port}`,
    close,
  };
}

async function loadWsForProxy() {
  try {
    return await import("ws");
  } catch (error) {
    throw new Error(
      `The ModCDP proxy requires the optional "ws" package, but it is not installed.\n\n` +
        `Install optional dependencies for modcdp, or add ws explicitly:\n` +
        `  pnpm add ws\n` +
        `  npm install ws\n` +
        `  yarn add ws\n\n` +
        `The ModCDP client does not require ws; it uses the native WebSocket implementation.`,
      { cause: error },
    );
  }
}

function rewriteWsUrls(value: unknown, host: string) {
  if (!value || typeof value !== "object") return;
  if ("webSocketDebuggerUrl" in value && typeof value.webSocketDebuggerUrl === "string") {
    value.webSocketDebuggerUrl = value.webSocketDebuggerUrl.replace(/ws:\/\/[^/]+/, `ws://${host}`);
  }
  for (const child of Array.isArray(value) ? value : Object.values(value)) rewriteWsUrls(child, host);
}

function errorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (error && typeof error === "object" && "message" in error && typeof error.message === "string")
    return error.message;
  return "";
}

function monitorUpstream(
  upstream: string,
  upstream_monitor_interval_ms: number,
  onClosed: () => void,
  WebSocketCtor: new (url: string) => WebSocket,
) {
  let stopped = false;
  let socket: WebSocket | null = null;
  let interval: NodeJS.Timeout | null = null;
  const close = () => {
    if (stopped) return;
    stopped = true;
    if (interval) clearInterval(interval);
    try {
      socket?.close();
    } catch {}
  };
  const upstreamClosed = () => {
    if (stopped) return;
    close();
    onClosed();
  };

  if (isWsUrl(upstream)) return close;

  const check = async () => {
    try {
      const response = await fetch(`${upstream}/json/version`);
      if (!response.ok) upstreamClosed();
    } catch {
      upstreamClosed();
    }
  };
  interval = setInterval(() => void check(), upstream_monitor_interval_ms);
  void check();
  return close;
}

type HostPort = { host: string; port: number };
type ReverseHello = {
  type: "modcdp.reverse.hello";
  role?: string;
  version?: number;
  extension_id?: string | null;
};
type ReversePeerState = {
  socket: WebSocket | null;
  info: ReverseHello | null;
  wait_timeout_ms: number;
  waiters: Set<{ resolve: (socket: WebSocket) => void; reject: (error: Error) => void; timeout: NodeJS.Timeout }>;
};
type ReverseConnectionState = {
  client: WebSocket;
  reverse: WebSocket;
  next_reverse_id: number;
  pending: Map<number, { client_id: number; client_session_id: string | null }>;
  client_session_ids: Set<string>;
  bootstrapped: boolean;
  closing: boolean;
  queued_from_client: RawData[];
};

function parseHostPort(value: string, defaultHost: string, defaultPort: number): HostPort {
  const raw = String(value || "");
  const parsed = new URL(/^[a-z][a-z\d+\-.]*:\/\//i.test(raw) ? raw : `ws://${raw}`);
  const host = parsed.hostname || defaultHost;
  const port = Number(parsed.port || defaultPort);
  if (!Number.isInteger(port) || port <= 0 || port > 65_535) throw new Error(`Invalid host:port ${value}`);
  return { host, port };
}

function createReversePeerState(wait_timeout_ms: number): ReversePeerState {
  return {
    socket: null,
    info: null,
    wait_timeout_ms,
    waiters: new Set(),
  };
}

function isOpenSocket(socket: WebSocket | null) {
  return socket != null && socket.readyState === socket.OPEN;
}

function waitForReversePeer(state: ReversePeerState) {
  if (isOpenSocket(state.socket)) return Promise.resolve(state.socket);
  return new Promise<WebSocket>((resolve, reject) => {
    const waiter = {
      resolve,
      reject,
      timeout: setTimeout(() => {
        state.waiters.delete(waiter);
        reject(new Error(`Timed out waiting ${state.wait_timeout_ms}ms for reverse ModCDP extension connection.`));
      }, state.wait_timeout_ms),
    };
    state.waiters.add(waiter);
  });
}

function resolveReverseWaiters(state: ReversePeerState, socket: WebSocket) {
  for (const waiter of state.waiters) {
    clearTimeout(waiter.timeout);
    waiter.resolve(socket);
  }
  state.waiters.clear();
}

function rejectReverseWaiters(state: ReversePeerState, error: Error) {
  for (const waiter of state.waiters) {
    clearTimeout(waiter.timeout);
    waiter.reject(error);
  }
  state.waiters.clear();
}

function acceptReversePeer(state: ReversePeerState, socket: WebSocket) {
  const fail = (message: string) => {
    try {
      socket.close(1008, message.slice(0, 120));
    } catch {}
  };
  const timeout = setTimeout(() => fail("reverse hello timeout"), state.wait_timeout_ms);
  socket.once("message", (buf) => {
    clearTimeout(timeout);
    let hello: ReverseHello;
    try {
      const parsed = JSON.parse(String(buf));
      if (parsed?.type !== "modcdp.reverse.hello") throw new Error("missing hello type");
      hello = parsed;
    } catch (error) {
      fail(`invalid reverse hello: ${errorMessage(error)}`);
      return;
    }

    if (isOpenSocket(state.socket) && state.socket !== socket) {
      try {
        state.socket?.close(1012, "reverse peer replaced");
      } catch {}
    }
    state.socket = socket;
    state.info = hello;
    log("reverse extension connected", hello.extension_id || "(unknown extension)");
    socket.addEventListener("close", () => {
      if (state.socket !== socket) return;
      state.socket = null;
      state.info = null;
      rejectReverseWaiters(state, new Error("Reverse ModCDP extension connection closed."));
    });
    socket.addEventListener("error", () => {
      if (state.socket !== socket) return;
      state.socket = null;
      state.info = null;
      rejectReverseWaiters(state, new Error("Reverse ModCDP extension connection errored."));
    });
    resolveReverseWaiters(state, socket);
  });
}

// --- per-connection bridging ----------------------------------------------

async function handleReverseConnection(
  client: WebSocket,
  earlyBuffer: RawData[],
  earlyHandler: (buf: RawData) => void,
  reversePeer: ReversePeerState,
) {
  const reverse = await waitForReversePeer(reversePeer);
  if (!isOpenSocket(reverse)) throw new Error("Reverse ModCDP extension connection is not open.");

  const state: ReverseConnectionState = {
    client,
    reverse,
    next_reverse_id: 1_000_000,
    pending: new Map(),
    client_session_ids: new Set(),
    bootstrapped: false,
    closing: false,
    queued_from_client: [],
  };

  const onReverseMessage = (event) => {
    let msg: CdpResponseMessage | CdpEventMessage;
    try {
      const parsed = JSON.parse(String(event.data));
      msg = "id" in parsed ? CdpResponseMessageSchema.parse(parsed) : CdpEventMessageSchema.parse(parsed);
    } catch (e) {
      log("reverse parse error", e.message);
      return;
    }
    dbg("reverse->", msg.id ?? "", msg.method ?? "(response)", msg.sessionId ?? "");
    handleReverseUpstreamMessage(state, msg);
  };
  reverse.addEventListener("message", onReverseMessage);
  reverse.addEventListener("close", () => {
    reverse.removeEventListener("message", onReverseMessage);
    state.closing = true;
    try {
      client.close();
    } catch {}
  });
  reverse.addEventListener("error", () => {
    reverse.removeEventListener("message", onReverseMessage);
    state.closing = true;
    try {
      client.close(1011, "reverse upstream error");
    } catch {}
  });
  client.on("close", () => {
    state.closing = true;
    reverse.removeEventListener("message", onReverseMessage);
  });

  client.off("message", earlyHandler);
  for (const buf of earlyBuffer) state.queued_from_client.push(buf);
  client.on("message", (buf) => {
    if (!state.bootstrapped) {
      state.queued_from_client.push(buf);
      return;
    }
    handleReverseClientMessage(state, buf);
  });
  state.bootstrapped = true;
  for (const buf of state.queued_from_client) handleReverseClientMessage(state, buf);
  state.queued_from_client = [];
}

function handleReverseClientMessage(state: ReverseConnectionState, buf: RawData) {
  let msg: CdpCommandMessage;
  try {
    msg = CdpCommandMessageSchema.parse(JSON.parse(String(buf)));
  } catch (e) {
    log("client parse error", e.message);
    return;
  }
  dbg("client->reverse", msg.id ?? "", msg.method, msg.sessionId ?? "");
  const upId = state.next_reverse_id++;
  state.pending.set(upId, { client_id: msg.id, client_session_id: msg.sessionId || null });
  const out: CdpCommandMessage = { id: upId, method: msg.method, params: msg.params ?? {} };
  if (msg.sessionId) out.sessionId = msg.sessionId;
  state.reverse.send(JSON.stringify(out));
}

function handleReverseUpstreamMessage(state: ReverseConnectionState, msg: CdpResponseMessage | CdpEventMessage) {
  if ("id" in msg && typeof msg.id === "number") {
    const response = CdpResponseMessageSchema.parse(msg);
    const pending = state.pending.get(response.id);
    if (!pending) return;
    state.pending.delete(response.id);
    const out: CdpResponseMessage = response.error
      ? { id: pending.client_id, error: response.error }
      : { id: pending.client_id, result: response.result ?? {} };
    if (pending.client_session_id) out.sessionId = pending.client_session_id;
    sendReverseToClient(state, out);
    return;
  }

  const event = CdpEventMessageSchema.parse(msg);
  const eventSessionId =
    event.params &&
    typeof event.params === "object" &&
    "sessionId" in event.params &&
    typeof event.params.sessionId === "string"
      ? event.params.sessionId
      : event.sessionId || null;
  if (event.method === "Target.attachedToTarget" && eventSessionId) {
    state.client_session_ids.add(eventSessionId);
  } else if (event.method === "Target.detachedFromTarget" && eventSessionId) {
    state.client_session_ids.delete(eventSessionId);
  }

  sendReverseToClient(state, event);
  if (!event.sessionId) {
    for (const sessionId of state.client_session_ids) sendReverseToClient(state, { ...event, sessionId });
  }
}

function sendReverseToClient(state: ReverseConnectionState, obj: CdpMessage) {
  if (DEBUG)
    dbg("client<-reverse", "id" in obj ? obj.id : "", "method" in obj ? obj.method : "(response)", obj.sessionId ?? "");
  state.client.send(JSON.stringify(obj));
}

async function handleConnection(
  client: WebSocket,
  earlyBuffer: RawData[],
  earlyHandler: (buf: RawData) => void,
  {
    launch,
    upstream,
    extension,
    client: clientOptions,
    server,
    forward_mirrored_upstream_events,
    onUpstreamClosed,
  }: {
    launch: { mode?: string; executable_path?: string | null; user_data_dir?: string | null; options?: Record<string, unknown> };
    upstream: { mode?: string; ws_url?: string | null };
    extension: { mode?: string; path?: string | null };
    client?: Record<string, unknown>;
    server?: Record<string, unknown> | null;
    forward_mirrored_upstream_events: boolean;
    onUpstreamClosed: () => void;
  },
) {
  const cdp = new ModCDPClient({
    launch: {
      mode: launch.mode as "local" | "remote" | "bb" | "none",
      executable_path: launch.executable_path,
      user_data_dir: launch.user_data_dir,
      options: launch.options as any,
    },
    upstream: { mode: "ws", ws_url: upstream.ws_url },
    extension: {
      mode: (extension.mode ?? "auto") as "auto" | "discover" | "inject" | "borrow" | "none",
      path: extension.path ?? DEFAULT_EXTENSION_PATH,
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
    },
    client: { hydrate_aliases: false, ...clientOptions },
    server: server as any,
  });
  await cdp.connect();
  const upstream_socket = (cdp.transport as unknown as { ws?: WebSocket } | null)?.ws ?? null;
  if (!upstream_socket) throw new Error("ModCDPClient connected without an upstream websocket.");

  // per-connection state
  const state: ProxyConnectionState = {
    client,
    upstream: upstream_socket,
    next_upstream_id: 1_000_000,
    pending: new Map(), // upstream_id -> { kind, client_id?, client_session_id?, ... }
    ext_session_id: cdp.ext_session_id,
    ext_target_id: cdp.ext_target_id,
    ext_execution_context_id: cdp.ext_execution_context_id,
    hidden_session_ids: new Set(), // sessions we attached for ourselves
    hidden_target_ids: new Set(), // SW target the client must never see
    target_session_ids: cdp.auto_target_sessions,
    client_session_ids: new Set(), // session ids the client has attached
    forward_mirrored_upstream_events: forward_mirrored_upstream_events,
    bootstrapped: false,
    closing: false,
    queued_from_client: [],
  };
  if (cdp.ext_session_id) state.hidden_session_ids.add(cdp.ext_session_id);
  if (cdp.ext_target_id) state.hidden_target_ids.add(cdp.ext_target_id);

  upstream_socket.addEventListener("message", (event) => {
    let msg: CdpResponseMessage | CdpEventMessage;
    try {
      const parsed = JSON.parse(String(event.data));
      msg = "id" in parsed ? CdpResponseMessageSchema.parse(parsed) : CdpEventMessageSchema.parse(parsed);
    } catch (e) {
      log("upstream parse error", e.message);
      return;
    }
    dbg("upstream->", msg.id ?? "", msg.method ?? "(response)", msg.sessionId ?? "");
    handleUpstreamMessage(state, msg);
  });
  upstream_socket.addEventListener("close", () => {
    const closedDuringDownstreamShutdown = state.closing;
    state.closing = true;
    try {
      client.close();
    } catch {}
    if (!closedDuringDownstreamShutdown) onUpstreamClosed();
  });
  upstream_socket.addEventListener("error", (event) => {
    if (state.closing || client.readyState === client.CLOSING || client.readyState === client.CLOSED) {
      dbg("upstream ws error during shutdown");
      return;
    }
    log("upstream ws error", errorMessage(event));
    try {
      client.close(1011, "upstream error");
    } catch {}
    onUpstreamClosed();
  });
  client.on("close", () => {
    state.closing = true;
    void cdp.close().catch(() => {});
  });
  log(`extension ${cdp.connect_timing?.extension_source} (${cdp.extension_id}); ext session ${cdp.ext_session_id}`);

  // Swap the early-buffer handler for the real one. Drain anything that
  // arrived before we got here.
  client.off("message", earlyHandler);
  for (const buf of earlyBuffer) state.queued_from_client.push(buf);
  client.on("message", (buf) => {
    if (!state.bootstrapped) {
      state.queued_from_client.push(buf);
      return;
    }
    handleClientMessage(state, buf);
  });
  state.bootstrapped = true;
  for (const buf of state.queued_from_client) handleClientMessage(state, buf);
  state.queued_from_client = [];
}

async function handleClientManagedConnection(
  client: WebSocket,
  earlyBuffer: RawData[],
  earlyHandler: (buf: RawData) => void,
  {
    launch,
    upstream,
    extension,
    client: clientOptions,
    server,
  }: {
    launch: { mode?: string; executable_path?: string | null; user_data_dir?: string | null; options?: Record<string, unknown> };
    upstream: {
      mode?: string;
      ws_url?: string | null;
      nats_url?: string | null;
      nats_subject_prefix?: string | null;
      nativemessaging_manifest?: string | null;
    };
    extension: { mode?: string; path?: string | null };
    client?: Record<string, unknown>;
    server?: Record<string, unknown> | null;
  },
) {
  const cdp = new ModCDPClient({
    launch: {
      mode: (launch.mode ?? "none") as "local" | "remote" | "bb" | "none",
      executable_path: launch.executable_path,
      user_data_dir: launch.user_data_dir,
      options: launch.options as any,
    },
    upstream: {
      mode: upstream.mode as "pipe" | "nativemessaging" | "nats",
      ws_url: upstream.ws_url,
      nats_url: upstream.nats_url,
      nats_subject_prefix: upstream.nats_subject_prefix,
      nativemessaging_manifest: upstream.nativemessaging_manifest,
    },
    extension: {
      mode: (extension.mode ?? "none") as "auto" | "discover" | "inject" | "borrow" | "none",
      path: extension.path ?? DEFAULT_EXTENSION_PATH,
      service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      trust_service_worker_target: true,
    },
    client: { hydrate_aliases: false, ...clientOptions },
    server: server as any,
  });
  await cdp.connect();
  wireClientManagedConnection(client, earlyBuffer, earlyHandler, cdp, true);
}

function wireClientManagedConnection(
  client: WebSocket,
  earlyBuffer: RawData[],
  earlyHandler: (buf: RawData) => void,
  cdp: ModCDPClient,
  close_cdp_on_client_close: boolean,
) {
  const event_listener = (event_name: string | symbol, payload: unknown, session_id?: string | null) => {
    const event: CdpEventMessage = { method: String(event_name), params: (payload ?? {}) as Record<string, unknown> };
    if (typeof session_id === "string" && session_id) event.sessionId = session_id;
    sendRawClientMessage(client, event);
  };
  cdp.on("*", event_listener);

  client.off("message", earlyHandler);
  const handle = (buf: RawData) => {
    let msg: CdpCommandMessage;
    try {
      msg = CdpCommandMessageSchema.parse(JSON.parse(String(buf)));
    } catch (e) {
      log("client parse error", e.message);
      return;
    }
    const service_worker_params =
      msg.sessionId && msg.params && typeof msg.params === "object" && !("cdpSessionId" in msg.params)
        ? { ...msg.params, cdpSessionId: msg.sessionId }
        : (msg.params ?? {});
    const command_promise = ROUTE_TO_SW_RE.test(msg.method)
      ? cdp.send(msg.method, service_worker_params)
      : cdp.sendRaw(msg.method, msg.params ?? {}, msg.sessionId ?? null);
    void command_promise
      .then((result) =>
        sendRawClientMessage(client, {
          id: msg.id,
          result: result ?? {},
          ...(msg.sessionId ? { sessionId: msg.sessionId } : {}),
        }),
      )
      .catch((error) =>
        sendRawClientMessage(client, {
          id: msg.id,
          error: { code: -32000, message: errorMessage(error) },
          ...(msg.sessionId ? { sessionId: msg.sessionId } : {}),
        }),
      );
  };
  client.on("message", handle);
  for (const buf of earlyBuffer) handle(buf);
  client.on("close", () => {
    cdp.off("*", event_listener);
    if (close_cdp_on_client_close) void cdp.close().catch(() => {});
  });
}

function sendRawClientMessage(client: WebSocket, obj: unknown) {
  client.send(JSON.stringify(obj));
}

function handleClientMessage(state: ProxyConnectionState, buf: RawData) {
  let msg: CdpCommandMessage;
  try {
    msg = CdpCommandMessageSchema.parse(JSON.parse(String(buf)));
  } catch (e) {
    log("client parse error", e.message);
    return;
  }
  dbg("client->", msg.id ?? "", msg.method, msg.sessionId ?? "");
  const { id, method, params = {}, sessionId } = msg;

  // route a Mod.* / Custom.* command into a Runtime.callFunctionOn against the
  // hidden ext session, while remembering the originating client id+session
  // so the response can be steered back to the right Playwright CDPSession.
  if (MAGIC_METHODS.has(method) || ROUTE_TO_SW_RE.test(method)) {
    const upId = state.next_upstream_id++;
    state.pending.set(upId, { kind: "modcdp_eval", client_id: id, client_session_id: sessionId || null });
    let runtimeParams;
    if (method === "Mod.evaluate") {
      const evaluateParams = ModCDPEvaluateParamsSchema.parse(params ?? {});
      runtimeParams = wrapModCDPEvaluate({
        ...evaluateParams,
        cdpSessionId: evaluateParams.cdpSessionId ?? sessionId ?? null,
      });
    } else if (method === "Mod.addCustomCommand") {
      runtimeParams = wrapModCDPAddCustomCommand(ModCDPAddCustomCommandParamsSchema.parse(params ?? {}));
    } else if (method === "Mod.addCustomEvent") {
      const eventParams = ModCDPAddCustomEventObjectParamsSchema.parse(params ?? {});
      runtimeParams = wrapModCDPAddCustomEvent({ name: normalizeModCDPName(eventParams.name) });
    } else if (method === "Mod.addMiddleware") {
      runtimeParams = wrapModCDPAddMiddleware(ModCDPAddMiddlewareParamsSchema.parse(params ?? {}));
    } else {
      const cdpSessionId =
        params && typeof params === "object" && "cdpSessionId" in params && typeof params.cdpSessionId === "string"
          ? params.cdpSessionId
          : (sessionId ?? null);
      runtimeParams = wrapCustomCommand(method, params, cdpSessionId);
    }
    if (state.ext_execution_context_id != null) runtimeParams.executionContextId = state.ext_execution_context_id;
    state.upstream.send(
      JSON.stringify({
        id: upId,
        method: "Runtime.callFunctionOn",
        params: runtimeParams,
        sessionId: state.ext_session_id,
      }),
    );
    return;
  }

  // passthrough
  const upId = state.next_upstream_id++;
  state.pending.set(upId, { kind: "passthrough", client_id: id, client_session_id: sessionId || null });
  const out: CdpCommandMessage = { id: upId, method, params };
  if (sessionId) out.sessionId = sessionId;
  state.upstream.send(JSON.stringify(out));
}

function handleUpstreamMessage(state: ProxyConnectionState, msg: CdpResponseMessage | CdpEventMessage) {
  // response
  if ("id" in msg && typeof msg.id === "number") {
    const response = CdpResponseMessageSchema.parse(msg);
    const p = state.pending.get(response.id);
    if (!p) return;
    state.pending.delete(response.id);

    if (p.kind === "internal") {
      if (response.error) p.reject?.(new Error(response.error.message));
      else p.resolve?.(response.result || {});
      return;
    }

    const replyToClient = (extra: Omit<CdpResponseMessage, "id">) => {
      const out: CdpResponseMessage = { id: p.client_id ?? 0, ...extra };
      if (p.client_session_id) out.sessionId = p.client_session_id;
      sendToClient(state, out);
    };

    if (p.kind === "modcdp_eval") {
      try {
        replyToClient({ result: unwrapResponseIfNeeded(response.result || {}, "runtime") ?? {} });
      } catch (e) {
        replyToClient({ error: { code: -32000, message: e.message } });
      }
      return;
    }
    // passthrough
    if (response.error) replyToClient({ error: response.error });
    else replyToClient({ result: response.result ?? {} });
    return;
  }

  const event = CdpEventMessageSchema.parse(msg);

  if (event.method === "Target.attachedToTarget") {
    const attached = TargetEvents["Target.attachedToTarget"].parse(event.params || {});
    if (attached.sessionId) {
      state.target_session_ids.set(attached.targetInfo.targetId, attached.sessionId);
      if (state.hidden_target_ids.has(attached.targetInfo.targetId)) state.hidden_session_ids.add(attached.sessionId);
    }
  } else if (event.method === "Target.detachedFromTarget") {
    const detached = TargetEvents["Target.detachedFromTarget"].parse(event.params || {});
    if (detached.sessionId) {
      state.hidden_session_ids.delete(detached.sessionId);
      for (const [targetId, sessionId] of state.target_session_ids) {
        if (sessionId !== detached.sessionId) continue;
        state.target_session_ids.delete(targetId);
        break;
      }
    }
  }

  // event
  if (event.method === "Runtime.bindingCalled" && event.sessionId === state.ext_session_id) {
    const binding = RuntimeEvents["Runtime.bindingCalled"].parse(event.params || {});
    if (binding.name === UPSTREAM_EVENT_BINDING_NAME && !state.forward_mirrored_upstream_events) return;
    const u = unwrapEventIfNeeded(event.method, binding, event.sessionId || null, null);
    if (!u) return;
    // emit to root + every known client session, so any CDPSession listener
    // (Playwright per-target sessions) fires.
    sendToClient(state, { method: u.event, params: (u.data ?? {}) as Record<string, unknown> });
    for (const sid of state.client_session_ids) {
      sendToClient(state, { method: u.event, params: (u.data ?? {}) as Record<string, unknown>, sessionId: sid });
    }
    return;
  }

  // hide bridge-attached session traffic from the client
  if (event.sessionId && state.hidden_session_ids.has(event.sessionId)) return;

  // If the client's auto-attach creates a fresh orphan session against the
  // hidden SW target, hide that session and detach it upstream. This MUST run
  // before the generic hidden_target_ids drop below: for an attachedToTarget
  // event, msg.params.targetInfo.targetId is the SW target (which we want to
  // act on), not a target the client owns.
  if (event.method === "Target.attachedToTarget") {
    const attached = TargetEvents["Target.attachedToTarget"].parse(event.params || {});
    if (state.hidden_target_ids.has(attached.targetInfo.targetId)) {
      const orphan = attached.sessionId;
      if (orphan && orphan !== state.ext_session_id) {
        state.hidden_session_ids.add(orphan);
        const upId = state.next_upstream_id++;
        state.pending.set(upId, { kind: "internal", resolve: () => {}, reject: () => {} });
        state.upstream.send(
          JSON.stringify({ id: upId, method: "Target.detachFromTarget", params: { sessionId: orphan } }),
        );
      }
      return;
    }
  }

  // hide all other events about the extension SW target.
  const targetId =
    event.params &&
    typeof event.params === "object" &&
    "targetInfo" in event.params &&
    event.params.targetInfo &&
    typeof event.params.targetInfo === "object" &&
    "targetId" in event.params.targetInfo &&
    typeof event.params.targetInfo.targetId === "string"
      ? event.params.targetInfo.targetId
      : event.params &&
          typeof event.params === "object" &&
          "targetId" in event.params &&
          typeof event.params.targetId === "string"
        ? event.params.targetId
        : null;
  if (targetId && state.hidden_target_ids.has(targetId)) return;
  const eventSessionId =
    event.params &&
    typeof event.params === "object" &&
    "sessionId" in event.params &&
    typeof event.params.sessionId === "string"
      ? event.params.sessionId
      : null;
  if (event.method.startsWith("Target.detached") && eventSessionId && state.hidden_session_ids.has(eventSessionId))
    return;

  if (!state.bootstrapped) return; // do not leak bootstrap events

  if (event.method === "Target.attachedToTarget" && eventSessionId) {
    state.client_session_ids.add(eventSessionId);
  }
  if (event.method === "Target.detachedFromTarget" && eventSessionId) {
    state.client_session_ids.delete(eventSessionId);
  }

  sendToClient(state, event);
}

function sendToClient(state: ProxyConnectionState, obj: CdpMessage) {
  if (DEBUG)
    dbg("client<-", "id" in obj ? obj.id : "", "method" in obj ? obj.method : "(response)", obj.sessionId ?? "");
  state.client.send(JSON.stringify(obj));
}

// --- CLI -------------------------------------------------------------------

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const argv = parseProxyArgs(process.argv.slice(2));
  const listen = argv.listen ? parseHostPort(String(argv.listen), DEFAULT_HOST, DEFAULT_PORT) : null;
  const host = listen?.host ?? DEFAULT_HOST;
  const port = listen?.port ?? Number(argv.port || DEFAULT_PORT);
  const extensionPath =
    typeof argv["extension-path"] === "string" && argv["extension-path"] !== "true"
      ? path.resolve(argv["extension-path"])
      : DEFAULT_EXTENSION_PATH;
  const forward_mirrored_upstream_events = argv["forward-mirrored-upstream-events"] === "true";
  const clientConfig =
    typeof argv.client === "string" && argv.client !== "true"
      ? JSON.parse(argv.client)
      : typeof argv["client-routes"] === "string" && argv["client-routes"] !== "true"
        ? { routes: JSON.parse(argv["client-routes"]) }
        : {};
  const serverConfig =
    typeof argv.server === "string" && argv.server !== "true"
      ? JSON.parse(argv.server)
      : typeof argv["server-routes"] === "string" && argv["server-routes"] !== "true"
        ? { routes: JSON.parse(argv["server-routes"]) }
        : {};
  startProxy({
    host,
    port,
    launch: {
      mode: String(argv.launch || "remote"),
      executable_path:
        typeof argv["launch-executable-path"] === "string" && argv["launch-executable-path"] !== "true"
          ? String(argv["launch-executable-path"])
          : null,
      user_data_dir:
        typeof argv["launch-user-data-dir"] === "string" && argv["launch-user-data-dir"] !== "true"
          ? String(argv["launch-user-data-dir"])
          : null,
      options:
        typeof argv["launch-options"] === "string" && argv["launch-options"] !== "true"
          ? JSON.parse(argv["launch-options"])
          : {},
    },
    upstream: {
      mode: String(argv.upstream || "ws"),
      ws_url:
        typeof argv["upstream-ws-url"] === "string" && argv["upstream-ws-url"] !== "true"
          ? String(argv["upstream-ws-url"])
          : DEFAULT_UPSTREAM,
      nats_url:
        typeof argv["upstream-nats-url"] === "string" && argv["upstream-nats-url"] !== "true"
          ? String(argv["upstream-nats-url"])
          : null,
      nats_subject_prefix:
        typeof argv["upstream-nats-subject-prefix"] === "string" &&
        argv["upstream-nats-subject-prefix"] !== "true"
          ? String(argv["upstream-nats-subject-prefix"])
          : null,
      reversews_bind:
        typeof argv["upstream-reversews-bind"] === "string" && argv["upstream-reversews-bind"] !== "true"
          ? String(argv["upstream-reversews-bind"])
          : "127.0.0.1:29292",
      reversews_wait_timeout_ms:
        typeof argv["upstream-reversews-wait-timeout-ms"] === "string" &&
        argv["upstream-reversews-wait-timeout-ms"] !== "true"
          ? Number(argv["upstream-reversews-wait-timeout-ms"])
          : null,
      nativemessaging_manifest:
        typeof argv["upstream-nativemessaging-manifest"] === "string" &&
        argv["upstream-nativemessaging-manifest"] !== "true"
          ? String(argv["upstream-nativemessaging-manifest"])
          : null,
    },
    extension: { mode: String(argv.extension || "auto"), path: extensionPath },
    client: clientConfig,
    server: serverConfig,
    forward_mirrored_upstream_events,
  }).catch((e) => {
    console.error(e);
    process.exitCode = 1;
  });
}

function parseProxyArgs(args: string[]) {
  const result: Record<string, string> = {};
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (!arg.startsWith("--")) continue;
    const raw = arg.slice(2);
    const equals = raw.indexOf("=");
    if (equals >= 0) {
      result[raw.slice(0, equals)] = raw.slice(equals + 1);
      continue;
    }
    const next = args[i + 1];
    result[raw] = next && !next.startsWith("--") ? next : "true";
    if (next && !next.startsWith("--")) i += 1;
  }
  return result;
}
