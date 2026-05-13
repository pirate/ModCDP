"""Pydantic v2 runtime adapters for JSON Schema.

This intentionally follows abxbus' dynamic JSON Schema loading pattern: raw
JSON Schema dictionaries are reconstructed as Pydantic-compatible runtime types,
then validated with ``TypeAdapter`` at the client boundary.
"""

from __future__ import annotations

from collections.abc import Mapping, Sequence
from typing import Annotated, Any, Literal, TypeAlias, Union, cast

from pydantic import BaseModel, ConfigDict, Field, TypeAdapter, create_model

JsonSchema: TypeAlias = Mapping[str, Any]
FieldDefinition: TypeAlias = Any | tuple[Any, Any]

_TYPE_MAPPING: dict[str, Any] = {
    "string": str,
    "integer": int,
    "number": float,
    "boolean": bool,
    "object": dict,
    "array": list,
    "null": type(None),
}

_CONSTRAINT_MAPPING: dict[str, str] = {
    "minimum": "ge",
    "maximum": "le",
    "exclusiveMinimum": "gt",
    "exclusiveMaximum": "lt",
    "minItems": "min_length",
    "maxItems": "max_length",
    "minLength": "min_length",
    "maxLength": "max_length",
    "pattern": "pattern",
    "multipleOf": "multiple_of",
}


def _as_string_key_dict(value: object) -> dict[str, Any] | None:
    if not isinstance(value, Mapping):
        return None
    return {key: raw_value for key, raw_value in cast(Mapping[object, Any], value).items() if isinstance(key, str)}


def _as_sequence(value: object) -> Sequence[Any] | None:
    if isinstance(value, Sequence) and not isinstance(value, (str, bytes, bytearray)):
        return cast(Sequence[Any], value)
    return None


def _iter_schema_objects(value: object) -> list[dict[str, Any]]:
    sequence = _as_sequence(value)
    if sequence is None:
        return []
    schemas: list[dict[str, Any]] = []
    for item in sequence:
        schema = _as_string_key_dict(item)
        if schema is not None:
            schemas.append(schema)
    return schemas


def _combine_union(types: list[Any]) -> Any:
    if not types:
        return Any
    return cast(Any, Union).__getitem__(tuple(types))


def _literal_type(values: Sequence[Any]) -> Any:
    return cast(Any, Literal).__getitem__(tuple(values))


def _create_dynamic_model(
    model_name: str,
    model_schema: Mapping[str, Any],
    fields: Mapping[str, FieldDefinition] | None = None,
) -> type[BaseModel]:
    field_definitions = cast(Any, dict(fields or {}))
    return create_model(
        model_name,
        __config__=ConfigDict(extra="forbid" if model_schema.get("additionalProperties") is False else "allow"),
        __doc__=str(model_schema.get("description", "")),
        **field_definitions,
    )


def _schema_type(schema: Mapping[str, Any]) -> str | None:
    raw_type = schema.get("type")
    if isinstance(raw_type, str):
        return raw_type
    raw_types = _as_sequence(raw_type)
    if raw_types is None:
        return None
    non_null = [item for item in raw_types if isinstance(item, str) and item != "null"]
    return non_null[0] if len(non_null) == 1 else None


def _allows_null(schema: Mapping[str, Any]) -> bool:
    raw_type = schema.get("type")
    if raw_type == "null":
        return True
    raw_types = _as_sequence(raw_type)
    if raw_types is not None and "null" in raw_types:
        return True
    return any(candidate.get("type") == "null" for candidate in _iter_schema_objects(schema.get("anyOf")))


def _nullable(resolved_type: Any, *, nullable: bool) -> Any:
    if not nullable or resolved_type is type(None):
        return resolved_type
    return resolved_type | None


def _field_params(schema: Mapping[str, Any]) -> dict[str, Any]:
    params: dict[str, Any] = {}
    for schema_key, field_key in _CONSTRAINT_MAPPING.items():
        if schema_key in schema:
            params[field_key] = schema[schema_key]
    if "description" in schema:
        params["description"] = schema["description"]
    if "default" in schema:
        params["default"] = schema["default"]
    return params


def _annotated(resolved_type: Any, schema: Mapping[str, Any]) -> Any:
    params = _field_params(schema)
    if not params:
        return resolved_type
    return Annotated[resolved_type, Field(**params)]


def pydantic_type_from_json_schema(schema_raw: JsonSchema | dict[str, Any]) -> Any:
    """Reconstruct a Pydantic-compatible runtime type from raw JSON Schema."""

    schema = _as_string_key_dict(schema_raw) or {}
    definitions = _as_string_key_dict(schema.get("$defs")) or _as_string_key_dict(schema.get("definitions")) or {}
    models: dict[str, type[BaseModel]] = {}
    build_stack: set[str] = set()

    def resolve_ref(reference: str) -> Any:
        name = reference.split("/")[-1]
        if name in models:
            return models[name]
        if name in build_stack:
            return Any
        ref_schema = _as_string_key_dict(definitions.get(name))
        if ref_schema is None:
            return Any

        build_stack.add(name)
        try:
            model = _create_dynamic_model(name or "ReferencedObject", ref_schema, build_model_fields(ref_schema))
            models[name] = model
            return model
        finally:
            build_stack.remove(name)

    def build_model_fields(object_schema: Mapping[str, Any]) -> dict[str, FieldDefinition]:
        fields: dict[str, FieldDefinition] = {}
        properties = _as_string_key_dict(object_schema.get("properties")) or {}
        required_raw = _as_sequence(object_schema.get("required")) or []
        required = {item for item in required_raw if isinstance(item, str)}
        for field_name, field_schema_raw in properties.items():
            field_schema = _as_string_key_dict(field_schema_raw) or {}
            field_type = resolve_schema(field_schema)
            field_params = _field_params(field_schema)
            if field_name in required:
                fields[field_name] = (field_type, Field(**field_params))
            else:
                default = field_params.pop("default", None)
                fields[field_name] = (_nullable(field_type, nullable=True), Field(default=default, **field_params))
        return fields

    def resolve_array(array_schema: Mapping[str, Any], *, nullable: bool) -> Any:
        prefix_items = _iter_schema_objects(array_schema.get("prefixItems"))
        if prefix_items:
            tuple_type = tuple.__class_getitem__(tuple(resolve_schema(item) for item in prefix_items))
            return _nullable(_annotated(tuple_type, array_schema), nullable=nullable)

        items_schema = _as_string_key_dict(array_schema.get("items"))
        item_type = resolve_schema(items_schema) if items_schema is not None else Any
        array_type = set[item_type] if array_schema.get("uniqueItems") is True else list[item_type]
        return _nullable(_annotated(array_type, array_schema), nullable=nullable)

    def resolve_object(object_schema: Mapping[str, Any], *, nullable: bool) -> Any:
        properties = _as_string_key_dict(object_schema.get("properties"))
        if properties is not None:
            model_name = str(object_schema.get("title") or "InlineObject")
            model = _create_dynamic_model(model_name, object_schema, build_model_fields(object_schema))
            return _nullable(model, nullable=nullable)

        additional_properties = object_schema.get("additionalProperties")
        additional_schema = _as_string_key_dict(additional_properties)
        if additional_schema is not None:
            additional_type = resolve_schema(additional_schema)
            return _nullable(dict[str, additional_type], nullable=nullable)
        if additional_properties is False:
            model = _create_dynamic_model(str(object_schema.get("title") or "InlineObject"), object_schema)
            return _nullable(model, nullable=nullable)
        return _nullable(dict[str, Any], nullable=nullable)

    def resolve_schema(candidate_raw: object) -> Any:
        candidate = _as_string_key_dict(candidate_raw) or {}
        if not candidate:
            return Any

        nullable = _allows_null(candidate)
        if "$ref" in candidate and isinstance(candidate["$ref"], str):
            return _nullable(resolve_ref(candidate["$ref"]), nullable=nullable)
        if "const" in candidate:
            return _nullable(_literal_type([candidate["const"]]), nullable=nullable)
        enum_values = _as_sequence(candidate.get("enum"))
        if enum_values is not None:
            non_null_values = [item for item in enum_values if item is not None]
            return _nullable(_literal_type(non_null_values), nullable=nullable or len(non_null_values) != len(enum_values))

        any_of = _iter_schema_objects(candidate.get("anyOf"))
        if any_of:
            resolved = [resolve_schema(item) for item in any_of if item.get("type") != "null"]
            return _nullable(_combine_union(resolved), nullable=nullable)
        one_of = _iter_schema_objects(candidate.get("oneOf"))
        if one_of:
            resolved = [resolve_schema(item) for item in one_of if item.get("type") != "null"]
            return _nullable(_combine_union(resolved), nullable=nullable)

        raw_types = _as_sequence(candidate.get("type"))
        if raw_types is not None:
            resolved = [
                resolve_schema({**candidate, "type": item})
                for item in raw_types
                if isinstance(item, str) and item != "null"
            ]
            return _nullable(_combine_union(resolved), nullable=nullable)

        candidate_type = _schema_type(candidate)
        if candidate_type == "array":
            return resolve_array(candidate, nullable=nullable)
        if candidate_type == "object":
            return resolve_object(candidate, nullable=nullable)
        if candidate_type == "null":
            return type(None)
        primitive_type = _TYPE_MAPPING.get(candidate_type or "")
        if primitive_type is not None:
            return _nullable(_annotated(primitive_type, candidate), nullable=nullable)
        return _nullable(Any, nullable=nullable)

    for definition_name in definitions:
        resolve_ref(definition_name)
    return resolve_schema(schema)


def type_adapter_from_json_schema(schema: JsonSchema | dict[str, Any]) -> TypeAdapter[Any]:
    """Return a Pydantic v2 adapter for a raw JSON Schema object."""

    return TypeAdapter(pydantic_type_from_json_schema(schema))


__all__ = ["pydantic_type_from_json_schema", "type_adapter_from_json_schema"]
