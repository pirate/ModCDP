package modcdp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func TestReverseWebSocketUpstreamTransportConfigOwnsBindUpdatesWaitTimeoutAndInjectorConfig(t *testing.T) {
	transport := NewReverseWebSocketUpstreamTransport("127.0.0.1:29292", 10)
	if transport.URL != "ws://127.0.0.1:29292" {
		t.Fatalf("URL = %q", transport.URL)
	}
	if transport.GetInjectorConfig().ReverseProxyURL != "ws://127.0.0.1:29292" {
		t.Fatalf("injector config = %#v", transport.GetInjectorConfig())
	}
	transport.Update(map[string]any{"reversews_bind": "127.0.0.1:29293", "reversews_wait_timeout_ms": 5})
	if transport.URL != "ws://127.0.0.1:29293" {
		t.Fatalf("URL after update = %q", transport.URL)
	}
	if transport.GetInjectorConfig().ReverseProxyURL != "ws://127.0.0.1:29293" {
		t.Fatalf("injector config after update = %#v", transport.GetInjectorConfig())
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms") {
		t.Fatalf("WaitForPeer error = %v", err)
	}
}

func TestReverseWebSocketUpstreamTransportCloseResetsPeerWaitState(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	transport := NewReverseWebSocketUpstreamTransport(fmt.Sprintf("127.0.0.1:%d", port), 5)
	if err := transport.Connect(); err != nil {
		t.Fatal(err)
	}
	conn, _, _, err := ws.Dial(context.Background(), transport.URL)
	if err != nil {
		t.Fatal(err)
	}
	hello, err := json.Marshal(map[string]any{"type": "modcdp.reverse.hello", "role": "test-peer", "version": 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := wsutil.WriteClientText(conn, hello); err != nil {
		t.Fatal(err)
	}

	defer conn.Close()
	defer transport.Close()
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer before close = %v", err)
	}
	if transport.PeerInfo["role"] != "test-peer" {
		t.Fatalf("PeerInfo before close = %#v", transport.PeerInfo)
	}
	if err := transport.Close(); err != nil {
		t.Fatalf("Close = %v", err)
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms") {
		t.Fatalf("WaitForPeer after close = %v", err)
	}
	if transport.PeerInfo != nil {
		t.Fatalf("PeerInfo after close = %#v", transport.PeerInfo)
	}
}

func TestReverseWebSocketUpstreamTransportWaitsAgainAfterPeerDisconnects(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	transport := NewReverseWebSocketUpstreamTransport(fmt.Sprintf("127.0.0.1:%d", port), 5)
	if err := transport.Connect(); err != nil {
		t.Fatal(err)
	}
	conn, _, _, err := ws.Dial(context.Background(), transport.URL)
	if err != nil {
		t.Fatal(err)
	}
	hello, err := json.Marshal(map[string]any{"type": "modcdp.reverse.hello", "role": "test-peer", "version": 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := wsutil.WriteClientText(conn, hello); err != nil {
		t.Fatal(err)
	}

	defer transport.Close()
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer before peer disconnect = %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	waitForReversePeerDisconnect(t, transport)
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms") {
		t.Fatalf("WaitForPeer after peer disconnect = %v", err)
	}
}

func TestReverseWebSocketUpstreamTransportAcceptsReplacementPeerAfterDisconnect(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	transport := NewReverseWebSocketUpstreamTransport(fmt.Sprintf("127.0.0.1:%d", port), 500)
	if err := transport.Connect(); err != nil {
		t.Fatal(err)
	}
	conn, _, _, err := ws.Dial(context.Background(), transport.URL)
	if err != nil {
		t.Fatal(err)
	}
	hello, err := json.Marshal(map[string]any{"type": "modcdp.reverse.hello", "role": "first-peer", "version": 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := wsutil.WriteClientText(conn, hello); err != nil {
		t.Fatal(err)
	}

	defer transport.Close()
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer before peer disconnect = %v", err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	waitForReversePeerDisconnect(t, transport)

	replacementConn, _, _, err := ws.Dial(context.Background(), transport.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer replacementConn.Close()
	replacementHello, err := json.Marshal(map[string]any{"type": "modcdp.reverse.hello", "role": "second-peer", "version": 1})
	if err != nil {
		t.Fatal(err)
	}
	if err := wsutil.WriteClientText(replacementConn, replacementHello); err != nil {
		t.Fatal(err)
	}
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer after replacement peer = %v", err)
	}
	if transport.PeerInfo["role"] != "second-peer" {
		t.Fatalf("PeerInfo after replacement peer = %#v", transport.PeerInfo)
	}
}

func TestReverseWebSocketUpstreamTransportCloseRejectsPendingPeerWaits(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	transport := NewReverseWebSocketUpstreamTransport(fmt.Sprintf("127.0.0.1:%d", port), 5_000)
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
		expected := fmt.Sprintf("reverse websocket transport at ws://127.0.0.1:%d closed before a peer connected", port)
		if err == nil || !strings.Contains(err.Error(), expected) {
			t.Fatalf("WaitForPeer close error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForPeer did not return after Close")
	}
}

func waitForReversePeerDisconnect(t *testing.T, transport *ReverseWebSocketUpstreamTransport) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if transport.Conn == nil && transport.PeerInfo == nil {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timed out waiting for reverse peer disconnect")
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
	transport, ok := cdp.transport.(*ReverseWebSocketUpstreamTransport)
	if !ok {
		t.Fatalf("transport = %T", cdp.transport)
	}
	if transport.PeerInfo["extension_id"] != DefaultModCDPExtensionID {
		t.Fatalf("extension_id = %#v", transport.PeerInfo["extension_id"])
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
