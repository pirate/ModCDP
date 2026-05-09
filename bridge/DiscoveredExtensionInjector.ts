import { ExtensionInjector } from "./ExtensionInjector.js";

export class DiscoveredExtensionInjector extends ExtensionInjector {
  async inject() {
    const discovered = await this.discoverReadyServiceWorker();
    if (discovered) return { ...discovered, source: "discovered" };
    if (this.options.trust_matched_service_worker) {
      const waited = await this.waitForReadyServiceWorker(this.options.service_worker_probe_timeout_ms ?? 10_000, {
        matched_only: true,
      });
      if (waited) return { ...waited, source: "discovered" };
    }
    const woke = await this.wakeConfiguredExtension();
    if (woke) {
      const waited = await this.waitForReadyServiceWorker(this.options.service_worker_probe_timeout_ms ?? 10_000, {
        matched_only: this.options.trust_matched_service_worker,
      });
      if (waited) return { ...waited, source: "discovered" };
    }
    if (!this.options.require_service_worker_target) return null;
    const waited = await this.waitForReadyServiceWorker(this.options.service_worker_ready_timeout_ms ?? 60_000, {
      matched_only: this.options.trust_matched_service_worker,
    });
    if (waited) return { ...waited, source: "discovered" };
    throw new Error(
      `Required ModCDP service worker target was not visible ` +
        `(${
          [
            ...(this.options.service_worker_url_includes ?? []),
            ...(this.options.service_worker_url_suffixes ?? []),
          ].join(", ") || "no matcher"
        }).`,
    );
  }
}
