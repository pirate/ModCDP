package modcdp

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

func TestNativeMessagingUpstreamTransportConfigOwnsManifestLoopbackAndInjectorConfig(t *testing.T) {
	transport := NewNativeMessagingUpstreamTransport("/tmp/modcdp-native-host.json")
	transport.Update(map[string]any{
		"extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"ws_url":       "ws://127.0.0.1:9222/devtools/browser/test",
	})
	if transport.GetInjectorConfig().NativeHostName != DefaultNativeMessagingHostName {
		t.Fatalf("injector config = %#v", transport.GetInjectorConfig())
	}
	if transport.GetServerConfig()["loopback_cdp_url"] != "ws://127.0.0.1:9222/devtools/browser/test" {
		t.Fatalf("server config = %#v", transport.GetServerConfig())
	}
	if transport.ManifestPath != "/tmp/modcdp-native-host.json" {
		t.Fatalf("ManifestPath = %q", transport.ManifestPath)
	}
	if transport.ExtensionID != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("ExtensionID = %q", transport.ExtensionID)
	}
}

func TestNativeMessagingUpstreamTransportInstallsLaunchProfileManifestAndConnectsToRealExtension(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("native messaging profile manifest path is not implemented on Windows")
	}
	cdp := New(Options{
		Launch: LaunchConfig{
			Mode: "local",
			Options: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{Mode: "nativemessaging"},
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
	transport, ok := cdp.transport.(*NativeMessagingUpstreamTransport)
	if !ok {
		t.Fatalf("transport = %T", cdp.transport)
	}
	if !regexp.MustCompile(`^native://com\.modcdp\.bridge@127\.0\.0\.1:\d+$`).MatchString(transport.URL) {
		t.Fatalf("transport.URL = %q", transport.URL)
	}
	if cdp.launchedBrowser == nil {
		t.Fatal("expected launched browser")
	}
	manifestPath := filepath.Join(
		cdp.launchedBrowser.ProfileDir,
		"NativeMessagingHosts",
		DefaultNativeMessagingHostName+".json",
	)
	if _, err := os.Stat(manifestPath); err != nil {
		t.Fatalf("native messaging profile manifest was not installed at %s", manifestPath)
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
