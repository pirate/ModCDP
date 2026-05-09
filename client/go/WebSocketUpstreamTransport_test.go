package modcdp

import (
	"strings"
	"testing"

	"github.com/gobwas/ws/wsutil"
)

func TestWebSocketUpstreamTransportConstructorUpdateAndServerConfigMatchTSShape(t *testing.T) {
	transport := NewWebSocketUpstreamTransport()
	if transport.URL != "" {
		t.Fatalf("URL = %q", transport.URL)
	}
	if len(transport.GetServerConfig()) != 0 {
		t.Fatalf("server config = %#v", transport.GetServerConfig())
	}
	transport.Update(map[string]any{"ws_url": "ws://127.0.0.1:1/devtools/browser/test"})
	if transport.URL != "ws://127.0.0.1:1/devtools/browser/test" {
		t.Fatalf("URL = %q", transport.URL)
	}
	if transport.GetServerConfig()["loopback_cdp_url"] != "ws://127.0.0.1:1/devtools/browser/test" {
		t.Fatalf("server config = %#v", transport.GetServerConfig())
	}
	if err := NewWebSocketUpstreamTransport().Connect(); err == nil || !strings.Contains(err.Error(), "upstream.mode=ws requires") {
		t.Fatalf("connect error = %v", err)
	}
}

func TestWebSocketUpstreamTransportLaunchesRealBrowserAndSpeaksRawCDP(t *testing.T) {
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
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["upstream_endpoint_kind"] != nil && cdp.ConnectTiming["upstream_endpoint_kind"] != UpstreamEndpointKindRawCDP {
		t.Fatalf("upstream_endpoint_kind = %v", cdp.ConnectTiming["upstream_endpoint_kind"])
	}
	if _, ok := cdp.transport.(*WebSocketUpstreamTransport); !ok {
		t.Fatalf("transport = %T", cdp.transport)
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
}

func TestWebSocketUpstreamTransportResolvesRealHTTPCDPEndpointToBrowserWebSocket(t *testing.T) {
	chrome, err := NewLocalBrowserLauncher(LaunchOptions{
		Headless: boolPtr(true),
		Sandbox:  boolPtr(false),
	}).Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()

	transport := NewWebSocketUpstreamTransport(chrome.CDPURL)
	if err := transport.Connect(); err != nil {
		t.Fatal(err)
	}
	defer transport.Close()
	if !strings.HasPrefix(transport.URL, "ws://") {
		t.Fatalf("transport.URL = %q", transport.URL)
	}
	if err := transport.Send(map[string]any{"id": 1, "method": "Browser.getVersion", "params": map[string]any{}}); err != nil {
		t.Fatal(err)
	}
	data, err := wsutil.ReadServerText(transport.Conn)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"id":1`) || !strings.Contains(string(data), `"product"`) {
		t.Fatalf("Browser.getVersion response = %s", string(data))
	}
}
