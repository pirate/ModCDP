// injector.js: inject the ModCDP extension service worker when needed in a
// running Chrome and return a CDP session attached to it.
//
// The caller hands in a `send(method, params, session_id?)` function bound to
// the upstream CDP websocket. The injector knows about Extensions.loadUnpacked,
// service-worker URL pattern matching, and probe-by-globalThis.ModCDP, but
// nothing about chrome binaries, the proxy, or wrap/unwrap.
//
// Precedence (single source of truth — do not duplicate this in proxy/client):
//   1. Look for an existing service-worker target whose JS context already has
//      globalThis.ModCDP. Use it. (source: "discovered")
//   2. Otherwise call Extensions.loadUnpacked(extension_path) and wait for that
//      extension's service worker to appear. (source: "injected")
//   3. If Chrome refuses extension loading, bootstrap ModCDP into every
//      already-running extension service worker target and use the best one.
//      (source: "borrowed")
//   4. Otherwise throw with explicit instructions for all failure modes.

import type { ProtocolParams, ProtocolResult } from "../types/modcdp.js";
import { commands as RuntimeCommands } from "../types/zod/Runtime.js";
import { commands as TargetCommands } from "../types/zod/Target.js";
import { installModCDPServer } from "../extension/ModCDPServer.js";

const EXT_ID_FROM_URL = /^chrome-extension:\/\/([a-z]+)\//;
const MODCDP_READY_EXPRESSION =
  "Boolean(globalThis.ModCDP?.__ModCDPServerVersion === 1 && globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)";
export const DEFAULT_CDP_SEND_TIMEOUT_MS = 10_000;
export const DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS = 10_000;
export const DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS = 10_000;
export const DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS = 60_000;
export const DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS = 100;
export const DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS = 20;

type SendCDP = (method: string, params?: ProtocolParams, session_id?: string | null) => Promise<ProtocolResult>;
type TargetInfo = { targetId: string; type?: string; url?: string };

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

export async function injectExtensionIfNeeded({
  send,
  session_id_for_target = null,
  attach_to_target = null,
  wait_for_execution_context = null,
  extension_path,
  service_worker_url_includes = [],
  service_worker_url_suffixes = [],
  trust_matched_service_worker = false,
  require_service_worker_target = false,
  service_worker_ready_expression = null,
  cdp_send_timeout_ms = DEFAULT_CDP_SEND_TIMEOUT_MS,
  execution_context_timeout_ms = DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS,
  service_worker_probe_timeout_ms = DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
  service_worker_ready_timeout_ms = DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS,
  service_worker_poll_interval_ms = DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS,
  target_session_poll_interval_ms = DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS,
}: {
  send: SendCDP;
  session_id_for_target?: ((target_id: string) => string | null | undefined) | null;
  attach_to_target?: ((target_id: string) => Promise<string | null | undefined>) | null;
  wait_for_execution_context?: ((session_id: string, timeout_ms: number) => Promise<number>) | null;
  extension_path?: string | null;
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
}) {
  if (typeof send !== "function") throw new Error("injectExtensionIfNeeded requires { send }");
  const ready_expression =
    service_worker_ready_expression == null || service_worker_ready_expression.length === 0
      ? MODCDP_READY_EXPRESSION
      : `(${MODCDP_READY_EXPRESSION}) && Boolean(${service_worker_ready_expression})`;
  const sendWithTimeout = (
    method: string,
    params: ProtocolParams = {},
    session_id: string | null = null,
    ms = cdp_send_timeout_ms,
  ) => {
    let timeout: ReturnType<typeof setTimeout> | null = null;
    return Promise.race([
      send(method, params, session_id),
      new Promise<never>((_, reject) => {
        timeout = setTimeout(() => reject(new Error(`${method} timed out after ${ms}ms`)), ms);
      }),
    ]).finally(() => {
      if (timeout != null) clearTimeout(timeout);
    });
  };
  // extension_path is only required as a fallback, when discovery does not turn
  // up an already-loaded ModCDP service worker. Validate at the point of use
  // (step 2) so callers running against a browser that already has the
  // extension loaded don't have to provide a path at all.

  const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));
  const bootstrapped_target_ids = new Set<string>();
  const unusable_target_ids = new Set<string>();
  const sessionIdForTarget = async (target_id: string, timeout_ms = 0) => {
    const deadline = Date.now() + timeout_ms;
    while (true) {
      const session_id = session_id_for_target?.(target_id);
      if (typeof session_id === "string" && session_id.length > 0) return session_id;
      if (Date.now() >= deadline) return null;
      await sleep(target_session_poll_interval_ms);
    }
  };
  const ensureSessionIdForTarget = async (target_id: string, timeout_ms = 0, allow_attach = false) => {
    const session_id = session_id_for_target?.(target_id);
    if (typeof session_id === "string" && session_id.length > 0) return session_id;
    if (allow_attach) {
      const attached_session_id = await attach_to_target?.(target_id);
      if (typeof attached_session_id === "string" && attached_session_id.length > 0) return attached_session_id;
    }
    return await sessionIdForTarget(target_id, timeout_ms);
  };
  const probeTarget = async (
    target: TargetInfo,
    session_timeout_ms = 0,
    { allow_attach = false }: { allow_attach?: boolean } = {},
  ) => {
    if (unusable_target_ids.has(target.targetId)) return null;
    const session_id = await ensureSessionIdForTarget(target.targetId, session_timeout_ms, allow_attach);
    if (session_id == null) return null;
    await sendWithTimeout("Runtime.enable", {}, session_id, cdp_send_timeout_ms);
    const probe = RuntimeCommands["Runtime.evaluate"].result.parse(
      await sendWithTimeout(
        "Runtime.evaluate",
        {
          expression: ready_expression,
          returnByValue: true,
        },
        session_id,
      ),
    );
    if (probe.result?.value !== true) return null;
    return {
      extension_id: target.url?.match(EXT_ID_FROM_URL)?.[1],
      target_id: target.targetId,
      url: target.url,
      session_id,
    };
  };
  const bootstrapTarget = async (target: TargetInfo) => {
    if (bootstrapped_target_ids.has(target.targetId)) return null;
    bootstrapped_target_ids.add(target.targetId);
    const session_id = await ensureSessionIdForTarget(target.targetId, service_worker_probe_timeout_ms, true);
    if (session_id == null) return null;
    await sendWithTimeout("Runtime.enable", {}, session_id, cdp_send_timeout_ms).catch(() => {});
    const bootstrap = RuntimeCommands["Runtime.evaluate"].result.parse(
      await sendWithTimeout(
        "Runtime.evaluate",
        {
          expression: `(${bootstrap_modcdp_server_expression})()`,
          awaitPromise: true,
          returnByValue: true,
        },
        session_id,
        cdp_send_timeout_ms,
      ),
    );
    const value = bootstrap.result?.value || {};
    if (!value.has_tabs || !value.has_debugger) {
      unusable_target_ids.add(target.targetId);
      return null;
    }
    let ready = Boolean(value.ok);
    if (ready && ready_expression !== MODCDP_READY_EXPRESSION) {
      const probe = RuntimeCommands["Runtime.evaluate"].result.parse(
        await sendWithTimeout(
          "Runtime.evaluate",
          {
            expression: ready_expression,
            returnByValue: true,
          },
          session_id,
          cdp_send_timeout_ms,
        ),
      );
      ready = probe.result?.value === true;
    }
    if (!ready) return null;
    return {
      extension_id: value.extension_id || target.url?.match(EXT_ID_FROM_URL)?.[1] || null,
      target_id: target.targetId,
      url: target.url,
      session_id,
      has_tabs: Boolean(value.has_tabs),
      has_debugger: Boolean(value.has_debugger),
    };
  };
  const discoverReadyServiceWorker = async ({ matched_only = false }: { matched_only?: boolean } = {}) => {
    const target_infos = TargetCommands["Target.getTargets"].result.parse(await send("Target.getTargets")).targetInfos;
    if (trust_matched_service_worker) {
      const trusted_target = target_infos.find((candidate) => serviceWorkerTargetMatches(candidate)) as
        | TargetInfo
        | undefined;
      if (trusted_target) {
        const probed = await probeTarget(trusted_target, service_worker_probe_timeout_ms, { allow_attach: true });
        if (probed) return { source: "trusted", ...probed };
        const bootstrapped = await bootstrapTarget(trusted_target);
        if (bootstrapped) return { source: "trusted", ...bootstrapped };
      }
    }
    if (trust_matched_service_worker || matched_only) return null;
    for (const candidate of target_infos) {
      if (candidate.type !== "service_worker") continue;
      if (!candidate.url.startsWith("chrome-extension://")) continue;
      try {
        const probed = await probeTarget(candidate as TargetInfo, service_worker_probe_timeout_ms);
        if (probed) return { source: "discovered", ...probed };
      } catch {
        continue;
      }
    }
    return null;
  };
  const waitForReadyServiceWorker = async (
    timeout_ms: number,
    { matched_only = false }: { matched_only?: boolean } = {},
  ) => {
    const deadline = Date.now() + timeout_ms;
    while (Date.now() < deadline) {
      const discovered = await discoverReadyServiceWorker({ matched_only });
      if (discovered) return discovered;
      await sleep(service_worker_poll_interval_ms);
    }
    return null;
  };

  // 1. Discover an existing ModCDP service worker from the current CDP target
  // snapshot. If no already-ready worker is visible, move on to the explicit
  // injection path instead of waiting on a guessed preinstalled-extension budget.
  const discovered = await discoverReadyServiceWorker();
  if (discovered) return discovered;
  if (require_service_worker_target) {
    const discovered = await waitForReadyServiceWorker(service_worker_ready_timeout_ms, {
      matched_only: trust_matched_service_worker,
    });
    if (discovered) return discovered;
    throw new Error(
      `Required ModCDP service worker target was not visible in the current CDP target snapshot ` +
        `(${[...service_worker_url_includes, ...service_worker_url_suffixes].join(", ") || "no matcher"}).`,
    );
  }

  // 2. Try Extensions.loadUnpacked.
  let load_unpacked_unavailable_error: Error | null = null;
  if (!extension_path) {
    load_unpacked_unavailable_error = new Error("No extension_path was provided.");
  } else {
    let load_result;
    try {
      load_result = await send("Extensions.loadUnpacked", { path: extension_path });
    } catch (error) {
      const load_error = error instanceof Error ? error : new Error(String(error));
      if (/Method not available|Method.*not.*found|wasn't found/i.test(load_error.message)) {
        load_unpacked_unavailable_error = load_error;
        const discovered = await waitForReadyServiceWorker(service_worker_ready_timeout_ms, {
          matched_only: trust_matched_service_worker,
        });
        if (discovered) return discovered;
      } else {
        throw new Error(
          `Extensions.loadUnpacked failed for ${extension_path}: ${load_error.message}\n` +
            `If the path is correct and the manifest is valid, load the ModCDP extension manually in chrome://extensions and reconnect.`,
        );
      }
    }

    if (!load_unpacked_unavailable_error) {
      const extension_id = load_result?.id || load_result?.extensionId;
      if (!extension_id) {
        throw new Error(`Extensions.loadUnpacked returned no extension id (got ${JSON.stringify(load_result)})`);
      }

      // 3. Wait for the loaded extension's service worker target. Custom extensions
      // can name the worker bundle anything; WXT uses background.js.
      const sw_url_prefix = `chrome-extension://${extension_id}/`;
      const deadline = Date.now() + service_worker_ready_timeout_ms;
      while (Date.now() < deadline) {
        const target_infos = TargetCommands["Target.getTargets"].result.parse(
          await send("Target.getTargets"),
        ).targetInfos;
        const target = target_infos.find(
          (candidate) => candidate.type === "service_worker" && candidate.url.startsWith(sw_url_prefix),
        ) as TargetInfo | undefined;
        if (target) {
          const probed = await probeTarget(target, service_worker_probe_timeout_ms, { allow_attach: true });
          if (probed)
            return {
              source: "injected",
              extension_id,
              target_id: target.targetId,
              url: target.url,
              session_id: probed.session_id,
            };
        }
        await sleep(service_worker_poll_interval_ms);
      }
      throw new Error(
        `Timed out after ${service_worker_ready_timeout_ms}ms waiting for service worker target for extension ${extension_id}.`,
      );
    }
  }

  // 4. Chrome's new chrome://inspect auto-connect flow exposes CDP without
  // exposing Extensions.loadUnpacked. In that case, inject the same server into
  // every currently running extension service worker and keep the best session.
  const borrowed: {
    target_id: string;
    url: string;
    session_id: string;
    extension_id?: string | null;
    has_tabs?: boolean;
    has_debugger?: boolean;
  }[] = [];
  const borrowed_target_infos = TargetCommands["Target.getTargets"].result.parse(
    await send("Target.getTargets"),
  ).targetInfos;
  for (const target of borrowed_target_infos) {
    if (target.type !== "service_worker") continue;
    if (!target.url.startsWith("chrome-extension://")) continue;

    let session_id: string | null = null;
    try {
      const bootstrapped = await bootstrapTarget(target as TargetInfo);
      if (bootstrapped) {
        session_id = bootstrapped.session_id;
        borrowed.push({
          target_id: target.targetId,
          url: target.url,
          session_id,
          extension_id: bootstrapped.extension_id,
          has_tabs: bootstrapped.has_tabs,
          has_debugger: bootstrapped.has_debugger,
        });
      }
    } catch {}
  }

  borrowed.sort((a, b) => Number(b.has_debugger) - Number(a.has_debugger) || Number(b.has_tabs) - Number(a.has_tabs));
  const selected = borrowed[0];
  if (selected) {
    return {
      source: "borrowed",
      extension_id: selected.extension_id,
      target_id: selected.target_id,
      url: selected.url,
      session_id: selected.session_id,
    };
  }

  throw new Error(
    `Cannot install or borrow ModCDP in the running browser.\n\n` +
      `  - No existing service worker with globalThis.ModCDP was found in the browser.\n` +
      `  - Extensions.loadUnpacked is unavailable ("${load_unpacked_unavailable_error.message}").\n` +
      `  - No running chrome-extension:// service worker target accepted the ModCDP bootstrap.\n\n` +
      `Fixes (any one of these):\n` +
      `  1. Open or wake an installed extension that has a service worker, then reconnect.\n` +
      `  2. Load the ModCDP extension once at chrome://extensions and reconnect.\n` +
      (extension_path ? `  3. For automated/test browsers, relaunch with --load-extension=${extension_path}.\n` : ""),
  );

  function serviceWorkerTargetMatches(candidate: { type?: string; url?: string }) {
    const url = candidate.url ?? "";
    if (candidate.type !== "service_worker") return false;
    if (!url.startsWith("chrome-extension://")) return false;
    if (service_worker_url_includes.length > 0 && !service_worker_url_includes.every((part) => url.includes(part))) {
      return false;
    }
    if (service_worker_url_suffixes.length > 0 && !service_worker_url_suffixes.some((suffix) => url.endsWith(suffix))) {
      return false;
    }
    return service_worker_url_includes.length > 0 || service_worker_url_suffixes.length > 0;
  }
}
