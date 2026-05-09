package modcdp

import (
	"path/filepath"
	"testing"
)

func TestDiscoveredExtensionInjectorAttachesToAlreadyLoadedRealModCDPExtension(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := NewLocalBrowserLauncher(LaunchOptions{
		Headless:  boolPtr(true),
		Sandbox:   boolPtr(false),
		ExtraArgs: []string{"--load-extension=" + extensionPath},
	}).Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	cdp := New(Options{
		Launch:   LaunchConfig{Mode: "remote"},
		Upstream: UpstreamConfig{Mode: "ws", WSURL: chrome.CDPURL},
		Extension: ExtensionConfig{
			Mode:                     "discover",
			ServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			TrustServiceWorkerTarget: true,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["extension_source"] != "discovered" {
		t.Fatalf("extension_source = %v", cdp.ConnectTiming["extension_source"])
	}
	if cdp.ExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("ExtensionID = %q", cdp.ExtensionID)
	}
	result, err := cdp.Send("Mod.evaluate", map[string]any{
		"expression": "chrome.runtime.getURL('modcdp/service_worker.js')",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js" {
		t.Fatalf("Mod.evaluate = %#v", result)
	}
}
