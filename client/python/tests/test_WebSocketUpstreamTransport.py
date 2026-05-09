from __future__ import annotations

import time
import unittest
from queue import Queue
from typing import Any, cast

from modcdp import ModCDPClient
from modcdp.LocalBrowserLauncher import LocalBrowserLauncher
from modcdp.WebSocketUpstreamTransport import WebSocketUpstreamTransport


class WebSocketUpstreamTransportTests(unittest.TestCase):
    def test_constructor_update_and_server_config_match_ts_shape(self) -> None:
        transport = WebSocketUpstreamTransport()
        self.assertEqual(transport.url, "")
        self.assertEqual(transport.getServerConfig(), {})
        self.assertIs(transport.update({"ws_url": "ws://127.0.0.1:1/devtools/browser/test"}), transport)
        self.assertEqual(transport.url, "ws://127.0.0.1:1/devtools/browser/test")
        self.assertEqual(transport.getServerConfig(), {"loopback_cdp_url": "ws://127.0.0.1:1/devtools/browser/test"})
        unconfigured = WebSocketUpstreamTransport()
        with self.assertRaisesRegex(RuntimeError, r"upstream\.mode='ws' requires"):
            unconfigured.connect()
        with self.assertRaisesRegex(RuntimeError, "CDP websocket is not connected"):
            unconfigured.send({"id": 1, "method": "Browser.getVersion"})

    def test_launches_real_browser_and_speaks_raw_cdp(self) -> None:
        cdp = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "ws"},
            extension={
                "mode": "auto",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
        )
        try:
            cdp.connect()
            self.assertEqual(cdp.transport.mode if cdp.transport else None, "ws")
            self.assertEqual(cdp.upstream_endpoint_kind, "raw_cdp")
            timing = cdp.connect_timing
            self.assertIsNotNone(timing)
            if timing is None:
                raise AssertionError("expected connect timing")
            self.assertEqual(timing["upstream_mode"], "ws")
            self.assertEqual(timing["upstream_endpoint_kind"], "raw_cdp")
            self.assertGreaterEqual(timing["transport_connected_at"], timing["transport_started_at"])
            self.assertEqual(
                timing["transport_duration_ms"],
                timing["transport_connected_at"] - timing["transport_started_at"],
            )
            self.assertRegex(cdp.cdp_url or "", r"^ws://")
            version = cdp.sendRaw("Browser.getVersion")
            self.assertIsInstance(version["product"], str)
            time.sleep(1.5)
            raw_target_infos = cdp.sendRaw("Target.getTargets").get("targetInfos", [])
            target_infos = cast(list[dict[str, Any]], raw_target_infos if isinstance(raw_target_infos, list) else [])
            self.assertTrue(
                any(
                    target.get("type") == "service_worker"
                    and str(target.get("url", "")).endswith("/modcdp/service_worker.js")
                    for target in target_infos
                )
            )
            self.assertTrue(
                cdp.Mod.evaluate(
                    expression="Boolean(globalThis.ModCDP?.handleCommand && chrome.runtime.getURL('modcdp/service_worker.js'))"
                )
            )
        finally:
            cdp.close()

    def test_resolves_real_http_cdp_endpoint_to_browser_websocket(self) -> None:
        chrome = LocalBrowserLauncher({"headless": True, "sandbox": False}).launch()
        transport = WebSocketUpstreamTransport(chrome["cdp_url"])
        received: Queue[dict] = Queue()
        transport.onRecv(lambda message: received.put(message))
        try:
            transport.connect()
            self.assertRegex(transport.url or "", r"^ws://")
            transport.send({"id": 1, "method": "Browser.getVersion", "params": {}})
            response = received.get(timeout=5)
            self.assertEqual(response["id"], 1)
            self.assertIsInstance(response["result"]["product"], str)
        finally:
            transport.close()
            chrome["close"]()


if __name__ == "__main__":
    unittest.main()
