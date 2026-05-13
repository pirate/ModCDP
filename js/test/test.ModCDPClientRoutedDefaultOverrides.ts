import assert from "node:assert/strict";
import { test } from "vitest";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { ModCDPClient } from "../src/client/ModCDPClient.js";
import { commands, events } from "../src/types/generated/zod.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");
const DEFAULT_ROUTED_OVERRIDES_TEST_TIMEOUT_MS = 45_000;

function hasTargetInfo(value: unknown): value is { targetInfo: Record<string, unknown> } {
  if (value == null || typeof value !== "object" || Array.isArray(value)) return false;
  const targetInfo = (value as Record<string, unknown>).targetInfo;
  return targetInfo != null && typeof targetInfo === "object" && !Array.isArray(targetInfo);
}

const getTargetsOverride = String.raw`
async (params) => {
  const [upstream, tabs] = await Promise.all([
    ModCDP.sendLoopback("Target.getTargets", params),
    chrome.tabs.query({}),
  ]);

  const tabIdByUrl = new Map();
  for (const tab of tabs) {
    for (const url of [tab.url, tab.pendingUrl].filter(Boolean)) {
      if (!tabIdByUrl.has(url)) tabIdByUrl.set(url, tab.id);
    }
  }

  return {
    ...upstream,
    targetInfos: (upstream.targetInfos || []).map(targetInfo => ({
      ...targetInfo,
      tabId: tabIdByUrl.get(targetInfo.url) ?? null,
    })),
  };
}
`;

const tabIdFromTargetIdCommand = String.raw`
async ({ targetId }) => {
  const targets = await chrome.debugger.getTargets();
  const target = targets.find(target => target.id === targetId);
  if (target?.tabId != null) return { tabId: target.tabId };
  const tabs = await chrome.tabs.query({});
  const tab = tabs.find(tab => target?.url && (tab.url === target.url || tab.pendingUrl === target.url));
  return { tabId: tab?.id ?? null };
}
`;

const addTabIdMiddleware = String.raw`
async (payload, next) => {
  const seen = new WeakSet();
  const visit = async value => {
    if (!value || typeof value !== "object" || seen.has(value)) return;
    seen.add(value);
    if (!Array.isArray(value) && typeof value.targetId === "string" && value.tabId == null) {
      const { tabId } = await cdp.send("Custom.tabIdFromTargetId", { targetId: value.targetId });
      if (tabId != null) value.tabId = tabId;
    }
    for (const child of Array.isArray(value) ? value : Object.values(value)) await visit(child);
  };
  await visit(payload);
  return next(payload);
}
`;

test(
  "service-worker routed standard CDP commands and events can be transformed",
  { timeout: DEFAULT_ROUTED_OVERRIDES_TEST_TIMEOUT_MS },
  async () => {
    const chrome = await new LocalBrowserLauncher().launch({
      headless: true,
      sandbox: process.platform !== "linux",
      extra_args: [`--load-extension=${EXTENSION_PATH}`],
    });
    const cdp = new ModCDPClient({
      launcher: { launcher_mode: "remote" },
      upstream: { upstream_mode: "ws", upstream_cdp_url: chrome.cdp_url },
      injector: {
        injector_mode: "auto",
        injector_extension_path: EXTENSION_PATH,
        injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
        injector_trust_service_worker_target: true,
      },
      client: {
        client_routes: {
          "Target.getTargets": "service_worker",
          "Target.createTarget": "service_worker",
          "Target.setDiscoverTargets": "service_worker",
        },
      },
      server: {
        server_loopback_cdp_url: chrome.cdp_url,
        server_routes: { "*.*": "loopback_cdp" },
      },
    });

    try {
      await cdp.connect();
      assert.equal(cdp.cdp_url, chrome.cdp_url);
      assert.equal(cdp.server.server_loopback_cdp_url, chrome.cdp_url);

      const rawTargets = commands["Target.getTargets"].result.parse(await cdp._sendMessage("Target.getTargets"));
      assert.ok(rawTargets.targetInfos?.length > 0, "expected raw Target.getTargets targetInfos");
      assert.equal(
        rawTargets.targetInfos.some((targetInfo) => Object.hasOwn(targetInfo, "tabId")),
        false,
        "raw CDP TargetInfo should not already contain tabId",
      );

      await cdp.Mod.addCustomCommand({
        name: "Custom.tabIdFromTargetId",
        expression: tabIdFromTargetIdCommand,
      });
      await cdp.Mod.addMiddleware({
        name: "*",
        phase: cdp.RESPONSE,
        expression: addTabIdMiddleware,
      });
      const middlewareTargets = await cdp.send("Target.getTargets");
      assert.ok(
        middlewareTargets.targetInfos.some(
          (targetInfo) => targetInfo.type === "page" && Number.isInteger(targetInfo.tabId),
        ),
        "wildcard response middleware should add tabId next to targetId inside TargetInfo",
      );

      await cdp.Mod.addMiddleware({
        name: "*",
        phase: cdp.EVENT,
        expression: addTabIdMiddleware,
      });

      await cdp.Mod.addCustomCommand({
        name: cdp.Target.getTargets,
        expression: getTargetsOverride,
      });

      const enrichedTargets = await cdp.send("Target.getTargets");
      assert.ok(enrichedTargets.targetInfos?.length > 0, "expected enriched Target.getTargets targetInfos");
      assert.equal(
        enrichedTargets.targetInfos.every((targetInfo) => Object.hasOwn(targetInfo, "tabId")),
        true,
        "every routed TargetInfo should include a tabId property",
      );
      assert.ok(
        enrichedTargets.targetInfos.some(
          (targetInfo) => targetInfo.type === "page" && Number.isInteger(targetInfo.tabId),
        ),
        "expected at least one page target to be matched to a chrome.tabs tab id",
      );

      await cdp.Mod.addCustomEvent({ name: cdp.Target.targetCreated });

      const transformedEvents: unknown[] = [];
      cdp.on("Target.targetCreated", (params) => {
        if (!hasTargetInfo(params)) return;
        if (!Object.hasOwn(params.targetInfo || {}, "tabId")) return;
        transformedEvents.push(params);
      });

      await cdp.Target.setDiscoverTargets({ discover: false });
      await cdp.Target.setDiscoverTargets({ discover: true });
      await cdp.Target.getTargets();

      if (transformedEvents.length === 0) {
        const createdTarget = await cdp.Target.createTarget({ url: "about:blank#modcdp-target-created" });
        await cdp.Target.getTargets();
        assert.ok(
          transformedEvents.some((params) => {
            if (!hasTargetInfo(params)) return;
            return params.targetInfo.targetId === createdTarget.targetId;
          }),
          `expected transformed Target.targetCreated for ${createdTarget.targetId}`,
        );
      }

      const event = events["Target.targetCreated"].parse(transformedEvents[0]);
      assert.ok(Object.hasOwn(event.targetInfo, "tabId"), "transformed event targetInfo should include tabId");
    } finally {
      try {
        await cdp.Target.setDiscoverTargets({ discover: false });
      } catch {}
      await cdp.close();
      await chrome.close();
    }
  },
);
