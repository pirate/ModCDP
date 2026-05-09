from __future__ import annotations

import asyncio
import json
import time
import unittest
from collections.abc import Mapping
from contextlib import redirect_stderr
from io import StringIO
from pathlib import Path
from queue import Queue
from typing import Any, cast

from websocket import create_connection

from modcdp import ModCDPClient
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
                "nativemessaging_host_name": "com.modcdp.custom",
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
                "hydrate_aliases": False,
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
        self.assertEqual(cdp._launch_options().get("executable_path"), "/tmp/chrome")
        self.assertEqual(cdp._launch_options().get("user_data_dir"), "/tmp/profile")
        self.assertEqual(cdp.upstream["reversews_wait_timeout_ms"], 456)
        self.assertEqual(cdp.upstream["nativemessaging_host_name"], "com.modcdp.custom")
        self.assertEqual(cdp.upstream["ws_connect_error_settle_timeout_ms"], 321)
        self.assertEqual(cdp.extension["execution_context_timeout_ms"], 4321)
        self.assertEqual(cdp.extension["service_worker_probe_timeout_ms"], 5432)
        self.assertEqual(cdp.extension["service_worker_ready_timeout_ms"], 6543)
        self.assertEqual(cdp.extension["service_worker_poll_interval_ms"], 76)
        self.assertEqual(cdp.extension["target_session_poll_interval_ms"], 87)
        self.assertEqual(cdp.client["routes"]["*.*"], "direct_cdp")
        self.assertEqual(cdp.client["hydrate_aliases"], False)
        self.assertEqual(cdp.client["mirror_upstream_events"], False)
        self.assertEqual(cdp.client["cdp_send_timeout_ms"], 1234)
        self.assertEqual(cdp.client["event_wait_timeout_ms"], 2345)
        self.assertNotIn("Browser", cdp.__dict__)
        with self.assertRaises(AttributeError):
            _ = cdp.Browser
        self.assertNotIn("routes", cdp.__dict__)
        self.assertNotIn("cdp_send_timeout_ms", cdp.__dict__)
        self.assertNotIn("service_worker_probe_timeout_ms", cdp.__dict__)

        params = cast(dict[str, Any], cdp._server_configure_params())
        self.assertEqual(params["client"]["routes"]["*.*"], "direct_cdp")
        self.assertEqual(params["server"]["browser_token"], "token-1")
        self.assertEqual(params["server"]["cdp_send_timeout_ms"], 9876)
        self.assertEqual(params["server"]["loopback_execution_context_timeout_ms"], 8765)
        self.assertEqual(params["server"]["ws_connect_error_settle_timeout_ms"], 7654)

    def test_preserves_explicit_empty_service_worker_suffix_config(self) -> None:
        cdp = ModCDPClient(extension={"mode": "borrow", "service_worker_url_suffixes": []})

        self.assertEqual(cdp.extension["service_worker_url_suffixes"], [])
        self.assertEqual(cdp._base_extension_injector_config(None).get("service_worker_url_suffixes"), [])

    def test_defaults_service_worker_suffix_config_to_modcdp_worker(self) -> None:
        cdp = ModCDPClient()

        self.assertEqual(cdp.extension["service_worker_url_suffixes"], ["/modcdp/service_worker.js"])
        self.assertEqual(
            cdp._base_extension_injector_config(None).get("service_worker_url_suffixes"),
            ["/modcdp/service_worker.js"],
        )

    def test_defaults_launched_modcdp_server_upstreams_to_extension_auto(self) -> None:
        for mode in ("nativemessaging", "reversews", "nats"):
            launched = ModCDPClient(launch={"mode": "local"}, upstream={"mode": mode})
            self.assertEqual(launched.launch["mode"], "local")
            self.assertEqual(launched.upstream_endpoint_kind, "modcdp_server")
            self.assertEqual(launched.extension["mode"], "auto")

            attach_only = ModCDPClient(upstream={"mode": mode})
            self.assertEqual(attach_only.launch["mode"], "none")
            self.assertEqual(attach_only.upstream_endpoint_kind, "modcdp_server")
            self.assertEqual(attach_only.extension["mode"], "none")

    def test_rejects_unknown_component_modes_at_their_owning_factory_boundary(self) -> None:
        with self.assertRaisesRegex(RuntimeError, r"unknown upstream\.mode=bogus"):
            ModCDPClient(upstream={"mode": "bogus"})._upstream_transport()
        with self.assertRaisesRegex(RuntimeError, r"unknown launch\.mode=bogus"):
            ModCDPClient(launch={"mode": "bogus"})._browser_launcher()
        with self.assertRaisesRegex(RuntimeError, r"unknown extension\.mode=bogus"):
            ModCDPClient(extension={"mode": "bogus"})._extension_injectors_for_config()

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
            self.assertEqual(cdp.connect_timing.get("extension_source") if cdp.connect_timing else None, "local_launch")
            self.assertEqual(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf")
            self.assertEqual(
                cdp.Mod.evaluate(expression="chrome.runtime.getURL('modcdp/service_worker.js')"),
                "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js",
            )
            sent_at = int(time.time() * 1000)
            pong: Queue[Mapping[str, Any]] = Queue()

            def on_pong(payload: Mapping[str, Any]) -> None:
                if payload.get("sent_at") == sent_at:
                    pong.put(payload)

            cdp.on("Mod.pong", on_pong)
            ping_result = cdp.Mod.ping(sent_at=sent_at)
            pong_payload = pong.get(timeout=10)
            self.assertEqual(ping_result["ok"], True)
            self.assertEqual(pong_payload["sent_at"], sent_at)
            self.assertIsInstance(pong_payload["received_at"], int | float)
            self.assertEqual(pong_payload["from"], "extension-service-worker")
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
        reverse_port = LocalBrowserLauncher.freePort()
        cdp = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "reversews", "reversews_bind": f"127.0.0.1:{reverse_port}"},
            extension={
                "mode": "auto",
                "path": str(EXTENSION_PATH),
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
            server={"routes": {"*.*": "loopback_cdp"}},
        )

        try:
            cdp.connect()
            injector = next(
                candidate
                for candidate in cdp._extension_injectors
                if type(candidate).__name__ == "LocalBrowserLaunchExtensionInjector"
            )
            unpacked_extension_path = getattr(injector, "unpacked_extension_path")
            self.assertIsInstance(unpacked_extension_path, str)
            self.assertNotEqual(unpacked_extension_path, str(EXTENSION_PATH))
            self.assertTrue((Path(unpacked_extension_path) / "config.js").exists())

            launched = cdp._launched_browser
            if launched is None:
                self.fail("expected launched browser")
            original_close = launched["close"]
            browser_close_saw_extension = False

            def close_browser() -> None:
                nonlocal browser_close_saw_extension
                browser_close_saw_extension = Path(unpacked_extension_path).exists()
                original_close()

            launched["close"] = close_browser

            cdp.close()

            self.assertTrue(browser_close_saw_extension)
            self.assertFalse(Path(unpacked_extension_path).exists())
        finally:
            cdp.close()

        self.assertIsNone(cdp.transport)
        self.assertIsNone(cdp._launched_browser)
        self.assertEqual(cdp._extension_injectors, [])

    def test_close_clears_top_level_connection_state(self) -> None:
        cdp = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "ws"},
            extension={
                "mode": "auto",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
        )

        cdp.connect()
        self.assertIsNotNone(cdp.transport)
        cdp.close()

        self.assertIsNone(cdp.transport)
        with self.assertRaisesRegex(RuntimeError, "ModCDP upstream is not connected"):
            cdp.sendRaw("Browser.getVersion")

    def test_generated_cdp_surface_exposes_direct_domain_commands(self) -> None:
        client = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "ws"},
            extension={
                "mode": "auto",
                "path": str(EXTENSION_PATH),
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
            client={"routes": {"Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp"}},
            server={"routes": {"*.*": "loopback_cdp"}},
        )
        target_ids: list[str] = []

        client.connect()
        try:
            result = client.Target.createTarget(url="https://example.com")
            raw_result = client.send("Target.createTarget", {"url": "https://example.org"})
            target_ids.append(str(result.targetId))
            target_ids.append(str(raw_result["targetId"]))

            self.assertRegex(str(result.targetId), r"^[A-F0-9]+$")
            self.assertRegex(str(raw_result["targetId"]), r"^[A-F0-9]+$")
            attached = client.Target.attachToTarget(targetId=result.targetId, flatten=True)
            evaluated = client.Runtime.evaluate(expression="1 + 1", returnByValue=True, session_id=str(attached.sessionId))
            self.assertEqual(evaluated.result["value"], 2)
            self.assertIsNotNone(client.last_command_timing)
            timing = cast(Mapping[str, Any], client.last_command_timing)
            self.assertEqual(timing["target"], "direct_cdp")
            raw_version = cast(Mapping[str, Any], client.send("Browser.getVersion"))
            self.assertIn("product", raw_version)
            self.assertIsNotNone(client.last_command_timing)
            timing = cast(Mapping[str, Any], client.last_command_timing)
            self.assertEqual(timing["target"], "direct_cdp")
        finally:
            for target_id in target_ids:
                try:
                    client.Target.closeTarget(targetId=target_id)
                except Exception:
                    pass
            client.close()

        with self.assertRaises(Exception):
            client.Target._CreateTargetParams.model_validate({"url": "https://example.com", "unknown": True})

        async def run_awaited_calls() -> None:
            awaited_result = await client.Target.createTarget(url="https://example.com")
            target_ids.append(str(awaited_result.targetId))
            self.assertRegex(str(awaited_result.targetId), r"^[A-F0-9]+$")
            awaited_raw_result = await client.send("Target.createTarget", {"url": "https://example.net"})
            target_ids.append(str(awaited_raw_result["targetId"]))
            self.assertRegex(str(awaited_raw_result["targetId"]), r"^[A-F0-9]+$")

        client = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "ws"},
            extension={
                "mode": "auto",
                "path": str(EXTENSION_PATH),
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
            client={"routes": {"Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp"}},
            server={"routes": {"*.*": "loopback_cdp"}},
        )
        target_ids = []
        client.connect()
        try:
            asyncio.run(run_awaited_calls())
        finally:
            for target_id in target_ids:
                try:
                    client.Target.closeTarget(targetId=target_id)
                except Exception:
                    pass
            client.close()

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
