from __future__ import annotations

import tempfile
import time
import zipfile
import shutil
from pathlib import Path

from ..injector.ExtensionInjector import DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS, DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS, DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS, ExtensionInjector, ExtensionInjectionResult, defaultModCDPExtensionPath


class ExtensionsLoadUnpackedInjector(ExtensionInjector):
    def __init__(self, options=None) -> None:
        super().__init__(options)
        self.unpacked_extension_path: str | None = None
        self.cleanup_dir: tempfile.TemporaryDirectory[str] | None = None

    def prepare(self) -> None:
        extension_path = self.options.get("injector_extension_path") or defaultModCDPExtensionPath()
        if not extension_path or self.unpacked_extension_path:
            super().prepare()
            return
        self.options["injector_extension_path"] = extension_path
        if not extension_path.endswith(".zip"):
            self.cleanup_dir = tempfile.TemporaryDirectory(prefix="modcdp-extension-")
            shutil.copytree(extension_path, self.cleanup_dir.name, dirs_exist_ok=True)
            self.unpacked_extension_path = _extension_root(self.cleanup_dir.name)
            super().prepare()
            return
        self.cleanup_dir = tempfile.TemporaryDirectory(prefix="modcdp-extension-")
        with zipfile.ZipFile(extension_path) as archive:
            archive.extractall(self.cleanup_dir.name)
        self.unpacked_extension_path = _extension_root(self.cleanup_dir.name)
        super().prepare()

    def inject(self) -> ExtensionInjectionResult | None:
        extension_path = self.unpacked_extension_path
        if not extension_path:
            return None
        try:
            load_result = self._sendWithTimeout("Extensions.loadUnpacked", {"path": extension_path})
        except RuntimeError as error:
            if "Method not available" in str(error) or "wasn't found" in str(error) or "Method not found" in str(error):
                self.last_error = error
                return None
            raise RuntimeError(
                f"Extensions.loadUnpacked failed for {extension_path}: {error}\n"
                "If the path is correct and the manifest is valid, load the ModCDP extension manually in chrome://extensions and reconnect."
            ) from error
        extension_id = load_result.get("id") or load_result.get("extensionId")
        if not isinstance(extension_id, str) or not extension_id:
            raise RuntimeError(f"Extensions.loadUnpacked returned no extension id (got {load_result})")
        self.options["injector_extension_id"] = extension_id
        self._wakeConfiguredExtension()

        sw_url_prefix = f"chrome-extension://{extension_id}/"
        deadline = time.monotonic() + (self.options.get("injector_service_worker_ready_timeout_ms") or DEFAULT_SERVICE_WORKER_READY_TIMEOUT_MS) / 1000
        while time.monotonic() < deadline:
            for target in self._targetInfos():
                if target["type"] != "service_worker" or not target["url"].startswith(sw_url_prefix):
                    continue
                probed = self._probeTarget(
                    target,
                    self.options.get("injector_service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS,
                    allow_attach=True,
                )
                if probed:
                    return {**probed, "source": "extensions_load_unpacked", "extension_id": extension_id}
            time.sleep((self.options.get("injector_service_worker_poll_interval_ms") or DEFAULT_SERVICE_WORKER_POLL_INTERVAL_MS) / 1000)
        raise RuntimeError(f"Timed out waiting for service worker target for extension {extension_id}.")

    def close(self) -> None:
        super().close()
        if self.cleanup_dir:
            self.cleanup_dir.cleanup()
            self.cleanup_dir = None


def _extension_root(unpacked_path: str) -> str:
    if (Path(unpacked_path) / "manifest.json").exists():
        return unpacked_path
    nested = Path(unpacked_path) / "extension"
    if (nested / "manifest.json").exists():
        return str(nested)
    return unpacked_path
