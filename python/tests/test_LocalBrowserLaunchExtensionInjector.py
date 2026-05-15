from __future__ import annotations

import unittest
import time
import tempfile
import zipfile
from pathlib import Path
from typing import Any, cast

from modcdp.injector.ExtensionInjector import DEFAULT_MODCDP_EXTENSION_ID
from modcdp.injector.LocalBrowserLaunchExtensionInjector import LocalBrowserLaunchExtensionInjector


ROOT = Path(__file__).resolve().parents[2]
EXTENSION_PATH = ROOT / "dist" / "extension"


class LocalBrowserLaunchExtensionInjectorTests(unittest.TestCase):
    def test_rejects_zip_entries_outside_extraction_directory(self) -> None:
        with tempfile.TemporaryDirectory(prefix="modcdp-bad-zip-") as temp_dir:
            zip_path = Path(temp_dir) / "extension.zip"
            with zipfile.ZipFile(zip_path, "w") as archive:
                archive.writestr("../evil.txt", "evil")

            injector = LocalBrowserLaunchExtensionInjector({"injector_extension_path": str(zip_path)})
            try:
                with self.assertRaisesRegex(RuntimeError, "escapes extension extraction directory"):
                    injector.prepare()
                self.assertFalse((Path(temp_dir) / "evil.txt").exists())
            finally:
                injector.close()

    def test_prepares_unpacked_extension_directory_for_load_extension(self) -> None:
        injector = LocalBrowserLaunchExtensionInjector({"injector_extension_path": str(EXTENSION_PATH)})
        try:
            injector.prepare()
            unpacked_extension_path = injector.unpacked_extension_path
            self.assertIsInstance(unpacked_extension_path, str)
            unpacked_extension_path = cast(str, unpacked_extension_path)
            self.assertNotEqual(unpacked_extension_path, str(EXTENSION_PATH))
            self.assertTrue((Path(unpacked_extension_path) / "manifest.json").exists())
            self.assertEqual(injector.getLauncherConfig(), {"extra_args": [f"--load-extension={unpacked_extension_path}"]})
            self.assertEqual(injector.options.get("injector_extension_id"), DEFAULT_MODCDP_EXTENSION_ID)
        finally:
            injector.close()

    def test_prepares_default_extension_zip_for_load_extension(self) -> None:
        injector = LocalBrowserLaunchExtensionInjector()
        try:
            injector.prepare()
            unpacked_extension_path = injector.unpacked_extension_path
            self.assertIsInstance(unpacked_extension_path, str)
            unpacked_extension_path = cast(str, unpacked_extension_path)
            self.assertTrue((Path(unpacked_extension_path) / "manifest.json").exists())
            self.assertIn("modcdp-extension-", unpacked_extension_path)
            self.assertEqual(injector.getLauncherConfig(), {"extra_args": [f"--load-extension={unpacked_extension_path}"]})
            self.assertEqual(injector.options.get("injector_extension_id"), DEFAULT_MODCDP_EXTENSION_ID)
        finally:
            injector.close()

    def test_returns_immediately_when_launched_extension_target_is_absent(self) -> None:
        methods: list[str] = []

        def send(method: str, params: dict[str, Any] | None = None, session_id: str | None = None) -> dict[str, Any]:
            methods.append(method)
            if method == "Target.getTargets":
                return {"targetInfos": []}
            raise RuntimeError(f"unexpected {method}")

        injector = LocalBrowserLaunchExtensionInjector(
            cast(Any, {
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_trust_service_worker_target": True,
                "send": send,
            })
        )
        try:
            injector.prepare()
            started_at = time.perf_counter()
            result = injector.inject()
            elapsed_ms = (time.perf_counter() - started_at) * 1000
            self.assertIsNone(result)
            self.assertEqual(methods, ["Target.getTargets"])
            self.assertLess(elapsed_ms, 200)
        finally:
            injector.close()


if __name__ == "__main__":
    unittest.main()
