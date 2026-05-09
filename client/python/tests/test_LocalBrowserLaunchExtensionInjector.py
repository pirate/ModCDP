from __future__ import annotations

import unittest
from pathlib import Path

from modcdp.ExtensionInjector import DEFAULT_MODCDP_EXTENSION_ID
from modcdp.LocalBrowserLaunchExtensionInjector import LocalBrowserLaunchExtensionInjector
from modcdp.ModCDPClient import ModCDPClient


ROOT = Path(__file__).resolve().parents[3]
EXTENSION_PATH = ROOT / "dist" / "extension"


class LocalBrowserLaunchExtensionInjectorTests(unittest.TestCase):
    def test_loads_real_extension_during_local_launch(self) -> None:
        cdp = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "ws"},
            extension={
                "mode": "inject",
                "path": str(EXTENSION_PATH),
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.connect_timing.get("extension_source") if cdp.connect_timing else None, "local_launch")
            self.assertEqual(cdp.extension_id, DEFAULT_MODCDP_EXTENSION_ID)
            self.assertRegex(cdp.ext_session_id or "", r"^.+$")
            self.assertEqual(
                cdp.Mod.evaluate(expression="chrome.runtime.getURL('modcdp/service_worker.js')"),
                f"chrome-extension://{DEFAULT_MODCDP_EXTENSION_ID}/modcdp/service_worker.js",
            )
        finally:
            cdp.close()

    def test_prepares_launcher_config(self) -> None:
        injector = LocalBrowserLaunchExtensionInjector({"extension_path": str(EXTENSION_PATH)})
        try:
            injector.prepare()
            extra_args = injector.getLauncherConfig().get("extra_args") or []
            self.assertEqual(len(extra_args), 1)
            self.assertTrue(extra_args[0].startswith("--load-extension="))
            self.assertEqual(injector.options.get("extension_id"), DEFAULT_MODCDP_EXTENSION_ID)
        finally:
            injector.close()


if __name__ == "__main__":
    unittest.main()
