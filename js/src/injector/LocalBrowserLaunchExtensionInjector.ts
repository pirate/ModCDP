import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import type { BrowserLaunchOptions } from "../launcher/BrowserLauncher.js";
import { defaultModCDPExtensionPath, ExtensionInjector } from "./ExtensionInjector.js";

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
    const extension_path = this.options.injector_extension_path ?? defaultModCDPExtensionPath();
    if (this.unpacked_extension_path) {
      await super.prepare();
      return;
    }
    if (!extension_path.endsWith(".zip")) {
      const unpacked_path = fs.mkdtempSync(path.join(os.tmpdir(), "modcdp-extension-"));
      fs.cpSync(extension_path, unpacked_path, { recursive: true });
      this.unpacked_extension_path = extensionRoot(unpacked_path);
      this.cleanup = async () => fs.rmSync(unpacked_path, { recursive: true, force: true });
      await this.resolveExtensionId();
      await super.prepare();
      return;
    }
    const { execFileSync } = await import("node:child_process");
    const unpacked_path = fs.mkdtempSync(path.join(os.tmpdir(), "modcdp-extension-"));
    execFileSync("unzip", ["-q", extension_path, "-d", unpacked_path]);
    this.unpacked_extension_path = extensionRoot(unpacked_path);
    this.cleanup = async () => fs.rmSync(unpacked_path, { recursive: true, force: true });
    await this.resolveExtensionId();
    await super.prepare();
  }

  getLauncherConfig(): BrowserLaunchOptions {
    const extension_path = this.unpacked_extension_path;
    if (!extension_path) return {};
    return { extra_args: [`--load-extension=${extension_path}`] };
  }

  async inject() {
    const timeout_ms = this.options.injector_service_worker_probe_timeout_ms ?? 10_000;
    const discovered = await this.waitForReadyServiceWorker(timeout_ms, {
      matched_only: this.options.injector_trust_service_worker_target,
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
    this.extension_id = firstString(this.options.injector_extension_id);
    if (!this.extension_id && this.unpacked_extension_path) {
      this.extension_id = await extensionIdFromManifestKey(this.unpacked_extension_path);
    }
    if (this.extension_id) this.options.injector_extension_id = this.extension_id;
    return this.extension_id;
  }
}

function extensionRoot(unpacked_path: string) {
  if (fs.existsSync(path.join(unpacked_path, "manifest.json"))) return unpacked_path;
  const nested_path = path.join(unpacked_path, "extension");
  if (fs.existsSync(path.join(nested_path, "manifest.json"))) return nested_path;
  return unpacked_path;
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
