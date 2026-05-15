package injector_test

import (
	"strings"
	"testing"

	. "github.com/browserbase/modcdp/go/modcdp/injector"
)

func TestExtensionInjectorOwnsSharedInjectorConfig(t *testing.T) {
	injector := NewExtensionInjector(ExtensionInjectorConfig{
		InjectorExtensionID:              "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
	})

	transportConfig := injector.GetTransportConfig()
	if transportConfig["injector_extension_id"] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("injector_extension_id = %v", transportConfig["injector_extension_id"])
	}
	if len(injector.GetLauncherConfig().ExtraArgs) != 0 {
		t.Fatalf("expected empty launcher config")
	}
	if !injector.ServiceWorkerTargetMatches(map[string]any{
		"targetId": "target-1",
		"type":     "service_worker",
		"url":      "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/modcdp/service_worker.js",
	}) {
		t.Fatal("expected service worker target to match")
	}
	if injector.ServiceWorkerTargetMatches(map[string]any{
		"targetId": "target-1",
		"type":     "service_worker",
		"url":      "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/background.js",
	}) {
		t.Fatal("expected background target not to match")
	}
}

func TestExtensionInjectorBaseInjectReportsTheClassName(t *testing.T) {
	injector := NewExtensionInjector(ExtensionInjectorConfig{})
	if _, err := injector.Inject(); err == nil || !strings.Contains(err.Error(), "ExtensionInjector.Inject is not implemented") {
		t.Fatalf("Inject error = %v", err)
	}
}
