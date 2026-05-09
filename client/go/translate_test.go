package modcdp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTranslateRoutesWrapsAndUnwrapsModCDPProtocolMessagesDeterministically(t *testing.T) {
	if routeFor("Browser.getVersion", map[string]string{"Browser.*": "direct_cdp", "*.*": "service_worker"}) != "direct_cdp" {
		t.Fatal("Browser.getVersion route mismatch")
	}
	if routeFor("Target.getTargets", map[string]string{"Browser.*": "direct_cdp", "*.*": "service_worker"}) != "service_worker" {
		t.Fatal("Target.getTargets route mismatch")
	}

	direct, err := wrapCommandIfNeeded("Browser.getVersion", map[string]any{}, map[string]string{"*.*": "direct_cdp"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if direct.Target != "direct_cdp" || len(direct.Steps) != 1 || direct.Steps[0].Method != "Browser.getVersion" {
		t.Fatalf("direct = %#v", direct)
	}

	wrapped, err := wrapCommandIfNeeded(
		"Mod.evaluate",
		map[string]any{"expression": "({ ok: true })", "params": map[string]any{"value": 1}},
		DefaultClientRoutes(),
		"session-1",
	)
	if err != nil {
		t.Fatal(err)
	}
	if wrapped.Target != "service_worker" {
		t.Fatalf("wrapped.Target = %q", wrapped.Target)
	}
	if wrapped.Steps[0].Method != "Runtime.evaluate" {
		t.Fatalf("wrapped step = %#v", wrapped.Steps[0])
	}
	if !strings.Contains(stringValue(wrapped.Steps[0].Params["expression"]), `attachToSession("session-1")`) {
		t.Fatalf("expression = %s", wrapped.Steps[0].Params["expression"])
	}
	if wrapped.Steps[0].Unwrap != "evaluate" {
		t.Fatalf("unwrap = %q", wrapped.Steps[0].Unwrap)
	}

	unwrapped, err := unwrapResponseIfNeeded(map[string]any{"result": map[string]any{"type": "object", "value": map[string]any{"ok": true}}}, "evaluate")
	if err != nil {
		t.Fatal(err)
	}
	if unwrapped.(map[string]any)["ok"] != true {
		t.Fatalf("unwrapped = %#v", unwrapped)
	}
	raw, err := unwrapResponseIfNeeded(map[string]any{"product": "Chrome/1"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if raw.(map[string]any)["product"] != "Chrome/1" {
		t.Fatalf("raw = %#v", raw)
	}

	payload, _ := json.Marshal(map[string]any{
		"event":        "Custom.ready",
		"data":         map[string]any{"ready": true},
		"cdpSessionId": "session-2",
	})
	event, data, ok := unwrapEventIfNeeded(
		"Runtime.bindingCalled",
		map[string]any{"name": customEventBindingName, "payload": string(payload)},
		"session-1",
		"session-1",
	)
	if !ok || event != "Custom.ready" || data.(map[string]any)["ready"] != true {
		t.Fatalf("event=%q data=%#v ok=%v", event, data, ok)
	}
	if _, _, ok := unwrapEventIfNeeded("Runtime.consoleAPICalled", map[string]any{"name": customEventBindingName, "payload": string(payload)}, "", ""); ok {
		t.Fatal("expected console event to ignore binding payload")
	}
}
