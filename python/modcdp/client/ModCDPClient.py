"""ModCDPClient (Python): importable, no CLI, no demo code.

Constructor option groups mirror the JS / Go ports:
    launcher          browser/session creation and cleanup
    upstream          message transport to raw CDP or a ModCDP server
    injector          raw-CDP extension discovery/injection/borrowing
    server            ModCDPServer.configure params
    client            client routing and client-owned send/event timings

Public methods: connect(), send(method, params), on(event, handler), close(), _cdp.send(), _cdp.on().
Synchronous (blocking) API; upstream transports own their read loops.
"""

import asyncio
import inspect
import threading
import time
from collections.abc import Mapping, Sequence
from queue import Queue, Empty
from typing import Any, Literal, cast

from pydantic import BaseModel, TypeAdapter, ValidationError
from pydantic_core import to_jsonable_python
from ..router.AutoSessionRouter import AutoSessionRouter
from ..types.jsonschema import type_adapter_from_json_schema
from ..types.generated import cdp as generated_cdp
from ..types.generated.cdp import AwaitableDict, CDPEvent, CDPModel, CDPParams, CDPSurfaceMixin, cdp_event_name, install_cdp_surface
from ..launcher.BrowserbaseBrowserLauncher import BrowserbaseBrowserLauncher
from ..injector.BBBrowserExtensionInjector import BBBrowserExtensionInjector
from ..injector.BorrowedExtensionInjector import BorrowedExtensionInjector
from ..injector.DiscoveredExtensionInjector import DiscoveredExtensionInjector
from ..injector.ExtensionInjector import (
    DEFAULT_MODCDP_SERVICE_WORKER_URL_SUFFIXES,
    ExtensionInjector,
    ExtensionInjectorConfig,
)
from ..injector.ExtensionsLoadUnpackedInjector import ExtensionsLoadUnpackedInjector
from ..injector.LocalBrowserLaunchExtensionInjector import LocalBrowserLaunchExtensionInjector
from ..launcher.LocalBrowserLauncher import LocalBrowserLauncher
from ..launcher.NoopBrowserLauncher import NoopBrowserLauncher
from ..launcher.RemoteBrowserLauncher import RemoteBrowserLauncher
from ..transport.NativeMessagingUpstreamTransport import DEFAULT_UPSTREAM_NATIVEMESSAGING_WAIT_TIMEOUT_MS, NativeMessagingUpstreamTransport
from ..transport.NatsUpstreamTransport import DEFAULT_UPSTREAM_NATS_WAIT_TIMEOUT_MS, NatsUpstreamTransport
from ..transport.PipeUpstreamTransport import PipeUpstreamTransport
from ..transport.ReverseWebSocketUpstreamTransport import (
    DEFAULT_UPSTREAM_REVERSEWS_BIND,
    DEFAULT_UPSTREAM_REVERSEWS_WAIT_TIMEOUT_MS,
    ReverseWebSocketUpstreamTransport,
)
from ..transport.UpstreamTransport import UpstreamTransport
from ..transport.WebSocketUpstreamTransport import WebSocketUpstreamTransport
from ..translate.translate import (
    CUSTOM_EVENT_BINDING_NAME,
    DEFAULT_CLIENT_ROUTES,
    UPSTREAM_EVENT_BINDING_NAME,
    wrap_command_if_needed,
    unwrap_event_if_needed,
    unwrap_response_if_needed,
)
from ..launcher.BrowserLauncher import BrowserLaunchOptions
from ..types.modcdp import (
    ModCDPAddCustomCommandParams,
    ModCDPAddCustomEventObjectParams,
    ModCDPAddCustomEventParams,
    ModCDPAddMiddlewareParams,
    ModCDPCommandTiming,
    ModCDPConnectTiming,
    ModCDPPingLatency,
    ModCDPRawTiming,
    ModCDPRoutes,
    ModCDPServerConfig,
    CdpMessage,
    ExtensionInfo,
    MessageParams,
    Handler,
    JsonValue,
    PendingEntry,
    ProtocolParams,
    ProtocolPayload,
    ProtocolResult,
    TranslatedCommand,
)


def _defaulted(value: Any, fallback: Any) -> Any:
    return fallback if value is None else value


class AwaitableValue:
    def __init__(self, value: Any) -> None:
        self.value = value

    def __await__(self):
        async def _value():
            return self.value

        return _value().__await__()

    def __bool__(self) -> bool:
        return bool(self.value)

    def __getattr__(self, name: str) -> Any:
        return getattr(self.value, name)

    def __getitem__(self, key: Any) -> Any:
        return self.value[key]

    def __eq__(self, other: object) -> bool:
        return self.value == other

    def __repr__(self) -> str:
        return repr(self.value)

    def __str__(self) -> str:
        return str(self.value)


class _ModDomain:
    def __init__(self, client: "ModCDPClient") -> None:
        self._client = client

    def evaluate(
        self,
        *,
        expression: str,
        params: Mapping[str, Any] | None = None,
        cdpSessionId: str | None = None,
    ) -> AwaitableDict | AwaitableValue:
        payload: dict[str, Any] = {"expression": expression}
        if params is not None:
            payload["params"] = dict(params)
        if cdpSessionId is not None:
            payload["cdpSessionId"] = cdpSessionId
        return self._client._send_command("Mod.evaluate", payload)

    def addCustomCommand(
        self,
        name: str,
        *,
        params_schema: Any | None = None,
        result_schema: Any | None = None,
        expression: str | None = None,
    ) -> AwaitableDict | AwaitableValue:
        payload: dict[str, Any] = {"name": name}
        if params_schema is not None:
            payload["params_schema"] = params_schema
        if result_schema is not None:
            payload["result_schema"] = result_schema
        if expression is not None:
            payload["expression"] = expression
        return self._client._send_command("Mod.addCustomCommand", payload)

    def addCustomEvent(self, name: str, *, event_schema: Any | None = None) -> AwaitableDict | AwaitableValue:
        payload: dict[str, Any] = {"name": name}
        if event_schema is not None:
            payload["event_schema"] = event_schema
        return self._client._send_command("Mod.addCustomEvent", payload)

    def addMiddleware(
        self,
        *,
        phase: Literal["request", "response", "event"],
        expression: str,
        name: str | None = None,
    ) -> AwaitableDict | AwaitableValue:
        payload: dict[str, Any] = {"phase": phase, "expression": expression}
        if name is not None:
            payload["name"] = name
        return self._client._send_command("Mod.addMiddleware", payload)

    def configure(self, **params: Any) -> AwaitableDict | AwaitableValue:
        return self._client._send_command("Mod.configure", params)

    def ping(self, **params: Any) -> AwaitableDict | AwaitableValue:
        return self._client._send_command("Mod.ping", params)

MODCDP_READY_EXPRESSION = (
    "Boolean(globalThis.ModCDP?.__ModCDPServerVersion >= 1 && "
    "globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)"
)
DEFAULT_SERVER = object()
DEFAULT_CDP_SEND_TIMEOUT_MS = 10_000
DEFAULT_EVENT_WAIT_TIMEOUT_MS = 10_000
DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS = 10_000
DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS = 10_000
DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS = 60_000
DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS = 100
DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS = 20
DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS = 250


class _RawCDP:
    def __init__(self, client: "ModCDPClient") -> None:
        self._client = client

    def send(
        self,
        method: str,
        params: ProtocolParams | None = None,
        session_id: str | None = None,
    ) -> ProtocolResult:
        return self._client._send_message(method, params or {}, session_id=session_id, record_raw_timing=True)

    def on(self, event: str, handler: Handler) -> "ModCDPClient":
        return self._client.on(event, handler)


def _json_object(value: JsonValue | None) -> ProtocolResult:
    return value if isinstance(value, dict) else {}


class ModCDPClient(CDPSurfaceMixin):
    def __init__(
        self,
        launcher: Mapping[str, Any] | None = None,
        upstream: Mapping[str, Any] | None = None,
        injector: Mapping[str, Any] | None = None,
        client: Mapping[str, Any] | None = None,
        server: Mapping[str, JsonValue] | None | object = DEFAULT_SERVER,
        custom_commands: Sequence[ModCDPAddCustomCommandParams] | None = None,
        custom_events: Sequence[ModCDPAddCustomEventParams] | None = None,
        custom_middlewares: Sequence[ModCDPAddMiddlewareParams] | None = None,
    ) -> None:
        launcher_input = dict(launcher or {})
        upstream_input = dict(upstream or {})
        injector_input = dict(injector or {})
        client_input = dict(client or {})
        upstream_mode = str(upstream_input.get("upstream_mode") or "ws")
        self.upstream: dict[str, Any] = {
            "upstream_mode": upstream_mode,
            "upstream_cdp_url": upstream_input.get("upstream_cdp_url"),
            "upstream_nats_url": upstream_input.get("upstream_nats_url"),
            "upstream_nats_subject_prefix": upstream_input.get("upstream_nats_subject_prefix"),
            "upstream_nats_wait_timeout_ms": int(
                _defaulted(upstream_input.get("upstream_nats_wait_timeout_ms"), DEFAULT_UPSTREAM_NATS_WAIT_TIMEOUT_MS)
            ),
            "upstream_reversews_bind": _defaulted(upstream_input.get("upstream_reversews_bind"), DEFAULT_UPSTREAM_REVERSEWS_BIND),
            "upstream_reversews_wait_timeout_ms": int(
                _defaulted(upstream_input.get("upstream_reversews_wait_timeout_ms"), DEFAULT_UPSTREAM_REVERSEWS_WAIT_TIMEOUT_MS)
            ),
            "upstream_nativemessaging_manifest": upstream_input.get("upstream_nativemessaging_manifest"),
            "upstream_nativemessaging_manifests": list(
                cast(Sequence[str], upstream_input.get("upstream_nativemessaging_manifests") or [])
            )
            if upstream_input.get("upstream_nativemessaging_manifests") is not None
            else None,
            "upstream_nativemessaging_host_name": upstream_input.get("upstream_nativemessaging_host_name"),
            "upstream_nativemessaging_wait_timeout_ms": int(
                _defaulted(
                    upstream_input.get("upstream_nativemessaging_wait_timeout_ms"),
                    DEFAULT_UPSTREAM_NATIVEMESSAGING_WAIT_TIMEOUT_MS,
                )
            ),
            "upstream_ws_connect_error_settle_timeout_ms": int(
                _defaulted(
                    upstream_input.get("upstream_ws_connect_error_settle_timeout_ms"),
                    DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS,
                )
            ),
        }
        self.upstream_endpoint_kind = "raw_cdp" if self.upstream["upstream_mode"] in ("ws", "pipe") else "modcdp_server"
        launcher_mode = launcher_input.get("launcher_mode") or (
            "none" if self.upstream_endpoint_kind == "modcdp_server" else "remote" if self.upstream.get("upstream_cdp_url") else "local"
        )
        self.launcher: dict[str, Any] = {
            "launcher_mode": launcher_mode,
            "launcher_executable_path": launcher_input.get("launcher_executable_path"),
            "launcher_user_data_dir": launcher_input.get("launcher_user_data_dir"),
            "launcher_options": dict(cast(Mapping[str, Any], launcher_input.get("launcher_options") or {})),
        }
        injector_mode = injector_input.get("injector_mode") or (
            "auto" if self.upstream_endpoint_kind == "raw_cdp" or launcher_mode != "none" else "none"
        )
        raw_service_worker_url_suffixes = injector_input.get("injector_service_worker_url_suffixes")
        self.injector: dict[str, Any] = {
            "injector_mode": injector_mode,
            "injector_extension_path": injector_input.get("injector_extension_path"),
            "injector_extension_id": injector_input.get("injector_extension_id"),
            "injector_service_worker_url_includes": list(cast(Sequence[str], injector_input.get("injector_service_worker_url_includes") or [])),
            "injector_service_worker_url_suffixes": list(
                cast(
                    Sequence[str],
                    DEFAULT_MODCDP_SERVICE_WORKER_URL_SUFFIXES
                    if raw_service_worker_url_suffixes is None
                    else raw_service_worker_url_suffixes,
                )
            ),
            "injector_trust_service_worker_target": bool(injector_input.get("injector_trust_service_worker_target", False)),
            "injector_require_service_worker_target": bool(injector_input.get("injector_require_service_worker_target", False)),
            "injector_service_worker_ready_expression": injector_input.get("injector_service_worker_ready_expression"),
            "injector_execution_context_timeout_ms": int(
                _defaulted(injector_input.get("injector_execution_context_timeout_ms"), DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS)
            ),
            "injector_service_worker_probe_timeout_ms": int(
                _defaulted(injector_input.get("injector_service_worker_probe_timeout_ms"), DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS)
            ),
            "injector_service_worker_ready_timeout_ms": int(
                _defaulted(injector_input.get("injector_service_worker_ready_timeout_ms"), DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS)
            ),
            "injector_service_worker_poll_interval_ms": int(
                _defaulted(injector_input.get("injector_service_worker_poll_interval_ms"), DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS)
            ),
            "injector_target_session_poll_interval_ms": int(
                _defaulted(injector_input.get("injector_target_session_poll_interval_ms"), DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS)
            ),
        }
        self.client: dict[str, Any] = {
            "client_routes": {
                **DEFAULT_CLIENT_ROUTES,
                **dict(cast(Mapping[str, str], client_input.get("client_routes") or {})),
            },
            "client_hydrate_aliases": bool(client_input.get("client_hydrate_aliases", True)),
            "client_mirror_upstream_events": bool(client_input.get("client_mirror_upstream_events", True)),
            "client_cdp_send_timeout_ms": int(_defaulted(client_input.get("client_cdp_send_timeout_ms"), DEFAULT_CDP_SEND_TIMEOUT_MS)),
            "client_event_wait_timeout_ms": int(_defaulted(client_input.get("client_event_wait_timeout_ms"), DEFAULT_EVENT_WAIT_TIMEOUT_MS)),
        }
        self.cdp_url: str | None = cast(str | None, self.upstream.get("upstream_cdp_url"))
        if server is DEFAULT_SERVER:
            self.server: ModCDPServerConfig | None = {"server_routes": {"*.*": "chrome_debugger"}} if self.upstream_endpoint_kind == "modcdp_server" else {}
        elif server is None:
            self.server = None
        elif isinstance(server, Mapping):
            self.server = cast(ModCDPServerConfig, {
                **({"server_routes": {"*.*": "chrome_debugger"}} if self.upstream_endpoint_kind == "modcdp_server" else {}),
                **dict(server),
            })
        else:
            raise TypeError("server must be a mapping, None, or omitted")
        self.custom_commands: list[ModCDPAddCustomCommandParams] = list(custom_commands or [])
        self.custom_events: list[ModCDPAddCustomEventParams] = list(custom_events or [])
        self.custom_middlewares: list[ModCDPAddMiddlewareParams] = list(custom_middlewares or [])

        self.extension_id: str | None = None
        self.ext_target_id: str | None = None
        self.ext_session_id: str | None = None
        self.ext_execution_context_id: int | None = None
        self.latency: ModCDPPingLatency | None = None
        self.connect_timing: ModCDPConnectTiming | None = None
        self.last_command_timing: ModCDPCommandTiming | None = None
        self.last_raw_timing: ModCDPRawTiming | None = None

        self.transport: UpstreamTransport | None = None
        self._next_id = 0
        self._pending: dict[int, PendingEntry] = {}
        self._handlers: dict[str, list[Handler]] = {}
        self._handler_wrappers: dict[tuple[str, Handler], Handler] = {}
        self._lock = threading.Lock()
        self.auto_sessions = AutoSessionRouter(
            lambda method, params=None, session_id=None: self._send_message(method, params or {}, session_id),
            lambda: self.injector["injector_execution_context_timeout_ms"],
        )
        self._schema_lock = threading.RLock()
        self._command_params_schemas: dict[str, TypeAdapter[Any]] = {}
        self._command_result_schemas: dict[str, TypeAdapter[Any]] = {}
        self._event_schemas: dict[str, TypeAdapter[Any]] = {}
        self._command_result_model_schemas: set[str] = set()
        self._event_model_schemas: set[str] = set()
        self._event_classes: dict[str, type[CDPEvent]] = {}
        if self.client["client_hydrate_aliases"]:
            install_cdp_surface(self)
        self.Mod = _ModDomain(self)
        self._closed = False
        self._launched_browser: Any | None = None
        self._extension_injectors: list[ExtensionInjector] = []
        self._cdp = _RawCDP(self)
        self._hydrate_native_protocol_schemas()
        self._hydrate_custom_surface()

    def connect(self) -> "ModCDPClient":
        connect_started_at = int(time.time() * 1000)
        transport_started_at = int(time.time() * 1000)
        self._connect_upstream_transport()
        transport_connected_at = int(time.time() * 1000)
        if self.transport is None:
            raise RuntimeError("upstream transport did not connect.")
        self.transport.onRecv(lambda message: self._on_recv(cast(CdpMessage, message)))
        self.transport.onClose(lambda error: self._reject_all(error))

        if self.upstream_endpoint_kind == "modcdp_server":
            self.transport.waitForPeer()
            if self.server is not None:
                self._send_message("Mod.configure", cast(ProtocolParams, self._server_configure_params()))
            threading.Thread(target=self._measure_ping_latency, daemon=True).start()
            connected_at = int(time.time() * 1000)
            self.connect_timing = cast(ModCDPConnectTiming, {
                "started_at": connect_started_at,
                "upstream_mode": self.upstream.get("upstream_mode"),
                "upstream_endpoint_kind": self.upstream_endpoint_kind,
                "transport_started_at": transport_started_at,
                "transport_connected_at": transport_connected_at,
                "transport_duration_ms": transport_connected_at - transport_started_at,
                "connected_at": connected_at,
                "duration_ms": connected_at - connect_started_at,
            })
            return self

        self._initialize_raw_cdp_transport()

        injector_started_at = int(time.time() * 1000)
        if self.injector.get("injector_mode") == "none":
            raise RuntimeError("injector.injector_mode='none' cannot be used with a raw_cdp upstream.")
        ext = self._inject_extension(self._extension_injectors)
        injector_completed_at = int(time.time() * 1000)
        self.extension_id = ext["extension_id"]
        self.ext_target_id = ext["target_id"]
        self.ext_session_id = ext["session_id"]
        self._send_message("Runtime.enable", {}, self.ext_session_id)
        self.ext_execution_context_id = self.auto_sessions.waitForExecutionContext(
            self.ext_session_id,
            self.injector["injector_execution_context_timeout_ms"],
        )
        self._send_message("Runtime.addBinding", {"name": CUSTOM_EVENT_BINDING_NAME}, self.ext_session_id)
        if self.client["client_mirror_upstream_events"]:
            self._send_message("Runtime.addBinding", {"name": UPSTREAM_EVENT_BINDING_NAME}, self.ext_session_id)

        if self.server is not None:
            self._send_raw(wrap_command_if_needed(
                "Mod.configure",
                cast(ProtocolParams, self._server_configure_params()),
                routes=cast(ModCDPRoutes, self.client["client_routes"]),
                cdp_session_id=self.ext_session_id,
            ))
        threading.Thread(target=self._measure_ping_latency, daemon=True).start()
        connected_at = int(time.time() * 1000)
        self.connect_timing = cast(ModCDPConnectTiming, {
            "started_at": connect_started_at,
            "upstream_mode": self.upstream.get("upstream_mode"),
            "upstream_endpoint_kind": self.upstream_endpoint_kind,
            "transport_started_at": transport_started_at,
            "transport_connected_at": transport_connected_at,
            "transport_duration_ms": transport_connected_at - transport_started_at,
            "injector_source": ext.get("source"),
            "injector_started_at": injector_started_at,
            "injector_completed_at": injector_completed_at,
            "injector_duration_ms": injector_completed_at - injector_started_at,
            "connected_at": connected_at,
            "duration_ms": connected_at - connect_started_at,
        })
        return self

    def _send_command(
        self,
        method: str,
        params: Mapping[str, Any] | None = None,
        session_id: str | None = None,
        validate_custom_schema: bool = True,
    ) -> Any:
        started_at = int(time.time() * 1000)
        command_params = cast(ProtocolParams, dict(params or {}))
        if method == "Mod.addCustomCommand":
            self._register_custom_command(command_params)
            expression = command_params.get("expression")
            if not isinstance(expression, str) or not expression:
                completed_at = int(time.time() * 1000)
                self.last_command_timing = {
                    "method": method,
                    "target": "client",
                    "started_at": started_at,
                    "completed_at": completed_at,
                    "duration_ms": completed_at - started_at,
                }
                return AwaitableDict({"name": cast(str, command_params.get("name")), "registered": True})
            command_params = self._custom_command_wire_params(command_params)
        elif method == "Mod.addCustomEvent":
            self._register_custom_event(command_params)
            if self.ext_session_id is None and self.upstream_endpoint_kind != "modcdp_server":
                completed_at = int(time.time() * 1000)
                self.last_command_timing = {
                    "method": method,
                    "target": "client",
                    "started_at": started_at,
                    "completed_at": completed_at,
                    "duration_ms": completed_at - started_at,
                }
                return AwaitableDict({"name": cast(str, command_params.get("name")), "registered": True})
            command_params = self._custom_event_wire_params(command_params)
        should_validate_params = validate_custom_schema or method in self._command_params_schemas
        should_validate_result = validate_custom_schema or method in self._command_result_schemas
        if method not in {"Mod.addCustomCommand", "Mod.addCustomEvent"} and should_validate_params:
            command_params = self._validate_command_params(method, command_params)

        if self.upstream_endpoint_kind == "modcdp_server":
            result = self._send_message(method, command_params)
            if should_validate_result and method != "Mod.addCustomCommand":
                result = self._validate_command_result(method, result)
            completed_at = int(time.time() * 1000)
            self.last_command_timing = {
                "method": method,
                "target": "modcdp_server",
                "started_at": started_at,
                "completed_at": completed_at,
                "duration_ms": completed_at - started_at,
            }
            return AwaitableDict(result) if isinstance(result, dict) else AwaitableValue(result)

        command = wrap_command_if_needed(
            method,
            command_params,
            routes=cast(ModCDPRoutes, self.client["client_routes"]),
            cdp_session_id=self.ext_session_id,
            target_cdp_session_id=session_id,
        )
        result = self._send_raw(command)
        if should_validate_result and method != "Mod.addCustomCommand":
            result = self._validate_command_result(method, result)
        completed_at = int(time.time() * 1000)
        self.last_command_timing = {
            "method": method,
            "target": command["target"],
            "started_at": started_at,
            "completed_at": completed_at,
            "duration_ms": completed_at - started_at,
        }
        return AwaitableDict(result) if isinstance(result, dict) else AwaitableValue(result)

    def sendRaw(
        self,
        method: str,
        params: Mapping[str, Any] | None = None,
        session_id: str | None = None,
    ) -> ProtocolResult:
        result = self._send_message(method, cast(ProtocolParams, dict(params or {})), session_id, record_raw_timing=True)
        if not isinstance(result, dict):
            raise RuntimeError(f"{method} returned non-object value: {result!r}")
        return result

    def send(
        self,
        method: str,
        params: Mapping[str, Any] | None = None,
        session_id: str | None = None,
    ) -> AwaitableDict | AwaitableValue:
        result = self._send_command(method, params, session_id=session_id, validate_custom_schema=False)
        if isinstance(result, AwaitableDict):
            return result
        if isinstance(result, AwaitableValue):
            return result
        return AwaitableDict(result) if isinstance(result, dict) else AwaitableValue(result)

    def on(self, event: str | type[CDPEvent], handler: Handler) -> "ModCDPClient":
        event_name = cdp_event_name(event) if not isinstance(event, str) else event
        if event_name is None:
            raise TypeError("event must be a CDP event class or event name string")
        wrapped_handler = handler
        if not isinstance(event, str):
            event_class = event

            def typed_handler(payload):
                typed_payload = event_class.model_validate(payload) if isinstance(payload, Mapping) else payload
                return handler(typed_payload)

            wrapped_handler = typed_handler
        self._handler_wrappers[(event_name, handler)] = wrapped_handler
        handlers = self._handlers.setdefault(event_name, [])
        if wrapped_handler not in handlers:
            handlers.append(wrapped_handler)
        return self

    def once(self, event: str | type[CDPEvent], handler: Handler) -> "ModCDPClient":
        def wrapped_handler(payload: Any) -> Any:
            self.off(event, wrapped_handler)
            return handler(payload)

        return self.on(event, wrapped_handler)

    def off(self, event: str | type[CDPEvent], handler: Handler) -> "ModCDPClient":
        event_name = cdp_event_name(event) if not isinstance(event, str) else event
        if event_name is None:
            raise TypeError("event must be a CDP event class or event name string")
        wrapped_handler = self._handler_wrappers.pop((event_name, handler), handler)
        handlers = self._handlers.get(event_name)
        if handlers and wrapped_handler in handlers:
            handlers.remove(wrapped_handler)
            if not handlers:
                self._handlers.pop(event_name, None)
        return self

    def __await__(self):
        async def _value():
            return self

        return _value().__await__()

    def _run_handler(self, handler: Handler, payload: Any, event_name: str) -> None:
        try:
            result = handler(payload)
            if inspect.iscoroutine(result):
                asyncio.run(result)
        except Exception as e:
            print(f"[ModCDPClient] handler error for {event_name}: {e}")

    def __getattr__(self, domain: str):
        if domain.startswith("_"):
            raise AttributeError(domain)
        if not self.client["client_hydrate_aliases"]:
            raise AttributeError(domain)
        from ..types.generated.cdp import DynamicDomain

        dynamic = DynamicDomain(self, domain)
        setattr(self, domain, dynamic)
        return dynamic

    def _server_configure_params(self) -> ModCDPServerConfig:
        server = dict(self.server or {})
        server_routes = server.pop("server_routes", None)
        server_loopback_cdp_url = server.pop("server_loopback_cdp_url", None)
        server_browser_token = server.pop("server_browser_token", None)
        server_cdp_send_timeout_ms = server.pop(
            "server_cdp_send_timeout_ms",
            self.client["client_cdp_send_timeout_ms"],
        )
        server_loopback_execution_context_timeout_ms = server.pop(
            "server_loopback_execution_context_timeout_ms",
            self.injector["injector_execution_context_timeout_ms"],
        )
        server_ws_connect_error_settle_timeout_ms = server.pop(
            "server_ws_connect_error_settle_timeout_ms",
            self.upstream["upstream_ws_connect_error_settle_timeout_ms"],
        )
        custom_events: list[ModCDPAddCustomEventObjectParams] = []
        for event in self.custom_events:
            custom_events.append(
                {"name": event}
                if isinstance(event, str)
                else cast(ModCDPAddCustomEventObjectParams, self._custom_event_wire_params(cast(ProtocolParams, event)))
            )
        custom_commands: list[ModCDPAddCustomCommandParams] = [
            cast(ModCDPAddCustomCommandParams, self._custom_command_wire_params(cast(ProtocolParams, command)))
            for command in self.custom_commands
            if isinstance(command.get("expression"), str) and command.get("expression")
        ]
        custom_middlewares: list[ModCDPAddMiddlewareParams] = list(self.custom_middlewares)
        return cast(ModCDPServerConfig, {
            "upstream": {
                "upstream_mode": self.upstream.get("upstream_mode"),
                **({"upstream_nats_url": self.upstream.get("upstream_nats_url")} if self.upstream.get("upstream_nats_url") else {}),
                **(
                    {"upstream_nats_subject_prefix": self.upstream.get("upstream_nats_subject_prefix")}
                    if self.upstream.get("upstream_nats_subject_prefix")
                    else {}
                ),
            },
            "client": {"client_routes": self.client["client_routes"]},
            "server": {
                "server_cdp_send_timeout_ms": server_cdp_send_timeout_ms,
                "server_loopback_execution_context_timeout_ms": server_loopback_execution_context_timeout_ms,
                "server_ws_connect_error_settle_timeout_ms": server_ws_connect_error_settle_timeout_ms,
                **({"server_routes": server_routes} if server_routes is not None else {}),
                **({"server_loopback_cdp_url": server_loopback_cdp_url} if server_loopback_cdp_url is not None else {}),
                **({"server_browser_token": server_browser_token} if server_browser_token is not None else {}),
                **server,
            },
            "custom_events": custom_events,
            "custom_commands": custom_commands,
            "custom_middlewares": custom_middlewares,
        })

    def close(self) -> None:
        if self._closed:
            return
        self._closed = True
        if self._launched_browser is not None:
            self._launched_browser["close"]()
            self._launched_browser = None
        try:
            if self.transport:
                self.transport.close()
        except Exception:
            pass
        self.transport = None
        for injector in self._extension_injectors:
            try:
                injector.close()
            except Exception:
                pass
        self._extension_injectors = []

    def _session_id_for_target(self, target_id: str, timeout: float = 0) -> str | None:
        if timeout <= 0:
            return self.auto_sessions.sessionIdForTarget(target_id)
        deadline = time.time() + timeout
        while time.time() <= deadline:
            session_id = self.auto_sessions.sessionIdForTarget(target_id)
            if session_id:
                return session_id
            time.sleep(self.injector["injector_target_session_poll_interval_ms"] / 1000)
        return None

    def _ensure_session_id_for_target(self, target_id: str, timeout: float = 0, allow_attach: bool = False) -> str | None:
        session_id = self.auto_sessions.sessionIdForTarget(target_id)
        if session_id:
            return session_id
        if allow_attach:
            attached_session_id = self.auto_sessions.attachToTarget(target_id)
            if attached_session_id:
                return attached_session_id
        return self._session_id_for_target(target_id, timeout=timeout)

    def _browser_launcher(self):
        if self.launcher.get("launcher_mode") == "local":
            return LocalBrowserLauncher(self._launch_options())
        if self.launcher.get("launcher_mode") == "remote":
            return RemoteBrowserLauncher(self._launch_options(), self.cdp_url)
        if self.launcher.get("launcher_mode") == "bb":
            return BrowserbaseBrowserLauncher(self._launch_options())
        if self.launcher.get("launcher_mode") == "none":
            return NoopBrowserLauncher(self._launch_options())
        raise RuntimeError(f"unknown launcher.launcher_mode={self.launcher.get('launcher_mode')}")

    def _launch_options(self) -> BrowserLaunchOptions:
        launch_options = cast(BrowserLaunchOptions, dict(cast(Mapping[str, Any], self.launcher.get("launcher_options") or {})))
        if self.launcher.get("launcher_executable_path"):
            launch_options["executable_path"] = cast(str, self.launcher["launcher_executable_path"])
        if self.launcher.get("launcher_user_data_dir"):
            launch_options["user_data_dir"] = cast(str, self.launcher["launcher_user_data_dir"])
        return launch_options

    def _connect_upstream_transport(self) -> None:
        if self.transport is not None:
            return
        launcher = self._browser_launcher()
        transport = self._upstream_transport()
        injectors = self._extension_injectors_for_config()
        self._extension_injectors = injectors
        initial_transport_config = self._upstream_transport_config()

        transport.update(initial_transport_config)
        launcher.update(self._launch_options())
        for injector in injectors:
            injector.update(self._base_extension_injector_config(None))
        for injector in injectors:
            injector.update(cast(ExtensionInjectorConfig, launcher.getInjectorConfig()))
        for injector in injectors:
            injector.update(cast(ExtensionInjectorConfig, transport.getInjectorConfig()))
        for injector in injectors:
            injector.prepare()
        for injector in injectors:
            launcher.update(injector.getLauncherConfig())
        for injector in injectors:
            transport.update(injector.getTransportConfig())
        launcher.update(cast(BrowserLaunchOptions, transport.getLauncherConfig()))
        launcher.update({"loopback_cdp": self._server_needs_loopback_cdp()})
        transport.update(launcher.getTransportConfig())

        if transport.endpoint_kind == "modcdp_server":
            transport.connect()
        if self.launcher.get("launcher_mode") != "none":
            launched = launcher.launch()
            self._launched_browser = launched
            transport.update(launcher.getTransportConfig())
            for injector in injectors:
                injector.update(cast(ExtensionInjectorConfig, launcher.getInjectorConfig()))
            for injector in injectors:
                transport.update(injector.getTransportConfig())
        launched_cdp_url = cast(str | None, self._launched_browser.get("cdp_url")) if self._launched_browser else None
        if transport.endpoint_kind == "raw_cdp":
            transport.connect()

        self.transport = transport
        self.cdp_url = cast(
            str | None,
            (transport.url or launched_cdp_url) if transport.endpoint_kind == "raw_cdp" else launched_cdp_url,
        )
        if transport.mode == "ws" and transport.url:
            # For ws mode, cdp_url has been resolved to the concrete WebSocket CDP endpoint after connect().
            self.upstream["upstream_cdp_url"] = transport.url
        server_config = (
            {"server_loopback_cdp_url": launched_cdp_url}
            if transport.endpoint_kind == "modcdp_server" and launched_cdp_url and not launched_cdp_url.startswith("pipe://")
            else {}
        )
        transport_server_config = transport.getServerConfig()
        server_config.update(launcher.getServerConfig())
        server_config.update(transport_server_config)
        if self.server is not None and server_config.get("server_loopback_cdp_url"):
            configured_loopback = self.server.get("server_loopback_cdp_url")
            if "server_loopback_cdp_url" not in self.server or configured_loopback in (
                initial_transport_config.get("cdp_url"),
                launched_cdp_url,
            ):
                self.server = cast(ModCDPServerConfig, {**self.server, **server_config})

    def _server_needs_loopback_cdp(self) -> bool:
        if self.server is None or self.server.get("server_loopback_cdp_url"):
            return False
        return "loopback_cdp" in set((self.server.get("server_routes") or {}).values())

    def _upstream_transport_config(self) -> dict[str, Any]:
        return {
            "cdp_url": self.upstream.get("upstream_cdp_url"),
            "upstream_nats_url": self.upstream.get("upstream_nats_url"),
            "upstream_nats_subject_prefix": self.upstream.get("upstream_nats_subject_prefix"),
            "upstream_nats_wait_timeout_ms": self.upstream.get("upstream_nats_wait_timeout_ms"),
            "upstream_reversews_bind": self.upstream.get("upstream_reversews_bind"),
            "upstream_reversews_wait_timeout_ms": self.upstream.get("upstream_reversews_wait_timeout_ms"),
            "upstream_nativemessaging_manifest": self.upstream.get("upstream_nativemessaging_manifest"),
            "upstream_nativemessaging_manifests": self.upstream.get("upstream_nativemessaging_manifests"),
            "upstream_nativemessaging_host_name": self.upstream.get("upstream_nativemessaging_host_name"),
            "upstream_nativemessaging_wait_timeout_ms": self.upstream.get("upstream_nativemessaging_wait_timeout_ms"),
            "injector_extension_id": self.injector.get("injector_extension_id"),
        }

    def _initialize_raw_cdp_transport(self) -> None:
        self._send_message("Target.setAutoAttach", {
            "autoAttach": True,
            "waitForDebuggerOnStart": False,
            "flatten": True,
        })
        self._send_message("Target.setDiscoverTargets", {"discover": True})

    def _upstream_transport(self):
        mode = self.upstream.get("upstream_mode")
        if mode == "ws":
            return WebSocketUpstreamTransport()
        if mode == "pipe":
            return PipeUpstreamTransport()
        if mode == "reversews":
            return ReverseWebSocketUpstreamTransport()
        if mode == "nativemessaging":
            return NativeMessagingUpstreamTransport()
        if mode == "nats":
            return NatsUpstreamTransport()
        raise RuntimeError(f"unknown upstream.upstream_mode={mode}")

    def _extension_injectors_for_config(self) -> list[ExtensionInjector]:
        mode = self.injector.get("injector_mode")
        if mode == "none":
            return []
        injectors: list[ExtensionInjector] = []
        prefer_launch_injection = mode == "auto" and self.launcher.get("launcher_mode") == "local"
        if mode in ("auto", "discover") and not prefer_launch_injection:
            injectors.append(DiscoveredExtensionInjector())
        if mode in ("auto", "inject"):
            if self.launcher.get("launcher_mode") == "bb":
                injectors.append(BBBrowserExtensionInjector())
            if self.launcher.get("launcher_mode") == "local":
                injectors.append(LocalBrowserLaunchExtensionInjector())
            injectors.append(ExtensionsLoadUnpackedInjector())
        if prefer_launch_injection:
            injectors.append(DiscoveredExtensionInjector())
        if mode in ("auto", "borrow"):
            injectors.append(BorrowedExtensionInjector())
        if not injectors:
            raise RuntimeError(f"unknown injector.injector_mode={mode}")
        return injectors

    def _base_extension_injector_config(self, send: Any | None) -> ExtensionInjectorConfig:
        trust_service_worker_target = (
            self.injector["injector_trust_service_worker_target"]
            or len(self.injector["injector_service_worker_url_includes"]) > 0
            or any(
                len([part for part in suffix.split("/") if part]) > 1
                for suffix in self.injector["injector_service_worker_url_suffixes"]
            )
        )

        def send_cdp(method: str, params: ProtocolParams | None = None, session_id: str | None = None) -> ProtocolResult:
            if send is None:
                raise RuntimeError("Extension injector CDP send is not connected.")
            return self._send_message(
                method,
                params or {},
                session_id,
                timeout=self.client["client_cdp_send_timeout_ms"] / 1000,
            )

        def attach_to_target(target_id: str) -> str | None:
            return self._ensure_session_id_for_target(
                target_id,
                timeout=self.injector["injector_service_worker_probe_timeout_ms"] / 1000,
                allow_attach=True,
            )

        return {
            "send": send_cdp if send is not None else None,
            "sessionIdForTarget": self.auto_sessions.sessionIdForTarget,
            "attachToTarget": attach_to_target if send is not None else None,
            "waitForExecutionContext": self.auto_sessions.waitForExecutionContext,
            "injector_extension_path": cast(str | None, self.injector.get("injector_extension_path")),
            "injector_extension_id": cast(str | None, self.injector.get("injector_extension_id")),
            "injector_service_worker_url_includes": cast(list[str], self.injector["injector_service_worker_url_includes"]),
            "injector_service_worker_url_suffixes": cast(list[str], self.injector["injector_service_worker_url_suffixes"]),
            "injector_trust_service_worker_target": trust_service_worker_target,
            "injector_require_service_worker_target": self.injector["injector_require_service_worker_target"] or self.injector.get("injector_mode") == "discover",
            "injector_service_worker_ready_expression": cast(str | None, self.injector.get("injector_service_worker_ready_expression")),
            "injector_cdp_send_timeout_ms": self.client["client_cdp_send_timeout_ms"],
            "injector_execution_context_timeout_ms": self.injector["injector_execution_context_timeout_ms"],
            "injector_service_worker_probe_timeout_ms": self.injector["injector_service_worker_probe_timeout_ms"],
            "injector_service_worker_ready_timeout_ms": self.injector["injector_service_worker_ready_timeout_ms"],
            "injector_service_worker_poll_interval_ms": self.injector["injector_service_worker_poll_interval_ms"],
            "injector_target_session_poll_interval_ms": self.injector["injector_target_session_poll_interval_ms"],
        }

    def _inject_extension(self, injectors: list[ExtensionInjector] | None = None) -> ExtensionInfo:
        injectors = injectors or self._extension_injectors_for_config()
        errors: list[str] = []
        for injector in injectors:
            injector.update(self._base_extension_injector_config(self._send_message))
            try:
                injector.prepare()
                result = injector.inject()
                if result:
                    return cast(ExtensionInfo, result)
            except Exception as error:
                injector.last_error = error
                errors.append(f"{type(injector).__name__}: {error}")
        detail = f"\n\n{chr(10).join(errors)}" if errors else ""
        raise RuntimeError(f"Cannot install, discover, or borrow the ModCDP extension in the running browser.{detail}")

    # --- internals ---------------------------------------------------------

    def _send_raw(self, wrapped: TranslatedCommand) -> Any:
        if wrapped["target"] == "direct_cdp":
            step = wrapped["steps"][0]
            return self._send_message(step["method"], step.get("params") or {}, step.get("sessionId"))
        if wrapped["target"] != "service_worker":
            raise RuntimeError(f"Unsupported command target {wrapped['target']!r}")

        result: ProtocolResult = {}
        unwrap: str | None = None
        for step in wrapped["steps"]:
            params = dict(step.get("params") or {})
            if step["method"] == "Runtime.callFunctionOn" and "executionContextId" not in params:
                if self.ext_execution_context_id is None:
                    self.ext_execution_context_id = self.auto_sessions.waitForExecutionContext(
                        self.ext_session_id,
                        self.injector["injector_execution_context_timeout_ms"],
                    )
                params["executionContextId"] = self.ext_execution_context_id
            result = self._send_message(step["method"], params, self.ext_session_id)
            unwrap = step.get("unwrap")
        return unwrap_response_if_needed(result, unwrap)

    def _hydrate_custom_surface(self) -> None:
        for command in self.custom_commands:
            self._register_custom_command(cast(ProtocolParams, command))
        for event in self.custom_events:
            if isinstance(event, str):
                continue
            self._register_custom_event(cast(ProtocolParams, event))

    def _hydrate_native_protocol_schemas(self) -> None:
        with self._schema_lock:
            for domain_name, domain_class in vars(generated_cdp).items():
                if not domain_name.endswith("Domain") or not isinstance(domain_class, type):
                    continue
                domain = domain_name.removesuffix("Domain")
                nested_classes = {
                    name: value
                    for name, value in vars(domain_class).items()
                    if isinstance(value, type) and issubclass(value, CDPModel)
                }
                for class_name, params_class in nested_classes.items():
                    if issubclass(params_class, CDPEvent):
                        event_name = getattr(params_class, "cdp_event_name", None)
                        if isinstance(event_name, str):
                            self._event_schemas[event_name] = TypeAdapter(params_class)
                        continue
                    if not class_name.startswith("_") or not class_name.endswith("Params"):
                        continue
                    command_base = class_name[1:-6]
                    result_class = nested_classes.get(f"_{command_base}Result")
                    if result_class is None:
                        continue
                    method = f"{domain}.{command_base[:1].lower()}{command_base[1:]}"
                    if issubclass(params_class, CDPParams):
                        self._command_params_schemas[method] = TypeAdapter(params_class)
                    self._command_result_schemas[method] = TypeAdapter(result_class)
                    self._command_result_model_schemas.add(method)

    def _register_custom_command(self, params: ProtocolParams) -> None:
        name = params.get("name")
        if not isinstance(name, str) or not name:
            raise TypeError("name must be a non-empty string")
        params_schema, _, _ = self._adapter_from_optional_schema(params.get("params_schema"), "params_schema")
        result_schema, _, result_is_model = self._adapter_from_optional_schema(params.get("result_schema"), "result_schema")
        with self._schema_lock:
            if params_schema is not None:
                self._command_params_schemas[name] = params_schema
            if result_schema is not None:
                self._command_result_schemas[name] = result_schema
            if result_is_model:
                self._command_result_model_schemas.add(name)
            else:
                self._command_result_model_schemas.discard(name)

    def _custom_command_wire_params(self, params: ProtocolParams) -> ProtocolParams:
        wire = dict(params)
        _, params_schema, _ = self._adapter_from_optional_schema(wire.get("params_schema"), "params_schema")
        _, result_schema, _ = self._adapter_from_optional_schema(wire.get("result_schema"), "result_schema")
        if "params_schema" in wire:
            wire["params_schema"] = cast(JsonValue, params_schema)
        if "result_schema" in wire:
            wire["result_schema"] = cast(JsonValue, result_schema)
        return cast(ProtocolParams, wire)

    def _custom_event_wire_params(self, params: ProtocolParams) -> ProtocolParams:
        wire = dict(params)
        _, event_schema, _ = self._adapter_from_optional_schema(wire.get("event_schema"), "event_schema")
        if "event_schema" in wire:
            wire["event_schema"] = cast(JsonValue, event_schema)
        return cast(ProtocolParams, wire)

    def _register_custom_event(self, params: ProtocolParams) -> None:
        name = params.get("name")
        if not isinstance(name, str) or not name:
            raise TypeError("name must be a non-empty string")
        event_schema, _, event_is_model = self._adapter_from_optional_schema(params.get("event_schema"), "event_schema")
        if event_schema is not None:
            with self._schema_lock:
                self._event_schemas[name] = event_schema
                if event_is_model:
                    self._event_model_schemas.add(name)
                else:
                    self._event_model_schemas.discard(name)

    def _adapter_from_optional_schema(self, schema: Any | None, field_name: str) -> tuple[TypeAdapter[Any] | None, dict[str, Any] | None, bool]:
        if schema is None:
            return None, None, False
        if isinstance(schema, type) and issubclass(schema, BaseModel):
            return TypeAdapter(schema), schema.model_json_schema(), True
        if not isinstance(schema, Mapping):
            raise TypeError(f"{field_name} must be a JSON Schema object")
        schema_dict = dict(cast(Mapping[str, Any], schema))
        return type_adapter_from_json_schema(schema_dict), schema_dict, False

    def _validate_command_params(self, method: str, params: ProtocolParams) -> ProtocolParams:
        with self._schema_lock:
            adapter = self._command_params_schemas.get(method)
        if adapter is None:
            return params
        try:
            validated = adapter.validate_python(dict(params), strict=True)
        except ValidationError as e:
            raise ValueError(f"{method} params did not match params_schema: {e}") from e
        jsonable = (
            validated.model_dump(mode="json", exclude_none=True, by_alias=True)
            if isinstance(validated, BaseModel)
            else to_jsonable_python(validated)
        )
        if not isinstance(jsonable, Mapping):
            raise ValueError(f"{method} params_schema must validate to a JSON object")
        return cast(ProtocolParams, dict(jsonable))

    def _validate_command_result(self, method: str, result: Any) -> Any:
        with self._schema_lock:
            adapter = self._command_result_schemas.get(method)
        if adapter is None:
            return result
        try:
            validated = adapter.validate_python(result, strict=True)
        except ValidationError as e:
            raise ValueError(f"{method} result did not match result_schema: {e}") from e
        if isinstance(validated, CDPModel):
            return cast(JsonValue, validated.model_dump(mode="json", exclude_none=True, by_alias=True))
        if method in self._command_result_model_schemas and isinstance(validated, BaseModel):
            fields = list(type(validated).model_fields)
            if len(fields) == 1:
                return cast(JsonValue, getattr(validated, fields[0]))
            return cast(JsonValue, validated)
        return cast(JsonValue, to_jsonable_python(validated))

    def _validate_event_payload(self, event: str, payload: ProtocolPayload) -> Any | None:
        with self._schema_lock:
            adapter = self._event_schemas.get(event)
            event_class = self._event_classes.get(event)
        if adapter is None and event_class is None:
            return dict(payload)
        if adapter is None and event_class is not None:
            try:
                return cast(ProtocolPayload, event_class.model_validate(dict(payload)).model_dump(mode="json", exclude_none=True, by_alias=True))
            except ValidationError as e:
                raise ValueError(f"{event} event did not match native event schema: {e}") from e
        assert adapter is not None
        try:
            validated = adapter.validate_python(dict(payload), strict=True)
        except ValidationError as direct_error:
            if set(payload.keys()) != {"value"}:
                raise ValueError(f"{event} event did not match event_schema: {direct_error}") from direct_error
            try:
                validated = adapter.validate_python(payload["value"], strict=True)
            except ValidationError as value_error:
                raise ValueError(f"{event} event did not match event_schema: {value_error}") from value_error
            if event in self._event_model_schemas:
                return cast(ProtocolPayload, validated)
            return {"value": cast(JsonValue, to_jsonable_python(validated))}
        if isinstance(validated, CDPModel):
            return cast(ProtocolPayload, validated.model_dump(mode="json", exclude_none=True, by_alias=True))
        if event in self._event_model_schemas:
            return cast(ProtocolPayload, validated)
        jsonable = to_jsonable_python(validated)
        if isinstance(jsonable, Mapping):
            return cast(ProtocolPayload, dict(jsonable))
        return {"value": cast(JsonValue, jsonable)}

    def _measure_ping_latency(self) -> ModCDPPingLatency | None:
        sent_at = int(time.time() * 1000)
        done: Queue[ProtocolPayload] = Queue()

        def on_pong(payload: ProtocolPayload) -> None:
            done.put(payload or {})

        self._handlers.setdefault("Mod.pong", []).append(on_pong)
        try:
            self.send("Mod.ping", {"sent_at": sent_at})
            payload = done.get(timeout=10)
        except Empty:
            return self.latency
        except Exception:
            return self.latency
        finally:
            handlers = self._handlers.get("Mod.pong") or []
            if on_pong in handlers:
                handlers.remove(on_pong)

        returned_at = int(time.time() * 1000)
        raw_received_at = payload.get("received_at")
        received_at = raw_received_at if isinstance(raw_received_at, (int, float)) else None
        latency: ModCDPPingLatency = {
            "sent_at": sent_at,
            "received_at": received_at,
            "returned_at": returned_at,
            "round_trip_ms": returned_at - sent_at,
            "service_worker_ms": received_at - sent_at if received_at is not None else None,
            "return_path_ms": returned_at - received_at if received_at is not None else None,
        }
        self.latency = latency
        return latency

    def _send_message(
        self,
        method: str,
        params: MessageParams | None = None,
        session_id: str | None = None,
        timeout: float | None = None,
        record_raw_timing: bool = False,
    ) -> ProtocolResult:
        effective_timeout = self.client["client_cdp_send_timeout_ms"] / 1000 if timeout is None else timeout
        with self._lock:
            self._next_id += 1
            msg_id = self._next_id
            done: Queue[CdpMessage] = Queue()
            self._pending[msg_id] = (method, done)
        started_at = int(time.time() * 1000)
        msg: CdpMessage = {"id": msg_id, "method": method, "params": params or {}}
        if session_id:
            msg["sessionId"] = session_id
        try:
            if self.transport is not None:
                self.transport.send(cast(dict[str, Any], msg))
            else:
                raise RuntimeError("ModCDP upstream is not connected")
        except Exception:
            with self._lock:
                self._pending.pop(msg_id, None)
            raise
        try:
            response = done.get(timeout=effective_timeout)
        except Empty:
            with self._lock:
                self._pending.pop(msg_id, None)
            raise RuntimeError(f"{method} timed out after {int(effective_timeout * 1000)}ms")
        err = response.get("error")
        if err:
            raise RuntimeError(f"{method} failed: {err.get('message', err)}")
        if record_raw_timing:
            completed_at = int(time.time() * 1000)
            self.last_raw_timing = {
                "method": method,
                "started_at": started_at,
                "completed_at": completed_at,
                "duration_ms": completed_at - started_at,
            }
        return _json_object(response.get("result"))

    def _reject_all(self, error: Exception) -> None:
        with self._lock:
            pending = list(self._pending.values())
            self._pending.clear()
        for _, done in pending:
            done.put({"error": {"message": str(error)}})

    def _on_recv(self, msg: CdpMessage) -> None:
        if "id" in msg and msg["id"] is not None:
            with self._lock:
                entry = self._pending.pop(msg["id"], None)
            if entry:
                entry[1].put(msg)
            return
        method = msg.get("method")
        raw_params = msg.get("params")
        params = cast(ProtocolParams, raw_params) if isinstance(raw_params, Mapping) else {}
        if isinstance(method, str):
            session_id = msg.get("sessionId")
            self.auto_sessions.recordProtocolEvent(method, params, session_id if isinstance(session_id, str) else None)
        if method and self.ext_session_id is not None and msg.get("sessionId") == self.ext_session_id:
            session_id = msg.get("sessionId")
            u = unwrap_event_if_needed(
                method,
                params,
                session_id if isinstance(session_id, str) else None,
                self.ext_session_id,
            )
            if u:
                validated_payload = self._validate_event_payload(u["event"], u["data"])
                if validated_payload is None:
                    return
                for handler in list(self._handlers.get(u["event"], [])):
                    def run_wrapped_event(handler=handler, payload=validated_payload, event_name=u["event"]):
                        self._run_handler(handler, payload, event_name)
                    threading.Thread(target=run_wrapped_event, daemon=True).start()
            return
        if method:
            validated_payload = self._validate_event_payload(method, dict(params))
            if validated_payload is None:
                return
            for handler in list(self._handlers.get(method, [])):
                def run_method_event(handler=handler, payload=validated_payload, event_name=method):
                    self._run_handler(handler, payload, event_name)
                threading.Thread(target=run_method_event, daemon=True).start()
