package modcdp

import (
	"strings"
	"testing"
)

func TestCustomCommandSchemasValidateParamsAndResults(t *testing.T) {
	cdp := New(Options{
		CustomCommands: []CustomCommand{
			{
				Name: "Custom.echo",
				ParamsSchema: map[string]any{
					"type":                 "object",
					"required":             []any{"value"},
					"properties":           map[string]any{"value": map[string]any{"type": "string"}},
					"additionalProperties": false,
				},
				ResultSchema: map[string]any{
					"type":                 "object",
					"required":             []any{"value"},
					"properties":           map[string]any{"value": map[string]any{"type": "string"}},
					"additionalProperties": false,
				},
			},
		},
	})

	if err := cdp.validateCommandParams("Custom.echo", map[string]any{"value": "ok"}); err != nil {
		t.Fatalf("expected valid params, got %v", err)
	}
	if err := cdp.validateCommandParams("Custom.echo", map[string]any{"value": 42}); err == nil || !strings.Contains(err.Error(), "paramsSchema") {
		t.Fatalf("expected params schema error, got %v", err)
	}
	if err := cdp.validateCommandResult("Custom.echo", map[string]any{"value": "ok"}); err != nil {
		t.Fatalf("expected valid result, got %v", err)
	}
	if err := cdp.validateCommandResult("Custom.echo", map[string]any{"value": 42}); err == nil || !strings.Contains(err.Error(), "resultSchema") {
		t.Fatalf("expected result schema error, got %v", err)
	}
}

func TestSchemaOnlyAddCustomCommandRegistersWithoutConnection(t *testing.T) {
	cdp := New(Options{})
	result, err := cdp.Send("Mod.addCustomCommand", map[string]any{
		"name": "Custom.clientOnly",
		"paramsSchema": map[string]any{
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

func TestCustomEventSchemasValidatePayloads(t *testing.T) {
	cdp := New(Options{
		CustomEvents: []CustomEvent{
			{
				Name: "Custom.changed",
				EventSchema: map[string]any{
					"type":                 "object",
					"required":             []any{"targetId"},
					"properties":           map[string]any{"targetId": map[string]any{"type": "string"}},
					"additionalProperties": false,
				},
			},
		},
	})

	if _, ok := cdp.validateEventData("Custom.changed", map[string]any{"targetId": "target-1"}); !ok {
		t.Fatal("expected valid event payload")
	}
	if _, ok := cdp.validateEventData("Custom.changed", map[string]any{"targetId": 1}); ok {
		t.Fatal("expected invalid event payload")
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
	command, err := AddCustomCommand[ParamsSchema, ResultSchema](cdp, "Custom.doSomething")
	if err != nil {
		t.Fatal(err)
	}
	if command.name != "Custom.doSomething" {
		t.Fatalf("unexpected command name %q", command.name)
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
	if err := AddCustomEvent[EventSchema](cdp, "Custom.someEvent"); err != nil {
		t.Fatal(err)
	}
	seen := make(chan string, 1)
	OnTyped[EventSchema](cdp, "Custom.someEvent", func(event EventSchema) {
		seen <- event.Data
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

func TestTypedCDPSurfaceInitializesAndEncodesParams(t *testing.T) {
	cdp := New(Options{})
	if cdp.Target.client != cdp {
		t.Fatal("expected Target domain to be initialized with the client")
	}

	params := TargetCreateTargetParams{
		URL:        "https://example.com",
		Background: Bool(true),
	}
	raw, err := cdpParamsMap(params)
	if err != nil {
		t.Fatal(err)
	}
	if raw["url"] != "https://example.com" || raw["background"] != true {
		t.Fatalf("unexpected encoded Target.createTarget params: %#v", raw)
	}
	if _, ok := raw["sessionId"]; ok {
		t.Fatalf("SessionID must stay transport-only, got %#v", raw)
	}
}

func TestTypedCDPEventsWrapRawHandlers(t *testing.T) {
	cdp := New(Options{})
	typedEvents := make(chan TargetTargetCreatedEvent, 1)
	rawEvents := make(chan any, 1)

	cdp.Target.On.TargetCreated(func(event TargetTargetCreatedEvent) {
		typedEvents <- event
	})
	cdp.On("Target.targetCreated", func(event any) {
		rawEvents <- event
	})

	payload := map[string]any{
		"targetInfo": map[string]any{
			"targetId": "target-1",
			"type":     "page",
			"url":      "https://example.com",
		},
	}
	for _, handler := range cdp.handlers["Target.targetCreated"] {
		handler(payload)
	}

	typed := <-typedEvents
	if typed.TargetID() != "target-1" || typed.TargetInfo.URL != "https://example.com" {
		t.Fatalf("unexpected typed event: %#v", typed)
	}
	raw := <-rawEvents
	rawMap, ok := raw.(map[string]any)
	if !ok || rawMap["targetInfo"] == nil {
		t.Fatalf("unexpected raw event: %#v", raw)
	}
}
