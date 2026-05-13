package transport_test

import (
	modcdp "github.com/browserbase/modcdp/go/modcdp/client"
	. "github.com/browserbase/modcdp/go/modcdp/transport"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestWebSocketUpstreamTransportConstructorUpdateAndServerConfigMatchTSShape(t *testing.T) {
	transport := NewWebSocketUpstreamTransport(WebSocketUpstreamTransportOptions{})
	if transport.URL != "" {
		t.Fatalf("URL = %q", transport.URL)
	}
	if len(transport.GetServerConfig()) != 0 {
		t.Fatalf("server config = %#v", transport.GetServerConfig())
	}
	transport.Update(map[string]any{"cdp_url": "ws://127.0.0.1:1/devtools/browser/test"})
	if transport.URL != "ws://127.0.0.1:1/devtools/browser/test" {
		t.Fatalf("URL = %q", transport.URL)
	}
	if transport.GetServerConfig()["server_loopback_cdp_url"] != "ws://127.0.0.1:1/devtools/browser/test" {
		t.Fatalf("server config = %#v", transport.GetServerConfig())
	}
	if err := NewWebSocketUpstreamTransport(WebSocketUpstreamTransportOptions{}).Connect(); err == nil || !strings.Contains(err.Error(), "upstream.upstream_mode=ws requires") {
		t.Fatalf("connect error = %v", err)
	}
	if err := NewWebSocketUpstreamTransport(WebSocketUpstreamTransportOptions{}).Send(map[string]any{"id": 1, "method": "Browser.getVersion"}); err == nil || !strings.Contains(err.Error(), "CDP websocket is not connected") {
		t.Fatalf("send error = %v", err)
	}
}

func TestWebSocketUpstreamTransportLaunchesRealBrowserAndSpeaksRawCDP(t *testing.T) {
	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "local",
			LauncherOptions: modcdp.LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "ws"},
		Injector: modcdp.InjectorConfig{
			InjectorMode:                     "auto",
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["upstream_mode"] != "ws" {
		t.Fatalf("upstream_mode = %v", cdp.ConnectTiming["upstream_mode"])
	}
	if cdp.ConnectTiming["upstream_endpoint_kind"] != UpstreamEndpointKindRawCDP {
		t.Fatalf("upstream_endpoint_kind = %v", cdp.ConnectTiming["upstream_endpoint_kind"])
	}
	transportStartedAt, ok := cdp.ConnectTiming["transport_started_at"].(int64)
	if !ok {
		t.Fatalf("transport_started_at = %#v", cdp.ConnectTiming["transport_started_at"])
	}
	transportConnectedAt, ok := cdp.ConnectTiming["transport_connected_at"].(int64)
	if !ok {
		t.Fatalf("transport_connected_at = %#v", cdp.ConnectTiming["transport_connected_at"])
	}
	if transportConnectedAt < transportStartedAt {
		t.Fatalf("transport timing went backwards: %d < %d", transportConnectedAt, transportStartedAt)
	}
	if cdp.ConnectTiming["transport_duration_ms"] != transportConnectedAt-transportStartedAt {
		t.Fatalf("transport_duration_ms = %v", cdp.ConnectTiming["transport_duration_ms"])
	}
	if _, ok := cdp.Transport().(*WebSocketUpstreamTransport); !ok {
		t.Fatalf("transport = %T", cdp.Transport())
	}
	if !strings.HasPrefix(cdp.CDPURL, "ws://") {
		t.Fatalf("CDPURL = %q", cdp.CDPURL)
	}
	version, err := cdp.SendRaw("Browser.getVersion", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := version["product"].(string); !ok {
		t.Fatalf("Browser.getVersion product = %#v", version["product"])
	}
	time.Sleep(1500 * time.Millisecond)
	targets, err := cdp.SendRaw("Target.getTargets", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	foundServiceWorker := false
	for _, target := range targets["targetInfos"].([]any) {
		targetMap := target.(map[string]any)
		if targetMap["type"] == "service_worker" && strings.HasSuffix(targetMap["url"].(string), "/modcdp/service_worker.js") {
			foundServiceWorker = true
			break
		}
	}
	if !foundServiceWorker {
		t.Fatalf("ModCDP service worker target not found after connect: %#v", targets["targetInfos"])
	}
	evaluated, err := cdp.Mod.Evaluate(map[string]any{
		"expression": "Boolean(globalThis.ModCDP?.handleCommand && chrome.runtime.getURL('modcdp/service_worker.js'))",
	})
	if err != nil {
		t.Fatal(err)
	}
	if evaluated != true {
		t.Fatalf("Mod.evaluate liveness = %#v", evaluated)
	}
}

func TestWebSocketUpstreamTransportResolvesRealHTTPCDPEndpointToBrowserWebSocket(t *testing.T) {
	chrome, err := modcdp.NewLocalBrowserLauncher(modcdp.LaunchOptions{
		Headless: boolPtr(true),
		Sandbox:  boolPtr(false),
	}).Launch(modcdp.LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	transport := NewWebSocketUpstreamTransport(WebSocketUpstreamTransportOptions{CDPURL: chrome.CDPURL})
	if err := transport.Connect(); err != nil {
		t.Fatal(err)
	}
	defer transport.Close()
	if !strings.HasPrefix(transport.URL, "ws://") {
		t.Fatalf("transport.URL = %q", transport.URL)
	}
	received := make(chan map[string]any, 1)
	transport.OnRecv(func(message map[string]any) { received <- message })
	if err := transport.Send(map[string]any{"id": 1, "method": "Browser.getVersion", "params": map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	message := <-received
	if message["id"] != float64(1) && message["id"] != 1 {
		t.Fatalf("Browser.getVersion id = %#v", message["id"])
	}
	result, _ := message["result"].(map[string]any)
	if _, ok := result["product"].(string); !ok {
		t.Fatalf("Browser.getVersion response = %#v", message)
	}

	parsedCDPURL, err := url.Parse(chrome.CDPURL)
	if err != nil {
		t.Fatal(err)
	}
	hostPortTransport := NewWebSocketUpstreamTransport(WebSocketUpstreamTransportOptions{CDPURL: parsedCDPURL.Host})
	if err := hostPortTransport.Connect(); err != nil {
		t.Fatal(err)
	}
	defer hostPortTransport.Close()
	if !strings.HasPrefix(hostPortTransport.URL, "ws://") && !strings.HasPrefix(hostPortTransport.URL, "wss://") {
		t.Fatalf("hostPortTransport.URL = %q", hostPortTransport.URL)
	}
}

func TestWebSocketUpstreamTransportCloseClearsConnectionState(t *testing.T) {
	chrome, err := modcdp.NewLocalBrowserLauncher(modcdp.LaunchOptions{
		Headless: boolPtr(true),
		Sandbox:  boolPtr(false),
	}).Launch(modcdp.LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	transport := NewWebSocketUpstreamTransport(WebSocketUpstreamTransportOptions{CDPURL: chrome.CDPURL})
	if err := transport.Connect(); err != nil {
		t.Fatal(err)
	}
	if transport.Conn == nil {
		t.Fatal("expected connected websocket")
	}
	if err := transport.Close(); err != nil {
		t.Fatal(err)
	}
	if transport.Conn != nil {
		t.Fatal("Close left Conn set")
	}
	if err := transport.Send(map[string]any{"id": 1, "method": "Browser.getVersion"}); err == nil || !strings.Contains(err.Error(), "CDP websocket is not connected") {
		t.Fatalf("Send after close error = %v", err)
	}
}
