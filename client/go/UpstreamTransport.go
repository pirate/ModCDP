package modcdp

import (
	"fmt"
	"reflect"
)

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
	return func() {
		pointer := reflect.ValueOf(listener).Pointer()
		for index, candidate := range e.recvListeners {
			if reflect.ValueOf(candidate).Pointer() == pointer {
				e.recvListeners = append(e.recvListeners[:index], e.recvListeners[index+1:]...)
				return
			}
		}
	}
}

func (e *UpstreamTransport) OnClose(listener func(error)) func() {
	e.closeListeners = append(e.closeListeners, listener)
	return func() {
		pointer := reflect.ValueOf(listener).Pointer()
		for index, candidate := range e.closeListeners {
			if reflect.ValueOf(candidate).Pointer() == pointer {
				e.closeListeners = append(e.closeListeners[:index], e.closeListeners[index+1:]...)
				return
			}
		}
	}
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
