package modcdp

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func TestAutoSessionRouterTracksRealTargetSessionsAndExecutionContexts(t *testing.T) {
	headless := true
	sandbox := false
	chrome, err := NewLocalBrowserLauncher(LaunchOptions{
		Headless: &headless,
		Sandbox:  &sandbox,
	}).Launch(LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()
	conn, _, _, err := ws.Dial(context.Background(), chrome.WSURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	type pendingResponse struct {
		ch chan map[string]any
	}
	nextID := int64(0)
	pending := map[int64]pendingResponse{}
	done := make(chan struct{})
	var writeMu sync.Mutex
	var pendingMu sync.Mutex
	var router *AutoSessionRouter
	send := func(method string, params map[string]any, sessionID string) (map[string]any, error) {
		pendingMu.Lock()
		nextID += 1
		id := nextID
		response := pendingResponse{ch: make(chan map[string]any, 1)}
		pending[id] = response
		pendingMu.Unlock()
		message := map[string]any{"id": id, "method": method, "params": params}
		if sessionID != "" {
			message["sessionId"] = sessionID
		}
		body, _ := json.Marshal(message)
		writeMu.Lock()
		err := wsutil.WriteClientText(conn, body)
		writeMu.Unlock()
		if err != nil {
			return nil, err
		}
		select {
		case received := <-response.ch:
			if rawError, ok := received["error"]; ok {
				return nil, fmt.Errorf("%v", rawError)
			}
			result, _ := received["result"].(map[string]any)
			if result == nil {
				result = map[string]any{}
			}
			return result, nil
		case <-time.After(10 * time.Second):
			return nil, fmt.Errorf("%s timed out", method)
		}
	}
	router = NewAutoSessionRouter(send, func() int { return 5000 })

	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			data, err := wsutil.ReadServerText(conn)
			if err != nil {
				return
			}
			var message map[string]any
			if err := json.Unmarshal(data, &message); err != nil {
				return
			}
			if id, ok := int64FromAny(message["id"]); ok {
				pendingMu.Lock()
				response, found := pending[id]
				delete(pending, id)
				pendingMu.Unlock()
				if found {
					response.ch <- message
				}
				continue
			}
			method, _ := message["method"].(string)
			sessionID, _ := message["sessionId"].(string)
			router.RecordProtocolEvent(method, message["params"], sessionID)
		}
	}()

	var targetID string
	defer func() {
		if targetID != "" {
			_, _ = send("Target.closeTarget", map[string]any{"targetId": targetID}, "")
		}
		close(done)
	}()

	if _, err := send("Target.setAutoAttach", map[string]any{"autoAttach": true, "waitForDebuggerOnStart": false, "flatten": true}, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := send("Target.setDiscoverTargets", map[string]any{"discover": true}, ""); err != nil {
		t.Fatal(err)
	}
	created, err := send("Target.createTarget", map[string]any{"url": "about:blank#modcdp-auto-session-router"}, "")
	if err != nil {
		t.Fatal(err)
	}
	targetID, _ = created["targetId"].(string)
	sessionID := waitForString(t, func() string { return router.SessionIDForTarget(targetID) })
	contextResult := make(chan int, 1)
	contextError := make(chan error, 1)
	go func() {
		contextID, err := router.WaitForExecutionContext(sessionID, 5000)
		if err != nil {
			contextError <- err
			return
		}
		contextResult <- contextID
	}()
	if _, err := send("Runtime.enable", map[string]any{}, sessionID); err != nil {
		t.Fatal(err)
	}
	select {
	case contextID := <-contextResult:
		if contextID == 0 {
			t.Fatal("context id was zero")
		}
	case err := <-contextError:
		t.Fatal(err)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for execution context")
	}
	if _, err := send("Target.detachFromTarget", map[string]any{"sessionId": sessionID}, ""); err != nil {
		t.Fatal(err)
	}
	waitForString(t, func() string {
		if router.SessionIDForTarget(targetID) == "" {
			return "detached"
		}
		return ""
	})
}

func waitForString(t *testing.T, fn func() string) string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		value := fn()
		if value != "" {
			return value
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("timed out waiting for string")
	return ""
}

func int64FromAny(value any) (int64, bool) {
	switch typed := value.(type) {
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case float64:
		return int64(typed), true
	default:
		return 0, false
	}
}
