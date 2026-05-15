from __future__ import annotations

import unittest
from pathlib import Path

from modcdp import ModCDPClient


ROOT = Path(__file__).resolve().parents[2]
EXTENSION_PATH = ROOT / "dist" / "extension"


class BorrowedExtensionInjectorTests(unittest.TestCase):
    def test_bootstraps_modcdp_inside_live_extension_service_worker(self) -> None:
        owner = ModCDPClient(
            launcher={"launcher_mode": "local", "launcher_options": {"headless": True}},
            upstream={"upstream_mode": "ws"},
            injector={
                "injector_mode": "auto",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
        )
        try:
            owner.connect()
            cdp = ModCDPClient(
                launcher={"launcher_mode": "remote"},
                upstream={"upstream_mode": "ws", "upstream_cdp_url": owner.cdp_url},
                injector={
                    "injector_mode": "borrow",
                    "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                    "injector_trust_service_worker_target": True,
                },
            )
            try:
                cdp.connect()
                self.assertEqual(cdp.connect_timing.get("injector_source") if cdp.connect_timing else None, "borrowed")
                self.assertEqual(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf")
                self.assertEqual(
                    cdp.Mod.evaluate(expression="chrome.runtime.getURL('modcdp/service_worker.js')"),
                    "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js",
                )
            finally:
                cdp.close()
        finally:
            owner.close()


if __name__ == "__main__":
    unittest.main()
