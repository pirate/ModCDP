package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type WebSocketUpstreamTransport struct {
	UpstreamTransport
	URL     string
	Conn    net.Conn
	writeMu sync.Mutex
	stateMu sync.Mutex
	ctx     context.Context
	cancel  context.CancelFunc
	closed  bool
}

type WebSocketUpstreamTransportOptions struct {
	CDPURL string `json:"cdp_url,omitempty"`
}

func NewWebSocketUpstreamTransport(options WebSocketUpstreamTransportOptions) *WebSocketUpstreamTransport {
	return &WebSocketUpstreamTransport{URL: options.CDPURL}
}

func (t *WebSocketUpstreamTransport) Update(config map[string]any) {
	if config == nil {
		return
	}
	if value, ok := config["cdp_url"].(string); ok && value != "" {
		t.URL = value
	}
}

func (t *WebSocketUpstreamTransport) GetServerConfig() map[string]any {
	if t.URL == "" {
		return map[string]any{}
	}
	return map[string]any{"server_loopback_cdp_url": t.URL}
}

func (t *WebSocketUpstreamTransport) Connect() error {
	if t.URL == "" {
		return fmt.Errorf("upstream.upstream_mode=ws requires upstream.upstream_cdp_url or launcher-provided cdp_url")
	}
	// URL may start as an HTTP cdp_url; from here on it is the resolved WebSocket CDP endpoint.
	resolvedURL, err := websocketURLFor(t.URL)
	if err != nil {
		return err
	}
	t.URL = resolvedURL
	t.ctx, t.cancel = context.WithCancel(context.Background())
	conn, _, _, err := ws.Dial(t.ctx, t.URL)
	if err != nil {
		return err
	}
	t.writeMu.Lock()
	t.Conn = conn
	t.writeMu.Unlock()
	t.setClosed(false)
	go t.readLoop(conn)
	return nil
}

func (t *WebSocketUpstreamTransport) Send(message map[string]any) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	conn := t.Conn
	if conn == nil {
		return fmt.Errorf("CDP websocket is not connected")
	}
	return wsutil.WriteClientText(conn, body)
}

func (t *WebSocketUpstreamTransport) Close() error {
	t.setClosed(true)
	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	if t.Conn != nil {
		err := t.Conn.Close()
		t.Conn = nil
		return err
	}
	return nil
}

func (t *WebSocketUpstreamTransport) readLoop(conn net.Conn) {
	for !t.isClosed() {
		data, err := wsutil.ReadServerText(conn)
		if err != nil {
			if !t.isClosed() {
				t.emitClose(err)
			}
			return
		}
		var message map[string]any
		if err := json.Unmarshal(data, &message); err == nil {
			t.emitRecv(message)
		}
	}
}

func (t *WebSocketUpstreamTransport) setClosed(closed bool) {
	t.stateMu.Lock()
	t.closed = closed
	t.stateMu.Unlock()
}

func (t *WebSocketUpstreamTransport) isClosed() bool {
	t.stateMu.Lock()
	defer t.stateMu.Unlock()
	return t.closed
}
