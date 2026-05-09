from __future__ import annotations

from .ExtensionInjector import DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS, DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS, ExtensionInjector, ExtensionInjectionResult


class DiscoveredExtensionInjector(ExtensionInjector):
    def inject(self) -> ExtensionInjectionResult | None:
        discovered = self._discoverReadyServiceWorker()
        if discovered:
            return {**discovered, "source": "discovered"}
        if self.options.get("trust_matched_service_worker"):
            waited = self._waitForReadyServiceWorker(
                self.options.get("service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
                matched_only=True,
            )
            if waited:
                return {**waited, "source": "discovered"}
        woke = self._wakeConfiguredExtension()
        if woke:
            waited = self._waitForReadyServiceWorker(
                self.options.get("service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
                matched_only=bool(self.options.get("trust_matched_service_worker")),
            )
            if waited:
                return {**waited, "source": "discovered"}
        if not self.options.get("require_service_worker_target"):
            return None
        waited = self._waitForReadyServiceWorker(
            self.options.get("service_worker_ready_timeout_ms") or DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS,
            matched_only=bool(self.options.get("trust_matched_service_worker")),
        )
        if waited:
            return {**waited, "source": "discovered"}
        matchers = ", ".join(
            [
                *(self.options.get("service_worker_url_includes") or []),
                *(self.options.get("service_worker_url_suffixes") or []),
            ]
        )
        raise RuntimeError(f"Required ModCDP service worker target was not visible ({matchers or 'no matcher'}).")
