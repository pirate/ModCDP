// ModCDPClient (JS): importable, no CLI, no demo code.
//
// Constructor option groups mirror the owning runtime components:
//   launch            browser/session creation and cleanup
//   upstream          message transport to either raw CDP or a ModCDP server
//   extension         raw-CDP extension discovery/injection/borrowing
//   client            client-side routing, alias hydration, event mirroring, send/event timeouts
//   server            ModCDPServer.configure params
//
// Public methods: connect, send(method, params), on(event, handler), close.

// oxlint-disable typescript-eslint/no-unsafe-declaration-merging -- alias members are assigned by connect().
import type { z } from "zod";

import { createCdpAliases, type CdpAliases } from "../../types/aliases.js";
export type { CdpAliases } from "../../types/aliases.js";
import {
  CUSTOM_EVENT_BINDING_NAME,
  DEFAULT_CLIENT_ROUTES,
  UPSTREAM_EVENT_BINDING_NAME,
  wrapCommandIfNeeded,
  unwrapResponseIfNeeded,
  unwrapEventIfNeeded,
} from "../../bridge/translate.js";
import {
  endpointKindForUpstream,
  type UpstreamEndpointKind,
  type UpstreamMode,
  type UpstreamTransport,
} from "../../bridge/UpstreamTransport.js";
import {
  DEFAULT_REVERSEWS_BIND,
  DEFAULT_REVERSEWS_WAIT_TIMEOUT_MS,
  ReverseWebSocketUpstreamTransport,
} from "../../bridge/ReverseWebSocketUpstreamTransport.js";
import { WebSocketUpstreamTransport } from "../../bridge/WebSocketUpstreamTransport.js";
import { NativeMessagingUpstreamTransport } from "../../bridge/NativeMessagingUpstreamTransport.js";
import { PipeUpstreamTransport } from "../../bridge/PipeUpstreamTransport.js";
import { NatsUpstreamTransport } from "../../bridge/NatsUpstreamTransport.js";
import { LocalBrowserLauncher } from "../../bridge/LocalBrowserLauncher.js";
import { RemoteBrowserLauncher } from "../../bridge/RemoteBrowserLauncher.js";
import { BrowserbaseBrowserLauncher } from "../../bridge/BrowserbaseBrowserLauncher.js";
import { NoopBrowserLauncher } from "../../bridge/NoopBrowserLauncher.js";
import type { BrowserLauncher, BrowserLaunchOptions, LaunchedBrowser } from "../../bridge/BrowserLauncher.js";
import { BBBrowserExtensionInjector } from "../../bridge/BBBrowserExtensionInjector.js";
import { BorrowedExtensionInjector } from "../../bridge/BorrowedExtensionInjector.js";
import { DiscoveredExtensionInjector } from "../../bridge/DiscoveredExtensionInjector.js";
import {
  DEFAULT_MODCDP_SERVICE_WORKER_URL_SUFFIXES,
  DEFAULT_MODCDP_WAKE_PATH,
  ExtensionInjector,
  type ExtensionInjectorConfig,
  type SendCDP,
} from "../../bridge/ExtensionInjector.js";
import { ExtensionsLoadUnpackedInjector } from "../../bridge/ExtensionsLoadUnpackedInjector.js";
import { LocalBrowserLaunchExtensionInjector } from "../../bridge/LocalBrowserLaunchExtensionInjector.js";
import { AutoSessionRouter } from "./AutoSessionRouter.js";
import type {
  CdpCommandMessage,
  CdpError,
  CdpEventMessage,
  CdpResponseMessage,
  RuntimeBindingCalledEvent,
  ModCDPConfigureParams,
  ModCDPServerOptions,
  ModCDPCustomPayload,
  ModCDPAddCustomCommandParams,
  ModCDPAddCustomEventObjectParams,
  ModCDPAddMiddlewareParams,
  ModCDPNamedValue,
  ModCDPPingLatency,
  ModCDPPongEvent,
  ModCDPRoutes,
  ProtocolPayload,
  ProtocolParams,
  ProtocolResult,
  TranslatedCommand,
} from "../../types/modcdp.js";
import {
  CdpEventMessageSchema,
  CdpResponseMessageSchema,
  Mod,
  normalizeModCDPName,
  normalizeModCDPPayloadSchema,
} from "../../types/modcdp.js";

export const DEFAULT_CDP_SEND_TIMEOUT_MS = 10_000;
export const DEFAULT_EVENT_WAIT_TIMEOUT_MS = 10_000;
export const DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS = 10_000;
export const DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS = 10_000;
export const DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS = 60_000;
export const DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS = 100;
export const DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS = 20;
export const DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS = 250;

type PendingCommand = {
  method: string;
  resolve: (value: ProtocolResult) => void;
  reject: (error: Error) => void;
};
type LaunchMode = "local" | "remote" | "bb" | "none";
type ExtensionMode = "auto" | "discover" | "inject" | "borrow" | "none";
type LaunchOptions = {
  mode?: LaunchMode;
  executable_path?: string | null;
  user_data_dir?: string | null;
  options?: Record<string, unknown>;
};
type UpstreamOptions = {
  mode?: UpstreamMode;
  ws_url?: string | null;
  nats_url?: string | null;
  nats_subject_prefix?: string | null;
  reversews_bind?: string | null;
  reversews_wait_timeout_ms?: number;
  nativemessaging_manifest?: string | null;
  ws_connect_error_settle_timeout_ms?: number;
};
type ExtensionOptions = {
  mode?: ExtensionMode;
  path?: string | null;
  extension_id?: string | null;
  wake_path?: string | null;
  wake_url?: string | null;
  service_worker_url_includes?: string[];
  service_worker_url_suffixes?: string[] | null;
  trust_service_worker_target?: boolean;
  require_service_worker_target?: boolean;
  service_worker_ready_expression?: string | null;
  execution_context_timeout_ms?: number;
  service_worker_probe_timeout_ms?: number;
  service_worker_ready_timeout_ms?: number;
  service_worker_poll_interval_ms?: number;
  target_session_poll_interval_ms?: number;
};
type ClientConfigOptions = {
  routes?: ModCDPRoutes;
  hydrate_aliases?: boolean;
  mirror_upstream_events?: boolean;
  cdp_send_timeout_ms?: number;
  event_wait_timeout_ms?: number;
};
type ClientOptions = {
  launch?: LaunchOptions;
  upstream?: UpstreamOptions;
  extension?: ExtensionOptions;
  client?: ClientConfigOptions;
  server?: ModCDPServerOptions | null;
  custom_commands?: ModCDPClientCustomCommandParams[];
  custom_events?: ModCDPAddCustomEventObjectParams[];
  custom_middlewares?: ModCDPAddMiddlewareParams[];
};
type ClientConfig = {
  routes: ModCDPRoutes;
  hydrate_aliases: boolean;
  mirror_upstream_events: boolean;
  cdp_send_timeout_ms: number;
  event_wait_timeout_ms: number;
};
type NormalizedClientOptions = {
  launch: Required<LaunchOptions>;
  upstream: Required<UpstreamOptions>;
  extension: Required<ExtensionOptions>;
  client: ClientConfig;
  server: ModCDPServerOptions | null;
  upstream_endpoint_kind: UpstreamEndpointKind;
};
type ModCDPEventNameInput = string | symbol | (z.ZodType & ModCDPNamedValue);
type ModCDPEventPayload<TEvent extends z.ZodType> = TEvent extends z.ZodType<infer TPayload> ? TPayload : never;
type ModCDPClientCustomCommandParams = Omit<ModCDPAddCustomCommandParams, "expression"> & {
  expression?: string | null;
};

export type ModCDPCommandSpec<Params = unknown, Result = unknown> = {
  params: Params;
  result: Result;
};
export type ModCDPCommandMap = Record<string, ModCDPCommandSpec>;
export type ModCDPEventMap = Record<string, unknown>;
type MethodName<TName extends string> = TName extends `${string}.${infer TMethod}` ? TMethod : never;
type DomainName<TName extends string> = TName extends `${infer TDomain}.${string}` ? TDomain : never;
type CommandsForDomain<TCommands extends ModCDPCommandMap, TDomain extends string> = {
  [TName in keyof TCommands as TName extends `${TDomain}.${string}`
    ? MethodName<Extract<TName, string>>
    : never]: undefined extends TCommands[TName]["params"]
    ? (params?: TCommands[TName]["params"]) => Promise<TCommands[TName]["result"]>
    : (params: TCommands[TName]["params"]) => Promise<TCommands[TName]["result"]>;
};
export type ModCDPClientInstance<
  TCommands extends ModCDPCommandMap = Record<never, never>,
  TEvents extends ModCDPEventMap = Record<never, never>,
> = ModCDPClient & {
  [TDomain in DomainName<Extract<keyof TCommands, string>>]: CommandsForDomain<TCommands, TDomain>;
} & {
  on<TName extends Extract<keyof TEvents, string>>(
    eventName: TName,
    listener: (event: TEvents[TName]) => void,
  ): ModCDPClient;
  once<TName extends Extract<keyof TEvents, string>>(
    eventName: TName,
    listener: (event: TEvents[TName]) => void,
  ): ModCDPClient;
};

class ModCDPEventEmitter {
  private listeners = new Map<string | symbol, Set<(...args: unknown[]) => void>>();

  on(event_name: string | symbol, listener: (...args: unknown[]) => void) {
    const listeners = this.listeners.get(event_name);
    if (listeners) listeners.add(listener);
    else this.listeners.set(event_name, new Set([listener]));
    return this;
  }

  once(event_name: string | symbol, listener: (...args: unknown[]) => void) {
    const wrapped = (...args: unknown[]) => {
      this.listeners.get(event_name)?.delete(wrapped);
      listener(...args);
    };
    return this.on(event_name, wrapped);
  }

  off(event_name: string | symbol, listener: (...args: unknown[]) => void) {
    this.listeners.get(event_name)?.delete(listener);
    return this;
  }

  emit(event_name: string | symbol, ...args: unknown[]) {
    for (const listener of this.listeners.get(event_name) ?? []) listener(...args);
    if (event_name !== "*") {
      for (const listener of this.listeners.get("*") ?? []) listener(event_name, ...args);
    }
    return true;
  }
}

function normalizeClientOptions({
  launch = {},
  upstream = {},
  extension = {},
  client = {},
  server = {},
}: ClientOptions) {
  const upstream_mode = upstream.mode ?? "ws";
  const upstream_endpoint_kind = endpointKindForUpstream(upstream_mode);
  const launch_mode =
    launch.mode ?? (upstream_endpoint_kind === "modcdp_server" ? "none" : upstream.ws_url ? "remote" : "local");
  const extension_mode = extension.mode ?? (upstream_endpoint_kind === "raw_cdp" ? "auto" : "none");
  return {
    launch: {
      mode: launch_mode,
      executable_path: launch.executable_path ?? null,
      user_data_dir: launch.user_data_dir ?? null,
      options: launch.options ?? {},
    },
    upstream: {
      mode: upstream_mode,
      ws_url: upstream.ws_url ?? null,
      nats_url: upstream.nats_url ?? null,
      nats_subject_prefix: upstream.nats_subject_prefix ?? null,
      reversews_bind: upstream.reversews_bind ?? DEFAULT_REVERSEWS_BIND,
      reversews_wait_timeout_ms: upstream.reversews_wait_timeout_ms ?? DEFAULT_REVERSEWS_WAIT_TIMEOUT_MS,
      nativemessaging_manifest: upstream.nativemessaging_manifest ?? null,
      ws_connect_error_settle_timeout_ms:
        upstream.ws_connect_error_settle_timeout_ms ?? DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS,
    },
    extension: {
      mode: extension_mode,
      path: extension.path ?? defaultExtensionPath(),
      extension_id: extension.extension_id ?? null,
      wake_path: extension.wake_path ?? DEFAULT_MODCDP_WAKE_PATH,
      wake_url: extension.wake_url ?? null,
      service_worker_url_includes: extension.service_worker_url_includes ?? [],
      service_worker_url_suffixes: extension.service_worker_url_suffixes ?? null,
      trust_service_worker_target: extension.trust_service_worker_target ?? false,
      require_service_worker_target: extension.require_service_worker_target ?? false,
      service_worker_ready_expression: extension.service_worker_ready_expression ?? null,
      execution_context_timeout_ms: extension.execution_context_timeout_ms ?? DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS,
      service_worker_probe_timeout_ms:
        extension.service_worker_probe_timeout_ms ?? DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
      service_worker_ready_timeout_ms:
        extension.service_worker_ready_timeout_ms ?? DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS,
      service_worker_poll_interval_ms:
        extension.service_worker_poll_interval_ms ?? DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS,
      target_session_poll_interval_ms:
        extension.target_session_poll_interval_ms ?? DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS,
    },
    client: {
      routes: { ...DEFAULT_CLIENT_ROUTES, ...(client.routes ?? {}) },
      hydrate_aliases: client.hydrate_aliases ?? true,
      mirror_upstream_events: client.mirror_upstream_events ?? true,
      cdp_send_timeout_ms: client.cdp_send_timeout_ms ?? DEFAULT_CDP_SEND_TIMEOUT_MS,
      event_wait_timeout_ms: client.event_wait_timeout_ms ?? DEFAULT_EVENT_WAIT_TIMEOUT_MS,
    },
    server:
      server === null
        ? null
        : {
            ...(upstream_endpoint_kind === "modcdp_server" ? { routes: { "*.*": "chrome_debugger" } } : {}),
            ...(server ?? {}),
          },
    upstream_endpoint_kind,
  } satisfies NormalizedClientOptions;
}

function defineCustomCommandMethod(client: ModCDPClient, name: string) {
  const parts = name.split(".");
  if (parts.length !== 2 || !parts[0] || !parts[1]) {
    throw new Error(`Custom command must use Domain.method format, got ${name}`);
  }
  const [domain, method] = parts;
  const target = client as unknown as Record<string, Record<string, unknown>>;
  if (method === "*") {
    target[domain] = new Proxy(target[domain] ?? {}, {
      get(existing, property, receiver) {
        if (typeof property !== "string") return Reflect.get(existing, property, receiver);
        if (property in existing) return Reflect.get(existing, property, receiver);
        const command_name = `${domain}.${property}`;
        const alias = (params?: unknown) => client.send(command_name, params ?? {});
        Object.defineProperties(alias, {
          cdp_command_name: { value: command_name, enumerable: true, configurable: true },
          id: { value: command_name, enumerable: true, configurable: true },
          name: { value: command_name, configurable: true },
          kind: { value: "command", enumerable: true, configurable: true },
          meta: {
            value: () => ({
              cdp_command_name: command_name,
              id: command_name,
              name: command_name,
              kind: "command",
            }),
            configurable: true,
          },
        });
        existing[property] = alias;
        return alias;
      },
    });
    return;
  }
  target[domain] ??= {};
  const alias = (params?: unknown) => client.send(name, params ?? {});
  Object.defineProperties(alias, {
    cdp_command_name: { value: name, enumerable: true, configurable: true },
    id: { value: name, enumerable: true, configurable: true },
    name: { value: name, configurable: true },
    kind: { value: "command", enumerable: true, configurable: true },
    meta: { value: () => ({ cdp_command_name: name, id: name, name, kind: "command" }), configurable: true },
  });
  target[domain][method] = alias;
}

function defaultExtensionPath() {
  if (typeof process === "object" && process?.versions?.node && import.meta.url.startsWith("file:")) {
    return decodeURIComponent(new URL(/* @vite-ignore */ "../../extension", import.meta.url).pathname);
  }
  return "../../extension";
}

function hasCommandExpression(
  command: ModCDPClientCustomCommandParams,
): command is ModCDPClientCustomCommandParams & { expression: string } {
  return typeof command.expression === "string" && command.expression.length > 0;
}

export class ModCDPClient extends ModCDPEventEmitter {
  launch: NormalizedClientOptions["launch"];
  upstream: NormalizedClientOptions["upstream"];
  extension: NormalizedClientOptions["extension"];
  client: NormalizedClientOptions["client"];
  upstream_endpoint_kind: UpstreamEndpointKind;
  cdp_url: string | null;
  server: ModCDPServerOptions | null;
  custom_commands: ModCDPClientCustomCommandParams[];
  custom_events: ModCDPAddCustomEventObjectParams[];
  custom_middlewares: ModCDPAddMiddlewareParams[];
  transport: UpstreamTransport | null;
  next_id: number;
  pending: Map<number, PendingCommand>;
  ext_session_id: string | null;
  ext_target_id: string | null;
  ext_execution_context_id: number | null;
  extension_id: string | null;
  latency: ModCDPPingLatency | null;
  connect_timing: Record<string, unknown> | null;
  last_command_timing: Record<string, unknown> | null;
  last_raw_timing: Record<string, unknown> | null;
  event_schemas: Map<string, z.ZodType>;
  command_params_schemas: Map<string, z.ZodType>;
  command_result_schemas: Map<string, z.ZodType>;
  command_result_unwrap_keys: Map<string, string>;
  cdp_aliases_hydrated: boolean;
  event_wait_cleanups: Set<() => void>;
  auto_sessions: AutoSessionRouter;
  _extension_injectors: ExtensionInjector[];
  _cdp: {
    send: (method: string, params?: ProtocolParams, sessionId?: string | null) => Promise<ProtocolResult>;
    on: (eventName: string | symbol, listener: (...args: unknown[]) => void) => ModCDPClient;
    once: (eventName: string | symbol, listener: (...args: unknown[]) => void) => ModCDPClient;
  };
  _launched: LaunchedBrowser | null;

  constructor({
    launch = {},
    upstream = {},
    extension = {},
    client = {},
    server = {},
    custom_commands = [],
    custom_events = [],
    custom_middlewares = [],
  }: ClientOptions = {}) {
    super();
    const normalized = normalizeClientOptions({ launch, upstream, extension, client, server });
    this.launch = normalized.launch;
    this.upstream = normalized.upstream;
    this.extension = normalized.extension;
    this.client = normalized.client;
    this.upstream_endpoint_kind = normalized.upstream_endpoint_kind;
    this.cdp_url = this.upstream.ws_url;
    this.server = normalized.server;
    this.custom_commands = custom_commands;
    this.custom_events = custom_events;
    this.custom_middlewares = custom_middlewares;

    this.transport = null;
    this.next_id = 1;
    this.pending = new Map();
    this.ext_session_id = null;
    this.ext_target_id = null;
    this.ext_execution_context_id = null;
    this.extension_id = null;
    this.latency = null;
    this.connect_timing = null;
    this.last_command_timing = null;
    this.last_raw_timing = null;
    this.event_schemas = new Map();
    this.command_params_schemas = new Map();
    this.command_result_schemas = new Map();
    this.command_result_unwrap_keys = new Map();
    this.cdp_aliases_hydrated = false;
    this.event_wait_cleanups = new Set();
    this.auto_sessions = new AutoSessionRouter(
      (method, params = {}, session_id = null) =>
        this._sendMessage(method, params, session_id) as Promise<ProtocolResult>,
      () => this.extension.execution_context_timeout_ms,
    );
    this._extension_injectors = [];
    this._launched = null;

    this._cdp = {
      send: (method: string, params: ProtocolParams = {}, session_id: string | null = null) =>
        this._sendMessage(method, params, session_id, { record_raw_timing: true }) as Promise<ProtocolResult>,
      on: (event_name: string | symbol, listener: (...args: unknown[]) => void) => this.on(event_name, listener),
      once: (event_name: string | symbol, listener: (...args: unknown[]) => void) => this.once(event_name, listener),
    };
    void this._hydrateCdpAliases();
    this._hydrateCustomSurface();
  }

  async connect() {
    const connect_started_at = Date.now();
    await this._hydrateCdpAliases();

    const transport_started_at = Date.now();
    await this._connectUpstreamTransport();
    const transport_connected_at = Date.now();
    this.transport?.onRecv((message) => this._onRecv(message));
    this.transport?.onClose((error) => {
      if (this.pending.size > 0) this._rejectAll(error);
    });

    if (this.upstream_endpoint_kind === "modcdp_server") {
      await this.transport?.waitForPeer?.();
      this.event_schemas.set("Mod.pong", Mod.PongEvent);
      if (this.server !== null) {
        await this._sendMessage("Mod.configure", this._serverConfigureParams(), null);
      }
      void this._measurePingLatency().catch(() => {});
      const connected_at = Date.now();
      this.connect_timing = {
        started_at: connect_started_at,
        upstream_mode: this.upstream.mode,
        upstream_endpoint_kind: this.upstream_endpoint_kind,
        transport_started_at,
        transport_connected_at,
        transport_duration_ms: transport_connected_at - transport_started_at,
        connected_at,
        duration_ms: connected_at - connect_started_at,
      };
      return this;
    }

    await this._initializeRawCDPTransport();

    const extension_started_at = Date.now();
    if (this.extension.mode === "none") {
      throw new Error("extension.mode=none cannot be used with a raw_cdp upstream.");
    }
    const ext = await this._injectExtension(
      (method, params, session_id) => this._sendMessage(method, params, session_id) as Promise<ProtocolResult>,
    );
    const extension_completed_at = Date.now();
    this.extension_id = typeof ext.extension_id === "string" ? ext.extension_id : null;
    this.ext_target_id = ext.target_id as string;
    this.ext_session_id = ext.session_id as string;
    this.event_schemas.set("Mod.pong", Mod.PongEvent);

    const ext_context = this.auto_sessions.waitForExecutionContext(this.ext_session_id, {
      timeout_ms: this.extension.execution_context_timeout_ms,
    });
    await this._sendMessage("Runtime.enable", {}, this.ext_session_id);
    this.ext_execution_context_id = await ext_context;
    await Promise.all([
      this._sendMessage("Runtime.addBinding", { name: CUSTOM_EVENT_BINDING_NAME }, this.ext_session_id),
      this.client.mirror_upstream_events
        ? this._sendMessage("Runtime.addBinding", { name: UPSTREAM_EVENT_BINDING_NAME }, this.ext_session_id)
        : Promise.resolve(),
    ]);
    if (this.server !== null) {
      await this._sendRaw(
        wrapCommandIfNeeded("Mod.configure", this._serverConfigureParams(), {
          routes: this.client.routes,
          cdpSessionId: this.ext_session_id,
        }),
      );
    }

    void this._measurePingLatency().catch(() => {});
    const connected_at = Date.now();
    this.connect_timing = {
      started_at: connect_started_at,
      upstream_mode: this.upstream.mode,
      upstream_endpoint_kind: this.upstream_endpoint_kind,
      transport_started_at,
      transport_connected_at,
      transport_duration_ms: transport_connected_at - transport_started_at,
      extension_source: ext.source,
      extension_started_at,
      extension_completed_at,
      extension_duration_ms: extension_completed_at - extension_started_at,
      connected_at,
      duration_ms: connected_at - connect_started_at,
    };
    return this;
  }

  async send(method: string, params: unknown = {}) {
    const started_at = Date.now();
    let command_params = this.command_params_schemas.get(method)?.parse(params ?? {}) ?? params ?? {};
    if (method === "Mod.addCustomCommand") {
      const parsed = Mod.AddCustomCommandParams.parse(command_params);
      const name = normalizeModCDPName(parsed.name);
      const params_schema = normalizeModCDPPayloadSchema(parsed.params_schema);
      const result_schema = normalizeModCDPPayloadSchema(parsed.result_schema);
      if (params_schema) this.command_params_schemas.set(name, params_schema);
      if (result_schema) {
        this.command_result_schemas.set(name, result_schema);
        this._setResultUnwrapKey(name, result_schema);
      }
      defineCustomCommandMethod(this, name);
      if (!parsed.expression) {
        this.last_command_timing = {
          method,
          target: "client",
          started_at,
          completed_at: Date.now(),
          duration_ms: Date.now() - started_at,
        };
        return { name, registered: true };
      }
      command_params = { ...parsed, name, params_schema: null, result_schema: null };
    } else if (method === "Mod.addCustomEvent") {
      const parsed = Mod.AddCustomEventObjectParams.parse(params ?? {});
      const name = normalizeModCDPName(parsed.name);
      const event_schema = normalizeModCDPPayloadSchema(parsed.event_schema);
      if (event_schema) this.event_schemas.set(name, event_schema);
      if (!this.ext_session_id && this.upstream_endpoint_kind !== "modcdp_server") {
        this.last_command_timing = {
          method,
          target: "client",
          started_at,
          completed_at: Date.now(),
          duration_ms: Date.now() - started_at,
        };
        return { name, registered: true };
      }
      command_params = { ...parsed, name, event_schema: null };
    }
    if (this.upstream_endpoint_kind === "modcdp_server") {
      const result = await this._sendMessage(method, command_params as ProtocolParams);
      const completed_at = Date.now();
      this.last_command_timing = {
        method,
        target: "modcdp_server",
        started_at,
        completed_at,
        duration_ms: completed_at - started_at,
      };
      return result;
    }
    const command = wrapCommandIfNeeded(method, command_params as ProtocolParams, {
      routes: this.client.routes,
    });
    const result = await this._sendRaw(command);
    const completed_at = Date.now();
    this.last_command_timing = {
      method,
      target: command.target,
      started_at,
      completed_at,
      duration_ms: completed_at - started_at,
    };
    const result_schema = this.command_result_schemas.get(method);
    if (!result_schema) return result;
    const parsed_result = result_schema.parse(result);
    const unwrap_key = this.command_result_unwrap_keys.get(method);
    return unwrap_key && parsed_result && typeof parsed_result === "object"
      ? (parsed_result as Record<string, unknown>)[unwrap_key]
      : parsed_result;
  }

  async sendRaw(method: string, params: ProtocolParams = {}, session_id: string | null = null) {
    return await this._sendMessage(method, params, session_id);
  }

  async _hydrateCdpAliases() {
    if (!this.client.hydrate_aliases || this.cdp_aliases_hydrated) return;
    Object.assign(
      this,
      createCdpAliases((method, params) => this.send(method, params), {
        onCustomCommand: (name, params_schema, result_schema) => {
          if (params_schema) this.command_params_schemas.set(name, params_schema);
          if (result_schema) {
            this.command_result_schemas.set(name, result_schema);
            this._setResultUnwrapKey(name, result_schema);
          }
          defineCustomCommandMethod(this, name);
        },
        onCustomEvent: (name, event_schema) => {
          if (event_schema) this.event_schemas.set(name, event_schema);
        },
      }),
    );
    this.cdp_aliases_hydrated = true;
  }

  _hydrateCustomSurface() {
    for (const command of this.custom_commands) {
      const name = normalizeModCDPName(command.name);
      const params_schema = command.params_schema ? Mod.PayloadSchemaSpec.parse(command.params_schema) : null;
      const result_schema = command.result_schema ? Mod.PayloadSchemaSpec.parse(command.result_schema) : null;
      const normalized_params_schema = params_schema == null ? null : this._normalizePayloadSchema(params_schema);
      const normalized_result_schema = result_schema == null ? null : this._normalizePayloadSchema(result_schema);
      if (normalized_params_schema) this.command_params_schemas.set(name, normalized_params_schema);
      if (normalized_result_schema) {
        this.command_result_schemas.set(name, normalized_result_schema);
        this._setResultUnwrapKey(name, normalized_result_schema);
      }
      defineCustomCommandMethod(this, name);
    }
    for (const event of this.custom_events) {
      const name = normalizeModCDPName(event.name);
      const event_schema = event.event_schema ? this._normalizePayloadSchema(event.event_schema) : null;
      if (event_schema) this.event_schemas.set(name, event_schema);
    }
  }

  _normalizePayloadSchema(schema: unknown) {
    return normalizeModCDPPayloadSchema(Mod.PayloadSchemaSpec.parse(schema));
  }

  _setResultUnwrapKey(name: string, schema: z.ZodType) {
    const shape = "shape" in schema && schema.shape && typeof schema.shape === "object" ? schema.shape : null;
    const keys = shape ? Object.keys(shape) : [];
    if (keys.length === 1) this.command_result_unwrap_keys.set(name, keys[0]);
    else this.command_result_unwrap_keys.delete(name);
  }

  async _serviceWorkerUrlSuffixes() {
    if (this.extension.service_worker_url_suffixes != null) return this.extension.service_worker_url_suffixes;
    return ["/modcdp/service_worker.js"];
  }

  _serverConfigureParams(): ModCDPConfigureParams {
    return {
      upstream: {
        mode: this.upstream.mode,
        ...(this.upstream.nats_url ? { nats_url: this.upstream.nats_url } : {}),
        ...(this.upstream.nats_subject_prefix ? { nats_subject_prefix: this.upstream.nats_subject_prefix } : {}),
      },
      client: {
        routes: this.client.routes,
      },
      server: {
        ...this._serverDefaults(),
        ...(this.server ?? {}),
      },
      custom_commands: this.custom_commands.filter(hasCommandExpression).map((command) => ({
        name: normalizeModCDPName(command.name),
        expression: command.expression,
        params_schema: null as null,
        result_schema: null as null,
      })),
      custom_events: this.custom_events.map((event) => ({
        name: normalizeModCDPName(event.name),
        event_schema: null as null,
      })),
      custom_middlewares: this.custom_middlewares.map(({ name, phase, expression }) => ({
        ...(name == null ? {} : { name: normalizeModCDPName(name) }),
        phase,
        expression,
      })),
    };
  }

  _serverDefaults(): ModCDPServerOptions {
    return {
      cdp_send_timeout_ms: this.client.cdp_send_timeout_ms,
      loopback_execution_context_timeout_ms: this.extension.execution_context_timeout_ms,
      ws_connect_error_settle_timeout_ms: this.upstream.ws_connect_error_settle_timeout_ms,
    };
  }

  async _connectUpstreamTransport() {
    if (this.transport) return;
    const launcher = this._browserLauncher();
    const transport = this._upstreamTransport();
    const injectors = await this._extensionInjectors();
    this._extension_injectors = injectors;
    const initial_transport_config = this._upstreamTransportConfig();
    transport.update(initial_transport_config);
    launcher.update(await this._launchOptions());
    for (const injector of injectors) injector.update(await this._baseExtensionInjectorConfig());
    for (const injector of injectors) injector.update(launcher.getInjectorConfig());
    for (const injector of injectors) injector.update(transport.getInjectorConfig());
    for (const injector of injectors) await injector.prepare();
    for (const injector of injectors) launcher.update(injector.getLauncherConfig());
    for (const injector of injectors) transport.update(injector.getTransportConfig());
    launcher.update(transport.getLauncherConfig());
    transport.update(launcher.getTransportConfig());

    if (transport.endpoint_kind === "modcdp_server") await transport.connect();
    if (this.launch.mode !== "none") {
      this._launched = await launcher.launch();
      transport.update(launcher.getTransportConfig());
      for (const injector of injectors) injector.update(launcher.getInjectorConfig());
      for (const injector of injectors) transport.update(injector.getTransportConfig());
    }
    const launched_cdp_url = this._launched?.ws_url ?? this._launched?.cdp_url ?? null;
    if (transport.endpoint_kind === "raw_cdp") await transport.connect();

    this.transport = transport;
    this.cdp_url =
      transport.endpoint_kind === "raw_cdp" ? ((transport.url || launched_cdp_url) ?? null) : launched_cdp_url;
    if (transport.mode === "ws" && transport.url) this.upstream.ws_url = transport.url;
    const server_config = {
      ...(transport.endpoint_kind === "modcdp_server" && launched_cdp_url
        ? { loopback_cdp_url: launched_cdp_url }
        : {}),
      ...transport.getServerConfig(),
    };
    if (this.server !== null && server_config.loopback_cdp_url) {
      const configured_loopback = this.server.loopback_cdp_url;
      if (
        !Object.hasOwn(this.server, "loopback_cdp_url") ||
        configured_loopback === initial_transport_config.ws_url ||
        configured_loopback === launched_cdp_url
      ) {
        this.server = { ...this.server, ...server_config };
      }
    }
  }

  _upstreamTransport(): UpstreamTransport {
    const factories = {
      ws: () => new WebSocketUpstreamTransport(),
      pipe: () => new PipeUpstreamTransport(),
      reversews: () => new ReverseWebSocketUpstreamTransport(),
      nativemessaging: () => new NativeMessagingUpstreamTransport(),
      nats: () => new NatsUpstreamTransport(),
    } satisfies Record<UpstreamMode, () => UpstreamTransport>;
    return factories[this.upstream.mode]();
  }

  _browserLauncher(): BrowserLauncher {
    const factories = {
      local: () => new LocalBrowserLauncher(this.launch.options),
      remote: () => new RemoteBrowserLauncher(this.launch.options, this.upstream.ws_url),
      bb: () => new BrowserbaseBrowserLauncher(this.launch.options),
      none: () => new NoopBrowserLauncher(this.launch.options),
    } satisfies Record<LaunchMode, () => BrowserLauncher>;
    return factories[this.launch.mode]();
  }

  async _launchOptions(): Promise<BrowserLaunchOptions> {
    return {
      ...this.launch.options,
      ...(this.launch.executable_path ? { executable_path: this.launch.executable_path } : {}),
      ...(this.launch.user_data_dir ? { user_data_dir: this.launch.user_data_dir } : {}),
    };
  }

  _upstreamTransportConfig() {
    return {
      ws_url: this.upstream.ws_url,
      nats_url: this.upstream.nats_url,
      nats_subject_prefix: this.upstream.nats_subject_prefix,
      reversews_bind: this.upstream.reversews_bind,
      reversews_wait_timeout_ms: this.upstream.reversews_wait_timeout_ms,
      manifest_path: this.upstream.nativemessaging_manifest,
      extension_id: this.extension.extension_id,
    };
  }

  async _extensionInjectors() {
    if (this.extension.mode === "none") return [];
    const injectors: ExtensionInjector[] = [];
    if (this.extension.mode === "auto" || this.extension.mode === "discover") {
      injectors.push(new DiscoveredExtensionInjector());
    }
    if (this.extension.mode === "auto" || this.extension.mode === "inject") {
      if (this.launch.mode === "bb") injectors.push(new BBBrowserExtensionInjector());
      if (this.launch.mode === "local") injectors.push(new LocalBrowserLaunchExtensionInjector());
      injectors.push(new ExtensionsLoadUnpackedInjector());
    }
    if (this.extension.mode === "auto" || this.extension.mode === "borrow") {
      injectors.push(new BorrowedExtensionInjector());
    }
    return injectors;
  }

  async _baseExtensionInjectorConfig(send: SendCDP | null = null): Promise<ExtensionInjectorConfig> {
    const service_worker_url_suffixes = await this._serviceWorkerUrlSuffixes();
    const trust_matched_service_worker =
      this.extension.trust_service_worker_target ||
      this.extension.service_worker_url_includes.length > 0 ||
      service_worker_url_suffixes.some((suffix) => suffix.split("/").filter(Boolean).length > 1);
    return {
      send,
      sessionIdForTarget: (target_id) => this.auto_sessions.sessionIdForTarget(target_id),
      attachToTarget: send ? (target_id) => this.auto_sessions.attachToTarget(target_id) : async () => null,
      waitForExecutionContext: (session_id, timeout_ms) =>
        this.auto_sessions.waitForExecutionContext(session_id, { timeout_ms }),
      extension_path: this.extension.path,
      extension_id: this.extension.extension_id,
      wake_path: this.extension.wake_path,
      wake_url: this.extension.wake_url,
      service_worker_url_includes: this.extension.service_worker_url_includes,
      service_worker_url_suffixes,
      trust_matched_service_worker,
      require_service_worker_target: this.extension.require_service_worker_target || this.extension.mode === "discover",
      service_worker_ready_expression: this.extension.service_worker_ready_expression,
      cdp_send_timeout_ms: this.client.cdp_send_timeout_ms,
      execution_context_timeout_ms: this.extension.execution_context_timeout_ms,
      service_worker_probe_timeout_ms: this.extension.service_worker_probe_timeout_ms,
      service_worker_ready_timeout_ms: this.extension.service_worker_ready_timeout_ms,
      service_worker_poll_interval_ms: this.extension.service_worker_poll_interval_ms,
      target_session_poll_interval_ms: this.extension.target_session_poll_interval_ms,
    };
  }

  async _injectExtension(send: SendCDP, injectors: ExtensionInjector[] | null = null) {
    injectors ??= await this._extensionInjectors();
    const errors: string[] = [];
    for (const injector of injectors) {
      injector.update(await this._baseExtensionInjectorConfig(send));
      try {
        await injector.prepare();
        const result = await injector.inject();
        if (result) return result;
      } catch (error) {
        const message = error instanceof Error ? error.message : String(error);
        injector.last_error = error instanceof Error ? error : new Error(message);
        errors.push(`${injector.constructor.name}: ${message}`);
      }
    }
    throw new Error(
      `Cannot install, discover, or borrow the ModCDP extension in the running browser.` +
        (errors.length ? `\n\n${errors.join("\n")}` : ""),
    );
  }

  async _initializeRawCDPTransport() {
    await Promise.all([
      this._sendMessage("Target.setAutoAttach", {
        autoAttach: true,
        waitForDebuggerOnStart: false,
        flatten: true,
      }),
      this._sendMessage("Target.setDiscoverTargets", { discover: true }),
    ]);
  }

  async close() {
    for (const cleanup of this.event_wait_cleanups) cleanup();
    this.event_wait_cleanups.clear();
    await this.transport?.close();
    this.transport = null;
    if (this._launched) await this._launched.close();
    this._launched = null;
    for (const injector of this._extension_injectors) await injector.close();
    this._extension_injectors = [];
  }

  on<TEvent extends z.ZodType & ModCDPNamedValue>(
    event_name: TEvent,
    listener: (event: ModCDPEventPayload<TEvent>) => void,
  ): this;
  on(event_name: string | symbol, listener: (...args: unknown[]) => void): this;
  on(event_name: ModCDPEventNameInput, listener: (...args: unknown[]) => void): this;
  on(event_name: ModCDPEventNameInput, listener: (...args: unknown[]) => void) {
    if (typeof event_name !== "string" && typeof event_name !== "symbol") {
      const name = normalizeModCDPName(event_name);
      this.event_schemas.set(name, event_name);
      return super.on(name, listener);
    }
    return super.on(typeof event_name === "symbol" ? event_name : normalizeModCDPName(event_name), listener);
  }

  once<TEvent extends z.ZodType & ModCDPNamedValue>(
    event_name: TEvent,
    listener: (event: ModCDPEventPayload<TEvent>) => void,
  ): this;
  once(event_name: string | symbol, listener: (...args: unknown[]) => void): this;
  once(event_name: ModCDPEventNameInput, listener: (...args: unknown[]) => void): this;
  once(event_name: ModCDPEventNameInput, listener: (...args: unknown[]) => void) {
    if (typeof event_name !== "string" && typeof event_name !== "symbol") {
      const name = normalizeModCDPName(event_name);
      this.event_schemas.set(name, event_name);
      return super.once(name, listener);
    }
    return super.once(typeof event_name === "symbol" ? event_name : normalizeModCDPName(event_name), listener);
  }

  off<TEvent extends z.ZodType & ModCDPNamedValue>(
    event_name: TEvent,
    listener: (event: ModCDPEventPayload<TEvent>) => void,
  ): this;
  off(event_name: string | symbol, listener: (...args: unknown[]) => void): this;
  off(event_name: ModCDPEventNameInput, listener: (...args: unknown[]) => void): this;
  off(event_name: ModCDPEventNameInput, listener: (...args: unknown[]) => void) {
    if (typeof event_name !== "string" && typeof event_name !== "symbol") {
      return super.off(normalizeModCDPName(event_name), listener);
    }
    return super.off(typeof event_name === "symbol" ? event_name : normalizeModCDPName(event_name), listener);
  }

  _waitForEvent(event_name: ModCDPEventNameInput, { timeout_ms }: { timeout_ms?: number } = {}) {
    const effective_timeout_ms = timeout_ms ?? this.client.event_wait_timeout_ms;
    let settled = false;
    let timeout: ReturnType<typeof setTimeout> | null = null;
    let cancel: () => void = () => {};
    let listener: (...args: unknown[]) => void = () => {};
    const promise = new Promise((resolve) => {
      const cleanup = () => {
        if (timeout != null) clearTimeout(timeout);
        timeout = null;
        this.off(event_name, listener);
        this.event_wait_cleanups.delete(cancel);
      };
      const finish = (value: unknown) => {
        if (settled) return;
        settled = true;
        cleanup();
        resolve(value);
      };
      cancel = () => finish(null);
      listener = (payload) => finish(payload || {});
      this.event_wait_cleanups.add(cancel);
      this.on(event_name, listener);
      timeout = setTimeout(() => finish(null), effective_timeout_ms);
    });
    return { promise, cancel };
  }

  async _measurePingLatency() {
    const sentAt = Date.now();
    const pong = this._waitForEvent("Mod.pong");
    try {
      await this.send("Mod.ping", { sentAt });
      const payload = (await pong.promise) as ModCDPPongEvent | null;
      if (payload == null) return this.latency;
      const returnedAt = Date.now();
      this.latency = {
        sentAt,
        receivedAt: payload.receivedAt ?? null,
        returnedAt,
        roundTripMs: returnedAt - sentAt,
        serviceWorkerMs: typeof payload.receivedAt === "number" ? payload.receivedAt - sentAt : null,
        returnPathMs: typeof payload.receivedAt === "number" ? returnedAt - payload.receivedAt : null,
      };
      return this.latency;
    } finally {
      pong.cancel();
    }
  }

  async _sendRaw(command: TranslatedCommand) {
    if (command.target === "direct_cdp") {
      const [step] = command.steps;
      return this._sendMessage(step.method, step.params ?? {}) as Promise<ProtocolResult>;
    }
    if (command.target !== "service_worker") {
      throw new Error(`Unsupported command target "${command.target}"`);
    }

    let result: ProtocolResult = {};
    let unwrap = null;
    for (const step of command.steps) {
      const step_params =
        step.method === "Runtime.callFunctionOn" && step.params && !Object.hasOwn(step.params, "executionContextId")
          ? {
              ...step.params,
              executionContextId:
                this.ext_execution_context_id ??
                (await this.auto_sessions.waitForExecutionContext(this.ext_session_id, {
                  timeout_ms: this.extension.execution_context_timeout_ms,
                })),
            }
          : (step.params ?? {});
      result = (await this._sendMessage(step.method, step_params, this.ext_session_id)) as ProtocolResult;
      unwrap = step.unwrap ?? null;
    }
    return unwrapResponseIfNeeded(result, unwrap);
  }

  _sendMessage(
    method: string,
    params: ProtocolParams = {},
    session_id: string | null = null,
    options: { record_raw_timing?: boolean; timeout_ms?: number | null } = {},
  ) {
    if (!this.transport) return Promise.reject(new Error("ModCDP upstream is not connected."));
    const id = this.next_id++;
    const started_at = Date.now();
    const message: CdpCommandMessage = { id, method, params };
    if (session_id) message.sessionId = session_id;
    return new Promise((resolve, reject) => {
      const timeout_ms = options.timeout_ms ?? this.client.cdp_send_timeout_ms;
      let timeout: ReturnType<typeof setTimeout> | null = null;
      const finish = (callback: () => void) => {
        if (timeout != null) clearTimeout(timeout);
        timeout = null;
        callback();
      };
      this.pending.set(id, {
        method,
        resolve: (value: ProtocolResult) => {
          finish(() => {
            if (options.record_raw_timing) {
              const completed_at = Date.now();
              this.last_raw_timing = {
                method,
                started_at,
                completed_at,
                duration_ms: completed_at - started_at,
              };
            }
            resolve(value);
          });
        },
        reject: (error: Error) => {
          finish(() => reject(error));
        },
      });
      if (timeout_ms != null && timeout_ms > 0) {
        timeout = setTimeout(() => {
          if (!this.pending.delete(id)) return;
          reject(new Error(`${method} timed out after ${timeout_ms}ms`));
        }, timeout_ms);
      }
      this.transport?.send(message);
    });
  }

  _rejectAll(error: Error) {
    const pending_methods = [...this.pending.values()].map((pending) => pending.method);
    const reject_error =
      pending_methods.length === 0 ? error : new Error(`${error.message}; pending=${pending_methods.join(",")}`);
    for (const pending of this.pending.values()) pending.reject(reject_error);
    this.pending.clear();
  }

  _onRecv(msg: CdpResponseMessage | CdpEventMessage) {
    if ("id" in msg && typeof msg.id === "number") {
      const response = CdpResponseMessageSchema.parse(msg);
      const pending = this.pending.get(response.id);
      if (!pending) return;
      this.pending.delete(response.id);
      if (response.error) {
        const err = new Error(`${pending.method} failed: ${response.error.message}`) as Error & { cdp?: CdpError };
        err.cdp = response.error;
        pending.reject(err);
      } else {
        pending.resolve(response.result || {});
      }
      return;
    }
    const event = CdpEventMessageSchema.parse(msg);
    if (event.sessionId === this.ext_session_id) {
      if (event.method === "Runtime.executionContextCreated") {
        this.auto_sessions.recordProtocolEvent(event.method, event.params || {}, event.sessionId || null);
      }
      if (event.method !== "Runtime.bindingCalled") return;
      const u = unwrapEventIfNeeded(
        event.method,
        (event.params || {}) as RuntimeBindingCalledEvent,
        event.sessionId || null,
        this.ext_session_id,
      );
      if (u) {
        this.auto_sessions.recordProtocolEvent(u.event, u.data, u.sessionId);
        this.emit(u.event, this.event_schemas.get(u.event)?.parse(u.data) ?? u.data, u.sessionId);
      }
      return;
    }
    if (event.method) {
      const data = event.params || {};
      this.auto_sessions.recordProtocolEvent(event.method, data, event.sessionId || null);
      this.emit(event.method, this.event_schemas.get(event.method)?.parse(data) ?? data, event.sessionId || null);
    }
  }

  get auto_target_sessions() {
    return this.auto_sessions.target_sessions;
  }

  get auto_session_targets() {
    return this.auto_sessions.session_targets;
  }

  get runtime_execution_contexts() {
    return this.auto_sessions.execution_contexts;
  }
}

export interface ModCDPClient extends CdpAliases {}
