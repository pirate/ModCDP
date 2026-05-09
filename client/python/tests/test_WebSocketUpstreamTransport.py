from __future__ import annotations

import unittest

from modcdp import ModCDPClient
from modcdp.WebSocketUpstreamTransport import WebSocketUpstreamTransport


class WebSocketUpstreamTransportTests(unittest.TestCase):
    def test_constructor_update_and_server_config_match_ts_shape(self) -> None:
        transport = WebSocketUpstreamTransport()
        self.assertEqual(transport.url, "")
        self.assertEqual(transport.getServerConfig(), {})
        self.assertIs(transport.update({"ws_url": "ws://127.0.0.1:1/devtools/browser/test"}), transport)
        self.assertEqual(transport.url, "ws://127.0.0.1:1/devtools/browser/test")
        self.assertEqual(transport.getServerConfig(), {"loopback_cdp_url": "ws://127.0.0.1:1/devtools/browser/test"})

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
            self.assertRegex(cdp.cdp_url or "", r"^ws://")
            version = cdp.send_raw("Browser.getVersion")
            self.assertIsInstance(version["product"], str)
        finally:
            cdp.close()


if __name__ == "__main__":
    unittest.main()
