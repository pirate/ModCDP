package transport

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const DefaultUpstreamReverseWSBind = "127.0.0.1:29292"
const DefaultUpstreamReverseWSWaitTimeoutMS = 10_000

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
	closeCh       chan struct{}
	stateMu       sync.Mutex
	PeerInfo      map[string]any
	generation    int64
}

type ReverseWebSocketUpstreamTransportOptions struct {
	UpstreamReverseWSBind          string `json:"upstream_reversews_bind,omitempty"`
	UpstreamReverseWSWaitTimeoutMS int    `json:"upstream_reversews_wait_timeout_ms,omitempty"`
}

func NewReverseWebSocketUpstreamTransport(options ReverseWebSocketUpstreamTransportOptions) *ReverseWebSocketUpstreamTransport {
	reverseWSBind := options.UpstreamReverseWSBind
	if reverseWSBind == "" {
		reverseWSBind = DefaultUpstreamReverseWSBind
	}
	reverseWSWaitTimeoutMS := options.UpstreamReverseWSWaitTimeoutMS
	if reverseWSWaitTimeoutMS == 0 {
		reverseWSWaitTimeoutMS = DefaultUpstreamReverseWSWaitTimeoutMS
	}
	t := &ReverseWebSocketUpstreamTransport{WaitTimeoutMS: reverseWSWaitTimeoutMS, peerCh: make(chan struct{}), closeCh: make(chan struct{})}
	t.setBind(reverseWSBind)
	return t
}

func (t *ReverseWebSocketUpstreamTransport) setBind(bind string) {
	raw := bind
	if !strings.Contains(raw, "://") {
		raw = "ws://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		parsed, _ = url.Parse("ws://" + DefaultUpstreamReverseWSBind)
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
	if bind, _ := config["upstream_reversews_bind"].(string); bind != "" {
		t.setBind(bind)
	}
	if waitTimeoutMS, ok := intFromConfig(config["upstream_reversews_wait_timeout_ms"]); ok {
		t.WaitTimeoutMS = waitTimeoutMS
	}
}

func (t *ReverseWebSocketUpstreamTransport) Connect() error {
	t.stateMu.Lock()
	t.closeCh = make(chan struct{})
	t.stateMu.Unlock()
	listener, err := net.Listen("tcp", t.Bind)
	if err != nil {
		return err
	}
	t.Listener = listener
	go t.acceptLoop()
	return nil
}

func (t *ReverseWebSocketUpstreamTransport) Send(message map[string]any) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	conn := t.Conn
	if conn == nil {
		return fmt.Errorf("no reverse ModCDP extension peer is connected at %s", t.URL)
	}
	return wsutil.WriteServerText(conn, body)
}

func (t *ReverseWebSocketUpstreamTransport) GetInjectorConfig() ExtensionInjectorConfig {
	return ExtensionInjectorConfig{}
}

func (t *ReverseWebSocketUpstreamTransport) WaitForPeer() error {
	t.writeMu.Lock()
	connected := t.Conn != nil
	t.writeMu.Unlock()
	if connected {
		return nil
	}
	t.stateMu.Lock()
	closeCh := t.closeCh
	peerCh := t.peerCh
	t.stateMu.Unlock()
	select {
	case <-peerCh:
		return nil
	case <-closeCh:
		return fmt.Errorf("reverse websocket transport at %s closed before a peer connected", t.URL)
	case <-time.After(time.Duration(t.WaitTimeoutMS) * time.Millisecond):
		return fmt.Errorf("timed out waiting %dms for reverse ModCDP extension connection", t.WaitTimeoutMS)
	}
}

func (t *ReverseWebSocketUpstreamTransport) PeerGeneration() int64 {
	t.stateMu.Lock()
	defer t.stateMu.Unlock()
	return t.generation
}

func (t *ReverseWebSocketUpstreamTransport) Close() error {
	t.stateMu.Lock()
	closeCh := t.closeCh
	t.closeCh = make(chan struct{})
	t.peerCh = make(chan struct{})
	t.peerOnce = sync.Once{}
	t.stateMu.Unlock()
	close(closeCh)
	t.writeMu.Lock()
	if t.Conn != nil {
		_ = t.Conn.Close()
		t.Conn = nil
	}
	t.writeMu.Unlock()
	if t.Listener != nil {
		_ = t.Listener.Close()
		t.Listener = nil
	}
	t.stateMu.Lock()
	t.PeerInfo = nil
	t.stateMu.Unlock()
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
	_ = conn.SetReadDeadline(time.Now().Add(time.Duration(t.WaitTimeoutMS) * time.Millisecond))
	helloBytes, err := wsutil.ReadClientText(conn)
	_ = conn.SetReadDeadline(time.Time{})
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
	t.writeMu.Lock()
	if t.Conn != nil && t.Conn != conn {
		_ = t.Conn.Close()
	}
	t.Conn = conn
	t.writeMu.Unlock()
	t.stateMu.Lock()
	t.PeerInfo = hello
	t.generation++
	t.peerOnce.Do(func() { close(t.peerCh) })
	t.stateMu.Unlock()
	for {
		data, err := wsutil.ReadClientText(conn)
		if err != nil {
			t.writeMu.Lock()
			if t.Conn == conn {
				t.Conn = nil
				t.writeMu.Unlock()
				t.stateMu.Lock()
				t.PeerInfo = nil
				t.stateMu.Unlock()
				t.resetPeerWait()
			} else {
				t.writeMu.Unlock()
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

func (t *ReverseWebSocketUpstreamTransport) resetPeerWait() {
	t.stateMu.Lock()
	defer t.stateMu.Unlock()
	t.peerCh = make(chan struct{})
	t.peerOnce = sync.Once{}
}
