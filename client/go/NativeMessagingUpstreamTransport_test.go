package modcdp

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNativeMessagingUpstreamTransportConfigOwnsManifestHostWaitTimeoutLoopbackAndInjectorConfig(t *testing.T) {
	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
		ManifestPath:  "/tmp/modcdp-native-host.json",
		ManifestPaths: []string{"/tmp/modcdp-native-host-extra.json"},
		HostName:      "com.modcdp.test",
		ExtensionID:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		WaitTimeoutMS: 10,
	})
	transport.Update(map[string]any{
		"ws_url":           "ws://127.0.0.1:9222/devtools/browser/test",
		"manifest_paths":   []string{},
		"native_host_name": "com.modcdp.updated",
		"wait_timeout_ms":  5,
	})
	if transport.GetInjectorConfig().NativeHostName != "com.modcdp.updated" {
		t.Fatalf("updated injector config = %#v", transport.GetInjectorConfig())
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
	if transport.IncludeDefaultManifestPaths {
		t.Fatal("IncludeDefaultManifestPaths should stay false while ManifestPath is set")
	}
	transport.Update(map[string]any{"manifest_path": ""})
	if !transport.IncludeDefaultManifestPaths {
		t.Fatal("IncludeDefaultManifestPaths should be true after clearing ManifestPath and ManifestPaths")
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for native messaging host com.modcdp.updated") {
		t.Fatalf("WaitForPeer error = %v", err)
	}
}

func TestNativeMessagingUpstreamTransportCloseResetsPeerWaitState(t *testing.T) {
	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{WaitTimeoutMS: 5})

	transport.peerOnce.Do(func() { close(transport.peerCh) })
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer before close = %v", err)
	}
	if err := transport.Close(); err != nil {
		t.Fatalf("Close = %v", err)
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for native messaging host com.modcdp.bridge") {
		t.Fatalf("WaitForPeer after close = %v", err)
	}
	if !transport.closed {
		t.Fatalf("closed after Close = %v", transport.closed)
	}
}

func TestNativeMessagingUpstreamTransportCloseRejectsPendingPeerWaits(t *testing.T) {
	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
		HostName:      "com.modcdp.close",
		WaitTimeoutMS: 5_000,
	})
	done := make(chan error, 1)
	go func() {
		done <- transport.WaitForPeer()
	}()
	time.Sleep(50 * time.Millisecond)
	if err := transport.Close(); err != nil {
		t.Fatalf("Close = %v", err)
	}
	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "native messaging transport for com.modcdp.close closed before a peer connected") {
			t.Fatalf("WaitForPeer close error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForPeer did not return after Close")
	}
}

func TestNativeMessagingUpstreamTransportInstallsLaunchProfileManifestAndConnectsToRealExtension(t *testing.T) {
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
