package client

import (
	"path/filepath"
	"testing"
	"time"

	abxjsonschema "github.com/ArchiveBox/abxbus/abxbus-go/jsonschema"
)

func TestCustomCommandsInstallFlatNamespaceThroughRealServiceWorker(t *testing.T) {
	type ParamsSchema struct {
		ID string `json:"id"`
	}
	type ResultSchema struct {
		Success bool `json:"success"`
	}

	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	cdp := New(Options{
		Launcher: LauncherConfig{LauncherMode: "local",
			LauncherOptions: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{UpstreamMode: "ws"},
		Injector: InjectorConfig{
			InjectorMode:                     "auto",
			InjectorExtensionPath:                     extensionPath,
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
		Client: ClientConfig{ClientRoutes: map[string]string{
			"Mod.*":    "service_worker",
			"Custom.*": "service_worker",
			"*.*":      "direct_cdp",
		}},
		Server: &ServerConfig{ServerRoutes: map[string]string{"*.*": "loopback_cdp"}},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	registered, err := cdp.Mod.AddCustomCommand(CustomCommand{
		Name:         "Custom.doSomething",
		ParamsSchema: abxjsonschema.SchemaFor[ParamsSchema](),
		ResultSchema: abxjsonschema.SchemaFor[ResultSchema](),
		Expression:   "async ({ id }) => ({ success: id === 'abc' })",
	})
	if err != nil {
		t.Fatal(err)
	}
	registration, ok := registered.(map[string]any)
	if !ok || registration["name"] != "Custom.doSomething" || registration["registered"] != true {
		t.Fatalf("unexpected custom command registration: %#v", registered)
	}
	result, err := cdp.Send("Custom.doSomething", map[string]any{"id": "abc"})
	if err != nil {
		t.Fatal(err)
	}
	if result != true {
		t.Fatalf("Custom.doSomething = %#v", result)
	}
	if _, err := cdp.Send("Custom.doSomething", map[string]any{"id": 123}); err == nil {
		t.Fatal("expected custom command params schema to reject non-string id")
	}
}

func TestCustomEventsValidateRawStringHandlersThroughRealServiceWorker(t *testing.T) {
	type EventSchema struct {
		Data string `json:"data"`
	}

	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	cdp := New(Options{
		Launcher: LauncherConfig{LauncherMode: "local",
			LauncherOptions: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{UpstreamMode: "ws"},
		Injector: InjectorConfig{
			InjectorMode:                     "auto",
			InjectorExtensionPath:                     extensionPath,
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
		Client: ClientConfig{ClientRoutes: map[string]string{
			"Mod.*":    "service_worker",
			"Custom.*": "service_worker",
			"*.*":      "direct_cdp",
		}},
		Server: &ServerConfig{ServerRoutes: map[string]string{"*.*": "loopback_cdp"}},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	registered, err := cdp.Mod.AddCustomEvent(CustomEvent{
		Name:        "Custom.someEvent",
		EventSchema: abxjsonschema.SchemaFor[EventSchema](),
	})
	if err != nil {
		t.Fatal(err)
	}
	registration, ok := registered.(map[string]any)
	if !ok || registration["name"] != "Custom.someEvent" || registration["registered"] != true {
		t.Fatalf("unexpected custom event registration: %#v", registered)
	}
	seen := make(chan string, 1)
	cdp.On("Custom.someEvent", func(data any) {
		event, _ := data.(map[string]any)
		if event != nil {
			seen <- event["data"].(string)
		}
	})
	if _, err := cdp.Mod.Evaluate(map[string]any{
		"expression": "async () => await globalThis.ModCDP.emit('Custom.someEvent', { data: 'ok' })",
	}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-seen:
		if got != "ok" {
			t.Fatalf("Custom.someEvent data = %q", got)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for Custom.someEvent")
	}
}

func TestSchemaOnlyAddCustomCommandRegistersWithoutConnection(t *testing.T) {
	cdp := New(Options{})
	result, err := cdp.Mod.AddCustomCommand(CustomCommand{
		Name: "Custom.clientOnly",
		ParamsSchema: map[string]any{
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
	result, err := cdp.Mod.AddCustomCommand(CustomCommand{
		Name:         "Custom.doSomething",
		ParamsSchema: abxjsonschema.SchemaFor[ParamsSchema](),
		ResultSchema: abxjsonschema.SchemaFor[ResultSchema](),
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
	result, err := cdp.Mod.AddCustomEvent(CustomEvent{
		Name:        "Custom.someEvent",
		EventSchema: abxjsonschema.SchemaFor[EventSchema](),
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
		for _, entry := range cdp.handlers["Custom.someEvent"] {
			entry.handler(data)
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
