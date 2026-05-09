package modcdp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const DefaultModCDPExtensionID = "mdedooklbnfejodmnhmkdpkaedafkehf"
const DefaultModCDPWakePath = "/modcdp/wake.html"

var DefaultModCDPServiceWorkerURLSuffixes = []string{"/modcdp/service_worker.js"}

type SendCDP func(method string, params map[string]any, sessionID string) (map[string]any, error)
type SessionIDForTarget func(targetID string) string
type AttachToTarget func(targetID string) string
type WaitForExecutionContext func(sessionID string, timeoutMS int) int

type ExtensionInjectorConfig struct {
	Send                         SendCDP
	SessionIDForTarget           SessionIDForTarget
	AttachToTarget               AttachToTarget
	WaitForExecutionContext      WaitForExecutionContext
	ExtensionPath                string
	ExtensionID                  string
	WakePath                     string
	WakeURL                      string
	ServiceWorkerURLIncludes     []string
	ServiceWorkerURLSuffixes     []string
	TrustMatchedServiceWorker    bool
	RequireServiceWorkerTarget   bool
	ServiceWorkerReadyExpression string
	CDPSendTimeoutMS             int
	ExecutionContextTimeoutMS    int
	ServiceWorkerProbeTimeoutMS  int
	ServiceWorkerReadyTimeoutMS  int
	ServiceWorkerPollIntervalMS  int
	TargetSessionPollIntervalMS  int
	BrowserbaseAPIKey            string
	BaseURL                      string
	BrowserbaseBaseURL           string
	ReverseProxyURL              string
	NativeHostName               string
	NATSURL                      string
	NATSSubjectPrefix            string
}

type ExtensionInjectionResult struct {
	Source      string
	ExtensionID string
	TargetID    string
	URL         string
	SessionID   string
	HasTabs     bool
	HasDebugger bool
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
	if config.TrustMatchedServiceWorker {
		i.Options.TrustMatchedServiceWorker = true
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
	if config.BaseURL != "" {
		i.Options.BaseURL = config.BaseURL
	}
	if config.BrowserbaseBaseURL != "" {
		i.Options.BrowserbaseBaseURL = config.BrowserbaseBaseURL
	}
	if config.ReverseProxyURL != "" {
		i.Options.ReverseProxyURL = config.ReverseProxyURL
	}
	if config.NativeHostName != "" {
		i.Options.NativeHostName = config.NativeHostName
	}
	if config.NATSURL != "" {
		i.Options.NATSURL = config.NATSURL
	}
	if config.NATSSubjectPrefix != "" {
		i.Options.NATSSubjectPrefix = config.NATSSubjectPrefix
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

func (i ExtensionInjector) ExtensionRuntimeConfig() map[string]string {
	config := map[string]string{}
	if i.Options.ReverseProxyURL != "" {
		config["reverse_proxy_url"] = i.Options.ReverseProxyURL
	}
	if i.Options.NativeHostName != "" {
		config["native_host_name"] = i.Options.NativeHostName
	}
	if i.Options.NATSURL != "" {
		config["nats_url"] = i.Options.NATSURL
	}
	if i.Options.NATSSubjectPrefix != "" {
		config["nats_subject_prefix"] = i.Options.NATSSubjectPrefix
	}
	return config
}

func (i ExtensionInjector) WriteExtensionRuntimeConfig(unpackedExtensionPath string) error {
	config := i.ExtensionRuntimeConfig()
	if len(config) == 0 {
		return nil
	}
	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(unpackedExtensionPath, "modcdp"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(unpackedExtensionPath, "modcdp", "config.json"), append(configBytes, '\n'), 0o644); err != nil {
		return err
	}
	configJS := "globalThis.__MODCDP_RUNTIME_CONFIG__ = " + string(configBytes) + ";\nexport {};\n"
	return os.WriteFile(filepath.Join(unpackedExtensionPath, "config.js"), []byte(configJS), 0o644)
}

func (i ExtensionInjector) ReadyExpression() string {
	if i.Options.ServiceWorkerReadyExpression == "" {
		return modcdpReadyExpression
	}
	return fmt.Sprintf("(%s) && Boolean(%s)", modcdpReadyExpression, i.Options.ServiceWorkerReadyExpression)
}

func (i ExtensionInjector) ServiceWorkerTargetMatches(target map[string]any) bool {
	targetURL, _ := target["url"].(string)
	targetType, _ := target["type"].(string)
	if targetType != "service_worker" || !strings.HasPrefix(targetURL, "chrome-extension://") {
		return false
	}
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
	return len(i.Options.ServiceWorkerURLIncludes) > 0 || len(i.Options.ServiceWorkerURLSuffixes) > 0
}
