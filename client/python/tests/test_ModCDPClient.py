from __future__ import annotations

import asyncio
import json
import time
import unittest
from collections.abc import Mapping
from contextlib import redirect_stderr
from io import StringIO
from pathlib import Path
from typing import Any, cast

from websocket import create_connection

from modcdp import ModCDPClient
from modcdp.cdp import AwaitableDict
from modcdp.LocalBrowserLauncher import LocalBrowserLauncher
from modcdp.types import JsonValue


HERE = Path(__file__).resolve().parent
EXTENSION_PATH = HERE.parents[2] / "dist" / "extension"


class ModCDPClientTests(unittest.TestCase):
    def test_constructor_normalizes_nested_config_owners(self) -> None:
        cdp = ModCDPClient(
            launch={
                "mode": "local",
                "executable_path": "/tmp/chrome",
                "user_data_dir": "/tmp/profile",
                "options": {"headless": True},
            },
            upstream={
                "mode": "ws",
                "ws_url": "http://127.0.0.1:9222",
                "reversews_wait_timeout_ms": 456,
                "ws_connect_error_settle_timeout_ms": 321,
            },
            extension={
                "mode": "discover",
                "path": "/tmp/ext",
                "extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "service_worker_url_includes": ["modcdp"],
                "service_worker_url_suffixes": ["/custom/service_worker.js"],
                "trust_service_worker_target": True,
                "require_service_worker_target": True,
                "execution_context_timeout_ms": 4321,
                "service_worker_probe_timeout_ms": 5432,
                "service_worker_ready_timeout_ms": 6543,
                "service_worker_poll_interval_ms": 76,
                "target_session_poll_interval_ms": 87,
            },
            client={
                "routes": {"*.*": "direct_cdp"},
                "mirror_upstream_events": False,
                "cdp_send_timeout_ms": 1234,
                "event_wait_timeout_ms": 2345,
            },
            server={
                "routes": {"*.*": "loopback_cdp"},
                "browser_token": "token-1",
                "cdp_send_timeout_ms": 9876,
                "loopback_execution_context_timeout_ms": 8765,
                "ws_connect_error_settle_timeout_ms": 7654,
            },
        )

        self.assertEqual(cdp.launch["options"], {"headless": True})
        self.assertEqual(cdp._launch_options()["executable_path"], "/tmp/chrome")
        self.assertEqual(cdp._launch_options()["user_data_dir"], "/tmp/profile")
        self.assertEqual(cdp.upstream["reversews_wait_timeout_ms"], 456)
        self.assertEqual(cdp.upstream["ws_connect_error_settle_timeout_ms"], 321)
        self.assertEqual(cdp.extension["execution_context_timeout_ms"], 4321)
        self.assertEqual(cdp.extension["service_worker_probe_timeout_ms"], 5432)
        self.assertEqual(cdp.extension["service_worker_ready_timeout_ms"], 6543)
        self.assertEqual(cdp.extension["service_worker_poll_interval_ms"], 76)
        self.assertEqual(cdp.extension["target_session_poll_interval_ms"], 87)
        self.assertEqual(cdp.client["routes"]["*.*"], "direct_cdp")
        self.assertEqual(cdp.client["mirror_upstream_events"], False)
        self.assertEqual(cdp.client["cdp_send_timeout_ms"], 1234)
        self.assertEqual(cdp.client["event_wait_timeout_ms"], 2345)
        self.assertNotIn("routes", cdp.__dict__)
        self.assertNotIn("cdp_send_timeout_ms", cdp.__dict__)
        self.assertNotIn("service_worker_probe_timeout_ms", cdp.__dict__)

        params = cast(dict[str, Any], cdp._server_configure_params())
        self.assertEqual(params["client"]["routes"]["*.*"], "direct_cdp")
        self.assertEqual(params["server"]["browser_token"], "token-1")
        self.assertEqual(params["server"]["cdp_send_timeout_ms"], 9876)
        self.assertEqual(params["server"]["loopback_execution_context_timeout_ms"], 8765)
        self.assertEqual(params["server"]["ws_connect_error_settle_timeout_ms"], 7654)

    def test_connects_with_local_launch_and_injector_chain(self) -> None:
        cdp = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "ws"},
            extension={
                "mode": "inject",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.connect_timing["extension_source"] if cdp.connect_timing else None, "local_launch")
            self.assertEqual(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf")
            self.assertEqual(
                cdp.Mod.evaluate(expression="chrome.runtime.getURL('modcdp/service_worker.js')"),
                "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js",
            )
        finally:
            cdp.close()

    def test_close_does_not_close_a_remote_browser_it_did_not_launch(self) -> None:
        chrome = LocalBrowserLauncher(
            {
                "headless": True,
                "sandbox": False,
                "extra_args": [f"--load-extension={EXTENSION_PATH}"],
            }
        ).launch()
        raw_ws = create_connection(cast(str, chrome["ws_url"]), timeout=5)
        cdp = ModCDPClient(
            launch={"mode": "remote"},
            upstream={"mode": "ws", "ws_url": chrome["cdp_url"]},
            extension={
                "mode": "discover",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
        )

        try:
            cdp.connect()
            cdp.close()
            time.sleep(0.5)
            raw_ws.send(json.dumps({"id": 1, "method": "Browser.getVersion", "params": {}}))
            response = json.loads(raw_ws.recv())
            self.assertEqual(response["id"], 1)
            self.assertRegex(response["result"]["product"], r"Chrome|Chromium")
        finally:
            raw_ws.close()
            cdp.close()
            chrome["close"]()

    def test_close_keeps_injector_files_until_after_launched_browser_shutdown(self) -> None:
        order: list[str] = []
        cdp = ModCDPClient()

        class FakeTransport:
            def close(self) -> None:
                order.append("transport")

        class FakeInjector:
            def close(self) -> None:
                order.append("injector")

        cdp.transport = cast(Any, FakeTransport())
        cdp._launched_browser = {"close": lambda: order.append("browser")}
        cdp._extension_injectors = cast(Any, [FakeInjector()])

        cdp.close()

        self.assertEqual(order, ["transport", "browser", "injector"])
        self.assertIsNone(cdp.transport)
        self.assertIsNone(cdp._launched_browser)
        self.assertEqual(cdp._extension_injectors, [])

    def test_generated_cdp_surface_exposes_direct_domain_commands(self) -> None:
        sent: list[tuple[str, dict[str, object], str | None, bool]] = []

        class RecordingClient(ModCDPClient):
            def _send_command(
                self,
                method: str,
                params: Mapping[str, object] | None = None,
                session_id: str | None = None,
                validate_custom_schema: bool = True,
            ) -> JsonValue:
                sent.append((method, dict(params or {}), session_id, validate_custom_schema))
                return AwaitableDict({"targetId": "target-1"})

        client = RecordingClient(client={"routes": {"Custom.*": "direct_cdp"}})

        result = client.Target.createTarget(url="https://example.com", session_id="session-1")
        raw_result = client.send("Target.createTarget", {"url": "https://example.com"})

        self.assertEqual(result.targetId, "target-1")
        self.assertEqual(raw_result, {"targetId": "target-1"})
        self.assertEqual(sent[0], ("Target.createTarget", {"url": "https://example.com"}, "session-1", True))
        self.assertEqual(sent[1], ("Target.createTarget", {"url": "https://example.com"}, None, False))
        with self.assertRaises(Exception):
            client.Target._CreateTargetParams.model_validate({"url": "https://example.com", "unknown": True})

        async def run_awaited_calls() -> None:
            awaited_result = await client.Target.createTarget(url="https://example.com")
            awaited_raw_result = await client.send("Target.createTarget", {"url": "https://example.com"})
            self.assertEqual(awaited_result.targetId, "target-1")
            self.assertEqual(awaited_raw_result["targetId"], "target-1")

        asyncio.run(run_awaited_calls())

    def test_generated_event_surface_supports_awaited_on_and_async_callbacks(self) -> None:
        client = ModCDPClient()
        seen: list[str] = []

        async def callback(event: Any) -> None:
            seen.append(event.targetId)

        async def register() -> None:
            await client.on(client.Target.targetCreated, callback)

        asyncio.run(register())
        client._run_handler(
            client._handlers["Target.targetCreated"][0],
            {"targetInfo": {"targetId": "target-1", "type": "page", "url": "https://example.com"}},
            "Target.targetCreated",
        )
        self.assertEqual(seen, ["target-1"])

    def test_schema_only_custom_command_registers_without_websocket(self) -> None:
        client = ModCDPClient()

        result = client.send(
            "Mod.addCustomCommand",
            {
                "name": "Custom.echo",
                "params_schema": {
                    "type": "object",
                    "properties": {"text": {"type": "string", "minLength": 1}},
                    "required": ["text"],
                    "additionalProperties": False,
                },
                "result_schema": {
                    "type": "object",
                    "properties": {"text": {"type": "string"}},
                    "required": ["text"],
                    "additionalProperties": False,
                },
            },
        )

        self.assertEqual(result, {"name": "Custom.echo", "registered": True})
        self.assertEqual(client._validate_command_params("Custom.echo", {"text": "ok"}), {"text": "ok"})
        with self.assertRaises(ValueError):
            client._validate_command_params("Custom.echo", {"text": ""})
        with self.assertRaises(ValueError):
            client._validate_command_params("Custom.echo", {"text": "ok", "extra": True})
        self.assertEqual(client._validate_command_result("Custom.echo", {"text": "ok"}), {"text": "ok"})
        with self.assertRaises(ValueError):
            client._validate_command_result("Custom.echo", {"text": 123})

    def test_constructor_custom_command_schemas_validate_nested_json(self) -> None:
        client = ModCDPClient(
            custom_commands=[
                {
                    "name": "Custom.collect",
                    "params_schema": {
                        "type": "object",
                        "properties": {
                            "items": {
                                "type": "array",
                                "minItems": 1,
                                "items": {
                                    "type": "object",
                                    "properties": {
                                        "id": {"type": "string"},
                                        "count": {"type": "integer", "minimum": 1},
                                    },
                                    "required": ["id", "count"],
                                    "additionalProperties": False,
                                },
                            }
                        },
                        "required": ["items"],
                        "additionalProperties": False,
                    },
                }
            ]
        )

        valid: dict[str, JsonValue] = {"items": [{"id": "a", "count": 1}]}
        self.assertEqual(client._validate_command_params("Custom.collect", valid), valid)
        with self.assertRaises(ValueError):
            client._validate_command_params("Custom.collect", {"items": [{"id": "a", "count": 0}]})
        with self.assertRaises(ValueError):
            client._validate_command_params("Custom.collect", {"items": []})

    def test_custom_event_schema_validates_payload_before_handlers(self) -> None:
        client = ModCDPClient(
            custom_events=[
                {
                    "name": "Custom.ready",
                    "event_schema": {
                        "type": "object",
                        "properties": {
                            "url": {"type": "string", "pattern": "^https://"},
                            "ready": {"type": "boolean"},
                        },
                        "required": ["url", "ready"],
                        "additionalProperties": False,
                    },
                }
            ]
        )

        payload: dict[str, JsonValue] = {"url": "https://example.com", "ready": True}
        self.assertEqual(client._validate_event_payload("Custom.ready", payload), payload)
        with redirect_stderr(StringIO()):
            self.assertIsNone(client._validate_event_payload("Custom.ready", {"url": "http://example.com", "ready": True}))
            self.assertIsNone(
                client._validate_event_payload("Custom.ready", {"url": "https://example.com", "ready": True, "x": 1})
            )

    def test_scalar_event_schema_validates_value_payloads(self) -> None:
        client = ModCDPClient(custom_events=[{"name": "Custom.count", "event_schema": {"type": "integer", "minimum": 1}}])

        self.assertEqual(client._validate_event_payload("Custom.count", {"value": 3}), {"value": 3})
        with redirect_stderr(StringIO()):
            self.assertIsNone(client._validate_event_payload("Custom.count", {"value": 0}))


if __name__ == "__main__":
    unittest.main()
