package client

import (
	"testing"

	abxjsonschema "github.com/ArchiveBox/abxbus/abxbus-go/v2/jsonschema"
)

type protocolValidationCustomEchoParams struct {
	Text string `json:"text"`
}

type protocolValidationCustomEchoResult struct {
	OK bool `json:"ok"`
}

type protocolValidationCustomReadyEvent struct {
	OK bool `json:"ok"`
}

func TestProtocolValidationCoversNativeMethodsNativeEventsCustomMethodsCustomEventsAndNativeOverrides(t *testing.T) {
	cdp := New(Options{})

	runtimeParams := RuntimeEvaluateParams{Expression: "1 + 1", ReturnByValue: Bool(true)}
	runtimeResult := RuntimeEvaluateResult{Result: RuntimeRemoteObject{Type: "number", Value: 2}}
	nativeEvent := TargetTargetCreatedEvent{
		TargetInfo: TargetTargetInfo{
			TargetID:        TargetTargetID("target-1"),
			Type:            "page",
			Title:           "Example",
			URL:             "https://example.com",
			Attached:        false,
			CanAccessOpener: false,
		},
	}
	customParams := protocolValidationCustomEchoParams{Text: "ok"}
	customResult := protocolValidationCustomEchoResult{OK: true}
	customEvent := protocolValidationCustomReadyEvent{OK: true}

	if err := cdp.validateCommandParams("Runtime.evaluate", mustParamsMap(t, runtimeParams)); err != nil {
		t.Fatalf("native Runtime.evaluate params should validate: %v", err)
	}
	if err := cdp.validateCommandResult("Runtime.evaluate", runtimeResult); err != nil {
		t.Fatalf("native Runtime.evaluate result should validate: %v", err)
	}
	if _, ok := cdp.validateEventData("Target.targetCreated", nativeEvent); !ok {
		t.Fatal("native Target.targetCreated event should validate")
	}
	if err := cdp.validateCommandParams("Runtime.evaluate", map[string]any{}); err == nil {
		t.Fatal("expected Runtime.evaluate params validation to reject missing expression")
	}
	if err := cdp.validateCommandResult("Runtime.evaluate", map[string]any{}); err == nil {
		t.Fatal("expected Runtime.evaluate result validation to reject missing result")
	}
	expectPanic(t, func() { cdp.validateEventData("Target.targetCreated", map[string]any{}) })

	if _, err := cdp.Mod.AddCustomCommand(CustomCommand{
		Name:         "Custom.echo",
		ParamsSchema: abxjsonschema.SchemaFor[protocolValidationCustomEchoParams](),
		ResultSchema: abxjsonschema.SchemaFor[protocolValidationCustomEchoResult](),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := cdp.Mod.AddCustomEvent(CustomEvent{
		Name:        "Custom.ready",
		EventSchema: abxjsonschema.SchemaFor[protocolValidationCustomReadyEvent](),
	}); err != nil {
		t.Fatal(err)
	}

	if err := cdp.validateCommandParams("Custom.echo", mustParamsMap(t, customParams)); err != nil {
		t.Fatalf("custom params should validate: %v", err)
	}
	if err := cdp.validateCommandResult("Custom.echo", customResult); err != nil {
		t.Fatalf("custom result should validate: %v", err)
	}
	if _, ok := cdp.validateEventData("Custom.ready", customEvent); !ok {
		t.Fatal("custom event should validate")
	}
	if err := cdp.validateCommandParams("Custom.echo", map[string]any{"text": 1}); err == nil {
		t.Fatal("expected custom params validation to reject wrong text type")
	}
	if err := cdp.validateCommandResult("Custom.echo", map[string]any{"ok": "yes"}); err == nil {
		t.Fatal("expected custom result validation to reject wrong ok type")
	}
	expectPanic(t, func() { cdp.validateEventData("Custom.ready", map[string]any{"ok": "yes"}) })

	if _, err := cdp.Mod.AddCustomCommand(CustomCommand{
		Name: "Target.getTargets",
		ResultSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"targetInfos": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"targetId":        map[string]any{"type": "string"},
							"type":            map[string]any{"type": "string"},
							"title":           map[string]any{"type": "string"},
							"url":             map[string]any{"type": "string"},
							"attached":        map[string]any{"type": "boolean"},
							"canAccessOpener": map[string]any{"type": "boolean"},
							"tabId":           map[string]any{"type": "integer"},
						},
						"required":             []any{"targetId", "type", "title", "url", "attached", "canAccessOpener"},
						"additionalProperties": true,
					},
				},
			},
			"required":             []any{"targetInfos"},
			"additionalProperties": true,
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := cdp.Mod.AddCustomEvent(CustomEvent{Name: "Target.targetCreated"}); err != nil {
		t.Fatal(err)
	}

	extendedTargetInfo := map[string]any{
		"targetId":        "target-1",
		"type":            "page",
		"title":           "Example",
		"url":             "https://example.com",
		"attached":        false,
		"canAccessOpener": false,
		"tabId":           7,
	}
	if err := cdp.validateCommandResult("Target.getTargets", map[string]any{"targetInfos": []any{extendedTargetInfo}}); err != nil {
		t.Fatalf("extended native command result should validate: %v", err)
	}
	if _, ok := cdp.validateEventData("Target.targetCreated", map[string]any{"targetInfo": extendedTargetInfo}); !ok {
		t.Fatal("extended native event should validate")
	}
}

func mustParamsMap(t *testing.T, value any) map[string]any {
	t.Helper()
	params, err := cdpParamsMap(value)
	if err != nil {
		t.Fatal(err)
	}
	return params
}
