import type { BrowserLaunchOptions } from "./BrowserLauncher.js";
import { ExtensionInjector } from "./ExtensionInjector.js";

function firstString(...values: unknown[]) {
  for (const value of values) {
    if (typeof value === "string" && value.trim()) return value.trim();
  }
  return null;
}

export class LocalBrowserLaunchExtensionInjector extends ExtensionInjector {
  private unpacked_extension_path: string | null = null;
  private extension_id: string | null = null;
  private cleanup: (() => Promise<void>) | null = null;

  async prepare() {
    const extension_path = this.options.extension_path;
    if (!extension_path) {
      await super.prepare();
      return;
    }
    if (this.unpacked_extension_path) {
      await super.prepare();
      return;
    }
    if (!extension_path.endsWith(".zip")) {
      this.unpacked_extension_path = extension_path;
      await this.resolveExtensionId();
      await super.prepare();
      return;
    }
    const [{ execFileSync }, fs, os, path] = await Promise.all([
      import("node:child_process"),
      import("node:fs"),
      import("node:os"),
      import("node:path"),
    ]);
    const unpacked_path = fs.mkdtempSync(path.join(os.tmpdir(), "modcdp-extension-"));
    execFileSync("unzip", ["-q", extension_path, "-d", unpacked_path]);
    this.unpacked_extension_path = unpacked_path;
    this.cleanup = async () => fs.rmSync(unpacked_path, { recursive: true, force: true });
    await this.resolveExtensionId();
    await super.prepare();
  }

  getLauncherConfig(): BrowserLaunchOptions {
    const extension_path = this.unpacked_extension_path;
    if (!extension_path) return {};
    const wake_url = this.configuredWakeUrl();
    return { extra_args: [`--load-extension=${extension_path}`, ...(wake_url ? [wake_url] : [])] };
  }

  async inject() {
    const timeout_ms = Math.min(this.options.service_worker_probe_timeout_ms ?? 10_000, 3_000);
    const discovered = await this.waitForReadyServiceWorker(timeout_ms, {
      matched_only: this.options.trust_matched_service_worker,
    });
    return discovered ? { ...discovered, source: "local_launch" } : null;
  }

  async close() {
    await super.close();
    await this.cleanup?.();
    this.cleanup = null;
  }

  private async resolveExtensionId() {
    if (this.extension_id) return this.extension_id;
    this.extension_id = firstString(this.options.extension_id);
    if (!this.extension_id && this.unpacked_extension_path) {
      this.extension_id = await extensionIdFromManifestKey(this.unpacked_extension_path);
    }
    if (this.extension_id) this.options.extension_id = this.extension_id;
    return this.extension_id;
  }
}

async function extensionIdFromManifestKey(extension_path: string) {
  const [crypto, fs, path] = await Promise.all([import("node:crypto"), import("node:fs"), import("node:path")]);
  const manifest_path = path.join(extension_path, "manifest.json");
  if (!fs.existsSync(manifest_path)) return null;
  const manifest = JSON.parse(fs.readFileSync(manifest_path, "utf8")) as Record<string, unknown>;
  const key = firstString(manifest.key);
  if (!key) return null;
  const digest = crypto.createHash("sha256").update(Buffer.from(key, "base64")).digest().subarray(0, 16);
  const alphabet = "abcdefghijklmnop";
  return [...digest].map((byte) => alphabet[byte >> 4] + alphabet[byte & 0x0f]).join("");
}
