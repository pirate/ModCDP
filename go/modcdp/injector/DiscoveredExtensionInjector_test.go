package injector_test

import (
	modcdp "github.com/browserbase/modcdp/go/modcdp/client"
	. "github.com/browserbase/modcdp/go/modcdp/injector"
	"path/filepath"
	"testing"
)

func TestDiscoveredExtensionInjectorAttachesToAlreadyLoadedRealModCDPExtension(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	headless := true
	owner := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "local", LauncherOptions: modcdp.LaunchOptions{Headless: &headless}},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "ws"},
		Injector: modcdp.InjectorConfig{
			InjectorMode:                     "auto",
			InjectorExtensionPath:            extensionPath,
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
	})
	defer owner.Close()

	if err := owner.Connect(); err != nil {
		t.Fatal(err)
	}
	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "remote"},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "ws", UpstreamCDPURL: owner.CDPURL},
		Injector: modcdp.InjectorConfig{
			InjectorMode:                     "discover",
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["injector_source"] != "discovered" {
		t.Fatalf("injector_source = %v", cdp.ConnectTiming["injector_source"])
	}
	if cdp.ExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("ExtensionID = %q", cdp.ExtensionID)
	}
	result, err := cdp.Mod.Evaluate(map[string]any{
		"expression": "chrome.runtime.getURL('modcdp/service_worker.js')",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js" {
		t.Fatalf("Mod.evaluate = %#v", result)
	}
}
