package modcdp

import (
	"strings"
	"testing"

	abxjsonschema "github.com/ArchiveBox/abxbus/abxbus-go/jsonschema"
)

func TestPayloadSchemaNormalizationAcceptsEmptyJSONSchemaObjects(t *testing.T) {
	schema := cloneSchema(map[string]any{})
	if schema == nil {
		t.Fatal("expected empty schema object to normalize")
	}
	if err := abxjsonschema.Validate(schema, map[string]any{"value": 1}); err != nil {
		t.Fatalf("expected empty schema to accept payload: %v", err)
	}
}

func TestPayloadSchemaNormalizationRejectsUnsupportedSchemaSpecs(t *testing.T) {
	_, err := New(Options{}).Send("Mod.addCustomCommand", map[string]any{
		"name":          "Custom.bad",
		"params_schema": "not-a-schema",
	})
	if err == nil || !strings.Contains(err.Error(), "params_schema must be a JSON Schema object") {
		t.Fatalf("expected unsupported schema error, got %v", err)
	}
}

func TestPayloadSchemaNormalizationAcceptsNonEmptyJSONSchemaObjects(t *testing.T) {
	schema := cloneSchema(map[string]any{
		"type":       "object",
		"required":   []any{"value"},
		"properties": map[string]any{"value": map[string]any{"type": "string"}},
	})
	if schema == nil {
		t.Fatal("expected schema object to normalize")
	}
	if err := abxjsonschema.Validate(schema, map[string]any{"value": "ok", "extra": true}); err != nil {
		t.Fatalf("expected valid payload: %v", err)
	}
	if err := abxjsonschema.Validate(schema, map[string]any{"value": 1}); err == nil {
		t.Fatal("expected invalid payload to fail validation")
	}
}
