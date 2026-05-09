package modcdp

import (
	"strings"
	"testing"
)

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
