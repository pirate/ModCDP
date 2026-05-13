from __future__ import annotations

import os
from pathlib import Path
import unittest
from unittest.mock import patch

from modcdp import ModCDPClient
from modcdp.injector.BBBrowserExtensionInjector import BBBrowserExtensionInjector

HERE = Path(__file__).resolve().parent
EXTENSION_PATH = HERE.parents[1] / "dist" / "extension"


class BBBrowserExtensionInjectorTests(unittest.TestCase):
    def test_prepares_default_packaged_extension_zip_when_path_is_omitted(self) -> None:
        injector = BBBrowserExtensionInjector()
        try:
            with patch.object(injector, "_uploadExtension", return_value="aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") as upload:
                injector.prepare()
            self.assertTrue(str(injector.options.get("injector_extension_path", "")).endswith("extension.zip"))
            self.assertTrue(str(injector.zip_path or "").endswith("extension.zip"))
            upload.assert_called_once_with(injector.zip_path)
            self.assertEqual(injector.getLauncherConfig(), {"injector_extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
        finally:
            injector.close()

    def test_uploads_real_extension_and_launches_browserbase_browser_with_it_installed(self) -> None:
        if not os.environ.get("BROWSERBASE_API_KEY", "").strip():
            self.fail("BROWSERBASE_API_KEY is required for live Browserbase tests")
        cdp = ModCDPClient(
            launcher={
                "launcher_mode": "bb",
                "launcher_options": {
                    "timeout": 120,
                    **({"region": os.environ["BROWSERBASE_REGION"]} if os.environ.get("BROWSERBASE_REGION") else {}),
                },
            },
            upstream={"upstream_mode": "ws"},
            injector={
                "injector_mode": "inject",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.connect_timing.get("injector_source") if cdp.connect_timing else None, "bb")
            self.assertIsInstance(cdp.extension_id, str)
            service_worker_url = cdp.Mod.evaluate(expression="chrome.runtime.getURL('modcdp/service_worker.js')")
            self.assertRegex(str(service_worker_url), r"^chrome-extension://[a-z]{32}/modcdp/service_worker\.js$")
        finally:
            cdp.close()


if __name__ == "__main__":
    unittest.main()
