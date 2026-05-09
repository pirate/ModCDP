from __future__ import annotations

import unittest

from modcdp.NoopBrowserLauncher import NoopBrowserLauncher


class NoopBrowserLauncherTests(unittest.TestCase):
    def test_uses_no_browser_lifecycle_and_returns_no_cdp_endpoints(self) -> None:
        browser = NoopBrowserLauncher(
            {
                "cdp_url": "ws://127.0.0.1:1/devtools/browser/not-used",
                "user_data_dir": "/tmp/not-used-by-noop",
            }
        ).launch()
        self.assertIsNone(browser.get("cdp_url"))
        self.assertIsNone(browser.get("ws_url"))
        self.assertNotIn("pipe_read", browser)
        self.assertNotIn("pipe_write", browser)
        browser["close"]()
        browser["close"]()


if __name__ == "__main__":
    unittest.main()
