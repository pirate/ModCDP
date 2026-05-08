from __future__ import annotations

import unittest
from contextlib import redirect_stderr
from io import StringIO

from modcdp import ModCDPClient
from modcdp.types import JsonValue


class SchemaValidationTests(unittest.TestCase):
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
