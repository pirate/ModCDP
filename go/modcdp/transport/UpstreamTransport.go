package transport

import (
	"fmt"
	"net"
	"sync"

	"github.com/pirate/ModCDP/go/modcdp/injector"
	"github.com/pirate/ModCDP/go/modcdp/launcher"
	"github.com/pirate/ModCDP/go/modcdp/types"
)

type ExtensionInjectorConfig = types.ExtensionInjectorConfig
type LaunchOptions = types.LaunchOptions

const DefaultModCDPExtensionID = injector.DefaultModCDPExtensionID

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func boolPtr(value bool) *bool {
	return &value
}

func freePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func websocketURLFor(endpoint string) (string, error) {
	return launcher.WebsocketURLFor(endpoint)
}

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
	recvListeners  []recvListener
	closeListeners []closeListener
	listenerMu     sync.Mutex
	nextListenerID int64
}

type recvListener struct {
	id int64
	fn func(map[string]any)
}

type closeListener struct {
	id int64
	fn func(error)
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
	e.listenerMu.Lock()
	e.nextListenerID++
	id := e.nextListenerID
	e.recvListeners = append(e.recvListeners, recvListener{id: id, fn: listener})
	e.listenerMu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			e.listenerMu.Lock()
			defer e.listenerMu.Unlock()
			for index, candidate := range e.recvListeners {
				if candidate.id != id {
					continue
				}
				e.recvListeners = append(e.recvListeners[:index], e.recvListeners[index+1:]...)
				return
			}
		})
	}
}

func (e *UpstreamTransport) OnClose(listener func(error)) func() {
	e.listenerMu.Lock()
	e.nextListenerID++
	id := e.nextListenerID
	e.closeListeners = append(e.closeListeners, closeListener{id: id, fn: listener})
	e.listenerMu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			e.listenerMu.Lock()
			defer e.listenerMu.Unlock()
			for index, candidate := range e.closeListeners {
				if candidate.id != id {
					continue
				}
				e.closeListeners = append(e.closeListeners[:index], e.closeListeners[index+1:]...)
				return
			}
		})
	}
}

func (e *UpstreamTransport) emitRecv(message map[string]any) {
	e.listenerMu.Lock()
	listeners := append([]recvListener(nil), e.recvListeners...)
	e.listenerMu.Unlock()
	for _, listener := range listeners {
		listener.fn(message)
	}
}

func (e *UpstreamTransport) EmitRecv(message map[string]any) {
	e.emitRecv(message)
}

func (e *UpstreamTransport) emitClose(err error) {
	e.listenerMu.Lock()
	listeners := append([]closeListener(nil), e.closeListeners...)
	e.listenerMu.Unlock()
	for _, listener := range listeners {
		listener.fn(err)
	}
}

func (e *UpstreamTransport) EmitClose(err error) {
	e.emitClose(err)
}

func (e *UpstreamTransport) WaitForPeer() error {
	return nil
}

func (e *UpstreamTransport) PeerGeneration() int64 {
	return 0
}

func EndpointKindForUpstream(mode string) UpstreamEndpointKind {
	if mode == "ws" || mode == "pipe" {
		return UpstreamEndpointKindRawCDP
	}
	return UpstreamEndpointKindModCDPServer
}

func intFromConfig(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case float32:
		return int(typed), true
	default:
		return 0, false
	}
}
