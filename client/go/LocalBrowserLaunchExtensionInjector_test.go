package modcdp

import (
	"path/filepath"
	"strings"
	"testing"
)

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
