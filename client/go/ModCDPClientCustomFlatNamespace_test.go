package modcdp

import (
	"testing"

	abxjsonschema "github.com/ArchiveBox/abxbus/abxbus-go/jsonschema"
)

func TestSchemaOnlyAddCustomCommandRegistersWithoutConnection(t *testing.T) {
	cdp := New(Options{})
	result, err := cdp.Send("Mod.addCustomCommand", map[string]any{
		"name": "Custom.clientOnly",
		"params_schema": map[string]any{
			"type":                 "object",
			"required":             []any{"tabId"},
			"properties":           map[string]any{"tabId": map[string]any{"type": "integer"}},
			"additionalProperties": false,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	registration, ok := result.(map[string]any)
	if !ok || registration["name"] != "Custom.clientOnly" || registration["registered"] != true {
		t.Fatalf("unexpected schema-only registration result: %#v", result)
	}
	if err := cdp.validateCommandParams("Custom.clientOnly", map[string]any{"tabId": 1}); err != nil {
		t.Fatalf("expected registered schema to validate params, got %v", err)
	}
	if err := cdp.validateCommandParams("Custom.clientOnly", map[string]any{"tabId": "1"}); err == nil {
		t.Fatal("expected registered schema to reject wrong params")
	}
}

func TestTypedCustomCommandRegistrationBuildsSchemas(t *testing.T) {
	type ParamsSchema struct {
		ID string `json:"id"`
	}
	type ResultSchema struct {
		Success bool `json:"success"`
	}

	cdp := New(Options{})
	result, err := cdp.Send("Mod.addCustomCommand", map[string]any{
		"name":          "Custom.doSomething",
		"params_schema": abxjsonschema.SchemaFor[ParamsSchema](),
		"result_schema": abxjsonschema.SchemaFor[ResultSchema](),
	})
	if err != nil {
		t.Fatal(err)
	}
	registration, ok := result.(map[string]any)
	if !ok || registration["name"] != "Custom.doSomething" || registration["registered"] != true {
		t.Fatalf("unexpected custom command registration: %#v", result)
	}
	params, err := cdpParamsMap(ParamsSchema{ID: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	if err := cdp.validateCommandParams("Custom.doSomething", params); err != nil {
		t.Fatalf("expected typed params schema to validate: %v", err)
	}
	if err := cdp.validateCommandParams("Custom.doSomething", map[string]any{"id": 123}); err == nil {
		t.Fatal("expected typed params schema to reject wrong id type")
	}
	if err := cdp.validateCommandResult("Custom.doSomething", ResultSchema{Success: true}); err != nil {
		t.Fatalf("expected typed result schema to validate: %v", err)
	}
	if err := cdp.validateCommandResult("Custom.doSomething", map[string]any{"success": "yes"}); err == nil {
		t.Fatal("expected typed result schema to reject wrong success type")
	}
}

func TestTypedCustomEventRegistrationAndHandler(t *testing.T) {
	type EventSchema struct {
		Data string `json:"data"`
	}

	cdp := New(Options{})
	result, err := cdp.Send("Mod.addCustomEvent", map[string]any{
		"name":         "Custom.someEvent",
		"event_schema": abxjsonschema.SchemaFor[EventSchema](),
	})
	if err != nil {
		t.Fatal(err)
	}
	registration, ok := result.(map[string]any)
	if !ok || registration["name"] != "Custom.someEvent" || registration["registered"] != true {
		t.Fatalf("unexpected custom event registration: %#v", result)
	}
	seen := make(chan string, 1)
	cdp.On("Custom.someEvent", func(data any) {
		event := data.(map[string]any)
		seen <- event["data"].(string)
	})
	if data, ok := cdp.validateEventData("Custom.someEvent", map[string]any{"data": "ok"}); ok {
		for _, handler := range cdp.handlers["Custom.someEvent"] {
			handler(data)
		}
	} else {
		t.Fatal("expected valid typed event payload")
	}
	if got := <-seen; got != "ok" {
		t.Fatalf("unexpected typed event data %q", got)
	}
	if _, ok := cdp.validateEventData("Custom.someEvent", map[string]any{"data": 123}); ok {
		t.Fatal("expected typed event schema to reject wrong data type")
	}
}
