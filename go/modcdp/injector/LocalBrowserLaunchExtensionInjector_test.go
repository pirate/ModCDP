package injector_test

import (
	modcdp "github.com/pirate/ModCDP/go/modcdp/client"
	. "github.com/pirate/ModCDP/go/modcdp/injector"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalBrowserLaunchExtensionInjectorLoadsRealExtensionDuringLocalLaunch(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "local",
			LauncherOptions: modcdp.LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "ws"},
		Injector: modcdp.InjectorConfig{
			InjectorMode:                     "inject",
			InjectorExtensionPath:            extensionPath,
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["injector_source"] != "local_launch" {
		t.Fatalf("injector_source = %v", cdp.ConnectTiming["injector_source"])
	}
	if cdp.ExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("ExtensionID = %q", cdp.ExtensionID)
	}
	if cdp.ExtSessionID == "" {
		t.Fatal("expected ExtSessionID")
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

func TestLocalBrowserLaunchExtensionInjectorPreparesLauncherConfig(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	injector := NewLocalBrowserLaunchExtensionInjector(ExtensionInjectorConfig{InjectorExtensionPath: extensionPath})
	if err := injector.Prepare(); err != nil {
		t.Fatal(err)
	}
	defer injector.Close()

	launchConfig := injector.GetLauncherConfig()
	if len(launchConfig.ExtraArgs) != 1 || !strings.HasPrefix(launchConfig.ExtraArgs[0], "--load-extension=") {
		t.Fatalf("ExtraArgs = %v", launchConfig.ExtraArgs)
	}
	if injector.ExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("ExtensionID = %q", injector.ExtensionID)
	}
}
