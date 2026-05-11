package transport_test

import (
	. "github.com/pirate/ModCDP/go/modcdp/transport"
	"strings"
	"testing"
)

type testUpstreamTransport struct {
	UpstreamTransport
}

func (t *testUpstreamTransport) emit(message map[string]any) {
	t.EmitRecv(message)
}

func TestUpstreamTransportSharedConfigEndpointClassificationAndRecvCallbacks(t *testing.T) {
	transport := &UpstreamTransport{}
	received := []map[string]any{}
	stop := transport.OnRecv(func(message map[string]any) { received = append(received, message) })

	if EndpointKindForUpstream("ws") != UpstreamEndpointKindRawCDP {
		t.Fatal("ws endpoint kind mismatch")
	}
	if EndpointKindForUpstream("pipe") != UpstreamEndpointKindRawCDP {
		t.Fatal("pipe endpoint kind mismatch")
	}
	if EndpointKindForUpstream("nativemessaging") != UpstreamEndpointKindModCDPServer {
		t.Fatal("native endpoint kind mismatch")
	}
	if EndpointKindForUpstream("reversews") != UpstreamEndpointKindModCDPServer {
		t.Fatal("reverse endpoint kind mismatch")
	}
	if EndpointKindForUpstream("nats") != UpstreamEndpointKindModCDPServer {
		t.Fatal("nats endpoint kind mismatch")
	}
	transport.Update(nil)
	if len(transport.GetLauncherConfig().ExtraArgs) != 0 {
		t.Fatal("expected empty launcher config")
	}
	if transport.GetInjectorConfig().InjectorExtensionID != "" {
		t.Fatal("expected empty injector config")
	}
	if len(transport.GetServerConfig()) != 0 {
		t.Fatal("expected empty server config")
	}

	testTransport := &testUpstreamTransport{}
	parsed := []map[string]any{}
	testTransport.OnRecv(func(message map[string]any) { parsed = append(parsed, message) })
	testTransport.emit(map[string]any{"id": 1, "result": map[string]any{"ok": true}})
	testTransport.emit(map[string]any{"method": "Runtime.executionContextCreated", "params": map[string]any{}})
	if len(parsed) != 2 {
		t.Fatalf("parsed = %#v", parsed)
	}

	transport.EmitRecv(map[string]any{"method": "before.stop"})
	if len(received) != 1 {
		t.Fatalf("received = %#v", received)
	}
	stop()
	transport.EmitRecv(map[string]any{"method": "after.stop"})
	if len(received) != 1 {
		t.Fatalf("received after stop = %#v", received)
	}
	if err := transport.Connect(); err == nil || !strings.Contains(err.Error(), "Connect is not implemented") {
		t.Fatalf("connect error = %v", err)
	}
	if err := transport.Send(map[string]any{"id": 1, "method": "Browser.getVersion", "params": map[string]any{}}); err == nil || !strings.Contains(err.Error(), "Send is not implemented") {
		t.Fatalf("send error = %v", err)
	}
}
