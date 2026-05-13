// ModCDPServer: lives inside an extension service worker. Owns the registry
// of custom commands and event bindings, and emits events through the binding
// API installed by the client (Runtime.addBinding -> globalThis[__ModCDP_custom_event__]).
//
// The installer is intentionally self-contained so the bridge can inject the
// same server implementation into an already-running extension service worker
// when Chrome refuses Extensions.loadUnpacked.

import type { cdp } from "../types/generated/cdp.js";
import type {
  CdpCommandMessage,
  CdpEventMessage,
  CdpDebuggeeCommandParams,
  ModCDPConfigureParams,
  ModCDPCustomCommandRegistration,
  ModCDPCustomEventRegistration,
  ModCDPMiddlewareRegistration,
  ModCDPPingParams,
  ModCDPRoutes,
  ProtocolParams,
  ProtocolPayload,
  ProtocolResult,
} from "../types/modcdp.js";

export const DEFAULT_CDP_SEND_TIMEOUT_MS = 10_000;
export const DEFAULT_LOOPBACK_EXECUTION_CONTEXT_TIMEOUT_MS = 10_000;
export const DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS = 250;
export const DEFAULT_REVERSE_BRIDGE_RECONNECT_INTERVAL_MS = 2_000;
export const DEFAULT_NATIVE_BRIDGE_HOST_NAME = "com.modcdp.bridge";
export const DEFAULT_NATIVE_BRIDGE_RECONNECT_INTERVAL_MS = 2_000;
export const DEFAULT_NATS_BRIDGE_RECONNECT_INTERVAL_MS = 2_000;
export const DEFAULT_NATS_BRIDGE_SUBJECT_PREFIX = "modcdp.default";

type MiddlewarePhase = "request" | "response" | "event";
type ModCDPGlobalScope = typeof globalThis &
  Record<string, unknown> & {
    ModCDP?: {
      __ModCDPServerVersion?: number;
      addCustomEvent?: unknown;
      handleCommand?: unknown;
    };
  };

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return value != null && typeof value === "object" && !Array.isArray(value);
}

export function installModCDPServer(globalScope: ModCDPGlobalScope = globalThis as ModCDPGlobalScope) {
  const MODCDP_SERVER_VERSION = 1;
  const DEFAULT_CDP_SEND_TIMEOUT_MS = 10_000;
  const DEFAULT_LOOPBACK_EXECUTION_CONTEXT_TIMEOUT_MS = 10_000;
  const DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS = 250;
  const DEFAULT_REVERSE_BRIDGE_RECONNECT_INTERVAL_MS = 2_000;
  const DEFAULT_NATIVE_BRIDGE_HOST_NAME = "com.modcdp.bridge";
  const DEFAULT_NATIVE_BRIDGE_RECONNECT_INTERVAL_MS = 2_000;
  const DEFAULT_NATS_BRIDGE_RECONNECT_INTERVAL_MS = 2_000;
  const DEFAULT_NATS_BRIDGE_SUBJECT_PREFIX = "modcdp.default";
  if (
    globalScope.ModCDP?.__ModCDPServerVersion === MODCDP_SERVER_VERSION &&
    globalScope.ModCDP?.handleCommand &&
    globalScope.ModCDP?.addCustomEvent
  )
    return globalScope.ModCDP;

  const UPSTREAM_EVENT_BINDING_NAME = "__ModCDP_event_from_upstream__";
  const CUSTOM_EVENT_BINDING_NAME = "__ModCDP_custom_event__";
  const encodeBindingPayload = ({
    event,
    data,
    cdpSessionId = null,
  }: {
    event: string;
    data: ProtocolPayload;
    cdpSessionId?: string | null;
  }) => JSON.stringify({ event, data, cdpSessionId });

  const commandHandlers = new Map<string, ModCDPCustomCommandRegistration>();
  const eventBindings = new Map<string, ModCDPCustomEventRegistration>();
  const eventListeners = new Set<(event: string, data: ProtocolPayload, cdpSessionId: string | null) => void>();
  const middlewares: Record<MiddlewarePhase, ModCDPMiddlewareRegistration[]> = {
    request: [],
    response: [],
    event: [],
  };
  const attachedDebuggees = new Set<string>();
  let runtime_types_promise: Promise<unknown> | null = null;

  async function publishEvent(eventName: string, payload: ProtocolPayload = {}, cdpSessionId: string | null = null) {
    payload = await ModCDPServer.runMiddleware("event", eventName, payload, {
      cdpSessionId,
      event: { name: eventName, payload },
    });
    if (payload === undefined) return { event: eventName, emitted: false, reason: "middleware_dropped" };

    for (const listener of eventListeners) {
      try {
        listener(eventName, payload, cdpSessionId);
      } catch (error) {
        console.error("[ModCDPServer] event listener failed", error);
      }
    }
    let emittedThroughReverseBridge = false;
    if (reverseBridgeSocket?.readyState === WebSocket.OPEN) {
      const message: CdpEventMessage = {
        method: eventName,
        params: (payload ?? {}) as CdpEventMessage["params"],
      };
      if (cdpSessionId) message.sessionId = cdpSessionId;
      reverseBridgeSocket.send(JSON.stringify(message));
      emittedThroughReverseBridge = true;
    }
    let emittedThroughNativeBridge = false;
    if (nativeBridgePort) {
      const message: CdpEventMessage = {
        method: eventName,
        params: (payload ?? {}) as CdpEventMessage["params"],
      };
      if (cdpSessionId) message.sessionId = cdpSessionId;
      nativeBridgePort.postMessage(message);
      emittedThroughNativeBridge = true;
    }
    let emittedThroughNatsBridge = false;
    if (nats_bridge_socket?.readyState === WebSocket.OPEN) {
      const message: CdpEventMessage = {
        method: eventName,
        params: (payload ?? {}) as CdpEventMessage["params"],
      };
      if (cdpSessionId) message.sessionId = cdpSessionId;
      publishNats(`${nats_bridge_subject_prefix}.browser_to_client`, { type: "modcdp.nats.message", message });
      emittedThroughNatsBridge = true;
    }

    const isCustomEvent = registryMatch(eventBindings, eventName) != null;
    let emittedThroughBinding = false;
    if (isCustomEvent) {
      const customBinding = globalScope[CUSTOM_EVENT_BINDING_NAME];
      if (typeof customBinding === "function") {
        customBinding(encodeBindingPayload({ event: eventName, data: payload, cdpSessionId }));
        emittedThroughBinding = true;
      }
    } else {
      const mirrorBinding = globalScope[UPSTREAM_EVENT_BINDING_NAME];
      if (typeof mirrorBinding === "function") {
        mirrorBinding(encodeBindingPayload({ event: eventName, data: payload, cdpSessionId }));
        emittedThroughBinding = true;
      }
    }
    return emittedThroughBinding ||
      emittedThroughReverseBridge ||
      emittedThroughNativeBridge ||
      emittedThroughNatsBridge
      ? { event: eventName, emitted: true }
      : { event: eventName, emitted: false, reason: "binding_not_installed" };
  }

  const targetAutoAttachParams = {
    autoAttach: true,
    waitForDebuggerOnStart: false,
    flatten: true,
  } satisfies cdp.types.ts.Target.SetAutoAttachParams;

  const defaultRoutes = {
    "Mod.*": "service_worker",
    "Custom.*": "service_worker",
    "*.*": "auto",
  } satisfies ModCDPRoutes;

  const browserLevelDomains = new Set(["Browser", "Target", "SystemInfo"]);

  let nextLoopbackId = 1;
  const loopbackSockets = new Map<string, WebSocket>();
  const loopbackSocketPromises = new Map<string, Promise<WebSocket>>();
  const loopbackTargetSessions = new Map<string, string>();
  const loopbackSessionTargets = new Map<string, string>();
  const loopbackSessionContexts = new Map<string, number>();
  const loopbackContextWaiters = new Map<string, Set<(contextId: number) => void>>();
  const initializedLoopbackSockets = new WeakSet<WebSocket>();
  const loopbackPending = new Map<
    number,
    { resolve: (value: ProtocolResult) => void; reject: (error: Error) => void }
  >();
  let reverseBridgeSocket: WebSocket | null = null;
  let reverseBridgeUrl: string | null = null;
  let reverseBridgeReconnectIntervalMs = DEFAULT_REVERSE_BRIDGE_RECONNECT_INTERVAL_MS;
  let reverseBridgeReconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let nativeBridgePort: chrome.runtime.Port | null = null;
  let nativeBridgeHostName: string | null = null;
  let nativeBridgeReconnectIntervalMs = DEFAULT_NATIVE_BRIDGE_RECONNECT_INTERVAL_MS;
  let nativeBridgeReconnectTimer: ReturnType<typeof setTimeout> | null = null;
  let nats_bridge_socket: WebSocket | null = null;
  let nats_bridge_url: string | null = null;
  let nats_bridge_subject_prefix = DEFAULT_NATS_BRIDGE_SUBJECT_PREFIX;
  let nats_bridge_reconnect_interval_ms = DEFAULT_NATS_BRIDGE_RECONNECT_INTERVAL_MS;
  let nats_bridge_reconnect_timer: ReturnType<typeof setTimeout> | null = null;
  let nats_bridge_buffer = "";
  let selfDebuggee: chrome.debugger.Debuggee | null = null;
  const offscreenKeepAlivePortName = "ModCDPOffscreenKeepAlive";
  const offscreenKeepAlivePath = "offscreen/keepalive.html";
  let creatingOffscreenKeepAlive: Promise<void> | null = null;
  let offscreenKeepAlivePort: chrome.runtime.Port | null = null;

  function registryMatch<T>(registry: Map<string, T>, name: string): T | null {
    const exact = registry.get(name);
    if (exact) return exact;
    let match: T | null = null;
    let matchPrefixLength = -1;
    for (const [pattern, value] of registry) {
      if (!pattern.endsWith(".*")) continue;
      const prefix = pattern.slice(0, -1);
      if (!name.startsWith(prefix) || prefix.length <= matchPrefixLength) continue;
      match = value;
      matchPrefixLength = prefix.length;
    }
    return match;
  }

  function normalizeModCDPName(
    value:
      | {
          cdp_command_name?: string;
          cdp_event_name?: string;
          id?: string;
          name?: string;
          meta?: () =>
            | { cdp_command_name?: unknown; cdp_event_name?: unknown; id?: unknown; name?: unknown }
            | undefined;
        }
      | string,
  ) {
    if (typeof value === "string") return value;
    const meta = typeof value?.meta === "function" ? value.meta() : undefined;
    const name =
      value?.cdp_command_name ??
      value?.cdp_event_name ??
      (typeof meta?.cdp_command_name === "string" ? meta.cdp_command_name : undefined) ??
      (typeof meta?.cdp_event_name === "string" ? meta.cdp_event_name : undefined) ??
      value?.id ??
      (typeof meta?.id === "string" ? meta.id : undefined) ??
      (typeof meta?.name === "string" ? meta.name : undefined) ??
      value?.name;
    if (typeof name !== "string" || !name) throw new Error("Expected a CDP name string or a named CDP schema/alias.");
    return name;
  }

  function errorMessage(error: unknown): string {
    return error instanceof Error ? error.message : String(error);
  }

  function compactDebuggee(input: {
    [Key in keyof chrome.debugger.Debuggee]?: chrome.debugger.Debuggee[Key] | null;
  }): chrome.debugger.Debuggee {
    return {
      ...(typeof input.tabId === "number" ? { tabId: input.tabId } : {}),
      ...(typeof input.targetId === "string" ? { targetId: input.targetId } : {}),
      ...(typeof input.extensionId === "string" ? { extensionId: input.extensionId } : {}),
    };
  }

  async function resolveCDPEndpoint(endpoint: string | null) {
    if (!endpoint || /^wss?:\/\//i.test(endpoint)) return endpoint;
    if (!/^https?:\/\//i.test(endpoint)) {
      throw new Error(`loopback_cdp_url must be a ws://, wss://, http://, or https:// CDP endpoint, got ${endpoint}.`);
    }
    const { webSocketDebuggerUrl } = await fetch(`${endpoint}/json/version`).then((r) => r.json());
    if (!webSocketDebuggerUrl) throw new Error(`loopback_cdp_url HTTP discovery returned no webSocketDebuggerUrl.`);
    return webSocketDebuggerUrl;
  }

  async function openCDPSocket(endpoint: string): Promise<WebSocket> {
    if (!/^wss?:\/\//i.test(endpoint)) {
      throw new Error(`loopback_cdp_url must be a ws:// or wss:// CDP websocket URL, got ${endpoint}.`);
    }
    return new Promise<WebSocket>((resolve, reject) => {
      const w = new WebSocket(endpoint);
      let settled = false;
      let errorEvent: Event | null = null;
      const describe = (prefix: string, closeEvent?: CloseEvent) => {
        const parts = [`${prefix} ${endpoint}`, `readyState=${w.readyState}`];
        if (errorEvent) parts.push(`error.type=${errorEvent.type}`);
        if (closeEvent) {
          parts.push(`close.code=${closeEvent.code}`);
          parts.push(`close.reason=${closeEvent.reason || ""}`);
          parts.push(`close.wasClean=${closeEvent.wasClean}`);
        }
        return parts.join(" ");
      };
      const fail = (error: Error) => {
        if (settled) return;
        settled = true;
        reject(error);
      };
      w.addEventListener(
        "open",
        () => {
          if (settled) return;
          settled = true;
          resolve(w);
        },
        { once: true },
      );
      w.addEventListener(
        "error",
        (event) => {
          errorEvent = event;
          setTimeout(
            () => fail(new Error(describe("CDP socket error"))),
            ModCDPServer.ws_connect_error_settle_timeout_ms,
          );
        },
        { once: true },
      );
      w.addEventListener("close", (event) => fail(new Error(describe("CDP socket closed", event))), { once: true });
    });
  }

  function startOffscreenKeepAlive() {
    void ensureOffscreenKeepAlive().catch(() => {});
  }

  function rejectLoopbackPending(error: Error) {
    for (const pending of loopbackPending.values()) pending.reject(error);
    loopbackPending.clear();
  }

  function scheduleReverseBridgeReconnect(delayMs: number) {
    if (!reverseBridgeUrl) return;
    if (reverseBridgeReconnectTimer) return;
    reverseBridgeReconnectTimer = setTimeout(() => {
      reverseBridgeReconnectTimer = null;
      void connectReverseBridge(reverseBridgeUrl).catch(() => {});
    }, delayMs);
  }

  function stopReverseBridge(reason = "stopped") {
    const upstream_reversews_url = reverseBridgeUrl;
    reverseBridgeUrl = null;
    if (reverseBridgeReconnectTimer) {
      clearTimeout(reverseBridgeReconnectTimer);
      reverseBridgeReconnectTimer = null;
    }
    const socket = reverseBridgeSocket;
    reverseBridgeSocket = null;
    if (socket?.readyState === WebSocket.OPEN || socket?.readyState === WebSocket.CONNECTING) {
      socket.close(1000, reason);
    }
    return { upstream_reversews_url, stopped: true, reason };
  }

  function scheduleNativeBridgeReconnect(delayMs: number) {
    if (!nativeBridgeHostName) return;
    if (nativeBridgeReconnectTimer) return;
    nativeBridgeReconnectTimer = setTimeout(() => {
      nativeBridgeReconnectTimer = null;
      connectNativeBridge(nativeBridgeHostName);
    }, delayMs);
  }

  function scheduleNatsBridgeReconnect(delayMs: number) {
    if (!nats_bridge_url) return;
    if (nats_bridge_reconnect_timer) return;
    nats_bridge_reconnect_timer = setTimeout(() => {
      nats_bridge_reconnect_timer = null;
      void connectNatsBridge(nats_bridge_url).catch(() => {});
    }, delayMs);
  }

  async function handleReverseBridgeMessage(ws: WebSocket, data: unknown) {
    let message: CdpCommandMessage;
    try {
      const parsed = JSON.parse(typeof data === "string" ? data : String(data));
      if (typeof parsed?.id !== "number" || typeof parsed?.method !== "string") return;
      message = parsed as CdpCommandMessage;
    } catch {
      return;
    }

    try {
      const result = await ModCDPServer.handleCommand(message.method, message.params ?? {}, message.sessionId ?? null);
      ws.send(JSON.stringify({ id: message.id, result }));
    } catch (error) {
      ws.send(
        JSON.stringify({
          id: message.id,
          error: {
            code: -32000,
            message: errorMessage(error),
          },
        }),
      );
    }
  }

  async function handleNativeBridgeMessage(port: chrome.runtime.Port, data: unknown) {
    let message: CdpCommandMessage;
    try {
      if (
        typeof (data as CdpCommandMessage)?.id !== "number" ||
        typeof (data as CdpCommandMessage)?.method !== "string"
      )
        return;
      message = data as CdpCommandMessage;
    } catch {
      return;
    }

    try {
      const result = await ModCDPServer.handleCommand(message.method, message.params ?? {}, message.sessionId ?? null);
      port.postMessage({ id: message.id, result });
    } catch (error) {
      port.postMessage({
        id: message.id,
        error: {
          code: -32000,
          message: errorMessage(error),
        },
      });
    }
  }

  async function handleNatsBridgePayload(payload: string) {
    let parsed: unknown;
    try {
      parsed = JSON.parse(payload);
    } catch {
      return;
    }
    const record = parsed && typeof parsed === "object" ? (parsed as Record<string, unknown>) : null;
    if (record?.type === "modcdp.nats.hello") {
      publishNats(`${nats_bridge_subject_prefix}.browser_to_client`, {
        type: "modcdp.nats.hello",
        role: "extension-service-worker",
        version: 1,
        extension_id: globalScope.chrome?.runtime?.id ?? null,
      });
      return;
    }
    const candidate = record?.type === "modcdp.nats.message" ? record.message : parsed;
    if (
      !candidate ||
      typeof candidate !== "object" ||
      typeof (candidate as CdpCommandMessage).id !== "number" ||
      typeof (candidate as CdpCommandMessage).method !== "string"
    )
      return;
    const message = candidate as CdpCommandMessage;
    try {
      const result = await ModCDPServer.handleCommand(message.method, message.params ?? {}, message.sessionId ?? null);
      publishNats(`${nats_bridge_subject_prefix}.browser_to_client`, {
        type: "modcdp.nats.message",
        message: { id: message.id, result },
      });
    } catch (error) {
      publishNats(`${nats_bridge_subject_prefix}.browser_to_client`, {
        type: "modcdp.nats.message",
        message: {
          id: message.id,
          error: {
            code: -32000,
            message: errorMessage(error),
          },
        },
      });
    }
  }

  async function connectReverseBridge(endpoint: string) {
    if (
      reverseBridgeSocket?.readyState === WebSocket.OPEN ||
      reverseBridgeSocket?.readyState === WebSocket.CONNECTING
    ) {
      return { upstream_reversews_url: endpoint, connected: reverseBridgeSocket.readyState === WebSocket.OPEN };
    }

    const ws = new WebSocket(endpoint);
    reverseBridgeSocket = ws;
    ws.addEventListener("open", () => {
      startOffscreenKeepAlive();
      ws.send(
        JSON.stringify({
          type: "modcdp.reverse.hello",
          role: "extension-service-worker",
          version: 1,
          extension_id: globalScope.chrome?.runtime?.id ?? null,
        }),
      );
    });
    ws.addEventListener("message", (event) => {
      void handleReverseBridgeMessage(ws, event.data);
    });
    ws.addEventListener("error", () => {
      if (reverseBridgeSocket === ws) reverseBridgeSocket = null;
      scheduleReverseBridgeReconnect(reverseBridgeReconnectIntervalMs);
    });
    ws.addEventListener("close", () => {
      if (reverseBridgeSocket === ws) reverseBridgeSocket = null;
      scheduleReverseBridgeReconnect(reverseBridgeReconnectIntervalMs);
    });
    return { upstream_reversews_url: endpoint, connected: false };
  }

  function connectNativeBridge(hostName: string) {
    const chromeApi = globalScope.chrome;
    if (!chromeApi?.runtime?.connectNative) {
      scheduleNativeBridgeReconnect(nativeBridgeReconnectIntervalMs);
      return { upstream_nativemessaging_host_name: hostName, connected: false, reason: "native_messaging_unavailable" };
    }
    if (nativeBridgePort) return { upstream_nativemessaging_host_name: hostName, connected: true };
    try {
      ModCDPServer.native_bridge_attempts += 1;
      ModCDPServer.native_bridge_last_error = null;
      const port = chromeApi.runtime.connectNative(hostName);
      nativeBridgePort = port;
      ModCDPServer.native_bridge_connected = true;
      startOffscreenKeepAlive();
      port.postMessage({
        type: "modcdp.native.hello",
        role: "extension-service-worker",
        version: 1,
        extension_id: globalScope.chrome?.runtime?.id ?? null,
      });
      port.onMessage.addListener((message) => {
        void handleNativeBridgeMessage(port, message);
      });
      port.onDisconnect.addListener(() => {
        if (nativeBridgePort === port) nativeBridgePort = null;
        ModCDPServer.native_bridge_connected = false;
        ModCDPServer.native_bridge_last_error =
          chromeApi.runtime.lastError?.message ?? "Native messaging port disconnected.";
        scheduleNativeBridgeReconnect(nativeBridgeReconnectIntervalMs);
      });
      return { upstream_nativemessaging_host_name: hostName, connected: true };
    } catch (error) {
      nativeBridgePort = null;
      ModCDPServer.native_bridge_connected = false;
      ModCDPServer.native_bridge_last_error = errorMessage(error);
      scheduleNativeBridgeReconnect(nativeBridgeReconnectIntervalMs);
      return { upstream_nativemessaging_host_name: hostName, connected: false, reason: errorMessage(error) };
    }
  }

  async function connectNatsBridge(endpoint: string) {
    if (!/^wss?:\/\//i.test(endpoint)) {
      throw new Error(`nats bridge endpoint must be a ws:// or wss:// URL for extension transport, got ${endpoint}.`);
    }
    if (nats_bridge_socket?.readyState === WebSocket.OPEN || nats_bridge_socket?.readyState === WebSocket.CONNECTING) {
      return {
        upstream_nats_url: endpoint,
        upstream_nats_subject_prefix: nats_bridge_subject_prefix,
        connected: nats_bridge_socket.readyState === WebSocket.OPEN,
      };
    }
    const ws = new WebSocket(endpoint);
    nats_bridge_socket = ws;
    nats_bridge_buffer = "";
    ws.addEventListener("open", () => {
      startOffscreenKeepAlive();
      writeNats(`CONNECT ${JSON.stringify(natsConnectOptions())}\r\nPING\r\n`);
      writeNats(`SUB ${nats_bridge_subject_prefix}.client_to_browser 1\r\n`);
      publishNats(`${nats_bridge_subject_prefix}.browser_to_client`, {
        type: "modcdp.nats.hello",
        role: "extension-service-worker",
        version: 1,
        extension_id: globalScope.chrome?.runtime?.id ?? null,
      });
    });
    ws.addEventListener("message", (event) => {
      void readNatsWebSocketData(event.data);
    });
    ws.addEventListener("error", () => {
      if (nats_bridge_socket === ws) nats_bridge_socket = null;
      scheduleNatsBridgeReconnect(nats_bridge_reconnect_interval_ms);
    });
    ws.addEventListener("close", () => {
      if (nats_bridge_socket === ws) nats_bridge_socket = null;
      scheduleNatsBridgeReconnect(nats_bridge_reconnect_interval_ms);
    });
    return { upstream_nats_url: endpoint, upstream_nats_subject_prefix: nats_bridge_subject_prefix, connected: false };
  }

  function writeNats(data: string) {
    if (nats_bridge_socket?.readyState === WebSocket.OPEN) nats_bridge_socket.send(data);
  }

  function publishNats(subject: string, message: unknown) {
    const body = JSON.stringify(message);
    writeNats(`PUB ${subject} ${new TextEncoder().encode(body).byteLength}\r\n${body}\r\n`);
  }

  async function readNatsWebSocketData(data: unknown) {
    if (typeof data === "string") nats_bridge_buffer += data;
    else if (data instanceof ArrayBuffer) nats_bridge_buffer += new TextDecoder().decode(data);
    else if (ArrayBuffer.isView(data)) nats_bridge_buffer += new TextDecoder().decode(data);
    else if (typeof Blob !== "undefined" && data instanceof Blob) nats_bridge_buffer += await data.text();
    else return;
    nats_bridge_buffer = consumeNatsProtocol(nats_bridge_buffer);
  }

  function consumeNatsProtocol(buffer: string) {
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
        void handleNatsBridgePayload(payload);
        continue;
      }
      buffer = buffer.slice(lineEnd + 2);
      if (upper === "PING") writeNats("PONG\r\n");
    }
  }

  function natsConnectOptions() {
    return { verbose: false, pedantic: false, lang: "modcdp-extension", version: "1", protocol: 1 };
  }

  function debuggerSendCommand(
    debuggee: chrome.debugger.Debuggee,
    method: string,
    params: Record<string, unknown> = {},
  ): Promise<ProtocolResult> {
    const chromeApi = globalScope.chrome;
    return new Promise<ProtocolResult>((resolve, reject) =>
      chromeApi.debugger.sendCommand(debuggee, method, params, (result) => {
        const error = chromeApi.runtime.lastError;
        if (error) reject(new Error(error.message));
        else resolve(result as ProtocolResult);
      }),
    );
  }

  async function getSelfDebuggee(): Promise<chrome.debugger.Debuggee> {
    if (selfDebuggee) return selfDebuggee;
    const chromeApi = globalScope.chrome;
    if (!chromeApi?.debugger?.getTargets || !chromeApi?.debugger?.attach) {
      throw new Error("chrome.debugger is unavailable for reverse expression evaluation.");
    }
    const serviceWorkerUrl = currentServiceWorkerUrl();
    const targets = await chromeApi.debugger.getTargets();
    const target = targets.find((candidate) => candidate.url === serviceWorkerUrl);
    if (!target?.id) throw new Error(`Could not find ModCDP service worker debugger target ${serviceWorkerUrl}.`);
    const debuggee = { targetId: target.id };
    await new Promise<void>((resolve, reject) =>
      chromeApi.debugger.attach(debuggee, "1.3", () => {
        const error = chromeApi.runtime.lastError;
        if (!error || error.message?.includes("Another debugger is already attached")) resolve();
        else reject(new Error(error.message));
      }),
    );
    selfDebuggee = debuggee;
    return debuggee;
  }

  function currentServiceWorkerUrl() {
    const chromeApi = globalScope.chrome;
    const manifest = chromeApi?.runtime?.getManifest?.();
    const service_worker =
      manifest && typeof manifest === "object" && "background" in manifest
        ? (manifest.background as { service_worker?: unknown } | undefined)?.service_worker
        : null;
    const service_worker_path =
      typeof service_worker === "string" && service_worker.length > 0
        ? service_worker.replace(/^\//, "")
        : "modcdp/service_worker.js";
    return chromeApi.runtime.getURL(service_worker_path);
  }

  async function evaluateInSelf(expression: string): Promise<ProtocolResult> {
    const debuggee = await getSelfDebuggee();
    const result = (await debuggerSendCommand(debuggee, "Runtime.evaluate", {
      expression,
      awaitPromise: true,
      returnByValue: true,
    })) as cdp.types.ts.Runtime.EvaluateResult;
    if (result.exceptionDetails) {
      const ex = result.exceptionDetails;
      throw new Error(ex.exception?.description || ex.text || "Runtime evaluation failed");
    }
    return (result.result?.value ?? {}) as ProtocolResult;
  }

  async function evaluateUserExpression({
    expression,
    params = {},
    cdpSessionId = null,
    method = null,
  }: {
    expression: string;
    params?: ProtocolPayload;
    cdpSessionId?: string | null;
    method?: string | null;
  }): Promise<ProtocolResult> {
    return evaluateInSelf(`
      (async () => {
        const params = ${JSON.stringify(params ?? {})};
        const method = ${JSON.stringify(method)};
        const cdp = globalThis.ModCDP.attachToSession(${JSON.stringify(cdpSessionId)});
        const ModCDP = globalThis.ModCDP;
        const chrome = globalThis.chrome;
        const value = (${expression});
        return typeof value === "function" ? await value(params || {}, method) : value;
      })()
    `);
  }

  async function loopbackWS(endpoint: string): Promise<WebSocket> {
    const existing = loopbackSockets.get(endpoint);
    if (existing?.readyState === WebSocket.OPEN) return existing;
    const pending = loopbackSocketPromises.get(endpoint);
    if (pending) return pending;

    const nextSocket = openCDPSocket(endpoint).then((ws) => {
      loopbackSockets.set(endpoint, ws);
      loopbackSocketPromises.delete(endpoint);
      ws.addEventListener("message", (event) => {
        const msg = JSON.parse(event.data);
        const id = typeof msg.id === "number" ? msg.id : null;
        if (id == null) {
          const method = typeof msg.method === "string" ? msg.method : null;
          if (!method) return;
          const payload =
            msg.params && typeof msg.params === "object" && !Array.isArray(msg.params)
              ? (msg.params as ProtocolPayload)
              : {};
          const cdpSessionId = typeof msg.sessionId === "string" ? msg.sessionId : null;
          const payloadRecord = payload as Record<string, unknown>;
          const targetInfo =
            payloadRecord.targetInfo &&
            typeof payloadRecord.targetInfo === "object" &&
            !Array.isArray(payloadRecord.targetInfo)
              ? (payloadRecord.targetInfo as Record<string, unknown>)
              : null;
          const attachedSessionId = typeof payloadRecord.sessionId === "string" ? payloadRecord.sessionId : null;
          const attachedTargetId = typeof targetInfo?.targetId === "string" ? targetInfo.targetId : null;
          if (method === "Target.attachedToTarget" && attachedSessionId != null && attachedTargetId != null) {
            loopbackTargetSessions.set(attachedTargetId, attachedSessionId);
            loopbackSessionTargets.set(attachedSessionId, attachedTargetId);
          } else if (method === "Target.detachedFromTarget") {
            const detachedSessionId =
              typeof payloadRecord.sessionId === "string" ? payloadRecord.sessionId : cdpSessionId;
            const detachedTargetId =
              typeof payloadRecord.targetId === "string"
                ? payloadRecord.targetId
                : detachedSessionId == null
                  ? null
                  : (loopbackSessionTargets.get(detachedSessionId) ?? null);
            if (detachedTargetId != null) loopbackTargetSessions.delete(detachedTargetId);
            if (detachedSessionId != null) loopbackSessionTargets.delete(detachedSessionId);
            if (detachedSessionId != null) loopbackSessionContexts.delete(detachedSessionId);
          } else if (method === "Runtime.executionContextCreated" && cdpSessionId != null) {
            const context = payloadRecord.context;
            const contextId =
              context && typeof context === "object" && "id" in context && typeof context.id === "number"
                ? context.id
                : null;
            if (contextId != null) {
              loopbackSessionContexts.set(cdpSessionId, contextId);
              const waiters = loopbackContextWaiters.get(cdpSessionId);
              if (waiters) {
                loopbackContextWaiters.delete(cdpSessionId);
                for (const resolve of waiters) resolve(contextId);
              }
            }
          }
          void (async () => {
            if (
              method === "Target.attachedToTarget" &&
              attachedSessionId != null &&
              (targetInfo?.type === "page" || targetInfo?.type === "iframe")
            ) {
              await ModCDPServer.handleCommand("Page.enable", {}, attachedSessionId).catch((error) =>
                console.error("[ModCDPServer] Page.enable failed for attached target", error),
              );
              await ModCDPServer.handleCommand(
                "Page.setLifecycleEventsEnabled",
                { enabled: true },
                attachedSessionId,
              ).catch((error) =>
                console.error("[ModCDPServer] Page.setLifecycleEventsEnabled failed for attached target", error),
              );
            }
            await publishEvent(method, payload, cdpSessionId);
          })().catch((error) => console.error("[ModCDPServer] loopback event listener failed", error));
          return;
        }
        const pending = loopbackPending.get(id);
        if (!pending) return;
        loopbackPending.delete(id);
        if (msg.error) pending.reject(new Error(msg.error.message));
        else pending.resolve(msg.result || {});
      });
      ws.addEventListener("error", () => {
        if (loopbackSockets.get(endpoint) === ws) loopbackSockets.delete(endpoint);
        loopbackTargetSessions.clear();
        loopbackSessionTargets.clear();
        loopbackSessionContexts.clear();
        rejectLoopbackPending(new Error(`CDP socket error ${endpoint}`));
      });
      ws.addEventListener("close", (event) => {
        if (loopbackSockets.get(endpoint) === ws) loopbackSockets.delete(endpoint);
        loopbackTargetSessions.clear();
        loopbackSessionTargets.clear();
        loopbackSessionContexts.clear();
        rejectLoopbackPending(
          new Error(
            `CDP socket closed ${endpoint} close.code=${event.code} close.reason=${event.reason || ""} close.wasClean=${
              event.wasClean
            }`,
          ),
        );
      });
      return ws;
    });
    loopbackSocketPromises.set(endpoint, nextSocket);
    return nextSocket;
  }

  async function callLoopbackWS(method: string, params: ProtocolParams = {}, sessionId: string | null = null) {
    if (!ModCDPServer.loopback_cdp_url) throw new Error(`No loopback_cdp_url configured for ${method}.`);
    const ws = await loopbackWS(ModCDPServer.loopback_cdp_url);
    const id = nextLoopbackId++;
    const message: { id: number; method: string; params: ProtocolParams; sessionId?: string } = {
      id,
      method,
      params,
    };
    if (sessionId) message.sessionId = sessionId;
    ws.send(JSON.stringify(message));
    return new Promise<ProtocolResult>((resolve, reject) => {
      const timeout = setTimeout(() => {
        if (!loopbackPending.delete(id)) return;
        reject(new Error(`${method} timed out after ${ModCDPServer.cdp_send_timeout_ms}ms`));
      }, ModCDPServer.cdp_send_timeout_ms);
      loopbackPending.set(id, {
        resolve: (value) => {
          clearTimeout(timeout);
          resolve(value);
        },
        reject: (error) => {
          clearTimeout(timeout);
          reject(error);
        },
      });
    });
  }

  async function initializeLoopbackCDP() {
    if (!ModCDPServer.loopback_cdp_url) return;
    const ws = await loopbackWS(ModCDPServer.loopback_cdp_url);
    if (initializedLoopbackSockets.has(ws)) return;
    await callLoopbackWS("Target.setAutoAttach", targetAutoAttachParams);
    await callLoopbackWS("Target.setDiscoverTargets", { discover: true });
    initializedLoopbackSockets.add(ws);
  }

  function waitForLoopbackExecutionContext(
    sessionId: string,
    timeoutMs = ModCDPServer.loopback_execution_context_timeout_ms,
  ) {
    const existing = loopbackSessionContexts.get(sessionId);
    if (existing != null) return Promise.resolve(existing);
    return new Promise<number>((resolve, reject) => {
      const timeout = setTimeout(() => {
        const waiters = loopbackContextWaiters.get(sessionId);
        waiters?.delete(complete);
        if (waiters?.size === 0) loopbackContextWaiters.delete(sessionId);
        reject(new Error(`Timed out waiting for Runtime.executionContextCreated for session ${sessionId}.`));
      }, timeoutMs);
      const complete = (contextId: number) => {
        clearTimeout(timeout);
        resolve(contextId);
      };
      const waiters = loopbackContextWaiters.get(sessionId);
      if (waiters) waiters.add(complete);
      else loopbackContextWaiters.set(sessionId, new Set([complete]));
    });
  }

  async function ensureOffscreenKeepAlive() {
    const chromeApi = globalScope.chrome;
    const offscreen = chromeApi?.offscreen;
    if (!offscreen || !chromeApi?.runtime?.getURL) return { started: false, reason: "offscreen_unavailable" };

    const offscreenUrl = chromeApi.runtime.getURL(offscreenKeepAlivePath);
    try {
      const existingContexts = chromeApi.runtime.getContexts
        ? await chromeApi.runtime.getContexts({
            contextTypes: ["OFFSCREEN_DOCUMENT"],
            documentUrls: [offscreenUrl],
          })
        : [];
      if (existingContexts.length > 0) return { started: true, existing: true };

      creatingOffscreenKeepAlive ??= offscreen
        .createDocument({
          url: offscreenKeepAlivePath,
          reasons: ["BLOBS"],
          justification: "Keep ModCDP service worker active while CDP clients route commands through it.",
        })
        .finally(() => {
          creatingOffscreenKeepAlive = null;
        });
      await creatingOffscreenKeepAlive;
      return { started: true };
    } catch (error) {
      return { started: false, reason: errorMessage(error) };
    }
  }

  const ModCDPServer = {
    __ModCDPServerVersion: MODCDP_SERVER_VERSION,
    routes: { ...defaultRoutes },
    loopback_cdp_url: null as string | null,
    browser_token: null as string | null,
    native_bridge_attempts: 0,
    native_bridge_last_error: null as string | null,
    native_bridge_connected: false,
    cdp_send_timeout_ms: DEFAULT_CDP_SEND_TIMEOUT_MS,
    loopback_execution_context_timeout_ms: DEFAULT_LOOPBACK_EXECUTION_CONTEXT_TIMEOUT_MS,
    ws_connect_error_settle_timeout_ms: DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS,
    types: null as (typeof import("../types/generated/zod.js"))["types"] | null,
    commands: null as (typeof import("../types/generated/zod.js"))["commands"] | null,
    events: null as (typeof import("../types/generated/zod.js"))["events"] | null,
    startOffscreenKeepAlive,
    startReverseBridge(
      endpoint: string,
      {
        reconnect_interval_ms = DEFAULT_REVERSE_BRIDGE_RECONNECT_INTERVAL_MS,
      }: {
        reconnect_interval_ms?: number;
      } = {},
    ) {
      if (!/^wss?:\/\//i.test(endpoint)) {
        throw new Error(`reverse proxy endpoint must be a ws:// or wss:// URL, got ${endpoint}.`);
      }
      reverseBridgeUrl = endpoint;
      reverseBridgeReconnectIntervalMs = reconnect_interval_ms;
      void connectReverseBridge(endpoint).catch(() => {
        scheduleReverseBridgeReconnect(reverseBridgeReconnectIntervalMs);
      });
      return { upstream_reversews_url: endpoint, reconnect_interval_ms, connecting: true };
    },
    stopReverseBridge,
    startNativeBridge(
      hostName = DEFAULT_NATIVE_BRIDGE_HOST_NAME,
      {
        reconnect_interval_ms = DEFAULT_NATIVE_BRIDGE_RECONNECT_INTERVAL_MS,
      }: {
        reconnect_interval_ms?: number;
      } = {},
    ) {
      nativeBridgeHostName = hostName;
      nativeBridgeReconnectIntervalMs = reconnect_interval_ms;
      return connectNativeBridge(hostName);
    },
    startNatsBridge(
      endpoint: string,
      {
        upstream_nats_subject_prefix = DEFAULT_NATS_BRIDGE_SUBJECT_PREFIX,
        reconnect_interval_ms = DEFAULT_NATS_BRIDGE_RECONNECT_INTERVAL_MS,
      }: {
        upstream_nats_subject_prefix?: string;
        reconnect_interval_ms?: number;
      } = {},
    ) {
      if (!upstream_nats_subject_prefix || /[\s*>]/.test(upstream_nats_subject_prefix))
        throw new Error(`Invalid NATS subject prefix ${upstream_nats_subject_prefix}`);
      nats_bridge_url = endpoint;
      nats_bridge_subject_prefix = upstream_nats_subject_prefix;
      nats_bridge_reconnect_interval_ms = reconnect_interval_ms;
      void connectNatsBridge(endpoint).catch(() => {
        scheduleNatsBridgeReconnect(nats_bridge_reconnect_interval_ms);
      });
      return { upstream_nats_url: endpoint, upstream_nats_subject_prefix, reconnect_interval_ms, connecting: true };
    },
    ensureOffscreenKeepAlive,

    async loadTypes() {
      runtime_types_promise ??= import("../types/generated/zod.js").then((module) => {
        this.types = module.types;
        this.commands = module.commands;
        this.events = module.events;
        return module.types;
      });
      return runtime_types_promise;
    },

    async configure(params: ModCDPConfigureParams = {}) {
      const upstream = params.upstream ?? {};
      const server = params.server ?? {};
      const {
        server_loopback_cdp_url = this.loopback_cdp_url,
        server_routes,
        server_browser_token = this.browser_token,
        server_cdp_send_timeout_ms = this.cdp_send_timeout_ms,
        server_loopback_execution_context_timeout_ms = this.loopback_execution_context_timeout_ms,
        server_ws_connect_error_settle_timeout_ms = this.ws_connect_error_settle_timeout_ms,
      } = server;
      const { custom_commands = [], custom_events = [], custom_middlewares = [] } = params;
      this.loopback_cdp_url = await resolveCDPEndpoint(server_loopback_cdp_url);
      this.browser_token = server_browser_token;
      this.cdp_send_timeout_ms = server_cdp_send_timeout_ms;
      this.loopback_execution_context_timeout_ms = server_loopback_execution_context_timeout_ms;
      this.ws_connect_error_settle_timeout_ms = server_ws_connect_error_settle_timeout_ms;
      if (upstream.upstream_mode === "nats" && upstream.upstream_nats_url) {
        this.startNatsBridge(upstream.upstream_nats_url, {
          upstream_nats_subject_prefix: upstream.upstream_nats_subject_prefix ?? DEFAULT_NATS_BRIDGE_SUBJECT_PREFIX,
        });
      }
      if (upstream.upstream_mode === "reversews" && upstream.upstream_reversews_url) {
        this.startReverseBridge(upstream.upstream_reversews_url);
      }
      if (server_routes) this.routes = { ...defaultRoutes, ...server_routes };
      else {
        this.routes = { ...defaultRoutes };
        await this.discoverLoopbackCDP();
      }
      for (const command of custom_commands) this.addCustomCommand(command as ModCDPCustomCommandRegistration);
      for (const event of custom_events) this.addCustomEvent(event as ModCDPCustomEventRegistration);
      for (const middleware of custom_middlewares) this.addMiddleware(middleware as ModCDPMiddlewareRegistration);
      await initializeLoopbackCDP();
      return { loopback_cdp_url: this.loopback_cdp_url, routes: this.routes };
    },

    addCustomCommand({
      name,
      params_schema = null,
      result_schema = null,
      expression = null,
      handler,
    }: ModCDPCustomCommandRegistration) {
      name = normalizeModCDPName(name);
      if (!/^[^.]+\.[^.]+$/.test(name)) throw new Error("name must be in Domain.method form.");
      if (typeof handler !== "function" && typeof expression === "string") {
        handler = async (params: ProtocolParams = {}, cdpSessionId: string | null = null, method: string = name) => {
          return await evaluateUserExpression({ expression, params, cdpSessionId, method });
        };
      }
      if (typeof handler !== "function") throw new Error(`Custom command ${name} was registered without a handler.`);
      commandHandlers.set(name, { name, handler, params_schema, result_schema, expression });
      return { name, registered: true };
    },

    addCustomEvent({ name, event_schema = null }: ModCDPCustomEventRegistration) {
      name = normalizeModCDPName(name);
      if (!/^[^.]+\.[^.]+$/.test(name)) throw new Error("name must be in Domain.event form.");
      eventBindings.set(name, { name, event_schema });
      return { name, registered: true };
    },

    addEventListener(listener: (event: string, data: ProtocolPayload, cdpSessionId: string | null) => void) {
      eventListeners.add(listener);
      return { remove: () => eventListeners.delete(listener) };
    },

    addMiddleware({ name = "*", phase, expression = null, handler }: ModCDPMiddlewareRegistration) {
      name = normalizeModCDPName(name);
      if (!["request", "response", "event"].includes(phase))
        throw new Error("phase must be request, response, or event.");
      if (name !== "*" && (!name || !name.includes("."))) throw new Error("name must be '*' or Domain.name form.");
      if (typeof handler !== "function" && typeof expression === "string") {
        handler = async (payload: ProtocolPayload, next: unknown, context: ProtocolPayload = {}) => {
          const context_object = context && typeof context === "object" ? (context as Record<string, unknown>) : {};
          const cdpSessionId = typeof context_object.cdpSessionId === "string" ? context_object.cdpSessionId : null;
          const result = (await evaluateInSelf(`
            (async () => {
              const payload = ${JSON.stringify(payload ?? {})};
              const context = ${JSON.stringify(context ?? {})};
              const cdp = globalThis.ModCDP.attachToSession(${JSON.stringify(cdpSessionId)});
              const ModCDP = globalThis.ModCDP;
              const chrome = globalThis.chrome;
              const next = async (nextValue = payload) => ({ __ModCDP_middleware_next__: true, value: nextValue });
              const handler = (${expression});
              return await handler(payload, next, context);
            })()
          `)) as Record<string, unknown>;
          if (result?.__ModCDP_middleware_next__ === true && typeof next === "function") {
            const nextResult = await next(result.value);
            const { __ModCDP_middleware_next__, value, ...overrides } = result;
            if (Object.keys(overrides).length === 0) return nextResult;
            return isPlainObject(nextResult) ? { ...nextResult, ...overrides } : overrides;
          }
          return result;
        };
      }
      if (typeof handler !== "function") {
        throw new Error(`Middleware ${name}:${phase} was registered without a handler.`);
      }
      middlewares[phase].push({ name, phase, expression, handler });
      return { name, phase, registered: true };
    },

    async runMiddleware(phase: MiddlewarePhase, name: string, payload: ProtocolPayload, context: ProtocolPayload = {}) {
      const matching = (middlewares[phase] || []).filter(
        (middleware) => middleware.name === "*" || middleware.name === name,
      );
      const dispatch = async (index: number, value: ProtocolPayload): Promise<ProtocolPayload> => {
        const middleware = matching[index];
        if (!middleware) return value;
        let nextCalled = false;
        const next = async (nextValue = value) => {
          if (nextCalled)
            throw new Error(`Middleware ${middleware.name}:${middleware.phase} called next() more than once.`);
          nextCalled = true;
          return dispatch(index + 1, nextValue);
        };
        const ctx = context && typeof context === "object" ? context : {};
        return middleware.handler(value, next, { ...ctx, name, phase });
      };
      return dispatch(0, payload);
    },

    async handleCommand(method: string, params: ProtocolParams = {}, cdpSessionId: string | null = null) {
      const request = { method, params, cdpSessionId };
      const middlewareParams = await this.runMiddleware("request", method, params, { cdpSessionId, request });
      if (middlewareParams == null) throw new Error(`Request middleware for ${method} returned no params.`);
      params = middlewareParams as ProtocolParams;

      const command = registryMatch(commandHandlers, method);
      let result;
      if (command) {
        result = await command.handler(params, cdpSessionId, method);
        return this.runMiddleware("response", method, result, {
          cdpSessionId,
          request: { ...request, params },
          response: { result },
        });
      }

      let upstream = "chrome_debugger";
      for (const [pattern, route] of Object.entries(this.routes || {}) as [string, string][]) {
        if (pattern === "*.*") {
          upstream = route;
          continue;
        }
        if (pattern.endsWith(".*") && method.startsWith(pattern.slice(0, -1))) {
          upstream = route;
          break;
        }
        if (pattern === method) {
          upstream = route;
          break;
        }
      }

      if (upstream === "auto") {
        if (this.loopback_cdp_url) {
          try {
            result = await this.sendLoopback(method, params, cdpSessionId);
          } catch {
            result = await this.sendChromeDebugger(method, params);
          }
        } else {
          result = await this.sendChromeDebugger(method, params);
        }
      } else if (upstream === "loopback_cdp") result = await this.sendLoopback(method, params, cdpSessionId);
      else if (upstream === "chrome_debugger") result = await this.sendChromeDebugger(method, params);
      else throw new Error(`No ModCDP command registered for ${method}.`);

      return this.runMiddleware("response", method, result, {
        cdpSessionId,
        request: { ...request, params },
        response: { result },
      });
    },

    attachToSession(cdpSessionId: string | null = null) {
      return {
        sessionId: cdpSessionId,
        get types() {
          return ModCDPServer.types;
        },
        get commands() {
          return ModCDPServer.commands;
        },
        get events() {
          return ModCDPServer.events;
        },
        send: (method: string, params: ProtocolParams = {}) => this.handleCommand(method, params, cdpSessionId),
        emit: (eventName: string, payload: ProtocolPayload = {}) => this.emit(eventName, payload, cdpSessionId),
      };
    },

    async emit(eventName: string, payload: ProtocolPayload = {}, cdpSessionId: string | null = null) {
      const event = registryMatch(eventBindings, eventName);
      if (!event) return { event: eventName, emitted: false, reason: "event_not_registered" };
      const customBinding = globalScope[CUSTOM_EVENT_BINDING_NAME];
      if (
        typeof customBinding !== "function" &&
        reverseBridgeSocket?.readyState !== WebSocket.OPEN &&
        !nativeBridgePort &&
        nats_bridge_socket?.readyState !== WebSocket.OPEN
      )
        return { event: eventName, emitted: false, reason: "binding_not_installed" };
      return publishEvent(eventName, payload, cdpSessionId);
    },

    async discoverLoopbackCDP(): Promise<{ loopback_cdp_url: string | null; verified: boolean; version?: unknown }> {
      if (!this.browser_token) return { loopback_cdp_url: null as null, verified: false };

      const url = "http://127.0.0.1:9222";
      const previousLoopbackUrl = this.loopback_cdp_url;
      const fail = (version?: unknown) => {
        this.loopback_cdp_url = previousLoopbackUrl ?? null;
        return { loopback_cdp_url: null as null, verified: false, ...(version ? { version } : {}) };
      };
      try {
        const version = await fetch(`${url}/json/version`).then((response) => response.ok && response.json());
        if (!version?.webSocketDebuggerUrl) return fail();

        this.loopback_cdp_url = version.webSocketDebuggerUrl;
        const { targetInfos } = (await callLoopbackWS("Target.getTargets")) as cdp.types.ts.Target.GetTargetsResult;
        const serviceWorkerUrl = currentServiceWorkerUrl();
        const worker = targetInfos.find(
          (target) => target.type === "service_worker" && target.url === serviceWorkerUrl,
        );
        if (!worker) return fail(version);

        const { sessionId } = (await callLoopbackWS("Target.attachToTarget", {
          targetId: worker.targetId,
          flatten: true,
        })) as cdp.types.ts.Target.AttachToTargetResult;
        loopbackTargetSessions.set(worker.targetId, sessionId);
        loopbackSessionTargets.set(sessionId, worker.targetId);
        const contextIdPromise = waitForLoopbackExecutionContext(sessionId);
        await callLoopbackWS("Runtime.enable", {}, sessionId);
        const executionContextId = await contextIdPromise;
        const result = (await callLoopbackWS(
          "Runtime.callFunctionOn",
          {
            functionDeclaration: `function() { return globalThis.ModCDP?.browser_token === ${JSON.stringify(this.browser_token)}; }`,
            executionContextId,
            returnByValue: true,
          },
          sessionId,
        )) as cdp.types.ts.Runtime.EvaluateResult;
        if (result.result?.value !== true) return fail(version);

        await initializeLoopbackCDP();
        return { loopback_cdp_url: this.loopback_cdp_url, verified: true, version };
      } catch {
        return fail();
      }
    },

    async sendLoopback(method: string, params: ProtocolParams = {}, cdpSessionId: string | null = null) {
      if (!this.loopback_cdp_url) throw new Error(`No loopback_cdp_url configured for ${method}.`);

      await initializeLoopbackCDP();

      const domain = method.split(".")[0] ?? "";
      if (browserLevelDomains.has(domain)) return await callLoopbackWS(method, params);
      if (cdpSessionId) return await callLoopbackWS(method, params, cdpSessionId);

      const {
        debuggee = null,
        tabId = null,
        targetId = null,
        extensionId = null,
        ...commandParams
      } = params as CdpDebuggeeCommandParams;
      const resolvedDebuggee = debuggee ?? compactDebuggee({ tabId, targetId, extensionId });

      const chromeApi = globalScope.chrome;
      let resolvedTargetId = resolvedDebuggee.targetId || null;
      if (!resolvedTargetId) {
        let resolvedTabId = resolvedDebuggee.tabId || null;
        let resolvedTabUrl: string | null = null;
        if (!resolvedTabId) {
          const [tab] = chromeApi.tabs?.query
            ? await chromeApi.tabs.query({ active: true, lastFocusedWindow: true })
            : [];
          resolvedTabId = tab?.id || null;
          resolvedTabUrl = tab?.url || tab?.pendingUrl || null;
        } else if (chromeApi.tabs?.get) {
          const tab = await chromeApi.tabs.get(resolvedTabId).catch((): null => null);
          resolvedTabUrl = tab?.url || tab?.pendingUrl || null;
        }
        if (resolvedTabId && chromeApi.debugger?.getTargets) {
          const targets = await chromeApi.debugger.getTargets();
          resolvedTargetId =
            targets.find((target) => target.tabId === resolvedTabId && target.type === "page")?.id || null;
        }
        if (!resolvedTargetId) {
          const { targetInfos } = (await callLoopbackWS("Target.getTargets")) as cdp.types.ts.Target.GetTargetsResult;
          const pageTargets = targetInfos.filter((target) => target.type === "page");
          resolvedTargetId =
            pageTargets.find((target) => resolvedTabUrl && target.url === resolvedTabUrl)?.targetId ||
            pageTargets[0]?.targetId ||
            null;
        }
        if (!resolvedTargetId) {
          const created = (await callLoopbackWS("Target.createTarget", {
            url: "about:blank#modcdp",
          })) as cdp.types.ts.Target.CreateTargetResult;
          resolvedTargetId = created.targetId || null;
        }
      }
      if (!resolvedTargetId) throw new Error(`loopback_cdp route for ${method} could not resolve a page target.`);

      const existingSessionId = loopbackTargetSessions.get(resolvedTargetId);
      if (existingSessionId) return await callLoopbackWS(method, commandParams, existingSessionId);

      const attached = (await callLoopbackWS("Target.attachToTarget", {
        targetId: resolvedTargetId,
        flatten: true,
      })) as cdp.types.ts.Target.AttachToTargetResult;
      const sessionId = attached.sessionId;
      loopbackTargetSessions.set(resolvedTargetId, sessionId);
      loopbackSessionTargets.set(sessionId, resolvedTargetId);
      await callLoopbackWS("Target.setAutoAttach", targetAutoAttachParams, sessionId).catch(() => {});
      return await callLoopbackWS(method, commandParams, sessionId);
    },

    async sendChromeDebugger(method: string, params: ProtocolParams = {}) {
      const chromeApi = globalScope.chrome;
      if (!chromeApi?.debugger?.sendCommand) throw new Error("chrome.debugger is unavailable.");

      const {
        debuggee = null,
        tabId = null,
        targetId = null,
        extensionId = null,
        ...commandParams
      } = params as CdpDebuggeeCommandParams;
      const resolvedDebuggee = debuggee ?? compactDebuggee({ tabId, targetId, extensionId });
      if (Object.keys(resolvedDebuggee).length === 0) {
        let tab: chrome.tabs.Tab | undefined;
        [tab] = await chromeApi.tabs.query({ active: true, lastFocusedWindow: true });
        if (!tab?.id) [tab] = await chromeApi.tabs.query({});
        if (!tab?.id) {
          try {
            tab = await chromeApi.tabs.create({ url: "https://example.com/#modcdp", active: true });
          } catch {
            const win = await chromeApi.windows.create({ url: "https://example.com/#modcdp", focused: true });
            tab = win?.tabs?.[0];
          }
        }
        if (!tab?.id) throw new Error(`chrome_debugger route for ${method} could not find an active tab.`);
        resolvedDebuggee.tabId = tab.id;
      }

      const key = JSON.stringify(resolvedDebuggee);
      if (!attachedDebuggees.has(key)) {
        try {
          await new Promise<void>((resolve, reject) =>
            chromeApi.debugger.attach(resolvedDebuggee, "1.3", () => {
              const error = chromeApi.runtime.lastError;
              if (error) reject(new Error(error.message));
              else resolve();
            }),
          );
        } catch (error) {
          if (!errorMessage(error).includes("Another debugger is already attached")) throw error;
        }
        await new Promise<void>((resolve, reject) =>
          chromeApi.debugger.sendCommand(resolvedDebuggee, "Target.setAutoAttach", targetAutoAttachParams, () => {
            const error = chromeApi.runtime.lastError;
            if (error) reject(new Error(error.message));
            else resolve();
          }),
        );
        attachedDebuggees.add(key);
      }

      return new Promise<ProtocolResult>((resolve, reject) =>
        chromeApi.debugger.sendCommand(resolvedDebuggee, method, commandParams, (result) => {
          const error = chromeApi.runtime.lastError;
          if (error) reject(new Error(error.message));
          else resolve(result as ProtocolResult);
        }),
      );
    },
  };

  globalScope.ModCDP = ModCDPServer;

  ModCDPServer.addCustomEvent({
    name: "Mod.pong",
  });

  ModCDPServer.addCustomCommand({
    name: "Mod.ping",
    handler: async (raw_params: ProtocolParams = {}, cdpSessionId: string | null = null) => {
      const params = raw_params as ModCDPPingParams;
      const received_at = Date.now();
      await ModCDPServer.emit(
        "Mod.pong",
        {
          sent_at: typeof params.sent_at === "number" ? params.sent_at : received_at,
          received_at,
          from: "extension-service-worker",
        },
        cdpSessionId,
      );
      return { ok: true };
    },
  });

  ModCDPServer.addCustomCommand({
    name: "Mod.configure",
    handler: async (params: ProtocolParams = {}) => ModCDPServer.configure(params as ModCDPConfigureParams),
  });

  ModCDPServer.addCustomCommand({
    name: "Mod.evaluate",
    handler: async (raw_params: ProtocolParams = {}) => {
      const { expression, params = {}, cdpSessionId = null } = raw_params as Record<string, unknown>;
      return await evaluateUserExpression({
        expression: String(expression),
        params: params as ProtocolPayload,
        cdpSessionId: typeof cdpSessionId === "string" ? cdpSessionId : null,
      });
    },
  });

  ModCDPServer.addCustomCommand({
    name: "Mod.addCustomCommand",
    handler: async (params: ProtocolParams = {}) =>
      ModCDPServer.addCustomCommand(params as ModCDPCustomCommandRegistration),
  });

  ModCDPServer.addCustomCommand({
    name: "Mod.addCustomEvent",
    handler: async (params: ProtocolParams = {}) =>
      ModCDPServer.addCustomEvent(params as ModCDPCustomEventRegistration),
  });

  ModCDPServer.addCustomCommand({
    name: "Mod.addMiddleware",
    handler: async (params: ProtocolParams = {}) => ModCDPServer.addMiddleware(params as ModCDPMiddlewareRegistration),
  });

  const chromeApi = globalScope.chrome;
  try {
    chromeApi?.runtime?.onStartup?.addListener(startOffscreenKeepAlive);
  } catch {}
  try {
    chromeApi?.runtime?.onInstalled?.addListener(startOffscreenKeepAlive);
  } catch {}
  try {
    chromeApi?.tabs?.onCreated?.addListener(startOffscreenKeepAlive);
  } catch {}
  try {
    chromeApi?.runtime?.onConnect?.addListener((port) => {
      if (port.name !== offscreenKeepAlivePortName) return;
      offscreenKeepAlivePort = port;
      port.onMessage.addListener(() => {});
      port.onDisconnect.addListener(() => {
        if (offscreenKeepAlivePort === port) offscreenKeepAlivePort = null;
      });
    });
  } catch {}
  startOffscreenKeepAlive();

  return ModCDPServer;
}

export const ModCDPServer = installModCDPServer(globalThis);
