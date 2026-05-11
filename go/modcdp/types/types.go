package types

type LaunchOptions struct {
	ExecutablePath                 string         `json:"executable_path,omitempty"`
	ExtraArgs                      []string       `json:"extra_args,omitempty"`
	Args                           []string       `json:"args,omitempty"`
	Headless                       *bool          `json:"headless,omitempty"`
	Port                           int            `json:"port,omitempty"`
	RemoteDebugging                string         `json:"remote_debugging,omitempty"`
	Sandbox                        *bool          `json:"sandbox,omitempty"`
	UserDataDir                    string         `json:"user_data_dir,omitempty"`
	CleanupUserDataDir             *bool          `json:"cleanup_user_data_dir,omitempty"`
	ChromeReadyTimeoutMS           int            `json:"chrome_ready_timeout_ms,omitempty"`
	ChromeReadyPollIntervalMS      int            `json:"chrome_ready_poll_interval_ms,omitempty"`
	CDPURL                         string         `json:"cdp_url,omitempty"`
	BrowserbaseAPIKey              string         `json:"browserbase_api_key,omitempty"`
	BrowserbaseBaseURL             string         `json:"browserbase_base_url,omitempty"`
	BrowserbaseSessionID           string         `json:"browserbase_session_id,omitempty"`
	BrowserbaseKeepAlive           *bool          `json:"browserbase_keep_alive,omitempty"`
	BrowserbaseCloseSessionOnClose *bool          `json:"browserbase_close_session_on_close,omitempty"`
	Region                         string         `json:"region,omitempty"`
	Timeout                        int            `json:"timeout,omitempty"`
	ExtensionID                    string         `json:"extension_id,omitempty"`
	BrowserbaseBrowserSettings     map[string]any `json:"browserbase_browser_settings,omitempty"`
	BrowserbaseUserMetadata        map[string]any `json:"browserbase_user_metadata,omitempty"`
	BrowserbaseSessionCreateParams map[string]any `json:"browserbase_session_create_params,omitempty"`
}

type SendCDP func(method string, params map[string]any, sessionID string) (map[string]any, error)
type SessionIDForTarget func(targetID string) string
type AttachToTarget func(targetID string) string
type WaitForExecutionContext func(sessionID string, timeoutMS int) int

type ExtensionInjectorConfig struct {
	Send                            SendCDP                 `json:"-"`
	SessionIDForTarget              SessionIDForTarget      `json:"-"`
	AttachToTarget                  AttachToTarget          `json:"-"`
	WaitForExecutionContext         WaitForExecutionContext `json:"-"`
	ExtensionPath                   string                  `json:"extension_path,omitempty"`
	ExtensionID                     string                  `json:"extension_id,omitempty"`
	WakePath                        string                  `json:"wake_path,omitempty"`
	WakeURL                         string                  `json:"wake_url,omitempty"`
	ServiceWorkerURLIncludes        []string                `json:"service_worker_url_includes,omitempty"`
	ServiceWorkerURLSuffixes        []string                `json:"service_worker_url_suffixes,omitempty"`
	TrustServiceWorkerTarget        bool                    `json:"trust_service_worker_target,omitempty"`
	RequireServiceWorkerTarget      bool                    `json:"require_service_worker_target,omitempty"`
	ServiceWorkerReadyExpression    string                  `json:"service_worker_ready_expression,omitempty"`
	CDPSendTimeoutMS                int                     `json:"cdp_send_timeout_ms,omitempty"`
	ExecutionContextTimeoutMS       int                     `json:"execution_context_timeout_ms,omitempty"`
	ServiceWorkerProbeTimeoutMS     int                     `json:"service_worker_probe_timeout_ms,omitempty"`
	ServiceWorkerReadyTimeoutMS     int                     `json:"service_worker_ready_timeout_ms,omitempty"`
	ServiceWorkerPollIntervalMS     int                     `json:"service_worker_poll_interval_ms,omitempty"`
	TargetSessionPollIntervalMS     int                     `json:"target_session_poll_interval_ms,omitempty"`
	BrowserbaseAPIKey               string                  `json:"browserbase_api_key,omitempty"`
	BrowserbaseBaseURL              string                  `json:"browserbase_base_url,omitempty"`
	UpstreamReverseWSURL            string                  `json:"upstream_reversews_url,omitempty"`
	UpstreamNativeMessagingHostName string                  `json:"upstream_nativemessaging_host_name,omitempty"`
	UpstreamNATSURL                 string                  `json:"upstream_nats_url,omitempty"`
	UpstreamNATSSubjectPrefix       string                  `json:"upstream_nats_subject_prefix,omitempty"`
}

type ExtensionInjectionResult struct {
	Source      string `json:"source"`
	ExtensionID string `json:"extension_id,omitempty"`
	TargetID    string `json:"target_id"`
	URL         string `json:"url,omitempty"`
	SessionID   string `json:"session_id"`
	HasTabs     bool   `json:"has_tabs,omitempty"`
	HasDebugger bool   `json:"has_debugger,omitempty"`
}
