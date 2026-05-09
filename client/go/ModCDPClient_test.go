package modcdp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
			NativeMessagingHostName:       "com.modcdp.custom",
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
			HydrateAliases:       boolPtr(false),
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
	if cdp.opts.Upstream.NativeMessagingHostName != "com.modcdp.custom" {
		t.Fatalf("Upstream.NativeMessagingHostName = %q", cdp.opts.Upstream.NativeMessagingHostName)
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
	if cdp.opts.Client.HydrateAliases == nil || *cdp.opts.Client.HydrateAliases {
		t.Fatalf("Client.HydrateAliases = %#v", cdp.opts.Client.HydrateAliases)
	}
	if _, err := cdp.Browser.GetVersion(); err == nil || !strings.Contains(err.Error(), "client.hydrate_aliases is false") {
		t.Fatalf("Browser.GetVersion with aliases disabled error = %v", err)
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

func TestModCDPClientOptionsMarshalToSnakeCaseConfigShape(t *testing.T) {
	encoded, err := json.Marshal(Options{
		Launch: LaunchConfig{
			Mode:           "local",
			ExecutablePath: "/tmp/chrome",
			UserDataDir:    "/tmp/profile",
			Options: LaunchOptions{
				RemoteDebugging:                "pipe",
				ChromeReadyTimeoutMS:           45_000,
				BrowserbaseAPIKey:              "test-key",
				BrowserbaseSessionCreateParams: map[string]any{"projectId": "project-1"},
			},
		},
		Upstream: UpstreamConfig{
			Mode:                          "nativemessaging",
			NATSSubjectPrefix:             "modcdp.test",
			ReverseWSWaitTimeoutMS:        1_234,
			NativeMessagingManifest:       "/tmp/native.json",
			NativeMessagingHostName:       "com.modcdp.custom",
			WSConnectErrorSettleTimeoutMS: 321,
		},
		Extension: ExtensionConfig{
			Mode:                         "discover",
			ExtensionID:                  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			ServiceWorkerURLSuffixes:     []string{"/modcdp/service_worker.js"},
			TrustServiceWorkerTarget:     true,
			RequireServiceWorkerTarget:   true,
			ServiceWorkerReadyExpression: "Boolean(globalThis.ModCDP)",
			ExecutionContextTimeoutMS:    4_321,
		},
		Client: ClientConfig{
			Routes:               map[string]string{"*.*": "service_worker"},
			HydrateAliases:       boolPtr(false),
			MirrorUpstreamEvents: boolPtr(false),
			CDPSendTimeoutMS:     987,
		},
		Server: &ServerConfig{
			LoopbackCDPURL: "http://127.0.0.1:9222",
		},
		CustomCommands: []CustomCommand{{Name: "Custom.echo", Expression: "async () => null"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw := string(encoded)
	for _, wrong := range []string{
		"Launch", "ExecutablePath", "RemoteDebugging", "BrowserbaseAPIKey",
		"Upstream", "NATSSubjectPrefix", "ReverseWSWaitTimeoutMS", "NativeMessagingHostName",
		"Extension", "ServiceWorkerURLSuffixes", "TrustServiceWorkerTarget",
		"Client", "HydrateAliases", "CustomCommands",
	} {
		if strings.Contains(raw, wrong) {
			t.Fatalf("encoded options leaked Go field name %q in %s", wrong, raw)
		}
	}
	for _, expected := range []string{
		`"launch"`,
		`"executable_path"`,
		`"remote_debugging"`,
		`"browserbase_api_key"`,
		`"browserbase_session_create_params"`,
		`"upstream"`,
		`"nats_subject_prefix"`,
		`"reversews_wait_timeout_ms"`,
		`"nativemessaging_manifest"`,
		`"nativemessaging_host_name"`,
		`"extension"`,
		`"service_worker_url_suffixes"`,
		`"trust_service_worker_target"`,
		`"require_service_worker_target"`,
		`"service_worker_ready_expression"`,
		`"execution_context_timeout_ms"`,
		`"client"`,
		`"hydrate_aliases"`,
		`"mirror_upstream_events"`,
		`"cdp_send_timeout_ms"`,
		`"custom_commands"`,
	} {
		if !strings.Contains(raw, expected) {
			t.Fatalf("encoded options missing %s in %s", expected, raw)
		}
	}
}

func TestModCDPClientPreservesExplicitEmptyServiceWorkerSuffixConfig(t *testing.T) {
	cdp := New(Options{
		Extension: ExtensionConfig{
			Mode:                     "borrow",
			ServiceWorkerURLSuffixes: []string{},
		},
	})

	if len(cdp.opts.Extension.ServiceWorkerURLSuffixes) != 0 {
		t.Fatalf("ServiceWorkerURLSuffixes = %#v", cdp.opts.Extension.ServiceWorkerURLSuffixes)
	}
	injectorConfig := cdp.baseExtensionInjectorConfig(nil)
	if len(injectorConfig.ServiceWorkerURLSuffixes) != 0 {
		t.Fatalf("injector ServiceWorkerURLSuffixes = %#v", injectorConfig.ServiceWorkerURLSuffixes)
	}
}

func TestModCDPClientDefaultsServiceWorkerSuffixConfigToModCDPWorker(t *testing.T) {
	cdp := New(Options{})

	if len(cdp.opts.Extension.ServiceWorkerURLSuffixes) != 1 || cdp.opts.Extension.ServiceWorkerURLSuffixes[0] != "/modcdp/service_worker.js" {
		t.Fatalf("ServiceWorkerURLSuffixes = %#v", cdp.opts.Extension.ServiceWorkerURLSuffixes)
	}
	injectorConfig := cdp.baseExtensionInjectorConfig(nil)
	if len(injectorConfig.ServiceWorkerURLSuffixes) != 1 || injectorConfig.ServiceWorkerURLSuffixes[0] != "/modcdp/service_worker.js" {
		t.Fatalf("injector ServiceWorkerURLSuffixes = %#v", injectorConfig.ServiceWorkerURLSuffixes)
	}
}

func TestModCDPClientDefaultsLaunchedModCDPServerUpstreamsToExtensionAuto(t *testing.T) {
	for _, mode := range []string{"nativemessaging", "reversews", "nats"} {
		launched := New(Options{
			Launch:   LaunchConfig{Mode: "local"},
			Upstream: UpstreamConfig{Mode: mode},
		})
		if launched.opts.Launch.Mode != "local" {
			t.Fatalf("%s launched Launch.Mode = %q", mode, launched.opts.Launch.Mode)
		}
		if endpointKindForUpstream(launched.opts.Upstream.Mode) != UpstreamEndpointKindModCDPServer {
			t.Fatalf("%s launched endpoint kind = %q", mode, endpointKindForUpstream(launched.opts.Upstream.Mode))
		}
		if launched.opts.Extension.Mode != "auto" {
			t.Fatalf("%s launched Extension.Mode = %q", mode, launched.opts.Extension.Mode)
		}

		attachOnly := New(Options{
			Upstream: UpstreamConfig{Mode: mode},
		})
		if attachOnly.opts.Launch.Mode != "none" {
			t.Fatalf("%s attach-only Launch.Mode = %q", mode, attachOnly.opts.Launch.Mode)
		}
		if endpointKindForUpstream(attachOnly.opts.Upstream.Mode) != UpstreamEndpointKindModCDPServer {
			t.Fatalf("%s attach-only endpoint kind = %q", mode, endpointKindForUpstream(attachOnly.opts.Upstream.Mode))
		}
		if attachOnly.opts.Extension.Mode != "none" {
			t.Fatalf("%s attach-only Extension.Mode = %q", mode, attachOnly.opts.Extension.Mode)
		}
	}
}

func TestModCDPClientRejectsUnknownComponentModesAtTheirOwningFactoryBoundary(t *testing.T) {
	cases := []struct {
		name string
		cdp  *ModCDPClient
		want string
	}{
		{
			name: "upstream",
			cdp:  New(Options{Upstream: UpstreamConfig{Mode: "bogus"}}),
			want: "unknown upstream.mode=bogus",
		},
		{
			name: "launch",
			cdp: New(Options{
				Launch:   LaunchConfig{Mode: "bogus"},
				Upstream: UpstreamConfig{Mode: "ws", WSURL: "ws://127.0.0.1:1/devtools/browser/test"},
			}),
			want: "unknown launch.mode=bogus",
		},
		{
			name: "extension",
			cdp: New(Options{
				Launch:    LaunchConfig{Mode: "none"},
				Upstream:  UpstreamConfig{Mode: "ws", WSURL: "ws://127.0.0.1:1/devtools/browser/test"},
				Extension: ExtensionConfig{Mode: "bogus"},
			}),
			want: "unknown extension.mode=bogus",
		},
	}
	for _, testCase := range cases {
		if err := testCase.cdp.Connect(); err == nil || !strings.Contains(err.Error(), testCase.want) {
			t.Fatalf("%s Connect error = %v", testCase.name, err)
		}
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
		Client: ClientConfig{
			Routes: map[string]string{"Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp"},
		},
		Server: &ServerConfig{
			Routes: map[string]string{"*.*": "loopback_cdp"},
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
	result, err := cdp.Mod.Evaluate(map[string]any{
		"expression": "chrome.runtime.getURL('modcdp/service_worker.js')",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result != "chrome-extension://mdedooklbnfejodmnhmkdpkaedafkehf/modcdp/service_worker.js" {
		t.Fatalf("Mod.evaluate = %#v", result)
	}
	directTargetRaw, err := cdp.Send("Target.createTarget", map[string]any{"url": "about:blank#direct-session-routing"})
	if err != nil {
		t.Fatal(err)
	}
	directTarget, _ := directTargetRaw.(map[string]any)
	directTargetID, _ := directTarget["targetId"].(string)
	if directTargetID == "" {
		t.Fatalf("Target.createTarget = %#v", directTargetRaw)
	}
	defer cdp.Send("Target.closeTarget", map[string]any{"targetId": directTargetID})
	directSessionRaw, err := cdp.Send("Target.attachToTarget", map[string]any{"targetId": directTargetID, "flatten": true})
	if err != nil {
		t.Fatal(err)
	}
	directSession, _ := directSessionRaw.(map[string]any)
	directSessionID, _ := directSession["sessionId"].(string)
	if directSessionID == "" {
		t.Fatalf("Target.attachToTarget = %#v", directSessionRaw)
	}
	directEvalRaw, err := cdp.Send("Runtime.evaluate", map[string]any{"expression": "1 + 1", "returnByValue": true}, directSessionID)
	if err != nil {
		t.Fatal(err)
	}
	directEval, _ := directEvalRaw.(map[string]any)
	directRemoteObject, _ := directEval["result"].(map[string]any)
	if directRemoteObject["value"] != float64(2) {
		t.Fatalf("Runtime.evaluate = %#v", directEvalRaw)
	}
	sent_at := time.Now().UnixMilli()
	pong := make(chan map[string]any, 1)
	muted := make(chan any, 1)
	mutedHandler := func(payload any) {
		muted <- payload
	}
	cdp.On("Mod.pong", mutedHandler)
	cdp.Off("Mod.pong", mutedHandler)
	cdp.Once("Mod.pong", func(payload any) {
		event, _ := payload.(map[string]any)
		if event == nil {
			return
		}
		if event["sent_at"] == float64(sent_at) || event["sent_at"] == sent_at {
			pong <- event
		}
	})
	ping_raw, err := cdp.Mod.Ping(map[string]any{"sent_at": sent_at})
	if err != nil {
		t.Fatal(err)
	}
	ping_result, _ := ping_raw.(map[string]any)
	if ping_result["ok"] != true {
		t.Fatalf("Mod.ping = %#v", ping_result)
	}
	select {
	case pong_payload := <-pong:
		if pong_payload["from"] != "extension-service-worker" {
			t.Fatalf("Mod.pong from = %#v", pong_payload["from"])
		}
		if _, ok := numberAsInt64(pong_payload["received_at"]); !ok {
			t.Fatalf("Mod.pong received_at = %#v", pong_payload["received_at"])
		}
		select {
		case payload := <-muted:
			t.Fatalf("off handler received payload %#v", payload)
		case <-time.After(200 * time.Millisecond):
		}
		if _, err := cdp.Mod.Ping(map[string]any{"sent_at": sent_at + 1}); err != nil {
			t.Fatal(err)
		}
		select {
		case event := <-pong:
			t.Fatalf("once handler received second event %#v", event)
		case <-time.After(200 * time.Millisecond):
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for Mod.pong")
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

func TestModCDPClientCloseKeepsInjectorFilesUntilAfterLaunchedBrowserShutdown(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	reversePort, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	cdp := New(Options{
		Launch: LaunchConfig{
			Mode: "local",
			Options: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{Mode: "reversews", ReverseWSBind: "127.0.0.1:" + fmt.Sprint(reversePort)},
		Extension: ExtensionConfig{
			Mode:                     "auto",
			Path:                     extensionPath,
			ServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			TrustServiceWorkerTarget: true,
		},
		Server: &ServerConfig{Routes: map[string]string{"*.*": "loopback_cdp"}},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	var localInjector *LocalBrowserLaunchExtensionInjector
	for _, injector := range cdp.extensionInjectors {
		if typed, ok := injector.(*LocalBrowserLaunchExtensionInjector); ok {
			localInjector = typed
		}
	}
	if localInjector == nil {
		t.Fatal("expected LocalBrowserLaunchExtensionInjector")
	}
	unpackedExtensionPath := localInjector.UnpackedExtensionPath
	if unpackedExtensionPath == "" || unpackedExtensionPath == extensionPath {
		t.Fatalf("UnpackedExtensionPath = %q", unpackedExtensionPath)
	}
	if _, err := os.Stat(filepath.Join(unpackedExtensionPath, "config.js")); err != nil {
		t.Fatalf("expected runtime config.js before close: %v", err)
	}

	originalClose := cdp.launchedBrowser.Close
	browserCloseSawExtension := false
	cdp.launchedBrowser.Close = func() {
		_, err := os.Stat(unpackedExtensionPath)
		browserCloseSawExtension = err == nil
		originalClose()
	}
	cdp.Close()

	if !browserCloseSawExtension {
		t.Fatal("browser close did not see prepared extension files")
	}
	if _, err := os.Stat(unpackedExtensionPath); !os.IsNotExist(err) {
		t.Fatalf("expected prepared extension files to be removed after close, got %v", err)
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

func TestModCDPClientCloseClearsTopLevelConnectionState(t *testing.T) {
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
			Mode:                     "auto",
			ServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			TrustServiceWorkerTarget: true,
		},
	})
	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	transport, ok := cdp.transport.(*WebSocketUpstreamTransport)
	if !ok {
		t.Fatalf("transport = %T", cdp.transport)
	}
	if transport.Conn == nil {
		t.Fatal("expected transport-owned websocket conn")
	}
	cdp.Close()
	if cdp.transport != nil {
		t.Fatal("Close left transport set")
	}
	if _, err := cdp.SendRaw("Browser.getVersion", map[string]any{}); err == nil || !strings.Contains(err.Error(), "ModCDP upstream is not connected") {
		t.Fatalf("SendRaw after close error = %v", err)
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
