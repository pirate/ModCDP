package modcdp

import "fmt"

type UpstreamMode string
type UpstreamEndpointKind string

const (
	UpstreamModeWS              UpstreamMode = "ws"
	UpstreamModePipe            UpstreamMode = "pipe"
	UpstreamModeNativeMessaging UpstreamMode = "nativemessaging"
	UpstreamModeReverseWS       UpstreamMode = "reversews"
	UpstreamModeNATS            UpstreamMode = "nats"

	UpstreamEndpointKindRawCDP       UpstreamEndpointKind = "raw_cdp"
	UpstreamEndpointKindModCDPServer UpstreamEndpointKind = "modcdp_server"
)

type UpstreamTransport struct {
	recvListeners  []func(map[string]any)
	closeListeners []func(error)
}

func (e *UpstreamTransport) Update(config map[string]any) {
}

func (e *UpstreamTransport) Connect() error {
	return fmt.Errorf("%T.Connect is not implemented", e)
}

func (e *UpstreamTransport) Close() error {
	return nil
}

func (e *UpstreamTransport) Send(message map[string]any) error {
	return fmt.Errorf("%T.Send is not implemented", e)
}

func (e *UpstreamTransport) GetInjectorConfig() ExtensionInjectorConfig {
	return ExtensionInjectorConfig{}
}

func (e *UpstreamTransport) GetLauncherConfig() LaunchOptions {
	return LaunchOptions{}
}

func (e *UpstreamTransport) GetServerConfig() map[string]any {
	return map[string]any{}
}

func (e *UpstreamTransport) OnRecv(listener func(map[string]any)) func() {
	e.recvListeners = append(e.recvListeners, listener)
	return func() {}
}

func (e *UpstreamTransport) OnClose(listener func(error)) func() {
	e.closeListeners = append(e.closeListeners, listener)
	return func() {}
}

func (e *UpstreamTransport) emitRecv(message map[string]any) {
	for _, listener := range e.recvListeners {
		listener(message)
	}
}

func (e *UpstreamTransport) emitClose(err error) {
	for _, listener := range e.closeListeners {
		listener(err)
	}
}

func (e *UpstreamTransport) WaitForPeer() error {
	return nil
}

func endpointKindForUpstream(mode string) UpstreamEndpointKind {
	if mode == "ws" || mode == "pipe" {
		return UpstreamEndpointKindRawCDP
	}
	return UpstreamEndpointKindModCDPServer
}

func unimplementedUpstream(mode string) error {
	return fmt.Errorf("upstream.mode=%s is not implemented by the Go client yet", mode)
}
