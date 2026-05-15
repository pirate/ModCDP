from __future__ import annotations

import time
from pathlib import Path
from typing import Any, Mapping, cast

from ..injector.ExtensionInjector import DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS, DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS, DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS, EXT_ID_FROM_URL_RE, ExtensionInjector, ExtensionInjectionResult, MODCDP_READY_EXPRESSION


class BorrowedExtensionInjector(ExtensionInjector):
    def inject(self) -> ExtensionInjectionResult | None:
        deadline = time.monotonic() + (self.options.get("injector_service_worker_ready_timeout_ms") or DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS) / 1000
        while True:
            borrowed = self._borrowVisibleServiceWorkers()
            if borrowed:
                return borrowed
            if time.monotonic() >= deadline:
                return None
            time.sleep((self.options.get("injector_service_worker_poll_interval_ms") or DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS) / 1000)

    def _borrowVisibleServiceWorkers(self) -> ExtensionInjectionResult | None:
        borrowed: list[ExtensionInjectionResult] = []
        visible_service_workers = [
            target
            for target in self._targetInfos()
            if target.get("type") == "service_worker" and isinstance(target.get("url"), str) and target["url"].startswith("chrome-extension://")
        ]
        has_configured_matcher = bool(
            self.options.get("injector_extension_id")
            or self.options.get("injector_service_worker_url_includes")
            or self.options.get("injector_service_worker_url_suffixes")
        )
        candidates = [target for target in visible_service_workers if self._serviceWorkerTargetMatches(target)] if has_configured_matcher else visible_service_workers
        for target in candidates:
            try:
                bootstrapped = self._bootstrapTarget(target)
            except Exception:
                bootstrapped = None
            if bootstrapped:
                borrowed.append({**bootstrapped, "source": "borrowed"})
        borrowed.sort(key=lambda item: (bool(item.get("has_debugger")), bool(item.get("has_tabs"))), reverse=True)
        return borrowed[0] if borrowed else None

    def _bootstrapTarget(self, target) -> ExtensionInjectionResult | None:
        session_id = self._ensureSessionIdForTarget(
            target["targetId"],
            self.options.get("injector_service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
            True,
        )
        if session_id is None:
            return None
        try:
            self._sendWithTimeout("Runtime.enable", {}, session_id)
        except Exception:
            pass
        bootstrap = self._sendWithTimeout(
            "Runtime.evaluate",
            {
                "expression": f"({bootstrap_modcdp_server_expression()})()",
                "awaitPromise": True,
                "returnByValue": True,
            },
            session_id,
        )
        result = cast(Mapping[str, Any], bootstrap.get("result")) if isinstance(bootstrap.get("result"), Mapping) else {}
        raw_value = result.get("value")
        value = cast(Mapping[str, Any], raw_value) if isinstance(raw_value, Mapping) else {}
        if not bool(value.get("has_tabs")) or not bool(value.get("has_debugger")):
            return None
        ready = bool(value.get("ok"))
        if ready and self._readyExpression() != MODCDP_READY_EXPRESSION:
            probe = self._sendWithTimeout(
                "Runtime.evaluate",
                {
                    "expression": self._readyExpression(),
                    "returnByValue": True,
                },
                session_id,
            )
            probe_result = cast(Mapping[str, Any], probe.get("result")) if isinstance(probe.get("result"), Mapping) else {}
            ready = bool(probe_result.get("value"))
        if not ready:
            return None
        match = EXT_ID_FROM_URL_RE.match(target["url"])
        extension_id = value.get("extension_id") if isinstance(value.get("extension_id"), str) else None
        return {
            "source": "borrowed",
            "extension_id": extension_id or (match.group(1) if match else None),
            "target_id": target["targetId"],
            "url": target["url"],
            "session_id": session_id,
            "has_tabs": bool(value.get("has_tabs")),
            "has_debugger": bool(value.get("has_debugger")),
        }


def bootstrap_modcdp_server_expression() -> str:
    source = modcdp_server_source()
    start = source.index("export function installModCDPServer")
    end = source.index("export const ModCDPServer")
    installer = source[start:end].replace("export function", "function", 1)
    return (
        "function() {\n"
        "const __name = (fn) => fn;\n"
        f"{installer}\n"
        "const ModCDP = installModCDPServer(globalThis);\n"
        "return {\n"
        "  ok: Boolean(ModCDP?.__ModCDPServerVersion >= 1 && ModCDP?.handleCommand && ModCDP?.addCustomEvent),\n"
        "  extension_id: globalThis.chrome?.runtime?.id ?? null,\n"
        "  has_tabs: Boolean(globalThis.chrome?.tabs?.query),\n"
        "  has_debugger: Boolean(globalThis.chrome?.debugger?.sendCommand && globalThis.chrome?.debugger?.getTargets),\n"
        "};\n"
        "}"
    )


def modcdp_server_source() -> str:
    candidates: list[Path] = []
    for parent in Path(__file__).resolve().parents:
        candidates.append(parent / "dist" / "js" / "src" / "server" / "ModCDPServer.js")
        candidates.append(parent / "dist" / "extension" / "js" / "src" / "server" / "ModCDPServer.js")
        candidates.append(parent / "dist" / "extension" / "ModCDPServer.js")
        candidates.append(parent / "ModCDPServer.js")
    for candidate in candidates:
        if candidate.exists():
            return candidate.read_text()
    checked = ", ".join(str(candidate) for candidate in candidates)
    raise FileNotFoundError(f"Unable to locate ModCDPServer.js; checked: {checked}")
