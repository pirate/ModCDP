package injector_test

import (
	"archive/zip"
	modcdp "github.com/pirate/ModCDP/go/modcdp/client"
	. "github.com/pirate/ModCDP/go/modcdp/injector"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLocalBrowserLaunchExtensionInjectorRejectsZipEntriesOutsideExtractionDir(t *testing.T) {
	tempDir := t.TempDir()
	zipPath := filepath.Join(tempDir, "extension.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	entry, err := writer.Create("../evil.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte("evil")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	injector := NewLocalBrowserLaunchExtensionInjector(ExtensionInjectorConfig{InjectorExtensionPath: zipPath})
	if err := injector.Prepare(); err == nil || !strings.Contains(err.Error(), "escapes extension extraction directory") {
		t.Fatalf("Prepare error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(tempDir, "evil.txt")); !os.IsNotExist(err) {
		t.Fatalf("outside file was created: %v", err)
	}
}

func TestLocalBrowserLaunchExtensionInjectorLoadsRealExtensionDuringLocalLaunch(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	headless := runtime.GOOS == "linux" && os.Getenv("DISPLAY") == ""
	sandbox := runtime.GOOS != "linux"
	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "local",
			LauncherOptions: modcdp.LaunchOptions{
				Headless: boolPtr(headless),
				Sandbox:  boolPtr(sandbox),
			},
		},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "ws"},
		Injector: modcdp.InjectorConfig{
			InjectorMode:                        "inject",
			InjectorExtensionPath:               extensionPath,
			InjectorServiceWorkerURLSuffixes:    []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget:    true,
			InjectorServiceWorkerProbeTimeoutMS: 30_000,
		},
		Client: modcdp.ClientConfig{
			ClientCDPSendTimeoutMS: 30_000,
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
