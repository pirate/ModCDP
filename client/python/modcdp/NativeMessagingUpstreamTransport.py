from __future__ import annotations

import json
import socket
import struct
import sys
import threading
from pathlib import Path
from typing import Any

from .ExtensionInjector import DEFAULT_MODCDP_EXTENSION_ID
from .UpstreamTransport import UpstreamTransport


DEFAULT_NATIVE_MESSAGING_HOST_NAME = "com.modcdp.bridge"
DEFAULT_NATIVE_MESSAGING_WAIT_TIMEOUT_MS = 10_000


class NativeMessagingUpstreamTransport(UpstreamTransport):
    mode = "nativemessaging"
    endpoint_kind = "modcdp_server"

    def __init__(
        self,
        manifest_path: str | None = None,
        host_name: str = DEFAULT_NATIVE_MESSAGING_HOST_NAME,
        extension_id: str = DEFAULT_MODCDP_EXTENSION_ID,
        wait_timeout_ms: int = DEFAULT_NATIVE_MESSAGING_WAIT_TIMEOUT_MS,
    ) -> None:
        super().__init__()
        self.manifest_path = manifest_path
        self.manifest_paths: list[str] = []
        self.include_default_manifest_paths = manifest_path is None
        self.host_name = host_name
        self.extension_id = extension_id
        self.wait_timeout_ms = wait_timeout_ms
        self.socket: socket.socket | None = None
        self.server: socket.socket | None = None
        self.peer_seen = threading.Event()
        self.bound_port: int | None = None
        self.cdp_url: str | None = None
        self.url = ""

    def update(self, config: dict[str, Any] | None = None) -> "NativeMessagingUpstreamTransport":
        config = config or {}
        if "manifest_path" in config:
            self.manifest_path = config.get("manifest_path")
            self.include_default_manifest_paths = self.manifest_path is None
        if "manifest_paths" in config:
            self.manifest_paths = list(config.get("manifest_paths") or [])
            self.include_default_manifest_paths = len(self.manifest_paths) == 0
        self.extension_id = str(config.get("extension_id") or self.extension_id)
        user_data_dir = config.get("user_data_dir")
        if isinstance(user_data_dir, str) and user_data_dir:
            self._set_profile_manifest_paths(user_data_dir)
            if self.bound_port is not None:
                self._install_native_host(self.bound_port)
        cdp_url = config.get("ws_url") or config.get("cdp_url")
        if isinstance(cdp_url, str) and cdp_url:
            self.cdp_url = cdp_url
        return self

    def getServerConfig(self) -> dict[str, Any]:
        return {"loopback_cdp_url": self.cdp_url} if self.cdp_url else {}

    def getInjectorConfig(self) -> dict[str, Any]:
        return {"native_host_name": self.host_name}

    def connect(self) -> None:
        server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server.bind(("127.0.0.1", 0))
        server.listen(1)
        self.server = server
        self.bound_port = int(server.getsockname()[1])
        self.url = f"native://{self.host_name}@127.0.0.1:{self.bound_port}"
        self._install_native_host(self.bound_port)
        threading.Thread(target=self._accept_loop, daemon=True).start()

    def send(self, message: dict[str, Any]) -> None:
        if self.socket is None:
            raise RuntimeError(f"No native messaging peer is connected for {self.host_name}.")
        _write_length_prefixed_json(self.socket, message)

    def waitForPeer(self) -> None:
        if not self.peer_seen.wait(self.wait_timeout_ms / 1000):
            raise RuntimeError(f"Timed out waiting {self.wait_timeout_ms}ms for native messaging host {self.host_name}.")

    def close(self) -> None:
        for sock in (self.socket, self.server):
            try:
                if sock is not None:
                    sock.close()
            except Exception:
                pass
        self.socket = None
        self.server = None

    def _accept_loop(self) -> None:
        server = self.server
        if server is None:
            return
        try:
            conn, _ = server.accept()
        except Exception as error:
            self._emit_close(error if isinstance(error, Exception) else Exception(str(error)))
            return
        if self.socket is not None:
            try:
                self.socket.close()
            except Exception:
                pass
        self.socket = conn
        self.peer_seen.set()
        threading.Thread(target=self._read_loop, args=(conn,), daemon=True).start()

    def _read_loop(self, conn: socket.socket) -> None:
        try:
            for message in _read_length_prefixed_json_messages(conn):
                if message.get("type") == "modcdp.native.hello":
                    continue
                self._emit_recv(message)
        except Exception as error:
            if self.socket is conn:
                self.socket = None
                self._emit_close(error if isinstance(error, Exception) else Exception(str(error)))

    def _install_native_host(self, port: int) -> None:
        host_dir = Path.home() / ".modcdp" / "native-messaging"
        host_dir.mkdir(parents=True, exist_ok=True)
        config_path = host_dir / f"{self.host_name}.config.json"
        host_script_path = host_dir / f"{self.host_name}.py"
        host_executable_path = host_dir / f"{self.host_name}.sh"
        config_path.write_text(json.dumps({"host": "127.0.0.1", "port": port}, indent=2) + "\n")
        host_script_path.write_text(_native_host_script(str(config_path)))
        host_executable_path.write_text(f"#!/bin/sh\nexec {json.dumps(sys.executable)} {json.dumps(str(host_script_path))}\n")
        host_executable_path.chmod(0o755)

        manifest_paths = []
        if self.manifest_path:
            manifest_paths.append(self.manifest_path)
        manifest_paths.extend(self.manifest_paths)
        if self.include_default_manifest_paths:
            manifest_paths.extend(_default_native_messaging_manifest_paths(self.host_name))
        manifest = {
            "name": self.host_name,
            "description": "ModCDP Native Messaging bridge",
            "path": str(host_executable_path),
            "type": "stdio",
            "allowed_origins": [f"chrome-extension://{self.extension_id}/"],
        }
        manifest_text = json.dumps(manifest, indent=2) + "\n"
        for manifest_path in manifest_paths:
            path = Path(manifest_path)
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(manifest_text)

    def _set_profile_manifest_paths(self, user_data_dir: str) -> None:
        self.manifest_paths = [
            str(Path(user_data_dir) / "NativeMessagingHosts" / f"{self.host_name}.json"),
            str(Path(user_data_dir) / "Default" / "NativeMessagingHosts" / f"{self.host_name}.json"),
            *self.manifest_paths,
        ]


def _default_native_messaging_manifest_paths(host_name: str) -> list[str]:
    home = str(Path.home())
    if sys.platform == "darwin":
        return [
            f"{home}/Library/Application Support/Google/Chrome/NativeMessagingHosts/{host_name}.json",
            f"{home}/Library/Application Support/Google/Chrome Canary/NativeMessagingHosts/{host_name}.json",
            f"{home}/Library/Application Support/Google/ChromeForTesting/NativeMessagingHosts/{host_name}.json",
            f"{home}/Library/Application Support/Google/Chrome for Testing/NativeMessagingHosts/{host_name}.json",
            f"{home}/Library/Application Support/Google/Chrome SxS/NativeMessagingHosts/{host_name}.json",
            f"{home}/Library/Application Support/Chromium/NativeMessagingHosts/{host_name}.json",
        ]
    if sys.platform.startswith("linux"):
        return [
            f"{home}/.config/google-chrome/NativeMessagingHosts/{host_name}.json",
            f"{home}/.config/google-chrome-for-testing/NativeMessagingHosts/{host_name}.json",
            f"{home}/.config/chromium/NativeMessagingHosts/{host_name}.json",
            f"{home}/.config/chromium-browser/NativeMessagingHosts/{host_name}.json",
        ]
    raise RuntimeError("upstream nativemessaging manifest_path is required on this platform.")


def _write_length_prefixed_json(sock: socket.socket, message: dict[str, Any]) -> None:
    body = json.dumps(message).encode()
    sock.sendall(struct.pack("<I", len(body)) + body)


def _read_exact(sock: socket.socket, length: int) -> bytes:
    chunks: list[bytes] = []
    remaining = length
    while remaining > 0:
        chunk = sock.recv(remaining)
        if not chunk:
            raise RuntimeError("native messaging socket closed")
        chunks.append(chunk)
        remaining -= len(chunk)
    return b"".join(chunks)


def _read_length_prefixed_json_messages(sock: socket.socket):
    while True:
        header = _read_exact(sock, 4)
        length = struct.unpack("<I", header)[0]
        body = _read_exact(sock, length)
        message = json.loads(body.decode())
        if isinstance(message, dict):
            yield message


def _native_host_script(config_path: str) -> str:
    return f"""
import json
import socket
import struct
import sys
import threading

config = json.loads(open({config_path!r}, "r", encoding="utf-8").read())
sock = socket.create_connection((config["host"], config["port"]))

def write_native(message):
    body = json.dumps(message).encode()
    sys.stdout.buffer.write(struct.pack("<I", len(body)) + body)
    sys.stdout.buffer.flush()

def write_tcp(message):
    body = json.dumps(message).encode()
    sock.sendall(struct.pack("<I", len(body)) + body)

def read_exact(source, length):
    chunks = []
    remaining = length
    while remaining > 0:
        chunk = source.read(remaining) if hasattr(source, "read") else source.recv(remaining)
        if not chunk:
            raise SystemExit(0)
        chunks.append(chunk)
        remaining -= len(chunk)
    return b"".join(chunks)

def read_messages(source, on_message):
    while True:
        header = read_exact(source, 4)
        length = struct.unpack("<I", header)[0]
        body = read_exact(source, length)
        on_message(json.loads(body.decode()))

write_tcp({{"type": "modcdp.native.hello", "role": "native-host", "version": 1}})
threading.Thread(target=read_messages, args=(sys.stdin.buffer, write_tcp), daemon=True).start()
read_messages(sock, write_native)
"""
