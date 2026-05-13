import type { BrowserLaunchOptions } from "../launcher/BrowserLauncher.js";
import { ExtensionInjector } from "./ExtensionInjector.js";

const DEFAULT_BROWSERBASE_BASE_URL = "https://api.browserbase.com";

function firstString(...values: unknown[]) {
  for (const value of values) {
    if (typeof value === "string" && value.trim()) return value.trim();
  }
  return null;
}

export class BBBrowserExtensionInjector extends ExtensionInjector {
  private extension_id: string | null = null;
  private zip_path: string | null = null;
  private cleanup: (() => Promise<void>) | null = null;

  async prepare() {
    const configured_extension_id = firstString(this.options.injector_extension_id);
    if (configured_extension_id) {
      this.extension_id = configured_extension_id;
      return;
    }
    if (this.extension_id) return;
    const extension_path = this.options.injector_extension_path;
    if (!extension_path) return;
    this.zip_path = extension_path.endsWith(".zip") ? extension_path : await this.zipExtensionDir(extension_path);
    try {
      this.extension_id = await this.uploadExtension(this.zip_path);
    } catch (error) {
      await this.close();
      throw error;
    }
  }

  getLauncherConfig(): BrowserLaunchOptions {
    if (!this.extension_id) return {};
    return { injector_extension_id: this.extension_id };
  }

  async inject() {
    const extension_id = this.options.injector_extension_id;
    this.options.injector_extension_id = null;
    try {
      const discovered = await this.waitForReadyServiceWorker(
        this.options.injector_service_worker_ready_timeout_ms ?? 60_000,
        {
          matched_only: this.options.injector_trust_service_worker_target,
        },
      );
      return discovered ? { ...discovered, source: "bb" } : null;
    } finally {
      this.options.injector_extension_id = extension_id;
    }
  }

  async close() {
    await this.cleanup?.();
    this.cleanup = null;
  }

  private async zipExtensionDir(extension_path: string) {
    const [{ execFileSync }, fs, os, path] = await Promise.all([
      import("node:child_process"),
      import("node:fs"),
      import("node:os"),
      import("node:path"),
    ]);
    const zip_path = path.join(fs.mkdtempSync(path.join(os.tmpdir(), "modcdp-bb-extension-")), "extension.zip");
    execFileSync("zip", ["-X", "-qr", zip_path, "."], { cwd: extension_path });
    this.cleanup = async () => fs.rmSync(path.dirname(zip_path), { recursive: true, force: true });
    return zip_path;
  }

  private async uploadExtension(zip_path: string) {
    const browserbase_api_key = firstString(this.options.injector_browserbase_api_key, process.env.BROWSERBASE_API_KEY);
    if (!browserbase_api_key) {
      throw new Error(
        "BBBrowserExtensionInjector requires BROWSERBASE_API_KEY or launcher.launcher_options.browserbase_api_key.",
      );
    }
    const base_url =
      firstString(this.options.injector_browserbase_base_url, process.env.BROWSERBASE_BASE_URL) ??
      DEFAULT_BROWSERBASE_BASE_URL;
    const fs = await import("node:fs");
    const path = await import("node:path");
    const form = new FormData();
    const zip_bytes = fs.readFileSync(zip_path);
    const zip_array_buffer = zip_bytes.buffer.slice(
      zip_bytes.byteOffset,
      zip_bytes.byteOffset + zip_bytes.byteLength,
    ) as ArrayBuffer;
    form.append("file", new Blob([zip_array_buffer]), path.basename(zip_path));
    const response = await fetch(new URL("/v1/extensions", `${base_url.replace(/\/$/, "")}/`), {
      method: "POST",
      headers: { "X-BB-API-Key": browserbase_api_key },
      body: form,
    });
    if (!response.ok) {
      const text = await response.text().catch(() => "");
      throw new Error(`Browserbase POST /v1/extensions -> ${response.status}${text ? `: ${text}` : ""}`);
    }
    const extension = (await response.json()) as Record<string, unknown>;
    if (typeof extension.id !== "string" || !extension.id) {
      throw new Error(`Browserbase extension upload returned no id (got ${JSON.stringify(extension)})`);
    }
    return extension.id;
  }
}
