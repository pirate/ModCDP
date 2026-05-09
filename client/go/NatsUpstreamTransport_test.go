package modcdp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
)

func TestNatsUpstreamTransportConfigOwnsURLSubjectPrefixWaitTimeoutAndInjectorConfig(t *testing.T) {
	transport := NewNatsUpstreamTransport(NatsUpstreamTransportOptions{
		URL:           "ws://127.0.0.1:4223",
		SubjectPrefix: "modcdp.one",
	})
	if transport.URL != "ws://127.0.0.1:4223/" {
		t.Fatalf("URL = %q", transport.URL)
	}
	if transport.SubjectPrefix != "modcdp.one" {
		t.Fatalf("SubjectPrefix = %q", transport.SubjectPrefix)
	}
	injectorConfig := transport.GetInjectorConfig()
	if injectorConfig.NATSURL != "ws://127.0.0.1:4223/" || injectorConfig.NATSSubjectPrefix != "modcdp.one" {
		t.Fatalf("injector config = %#v", injectorConfig)
	}
	transport.Update(map[string]any{
		"nats_url":            "nats://127.0.0.1:4222",
		"nats_subject_prefix": "modcdp.two",
		"role":                "browser",
		"wait_timeout_ms":     5,
	})
	if transport.URL != "nats://127.0.0.1:4222" {
		t.Fatalf("URL after update = %q", transport.URL)
	}
	if transport.SubjectPrefix != "modcdp.two" {
		t.Fatalf("SubjectPrefix after update = %q", transport.SubjectPrefix)
	}
	if transport.Role != "browser" {
		t.Fatalf("Role after update = %q", transport.Role)
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for NATS ModCDP peer") {
		t.Fatalf("WaitForPeer error = %v", err)
	}
}

func TestNatsUpstreamTransportCloseResetsPeerWaitState(t *testing.T) {
	transport := NewNatsUpstreamTransport(NatsUpstreamTransportOptions{WaitTimeoutMS: 5})

	transport.handlePayload(`{"type":"modcdp.nats.hello","role":"browser","version":1}`)
	if err := transport.WaitForPeer(); err != nil {
		t.Fatalf("WaitForPeer before close = %v", err)
	}
	if err := transport.Close(); err != nil {
		t.Fatalf("Close = %v", err)
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms for NATS ModCDP peer") {
		t.Fatalf("WaitForPeer after close = %v", err)
	}
	if transport.closed != true {
		t.Fatalf("closed after Close = %v", transport.closed)
	}
}

func TestNatsUpstreamTransportCloseRejectsPendingPeerWaits(t *testing.T) {
	transport := NewNatsUpstreamTransport(NatsUpstreamTransportOptions{
		URL:           "ws://127.0.0.1:4223",
		SubjectPrefix: "modcdp.close",
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
		if err == nil || !strings.Contains(err.Error(), "NATS transport for modcdp.close closed before a peer connected") {
			t.Fatalf("WaitForPeer close error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("WaitForPeer did not return after Close")
	}
}

func TestNatsUpstreamTransportRelaysCDPThroughRealExtensionOverRealNATSServer(t *testing.T) {
	nats := startNATSServer(t)
	subjectPrefix := fmt.Sprintf("modcdp.test.%d", time.Now().UnixMilli())
	cdp := New(Options{
		Launch: LaunchConfig{
			Mode: "local",
			Options: LaunchOptions{
				Headless: boolPtr(true),
				Sandbox:  boolPtr(false),
			},
		},
		Upstream: UpstreamConfig{
			Mode:              "nats",
			NATSURL:           nats.url,
			NATSSubjectPrefix: subjectPrefix,
		},
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
	transport, ok := cdp.transport.(*NatsUpstreamTransport)
	if !ok {
		t.Fatalf("transport = %T", cdp.transport)
	}
	if transport.URL != nats.url+"/" {
		t.Fatalf("transport.URL = %q", transport.URL)
	}
	if transport.SubjectPrefix != subjectPrefix {
		t.Fatalf("transport.SubjectPrefix = %q", transport.SubjectPrefix)
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

type natsTestServer struct {
	url   string
	close func()
}

func startNATSServer(t *testing.T) natsTestServer {
	t.Helper()
	websocketPort, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	clientPort, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nats.conf")
	config := strings.Join([]string{
		`host: "127.0.0.1"`,
		fmt.Sprintf("port: %d", clientPort),
		"websocket {",
		`  host: "127.0.0.1"`,
		fmt.Sprintf("  port: %d", websocketPort),
		"  no_tls: true",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatal(err)
	}
	binaryPath := natsServerBinaryPath(t)
	cmd := exec.Command(binaryPath, "-c", configPath)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	url := fmt.Sprintf("ws://127.0.0.1:%d", websocketPort)
	closeServer := func() {
		if cmd.Process == nil {
			return
		}
		_ = cmd.Process.Kill()
		_, _ = cmd.Process.Wait()
	}
	t.Cleanup(closeServer)
	if err := waitForNATSWebSocket(url, 10*time.Second); err != nil {
		closeServer()
		t.Fatal(err)
	}
	return natsTestServer{url: url, close: closeServer}
}

func natsServerBinaryPath(t *testing.T) string {
	t.Helper()
	cmd := exec.Command(
		"pnpm",
		"exec",
		"node",
		"--input-type=module",
		"-e",
		"import { getBinaryPath } from '@eplightning/nats-server'; console.log(await getBinaryPath())",
	)
	cmd.Dir = filepath.Clean(filepath.Join("..", ".."))
	body, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(body))
}

func waitForNATSWebSocket(rawURL string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		conn, _, _, err := ws.Dial(context.Background(), rawURL)
		if err == nil {
			_ = conn.Close()
			return nil
		}
		lastErr = err
		time.Sleep(50 * time.Millisecond)
	}
	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("timed out waiting for %s", rawURL)
}
