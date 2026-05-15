from __future__ import annotations

import unittest

from modcdp.injector.ExtensionInjector import ExtensionInjector


class ExtensionInjectorTests(unittest.TestCase):
    def test_owns_shared_injector_config(self) -> None:
        injector = ExtensionInjector(
            {
                "injector_extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
            }
        )

        self.assertEqual(injector.getTransportConfig(), {"injector_extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
        self.assertEqual(injector.getLauncherConfig(), {})
        self.assertTrue(
            injector._serviceWorkerTargetMatches(
                {
                    "targetId": "target-1",
                    "type": "service_worker",
                    "url": "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/modcdp/service_worker.js",
                }
            )
        )
        self.assertFalse(
            injector._serviceWorkerTargetMatches(
                {
                    "targetId": "target-1",
                    "type": "service_worker",
                    "url": "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/background.js",
                }
            )
        )

    def test_base_inject_reports_the_class_name(self) -> None:
        with self.assertRaisesRegex(NotImplementedError, "ExtensionInjector.inject is not implemented"):
            ExtensionInjector().inject()


if __name__ == "__main__":
    unittest.main()
