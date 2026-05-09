from __future__ import annotations

import os
from pathlib import Path
import unittest

from modcdp import ModCDPClient
from modcdp.BBBrowserExtensionInjector import BBBrowserExtensionInjector

HERE = Path(__file__).resolve().parent
EXTENSION_PATH = HERE.parents[2] / "dist" / "extension"


class BBBrowserExtensionInjectorTests(unittest.TestCase):
    def test_uses_configured_extension_id(self) -> None:
        injector = BBBrowserExtensionInjector({"extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
        injector.prepare()
        self.assertEqual(injector.getLauncherConfig().get("extension_id"), "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

    @unittest.skipIf(os.environ.get("BROWSERBASE_API_KEY", "").strip(), "BROWSERBASE_API_KEY is set")
    def test_requires_browserbase_api_key_when_extension_upload_is_needed(self) -> None:
        injector = BBBrowserExtensionInjector({"extension_path": str(EXTENSION_PATH)})

        try:
            with self.assertRaisesRegex(RuntimeError, "BROWSERBASE_API_KEY"):
                injector.prepare()
            self.assertIsNone(injector.cleanup_dir)
        finally:
            injector.close()

    @unittest.skipUnless(os.environ.get("BROWSERBASE_API_KEY", "").strip(), "BROWSERBASE_API_KEY is required for live Browserbase tests")
    def test_uploads_real_extension_and_launches_browserbase_browser_with_it_installed(self) -> None:
        cdp = ModCDPClient(
            launch={
                "mode": "bb",
                "options": {
                    "project_id": os.environ.get("BROWSERBASE_PROJECT_ID"),
                    "timeout": 120,
                    **({"region": os.environ["BROWSERBASE_REGION"]} if os.environ.get("BROWSERBASE_REGION") else {}),
                },
            },
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
            self.assertEqual(cdp.connect_timing.get("extension_source") if cdp.connect_timing else None, "bb")
            self.assertIsInstance(cdp.extension_id, str)
            service_worker_url = cdp.Mod.evaluate(expression="chrome.runtime.getURL('modcdp/service_worker.js')")
            self.assertRegex(str(service_worker_url), r"^chrome-extension://[a-z]{32}/modcdp/service_worker\.js$")
        finally:
            cdp.close()


if __name__ == "__main__":
    unittest.main()
