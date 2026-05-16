package translate

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
	if wrapped.Steps[0].Method != "Runtime.callFunctionOn" {
		t.Fatalf("wrapped step = %#v", wrapped.Steps[0])
	}
	if !strings.Contains(stringValue(wrapped.Steps[0].Params["functionDeclaration"]), `attachToSession("session-1")`) {
		t.Fatalf("functionDeclaration = %s", wrapped.Steps[0].Params["functionDeclaration"])
	}
	if wrapped.Steps[0].Unwrap != "runtime" {
		t.Fatalf("unwrap = %q", wrapped.Steps[0].Unwrap)
	}

	configured, err := wrapCommandIfNeeded("Mod.configure", map[string]any{"server": map[string]any{"server_routes": map[string]any{"*.*": "loopback_cdp"}}}, DefaultClientRoutes(), "session-1")
	if err != nil {
		t.Fatal(err)
	}
	if configured.Steps[0].Unwrap != "runtime_json" {
		t.Fatalf("configure unwrap = %q", configured.Steps[0].Unwrap)
	}

	custom, err := wrapCommandIfNeeded(
		"Custom.echo",
		map[string]any{"secret": strings.Repeat("x", 100), "nested": map[string]any{"ok": true}},
		DefaultClientRoutes(),
		"session-1",
	)
	if err != nil {
		t.Fatal(err)
	}
	customParams := custom.Steps[0].Params
	if !strings.Contains(stringValue(customParams["functionDeclaration"]), "JSON.parse(paramsJson)") {
		t.Fatalf("functionDeclaration = %s", customParams["functionDeclaration"])
	}
	if strings.Contains(stringValue(customParams["functionDeclaration"]), "xxxxxxxxxx") {
		t.Fatalf("functionDeclaration includes custom params: %s", customParams["functionDeclaration"])
	}
	customArguments := customParams["arguments"].([]map[string]any)
	if customArguments[0]["value"] != "Custom.echo" {
		t.Fatalf("method argument = %#v", customArguments[0])
	}
	var customPayload map[string]any
	if err := json.Unmarshal([]byte(customArguments[1]["value"].(string)), &customPayload); err != nil {
		t.Fatal(err)
	}
	if customPayload["secret"] != strings.Repeat("x", 100) || customPayload["nested"].(map[string]any)["ok"] != true {
		t.Fatalf("params argument = %#v", customPayload)
	}
	if customArguments[2]["value"] != "session-1" {
		t.Fatalf("session argument = %#v", customArguments[2])
	}

	unwrapped, err := unwrapResponseIfNeeded(map[string]any{"result": map[string]any{"type": "object", "value": map[string]any{"ok": true}}}, "runtime")
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
