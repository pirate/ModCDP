from __future__ import annotations

import unittest

from modcdp import ModCDPClient


class WebSocketUpstreamTransportTests(unittest.TestCase):
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
