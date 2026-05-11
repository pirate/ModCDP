package transport

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const DefaultUpstreamNATSURL = "ws://127.0.0.1:4223"
const DefaultUpstreamNATSSubjectPrefix = "modcdp.default"
const DefaultUpstreamNATSWaitTimeoutMS = 10_000

type NatsUpstreamTransport struct {
	UpstreamTransport
	URL                       string
	UpstreamNATSSubjectPrefix string
	UpstreamNATSRole          string
	WaitTimeoutMS             int
	Conn                      net.Conn
	IsWebSocket               bool
	buffer                    string
	connected                 bool
	closed                    bool
	writeMu                   sync.Mutex
	peerCh                    chan struct{}
	peerOnce                  sync.Once
	closeCh                   chan struct{}
	stateMu                   sync.Mutex
}

type NatsUpstreamTransportOptions struct {
	UpstreamNATSURL           string `json:"upstream_nats_url,omitempty"`
	UpstreamNATSSubjectPrefix string `json:"upstream_nats_subject_prefix,omitempty"`
	UpstreamNATSRole          string `json:"upstream_nats_role,omitempty"`
	UpstreamNATSWaitTimeoutMS int    `json:"upstream_nats_wait_timeout_ms,omitempty"`
}

func NewNatsUpstreamTransport(options NatsUpstreamTransportOptions) *NatsUpstreamTransport {
	normalizedURL, subjectPrefix := normalizeUpstreamNATSURL(firstNonEmptyString(options.UpstreamNATSURL, DefaultUpstreamNATSURL), options.UpstreamNATSSubjectPrefix)
	role := firstNonEmptyString(options.UpstreamNATSRole, "client")
	waitTimeoutMS := options.UpstreamNATSWaitTimeoutMS
	if waitTimeoutMS == 0 {
		waitTimeoutMS = DefaultUpstreamNATSWaitTimeoutMS
	}
	return &NatsUpstreamTransport{
		URL:                       normalizedURL,
		UpstreamNATSSubjectPrefix: subjectPrefix,
		UpstreamNATSRole:          role,
		WaitTimeoutMS:             waitTimeoutMS,
		peerCh:                    make(chan struct{}),
		closeCh:                   make(chan struct{}),
	}
}

func (t *NatsUpstreamTransport) Update(config map[string]any) {
	if config == nil {
		return
	}
	natsURL, _ := config["upstream_nats_url"].(string)
	subjectPrefix, _ := config["upstream_nats_subject_prefix"].(string)
	if natsURL != "" || subjectPrefix != "" {
		t.URL, t.UpstreamNATSSubjectPrefix = normalizeUpstreamNATSURL(firstNonEmptyString(natsURL, t.URL), firstNonEmptyString(subjectPrefix, t.UpstreamNATSSubjectPrefix))
	}
	if role, _ := config["upstream_nats_role"].(string); role == "client" || role == "browser" {
		t.UpstreamNATSRole = role
	}
	if waitTimeoutMS, ok := intFromConfig(config["upstream_nats_wait_timeout_ms"]); ok {
		t.WaitTimeoutMS = waitTimeoutMS
	}
}

func (t *NatsUpstreamTransport) GetInjectorConfig() ExtensionInjectorConfig {
	return ExtensionInjectorConfig{UpstreamNATSURL: t.URL, UpstreamNATSSubjectPrefix: t.UpstreamNATSSubjectPrefix}
}

func (t *NatsUpstreamTransport) Connect() error {
	if t.connected {
		return nil
	}
	t.stateMu.Lock()
	t.closed = false
	t.closeCh = make(chan struct{})
	closeCh := t.closeCh
	t.stateMu.Unlock()
	if !validUpstreamNATSSubjectPrefix(t.UpstreamNATSSubjectPrefix) {
		return fmt.Errorf("invalid NATS subject prefix %q", t.UpstreamNATSSubjectPrefix)
	}
	parsed, err := url.Parse(t.URL)
	if err != nil {
		return err
	}
	switch parsed.Scheme {
	case "ws", "wss":
		conn, _, _, err := ws.Dial(context.Background(), t.URL)
		if err != nil {
			return err
		}
		t.Conn = conn
		t.IsWebSocket = true
		if err := t.writeProtocol("CONNECT " + mustJSON(connectNATSOptions()) + "\r\nPING\r\n"); err != nil {
			_ = conn.Close()
			return err
		}
		go t.readWebSocketLoop(conn, closeCh)
	case "nats", "tls":
		host := parsed.Hostname()
		if host == "" {
			host = "127.0.0.1"
		}
		port := parsed.Port()
		if port == "" {
			port = "4222"
		}
		var conn net.Conn
		if parsed.Scheme == "tls" {
			conn, err = tls.Dial("tcp", net.JoinHostPort(host, port), &tls.Config{ServerName: host})
		} else {
			conn, err = net.Dial("tcp", net.JoinHostPort(host, port))
		}
		if err != nil {
			return err
		}
		t.Conn = conn
		if err := t.writeProtocol("CONNECT " + mustJSON(connectNATSOptions()) + "\r\nPING\r\n"); err != nil {
			_ = conn.Close()
			return err
		}
		go t.readTCPLoop(conn, closeCh)
	default:
		return fmt.Errorf("upstream.upstream_mode=nats requires ws://, wss://, nats://, or tls:// URL, got %s", t.URL)
	}
	t.connected = true
	if err := t.subscribe(); err != nil {
		return err
	}
	return t.publish(t.outgoingSubject(), map[string]any{"type": "modcdp.nats.hello", "role": t.UpstreamNATSRole, "version": 1})
}

func (t *NatsUpstreamTransport) Send(message map[string]any) error {
	if !t.connected || t.Conn == nil {
		return fmt.Errorf("NATS transport is not connected")
	}
	return t.publish(t.outgoingSubject(), map[string]any{"type": "modcdp.nats.message", "message": message})
}

func (t *NatsUpstreamTransport) WaitForPeer() error {
	t.stateMu.Lock()
	closeCh := t.closeCh
	t.stateMu.Unlock()
	select {
	case <-t.peerCh:
		return nil
	case <-closeCh:
		return fmt.Errorf("NATS transport for %s closed before a peer connected", t.UpstreamNATSSubjectPrefix)
	case <-time.After(time.Duration(t.WaitTimeoutMS) * time.Millisecond):
		return fmt.Errorf("timed out waiting %dms for NATS ModCDP peer", t.WaitTimeoutMS)
	}
}

func (t *NatsUpstreamTransport) Close() error {
	t.stateMu.Lock()
	t.closed = true
	closeCh := t.closeCh
	t.closeCh = make(chan struct{})
	t.stateMu.Unlock()
	close(closeCh)
	t.writeMu.Lock()
	t.connected = false
	if t.Conn != nil {
		_ = t.Conn.Close()
		t.Conn = nil
	}
	t.writeMu.Unlock()
	t.buffer = ""
	t.peerCh = make(chan struct{})
	t.peerOnce = sync.Once{}
	return nil
}

func (t *NatsUpstreamTransport) Connected() bool {
	return t.connected
}

func (t *NatsUpstreamTransport) Closed() bool {
	return t.closed
}

func natsClosed(closeCh chan struct{}) bool {
	select {
	case <-closeCh:
		return true
	default:
		return false
	}
}

func (t *NatsUpstreamTransport) subscribe() error {
	return t.writeProtocol("SUB " + t.incomingSubject() + " 1\r\n")
}

func (t *NatsUpstreamTransport) publish(subject string, message map[string]any) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}
	return t.writeProtocol(fmt.Sprintf("PUB %s %d\r\n%s\r\n", subject, len(body), string(body)))
}

func (t *NatsUpstreamTransport) writeProtocol(data string) error {
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	conn := t.Conn
	if conn == nil {
		return fmt.Errorf("NATS transport is not connected")
	}
	if t.IsWebSocket {
		return wsutil.WriteClientText(conn, []byte(data))
	}
	_, err := conn.Write([]byte(data))
	return err
}

func (t *NatsUpstreamTransport) incomingSubject() string {
	if t.UpstreamNATSRole == "client" {
		return t.UpstreamNATSSubjectPrefix + ".browser_to_client"
	}
	return t.UpstreamNATSSubjectPrefix + ".client_to_browser"
}

func (t *NatsUpstreamTransport) outgoingSubject() string {
	if t.UpstreamNATSRole == "client" {
		return t.UpstreamNATSSubjectPrefix + ".client_to_browser"
	}
	return t.UpstreamNATSSubjectPrefix + ".browser_to_client"
}

func (t *NatsUpstreamTransport) readWebSocketLoop(conn net.Conn, closeCh chan struct{}) {
	for !natsClosed(closeCh) {
		data, _, err := wsutil.ReadServerData(conn)
		if err != nil {
			if !natsClosed(closeCh) {
				t.emitClose(err)
			}
			return
		}
		t.buffer = t.consumeProtocol(t.buffer + string(data))
	}
}

func (t *NatsUpstreamTransport) readTCPLoop(conn net.Conn, closeCh chan struct{}) {
	chunk := make([]byte, 65536)
	for !natsClosed(closeCh) {
		n, err := conn.Read(chunk)
		if err != nil {
			if !natsClosed(closeCh) {
				t.emitClose(err)
			}
			return
		}
		t.buffer = t.consumeProtocol(t.buffer + string(chunk[:n]))
	}
}

func (t *NatsUpstreamTransport) consumeProtocol(buffer string) string {
	for {
		lineEnd := strings.Index(buffer, "\r\n")
		if lineEnd < 0 {
			return buffer
		}
		line := buffer[:lineEnd]
		upper := strings.ToUpper(line)
		if strings.HasPrefix(upper, "MSG ") {
			parts := strings.Fields(line)
			size, err := strconv.Atoi(parts[len(parts)-1])
			payloadStart := lineEnd + 2
			payloadEnd := payloadStart + size
			if err != nil || len(buffer) < payloadEnd+2 {
				return buffer
			}
			payload := buffer[payloadStart:payloadEnd]
			buffer = buffer[payloadEnd+2:]
			t.handlePayload(payload)
			continue
		}
		buffer = buffer[lineEnd+2:]
		if upper == "PING" {
			_ = t.writeProtocol("PONG\r\n")
		} else if strings.HasPrefix(upper, "-ERR") {
			t.emitClose(fmt.Errorf("NATS error: %s", line))
		}
	}
}

func (t *NatsUpstreamTransport) handlePayload(payload string) {
	var parsed any
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return
	}
	record, _ := parsed.(map[string]any)
	if record["type"] == "modcdp.nats.hello" {
		t.peerOnce.Do(func() { close(t.peerCh) })
		return
	}
	message := parsed
	if record["type"] == "modcdp.nats.message" {
		message = record["message"]
	}
	if cdpMessage, ok := message.(map[string]any); ok {
		t.emitRecv(cdpMessage)
	}
}

func (t *NatsUpstreamTransport) HandlePayload(payload string) {
	t.handlePayload(payload)
}

func connectNATSOptions() map[string]any {
	return map[string]any{"verbose": false, "pedantic": false, "lang": "modcdp", "version": "1", "protocol": 1}
}

func normalizeUpstreamNATSURL(rawURL string, subjectPrefix string) (string, string) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, sanitizeUpstreamNATSSubjectPrefix(firstNonEmptyString(subjectPrefix, DefaultUpstreamNATSSubjectPrefix))
	}
	query := parsed.Query()
	subject := firstNonEmptyString(subjectPrefix, query.Get("upstream_nats_subject_prefix"), DefaultUpstreamNATSSubjectPrefix)
	query.Del("upstream_nats_subject_prefix")
	parsed.RawQuery = query.Encode()
	if parsed.Path == "" && (parsed.Scheme == "ws" || parsed.Scheme == "wss") {
		parsed.Path = "/"
	}
	return parsed.String(), sanitizeUpstreamNATSSubjectPrefix(subject)
}

func sanitizeUpstreamNATSSubjectPrefix(value string) string {
	return strings.TrimSpace(value)
}

func validUpstreamNATSSubjectPrefix(value string) bool {
	subject := strings.TrimSpace(value)
	return subject != "" && !strings.ContainsAny(subject, " \t\r\n*>")
}

func mustJSON(value any) string {
	body, _ := json.Marshal(value)
	return string(body)
}
