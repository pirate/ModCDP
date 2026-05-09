package modcdp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtensionInjectorOwnsSharedConfigAndRuntimeTransportConfig(t *testing.T) {
	injector := NewExtensionInjector(ExtensionInjectorConfig{
		ExtensionID:              "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		ServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
		ReverseProxyURL:          "ws://127.0.0.1:29292",
	})
	injector.Update(ExtensionInjectorConfig{NativeHostName: "com.modcdp.bridge"})

	transportConfig := injector.GetTransportConfig()
	if transportConfig["extension_id"] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("extension_id = %v", transportConfig["extension_id"])
	}
	if len(injector.GetLauncherConfig().ExtraArgs) != 0 {
		t.Fatalf("expected empty launcher config")
	}
	if !injector.ServiceWorkerTargetMatches(map[string]any{
		"targetId": "target-1",
		"type":     "service_worker",
		"url":      "chrome-extension://aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/modcdp/service_worker.js",
	}) {
		t.Fatalf("expected service worker target to match")
	}

	extensionPath := t.TempDir()
	if err := injector.WriteExtensionRuntimeConfig(extensionPath); err != nil {
		t.Fatal(err)
	}
	config, err := os.ReadFile(filepath.Join(extensionPath, "modcdp", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  \"native_host_name\": \"com.modcdp.bridge\",\n  \"reverse_proxy_url\": \"ws://127.0.0.1:29292\"\n}\n"
	if string(config) != expected {
		t.Fatalf("config.json = %s", config)
	}
	if _, err := injector.Inject(); err == nil {
		t.Fatalf("expected base Inject to fail")
	}
}
