from __future__ import annotations

import asyncio
import unittest
from contextlib import redirect_stderr
from io import StringIO
from typing import Any

from pydantic import BaseModel

from modcdp import ModCDPClient
from modcdp.types import JsonValue


class ModCDPClientCustomFlatNamespaceTests(unittest.TestCase):
    def test_pydantic_custom_command_installs_flat_dynamic_method(self) -> None:
        class ParamsSchema(BaseModel):
            id: str

        class ResultSchema(BaseModel):
            success: bool

        class RecordingClient(ModCDPClient):
            def _send_raw(self, wrapped: Any) -> JsonValue:
                self.last_wrapped = wrapped
                return {"success": True}

        client = RecordingClient(client={"routes": {"Custom.*": "direct_cdp"}})

        async def run() -> None:
            registered = await client.Mod.addCustomCommand(
                "Custom.doSomething",
                params_schema=ParamsSchema,
                result_schema=ResultSchema,
            )
            self.assertEqual(registered, {"name": "Custom.doSomething", "registered": True})
            success: bool = await client.Custom.doSomething(id="abc")
            raw_success: bool = bool(await client.send("Custom.doSomething", {"id": "abc"}))
            self.assertIs(success, True)
            self.assertIs(raw_success, True)

        asyncio.run(run())
        with self.assertRaises(ValueError):
            client.Custom.doSomething(id=123)

    def test_pydantic_custom_event_schema_coerces_raw_string_handlers(self) -> None:
        class EventSchema(BaseModel):
            data: str

        client = ModCDPClient()
        seen: list[str] = []

        async def callback(event: EventSchema) -> None:
            seen.append(event.data)

        async def register() -> None:
            await client.Mod.addCustomEvent("Custom.someEvent", event_schema=EventSchema)
            await client.on("Custom.someEvent", callback)

        asyncio.run(register())
        client._run_handler(
            client._handlers["Custom.someEvent"][0],
            client._validate_event_payload("Custom.someEvent", {"data": "ok"}),
            "Custom.someEvent",
        )
        self.assertEqual(seen, ["ok"])
        with redirect_stderr(StringIO()):
            self.assertIsNone(client._validate_event_payload("Custom.someEvent", {"data": 123}))


if __name__ == "__main__":
    unittest.main()
