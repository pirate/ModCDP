from __future__ import annotations

import unittest
import json

from modcdp.translate import (
    CUSTOM_EVENT_BINDING_NAME,
    route_for,
    unwrap_event_if_needed,
    unwrap_response_if_needed,
    wrap_command_if_needed,
)


class TranslateTests(unittest.TestCase):
    def test_routes_wraps_and_unwraps_modcdp_protocol_messages_deterministically(self) -> None:
        self.assertEqual(route_for("Browser.getVersion", {"Browser.*": "direct_cdp", "*.*": "service_worker"}), "direct_cdp")
        self.assertEqual(route_for("Target.getTargets", {"Browser.*": "direct_cdp", "*.*": "service_worker"}), "service_worker")

        direct = wrap_command_if_needed("Browser.getVersion", {}, routes={"*.*": "direct_cdp"})
        self.assertEqual(direct["target"], "direct_cdp")
        self.assertEqual(direct["steps"], [{"method": "Browser.getVersion", "params": {}}])

        wrapped = wrap_command_if_needed(
            "Mod.evaluate",
            {"expression": "({ ok: true })", "params": {"value": 1}},
            cdp_session_id="session-1",
        )
        self.assertEqual(wrapped["target"], "service_worker")
        self.assertEqual(wrapped["steps"][0]["method"], "Runtime.callFunctionOn")
        self.assertIn('attachToSession("session-1")', str(wrapped["steps"][0].get("params", {}).get("functionDeclaration")))
        self.assertEqual(wrapped["steps"][0].get("unwrap"), "runtime")

        configured = wrap_command_if_needed(
            "Mod.configure",
            {"server": {"server_routes": {"*.*": "loopback_cdp"}}},
            cdp_session_id="session-1",
        )
        self.assertEqual(configured["steps"][0].get("unwrap"), "runtime_json")

        self.assertEqual(unwrap_response_if_needed({"result": {"type": "object", "value": {"ok": True}}}, "runtime"), {"ok": True})
        self.assertEqual(unwrap_response_if_needed({"product": "Chrome/1"}, None), {"product": "Chrome/1"})

        payload = json.dumps(
            {"event": "Custom.ready", "data": {"ready": True}, "cdpSessionId": "session-2"},
            separators=(",", ":"),
        )
        self.assertEqual(
            unwrap_event_if_needed(
                "Runtime.bindingCalled",
                {"name": CUSTOM_EVENT_BINDING_NAME, "payload": payload},
                "session-1",
                "session-1",
            ),
            {"event": "Custom.ready", "data": {"ready": True}, "sessionId": "session-2"},
        )
        self.assertIsNone(unwrap_event_if_needed("Runtime.consoleAPICalled", {"name": CUSTOM_EVENT_BINDING_NAME, "payload": payload}))


if __name__ == "__main__":
    unittest.main()
