from __future__ import annotations

import asyncio
from queue import Queue
import unittest
from pathlib import Path

from pydantic import BaseModel

from modcdp import ModCDPClient


ROOT = Path(__file__).resolve().parents[3]
EXTENSION_PATH = ROOT / "dist" / "extension"


class ModCDPClientCustomFlatNamespaceTests(unittest.TestCase):
    def test_pydantic_custom_command_installs_flat_dynamic_method_through_real_service_worker(self) -> None:
        class ParamsSchema(BaseModel):
            id: str

        class ResultSchema(BaseModel):
            success: bool

        client = ModCDPClient(
            launch={
                "mode": "local",
                "options": {"headless": True, "sandbox": False},
            },
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
            launch={
                "mode": "local",
                "options": {"headless": True, "sandbox": False},
            },
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


if __name__ == "__main__":
    unittest.main()
