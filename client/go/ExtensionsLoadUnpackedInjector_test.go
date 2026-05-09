package modcdp

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtensionsLoadUnpackedInjectorPreparesRuntimeConfigCopy(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	injector := NewExtensionsLoadUnpackedInjector(ExtensionInjectorConfig{
		ExtensionPath:   extensionPath,
		ReverseProxyURL: "ws://127.0.0.1:29292",
	})
	if err := injector.Prepare(); err != nil {
		t.Fatal(err)
	}
	defer injector.Close()

	if injector.UnpackedExtensionPath == extensionPath {
		t.Fatalf("expected runtime config to use a copied extension path")
	}
	config, err := os.ReadFile(filepath.Join(injector.UnpackedExtensionPath, "modcdp", "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(config) != "{\n  \"reverse_proxy_url\": \"ws://127.0.0.1:29292\"\n}\n" {
		t.Fatalf("config.json = %s", config)
	}
}
