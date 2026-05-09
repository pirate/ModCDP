package modcdp

import (
	"fmt"
	"sync"
	"time"
)

type AutoSessionRouterSend func(method string, params map[string]any, sessionID string) (map[string]any, error)

type AutoSessionRouter struct {
	TargetSessions                   map[string]string
	SessionTargets                   map[string]map[string]any
	ExecutionContexts                map[string]int
	send                             AutoSessionRouterSend
	defaultExecutionContextTimeoutMS func() int
	executionContextWaiters          map[string][]chan int
	mu                               sync.Mutex
}

func NewAutoSessionRouter(send AutoSessionRouterSend, defaultExecutionContextTimeoutMS func() int) *AutoSessionRouter {
	return &AutoSessionRouter{
		TargetSessions:                   map[string]string{},
		SessionTargets:                   map[string]map[string]any{},
		ExecutionContexts:                map[string]int{},
		send:                             send,
		defaultExecutionContextTimeoutMS: defaultExecutionContextTimeoutMS,
		executionContextWaiters:          map[string][]chan int{},
	}
}

func (r *AutoSessionRouter) SessionIDForTarget(targetID string) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.TargetSessions[targetID]
}

func (r *AutoSessionRouter) AttachToTarget(targetID string) string {
	if sessionID := r.SessionIDForTarget(targetID); sessionID != "" {
		return sessionID
	}
	result, err := r.send("Target.attachToTarget", map[string]any{"targetId": targetID, "flatten": true}, "")
	if err != nil {
		return ""
	}
	sessionID, _ := result["sessionId"].(string)
	return sessionID
}

func (r *AutoSessionRouter) RecordProtocolEvent(method string, data any, sessionID string) {
	eventData, _ := data.(map[string]any)
	if eventData == nil {
		eventData = map[string]any{}
	}
	switch method {
	case "Target.attachedToTarget":
		attachedSessionID, _ := eventData["sessionId"].(string)
		if attachedSessionID == "" {
			attachedSessionID = sessionID
		}
		targetInfo, _ := eventData["targetInfo"].(map[string]any)
		targetID, _ := targetInfo["targetId"].(string)
		if attachedSessionID != "" && targetID != "" {
			r.mu.Lock()
			r.TargetSessions[targetID] = attachedSessionID
			r.SessionTargets[attachedSessionID] = targetInfo
			r.mu.Unlock()
		}
	case "Runtime.executionContextCreated":
		context, _ := eventData["context"].(map[string]any)
		contextID, ok := intFromAny(context["id"])
		if sessionID != "" && ok {
			r.recordExecutionContext(sessionID, contextID)
		}
	case "Target.detachedFromTarget":
		detachedSessionID, _ := eventData["sessionId"].(string)
		if detachedSessionID == "" {
			detachedSessionID = sessionID
		}
		if detachedSessionID != "" {
			r.forgetSession(detachedSessionID)
		}
	}
}

func (r *AutoSessionRouter) WaitForExecutionContext(sessionID string, timeoutMS int) (int, error) {
	if timeoutMS == 0 {
		timeoutMS = r.defaultExecutionContextTimeoutMS()
	}
	if sessionID == "" {
		return 0, fmt.Errorf("cannot wait for a Runtime execution context without a session")
	}
	r.mu.Lock()
	if contextID, ok := r.ExecutionContexts[sessionID]; ok {
		r.mu.Unlock()
		return contextID, nil
	}
	waiter := make(chan int, 1)
	r.executionContextWaiters[sessionID] = append(r.executionContextWaiters[sessionID], waiter)
	r.mu.Unlock()

	select {
	case contextID := <-waiter:
		return contextID, nil
	case <-time.After(time.Duration(timeoutMS) * time.Millisecond):
		r.mu.Lock()
		waiters := r.executionContextWaiters[sessionID]
		filtered := waiters[:0]
		for _, candidate := range waiters {
			if candidate != waiter {
				filtered = append(filtered, candidate)
			}
		}
		if len(filtered) == 0 {
			delete(r.executionContextWaiters, sessionID)
		} else {
			r.executionContextWaiters[sessionID] = filtered
		}
		r.mu.Unlock()
		return 0, fmt.Errorf("timed out waiting for Runtime.executionContextCreated for session %s", sessionID)
	}
}

func (r *AutoSessionRouter) recordExecutionContext(sessionID string, contextID int) {
	r.mu.Lock()
	r.ExecutionContexts[sessionID] = contextID
	waiters := r.executionContextWaiters[sessionID]
	delete(r.executionContextWaiters, sessionID)
	r.mu.Unlock()
	for _, waiter := range waiters {
		waiter <- contextID
	}
}

func (r *AutoSessionRouter) forgetSession(sessionID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	targetInfo := r.SessionTargets[sessionID]
	delete(r.SessionTargets, sessionID)
	if targetID, _ := targetInfo["targetId"].(string); targetID != "" {
		delete(r.TargetSessions, targetID)
	}
	delete(r.ExecutionContexts, sessionID)
}

func intFromAny(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	default:
		return 0, false
	}
}
