from __future__ import annotations

import json
import unittest
from unittest.mock import patch

from modcdp.launcher.BrowserLauncher import BrowserLauncher, resolveCdpWebSocketUrl


class BrowserLauncherTests(unittest.TestCase):
    def test_merges_launch_config_and_exposes_transport_and_injector_config(self) -> None:
        launcher = BrowserLauncher(
            {
                "cdp_url": "ws://127.0.0.1:9222/devtools/browser/initial",
                "user_data_dir": "/tmp/modcdp-browser-launcher",
                "browserbase_api_key": "test-key",
                "injector_extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "args": ["--load-extension=/tmp/args-one"],
                "extra_args": ["--load-extension=/tmp/one"],
            }
        )
        launcher.update(
            {
                "cdp_url": "ws://127.0.0.1:9222/devtools/browser/updated",
                "args": ["--load-extension=/tmp/args-two", "--lang=en-US"],
                "extra_args": ["--load-extension=/tmp/two", "--window-size=900,700"],
            }
        )

        self.assertEqual(
            launcher.options.get("args"),
            ["--lang=en-US", "--load-extension=/tmp/args-one,/tmp/args-two"],
        )
        self.assertEqual(
            launcher.options.get("extra_args"),
            ["--window-size=900,700", "--load-extension=/tmp/one,/tmp/two"],
        )
        self.assertEqual(
            {
                "cdp_url": launcher.getTransportConfig()["cdp_url"],
                "user_data_dir": launcher.getTransportConfig()["user_data_dir"],
            },
            {
                "cdp_url": "ws://127.0.0.1:9222/devtools/browser/updated",
                "user_data_dir": "/tmp/modcdp-browser-launcher",
            },
        )
        self.assertEqual(
            {
                "injector_browserbase_api_key": launcher.getInjectorConfig()["injector_browserbase_api_key"],
                "injector_extension_id": launcher.getInjectorConfig()["injector_extension_id"],
            },
            {
                "injector_browserbase_api_key": "test-key",
                "injector_extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
            },
        )
        with self.assertRaisesRegex(NotImplementedError, "BrowserLauncher.launch is not implemented"):
            launcher.launch()

    def test_resolve_cdp_websocket_url_accepts_host_http_https_ws_and_wss_shapes(self) -> None:
        self.assertEqual(
            resolveCdpWebSocketUrl("ws://127.0.0.1:9222/devtools/browser/one"),
            "ws://127.0.0.1:9222/devtools/browser/one",
        )
        self.assertEqual(
            resolveCdpWebSocketUrl("wss://example.test/devtools/browser/two"),
            "wss://example.test/devtools/browser/two",
        )

        class FakeResponse:
            def __init__(self, cdp_url: str) -> None:
                self.cdp_url = cdp_url

            def __enter__(self) -> "FakeResponse":
                return self

            def __exit__(self, *_args: object) -> None:
                return None

            def read(self) -> bytes:
                return json.dumps({"webSocketDebuggerUrl": self.cdp_url}).encode()

        with patch("urllib.request.urlopen", return_value=FakeResponse("ws://127.0.0.1:9222/devtools/browser/three")) as urlopen:
            self.assertEqual(resolveCdpWebSocketUrl("127.0.0.1:9222"), "ws://127.0.0.1:9222/devtools/browser/three")
            urlopen.assert_called_once_with("http://127.0.0.1:9222/json/version", timeout=10)

        with patch("urllib.request.urlopen", return_value=FakeResponse("wss://example.test/devtools/browser/four")) as urlopen:
            self.assertEqual(resolveCdpWebSocketUrl("https://example.test"), "wss://example.test/devtools/browser/four")
            urlopen.assert_called_once_with("https://example.test/json/version", timeout=10)


if __name__ == "__main__":
    unittest.main()
