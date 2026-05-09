from __future__ import annotations

import unittest
from pathlib import Path

from modcdp import ModCDPClient
from modcdp.LocalBrowserLauncher import LocalBrowserLauncher


ROOT = Path(__file__).resolve().parents[3]
EXTENSION_PATH = ROOT / "dist" / "extension"


class DiscoveredExtensionInjectorTests(unittest.TestCase):
    def test_attaches_to_already_loaded_real_modcdp_extension(self) -> None:
        chrome = LocalBrowserLauncher(
            {
                "headless": True,
                "sandbox": False,
                "extra_args": [f"--load-extension={EXTENSION_PATH}"],
            }
        ).launch()
        cdp = ModCDPClient(
            launch={"mode": "remote"},
            upstream={"mode": "ws", "ws_url": chrome["cdp_url"]},
            extension={
                "mode": "discover",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.connect_timing.get("extension_source") if cdp.connect_timing else None, "discovered")
            self.assertEqual(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf")
            self.assertEqual(
                cdp.Mod.evaluate(expression="chrome.runtime.getURL('modcdp/service_worker.js')"),
                "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js",
            )
        finally:
            cdp.close()
            chrome["close"]()


if __name__ == "__main__":
    unittest.main()
