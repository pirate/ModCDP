import type { ExtensionInjectorConfig } from "../injector/ExtensionInjector.js";
import type { UpstreamTransportConfig } from "../transport/UpstreamTransport.js";

export type BrowserLaunchOptions = {
  executable_path?: string | null;
  port?: number | null;
  user_data_dir?: string | null;
  headless?: boolean;
  sandbox?: boolean;
  args?: string[];
  extra_args?: string[];
  remote_debugging?: "port" | "pipe";
  cleanup_user_data_dir?: boolean;
  chrome_ready_timeout_ms?: number;
  chrome_ready_poll_interval_ms?: number;
  cdp_url?: string | null;
  browserbase_api_key?: string | null;
  browserbase_base_url?: string | null;
  browserbase_session_id?: string | null;
  browserbase_keep_alive?: boolean;
  browserbase_close_session_on_close?: boolean;
  region?: string | null;
  timeout?: number | null;
  injector_extension_id?: string | null;
  browserbase_browser_settings?: Record<string, unknown> | null;
  browserbase_user_metadata?: Record<string, unknown> | null;
  browserbase_session_create_params?: Record<string, unknown> | null;
};

export type LaunchedBrowser = {
  proc?: unknown;
  port?: number;
  // Effective CDP endpoint for the selected transport; launchers resolve HTTP discovery endpoints to ws:// before returning when they can.
  cdp_url: string | null;
  pipe_read?: NodeJS.ReadableStream | null;
  pipe_write?: NodeJS.WritableStream | null;
  profile_dir?: string | null;
  browserbase_session_id?: string | null;
  browserbase_session_url?: string | null;
  browserbase_debug_url?: string | null;
  close: () => Promise<void> | void;
};

export const DEFAULT_CHROME_READY_TIMEOUT_MS = 45_000;
export const DEFAULT_CHROME_READY_POLL_INTERVAL_MS = 100;

function mergeChromeArgs(existing: string[] = [], incoming: string[] = []) {
  const args = [...existing, ...incoming];
  const load_extension_paths: string[] = [];
  const merged: string[] = [];
  for (const arg of args) {
    if (!arg.startsWith("--load-extension=")) {
      merged.push(arg);
      continue;
    }
    for (const extension_path of arg.slice("--load-extension=".length).split(",")) {
      if (extension_path && !load_extension_paths.includes(extension_path)) load_extension_paths.push(extension_path);
    }
  }
  if (load_extension_paths.length > 0) {
    const first_url_index = merged.findIndex((arg) => !arg.startsWith("-"));
    const load_extension_arg = `--load-extension=${load_extension_paths.join(",")}`;
    if (first_url_index === -1) merged.push(load_extension_arg);
    else merged.splice(first_url_index, 0, load_extension_arg);
  }
  return merged;
}

export class BrowserLauncher {
  options: BrowserLaunchOptions;
  launched: LaunchedBrowser | null = null;

  constructor(options: BrowserLaunchOptions = {}) {
    this.options = { ...options };
  }

  update(config: BrowserLaunchOptions = {}) {
    this.options = {
      ...this.options,
      ...config,
      ...(config.args ? { args: mergeChromeArgs(this.options.args, config.args) } : {}),
      ...(config.extra_args
        ? {
            extra_args: mergeChromeArgs(this.options.extra_args, config.extra_args),
          }
        : {}),
    };
    return this;
  }

  getTransportConfig(): UpstreamTransportConfig {
    return {
      cdp_url: this.launched?.cdp_url ?? this.options.cdp_url ?? null,
      user_data_dir: this.launched?.profile_dir ?? this.options.user_data_dir ?? null,
      pipe_read: this.launched?.pipe_read ?? null,
      pipe_write: this.launched?.pipe_write ?? null,
    };
  }

  getInjectorConfig(): ExtensionInjectorConfig {
    return {
      injector_browserbase_api_key: this.options.browserbase_api_key ?? null,
      injector_browserbase_base_url: this.options.browserbase_base_url ?? null,
      injector_extension_id: this.options.injector_extension_id ?? null,
    };
  }

  async launch(_options: BrowserLaunchOptions = {}): Promise<LaunchedBrowser> {
    throw new Error(`${this.constructor.name}.launch is not implemented.`);
  }
}

export async function resolveCdpWebSocketUrl(endpoint: string, name = "cdp_url") {
  if (/^wss?:\/\//i.test(endpoint)) return endpoint;
  const httpEndpoint = /^[a-z][a-z\d+\-.]*:\/\//i.test(endpoint) ? endpoint : `http://${endpoint}`;
  const response = await fetch(`${httpEndpoint.replace(/\/$/, "")}/json/version`);
  if (!response.ok) throw new Error(`GET ${httpEndpoint}/json/version -> ${response.status}`);
  const version = await response.json();
  if (!version.webSocketDebuggerUrl) throw new Error(`${name} HTTP discovery returned no webSocketDebuggerUrl`);
  return version.webSocketDebuggerUrl as string;
}
