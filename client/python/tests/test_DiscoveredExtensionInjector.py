from __future__ import annotations

import unittest

from modcdp.DiscoveredExtensionInjector import DiscoveredExtensionInjector


class DiscoveredExtensionInjectorTests(unittest.TestCase):
    def test_returns_none_when_no_worker_is_visible(self) -> None:
        injector = DiscoveredExtensionInjector({"send": lambda method, params=None, session_id=None: {"targetInfos": []}})
        self.assertIsNone(injector.inject())


if __name__ == "__main__":
    unittest.main()
