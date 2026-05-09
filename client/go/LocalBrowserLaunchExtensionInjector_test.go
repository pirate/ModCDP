package modcdp

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalBrowserLaunchExtensionInjectorLoadsRealExtensionDuringLocalLaunch(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	cdp := New(Options{
		Launch: LaunchConfig{
			Mode: "local",
			Options: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{Mode: "ws"},
		Extension: ExtensionConfig{
			Mode:                     "inject",
			Path:                     extensionPath,
			ServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			TrustServiceWorkerTarget: true,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["extension_source"] != "local_launch" {
		t.Fatalf("extension_source = %v", cdp.ConnectTiming["extension_source"])
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
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	injector := NewLocalBrowserLaunchExtensionInjector(ExtensionInjectorConfig{ExtensionPath: extensionPath})
	if err := injector.Prepare(); err != nil {
		t.Fatal(err)
	}
	defer injector.Close()

	launchConfig := injector.GetLauncherConfig()
	if len(launchConfig.ExtraArgs) != 1 || !strings.HasPrefix(launchConfig.ExtraArgs[0], "--load-extension=") {
		t.Fatalf("ExtraArgs = %v", launchConfig.ExtraArgs)
	}
	if injector.Options.ExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("ExtensionID = %q", injector.Options.ExtensionID)
	}
}
