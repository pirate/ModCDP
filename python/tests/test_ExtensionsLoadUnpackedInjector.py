from __future__ import annotations

import unittest
from pathlib import Path
from typing import cast

from modcdp.injector.ExtensionsLoadUnpackedInjector import ExtensionsLoadUnpackedInjector


class ExtensionsLoadUnpackedInjectorTests(unittest.TestCase):
    def test_prepares_default_packaged_extension_zip(self) -> None:
        injector = ExtensionsLoadUnpackedInjector()
        try:
            injector.prepare()
            unpacked_extension_path = injector.unpacked_extension_path
            self.assertIsInstance(unpacked_extension_path, str)
            unpacked_extension_path = cast(str, unpacked_extension_path)
            self.assertIn("modcdp-extension-", unpacked_extension_path)
            self.assertTrue((Path(unpacked_extension_path) / "manifest.json").exists())
        finally:
            injector.close()


if __name__ == "__main__":
    unittest.main()
