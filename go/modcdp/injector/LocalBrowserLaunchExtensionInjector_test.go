package injector_test

import (
	"archive/zip"
	. "github.com/browserbase/modcdp/go/modcdp/injector"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestLocalBrowserLaunchExtensionInjectorPreparesUnpackedExtensionDirectoryForLoadExtension(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	injector := NewLocalBrowserLaunchExtensionInjector(ExtensionInjectorConfig{InjectorExtensionPath: extensionPath})
	if err := injector.Prepare(); err != nil {
		t.Fatal(err)
	}
	defer injector.Close()

	if injector.UnpackedExtensionPath == "" {
		t.Fatal("expected UnpackedExtensionPath")
	}
	if injector.UnpackedExtensionPath == extensionPath {
		t.Fatalf("UnpackedExtensionPath = %q", injector.UnpackedExtensionPath)
	}
	if _, err := os.Stat(filepath.Join(injector.UnpackedExtensionPath, "manifest.json")); err != nil {
		t.Fatalf("expected unpacked manifest: %v", err)
	}
	launcherConfig := injector.GetLauncherConfig()
	if len(launcherConfig.ExtraArgs) != 1 || launcherConfig.ExtraArgs[0] != "--load-extension="+injector.UnpackedExtensionPath {
		t.Fatalf("ExtraArgs = %#v", launcherConfig.ExtraArgs)
	}
	if injector.Options.InjectorExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("InjectorExtensionID = %q", injector.Options.InjectorExtensionID)
	}
}

func TestLocalBrowserLaunchExtensionInjectorPreparesDefaultExtensionZipForLoadExtension(t *testing.T) {
	injector := NewLocalBrowserLaunchExtensionInjector(ExtensionInjectorConfig{})
	if err := injector.Prepare(); err != nil {
		t.Fatal(err)
	}
	defer injector.Close()

	if injector.UnpackedExtensionPath == "" {
		t.Fatal("expected UnpackedExtensionPath")
	}
	if !strings.Contains(injector.UnpackedExtensionPath, "modcdp-extension-") {
		t.Fatalf("UnpackedExtensionPath = %q", injector.UnpackedExtensionPath)
	}
	if _, err := os.Stat(filepath.Join(injector.UnpackedExtensionPath, "manifest.json")); err != nil {
		t.Fatalf("expected unpacked manifest: %v", err)
	}
	launcherConfig := injector.GetLauncherConfig()
	if len(launcherConfig.ExtraArgs) != 1 || launcherConfig.ExtraArgs[0] != "--load-extension="+injector.UnpackedExtensionPath {
		t.Fatalf("ExtraArgs = %#v", launcherConfig.ExtraArgs)
	}
	if injector.Options.InjectorExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("InjectorExtensionID = %q", injector.Options.InjectorExtensionID)
	}
}

func TestLocalBrowserLaunchExtensionInjectorReturnsImmediatelyWhenLaunchedExtensionTargetIsAbsent(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	methods := []string{}
	injector := NewLocalBrowserLaunchExtensionInjector(ExtensionInjectorConfig{
		InjectorExtensionPath:            extensionPath,
		InjectorTrustServiceWorkerTarget: true,
		Send: func(method string, params map[string]any, sessionID string) (map[string]any, error) {
			methods = append(methods, method)
			if method == "Target.getTargets" {
				return map[string]any{"targetInfos": []any{}}, nil
			}
			t.Fatalf("unexpected %s", method)
			return nil, nil
		},
	})
	if err := injector.Prepare(); err != nil {
		t.Fatal(err)
	}
	defer injector.Close()

	startedAt := time.Now()
	result, err := injector.Inject()
	if err != nil {
		t.Fatal(err)
	}
	elapsed := time.Since(startedAt)
	if result != nil {
		t.Fatalf("result = %#v", result)
	}
	if strings.Join(methods, ",") != "Target.getTargets" {
		t.Fatalf("methods = %#v", methods)
	}
	if elapsed >= 200*time.Millisecond {
		t.Fatalf("Inject took %s", elapsed)
	}
}
