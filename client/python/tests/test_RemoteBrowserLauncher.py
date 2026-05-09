from __future__ import annotations

import json
import unittest

from websocket import create_connection

from modcdp.LocalBrowserLauncher import LocalBrowserLauncher
from modcdp.RemoteBrowserLauncher import RemoteBrowserLauncher


class RemoteBrowserLauncherTests(unittest.TestCase):
    def test_connects_to_real_browser_from_http_and_websocket_cdp_endpoints(self) -> None:
        local = LocalBrowserLauncher().launch({"headless": True, "sandbox": False, "chrome_ready_timeout_ms": 45_000})
        ws = None
        try:
            from_http = RemoteBrowserLauncher(cdp_url=local["cdp_url"]).launch()
            self.assertEqual(from_http["cdp_url"], local["cdp_url"])
            self.assertEqual(from_http["ws_url"], local["ws_url"])
            from_http_ws_url = from_http.get("ws_url")
            if not isinstance(from_http_ws_url, str):
                self.fail(f"ws_url = {from_http_ws_url!r}")
            ws = create_connection(from_http_ws_url, timeout=10)
            _expect_cdp_browser_surface(ws)
            from_http["close"]()

            from_ws = RemoteBrowserLauncher().launch({"ws_url": local["ws_url"]})
            self.assertEqual(from_ws["cdp_url"], local["ws_url"])
            self.assertEqual(from_ws["ws_url"], local["ws_url"])
            _expect_cdp_browser_surface(ws)
            from_ws["close"]()
        finally:
            if ws is not None:
                ws.close()
            local["close"]()


def _expect_cdp_browser_surface(ws) -> None:
    ws.send(json.dumps({"id": 1, "method": "Browser.getVersion", "params": {}}))
    message = json.loads(ws.recv())
    if not isinstance(message.get("result", {}).get("product"), str):
        raise AssertionError(f"Browser.getVersion result = {message!r}")


if __name__ == "__main__":
    unittest.main()
