package modcdp

import (
	"fmt"
	"strings"
)

type DiscoveredExtensionInjector struct {
	ExtensionInjector
}

func NewDiscoveredExtensionInjector(options ExtensionInjectorConfig) DiscoveredExtensionInjector {
	return DiscoveredExtensionInjector{ExtensionInjector: NewExtensionInjector(options)}
}

func (i *DiscoveredExtensionInjector) Inject() (*ExtensionInjectionResult, error) {
	discovered, err := i.DiscoverReadyServiceWorker(false)
	if err != nil || discovered != nil {
		if discovered != nil {
			discovered.Source = "discovered"
		}
		return discovered, err
	}
	if i.Options.TrustMatchedServiceWorker {
		waited, err := i.WaitForReadyServiceWorker(i.Options.ServiceWorkerProbeTimeoutMS, true)
		if err != nil || waited != nil {
			if waited != nil {
				waited.Source = "discovered"
			}
			return waited, err
		}
	}
	if i.WakeConfiguredExtension() {
		waited, err := i.WaitForReadyServiceWorker(i.Options.ServiceWorkerProbeTimeoutMS, i.Options.TrustMatchedServiceWorker)
		if err != nil || waited != nil {
			if waited != nil {
				waited.Source = "discovered"
			}
			return waited, err
		}
	}
	if !i.Options.RequireServiceWorkerTarget {
		return nil, nil
	}
	waited, err := i.WaitForReadyServiceWorker(i.Options.ServiceWorkerReadyTimeoutMS, i.Options.TrustMatchedServiceWorker)
	if err != nil || waited != nil {
		if waited != nil {
			waited.Source = "discovered"
		}
		return waited, err
	}
	matchers := append(append([]string{}, i.Options.ServiceWorkerURLIncludes...), i.Options.ServiceWorkerURLSuffixes...)
	matcherText := strings.Join(matchers, ", ")
	if matcherText == "" {
		matcherText = "no matcher"
	}
	return nil, fmt.Errorf("required ModCDP service worker target was not visible (%s)", matcherText)
}
