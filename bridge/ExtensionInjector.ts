import type { BrowserLaunchOptions } from "./BrowserLauncher.js";
import type { UpstreamTransportConfig } from "./UpstreamTransport.js";
import type { ProtocolParams, ProtocolResult } from "../types/modcdp.js";
import fs from "node:fs";
import path from "node:path";
import { commands as RuntimeCommands } from "../types/zod/Runtime.js";
import { commands as TargetCommands } from "../types/zod/Target.js";

const EXT_ID_FROM_URL = /^chrome-extension:\/\/([a-z]+)\//;
export const DEFAULT_MODCDP_EXTENSION_ID = "mdedooklbnfejodmnhmkdpkaedafkehf";
export const DEFAULT_MODCDP_SERVICE_WORKER_URL_SUFFIXES = ["/modcdp/service_worker.js"];
export const DEFAULT_MODCDP_WAKE_PATH = "/modcdp/wake.html";
const MODCDP_READY_EXPRESSION =
  "Boolean(globalThis.ModCDP?.__ModCDPServerVersion === 1 && globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)";
export const DEFAULT_CDP_SEND_TIMEOUT_MS = 10_000;
export const DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS = 10_000;
export const DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS = 10_000;
export const DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS = 60_000;
export const DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS = 100;
export const DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS = 20;

export type SendCDP = (method: string, params?: ProtocolParams, session_id?: string | null) => Promise<ProtocolResult>;
export type TargetInfo = { targetId: string; type?: string; url?: string };

export type ExtensionInjectorConfig = {
  send?: SendCDP | null;
  sessionIdForTarget?: ((target_id: string) => string | null | undefined) | null;
  attachToTarget?: ((target_id: string) => Promise<string | null | undefined>) | null;
  waitForExecutionContext?: ((session_id: string, timeout_ms: number) => Promise<number>) | null;
  extension_path?: string | null;
  extension_id?: string | null;
  wake_path?: string | null;
  wake_url?: string | null;
  service_worker_url_includes?: string[];
  service_worker_url_suffixes?: string[];
  trust_matched_service_worker?: boolean;
  require_service_worker_target?: boolean;
  service_worker_ready_expression?: string | null;
  cdp_send_timeout_ms?: number;
  execution_context_timeout_ms?: number;
  service_worker_probe_timeout_ms?: number;
  service_worker_ready_timeout_ms?: number;
  service_worker_poll_interval_ms?: number;
  target_session_poll_interval_ms?: number;
  browserbase_api_key?: string | null;
  base_url?: string | null;
  browserbase_base_url?: string | null;
  reverse_proxy_url?: string | null;
  native_host_name?: string | null;
  nats_url?: string | null;
  nats_subject_prefix?: string | null;
};

export type ExtensionInjectionResult = {
  source: string;
  extension_id?: string | null;
  target_id: string;
  url?: string;
  session_id: string;
  has_tabs?: boolean;
  has_debugger?: boolean;
};

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

export class ExtensionInjector {
  options: ExtensionInjectorConfig;
  protected unusable_target_ids = new Set<string>();
  last_error: Error | null = null;

  constructor(options: ExtensionInjectorConfig = {}) {
    this.options = {
      send: null,
      sessionIdForTarget: null,
      attachToTarget: null,
      waitForExecutionContext: null,
      extension_path: null,
      extension_id: null,
      wake_path: DEFAULT_MODCDP_WAKE_PATH,
      wake_url: null,
      service_worker_url_includes: [],
      service_worker_url_suffixes: [],
      trust_matched_service_worker: false,
      require_service_worker_target: false,
      service_worker_ready_expression: null,
      cdp_send_timeout_ms: DEFAULT_CDP_SEND_TIMEOUT_MS,
      execution_context_timeout_ms: DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS,
      service_worker_probe_timeout_ms: DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
      service_worker_ready_timeout_ms: DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS,
      service_worker_poll_interval_ms: DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS,
      target_session_poll_interval_ms: DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS,
      browserbase_api_key: null,
      base_url: null,
      browserbase_base_url: null,
      reverse_proxy_url: null,
      native_host_name: null,
      nats_url: null,
      nats_subject_prefix: null,
      ...options,
    };
  }

  update(config: ExtensionInjectorConfig = {}) {
    this.options = {
      ...this.options,
      ...config,
      service_worker_url_includes: config.service_worker_url_includes ?? this.options.service_worker_url_includes ?? [],
      service_worker_url_suffixes: config.service_worker_url_suffixes ?? this.options.service_worker_url_suffixes ?? [],
    };
    return this;
  }

  getInjectorConfig(): ExtensionInjectorConfig {
    return { ...this.options };
  }

  getLauncherConfig(): BrowserLaunchOptions {
    return {};
  }

  getTransportConfig(): UpstreamTransportConfig {
    return this.options.extension_id ? { extension_id: this.options.extension_id } : {};
  }

  async prepare() {}

  async close() {}

  async inject(): Promise<ExtensionInjectionResult | null> {
    throw new Error(`${this.constructor.name}.inject is not implemented.`);
  }

  protected get send(): SendCDP {
    if (typeof this.options.send !== "function")
      throw new Error(`${this.constructor.name} requires a CDP send function.`);
    return this.options.send;
  }

  protected extensionRuntimeConfig() {
    const config = {
      ...(this.options.reverse_proxy_url ? { reverse_proxy_url: this.options.reverse_proxy_url } : {}),
      ...(this.options.native_host_name ? { native_host_name: this.options.native_host_name } : {}),
      ...(this.options.nats_url ? { nats_url: this.options.nats_url } : {}),
      ...(this.options.nats_subject_prefix ? { nats_subject_prefix: this.options.nats_subject_prefix } : {}),
    };
    return Object.keys(config).length > 0 ? config : null;
  }

  protected writeExtensionRuntimeConfig(unpacked_extension_path: string) {
    const config = this.extensionRuntimeConfig();
    if (!config) return;
    const config_dir = path.join(unpacked_extension_path, "modcdp");
    fs.mkdirSync(config_dir, { recursive: true });
    fs.writeFileSync(path.join(config_dir, "config.json"), `${JSON.stringify(config, null, 2)}\n`);
    fs.writeFileSync(
      path.join(unpacked_extension_path, "config.js"),
      `globalThis.__MODCDP_RUNTIME_CONFIG__ = ${JSON.stringify(config, null, 2)};\nexport {};\n`,
    );
  }

  protected readyExpression() {
    const expression = this.options.service_worker_ready_expression;
    return expression == null || expression.length === 0
      ? MODCDP_READY_EXPRESSION
      : `(${MODCDP_READY_EXPRESSION}) && Boolean(${expression})`;
  }

  protected async sendWithTimeout(
    method: string,
    params: ProtocolParams = {},
    session_id: string | null = null,
    timeout_ms = this.options.cdp_send_timeout_ms ?? DEFAULT_CDP_SEND_TIMEOUT_MS,
  ) {
    let timeout: ReturnType<typeof setTimeout> | null = null;
    return Promise.race([
      this.send(method, params, session_id),
      new Promise<never>((_, reject) => {
        timeout = setTimeout(() => reject(new Error(`${method} timed out after ${timeout_ms}ms`)), timeout_ms);
      }),
    ]).finally(() => {
      if (timeout != null) clearTimeout(timeout);
    });
  }

  protected async sessionIdForTarget(target_id: string, timeout_ms = 0) {
    const deadline = Date.now() + timeout_ms;
    while (true) {
      const session_id = this.options.sessionIdForTarget?.(target_id);
      if (typeof session_id === "string" && session_id.length > 0) return session_id;
      if (Date.now() >= deadline) return null;
      await delay(this.options.target_session_poll_interval_ms ?? DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS);
    }
  }

  protected async ensureSessionIdForTarget(target_id: string, timeout_ms = 0, allow_attach = false) {
    const session_id = this.options.sessionIdForTarget?.(target_id);
    if (typeof session_id === "string" && session_id.length > 0) return session_id;
    if (allow_attach) {
      const attached_session_id = await this.options.attachToTarget?.(target_id);
      if (typeof attached_session_id === "string" && attached_session_id.length > 0) return attached_session_id;
    }
    return await this.sessionIdForTarget(target_id, timeout_ms);
  }

  protected async targetInfos() {
    return TargetCommands["Target.getTargets"].result.parse(await this.send("Target.getTargets")).targetInfos;
  }

  protected configuredWakeUrl() {
    if (this.options.wake_url) return this.options.wake_url;
    if (!this.options.extension_id) return null;
    const wake_path = this.options.wake_path ?? DEFAULT_MODCDP_WAKE_PATH;
    if (!wake_path) return null;
    return `chrome-extension://${this.options.extension_id}${wake_path.startsWith("/") ? wake_path : `/${wake_path}`}`;
  }

  protected async wakeConfiguredExtension() {
    const wake_url = this.configuredWakeUrl();
    if (!wake_url || typeof this.options.send !== "function") return false;
    try {
      await this.sendWithTimeout("Target.createTarget", {
        url: wake_url,
        background: true,
        hidden: true,
        focus: false,
      });
      return true;
    } catch {
      return false;
    }
  }

  protected async probeTarget(
    target: TargetInfo,
    session_timeout_ms = 0,
    { allow_attach = false }: { allow_attach?: boolean } = {},
  ): Promise<ExtensionInjectionResult | null> {
    if (this.unusable_target_ids.has(target.targetId)) return null;
    const session_id = await this.ensureSessionIdForTarget(target.targetId, session_timeout_ms, allow_attach);
    if (session_id == null) return null;
    await this.sendWithTimeout("Runtime.enable", {}, session_id);
    const probe = RuntimeCommands["Runtime.evaluate"].result.parse(
      await this.sendWithTimeout(
        "Runtime.evaluate",
        {
          expression: this.readyExpression(),
          returnByValue: true,
        },
        session_id,
      ),
    );
    if (probe.result?.value !== true) return null;
    return {
      source: "discovered",
      extension_id: target.url?.match(EXT_ID_FROM_URL)?.[1],
      target_id: target.targetId,
      url: target.url,
      session_id,
    };
  }

  protected async discoverReadyServiceWorker({ matched_only = false }: { matched_only?: boolean } = {}) {
    const target_infos = await this.targetInfos();
    if (this.options.trust_matched_service_worker) {
      const trusted_target = target_infos.find((candidate) => this.serviceWorkerTargetMatches(candidate)) as
        | TargetInfo
        | undefined;
      if (trusted_target) {
        const probed = await this.probeTarget(trusted_target, this.options.service_worker_probe_timeout_ms, {
          allow_attach: true,
        });
        if (probed) return { ...probed, source: "trusted" };
      }
    }
    if (this.options.trust_matched_service_worker || matched_only) return null;
    for (const candidate of target_infos) {
      if (candidate.type !== "service_worker") continue;
      if (!candidate.url.startsWith("chrome-extension://")) continue;
      try {
        const probed = await this.probeTarget(candidate as TargetInfo, this.options.service_worker_probe_timeout_ms);
        if (probed) return probed;
      } catch {
        continue;
      }
    }
    return null;
  }

  protected async waitForReadyServiceWorker(
    timeout_ms: number,
    { matched_only = false }: { matched_only?: boolean } = {},
  ) {
    const deadline = Date.now() + timeout_ms;
    while (Date.now() < deadline) {
      const discovered = await this.discoverReadyServiceWorker({ matched_only });
      if (discovered) return discovered;
      await delay(this.options.service_worker_poll_interval_ms ?? DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS);
    }
    return null;
  }

  protected serviceWorkerTargetMatches(candidate: { type?: string; url?: string }) {
    const url = candidate.url ?? "";
    if (candidate.type !== "service_worker") return false;
    if (!url.startsWith("chrome-extension://")) return false;
    const has_extension_id = Boolean(this.options.extension_id);
    if (this.options.extension_id && !url.startsWith(`chrome-extension://${this.options.extension_id}/`)) return false;
    const includes = this.options.service_worker_url_includes ?? [];
    const suffixes = this.options.service_worker_url_suffixes ?? [];
    if (includes.length > 0 && !includes.every((part) => url.includes(part))) return false;
    if (suffixes.length > 0 && !suffixes.some((suffix) => url.endsWith(suffix))) return false;
    return has_extension_id || includes.length > 0 || suffixes.length > 0;
  }
}
