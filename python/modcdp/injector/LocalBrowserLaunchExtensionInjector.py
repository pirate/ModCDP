from __future__ import annotations

import tempfile

from ..launcher.BrowserLauncher import BrowserLaunchOptions
from ..injector.ExtensionInjector import (
    ExtensionInjector,
    ExtensionInjectionResult,
    defaultModCDPExtensionPath,
    extensionIdFromManifestKey,
    prepareUnpackedExtension,
)


class LocalBrowserLaunchExtensionInjector(ExtensionInjector):
    def __init__(self, options=None) -> None:
        super().__init__(options)
        self.unpacked_extension_path: str | None = None
        self.extension_id: str | None = None
        self.cleanup_dir: tempfile.TemporaryDirectory[str] | None = None

    def prepare(self) -> None:
        extension_path = self.options.get("injector_extension_path") or defaultModCDPExtensionPath()
        if not extension_path or self.unpacked_extension_path:
            super().prepare()
            return
        self.options["injector_extension_path"] = extension_path
        self.unpacked_extension_path, self.cleanup_dir = prepareUnpackedExtension(extension_path)
        self._resolveExtensionId()
        super().prepare()

    def getLauncherConfig(self) -> BrowserLaunchOptions:
        if not self.unpacked_extension_path:
            return {}
        return {"extra_args": [f"--load-extension={self.unpacked_extension_path}"]}

    def inject(self) -> ExtensionInjectionResult | None:
        discovered = self._discoverReadyServiceWorker(
            matched_only=bool(self.options.get("injector_trust_service_worker_target")),
        )
        return {**discovered, "source": "local_launch"} if discovered else None

    def close(self) -> None:
        super().close()
        if self.cleanup_dir:
            self.cleanup_dir.cleanup()
            self.cleanup_dir = None

    def _resolveExtensionId(self) -> str | None:
        if self.extension_id:
            return self.extension_id
        configured_extension_id = self.options.get("injector_extension_id")
        if configured_extension_id:
            self.extension_id = configured_extension_id
        elif self.unpacked_extension_path:
            self.extension_id = extensionIdFromManifestKey(self.unpacked_extension_path)
        if self.extension_id:
            self.options["injector_extension_id"] = self.extension_id
        return self.extension_id
