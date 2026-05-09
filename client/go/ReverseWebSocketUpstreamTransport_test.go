package modcdp

import (
	"fmt"
	"testing"
)

func TestReverseWebSocketUpstreamTransportConfigOwnsBindUpdatesAndInjectorConfig(t *testing.T) {
	transport := NewReverseWebSocketUpstreamTransport("127.0.0.1:29292")
	if transport.URL != "ws://127.0.0.1:29292" {
		t.Fatalf("URL = %q", transport.URL)
	}
	if transport.GetInjectorConfig().ReverseProxyURL != "ws://127.0.0.1:29292" {
		t.Fatalf("injector config = %#v", transport.GetInjectorConfig())
	}
	transport.Update(map[string]any{"reversews_bind": "127.0.0.1:29293"})
	if transport.URL != "ws://127.0.0.1:29293" {
		t.Fatalf("URL after update = %q", transport.URL)
	}
	if transport.GetInjectorConfig().ReverseProxyURL != "ws://127.0.0.1:29293" {
		t.Fatalf("injector config after update = %#v", transport.GetInjectorConfig())
	}
}

func TestReverseWebSocketUpstreamTransportAcceptsRealExtensionReverseConnectionAndRoutesCDPThroughLoopback(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	reverseBind := fmt.Sprintf("127.0.0.1:%d", port)
	cdp := New(Options{
		Launch: LaunchConfig{
			Mode: "local",
			Options: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{Mode: "reversews", ReverseWSBind: reverseBind},
		Extension: ExtensionConfig{
			Mode:                     "auto",
			ServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			TrustServiceWorkerTarget: true,
		},
		Server: &ServerConfig{Routes: map[string]string{"*.*": "loopback_cdp"}},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["upstream_endpoint_kind"] != UpstreamEndpointKindModCDPServer {
		t.Fatalf("upstream_endpoint_kind = %v", cdp.ConnectTiming["upstream_endpoint_kind"])
	}
	if cdp.transport == nil {
		t.Fatal("expected reverse transport to be connected")
	}
	result, err := cdp.Send("Browser.getVersion", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	version, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Browser.getVersion result = %#v", result)
	}
	if _, ok := version["product"].(string); !ok {
		t.Fatalf("Browser.getVersion product = %#v", version["product"])
	}
}
