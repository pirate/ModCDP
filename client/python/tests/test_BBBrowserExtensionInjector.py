from __future__ import annotations

import unittest

from modcdp.BBBrowserExtensionInjector import BBBrowserExtensionInjector


class BBBrowserExtensionInjectorTests(unittest.TestCase):
    def test_uses_configured_extension_id(self) -> None:
        injector = BBBrowserExtensionInjector({"extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
        injector.prepare()
        self.assertEqual(injector.getLauncherConfig()["extension_id"], "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")


if __name__ == "__main__":
    unittest.main()
