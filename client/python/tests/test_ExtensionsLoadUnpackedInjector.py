from __future__ import annotations

import json
import re
import unittest
from pathlib import Path
from typing import Any, cast

from modcdp.ExtensionsLoadUnpackedInjector import ExtensionsLoadUnpackedInjector
from modcdp.LocalBrowserLauncher import LocalBrowserLauncher
from websocket import create_connection


ROOT = Path(__file__).resolve().parents[3]
EXTENSION_PATH = ROOT / "dist" / "extension"


class ExtensionsLoadUnpackedInjectorTests(unittest.TestCase):
    def test_exercises_real_cdp_load_unpacked_path(self) -> None:
        chrome = LocalBrowserLauncher({"headless": True, "sandbox": False}).launch()
        ws = create_connection(cast(str, chrome["ws_url"]), timeout=10)
        next_id = 0

        def send(method: str, params: dict[str, Any] | None = None, session_id: str | None = None) -> dict[str, Any]:
            nonlocal next_id
            next_id += 1
            message: dict[str, Any] = {"id": next_id, "method": method, "params": params or {}}
            if session_id:
                message["sessionId"] = session_id
            ws.send(json.dumps(message))
            while True:
                response = json.loads(ws.recv())
                if response.get("id") != next_id:
                    continue
                error = response.get("error")
                if isinstance(error, dict):
                    raise RuntimeError(str(error.get("message") or error))
                return cast(dict[str, Any], response.get("result") or {})

        injector = ExtensionsLoadUnpackedInjector(
            cast(Any, {
                "send": send,
                "extension_path": str(EXTENSION_PATH),
            })
        )
        try:
            injector.prepare()
            result = injector.inject()
            self.assertIsNone(result)
            self.assertIsNotNone(injector.last_error)
            self.assertRegex(
                str(injector.last_error),
                re.compile(r"Method not available|Method.*not.*found|wasn't found", re.I),
            )
        finally:
            injector.close()
            ws.close()
            chrome["close"]()

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
