from __future__ import annotations

import unittest
from pathlib import Path

from modcdp.ExtensionsLoadUnpackedInjector import ExtensionsLoadUnpackedInjector


ROOT = Path(__file__).resolve().parents[3]
EXTENSION_PATH = ROOT / "dist" / "extension"


class ExtensionsLoadUnpackedInjectorTests(unittest.TestCase):
    def test_prepares_runtime_config_copy(self) -> None:
        injector = ExtensionsLoadUnpackedInjector(
            {
                "extension_path": str(EXTENSION_PATH),
                "reverse_proxy_url": "ws://127.0.0.1:29292",
            }
        )
        try:
            injector.prepare()
            self.assertNotEqual(injector.unpacked_extension_path, str(EXTENSION_PATH))
            config = Path(injector.unpacked_extension_path or "", "modcdp", "config.json").read_text()
            self.assertEqual(config, '{\n  "reverse_proxy_url": "ws://127.0.0.1:29292"\n}\n')
        finally:
            injector.close()


if __name__ == "__main__":
    unittest.main()
