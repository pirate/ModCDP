package modcdp

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNativeMessagingUpstreamTransportConfigOwnsManifestHostWaitTimeoutLoopbackAndInjectorConfig(t *testing.T) {
	encoded, err := json.Marshal(NativeMessagingUpstreamTransportOptions{
		ManifestPath:  "/tmp/modcdp-native-host.json",
		ManifestPaths: []string{"/tmp/modcdp-native-host-extra.json"},
		HostName:      "com.modcdp.test",
		ExtensionID:   "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		WaitTimeoutMS: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if raw := string(encoded); raw != `{"manifest_path":"/tmp/modcdp-native-host.json","manifest_paths":["/tmp/modcdp-native-host-extra.json"],"host_name":"com.modcdp.test","extension_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","wait_timeout_ms":10}` {
		t.Fatalf("NativeMessagingUpstreamTransportOptions JSON = %s", raw)
	}

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
	transport.Update(map[string]any{"user_data_dir": "/tmp/modcdp-profile-one"})
	transport.Update(map[string]any{"user_data_dir": "/tmp/modcdp-profile-one"})
	transport.Update(map[string]any{"user_data_dir": "/tmp/modcdp-profile-two"})
	expectedManifestPaths := []string{
		filepath.Join("/tmp/modcdp-profile-two", "NativeMessagingHosts", "com.modcdp.updated.json"),
		filepath.Join("/tmp/modcdp-profile-two", "Default", "NativeMessagingHosts", "com.modcdp.updated.json"),
	}
	if strings.Join(transport.ManifestPaths, "\n") != strings.Join(expectedManifestPaths, "\n") {
		t.Fatalf("ManifestPaths = %#v", transport.ManifestPaths)
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for native messaging host com.modcdp.updated") {
		t.Fatalf("WaitForPeer error = %v", err)
	}
}

func TestNativeMessagingUpstreamTransportCloseResetsPeerWaitState(t *testing.T) {
	hostName := fmt.Sprintf("com.modcdp.close.reset.go.%d", os.Getpid())
	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
		HostName:      hostName,
		WaitTimeoutMS: 5,
	})
	if err := transport.Connect(); err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", transport.BoundPort))
	if err != nil {
		t.Fatal(err)
	}

	defer conn.Close()
	defer transport.Close()
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer before close = %v", err)
	}
	if err := transport.Close(); err != nil {
		t.Fatalf("Close = %v", err)
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for native messaging host "+hostName) {
		t.Fatalf("WaitForPeer after close = %v", err)
	}
	if !transport.closed {
		t.Fatalf("closed after Close = %v", transport.closed)
	}
}

func TestNativeMessagingUpstreamTransportWaitsAgainAfterPeerDisconnects(t *testing.T) {
	hostName := fmt.Sprintf("com.modcdp.disconnect.reset.go.%d", os.Getpid())
	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
		HostName:      hostName,
		WaitTimeoutMS: 5,
	})
	if err := transport.Connect(); err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", transport.BoundPort))
	if err != nil {
		t.Fatal(err)
	}

	defer transport.Close()
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer before peer disconnect = %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	waitForNativePeerDisconnect(t, transport)
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for native messaging host "+hostName) {
		t.Fatalf("WaitForPeer after peer disconnect = %v", err)
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

func waitForNativePeerDisconnect(t *testing.T, transport *NativeMessagingUpstreamTransport) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if transport.Conn == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timed out waiting for native peer disconnect")
}

func TestNativeMessagingUpstreamTransportInstallsLaunchProfileManifestAndConnectsToRealExtension(t *testing.T) {
	hostName := fmt.Sprintf("com.modcdp.test.go.%d", os.Getpid())
	cdp := New(Options{
		Launch: LaunchConfig{
			Mode: "local",
			Options: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{Mode: "nativemessaging", NativeMessagingHostName: hostName},
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
	if !regexp.MustCompile(`^native://` + regexp.QuoteMeta(hostName) + `@127\.0\.0\.1:\d+$`).MatchString(transport.URL) {
		t.Fatalf("transport.URL = %q", transport.URL)
	}
	if cdp.launchedBrowser == nil {
		t.Fatal("expected launched browser")
	}
	manifestPath := filepath.Join(
		cdp.launchedBrowser.ProfileDir,
		"NativeMessagingHosts",
		hostName+".json",
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
