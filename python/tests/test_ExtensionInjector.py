from __future__ import annotations

import json
import time
import unittest
from pathlib import Path
from typing import Any, cast

import modcdp
from modcdp.injector.ExtensionInjector import ExtensionInjector, ExtensionInjectorConfig
from modcdp.launcher.LocalBrowserLauncher import LocalBrowserLauncher
from modcdp.types import ProtocolParams, ProtocolResult
from websocket import create_connection


ROOT = Path(__file__).resolve().parents[2]
EXTENSION_PATH = ROOT / "dist" / "extension"


class ProbeExtensionInjector(ExtensionInjector):
    def inject(self):
        return self._waitForReadyServiceWorker(
            self.options.get("injector_service_worker_ready_timeout_ms") or 60_000,
            matched_only=True,
        )

    def sendTimed(
        self,
        method: str,
        params: ProtocolParams | None,
        session_id: str | None,
        timeout_ms: int,
    ) -> ProtocolResult:
        return self._sendWithTimeout(method, params or {}, session_id, timeout_ms)


class ExtensionInjectorTests(unittest.TestCase):
    def test_probes_real_extension_service_worker_with_shared_base_config(self) -> None:
        chrome = LocalBrowserLauncher(
            {
                "headless": True,
                "sandbox": False,
                "extra_args": [f"--load-extension={EXTENSION_PATH}"],
            }
        ).launch()
        ws = create_connection(cast(str, chrome["cdp_url"]), timeout=10)
        next_id = 0

        def send(method: str, params: ProtocolParams | None = None, session_id: str | None = None) -> ProtocolResult:
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
                return cast(ProtocolResult, response.get("result") or {})

        def attach_to_target(target_id: str) -> str | None:
            result = send("Target.attachToTarget", {"targetId": target_id, "flatten": True})
            session_id = result.get("sessionId")
            return session_id if isinstance(session_id, str) else None

        injector = ProbeExtensionInjector(
            cast(ExtensionInjectorConfig, {
                "send": send,
                "attachToTarget": attach_to_target,
                "injector_extension_id": "mdedooklbnfejodmnhmkdpkaedafkehf",
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            })
        )

        try:
            self.assertEqual(injector.getLauncherConfig(), {})
            self.assertEqual(injector.getTransportConfig(), {"injector_extension_id": "mdedooklbnfejodmnhmkdpkaedafkehf"})
            result = injector.inject()
            self.assertEqual(result["extension_id"] if result else None, "mdedooklbnfejodmnhmkdpkaedafkehf")
            self.assertTrue(str(result["url"] if result else "").endswith("/modcdp/service_worker.js"))
        finally:
            injector.close()
            ws.close()
            chrome["close"]()

    def test_keeps_modcdp_service_worker_alive_through_offscreen_keepalive(self) -> None:
        chrome = LocalBrowserLauncher(
            {
                "headless": True,
                "sandbox": False,
                "extra_args": [f"--load-extension={EXTENSION_PATH}"],
            }
        ).launch()
        ws = create_connection(cast(str, chrome["cdp_url"]), timeout=10)
        next_id = 0

        def send(method: str, params: ProtocolParams | None = None, session_id: str | None = None) -> ProtocolResult:
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
                return cast(ProtocolResult, response.get("result") or {})

        def attach_to_target(target_id: str) -> str | None:
            result = send("Target.attachToTarget", {"targetId": target_id, "flatten": True})
            session_id = result.get("sessionId")
            return session_id if isinstance(session_id, str) else None

        injector = ProbeExtensionInjector(
            cast(ExtensionInjectorConfig, {
                "send": send,
                "attachToTarget": attach_to_target,
                "injector_extension_id": "mdedooklbnfejodmnhmkdpkaedafkehf",
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            })
        )

        try:
            result = injector.inject()
            self.assertEqual(result["extension_id"] if result else None, "mdedooklbnfejodmnhmkdpkaedafkehf")
            session_id = result["session_id"] if result else None
            self.assertIsInstance(session_id, str)

            value: list[Any] = []
            for _ in range(50):
                contexts = send(
                    "Runtime.evaluate",
                    {
                        "expression": (
                            "chrome.runtime.getContexts({}).then((contexts) => contexts.map((context) => "
                            "({ type: context.contextType, url: context.documentUrl || context.origin || '' })))"
                        ),
                        "awaitPromise": True,
                        "returnByValue": True,
                    },
                    cast(str, session_id),
                )
                raw_value = cast(dict[str, Any], contexts.get("result") or {}).get("value")
                value = raw_value if isinstance(raw_value, list) else []
                if any(
                    isinstance(context, dict)
                    and context.get("type") == "OFFSCREEN_DOCUMENT"
                    and context.get("url")
                    == "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/offscreen/keepalive.html"
                    for context in value
                ):
                    break
                time.sleep(0.1)
            self.assertTrue(
                any(
                    isinstance(context, dict)
                    and context.get("type") == "OFFSCREEN_DOCUMENT"
                    and context.get("url")
                    == "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/offscreen/keepalive.html"
                    for context in value
                )
            )

            time.sleep(3)
            targets = cast(dict[str, Any], send("Target.getTargets"))
            target_infos = targets.get("targetInfos")
            self.assertIsInstance(target_infos, list)
            target_infos = cast(list[Any], target_infos)
            self.assertTrue(
                any(
                    isinstance(target, dict)
                    and target.get("type") == "service_worker"
                    and target.get("url")
                    == "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js"
                    for target in target_infos
                )
            )
            version = send(
                "Runtime.evaluate",
                {"expression": "globalThis.ModCDP?.__ModCDPServerVersion", "returnByValue": True},
                cast(str, session_id),
            )
            self.assertEqual(cast(dict[str, Any], version.get("result") or {}).get("value"), 2)
        finally:
            injector.close()
            ws.close()
            chrome["close"]()

    def test_owns_shared_injector_config_and_runtime_transport_config(self) -> None:
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

        with self.assertRaisesRegex(NotImplementedError, "ExtensionInjector.inject is not implemented"):
            injector.inject()

    def test_send_with_timeout_enforces_cdp_send_timeout(self) -> None:
        chrome = LocalBrowserLauncher({"headless": True, "sandbox": False}).launch()
        ws = create_connection(cast(str, chrome["cdp_url"]), timeout=10)
        next_id = 0

        def send(method: str, params: ProtocolParams | None = None, session_id: str | None = None) -> ProtocolResult:
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
                return cast(ProtocolResult, response.get("result") or {})

        injector = ProbeExtensionInjector(cast(ExtensionInjectorConfig, {"send": send}))
        target_id: str | None = None

        try:
            created = send("Target.createTarget", {"url": "about:blank#modcdp-timeout"})
            target_id = cast(str, created["targetId"])
            attached = send("Target.attachToTarget", {"targetId": target_id, "flatten": True})
            session_id = cast(str, attached["sessionId"])
            send("Runtime.enable", {}, session_id)
            with self.assertRaisesRegex(TimeoutError, r"Runtime\.evaluate timed out after 5ms"):
                injector.sendTimed(
                    "Runtime.evaluate",
                    {"expression": "new Promise(() => {})", "awaitPromise": True},
                    session_id,
                    5,
                )
        finally:
            if target_id:
                try:
                    send("Target.closeTarget", {"targetId": target_id})
                except Exception:
                    pass
            injector.close()
            ws.close()
            chrome["close"]()

    def test_wakes_configured_extension_with_hidden_background_target(self) -> None:
        chrome = LocalBrowserLauncher(
            {
                "headless": True,
                "sandbox": False,
                "extra_args": [f"--load-extension={EXTENSION_PATH}"],
            }
        ).launch()
        ws = create_connection(cast(str, chrome["cdp_url"]), timeout=10)
        next_id = 0

        def send(method: str, params: ProtocolParams | None = None, _session_id: str | None = None) -> ProtocolResult:
            nonlocal next_id
            next_id += 1
            message: dict[str, Any] = {"id": next_id, "method": method, "params": params or {}}
            ws.send(json.dumps(message))
            while True:
                response = json.loads(ws.recv())
                if response.get("id") != next_id:
                    continue
                error = response.get("error")
                if isinstance(error, dict):
                    raise RuntimeError(str(error.get("message") or error))
                return cast(ProtocolResult, response.get("result") or {})

        injector = ProbeExtensionInjector(
            cast(ExtensionInjectorConfig, {
                "injector_extension_id": "mdedooklbnfejodmnhmkdpkaedafkehf",
                "send": send,
            })
        )

        try:
            self.assertTrue(injector._wakeConfiguredExtension())
            targets = cast(dict[str, Any], send("Target.getTargets"))
            target_infos = targets.get("targetInfos")
            self.assertIsInstance(target_infos, list)
            target_infos = cast(list[Any], target_infos)
            self.assertTrue(
                any(
                    isinstance(target, dict)
                    and target.get("url") == "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/wake.html"
                    for target in target_infos
                )
            )
        finally:
            injector.close()
            ws.close()
            chrome["close"]()

    def test_package_exports_all_injector_classes(self) -> None:
        self.assertIs(modcdp.ExtensionInjector, ExtensionInjector)
        self.assertEqual(modcdp.DiscoveredExtensionInjector.__name__, "DiscoveredExtensionInjector")
        self.assertEqual(modcdp.LocalBrowserLaunchExtensionInjector.__name__, "LocalBrowserLaunchExtensionInjector")
        self.assertEqual(modcdp.BBBrowserExtensionInjector.__name__, "BBBrowserExtensionInjector")
        self.assertEqual(modcdp.ExtensionsLoadUnpackedInjector.__name__, "ExtensionsLoadUnpackedInjector")
        self.assertEqual(modcdp.BorrowedExtensionInjector.__name__, "BorrowedExtensionInjector")


if __name__ == "__main__":
    unittest.main()
