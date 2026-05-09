package modcdp

import "testing"

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
