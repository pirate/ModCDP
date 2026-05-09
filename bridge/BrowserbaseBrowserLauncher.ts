import { BrowserLauncher, type BrowserLaunchOptions, type LaunchedBrowser } from "./BrowserLauncher.js";

const DEFAULT_BROWSERBASE_BASE_URL = "https://api.browserbase.com";
const DEFAULT_BROWSERBASE_VIEWPORT = { width: 1288, height: 711 };

type BrowserbaseSession = {
  id?: string;
  connectUrl?: string;
  debuggerUrl?: string;
  debuggerFullscreenUrl?: string;
  status?: string;
};

function firstString(...values: unknown[]) {
  for (const value of values) {
    if (typeof value === "string" && value.trim()) return value.trim();
  }
  return null;
}

function firstBoolean(...values: unknown[]) {
  for (const value of values) {
    if (typeof value === "boolean") return value;
  }
  return null;
}

function objectValue(value: unknown): Record<string, unknown> {
  return value && typeof value === "object" && !Array.isArray(value) ? (value as Record<string, unknown>) : {};
}

function browserbaseUrl(base_url: string, pathname: string) {
  return new URL(pathname, `${base_url.replace(/\/$/, "")}/`).toString();
}

async function browserbaseRequest<T>({
  base_url,
  browserbase_api_key,
  method,
  pathname,
  body,
}: {
  base_url: string;
  browserbase_api_key: string;
  method: "GET" | "POST";
  pathname: string;
  body?: Record<string, unknown>;
}) {
  const response = await fetch(browserbaseUrl(base_url, pathname), {
    method,
    headers: {
      "content-type": "application/json",
      "x-bb-api-key": browserbase_api_key,
    },
    body: body == null ? undefined : JSON.stringify(body),
  });
  if (!response.ok) {
    const text = await response.text().catch(() => "");
    throw new Error(`Browserbase ${method} ${pathname} -> ${response.status}${text ? `: ${text}` : ""}`);
  }
  return (await response.json()) as T;
}

export class BrowserbaseBrowserLauncher extends BrowserLauncher {
  async launch(options: BrowserLaunchOptions = {}): Promise<LaunchedBrowser> {
    const merged = { ...this.options, ...options };
    const browserbase_api_key = firstString(merged.browserbase_api_key, process.env.BROWSERBASE_API_KEY);
    if (!browserbase_api_key) {
      throw new Error("launch.mode=bb requires BROWSERBASE_API_KEY or launch.options.browserbase_api_key.");
    }

    const project_id = firstString(
      merged.project_id,
      merged.browserbase_project_id,
      process.env.BROWSERBASE_PROJECT_ID,
      process.env.BB_PROJECT_ID,
    );
    const base_url =
      firstString(merged.base_url, merged.browserbase_base_url, process.env.BROWSERBASE_BASE_URL) ??
      DEFAULT_BROWSERBASE_BASE_URL;
    const resume_session_id = firstString(merged.session_id, merged.resume_session_id);
    const keep_alive = firstBoolean(merged.keep_alive) ?? false;
    const close_session_on_close = firstBoolean(merged.close_session_on_close) ?? !keep_alive;

    let created_session = false;
    let session: BrowserbaseSession;
    if (resume_session_id) {
      session = await browserbaseRequest<BrowserbaseSession>({
        base_url,
        browserbase_api_key,
        method: "GET",
        pathname: `/v1/sessions/${resume_session_id}`,
      });
    } else {
      const session_create_params = objectValue(
        merged.session_create_params ?? merged.browserbase_session_create_params,
      );
      const browser_settings = {
        ...objectValue(session_create_params.browserSettings),
        ...objectValue(merged.browser_settings),
      };
      const user_metadata = {
        ...objectValue(session_create_params.userMetadata),
        ...objectValue(merged.user_metadata),
      };
      const extension_id = firstString(
        merged.extension_id,
        session_create_params.extensionId,
        objectValue(session_create_params.browserSettings).extensionId,
      );
      const body = {
        ...session_create_params,
        ...(project_id ? { projectId: project_id } : {}),
        ...(keep_alive ? { keepAlive: true } : {}),
        ...(firstString(merged.region, session_create_params.region)
          ? { region: firstString(merged.region, session_create_params.region) }
          : {}),
        ...(typeof merged.timeout === "number" ? { timeout: merged.timeout } : {}),
        ...(extension_id ? { extensionId: extension_id } : {}),
        browserSettings: {
          ...browser_settings,
          ...(extension_id ? { extensionId: extension_id } : {}),
          viewport: objectValue(browser_settings.viewport).width
            ? browser_settings.viewport
            : DEFAULT_BROWSERBASE_VIEWPORT,
        },
        userMetadata: {
          ...user_metadata,
          modcdp: "true",
        },
      };
      session = await browserbaseRequest<BrowserbaseSession>({
        base_url,
        browserbase_api_key,
        method: "POST",
        pathname: "/v1/sessions",
        body,
      });
      created_session = true;
    }

    if (!session.id || !session.connectUrl) {
      throw new Error("Browserbase session creation returned an unexpected shape.");
    }

    let closed = false;
    const close = async () => {
      if (closed) return;
      closed = true;
      if (!created_session || !close_session_on_close) return;
      await browserbaseRequest({
        base_url,
        browserbase_api_key,
        method: "POST",
        pathname: `/v1/sessions/${session.id}`,
        body: {
          status: "REQUEST_RELEASE",
          ...(project_id ? { projectId: project_id } : {}),
        },
      }).catch(() => {});
    };

    this.launched = {
      cdp_url: session.connectUrl,
      ws_url: session.connectUrl,
      browserbase_session_id: session.id,
      browserbase_session_url: `https://www.browserbase.com/sessions/${session.id}`,
      browserbase_debug_url: session.debuggerUrl ?? session.debuggerFullscreenUrl ?? null,
      close,
    };
    return this.launched;
  }
}
