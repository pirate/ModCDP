from __future__ import annotations

import unittest
from pathlib import Path

from modcdp.ExtensionInjector import DEFAULT_MODCDP_EXTENSION_ID
from modcdp.LocalBrowserLaunchExtensionInjector import LocalBrowserLaunchExtensionInjector


ROOT = Path(__file__).resolve().parents[3]
EXTENSION_PATH = ROOT / "dist" / "extension"


class LocalBrowserLaunchExtensionInjectorTests(unittest.TestCase):
    def test_prepares_launcher_config(self) -> None:
        injector = LocalBrowserLaunchExtensionInjector({"extension_path": str(EXTENSION_PATH)})
        try:
            injector.prepare()
            extra_args = injector.getLauncherConfig()["extra_args"]
            self.assertEqual(len(extra_args), 1)
            self.assertTrue(extra_args[0].startswith("--load-extension="))
            self.assertEqual(injector.options["extension_id"], DEFAULT_MODCDP_EXTENSION_ID)
        finally:
            injector.close()


if __name__ == "__main__":
    unittest.main()
