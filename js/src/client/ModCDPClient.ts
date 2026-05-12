// ModCDPClient (JS): importable, no CLI, no demo code.
//
// Constructor option groups mirror the owning runtime components:
//   launcher          browser/session creation and cleanup
//   upstream          message transport to either raw CDP or a ModCDP server
//   injector          raw-CDP extension discovery/injection/borrowing
//   client            client-side routing, alias hydration, event mirroring, send/event timeouts
//   server            ModCDPServer.configure params
//
// Public methods: connect, send(method, params), on(event, handler), close.

// oxlint-disable typescript-eslint/no-unsafe-declaration-merging -- alias members are assigned by connect().
import { existsSync } from "node:fs";
import type { z } from "zod";

import { createCdpAliases, type CdpAliases } from "../types/generated/aliases.js";
export type { CdpAliases } from "../types/generated/aliases.js";
import {
  CUSTOM_EVENT_BINDING_NAME,
  DEFAULT_CLIENT_ROUTES,
  UPSTREAM_EVENT_BINDING_NAME,
  wrapCommandIfNeeded,
  unwrapResponseIfNeeded,
  unwrapEventIfNeeded,
} from "../translate/translate.js";
import {
  endpointKindForUpstream,
  type UpstreamEndpointKind,
  type UpstreamMode,
  type UpstreamTransport,
} from "../transport/UpstreamTransport.js";
import {
  DEFAULT_UPSTREAM_REVERSEWS_BIND,
  DEFAULT_UPSTREAM_REVERSEWS_WAIT_TIMEOUT_MS,
  ReverseWebSocketUpstreamTransport,
} from "../transport/ReverseWebSocketUpstreamTransport.js";
import { WebSocketUpstreamTransport } from "../transport/WebSocketUpstreamTransport.js";
import {
  DEFAULT_UPSTREAM_NATIVEMESSAGING_WAIT_TIMEOUT_MS,
  NativeMessagingUpstreamTransport,
} from "../transport/NativeMessagingUpstreamTransport.js";
import { PipeUpstreamTransport } from "../transport/PipeUpstreamTransport.js";
import { DEFAULT_UPSTREAM_NATS_WAIT_TIMEOUT_MS, NatsUpstreamTransport } from "../transport/NatsUpstreamTransport.js";
import { LocalBrowserLauncher } from "../launcher/LocalBrowserLauncher.js";
import { RemoteBrowserLauncher } from "../launcher/RemoteBrowserLauncher.js";
import { BrowserbaseBrowserLauncher } from "../launcher/BrowserbaseBrowserLauncher.js";
import { NoopBrowserLauncher } from "../launcher/NoopBrowserLauncher.js";
import type { BrowserLauncher, BrowserLaunchOptions, LaunchedBrowser } from "../launcher/BrowserLauncher.js";
import { BBBrowserExtensionInjector } from "../injector/BBBrowserExtensionInjector.js";
import { BorrowedExtensionInjector } from "../injector/BorrowedExtensionInjector.js";
import { DiscoveredExtensionInjector } from "../injector/DiscoveredExtensionInjector.js";
import {
  DEFAULT_MODCDP_SERVICE_WORKER_URL_SUFFIXES,
  DEFAULT_MODCDP_WAKE_PATH,
  ExtensionInjector,
  type ExtensionInjectorConfig,
  type SendCDP,
} from "../injector/ExtensionInjector.js";
import { ExtensionsLoadUnpackedInjector } from "../injector/ExtensionsLoadUnpackedInjector.js";
import { LocalBrowserLaunchExtensionInjector } from "../injector/LocalBrowserLaunchExtensionInjector.js";
import { AutoSessionRouter } from "../router/AutoSessionRouter.js";
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
} from "../types/modcdp.js";
import {
  CdpEventMessageSchema,
  CdpResponseMessageSchema,
  Mod,
  normalizeModCDPName,
  normalizeModCDPPayloadSchema,
} from "../types/modcdp.js";

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
export type LauncherMode = "local" | "remote" | "bb" | "none";
export type InjectorMode = "auto" | "discover" | "inject" | "borrow" | "none";
export type LauncherOptions = {
  launcher_mode?: LauncherMode;
  launcher_executable_path?: string | null;
  launcher_user_data_dir?: string | null;
  launcher_options?: Record<string, unknown>;
};
export type UpstreamOptions = {
  upstream_mode?: UpstreamMode;
  upstream_cdp_url?: string | null;
  upstream_nats_url?: string | null;
  upstream_nats_subject_prefix?: string | null;
  upstream_nats_wait_timeout_ms?: number;
  upstream_reversews_bind?: string | null;
  upstream_reversews_wait_timeout_ms?: number;
  upstream_nativemessaging_manifest?: string | null;
  upstream_nativemessaging_host_name?: string | null;
  upstream_nativemessaging_wait_timeout_ms?: number;
  upstream_ws_connect_error_settle_timeout_ms?: number;
};
export type InjectorOptions = {
  injector_mode?: InjectorMode;
  injector_extension_path?: string | null;
  injector_extension_id?: string | null;
  injector_wake_path?: string | null;
  injector_wake_url?: string | null;
  injector_service_worker_url_includes?: string[];
  injector_service_worker_url_suffixes?: string[] | null;
  injector_trust_service_worker_target?: boolean;
  injector_require_service_worker_target?: boolean;
  injector_service_worker_ready_expression?: string | null;
  injector_execution_context_timeout_ms?: number;
  injector_service_worker_probe_timeout_ms?: number;
  injector_service_worker_ready_timeout_ms?: number;
  injector_service_worker_poll_interval_ms?: number;
  injector_target_session_poll_interval_ms?: number;
};
type InjectorConfig = Omit<Required<InjectorOptions>, "injector_service_worker_url_suffixes"> & {
  injector_service_worker_url_suffixes: string[];
};
export type ClientConfigOptions = {
  client_routes?: ModCDPRoutes;
  client_hydrate_aliases?: boolean;
  client_mirror_upstream_events?: boolean;
  client_cdp_send_timeout_ms?: number;
  client_event_wait_timeout_ms?: number;
};
export type ClientOptions = {
  launcher?: LauncherOptions;
  upstream?: UpstreamOptions;
  injector?: InjectorOptions;
  client?: ClientConfigOptions;
  server?: ModCDPServerOptions | null;
  custom_commands?: ModCDPClientCustomCommandParams[];
  custom_events?: ModCDPAddCustomEventObjectParams[];
  custom_middlewares?: ModCDPAddMiddlewareParams[];
};
type ClientConfig = {
  client_routes: ModCDPRoutes;
  client_hydrate_aliases: boolean;
  client_mirror_upstream_events: boolean;
  client_cdp_send_timeout_ms: number;
  client_event_wait_timeout_ms: number;
};
type NormalizedClientOptions = {
  launcher: Required<LauncherOptions>;
  upstream: Required<UpstreamOptions>;
  injector: InjectorConfig;
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
  launcher = {},
  upstream = {},
  injector = {},
  client = {},
  server = {},
}: ClientOptions) {
  const upstream_mode = upstream.upstream_mode ?? "ws";
  const upstream_endpoint_kind = endpointKindForUpstream(upstream_mode);
  const launcher_mode =
    launcher.launcher_mode ??
    (upstream_endpoint_kind === "modcdp_server" ? "none" : upstream.upstream_cdp_url ? "remote" : "local");
  const injector_mode =
    injector.injector_mode ?? (upstream_endpoint_kind === "raw_cdp" || launcher_mode !== "none" ? "auto" : "none");
  return {
    launcher: {
      launcher_mode,
      launcher_executable_path: launcher.launcher_executable_path ?? null,
      launcher_user_data_dir: launcher.launcher_user_data_dir ?? null,
      launcher_options: launcher.launcher_options ?? {},
    },
    upstream: {
      upstream_mode,
      upstream_cdp_url: upstream.upstream_cdp_url ?? null,
      upstream_nats_url: upstream.upstream_nats_url ?? null,
      upstream_nats_subject_prefix: upstream.upstream_nats_subject_prefix ?? null,
      upstream_nats_wait_timeout_ms: upstream.upstream_nats_wait_timeout_ms ?? DEFAULT_UPSTREAM_NATS_WAIT_TIMEOUT_MS,
      upstream_reversews_bind: upstream.upstream_reversews_bind ?? DEFAULT_UPSTREAM_REVERSEWS_BIND,
      upstream_reversews_wait_timeout_ms:
        upstream.upstream_reversews_wait_timeout_ms ?? DEFAULT_UPSTREAM_REVERSEWS_WAIT_TIMEOUT_MS,
      upstream_nativemessaging_manifest: upstream.upstream_nativemessaging_manifest ?? null,
      upstream_nativemessaging_host_name: upstream.upstream_nativemessaging_host_name ?? null,
      upstream_nativemessaging_wait_timeout_ms:
        upstream.upstream_nativemessaging_wait_timeout_ms ?? DEFAULT_UPSTREAM_NATIVEMESSAGING_WAIT_TIMEOUT_MS,
      upstream_ws_connect_error_settle_timeout_ms:
        upstream.upstream_ws_connect_error_settle_timeout_ms ?? DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS,
    },
    injector: {
      injector_mode,
      injector_extension_path: injector.injector_extension_path ?? defaultExtensionPath(),
      injector_extension_id: injector.injector_extension_id ?? null,
      injector_wake_path: injector.injector_wake_path ?? DEFAULT_MODCDP_WAKE_PATH,
      injector_wake_url: injector.injector_wake_url ?? null,
      injector_service_worker_url_includes: injector.injector_service_worker_url_includes ?? [],
      injector_service_worker_url_suffixes:
        injector.injector_service_worker_url_suffixes ?? DEFAULT_MODCDP_SERVICE_WORKER_URL_SUFFIXES,
      injector_trust_service_worker_target: injector.injector_trust_service_worker_target ?? false,
      injector_require_service_worker_target: injector.injector_require_service_worker_target ?? false,
      injector_service_worker_ready_expression: injector.injector_service_worker_ready_expression ?? null,
      injector_execution_context_timeout_ms:
        injector.injector_execution_context_timeout_ms ?? DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS,
      injector_service_worker_probe_timeout_ms:
        injector.injector_service_worker_probe_timeout_ms ?? DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
      injector_service_worker_ready_timeout_ms:
        injector.injector_service_worker_ready_timeout_ms ?? DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS,
      injector_service_worker_poll_interval_ms:
        injector.injector_service_worker_poll_interval_ms ?? DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS,
      injector_target_session_poll_interval_ms:
        injector.injector_target_session_poll_interval_ms ?? DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS,
    },
    client: {
      client_routes: { ...DEFAULT_CLIENT_ROUTES, ...(client.client_routes ?? {}) },
      client_hydrate_aliases: client.client_hydrate_aliases ?? true,
      client_mirror_upstream_events: client.client_mirror_upstream_events ?? true,
      client_cdp_send_timeout_ms: client.client_cdp_send_timeout_ms ?? DEFAULT_CDP_SEND_TIMEOUT_MS,
      client_event_wait_timeout_ms: client.client_event_wait_timeout_ms ?? DEFAULT_EVENT_WAIT_TIMEOUT_MS,
    },
    server:
      server === null
        ? null
        : {
            ...(upstream_endpoint_kind === "modcdp_server" ? { server_routes: { "*.*": "chrome_debugger" } } : {}),
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
          cdp_command_name: {
            value: command_name,
            enumerable: true,
            configurable: true,
          },
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
    meta: {
      value: () => ({
        cdp_command_name: name,
        id: name,
        name,
        kind: "command",
      }),
      configurable: true,
    },
  });
  target[domain][method] = alias;
}

function defaultExtensionPath() {
  if (typeof process === "object" && process?.versions?.node && import.meta.url.startsWith("file:")) {
    const candidates = [
      "../../../dist/extension",
      "../../../extension",
      "../../../../dist/extension",
      "../../../../extension",
    ].map((relative_path) => decodeURIComponent(new URL(/* @vite-ignore */ relative_path, import.meta.url).pathname));
    return candidates.find(isBuiltExtensionPath) ?? candidates[0];
  }
  return "../../../dist/extension";
}

function isBuiltExtensionPath(candidate: string) {
  return existsSync(`${candidate}/modcdp/service_worker.js`);
}

function hasCommandExpression(
  command: ModCDPClientCustomCommandParams,
): command is ModCDPClientCustomCommandParams & { expression: string } {
  return typeof command.expression === "string" && command.expression.length > 0;
}

export class ModCDPClient extends ModCDPEventEmitter {
  launcher: NormalizedClientOptions["launcher"];
  upstream: NormalizedClientOptions["upstream"];
  injector: NormalizedClientOptions["injector"];
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
  _injectors: ExtensionInjector[];
  _cdp: {
    send: (method: string, params?: ProtocolParams, sessionId?: string | null) => Promise<ProtocolResult>;
    on: (eventName: string | symbol, listener: (...args: unknown[]) => void) => ModCDPClient;
    once: (eventName: string | symbol, listener: (...args: unknown[]) => void) => ModCDPClient;
  };
  _launched: LaunchedBrowser | null;

  constructor({
    launcher = {},
    upstream = {},
    injector = {},
    client = {},
    server = {},
    custom_commands = [],
    custom_events = [],
    custom_middlewares = [],
  }: ClientOptions = {}) {
    super();
    const normalized = normalizeClientOptions({
      launcher,
      upstream,
      injector,
      client,
      server,
    });
    this.launcher = normalized.launcher;
    this.upstream = normalized.upstream;
    this.injector = normalized.injector;
    this.client = normalized.client;
    this.upstream_endpoint_kind = normalized.upstream_endpoint_kind;
    this.cdp_url = this.upstream.upstream_cdp_url;
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
      () => this.injector.injector_execution_context_timeout_ms,
    );
    this._injectors = [];
    this._launched = null;

    this._cdp = {
      send: (method: string, params: ProtocolParams = {}, session_id: string | null = null) =>
        this._sendMessage(method, params, session_id, {
          record_raw_timing: true,
        }) as Promise<ProtocolResult>,
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
        upstream_mode: this.upstream.upstream_mode,
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

    const injector_started_at = Date.now();
    if (this.injector.injector_mode === "none") {
      throw new Error("injector.injector_mode=none cannot be used with a raw_cdp upstream.");
    }
    const ext = await this._runInjectors(
      (method, params, session_id) => this._sendMessage(method, params, session_id) as Promise<ProtocolResult>,
    );
    const injector_completed_at = Date.now();
    this.extension_id = typeof ext.extension_id === "string" ? ext.extension_id : null;
    this.ext_target_id = ext.target_id as string;
    this.ext_session_id = ext.session_id as string;
    this.event_schemas.set("Mod.pong", Mod.PongEvent);

    const ext_context = this.auto_sessions.waitForExecutionContext(this.ext_session_id, {
      timeout_ms: this.injector.injector_execution_context_timeout_ms,
    });
    await this._sendMessage("Runtime.enable", {}, this.ext_session_id);
    this.ext_execution_context_id = await ext_context;
    await Promise.all([
      this._sendMessage("Runtime.addBinding", { name: CUSTOM_EVENT_BINDING_NAME }, this.ext_session_id),
      this.client.client_mirror_upstream_events
        ? this._sendMessage("Runtime.addBinding", { name: UPSTREAM_EVENT_BINDING_NAME }, this.ext_session_id)
        : Promise.resolve(),
    ]);
    if (this.server !== null) {
      await this._sendRaw(
        wrapCommandIfNeeded("Mod.configure", this._serverConfigureParams(), {
          routes: this.client.client_routes,
          cdpSessionId: this.ext_session_id,
        }),
      );
    }

    void this._measurePingLatency().catch(() => {});
    const connected_at = Date.now();
    this.connect_timing = {
      started_at: connect_started_at,
      upstream_mode: this.upstream.upstream_mode,
      upstream_endpoint_kind: this.upstream_endpoint_kind,
      transport_started_at,
      transport_connected_at,
      transport_duration_ms: transport_connected_at - transport_started_at,
      injector_source: ext.source,
      injector_started_at,
      injector_completed_at,
      injector_duration_ms: injector_completed_at - injector_started_at,
      connected_at,
      duration_ms: connected_at - connect_started_at,
    };
    return this;
  }

  async send(method: string, params: unknown = {}, session_id: string | null = null) {
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
      command_params = {
        ...parsed,
        name,
        params_schema: null,
        result_schema: null,
      };
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
      routes: this.client.client_routes,
      targetCdpSessionId: session_id,
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
    if (!this.client.client_hydrate_aliases || this.cdp_aliases_hydrated) return;
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

  _serviceWorkerUrlSuffixes() {
    return this.injector.injector_service_worker_url_suffixes;
  }

  _serverConfigureParams(): ModCDPConfigureParams {
    return {
      upstream: {
        upstream_mode: this.upstream.upstream_mode,
        ...(this.upstream.upstream_nats_url ? { upstream_nats_url: this.upstream.upstream_nats_url } : {}),
        ...(this.upstream.upstream_nats_subject_prefix
          ? {
              upstream_nats_subject_prefix: this.upstream.upstream_nats_subject_prefix,
            }
          : {}),
      },
      client: {
        client_routes: this.client.client_routes,
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
      server_cdp_send_timeout_ms: this.client.client_cdp_send_timeout_ms,
      server_loopback_execution_context_timeout_ms: this.injector.injector_execution_context_timeout_ms,
      server_ws_connect_error_settle_timeout_ms: this.upstream.upstream_ws_connect_error_settle_timeout_ms,
    };
  }

  async _connectUpstreamTransport() {
    if (this.transport) return;
    const launcher = this._browserLauncher();
    const transport = this._upstreamTransport();
    const injectors = this._injectorsForConfig();
    this._injectors = injectors;
    const initial_transport_config = this._upstreamTransportConfig();
    transport.update(initial_transport_config);
    launcher.update(this._launcherOptions());
    for (const injector of injectors) injector.update(this._baseInjectorConfig());
    for (const injector of injectors) injector.update(launcher.getInjectorConfig());
    for (const injector of injectors) injector.update(transport.getInjectorConfig());
    for (const injector of injectors) await injector.prepare();
    for (const injector of injectors) launcher.update(injector.getLauncherConfig());
    for (const injector of injectors) transport.update(injector.getTransportConfig());
    launcher.update(transport.getLauncherConfig());
    launcher.update({ loopback_cdp: this._serverNeedsLoopbackCdp() });
    transport.update(launcher.getTransportConfig());

    if (transport.endpoint_kind === "modcdp_server") await transport.connect();
    if (this.launcher.launcher_mode !== "none") {
      this._launched = await launcher.launch();
      transport.update(launcher.getTransportConfig());
      for (const injector of injectors) injector.update(launcher.getInjectorConfig());
      for (const injector of injectors) transport.update(injector.getTransportConfig());
    }
    const launched_cdp_url = this._launched?.cdp_url ?? null;
    if (transport.endpoint_kind === "raw_cdp") await transport.connect();

    this.transport = transport;
    this.cdp_url =
      transport.endpoint_kind === "raw_cdp" ? ((transport.url || launched_cdp_url) ?? null) : launched_cdp_url;
    // For ws mode, cdp_url has been resolved to the concrete WebSocket CDP endpoint after connect().
    if (transport.mode === "ws" && transport.url) this.upstream.upstream_cdp_url = transport.url;
    const server_config = {
      ...(transport.endpoint_kind === "modcdp_server" && launched_cdp_url && !launched_cdp_url.startsWith("pipe://")
        ? { server_loopback_cdp_url: launched_cdp_url }
        : {}),
      ...launcher.getServerConfig(),
      ...transport.getServerConfig(),
    };
    if (this.server !== null && server_config.server_loopback_cdp_url) {
      const configured_loopback = this.server.server_loopback_cdp_url;
      if (
        !Object.hasOwn(this.server, "server_loopback_cdp_url") ||
        configured_loopback === initial_transport_config.cdp_url ||
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
    const factory = factories[this.upstream.upstream_mode as UpstreamMode];
    if (!factory) throw new Error(`unknown upstream.upstream_mode=${this.upstream.upstream_mode}`);
    return factory();
  }

  _browserLauncher(): BrowserLauncher {
    const factories = {
      local: () => new LocalBrowserLauncher(this.launcher.launcher_options),
      remote: () => new RemoteBrowserLauncher(this.launcher.launcher_options, this.upstream.upstream_cdp_url),
      bb: () => new BrowserbaseBrowserLauncher(this.launcher.launcher_options),
      none: () => new NoopBrowserLauncher(this.launcher.launcher_options),
    } satisfies Record<LauncherMode, () => BrowserLauncher>;
    const factory = factories[this.launcher.launcher_mode as LauncherMode];
    if (!factory) throw new Error(`unknown launcher.launcher_mode=${this.launcher.launcher_mode}`);
    return factory();
  }

  _launcherOptions(): BrowserLaunchOptions {
    return {
      ...this.launcher.launcher_options,
      ...(this.launcher.launcher_executable_path ? { executable_path: this.launcher.launcher_executable_path } : {}),
      ...(this.launcher.launcher_user_data_dir ? { user_data_dir: this.launcher.launcher_user_data_dir } : {}),
    };
  }

  _serverNeedsLoopbackCdp() {
    if (!this.server || this.server.server_loopback_cdp_url) return false;
    return Object.values(this.server.server_routes ?? {}).includes("loopback_cdp");
  }

  _upstreamTransportConfig() {
    return {
      cdp_url: this.upstream.upstream_cdp_url,
      upstream_nats_url: this.upstream.upstream_nats_url,
      upstream_nats_subject_prefix: this.upstream.upstream_nats_subject_prefix,
      upstream_nats_wait_timeout_ms: this.upstream.upstream_nats_wait_timeout_ms,
      upstream_reversews_bind: this.upstream.upstream_reversews_bind,
      upstream_reversews_wait_timeout_ms: this.upstream.upstream_reversews_wait_timeout_ms,
      upstream_nativemessaging_manifest: this.upstream.upstream_nativemessaging_manifest,
      upstream_nativemessaging_host_name: this.upstream.upstream_nativemessaging_host_name,
      upstream_nativemessaging_wait_timeout_ms: this.upstream.upstream_nativemessaging_wait_timeout_ms,
      injector_extension_id: this.injector.injector_extension_id,
    };
  }

  _injectorsForConfig() {
    if (this.injector.injector_mode === "none") return [];
    const injectors: ExtensionInjector[] = [];
    if (this.injector.injector_mode === "auto" || this.injector.injector_mode === "discover") {
      injectors.push(new DiscoveredExtensionInjector());
    }
    if (this.injector.injector_mode === "auto" || this.injector.injector_mode === "inject") {
      if (this.launcher.launcher_mode === "bb") injectors.push(new BBBrowserExtensionInjector());
      if (this.launcher.launcher_mode === "local") injectors.push(new LocalBrowserLaunchExtensionInjector());
      injectors.push(new ExtensionsLoadUnpackedInjector());
    }
    if (this.injector.injector_mode === "auto" || this.injector.injector_mode === "borrow") {
      injectors.push(new BorrowedExtensionInjector());
    }
    if (injectors.length === 0) throw new Error(`unknown injector.injector_mode=${this.injector.injector_mode}`);
    return injectors;
  }

  _baseInjectorConfig(send: SendCDP | null = null): ExtensionInjectorConfig {
    const service_worker_url_suffixes = this._serviceWorkerUrlSuffixes();
    const trust_service_worker_target =
      this.injector.injector_trust_service_worker_target ||
      this.injector.injector_service_worker_url_includes.length > 0 ||
      service_worker_url_suffixes.some((suffix) => suffix.split("/").filter(Boolean).length > 1);
    return {
      send,
      sessionIdForTarget: (target_id) => this.auto_sessions.sessionIdForTarget(target_id),
      attachToTarget: send ? (target_id) => this.auto_sessions.attachToTarget(target_id) : null,
      waitForExecutionContext: (session_id, timeout_ms) =>
        this.auto_sessions.waitForExecutionContext(session_id, { timeout_ms }),
      injector_extension_path: this.injector.injector_extension_path,
      injector_extension_id: this.injector.injector_extension_id,
      injector_wake_path: this.injector.injector_wake_path,
      injector_wake_url: this.injector.injector_wake_url,
      injector_service_worker_url_includes: this.injector.injector_service_worker_url_includes,
      injector_service_worker_url_suffixes: service_worker_url_suffixes,
      injector_trust_service_worker_target: trust_service_worker_target,
      injector_require_service_worker_target:
        this.injector.injector_require_service_worker_target || this.injector.injector_mode === "discover",
      injector_service_worker_ready_expression: this.injector.injector_service_worker_ready_expression,
      injector_cdp_send_timeout_ms: this.client.client_cdp_send_timeout_ms,
      injector_execution_context_timeout_ms: this.injector.injector_execution_context_timeout_ms,
      injector_service_worker_probe_timeout_ms: this.injector.injector_service_worker_probe_timeout_ms,
      injector_service_worker_ready_timeout_ms: this.injector.injector_service_worker_ready_timeout_ms,
      injector_service_worker_poll_interval_ms: this.injector.injector_service_worker_poll_interval_ms,
      injector_target_session_poll_interval_ms: this.injector.injector_target_session_poll_interval_ms,
    };
  }

  async _runInjectors(send: SendCDP, injectors: ExtensionInjector[] | null = null) {
    injectors ??= this._injectorsForConfig();
    const errors: string[] = [];
    for (const injector of injectors) {
      injector.update(this._baseInjectorConfig(send));
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
    for (const injector of this._injectors) await injector.close();
    this._injectors = [];
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
    const effective_timeout_ms = timeout_ms ?? this.client.client_event_wait_timeout_ms;
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
    const sent_at = Date.now();
    const pong = this._waitForEvent("Mod.pong");
    try {
      await this.send("Mod.ping", { sent_at });
      const payload = (await pong.promise) as ModCDPPongEvent | null;
      if (payload == null) return this.latency;
      const returned_at = Date.now();
      this.latency = {
        sent_at,
        received_at: payload.received_at ?? null,
        returned_at,
        round_trip_ms: returned_at - sent_at,
        service_worker_ms: typeof payload.received_at === "number" ? payload.received_at - sent_at : null,
        return_path_ms: typeof payload.received_at === "number" ? returned_at - payload.received_at : null,
      };
      return this.latency;
    } finally {
      pong.cancel();
    }
  }

  async _sendRaw(command: TranslatedCommand) {
    if (command.target === "direct_cdp") {
      const [step] = command.steps;
      return this._sendMessage(step.method, step.params ?? {}, step.sessionId ?? null) as Promise<ProtocolResult>;
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
                  timeout_ms: this.injector.injector_execution_context_timeout_ms,
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
      const timeout_ms = options.timeout_ms ?? this.client.client_cdp_send_timeout_ms;
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
        pending.resolve((response.result === undefined ? {} : response.result) as ProtocolResult);
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
