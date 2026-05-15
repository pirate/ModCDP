from __future__ import annotations

import asyncio
from queue import Queue
import unittest
from pathlib import Path

from pydantic import BaseModel

from modcdp import ModCDPClient


ROOT = Path(__file__).resolve().parents[2]
EXTENSION_PATH = ROOT / "dist" / "extension"


class ModCDPClientCustomFlatNamespaceTests(unittest.TestCase):
    def test_pydantic_custom_command_installs_flat_dynamic_method_through_real_service_worker(self) -> None:
        class ParamsSchema(BaseModel):
            id: str

        class ResultSchema(BaseModel):
            success: bool

        client = ModCDPClient(
            launcher={
                "launcher_mode": "local",
                "launcher_options": {"headless": True},
            },
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

        async def run() -> None:
            client.connect()
            registered = await client.Mod.addCustomCommand(
                "Custom.doSomething",
                params_schema=ParamsSchema,
                result_schema=ResultSchema,
                expression="async ({ id }) => ({ success: id === 'abc' })",
            )
            self.assertEqual(registered, {"name": "Custom.doSomething", "registered": True})
            success: bool = await client.Custom.doSomething(id="abc")
            raw_success: bool = bool(await client.send("Custom.doSomething", {"id": "abc"}))
            self.assertIs(success, True)
            self.assertIs(raw_success, True)

        try:
            asyncio.run(run())
            with self.assertRaises(ValueError):
                client.Custom.doSomething(id=123)
        finally:
            client.close()

    def test_pydantic_custom_event_schema_coerces_raw_string_handlers_through_real_service_worker(self) -> None:
        class EventSchema(BaseModel):
            data: str

        client = ModCDPClient(
            launcher={
                "launcher_mode": "local",
                "launcher_options": {"headless": True},
            },
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
        seen: Queue[str] = Queue()

        async def callback(event: EventSchema) -> None:
            seen.put(event.data)

        async def run() -> None:
            client.connect()
            await client.Mod.addCustomEvent("Custom.someEvent", event_schema=EventSchema)
            await client.on("Custom.someEvent", callback)
            await client.Mod.evaluate(
                expression="async () => await globalThis.ModCDP.emit('Custom.someEvent', { data: 'ok' })"
            )

        try:
            asyncio.run(run())
            self.assertEqual(seen.get(timeout=10), "ok")
        finally:
            client.close()

    def test_schema_only_custom_event_registers_without_websocket(self) -> None:
        client = ModCDPClient()

        result = client.send(
            "Mod.addCustomEvent",
            {
                "name": "Custom.schemaOnly",
                "event_schema": {
                    "type": "object",
                    "properties": {"ok": {"type": "boolean"}},
                    "required": ["ok"],
                    "additionalProperties": False,
                },
            },
        )

        self.assertEqual(result, {"name": "Custom.schemaOnly", "registered": True})
        self.assertEqual(client._validate_event_payload("Custom.schemaOnly", {"ok": True}), {"ok": True})
        with self.assertRaises(ValueError):
            client._validate_event_payload("Custom.schemaOnly", {"ok": True, "extra": True})


if __name__ == "__main__":
    unittest.main()
