from __future__ import annotations

import unittest
from typing import Any, cast

from modcdp import ModCDPClient


class ModCDPClientTests(unittest.TestCase):
    def test_constructor_normalizes_nested_config_owners(self) -> None:
        cdp = ModCDPClient(
            launch={
                "mode": "local",
                "executable_path": "/tmp/chrome",
                "user_data_dir": "/tmp/profile",
                "options": {"headless": True},
            },
            upstream={
                "mode": "ws",
                "ws_url": "http://127.0.0.1:9222",
                "ws_connect_error_settle_timeout_ms": 321,
            },
            extension={
                "mode": "discover",
                "path": "/tmp/ext",
                "extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "service_worker_url_includes": ["modcdp"],
                "service_worker_url_suffixes": ["/custom/service_worker.js"],
                "trust_service_worker_target": True,
                "require_service_worker_target": True,
                "execution_context_timeout_ms": 4321,
                "service_worker_probe_timeout_ms": 5432,
                "service_worker_ready_timeout_ms": 6543,
                "service_worker_poll_interval_ms": 76,
                "target_session_poll_interval_ms": 87,
            },
            client={
                "routes": {"*.*": "direct_cdp"},
                "mirror_upstream_events": False,
                "cdp_send_timeout_ms": 1234,
                "event_wait_timeout_ms": 2345,
            },
            server={"routes": {"*.*": "loopback_cdp"}},
        )

        self.assertEqual(cdp.launch["options"], {"headless": True})
        self.assertEqual(cdp._launch_options()["executable_path"], "/tmp/chrome")
        self.assertEqual(cdp._launch_options()["user_data_dir"], "/tmp/profile")
        self.assertEqual(cdp.upstream["ws_connect_error_settle_timeout_ms"], 321)
        self.assertEqual(cdp.extension["execution_context_timeout_ms"], 4321)
        self.assertEqual(cdp.extension["service_worker_probe_timeout_ms"], 5432)
        self.assertEqual(cdp.extension["service_worker_ready_timeout_ms"], 6543)
        self.assertEqual(cdp.extension["service_worker_poll_interval_ms"], 76)
        self.assertEqual(cdp.extension["target_session_poll_interval_ms"], 87)
        self.assertEqual(cdp.client["routes"]["*.*"], "direct_cdp")
        self.assertEqual(cdp.client["mirror_upstream_events"], False)
        self.assertEqual(cdp.client["cdp_send_timeout_ms"], 1234)
        self.assertEqual(cdp.client["event_wait_timeout_ms"], 2345)
        self.assertNotIn("routes", cdp.__dict__)
        self.assertNotIn("cdp_send_timeout_ms", cdp.__dict__)
        self.assertNotIn("service_worker_probe_timeout_ms", cdp.__dict__)

        params = cast(dict[str, Any], cdp._server_configure_params())
        self.assertEqual(params["client"]["routes"]["*.*"], "direct_cdp")
        self.assertEqual(params["server"]["cdp_send_timeout_ms"], 1234)
        self.assertEqual(params["server"]["loopback_execution_context_timeout_ms"], 4321)
        self.assertEqual(params["server"]["ws_connect_error_settle_timeout_ms"], 321)

    def test_connects_with_local_launch_and_injector_chain(self) -> None:
        cdp = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "ws"},
            extension={
                "mode": "inject",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.connect_timing["extension_source"] if cdp.connect_timing else None, "local_launch")
            self.assertEqual(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf")
            self.assertEqual(
                cdp.Mod.evaluate(expression="chrome.runtime.getURL('modcdp/service_worker.js')"),
                "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js",
            )
        finally:
            cdp.close()


if __name__ == "__main__":
    unittest.main()
