package modcdp

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

const DefaultNATSURL = "ws://127.0.0.1:4223"
const DefaultNATSSubjectPrefix = "modcdp.default"
const DefaultNATSWaitTimeoutMS = 10_000

type NatsUpstreamTransport struct {
	UpstreamTransport
	URL           string
	SubjectPrefix string
	Role          string
	WaitTimeoutMS int
	Conn          net.Conn
	IsWebSocket   bool
	buffer        string
	connected     bool
	closed        bool
	writeMu       sync.Mutex
	peerCh        chan struct{}
	peerOnce      sync.Once
}

func NewNatsUpstreamTransport(natsURL string) *NatsUpstreamTransport {
	normalizedURL, subjectPrefix := normalizeNATSURL(firstNonEmptyString(natsURL, DefaultNATSURL), "")
	return &NatsUpstreamTransport{
		URL:           normalizedURL,
		SubjectPrefix: subjectPrefix,
		Role:          "client",
		WaitTimeoutMS: DefaultNATSWaitTimeoutMS,
		peerCh:        make(chan struct{}),
	}
}

func (t *NatsUpstreamTransport) Update(config map[string]any) {
	if config == nil {
		return
	}
	natsURL, _ := config["nats_url"].(string)
	subjectPrefix, _ := config["nats_subject_prefix"].(string)
	if natsURL != "" || subjectPrefix != "" {
		t.URL, t.SubjectPrefix = normalizeNATSURL(firstNonEmptyString(natsURL, t.URL), firstNonEmptyString(subjectPrefix, t.SubjectPrefix))
	}
}

func (t *NatsUpstreamTransport) GetInjectorConfig() ExtensionInjectorConfig {
	return ExtensionInjectorConfig{NATSURL: t.URL, NATSSubjectPrefix: t.SubjectPrefix}
}

func (t *NatsUpstreamTransport) Connect() error {
	if t.connected {
		return nil
	}
	if !validNATSSubjectPrefix(t.SubjectPrefix) {
		return fmt.Errorf("invalid NATS subject prefix %q", t.SubjectPrefix)
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
		go t.readWebSocketLoop()
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
		go t.readTCPLoop()
	default:
		return fmt.Errorf("upstream.mode=nats requires ws://, wss://, nats://, or tls:// URL, got %s", t.URL)
	}
	t.connected = true
	if err := t.subscribe(); err != nil {
		return err
	}
	return t.publish(t.outgoingSubject(), map[string]any{"type": "modcdp.nats.hello", "role": t.Role, "version": 1})
}

func (t *NatsUpstreamTransport) Send(message map[string]any) error {
	if !t.connected || t.Conn == nil {
		return fmt.Errorf("NATS transport is not connected")
	}
	return t.publish(t.outgoingSubject(), map[string]any{"type": "modcdp.nats.message", "message": message})
}

func (t *NatsUpstreamTransport) WaitForPeer() error {
	select {
	case <-t.peerCh:
		return nil
	case <-time.After(time.Duration(t.WaitTimeoutMS) * time.Millisecond):
		return fmt.Errorf("timed out waiting %dms for NATS ModCDP peer", t.WaitTimeoutMS)
	}
}

func (t *NatsUpstreamTransport) Close() error {
	t.closed = true
	t.connected = false
	if t.Conn != nil {
		_ = t.Conn.Close()
		t.Conn = nil
	}
	return nil
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
	if t.Conn == nil {
		return fmt.Errorf("NATS transport is not connected")
	}
	t.writeMu.Lock()
	defer t.writeMu.Unlock()
	if t.IsWebSocket {
		return wsutil.WriteClientText(t.Conn, []byte(data))
	}
	_, err := t.Conn.Write([]byte(data))
	return err
}

func (t *NatsUpstreamTransport) incomingSubject() string {
	if t.Role == "client" {
		return t.SubjectPrefix + ".browser_to_client"
	}
	return t.SubjectPrefix + ".client_to_browser"
}

func (t *NatsUpstreamTransport) outgoingSubject() string {
	if t.Role == "client" {
		return t.SubjectPrefix + ".client_to_browser"
	}
	return t.SubjectPrefix + ".browser_to_client"
}

func (t *NatsUpstreamTransport) readWebSocketLoop() {
	for !t.closed {
		data, _, err := wsutil.ReadServerData(t.Conn)
		if err != nil {
			if !t.closed {
				t.emitClose(err)
			}
			return
		}
		t.buffer = t.consumeProtocol(t.buffer + string(data))
	}
}

func (t *NatsUpstreamTransport) readTCPLoop() {
	chunk := make([]byte, 65536)
	for !t.closed {
		n, err := t.Conn.Read(chunk)
		if err != nil {
			if !t.closed {
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

func connectNATSOptions() map[string]any {
	return map[string]any{"verbose": false, "pedantic": false, "lang": "modcdp", "version": "1", "protocol": 1}
}

func normalizeNATSURL(rawURL string, subjectPrefix string) (string, string) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL, sanitizeNATSSubjectPrefix(firstNonEmptyString(subjectPrefix, DefaultNATSSubjectPrefix))
	}
	query := parsed.Query()
	subject := firstNonEmptyString(subjectPrefix, query.Get("subject"), query.Get("subject_prefix"), DefaultNATSSubjectPrefix)
	query.Del("subject")
	query.Del("subject_prefix")
	parsed.RawQuery = query.Encode()
	if parsed.Path == "" && (parsed.Scheme == "ws" || parsed.Scheme == "wss") {
		parsed.Path = "/"
	}
	return parsed.String(), sanitizeNATSSubjectPrefix(subject)
}

func sanitizeNATSSubjectPrefix(value string) string {
	return strings.TrimSpace(value)
}

func validNATSSubjectPrefix(value string) bool {
	subject := strings.TrimSpace(value)
	return subject != "" && !strings.ContainsAny(subject, " \t\r\n*>")
}

func mustJSON(value any) string {
	body, _ := json.Marshal(value)
	return string(body)
}
