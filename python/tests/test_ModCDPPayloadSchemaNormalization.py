from __future__ import annotations

import unittest

from pydantic_core import to_jsonable_python

from modcdp import ModCDPClient


class ModCDPPayloadSchemaNormalizationTests(unittest.TestCase):
    def test_payload_schema_normalization_accepts_empty_json_schema_objects(self) -> None:
        adapter, schema, _ = ModCDPClient()._adapter_from_optional_schema({}, "params_schema")

        assert adapter is not None
        self.assertEqual(schema, {})
        self.assertEqual(adapter.validate_python({"value": 1}), {"value": 1})

    def test_payload_schema_normalization_rejects_unsupported_schema_specs(self) -> None:
        with self.assertRaises(TypeError):
            ModCDPClient()._adapter_from_optional_schema("not-a-schema", "params_schema")

    def test_payload_schema_normalization_accepts_non_empty_json_schema_objects(self) -> None:
        adapter, schema, _ = ModCDPClient()._adapter_from_optional_schema(
            {
                "type": "object",
                "properties": {"value": {"type": "string"}},
                "required": ["value"],
            },
            "params_schema",
        )

        assert adapter is not None
        assert schema is not None
        self.assertEqual(schema["type"], "object")
        self.assertEqual(to_jsonable_python(adapter.validate_python({"value": "ok", "extra": True})), {"value": "ok", "extra": True})


if __name__ == "__main__":
    unittest.main()
