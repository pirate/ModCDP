from __future__ import annotations

import re
import threading
import time
from collections.abc import Callable, Mapping
from queue import Empty, Queue
from typing import Any, TypedDict, cast

from typing_extensions import NotRequired

from ..launcher.BrowserLauncher import BrowserLaunchOptions
from ..types.modcdp import ProtocolParams, ProtocolResult, TargetInfo

EXT_ID_FROM_URL_RE = re.compile(r"^chrome-extension://([a-z]+)/")
DEFAULT_MODCDP_EXTENSION_ID = "mdedooklbnfejodmnhmkdpkaedafkehf"
DEFAULT_MODCDP_SERVICE_WORKER_URL_SUFFIXES = ["/modcdp/service_worker.js"]
DEFAULT_MODCDP_WAKE_PATH = "/modcdp/wake.html"
MODCDP_READY_EXPRESSION = (
    "Boolean(globalThis.ModCDP?.__ModCDPServerVersion === 1 && globalThis.ModCDP?.handleCommand && globalThis.ModCDP?.addCustomEvent)"
)
DEFAULT_CDP_SEND_TIMEOUT_MS = 10_000
DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS = 10_000
DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS = 10_000
DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS = 60_000
DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS = 100
DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS = 20

SendCDP = Callable[[str, ProtocolParams | None, str | None], ProtocolResult]
SessionIdForTarget = Callable[[str], str | None]
AttachToTarget = Callable[[str], str | None]
WaitForExecutionContext = Callable[[str, int], int]


class ExtensionInjectorConfig(TypedDict, total=False):
    send: SendCDP | None
    sessionIdForTarget: SessionIdForTarget | None
    attachToTarget: AttachToTarget | None
    waitForExecutionContext: WaitForExecutionContext | None
    injector_extension_path: str | None
    injector_extension_id: str | None
    injector_wake_path: str | None
    injector_wake_url: str | None
    injector_service_worker_url_includes: list[str]
    injector_service_worker_url_suffixes: list[str]
    injector_trust_service_worker_target: bool
    injector_require_service_worker_target: bool
    injector_service_worker_ready_expression: str | None
    injector_cdp_send_timeout_ms: int
    injector_execution_context_timeout_ms: int
    injector_service_worker_probe_timeout_ms: int
    injector_service_worker_ready_timeout_ms: int
    injector_service_worker_poll_interval_ms: int
    injector_target_session_poll_interval_ms: int
    injector_browserbase_api_key: str | None
    injector_browserbase_base_url: str | None
    upstream_reversews_url: str | None
    upstream_nativemessaging_host_name: str | None
    upstream_nats_url: str | None
    upstream_nats_subject_prefix: str | None


class ExtensionInjectionResult(TypedDict):
    source: str
    extension_id: str | None
    target_id: str
    url: str
    session_id: str
    has_tabs: NotRequired[bool]
    has_debugger: NotRequired[bool]


class ExtensionInjector:
    def __init__(self, options: ExtensionInjectorConfig | None = None) -> None:
        self.options = cast(ExtensionInjectorConfig, {
            "send": None,
            "sessionIdForTarget": None,
            "attachToTarget": None,
            "waitForExecutionContext": None,
            "injector_extension_path": None,
            "injector_extension_id": None,
            "injector_wake_path": DEFAULT_MODCDP_WAKE_PATH,
            "injector_wake_url": None,
            "injector_service_worker_url_includes": [],
            "injector_service_worker_url_suffixes": [],
            "injector_trust_service_worker_target": False,
            "injector_require_service_worker_target": False,
            "injector_service_worker_ready_expression": None,
            "injector_cdp_send_timeout_ms": DEFAULT_CDP_SEND_TIMEOUT_MS,
            "injector_execution_context_timeout_ms": DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS,
            "injector_service_worker_probe_timeout_ms": DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
            "injector_service_worker_ready_timeout_ms": DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS,
            "injector_service_worker_poll_interval_ms": DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS,
            "injector_target_session_poll_interval_ms": DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS,
            "injector_browserbase_api_key": None,
            "injector_browserbase_base_url": None,
            "upstream_reversews_url": None,
            "upstream_nativemessaging_host_name": None,
            "upstream_nats_url": None,
            "upstream_nats_subject_prefix": None,
            **dict(options or {}),
        })
        self.unusable_target_ids: set[str] = set()
        self.last_error: Exception | None = None

    def update(self, config: ExtensionInjectorConfig | None = None) -> "ExtensionInjector":
        config = cast(ExtensionInjectorConfig, dict(config or {}))
        self.options = cast(
            ExtensionInjectorConfig,
            {
                **self.options,
                **config,
                "injector_service_worker_url_includes": config.get(
                    "injector_service_worker_url_includes",
                    self.options.get("injector_service_worker_url_includes") or [],
                ),
                "injector_service_worker_url_suffixes": config.get(
                    "injector_service_worker_url_suffixes",
                    self.options.get("injector_service_worker_url_suffixes") or [],
                ),
            },
        )
        return self

    def getInjectorConfig(self) -> ExtensionInjectorConfig:
        return cast(ExtensionInjectorConfig, dict(self.options))

    def getLauncherConfig(self) -> BrowserLaunchOptions:
        return {}

    def getTransportConfig(self) -> dict[str, Any]:
        extension_id = self.options.get("injector_extension_id")
        return {"injector_extension_id": extension_id} if extension_id else {}

    def prepare(self) -> None:
        return None

    def close(self) -> None:
        return None

    def inject(self) -> ExtensionInjectionResult | None:
        raise NotImplementedError(f"{type(self).__name__}.inject is not implemented.")

    def _readyExpression(self) -> str:
        expression = self.options.get("injector_service_worker_ready_expression")
        if not expression:
            return MODCDP_READY_EXPRESSION
        return f"({MODCDP_READY_EXPRESSION}) && Boolean({expression})"

    def _sendWithTimeout(
        self,
        method: str,
        params: ProtocolParams | None = None,
        session_id: str | None = None,
        timeout_ms: int | None = None,
    ) -> ProtocolResult:
        send = self.options.get("send")
        if send is None:
            raise RuntimeError(f"{type(self).__name__} requires a CDP send function.")
        effective_timeout_ms = timeout_ms or self.options.get("injector_cdp_send_timeout_ms") or DEFAULT_CDP_SEND_TIMEOUT_MS
        if effective_timeout_ms <= 0:
            return send(method, params or {}, session_id)

        result_queue: Queue[tuple[ProtocolResult | None, BaseException | None]] = Queue(maxsize=1)

        def runSend() -> None:
            try:
                result_queue.put((send(method, params or {}, session_id), None))
            except BaseException as error:
                result_queue.put((None, error))

        threading.Thread(target=runSend, daemon=True).start()
        try:
            result, error = result_queue.get(timeout=effective_timeout_ms / 1000)
        except Empty as error:
            raise TimeoutError(f"{method} timed out after {effective_timeout_ms}ms") from error
        if error is not None:
            raise error
        return result or {}

    def _sessionIdForTarget(self, target_id: str, timeout_ms: int = 0) -> str | None:
        deadline = time.monotonic() + timeout_ms / 1000
        while True:
            session_id = self.options.get("sessionIdForTarget")
            if session_id is not None:
                value = session_id(target_id)
                if value:
                    return value
            if time.monotonic() >= deadline:
                return None
            time.sleep((self.options.get("injector_target_session_poll_interval_ms") or DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS) / 1000)

    def _ensureSessionIdForTarget(self, target_id: str, timeout_ms: int = 0, allow_attach: bool = False) -> str | None:
        session_id = self.options.get("sessionIdForTarget")
        if session_id is not None:
            value = session_id(target_id)
            if value:
                return value
        if allow_attach:
            attach_to_target = self.options.get("attachToTarget")
            if attach_to_target is not None:
                attached_session_id = attach_to_target(target_id)
                if attached_session_id:
                    return attached_session_id
        return self._sessionIdForTarget(target_id, timeout_ms)

    def _targetInfos(self) -> list[TargetInfo]:
        result = self._sendWithTimeout("Target.getTargets")
        raw_targets = result.get("targetInfos")
        if not isinstance(raw_targets, list):
            return []
        targets: list[TargetInfo] = []
        for raw_target in raw_targets:
            if not isinstance(raw_target, Mapping):
                continue
            target_id = raw_target.get("targetId")
            target_type = raw_target.get("type")
            target_url = raw_target.get("url")
            if isinstance(target_id, str) and isinstance(target_type, str) and isinstance(target_url, str):
                targets.append({"targetId": target_id, "type": target_type, "url": target_url})
        return targets

    def _configuredWakeUrl(self) -> str | None:
        wake_url = self.options.get("injector_wake_url")
        if wake_url:
            return wake_url
        extension_id = self.options.get("injector_extension_id")
        if not extension_id:
            return None
        wake_path = self.options.get("injector_wake_path") or DEFAULT_MODCDP_WAKE_PATH
        return f"chrome-extension://{extension_id}{wake_path if wake_path.startswith('/') else f'/{wake_path}'}"

    def _wakeConfiguredExtension(self) -> bool:
        wake_url = self._configuredWakeUrl()
        if not wake_url or self.options.get("send") is None:
            return False
        try:
            self._sendWithTimeout(
                "Target.createTarget",
                {
                    "url": wake_url,
                    "background": True,
                    "hidden": True,
                    "focus": False,
                },
            )
            return True
        except Exception:
            return False

    def _probeTarget(
        self,
        target: TargetInfo,
        session_timeout_ms: int = 0,
        *,
        allow_attach: bool = False,
    ) -> ExtensionInjectionResult | None:
        target_id = target["targetId"]
        if target_id in self.unusable_target_ids:
            return None
        session_id = self._ensureSessionIdForTarget(target_id, session_timeout_ms, allow_attach)
        if session_id is None:
            return None
        self._sendWithTimeout("Runtime.enable", {}, session_id)
        probe = self._sendWithTimeout(
            "Runtime.evaluate",
            {
                "expression": self._readyExpression(),
                "returnByValue": True,
            },
            session_id,
        )
        result = cast(Mapping[str, Any], probe.get("result")) if isinstance(probe.get("result"), Mapping) else {}
        value = result.get("value")
        if value is not True:
            return None
        match = EXT_ID_FROM_URL_RE.match(target.get("url") or "")
        return {
            "source": "discovered",
            "extension_id": match.group(1) if match else None,
            "target_id": target_id,
            "url": target["url"],
            "session_id": session_id,
        }

    def _discoverReadyServiceWorker(self, *, matched_only: bool = False) -> ExtensionInjectionResult | None:
        target_infos = self._targetInfos()
        if self.options.get("injector_trust_service_worker_target"):
            for candidate in target_infos:
                if not self._serviceWorkerTargetMatches(candidate):
                    continue
                probed = self._probeTarget(
                    candidate,
                    self.options.get("injector_service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
                    allow_attach=True,
                )
                if probed:
                    return {**probed, "source": "trusted"}
        if self.options.get("injector_trust_service_worker_target") or matched_only:
            return None
        for candidate in target_infos:
            if candidate["type"] != "service_worker":
                continue
            if not candidate["url"].startswith("chrome-extension://"):
                continue
            try:
                probed = self._probeTarget(
                    candidate,
                    self.options.get("injector_service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
                )
            except Exception:
                continue
            if probed:
                return probed
        return None

    def _waitForReadyServiceWorker(self, timeout_ms: int, *, matched_only: bool = False) -> ExtensionInjectionResult | None:
        deadline = time.monotonic() + timeout_ms / 1000
        while time.monotonic() < deadline:
            discovered = self._discoverReadyServiceWorker(matched_only=matched_only)
            if discovered:
                return discovered
            time.sleep((self.options.get("injector_service_worker_poll_interval_ms") or DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS) / 1000)
        return None

    def _serviceWorkerTargetMatches(self, candidate: Mapping[str, object]) -> bool:
        raw_target_url = candidate.get("url")
        target_url = raw_target_url if isinstance(raw_target_url, str) else ""
        if candidate.get("type") != "service_worker":
            return False
        if not target_url.startswith("chrome-extension://"):
            return False
        extension_id = self.options.get("injector_extension_id")
        has_extension_id = bool(extension_id)
        if extension_id and not target_url.startswith(f"chrome-extension://{extension_id}/"):
            return False
        includes = self.options.get("injector_service_worker_url_includes") or []
        suffixes = self.options.get("injector_service_worker_url_suffixes") or []
        if includes and not all(part in target_url for part in includes):
            return False
        if suffixes and not any(target_url.endswith(suffix) for suffix in suffixes):
            return False
        return bool(has_extension_id or includes or suffixes)
