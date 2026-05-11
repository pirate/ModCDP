package client

import (
	"context"
	"encoding/json"
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
		Launcher: LauncherConfig{
			LauncherMode:           "local",
			LauncherExecutablePath: "/tmp/chrome",
			LauncherUserDataDir:    "/tmp/profile",
			LauncherOptions: LaunchOptions{
				Headless: boolPtr(true),
			},
		},
		Upstream: UpstreamConfig{
			UpstreamMode:                          "ws",
			UpstreamCDPURL:                        "http://127.0.0.1:9222",
			UpstreamNATSWaitTimeoutMS:             345,
			UpstreamReverseWSWaitTimeoutMS:        456,
			UpstreamNativeMessagingHostName:       "com.modcdp.custom",
			UpstreamNativeMessagingWaitTimeoutMS:  567,
			UpstreamWSConnectErrorSettleTimeoutMS: 321,
		},
		Injector: InjectorConfig{
			InjectorMode:                        "discover",
			InjectorExtensionPath:               "/tmp/ext",
			InjectorExtensionID:                 "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			InjectorServiceWorkerURLIncludes:    []string{"modcdp"},
			InjectorServiceWorkerURLSuffixes:    []string{"/custom/service_worker.js"},
			InjectorTrustServiceWorkerTarget:    true,
			InjectorRequireServiceWorkerTarget:  true,
			InjectorExecutionContextTimeoutMS:   4321,
			InjectorServiceWorkerProbeTimeoutMS: 5432,
			InjectorServiceWorkerReadyTimeoutMS: 6543,
			InjectorServiceWorkerPollIntervalMS: 76,
			InjectorTargetSessionPollIntervalMS: 87,
		},
		Client: ClientConfig{
			ClientRoutes:               map[string]string{"*.*": "direct_cdp"},
			ClientHydrateAliases:       boolPtr(false),
			ClientMirrorUpstreamEvents: boolPtr(false),
			ClientCDPSendTimeoutMS:     1234,
			ClientEventWaitTimeoutMS:   2345,
		},
		Server: &ServerConfig{
			ServerRoutes:                            map[string]string{"*.*": "loopback_cdp"},
			ServerBrowserToken:                      "token-1",
			ServerCDPSendTimeoutMS:                  9876,
			ServerLoopbackExecutionContextTimeoutMS: 8765,
			ServerWSConnectErrorSettleTimeoutMS:     7654,
		},
	})

	if cdp.Launcher.LauncherOptions.ExecutablePath != "/tmp/chrome" {
		t.Fatalf("Launcher.LauncherOptions.ExecutablePath = %q", cdp.Launcher.LauncherOptions.ExecutablePath)
	}
	if cdp.Launcher.LauncherOptions.UserDataDir != "/tmp/profile" {
		t.Fatalf("Launcher.LauncherOptions.UserDataDir = %q", cdp.Launcher.LauncherOptions.UserDataDir)
	}
	if cdp.Upstream.UpstreamWSConnectErrorSettleTimeoutMS != 321 {
		t.Fatalf("Upstream.UpstreamWSConnectErrorSettleTimeoutMS = %d", cdp.Upstream.UpstreamWSConnectErrorSettleTimeoutMS)
	}
	if cdp.Upstream.UpstreamReverseWSWaitTimeoutMS != 456 {
		t.Fatalf("Upstream.UpstreamReverseWSWaitTimeoutMS = %d", cdp.Upstream.UpstreamReverseWSWaitTimeoutMS)
	}
	if cdp.Upstream.UpstreamNATSWaitTimeoutMS != 345 {
		t.Fatalf("Upstream.UpstreamNATSWaitTimeoutMS = %d", cdp.Upstream.UpstreamNATSWaitTimeoutMS)
	}
	if cdp.Upstream.UpstreamNativeMessagingHostName != "com.modcdp.custom" {
		t.Fatalf("Upstream.UpstreamNativeMessagingHostName = %q", cdp.Upstream.UpstreamNativeMessagingHostName)
	}
	if cdp.Upstream.UpstreamNativeMessagingWaitTimeoutMS != 567 {
		t.Fatalf("Upstream.UpstreamNativeMessagingWaitTimeoutMS = %d", cdp.Upstream.UpstreamNativeMessagingWaitTimeoutMS)
	}
	if cdp.Injector.InjectorExecutionContextTimeoutMS != 4321 {
		t.Fatalf("Injector.InjectorExecutionContextTimeoutMS = %d", cdp.Injector.InjectorExecutionContextTimeoutMS)
	}
	if cdp.Injector.InjectorServiceWorkerProbeTimeoutMS != 5432 {
		t.Fatalf("Injector.InjectorServiceWorkerProbeTimeoutMS = %d", cdp.Injector.InjectorServiceWorkerProbeTimeoutMS)
	}
	if cdp.Injector.InjectorServiceWorkerReadyTimeoutMS != 6543 {
		t.Fatalf("Injector.InjectorServiceWorkerReadyTimeoutMS = %d", cdp.Injector.InjectorServiceWorkerReadyTimeoutMS)
	}
	if cdp.Injector.InjectorServiceWorkerPollIntervalMS != 76 {
		t.Fatalf("Injector.InjectorServiceWorkerPollIntervalMS = %d", cdp.Injector.InjectorServiceWorkerPollIntervalMS)
	}
	if cdp.Injector.InjectorTargetSessionPollIntervalMS != 87 {
		t.Fatalf("Injector.InjectorTargetSessionPollIntervalMS = %d", cdp.Injector.InjectorTargetSessionPollIntervalMS)
	}
	if cdp.Client.ClientRoutes["*.*"] != "direct_cdp" {
		t.Fatalf("Client.ClientRoutes[*.*] = %q", cdp.Client.ClientRoutes["*.*"])
	}
	if cdp.Client.ClientHydrateAliases == nil || *cdp.Client.ClientHydrateAliases {
		t.Fatalf("Client.ClientHydrateAliases = %#v", cdp.Client.ClientHydrateAliases)
	}
	if _, err := cdp.Browser.GetVersion(); err == nil || !strings.Contains(err.Error(), "client.hydrate_aliases is false") {
		t.Fatalf("Browser.GetVersion with aliases disabled error = %v", err)
	}
	if cdp.Client.ClientMirrorUpstreamEvents == nil || *cdp.Client.ClientMirrorUpstreamEvents {
		t.Fatalf("Client.ClientMirrorUpstreamEvents = %#v", cdp.Client.ClientMirrorUpstreamEvents)
	}
	if cdp.Client.ClientCDPSendTimeoutMS != 1234 {
		t.Fatalf("Client.ClientCDPSendTimeoutMS = %d", cdp.Client.ClientCDPSendTimeoutMS)
	}
	if cdp.Client.ClientEventWaitTimeoutMS != 2345 {
		t.Fatalf("Client.ClientEventWaitTimeoutMS = %d", cdp.Client.ClientEventWaitTimeoutMS)
	}
	if cdp.UpstreamEndpointKind != UpstreamEndpointKindRawCDP {
		t.Fatalf("UpstreamEndpointKind = %q", cdp.UpstreamEndpointKind)
	}

	params := cdp.serverConfigureParams(nil, nil, nil)
	clientConfig := params["client"].(map[string]any)
	routes := clientConfig["client_routes"].(map[string]string)
	serverConfig := params["server"].(map[string]any)
	if routes["*.*"] != "direct_cdp" {
		t.Fatalf("configure client routes = %#v", routes)
	}
	if serverConfig["server_browser_token"] != "token-1" {
		t.Fatalf("configure browser_token = %#v", serverConfig["server_browser_token"])
	}
	if serverConfig["server_cdp_send_timeout_ms"] != 9876 {
		t.Fatalf("configure cdp_send_timeout_ms = %#v", serverConfig["server_cdp_send_timeout_ms"])
	}
	if serverConfig["server_loopback_execution_context_timeout_ms"] != 8765 {
		t.Fatalf("configure loopback_execution_context_timeout_ms = %#v", serverConfig["server_loopback_execution_context_timeout_ms"])
	}
	if serverConfig["server_ws_connect_error_settle_timeout_ms"] != 7654 {
		t.Fatalf("configure ws_connect_error_settle_timeout_ms = %#v", serverConfig["server_ws_connect_error_settle_timeout_ms"])
	}
}

func TestModCDPClientDispatchesRootEventsBeforeExtensionSessionAttached(t *testing.T) {
	cdp := New(Options{})
	seen := make(chan string, 1)
	cdp.On("Target.targetCreated", func(payload any) {
		event, _ := payload.(map[string]any)
		targetInfo, _ := event["targetInfo"].(map[string]any)
		targetID, _ := targetInfo["targetId"].(string)
		seen <- targetID
	})

	cdp.handleEventMessage(map[string]any{
		"method": "Target.targetCreated",
		"params": map[string]any{
			"targetInfo": map[string]any{"targetId": "target-1", "type": "page", "url": "about:blank"},
		},
	})

	select {
	case got := <-seen:
		if got != "target-1" {
			t.Fatalf("Target.targetCreated targetId = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for root event")
	}
}

func TestModCDPClientEventDispatchSnapshotsHandlersWhenOnceRemovesItself(t *testing.T) {
	cdp := New(Options{})
	cdp.ExtSessionID = "ext-session"
	seen := make(chan string, 3)
	cdp.Once("Target.targetCreated", func(payload any) {
		seen <- "once"
	})
	cdp.On("Target.targetCreated", func(payload any) {
		seen <- "persistent"
	})

	cdp.handleEventMessage(map[string]any{
		"method": "Target.targetCreated",
		"params": map[string]any{
			"targetInfo": map[string]any{"targetId": "target-1", "type": "page", "url": "about:blank"},
		},
	})

	first := map[string]bool{}
	for len(first) < 2 {
		select {
		case got := <-seen:
			first[got] = true
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for first dispatch, saw %#v", first)
		}
	}
	if !first["once"] || !first["persistent"] {
		t.Fatalf("first dispatch handlers = %#v", first)
	}

	cdp.handleEventMessage(map[string]any{
		"method": "Target.targetCreated",
		"params": map[string]any{
			"targetInfo": map[string]any{"targetId": "target-2", "type": "page", "url": "about:blank"},
		},
	})

	select {
	case got := <-seen:
		if got != "persistent" {
			t.Fatalf("second dispatch handler = %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for second dispatch")
	}
	select {
	case got := <-seen:
		t.Fatalf("unexpected extra handler on second dispatch: %q", got)
	case <-time.After(100 * time.Millisecond):
	}
}

func TestModCDPClientOptionsMarshalToSnakeCaseConfigShape(t *testing.T) {
	encoded, err := json.Marshal(Options{
		Launcher: LauncherConfig{
			LauncherMode:           "local",
			LauncherExecutablePath: "/tmp/chrome",
			LauncherUserDataDir:    "/tmp/profile",
			LauncherOptions: LaunchOptions{
				RemoteDebugging:                "pipe",
				ChromeReadyTimeoutMS:           45_000,
				BrowserbaseAPIKey:              "test-key",
				BrowserbaseSessionCreateParams: map[string]any{"keepAlive": true},
			},
		},
		Upstream: UpstreamConfig{
			UpstreamMode:                          "nativemessaging",
			UpstreamNATSSubjectPrefix:             "modcdp.test",
			UpstreamNATSWaitTimeoutMS:             789,
			UpstreamReverseWSWaitTimeoutMS:        1_234,
			UpstreamNativeMessagingManifest:       "/tmp/native.json",
			UpstreamNativeMessagingHostName:       "com.modcdp.custom",
			UpstreamNativeMessagingWaitTimeoutMS:  2_345,
			UpstreamWSConnectErrorSettleTimeoutMS: 321,
		},
		Injector: InjectorConfig{
			InjectorMode:                         "discover",
			InjectorExtensionID:                  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			InjectorServiceWorkerURLSuffixes:     []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget:     true,
			InjectorRequireServiceWorkerTarget:   true,
			InjectorServiceWorkerReadyExpression: "Boolean(globalThis.ModCDP)",
			InjectorExecutionContextTimeoutMS:    4_321,
		},
		Client: ClientConfig{
			ClientRoutes:               map[string]string{"*.*": "service_worker"},
			ClientHydrateAliases:       boolPtr(false),
			ClientMirrorUpstreamEvents: boolPtr(false),
			ClientCDPSendTimeoutMS:     987,
		},
		Server: &ServerConfig{
			ServerLoopbackCDPURL: "http://127.0.0.1:9222",
		},
		CustomCommands: []CustomCommand{{Name: "Custom.echo", Expression: "async () => null"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw := string(encoded)
	for _, wrong := range []string{
		"Launcher", "ExecutablePath", "RemoteDebugging", "BrowserbaseAPIKey",
		"Upstream", "UpstreamNATSSubjectPrefix", "UpstreamNATSWaitTimeoutMS", "UpstreamReverseWSWaitTimeoutMS", "UpstreamNativeMessagingHostName", "UpstreamNativeMessagingWaitTimeoutMS",
		"Injector", "ServiceWorkerURLSuffixes", "TrustServiceWorkerTarget",
		"Client", "HydrateAliases", "CustomCommands",
	} {
		if strings.Contains(raw, wrong) {
			t.Fatalf("encoded options leaked Go field name %q in %s", wrong, raw)
		}
	}
	for _, expected := range []string{
		`"launcher"`,
		`"launcher_mode"`,
		`"launcher_executable_path"`,
		`"launcher_user_data_dir"`,
		`"launcher_options"`,
		`"remote_debugging"`,
		`"browserbase_api_key"`,
		`"browserbase_session_create_params"`,
		`"upstream"`,
		`"upstream_mode"`,
		`"upstream_nats_subject_prefix"`,
		`"upstream_nats_wait_timeout_ms"`,
		`"upstream_reversews_wait_timeout_ms"`,
		`"upstream_nativemessaging_manifest"`,
		`"upstream_nativemessaging_host_name"`,
		`"upstream_nativemessaging_wait_timeout_ms"`,
		`"injector"`,
		`"injector_mode"`,
		`"injector_service_worker_url_suffixes"`,
		`"injector_trust_service_worker_target"`,
		`"injector_require_service_worker_target"`,
		`"injector_service_worker_ready_expression"`,
		`"injector_execution_context_timeout_ms"`,
		`"client"`,
		`"client_hydrate_aliases"`,
		`"client_mirror_upstream_events"`,
		`"client_cdp_send_timeout_ms"`,
		`"custom_commands"`,
	} {
		if !strings.Contains(raw, expected) {
			t.Fatalf("encoded options missing %s in %s", expected, raw)
		}
	}
}

func TestModCDPClientOptionsUnmarshalNullServerDisablesServerConfig(t *testing.T) {
	var options Options
	if err := json.Unmarshal([]byte(`{"server": null}`), &options); err != nil {
		t.Fatal(err)
	}
	cdp := New(options)

	if cdp.Server != nil {
		t.Fatalf("Server = %#v", cdp.Server)
	}
}

func TestModCDPClientPreservesExplicitEmptyServiceWorkerSuffixConfig(t *testing.T) {
	cdp := New(Options{
		Injector: InjectorConfig{
			InjectorMode:                     "borrow",
			InjectorServiceWorkerURLSuffixes: []string{},
		},
	})

	if len(cdp.Injector.InjectorServiceWorkerURLSuffixes) != 0 {
		t.Fatalf("InjectorServiceWorkerURLSuffixes = %#v", cdp.Injector.InjectorServiceWorkerURLSuffixes)
	}
	injectorConfig := cdp.baseExtensionInjectorConfig(nil)
	if len(injectorConfig.ServiceWorkerURLSuffixes) != 0 {
		t.Fatalf("injector ServiceWorkerURLSuffixes = %#v", injectorConfig.ServiceWorkerURLSuffixes)
	}
}

func TestModCDPClientPreservesExplicitNoneServerConfig(t *testing.T) {
	cdp := New(Options{Server: ServerNone})

	if cdp.Server != nil {
		t.Fatalf("Server = %#v", cdp.Server)
	}
}

func TestModCDPClientDefaultsServiceWorkerSuffixConfigToModCDPWorker(t *testing.T) {
	cdp := New(Options{})

	if len(cdp.Injector.InjectorServiceWorkerURLSuffixes) != 1 || cdp.Injector.InjectorServiceWorkerURLSuffixes[0] != "/modcdp/service_worker.js" {
		t.Fatalf("InjectorServiceWorkerURLSuffixes = %#v", cdp.Injector.InjectorServiceWorkerURLSuffixes)
	}
	injectorConfig := cdp.baseExtensionInjectorConfig(nil)
	if len(injectorConfig.ServiceWorkerURLSuffixes) != 1 || injectorConfig.ServiceWorkerURLSuffixes[0] != "/modcdp/service_worker.js" {
		t.Fatalf("injector ServiceWorkerURLSuffixes = %#v", injectorConfig.ServiceWorkerURLSuffixes)
	}
}

func TestModCDPClientDefaultsLaunchedModCDPServerUpstreamsToExtensionAuto(t *testing.T) {
	for _, mode := range []string{"nativemessaging", "reversews", "nats"} {
		launched := New(Options{
			Launcher: LauncherConfig{LauncherMode: "local"},
			Upstream: UpstreamConfig{UpstreamMode: mode},
		})
		if launched.Launcher.LauncherMode != "local" {
			t.Fatalf("%s launched Launcher.LauncherMode = %q", mode, launched.Launcher.LauncherMode)
		}
		if endpointKindForUpstream(launched.Upstream.UpstreamMode) != UpstreamEndpointKindModCDPServer {
			t.Fatalf("%s launched endpoint kind = %q", mode, endpointKindForUpstream(launched.Upstream.UpstreamMode))
		}
		if launched.UpstreamEndpointKind != UpstreamEndpointKindModCDPServer {
			t.Fatalf("%s launched UpstreamEndpointKind = %q", mode, launched.UpstreamEndpointKind)
		}
		if launched.Injector.InjectorMode != "auto" {
			t.Fatalf("%s launched Injector.InjectorMode = %q", mode, launched.Injector.InjectorMode)
		}

		attachOnly := New(Options{
			Upstream: UpstreamConfig{UpstreamMode: mode},
		})
		if attachOnly.Launcher.LauncherMode != "none" {
			t.Fatalf("%s attach-only Launcher.LauncherMode = %q", mode, attachOnly.Launcher.LauncherMode)
		}
		if endpointKindForUpstream(attachOnly.Upstream.UpstreamMode) != UpstreamEndpointKindModCDPServer {
			t.Fatalf("%s attach-only endpoint kind = %q", mode, endpointKindForUpstream(attachOnly.Upstream.UpstreamMode))
		}
		if attachOnly.UpstreamEndpointKind != UpstreamEndpointKindModCDPServer {
			t.Fatalf("%s attach-only UpstreamEndpointKind = %q", mode, attachOnly.UpstreamEndpointKind)
		}
		if attachOnly.Injector.InjectorMode != "none" {
			t.Fatalf("%s attach-only Injector.InjectorMode = %q", mode, attachOnly.Injector.InjectorMode)
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
			cdp:  New(Options{Upstream: UpstreamConfig{UpstreamMode: "bogus"}}),
			want: "unknown upstream.upstream_mode=bogus",
		},
		{
			name: "launch",
			cdp: New(Options{
				Launcher: LauncherConfig{LauncherMode: "bogus"},
				Upstream: UpstreamConfig{UpstreamMode: "ws", UpstreamCDPURL: "ws://127.0.0.1:1/devtools/browser/test"},
			}),
			want: "unknown launcher.launcher_mode=bogus",
		},
		{
			name: "injector",
			cdp: New(Options{
				Launcher: LauncherConfig{LauncherMode: "none"},
				Upstream: UpstreamConfig{UpstreamMode: "ws", UpstreamCDPURL: "ws://127.0.0.1:1/devtools/browser/test"},
				Injector: InjectorConfig{InjectorMode: "bogus"},
			}),
			want: "unknown injector.injector_mode=bogus",
		},
	}
	for _, testCase := range cases {
		if err := testCase.cdp.Connect(); err == nil || !strings.Contains(err.Error(), testCase.want) {
			t.Fatalf("%s Connect error = %v", testCase.name, err)
		}
	}
}

func TestModCDPClientOnlyExposesInjectorAttachAfterCDPSendIsAvailable(t *testing.T) {
	cdp := New(Options{})
	disconnectedConfig := cdp.baseExtensionInjectorConfig(nil)
	if disconnectedConfig.Send != nil {
		t.Fatalf("disconnected Send = %#v", disconnectedConfig.Send)
	}
	if disconnectedConfig.AttachToTarget != nil {
		t.Fatalf("disconnected AttachToTarget = %#v", disconnectedConfig.AttachToTarget)
	}

	connectedConfig := cdp.baseExtensionInjectorConfig(func(method string, params map[string]any, sessionID string) (map[string]any, error) {
		return map[string]any{}, nil
	})
	if connectedConfig.Send == nil {
		t.Fatal("connected Send is nil")
	}
	if connectedConfig.AttachToTarget == nil {
		t.Fatal("connected AttachToTarget is nil")
	}
}

func TestModCDPClientConnectsWithLocalLaunchAndInjectorChain(t *testing.T) {
	cdp := New(Options{
		Launcher: LauncherConfig{LauncherMode: "local",
			LauncherOptions: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{UpstreamMode: "ws"},
		Injector: InjectorConfig{
			InjectorMode:                     "inject",
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
		Client: ClientConfig{
			ClientRoutes: map[string]string{"Mod.*": "service_worker", "Custom.*": "service_worker", "*.*": "direct_cdp"},
		},
		Server: &ServerConfig{
			ServerRoutes: map[string]string{"*.*": "loopback_cdp"},
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["injector_source"] != "local_launch" {
		t.Fatalf("injector_source = %v", cdp.ConnectTiming["injector_source"])
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
	extensionPath, err := filepath.Abs("../../../dist/extension")
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
	rawConn, _, _, err := ws.Dial(ctx, chrome.CDPURL)
	if err != nil {
		t.Fatal(err)
	}
	defer rawConn.Close()

	cdp := New(Options{
		Launcher: LauncherConfig{LauncherMode: "remote"},
		Upstream: UpstreamConfig{UpstreamMode: "ws", UpstreamCDPURL: chrome.CDPURL},
		Injector: InjectorConfig{
			InjectorMode:                     "discover",
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
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
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	cdp := New(Options{
		Launcher: LauncherConfig{LauncherMode: "local",
			LauncherOptions: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{
			UpstreamMode:                   "reversews",
			UpstreamReverseWSBind:          "127.0.0.1:29292",
			UpstreamReverseWSWaitTimeoutMS: 30_000,
		},
		Injector: InjectorConfig{
			InjectorMode:                     "auto",
			InjectorExtensionPath:            extensionPath,
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
		Server: &ServerConfig{ServerRoutes: map[string]string{"*.*": "loopback_cdp"}},
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
	if unpackedExtensionPath != extensionPath {
		t.Fatalf("UnpackedExtensionPath = %q", unpackedExtensionPath)
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
	if _, err := os.Stat(unpackedExtensionPath); err != nil {
		t.Fatalf("expected static extension files to remain after close, got %v", err)
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
		Launcher: LauncherConfig{LauncherMode: "local",
			LauncherOptions: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{UpstreamMode: "ws"},
		Injector: InjectorConfig{
			InjectorMode:                     "auto",
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
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
