package modcdp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func boolPtr(value bool) *bool {
	return &value
}

func TestModCDPClientNormalizesNestedConfigOwners(t *testing.T) {
	cdp := New(Options{
		Launch: LaunchConfig{
			Mode:           "local",
			ExecutablePath: "/tmp/chrome",
			UserDataDir:    "/tmp/profile",
			Options: LaunchOptions{
				Headless: boolPtr(true),
			},
		},
		Upstream: UpstreamConfig{
			Mode:                          "ws",
			WSURL:                         "http://127.0.0.1:9222",
			ReverseWSWaitTimeoutMS:        456,
			WSConnectErrorSettleTimeoutMS: 321,
		},
		Extension: ExtensionConfig{
			Mode:                        "discover",
			Path:                        "/tmp/ext",
			ExtensionID:                 "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ServiceWorkerURLIncludes:    []string{"modcdp"},
			ServiceWorkerURLSuffixes:    []string{"/custom/service_worker.js"},
			TrustServiceWorkerTarget:    true,
			RequireServiceWorkerTarget:  true,
			ExecutionContextTimeoutMS:   4321,
			ServiceWorkerProbeTimeoutMS: 5432,
			ServiceWorkerReadyTimeoutMS: 6543,
			ServiceWorkerPollIntervalMS: 76,
			TargetSessionPollIntervalMS: 87,
		},
		Client: ClientConfig{
			Routes:               map[string]string{"*.*": "direct_cdp"},
			MirrorUpstreamEvents: boolPtr(false),
			CDPSendTimeoutMS:     1234,
			EventWaitTimeoutMS:   2345,
		},
		Server: &ServerConfig{
			Routes:                            map[string]string{"*.*": "loopback_cdp"},
			BrowserToken:                      "token-1",
			CDPSendTimeoutMS:                  9876,
			LoopbackExecutionContextTimeoutMS: 8765,
			WSConnectErrorSettleTimeoutMS:     7654,
		},
	})

	if cdp.opts.Launch.Options.ExecutablePath != "/tmp/chrome" {
		t.Fatalf("Launch.Options.ExecutablePath = %q", cdp.opts.Launch.Options.ExecutablePath)
	}
	if cdp.opts.Launch.Options.UserDataDir != "/tmp/profile" {
		t.Fatalf("Launch.Options.UserDataDir = %q", cdp.opts.Launch.Options.UserDataDir)
	}
	if cdp.opts.Upstream.WSConnectErrorSettleTimeoutMS != 321 {
		t.Fatalf("Upstream.WSConnectErrorSettleTimeoutMS = %d", cdp.opts.Upstream.WSConnectErrorSettleTimeoutMS)
	}
	if cdp.opts.Upstream.ReverseWSWaitTimeoutMS != 456 {
		t.Fatalf("Upstream.ReverseWSWaitTimeoutMS = %d", cdp.opts.Upstream.ReverseWSWaitTimeoutMS)
	}
	if cdp.opts.Extension.ExecutionContextTimeoutMS != 4321 {
		t.Fatalf("Extension.ExecutionContextTimeoutMS = %d", cdp.opts.Extension.ExecutionContextTimeoutMS)
	}
	if cdp.opts.Extension.ServiceWorkerProbeTimeoutMS != 5432 {
		t.Fatalf("Extension.ServiceWorkerProbeTimeoutMS = %d", cdp.opts.Extension.ServiceWorkerProbeTimeoutMS)
	}
	if cdp.opts.Extension.ServiceWorkerReadyTimeoutMS != 6543 {
		t.Fatalf("Extension.ServiceWorkerReadyTimeoutMS = %d", cdp.opts.Extension.ServiceWorkerReadyTimeoutMS)
	}
	if cdp.opts.Extension.ServiceWorkerPollIntervalMS != 76 {
		t.Fatalf("Extension.ServiceWorkerPollIntervalMS = %d", cdp.opts.Extension.ServiceWorkerPollIntervalMS)
	}
	if cdp.opts.Extension.TargetSessionPollIntervalMS != 87 {
		t.Fatalf("Extension.TargetSessionPollIntervalMS = %d", cdp.opts.Extension.TargetSessionPollIntervalMS)
	}
	if cdp.opts.Client.Routes["*.*"] != "direct_cdp" {
		t.Fatalf("Client.Routes[*.*] = %q", cdp.opts.Client.Routes["*.*"])
	}
	if cdp.opts.Client.MirrorUpstreamEvents == nil || *cdp.opts.Client.MirrorUpstreamEvents {
		t.Fatalf("Client.MirrorUpstreamEvents = %#v", cdp.opts.Client.MirrorUpstreamEvents)
	}
	if cdp.opts.Client.CDPSendTimeoutMS != 1234 {
		t.Fatalf("Client.CDPSendTimeoutMS = %d", cdp.opts.Client.CDPSendTimeoutMS)
	}
	if cdp.opts.Client.EventWaitTimeoutMS != 2345 {
		t.Fatalf("Client.EventWaitTimeoutMS = %d", cdp.opts.Client.EventWaitTimeoutMS)
	}

	params := cdp.serverConfigureParams(nil, nil, nil)
	clientConfig := params["client"].(map[string]any)
	routes := clientConfig["routes"].(map[string]string)
	serverConfig := params["server"].(map[string]any)
	if routes["*.*"] != "direct_cdp" {
		t.Fatalf("configure client routes = %#v", routes)
	}
	if serverConfig["browser_token"] != "token-1" {
		t.Fatalf("configure browser_token = %#v", serverConfig["browser_token"])
	}
	if serverConfig["cdp_send_timeout_ms"] != 9876 {
		t.Fatalf("configure cdp_send_timeout_ms = %#v", serverConfig["cdp_send_timeout_ms"])
	}
	if serverConfig["loopback_execution_context_timeout_ms"] != 8765 {
		t.Fatalf("configure loopback_execution_context_timeout_ms = %#v", serverConfig["loopback_execution_context_timeout_ms"])
	}
	if serverConfig["ws_connect_error_settle_timeout_ms"] != 7654 {
		t.Fatalf("configure ws_connect_error_settle_timeout_ms = %#v", serverConfig["ws_connect_error_settle_timeout_ms"])
	}
}

func TestModCDPClientConnectsWithLocalLaunchAndInjectorChain(t *testing.T) {
	cdp := New(Options{
		Launch: LaunchConfig{
			Mode: "local",
			Options: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{Mode: "ws"},
		Extension: ExtensionConfig{
			Mode:                     "inject",
			ServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			TrustServiceWorkerTarget: true,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["extension_source"] != "local_launch" {
		t.Fatalf("extension_source = %v", cdp.ConnectTiming["extension_source"])
	}
	if cdp.ExtensionID != DefaultModCDPExtensionID {
		t.Fatalf("ExtensionID = %q", cdp.ExtensionID)
	}
	result, err := cdp.Send("Mod.evaluate", map[string]any{
		"expression": "chrome.runtime.getURL('modcdp/service_worker.js')",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js" {
		t.Fatalf("Mod.evaluate = %#v", result)
	}
}

func TestModCDPClientCloseDoesNotCloseRemoteBrowserItDidNotLaunch(t *testing.T) {
	headless := true
	sandbox := false
	extensionPath, err := filepath.Abs("../../dist/extension")
	if err != nil {
		t.Fatal(err)
	}
	chrome, err := NewLocalBrowserLauncher(LaunchOptions{
		Headless:  &headless,
		Sandbox:   &sandbox,
		ExtraArgs: []string{"--load-extension=" + extensionPath},
	}).Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	rawConn, _, _, err := ws.Dial(ctx, chrome.WSURL)
	if err != nil {
		t.Fatal(err)
	}
	defer rawConn.Close()

	cdp := New(Options{
		Launch:   LaunchConfig{Mode: "remote"},
		Upstream: UpstreamConfig{Mode: "ws", WSURL: chrome.CDPURL},
		Extension: ExtensionConfig{
			Mode:                     "discover",
			ServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			TrustServiceWorkerTarget: true,
		},
	})
	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	cdp.Close()
	time.Sleep(500 * time.Millisecond)

	if err := wsutil.WriteClientText(rawConn, []byte(`{"id":1,"method":"Browser.getVersion","params":{}}`)); err != nil {
		t.Fatal(err)
	}
	body, err := wsutil.ReadServerText(rawConn)
	if err != nil {
		t.Fatal(err)
	}
	var response struct {
		ID     int `json:"id"`
		Result struct {
			Product string `json:"product"`
		} `json:"result"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatal(err)
	}
	if response.ID != 1 {
		t.Fatalf("unexpected response id %d", response.ID)
	}
	if !strings.Contains(response.Result.Product, "Chrome") && !strings.Contains(response.Result.Product, "Chromium") {
		t.Fatalf("unexpected product %q", response.Result.Product)
	}
}

type closeOrderTransport struct {
	order *[]string
}

func (t closeOrderTransport) Update(map[string]any)            {}
func (t closeOrderTransport) Connect() error                   { return nil }
func (t closeOrderTransport) Close() error                     { *t.order = append(*t.order, "transport"); return nil }
func (t closeOrderTransport) Send(map[string]any) error        { return nil }
func (t closeOrderTransport) GetLauncherConfig() LaunchOptions { return LaunchOptions{} }
func (t closeOrderTransport) GetInjectorConfig() ExtensionInjectorConfig {
	return ExtensionInjectorConfig{}
}
func (t closeOrderTransport) GetServerConfig() map[string]any    { return map[string]any{} }
func (t closeOrderTransport) OnRecv(func(map[string]any)) func() { return func() {} }
func (t closeOrderTransport) OnClose(func(error)) func()         { return func() {} }
func (t closeOrderTransport) WaitForPeer() error                 { return nil }

type closeOrderInjector struct {
	order *[]string
}

func (i *closeOrderInjector) Update(ExtensionInjectorConfig) *ExtensionInjector {
	return &ExtensionInjector{}
}
func (i *closeOrderInjector) GetLauncherConfig() LaunchOptions           { return LaunchOptions{} }
func (i *closeOrderInjector) GetTransportConfig() map[string]any         { return map[string]any{} }
func (i *closeOrderInjector) Prepare() error                             { return nil }
func (i *closeOrderInjector) Inject() (*ExtensionInjectionResult, error) { return nil, nil }
func (i *closeOrderInjector) Close() error                               { *i.order = append(*i.order, "injector"); return nil }

func TestModCDPClientCloseKeepsInjectorFilesUntilAfterLaunchedBrowserShutdown(t *testing.T) {
	order := []string{}
	cdp := New(Options{})
	cdp.transport = closeOrderTransport{order: &order}
	cdp.launchedBrowser = &LaunchedBrowser{
		Close: func() {
			order = append(order, "browser")
		},
	}
	cdp.extensionInjectors = []extensionInjector{&closeOrderInjector{order: &order}}

	cdp.Close()

	if strings.Join(order, ",") != "transport,browser,injector" {
		t.Fatalf("close order = %#v", order)
	}
	if cdp.transport != nil {
		t.Fatal("expected transport to be nil")
	}
	if cdp.launchedBrowser != nil {
		t.Fatal("expected launchedBrowser to be nil")
	}
	if cdp.extensionInjectors != nil {
		t.Fatal("expected extensionInjectors to be nil")
	}
}

func TestCustomCommandSchemasValidateParamsAndResults(t *testing.T) {
	cdp := New(Options{
		CustomCommands: []CustomCommand{
			{
				Name: "Custom.echo",
				ParamsSchema: map[string]any{
					"type":                 "object",
					"required":             []any{"value"},
					"properties":           map[string]any{"value": map[string]any{"type": "string"}},
					"additionalProperties": false,
				},
				ResultSchema: map[string]any{
					"type":                 "object",
					"required":             []any{"value"},
					"properties":           map[string]any{"value": map[string]any{"type": "string"}},
					"additionalProperties": false,
				},
			},
		},
	})

	if err := cdp.validateCommandParams("Custom.echo", map[string]any{"value": "ok"}); err != nil {
		t.Fatalf("expected valid params, got %v", err)
	}
	if err := cdp.validateCommandParams("Custom.echo", map[string]any{"value": 42}); err == nil || !strings.Contains(err.Error(), "params_schema") {
		t.Fatalf("expected params schema error, got %v", err)
	}
	if err := cdp.validateCommandResult("Custom.echo", map[string]any{"value": "ok"}); err != nil {
		t.Fatalf("expected valid result, got %v", err)
	}
	if err := cdp.validateCommandResult("Custom.echo", map[string]any{"value": 42}); err == nil || !strings.Contains(err.Error(), "result_schema") {
		t.Fatalf("expected result schema error, got %v", err)
	}
}

func TestCustomEventSchemasValidatePayloads(t *testing.T) {
	cdp := New(Options{
		CustomEvents: []CustomEvent{
			{
				Name: "Custom.changed",
				EventSchema: map[string]any{
					"type":                 "object",
					"required":             []any{"targetId"},
					"properties":           map[string]any{"targetId": map[string]any{"type": "string"}},
					"additionalProperties": false,
				},
			},
		},
	})

	if _, ok := cdp.validateEventData("Custom.changed", map[string]any{"targetId": "target-1"}); !ok {
		t.Fatal("expected valid event payload")
	}
	if _, ok := cdp.validateEventData("Custom.changed", map[string]any{"targetId": 1}); ok {
		t.Fatal("expected invalid event payload")
	}
}

func TestTypedCDPSurfaceInitializesAndEncodesParams(t *testing.T) {
	cdp := New(Options{})
	if cdp.Target.client != cdp {
		t.Fatal("expected Target domain to be initialized with the client")
	}

	params := TargetCreateTargetParams{
		URL:        "https://example.com",
		Background: Bool(true),
	}
	raw, err := cdpParamsMap(params)
	if err != nil {
		t.Fatal(err)
	}
	if raw["url"] != "https://example.com" || raw["background"] != true {
		t.Fatalf("unexpected encoded Target.createTarget params: %#v", raw)
	}
	if _, ok := raw["sessionId"]; ok {
		t.Fatalf("SessionID must stay transport-only, got %#v", raw)
	}
}
