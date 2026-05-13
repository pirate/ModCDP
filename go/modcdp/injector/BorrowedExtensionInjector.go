package injector

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type BorrowedExtensionInjector struct {
	ExtensionInjector
}

func NewBorrowedExtensionInjector(options ExtensionInjectorConfig) BorrowedExtensionInjector {
	return BorrowedExtensionInjector{ExtensionInjector: NewExtensionInjector(options)}
}

func (i *BorrowedExtensionInjector) Inject() (*ExtensionInjectionResult, error) {
	deadline := time.Now().Add(time.Duration(i.Options.InjectorServiceWorkerReadyTimeoutMS) * time.Millisecond)
	for {
		borrowed, err := i.borrowVisibleServiceWorkers()
		if err != nil || borrowed != nil {
			return borrowed, err
		}
		if time.Now().After(deadline) {
			return nil, nil
		}
		time.Sleep(time.Duration(i.Options.InjectorServiceWorkerPollIntervalMS) * time.Millisecond)
	}
}

func (i *BorrowedExtensionInjector) borrowVisibleServiceWorkers() (*ExtensionInjectionResult, error) {
	targets, err := i.targetInfos()
	if err != nil {
		return nil, err
	}
	hasConfiguredMatcher := i.Options.InjectorExtensionID != "" || len(i.Options.InjectorServiceWorkerURLIncludes) > 0 || len(i.Options.InjectorServiceWorkerURLSuffixes) > 0
	candidates := []map[string]any{}
	for _, target := range targets {
		targetType, _ := target["type"].(string)
		targetURL, _ := target["url"].(string)
		if targetType != "service_worker" || !strings.HasPrefix(targetURL, "chrome-extension://") {
			continue
		}
		if hasConfiguredMatcher && !i.serviceWorkerTargetMatches(target) {
			continue
		}
		candidates = append(candidates, target)
	}
	var borrowed []*ExtensionInjectionResult
	for _, target := range candidates {
		bootstrapped, err := i.bootstrapTarget(target)
		if err == nil && bootstrapped != nil {
			bootstrapped.Source = "borrowed"
			borrowed = append(borrowed, bootstrapped)
		}
	}
	sort.SliceStable(borrowed, func(left, right int) bool {
		if borrowed[left].HasDebugger != borrowed[right].HasDebugger {
			return borrowed[left].HasDebugger
		}
		return borrowed[left].HasTabs && !borrowed[right].HasTabs
	})
	if len(borrowed) == 0 {
		return nil, nil
	}
	return borrowed[0], nil
}

func (i *BorrowedExtensionInjector) bootstrapTarget(target map[string]any) (*ExtensionInjectionResult, error) {
	targetID, _ := target["targetId"].(string)
	targetURL, _ := target["url"].(string)
	sessionID := i.ensureSessionIDForTarget(targetID, i.Options.InjectorServiceWorkerProbeTimeoutMS, true)
	if sessionID == "" {
		return nil, nil
	}
	_, _ = i.sendWithTimeout("Runtime.enable", map[string]any{}, sessionID, i.Options.InjectorCDPSendTimeoutMS)
	bootstrap, err := modcdpServerBootstrapExpressionFromPath(i.Options.InjectorExtensionPath)
	if err != nil {
		return nil, err
	}
	probe, err := i.sendWithTimeout("Runtime.evaluate", map[string]any{
		"expression":    fmt.Sprintf("(%s)()", bootstrap),
		"awaitPromise":  true,
		"returnByValue": true,
	}, sessionID, i.Options.InjectorCDPSendTimeoutMS)
	if err != nil {
		return nil, err
	}
	result, _ := probe["result"].(map[string]any)
	value, _ := result["value"].(map[string]any)
	if hasTabs, _ := value["has_tabs"].(bool); !hasTabs {
		return nil, nil
	}
	if hasDebugger, _ := value["has_debugger"].(bool); !hasDebugger {
		return nil, nil
	}
	ready, _ := value["ok"].(bool)
	if ready && i.readyExpression() != modcdpReadyExpression {
		readyProbe, err := i.sendWithTimeout("Runtime.evaluate", map[string]any{
			"expression":    i.readyExpression(),
			"returnByValue": true,
		}, sessionID, i.Options.InjectorCDPSendTimeoutMS)
		if err != nil {
			return nil, err
		}
		readyResult, _ := readyProbe["result"].(map[string]any)
		ready, _ = readyResult["value"].(bool)
	}
	if !ready {
		return nil, nil
	}
	extensionID, _ := value["extension_id"].(string)
	if extensionID == "" {
		if match := extIDFromURL.FindStringSubmatch(targetURL); len(match) > 1 {
			extensionID = match[1]
		}
	}
	hasTabs, _ := value["has_tabs"].(bool)
	hasDebugger, _ := value["has_debugger"].(bool)
	return &ExtensionInjectionResult{
		Source:      "borrowed",
		ExtensionID: extensionID,
		TargetID:    targetID,
		URL:         targetURL,
		SessionID:   sessionID,
		HasTabs:     hasTabs,
		HasDebugger: hasDebugger,
	}, nil
}

func modcdpServerBootstrapExpressionFromPath(extensionPath string) (string, error) {
	serverPath, err := modcdpServerPathFromExtensionPath(extensionPath)
	if err != nil {
		return "", err
	}
	body, err := os.ReadFile(serverPath)
	if err != nil {
		return "", err
	}
	source := string(body)
	start := strings.Index(source, "export function installModCDPServer")
	end := strings.Index(source, "export const ModCDPServer")
	if start < 0 || end < start {
		return "", fmt.Errorf("could not find installModCDPServer in ModCDPServer.js")
	}
	installer := strings.Replace(source[start:end], "export function", "function", 1)
	return fmt.Sprintf(`function() {
const __name = (fn) => fn;
%s
const ModCDP = installModCDPServer(globalThis);
return {
  ok: Boolean(ModCDP?.__ModCDPServerVersion === 1 && ModCDP?.handleCommand && ModCDP?.addCustomEvent),
  extension_id: globalThis.chrome?.runtime?.id ?? null,
  has_tabs: Boolean(globalThis.chrome?.tabs?.query),
  has_debugger: Boolean(globalThis.chrome?.debugger?.sendCommand && globalThis.chrome?.debugger?.getTargets),
};
}`, installer), nil
}

func modcdpServerPathFromExtensionPath(extensionPath string) (string, error) {
	candidates := []string{}
	if extensionPath != "" {
		candidates = append(candidates, filepath.Join(extensionPath, "ModCDPServer.js"))
		candidates = append(candidates, filepath.Join(extensionPath, "js", "src", "server", "ModCDPServer.js"))
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		for dir := filepath.Dir(file); ; dir = filepath.Dir(dir) {
			candidates = append(candidates, filepath.Join(dir, "dist", "extension", "js", "src", "server", "ModCDPServer.js"))
			candidates = append(candidates, filepath.Join(dir, "dist", "extension", "ModCDPServer.js"))
			if parent := filepath.Dir(dir); parent == dir {
				break
			}
		}
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unable to locate ModCDPServer.js; checked: %s", strings.Join(candidates, ", "))
}
