"""Pure ModCDP <-> CDP translation helpers for the Python client."""

import json
import time
from typing import cast

from ..types.modcdp import (
    ModCDPRoutes,
    JsonObject,
    JsonValue,
    ProtocolParams,
    ProtocolPayload,
    ProtocolResult,
    RuntimeCallFunctionOnParams,
    TranslatedCommand,
    TranslatedStep,
    UnwrappedModCDPEvent,
)

UPSTREAM_EVENT_BINDING_NAME = "__ModCDP_event_from_upstream__"
CUSTOM_EVENT_BINDING_NAME = "__ModCDP_custom_event__"

DEFAULT_CLIENT_ROUTES: ModCDPRoutes = {
    "Mod.*": "service_worker",
    "Custom.*": "service_worker",
    "*.*": "service_worker",
}


def route_for(method: str, routes: ModCDPRoutes) -> str:
    routes = routes or {}
    if method in routes:
        return routes[method]
    best_prefix_len = -1
    best_route = None
    for pattern, route in routes.items():
        if pattern == "*.*" or not pattern.endswith(".*"):
            continue
        prefix = pattern[:-1]
        if method.startswith(prefix) and len(prefix) > best_prefix_len:
            best_prefix_len = len(prefix)
            best_route = route
    if best_route is not None:
        return best_route
    if "*.*" in routes:
        return routes["*.*"]
    return "direct_cdp"


def _required_string(params: ProtocolParams, name: str) -> str:
    value = params.get(name)
    if not isinstance(value, str) or not value:
        raise TypeError(f"{name} must be a non-empty string")
    return value


def _optional_string(params: ProtocolParams, name: str) -> str | None:
    value = params.get(name)
    if value is None:
        return None
    if not isinstance(value, str):
        raise TypeError(f"{name} must be a string")
    return value


def _object_or_empty(value: JsonValue | None) -> JsonObject:
    return value if isinstance(value, dict) else {}


def _call_function_params(function_declaration: str) -> RuntimeCallFunctionOnParams:
    return {
        "functionDeclaration": function_declaration,
        "awaitPromise": True,
        "returnByValue": True,
    }


def _wrap_modcdp_evaluate(
    params: ProtocolParams,
    session_id: str,
    target_session_id: str | None = None,
) -> RuntimeCallFunctionOnParams:
    expression = _required_string(params, "expression")
    user_params = params.get("params", {})
    cdp_session_id = target_session_id or _optional_string(params, "cdpSessionId") or session_id
    return _call_function_params(
        "async function() {\n"
        f"  const params = {json.dumps(user_params)};\n"
        f"  const cdp = globalThis.ModCDP.attachToSession({json.dumps(cdp_session_id)});\n"
        "  const ModCDP = globalThis.ModCDP;\n"
        "  const chrome = globalThis.chrome;\n"
        f"  const value = ({expression});\n"
        "  return typeof value === 'function' ? await value(params) : value;\n"
        "}"
    )


def _wrap_modcdp_add_custom_command(params: ProtocolParams) -> RuntimeCallFunctionOnParams:
    name = _required_string(params, "name")
    expression = _required_string(params, "expression")
    return _call_function_params(
        "function() {\n"
        "  return globalThis.ModCDP.addCustomCommand({\n"
        f"    name: {json.dumps(name)},\n"
        "    params_schema: null,\n"
        "    result_schema: null,\n"
        f"    expression: {json.dumps(expression)},\n"
        "    handler: async (params, cdpSessionId, method) => {\n"
        "      const cdp = globalThis.ModCDP.attachToSession(cdpSessionId);\n"
        "      const ModCDP = globalThis.ModCDP;\n"
        "      const chrome = globalThis.chrome;\n"
        f"      const handler = ({expression});\n"
        "      return await handler(params || {}, method);\n"
        "    },\n"
        "  });\n"
        "}"
    )


def _wrap_modcdp_add_custom_event(params: ProtocolParams) -> RuntimeCallFunctionOnParams:
    name = _required_string(params, "name")
    return _call_function_params(
        "function() {\n"
        "  return globalThis.ModCDP.addCustomEvent({\n"
        f"  name: {json.dumps(name)},\n"
        "  event_schema: null,\n"
        "  });\n"
        "}"
    )


def _wrap_modcdp_add_middleware(params: ProtocolParams) -> RuntimeCallFunctionOnParams:
    phase = _required_string(params, "phase")
    expression = _required_string(params, "expression")
    name = _optional_string(params, "name") or "*"
    return _call_function_params(
        "function() {\n"
        "  return globalThis.ModCDP.addMiddleware({\n"
        f"    name: {json.dumps(name)},\n"
        f"    phase: {json.dumps(phase)},\n"
        f"    expression: {json.dumps(expression)},\n"
        "    handler: async (payload, next, context = {}) => {\n"
        "      const cdp = globalThis.ModCDP.attachToSession(context.cdpSessionId ?? null);\n"
        "      const ModCDP = globalThis.ModCDP;\n"
        "      const chrome = globalThis.chrome;\n"
        f"      const middleware = ({expression});\n"
        "      return await middleware(payload, next, context);\n"
        "    },\n"
        "  });\n"
        "}"
    )


def _wrap_custom_command(method: str, params: ProtocolParams, session_id: str) -> RuntimeCallFunctionOnParams:
    return _call_function_params(
        "async function() { return JSON.stringify(await globalThis.ModCDP.handleCommand("
        f"{json.dumps(method)}, {json.dumps(params)}, {json.dumps(session_id)})); }}"
    )


def _wrap_service_worker_command(
    method: str,
    params: ProtocolParams,
    session_id: str,
    target_session_id: str | None = None,
) -> list[TranslatedStep]:
    if method == "Mod.ping" and "sent_at" not in params:
        params = {**params, "sent_at": int(time.time() * 1000)}

    if method == "Mod.addCustomEvent":
        return [
            {
                "method": "Runtime.callFunctionOn",
                "params": _wrap_modcdp_add_custom_event(params),
                "unwrap": "runtime",
            },
        ]
    if method == "Mod.evaluate":
        runtime_params = _wrap_modcdp_evaluate(params, session_id, target_session_id)
    elif method == "Mod.addCustomCommand":
        runtime_params = _wrap_modcdp_add_custom_command(params)
    elif method == "Mod.addMiddleware":
        runtime_params = _wrap_modcdp_add_middleware(params)
    else:
        runtime_params = _wrap_custom_command(method, params, target_session_id or _optional_string(params, "cdpSessionId") or session_id)
    unwrap = "runtime" if method.startswith("Mod.") else "runtime_json"
    return [{"method": "Runtime.callFunctionOn", "params": runtime_params, "unwrap": unwrap}]


def wrap_command_if_needed(
    method: str,
    params: ProtocolParams | None = None,
    *,
    routes: ModCDPRoutes | None = None,
    cdp_session_id: str | None = None,
    target_cdp_session_id: str | None = None,
) -> TranslatedCommand:
    params = params or {}
    route = route_for(method, routes or DEFAULT_CLIENT_ROUTES)
    if route == "direct_cdp":
        step: TranslatedStep = {"method": method, "params": params}
        if target_cdp_session_id:
            step["sessionId"] = target_cdp_session_id
        return {"route": route, "target": "direct_cdp", "steps": [step]}
    if route == "service_worker":
        if cdp_session_id is None:
            raise RuntimeError(f"service_worker route requires a CDP session id for {method}")
        return {
            "route": route,
            "target": "service_worker",
            "steps": _wrap_service_worker_command(method, params, cdp_session_id, target_cdp_session_id),
        }
    raise RuntimeError(f"Unsupported client route '{route}' for {method}")


def _unwrap_evaluate_response(result: ProtocolResult) -> JsonValue:
    if result.get("exceptionDetails"):
        ex = _object_or_empty(result.get("exceptionDetails"))
        exception = _object_or_empty(ex.get("exception"))
        description = exception.get("description")
        text = ex.get("text")
        msg = (
            description
            if isinstance(description, str)
            else text
            if isinstance(text, str)
            else "Runtime.evaluate failed"
        )
        raise RuntimeError(msg)
    inner = _object_or_empty(result.get("result"))
    return inner.get("value")


def unwrap_response_if_needed(result: ProtocolResult, unwrap: str | None = None) -> JsonValue:
    if unwrap == "runtime_json":
        value = _unwrap_evaluate_response(result)
        return cast(JsonValue, json.loads(value)) if isinstance(value, str) else value
    return _unwrap_evaluate_response(result) if unwrap == "runtime" else (result or {})


def unwrap_event_if_needed(
    method: str,
    params: ProtocolParams,
    session_id: str | None = None,
    our_session_id: str | None = None,
) -> UnwrappedModCDPEvent | None:
    if method != "Runtime.bindingCalled":
        return None
    binding_name = params.get("name")
    if not isinstance(binding_name, str):
        return None
    raw_payload = params.get("payload")
    if not isinstance(raw_payload, str):
        return None
    try:
        parsed: object = json.loads(raw_payload)
    except json.JSONDecodeError:
        return None
    if not isinstance(parsed, dict):
        return None
    payload = cast(ProtocolPayload, parsed)
    is_upstream_event_binding = binding_name == UPSTREAM_EVENT_BINDING_NAME
    is_custom_event_binding = binding_name == CUSTOM_EVENT_BINDING_NAME
    if not is_upstream_event_binding and not is_custom_event_binding:
        return None
    payload_event = payload.get("event")
    if not isinstance(payload_event, str) or not payload_event:
        payload_event = None
    if payload_event is None:
        return None
    resolved_event = payload_event
    if resolved_event == UPSTREAM_EVENT_BINDING_NAME or resolved_event == CUSTOM_EVENT_BINDING_NAME:
        return None
    data_value = payload["data"] if "data" in payload else payload
    data: ProtocolPayload = data_value if isinstance(data_value, dict) else {"value": data_value}
    raw_source_session_id = payload.get("cdpSessionId")
    source_session_id = raw_source_session_id if isinstance(raw_source_session_id, str) else session_id
    unwrapped: UnwrappedModCDPEvent = {"event": resolved_event, "data": data, "sessionId": source_session_id}
    return unwrapped
