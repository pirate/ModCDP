from __future__ import annotations

import unittest

from modcdp import ModCDPClient
from modcdp.transport.PipeUpstreamTransport import PipeUpstreamTransport


class PipeUpstreamTransportTests(unittest.TestCase):
    def test_constructor_update_launcher_config_and_unconnected_errors_match_transport_surface(self) -> None:
        transport = PipeUpstreamTransport({"cdp_url": "pipe://constructor"})
        self.assertEqual(transport.mode, "pipe")
        self.assertEqual(transport.endpoint_kind, "raw_cdp")
        self.assertEqual(transport.url, "pipe://constructor")
        self.assertEqual(transport.getLauncherConfig(), {"remote_debugging": "pipe"})
        self.assertIs(transport.update({"cdp_url": "pipe://1234"}), transport)
        self.assertEqual(transport.url, "pipe://1234")
        with self.assertRaisesRegex(RuntimeError, r"upstream\.mode='pipe' requires"):
            transport.connect()
        with self.assertRaisesRegex(RuntimeError, "CDP pipe is not connected"):
            transport.send({"id": 1, "method": "Browser.getVersion"})

    def test_launches_real_browser_and_uses_pid_scoped_pipe_url(self) -> None:
        cdp = ModCDPClient(
            launcher={"launcher_mode": "local", "launcher_options": {"headless": True, "sandbox": False}},
            upstream={"upstream_mode": "pipe"},
            injector={
                "injector_mode": "auto",
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.transport.mode if cdp.transport else None, "pipe")
            self.assertEqual(cdp.upstream_endpoint_kind, "raw_cdp")
            self.assertRegex(cdp.cdp_url or "", r"^pipe://\d+$")
            self.assertEqual(cdp.transport.url if cdp.transport else None, cdp.cdp_url)
            version = cdp.sendRaw("Browser.getVersion")
            self.assertIsInstance(version["product"], str)
        finally:
            cdp.close()


if __name__ == "__main__":
    unittest.main()
