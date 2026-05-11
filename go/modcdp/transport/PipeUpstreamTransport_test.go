package transport_test

import (
	modcdp "github.com/pirate/ModCDP/go/modcdp/client"
	. "github.com/pirate/ModCDP/go/modcdp/transport"
	"regexp"
	"testing"
)

func TestPipeUpstreamTransportConstructorUpdateLauncherConfigAndUnconnectedErrorsMatchTransportSurface(t *testing.T) {
	transport := NewPipeUpstreamTransport(PipeUpstreamTransportOptions{CDPURL: "pipe://constructor"})
	if transport.URL != "pipe://constructor" {
		t.Fatalf("URL = %q", transport.URL)
	}
	if launcherConfig := transport.GetLauncherConfig(); launcherConfig.RemoteDebugging != "pipe" {
		t.Fatalf("launcher config = %#v", launcherConfig)
	}
	transport.Update(map[string]any{"cdp_url": "pipe://1234"})
	if transport.URL != "pipe://1234" {
		t.Fatalf("URL after update = %q", transport.URL)
	}
	if err := transport.Connect(); err == nil {
		t.Fatal("expected Connect to require pipe handles")
	}
	if err := transport.Send(map[string]any{"id": 1, "method": "Browser.getVersion"}); err == nil {
		t.Fatal("expected Send to require a connected pipe")
	}
}

func TestPipeUpstreamTransportLaunchesRealBrowserAndUsesPIDScopedPipeURL(t *testing.T) {
	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "local",
			LauncherOptions: modcdp.LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "pipe"},
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
	if cdp.ConnectTiming["upstream_endpoint_kind"] != UpstreamEndpointKindRawCDP {
		t.Fatalf("upstream_endpoint_kind = %v", cdp.ConnectTiming["upstream_endpoint_kind"])
	}
	if cdp.Transport() == nil {
		t.Fatal("expected pipe transport")
	}
	if !regexp.MustCompile(`^pipe://\d+$`).MatchString(cdp.CDPURL) {
		t.Fatalf("CDPURL = %q", cdp.CDPURL)
	}
	pipeTransport, ok := cdp.Transport().(*PipeUpstreamTransport)
	if !ok {
		t.Fatalf("transport = %T", cdp.Transport())
	}
	if pipeTransport.URL != cdp.CDPURL {
		t.Fatalf("pipe transport URL = %q, CDPURL = %q", pipeTransport.URL, cdp.CDPURL)
	}
	version, err := cdp.SendRaw("Browser.getVersion", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := version["product"].(string); !ok {
		t.Fatalf("Browser.getVersion product = %#v", version["product"])
	}
}
