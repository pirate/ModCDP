package modcdp

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
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewWebSocketUpstreamTransport(url ...string) *WebSocketUpstreamTransport {
	transport := &WebSocketUpstreamTransport{}
	if len(url) > 0 {
		transport.URL = url[0]
	}
	return transport
}

func (t *WebSocketUpstreamTransport) Update(config map[string]any) {
	if config == nil {
		return
	}
	for _, key := range []string{"ws_url", "cdp_url", "url"} {
		if value, ok := config[key].(string); ok && value != "" {
			t.URL = value
			return
		}
	}
}

func (t *WebSocketUpstreamTransport) GetServerConfig() map[string]any {
	if t.URL == "" {
		return map[string]any{}
	}
	return map[string]any{"loopback_cdp_url": t.URL}
}

func (t *WebSocketUpstreamTransport) Connect() error {
	if t.URL == "" {
		return fmt.Errorf("upstream.mode=ws requires upstream.ws_url or launcher-provided ws_url")
	}
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
	t.Conn = conn
	return nil
}

func (t *WebSocketUpstreamTransport) Send(message map[string]any) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	return wsutil.WriteClientText(t.Conn, body)
}

func (t *WebSocketUpstreamTransport) Close() error {
	if t.cancel != nil {
		t.cancel()
	}
	if t.Conn != nil {
		return t.Conn.Close()
	}
	return nil
}
