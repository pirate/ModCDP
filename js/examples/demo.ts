// JS demo for ModCDPClient with --direct / --loopback / --debugger route modes
// and --upstream=ws|pipe|reversews|nativemessaging|nats transport modes.
//
// Modes select where non-ModCDP commands ultimately get serviced:
//   --live        use the running Google Chrome enabled via chrome://inspect.
//   --direct      client sends standard CDP straight to the upstream WS.
//   --loopback    client routes *.* through the extension service worker,
//                 which opens a verified WebSocket back to localhost:9222 and
//                 forwards the command. (*.* -> service_worker on client,
//                 *.* -> loopback_cdp on server. Default mode.)
//   --debugger    client routes *.* through the extension service worker,
//                 which uses chrome.debugger.sendCommand against the active
//                 tab. (*.* -> service_worker on client, *.* -> chrome_debugger
//                 on server.)
//
//   --upstream    select the browser/upstream transport. Defaults to ws.
//                 reversews and nativemessaging use the fixed extension
//                 defaults: ws://127.0.0.1:29292 and com.modcdp.bridge.
//
// Valid CI/local demo combinations exercise the same surface: raw Browser.getVersion, raw
// Target.targetCreated event handling, Mod.evaluate, Custom.* commands,
// Custom.* events, and response middleware.

import path from "node:path";
import { existsSync } from "node:fs";
import { readFile, stat } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { setTimeout as sleep } from "node:timers/promises";
import { createInterface } from "node:readline/promises";
import { spawn } from "node:child_process";
import { z } from "zod";

import { ModCDPClient } from "../src/client/ModCDPClient.js";

type TargetCreatedPayload = {
  targetInfo?: { targetId?: string } & Record<string, unknown>;
};

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH =
  [path.resolve(HERE, "..", "..", "extension"), path.resolve(HERE, "..", "..", "dist", "extension")].find((candidate) =>
    existsSync(path.join(candidate, "modcdp/service_worker.js")),
  ) ?? path.resolve(HERE, "..", "..", "extension");
const DEFAULT_DEMO_EVENT_TIMEOUT_MS = 10_000;
const DEFAULT_REVERSE_TRANSPORT_WAIT_TIMEOUT_MS = 60_000;
const DEFAULT_LIVE_CDP_POLL_INTERVAL_MS = 250;
const DEFAULT_LIVE_CDP_ACTIVE_PORT_STALE_MS = 1_000;
const DEFAULT_TARGET_EVENT_TIMEOUT_MS = 10_000;
const DEFAULT_PAGE_TARGET_EVENT_TIMEOUT_MS = 10_000;
const DEFAULT_DEMO_EVENT_POLL_INTERVAL_MS = 20;

const UPSTREAM_MODES = new Set(["ws", "pipe", "reversews", "nativemessaging", "nats"]);

function parseArgs(argv) {
  const flags = new Set(argv.filter((a) => a.startsWith("--")).map((a) => a.slice(2)));
  const upstream_arg = argv.find((a) => a === "--upstream" || a.startsWith("--upstream="));
  const upstream_value =
    upstream_arg === "--upstream"
      ? argv[argv.indexOf(upstream_arg) + 1]
      : upstream_arg?.startsWith("--upstream=")
        ? upstream_arg.slice("--upstream=".length)
        : [...UPSTREAM_MODES].find((mode) => flags.has(mode));
  const upstream_mode = upstream_value || "ws";
  if (!UPSTREAM_MODES.has(upstream_mode)) {
    throw new Error(`unknown --upstream=${upstream_mode}; expected ${[...UPSTREAM_MODES].join("|")}`);
  }
  const live = flags.has("live");
  const mode = flags.has("debugger")
    ? "debugger"
    : flags.has("direct")
      ? "direct"
      : flags.has("loopback")
        ? "loopback"
        : live
          ? "direct"
          : "loopback";
  if (live && upstream_mode === "pipe") {
    throw new Error(
      "--live cannot be combined with --upstream=pipe because pipe handles only exist for launched browsers.",
    );
  }
  if (
    mode === "direct" &&
    (upstream_mode === "reversews" || upstream_mode === "nativemessaging" || upstream_mode === "nats")
  ) {
    throw new Error(
      `--direct cannot be combined with --upstream=${upstream_mode}; reverse transports terminate at ModCDPServer.`,
    );
  }
  return { mode, live, upstream_mode };
}

function serverRoutesFor(mode, upstream_mode) {
  const routes = {
    "Mod.*": "service_worker",
    "Custom.*": "service_worker",
    "*.*": mode === "loopback" ? "loopback_cdp" : mode === "debugger" ? "chrome_debugger" : "auto",
  };
  if (mode === "loopback" || ["reversews", "nativemessaging", "nats"].includes(upstream_mode)) {
    routes["Target.setDiscoverTargets"] = "loopback_cdp";
    routes["Target.createTarget"] = "loopback_cdp";
    routes["Target.activateTarget"] = "loopback_cdp";
  }
  return routes;
}

function clientRoutesFor(mode) {
  const directNormalEventRoutes = {
    "Target.setDiscoverTargets": "direct_cdp",
    "Target.createTarget": "direct_cdp",
    "Target.activateTarget": "direct_cdp",
  };
  return {
    "Mod.*": "service_worker",
    "Custom.*": "service_worker",
    "*.*": mode === "direct" ? "direct_cdp" : "service_worker",
    ...directNormalEventRoutes,
  };
}

function clientOptionsFor(mode, upstream_mode, cdp_url, launch_options = {}) {
  const launcher = cdp_url
    ? ({ launcher_mode: "remote" } as const)
    : ({ launcher_mode: "local", launcher_options: launch_options } as const);
  const upstream = {
    upstream_mode,
    upstream_cdp_url: cdp_url,
    ...(upstream_mode === "reversews"
      ? { upstream_reversews_wait_timeout_ms: DEFAULT_REVERSE_TRANSPORT_WAIT_TIMEOUT_MS }
      : {}),
    ...(upstream_mode === "nativemessaging"
      ? { upstream_nativemessaging_wait_timeout_ms: DEFAULT_REVERSE_TRANSPORT_WAIT_TIMEOUT_MS }
      : {}),
    ...(upstream_mode === "nats" ? { upstream_nats_wait_timeout_ms: DEFAULT_REVERSE_TRANSPORT_WAIT_TIMEOUT_MS } : {}),
  };
  const injector = {
    injector_mode: "auto" as const,
    injector_extension_path: EXTENSION_PATH,
    injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
  };
  if (mode === "direct") {
    return {
      launcher,
      upstream,
      injector,
      client: {
        client_routes: clientRoutesFor(mode),
      },
    };
  }
  return {
    launcher,
    upstream,
    injector,
    client: {
      client_routes: clientRoutesFor(mode),
    },
    server: {
      server_routes: serverRoutesFor(mode, upstream_mode),
    },
  };
}

function assertObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} returned non-object value ${JSON.stringify(value)}`);
  }
  return value;
}

function isTargetCreatedPayload(value: unknown): value is TargetCreatedPayload {
  if (value == null || typeof value !== "object" || Array.isArray(value)) return false;
  const targetInfo = (value as Record<string, unknown>).targetInfo;
  return targetInfo == null || (typeof targetInfo === "object" && !Array.isArray(targetInfo));
}

async function waitForEvent(cdp, eventName, predicate = (_payload) => true, timeoutMs = DEFAULT_DEMO_EVENT_TIMEOUT_MS) {
  return await new Promise((resolve, reject) => {
    const timeout = setTimeout(() => {
      cdp.off(eventName, onEvent);
      reject(new Error(`timed out waiting for ${eventName}`));
    }, timeoutMs);
    const onEvent = (payload) => {
      if (!predicate(payload)) return;
      clearTimeout(timeout);
      cdp.off(eventName, onEvent);
      resolve(payload);
    };
    cdp.on(eventName, onEvent);
  });
}

function openLiveInspectPage() {
  if (process.platform === "darwin") {
    spawn("open", ["chrome://inspect/#remote-debugging"], {
      detached: true,
      stdio: "ignore",
    }).unref();
  } else {
    spawn("xdg-open", ["chrome://inspect/#remote-debugging"], {
      detached: true,
      stdio: "ignore",
    }).unref();
  }
}

async function waitForLiveCdpUrl() {
  const startedAt = Date.now();
  openLiveInspectPage();
  console.log("opened chrome://inspect/#remote-debugging");
  console.log("waiting for Chrome to expose DevToolsActivePort; click Allow when Chrome asks.");

  const candidates =
    process.platform === "darwin"
      ? [
          path.join(process.env.HOME || "", "Library/Application Support/Google/Chrome/DevToolsActivePort"),
          path.join(process.env.HOME || "", "Library/Application Support/Google/Chrome Beta/DevToolsActivePort"),
        ]
      : [
          path.join(process.env.HOME || "", ".config/google-chrome/DevToolsActivePort"),
          path.join(process.env.HOME || "", ".config/chromium/DevToolsActivePort"),
        ];
  while (true) {
    for (const file of candidates) {
      try {
        const info = await stat(file);
        if (info.mtimeMs < startedAt - DEFAULT_LIVE_CDP_ACTIVE_PORT_STALE_MS) continue;
        const [port, browserPath] = (await readFile(file, "utf8"))
          .trim()
          .split(/\n/)
          .map((line) => line.trim());
        if (port && browserPath) return `ws://127.0.0.1:${port}${browserPath}`;
      } catch {}
    }
    await sleep(DEFAULT_LIVE_CDP_POLL_INTERVAL_MS);
  }
}

async function main() {
  const { mode, live, upstream_mode } = parseArgs(process.argv.slice(2));
  console.log(`== mode: ${live ? "live/" : ""}${mode}; upstream: ${upstream_mode} ==`);
  if (!existsSync(path.join(EXTENSION_PATH, "modcdp/service_worker.js"))) {
    throw new Error(`Built extension not found at ${EXTENSION_PATH}. Run pnpm run build first.`);
  }

  let cdp_url;
  let launch_options = {};
  if (live) {
    cdp_url = await waitForLiveCdpUrl();
  } else {
    cdp_url = null;
    launch_options = {
      chrome_ready_timeout_ms: 60_000,
      headless: process.platform === "linux" && !process.env.DISPLAY,
      sandbox: process.platform !== "linux",
    };
  }

  const cdp = new ModCDPClient(clientOptionsFor(mode, upstream_mode, cdp_url, launch_options));
  const pageTargetEvents = [];
  const targetCreatedEvents: TargetCreatedPayload[] = [];

  try {
    await cdp.connect();
    console.log("upstream cdp:", cdp.cdp_url);
    cdp.on(cdp.Target.targetCreated, (payload) => {
      const event = isTargetCreatedPayload(payload) ? payload : {};
      console.log("Target.targetCreated ->", event.targetInfo?.targetId);
      targetCreatedEvents.push(event);
    });
    console.log("connected; ext", cdp.extension_id, "session", cdp.ext_session_id);
    console.log("connect timing    ->", cdp.connect_timing);

    const configureResult = assertObject(
      await cdp.Mod.configure({
        upstream: {
          upstream_mode,
        },
        client: {
          client_routes: clientRoutesFor(mode),
        },
        server: {
          server_routes: serverRoutesFor(mode, upstream_mode),
        },
      }),
      "Mod.configure",
    );
    if (configureResult.routes?.["*.*"] !== serverRoutesFor(mode, upstream_mode)["*.*"]) {
      throw new Error(`unexpected Mod.configure result ${JSON.stringify(configureResult)}`);
    }
    console.log("Mod.configure    ->", configureResult.routes);

    const ping_sent_at = Date.now();
    const pongPromise = waitForEvent(cdp, "Mod.pong", (event) => event?.sent_at === ping_sent_at);
    const ping_result = assertObject(await cdp.Mod.ping({ sent_at: ping_sent_at }), "Mod.ping");
    const pong = assertObject(await pongPromise, "Mod.pong");
    const ping_returned_at = Date.now();
    if (ping_result.ok !== true || pong.received_at == null || pong.from !== "extension-service-worker") {
      throw new Error(`unexpected Mod.ping/Mod.pong result ${JSON.stringify({ ping_result, pong })}`);
    }
    console.log("Mod.ping/pong    ->", { ping_result, pong });
    console.log("ping latency      ->", {
      round_trip_ms: ping_returned_at - ping_sent_at,
      service_worker_ms: typeof pong.received_at === "number" ? pong.received_at - ping_sent_at : null,
      return_path_ms: typeof pong.received_at === "number" ? ping_returned_at - pong.received_at : null,
    });

    // Browser.getVersion is browser-scoped and chrome.debugger is tab-scoped,
    // so debugger mode asserts a positive raw CDP Runtime.evaluate instead.
    if (mode === "debugger") {
      try {
        const version = assertObject(await cdp.Browser.getVersion(), "Browser.getVersion");
        if (typeof version.protocolVersion !== "string" || typeof version.product !== "string") {
          throw new Error(`unexpected Browser.getVersion result ${JSON.stringify(version)}`);
        }
        console.log("Browser.getVersion ->", version);
      } catch (e) {
        console.log("Browser.getVersion -> (debugger route rejected:", e.message.replace(/\n/g, " "), ")");
      }
      const runtimeEval = assertObject(
        await cdp.Runtime.evaluate({
          expression: "(() => 42)()",
          returnByValue: true,
        }),
        "Runtime.evaluate",
      );
      if (runtimeEval.result?.value !== 42)
        throw new Error(`unexpected Runtime.evaluate result ${JSON.stringify(runtimeEval)}`);
      console.log("Runtime.evaluate ->", runtimeEval);
    } else {
      const version = assertObject(await cdp.Browser.getVersion(), "Browser.getVersion");
      if (typeof version.protocolVersion !== "string" || typeof version.product !== "string") {
        throw new Error(`unexpected Browser.getVersion result ${JSON.stringify(version)}`);
      }
      console.log("Browser.getVersion ->", version);
    }

    const modcdpEval = (await cdp.Mod.evaluate({
      expression: "({ extension_id: chrome.runtime.id })",
    })) as {
      extension_id?: string;
    };
    if (
      typeof modcdpEval.extension_id !== "string" ||
      (cdp.extension_id && modcdpEval.extension_id !== cdp.extension_id)
    )
      throw new Error(`unexpected Mod.evaluate result ${JSON.stringify(modcdpEval)}`);
    console.log("Mod.evaluate     ->", modcdpEval);

    const echoRegistration = assertObject(
      await cdp.Mod.addCustomCommand({
        name: "Custom.echo",
        expression: `async (params, method) => ({ echoed: params.value, method })`,
      }),
      "Mod.addCustomCommand Custom.echo",
    );
    if (echoRegistration.registered !== true || echoRegistration.name !== "Custom.echo") {
      throw new Error(`unexpected Custom.echo registration ${JSON.stringify(echoRegistration)}`);
    }
    const echoResult = assertObject(await cdp.send("Custom.echo", { value: "custom-command-ok" }), "Custom.echo");
    if (echoResult.echoed !== "custom-command-ok" || echoResult.method !== "Custom.echo") {
      throw new Error(`unexpected Custom.echo result ${JSON.stringify(echoResult)}`);
    }
    console.log("Custom.echo      ->", echoResult);

    const tabCommandRegistration = assertObject(
      await cdp.Mod.addCustomCommand({
        name: "Custom.TabIdFromTargetId",
        params_schema: {
          targetId: cdp.types.zod.Target.TargetID,
        },
        result_schema: {
          tabId: z.number().nullable(),
        },
        expression: `async ({ targetId }) => {
        const targets = await chrome.debugger.getTargets();
        const target = targets.find(target => target.id === targetId);
        return { tabId: target?.tabId ?? null };
      }`,
      }),
      "Mod.addCustomCommand Custom.TabIdFromTargetId",
    );
    if (tabCommandRegistration.registered !== true) {
      throw new Error(`unexpected TabIdFromTargetId registration ${JSON.stringify(tabCommandRegistration)}`);
    }
    const targetCommandRegistration = assertObject(
      await cdp.Mod.addCustomCommand({
        name: "Custom.targetIdFromTabId",
        params_schema: {
          tabId: z.number(),
        },
        result_schema: {
          targetId: cdp.types.zod.Target.TargetID.nullable(),
          tabId: z.number().optional(),
        },
        expression: `async ({ tabId }) => {
        const targets = await chrome.debugger.getTargets();
        const target = targets.find(target => target.type === "page" && target.tabId === tabId);
        return { targetId: target?.id ?? null };
      }`,
      }),
      "Mod.addCustomCommand Custom.targetIdFromTabId",
    );
    if (targetCommandRegistration.registered !== true) {
      throw new Error(`unexpected targetIdFromTabId registration ${JSON.stringify(targetCommandRegistration)}`);
    }
    const responseMiddlewareRegistration = assertObject(
      await cdp.Mod.addMiddleware({
        name: "*",
        phase: cdp.RESPONSE,
        expression: `async (payload, next) => {
        const seen = new WeakSet();
        const visit = async value => {
          if (!value || typeof value !== "object" || seen.has(value)) return;
          seen.add(value);
          if (!Array.isArray(value) && typeof value.targetId === "string" && value.tabId == null) {
            const { tabId } = await cdp.send("Custom.TabIdFromTargetId", { targetId: value.targetId });
            if (tabId != null) value.tabId = tabId;
          }
          for (const child of Array.isArray(value) ? value : Object.values(value)) await visit(child);
        };
        await visit(payload);
        return next(payload);
      }`,
      }),
      "Mod.addMiddleware response",
    );
    if (responseMiddlewareRegistration.registered !== true || responseMiddlewareRegistration.phase !== cdp.RESPONSE) {
      throw new Error(`unexpected response middleware registration ${JSON.stringify(responseMiddlewareRegistration)}`);
    }

    const eventMiddlewareRegistration = assertObject(
      await cdp.Mod.addMiddleware({
        name: "*",
        phase: cdp.EVENT,
        expression: `async (payload, next) => {
        const seen = new WeakSet();
        const visit = async value => {
          if (!value || typeof value !== "object" || seen.has(value)) return;
          seen.add(value);
          if (!Array.isArray(value) && typeof value.targetId === "string" && value.tabId == null) {
            const { tabId } = await cdp.send("Custom.TabIdFromTargetId", { targetId: value.targetId });
            if (tabId != null) value.tabId = tabId;
          }
          for (const child of Array.isArray(value) ? value : Object.values(value)) await visit(child);
        };
        await visit(payload);
        return next(payload);
      }`,
      }),
      "Mod.addMiddleware event",
    );
    if (eventMiddlewareRegistration.registered !== true || eventMiddlewareRegistration.phase !== cdp.EVENT) {
      throw new Error(`unexpected event middleware registration ${JSON.stringify(eventMiddlewareRegistration)}`);
    }

    const demoEventRegistration = assertObject(
      await cdp.Mod.addCustomEvent({ name: "Custom.demoEvent" }),
      "Mod.addCustomEvent Custom.demoEvent",
    );
    if (demoEventRegistration.registered !== true || demoEventRegistration.name !== "Custom.demoEvent") {
      throw new Error(`unexpected Custom.demoEvent registration ${JSON.stringify(demoEventRegistration)}`);
    }
    const demoEventPromise = waitForEvent(cdp, "Custom.demoEvent", (event) => event?.value === "custom-event-ok");
    const emitResult = assertObject(
      await cdp.Mod.evaluate({
        expression: `async () => await ModCDP.emit("Custom.demoEvent", { value: "custom-event-ok" })`,
      }),
      "Custom.demoEvent emit",
    );
    if (emitResult.emitted !== true)
      throw new Error(`unexpected Custom.demoEvent emit result ${JSON.stringify(emitResult)}`);
    const demoEvent = assertObject(await demoEventPromise, "Custom.demoEvent");
    console.log("Custom.demoEvent ->", demoEvent);

    const PageTargetUpdated = z
      .object({
        targetId: cdp.types.zod.Target.TargetID,
        tabId: z.number().optional(),
        url: z.string().nullable().optional(),
      })
      .passthrough()
      .meta({ id: "Custom.pageTargetUpdated" });
    const pageTargetEventRegistration = assertObject(
      await cdp.Mod.addCustomEvent(PageTargetUpdated),
      "Mod.addCustomEvent Custom.pageTargetUpdated",
    );
    if (pageTargetEventRegistration.registered !== true) {
      throw new Error(`unexpected page target event registration ${JSON.stringify(pageTargetEventRegistration)}`);
    }
    cdp.on(PageTargetUpdated, (event) => {
      console.log("Custom.pageTargetUpdated ->", event);
      pageTargetEvents.push(event);
    });

    await cdp.Target.setDiscoverTargets({ discover: true });
    const createdTarget = await cdp.Target.createTarget({
      url: "https://example.com",
      background: true,
    });
    const targetDeadline = Date.now() + DEFAULT_TARGET_EVENT_TIMEOUT_MS;
    while (
      !targetCreatedEvents.some((event) => event?.targetInfo?.targetId === createdTarget.targetId) &&
      Date.now() < targetDeadline
    ) {
      await sleep(DEFAULT_DEMO_EVENT_POLL_INTERVAL_MS);
    }
    if (!targetCreatedEvents.some((event) => event?.targetInfo?.targetId === createdTarget.targetId)) {
      throw new Error(`expected Target.targetCreated for ${createdTarget.targetId}`);
    }
    console.log("normal event matched ->", createdTarget.targetId);

    const tabFromTarget = await cdp.send("Custom.TabIdFromTargetId", {
      targetId: createdTarget.targetId,
    });
    const tabFromTargetId =
      typeof tabFromTarget === "number"
        ? tabFromTarget
        : tabFromTarget && typeof tabFromTarget === "object"
          ? (tabFromTarget as { tabId?: unknown }).tabId
          : null;
    if (typeof tabFromTargetId !== "number")
      throw new Error(`unexpected Custom.TabIdFromTargetId result ${JSON.stringify(tabFromTarget)}`);
    console.log("Custom.TabIdFromTargetId ->", tabFromTarget);

    await cdp.Target.activateTarget({ targetId: createdTarget.targetId });
    const pageTargetEmitResult = assertObject(
      await cdp.Mod.evaluate({
        params: { targetId: createdTarget.targetId },
        expression: `async ({ targetId }) => {
          const targets = await chrome.debugger.getTargets();
          const target = targets.find(target => target.id === targetId);
          if (!target?.id) throw new Error(\`target \${targetId} not found\`);
          await cdp.emit("Custom.pageTargetUpdated", { targetId: target.id, url: target.url ?? null });
          return { emitted: true, targetId: target.id };
        }`,
      }),
      "Custom.pageTargetUpdated emit",
    );
    if (pageTargetEmitResult.emitted !== true || pageTargetEmitResult.targetId !== createdTarget.targetId) {
      throw new Error(`unexpected Custom.pageTargetUpdated emit result ${JSON.stringify(pageTargetEmitResult)}`);
    }
    const pageTargetDeadline = Date.now() + DEFAULT_PAGE_TARGET_EVENT_TIMEOUT_MS;
    while (
      !pageTargetEvents.some((event) => event.targetId === createdTarget.targetId) &&
      Date.now() < pageTargetDeadline
    ) {
      await sleep(DEFAULT_DEMO_EVENT_POLL_INTERVAL_MS);
    }
    const pageTarget = pageTargetEvents.find((event) => event.targetId === createdTarget.targetId);
    if (!pageTarget) throw new Error(`expected Custom.pageTargetUpdated for ${createdTarget.targetId}`);
    if (pageTarget.tabId !== tabFromTargetId)
      throw new Error(`unexpected Custom.pageTargetUpdated result ${JSON.stringify(pageTarget)}`);

    const targetFromTab = await cdp.send("Custom.targetIdFromTabId", {
      tabId: pageTarget.tabId,
    });
    if (targetFromTab.targetId !== createdTarget.targetId || targetFromTab.tabId !== pageTarget.tabId) {
      throw new Error(`unexpected Custom.targetIdFromTabId/middleware result ${JSON.stringify(targetFromTab)}`);
    }
    console.log("Custom.targetIdFromTabId ->", targetFromTab);

    console.log(
      `\nSUCCESS (${mode}/${upstream_mode}): normal command, normal event, custom commands, custom event, and middleware all passed`,
    );

    // Drop into an interactive prompt when stdin is a TTY. Lets you poke at
    // the live browser using the same JSON command syntax as the Python and Go
    // demos. Skip the prompt when run non-interactively (CI, piped stdin) so
    // the demo exits cleanly after assertions.
    if (process.stdin.isTTY) {
      cdp.on("Mod.pong", (e) => console.log("\n[event] Mod.pong", e));
      await runRepl(cdp, mode);
    }
  } finally {
    await cdp.close();
  }
}

async function runRepl(cdp, mode) {
  console.log(`\nBrowser remains running. Mode: ${mode}.`);
  console.log("Enter commands as Domain.method({...JSON params...}). Examples:");
  console.log("  Browser.getVersion({})");
  console.log('  Mod.evaluate({"expression": "chrome.tabs.query({active: true})"})');
  console.log('  Custom.TabIdFromTargetId({"targetId": "..."})');
  console.log("Type exit or quit to disconnect (browser keeps running).");

  const rl = createInterface({ input: process.stdin, output: process.stdout });
  try {
    while (true) {
      let line;
      try {
        line = (await rl.question("ModCDP> ")).trim();
      } catch {
        break;
      }
      if (!line) continue;
      if (line === "exit" || line === "quit") break;
      try {
        const match = line.match(/^([A-Za-z_][\w]*\.[A-Za-z_][\w]*)(?:\(([\s\S]*)\))?$/);
        if (!match) throw new Error("format: Domain.method({...})");
        const [, method, raw = ""] = match;
        const params = raw.trim() ? JSON.parse(raw) : {};
        const [domain, command] = method.split(".");
        const result = cdp[domain]?.[command] ? await cdp[domain][command](params) : await cdp.send(method, params);
        console.log(JSON.stringify(result, null, 2));
      } catch (e) {
        console.error("error:", e?.message || e);
      }
    }
  } finally {
    rl.close();
  }
}

main().catch((e) => {
  console.error("DEMO FAILED:", e);
  process.exitCode = 1;
});
