package modcdp

import "testing"

func TestRootExportsConcreteLaunchersInjectorsAndTransports(t *testing.T) {
	if NewLocalBrowserLauncher(LaunchOptions{}) == nil {
		t.Fatal("NewLocalBrowserLauncher returned nil")
	}
	if NewRemoteBrowserLauncher(LaunchOptions{}, "ws://127.0.0.1:9222/devtools/browser/test") == nil {
		t.Fatal("NewRemoteBrowserLauncher returned nil")
	}
	if NewBrowserbaseBrowserLauncher(LaunchOptions{}) == nil {
		t.Fatal("NewBrowserbaseBrowserLauncher returned nil")
	}
	if NewNoopBrowserLauncher(LaunchOptions{}) == nil {
		t.Fatal("NewNoopBrowserLauncher returned nil")
	}

	extensionInjector := NewExtensionInjector(ExtensionInjectorConfig{})
	discoveredInjector := NewDiscoveredExtensionInjector(ExtensionInjectorConfig{})
	bbInjector := NewBBBrowserExtensionInjector(ExtensionInjectorConfig{})
	localLaunchInjector := NewLocalBrowserLaunchExtensionInjector(ExtensionInjectorConfig{})
	loadUnpackedInjector := NewExtensionsLoadUnpackedInjector(ExtensionInjectorConfig{})
	borrowedInjector := NewBorrowedExtensionInjector(ExtensionInjectorConfig{})
	_ = []any{extensionInjector, discoveredInjector, bbInjector, localLaunchInjector, loadUnpackedInjector, borrowedInjector}

	if NewWebSocketUpstreamTransport(WebSocketUpstreamTransportOptions{}) == nil {
		t.Fatal("NewWebSocketUpstreamTransport returned nil")
	}
	if NewPipeUpstreamTransport(PipeUpstreamTransportOptions{}) == nil {
		t.Fatal("NewPipeUpstreamTransport returned nil")
	}
	if NewReverseWebSocketUpstreamTransport(ReverseWebSocketUpstreamTransportOptions{}) == nil {
		t.Fatal("NewReverseWebSocketUpstreamTransport returned nil")
	}
	if NewNativeMessagingUpstreamTransport(NativeMessagingUpstreamTransportOptions{}) == nil {
		t.Fatal("NewNativeMessagingUpstreamTransport returned nil")
	}
	if NewNatsUpstreamTransport(NatsUpstreamTransportOptions{}) == nil {
		t.Fatal("NewNatsUpstreamTransport returned nil")
	}

	if UpstreamModeWS != "ws" || UpstreamModePipe != "pipe" || UpstreamModeNativeMessaging != "nativemessaging" || UpstreamModeReverseWS != "reversews" || UpstreamModeNATS != "nats" {
		t.Fatal("upstream mode constants drifted")
	}
	if UpstreamEndpointKindRawCDP != "raw_cdp" || UpstreamEndpointKindModCDPServer != "modcdp_server" {
		t.Fatal("upstream endpoint kind constants drifted")
	}
}
