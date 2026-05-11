package injector

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pirate/ModCDP/go/modcdp/types"
)

const DefaultModCDPExtensionID = "mdedooklbnfejodmnhmkdpkaedafkehf"
const DefaultModCDPWakePath = "/modcdp/wake.html"
const DefaultCDPSendTimeoutMS = 10_000
const DefaultExecutionContextTimeoutMS = 10_000
const DefaultServiceWorkerProbeTimeoutMS = 10_000
const DefaultServiceWorkerReadyTimeoutMS = 60_000
const DefaultServiceWorkerPollIntervalMS = 100
const DefaultTargetSessionPollIntervalMS = 20

var DefaultModCDPServiceWorkerURLSuffixes = []string{"/modcdp/service_worker.js"}
var extIDFromURL = regexp.MustCompile(`^chrome-extension://([a-z]+)/`)

const modcdpReadyExpression = `Boolean(globalThis.ModCDP?.__ModCDPServerVersion === 1 && globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)`

type SendCDP = types.SendCDP
type SessionIDForTarget = types.SessionIDForTarget
type AttachToTarget = types.AttachToTarget
type WaitForExecutionContext = types.WaitForExecutionContext
type LaunchOptions = types.LaunchOptions
type ExtensionInjectorConfig = types.ExtensionInjectorConfig
type ExtensionInjectionResult = types.ExtensionInjectionResult

func boolPtr(value bool) *bool {
	return &value
}

type ExtensionInjector struct {
	Options           ExtensionInjectorConfig
	UnusableTargetIDs map[string]bool
	LastError         error
}

func NewExtensionInjector(options ExtensionInjectorConfig) ExtensionInjector {
	if options.WakePath == "" {
		options.WakePath = DefaultModCDPWakePath
	}
	if options.CDPSendTimeoutMS == 0 {
		options.CDPSendTimeoutMS = DefaultCDPSendTimeoutMS
	}
	if options.ExecutionContextTimeoutMS == 0 {
		options.ExecutionContextTimeoutMS = DefaultExecutionContextTimeoutMS
	}
	if options.ServiceWorkerProbeTimeoutMS == 0 {
		options.ServiceWorkerProbeTimeoutMS = DefaultServiceWorkerProbeTimeoutMS
	}
	if options.ServiceWorkerReadyTimeoutMS == 0 {
		options.ServiceWorkerReadyTimeoutMS = DefaultServiceWorkerReadyTimeoutMS
	}
	if options.ServiceWorkerPollIntervalMS == 0 {
		options.ServiceWorkerPollIntervalMS = DefaultServiceWorkerPollIntervalMS
	}
	if options.TargetSessionPollIntervalMS == 0 {
		options.TargetSessionPollIntervalMS = DefaultTargetSessionPollIntervalMS
	}
	return ExtensionInjector{Options: options, UnusableTargetIDs: map[string]bool{}}
}

func (i *ExtensionInjector) Update(config ExtensionInjectorConfig) *ExtensionInjector {
	if config.Send != nil {
		i.Options.Send = config.Send
	}
	if config.SessionIDForTarget != nil {
		i.Options.SessionIDForTarget = config.SessionIDForTarget
	}
	if config.AttachToTarget != nil {
		i.Options.AttachToTarget = config.AttachToTarget
	}
	if config.WaitForExecutionContext != nil {
		i.Options.WaitForExecutionContext = config.WaitForExecutionContext
	}
	if config.ExtensionPath != "" {
		i.Options.ExtensionPath = config.ExtensionPath
	}
	if config.ExtensionID != "" {
		i.Options.ExtensionID = config.ExtensionID
	}
	if config.WakePath != "" {
		i.Options.WakePath = config.WakePath
	}
	if config.WakeURL != "" {
		i.Options.WakeURL = config.WakeURL
	}
	if config.ServiceWorkerURLIncludes != nil {
		i.Options.ServiceWorkerURLIncludes = append([]string{}, config.ServiceWorkerURLIncludes...)
	}
	if config.ServiceWorkerURLSuffixes != nil {
		i.Options.ServiceWorkerURLSuffixes = append([]string{}, config.ServiceWorkerURLSuffixes...)
	}
	if config.TrustServiceWorkerTarget {
		i.Options.TrustServiceWorkerTarget = true
	}
	if config.RequireServiceWorkerTarget {
		i.Options.RequireServiceWorkerTarget = true
	}
	if config.ServiceWorkerReadyExpression != "" {
		i.Options.ServiceWorkerReadyExpression = config.ServiceWorkerReadyExpression
	}
	if config.CDPSendTimeoutMS != 0 {
		i.Options.CDPSendTimeoutMS = config.CDPSendTimeoutMS
	}
	if config.ExecutionContextTimeoutMS != 0 {
		i.Options.ExecutionContextTimeoutMS = config.ExecutionContextTimeoutMS
	}
	if config.ServiceWorkerProbeTimeoutMS != 0 {
		i.Options.ServiceWorkerProbeTimeoutMS = config.ServiceWorkerProbeTimeoutMS
	}
	if config.ServiceWorkerReadyTimeoutMS != 0 {
		i.Options.ServiceWorkerReadyTimeoutMS = config.ServiceWorkerReadyTimeoutMS
	}
	if config.ServiceWorkerPollIntervalMS != 0 {
		i.Options.ServiceWorkerPollIntervalMS = config.ServiceWorkerPollIntervalMS
	}
	if config.TargetSessionPollIntervalMS != 0 {
		i.Options.TargetSessionPollIntervalMS = config.TargetSessionPollIntervalMS
	}
	if config.BrowserbaseAPIKey != "" {
		i.Options.BrowserbaseAPIKey = config.BrowserbaseAPIKey
	}
	if config.BrowserbaseBaseURL != "" {
		i.Options.BrowserbaseBaseURL = config.BrowserbaseBaseURL
	}
	if config.UpstreamReverseWSURL != "" {
		i.Options.UpstreamReverseWSURL = config.UpstreamReverseWSURL
	}
	if config.UpstreamNativeMessagingHostName != "" {
		i.Options.UpstreamNativeMessagingHostName = config.UpstreamNativeMessagingHostName
	}
	if config.UpstreamNATSURL != "" {
		i.Options.UpstreamNATSURL = config.UpstreamNATSURL
	}
	if config.UpstreamNATSSubjectPrefix != "" {
		i.Options.UpstreamNATSSubjectPrefix = config.UpstreamNATSSubjectPrefix
	}
	return i
}

func (i ExtensionInjector) GetInjectorConfig() ExtensionInjectorConfig {
	return i.Options
}

func (i ExtensionInjector) GetLauncherConfig() LaunchOptions {
	return LaunchOptions{}
}

func (i ExtensionInjector) GetTransportConfig() map[string]any {
	if i.Options.ExtensionID == "" {
		return map[string]any{}
	}
	return map[string]any{"extension_id": i.Options.ExtensionID}
}

func (i *ExtensionInjector) Prepare() error {
	return nil
}

func (i *ExtensionInjector) Close() error {
	return nil
}

func (i *ExtensionInjector) Inject() (*ExtensionInjectionResult, error) {
	return nil, fmt.Errorf("%T.Inject is not implemented", i)
}

func (i ExtensionInjector) readyExpression() string {
	if i.Options.ServiceWorkerReadyExpression == "" {
		return modcdpReadyExpression
	}
	return fmt.Sprintf("(%s) && Boolean(%s)", modcdpReadyExpression, i.Options.ServiceWorkerReadyExpression)
}

func (i ExtensionInjector) sendWithTimeout(method string, params map[string]any, sessionID string, timeoutMS int) (map[string]any, error) {
	if i.Options.Send == nil {
		return nil, fmt.Errorf("%T requires a CDP send function", i)
	}
	if params == nil {
		params = map[string]any{}
	}
	if timeoutMS == 0 {
		timeoutMS = i.Options.CDPSendTimeoutMS
	}
	if timeoutMS <= 0 {
		return i.Options.Send(method, params, sessionID)
	}
	type sendResult struct {
		result map[string]any
		err    error
	}
	done := make(chan sendResult, 1)
	go func() {
		result, err := i.Options.Send(method, params, sessionID)
		done <- sendResult{result: result, err: err}
	}()
	select {
	case result := <-done:
		return result.result, result.err
	case <-time.After(time.Duration(timeoutMS) * time.Millisecond):
		return nil, fmt.Errorf("%s timed out after %dms", method, timeoutMS)
	}
}

func (i ExtensionInjector) SendWithTimeout(method string, params map[string]any, sessionID string, timeoutMS int) (map[string]any, error) {
	return i.sendWithTimeout(method, params, sessionID, timeoutMS)
}

func (i ExtensionInjector) sessionIDForTarget(targetID string, timeoutMS int) string {
	deadline := time.Now().Add(time.Duration(timeoutMS) * time.Millisecond)
	for {
		if i.Options.SessionIDForTarget != nil {
			if sessionID := i.Options.SessionIDForTarget(targetID); sessionID != "" {
				return sessionID
			}
		}
		if timeoutMS <= 0 || time.Now().After(deadline) {
			return ""
		}
		time.Sleep(time.Duration(i.Options.TargetSessionPollIntervalMS) * time.Millisecond)
	}
}

func (i ExtensionInjector) ensureSessionIDForTarget(targetID string, timeoutMS int, allowAttach bool) string {
	if i.Options.SessionIDForTarget != nil {
		if sessionID := i.Options.SessionIDForTarget(targetID); sessionID != "" {
			return sessionID
		}
	}
	if allowAttach && i.Options.AttachToTarget != nil {
		if sessionID := i.Options.AttachToTarget(targetID); sessionID != "" {
			return sessionID
		}
	}
	return i.sessionIDForTarget(targetID, timeoutMS)
}

func (i ExtensionInjector) targetInfos() ([]map[string]any, error) {
	result, err := i.sendWithTimeout("Target.getTargets", map[string]any{}, "", i.Options.CDPSendTimeoutMS)
	if err != nil {
		return nil, err
	}
	rawTargets, _ := result["targetInfos"].([]any)
	targets := make([]map[string]any, 0, len(rawTargets))
	for _, rawTarget := range rawTargets {
		target, _ := rawTarget.(map[string]any)
		if target != nil {
			targets = append(targets, target)
		}
	}
	return targets, nil
}

func (i ExtensionInjector) configuredWakeURL() string {
	if i.Options.WakeURL != "" {
		return i.Options.WakeURL
	}
	if i.Options.ExtensionID == "" {
		return ""
	}
	wakePath := i.Options.WakePath
	if wakePath == "" {
		wakePath = DefaultModCDPWakePath
	}
	if !strings.HasPrefix(wakePath, "/") {
		wakePath = "/" + wakePath
	}
	return "chrome-extension://" + i.Options.ExtensionID + wakePath
}

func (i ExtensionInjector) wakeConfiguredExtension() bool {
	wakeURL := i.configuredWakeURL()
	if wakeURL == "" || i.Options.Send == nil {
		return false
	}
	_, err := i.sendWithTimeout("Target.createTarget", map[string]any{
		"url":        wakeURL,
		"background": true,
		"hidden":     true,
		"focus":      false,
	}, "", i.Options.CDPSendTimeoutMS)
	return err == nil
}

func (i ExtensionInjector) WakeConfiguredExtension() bool {
	return i.wakeConfiguredExtension()
}

func (i ExtensionInjector) probeTarget(target map[string]any, sessionTimeoutMS int, allowAttach bool) (*ExtensionInjectionResult, error) {
	targetID, _ := target["targetId"].(string)
	targetURL, _ := target["url"].(string)
	if targetID == "" || i.UnusableTargetIDs[targetID] {
		return nil, nil
	}
	sessionID := i.ensureSessionIDForTarget(targetID, sessionTimeoutMS, allowAttach)
	if sessionID == "" {
		return nil, nil
	}
	if _, err := i.sendWithTimeout("Runtime.enable", map[string]any{}, sessionID, i.Options.CDPSendTimeoutMS); err != nil {
		return nil, err
	}
	probe, err := i.sendWithTimeout("Runtime.evaluate", map[string]any{
		"expression":    i.readyExpression(),
		"returnByValue": true,
	}, sessionID, i.Options.CDPSendTimeoutMS)
	if err != nil {
		return nil, err
	}
	result, _ := probe["result"].(map[string]any)
	if ready, _ := result["value"].(bool); !ready {
		return nil, nil
	}
	extensionID := ""
	if m := extIDFromURL.FindStringSubmatch(targetURL); len(m) > 1 {
		extensionID = m[1]
	}
	return &ExtensionInjectionResult{
		Source:      "discovered",
		ExtensionID: extensionID,
		TargetID:    targetID,
		URL:         targetURL,
		SessionID:   sessionID,
	}, nil
}

func (i ExtensionInjector) discoverReadyServiceWorker(matchedOnly bool) (*ExtensionInjectionResult, error) {
	targets, err := i.targetInfos()
	if err != nil {
		return nil, err
	}
	if i.Options.TrustServiceWorkerTarget {
		for _, target := range targets {
			if !i.serviceWorkerTargetMatches(target) {
				continue
			}
			probed, err := i.probeTarget(target, i.Options.ServiceWorkerProbeTimeoutMS, true)
			if err != nil {
				return nil, err
			}
			if probed != nil {
				probed.Source = "trusted"
				return probed, nil
			}
		}
	}
	if i.Options.TrustServiceWorkerTarget || matchedOnly {
		return nil, nil
	}
	for _, target := range targets {
		targetType, _ := target["type"].(string)
		targetURL, _ := target["url"].(string)
		if targetType != "service_worker" || !strings.HasPrefix(targetURL, "chrome-extension://") {
			continue
		}
		probed, err := i.probeTarget(target, i.Options.ServiceWorkerProbeTimeoutMS, false)
		if err == nil && probed != nil {
			return probed, nil
		}
	}
	return nil, nil
}

func (i ExtensionInjector) waitForReadyServiceWorker(timeoutMS int, matchedOnly bool) (*ExtensionInjectionResult, error) {
	deadline := time.Now().Add(time.Duration(timeoutMS) * time.Millisecond)
	for time.Now().Before(deadline) {
		discovered, err := i.discoverReadyServiceWorker(matchedOnly)
		if err != nil || discovered != nil {
			return discovered, err
		}
		time.Sleep(time.Duration(i.Options.ServiceWorkerPollIntervalMS) * time.Millisecond)
	}
	return nil, nil
}

func (i ExtensionInjector) WaitForReadyServiceWorker(timeoutMS int, matchedOnly bool) (*ExtensionInjectionResult, error) {
	return i.waitForReadyServiceWorker(timeoutMS, matchedOnly)
}

func (i ExtensionInjector) serviceWorkerTargetMatches(target map[string]any) bool {
	targetURL, _ := target["url"].(string)
	targetType, _ := target["type"].(string)
	if targetType != "service_worker" || !strings.HasPrefix(targetURL, "chrome-extension://") {
		return false
	}
	hasExtensionID := i.Options.ExtensionID != ""
	if i.Options.ExtensionID != "" && !strings.HasPrefix(targetURL, "chrome-extension://"+i.Options.ExtensionID+"/") {
		return false
	}
	for _, part := range i.Options.ServiceWorkerURLIncludes {
		if !strings.Contains(targetURL, part) {
			return false
		}
	}
	if len(i.Options.ServiceWorkerURLSuffixes) > 0 {
		matched := false
		for _, suffix := range i.Options.ServiceWorkerURLSuffixes {
			if strings.HasSuffix(targetURL, suffix) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	return hasExtensionID || len(i.Options.ServiceWorkerURLIncludes) > 0 || len(i.Options.ServiceWorkerURLSuffixes) > 0
}

func (i ExtensionInjector) ServiceWorkerTargetMatches(target map[string]any) bool {
	return i.serviceWorkerTargetMatches(target)
}
