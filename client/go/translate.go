package modcdp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const upstreamEventBindingName = "__ModCDP_event_from_upstream__"
const customEventBindingName = "__ModCDP_custom_event__"

func DefaultClientRoutes() map[string]string {
	return map[string]string{
		"Mod.*":    "service_worker",
		"Custom.*": "service_worker",
		"*.*":      "service_worker",
	}
}

func routeFor(method string, routes map[string]string) string {
	if route, ok := routes[method]; ok {
		return route
	}
	bestPrefixLen := -1
	bestRoute := ""
	for pattern, route := range routes {
		if pattern == "*.*" || !strings.HasSuffix(pattern, ".*") {
			continue
		}
		prefix := pattern[:len(pattern)-1]
		if strings.HasPrefix(method, prefix) && len(prefix) > bestPrefixLen {
			bestPrefixLen = len(prefix)
			bestRoute = route
		}
	}
	if bestPrefixLen >= 0 {
		return bestRoute
	}
	if route, ok := routes["*.*"]; ok {
		return route
	}
	return "direct_cdp"
}

type rawStep struct {
	Method    string
	Params    map[string]any
	Unwrap    string
	SessionID string
}

type rawCommand struct {
	Route  string
	Target string
	Steps  []rawStep
}

func evalParams(expression string) map[string]any {
	return map[string]any{
		"expression":                  expression,
		"awaitPromise":                true,
		"returnByValue":               true,
		"allowUnsafeEvalBlockedByCSP": true,
	}
}

func wrapModCDPEvaluate(params map[string]any, sessionID string) map[string]any {
	expr, _ := params["expression"].(string)
	userParams := params["params"]
	if userParams == nil {
		userParams = map[string]any{}
	}
	cdpSessionID, _ := params["cdpSessionId"].(string)
	if cdpSessionID == "" {
		cdpSessionID = sessionID
	}
	up, _ := json.Marshal(userParams)
	sid, _ := json.Marshal(cdpSessionID)
	return evalParams(fmt.Sprintf(
		`(async () => { const params = %s; const cdp = globalThis.ModCDP.attachToSession(%s); const ModCDP = globalThis.ModCDP; const chrome = globalThis.chrome; const value = (%s); return typeof value === 'function' ? await value(params) : value; })()`,
		string(up), string(sid), expr,
	))
}

func wrapModCDPAddCustomCommand(params map[string]any) map[string]any {
	name, _ := json.Marshal(params["name"])
	expr, _ := params["expression"].(string)
	exprJSON, _ := json.Marshal(expr)
	pSchema, _ := json.Marshal(params["params_schema"])
	rSchema, _ := json.Marshal(params["result_schema"])
	return evalParams(fmt.Sprintf(
		`(() => { return globalThis.ModCDP.addCustomCommand({ name: %s, params_schema: %s, result_schema: %s, expression: %s, handler: async (params, cdpSessionId, method) => { const cdp = globalThis.ModCDP.attachToSession(cdpSessionId); const ModCDP = globalThis.ModCDP; const chrome = globalThis.chrome; const handler = (%s); return await handler(params || {}, method); }, }); })()`,
		string(name), string(pSchema), string(rSchema), string(exprJSON), expr,
	))
}

func wrapModCDPAddCustomEvent(params map[string]any) map[string]any {
	rawName, _ := params["name"].(string)
	name, _ := json.Marshal(rawName)
	pSchema, _ := json.Marshal(params["event_schema"])
	return evalParams(fmt.Sprintf(
		`globalThis.ModCDP.addCustomEvent({ name: %s, event_schema: %s })`,
		string(name), string(pSchema),
	))
}

func wrapModCDPAddMiddleware(params map[string]any) map[string]any {
	name := params["name"]
	if name == nil {
		name = "*"
	}
	rawExpr, _ := params["expression"].(string)
	nameJSON, _ := json.Marshal(name)
	phaseJSON, _ := json.Marshal(params["phase"])
	exprJSON, _ := json.Marshal(rawExpr)
	return evalParams(fmt.Sprintf(
		`(() => { return globalThis.ModCDP.addMiddleware({ name: %s, phase: %s, expression: %s, handler: async (payload, next, context = {}) => { const cdp = globalThis.ModCDP.attachToSession(context.cdpSessionId ?? null); const ModCDP = globalThis.ModCDP; const chrome = globalThis.chrome; const middleware = (%s); return await middleware(payload, next, context); }, }); })()`,
		string(nameJSON), string(phaseJSON), string(exprJSON), rawExpr,
	))
}

func wrapCustomCommand(method string, params map[string]any, sessionID string) map[string]any {
	m, _ := json.Marshal(method)
	p, _ := json.Marshal(params)
	sid, _ := json.Marshal(sessionID)
	return evalParams(fmt.Sprintf(`globalThis.ModCDP.handleCommand(%s, %s, %s)`, string(m), string(p), string(sid)))
}

func wrapServiceWorkerCommand(method string, params map[string]any, sessionID string, targetSessionID string) []rawStep {
	if params == nil {
		params = map[string]any{}
	}
	if targetSessionID == "" {
		targetSessionID = sessionID
	}
	if method == "Mod.ping" {
		if _, ok := params["sentAt"]; !ok {
			next := map[string]any{}
			for key, value := range params {
				next[key] = value
			}
			next["sentAt"] = time.Now().UnixMilli()
			params = next
		}
	}

	if method == "Mod.addCustomEvent" {
		return []rawStep{
			{Method: "Runtime.evaluate", Params: wrapModCDPAddCustomEvent(params), Unwrap: "evaluate"},
		}
	}
	runtimeParams := map[string]any{}
	switch method {
	case "Mod.evaluate":
		runtimeParams = wrapModCDPEvaluate(params, targetSessionID)
	case "Mod.addCustomCommand":
		runtimeParams = wrapModCDPAddCustomCommand(params)
	case "Mod.addMiddleware":
		runtimeParams = wrapModCDPAddMiddleware(params)
	default:
		cdpSessionID, _ := params["cdpSessionId"].(string)
		if cdpSessionID == "" {
			cdpSessionID = targetSessionID
		}
		runtimeParams = wrapCustomCommand(method, params, cdpSessionID)
	}
	return []rawStep{{Method: "Runtime.evaluate", Params: runtimeParams, Unwrap: "evaluate"}}
}

func wrapCommandIfNeeded(method string, params map[string]any, routes map[string]string, sessionID string, targetSessionID ...string) (rawCommand, error) {
	targetSession := ""
	if len(targetSessionID) > 0 {
		targetSession = targetSessionID[0]
	}
	route := routeFor(method, routes)
	if route == "direct_cdp" {
		return rawCommand{Route: route, Target: "direct_cdp", Steps: []rawStep{{Method: method, Params: params, SessionID: targetSession}}}, nil
	}
	if route == "service_worker" {
		return rawCommand{Route: route, Target: "service_worker", Steps: wrapServiceWorkerCommand(method, params, sessionID, targetSession)}, nil
	}
	return rawCommand{}, fmt.Errorf("unsupported client route %q for %s", route, method)
}

func unwrapResponseIfNeeded(result map[string]any, unwrap string) (any, error) {
	if unwrap != "evaluate" {
		return result, nil
	}
	if ex, ok := result["exceptionDetails"].(map[string]any); ok {
		msg := ""
		if e, ok := ex["exception"].(map[string]any); ok {
			if d, ok := e["description"].(string); ok {
				msg = d
			}
		}
		if msg == "" {
			if t, ok := ex["text"].(string); ok {
				msg = t
			}
		}
		if msg == "" {
			msg = "Runtime.evaluate failed"
		}
		return nil, fmt.Errorf("%s", msg)
	}
	inner, _ := result["result"].(map[string]any)
	return inner["value"], nil
}

func unwrapEventIfNeeded(method string, params map[string]any, sessionID string, ourSessionID string) (string, any, bool) {
	if method != "Runtime.bindingCalled" {
		return "", nil, false
	}
	name, _ := params["name"].(string)
	payloadStr, _ := params["payload"].(string)
	var payload map[string]any
	if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil || payload == nil {
		return "", nil, false
	}
	isUpstreamEventBinding := name == upstreamEventBindingName
	isCustomEventBinding := name == customEventBindingName
	if !isUpstreamEventBinding && !isCustomEventBinding {
		return "", nil, false
	}
	payloadEvent, _ := payload["event"].(string)
	if payloadEvent == "" {
		return "", nil, false
	}
	resolvedEvent := payloadEvent
	if resolvedEvent == upstreamEventBindingName || resolvedEvent == customEventBindingName {
		return "", nil, false
	}
	if data, ok := payload["data"]; ok {
		return resolvedEvent, data, true
	}
	return resolvedEvent, payload, true
}
