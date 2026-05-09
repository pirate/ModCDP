"""ModCDPClient (Python): importable, no CLI, no demo code.

Constructor option groups mirror the JS / Go ports:
    launch            browser/session creation and cleanup
    upstream          message transport to raw CDP or a ModCDP server
    extension         raw-CDP extension discovery/injection/borrowing
    client.routes     client-side direct_cdp/service_worker routing
    server            ModCDPServer.configure params
    client            client routing and client-owned send/event timings
    extension         extension discovery, wake, probe, and keepalive timings
    upstream          upstream transport options and upstream-owned timings

Public methods: connect(), send(method, params), on(event, handler), close(), _cdp.send(), _cdp.on().
Synchronous (blocking) API; one background thread reads messages off the WS.
"""

import json
import asyncio
import inspect
import re
import sys
import threading
import time
import urllib.error
import urllib.parse
import urllib.request
from collections.abc import Mapping, Sequence
from pathlib import Path
from queue import Queue, Empty
from typing import Any, cast

from pydantic import BaseModel, TypeAdapter, ValidationError
from pydantic_core import to_jsonable_python
from .AutoSessionRouter import AutoSessionRouter
from .jsonschema import type_adapter_from_json_schema
from .cdp.surface import AwaitableDict, CDPEvent, CDPSurfaceMixin, cdp_event_name, install_cdp_surface
from .BrowserbaseBrowserLauncher import BrowserbaseBrowserLauncher
from .BBBrowserExtensionInjector import BBBrowserExtensionInjector
from .BorrowedExtensionInjector import BorrowedExtensionInjector
from .DiscoveredExtensionInjector import DiscoveredExtensionInjector
from .ExtensionInjector import (
    DEFAULT_MODCDP_WAKE_PATH,
    DEFAULT_MODCDP_SERVICE_WORKER_URL_SUFFIXES,
    ExtensionInjector,
    ExtensionInjectorConfig,
)
from .ExtensionsLoadUnpackedInjector import ExtensionsLoadUnpackedInjector
from .LocalBrowserLaunchExtensionInjector import LocalBrowserLaunchExtensionInjector
from .LocalBrowserLauncher import LocalBrowserLauncher
from .NoopBrowserLauncher import NoopBrowserLauncher
from .RemoteBrowserLauncher import RemoteBrowserLauncher
from .NativeMessagingUpstreamTransport import NativeMessagingUpstreamTransport
from .NatsUpstreamTransport import NatsUpstreamTransport
from .PipeUpstreamTransport import PipeUpstreamTransport
from .ReverseWebSocketUpstreamTransport import ReverseWebSocketUpstreamTransport
from .UpstreamTransport import UpstreamTransport
from .WebSocketUpstreamTransport import WebSocketUpstreamTransport
from .translate import (
    CUSTOM_EVENT_BINDING_NAME,
    DEFAULT_CLIENT_ROUTES,
    UPSTREAM_EVENT_BINDING_NAME,
    wrap_command_if_needed,
    unwrap_event_if_needed,
    unwrap_response_if_needed,
)
from .BrowserLauncher import BrowserLaunchOptions
from .types import (
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
    WebSocketLike,
)


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

    def evaluate(self, *, expression: str, params: Mapping[str, Any] | None = None, cdpSessionId: str | None = None):
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
    ):
        payload: dict[str, Any] = {"name": name}
        if params_schema is not None:
            payload["params_schema"] = params_schema
        if result_schema is not None:
            payload["result_schema"] = result_schema
        if expression is not None:
            payload["expression"] = expression
        return self._client._send_command("Mod.addCustomCommand", payload)

    def addCustomEvent(self, name: str, *, event_schema: Any | None = None):
        payload: dict[str, Any] = {"name": name}
        if event_schema is not None:
            payload["event_schema"] = event_schema
        return self._client._send_command("Mod.addCustomEvent", payload)

    def addMiddleware(self, *, phase: str, expression: str, name: str | None = None):
        payload: dict[str, Any] = {"phase": phase, "expression": expression}
        if name is not None:
            payload["name"] = name
        return self._client._send_command("Mod.addMiddleware", payload)

    def configure(self, **params: Any):
        return self._client._send_command("Mod.configure", params)

    def ping(self, **params: Any):
        return self._client._send_command("Mod.ping", params)

MODCDP_READY_EXPRESSION = (
    "Boolean(globalThis.ModCDP?.__ModCDPServerVersion === 1 && "
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


def websocket_url_for(endpoint: str) -> str:
    if re.match(r"^wss?://", endpoint, re.I):
        return endpoint
    http_endpoint = endpoint if re.match(r"^[a-z][a-z\d+\-.]*://", endpoint, re.I) else f"http://{endpoint}"
    try:
        r = urllib.request.urlopen(f"{http_endpoint}/json/version", timeout=5)
    except urllib.error.HTTPError as e:
        if e.code != 404:
            raise
        parsed = urllib.parse.urlparse(http_endpoint)
        scheme = "wss" if parsed.scheme == "https" else "ws"
        return urllib.parse.urlunparse((scheme, parsed.netloc, "/devtools/browser", "", "", ""))
    with r:
        parsed: object = json.loads(r.read())
    if not isinstance(parsed, dict):
        raise RuntimeError(f"HTTP discovery for {endpoint} returned invalid JSON")
    parsed_obj = cast(Mapping[str, object], parsed)
    ws_url = parsed_obj.get("webSocketDebuggerUrl")
    if not ws_url:
        raise RuntimeError(f"HTTP discovery for {endpoint} returned no webSocketDebuggerUrl")
    if not isinstance(ws_url, str):
        raise RuntimeError(f"HTTP discovery for {endpoint} returned a non-string webSocketDebuggerUrl")
    return ws_url


def _json_object(value: JsonValue | None) -> ProtocolResult:
    return value if isinstance(value, dict) else {}


def default_extension_path() -> str | None:
    bundled_extension = Path(__file__).resolve().parent / "extension.zip"
    return str(bundled_extension) if bundled_extension.exists() else None


class ModCDPClient(CDPSurfaceMixin):
    def __init__(
        self,
        launch: Mapping[str, Any] | None = None,
        upstream: Mapping[str, Any] | None = None,
        extension: Mapping[str, Any] | None = None,
        client: Mapping[str, Any] | None = None,
        server: Mapping[str, JsonValue] | None | object = DEFAULT_SERVER,
        custom_commands: Sequence[ModCDPAddCustomCommandParams] | None = None,
        custom_events: Sequence[ModCDPAddCustomEventParams] | None = None,
        custom_middlewares: Sequence[ModCDPAddMiddlewareParams] | None = None,
    ) -> None:
        launch_input = dict(launch or {})
        upstream_input = dict(upstream or {})
        extension_input = dict(extension or {})
        client_input = dict(client or {})
        upstream_mode = str(upstream_input.get("mode") or "ws")
        self.upstream: dict[str, Any] = {
            "mode": upstream_mode,
            "ws_url": upstream_input.get("ws_url"),
            "nats_url": upstream_input.get("nats_url"),
            "nats_subject_prefix": upstream_input.get("nats_subject_prefix"),
            "reversews_bind": upstream_input.get("reversews_bind"),
            "nativemessaging_manifest": upstream_input.get("nativemessaging_manifest"),
            "ws_connect_error_settle_timeout_ms": int(
                upstream_input.get("ws_connect_error_settle_timeout_ms")
                or DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS
            ),
        }
        self.upstream_endpoint_kind = "raw_cdp" if self.upstream["mode"] in ("ws", "pipe") else "modcdp_server"
        launch_mode = launch_input.get("mode") or (
            "none" if self.upstream_endpoint_kind == "modcdp_server" else "remote" if self.upstream.get("ws_url") else "local"
        )
        self.launch: dict[str, Any] = {
            "mode": launch_mode,
            "executable_path": launch_input.get("executable_path"),
            "user_data_dir": launch_input.get("user_data_dir"),
            "options": dict(cast(Mapping[str, Any], launch_input.get("options") or {})),
        }
        extension_mode = extension_input.get("mode") or (
            "auto" if self.upstream_endpoint_kind == "raw_cdp" else "none"
        )
        self.extension: dict[str, Any] = {
            "mode": extension_mode,
            "path": extension_input.get("path") or default_extension_path(),
            "extension_id": extension_input.get("extension_id"),
            "wake_path": extension_input.get("wake_path") or DEFAULT_MODCDP_WAKE_PATH,
            "wake_url": extension_input.get("wake_url"),
            "service_worker_url_includes": list(cast(Sequence[str], extension_input.get("service_worker_url_includes") or [])),
            "service_worker_url_suffixes": list(
                cast(
                    Sequence[str],
                    extension_input.get("service_worker_url_suffixes")
                    or DEFAULT_MODCDP_SERVICE_WORKER_URL_SUFFIXES,
                )
            ),
            "trust_service_worker_target": bool(extension_input.get("trust_service_worker_target", False)),
            "require_service_worker_target": bool(extension_input.get("require_service_worker_target", False)),
            "service_worker_ready_expression": extension_input.get("service_worker_ready_expression"),
            "execution_context_timeout_ms": int(
                extension_input.get("execution_context_timeout_ms") or DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS
            ),
            "service_worker_probe_timeout_ms": int(
                extension_input.get("service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS
            ),
            "service_worker_ready_timeout_ms": int(
                extension_input.get("service_worker_ready_timeout_ms") or DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS
            ),
            "service_worker_poll_interval_ms": int(
                extension_input.get("service_worker_poll_interval_ms") or DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS
            ),
            "target_session_poll_interval_ms": int(
                extension_input.get("target_session_poll_interval_ms") or DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS
            ),
        }
        self.client: dict[str, Any] = {
            "routes": {
                **DEFAULT_CLIENT_ROUTES,
                **dict(cast(Mapping[str, str], client_input.get("routes") or {})),
            },
            "hydrate_aliases": bool(client_input.get("hydrate_aliases", True)),
            "mirror_upstream_events": bool(client_input.get("mirror_upstream_events", True)),
            "cdp_send_timeout_ms": int(client_input.get("cdp_send_timeout_ms") or DEFAULT_CDP_SEND_TIMEOUT_MS),
            "event_wait_timeout_ms": int(client_input.get("event_wait_timeout_ms") or DEFAULT_EVENT_WAIT_TIMEOUT_MS),
        }
        self.cdp_url: str | None = cast(str | None, self.upstream.get("ws_url"))
        if server is DEFAULT_SERVER:
            self.server: ModCDPServerConfig | None = {"routes": {"*.*": "chrome_debugger"}} if self.upstream_endpoint_kind == "modcdp_server" else {}
        elif server is None:
            self.server = None
        elif isinstance(server, Mapping):
            self.server = cast(ModCDPServerConfig, {
                **({"routes": {"*.*": "chrome_debugger"}} if self.upstream_endpoint_kind == "modcdp_server" else {}),
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
        self.latency: ModCDPPingLatency | None = None
        self.connect_timing: ModCDPConnectTiming | None = None
        self.last_command_timing: ModCDPCommandTiming | None = None
        self.last_raw_timing: ModCDPRawTiming | None = None

        self._ws: WebSocketLike | None = None
        self.transport: UpstreamTransport | None = None
        self._next_id = 0
        self._pending: dict[int, PendingEntry] = {}
        self._handlers: dict[str, list[Handler]] = {}
        self._lock = threading.Lock()
        self.auto_sessions = AutoSessionRouter(
            lambda method, params=None, session_id=None: self._send_message(method, params or {}, session_id),
            lambda: self.extension["execution_context_timeout_ms"],
        )
        self._schema_lock = threading.RLock()
        self._command_params_schemas: dict[str, TypeAdapter[Any]] = {}
        self._command_result_schemas: dict[str, TypeAdapter[Any]] = {}
        self._event_schemas: dict[str, TypeAdapter[Any]] = {}
        self._command_result_model_schemas: set[str] = set()
        self._event_model_schemas: set[str] = set()
        self._event_classes: dict[str, type[CDPEvent]] = {}
        install_cdp_surface(self)
        self.Mod = _ModDomain(self)
        self._reader_thread: threading.Thread | None = None
        self._closed = False
        self._launched_browser: Any | None = None
        self._extension_injectors: list[ExtensionInjector] = []
        self._cdp = _RawCDP(self)
        self._hydrate_custom_surface()

    def connect(self) -> "ModCDPClient":
        connect_started_at = int(time.time() * 1000)
        if self.upstream_endpoint_kind == "modcdp_server":
            transport_started_at = int(time.time() * 1000)
            launcher = self._browser_launcher()
            launcher.update(self._launch_options())
            transport = self._upstream_transport()
            transport.update(
                {
                    "ws_url": self.upstream.get("ws_url"),
                    "cdp_url": self.upstream.get("ws_url"),
                    "nats_url": self.upstream.get("nats_url"),
                    "nats_subject_prefix": self.upstream.get("nats_subject_prefix"),
                    "reversews_bind": self.upstream.get("reversews_bind"),
                    "manifest_path": self.upstream.get("nativemessaging_manifest"),
                    "extension_id": self.extension.get("extension_id"),
                }
            )
            injectors = self._extension_injectors_for_config()
            self._extension_injectors = injectors
            for injector in injectors:
                injector.update(self._base_extension_injector_config(None))
                injector.update(cast(ExtensionInjectorConfig, transport.getInjectorConfig()))
                injector.prepare()
                launcher.update(injector.getLauncherConfig())
                transport.update(injector.getTransportConfig())
            launcher.update(transport.getLauncherConfig())
            transport.update(launcher.getTransportConfig())
            transport.connect()
            self.transport = transport
            transport.onRecv(lambda message: self._on_recv(cast(CdpMessage, message)))
            transport.onClose(lambda error: self._reject_all(error))
            if self.launch.get("mode") != "none":
                launched = launcher.launch()
                self._launched_browser = launched
                self.cdp_url = cast(str | None, launched.get("ws_url") or launched.get("cdp_url"))
                transport.update(launcher.getTransportConfig())
                for injector in injectors:
                    transport.update(injector.getTransportConfig())
                server_config = {"loopback_cdp_url": self.cdp_url} if self.cdp_url else {}
                server_config.update(transport.getServerConfig())
                if self.server is not None and server_config.get("loopback_cdp_url") and "loopback_cdp_url" not in self.server:
                    self.server = {**self.server, **server_config}
            transport_connected_at = int(time.time() * 1000)
            transport.waitForPeer()
            if self.server is not None:
                self._send_message("Mod.configure", cast(ProtocolParams, self._server_configure_params()))
            threading.Thread(target=self._measure_ping_latency, daemon=True).start()
            connected_at = int(time.time() * 1000)
            self.connect_timing = cast(ModCDPConnectTiming, {
                "started_at": connect_started_at,
                "upstream_mode": self.upstream.get("mode"),
                "upstream_endpoint_kind": self.upstream_endpoint_kind,
                "transport_started_at": transport_started_at,
                "transport_connected_at": transport_connected_at,
                "transport_duration_ms": transport_connected_at - transport_started_at,
                "connected_at": connected_at,
                "duration_ms": connected_at - connect_started_at,
            })
            return self
        launcher = self._browser_launcher()
        launcher.update(self._launch_options())
        transport: UpstreamTransport | None = None
        if self.upstream.get("mode") != "ws" or self.cdp_url is not None:
            transport = self._upstream_transport()
            launcher.update(transport.getLauncherConfig())
        injectors = self._extension_injectors_for_config()
        self._extension_injectors = injectors
        for injector in injectors:
            injector.update(self._base_extension_injector_config(None))
            injector.update(launcher.getInjectorConfig())
            injector.prepare()
            launcher.update(injector.getLauncherConfig())
        if self.cdp_url is None:
            if self.launch.get("mode") == "none":
                raise RuntimeError("upstream.mode=ws requires upstream.ws_url or a launch mode that creates a browser.")
            launched = launcher.launch()
            self._launched_browser = launched
            self.cdp_url = cast(str | None, launched.get("ws_url") or launched.get("cdp_url"))
            if transport is not None:
                transport.update(launcher.getTransportConfig())
            for injector in injectors:
                injector.update(launcher.getInjectorConfig())
        input_cdp_url = self.cdp_url
        if input_cdp_url is None:
            raise RuntimeError("upstream.mode=ws requires an upstream CDP endpoint.")
        if self.upstream.get("mode") == "ws":
            self.cdp_url = websocket_url_for(input_cdp_url)
            self.upstream["ws_url"] = self.cdp_url
            if transport is None:
                transport = self._upstream_transport()
            transport.update({"ws_url": self.cdp_url, "cdp_url": self.cdp_url, "url": self.cdp_url})
        if transport is None:
            transport = self._upstream_transport()
        can_autoconfigure_loopback = self.cdp_url is not None and not self.cdp_url.startswith("pipe://")
        if self.server is not None and can_autoconfigure_loopback and "loopback_cdp_url" not in self.server:
            self.server = {**self.server, "loopback_cdp_url": self.cdp_url}
        elif self.server and isinstance(self.server.get("loopback_cdp_url"), str):
            loopback_url = self.server["loopback_cdp_url"]
            if loopback_url == input_cdp_url or loopback_url == self.cdp_url:
                self.server = {**self.server, "loopback_cdp_url": self.cdp_url}
        self.transport = transport
        self.transport.connect()
        if self.upstream.get("mode") == "ws":
            self._ws = cast(WebSocketLike, cast(WebSocketUpstreamTransport, self.transport).ws)
            self._reader_thread = threading.Thread(target=self._reader, daemon=True)
            self._reader_thread.start()
        else:
            self.transport.onRecv(lambda message: self._on_recv(cast(CdpMessage, message)))
            self.transport.onClose(lambda error: self._reject_all(error))

        self._send_message("Target.setAutoAttach", {
            "autoAttach": True,
            "waitForDebuggerOnStart": False,
            "flatten": True,
        })
        self._send_message("Target.setDiscoverTargets", {"discover": True})

        extension_started_at = int(time.time() * 1000)
        if self.extension.get("mode") == "none":
            raise RuntimeError("extension.mode='none' cannot be used with a raw_cdp upstream.")
        ext = self._inject_extension(injectors)
        extension_completed_at = int(time.time() * 1000)
        self.extension_id = ext["extension_id"]
        self.ext_target_id = ext["target_id"]
        self.ext_session_id = ext["session_id"]
        self._send_message("Runtime.enable", {}, self.ext_session_id)
        self._send_message("Runtime.addBinding", {"name": CUSTOM_EVENT_BINDING_NAME}, self.ext_session_id)
        if self.client["mirror_upstream_events"]:
            self._send_message("Runtime.addBinding", {"name": UPSTREAM_EVENT_BINDING_NAME}, self.ext_session_id)

        if self.server is not None:
            self._send_raw(wrap_command_if_needed(
                "Mod.configure",
                cast(ProtocolParams, self._server_configure_params()),
                routes=cast(ModCDPRoutes, self.client["routes"]),
                cdp_session_id=self.ext_session_id,
            ))
        threading.Thread(target=self._measure_ping_latency, daemon=True).start()
        connected_at = int(time.time() * 1000)
        self.connect_timing = {
            "started_at": connect_started_at,
            "extension_source": ext.get("source"),
            "extension_started_at": extension_started_at,
            "extension_completed_at": extension_completed_at,
            "extension_duration_ms": extension_completed_at - extension_started_at,
            "connected_at": connected_at,
            "duration_ms": connected_at - connect_started_at,
        }
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
            routes=cast(ModCDPRoutes, self.client["routes"]),
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
        self._handlers.setdefault(event_name, []).append(wrapped_handler)
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
        from .cdp.surface import DynamicDomain

        dynamic = DynamicDomain(self, domain)
        setattr(self, domain, dynamic)
        return dynamic

    def _server_configure_params(self) -> ModCDPServerConfig:
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
                "mode": self.upstream.get("mode"),
                **({"nats_url": self.upstream.get("nats_url")} if self.upstream.get("nats_url") else {}),
                **(
                    {"nats_subject_prefix": self.upstream.get("nats_subject_prefix")}
                    if self.upstream.get("nats_subject_prefix")
                    else {}
                ),
            },
            "client": {"routes": self.client["routes"]},
            "server": {
                "cdp_send_timeout_ms": self.client["cdp_send_timeout_ms"],
                "loopback_execution_context_timeout_ms": self.extension["execution_context_timeout_ms"],
                "ws_connect_error_settle_timeout_ms": self.upstream["ws_connect_error_settle_timeout_ms"],
                **(self.server or {}),
            },
            "custom_events": custom_events,
            "custom_commands": custom_commands,
            "custom_middlewares": custom_middlewares,
        })

    def close(self) -> None:
        if self._closed:
            return
        for injector in self._extension_injectors:
            try:
                injector.close()
            except Exception:
                pass
        self._extension_injectors = []
        self._closed = True
        try:
            if self.transport:
                self.transport.close()
            elif self._ws:
                self._ws.close()
        except Exception:
            pass
        if self._reader_thread is not None and self._reader_thread.is_alive():
            self._reader_thread.join(timeout=1)
        self.transport = None
        self._ws = None
        if self._launched_browser is not None:
            self._launched_browser["close"]()
            self._launched_browser = None

    def _session_id_for_target(self, target_id: str, timeout: float = 0) -> str | None:
        if timeout <= 0:
            return self.auto_sessions.sessionIdForTarget(target_id)
        deadline = time.time() + timeout
        while time.time() <= deadline:
            session_id = self.auto_sessions.sessionIdForTarget(target_id)
            if session_id:
                return session_id
            time.sleep(self.extension["target_session_poll_interval_ms"] / 1000)
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
        if self.launch.get("mode") == "local":
            return LocalBrowserLauncher(self._launch_options())
        if self.launch.get("mode") == "remote":
            return RemoteBrowserLauncher(self._launch_options(), self.cdp_url)
        if self.launch.get("mode") == "bb":
            return BrowserbaseBrowserLauncher(self._launch_options())
        return NoopBrowserLauncher(self._launch_options())

    def _launch_options(self) -> BrowserLaunchOptions:
        launch_options = cast(BrowserLaunchOptions, dict(cast(Mapping[str, Any], self.launch.get("options") or {})))
        if self.launch.get("executable_path"):
            launch_options["executable_path"] = cast(str, self.launch["executable_path"])
        if self.launch.get("user_data_dir"):
            launch_options["user_data_dir"] = cast(str, self.launch["user_data_dir"])
        return launch_options

    def _upstream_transport(self):
        mode = self.upstream.get("mode")
        if mode == "ws":
            return WebSocketUpstreamTransport(self.cdp_url, timeout_s=self.client["cdp_send_timeout_ms"] / 1000)
        if mode == "pipe":
            return PipeUpstreamTransport()
        if mode == "reversews":
            return ReverseWebSocketUpstreamTransport(str(self.upstream.get("reversews_bind") or "127.0.0.1:29292"))
        if mode == "nativemessaging":
            return NativeMessagingUpstreamTransport({"manifest_path": self.upstream.get("nativemessaging_manifest")})
        if mode == "nats":
            return NatsUpstreamTransport(
                {
                    "url": self.upstream.get("nats_url"),
                    "subject_prefix": self.upstream.get("nats_subject_prefix"),
                }
            )
        raise RuntimeError(f"unknown upstream.mode={mode}")

    def _extension_injectors_for_config(self) -> list[ExtensionInjector]:
        mode = self.extension.get("mode")
        if mode == "none":
            return []
        injectors: list[ExtensionInjector] = []
        if mode in ("auto", "discover"):
            injectors.append(DiscoveredExtensionInjector())
        if mode in ("auto", "inject"):
            if self.launch.get("mode") == "bb":
                injectors.append(BBBrowserExtensionInjector())
            if self.launch.get("mode") == "local":
                injectors.append(LocalBrowserLaunchExtensionInjector())
            injectors.append(ExtensionsLoadUnpackedInjector())
        if mode in ("auto", "borrow"):
            injectors.append(BorrowedExtensionInjector())
        return injectors

    def _base_extension_injector_config(self, send: Any | None) -> ExtensionInjectorConfig:
        trust_matched_service_worker = (
            self.extension["trust_service_worker_target"]
            or len(self.extension["service_worker_url_includes"]) > 0
            or any(
                len([part for part in suffix.split("/") if part]) > 1
                for suffix in self.extension["service_worker_url_suffixes"]
            )
        )

        def send_cdp(method: str, params: ProtocolParams | None = None, session_id: str | None = None) -> ProtocolResult:
            if send is None:
                raise RuntimeError("Extension injector CDP send is not connected.")
            return self._send_message(
                method,
                params or {},
                session_id,
                timeout=self.client["cdp_send_timeout_ms"] / 1000,
            )

        def attach_to_target(target_id: str) -> str | None:
            return self._ensure_session_id_for_target(
                target_id,
                timeout=self.extension["service_worker_probe_timeout_ms"] / 1000,
                allow_attach=True,
            )

        return {
            "send": send_cdp if send is not None else None,
            "sessionIdForTarget": self.auto_sessions.sessionIdForTarget,
            "attachToTarget": attach_to_target if send is not None else None,
            "waitForExecutionContext": self.auto_sessions.waitForExecutionContext,
            "extension_path": cast(str | None, self.extension.get("path")),
            "extension_id": cast(str | None, self.extension.get("extension_id")),
            "wake_path": cast(str | None, self.extension.get("wake_path")),
            "wake_url": cast(str | None, self.extension.get("wake_url")),
            "service_worker_url_includes": cast(list[str], self.extension["service_worker_url_includes"]),
            "service_worker_url_suffixes": cast(list[str], self.extension["service_worker_url_suffixes"]),
            "trust_matched_service_worker": trust_matched_service_worker,
            "require_service_worker_target": self.extension["require_service_worker_target"] or self.extension.get("mode") == "discover",
            "service_worker_ready_expression": cast(str | None, self.extension.get("service_worker_ready_expression")),
            "cdp_send_timeout_ms": self.client["cdp_send_timeout_ms"],
            "execution_context_timeout_ms": self.extension["execution_context_timeout_ms"],
            "service_worker_probe_timeout_ms": self.extension["service_worker_probe_timeout_ms"],
            "service_worker_ready_timeout_ms": self.extension["service_worker_ready_timeout_ms"],
            "service_worker_poll_interval_ms": self.extension["service_worker_poll_interval_ms"],
            "target_session_poll_interval_ms": self.extension["target_session_poll_interval_ms"],
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
            return self._send_message(step["method"], step.get("params") or {})
        if wrapped["target"] != "service_worker":
            raise RuntimeError(f"Unsupported command target {wrapped['target']!r}")

        result: ProtocolResult = {}
        unwrap: str | None = None
        for step in wrapped["steps"]:
            result = self._send_message(step["method"], step.get("params") or {}, self.ext_session_id)
            unwrap = step.get("unwrap")
        return unwrap_response_if_needed(result, unwrap)

    def _hydrate_custom_surface(self) -> None:
        for command in self.custom_commands:
            self._register_custom_command(cast(ProtocolParams, command))
        for event in self.custom_events:
            if isinstance(event, str):
                continue
            self._register_custom_event(cast(ProtocolParams, event))

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
            validated = adapter.validate_python(dict(params))
        except ValidationError as e:
            raise ValueError(f"{method} params did not match params_schema: {e}") from e
        jsonable = to_jsonable_python(validated)
        if not isinstance(jsonable, Mapping):
            raise ValueError(f"{method} params_schema must validate to a JSON object")
        return cast(ProtocolParams, dict(jsonable))

    def _validate_command_result(self, method: str, result: Any) -> Any:
        with self._schema_lock:
            adapter = self._command_result_schemas.get(method)
        if adapter is None:
            return result
        try:
            validated = adapter.validate_python(result)
        except ValidationError as e:
            raise ValueError(f"{method} result did not match result_schema: {e}") from e
        if method in self._command_result_model_schemas and isinstance(validated, BaseModel):
            fields = list(type(validated).model_fields)
            if len(fields) == 1:
                return cast(JsonValue, getattr(validated, fields[0]))
            return cast(JsonValue, validated)
        return cast(JsonValue, to_jsonable_python(validated))

    def _validate_event_payload(self, event: str, payload: ProtocolPayload) -> Any | None:
        with self._schema_lock:
            adapter = self._event_schemas.get(event)
        if adapter is None:
            return dict(payload)
        try:
            validated = adapter.validate_python(dict(payload))
        except ValidationError as direct_error:
            if set(payload.keys()) != {"value"}:
                print(f"[ModCDPClient] event {event} did not match event_schema: {direct_error}", file=sys.stderr)
                return None
            try:
                validated = adapter.validate_python(payload["value"])
            except ValidationError as value_error:
                print(f"[ModCDPClient] event {event} did not match event_schema: {value_error}", file=sys.stderr)
                return None
            if event in self._event_model_schemas:
                return cast(ProtocolPayload, validated)
            return {"value": cast(JsonValue, to_jsonable_python(validated))}
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
            self.send("Mod.ping", {"sentAt": sent_at})
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
        raw_received_at = payload.get("receivedAt")
        received_at = raw_received_at if isinstance(raw_received_at, (int, float)) else None
        latency: ModCDPPingLatency = {
            "sentAt": sent_at,
            "receivedAt": received_at,
            "returnedAt": returned_at,
            "roundTripMs": returned_at - sent_at,
            "serviceWorkerMs": received_at - sent_at if received_at is not None else None,
            "returnPathMs": returned_at - received_at if received_at is not None else None,
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
        effective_timeout = self.client["cdp_send_timeout_ms"] / 1000 if timeout is None else timeout
        with self._lock:
            self._next_id += 1
            msg_id = self._next_id
            done: Queue[CdpMessage] = Queue()
            self._pending[msg_id] = (method, done)
        started_at = int(time.time() * 1000)
        msg: CdpMessage = {"id": msg_id, "method": method, "params": params or {}}
        if session_id:
            msg["sessionId"] = session_id
        ws = self._ws
        try:
            if ws is not None:
                ws.send(json.dumps(msg))
            elif self.transport is not None:
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
        if method and msg.get("sessionId") == self.ext_session_id:
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
                for handler in self._handlers.get(u["event"], []):
                    def run_wrapped_event(handler=handler, payload=validated_payload, event_name=u["event"]):
                        self._run_handler(handler, payload, event_name)
                    threading.Thread(target=run_wrapped_event, daemon=True).start()
            return
        if method:
            validated_payload = self._validate_event_payload(method, dict(params))
            if validated_payload is None:
                return
            for handler in self._handlers.get(method, []):
                def run_method_event(handler=handler, payload=validated_payload, event_name=method):
                    self._run_handler(handler, payload, event_name)
                threading.Thread(target=run_method_event, daemon=True).start()

    def _reader(self) -> None:
        ws = self._ws
        if ws is None:
            return
        try:
            while True:
                raw = ws.recv()
                if not raw:
                    break
                if isinstance(raw, bytes):
                    raw = raw.decode()
                parsed: object = json.loads(raw)
                if not isinstance(parsed, dict):
                    continue
                self._on_recv(cast(CdpMessage, parsed))
        except Exception as e:
            if not self._closed:
                print(f"[ModCDPClient] reader exited: {e}")
        finally:
            with self._lock:
                pending = list(self._pending.values())
                self._pending.clear()
            for _, done in pending:
                done.put({"error": {"message": "connection closed"}})
