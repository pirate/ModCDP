package injector_test

import (
	"encoding/json"
	modcdp "github.com/browserbase/modcdp/go/modcdp/client"
	. "github.com/browserbase/modcdp/go/modcdp/injector"
	"os"
	"path/filepath"
	"testing"
)

const customDiscoveredExtensionID = "hhklgmbgnbeghnjidampacgmgnhelifg"
const customDiscoveredExtensionPublicKey = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzG1LUbtH0aHMKjTAUeT0saY8xfnRNENctFJme3C1qnsqT7PAXMxJC4nT7tBZy2gEGRirBb3zIZ3OyAu9a0QR8lTLupDp4qHWOhQ7dl9ZjxjQdYa4Gby0xuXLdQrJIxDbmuv+UVJvYa8vRTwQB8koygbzDDDP5/YiB6mc0hbh8XBb82Ossy7T280k8280o/rS0CXdioUraCHj58PDhfxbs18TBcYfOjuRqua9J2oddxobtGehSD0gDtbvn2IWDtRajOlgZZyuS1vLoSR7C1ulFzpRSYPEMhI2x+wphut7E3QImyJ577YeULVGpt988FcixOou7udjx3/IUWjpq8046wIDAQAB"

func TestDiscoveredExtensionInjectorAttachesToAlreadyLoadedRealModCDPExtension(t *testing.T) {
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

func TestDiscoveredExtensionInjectorSelectsConfiguredExtensionWhenMultipleModCDPWorkersExist(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	customExtensionPath := filepath.Join(t.TempDir(), "extension")
	if err := copyDir(extensionPath, customExtensionPath); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(customExtensionPath, "manifest.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatal(err)
	}
	manifest["key"] = customDiscoveredExtensionPublicKey
	manifest["name"] = "ModCDP Bridge Custom Test"
	updatedManifest, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifestPath, append(updatedManifest, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}

	chrome, err := modcdp.NewLocalBrowserLauncher(modcdp.LaunchOptions{
		Headless:  boolPtr(true),
		Sandbox:   boolPtr(false),
		ExtraArgs: []string{"--load-extension=" + extensionPath + "," + customExtensionPath},
	}).Launch(modcdp.LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "remote"},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "ws", UpstreamCDPURL: chrome.CDPURL},
		Injector: modcdp.InjectorConfig{
			InjectorMode:                       "discover",
			InjectorExtensionID:                customDiscoveredExtensionID,
			InjectorServiceWorkerURLSuffixes:   []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget:   true,
			InjectorRequireServiceWorkerTarget: true,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["injector_source"] != "discovered" {
		t.Fatalf("injector_source = %v", cdp.ConnectTiming["injector_source"])
	}
	if cdp.ExtensionID != customDiscoveredExtensionID {
		t.Fatalf("ExtensionID = %q", cdp.ExtensionID)
	}
	result, err := cdp.Mod.Evaluate(map[string]any{"expression": "chrome.runtime.id"})
	if err != nil {
		t.Fatal(err)
	}
	if result != customDiscoveredExtensionID {
		t.Fatalf("Mod.evaluate = %#v", result)
	}

	targetsResult, err := cdp.SendRaw("Target.getTargets", nil)
	if err != nil {
		t.Fatal(err)
	}
	targets, _ := targetsResult["targetInfos"].([]any)
	foundCustom := false
	foundDefault := false
	for _, rawTarget := range targets {
		target, _ := rawTarget.(map[string]any)
		if target["type"] != "service_worker" {
			continue
		}
		url, _ := target["url"].(string)
		switch url {
		case "chrome-extension://" + customDiscoveredExtensionID + "/modcdp/service_worker.js":
			foundCustom = true
		case "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js":
			foundDefault = true
		}
	}
	if !foundCustom || !foundDefault {
		t.Fatalf("expected custom and default ModCDP workers, custom=%v default=%v targets=%#v", foundCustom, foundDefault, targets)
	}
}
