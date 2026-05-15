package injector

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
	discovered, err := i.discoverReadyServiceWorker(false)
	if err != nil || discovered != nil {
		if discovered != nil {
			discovered.Source = "discovered"
		}
		return discovered, err
	}
	if i.Options.InjectorTrustServiceWorkerTarget {
		waited, err := i.waitForReadyServiceWorker(i.Options.InjectorServiceWorkerProbeTimeoutMS, true)
		if err != nil || waited != nil {
			if waited != nil {
				waited.Source = "discovered"
			}
			return waited, err
		}
	}
	if !i.Options.InjectorRequireServiceWorkerTarget {
		return nil, nil
	}
	waited, err := i.waitForReadyServiceWorker(i.Options.InjectorServiceWorkerReadyTimeoutMS, i.Options.InjectorTrustServiceWorkerTarget)
	if err != nil || waited != nil {
		if waited != nil {
			waited.Source = "discovered"
		}
		return waited, err
	}
	matchers := append(append([]string{}, i.Options.InjectorServiceWorkerURLIncludes...), i.Options.InjectorServiceWorkerURLSuffixes...)
	matcherText := strings.Join(matchers, ", ")
	if matcherText == "" {
		matcherText = "no matcher"
	}
	return nil, fmt.Errorf("required ModCDP service worker target was not visible (%s)", matcherText)
}
