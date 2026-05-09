package modcdp

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const DefaultReverseWSBind = "127.0.0.1:29292"
const DefaultReverseWSWaitTimeoutMS = 10_000

type ReverseWebSocketUpstreamTransport struct {
	UpstreamTransport
	Bind          string
	URL           string
	WaitTimeoutMS int
	Listener      net.Listener
	Conn          net.Conn
	writeMu       sync.Mutex
	peerCh        chan struct{}
	peerOnce      sync.Once
	PeerInfo      map[string]any
}

func NewReverseWebSocketUpstreamTransport(bind string) *ReverseWebSocketUpstreamTransport {
	if bind == "" {
		bind = DefaultReverseWSBind
	}
	t := &ReverseWebSocketUpstreamTransport{WaitTimeoutMS: DefaultReverseWSWaitTimeoutMS, peerCh: make(chan struct{})}
	t.SetBind(bind)
	return t
}

func (t *ReverseWebSocketUpstreamTransport) SetBind(bind string) {
	parsed, err := url.Parse(bind)
	if err != nil || parsed.Scheme == "" {
		parsed, _ = url.Parse("ws://" + bind)
	}
	host := parsed.Hostname()
	if host == "" {
		host = "127.0.0.1"
	}
	port := parsed.Port()
	if port == "" {
		port = "29292"
	}
	t.Bind = net.JoinHostPort(host, port)
	t.URL = "ws://" + t.Bind
}

func (t *ReverseWebSocketUpstreamTransport) Update(config map[string]any) {
	if config == nil {
		return
	}
	if bind, _ := config["reversews_bind"].(string); bind != "" {
		t.SetBind(bind)
	} else if rawURL, _ := config["url"].(string); rawURL != "" {
		t.SetBind(rawURL)
	}
}

func (t *ReverseWebSocketUpstreamTransport) Connect() error {
	listener, err := net.Listen("tcp", t.Bind)
	if err != nil {
		return err
	}
	t.Listener = listener
	go t.acceptLoop()
	return nil
}

func (t *ReverseWebSocketUpstreamTransport) Send(message map[string]any) error {
	if t.Conn == nil {
		return fmt.Errorf("no reverse ModCDP extension peer is connected at %s", t.URL)
	}
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	return wsutil.WriteServerText(t.Conn, body)
}

func (t *ReverseWebSocketUpstreamTransport) GetInjectorConfig() ExtensionInjectorConfig {
	return ExtensionInjectorConfig{ReverseProxyURL: t.URL}
}

func (t *ReverseWebSocketUpstreamTransport) WaitForPeer() error {
	if t.Conn != nil {
		return nil
	}
	select {
	case <-t.peerCh:
		return nil
	case <-time.After(time.Duration(t.WaitTimeoutMS) * time.Millisecond):
		return fmt.Errorf("timed out waiting %dms for reverse ModCDP extension connection", t.WaitTimeoutMS)
	}
}

func (t *ReverseWebSocketUpstreamTransport) Close() error {
	if t.Conn != nil {
		_ = t.Conn.Close()
		t.Conn = nil
	}
	if t.Listener != nil {
		_ = t.Listener.Close()
		t.Listener = nil
	}
	return nil
}

func (t *ReverseWebSocketUpstreamTransport) acceptLoop() {
	for {
		conn, err := t.Listener.Accept()
		if err != nil {
			return
		}
		go t.accept(conn)
	}
}

func (t *ReverseWebSocketUpstreamTransport) accept(conn net.Conn) {
	if _, err := ws.Upgrade(conn); err != nil {
		_ = conn.Close()
		t.emitClose(err)
		return
	}
	helloBytes, err := wsutil.ReadClientText(conn)
	if err != nil {
		_ = conn.Close()
		t.emitClose(err)
		return
	}
	var hello map[string]any
	if err := json.Unmarshal(helloBytes, &hello); err != nil || hello["type"] != "modcdp.reverse.hello" {
		_ = conn.Close()
		if err == nil {
			err = fmt.Errorf("invalid reverse hello")
		}
		t.emitClose(err)
		return
	}
	if t.Conn != nil && t.Conn != conn {
		_ = t.Conn.Close()
	}
	t.Conn = conn
	t.PeerInfo = hello
	t.peerOnce.Do(func() { close(t.peerCh) })
	for {
		data, err := wsutil.ReadClientText(conn)
		if err != nil {
			if t.Conn == conn {
				t.Conn = nil
				t.PeerInfo = nil
			}
			t.emitClose(err)
			return
		}
		var message map[string]any
		if err := json.Unmarshal(data, &message); err == nil {
			t.emitRecv(message)
		}
	}
}
