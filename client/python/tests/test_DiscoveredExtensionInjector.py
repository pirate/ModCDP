from __future__ import annotations

import json
import shutil
import tempfile
import unittest
from pathlib import Path
from typing import Any, cast

from modcdp import ModCDPClient
from modcdp.LocalBrowserLauncher import LocalBrowserLauncher


ROOT = Path(__file__).resolve().parents[3]
EXTENSION_PATH = ROOT / "dist" / "extension"
CUSTOM_EXTENSION_ID = "hhklgmbgnbeghnjidampacgmgnhelifg"
CUSTOM_EXTENSION_PUBLIC_KEY = (
    "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzG1LUbtH0aHMKjTAUeT0saY8xfnRNENctFJme3C1qnsqT7PAXMxJC4nT7tBZy2gEGRirBb3zIZ3OyAu9a0QR8lTLupDp4qHWOhQ7dl9ZjxjQdYa4Gby0xuXLdQrJIxDbmuv+UVJvYa8vRTwQB8koygbzDDDP5/YiB6mc0hbh8XBb82Ossy7T280k8280o/rS0CXdioUraCHj58PDhfxbs18TBcYfOjuRqua9J2oddxobtGehSD0gDtbvn2IWDtRajOlgZZyuS1vLoSR7C1ulFzpRSYPEMhI2x+wphut7E3QImyJ577YeULVGpt988FcixOou7udjx3/IUWjpq8046wIDAQAB"
)


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

    def test_selects_configured_extension_when_multiple_modcdp_workers_exist(self) -> None:
        with tempfile.TemporaryDirectory(prefix="modcdp-custom-extension-") as custom_extension_dir:
            custom_extension_path = Path(custom_extension_dir)
            shutil.copytree(EXTENSION_PATH, custom_extension_path, dirs_exist_ok=True)
            manifest_path = custom_extension_path / "manifest.json"
            manifest = json.loads(manifest_path.read_text())
            manifest["key"] = CUSTOM_EXTENSION_PUBLIC_KEY
            manifest["name"] = "ModCDP Bridge Custom Test"
            manifest_path.write_text(json.dumps(manifest, indent=2) + "\n")

            chrome = LocalBrowserLauncher(
                {
                    "headless": True,
                    "sandbox": False,
                    "extra_args": [f"--load-extension={EXTENSION_PATH},{custom_extension_path}"],
                }
            ).launch()
            cdp = ModCDPClient(
                launch={"mode": "remote"},
                upstream={"mode": "ws", "ws_url": chrome["cdp_url"]},
                extension={
                    "mode": "discover",
                    "extension_id": CUSTOM_EXTENSION_ID,
                    "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                    "trust_service_worker_target": True,
                    "require_service_worker_target": True,
                },
            )

            try:
                cdp.connect()
                self.assertEqual(cdp.connect_timing.get("extension_source") if cdp.connect_timing else None, "discovered")
                self.assertEqual(cdp.extension_id, CUSTOM_EXTENSION_ID)
                self.assertEqual(cdp.Mod.evaluate(expression="chrome.runtime.id"), CUSTOM_EXTENSION_ID)
                targets = cast(dict[str, Any], cdp.sendRaw("Target.getTargets"))
                target_infos = targets.get("targetInfos")
                self.assertIsInstance(target_infos, list)
                target_infos = cast(list[Any], target_infos)
                modcdp_workers = [
                    target
                    for target in target_infos
                    if isinstance(target, dict)
                    and target.get("type") == "service_worker"
                    and str(target.get("url", "")).endswith("/modcdp/service_worker.js")
                ]
                self.assertTrue(
                    any(
                        target.get("url") == f"chrome-extension://{CUSTOM_EXTENSION_ID}/modcdp/service_worker.js"
                        for target in modcdp_workers
                    )
                )
                self.assertTrue(
                    any(
                        target.get("url")
                        == "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js"
                        for target in modcdp_workers
                    )
                )
            finally:
                cdp.close()
                chrome["close"]()


if __name__ == "__main__":
    unittest.main()
