import type { BrowserLaunchOptions } from "./BrowserLauncher.js";
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
    const configured_extension_id = firstString(this.options.extension_id);
    if (configured_extension_id) {
      this.extension_id = configured_extension_id;
      return;
    }
    const extension_path = this.options.extension_path;
    if (!extension_path) return;
    this.zip_path = extension_path.endsWith(".zip") ? extension_path : await this.zipExtensionDir(extension_path);
    this.extension_id = await this.uploadExtension(this.zip_path);
  }

  getLauncherConfig(): BrowserLaunchOptions {
    if (!this.extension_id) return {};
    return { extension_id: this.extension_id };
  }

  async inject() {
    const discovered = await this.waitForReadyServiceWorker(this.options.service_worker_ready_timeout_ms ?? 60_000, {
      matched_only: this.options.trust_matched_service_worker,
    });
    return discovered ? { ...discovered, source: "bb" } : null;
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
    const browserbase_api_key = firstString(this.options.browserbase_api_key, process.env.BROWSERBASE_API_KEY);
    if (!browserbase_api_key) {
      throw new Error("BBBrowserExtensionInjector requires BROWSERBASE_API_KEY or launch.options.browserbase_api_key.");
    }
    const base_url =
      firstString(this.options.base_url, this.options.browserbase_base_url, process.env.BROWSERBASE_BASE_URL) ??
      DEFAULT_BROWSERBASE_BASE_URL;
    const fs = await import("node:fs");
    const path = await import("node:path");
    const form = new FormData();
    form.append("file", new Blob([fs.readFileSync(zip_path)]), path.basename(zip_path));
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
