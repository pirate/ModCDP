package transport_test

import (
	"context"
	"encoding/json"
	"fmt"
	modcdp "github.com/browserbase/modcdp/go/modcdp/client"
	. "github.com/browserbase/modcdp/go/modcdp/transport"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func TestReverseWebSocketUpstreamTransportConfigOwnsBindUpdatesAndWaitTimeout(t *testing.T) {
	transport := NewReverseWebSocketUpstreamTransport(ReverseWebSocketUpstreamTransportOptions{UpstreamReverseWSBind: "127.0.0.1:29292", UpstreamReverseWSWaitTimeoutMS: 10})
	if transport.URL != "ws://127.0.0.1:29292" {
		t.Fatalf("URL = %q", transport.URL)
	}
	if !reflect.DeepEqual(transport.GetInjectorConfig(), ExtensionInjectorConfig{}) {
		t.Fatalf("injector config = %#v", transport.GetInjectorConfig())
	}
	transport.Update(map[string]any{"upstream_reversews_bind": "127.0.0.1:29293", "upstream_reversews_wait_timeout_ms": 5})
	if transport.URL != "ws://127.0.0.1:29293" {
		t.Fatalf("URL after update = %q", transport.URL)
	}
	if !reflect.DeepEqual(transport.GetInjectorConfig(), ExtensionInjectorConfig{}) {
		t.Fatalf("injector config after update = %#v", transport.GetInjectorConfig())
	}
	transport.Update(map[string]any{"upstream_reversews_bind": "http://127.0.0.1:29294"})
	if transport.URL != "ws://127.0.0.1:29294" {
		t.Fatalf("URL after http update = %q", transport.URL)
	}
	if err := transport.WaitForPeer(); err == nil || !strings.Contains(err.Error(), "timed out waiting 5ms") {
		t.Fatalf("WaitForPeer error = %v", err)
	}
}

func TestReverseWebSocketUpstreamTransportSendBeforePeerErrorsImmediately(t *testing.T) {
	transport := NewReverseWebSocketUpstreamTransport(ReverseWebSocketUpstreamTransportOptions{UpstreamReverseWSBind: "127.0.0.1:29292", UpstreamReverseWSWaitTimeoutMS: 5_000})
	started := time.Now()
	err := transport.Send(map[string]any{"id": 1, "method": "Browser.getVersion"})
	if err == nil || !strings.Contains(err.Error(), "no reverse ModCDP extension peer is connected at ws://127.0.0.1:29292") {
		t.Fatalf("Send error = %v", err)
	}
	if elapsed := time.Since(started); elapsed > 250*time.Millisecond {
		t.Fatalf("Send waited for peer: elapsed = %s", elapsed)
	}
}

func TestReverseWebSocketUpstreamTransportCloseResetsPeerWaitState(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}
	transport := NewReverseWebSocketUpstreamTransport(ReverseWebSocketUpstreamTransportOptions{UpstreamReverseWSBind: fmt.Sprintf("127.0.0.1:%d", port), UpstreamReverseWSWaitTimeoutMS: 5})
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
	transport := NewReverseWebSocketUpstreamTransport(ReverseWebSocketUpstreamTransportOptions{UpstreamReverseWSBind: fmt.Sprintf("127.0.0.1:%d", port), UpstreamReverseWSWaitTimeoutMS: 5})
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
	transport := NewReverseWebSocketUpstreamTransport(ReverseWebSocketUpstreamTransportOptions{UpstreamReverseWSBind: fmt.Sprintf("127.0.0.1:%d", port), UpstreamReverseWSWaitTimeoutMS: 500})
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
	transport := NewReverseWebSocketUpstreamTransport(ReverseWebSocketUpstreamTransportOptions{UpstreamReverseWSBind: fmt.Sprintf("127.0.0.1:%d", port), UpstreamReverseWSWaitTimeoutMS: 5_000})
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
		err := transport.Send(map[string]any{"id": 99, "method": "Browser.getVersion"})
		if err != nil && strings.Contains(err.Error(), "no reverse ModCDP extension peer is connected") {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("timed out waiting for reverse peer disconnect")
}

func TestReverseWebSocketUpstreamTransportAcceptsRealExtensionReverseConnectionAndRoutesCDPThroughChromeDebugger(t *testing.T) {
	extensionPath, err := filepath.Abs(filepath.Join("..", "..", "..", "dist", "extension"))
	if err != nil {
		t.Fatal(err)
	}
	headless := runtime.GOOS == "linux" && os.Getenv("DISPLAY") == ""
	cdp := modcdp.New(modcdp.Options{
		Launcher: modcdp.LauncherConfig{LauncherMode: "local",
			LauncherOptions: modcdp.LaunchOptions{
				Headless: boolPtr(headless),
				// Reversews is browser -> client only. After explicit CHROME_PATH and
				// CI /usr/bin/chromium, these tests use Chrome for Testing because
				// Canary rejects --load-extension in this local test path.
				ExecutablePath: reverseWSTestBrowserPath(t),
			},
		},
		Upstream: modcdp.UpstreamConfig{UpstreamMode: "reversews"},
		Injector: modcdp.InjectorConfig{
			InjectorExtensionPath:               extensionPath,
			InjectorMode:                        "auto",
			InjectorServiceWorkerURLSuffixes:    []string{"/modcdp/service_worker.js"},
			InjectorTrustServiceWorkerTarget:    true,
			InjectorServiceWorkerProbeTimeoutMS: 1000,
		},
	})
	defer cdp.Close()

	if err := cdp.Connect(); err != nil {
		t.Fatal(err)
	}
	if cdp.ConnectTiming["upstream_endpoint_kind"] != UpstreamEndpointKindModCDPServer {
		t.Fatalf("upstream_endpoint_kind = %v", cdp.ConnectTiming["upstream_endpoint_kind"])
	}
	if cdp.Transport() == nil {
		t.Fatal("expected reverse transport to be connected")
	}
	transport, ok := cdp.Transport().(*ReverseWebSocketUpstreamTransport)
	if !ok {
		t.Fatalf("transport = %T", cdp.Transport())
	}
	if transport.URL != "ws://127.0.0.1:29292" {
		t.Fatalf("transport URL = %q", transport.URL)
	}
	if transport.PeerInfo["extension_id"] != DefaultModCDPExtensionID {
		t.Fatalf("extension_id = %#v", transport.PeerInfo["extension_id"])
	}
	result, err := cdp.Send("Runtime.evaluate", map[string]any{"expression": "location.href", "returnByValue": true})
	if err != nil {
		t.Fatal(err)
	}
	evaluated, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Runtime.evaluate result = %#v", result)
	}
	evaluatedResult, _ := evaluated["result"].(map[string]any)
	if evaluatedResult["value"] != "about:blank" {
		t.Fatalf("Runtime.evaluate value = %#v", evaluatedResult["value"])
	}
	time.Sleep(1500 * time.Millisecond)
	secondResult, err := cdp.Send("Runtime.evaluate", map[string]any{"expression": "document.readyState", "returnByValue": true})
	if err != nil {
		t.Fatal(err)
	}
	secondEvaluated, ok := secondResult.(map[string]any)
	if !ok {
		t.Fatalf("second Runtime.evaluate result = %#v", secondResult)
	}
	secondEvaluatedResult, _ := secondEvaluated["result"].(map[string]any)
	if secondEvaluatedResult["value"] != "complete" {
		t.Fatalf("second Runtime.evaluate value = %#v", secondEvaluatedResult["value"])
	}
}

func reverseWSTestBrowserPath(t *testing.T) string {
	t.Helper()
	explicitCandidates := []string{os.Getenv("CHROME_PATH")}
	if runtime.GOOS == "linux" {
		explicitCandidates = append(explicitCandidates, "/usr/bin/chromium")
	}
	for _, candidate := range explicitCandidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		home = "."
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = filepath.Join(home, "AppData", "Local")
	}
	var patterns []string
	switch runtime.GOOS {
	case "darwin":
		patterns = []string{
			filepath.Join(home, "Library", "Caches", "ms-playwright", "chromium-*", "chrome-mac*", "Google Chrome for Testing.app", "Contents", "MacOS", "Google Chrome for Testing"),
			filepath.Join(home, "Library", "Caches", "ms-playwright", "chromium-*", "chrome-mac*", "Chromium.app", "Contents", "MacOS", "Chromium"),
			filepath.Join(home, "Library", "Caches", "puppeteer", "chrome", "mac*-*", "chrome-mac*", "Google Chrome for Testing.app", "Contents", "MacOS", "Google Chrome for Testing"),
		}
	case "windows":
		patterns = []string{
			filepath.Join(localAppData, "ms-playwright", "chromium-*", "chrome-win*", "chrome.exe"),
			filepath.Join(home, ".cache", "puppeteer", "chrome", "win*-*", "chrome.exe"),
		}
	default:
		patterns = []string{
			filepath.Join(home, ".cache", "ms-playwright", "chromium-*", "chrome-linux*", "chrome"),
			filepath.Join("/opt", "pw-browsers", "chromium-*", "chrome-linux*", "chrome"),
			filepath.Join(home, ".cache", "puppeteer", "chrome", "linux-*", "chrome-linux*", "chrome"),
		}
	}
	var candidates []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		candidates = append(candidates, matches...)
	}
	candidates = newestChromeForTestingFirst(candidates)
	if len(candidates) > 0 {
		return candidates[0]
	}
	t.Fatal("Reversews tests require CHROME_PATH, /usr/bin/chromium, or Chrome for Testing.")
	return ""
}

func newestChromeForTestingFirst(candidates []string) []string {
	seen := map[string]bool{}
	deduped := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" || seen[candidate] {
			continue
		}
		seen[candidate] = true
		deduped = append(deduped, candidate)
	}
	sort.SliceStable(deduped, func(i, j int) bool {
		leftVersion := maxPathNumber(deduped[i])
		rightVersion := maxPathNumber(deduped[j])
		if leftVersion != rightVersion {
			return leftVersion > rightVersion
		}
		leftStat, leftErr := os.Stat(deduped[i])
		rightStat, rightErr := os.Stat(deduped[j])
		var leftMtime, rightMtime time.Time
		if leftErr == nil {
			leftMtime = leftStat.ModTime()
		}
		if rightErr == nil {
			rightMtime = rightStat.ModTime()
		}
		if !leftMtime.Equal(rightMtime) {
			return leftMtime.After(rightMtime)
		}
		return deduped[i] < deduped[j]
	})
	return deduped
}

func maxPathNumber(value string) int {
	maxValue := 0
	for _, raw := range regexp.MustCompile(`\d+`).FindAllString(value, -1) {
		number, err := strconv.Atoi(raw)
		if err == nil && number > maxValue {
			maxValue = number
		}
	}
	return maxValue
}
