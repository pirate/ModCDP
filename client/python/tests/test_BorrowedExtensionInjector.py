from __future__ import annotations

import unittest

from modcdp.BorrowedExtensionInjector import BorrowedExtensionInjector


class BorrowedExtensionInjectorTests(unittest.TestCase):
    def test_returns_none_when_no_worker_is_visible(self) -> None:
        injector = BorrowedExtensionInjector(
            {
                "send": lambda method, params=None, session_id=None: {"targetInfos": []},
                "service_worker_ready_timeout_ms": 1,
            }
        )
        self.assertIsNone(injector.inject())


if __name__ == "__main__":
    unittest.main()
