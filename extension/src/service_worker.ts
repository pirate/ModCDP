// Extension service worker entry point.

import { ModCDPServer } from "../../js/src/server/ModCDPServer.js";

const bridge = ModCDPServer as Record<string, any>;
const started_at = new Date().toISOString();
const DEFAULT_REVERSEWS_URL = "ws://127.0.0.1:29292";
const DEFAULT_REVERSEWS_RECONNECT_INTERVAL_MS = 2_000;
const DEFAULT_NATIVE_HOST_NAME = "com.modcdp.bridge";
const DEFAULT_NATIVE_RECONNECT_INTERVAL_MS = 2_000;
const downstream_clients: Record<string, any> = {};
const upstream_servers: Record<string, any> = {};
const client_id_by_config_session = new Map<string, string>();
let active_downstream_client_id: string | null = null;
let next_downstream_client_id = 1;
let next_log_id = 1;
const self_transports: Record<string, any> = {};
const self_custom = { commands: new Set<string>(), events: new Set<string>() };
const self_log: any[] = [];
const compact = (value: unknown) => {
  try {
    return JSON.parse(JSON.stringify(value ?? null));
  } catch (error) {
    return {
      unserializable: true,
      error: error instanceof Error ? error.message : String(error),
    };
  }
};
const trimLog = (log: any[]) => (log.length = Math.min(log.length, 80));
const routeFor = (method: string) => {
  if (method.startsWith("Mod.") || method.startsWith("Custom.")) return "service_worker";
  const routes = (bridge.routes ?? {}) as Record<string, string>;
  const route =
    routes[method] ??
    Object.entries(routes)
      .filter(([pattern]) => pattern.endsWith(".*") && method.startsWith(pattern.slice(0, -1)))
      .sort((a, b) => b[0].length - a[0].length)[0]?.[1] ??
    routes["*.*"] ??
    "chrome_debugger";
  if (route === "loopback_cdp") return "loopback";
  if (route === "chrome_debugger") return "debugger";
  if (route === "auto") return bridge.loopback_cdp_url ? "loopback" : "debugger";
  return route;
};
const upstreamServer = (id: string) =>
  (upstream_servers[id] ??= {
    id,
    log: [],
  });
const configuredClient = (params: unknown, session_id?: string | null) => {
  const at = new Date().toISOString();
  const id =
    (session_id && client_id_by_config_session.get(session_id)) || `downstream_client_${next_downstream_client_id++}`;
  if (session_id) client_id_by_config_session.set(session_id, id);
  active_downstream_client_id = id;
  const configure = compact(params);
  const client = (downstream_clients[id] ??= {
    id,
    configured_at: at,
    commands: 0,
    events: 0,
    sessions: {},
    recent: [],
  });
  client.updated_at = at;
  client.configure = configure;
  client.downstream_transport = configure?.upstream?.upstream_mode ?? "unknown";
  client.route_config = {
    upstream: configure?.upstream ?? {},
    client: configure?.client ?? {},
    server: configure?.server ?? {},
  };
  if (client.downstream_transport !== "reversews") {
    bridge.stopReverseBridge?.("non-reverse downstream connected");
  }
  return client;
};
const downstreamClient = (session_id?: string | null) => {
  const at = new Date().toISOString();
  const client_id =
    (session_id && client_id_by_config_session.get(session_id)) ||
    active_downstream_client_id ||
    "unconfigured_downstream_client";
  const client = (downstream_clients[client_id] ??= {
    id: client_id,
    commands: 0,
    events: 0,
    sessions: {},
    recent: [],
    first_seen: at,
    last_seen: at,
  });
  const id = session_id || "root";
  const session = (client.sessions[id] ??= {
    id,
    commands: 0,
    events: 0,
    first_seen: at,
    last_seen: at,
  });
  return { at, client_id, client, session };
};
const logTraffic = (direction: "command" | "event", name: string, payload: unknown, session_id?: string | null) => {
  const { at, client_id, client, session } = downstreamClient(session_id);
  const upstream = routeFor(name);
  const from = direction === "command" ? client_id : upstream;
  const to = direction === "command" ? upstream : client_id;
  const route_path = from === "service_worker" || to === "service_worker" ? [from, to] : [from, "service_worker", to];
  const entry: any = {
    id: `log_${next_log_id++}`,
    at,
    direction,
    method: name,
    payload: compact(payload),
    route_path,
    downstream_transport: client.downstream_transport ?? "unknown",
    cdp_session_id: session_id ?? null,
  };
  direction === "event" ? client.events++ : client.commands++;
  direction === "event" ? session.events++ : session.commands++;
  direction === "event" ? (session.last_event = name) : (session.last_command = name);
  client.last_seen = at;
  session.last_seen = at;
  client.log ??= client.recent ?? [];
  client.log.unshift(entry);
  trimLog(client.log);
  const endpointLog = upstream === "service_worker" ? self_log : upstreamServer(upstream).log;
  endpointLog.unshift(entry);
  trimLog(endpointLog);
  return entry;
};

if (bridge) {
  const handleCommand = bridge.handleCommand?.bind(bridge);
  if (handleCommand) {
    bridge.handleCommand = async (method: string, params?: unknown, session_id?: string | null) => {
      if (method === "Mod.configure") configuredClient(params, session_id);
      const entry = logTraffic("command", method, params, session_id);
      try {
        const result = await handleCommand(method, params, session_id);
        entry.result = compact(result);
        entry.completed_at = new Date().toISOString();
        return result;
      } catch (error) {
        entry.error = error instanceof Error ? error.message : String(error);
        entry.completed_at = new Date().toISOString();
        throw error;
      }
    };
  }
  bridge.addEventListener?.((event: string, _payload: unknown, session_id?: string | null) =>
    logTraffic("event", event, _payload, session_id),
  );
  for (const [method, key] of [
    ["startReverseBridge", "reverse"],
    ["stopReverseBridge", "reverse"],
    ["startNativeBridge", "native"],
    ["startNatsBridge", "nats"],
  ]) {
    const start = bridge[method]?.bind(bridge);
    if (start) {
      bridge[method] = (...args: unknown[]) => {
        const result = start(...args);
        self_transports[key] = {
          args: compact(args),
          result: compact(result),
          updated_at: new Date().toISOString(),
        };
        return result;
      };
    }
  }
  for (const [method, key] of [
    ["addCustomCommand", "commands"],
    ["addCustomEvent", "events"],
  ]) {
    const add = bridge[method]?.bind(bridge);
    if (add) {
      bridge[method] = (name: string, ...args: unknown[]) => {
        self_custom[key as "commands" | "events"].add(name);
        return add(name, ...args);
      };
    }
  }
}

const startConfiguredTransports = () => {
  bridge.startOffscreenKeepAlive?.();
  bridge.startReverseBridge?.(DEFAULT_REVERSEWS_URL, {
    reconnect_interval_ms: DEFAULT_REVERSEWS_RECONNECT_INTERVAL_MS,
  });
  bridge.startNativeBridge?.(DEFAULT_NATIVE_HOST_NAME, {
    reconnect_interval_ms: DEFAULT_NATIVE_RECONNECT_INTERVAL_MS,
  });
};

startConfiguredTransports();
chrome.runtime.onInstalled.addListener(startConfiguredTransports);
chrome.runtime.onStartup.addListener(startConfiguredTransports);

chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  if (message?.type !== "modcdp.options.status") return false;
  const self = {
    id: "self",
    runtime: {
      extension_id: chrome.runtime.id,
      service_worker_url: chrome.runtime.getURL("modcdp/service_worker.js"),
      options_url: chrome.runtime.getURL("options.html"),
      started_at,
    },
    server: {
      __ModCDPServerVersion: bridge.__ModCDPServerVersion,
      routes: bridge.routes,
      loopback_cdp_url: bridge.loopback_cdp_url,
      browser_token: bridge.browser_token ? "set" : null,
      cdp_send_timeout_ms: bridge.cdp_send_timeout_ms,
      loopback_execution_context_timeout_ms: bridge.loopback_execution_context_timeout_ms,
      ws_connect_error_settle_timeout_ms: bridge.ws_connect_error_settle_timeout_ms,
      native_bridge_attempts: bridge.native_bridge_attempts,
      native_bridge_connected: bridge.native_bridge_connected,
      native_bridge_last_error: bridge.native_bridge_last_error,
    },
    ...(Object.keys(self_transports).length ? { transports: self_transports } : {}),
    custom: {
      commands: [...self_custom.commands],
      events: [...self_custom.events],
    },
    log: self_log,
  };
  sendResponse({
    now: new Date().toISOString(),
    self,
    downstream_clients,
    upstream_servers: {
      ...upstream_servers,
      ...(bridge.loopback_cdp_url
        ? {
            loopback: {
              ...upstream_servers.loopback,
              id: "loopback",
              url: bridge.loopback_cdp_url,
              log: upstream_servers.loopback?.log ?? [],
            },
          }
        : {}),
      debugger: {
        ...upstream_servers.debugger,
        id: "debugger",
        log: upstream_servers.debugger?.log ?? [],
      },
    },
  });
  return false;
});

chrome.action?.onClicked.addListener(() => {
  void chrome.runtime.openOptionsPage();
});
