package injector_test

import (
	"fmt"
	modcdp "github.com/pirate/ModCDP/go/modcdp/client"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestBBBrowserExtensionInjectorUploadsRealExtensionAndLaunchesBrowserbaseBrowserWithItInstalled(t *testing.T) {
	if os.Getenv("BROWSERBASE_API_KEY") == "" {
		t.Fatal("BROWSERBASE_API_KEY is required for live Browserbase tests")
	}
	extensionPath, err := filepath.Abs("../../../dist/extension")
	if err != nil {
		t.Fatal(err)
	}
	launchOptions := modcdp.LaunchOptions{
		Timeout: 120,
	}
	if region := os.Getenv("BROWSERBASE_REGION"); region != "" {
		launchOptions.Region = region
	}
	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode:    "bb",
			LauncherOptions: launchOptions,
		},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "ws"},
		Injector: modcdp.InjectorConfig{
			InjectorMode:                     "inject",
			InjectorExtensionPath:                     extensionPath,
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["injector_source"] != "bb" {
		t.Fatalf("injector_source = %v", cdp.ConnectTiming["injector_source"])
	}
	if cdp.ExtensionID == "" {
		t.Fatal("expected ExtensionID")
	}
	result, err := cdp.Mod.Evaluate(map[string]any{
		"expression": "chrome.runtime.getURL('modcdp/service_worker.js')",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^chrome-extension://[a-z]{32}/modcdp/service_worker\.js$`).MatchString(fmt.Sprint(result)) {
		t.Fatalf("service worker url = %#v", result)
	}
}
