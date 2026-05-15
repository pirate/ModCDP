package injector_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/browserbase/modcdp/go/modcdp/injector"
)

func TestExtensionsLoadUnpackedInjectorPreparesDefaultPackagedExtensionZip(t *testing.T) {
	injector := NewExtensionsLoadUnpackedInjector(ExtensionInjectorConfig{})
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
}
