from __future__ import annotations

import unittest
from pathlib import Path

from modcdp import ModCDPClient
from modcdp.LocalBrowserLauncher import LocalBrowserLauncher


ROOT = Path(__file__).resolve().parents[3]
EXTENSION_PATH = ROOT / "dist" / "extension"


class BorrowedExtensionInjectorTests(unittest.TestCase):
    def test_bootstraps_modcdp_inside_live_extension_service_worker(self) -> None:
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
                "mode": "borrow",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.connect_timing.get("extension_source") if cdp.connect_timing else None, "borrowed")
            self.assertEqual(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf")
            target_infos = cdp.send("Target.getTargets")["targetInfos"]
            self.assertGreater(len(target_infos), 0)
        finally:
            cdp.close()
            chrome["close"]()


if __name__ == "__main__":
    unittest.main()
