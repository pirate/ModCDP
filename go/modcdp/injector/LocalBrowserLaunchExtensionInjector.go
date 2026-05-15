package injector

import (
	"os"
)

type LocalBrowserLaunchExtensionInjector struct {
	ExtensionInjector
	UnpackedExtensionPath string
	ExtensionID           string
	CleanupPath           string
}

func NewLocalBrowserLaunchExtensionInjector(options ExtensionInjectorConfig) LocalBrowserLaunchExtensionInjector {
	return LocalBrowserLaunchExtensionInjector{ExtensionInjector: NewExtensionInjector(options)}
}

func (i *LocalBrowserLaunchExtensionInjector) Prepare() error {
	extensionPath := i.Options.InjectorExtensionPath
	if i.UnpackedExtensionPath != "" {
		return nil
	}
	unpackedPath, cleanupPath, err := prepareUnpackedExtension(extensionPath)
	if err != nil {
		return err
	}
	i.UnpackedExtensionPath = unpackedPath
	i.CleanupPath = cleanupPath
	_, err = i.resolveExtensionID()
	return err
}

func (i *LocalBrowserLaunchExtensionInjector) GetLauncherConfig() LaunchOptions {
	if i.UnpackedExtensionPath == "" {
		return LaunchOptions{}
	}
	return LaunchOptions{ExtraArgs: []string{"--load-extension=" + i.UnpackedExtensionPath}}
}

func (i *LocalBrowserLaunchExtensionInjector) Inject() (*ExtensionInjectionResult, error) {
	discovered, err := i.discoverReadyServiceWorker(i.Options.InjectorTrustServiceWorkerTarget)
	if err != nil || discovered == nil {
		return discovered, err
	}
	discovered.Source = "local_launch"
	return discovered, nil
}

func (i *LocalBrowserLaunchExtensionInjector) Close() error {
	if i.CleanupPath != "" {
		_ = os.RemoveAll(i.CleanupPath)
		i.CleanupPath = ""
	}
	return nil
}

func (i *LocalBrowserLaunchExtensionInjector) resolveExtensionID() (string, error) {
	if i.ExtensionID != "" {
		return i.ExtensionID, nil
	}
	if i.Options.InjectorExtensionID != "" {
		i.ExtensionID = i.Options.InjectorExtensionID
	} else if i.UnpackedExtensionPath != "" {
		extensionID, err := extensionIDFromManifestKey(i.UnpackedExtensionPath)
		if err != nil {
			return "", err
		}
		i.ExtensionID = extensionID
	}
	if i.ExtensionID != "" {
		i.Options.InjectorExtensionID = i.ExtensionID
	}
	return i.ExtensionID, nil
}
