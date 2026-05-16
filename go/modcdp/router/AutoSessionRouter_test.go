package router

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/browserbase/modcdp/go/modcdp/launcher"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func TestAutoSessionRouterRejectsPendingExecutionContextWaitersWhenSessionDetaches(t *testing.T) {
	router := NewAutoSessionRouter(func(string, map[string]any, string) (map[string]any, error) {
		return map[string]any{}, nil
	}, func() int { return 5000 })

	result := make(chan error, 1)
	go func() {
		_, err := router.WaitForExecutionContext("detached-session", 5000)
		result <- err
	}()
	waitForString(t, func() string {
		router.mu.Lock()
		defer router.mu.Unlock()
		if len(router.executionContextWaiters["detached-session"]) > 0 {
			return "waiting"
		}
		return ""
	})
	router.RecordProtocolEvent("Target.attachedToTarget", map[string]any{
		"sessionId":  "detached-session",
		"targetInfo": map[string]any{"targetId": "target-1", "type": "page"},
	}, "")
	router.RecordProtocolEvent("Target.detachedFromTarget", map[string]any{"sessionId": "detached-session"}, "")
	router.RecordProtocolEvent("Runtime.executionContextCreated", map[string]any{"context": map[string]any{"id": 42}}, "detached-session")

	select {
	case err := <-result:
		if err == nil || !strings.Contains(err.Error(), "Runtime execution context wait cancelled because session detached-session detached") {
			t.Fatalf("wait error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for detach error")
	}
	if sessionID := router.SessionIDForTarget("target-1"); sessionID != "" {
		t.Fatalf("session id after detach = %q", sessionID)
	}
	if _, ok := router.ExecutionContexts["detached-session"]; ok {
		t.Fatal("stale execution context was recorded after detach")
	}
}

func TestAutoSessionRouterBoundsDetachedSessionGuardsAndClearsThemWhenSessionReattaches(t *testing.T) {
	router := NewAutoSessionRouter(func(string, map[string]any, string) (map[string]any, error) {
		return map[string]any{}, nil
	}, func() int { return 5000 })

	for index := 0; index < 1034; index++ {
		router.RecordProtocolEvent("Target.detachedFromTarget", map[string]any{"sessionId": fmt.Sprintf("detached-session-%d", index)}, "")
	}

	router.mu.Lock()
	detachedCount := len(router.detachedSessions)
	detachedOrderCount := len(router.detachedSessionOrder)
	router.mu.Unlock()
	if detachedCount > maxDetachedSessionGuards {
		t.Fatalf("detached session guard count = %d, want <= %d", detachedCount, maxDetachedSessionGuards)
	}
	if detachedOrderCount > maxDetachedSessionGuards {
		t.Fatalf("detached session guard order count = %d, want <= %d", detachedOrderCount, maxDetachedSessionGuards)
	}

	recentSessionID := "detached-session-1033"
	router.RecordProtocolEvent("Runtime.executionContextCreated", map[string]any{"context": map[string]any{"id": 42}}, recentSessionID)
	if _, ok := router.ExecutionContexts[recentSessionID]; ok {
		t.Fatal("stale execution context was recorded for detached session")
	}

	router.RecordProtocolEvent("Target.attachedToTarget", map[string]any{
		"sessionId":  recentSessionID,
		"targetInfo": map[string]any{"targetId": "target-reattached", "type": "page"},
	}, "")
	router.RecordProtocolEvent("Runtime.executionContextCreated", map[string]any{"context": map[string]any{"id": 43}}, recentSessionID)

	if sessionID := router.SessionIDForTarget("target-reattached"); sessionID != recentSessionID {
		t.Fatalf("session id = %q, want %q", sessionID, recentSessionID)
	}
	if contextID := router.ExecutionContexts[recentSessionID]; contextID != 43 {
		t.Fatalf("context id = %d, want 43", contextID)
	}
	router.mu.Lock()
	defer router.mu.Unlock()
	if router.detachedSessions[recentSessionID] {
		t.Fatal("reattached session was still marked detached")
	}
	for _, detachedSessionID := range router.detachedSessionOrder {
		if detachedSessionID == recentSessionID {
			t.Fatal("reattached session stayed in detached session order")
		}
	}
}

func TestAutoSessionRouterTracksRealTargetSessionsAndExecutionContexts(t *testing.T) {
	headless := true
	chrome, err := launcher.NewLocalBrowserLauncher(launcher.LaunchOptions{
		Headless: &headless,
	}).Launch(launcher.LaunchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer chrome.Close()
	conn, _, _, err := ws.Dial(context.Background(), chrome.CDPURL)
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
	router = NewAutoSessionRouter(send, func() int { return 30000 })

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
		contextID, err := router.WaitForExecutionContext(sessionID, 30000)
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
	case <-time.After(35 * time.Second):
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
