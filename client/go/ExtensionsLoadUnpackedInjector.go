package modcdp

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type ExtensionsLoadUnpackedInjector struct {
	ExtensionInjector
	UnpackedExtensionPath string
	CleanupPath           string
}

func NewExtensionsLoadUnpackedInjector(options ExtensionInjectorConfig) ExtensionsLoadUnpackedInjector {
	return ExtensionsLoadUnpackedInjector{ExtensionInjector: NewExtensionInjector(options)}
}

func (i *ExtensionsLoadUnpackedInjector) Prepare() error {
	extensionPath := i.Options.ExtensionPath
	if extensionPath == "" || i.UnpackedExtensionPath != "" {
		return nil
	}
	unpackedPath, cleanupPath, err := prepareUnpackedExtension(extensionPath, len(i.ExtensionRuntimeConfig()) > 0)
	if err != nil {
		return err
	}
	i.UnpackedExtensionPath = unpackedPath
	i.CleanupPath = cleanupPath
	return i.WriteExtensionRuntimeConfig(i.UnpackedExtensionPath)
}

func (i *ExtensionsLoadUnpackedInjector) Inject() (*ExtensionInjectionResult, error) {
	if i.UnpackedExtensionPath == "" {
		return nil, nil
	}
	loadResult, err := i.SendWithTimeout("Extensions.loadUnpacked", map[string]any{"path": i.UnpackedExtensionPath}, "", i.Options.CDPSendTimeoutMS)
	if err != nil {
		if strings.Contains(err.Error(), "Method not available") || strings.Contains(err.Error(), "Method not found") || strings.Contains(err.Error(), "wasn't found") {
			i.LastError = err
			return nil, nil
		}
		return nil, fmt.Errorf("Extensions.loadUnpacked failed for %s: %w", i.UnpackedExtensionPath, err)
	}
	extensionID, _ := loadResult["id"].(string)
	if extensionID == "" {
		extensionID, _ = loadResult["extensionId"].(string)
	}
	if extensionID == "" {
		return nil, fmt.Errorf("Extensions.loadUnpacked returned no extension id")
	}
	i.Options.ExtensionID = extensionID
	i.WakeConfiguredExtension()
	swURLPrefix := "chrome-extension://" + extensionID + "/"
	deadline := time.Now().Add(time.Duration(i.Options.ServiceWorkerReadyTimeoutMS) * time.Millisecond)
	for time.Now().Before(deadline) {
		targets, err := i.TargetInfos()
		if err != nil {
			return nil, err
		}
		for _, target := range targets {
			targetType, _ := target["type"].(string)
			targetURL, _ := target["url"].(string)
			if targetType != "service_worker" || !strings.HasPrefix(targetURL, swURLPrefix) {
				continue
			}
			probed, err := i.ProbeTarget(target, i.Options.ServiceWorkerProbeTimeoutMS, true)
			if err != nil {
				return nil, err
			}
			if probed != nil {
				probed.Source = "extensions_load_unpacked"
				probed.ExtensionID = extensionID
				return probed, nil
			}
		}
		time.Sleep(time.Duration(i.Options.ServiceWorkerPollIntervalMS) * time.Millisecond)
	}
	return nil, fmt.Errorf("timed out waiting for service worker target for extension %s", extensionID)
}

func (i *ExtensionsLoadUnpackedInjector) Close() error {
	if i.CleanupPath != "" {
		_ = os.RemoveAll(i.CleanupPath)
		i.CleanupPath = ""
	}
	return nil
}
