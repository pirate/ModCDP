package transport_test

import (
	"encoding/json"
	"fmt"
	modcdp "github.com/pirate/ModCDP/go/modcdp/client"
	. "github.com/pirate/ModCDP/go/modcdp/transport"
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
		UpstreamNativeMessagingManifest:      "/tmp/modcdp-native-host.json",
		UpstreamNativeMessagingManifests:     []string{"/tmp/modcdp-native-host-extra.json"},
		UpstreamNativeMessagingHostName:      "com.modcdp.test",
		ExtensionID:                          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		UpstreamNativeMessagingWaitTimeoutMS: 10,
	})
	if err != nil {
		t.Fatal(err)
	}
	if raw := string(encoded); raw != `{"upstream_nativemessaging_manifest":"/tmp/modcdp-native-host.json","upstream_nativemessaging_manifests":["/tmp/modcdp-native-host-extra.json"],"upstream_nativemessaging_host_name":"com.modcdp.test","extension_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","upstream_nativemessaging_wait_timeout_ms":10}` {
		t.Fatalf("NativeMessagingUpstreamTransportOptions JSON = %s", raw)
	}

	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
		UpstreamNativeMessagingManifest:      "/tmp/modcdp-native-host.json",
		UpstreamNativeMessagingManifests:     []string{"/tmp/modcdp-native-host-extra.json"},
		UpstreamNativeMessagingHostName:      "com.modcdp.test",
		ExtensionID:                          "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		UpstreamNativeMessagingWaitTimeoutMS: 10,
	})
	transport.Update(map[string]any{
		"cdp_url":                                  "ws://127.0.0.1:9222/devtools/browser/test",
		"upstream_nativemessaging_manifests":       []string{},
		"upstream_nativemessaging_host_name":       "com.modcdp.updated",
		"upstream_nativemessaging_wait_timeout_ms": 5,
	})
	if transport.GetInjectorConfig().UpstreamNativeMessagingHostName != "com.modcdp.updated" {
		t.Fatalf("updated injector config = %#v", transport.GetInjectorConfig())
	}
	if transport.GetServerConfig()["loopback_cdp_url"] != "ws://127.0.0.1:9222/devtools/browser/test" {
		t.Fatalf("server config = %#v", transport.GetServerConfig())
	}
	if transport.UpstreamNativeMessagingManifest != "/tmp/modcdp-native-host.json" {
		t.Fatalf("UpstreamNativeMessagingManifest = %q", transport.UpstreamNativeMessagingManifest)
	}
	if transport.ExtensionID != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("ExtensionID = %q", transport.ExtensionID)
	}
	if transport.IncludeDefaultUpstreamNativeMessagingManifests {
		t.Fatal("IncludeDefaultUpstreamNativeMessagingManifests should stay false while UpstreamNativeMessagingManifest is set")
	}
	transport.Update(map[string]any{"upstream_nativemessaging_manifest": ""})
	if !transport.IncludeDefaultUpstreamNativeMessagingManifests {
		t.Fatal("IncludeDefaultUpstreamNativeMessagingManifests should be true after clearing UpstreamNativeMessagingManifest and UpstreamNativeMessagingManifests")
	}
	transport.Update(map[string]any{"user_data_dir": "/tmp/modcdp-profile-one"})
	transport.Update(map[string]any{"user_data_dir": "/tmp/modcdp-profile-one"})
	transport.Update(map[string]any{"user_data_dir": "/tmp/modcdp-profile-two"})
	expectedUpstreamNativeMessagingManifests := []string{
		filepath.Join("/tmp/modcdp-profile-two", "NativeMessagingHosts", "com.modcdp.updated.json"),
		filepath.Join("/tmp/modcdp-profile-two", "Default", "NativeMessagingHosts", "com.modcdp.updated.json"),
	}
	if strings.Join(transport.UpstreamNativeMessagingManifests, "\n") != strings.Join(expectedUpstreamNativeMessagingManifests, "\n") {
		t.Fatalf("UpstreamNativeMessagingManifests = %#v", transport.UpstreamNativeMessagingManifests)
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for native messaging host com.modcdp.updated") {
		t.Fatalf("WaitForPeer error = %v", err)
	}
}

func TestNativeMessagingUpstreamTransportCloseResetsPeerWaitState(t *testing.T) {
	nativeHostName := fmt.Sprintf("com.modcdp.close.reset.go.%d", os.Getpid())
	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
		UpstreamNativeMessagingHostName:      nativeHostName,
		UpstreamNativeMessagingWaitTimeoutMS: 5,
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
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for native messaging host "+nativeHostName) {
		t.Fatalf("WaitForPeer after close = %v", err)
	}
	if !transport.Closed() {
		t.Fatalf("closed after Close = %v", transport.Closed())
	}
}

func TestNativeMessagingUpstreamTransportWaitsAgainAfterPeerDisconnects(t *testing.T) {
	nativeHostName := fmt.Sprintf("com.modcdp.disconnect.reset.go.%d", os.Getpid())
	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
		UpstreamNativeMessagingHostName:      nativeHostName,
		UpstreamNativeMessagingWaitTimeoutMS: 5,
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
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for native messaging host "+nativeHostName) {
		t.Fatalf("WaitForPeer after peer disconnect = %v", err)
	}
}

func TestNativeMessagingUpstreamTransportAcceptsReplacementPeerAfterDisconnect(t *testing.T) {
	nativeHostName := fmt.Sprintf("com.modcdp.replacement.go.%d", os.Getpid())
	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
		UpstreamNativeMessagingHostName:      nativeHostName,
		UpstreamNativeMessagingWaitTimeoutMS: 500,
	})
	if err := transport.Connect(); err != nil {
		t.Fatal(err)
	}
	firstConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", transport.BoundPort))
	if err != nil {
		t.Fatal(err)
	}

	defer transport.Close()
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer before peer disconnect = %v", err)
	}
	if err := firstConn.Close(); err != nil {
		t.Fatal(err)
	}
	waitForNativePeerDisconnect(t, transport)
	secondConn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", transport.BoundPort))
	if err != nil {
		t.Fatal(err)
	}
	defer secondConn.Close()
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer after replacement peer = %v", err)
	}
}

func TestNativeMessagingUpstreamTransportCloseRejectsPendingPeerWaits(t *testing.T) {
	transport := NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{
		UpstreamNativeMessagingHostName:      "com.modcdp.close",
		UpstreamNativeMessagingWaitTimeoutMS: 5_000,
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
	nativeHostName := "com.modcdp.bridge"
	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "local",
			LauncherOptions: modcdp.LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "nativemessaging", UpstreamNativeMessagingHostName: nativeHostName},
		Injector: modcdp.InjectorConfig{
			InjectorMode:                     "auto",
			InjectorServiceWorkerURLSuffixes: []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget: true,
		},
		Server: &modcdp.ServerConfig{ServerRoutes: map[string]string{"*.*": "loopback_cdp"}},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["upstream_endpoint_kind"] != UpstreamEndpointKindModCDPServer {
		t.Fatalf("upstream_endpoint_kind = %v", cdp.ConnectTiming["upstream_endpoint_kind"])
	}
	transport, ok := cdp.Transport().(*NativeMessagingUpstreamTransport)
	if !ok {
		t.Fatalf("transport = %T", cdp.Transport())
	}
	if !regexp.MustCompile(`^native://` + regexp.QuoteMeta(nativeHostName) + `@127\.0\.0\.1:\d+$`).MatchString(transport.URL) {
		t.Fatalf("transport.URL = %q", transport.URL)
	}
	if cdp.LaunchedBrowser() == nil {
		t.Fatal("expected launched browser")
	}
	manifestPath := filepath.Join(
		cdp.LaunchedBrowser().ProfileDir,
		"NativeMessagingHosts",
		nativeHostName+".json",
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
	time.Sleep(1500 * time.Millisecond)
	secondResult, err := cdp.Send("Browser.getVersion", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	secondVersion, ok := secondResult.(map[string]any)
	if !ok {
		t.Fatalf("second Browser.getVersion result = %#v", secondResult)
	}
	if _, ok := secondVersion["product"].(string); !ok {
		t.Fatalf("second Browser.getVersion product = %#v", secondVersion["product"])
	}
}
