from __future__ import annotations

import asyncio
import importlib
import json
import time
import unittest
from collections.abc import Mapping
from pathlib import Path
from queue import Empty, Queue
from typing import Any, cast

from websocket import create_connection

from modcdp import ModCDPClient
from modcdp.launcher.LocalBrowserLauncher import LocalBrowserLauncher
from modcdp.types import JsonValue
from tests.test_ReverseWebSocketUpstreamTransport import reversews_test_browser_path


HERE = Path(__file__).resolve().parent
EXTENSION_PATH = HERE.parents[1] / "dist" / "extension"


class ModCDPClientTests(unittest.TestCase):
    def test_constructor_normalizes_nested_config_owners(self) -> None:
        cdp = ModCDPClient(
            launcher={
                "launcher_mode": "local",
                "launcher_executable_path": "/tmp/chrome",
                "launcher_user_data_dir": "/tmp/profile",
                "launcher_options": {"headless": True},
            },
            upstream={
                "upstream_mode": "ws",
                "upstream_cdp_url": "http://127.0.0.1:9222",
                "upstream_nats_wait_timeout_ms": 345,
                "upstream_reversews_wait_timeout_ms": 456,
                "upstream_nativemessaging_manifest": "/tmp/native-host.json",
                "upstream_nativemessaging_manifests": ["/tmp/native-host-extra.json"],
                "upstream_nativemessaging_host_name": "com.modcdp.custom",
                "upstream_nativemessaging_wait_timeout_ms": 567,
                "upstream_ws_connect_error_settle_timeout_ms": 321,
            },
            injector={
                "injector_mode": "discover",
                "injector_extension_path": "/tmp/ext",
                "injector_extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "injector_service_worker_url_includes": ["modcdp"],
                "injector_service_worker_url_suffixes": ["/custom/service_worker.js"],
                "injector_trust_service_worker_target": True,
                "injector_require_service_worker_target": True,
                "injector_execution_context_timeout_ms": 4321,
                "injector_service_worker_probe_timeout_ms": 5432,
                "injector_service_worker_ready_timeout_ms": 6543,
                "injector_service_worker_poll_interval_ms": 76,
                "injector_target_session_poll_interval_ms": 87,
            },
            client={
                "client_routes": {"*.*": "direct_cdp"},
                "client_hydrate_aliases": False,
                "client_mirror_upstream_events": False,
                "client_cdp_send_timeout_ms": 1234,
                "client_event_wait_timeout_ms": 2345,
            },
            server={
                "server_routes": {"*.*": "loopback_cdp"},
                "server_browser_token": "token-1",
                "server_cdp_send_timeout_ms": 9876,
                "server_loopback_execution_context_timeout_ms": 8765,
                "server_ws_connect_error_settle_timeout_ms": 7654,
            },
        )

        self.assertEqual(cdp.launcher["launcher_options"], {"headless": True})
        self.assertEqual(cdp._launch_options().get("executable_path"), "/tmp/chrome")
        self.assertEqual(cdp._launch_options().get("user_data_dir"), "/tmp/profile")
        self.assertEqual(cdp.upstream["upstream_nats_wait_timeout_ms"], 345)
        self.assertEqual(cdp.upstream["upstream_reversews_wait_timeout_ms"], 456)
        self.assertEqual(cdp.upstream["upstream_nativemessaging_manifest"], "/tmp/native-host.json")
        self.assertEqual(cdp.upstream["upstream_nativemessaging_manifests"], ["/tmp/native-host-extra.json"])
        self.assertEqual(cdp.upstream["upstream_nativemessaging_host_name"], "com.modcdp.custom")
        self.assertEqual(cdp.upstream["upstream_nativemessaging_wait_timeout_ms"], 567)
        self.assertEqual(cdp.upstream["upstream_ws_connect_error_settle_timeout_ms"], 321)
        self.assertEqual(cdp.injector["injector_execution_context_timeout_ms"], 4321)
        self.assertEqual(cdp.injector["injector_service_worker_probe_timeout_ms"], 5432)
        self.assertEqual(cdp.injector["injector_service_worker_ready_timeout_ms"], 6543)
        self.assertEqual(cdp.injector["injector_service_worker_poll_interval_ms"], 76)
        self.assertEqual(cdp.injector["injector_target_session_poll_interval_ms"], 87)
        self.assertEqual(cdp.client["client_routes"]["*.*"], "direct_cdp")
        self.assertEqual(cdp.client["client_hydrate_aliases"], False)
        self.assertEqual(cdp.client["client_mirror_upstream_events"], False)
        self.assertEqual(cdp.client["client_cdp_send_timeout_ms"], 1234)
        self.assertEqual(cdp.client["client_event_wait_timeout_ms"], 2345)
        self.assertNotIn("Browser", cdp.__dict__)
        with self.assertRaises(AttributeError):
            _ = cdp.Browser
        self.assertNotIn("routes", cdp.__dict__)
        self.assertNotIn("cdp_send_timeout_ms", cdp.__dict__)
        self.assertNotIn("service_worker_probe_timeout_ms", cdp.__dict__)

        params = cast(dict[str, Any], cdp._server_configure_params())
        self.assertEqual(params["client"]["client_routes"]["*.*"], "direct_cdp")
        self.assertEqual(params["server"]["server_browser_token"], "token-1")
        self.assertEqual(params["server"]["server_cdp_send_timeout_ms"], 9876)
        self.assertEqual(params["server"]["server_loopback_execution_context_timeout_ms"], 8765)
        self.assertEqual(params["server"]["server_ws_connect_error_settle_timeout_ms"], 7654)

    def test_preserves_explicit_zero_timeout_config(self) -> None:
        cdp = ModCDPClient(
            upstream={
                "upstream_nats_wait_timeout_ms": 0,
                "upstream_reversews_wait_timeout_ms": 0,
                "upstream_nativemessaging_wait_timeout_ms": 0,
                "upstream_ws_connect_error_settle_timeout_ms": 0,
            },
            injector={
                "injector_execution_context_timeout_ms": 0,
                "injector_service_worker_probe_timeout_ms": 0,
                "injector_service_worker_ready_timeout_ms": 0,
                "injector_service_worker_poll_interval_ms": 0,
                "injector_target_session_poll_interval_ms": 0,
            },
            client={
                "client_cdp_send_timeout_ms": 0,
                "client_event_wait_timeout_ms": 0,
            },
        )

        self.assertEqual(cdp.upstream["upstream_nats_wait_timeout_ms"], 0)
        self.assertEqual(cdp.upstream["upstream_reversews_wait_timeout_ms"], 0)
        self.assertEqual(cdp.upstream["upstream_nativemessaging_wait_timeout_ms"], 0)
        self.assertEqual(cdp.upstream["upstream_ws_connect_error_settle_timeout_ms"], 0)
        self.assertEqual(cdp.injector["injector_execution_context_timeout_ms"], 0)
        self.assertEqual(cdp.injector["injector_service_worker_probe_timeout_ms"], 0)
        self.assertEqual(cdp.injector["injector_service_worker_ready_timeout_ms"], 0)
        self.assertEqual(cdp.injector["injector_service_worker_poll_interval_ms"], 0)
        self.assertEqual(cdp.injector["injector_target_session_poll_interval_ms"], 0)
        self.assertEqual(cdp.client["client_cdp_send_timeout_ms"], 0)
        self.assertEqual(cdp.client["client_event_wait_timeout_ms"], 0)

    def test_preserves_explicit_empty_service_worker_suffix_config(self) -> None:
        cdp = ModCDPClient(injector={"injector_mode": "borrow", "injector_service_worker_url_suffixes": []})

        self.assertEqual(cdp.injector["injector_service_worker_url_suffixes"], [])
        self.assertEqual(cdp._base_extension_injector_config(None).get("injector_service_worker_url_suffixes"), [])

    def test_defaults_service_worker_suffix_config_to_modcdp_worker(self) -> None:
        cdp = ModCDPClient()

        self.assertEqual(cdp.injector["injector_service_worker_url_suffixes"], ["/modcdp/service_worker.js"])
        self.assertEqual(
            cdp._base_extension_injector_config(None).get("injector_service_worker_url_suffixes"),
            ["/modcdp/service_worker.js"],
        )

    def test_preserves_explicit_none_server_config(self) -> None:
        cdp = ModCDPClient(server=None)

        self.assertIsNone(cdp.server)

    def test_only_exposes_injector_attach_after_cdp_send_is_available(self) -> None:
        cdp = ModCDPClient()
        disconnected_config = cdp._base_extension_injector_config(None)
        self.assertIsNone(disconnected_config.get("send"))
        self.assertIsNone(disconnected_config.get("attachToTarget"))

        connected_config = cdp._base_extension_injector_config(lambda method, params=None, session_id=None: {})
        self.assertTrue(callable(connected_config.get("send")))
        self.assertTrue(callable(connected_config.get("attachToTarget")))

    def test_defaults_launched_modcdp_server_upstreams_to_extension_auto(self) -> None:
        for mode in ("nativemessaging", "reversews", "nats"):
            launched = ModCDPClient(launcher={"launcher_mode": "local"}, upstream={"upstream_mode": mode})
            self.assertEqual(launched.launcher["launcher_mode"], "local")
            self.assertEqual(launched.upstream_endpoint_kind, "modcdp_server")
            self.assertEqual(launched.injector["injector_mode"], "auto")

            attach_only = ModCDPClient(upstream={"upstream_mode": mode})
            self.assertEqual(attach_only.launcher["launcher_mode"], "none")
            self.assertEqual(attach_only.upstream_endpoint_kind, "modcdp_server")
            self.assertEqual(attach_only.injector["injector_mode"], "none")

    def test_orders_local_auto_injection_as_launch_flag_then_load_unpacked_fallback(self) -> None:
        cdp = ModCDPClient(
            launcher={"launcher_mode": "local"},
            injector={"injector_mode": "auto"},
        )

        self.assertEqual(
            [type(injector).__name__ for injector in cdp._extension_injectors_for_config()],
            [
                "LocalBrowserLaunchExtensionInjector",
                "ExtensionsLoadUnpackedInjector",
                "DiscoveredExtensionInjector",
                "BorrowedExtensionInjector",
            ],
        )

    def test_rejects_unknown_component_modes_at_their_owning_factory_boundary(self) -> None:
        with self.assertRaisesRegex(RuntimeError, r"unknown upstream\.upstream_mode=bogus"):
            ModCDPClient(upstream={"upstream_mode": "bogus"})._upstream_transport()
        with self.assertRaisesRegex(RuntimeError, r"unknown launcher\.launcher_mode=bogus"):
            ModCDPClient(launcher={"launcher_mode": "bogus"})._browser_launcher()
        with self.assertRaisesRegex(RuntimeError, r"unknown injector\.injector_mode=bogus"):
            ModCDPClient(injector={"injector_mode": "bogus"})._extension_injectors_for_config()

    def test_connects_with_local_launch_injector_chain(self) -> None:
        cdp = ModCDPClient(
            launcher={"launcher_mode": "local", "launcher_options": {"headless": True, "chrome_ready_timeout_ms": 60_000}},
            upstream={"upstream_mode": "ws"},
            injector={
                "injector_mode": "auto",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
                "injector_service_worker_probe_timeout_ms": 30_000,
            },
            client={
                "client_cdp_send_timeout_ms": 30_000,
                "client_event_wait_timeout_ms": 30_000,
            },
        )

        try:
            cdp.connect()
            self.assertIn(
                cdp.connect_timing.get("injector_source") if cdp.connect_timing else None,
                ("discovered", "local_launch", "extensions_load_unpacked", "borrowed"),
            )
            self.assertEqual(cdp.extension_id, "mdedooklbnfejodmnhmkdpkaedafkehf")
            self.assertEqual(
                cdp.Mod.evaluate(expression="chrome.runtime.getURL('modcdp/service_worker.js')"),
                "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js",
            )
            contexts = cdp.Mod.evaluate(
                expression=(
                    "chrome.runtime.getContexts({}).then((contexts) => contexts.map((context) => "
                    "({ type: context.contextType, url: context.documentUrl || context.origin || '' })))"
                )
            )
            self.assertTrue(
                any(
                    isinstance(context, Mapping)
                    and context.get("type") == "OFFSCREEN_DOCUMENT"
                    and context.get("url") == "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/offscreen/keepalive.html"
                    for context in cast(list[Any], contexts)
                )
            )
            sent_at = int(time.time() * 1000)
            pong: Queue[Mapping[str, Any]] = Queue()

            def on_pong(payload: Mapping[str, Any]) -> None:
                if payload.get("sent_at") == sent_at:
                    pong.put(payload)

            muted: Queue[Mapping[str, Any]] = Queue()

            def muted_pong(payload: Mapping[str, Any]) -> None:
                muted.put(payload)

            cdp.on("Mod.pong", muted_pong)
            cdp.off("Mod.pong", muted_pong)
            cdp.on("Mod.pong", on_pong)
            try:
                ping_result = cdp.Mod.ping(sent_at=sent_at)
                pong_payload = pong.get(timeout=30)
            finally:
                cdp.off("Mod.pong", on_pong)
            with self.assertRaises(Empty):
                muted.get(timeout=0.2)
            self.assertEqual(ping_result["ok"], True)
            self.assertEqual(pong_payload["sent_at"], sent_at)
            self.assertIsInstance(pong_payload["received_at"], int | float)
            self.assertEqual(pong_payload["from"], "extension-service-worker")

            cdp.Mod.ping(sent_at=sent_at + 1)
            with self.assertRaises(Empty):
                pong.get(timeout=0.2)
        finally:
            cdp.close()

    def test_close_does_not_close_a_remote_browser_it_did_not_launch(self) -> None:
        chrome = LocalBrowserLauncher(
            {
                "headless": True,
                "chrome_ready_timeout_ms": 60_000,
                # This test manually supplies --load-extension, so it intentionally uses
                # the launch-flag browser path instead of relying on the client fallback.
                "executable_path": reversews_test_browser_path(),
                "extra_args": [f"--load-extension={EXTENSION_PATH}"],
            }
        ).launch()
        raw_ws = create_connection(cast(str, chrome["cdp_url"]), timeout=5)
        cdp = ModCDPClient(
            launcher={"launcher_mode": "remote"},
            upstream={"upstream_mode": "ws", "upstream_cdp_url": chrome["cdp_url"]},
            injector={
                "injector_mode": "auto",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
                "injector_service_worker_ready_timeout_ms": 30_000,
                "injector_service_worker_probe_timeout_ms": 30_000,
            },
            client={"client_routes": {"*.*": "direct_cdp"}},
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
        cdp = ModCDPClient(
            launcher={
                "launcher_mode": "local",
                "launcher_options": {
                    "headless": True,
                    # After explicit CHROME_PATH and CI /usr/bin/chromium, this test uses
                    # Chrome for Testing because Canary rejects --load-extension in this
                    # local launch injector path.
                    "executable_path": reversews_test_browser_path(),
                },
            },
            upstream={"upstream_mode": "ws"},
            injector={
                "injector_mode": "auto",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
            server={"server_routes": {"*.*": "loopback_cdp"}},
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
            launcher={"launcher_mode": "local", "launcher_options": {"headless": True}},
            upstream={"upstream_mode": "ws"},
            injector={
                "injector_mode": "auto",
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
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
            launcher={"launcher_mode": "local", "launcher_options": {"headless": True}},
            upstream={"upstream_mode": "ws"},
            injector={
                "injector_mode": "auto",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
            client={"client_routes": {"Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp"}},
            server={"server_routes": {"*.*": "loopback_cdp"}},
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
            launcher={"launcher_mode": "local", "launcher_options": {"headless": True}},
            upstream={"upstream_mode": "ws"},
            injector={
                "injector_mode": "auto",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
            client={"client_routes": {"Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp"}},
            server={"server_routes": {"*.*": "loopback_cdp"}},
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

    def test_event_dispatch_snapshots_handlers_when_once_removes_itself(self) -> None:
        client = ModCDPClient()
        client.ext_session_id = "ext-session"
        modcdp_client_module = importlib.import_module("modcdp.client.ModCDPClient")
        original_thread = modcdp_client_module.threading.Thread
        seen: list[str] = []

        class ImmediateThread:
            def __init__(self, target, daemon=False):  # noqa: ANN001
                self.target = target

            def start(self) -> None:
                self.target()

        def persistent(_payload: Mapping[str, Any]) -> None:
            seen.append("persistent")

        try:
            cast(Any, modcdp_client_module.threading).Thread = ImmediateThread
            client.once("Target.targetCreated", lambda _payload: seen.append("once"))
            client.on("Target.targetCreated", persistent)
            client._on_recv(
                {
                    "method": "Target.targetCreated",
                    "params": {"targetInfo": {"targetId": "target-1", "type": "page", "url": "about:blank"}},
                }
            )
            self.assertEqual(seen, ["once", "persistent"])

            seen.clear()
            client._on_recv(
                {
                    "method": "Target.targetCreated",
                    "params": {"targetInfo": {"targetId": "target-2", "type": "page", "url": "about:blank"}},
                }
            )
            self.assertEqual(seen, ["persistent"])
        finally:
            cast(Any, modcdp_client_module.threading).Thread = original_thread

    def test_root_events_dispatch_before_extension_session_is_attached(self) -> None:
        client = ModCDPClient()
        modcdp_client_module = importlib.import_module("modcdp.client.ModCDPClient")
        original_thread = modcdp_client_module.threading.Thread
        seen: list[str] = []

        class ImmediateThread:
            def __init__(self, target, daemon=False):  # noqa: ANN001
                self.target = target

            def start(self) -> None:
                self.target()

        try:
            cast(Any, modcdp_client_module.threading).Thread = ImmediateThread
            client.on("Target.targetCreated", lambda payload: seen.append(str(payload["targetInfo"]["targetId"])))
            client._on_recv(
                {
                    "method": "Target.targetCreated",
                    "params": {"targetInfo": {"targetId": "target-1", "type": "page", "url": "about:blank"}},
                }
            )
            self.assertEqual(seen, ["target-1"])
        finally:
            cast(Any, modcdp_client_module.threading).Thread = original_thread

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
        with self.assertRaises(ValueError):
            client._validate_event_payload("Custom.ready", {"url": "http://example.com", "ready": True})
        with self.assertRaises(ValueError):
            client._validate_event_payload("Custom.ready", {"url": "https://example.com", "ready": True, "x": 1})

    def test_scalar_event_schema_validates_value_payloads(self) -> None:
        client = ModCDPClient(custom_events=[{"name": "Custom.count", "event_schema": {"type": "integer", "minimum": 1}}])

        self.assertEqual(client._validate_event_payload("Custom.count", {"value": 3}), {"value": 3})
        with self.assertRaises(ValueError):
            client._validate_event_payload("Custom.count", {"value": 0})


if __name__ == "__main__":
    unittest.main()
