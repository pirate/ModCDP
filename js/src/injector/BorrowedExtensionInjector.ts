import { installModCDPServer } from "../server/ModCDPServer.js";
import { commands as RuntimeCommands } from "../types/generated/zod/Runtime.js";
import { ExtensionInjector, type ExtensionInjectionResult, type TargetInfo } from "./ExtensionInjector.js";

const EXT_ID_FROM_URL = /^chrome-extension:\/\/([a-z]+)\//;
const MODCDP_READY_EXPRESSION =
  "Boolean(globalThis.ModCDP?.__ModCDPServerVersion === 1 && globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)";
const bootstrap_modcdp_server_expression = `
  function() {
    const __name = (fn) => fn;
    const installModCDPServer = ${installModCDPServer.toString()};
    const ModCDP = installModCDPServer(globalThis);
    return {
      ok: Boolean(ModCDP?.__ModCDPServerVersion === 1 && ModCDP?.handleCommand && ModCDP?.addCustomEvent),
      extension_id: globalThis.chrome?.runtime?.id ?? null,
      has_tabs: Boolean(globalThis.chrome?.tabs?.query),
      has_debugger: Boolean(globalThis.chrome?.debugger?.sendCommand && globalThis.chrome?.debugger?.getTargets),
    };
  }
`;

export class BorrowedExtensionInjector extends ExtensionInjector {
  async inject() {
    const deadline = Date.now() + (this.options.injector_service_worker_ready_timeout_ms ?? 60_000);
    do {
      const borrowed = await this.borrowVisibleServiceWorkers();
      if (borrowed) return borrowed;
      await new Promise((resolve) => setTimeout(resolve, this.options.injector_service_worker_poll_interval_ms ?? 100));
    } while (Date.now() < deadline);
    return null;
  }

  private async borrowVisibleServiceWorkers() {
    const borrowed: ExtensionInjectionResult[] = [];
    const visible_service_workers = (await this.targetInfos()).filter((target) => {
      const target_url = target.url ?? "";
      return target.type === "service_worker" && target_url.startsWith("chrome-extension://");
    });
    const has_configured_matcher =
      Boolean(this.options.injector_extension_id) ||
      (this.options.injector_service_worker_url_includes?.length ?? 0) > 0 ||
      (this.options.injector_service_worker_url_suffixes?.length ?? 0) > 0;
    const candidates = has_configured_matcher
      ? visible_service_workers.filter((target) => this.serviceWorkerTargetMatches(target))
      : visible_service_workers;
    for (const target of candidates) {
      try {
        const bootstrapped = await this.bootstrapTarget(target as TargetInfo);
        if (bootstrapped) borrowed.push({ ...bootstrapped, source: "borrowed" });
      } catch {}
    }
    borrowed.sort((a, b) => Number(b.has_debugger) - Number(a.has_debugger) || Number(b.has_tabs) - Number(a.has_tabs));
    return borrowed[0] ?? null;
  }

  private async bootstrapTarget(target: TargetInfo): Promise<ExtensionInjectionResult | null> {
    const session_id = await this.ensureSessionIdForTarget(
      target.targetId,
      this.options.injector_service_worker_probe_timeout_ms,
      true,
    );
    if (session_id == null) return null;
    await this.sendWithTimeout("Runtime.enable", {}, session_id).catch(() => {});
    const bootstrap = RuntimeCommands["Runtime.evaluate"].result.parse(
      await this.sendWithTimeout(
        "Runtime.evaluate",
        {
          expression: `(${bootstrap_modcdp_server_expression})()`,
          awaitPromise: true,
          returnByValue: true,
        },
        session_id,
      ),
    );
    const value = bootstrap.result?.value || {};
    if (!value.has_tabs || !value.has_debugger) return null;
    let ready = Boolean(value.ok);
    if (ready && this.readyExpression() !== MODCDP_READY_EXPRESSION) {
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
      ready = probe.result?.value === true;
    }
    if (!ready) return null;
    return {
      source: "borrowed",
      extension_id: value.extension_id || target.url?.match(EXT_ID_FROM_URL)?.[1] || null,
      target_id: target.targetId,
      url: target.url,
      session_id,
      has_tabs: Boolean(value.has_tabs),
      has_debugger: Boolean(value.has_debugger),
    };
  }
}
