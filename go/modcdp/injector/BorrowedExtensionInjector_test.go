package injector_test

import (
	modcdp "github.com/pirate/ModCDP/go/modcdp/client"
	. "github.com/pirate/ModCDP/go/modcdp/injector"
	"path/filepath"
	"testing"
)

func TestBorrowedExtensionInjectorBootstrapsModCDPInsideLiveExtensionServiceWorker(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := modcdp.NewLocalBrowserLauncher(modcdp.LaunchOptions{
		Headless:  boolPtr(true),
		Sandbox:   boolPtr(false),
		ExtraArgs: []string{"--load-extension=" + extensionPath},
	}).Launch(modcdp.LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "remote"},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "ws", UpstreamCDPURL: chrome.CDPURL},
		Injector: modcdp.InjectorConfig{
			InjectorMode:                     "borrow",
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["injector_source"] != "borrowed" {
		t.Fatalf("injector_source = %v", cdp.ConnectTiming["injector_source"])
	}
	if cdp.ExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("ExtensionID = %q", cdp.ExtensionID)
	}
	result, err := cdp.Send("Target.getTargets", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	targetInfos, _ := result.(map[string]any)["targetInfos"].([]any)
	if len(targetInfos) == 0 {
		t.Fatalf("Target.getTargets = %#v", result)
	}
}
