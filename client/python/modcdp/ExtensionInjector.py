from __future__ import annotations

import json
import re
import threading
import time
from collections.abc import Callable, Mapping
from pathlib import Path
from queue import Empty, Queue
from typing import Any, TypedDict, cast

from typing_extensions import NotRequired

from .BrowserLauncher import BrowserLaunchOptions
from .types import ProtocolParams, ProtocolResult, TargetInfo

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
    extension_path: str | None
    extension_id: str | None
    wake_path: str | None
    wake_url: str | None
    service_worker_url_includes: list[str]
    service_worker_url_suffixes: list[str]
    trust_matched_service_worker: bool
    require_service_worker_target: bool
    service_worker_ready_expression: str | None
    cdp_send_timeout_ms: int
    execution_context_timeout_ms: int
    service_worker_probe_timeout_ms: int
    service_worker_ready_timeout_ms: int
    service_worker_poll_interval_ms: int
    target_session_poll_interval_ms: int
    browserbase_api_key: str | None
    base_url: str | None
    browserbase_base_url: str | None
    reverse_proxy_url: str | None
    native_host_name: str | None
    nats_url: str | None
    nats_subject_prefix: str | None


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
            "extension_path": None,
            "extension_id": None,
            "wake_path": DEFAULT_MODCDP_WAKE_PATH,
            "wake_url": None,
            "service_worker_url_includes": [],
            "service_worker_url_suffixes": [],
            "trust_matched_service_worker": False,
            "require_service_worker_target": False,
            "service_worker_ready_expression": None,
            "cdp_send_timeout_ms": DEFAULT_CDP_SEND_TIMEOUT_MS,
            "execution_context_timeout_ms": DEFAULT_EXECUTION_CONTEXT_TIMEOUT_MS,
            "service_worker_probe_timeout_ms": DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
            "service_worker_ready_timeout_ms": DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS,
            "service_worker_poll_interval_ms": DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS,
            "target_session_poll_interval_ms": DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS,
            "browserbase_api_key": None,
            "base_url": None,
            "browserbase_base_url": None,
            "reverse_proxy_url": None,
            "native_host_name": None,
            "nats_url": None,
            "nats_subject_prefix": None,
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
                "service_worker_url_includes": config.get(
                    "service_worker_url_includes",
                    self.options.get("service_worker_url_includes") or [],
                ),
                "service_worker_url_suffixes": config.get(
                    "service_worker_url_suffixes",
                    self.options.get("service_worker_url_suffixes") or [],
                ),
            },
        )
        return self

    def getInjectorConfig(self) -> ExtensionInjectorConfig:
        return cast(ExtensionInjectorConfig, dict(self.options))

    def getLauncherConfig(self) -> BrowserLaunchOptions:
        return {}

    def getTransportConfig(self) -> dict[str, Any]:
        extension_id = self.options.get("extension_id")
        return {"extension_id": extension_id} if extension_id else {}

    def prepare(self) -> None:
        return None

    def close(self) -> None:
        return None

    def inject(self) -> ExtensionInjectionResult | None:
        raise NotImplementedError(f"{type(self).__name__}.inject is not implemented.")

    def extensionRuntimeConfig(self) -> dict[str, str] | None:
        config = {
            key: value
            for key, value in {
                "reverse_proxy_url": self.options.get("reverse_proxy_url"),
                "native_host_name": self.options.get("native_host_name"),
                "nats_url": self.options.get("nats_url"),
                "nats_subject_prefix": self.options.get("nats_subject_prefix"),
            }.items()
            if isinstance(value, str) and value
        }
        return config or None

    def writeExtensionRuntimeConfig(self, unpacked_extension_path: str) -> None:
        config = self.extensionRuntimeConfig()
        if not config:
            return
        extension_path = Path(unpacked_extension_path)
        (extension_path / "modcdp").mkdir(parents=True, exist_ok=True)
        (extension_path / "modcdp" / "config.json").write_text(json.dumps(config, indent=2) + "\n")
        (extension_path / "config.js").write_text(
            f"globalThis.__MODCDP_RUNTIME_CONFIG__ = {json.dumps(config, indent=2)};\nexport {{}};\n"
        )

    def readyExpression(self) -> str:
        expression = self.options.get("service_worker_ready_expression")
        if not expression:
            return MODCDP_READY_EXPRESSION
        return f"({MODCDP_READY_EXPRESSION}) && Boolean({expression})"

    def sendWithTimeout(
        self,
        method: str,
        params: ProtocolParams | None = None,
        session_id: str | None = None,
        timeout_ms: int | None = None,
    ) -> ProtocolResult:
        send = self.options.get("send")
        if send is None:
            raise RuntimeError(f"{type(self).__name__} requires a CDP send function.")
        effective_timeout_ms = timeout_ms or self.options.get("cdp_send_timeout_ms") or DEFAULT_CDP_SEND_TIMEOUT_MS
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

    def sessionIdForTarget(self, target_id: str, timeout_ms: int = 0) -> str | None:
        deadline = time.monotonic() + timeout_ms / 1000
        while True:
            session_id = self.options.get("sessionIdForTarget")
            if session_id is not None:
                value = session_id(target_id)
                if value:
                    return value
            if time.monotonic() >= deadline:
                return None
            time.sleep((self.options.get("target_session_poll_interval_ms") or DEFAULT_TARGET_SESSION_POLL_INTERVAL_MS) / 1000)

    def ensureSessionIdForTarget(self, target_id: str, timeout_ms: int = 0, allow_attach: bool = False) -> str | None:
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
        return self.sessionIdForTarget(target_id, timeout_ms)

    def targetInfos(self) -> list[TargetInfo]:
        result = self.sendWithTimeout("Target.getTargets")
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

    def configuredWakeUrl(self) -> str | None:
        wake_url = self.options.get("wake_url")
        if wake_url:
            return wake_url
        extension_id = self.options.get("extension_id")
        if not extension_id:
            return None
        wake_path = self.options.get("wake_path") or DEFAULT_MODCDP_WAKE_PATH
        return f"chrome-extension://{extension_id}{wake_path if wake_path.startswith('/') else f'/{wake_path}'}"

    def wakeConfiguredExtension(self) -> bool:
        wake_url = self.configuredWakeUrl()
        if not wake_url or self.options.get("send") is None:
            return False
        try:
            self.sendWithTimeout(
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

    def probeTarget(
        self,
        target: TargetInfo,
        session_timeout_ms: int = 0,
        *,
        allow_attach: bool = False,
    ) -> ExtensionInjectionResult | None:
        target_id = target["targetId"]
        if target_id in self.unusable_target_ids:
            return None
        session_id = self.ensureSessionIdForTarget(target_id, session_timeout_ms, allow_attach)
        if session_id is None:
            return None
        self.sendWithTimeout("Runtime.enable", {}, session_id)
        probe = self.sendWithTimeout(
            "Runtime.evaluate",
            {
                "expression": self.readyExpression(),
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

    def discoverReadyServiceWorker(self, *, matched_only: bool = False) -> ExtensionInjectionResult | None:
        target_infos = self.targetInfos()
        if self.options.get("trust_matched_service_worker"):
            for candidate in target_infos:
                if not self.serviceWorkerTargetMatches(candidate):
                    continue
                probed = self.probeTarget(
                    candidate,
                    self.options.get("service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
                    allow_attach=True,
                )
                if probed:
                    return {**probed, "source": "trusted"}
        if self.options.get("trust_matched_service_worker") or matched_only:
            return None
        for candidate in target_infos:
            if candidate["type"] != "service_worker":
                continue
            if not candidate["url"].startswith("chrome-extension://"):
                continue
            try:
                probed = self.probeTarget(
                    candidate,
                    self.options.get("service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
                )
            except Exception:
                continue
            if probed:
                return probed
        return None

    def waitForReadyServiceWorker(self, timeout_ms: int, *, matched_only: bool = False) -> ExtensionInjectionResult | None:
        deadline = time.monotonic() + timeout_ms / 1000
        while time.monotonic() < deadline:
            discovered = self.discoverReadyServiceWorker(matched_only=matched_only)
            if discovered:
                return discovered
            time.sleep((self.options.get("service_worker_poll_interval_ms") or DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS) / 1000)
        return None

    def serviceWorkerTargetMatches(self, candidate: Mapping[str, object]) -> bool:
        raw_target_url = candidate.get("url")
        target_url = raw_target_url if isinstance(raw_target_url, str) else ""
        if candidate.get("type") != "service_worker":
            return False
        if not target_url.startswith("chrome-extension://"):
            return False
        extension_id = self.options.get("extension_id")
        has_extension_id = bool(extension_id)
        if extension_id and not target_url.startswith(f"chrome-extension://{extension_id}/"):
            return False
        includes = self.options.get("service_worker_url_includes") or []
        suffixes = self.options.get("service_worker_url_suffixes") or []
        if includes and not all(part in target_url for part in includes):
            return False
        if suffixes and not any(target_url.endswith(suffix) for suffix in suffixes):
            return False
        return bool(has_extension_id or includes or suffixes)
