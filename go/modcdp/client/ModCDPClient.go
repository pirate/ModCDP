// ModCDPClient (Go): importable, no CLI, no demo code.
//
// Option groups mirror the JS / Python ports:
//
//	Launcher       browser/session creation and cleanup.
//	Upstream       message transport to raw CDP or a ModCDP server.
//	Injector       raw-CDP extension discovery/injection/borrowing.
//	Server         ModCDPServer.configure params.
//	Client        client routing and client-owned send/event timings.
//	Upstream      upstream transport options and upstream-owned timings.
//
// Public methods: Connect, Send(method, params), SendRaw, On, OnRaw, Close.
// Synchronous; one background goroutine reads messages off the WS.
//
// Route and ModCDP wire translation lives in translate.go. Launchers and
// upstream transports live in their matching class files.
package client

import (
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
	"github.com/browserbase/modcdp/go/modcdp/injector"
	"github.com/browserbase/modcdp/go/modcdp/launcher"
	"github.com/browserbase/modcdp/go/modcdp/router"
	"github.com/browserbase/modcdp/go/modcdp/translate"
	transportpkg "github.com/browserbase/modcdp/go/modcdp/transport"
	"github.com/browserbase/modcdp/go/modcdp/types"
)

var (
	extIDFromURL = regexp.MustCompile(`^chrome-extension://([a-z]+)/`)
)

const modcdpReadyExpression = `Boolean(globalThis.ModCDP?.__ModCDPServerVersion >= 1 && globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)`

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

func boolPointer(value bool) *bool {
	return &value
}

type LaunchOptions = types.LaunchOptions
type LaunchedBrowser = launcher.LaunchedBrowser
type BrowserLauncher = launcher.BrowserLauncher
type LocalBrowserLauncher = launcher.LocalBrowserLauncher
type RemoteBrowserLauncher = launcher.RemoteBrowserLauncher
type BrowserbaseBrowserLauncher = launcher.BrowserbaseBrowserLauncher
type NoopBrowserLauncher = launcher.NoopBrowserLauncher
type ExtensionInjectorConfig = types.ExtensionInjectorConfig
type ExtensionInjectionResult = types.ExtensionInjectionResult
type SendCDP = types.SendCDP
type AttachToTarget = types.AttachToTarget
type ExtensionInjector = injector.ExtensionInjector
type DiscoveredExtensionInjector = injector.DiscoveredExtensionInjector
type BBBrowserExtensionInjector = injector.BBBrowserExtensionInjector
type LocalBrowserLaunchExtensionInjector = injector.LocalBrowserLaunchExtensionInjector
type ExtensionsLoadUnpackedInjector = injector.ExtensionsLoadUnpackedInjector
type BorrowedExtensionInjector = injector.BorrowedExtensionInjector
type UpstreamMode = transportpkg.UpstreamMode
type UpstreamEndpointKind = transportpkg.UpstreamEndpointKind
type UpstreamTransport = transportpkg.UpstreamTransport
type WebSocketUpstreamTransport = transportpkg.WebSocketUpstreamTransport
type WebSocketUpstreamTransportOptions = transportpkg.WebSocketUpstreamTransportOptions
type PipeUpstreamTransport = transportpkg.PipeUpstreamTransport
type PipeUpstreamTransportOptions = transportpkg.PipeUpstreamTransportOptions
type ReverseWebSocketUpstreamTransport = transportpkg.ReverseWebSocketUpstreamTransport
type ReverseWebSocketUpstreamTransportOptions = transportpkg.ReverseWebSocketUpstreamTransportOptions
type NativeMessagingUpstreamTransport = transportpkg.NativeMessagingUpstreamTransport
type NativeMessagingUpstreamTransportOptions = transportpkg.NativeMessagingUpstreamTransportOptions
type NatsUpstreamTransport = transportpkg.NatsUpstreamTransport
type NatsUpstreamTransportOptions = transportpkg.NatsUpstreamTransportOptions
type AutoSessionRouter = router.AutoSessionRouter

var NewLocalBrowserLauncher = launcher.NewLocalBrowserLauncher
var NewRemoteBrowserLauncher = launcher.NewRemoteBrowserLauncher
var NewBrowserbaseBrowserLauncher = launcher.NewBrowserbaseBrowserLauncher
var NewNoopBrowserLauncher = launcher.NewNoopBrowserLauncher
var NewDiscoveredExtensionInjector = injector.NewDiscoveredExtensionInjector
var NewBBBrowserExtensionInjector = injector.NewBBBrowserExtensionInjector
var NewLocalBrowserLaunchExtensionInjector = injector.NewLocalBrowserLaunchExtensionInjector
var NewExtensionsLoadUnpackedInjector = injector.NewExtensionsLoadUnpackedInjector
var NewBorrowedExtensionInjector = injector.NewBorrowedExtensionInjector
var NewWebSocketUpstreamTransport = transportpkg.NewWebSocketUpstreamTransport
var NewPipeUpstreamTransport = transportpkg.NewPipeUpstreamTransport
var NewReverseWebSocketUpstreamTransport = transportpkg.NewReverseWebSocketUpstreamTransport
var NewNativeMessagingUpstreamTransport = transportpkg.NewNativeMessagingUpstreamTransport
var NewNatsUpstreamTransport = transportpkg.NewNatsUpstreamTransport
var NewAutoSessionRouter = router.NewAutoSessionRouter

var DefaultModCDPServiceWorkerURLSuffixes = injector.DefaultModCDPServiceWorkerURLSuffixes

const DefaultModCDPWakePath = injector.DefaultModCDPWakePath
const DefaultModCDPExtensionID = injector.DefaultModCDPExtensionID
const DefaultUpstreamReverseWSBind = transportpkg.DefaultUpstreamReverseWSBind
const DefaultUpstreamReverseWSWaitTimeoutMS = transportpkg.DefaultUpstreamReverseWSWaitTimeoutMS
const DefaultUpstreamNATSWaitTimeoutMS = transportpkg.DefaultUpstreamNATSWaitTimeoutMS
const DefaultUpstreamNativeMessagingWaitTimeoutMS = transportpkg.DefaultUpstreamNativeMessagingWaitTimeoutMS
const UpstreamEndpointKindRawCDP = transportpkg.UpstreamEndpointKindRawCDP
const UpstreamEndpointKindModCDPServer = transportpkg.UpstreamEndpointKindModCDPServer

var endpointKindForUpstream = transportpkg.EndpointKindForUpstream

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

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
	ServerLoopbackCDPURL                    string            `json:"server_loopback_cdp_url,omitempty"`
	ServerRoutes                            map[string]string `json:"server_routes,omitempty"`
	ServerBrowserToken                      string            `json:"server_browser_token,omitempty"`
	ServerCDPSendTimeoutMS                  int               `json:"server_cdp_send_timeout_ms,omitempty"`
	ServerLoopbackExecutionContextTimeoutMS int               `json:"server_loopback_execution_context_timeout_ms,omitempty"`
	ServerWSConnectErrorSettleTimeoutMS     int               `json:"server_ws_connect_error_settle_timeout_ms,omitempty"`
	Options                                 map[string]any    `json:"-"`
	disabled                                bool
}

var ServerNone = &ServerConfig{disabled: true}

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

type LauncherConfig struct {
	LauncherMode           string        `json:"launcher_mode,omitempty"`
	LauncherExecutablePath string        `json:"launcher_executable_path,omitempty"`
	LauncherUserDataDir    string        `json:"launcher_user_data_dir,omitempty"`
	LauncherOptions        LaunchOptions `json:"launcher_options,omitempty"`
}

type UpstreamConfig struct {
	UpstreamMode                          string   `json:"upstream_mode,omitempty"`
	UpstreamCDPURL                        string   `json:"upstream_cdp_url,omitempty"`
	UpstreamNATSURL                       string   `json:"upstream_nats_url,omitempty"`
	UpstreamNATSSubjectPrefix             string   `json:"upstream_nats_subject_prefix,omitempty"`
	UpstreamNATSWaitTimeoutMS             int      `json:"upstream_nats_wait_timeout_ms,omitempty"`
	UpstreamReverseWSBind                 string   `json:"upstream_reversews_bind,omitempty"`
	UpstreamReverseWSWaitTimeoutMS        int      `json:"upstream_reversews_wait_timeout_ms,omitempty"`
	UpstreamNativeMessagingManifest       string   `json:"upstream_nativemessaging_manifest,omitempty"`
	UpstreamNativeMessagingManifests      []string `json:"upstream_nativemessaging_manifests,omitempty"`
	UpstreamNativeMessagingHostName       string   `json:"upstream_nativemessaging_host_name,omitempty"`
	UpstreamNativeMessagingWaitTimeoutMS  int      `json:"upstream_nativemessaging_wait_timeout_ms,omitempty"`
	UpstreamWSConnectErrorSettleTimeoutMS int      `json:"upstream_ws_connect_error_settle_timeout_ms,omitempty"`
}

type InjectorConfig struct {
	InjectorMode                         string   `json:"injector_mode,omitempty"`
	InjectorExtensionPath                string   `json:"injector_extension_path,omitempty"`
	InjectorExtensionID                  string   `json:"injector_extension_id,omitempty"`
	InjectorWakePath                     string   `json:"injector_wake_path,omitempty"`
	InjectorWakeURL                      string   `json:"injector_wake_url,omitempty"`
	InjectorServiceWorkerURLIncludes     []string `json:"injector_service_worker_url_includes,omitempty"`
	InjectorServiceWorkerURLSuffixes     []string `json:"injector_service_worker_url_suffixes,omitempty"`
	InjectorTrustServiceWorkerTarget     bool     `json:"injector_trust_service_worker_target,omitempty"`
	InjectorRequireServiceWorkerTarget   bool     `json:"injector_require_service_worker_target,omitempty"`
	InjectorServiceWorkerReadyExpression string   `json:"injector_service_worker_ready_expression,omitempty"`
	InjectorExecutionContextTimeoutMS    int      `json:"injector_execution_context_timeout_ms,omitempty"`
	InjectorServiceWorkerProbeTimeoutMS  int      `json:"injector_service_worker_probe_timeout_ms,omitempty"`
	InjectorServiceWorkerReadyTimeoutMS  int      `json:"injector_service_worker_ready_timeout_ms,omitempty"`
	InjectorServiceWorkerPollIntervalMS  int      `json:"injector_service_worker_poll_interval_ms,omitempty"`
	InjectorTargetSessionPollIntervalMS  int      `json:"injector_target_session_poll_interval_ms,omitempty"`
}

type ClientConfig struct {
	ClientRoutes               map[string]string `json:"client_routes,omitempty"`
	ClientHydrateAliases       *bool             `json:"client_hydrate_aliases,omitempty"`
	ClientMirrorUpstreamEvents *bool             `json:"client_mirror_upstream_events,omitempty"`
	ClientCDPSendTimeoutMS     int               `json:"client_cdp_send_timeout_ms,omitempty"`
	ClientEventWaitTimeoutMS   int               `json:"client_event_wait_timeout_ms,omitempty"`
}

type Options struct {
	Launcher          LauncherConfig     `json:"launcher,omitempty"`
	Upstream          UpstreamConfig     `json:"upstream,omitempty"`
	Injector          InjectorConfig     `json:"injector,omitempty"`
	Client            ClientConfig       `json:"client,omitempty"`
	Server            *ServerConfig      `json:"server,omitempty"`
	CustomCommands    []CustomCommand    `json:"custom_commands,omitempty"`
	CustomEvents      []CustomEvent      `json:"custom_events,omitempty"`
	CustomMiddlewares []CustomMiddleware `json:"custom_middlewares,omitempty"`
	serverConfigured  bool
}

func (o *Options) UnmarshalJSON(data []byte) error {
	type optionsAlias Options
	var decoded optionsAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	*o = Options(decoded)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	rawServer, hasServer := raw["server"]
	if !hasServer {
		return nil
	}
	o.serverConfigured = true
	if strings.TrimSpace(string(rawServer)) == "null" {
		o.Server = nil
		return nil
	}
	var server ServerConfig
	if err := json.Unmarshal(rawServer, &server); err != nil {
		return err
	}
	o.Server = &server
	return nil
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

	Launcher                 LauncherConfig
	Upstream                 UpstreamConfig
	Injector                 InjectorConfig
	Client                   ClientConfig
	Server                   *ServerConfig
	CustomCommands           []CustomCommand
	CustomEvents             []CustomEvent
	CustomMiddlewares        []CustomMiddleware
	UpstreamEndpointKind     UpstreamEndpointKind
	CDPURL                   string
	transport                upstreamTransportClient
	mu                       sync.Mutex
	nextID                   int64
	pending                  map[int64]chan map[string]any
	handlers                 map[string][]handlerEntry
	cdpHandlers              map[string][]func(CDPEvent)
	commandParamsSchemas     map[string]map[string]any
	commandResultSchemas     map[string]map[string]any
	commandResultUnwrapKeys  map[string]string
	eventSchemas             map[string]map[string]any
	schemaMu                 sync.RWMutex
	handlersMu               sync.Mutex
	autoSessions             *AutoSessionRouter
	ExtensionID              string
	ExtTargetID              string
	ExtSessionID             string
	ExtExecutionContextID    int
	Latency                  map[string]any
	ConnectTiming            map[string]any
	LastCommandTiming        map[string]any
	LastRawTiming            map[string]any
	launchedBrowser          *LaunchedBrowser
	extensionInjectors       []extensionInjector
	configuredPeerGeneration int64
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
	GetServerConfig() map[string]any
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
	PeerGeneration() int64
}

func New(opts Options) *ModCDPClient {
	if opts.Upstream.UpstreamMode == "" {
		opts.Upstream.UpstreamMode = "ws"
	}
	upstreamEndpointKind := UpstreamEndpointKindModCDPServer
	if opts.Upstream.UpstreamMode == "ws" || opts.Upstream.UpstreamMode == "pipe" {
		upstreamEndpointKind = UpstreamEndpointKindRawCDP
	}
	if opts.Launcher.LauncherMode == "" {
		if upstreamEndpointKind == UpstreamEndpointKindModCDPServer {
			opts.Launcher.LauncherMode = "none"
		} else if opts.Upstream.UpstreamCDPURL != "" {
			opts.Launcher.LauncherMode = "remote"
		} else {
			opts.Launcher.LauncherMode = "local"
		}
	}
	if opts.Injector.InjectorMode == "" {
		if upstreamEndpointKind == UpstreamEndpointKindRawCDP || opts.Launcher.LauncherMode != "none" {
			opts.Injector.InjectorMode = "auto"
		} else {
			opts.Injector.InjectorMode = "none"
		}
	}
	if opts.Launcher.LauncherExecutablePath != "" {
		opts.Launcher.LauncherOptions.ExecutablePath = opts.Launcher.LauncherExecutablePath
	}
	if opts.Launcher.LauncherUserDataDir != "" {
		opts.Launcher.LauncherOptions.UserDataDir = opts.Launcher.LauncherUserDataDir
	}
	if opts.Client.ClientRoutes == nil {
		opts.Client.ClientRoutes = translate.DefaultClientRoutes()
	} else {
		merged := translate.DefaultClientRoutes()
		for k, v := range opts.Client.ClientRoutes {
			merged[k] = v
		}
		opts.Client.ClientRoutes = merged
	}
	if opts.Client.ClientHydrateAliases == nil {
		value := true
		opts.Client.ClientHydrateAliases = &value
	}
	if opts.Server != nil && opts.Server.disabled {
		opts.Server = nil
		opts.serverConfigured = true
	}
	if opts.Server == nil && !opts.serverConfigured {
		opts.Server = &ServerConfig{}
	}
	if upstreamEndpointKind == UpstreamEndpointKindModCDPServer && opts.Server != nil && opts.Server.ServerRoutes == nil {
		opts.Server.ServerRoutes = map[string]string{"*.*": "chrome_debugger"}
	}
	if opts.Injector.InjectorServiceWorkerURLSuffixes == nil {
		opts.Injector.InjectorServiceWorkerURLSuffixes = append([]string{}, DefaultModCDPServiceWorkerURLSuffixes...)
	}
	if opts.Client.ClientCDPSendTimeoutMS == 0 {
		opts.Client.ClientCDPSendTimeoutMS = DefaultCDPSendTimeoutMS
	}
	if opts.Client.ClientEventWaitTimeoutMS == 0 {
		opts.Client.ClientEventWaitTimeoutMS = DefaultEventWaitTimeoutMS
	}
	if opts.Injector.InjectorExecutionContextTimeoutMS == 0 {
		opts.Injector.InjectorExecutionContextTimeoutMS = DefaultExecutionContextTimeoutMS
	}
	if opts.Injector.InjectorServiceWorkerProbeTimeoutMS == 0 {
		opts.Injector.InjectorServiceWorkerProbeTimeoutMS = DefaultServiceWorkerProbeTimeoutMS
	}
	if opts.Injector.InjectorServiceWorkerReadyTimeoutMS == 0 {
		opts.Injector.InjectorServiceWorkerReadyTimeoutMS = DefaultServiceWorkerReadyTimeoutMS
	}
	if opts.Injector.InjectorServiceWorkerPollIntervalMS == 0 {
		opts.Injector.InjectorServiceWorkerPollIntervalMS = DefaultServiceWorkerPollIntervalMS
	}
	if opts.Injector.InjectorTargetSessionPollIntervalMS == 0 {
		opts.Injector.InjectorTargetSessionPollIntervalMS = DefaultTargetSessionPollIntervalMS
	}
	if opts.Upstream.UpstreamWSConnectErrorSettleTimeoutMS == 0 {
		opts.Upstream.UpstreamWSConnectErrorSettleTimeoutMS = DefaultWSConnectErrorSettleTimeoutMS
	}
	if opts.Upstream.UpstreamReverseWSBind == "" {
		opts.Upstream.UpstreamReverseWSBind = DefaultUpstreamReverseWSBind
	}
	if opts.Upstream.UpstreamNATSWaitTimeoutMS == 0 {
		opts.Upstream.UpstreamNATSWaitTimeoutMS = DefaultUpstreamNATSWaitTimeoutMS
	}
	if opts.Upstream.UpstreamReverseWSWaitTimeoutMS == 0 {
		opts.Upstream.UpstreamReverseWSWaitTimeoutMS = DefaultUpstreamReverseWSWaitTimeoutMS
	}
	if opts.Upstream.UpstreamNativeMessagingWaitTimeoutMS == 0 {
		opts.Upstream.UpstreamNativeMessagingWaitTimeoutMS = DefaultUpstreamNativeMessagingWaitTimeoutMS
	}
	client := &ModCDPClient{
		Launcher:                opts.Launcher,
		Upstream:                opts.Upstream,
		Injector:                opts.Injector,
		Client:                  opts.Client,
		Server:                  opts.Server,
		CustomCommands:          opts.CustomCommands,
		CustomEvents:            opts.CustomEvents,
		CustomMiddlewares:       opts.CustomMiddlewares,
		UpstreamEndpointKind:    upstreamEndpointKind,
		pending:                 map[int64]chan map[string]any{},
		handlers:                map[string][]handlerEntry{},
		cdpHandlers:             map[string][]func(CDPEvent){},
		commandParamsSchemas:    map[string]map[string]any{},
		commandResultSchemas:    map[string]map[string]any{},
		commandResultUnwrapKeys: map[string]string{},
		eventSchemas:            map[string]map[string]any{},
	}
	client.Mod = ModDomain{client: client}
	client.autoSessions = NewAutoSessionRouter(
		func(method string, params map[string]any, sessionID string) (map[string]any, error) {
			return client.sendMessage(method, params, sessionID)
		},
		func() int { return client.Injector.InjectorExecutionContextTimeoutMS },
	)
	if *client.Client.ClientHydrateAliases {
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
	if transportpkg.EndpointKindForUpstream(c.Upstream.UpstreamMode) == UpstreamEndpointKindModCDPServer {
		if err := c.transport.WaitForPeer(); err != nil {
			c.Close()
			return err
		}
		if c.Server != nil {
			if _, err := c.sendMessage("Mod.configure", c.serverConfigureParams(nil, nil, nil), ""); err != nil {
				c.Close()
				return err
			}
			c.configuredPeerGeneration = c.transport.PeerGeneration()
		}
		c.startPingLatencyMeasurement()
		connectedAt := time.Now().UnixMilli()
		c.ConnectTiming = map[string]any{
			"started_at":             connectStartedAt,
			"upstream_mode":          c.Upstream.UpstreamMode,
			"upstream_endpoint_kind": transportpkg.EndpointKindForUpstream(c.Upstream.UpstreamMode),
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
	extExecutionContextID, err := c.autoSessions.WaitForExecutionContext(c.ExtSessionID, c.Injector.InjectorExecutionContextTimeoutMS)
	if err != nil {
		c.Close()
		return err
	}
	c.ExtExecutionContextID = extExecutionContextID
	if _, err := c.sendMessage("Runtime.addBinding", map[string]any{"name": translate.CustomEventBindingName}, c.ExtSessionID); err != nil {
		c.Close()
		return err
	}
	mirrorUpstreamEvents := true
	if c.Client.ClientMirrorUpstreamEvents != nil {
		mirrorUpstreamEvents = *c.Client.ClientMirrorUpstreamEvents
	}
	if mirrorUpstreamEvents {
		if _, err := c.sendMessage("Runtime.addBinding", map[string]any{"name": translate.UpstreamEventBindingName}, c.ExtSessionID); err != nil {
			c.Close()
			return err
		}
	}

	if c.Server != nil {
		customCommands := make([]map[string]any, 0, len(c.CustomCommands))
		for _, command := range c.CustomCommands {
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
		customEvents := make([]map[string]any, 0, len(c.CustomEvents))
		for _, event := range c.CustomEvents {
			customEvents = append(customEvents, map[string]any{
				"name":         event.Name,
				"event_schema": event.EventSchema,
			})
		}
		customMiddlewares := make([]map[string]any, 0, len(c.CustomMiddlewares))
		for _, middleware := range c.CustomMiddlewares {
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
		command, err := translate.WrapCommandIfNeeded("Mod.configure", configureParams, c.Client.ClientRoutes, c.ExtSessionID)
		if err != nil {
			c.Close()
			return fmt.Errorf("Mod.configure: %w", err)
		}
		if _, err := c.sendRaw(command); err != nil {
			c.Close()
			return fmt.Errorf("Mod.configure: %w", err)
		}
	}
	c.startPingLatencyMeasurement()
	connectedAt := time.Now().UnixMilli()
	c.ConnectTiming = map[string]any{
		"started_at":             connectStartedAt,
		"upstream_mode":          c.Upstream.UpstreamMode,
		"upstream_endpoint_kind": transportpkg.EndpointKindForUpstream(c.Upstream.UpstreamMode),
		"transport_started_at":   transportStartedAt,
		"transport_connected_at": transportConnectedAt,
		"transport_duration_ms":  transportConnectedAt - transportStartedAt,
		"injector_source":        ext.Source,
		"injector_started_at":    extensionStartedAt,
		"injector_completed_at":  extensionCompletedAt,
		"injector_duration_ms":   extensionCompletedAt - extensionStartedAt,
		"connected_at":           connectedAt,
		"duration_ms":            connectedAt - connectStartedAt,
	}
	return nil
}

func (c *ModCDPClient) connectUpstreamTransport() error {
	if c.transport != nil {
		return nil
	}
	if !isKnownLaunchMode(c.Launcher.LauncherMode) {
		return fmt.Errorf("unknown launcher.launcher_mode=%s", c.Launcher.LauncherMode)
	}
	if !isKnownUpstreamMode(c.Upstream.UpstreamMode) {
		return fmt.Errorf("unknown upstream.upstream_mode=%s", c.Upstream.UpstreamMode)
	}
	if !isKnownExtensionMode(c.Injector.InjectorMode) {
		return fmt.Errorf("unknown injector.injector_mode=%s", c.Injector.InjectorMode)
	}
	launcher := c.browserLauncher()
	transport := c.upstreamTransport()
	injectors := c.extensionInjectorsForConfig()
	c.extensionInjectors = injectors
	initialTransportConfig := c.upstreamTransportConfig()

	transport.Update(initialTransportConfig)
	launcher.Update(c.Launcher.LauncherOptions)
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
	launcher.Update(LaunchOptions{LoopbackCDP: boolPointer(c.serverNeedsLoopbackCDP())})
	transport.Update(launcher.GetTransportConfig())

	if transportpkg.EndpointKindForUpstream(c.Upstream.UpstreamMode) == UpstreamEndpointKindModCDPServer {
		if err := transport.Connect(); err != nil {
			return err
		}
	}
	if c.Launcher.LauncherMode != "none" {
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
		launchedCDPURL = c.launchedBrowser.CDPURL
	}
	if transportpkg.EndpointKindForUpstream(c.Upstream.UpstreamMode) == UpstreamEndpointKindRawCDP {
		if err := transport.Connect(); err != nil {
			return err
		}
	}

	c.transport = transport
	transportURL := transportURL(transport)
	if transportpkg.EndpointKindForUpstream(c.Upstream.UpstreamMode) == UpstreamEndpointKindRawCDP {
		c.CDPURL = firstNonEmptyString(transportURL, launchedCDPURL)
	} else {
		c.CDPURL = launchedCDPURL
	}
	if wsTransport, ok := transport.(*WebSocketUpstreamTransport); ok && wsTransport.URL != "" {
		// For ws mode, cdp_url has been resolved to the concrete WebSocket CDP endpoint after connect().
		c.Upstream.UpstreamCDPURL = wsTransport.URL
	}

	serverConfig := map[string]any{}
	if transportpkg.EndpointKindForUpstream(c.Upstream.UpstreamMode) == UpstreamEndpointKindModCDPServer &&
		launchedCDPURL != "" &&
		!strings.HasPrefix(launchedCDPURL, "pipe://") {
		serverConfig["server_loopback_cdp_url"] = launchedCDPURL
	}
	for key, value := range launcher.GetServerConfig() {
		serverConfig[key] = value
	}
	for key, value := range transport.GetServerConfig() {
		serverConfig[key] = value
	}
	if c.Server != nil {
		if loopbackCDPURL, _ := serverConfig["server_loopback_cdp_url"].(string); loopbackCDPURL != "" {
			initialCDPURL, _ := initialTransportConfig["cdp_url"].(string)
			if c.Server.ServerLoopbackCDPURL == "" ||
				c.Server.ServerLoopbackCDPURL == initialCDPURL ||
				c.Server.ServerLoopbackCDPURL == launchedCDPURL {
				c.Server.ServerLoopbackCDPURL = loopbackCDPURL
			}
		}
	}
	return nil
}

func (c *ModCDPClient) serverNeedsLoopbackCDP() bool {
	if c.Server == nil || c.Server.ServerLoopbackCDPURL != "" {
		return false
	}
	for _, route := range c.Server.ServerRoutes {
		if route == "loopback_cdp" {
			return true
		}
	}
	return false
}

func (c *ModCDPClient) ensureModCDPServerConfigured() error {
	if c.Server == nil || c.transport == nil {
		return nil
	}
	if err := c.transport.WaitForPeer(); err != nil {
		return err
	}
	peerGeneration := c.transport.PeerGeneration()
	if peerGeneration == c.configuredPeerGeneration {
		return nil
	}
	if _, err := c.sendMessage("Mod.configure", c.serverConfigureParams(nil, nil, nil), ""); err != nil {
		return err
	}
	c.configuredPeerGeneration = peerGeneration
	return nil
}

func (c *ModCDPClient) upstreamTransportConfig() map[string]any {
	return map[string]any{
		"cdp_url":                                  c.Upstream.UpstreamCDPURL,
		"upstream_nats_url":                        c.Upstream.UpstreamNATSURL,
		"upstream_nats_subject_prefix":             c.Upstream.UpstreamNATSSubjectPrefix,
		"upstream_nats_wait_timeout_ms":            c.Upstream.UpstreamNATSWaitTimeoutMS,
		"upstream_reversews_bind":                  c.Upstream.UpstreamReverseWSBind,
		"upstream_reversews_wait_timeout_ms":       c.Upstream.UpstreamReverseWSWaitTimeoutMS,
		"upstream_nativemessaging_manifest":        c.Upstream.UpstreamNativeMessagingManifest,
		"upstream_nativemessaging_manifests":       c.Upstream.UpstreamNativeMessagingManifests,
		"upstream_nativemessaging_host_name":       c.Upstream.UpstreamNativeMessagingHostName,
		"upstream_nativemessaging_wait_timeout_ms": c.Upstream.UpstreamNativeMessagingWaitTimeoutMS,
		"injector_extension_id":                    c.Injector.InjectorExtensionID,
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
		"server_cdp_send_timeout_ms":                   c.Client.ClientCDPSendTimeoutMS,
		"server_loopback_execution_context_timeout_ms": c.Injector.InjectorExecutionContextTimeoutMS,
		"server_ws_connect_error_settle_timeout_ms":    c.Upstream.UpstreamWSConnectErrorSettleTimeoutMS,
	}
	if c.Server != nil {
		server["server_loopback_cdp_url"] = c.Server.ServerLoopbackCDPURL
		server["server_routes"] = c.Server.ServerRoutes
		if c.Server.ServerBrowserToken != "" {
			server["server_browser_token"] = c.Server.ServerBrowserToken
		}
		if c.Server.ServerCDPSendTimeoutMS != 0 {
			server["server_cdp_send_timeout_ms"] = c.Server.ServerCDPSendTimeoutMS
		}
		if c.Server.ServerLoopbackExecutionContextTimeoutMS != 0 {
			server["server_loopback_execution_context_timeout_ms"] = c.Server.ServerLoopbackExecutionContextTimeoutMS
		}
		if c.Server.ServerWSConnectErrorSettleTimeoutMS != 0 {
			server["server_ws_connect_error_settle_timeout_ms"] = c.Server.ServerWSConnectErrorSettleTimeoutMS
		}
		for key, value := range c.Server.Options {
			server[key] = value
		}
	}
	upstream := map[string]any{"upstream_mode": c.Upstream.UpstreamMode}
	if c.Upstream.UpstreamNATSURL != "" {
		upstream["upstream_nats_url"] = c.Upstream.UpstreamNATSURL
	}
	if c.Upstream.UpstreamNATSSubjectPrefix != "" {
		upstream["upstream_nats_subject_prefix"] = c.Upstream.UpstreamNATSSubjectPrefix
	}
	if c.transport != nil {
		if reverseWSURL := c.transport.GetInjectorConfig().UpstreamReverseWSURL; reverseWSURL != "" {
			upstream["upstream_reversews_url"] = reverseWSURL
		}
	}
	return map[string]any{
		"upstream": upstream,
		"client": map[string]any{
			"client_routes": c.Client.ClientRoutes,
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

func resultUnwrapKeyFromSchema(schema map[string]any) string {
	properties, _ := schema["properties"].(map[string]any)
	if len(properties) != 1 {
		return ""
	}
	for key := range properties {
		return key
	}
	return ""
}

func (c *ModCDPClient) setCommandResultSchema(name string, schema map[string]any) {
	c.commandResultSchemas[name] = schema
	if unwrapKey := resultUnwrapKeyFromSchema(schema); unwrapKey != "" {
		c.commandResultUnwrapKeys[name] = unwrapKey
	} else {
		delete(c.commandResultUnwrapKeys, name)
	}
}

func (c *ModCDPClient) hydrateCustomSurface() {
	c.schemaMu.Lock()
	defer c.schemaMu.Unlock()
	for _, command := range c.CustomCommands {
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
			c.setCommandResultSchema(name, schema)
		}
	}
	for _, event := range c.CustomEvents {
		if event.Name == "" {
			continue
		}
		name, err := normalizeModCDPName(event.Name)
		if err != nil {
			continue
		}
		if schema := cloneSchema(event.EventSchema); schema != nil {
			c.eventSchemas[name] = schema
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
			c.setCommandResultSchema(name, schema)
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
			c.eventSchemas[name] = schema
		}
	}
	found := false
	for index, event := range c.CustomEvents {
		if event.Name == name {
			found = true
			if rawSchema, exists := params["event_schema"]; exists {
				if schemaObject, ok := rawSchema.(map[string]any); ok {
					c.CustomEvents[index].EventSchema = cloneSchema(schemaObject)
				}
			}
			break
		}
	}
	if !found {
		event := CustomEvent{Name: name}
		if rawSchema, exists := params["event_schema"]; exists {
			if schemaObject, ok := rawSchema.(map[string]any); ok {
				event.EventSchema = cloneSchema(schemaObject)
			}
		}
		c.CustomEvents = append(c.CustomEvents, event)
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

func (c *ModCDPClient) validateAndUnwrapCommandResult(method string, result any) (any, error) {
	if err := c.validateCommandResult(method, result); err != nil {
		return nil, err
	}
	c.schemaMu.RLock()
	unwrapKey := c.commandResultUnwrapKeys[method]
	c.schemaMu.RUnlock()
	if unwrapKey == "" {
		return result, nil
	}
	resultObject, ok := result.(map[string]any)
	if !ok {
		return result, nil
	}
	return resultObject[unwrapKey], nil
}

func (c *ModCDPClient) validateEventData(event string, data any) (any, bool) {
	c.schemaMu.RLock()
	schema := c.eventSchemas[event]
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
		if c.ExtSessionID == "" && transportpkg.EndpointKindForUpstream(c.Upstream.UpstreamMode) != UpstreamEndpointKindModCDPServer {
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
	if transportpkg.EndpointKindForUpstream(c.Upstream.UpstreamMode) == UpstreamEndpointKindModCDPServer {
		if method != "Mod.configure" {
			if err := c.ensureModCDPServerConfigured(); err != nil {
				return nil, err
			}
		}
		rawResult, err := c.sendMessage(method, params, "")
		var result any = rawResult
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
		if method == "Mod.configure" && c.transport != nil {
			c.configuredPeerGeneration = c.transport.PeerGeneration()
		}
		if validateSchema {
			var err error
			result, err = c.validateAndUnwrapCommandResult(method, result)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	}
	command, err := translate.WrapCommandIfNeeded(method, params, c.Client.ClientRoutes, c.ExtSessionID, targetSessionID)
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
		var err error
		result, err = c.validateAndUnwrapCommandResult(method, result)
		if err != nil {
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
	if c.launchedBrowser != nil {
		c.launchedBrowser.Close()
		c.launchedBrowser = nil
	}
	if c.transport != nil {
		_ = c.transport.Close()
		c.transport = nil
	}
	for _, injector := range c.extensionInjectors {
		_ = injector.Close()
	}
	c.extensionInjectors = nil
}

func (c *ModCDPClient) Transport() any {
	return c.transport
}

func (c *ModCDPClient) LaunchedBrowser() *LaunchedBrowser {
	return c.launchedBrowser
}

func (c *ModCDPClient) browserLauncher() browserLauncherClient {
	switch c.Launcher.LauncherMode {
	case "local":
		return NewLocalBrowserLauncher(c.Launcher.LauncherOptions)
	case "remote":
		return NewRemoteBrowserLauncher(c.Launcher.LauncherOptions, c.Upstream.UpstreamCDPURL)
	case "bb":
		return NewBrowserbaseBrowserLauncher(c.Launcher.LauncherOptions)
	case "none":
		return NewNoopBrowserLauncher(c.Launcher.LauncherOptions)
	default:
		return nil
	}
}

func (c *ModCDPClient) upstreamTransport() upstreamTransportClient {
	switch c.Upstream.UpstreamMode {
	case "ws":
		return NewWebSocketUpstreamTransport(WebSocketUpstreamTransportOptions{})
	case "pipe":
		return NewPipeUpstreamTransport(PipeUpstreamTransportOptions{})
	case "reversews":
		return NewReverseWebSocketUpstreamTransport(ReverseWebSocketUpstreamTransportOptions{})
	case "nativemessaging":
		return NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{})
	case "nats":
		return NewNatsUpstreamTransport(NatsUpstreamTransportOptions{})
	default:
		return nil
	}
}

func (c *ModCDPClient) extensionInjectorsForConfig() []extensionInjector {
	if c.Injector.InjectorMode == "none" {
		return nil
	}
	var injectors []extensionInjector
	if c.Injector.InjectorMode == "auto" || c.Injector.InjectorMode == "discover" {
		injector := NewDiscoveredExtensionInjector(ExtensionInjectorConfig{})
		injectors = append(injectors, &injector)
	}
	if c.Injector.InjectorMode == "auto" || c.Injector.InjectorMode == "inject" {
		if c.Launcher.LauncherMode == "bb" {
			injector := NewBBBrowserExtensionInjector(ExtensionInjectorConfig{})
			injectors = append(injectors, &injector)
		}
		if c.Launcher.LauncherMode == "local" {
			injector := NewLocalBrowserLaunchExtensionInjector(ExtensionInjectorConfig{})
			injectors = append(injectors, &injector)
		}
		injector := NewExtensionsLoadUnpackedInjector(ExtensionInjectorConfig{})
		injectors = append(injectors, &injector)
	}
	if c.Injector.InjectorMode == "auto" || c.Injector.InjectorMode == "borrow" {
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
	var attachToTarget AttachToTarget
	if send != nil {
		attachToTarget = func(targetID string) string {
			return c.ensureSessionIDForTarget(targetID, time.Duration(c.Injector.InjectorServiceWorkerProbeTimeoutMS)*time.Millisecond, true)
		}
	}
	return ExtensionInjectorConfig{
		Send:               send,
		SessionIDForTarget: func(targetID string) string { return c.autoSessions.SessionIDForTarget(targetID) },
		AttachToTarget:     attachToTarget,
		WaitForExecutionContext: func(sessionID string, timeoutMS int) int {
			contextID, _ := c.autoSessions.WaitForExecutionContext(sessionID, timeoutMS)
			return contextID
		},
		InjectorExtensionPath:                c.Injector.InjectorExtensionPath,
		InjectorExtensionID:                  c.Injector.InjectorExtensionID,
		InjectorWakePath:                     firstNonEmptyString(c.Injector.InjectorWakePath, DefaultModCDPWakePath),
		InjectorWakeURL:                      c.Injector.InjectorWakeURL,
		InjectorServiceWorkerURLIncludes:     c.Injector.InjectorServiceWorkerURLIncludes,
		InjectorServiceWorkerURLSuffixes:     c.Injector.InjectorServiceWorkerURLSuffixes,
		InjectorTrustServiceWorkerTarget:     trustMatchedServiceWorker,
		InjectorRequireServiceWorkerTarget:   c.Injector.InjectorRequireServiceWorkerTarget || c.Injector.InjectorMode == "discover",
		InjectorServiceWorkerReadyExpression: c.Injector.InjectorServiceWorkerReadyExpression,
		InjectorCDPSendTimeoutMS:             c.Client.ClientCDPSendTimeoutMS,
		InjectorExecutionContextTimeoutMS:    c.Injector.InjectorExecutionContextTimeoutMS,
		InjectorServiceWorkerProbeTimeoutMS:  c.Injector.InjectorServiceWorkerProbeTimeoutMS,
		InjectorServiceWorkerReadyTimeoutMS:  c.Injector.InjectorServiceWorkerReadyTimeoutMS,
		InjectorServiceWorkerPollIntervalMS:  c.Injector.InjectorServiceWorkerPollIntervalMS,
		InjectorTargetSessionPollIntervalMS:  c.Injector.InjectorTargetSessionPollIntervalMS,
	}
}

func (c *ModCDPClient) injectExtension(injectors []extensionInjector) (*ExtensionInjectionResult, error) {
	if len(injectors) == 0 {
		return nil, fmt.Errorf("injector.injector_mode='none' cannot be used with a raw_cdp upstream")
	}
	send := func(method string, params map[string]any, sessionID string) (map[string]any, error) {
		return c.sendMessageTimeout(method, params, sessionID, time.Duration(c.Client.ClientCDPSendTimeoutMS)*time.Millisecond)
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

func cloneMap(value map[string]any) map[string]any {
	cloned := map[string]any{}
	for key, item := range value {
		cloned[key] = item
	}
	return cloned
}

func (c *ModCDPClient) sendRaw(command translate.RawCommand) (any, error) {
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
		params := step.Params
		if params == nil {
			params = map[string]any{}
		}
		if step.Method == "Runtime.callFunctionOn" {
			if _, exists := params["executionContextId"]; !exists {
				if c.ExtExecutionContextID == 0 {
					contextID, err := c.autoSessions.WaitForExecutionContext(c.ExtSessionID, c.Injector.InjectorExecutionContextTimeoutMS)
					if err != nil {
						return nil, err
					}
					c.ExtExecutionContextID = contextID
				}
				params = cloneMap(params)
				params["executionContextId"] = c.ExtExecutionContextID
			}
		}
		r, err := c.sendMessage(step.Method, params, c.ExtSessionID)
		if err != nil {
			return nil, err
		}
		result = r
		unwrap = step.Unwrap
	}
	return translate.UnwrapResponseIfNeeded(result, unwrap)
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
	case <-time.After(time.Duration(c.Client.ClientEventWaitTimeoutMS) * time.Millisecond):
		return fmt.Errorf("Mod.pong timed out")
	}
}

func (c *ModCDPClient) startPingLatencyMeasurement() {
	go func() {
		_ = c.measurePingLatency()
	}()
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
	return c.sendMessageTimeout(method, params, sessionID, time.Duration(c.Client.ClientCDPSendTimeoutMS)*time.Millisecond)
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
	if c.ExtSessionID != "" && sessionID == c.ExtSessionID {
		bindingName, _ := params["name"].(string)
		if event, data, ok := translate.UnwrapEventIfNeeded(method, params, sessionID, c.ExtSessionID); ok {
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
			if bindingName == translate.UpstreamEventBindingName {
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
	if c.Injector.InjectorTrustServiceWorkerTarget || len(c.Injector.InjectorServiceWorkerURLIncludes) > 0 {
		return true
	}
	for _, suffix := range c.Injector.InjectorServiceWorkerURLSuffixes {
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
		time.Sleep(time.Duration(c.Injector.InjectorTargetSessionPollIntervalMS) * time.Millisecond)
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
