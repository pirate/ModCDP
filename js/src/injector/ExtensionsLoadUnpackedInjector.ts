import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { ExtensionInjector, type TargetInfo } from "./ExtensionInjector.js";

export class ExtensionsLoadUnpackedInjector extends ExtensionInjector {
  private unpacked_extension_path: string | null = null;
  private cleanup: (() => Promise<void>) | null = null;

  async prepare() {
    const extension_path = this.options.injector_extension_path;
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
      await super.prepare();
      return;
    }
    const { execFileSync } = await import("node:child_process");
    const unpacked_path = fs.mkdtempSync(path.join(os.tmpdir(), "modcdp-extension-"));
    execFileSync("unzip", ["-q", extension_path, "-d", unpacked_path]);
    this.unpacked_extension_path = unpacked_path;
    this.cleanup = async () => fs.rmSync(unpacked_path, { recursive: true, force: true });
    await super.prepare();
  }

  async inject() {
    const extension_path = this.unpacked_extension_path;
    if (!extension_path) return null;
    let load_result: Record<string, unknown>;
    try {
      load_result = (await this.send("Extensions.loadUnpacked", { path: extension_path })) as Record<string, unknown>;
    } catch (error) {
      const load_error = error instanceof Error ? error : new Error(String(error));
      if (/Method not available|Method.*not.*found|wasn't found/i.test(load_error.message)) {
        this.last_error = load_error;
        return null;
      }
      throw new Error(
        `Extensions.loadUnpacked failed for ${extension_path}: ${load_error.message}\n` +
          `If the path is correct and the manifest is valid, load the ModCDP extension manually in chrome://extensions and reconnect.`,
      );
    }

    const extension_id = load_result.id || load_result.extensionId;
    if (typeof extension_id !== "string" || !extension_id) {
      throw new Error(`Extensions.loadUnpacked returned no extension id (got ${JSON.stringify(load_result)})`);
    }
    this.options.injector_extension_id = extension_id;
    await this.wakeConfiguredExtension();

    const sw_url_prefix = `chrome-extension://${extension_id}/`;
    const deadline = Date.now() + (this.options.injector_service_worker_ready_timeout_ms ?? 60_000);
    while (Date.now() < deadline) {
      const target_infos = await this.targetInfos();
      const target = target_infos.find(
        (candidate) => candidate.type === "service_worker" && candidate.url.startsWith(sw_url_prefix),
      ) as TargetInfo | undefined;
      if (target) {
        const probed = await this.probeTarget(target, this.options.injector_service_worker_probe_timeout_ms, {
          allow_attach: true,
        });
        if (probed) return { ...probed, source: "extensions_load_unpacked", extension_id };
      }
      await new Promise((resolve) => setTimeout(resolve, this.options.injector_service_worker_poll_interval_ms ?? 100));
    }
    throw new Error(`Timed out waiting for service worker target for extension ${extension_id}.`);
  }

  async close() {
    await super.close();
    await this.cleanup?.();
    this.cleanup = null;
  }
}
