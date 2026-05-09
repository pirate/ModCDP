package modcdp

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestBBBrowserExtensionInjectorUsesConfiguredExtensionID(t *testing.T) {
	injector := NewBBBrowserExtensionInjector(ExtensionInjectorConfig{
		ExtensionID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if err := injector.Prepare(); err != nil {
		t.Fatal(err)
	}
	launchConfig := injector.GetLauncherConfig()
	if launchConfig.ExtensionID != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("ExtensionID = %q", launchConfig.ExtensionID)
	}
}

func TestBBBrowserExtensionInjectorRequiresAPIKeyWhenExtensionUploadIsNeeded(t *testing.T) {
	if strings.TrimSpace(os.Getenv("BROWSERBASE_API_KEY")) != "" {
		t.Skip("BROWSERBASE_API_KEY is set")
	}
	extensionPath, err := filepath.Abs("../../dist/extension")
	if err != nil {
		t.Fatal(err)
	}
	injector := NewBBBrowserExtensionInjector(ExtensionInjectorConfig{
		ExtensionPath: extensionPath,
	})
	if err := injector.Prepare(); err == nil || !strings.Contains(err.Error(), "BROWSERBASE_API_KEY") {
		t.Fatalf("expected missing key error, got %v", err)
	}
	if injector.CleanupPath != "" {
		t.Fatalf("CleanupPath = %q", injector.CleanupPath)
	}
}

func TestBBBrowserExtensionInjectorUploadsRealExtensionAndLaunchesBrowserbaseBrowserWithItInstalled(t *testing.T) {
	if strings.TrimSpace(os.Getenv("BROWSERBASE_API_KEY")) == "" {
		t.Skip("BROWSERBASE_API_KEY is required for live Browserbase tests")
	}
	extensionPath, err := filepath.Abs("../../dist/extension")
	if err != nil {
		t.Fatal(err)
	}
	launchOptions := LaunchOptions{
		ProjectID: os.Getenv("BROWSERBASE_PROJECT_ID"),
		Timeout:   120,
	}
	if region := os.Getenv("BROWSERBASE_REGION"); region != "" {
		launchOptions.Region = region
	}
	cdp := New(Options{
		Launch: LaunchConfig{
			Mode:    "bb",
			Options: launchOptions,
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
	if cdp.ConnectTiming["extension_source"] != "bb" {
		t.Fatalf("extension_source = %v", cdp.ConnectTiming["extension_source"])
	}
	if cdp.ExtensionID == "" {
		t.Fatal("expected ExtensionID")
	}
	result, err := cdp.Send("Mod.evaluate", map[string]any{
		"expression": "chrome.runtime.getURL('modcdp/service_worker.js')",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^chrome-extension://[a-z]{32}/modcdp/service_worker\.js$`).MatchString(fmt.Sprint(result)) {
		t.Fatalf("service worker url = %#v", result)
	}
}
