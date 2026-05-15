from __future__ import annotations

from pathlib import Path
from queue import Empty, Queue
from typing import Any, cast
import unittest

from modcdp import ModCDPClient


HERE = Path(__file__).resolve().parent
EXTENSION_PATH = HERE.parents[1] / "dist" / "extension"

GET_TARGETS_OVERRIDE = r"""
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
"""

TAB_ID_FROM_TARGET_ID_COMMAND = r"""
async ({ targetId }) => {
  const targets = await chrome.debugger.getTargets();
  const target = targets.find(target => target.id === targetId);
  if (target?.tabId != null) return { tabId: target.tabId };
  const tabs = await chrome.tabs.query({});
  const tab = tabs.find(tab => target?.url && (tab.url === target.url || tab.pendingUrl === target.url));
  return { tabId: tab?.id ?? null };
}
"""

ADD_TAB_ID_MIDDLEWARE = r"""
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
"""


def target_infos_from_result(result: Any) -> list[dict[str, Any]]:
    result_map = cast(dict[str, Any], result)
    return cast(list[dict[str, Any]], result_map["targetInfos"])


class ModCDPClientRoutedDefaultOverridesTests(unittest.TestCase):
    def test_service_worker_routed_standard_cdp_commands_and_events_can_be_transformed(self) -> None:
        owner = ModCDPClient(
            launcher={
                "launcher_mode": "local",
                "launcher_options": {"headless": True},
            },
            upstream={"upstream_mode": "ws"},
            injector={
                "injector_mode": "auto",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
        )
        owner.connect()
        cdp = ModCDPClient(
            launcher={"launcher_mode": "remote"},
            upstream={"upstream_mode": "ws", "upstream_cdp_url": owner.cdp_url},
            injector={
                "injector_mode": "discover",
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
            client={
                "client_routes": {
                    "Target.getTargets": "service_worker",
                    "Target.createTarget": "service_worker",
                    "Target.setDiscoverTargets": "service_worker",
                }
            },
            server={"server_loopback_cdp_url": owner.cdp_url, "server_routes": {"*.*": "loopback_cdp"}},
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.cdp_url, owner.cdp_url)
            self.assertIsNotNone(cdp.server)
            server = cast(dict[str, Any], cdp.server)
            self.assertEqual(server["server_loopback_cdp_url"], owner.cdp_url)

            raw_targets = cdp.send("Target.getTargets")
            raw_target_infos = target_infos_from_result(raw_targets)
            self.assertTrue(raw_target_infos)
            self.assertFalse(any("tabId" in target_info for target_info in raw_target_infos))

            cdp.Mod.addCustomCommand("Custom.tabIdFromTargetId", expression=TAB_ID_FROM_TARGET_ID_COMMAND)
            cdp.Mod.addMiddleware(name="*", phase="response", expression=ADD_TAB_ID_MIDDLEWARE)
            middleware_targets = cdp.send("Target.getTargets")
            self.assertTrue(
                any(
                    target_info.get("type") == "page" and isinstance(target_info.get("tabId"), int)
                    for target_info in target_infos_from_result(middleware_targets)
                )
            )

            cdp.Mod.addMiddleware(name="*", phase="event", expression=ADD_TAB_ID_MIDDLEWARE)
            cdp.Mod.addCustomCommand("Target.getTargets", expression=GET_TARGETS_OVERRIDE)

            enriched_targets = cdp.send("Target.getTargets")
            enriched_target_infos = target_infos_from_result(enriched_targets)
            self.assertTrue(enriched_target_infos)
            self.assertTrue(all("tabId" in target_info for target_info in enriched_target_infos))
            self.assertTrue(
                any(
                    target_info.get("type") == "page" and isinstance(target_info.get("tabId"), int)
                    for target_info in enriched_target_infos
                )
            )

            cdp.Mod.addCustomEvent("Target.targetCreated")
            transformed_events: Queue[dict] = Queue()

            def on_target_created(params):
                target_info = params.get("targetInfo") if isinstance(params, dict) else None
                if isinstance(target_info, dict) and "tabId" in target_info:
                    transformed_events.put(params)

            cdp.on("Target.targetCreated", on_target_created)
            cdp.Target.setDiscoverTargets(discover=False)
            cdp.Target.setDiscoverTargets(discover=True)
            cdp.Target.getTargets()

            try:
                event = transformed_events.get(timeout=2)
            except Empty:
                created_target = cdp.Target.createTarget(url="about:blank#modcdp-target-created")
                cdp.Target.getTargets()
                event = transformed_events.get(timeout=5)
                self.assertEqual(event["targetInfo"]["targetId"], created_target.targetId)
            self.assertIn("tabId", event["targetInfo"])
        finally:
            try:
                cdp.Target.setDiscoverTargets(discover=False)
            except Exception:
                pass
            cdp.close()
            owner.close()


if __name__ == "__main__":
    unittest.main()
