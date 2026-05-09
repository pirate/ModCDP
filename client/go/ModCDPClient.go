// ModCDPClient (Go): importable, no CLI, no demo code.
//
// Option groups mirror the JS / Python ports:
//
//	Launch         browser/session creation and cleanup.
//	Upstream       message transport to raw CDP or a ModCDP server.
//	Extension      raw-CDP extension discovery/injection/borrowing.
//	Client.Routes  client-side direct_cdp/service_worker routing.
//	Server         ModCDPServer.configure params.
//	Client        client routing and client-owned send/event timings.
//	Extension     extension discovery, wake, probe, and keepalive timings.
//	Upstream      upstream transport options and upstream-owned timings.
//
// Public methods: Connect, Send(method, params), SendRaw, On, OnRaw, Close.
// Synchronous; one background goroutine reads messages off the WS.
//
// Route and ModCDP wire translation lives in translate.go. Launchers and
// upstream transports live in their matching class files.
package modcdp

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	abxjsonschema "github.com/ArchiveBox/abxbus/abxbus-go/jsonschema"
)

var (
	extIDFromURL = regexp.MustCompile(`^chrome-extension://([a-z]+)/`)
)

const modcdpReadyExpression = `Boolean(globalThis.ModCDP?.__ModCDPServerVersion === 1 && globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)`

const DefaultCDPSendTimeoutMS = 10_000
const DefaultEventWaitTimeoutMS = 10_000
const DefaultExecutionContextTimeoutMS = 10_000
const DefaultChromeReadyTimeoutMS = 45_000
const DefaultChromeReadyPollIntervalMS = 100
const DefaultServiceWorkerProbeTimeoutMS = 10_000
const DefaultServiceWorkerReadyTimeoutMS = 60_000
const DefaultServiceWorkerPollIntervalMS = 100
const DefaultTargetSessionPollIntervalMS = 20
const DefaultWSConnectErrorSettleTimeoutMS = 250

//go:embed extension.zip
var bundledExtensionZip []byte

func websocketURLFor(endpoint string) (string, error) {
	if strings.HasPrefix(endpoint, "ws://") || strings.HasPrefix(endpoint, "wss://") {
		return endpoint, nil
	}
	httpEndpoint := endpoint
	if !strings.Contains(endpoint, "://") {
		httpEndpoint = "http://" + endpoint
	}
	resp, err := http.Get(httpEndpoint + "/json/version")
	if err != nil {
		return "", fmt.Errorf("GET /json/version: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		parsed, parseErr := url.Parse(httpEndpoint)
		if parseErr != nil {
			return "", parseErr
		}
		if parsed.Scheme == "https" {
			parsed.Scheme = "wss"
		} else {
			parsed.Scheme = "ws"
		}
		parsed.Path = "/devtools/browser"
		parsed.RawQuery = ""
		parsed.Fragment = ""
		return parsed.String(), nil
	}
	body, _ := io.ReadAll(resp.Body)
	var version map[string]any
	if err := json.Unmarshal(body, &version); err != nil {
		return "", fmt.Errorf("parse /json/version: %w", err)
	}
	wsURL, _ := version["webSocketDebuggerUrl"].(string)
	if wsURL == "" {
		return "", fmt.Errorf("HTTP discovery for %s returned no webSocketDebuggerUrl", endpoint)
	}
	return wsURL, nil
}

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

// --- public types --------------------------------------------------------

type ServerConfig struct {
	LoopbackCDPURL                    string            `json:"loopback_cdp_url,omitempty"`
	Routes                            map[string]string `json:"routes,omitempty"`
	BrowserToken                      string            `json:"browser_token,omitempty"`
	CDPSendTimeoutMS                  int               `json:"cdp_send_timeout_ms,omitempty"`
	LoopbackExecutionContextTimeoutMS int               `json:"loopback_execution_context_timeout_ms,omitempty"`
	WSConnectErrorSettleTimeoutMS     int               `json:"ws_connect_error_settle_timeout_ms,omitempty"`
	Options                           map[string]any    `json:"-"`
}

type CustomEvent struct {
	Name        string         `json:"name"`
	EventSchema map[string]any `json:"event_schema,omitempty"`
}

type CustomCommand struct {
	Name         string         `json:"name"`
	Expression   string         `json:"expression,omitempty"`
	ParamsSchema map[string]any `json:"params_schema,omitempty"`
	ResultSchema map[string]any `json:"result_schema,omitempty"`
}

type CustomMiddleware struct {
	Name       string `json:"name,omitempty"`
	Phase      string `json:"phase"`
	Expression string `json:"expression"`
}

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
	WSURL                          string         `json:"ws_url,omitempty"`
	BrowserbaseAPIKey              string         `json:"browserbase_api_key,omitempty"`
	ProjectID                      string         `json:"project_id,omitempty"`
	BrowserbaseProjectID           string         `json:"browserbase_project_id,omitempty"`
	BaseURL                        string         `json:"base_url,omitempty"`
	BrowserbaseBaseURL             string         `json:"browserbase_base_url,omitempty"`
	SessionID                      string         `json:"session_id,omitempty"`
	ResumeSessionID                string         `json:"resume_session_id,omitempty"`
	KeepAlive                      *bool          `json:"keep_alive,omitempty"`
	CloseSessionOnClose            *bool          `json:"close_session_on_close,omitempty"`
	Region                         string         `json:"region,omitempty"`
	Timeout                        int            `json:"timeout,omitempty"`
	ExtensionID                    string         `json:"extension_id,omitempty"`
	BrowserSettings                map[string]any `json:"browser_settings,omitempty"`
	UserMetadata                   map[string]any `json:"user_metadata,omitempty"`
	SessionCreateParams            map[string]any `json:"session_create_params,omitempty"`
	BrowserbaseSessionCreateParams map[string]any `json:"browserbase_session_create_params,omitempty"`
}

type LaunchConfig struct {
	Mode           string        `json:"mode,omitempty"`
	ExecutablePath string        `json:"executable_path,omitempty"`
	UserDataDir    string        `json:"user_data_dir,omitempty"`
	Options        LaunchOptions `json:"options,omitempty"`
}

type UpstreamConfig struct {
	Mode                          string `json:"mode,omitempty"`
	WSURL                         string `json:"ws_url,omitempty"`
	NATSURL                       string `json:"nats_url,omitempty"`
	NATSSubjectPrefix             string `json:"nats_subject_prefix,omitempty"`
	ReverseWSBind                 string `json:"reversews_bind,omitempty"`
	ReverseWSWaitTimeoutMS        int    `json:"reversews_wait_timeout_ms,omitempty"`
	NativeMessagingManifest       string `json:"nativemessaging_manifest,omitempty"`
	NativeMessagingHostName       string `json:"nativemessaging_host_name,omitempty"`
	WSConnectErrorSettleTimeoutMS int    `json:"ws_connect_error_settle_timeout_ms,omitempty"`
}

type ExtensionConfig struct {
	Mode                         string   `json:"mode,omitempty"`
	Path                         string   `json:"path,omitempty"`
	ExtensionID                  string   `json:"extension_id,omitempty"`
	WakePath                     string   `json:"wake_path,omitempty"`
	WakeURL                      string   `json:"wake_url,omitempty"`
	ServiceWorkerURLIncludes     []string `json:"service_worker_url_includes,omitempty"`
	ServiceWorkerURLSuffixes     []string `json:"service_worker_url_suffixes,omitempty"`
	TrustServiceWorkerTarget     bool     `json:"trust_service_worker_target,omitempty"`
	RequireServiceWorkerTarget   bool     `json:"require_service_worker_target,omitempty"`
	ServiceWorkerReadyExpression string   `json:"service_worker_ready_expression,omitempty"`
	ExecutionContextTimeoutMS    int      `json:"execution_context_timeout_ms,omitempty"`
	ServiceWorkerProbeTimeoutMS  int      `json:"service_worker_probe_timeout_ms,omitempty"`
	ServiceWorkerReadyTimeoutMS  int      `json:"service_worker_ready_timeout_ms,omitempty"`
	ServiceWorkerPollIntervalMS  int      `json:"service_worker_poll_interval_ms,omitempty"`
	TargetSessionPollIntervalMS  int      `json:"target_session_poll_interval_ms,omitempty"`
}

type ClientConfig struct {
	Routes               map[string]string `json:"routes,omitempty"`
	HydrateAliases       *bool             `json:"hydrate_aliases,omitempty"`
	MirrorUpstreamEvents *bool             `json:"mirror_upstream_events,omitempty"`
	CDPSendTimeoutMS     int               `json:"cdp_send_timeout_ms,omitempty"`
	EventWaitTimeoutMS   int               `json:"event_wait_timeout_ms,omitempty"`
}

type Options struct {
	Launch            LaunchConfig       `json:"launch,omitempty"`
	Upstream          UpstreamConfig     `json:"upstream,omitempty"`
	Extension         ExtensionConfig    `json:"extension,omitempty"`
	Client            ClientConfig       `json:"client,omitempty"`
	Server            *ServerConfig      `json:"server,omitempty"`
	CustomCommands    []CustomCommand    `json:"custom_commands,omitempty"`
	CustomEvents      []CustomEvent      `json:"custom_events,omitempty"`
	CustomMiddlewares []CustomMiddleware `json:"custom_middlewares,omitempty"`
}

type Handler func(data any)

type handlerEntry struct {
	handler Handler
	pointer uintptr
}

type CDPEvent struct {
	Method       string         `json:"method"`
	Params       map[string]any `json:"params,omitempty"`
	CDPSessionID string         `json:"cdpSessionId,omitempty"`
	SessionID    string         `json:"sessionId,omitempty"`
}

type ModDomain struct {
	client *ModCDPClient
}

type ModCDPClient struct {
	Accessibility        AccessibilityDomain
	Animation            AnimationDomain
	Audits               AuditsDomain
	Autofill             AutofillDomain
	BackgroundService    BackgroundServiceDomain
	BluetoothEmulation   BluetoothEmulationDomain
	Browser              BrowserDomain
	CSS                  CSSDomain
	CacheStorage         CacheStorageDomain
	Cast                 CastDomain
	Console              ConsoleDomain
	DOM                  DOMDomain
	DOMDebugger          DOMDebuggerDomain
	DOMSnapshot          DOMSnapshotDomain
	DOMStorage           DOMStorageDomain
	Debugger             DebuggerDomain
	DeviceAccess         DeviceAccessDomain
	DeviceOrientation    DeviceOrientationDomain
	Emulation            EmulationDomain
	EventBreakpoints     EventBreakpointsDomain
	Extensions           ExtensionsDomain
	FedCm                FedCmDomain
	Fetch                FetchDomain
	FileSystem           FileSystemDomain
	HeadlessExperimental HeadlessExperimentalDomain
	HeapProfiler         HeapProfilerDomain
	IO                   IODomain
	IndexedDB            IndexedDBDomain
	Input                InputDomain
	Inspector            InspectorDomain
	LayerTree            LayerTreeDomain
	Log                  LogDomain
	Media                MediaDomain
	Memory               MemoryDomain
	Network              NetworkDomain
	Overlay              OverlayDomain
	PWA                  PWADomain
	Page                 PageDomain
	Performance          PerformanceDomain
	PerformanceTimeline  PerformanceTimelineDomain
	Preload              PreloadDomain
	Profiler             ProfilerDomain
	Runtime              RuntimeDomain
	Schema               SchemaDomain
	Security             SecurityDomain
	ServiceWorker        ServiceWorkerDomain
	SmartCardEmulation   SmartCardEmulationDomain
	Storage              StorageDomain
	SystemInfo           SystemInfoDomain
	Target               TargetDomain
	Tethering            TetheringDomain
	Tracing              TracingDomain
	WebAudio             WebAudioDomain
	WebAuthn             WebAuthnDomain
	Mod                  ModDomain

	opts                 Options
	CDPURL               string
	transport            upstreamTransportClient
	mu                   sync.Mutex
	nextID               int64
	pending              map[int64]chan map[string]any
	handlers             map[string][]handlerEntry
	cdpHandlers          map[string][]func(CDPEvent)
	commandParamsSchemas map[string]map[string]any
	commandResultSchemas map[string]map[string]any
	event_schemas        map[string]map[string]any
	schemaMu             sync.RWMutex
	handlersMu           sync.Mutex
	autoSessions         *AutoSessionRouter
	ExtensionID          string
	ExtTargetID          string
	ExtSessionID         string
	Latency              map[string]any
	ConnectTiming        map[string]any
	LastCommandTiming    map[string]any
	LastRawTiming        map[string]any
	launchedBrowser      *LaunchedBrowser
	extensionInjectors   []extensionInjector
}

type extensionInjector interface {
	Update(ExtensionInjectorConfig) *ExtensionInjector
	GetLauncherConfig() LaunchOptions
	GetTransportConfig() map[string]any
	Prepare() error
	Inject() (*ExtensionInjectionResult, error)
	Close() error
}

type browserLauncherClient interface {
	Update(LaunchOptions) *BrowserLauncher
	GetInjectorConfig() ExtensionInjectorConfig
	GetTransportConfig() map[string]any
	Launch(LaunchOptions) (*LaunchedBrowser, error)
}

type upstreamTransportClient interface {
	Update(map[string]any)
	Connect() error
	Close() error
	Send(map[string]any) error
	GetLauncherConfig() LaunchOptions
	GetInjectorConfig() ExtensionInjectorConfig
	GetServerConfig() map[string]any
	OnRecv(func(map[string]any)) func()
	OnClose(func(error)) func()
	WaitForPeer() error
}

func New(opts Options) *ModCDPClient {
	if opts.Upstream.Mode == "" {
		opts.Upstream.Mode = "ws"
	}
	upstreamEndpointKind := "modcdp_server"
	if opts.Upstream.Mode == "ws" || opts.Upstream.Mode == "pipe" {
		upstreamEndpointKind = "raw_cdp"
	}
	if opts.Launch.Mode == "" {
		if upstreamEndpointKind == "modcdp_server" {
			opts.Launch.Mode = "none"
		} else if opts.Upstream.WSURL != "" {
			opts.Launch.Mode = "remote"
		} else {
			opts.Launch.Mode = "local"
		}
	}
	if opts.Extension.Mode == "" {
		if upstreamEndpointKind == "raw_cdp" || opts.Launch.Mode != "none" {
			opts.Extension.Mode = "auto"
		} else {
			opts.Extension.Mode = "none"
		}
	}
	if opts.Launch.ExecutablePath != "" {
		opts.Launch.Options.ExecutablePath = opts.Launch.ExecutablePath
	}
	if opts.Launch.UserDataDir != "" {
		opts.Launch.Options.UserDataDir = opts.Launch.UserDataDir
	}
	if opts.Client.Routes == nil {
		opts.Client.Routes = DefaultClientRoutes()
	} else {
		merged := DefaultClientRoutes()
		for k, v := range opts.Client.Routes {
			merged[k] = v
		}
		opts.Client.Routes = merged
	}
	if opts.Client.HydrateAliases == nil {
		value := true
		opts.Client.HydrateAliases = &value
	}
	if opts.Server == nil {
		opts.Server = &ServerConfig{}
	}
	if upstreamEndpointKind == "modcdp_server" && opts.Server.Routes == nil {
		opts.Server.Routes = map[string]string{"*.*": "chrome_debugger"}
	}
	if opts.Extension.ServiceWorkerURLSuffixes == nil {
		opts.Extension.ServiceWorkerURLSuffixes = append([]string{}, DefaultModCDPServiceWorkerURLSuffixes...)
	}
	if opts.Client.CDPSendTimeoutMS == 0 {
		opts.Client.CDPSendTimeoutMS = DefaultCDPSendTimeoutMS
	}
	if opts.Client.EventWaitTimeoutMS == 0 {
		opts.Client.EventWaitTimeoutMS = DefaultEventWaitTimeoutMS
	}
	if opts.Extension.ExecutionContextTimeoutMS == 0 {
		opts.Extension.ExecutionContextTimeoutMS = DefaultExecutionContextTimeoutMS
	}
	if opts.Extension.ServiceWorkerProbeTimeoutMS == 0 {
		opts.Extension.ServiceWorkerProbeTimeoutMS = DefaultServiceWorkerProbeTimeoutMS
	}
	if opts.Extension.ServiceWorkerReadyTimeoutMS == 0 {
		opts.Extension.ServiceWorkerReadyTimeoutMS = DefaultServiceWorkerReadyTimeoutMS
	}
	if opts.Extension.ServiceWorkerPollIntervalMS == 0 {
		opts.Extension.ServiceWorkerPollIntervalMS = DefaultServiceWorkerPollIntervalMS
	}
	if opts.Extension.TargetSessionPollIntervalMS == 0 {
		opts.Extension.TargetSessionPollIntervalMS = DefaultTargetSessionPollIntervalMS
	}
	if opts.Upstream.WSConnectErrorSettleTimeoutMS == 0 {
		opts.Upstream.WSConnectErrorSettleTimeoutMS = DefaultWSConnectErrorSettleTimeoutMS
	}
	if opts.Upstream.ReverseWSBind == "" {
		opts.Upstream.ReverseWSBind = DefaultReverseWSBind
	}
	if opts.Upstream.ReverseWSWaitTimeoutMS == 0 {
		opts.Upstream.ReverseWSWaitTimeoutMS = DefaultReverseWSWaitTimeoutMS
	}
	client := &ModCDPClient{
		opts:                 opts,
		pending:              map[int64]chan map[string]any{},
		handlers:             map[string][]handlerEntry{},
		cdpHandlers:          map[string][]func(CDPEvent){},
		commandParamsSchemas: map[string]map[string]any{},
		commandResultSchemas: map[string]map[string]any{},
		event_schemas:        map[string]map[string]any{},
	}
	client.Mod = ModDomain{client: client}
	client.autoSessions = NewAutoSessionRouter(
		func(method string, params map[string]any, sessionID string) (map[string]any, error) {
			return client.sendMessage(method, params, sessionID)
		},
		func() int { return client.opts.Extension.ExecutionContextTimeoutMS },
	)
	if *opts.Client.HydrateAliases {
		initCDPSurface(client)
	}
	client.hydrateCustomSurface()
	return client
}

func (c *ModCDPClient) Connect() error {
	connectStartedAt := time.Now().UnixMilli()
	transportStartedAt := time.Now().UnixMilli()
	if err := c.connectUpstreamTransport(); err != nil {
		return err
	}
	transportConnectedAt := time.Now().UnixMilli()
	if c.transport == nil {
		return fmt.Errorf("upstream transport did not connect")
	}
	c.transport.OnRecv(func(message map[string]any) { c.handleMessage(message) })
	c.transport.OnClose(func(err error) { c.rejectAll(err) })
	if endpointKindForUpstream(c.opts.Upstream.Mode) == UpstreamEndpointKindModCDPServer {
		if err := c.transport.WaitForPeer(); err != nil {
			c.Close()
			return err
		}
		if c.opts.Server != nil {
			if _, err := c.sendMessage("Mod.configure", c.serverConfigureParams(nil, nil, nil), ""); err != nil {
				c.Close()
				return err
			}
		}
		go func() { _ = c.measurePingLatency() }()
		connectedAt := time.Now().UnixMilli()
		c.ConnectTiming = map[string]any{
			"started_at":             connectStartedAt,
			"upstream_mode":          c.opts.Upstream.Mode,
			"upstream_endpoint_kind": endpointKindForUpstream(c.opts.Upstream.Mode),
			"transport_started_at":   transportStartedAt,
			"transport_connected_at": transportConnectedAt,
			"transport_duration_ms":  transportConnectedAt - transportStartedAt,
			"connected_at":           connectedAt,
			"duration_ms":            connectedAt - connectStartedAt,
		}
		return nil
	}
	if err := c.initializeRawCDPTransport(); err != nil {
		c.Close()
		return err
	}
	extensionStartedAt := time.Now().UnixMilli()
	ext, err := c.injectExtension(c.extensionInjectors)
	if err != nil {
		c.Close()
		return err
	}
	extensionCompletedAt := time.Now().UnixMilli()
	c.ExtensionID = ext.ExtensionID
	c.ExtTargetID = ext.TargetID
	c.ExtSessionID = ext.SessionID
	if _, err := c.sendMessage("Runtime.enable", map[string]any{}, c.ExtSessionID); err != nil {
		c.Close()
		return err
	}
	if _, err := c.sendMessage("Runtime.addBinding", map[string]any{"name": customEventBindingName}, c.ExtSessionID); err != nil {
		c.Close()
		return err
	}
	mirrorUpstreamEvents := true
	if c.opts.Client.MirrorUpstreamEvents != nil {
		mirrorUpstreamEvents = *c.opts.Client.MirrorUpstreamEvents
	}
	if mirrorUpstreamEvents {
		if _, err := c.sendMessage("Runtime.addBinding", map[string]any{"name": upstreamEventBindingName}, c.ExtSessionID); err != nil {
			c.Close()
			return err
		}
	}

	if c.opts.Server != nil {
		customCommands := make([]map[string]any, 0, len(c.opts.CustomCommands))
		for _, command := range c.opts.CustomCommands {
			if command.Expression == "" {
				continue
			}
			customCommands = append(customCommands, map[string]any{
				"name":          command.Name,
				"expression":    command.Expression,
				"params_schema": command.ParamsSchema,
				"result_schema": command.ResultSchema,
			})
		}
		customEvents := make([]map[string]any, 0, len(c.opts.CustomEvents))
		for _, event := range c.opts.CustomEvents {
			customEvents = append(customEvents, map[string]any{
				"name":         event.Name,
				"event_schema": event.EventSchema,
			})
		}
		customMiddlewares := make([]map[string]any, 0, len(c.opts.CustomMiddlewares))
		for _, middleware := range c.opts.CustomMiddlewares {
			item := map[string]any{
				"phase":      middleware.Phase,
				"expression": middleware.Expression,
			}
			if middleware.Name != "" {
				item["name"] = middleware.Name
			}
			customMiddlewares = append(customMiddlewares, item)
		}
		configureParams := c.serverConfigureParams(customCommands, customEvents, customMiddlewares)
		command, err := wrapCommandIfNeeded("Mod.configure", configureParams, c.opts.Client.Routes, c.ExtSessionID)
		if err != nil {
			c.Close()
			return fmt.Errorf("Mod.configure: %w", err)
		}
		if _, err := c.sendRaw(command); err != nil {
			c.Close()
			return fmt.Errorf("Mod.configure: %w", err)
		}
	}
	go func() { _ = c.measurePingLatency() }()
	connectedAt := time.Now().UnixMilli()
	c.ConnectTiming = map[string]any{
		"started_at":             connectStartedAt,
		"upstream_mode":          c.opts.Upstream.Mode,
		"upstream_endpoint_kind": endpointKindForUpstream(c.opts.Upstream.Mode),
		"transport_started_at":   transportStartedAt,
		"transport_connected_at": transportConnectedAt,
		"transport_duration_ms":  transportConnectedAt - transportStartedAt,
		"extension_source":       ext.Source,
		"extension_started_at":   extensionStartedAt,
		"extension_completed_at": extensionCompletedAt,
		"extension_duration_ms":  extensionCompletedAt - extensionStartedAt,
		"connected_at":           connectedAt,
		"duration_ms":            connectedAt - connectStartedAt,
	}
	return nil
}

func (c *ModCDPClient) connectUpstreamTransport() error {
	if c.transport != nil {
		return nil
	}
	if !isKnownLaunchMode(c.opts.Launch.Mode) {
		return fmt.Errorf("unknown launch.mode=%s", c.opts.Launch.Mode)
	}
	if !isKnownUpstreamMode(c.opts.Upstream.Mode) {
		return fmt.Errorf("unknown upstream.mode=%s", c.opts.Upstream.Mode)
	}
	if !isKnownExtensionMode(c.opts.Extension.Mode) {
		return fmt.Errorf("unknown extension.mode=%s", c.opts.Extension.Mode)
	}
	launcher := c.browserLauncher()
	transport := c.upstreamTransport()
	injectors := c.extensionInjectorsForConfig()
	c.extensionInjectors = injectors
	initialTransportConfig := c.upstreamTransportConfig()

	transport.Update(initialTransportConfig)
	launcher.Update(c.opts.Launch.Options)
	for _, injector := range injectors {
		injector.Update(c.baseExtensionInjectorConfig(nil))
	}
	for _, injector := range injectors {
		injector.Update(launcher.GetInjectorConfig())
	}
	for _, injector := range injectors {
		injector.Update(transport.GetInjectorConfig())
	}
	for _, injector := range injectors {
		if err := injector.Prepare(); err != nil {
			return err
		}
	}
	for _, injector := range injectors {
		launcher.Update(injector.GetLauncherConfig())
	}
	for _, injector := range injectors {
		transport.Update(injector.GetTransportConfig())
	}
	launcher.Update(transport.GetLauncherConfig())
	transport.Update(launcher.GetTransportConfig())

	if endpointKindForUpstream(c.opts.Upstream.Mode) == UpstreamEndpointKindModCDPServer {
		if err := transport.Connect(); err != nil {
			return err
		}
	}
	if c.opts.Launch.Mode != "none" {
		launched, err := launcher.Launch(LaunchOptions{})
		if err != nil {
			_ = transport.Close()
			return err
		}
		c.launchedBrowser = launched
		transport.Update(launcher.GetTransportConfig())
		for _, injector := range injectors {
			injector.Update(launcher.GetInjectorConfig())
		}
		for _, injector := range injectors {
			transport.Update(injector.GetTransportConfig())
		}
	}
	launchedCDPURL := ""
	if c.launchedBrowser != nil {
		launchedCDPURL = firstNonEmptyString(c.launchedBrowser.WSURL, c.launchedBrowser.CDPURL)
	}
	if endpointKindForUpstream(c.opts.Upstream.Mode) == UpstreamEndpointKindRawCDP {
		if err := transport.Connect(); err != nil {
			return err
		}
	}

	c.transport = transport
	transportURL := transportURL(transport)
	if endpointKindForUpstream(c.opts.Upstream.Mode) == UpstreamEndpointKindRawCDP {
		c.CDPURL = firstNonEmptyString(transportURL, launchedCDPURL)
	} else {
		c.CDPURL = launchedCDPURL
	}
	if wsTransport, ok := transport.(*WebSocketUpstreamTransport); ok && wsTransport.URL != "" {
		c.opts.Upstream.WSURL = wsTransport.URL
	}

	serverConfig := map[string]any{}
	if endpointKindForUpstream(c.opts.Upstream.Mode) == UpstreamEndpointKindModCDPServer && launchedCDPURL != "" {
		serverConfig["loopback_cdp_url"] = launchedCDPURL
	}
	for key, value := range transport.GetServerConfig() {
		serverConfig[key] = value
	}
	if c.opts.Server != nil {
		if loopbackCDPURL, _ := serverConfig["loopback_cdp_url"].(string); loopbackCDPURL != "" {
			initialWSURL, _ := initialTransportConfig["ws_url"].(string)
			if c.opts.Server.LoopbackCDPURL == "" ||
				c.opts.Server.LoopbackCDPURL == initialWSURL ||
				c.opts.Server.LoopbackCDPURL == launchedCDPURL {
				c.opts.Server.LoopbackCDPURL = loopbackCDPURL
			}
		}
	}
	return nil
}

func (c *ModCDPClient) upstreamTransportConfig() map[string]any {
	return map[string]any{
		"ws_url":                    c.opts.Upstream.WSURL,
		"cdp_url":                   c.opts.Upstream.WSURL,
		"nats_url":                  c.opts.Upstream.NATSURL,
		"nats_subject_prefix":       c.opts.Upstream.NATSSubjectPrefix,
		"reversews_bind":            c.opts.Upstream.ReverseWSBind,
		"reversews_wait_timeout_ms": c.opts.Upstream.ReverseWSWaitTimeoutMS,
		"manifest_path":             c.opts.Upstream.NativeMessagingManifest,
		"native_host_name":          c.opts.Upstream.NativeMessagingHostName,
		"extension_id":              c.opts.Extension.ExtensionID,
	}
}

func (c *ModCDPClient) initializeRawCDPTransport() error {
	if _, err := c.sendMessage("Target.setAutoAttach", map[string]any{
		"autoAttach":             true,
		"waitForDebuggerOnStart": false,
		"flatten":                true,
	}, ""); err != nil {
		return err
	}
	if _, err := c.sendMessage("Target.setDiscoverTargets", map[string]any{"discover": true}, ""); err != nil {
		return err
	}
	return nil
}

func transportURL(transport upstreamTransportClient) string {
	switch typed := transport.(type) {
	case *WebSocketUpstreamTransport:
		return typed.URL
	case *PipeUpstreamTransport:
		return typed.URL
	default:
		return ""
	}
}

func (c *ModCDPClient) serverConfigureParams(customCommands []map[string]any, customEvents []map[string]any, customMiddlewares []map[string]any) map[string]any {
	if customCommands == nil {
		customCommands = []map[string]any{}
	}
	if customEvents == nil {
		customEvents = []map[string]any{}
	}
	if customMiddlewares == nil {
		customMiddlewares = []map[string]any{}
	}
	server := map[string]any{
		"cdp_send_timeout_ms":                   c.opts.Client.CDPSendTimeoutMS,
		"loopback_execution_context_timeout_ms": c.opts.Extension.ExecutionContextTimeoutMS,
		"ws_connect_error_settle_timeout_ms":    c.opts.Upstream.WSConnectErrorSettleTimeoutMS,
	}
	if c.opts.Server != nil {
		server["loopback_cdp_url"] = c.opts.Server.LoopbackCDPURL
		server["routes"] = c.opts.Server.Routes
		if c.opts.Server.BrowserToken != "" {
			server["browser_token"] = c.opts.Server.BrowserToken
		}
		if c.opts.Server.CDPSendTimeoutMS != 0 {
			server["cdp_send_timeout_ms"] = c.opts.Server.CDPSendTimeoutMS
		}
		if c.opts.Server.LoopbackExecutionContextTimeoutMS != 0 {
			server["loopback_execution_context_timeout_ms"] = c.opts.Server.LoopbackExecutionContextTimeoutMS
		}
		if c.opts.Server.WSConnectErrorSettleTimeoutMS != 0 {
			server["ws_connect_error_settle_timeout_ms"] = c.opts.Server.WSConnectErrorSettleTimeoutMS
		}
		for key, value := range c.opts.Server.Options {
			server[key] = value
		}
	}
	upstream := map[string]any{"mode": c.opts.Upstream.Mode}
	if c.opts.Upstream.NATSURL != "" {
		upstream["nats_url"] = c.opts.Upstream.NATSURL
	}
	if c.opts.Upstream.NATSSubjectPrefix != "" {
		upstream["nats_subject_prefix"] = c.opts.Upstream.NATSSubjectPrefix
	}
	return map[string]any{
		"upstream": upstream,
		"client": map[string]any{
			"routes": c.opts.Client.Routes,
		},
		"server":             server,
		"custom_commands":    customCommands,
		"custom_events":      customEvents,
		"custom_middlewares": customMiddlewares,
	}
}

func normalizeModCDPName(name string) (string, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		return "", fmt.Errorf("name must be a non-empty string")
	}
	if strings.Count(normalized, ".") != 1 {
		return "", fmt.Errorf("name must be in Domain.method form")
	}
	parts := strings.Split(normalized, ".")
	if parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("name must be in Domain.method form")
	}
	return normalized, nil
}

func cloneSchema(schema map[string]any) map[string]any {
	if schema == nil {
		return nil
	}
	normalized, _ := abxjsonschema.Normalize(schema).(map[string]any)
	if normalized == nil {
		return nil
	}
	return normalized
}

func (c *ModCDPClient) hydrateCustomSurface() {
	c.schemaMu.Lock()
	defer c.schemaMu.Unlock()
	for _, command := range c.opts.CustomCommands {
		if command.Name == "" {
			continue
		}
		name, err := normalizeModCDPName(command.Name)
		if err != nil {
			continue
		}
		if schema := cloneSchema(command.ParamsSchema); schema != nil {
			c.commandParamsSchemas[name] = schema
		}
		if schema := cloneSchema(command.ResultSchema); schema != nil {
			c.commandResultSchemas[name] = schema
		}
	}
	for _, event := range c.opts.CustomEvents {
		if event.Name == "" {
			continue
		}
		name, err := normalizeModCDPName(event.Name)
		if err != nil {
			continue
		}
		if schema := cloneSchema(event.EventSchema); schema != nil {
			c.event_schemas[name] = schema
		}
	}
}

func (c *ModCDPClient) registerCustomCommandParams(params map[string]any) (string, bool, error) {
	rawName, _ := params["name"].(string)
	name, err := normalizeModCDPName(rawName)
	if err != nil {
		return "", false, err
	}
	c.schemaMu.Lock()
	defer c.schemaMu.Unlock()
	if rawSchema, exists := params["params_schema"]; exists {
		schemaObject, ok := rawSchema.(map[string]any)
		if !ok {
			return "", false, fmt.Errorf("params_schema must be a JSON Schema object")
		}
		if schema := cloneSchema(schemaObject); schema != nil {
			c.commandParamsSchemas[name] = schema
		}
	}
	if rawSchema, exists := params["result_schema"]; exists {
		schemaObject, ok := rawSchema.(map[string]any)
		if !ok {
			return "", false, fmt.Errorf("result_schema must be a JSON Schema object")
		}
		if schema := cloneSchema(schemaObject); schema != nil {
			c.commandResultSchemas[name] = schema
		}
	}
	expression, _ := params["expression"].(string)
	return name, expression != "", nil
}

func (c *ModCDPClient) registerCustomEventParams(params map[string]any) (string, error) {
	rawName, _ := params["name"].(string)
	name, err := normalizeModCDPName(rawName)
	if err != nil {
		return "", err
	}
	c.schemaMu.Lock()
	defer c.schemaMu.Unlock()
	if rawSchema, exists := params["event_schema"]; exists {
		schemaObject, ok := rawSchema.(map[string]any)
		if !ok {
			return "", fmt.Errorf("event_schema must be a JSON Schema object")
		}
		if schema := cloneSchema(schemaObject); schema != nil {
			c.event_schemas[name] = schema
		}
	}
	return name, nil
}

func (c *ModCDPClient) validateCommandParams(method string, params map[string]any) error {
	c.schemaMu.RLock()
	schema := c.commandParamsSchemas[method]
	c.schemaMu.RUnlock()
	if schema == nil {
		return nil
	}
	if err := abxjsonschema.Validate(schema, params); err != nil {
		return fmt.Errorf("%s params did not match params_schema: %w", method, err)
	}
	return nil
}

func (c *ModCDPClient) validateCommandResult(method string, result any) error {
	c.schemaMu.RLock()
	schema := c.commandResultSchemas[method]
	c.schemaMu.RUnlock()
	if schema == nil {
		return nil
	}
	if err := abxjsonschema.Validate(schema, result); err != nil {
		return fmt.Errorf("%s result did not match result_schema: %w", method, err)
	}
	return nil
}

func (c *ModCDPClient) validateEventData(event string, data any) (any, bool) {
	c.schemaMu.RLock()
	schema := c.event_schemas[event]
	c.schemaMu.RUnlock()
	if schema == nil {
		return data, true
	}
	if err := abxjsonschema.Validate(schema, data); err != nil {
		fmt.Fprintf(os.Stderr, "[ModCDPClient] %s event did not match event_schema: %v\n", event, err)
		return nil, false
	}
	return data, true
}

func (c *ModCDPClient) Send(method string, params map[string]any, sessionID ...string) (any, error) {
	targetSessionID := ""
	if len(sessionID) > 0 {
		targetSessionID = sessionID[0]
	}
	return c.sendCommand(method, params, targetSessionID, true)
}

func (d ModDomain) Evaluate(params map[string]any) (any, error) {
	return d.client.Send("Mod.evaluate", params)
}

func (d ModDomain) AddCustomCommand(params CustomCommand) (any, error) {
	commandParams := map[string]any{"name": params.Name}
	if params.Expression != "" {
		commandParams["expression"] = params.Expression
	}
	if params.ParamsSchema != nil {
		commandParams["params_schema"] = params.ParamsSchema
	}
	if params.ResultSchema != nil {
		commandParams["result_schema"] = params.ResultSchema
	}
	return d.client.Send("Mod.addCustomCommand", commandParams)
}

func (d ModDomain) AddCustomEvent(params CustomEvent) (any, error) {
	eventParams := map[string]any{"name": params.Name}
	if params.EventSchema != nil {
		eventParams["event_schema"] = params.EventSchema
	}
	return d.client.Send("Mod.addCustomEvent", eventParams)
}

func (d ModDomain) AddMiddleware(params CustomMiddleware) (any, error) {
	middlewareParams := map[string]any{
		"phase":      params.Phase,
		"expression": params.Expression,
	}
	if params.Name != "" {
		middlewareParams["name"] = params.Name
	}
	return d.client.Send("Mod.addMiddleware", middlewareParams)
}

func (d ModDomain) Configure(params map[string]any) (any, error) {
	return d.client.Send("Mod.configure", params)
}

func (d ModDomain) Ping(params map[string]any) (any, error) {
	return d.client.Send("Mod.ping", params)
}

func (c *ModCDPClient) sendCommand(method string, params map[string]any, targetSessionID string, validateSchema bool) (any, error) {
	startedAt := time.Now().UnixMilli()
	if params == nil {
		params = map[string]any{}
	}
	if method == "Mod.addCustomCommand" {
		name, hasExpression, err := c.registerCustomCommandParams(params)
		if err != nil {
			return nil, err
		}
		if !hasExpression {
			completedAt := time.Now().UnixMilli()
			c.LastCommandTiming = map[string]any{
				"method":       method,
				"target":       "client",
				"started_at":   startedAt,
				"completed_at": completedAt,
				"duration_ms":  completedAt - startedAt,
			}
			return map[string]any{"name": name, "registered": true}, nil
		}
	} else if method == "Mod.addCustomEvent" {
		name, err := c.registerCustomEventParams(params)
		if err != nil {
			return nil, err
		}
		if c.ExtSessionID == "" {
			completedAt := time.Now().UnixMilli()
			c.LastCommandTiming = map[string]any{
				"method":       method,
				"target":       "client",
				"started_at":   startedAt,
				"completed_at": completedAt,
				"duration_ms":  completedAt - startedAt,
			}
			return map[string]any{"name": name, "registered": true}, nil
		}
	}
	if validateSchema {
		if err := c.validateCommandParams(method, params); err != nil {
			return nil, err
		}
	}
	if endpointKindForUpstream(c.opts.Upstream.Mode) == UpstreamEndpointKindModCDPServer {
		result, err := c.sendMessage(method, params, "")
		completedAt := time.Now().UnixMilli()
		c.LastCommandTiming = map[string]any{
			"method":       method,
			"target":       "modcdp_server",
			"started_at":   startedAt,
			"completed_at": completedAt,
			"duration_ms":  completedAt - startedAt,
		}
		if err != nil {
			return nil, err
		}
		if validateSchema {
			if err := c.validateCommandResult(method, result); err != nil {
				return nil, err
			}
		}
		return result, nil
	}
	command, err := wrapCommandIfNeeded(method, params, c.opts.Client.Routes, c.ExtSessionID, targetSessionID)
	if err != nil {
		return nil, err
	}
	result, err := c.sendRaw(command)
	completedAt := time.Now().UnixMilli()
	c.LastCommandTiming = map[string]any{
		"method":       method,
		"target":       command.Target,
		"started_at":   startedAt,
		"completed_at": completedAt,
		"duration_ms":  completedAt - startedAt,
	}
	if err != nil {
		return nil, err
	}
	if validateSchema {
		if err := c.validateCommandResult(method, result); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (c *ModCDPClient) SendRaw(method string, params map[string]any, sessionID ...string) (map[string]any, error) {
	startedAt := time.Now().UnixMilli()
	if params == nil {
		params = map[string]any{}
	}
	targetSessionID := ""
	if len(sessionID) > 0 {
		targetSessionID = sessionID[0]
	}
	result, err := c.sendMessage(method, params, targetSessionID)
	completedAt := time.Now().UnixMilli()
	c.LastRawTiming = map[string]any{
		"method":       method,
		"started_at":   startedAt,
		"completed_at": completedAt,
		"duration_ms":  completedAt - startedAt,
	}
	return result, err
}

func (c *ModCDPClient) OnRaw(event string, handler Handler) *ModCDPClient {
	return c.On(event, handler)
}

func (c *ModCDPClient) OnCDP(event string, handler func(CDPEvent)) *ModCDPClient {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	c.cdpHandlers[event] = append(c.cdpHandlers[event], handler)
	return c
}

func (c *ModCDPClient) On(event string, handler Handler) *ModCDPClient {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	pointer := handlerPointer(handler)
	for _, existing := range c.handlers[event] {
		if existing.pointer == pointer {
			return c
		}
	}
	c.handlers[event] = append(c.handlers[event], handlerEntry{handler: handler, pointer: pointer})
	return c
}

func (c *ModCDPClient) Once(event string, handler Handler) *ModCDPClient {
	var wrapped Handler
	wrapped = func(data any) {
		c.Off(event, wrapped)
		handler(data)
	}
	return c.On(event, wrapped)
}

func (c *ModCDPClient) Off(event string, handler Handler) *ModCDPClient {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	pointer := handlerPointer(handler)
	entries := c.handlers[event]
	filtered := entries[:0]
	for _, entry := range entries {
		if entry.pointer != pointer {
			filtered = append(filtered, entry)
		}
	}
	if len(filtered) == 0 {
		delete(c.handlers, event)
	} else {
		c.handlers[event] = filtered
	}
	return c
}

func handlerPointer(handler Handler) uintptr {
	if handler == nil {
		return 0
	}
	return reflect.ValueOf(handler).Pointer()
}

func (c *ModCDPClient) Close() {
	if c.transport != nil {
		_ = c.transport.Close()
		c.transport = nil
	}
	if c.launchedBrowser != nil {
		c.launchedBrowser.Close()
		c.launchedBrowser = nil
	}
	for _, injector := range c.extensionInjectors {
		_ = injector.Close()
	}
	c.extensionInjectors = nil
}

func (c *ModCDPClient) browserLauncher() browserLauncherClient {
	switch c.opts.Launch.Mode {
	case "local":
		return NewLocalBrowserLauncher(c.opts.Launch.Options)
	case "remote":
		return NewRemoteBrowserLauncher(c.opts.Launch.Options, c.opts.Upstream.WSURL)
	case "bb":
		return NewBrowserbaseBrowserLauncher(c.opts.Launch.Options)
	case "none":
		return NewNoopBrowserLauncher(c.opts.Launch.Options)
	default:
		return nil
	}
}

func (c *ModCDPClient) upstreamTransport() upstreamTransportClient {
	switch c.opts.Upstream.Mode {
	case "ws":
		return NewWebSocketUpstreamTransport(c.opts.Upstream.WSURL)
	case "pipe":
		return NewPipeUpstreamTransport(nil, nil, "")
	case "reversews":
		return NewReverseWebSocketUpstreamTransport(c.opts.Upstream.ReverseWSBind, c.opts.Upstream.ReverseWSWaitTimeoutMS)
	case "nativemessaging":
		return NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
			ManifestPath: c.opts.Upstream.NativeMessagingManifest,
			HostName:     c.opts.Upstream.NativeMessagingHostName,
		})
	case "nats":
		return NewNatsUpstreamTransport(NatsUpstreamTransportOptions{
			URL:           c.opts.Upstream.NATSURL,
			SubjectPrefix: c.opts.Upstream.NATSSubjectPrefix,
		})
	default:
		return nil
	}
}

func (c *ModCDPClient) extensionInjectorsForConfig() []extensionInjector {
	if c.opts.Extension.Mode == "none" {
		return nil
	}
	var injectors []extensionInjector
	if c.opts.Extension.Mode == "auto" || c.opts.Extension.Mode == "discover" {
		injector := NewDiscoveredExtensionInjector(ExtensionInjectorConfig{})
		injectors = append(injectors, &injector)
	}
	if c.opts.Extension.Mode == "auto" || c.opts.Extension.Mode == "inject" {
		if c.opts.Launch.Mode == "bb" {
			injector := NewBBBrowserExtensionInjector(ExtensionInjectorConfig{})
			injectors = append(injectors, &injector)
		}
		if c.opts.Launch.Mode == "local" {
			injector := NewLocalBrowserLaunchExtensionInjector(ExtensionInjectorConfig{})
			injectors = append(injectors, &injector)
		}
		injector := NewExtensionsLoadUnpackedInjector(ExtensionInjectorConfig{})
		injectors = append(injectors, &injector)
	}
	if c.opts.Extension.Mode == "auto" || c.opts.Extension.Mode == "borrow" {
		injector := NewBorrowedExtensionInjector(ExtensionInjectorConfig{})
		injectors = append(injectors, &injector)
	}
	return injectors
}

func isKnownLaunchMode(mode string) bool {
	return mode == "local" || mode == "remote" || mode == "bb" || mode == "none"
}

func isKnownUpstreamMode(mode string) bool {
	return mode == "ws" || mode == "pipe" || mode == "nativemessaging" || mode == "reversews" || mode == "nats"
}

func isKnownExtensionMode(mode string) bool {
	return mode == "auto" || mode == "discover" || mode == "inject" || mode == "borrow" || mode == "none"
}

func (c *ModCDPClient) baseExtensionInjectorConfig(send SendCDP) ExtensionInjectorConfig {
	trustMatchedServiceWorker := c.trustServiceWorkerTarget()
	return ExtensionInjectorConfig{
		Send:               send,
		SessionIDForTarget: func(targetID string) string { return c.autoSessions.SessionIDForTarget(targetID) },
		AttachToTarget: func(targetID string) string {
			return c.ensureSessionIDForTarget(targetID, time.Duration(c.opts.Extension.ServiceWorkerProbeTimeoutMS)*time.Millisecond, true)
		},
		WaitForExecutionContext: func(sessionID string, timeoutMS int) int {
			contextID, _ := c.autoSessions.WaitForExecutionContext(sessionID, timeoutMS)
			return contextID
		},
		ExtensionPath:                c.opts.Extension.Path,
		ExtensionID:                  c.opts.Extension.ExtensionID,
		WakePath:                     firstNonEmptyString(c.opts.Extension.WakePath, DefaultModCDPWakePath),
		WakeURL:                      c.opts.Extension.WakeURL,
		ServiceWorkerURLIncludes:     c.opts.Extension.ServiceWorkerURLIncludes,
		ServiceWorkerURLSuffixes:     c.opts.Extension.ServiceWorkerURLSuffixes,
		TrustMatchedServiceWorker:    trustMatchedServiceWorker,
		RequireServiceWorkerTarget:   c.opts.Extension.RequireServiceWorkerTarget || c.opts.Extension.Mode == "discover",
		ServiceWorkerReadyExpression: c.opts.Extension.ServiceWorkerReadyExpression,
		CDPSendTimeoutMS:             c.opts.Client.CDPSendTimeoutMS,
		ExecutionContextTimeoutMS:    c.opts.Extension.ExecutionContextTimeoutMS,
		ServiceWorkerProbeTimeoutMS:  c.opts.Extension.ServiceWorkerProbeTimeoutMS,
		ServiceWorkerReadyTimeoutMS:  c.opts.Extension.ServiceWorkerReadyTimeoutMS,
		ServiceWorkerPollIntervalMS:  c.opts.Extension.ServiceWorkerPollIntervalMS,
		TargetSessionPollIntervalMS:  c.opts.Extension.TargetSessionPollIntervalMS,
	}
}

func (c *ModCDPClient) injectExtension(injectors []extensionInjector) (*ExtensionInjectionResult, error) {
	if len(injectors) == 0 {
		return nil, fmt.Errorf("extension.mode='none' cannot be used with a raw_cdp upstream")
	}
	send := func(method string, params map[string]any, sessionID string) (map[string]any, error) {
		return c.sendMessageTimeout(method, params, sessionID, time.Duration(c.opts.Client.CDPSendTimeoutMS)*time.Millisecond)
	}
	var errors []string
	for _, injector := range injectors {
		injector.Update(c.baseExtensionInjectorConfig(send))
		if err := injector.Prepare(); err != nil {
			errors = append(errors, fmt.Sprintf("%T: %v", injector, err))
			continue
		}
		result, err := injector.Inject()
		if err != nil {
			errors = append(errors, fmt.Sprintf("%T: %v", injector, err))
			continue
		}
		if result != nil {
			return result, nil
		}
	}
	return nil, fmt.Errorf("cannot install, discover, or borrow the ModCDP extension in the running browser.%s", formatInjectorErrors(errors))
}

func formatInjectorErrors(errors []string) string {
	if len(errors) == 0 {
		return ""
	}
	return "\n\n" + strings.Join(errors, "\n")
}

func (c *ModCDPClient) sendRaw(command rawCommand) (any, error) {
	if command.Target == "direct_cdp" {
		step := command.Steps[0]
		return c.sendMessage(step.Method, step.Params, step.SessionID)
	}
	if command.Target != "service_worker" {
		return nil, fmt.Errorf("unsupported command target %q", command.Target)
	}

	var result map[string]any
	unwrap := ""
	for _, step := range command.Steps {
		r, err := c.sendMessage(step.Method, step.Params, c.ExtSessionID)
		if err != nil {
			return nil, err
		}
		result = r
		unwrap = step.Unwrap
	}
	return unwrapResponseIfNeeded(result, unwrap)
}

func (c *ModCDPClient) measurePingLatency() error {
	sent_at := time.Now().UnixMilli()
	ch := make(chan any, 1)
	c.Once("Mod.pong", func(data any) {
		select {
		case ch <- data:
		default:
		}
	})
	if _, err := c.Send("Mod.ping", map[string]any{"sent_at": sent_at}); err != nil {
		return err
	}
	select {
	case payload := <-ch:
		returned_at := time.Now().UnixMilli()
		latency := map[string]any{
			"sent_at":           sent_at,
			"received_at":       nil,
			"returned_at":       returned_at,
			"round_trip_ms":     returned_at - sent_at,
			"service_worker_ms": nil,
			"return_path_ms":    nil,
		}
		if data, ok := payload.(map[string]any); ok {
			if received_at, ok := numberAsInt64(data["received_at"]); ok {
				latency["received_at"] = received_at
				latency["service_worker_ms"] = received_at - sent_at
				latency["return_path_ms"] = returned_at - received_at
			}
		}
		c.Latency = latency
		return nil
	case <-time.After(time.Duration(c.opts.Client.EventWaitTimeoutMS) * time.Millisecond):
		return fmt.Errorf("Mod.pong timed out")
	}
}

func numberAsInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int64:
		return v, true
	case int:
		return int64(v), true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}

func (c *ModCDPClient) sendMessage(method string, params map[string]any, sessionID string) (map[string]any, error) {
	return c.sendMessageTimeout(method, params, sessionID, time.Duration(c.opts.Client.CDPSendTimeoutMS)*time.Millisecond)
}

func (c *ModCDPClient) sendMessageTimeout(method string, params map[string]any, sessionID string, timeout time.Duration) (map[string]any, error) {
	c.mu.Lock()
	c.nextID++
	id := c.nextID
	ch := make(chan map[string]any, 1)
	c.pending[id] = ch
	c.mu.Unlock()

	msg := map[string]any{"id": id, "method": method, "params": params}
	if sessionID != "" {
		msg["sessionId"] = sessionID
	}
	var err error
	if c.transport != nil {
		err = c.transport.Send(msg)
	} else {
		err = fmt.Errorf("ModCDP upstream is not connected")
	}
	if err != nil {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, err
	}
	if timeout <= 0 {
		resp := <-ch
		if errObj, ok := resp["error"].(map[string]any); ok {
			return nil, fmt.Errorf("%s failed: %v", method, errObj["message"])
		}
		if r, ok := resp["result"].(map[string]any); ok {
			return r, nil
		}
		return map[string]any{}, nil
	}
	select {
	case <-time.After(timeout):
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
		return nil, fmt.Errorf("%s timed out after %s", method, timeout)
	case resp := <-ch:
		if errObj, ok := resp["error"].(map[string]any); ok {
			return nil, fmt.Errorf("%s failed: %v", method, errObj["message"])
		}
		if r, ok := resp["result"].(map[string]any); ok {
			return r, nil
		}
		return map[string]any{}, nil
	}
}

func (c *ModCDPClient) rejectAll(err error) {
	c.mu.Lock()
	pending := c.pending
	c.pending = map[int64]chan map[string]any{}
	c.mu.Unlock()
	for _, ch := range pending {
		ch <- map[string]any{"error": map[string]any{"message": fmt.Sprintf("connection closed: %v", err)}}
	}
}

func (c *ModCDPClient) handleMessage(msg map[string]any) {
	if idF, ok := msg["id"].(float64); ok {
		id := int64(idF)
		c.mu.Lock()
		ch, ok := c.pending[id]
		delete(c.pending, id)
		c.mu.Unlock()
		if ok {
			ch <- msg
		}
		return
	}
	if id, ok := msg["id"].(int); ok {
		c.mu.Lock()
		ch, found := c.pending[int64(id)]
		delete(c.pending, int64(id))
		c.mu.Unlock()
		if found {
			ch <- msg
		}
		return
	}
	c.handleEventMessage(msg)
}

func (c *ModCDPClient) handleEventMessage(msg map[string]any) {
	method, _ := msg["method"].(string)
	sessionID, _ := msg["sessionId"].(string)
	params, _ := msg["params"].(map[string]any)
	c.autoSessions.RecordProtocolEvent(method, params, sessionID)
	if sessionID == c.ExtSessionID {
		bindingName, _ := params["name"].(string)
		if event, data, ok := unwrapEventIfNeeded(method, params, sessionID, c.ExtSessionID); ok {
			validatedData, valid := c.validateEventData(event, data)
			if !valid {
				return
			}
			c.handlersMu.Lock()
			hs := append([]handlerEntry(nil), c.handlers[event]...)
			cdpHandlers := append([]func(CDPEvent){}, c.cdpHandlers["*"]...)
			cdpHandlers = append(cdpHandlers, c.cdpHandlers[event]...)
			c.handlersMu.Unlock()
			for _, h := range hs {
				go h.handler(validatedData)
			}
			if bindingName == upstreamEventBindingName {
				dataMap, _ := validatedData.(map[string]any)
				cdpEvent := CDPEvent{Method: event, Params: dataMap, CDPSessionID: sessionID, SessionID: sessionID}
				for _, h := range cdpHandlers {
					go h(cdpEvent)
				}
			}
		}
		return
	}
	if method != "" {
		validatedParams, valid := c.validateEventData(method, params)
		if !valid {
			return
		}
		validatedParamsMap, _ := validatedParams.(map[string]any)
		if validatedParamsMap == nil {
			validatedParamsMap = map[string]any{}
		}
		c.handlersMu.Lock()
		hs := append([]handlerEntry(nil), c.handlers[method]...)
		cdpHandlers := append([]func(CDPEvent){}, c.cdpHandlers["*"]...)
		cdpHandlers = append(cdpHandlers, c.cdpHandlers[method]...)
		c.handlersMu.Unlock()
		for _, h := range hs {
			go h.handler(validatedParams)
		}
		if len(cdpHandlers) > 0 {
			event := CDPEvent{Method: method, Params: validatedParamsMap, CDPSessionID: sessionID, SessionID: sessionID}
			for _, h := range cdpHandlers {
				go h(event)
			}
		}
	}
}

func (c *ModCDPClient) trustServiceWorkerTarget() bool {
	if c.opts.Extension.TrustServiceWorkerTarget || len(c.opts.Extension.ServiceWorkerURLIncludes) > 0 {
		return true
	}
	for _, suffix := range c.opts.Extension.ServiceWorkerURLSuffixes {
		parts := 0
		for _, part := range strings.Split(suffix, "/") {
			if part != "" {
				parts++
			}
		}
		if parts > 1 {
			return true
		}
	}
	return false
}

func (c *ModCDPClient) sessionIDForTarget(targetID string, timeout time.Duration) string {
	if timeout <= 0 {
		return c.autoSessions.SessionIDForTarget(targetID)
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline.Add(time.Millisecond)) {
		sessionID := c.autoSessions.SessionIDForTarget(targetID)
		if sessionID != "" {
			return sessionID
		}
		time.Sleep(time.Duration(c.opts.Extension.TargetSessionPollIntervalMS) * time.Millisecond)
	}
	return ""
}

func (c *ModCDPClient) ensureSessionIDForTarget(targetID string, timeout time.Duration, allowAttach bool) string {
	sessionID := c.autoSessions.SessionIDForTarget(targetID)
	if sessionID != "" {
		return sessionID
	}
	if allowAttach {
		attachedSessionID := c.autoSessions.AttachToTarget(targetID)
		if attachedSessionID != "" {
			return attachedSessionID
		}
	}
	return c.sessionIDForTarget(targetID, timeout)
}
