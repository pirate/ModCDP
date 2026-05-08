"""ModCDPClient (Python): importable, no CLI, no demo code.

Constructor parameter names mirror the JS / Go ports:
    cdp_url           upstream CDP URL (str)
    extension_path    extension directory (str)
    routes            client-side routing dict
    server            { 'loopback_cdp_url'?, 'routes'? } passed to ModCDPServer.configure
    scan_for_existing_localhost_9222
                      when true and cdp_url is unset, attach to localhost:9222 before autolaunching
    mirror_upstream_events
                      when false, do not mirror server-side upstream CDP events back through Runtime bindings
    *_timeout_ms / *_interval_ms
                      override default CDP send, service-worker probe, event, and poll timings

Public methods: connect(), send(method, params), on(event, handler), close(), _cdp.send(), _cdp.on().
Synchronous (blocking) API; one background thread reads frames off the WS.
"""

import json
import asyncio
import inspect
import os
import re
import shutil
import subprocess
import sys
import threading
import time
import tempfile
import urllib.error
import urllib.parse
import urllib.request
import socket
import zipfile
from collections.abc import Mapping, Sequence
from pathlib import Path
from queue import Queue, Empty
from typing import Any, cast

from pydantic import TypeAdapter, ValidationError
from pydantic_core import to_jsonable_python
from websocket import create_connection

from .jsonschema import type_adapter_from_json_schema
from .cdp.surface import AwaitableDict, CDPEvent, CDPSurfaceMixin, cdp_event_name, install_cdp_surface
from .translate import (
    CUSTOM_EVENT_BINDING_NAME,
    DEFAULT_CLIENT_ROUTES,
    UPSTREAM_EVENT_BINDING_NAME,
    wrap_command_if_needed,
    unwrap_event_if_needed,
    unwrap_response_if_needed,
)
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
    BorrowedExtensionInfo,
    CdpFrame,
    ExtensionInfo,
    ExtensionProbe,
    FrameParams,
    Handler,
    JsonValue,
    LaunchOptions,
    PendingEntry,
    ProtocolParams,
    ProtocolPayload,
    ProtocolResult,
    TargetInfo,
    TranslatedCommand,
    WebSocketLike,
)

EXT_ID_FROM_URL_RE = re.compile(r"^chrome-extension://([a-z]+)/")
MODCDP_READY_EXPRESSION = (
    "Boolean(globalThis.ModCDP?.__ModCDPServerVersion === 1 && "
    "globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)"
)
DEFAULT_SERVER = object()
DEFAULT_LIVE_CDP_URL = "http://127.0.0.1:9222"
DEFAULT_CDP_SEND_TIMEOUT_MS = 10_000
DEFAULT_EVENT_WAIT_TIMEOUT_MS = 10_000
DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS = 10_000
DEFAULT_CHROME_READY_TIMEOUT_MS = 45_000
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
        return self._client._send_frame(method, params or {}, session_id=session_id, record_raw_timing=True)

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


def live_websocket_url_for(endpoint: str = DEFAULT_LIVE_CDP_URL) -> str | None:
    try:
        return websocket_url_for(endpoint)
    except Exception:
        return None


def _free_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]


def _json_object(value: JsonValue | None) -> ProtocolResult:
    return value if isinstance(value, dict) else {}


def default_extension_path() -> str | None:
    bundled_extension = Path(__file__).resolve().parent / "extension.zip"
    return str(bundled_extension) if bundled_extension.exists() else None


def modcdp_server_path(extension_path: str) -> Path:
    candidates = [Path(extension_path) / "ModCDPServer.js"]
    for parent in Path(__file__).resolve().parents:
        candidates.append(parent / "dist" / "extension" / "ModCDPServer.js")
    for candidate in candidates:
        if candidate.exists():
            return candidate
    checked = ", ".join(str(candidate) for candidate in candidates)
    raise FileNotFoundError(f"Unable to locate ModCDPServer.js; checked: {checked}")


def modcdp_server_bootstrap_expression(extension_path: str) -> str:
    server_path = modcdp_server_path(extension_path)
    source = server_path.read_text()
    start = source.index("export function installModCDPServer")
    end = source.index("export const ModCDPServer")
    installer = source[start:end].replace("export function", "function", 1)
    return (
        "(() => {\n"
        f"{installer}\n"
        "const ModCDP = installModCDPServer(globalThis);\n"
        "return {\n"
        "  ok: Boolean(ModCDP?.__ModCDPServerVersion === 1 && ModCDP?.handleCommand && ModCDP?.addCustomEvent),\n"
        "  extension_id: globalThis.chrome?.runtime?.id ?? null,\n"
        "  has_tabs: Boolean(globalThis.chrome?.tabs?.query),\n"
        "  has_debugger: Boolean(globalThis.chrome?.debugger?.sendCommand),\n"
        "};\n"
        "})()"
    )


class ModCDPClient(CDPSurfaceMixin):
    def __init__(
        self,
        cdp_url: str | None = None,
        extension_path: str | None = None,
        routes: Mapping[str, str] | None = None,
        server: Mapping[str, JsonValue] | None | object = DEFAULT_SERVER,
        custom_commands: Sequence[ModCDPAddCustomCommandParams] | None = None,
        custom_events: Sequence[ModCDPAddCustomEventParams] | None = None,
        custom_middlewares: Sequence[ModCDPAddMiddlewareParams] | None = None,
        service_worker_url_includes: Sequence[str] | None = None,
        service_worker_url_suffixes: Sequence[str] | None = None,
        trust_service_worker_target: bool = False,
        require_service_worker_target: bool = False,
        service_worker_ready_expression: str | None = None,
        mirror_upstream_events: bool = True,
        scan_for_existing_localhost_9222: bool = False,
        cdp_send_timeout_ms: int = DEFAULT_CDP_SEND_TIMEOUT_MS,
        event_wait_timeout_ms: int = DEFAULT_EVENT_WAIT_TIMEOUT_MS,
        execution_context_timeout_ms: int = DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS,
        service_worker_probe_timeout_ms: int = DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
        service_worker_ready_timeout_ms: int = DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS,
        service_worker_poll_interval_ms: int = DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS,
        target_session_poll_interval_ms: int = DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS,
        ws_connect_error_settle_timeout_ms: int = DEFAULT_WS_CONNECT_ERROR_SETTLE_TIMEOUT_MS,
        launch_options: LaunchOptions | None = None,
    ) -> None:
        self.cdp_url: str | None = cdp_url
        self.extension_path: str | None = extension_path or default_extension_path()
        self.routes: ModCDPRoutes = {**DEFAULT_CLIENT_ROUTES, **dict(routes or {})}
        if server is DEFAULT_SERVER:
            self.server: ModCDPServerConfig | None = {}
        elif server is None:
            self.server = None
        elif isinstance(server, Mapping):
            self.server = cast(ModCDPServerConfig, dict(server))
        else:
            raise TypeError("server must be a mapping, None, or omitted")
        self.custom_commands: list[ModCDPAddCustomCommandParams] = list(custom_commands or [])
        self.custom_events: list[ModCDPAddCustomEventParams] = list(custom_events or [])
        self.custom_middlewares: list[ModCDPAddMiddlewareParams] = list(custom_middlewares or [])
        self.service_worker_url_includes: list[str] = list(service_worker_url_includes or [])
        self.service_worker_url_suffixes: list[str] = list(service_worker_url_suffixes or ["/service_worker.js", "/background.js"])
        self.trust_service_worker_target = trust_service_worker_target
        self.require_service_worker_target = require_service_worker_target
        self.service_worker_ready_expression = service_worker_ready_expression
        self.mirror_upstream_events = mirror_upstream_events
        self.scan_for_existing_localhost_9222 = scan_for_existing_localhost_9222
        self.cdp_send_timeout_ms = cdp_send_timeout_ms
        self.event_wait_timeout_ms = event_wait_timeout_ms
        self.execution_context_timeout_ms = execution_context_timeout_ms
        self.service_worker_probe_timeout_ms = service_worker_probe_timeout_ms
        self.service_worker_ready_timeout_ms = service_worker_ready_timeout_ms
        self.service_worker_poll_interval_ms = service_worker_poll_interval_ms
        self.target_session_poll_interval_ms = target_session_poll_interval_ms
        self.ws_connect_error_settle_timeout_ms = ws_connect_error_settle_timeout_ms
        self.server_options: ModCDPServerConfig = {
            "cdp_send_timeout_ms": cdp_send_timeout_ms,
            "loopback_execution_context_timeout_ms": execution_context_timeout_ms,
            "ws_connect_error_settle_timeout_ms": ws_connect_error_settle_timeout_ms,
        }
        self.launch_options = cast(LaunchOptions, dict(launch_options or {}))

        self.extension_id: str | None = None
        self.ext_target_id: str | None = None
        self.ext_session_id: str | None = None
        self.latency: ModCDPPingLatency | None = None
        self.connect_timing: ModCDPConnectTiming | None = None
        self.last_command_timing: ModCDPCommandTiming | None = None
        self.last_raw_timing: ModCDPRawTiming | None = None

        self._ws: WebSocketLike | None = None
        self._next_id = 0
        self._pending: dict[int, PendingEntry] = {}
        self._handlers: dict[str, list[Handler]] = {}
        self._lock = threading.Lock()
        self._target_sessions: dict[str, str] = {}
        self._session_targets: dict[str, TargetInfo] = {}
        self._schema_lock = threading.RLock()
        self._command_params_schemas: dict[str, TypeAdapter[Any]] = {}
        self._command_result_schemas: dict[str, TypeAdapter[Any]] = {}
        self._event_schemas: dict[str, TypeAdapter[Any]] = {}
        self._event_classes: dict[str, type[CDPEvent]] = {}
        install_cdp_surface(self)
        self._reader_thread: threading.Thread | None = None
        self._closed = False
        self._launched_process: subprocess.Popen[bytes] | None = None
        self._profile_dir: tempfile.TemporaryDirectory[str] | None = None
        self._prepared_extension_dir: tempfile.TemporaryDirectory[str] | None = None
        self._cdp = _RawCDP(self)
        self._hydrate_custom_surface()

    def connect(self) -> "ModCDPClient":
        connect_started_at = int(time.time() * 1000)
        if self.cdp_url is None:
            self.cdp_url = live_websocket_url_for() if self.scan_for_existing_localhost_9222 else None
            if self.cdp_url is None:
                self._prepare_extension_path()
                launched = self._launch_chrome()
                self.cdp_url = launched["cdp_url"]
        input_cdp_url = self.cdp_url
        self.cdp_url = websocket_url_for(input_cdp_url)
        if self.server is not None and "loopback_cdp_url" not in self.server:
            self.server = {**self.server, "loopback_cdp_url": self.cdp_url}
        elif self.server and isinstance(self.server.get("loopback_cdp_url"), str):
            loopback_url = self.server["loopback_cdp_url"]
            if loopback_url == input_cdp_url or loopback_url == self.cdp_url:
                self.server = {**self.server, "loopback_cdp_url": self.cdp_url}
        self._ws = cast(WebSocketLike, create_connection(self.cdp_url, timeout=self.cdp_send_timeout_ms / 1000))
        self._reader_thread = threading.Thread(target=self._reader, daemon=True)
        self._reader_thread.start()

        self._send_frame("Target.setAutoAttach", {
            "autoAttach": True,
            "waitForDebuggerOnStart": False,
            "flatten": True,
        })
        self._send_frame("Target.setDiscoverTargets", {"discover": True})

        extension_started_at = int(time.time() * 1000)
        self._prepare_extension_path()
        ext = self._ensure_extension()
        extension_completed_at = int(time.time() * 1000)
        self.extension_id = ext["extension_id"]
        self.ext_target_id = ext["target_id"]
        self.ext_session_id = ext["session_id"]
        self._send_frame("Runtime.enable", {}, self.ext_session_id)
        self._send_frame("Runtime.addBinding", {"name": CUSTOM_EVENT_BINDING_NAME}, self.ext_session_id)
        if self.mirror_upstream_events:
            self._send_frame("Runtime.addBinding", {"name": UPSTREAM_EVENT_BINDING_NAME}, self.ext_session_id)

        if self.server is not None:
            custom_events: list[ModCDPAddCustomEventObjectParams] = []
            for event in self.custom_events:
                custom_events.append({"name": event} if isinstance(event, str) else event)
            custom_commands: list[ModCDPAddCustomCommandParams] = [
                command
                for command in self.custom_commands
                if isinstance(command.get("expression"), str) and command.get("expression")
            ]
            custom_middlewares: list[ModCDPAddMiddlewareParams] = list(self.custom_middlewares)
            configure_params: ModCDPServerConfig = {
                **self.server_options,
                **self.server,
                "custom_events": custom_events,
                "custom_commands": custom_commands,
                "custom_middlewares": custom_middlewares,
            }
            self._send_raw(wrap_command_if_needed(
                "Mod.configure",
                cast(ProtocolParams, configure_params),
                routes=self.routes,
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
    ) -> JsonValue:
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
        elif method == "Mod.addCustomEvent":
            self._register_custom_event(command_params)
        elif validate_custom_schema:
            command_params = self._validate_command_params(method, command_params)

        command = wrap_command_if_needed(
            method,
            command_params,
            routes=self.routes,
            cdp_session_id=self.ext_session_id,
            target_cdp_session_id=session_id,
        )
        result = self._send_raw(command)
        if validate_custom_schema and method != "Mod.addCustomCommand":
            result = self._validate_command_result(method, result)
        completed_at = int(time.time() * 1000)
        self.last_command_timing = {
            "method": method,
            "target": command["target"],
            "started_at": started_at,
            "completed_at": completed_at,
            "duration_ms": completed_at - started_at,
        }
        return AwaitableDict(result) if isinstance(result, dict) else result

    def raw_send(self, method: str, params: ProtocolParams | None = None) -> ProtocolResult:
        return self._send_frame(method, params or {}, record_raw_timing=True)

    def send_raw(
        self,
        method: str,
        params: Mapping[str, Any] | None = None,
        session_id: str | None = None,
    ) -> ProtocolResult:
        result = self._send_command(method, params, session_id=session_id)
        if not isinstance(result, dict):
            raise RuntimeError(f"{method} returned non-object value: {result!r}")
        return result

    def send(
        self,
        method: str,
        params: Mapping[str, Any] | None = None,
        session_id: str | None = None,
    ) -> AwaitableDict:
        result = self._send_command(method, params, session_id=session_id, validate_custom_schema=False)
        if isinstance(result, AwaitableDict):
            return result
        if isinstance(result, dict):
            return AwaitableDict(result)
        return AwaitableDict({"value": result})

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

    def close(self) -> None:
        if self._closed:
            return
        if self._ws is not None:
            try:
                with self._lock:
                    self._next_id += 1
                    msg_id = self._next_id
                self._ws.send(json.dumps({"id": msg_id, "method": "Browser.close", "params": {}}))
            except Exception:
                pass
        self._closed = True
        try:
            if self._ws:
                self._ws.close()
        except Exception:
            pass
        if self._reader_thread is not None and self._reader_thread.is_alive():
            self._reader_thread.join(timeout=1)
        self._ws = None
        if self._launched_process is not None:
            self._launched_process.terminate()
            try:
                self._launched_process.wait(timeout=2)
            except subprocess.TimeoutExpired:
                self._launched_process.kill()
                self._launched_process.wait(timeout=2)
            self._launched_process = None
        if self._profile_dir is not None:
            self._cleanup_temp_dir(self._profile_dir)
            self._profile_dir = None
        if self._prepared_extension_dir is not None:
            self._cleanup_temp_dir(self._prepared_extension_dir)
            self._prepared_extension_dir = None

    def _cleanup_temp_dir(self, temp_dir: tempfile.TemporaryDirectory[str]) -> None:
        for attempt in range(20):
            try:
                temp_dir.cleanup()
                return
            except OSError:
                if attempt == 19:
                    shutil.rmtree(temp_dir.name, ignore_errors=True)
                    return
                time.sleep(0.1)

    def _ready_expression(self) -> str:
        if not self.service_worker_ready_expression:
            return MODCDP_READY_EXPRESSION
        return f"({MODCDP_READY_EXPRESSION}) && Boolean({self.service_worker_ready_expression})"

    def _session_id_for_target(self, target_id: str, timeout: float = 0) -> str | None:
        if timeout <= 0:
            return self._target_sessions.get(target_id)
        deadline = time.time() + timeout
        while time.time() <= deadline:
            session_id = self._target_sessions.get(target_id)
            if session_id:
                return session_id
            time.sleep(self.target_session_poll_interval_ms / 1000)
        return None

    def _ensure_session_id_for_target(self, target_id: str, timeout: float = 0, allow_attach: bool = False) -> str | None:
        session_id = self._target_sessions.get(target_id)
        if session_id:
            return session_id
        if allow_attach:
            result = self._send_frame(
                "Target.attachToTarget",
                {"targetId": target_id, "flatten": True},
                timeout=max(timeout, self.cdp_send_timeout_ms / 1000),
            )
            attached_session_id = result.get("sessionId")
            if isinstance(attached_session_id, str) and attached_session_id:
                self._target_sessions[target_id] = attached_session_id
                return attached_session_id
        return self._session_id_for_target(target_id, timeout=timeout)

    def _launch_chrome(self) -> dict[str, str]:
        executable_path = self.launch_options.get("executable_path") or os.environ.get("CHROME_PATH")
        candidates = [
            executable_path,
            "/Applications/Chromium.app/Contents/MacOS/Chromium",
            "/Applications/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing",
            str(Path.home() / "Library/Caches/ms-playwright/chromium-1217/chrome-mac-arm64/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing"),
            "/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
            "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
            "/usr/bin/chromium",
            "/usr/bin/chromium-browser",
            "/usr/bin/google-chrome-canary",
            "/usr/bin/google-chrome-stable",
            "/usr/bin/google-chrome",
        ]
        executable_path = next((candidate for candidate in candidates if candidate and Path(candidate).exists()), None)
        if executable_path is None:
            raise RuntimeError("No Chrome/Chromium binary found. Set CHROME_PATH or pass launch_options.executable_path.")
        port = int(self.launch_options.get("port") or _free_port())
        self._profile_dir = tempfile.TemporaryDirectory(prefix="modcdp.")
        args = [
            "--enable-unsafe-extension-debugging",
            "--remote-allow-origins=*",
            "--no-first-run",
            "--no-default-browser-check",
            "--disable-default-apps",
            "--disable-dev-shm-usage",
            "--disable-background-networking",
            "--disable-backgrounding-occluded-windows",
            "--disable-renderer-backgrounding",
            "--disable-background-timer-throttling",
            "--disable-sync",
            "--disable-features=DisableLoadExtensionCommandLineSwitch",
            "--password-store=basic",
            "--use-mock-keychain",
            "--disable-gpu",
            f"--user-data-dir={self._profile_dir.name}",
            "--remote-debugging-address=127.0.0.1",
            f"--remote-debugging-port={port}",
        ]
        default_headless = sys.platform.startswith("linux") and not os.environ.get("DISPLAY")
        if self.launch_options.get("headless", default_headless):
            args.append("--headless=new")
        if self.launch_options.get("sandbox", False) is False:
            args.append("--no-sandbox")
        if self.extension_path:
            args.append(f"--load-extension={self.extension_path}")
        extra_args = self.launch_options.get("extra_args") or []
        args.extend(extra_args)
        args.append("about:blank")
        self._launched_process = subprocess.Popen([executable_path, *args], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        cdp_url = f"http://127.0.0.1:{port}"
        chrome_ready_timeout_s = DEFAULT_CHROME_READY_TIMEOUT_MS / 1000
        deadline = time.time() + chrome_ready_timeout_s
        while time.time() < deadline:
            try:
                with urllib.request.urlopen(f"{cdp_url}/json/version", timeout=0.5) as response:
                    json.loads(response.read())
                    return {"cdp_url": cdp_url}
            except Exception:
                time.sleep(0.1)
        self.close()
        raise RuntimeError(f"Chrome at {cdp_url} did not become ready within {chrome_ready_timeout_s}s")

    # --- internals ---------------------------------------------------------

    def _prepare_extension_path(self) -> None:
        if not self.extension_path or not self.extension_path.endswith(".zip"):
            return
        self._prepared_extension_dir = tempfile.TemporaryDirectory(prefix="modcdp-extension.")
        with zipfile.ZipFile(self.extension_path) as archive:
            archive.extractall(self._prepared_extension_dir.name)
        self.extension_path = self._prepared_extension_dir.name

    def _send_raw(self, wrapped: TranslatedCommand) -> JsonValue:
        if wrapped["target"] == "direct_cdp":
            step = wrapped["steps"][0]
            return self._send_frame(step["method"], step.get("params") or {})
        if wrapped["target"] != "service_worker":
            raise RuntimeError(f"Unsupported command target {wrapped['target']!r}")

        result: ProtocolResult = {}
        unwrap: str | None = None
        for step in wrapped["steps"]:
            result = self._send_frame(step["method"], step.get("params") or {}, self.ext_session_id)
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
        params_schema = self._adapter_from_optional_schema(params.get("paramsSchema"), "paramsSchema")
        result_schema = self._adapter_from_optional_schema(params.get("resultSchema"), "resultSchema")
        with self._schema_lock:
            if params_schema is not None:
                self._command_params_schemas[name] = params_schema
            if result_schema is not None:
                self._command_result_schemas[name] = result_schema

    def _register_custom_event(self, params: ProtocolParams) -> None:
        name = params.get("name")
        if not isinstance(name, str) or not name:
            raise TypeError("name must be a non-empty string")
        event_schema = self._adapter_from_optional_schema(params.get("eventSchema"), "eventSchema")
        if event_schema is not None:
            with self._schema_lock:
                self._event_schemas[name] = event_schema

    def _adapter_from_optional_schema(self, schema: JsonValue | None, field_name: str) -> TypeAdapter[Any] | None:
        if schema is None:
            return None
        if not isinstance(schema, Mapping):
            raise TypeError(f"{field_name} must be a JSON Schema object")
        return type_adapter_from_json_schema(cast(Mapping[str, Any], schema))

    def _validate_command_params(self, method: str, params: ProtocolParams) -> ProtocolParams:
        with self._schema_lock:
            adapter = self._command_params_schemas.get(method)
        if adapter is None:
            return params
        try:
            validated = adapter.validate_python(dict(params))
        except ValidationError as e:
            raise ValueError(f"{method} params did not match paramsSchema: {e}") from e
        jsonable = to_jsonable_python(validated)
        if not isinstance(jsonable, Mapping):
            raise ValueError(f"{method} paramsSchema must validate to a JSON object")
        return cast(ProtocolParams, dict(jsonable))

    def _validate_command_result(self, method: str, result: JsonValue) -> JsonValue:
        with self._schema_lock:
            adapter = self._command_result_schemas.get(method)
        if adapter is None:
            return result
        try:
            return cast(JsonValue, to_jsonable_python(adapter.validate_python(result)))
        except ValidationError as e:
            raise ValueError(f"{method} result did not match resultSchema: {e}") from e

    def _validate_event_payload(self, event: str, payload: ProtocolPayload) -> ProtocolPayload | None:
        with self._schema_lock:
            adapter = self._event_schemas.get(event)
        if adapter is None:
            return dict(payload)
        try:
            validated = adapter.validate_python(dict(payload))
        except ValidationError as direct_error:
            if set(payload.keys()) != {"value"}:
                print(f"[ModCDPClient] event {event} did not match eventSchema: {direct_error}", file=sys.stderr)
                return None
            try:
                validated = adapter.validate_python(payload["value"])
            except ValidationError as value_error:
                print(f"[ModCDPClient] event {event} did not match eventSchema: {value_error}", file=sys.stderr)
                return None
            return {"value": cast(JsonValue, to_jsonable_python(validated))}
        jsonable = to_jsonable_python(validated)
        if isinstance(jsonable, Mapping):
            return cast(ProtocolPayload, dict(jsonable))
        return {"value": cast(JsonValue, jsonable)}

    def _measure_ping_latency(self) -> ModCDPPingLatency:
        sent_at = int(time.time() * 1000)
        done: Queue[ProtocolPayload] = Queue()

        def on_pong(payload: ProtocolPayload) -> None:
            done.put(payload or {})

        self._handlers.setdefault("Mod.pong", []).append(on_pong)
        try:
            self.send("Mod.ping", {"sentAt": sent_at})
            payload = done.get(timeout=10)
        except Empty:
            raise RuntimeError("Mod.pong timed out")
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

    def _send_frame(
        self,
        method: str,
        params: FrameParams | None = None,
        session_id: str | None = None,
        timeout: float | None = None,
        record_raw_timing: bool = False,
    ) -> ProtocolResult:
        effective_timeout = self.cdp_send_timeout_ms / 1000 if timeout is None else timeout
        with self._lock:
            self._next_id += 1
            msg_id = self._next_id
            done: Queue[CdpFrame] = Queue()
            self._pending[msg_id] = (method, done)
        started_at = int(time.time() * 1000)
        msg: CdpFrame = {"id": msg_id, "method": method, "params": params or {}}
        if session_id:
            msg["sessionId"] = session_id
        ws = self._ws
        if ws is None:
            with self._lock:
                self._pending.pop(msg_id, None)
            raise RuntimeError("CDP websocket is not connected")
        try:
            ws.send(json.dumps(msg))
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
                msg = cast(CdpFrame, parsed)
                if "id" in msg and msg["id"] is not None:
                    with self._lock:
                        entry = self._pending.pop(msg["id"], None)
                    if entry:
                        entry[1].put(msg)
                    continue
                method = msg.get("method")
                raw_params = msg.get("params")
                params = cast(ProtocolParams, raw_params) if isinstance(raw_params, Mapping) else {}
                if method == "Target.attachedToTarget":
                    session_id = params.get("sessionId")
                    raw_target_info = params.get("targetInfo")
                    target_info = cast(TargetInfo, raw_target_info) if isinstance(raw_target_info, dict) else None
                    target_id = target_info.get("targetId") if target_info else None
                    if isinstance(session_id, str) and isinstance(target_id, str) and target_info:
                        self._target_sessions[target_id] = session_id
                        self._session_targets[session_id] = target_info
                elif method == "Target.detachedFromTarget":
                    session_id = params.get("sessionId")
                    target_info = self._session_targets.pop(session_id, None) if isinstance(session_id, str) else None
                    target_id = target_info.get("targetId") if target_info else None
                    if target_id:
                        self._target_sessions.pop(target_id, None)
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
                            continue
                        for h in self._handlers.get(u["event"], []):
                            def run_wrapped_event(handler=h, payload=validated_payload, event_name=u["event"]):
                                self._run_handler(handler, payload, event_name)
                            threading.Thread(target=run_wrapped_event, daemon=True).start()
                    continue
                if method:
                    raw_session_id = msg.get("sessionId")
                    event_session_id = raw_session_id if isinstance(raw_session_id, str) else None
                    validated_payload = self._validate_event_payload(method, dict(params))
                    if validated_payload is None:
                        continue
                    for h in self._handlers.get(method, []):
                        def run_method_event(handler=h, payload=validated_payload, event_name=method):
                            self._run_handler(handler, payload, event_name)
                        threading.Thread(target=run_method_event, daemon=True).start()
        except Exception as e:
            if not self._closed:
                print(f"[ModCDPClient] reader exited: {e}")
        finally:
            with self._lock:
                pending = list(self._pending.values())
                self._pending.clear()
            for _, done in pending:
                done.put({"error": {"message": "connection closed"}})

    def _target_infos(self) -> list[TargetInfo]:
        value = self._send_frame("Target.getTargets").get("targetInfos")
        if not isinstance(value, list):
            return []
        targets: list[TargetInfo] = []
        for item in value:
            if not isinstance(item, dict):
                continue
            target_id = item.get("targetId")
            target_type = item.get("type")
            target_url = item.get("url")
            if isinstance(target_id, str) and isinstance(target_type, str) and isinstance(target_url, str):
                targets.append({"targetId": target_id, "type": target_type, "url": target_url})
        return targets

    def _ensure_extension(self) -> ExtensionInfo:
        ready_expression = self._ready_expression()
        trust_service_worker_target = (
            self.trust_service_worker_target
            or len(self.service_worker_url_includes) > 0
            or any(len([part for part in suffix.split("/") if part]) > 1 for suffix in self.service_worker_url_suffixes)
        )

        def probe_target(target: TargetInfo, timeout: float = 0, allow_attach: bool = False) -> ExtensionProbe | None:
            target_id = target.get("targetId")
            target_url = target.get("url")
            if not target_id or not target_url:
                return None
            session_id = self._ensure_session_id_for_target(target_id, timeout=timeout, allow_attach=allow_attach)
            if not session_id:
                return None
            self._send_frame("Runtime.enable", {}, session_id, timeout=self.cdp_send_timeout_ms / 1000)
            probe = self._send_frame("Runtime.evaluate", {
                "expression": ready_expression,
                "returnByValue": True,
            }, session_id, timeout=self.cdp_send_timeout_ms / 1000)
            if _json_object(probe.get("result")).get("value") is not True:
                return None
            match = EXT_ID_FROM_URL_RE.match(target_url)
            if match is None:
                return None
            return {
                "extension_id": match.group(1),
                "target_id": target_id,
                "url": target_url,
                "session_id": session_id,
            }

        def discover_ready_service_worker(matched_only: bool = False) -> ExtensionInfo | None:
            target_infos = self._target_infos()
            if trust_service_worker_target:
                for t in target_infos:
                    if self._service_worker_target_matches(t):
                        try:
                            result = probe_target(t, timeout=self.service_worker_probe_timeout_ms / 1000, allow_attach=True)
                        except Exception:
                            result = None
                        if result:
                            return {"source": "trusted", **result}
            if trust_service_worker_target or matched_only:
                return None
            for t in target_infos:
                if t["type"] != "service_worker": continue
                if not t["url"].startswith("chrome-extension://"): continue
                try:
                    result = probe_target(t, timeout=self.service_worker_probe_timeout_ms / 1000)
                except Exception:
                    continue
                if result:
                    return {"source": "discovered", **result}
            return None

        def wait_for_ready_service_worker(timeout: float, matched_only: bool = False) -> ExtensionInfo | None:
            deadline = time.monotonic() + timeout
            while time.monotonic() < deadline:
                result = discover_ready_service_worker(matched_only=matched_only)
                if result:
                    return result
                time.sleep(self.service_worker_poll_interval_ms / 1000)
            return None

        # 1. Discover an existing ModCDP service worker. Browserbase loads the
        # extension for the session, but the service-worker target can appear a
        # moment after the browser CDP websocket accepts connections.
        discovered = discover_ready_service_worker()
        if discovered:
            return discovered
        if self.require_service_worker_target:
            discovered = wait_for_ready_service_worker(
                self.service_worker_probe_timeout_ms / 1000,
                matched_only=trust_service_worker_target,
            )
            if discovered:
                return discovered
            raise RuntimeError(
                "Required ModCDP service worker target did not become ready "
                f"({', '.join([*self.service_worker_url_includes, *self.service_worker_url_suffixes]) or 'no matcher'})."
            )
        if self.extension_path is None:
            raise RuntimeError("extension_path is required when no existing ModCDP service worker can be discovered.")

        # 2. Try Extensions.loadUnpacked.
        try:
            r = self._send_frame("Extensions.loadUnpacked", {"path": self.extension_path})
            extension_id = r.get("id") or r.get("extensionId")
        except RuntimeError as e:
            if "Method not available" in str(e) or "wasn't found" in str(e):
                discovered = wait_for_ready_service_worker(
                    self.service_worker_probe_timeout_ms / 1000,
                    matched_only=trust_service_worker_target,
                )
                if discovered:
                    return discovered
                return self._borrow_extension_worker(str(e))
            raise
        if not isinstance(extension_id, str) or not extension_id:
            raise RuntimeError(f"Extensions.loadUnpacked returned no id: {r}")

        # 3. Wait for the loaded extension's SW.
        sw_url_prefix = f"chrome-extension://{extension_id}/"
        deadline = time.monotonic() + self.service_worker_ready_timeout_ms / 1000
        while time.monotonic() < deadline:
            for t in self._target_infos():
                target_url = t.get("url") or ""
                if t.get("type") == "service_worker" and target_url.startswith(sw_url_prefix):
                    result = probe_target(t, timeout=self.service_worker_probe_timeout_ms / 1000, allow_attach=True)
                    if result:
                        return {
                            "source": "injected", "extension_id": extension_id,
                            "target_id": t["targetId"], "url": target_url, "session_id": result["session_id"],
                        }
            time.sleep(self.service_worker_poll_interval_ms / 1000)
        raise RuntimeError(
            f"Timed out after {self.service_worker_ready_timeout_ms}ms waiting for service worker target for extension {extension_id}."
        )

    def _service_worker_target_matches(self, target: TargetInfo) -> bool:
        url = target.get("url") or ""
        if target.get("type") != "service_worker" or not url.startswith("chrome-extension://"):
            return False
        if self.service_worker_url_includes and not all(part in url for part in self.service_worker_url_includes):
            return False
        if self.service_worker_url_suffixes and not any(url.endswith(suffix) for suffix in self.service_worker_url_suffixes):
            return False
        return bool(self.service_worker_url_includes or self.service_worker_url_suffixes)

    def _borrow_extension_worker(self, load_error: str) -> ExtensionInfo:
        if self.extension_path is None:
            raise RuntimeError("extension_path is required to bootstrap a borrowed extension worker.")
        borrowed: list[BorrowedExtensionInfo] = []
        bootstrap = modcdp_server_bootstrap_expression(self.extension_path)
        for t in self._target_infos():
            target_id = t.get("targetId")
            target_url = t.get("url") or ""
            if t.get("type") != "service_worker": continue
            if not target_id or not target_url.startswith("chrome-extension://"): continue
            session_id = None
            try:
                session_id = self._session_id_for_target(target_id, timeout=2)
                if not session_id:
                    continue
                try: self._send_frame("Runtime.enable", {}, session_id, timeout=2)
                except Exception: pass
                result = self._send_frame("Runtime.evaluate", {
                    "expression": bootstrap,
                    "awaitPromise": True,
                    "returnByValue": True,
                    "allowUnsafeEvalBlockedByCSP": True,
                }, session_id, timeout=3)
                result_object = result.get("result")
                value = result_object.get("value") if isinstance(result_object, dict) else {}
                if not isinstance(value, dict):
                    value = {}
                ready = bool(value.get("ok"))
                if ready and self.service_worker_ready_expression:
                    probe = self._send_frame("Runtime.evaluate", {
                        "expression": self._ready_expression(),
                        "returnByValue": True,
                    }, session_id, timeout=2)
                    ready = _json_object(probe.get("result")).get("value") is True
                if ready:
                    m = EXT_ID_FROM_URL_RE.match(target_url)
                    extension_id = value.get("extension_id") or (m.group(1) if m else None)
                    if not isinstance(extension_id, str):
                        continue
                    borrowed.append({
                        "source": "borrowed",
                        "extension_id": extension_id,
                        "target_id": target_id,
                        "url": target_url,
                        "session_id": session_id,
                        "has_tabs": bool(value.get("has_tabs")),
                        "has_debugger": bool(value.get("has_debugger")),
                    })
            except Exception:
                pass
        borrowed.sort(key=lambda item: (item.get("has_debugger", False), item.get("has_tabs", False)), reverse=True)
        if borrowed:
            selected = borrowed[0]
            selected.pop("has_tabs", None)
            selected.pop("has_debugger", None)
            return selected
        raise RuntimeError(
            "Cannot install or borrow ModCDP in the running browser.\n"
            "  - No existing service worker with globalThis.ModCDP was found.\n"
            f"  - Extensions.loadUnpacked is unavailable ({load_error}).\n"
            "  - No running chrome-extension:// service worker accepted the ModCDP bootstrap."
        )
