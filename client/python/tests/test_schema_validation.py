from __future__ import annotations

import asyncio
import unittest
from collections.abc import Mapping
from contextlib import redirect_stderr
from io import StringIO
from typing import Any, cast

from modcdp import ModCDPClient
from modcdp.cdp import AwaitableDict
from modcdp.types import JsonValue


class SchemaValidationTests(unittest.TestCase):
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

        client = RecordingClient()

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

    def test_generated_event_surface_supports_typed_and_raw_on(self) -> None:
        client = ModCDPClient()
        typed_events: list[Any] = []
        raw_events: list[dict[str, JsonValue]] = []

        client.on(client.Target.targetCreated, typed_events.append)
        client.on("Target.targetCreated", raw_events.append)

        payload = {"targetInfo": {"targetId": "target-1", "type": "page", "url": "https://example.com"}}
        for handler in client._handlers["Target.targetCreated"]:
            handler(payload)

        self.assertEqual(typed_events[0].targetId, "target-1")
        target_info = cast(dict[str, JsonValue], raw_events[0]["targetInfo"])
        self.assertEqual(target_info["targetId"], "target-1")

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
                "paramsSchema": {
                    "type": "object",
                    "properties": {"text": {"type": "string", "minLength": 1}},
                    "required": ["text"],
                    "additionalProperties": False,
                },
                "resultSchema": {
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
                    "paramsSchema": {
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
                    "eventSchema": {
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
        client = ModCDPClient(custom_events=[{"name": "Custom.count", "eventSchema": {"type": "integer", "minimum": 1}}])

        self.assertEqual(client._validate_event_payload("Custom.count", {"value": 3}), {"value": 3})
        with redirect_stderr(StringIO()):
            self.assertIsNone(client._validate_event_payload("Custom.count", {"value": 0}))


if __name__ == "__main__":
    unittest.main()
