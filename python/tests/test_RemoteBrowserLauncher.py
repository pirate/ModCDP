from __future__ import annotations

import json
import unittest

from websocket import create_connection

from modcdp.launcher.LocalBrowserLauncher import LocalBrowserLauncher
from modcdp.launcher.RemoteBrowserLauncher import RemoteBrowserLauncher


class RemoteBrowserLauncherTests(unittest.TestCase):
    def test_requires_upstream_cdp_url(self) -> None:
        with self.assertRaisesRegex(RuntimeError, "launcher.launcher_mode=remote requires upstream.upstream_cdp_url"):
            RemoteBrowserLauncher().launch()

    def test_connects_to_real_browser_from_http_and_websocket_cdp_endpoints(self) -> None:
        port = LocalBrowserLauncher.freePort()
        local = LocalBrowserLauncher().launch(
            {"port": port, "headless": True, "sandbox": False, "chrome_ready_timeout_ms": 45_000}
        )
        ws = None
        try:
            from_http = RemoteBrowserLauncher(cdp_url=f"http://127.0.0.1:{port}").launch()
            self.assertEqual(from_http["cdp_url"], local["cdp_url"])
            from_http_cdp_url = from_http.get("cdp_url")
            if not isinstance(from_http_cdp_url, str):
                self.fail(f"cdp_url = {from_http_cdp_url!r}")
            ws = create_connection(from_http_cdp_url, timeout=10)
            _expect_cdp_browser_surface(ws)
            from_http["close"]()

            from_host_port = RemoteBrowserLauncher(cdp_url=f"127.0.0.1:{port}").launch()
            self.assertEqual(from_host_port["cdp_url"], local["cdp_url"])
            from_host_port["close"]()

            from_options = RemoteBrowserLauncher({"cdp_url": local["cdp_url"]}).launch()
            self.assertEqual(from_options["cdp_url"], local["cdp_url"])
            from_options["close"]()

            from_override = RemoteBrowserLauncher({"cdp_url": "http://127.0.0.1:1"}).launch({"cdp_url": f"127.0.0.1:{port}"})
            self.assertEqual(from_override["cdp_url"], local["cdp_url"])
            from_override["close"]()

            from_ws = RemoteBrowserLauncher().launch({"cdp_url": local["cdp_url"]})
            self.assertEqual(from_ws["cdp_url"], local["cdp_url"])
            _expect_cdp_browser_surface(ws)
            from_ws["close"]()
        finally:
            if ws is not None:
                ws.close()
            local["close"]()

    def test_accepts_wss_cdp_endpoint_without_http_discovery(self) -> None:
        launched = RemoteBrowserLauncher(cdp_url="wss://example.test/devtools/browser/test").launch()

        self.assertEqual(launched["cdp_url"], "wss://example.test/devtools/browser/test")
        launched["close"]()


def _expect_cdp_browser_surface(ws) -> None:
    ws.send(json.dumps({"id": 1, "method": "Browser.getVersion", "params": {}}))
    message = json.loads(ws.recv())
    if not isinstance(message.get("result", {}).get("product"), str):
        raise AssertionError(f"Browser.getVersion result = {message!r}")


if __name__ == "__main__":
    unittest.main()
