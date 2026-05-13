from __future__ import annotations

import unittest

from modcdp.launcher.NoopBrowserLauncher import NoopBrowserLauncher


class NoopBrowserLauncherTests(unittest.TestCase):
    def test_constructor_launch_and_config_match_ts_shape(self) -> None:
        launcher = NoopBrowserLauncher({"cdp_url": "ws://127.0.0.1:9222/devtools/browser/initial"})
        self.assertEqual(launcher.options.get("cdp_url"), "ws://127.0.0.1:9222/devtools/browser/initial")
        self.assertEqual(
            launcher.getTransportConfig().get("cdp_url"),
            "ws://127.0.0.1:9222/devtools/browser/initial",
        )

        launched = launcher.launch({"cdp_url": "ws://127.0.0.1:9222/devtools/browser/call"})
        self.assertIs(launcher.launched, launched)
        self.assertIsNone(launched["cdp_url"])
        self.assertEqual(launcher.getServerConfig(), {})
        launched["close"]()


if __name__ == "__main__":
    unittest.main()
