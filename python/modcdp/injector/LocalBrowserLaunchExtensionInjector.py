from __future__ import annotations

import base64
import hashlib
import json
import shutil
import tempfile
import zipfile
from pathlib import Path

from ..launcher.BrowserLauncher import BrowserLaunchOptions
from ..injector.ExtensionInjector import DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS, ExtensionInjector, ExtensionInjectionResult, defaultModCDPExtensionPath


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
        if not extension_path.endswith(".zip"):
            self.cleanup_dir = tempfile.TemporaryDirectory(prefix="modcdp-extension-")
            shutil.copytree(extension_path, self.cleanup_dir.name, dirs_exist_ok=True)
            self.unpacked_extension_path = _extension_root(self.cleanup_dir.name)
            self._resolveExtensionId()
            super().prepare()
            return
        self.cleanup_dir = tempfile.TemporaryDirectory(prefix="modcdp-extension-")
        with zipfile.ZipFile(extension_path) as archive:
            archive.extractall(self.cleanup_dir.name)
        self.unpacked_extension_path = _extension_root(self.cleanup_dir.name)
        self._resolveExtensionId()
        super().prepare()

    def getLauncherConfig(self) -> BrowserLaunchOptions:
        if not self.unpacked_extension_path:
            return {}
        return {"extra_args": [f"--load-extension={self.unpacked_extension_path}"]}

    def inject(self) -> ExtensionInjectionResult | None:
        timeout_ms = self.options.get("injector_service_worker_probe_timeout_ms") or DEFAULT_SERVICE_WORKER_PROBE_TIMEOUT_MS
        discovered = self._waitForReadyServiceWorker(
            timeout_ms,
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


def extensionIdFromManifestKey(extension_path: str) -> str | None:
    manifest_path = Path(extension_path) / "manifest.json"
    if not manifest_path.exists():
        return None
    manifest = json.loads(manifest_path.read_text())
    key = manifest.get("key") if isinstance(manifest, dict) else None
    if not isinstance(key, str) or not key.strip():
        return None
    digest = hashlib.sha256(base64.b64decode(key)).digest()[:16]
    alphabet = "abcdefghijklmnop"
    return "".join(alphabet[byte >> 4] + alphabet[byte & 0x0F] for byte in digest)


def _extension_root(unpacked_path: str) -> str:
    if (Path(unpacked_path) / "manifest.json").exists():
        return unpacked_path
    nested = Path(unpacked_path) / "extension"
    if (nested / "manifest.json").exists():
        return str(nested)
    return unpacked_path
