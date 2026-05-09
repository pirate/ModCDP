from __future__ import annotations

import json
import socket
import struct
import subprocess
import sys
import threading
import time
from pathlib import Path
from collections.abc import Mapping
from typing import Any, cast

from .ExtensionInjector import DEFAULT_MODCDP_EXTENSION_ID
from .UpstreamTransport import UpstreamTransport


DEFAULT_NATIVE_MESSAGING_HOST_NAME = "com.modcdp.bridge"
DEFAULT_NATIVE_MESSAGING_WAIT_TIMEOUT_MS = 10_000


class NativeMessagingUpstreamTransport(UpstreamTransport):
    mode = "nativemessaging"
    endpoint_kind = "modcdp_server"

    def __init__(
        self,
        options: Mapping[str, Any] | None = None,
    ) -> None:
        super().__init__()
        normalized_options = dict(options or {})
        self.manifest_path = cast(str | None, normalized_options.get("manifest_path"))
        self.manifest_paths = list(cast(list[str], normalized_options.get("manifest_paths") or []))
        self.include_default_manifest_paths = self.manifest_path is None and not self.manifest_paths
        self.host_name = str(normalized_options.get("host_name") or DEFAULT_NATIVE_MESSAGING_HOST_NAME)
        self.extension_id = str(normalized_options.get("extension_id") or DEFAULT_MODCDP_EXTENSION_ID)
        self.wait_timeout_ms = int(normalized_options.get("wait_timeout_ms") or DEFAULT_NATIVE_MESSAGING_WAIT_TIMEOUT_MS)
        self.socket: socket.socket | None = None
        self.server: socket.socket | None = None
        self.peer_seen = threading.Event()
        self._peer_condition = threading.Condition()
        self._close_generation = 0
        self.bound_port: int | None = None
        self.cdp_url: str | None = None
        self.url = ""

    def update(self, config: dict[str, Any] | None = None) -> "NativeMessagingUpstreamTransport":
        config = config or {}
        should_install_native_host = False
        if "manifest_path" in config:
            self.manifest_path = config.get("manifest_path")
            should_install_native_host = True
        if "manifest_paths" in config:
            self.manifest_paths = list(config.get("manifest_paths") or [])
            should_install_native_host = True
        self.include_default_manifest_paths = self.manifest_path is None and not self.manifest_paths
        host_name = config.get("host_name") or config.get("native_host_name")
        if isinstance(host_name, str) and host_name:
            self.host_name = host_name
            should_install_native_host = True
        wait_timeout_ms = config.get("wait_timeout_ms")
        if isinstance(wait_timeout_ms, int | float):
            self.wait_timeout_ms = int(wait_timeout_ms)
        extension_id = config.get("extension_id")
        if isinstance(extension_id, str) and extension_id:
            self.extension_id = extension_id
            should_install_native_host = True
        user_data_dir = config.get("user_data_dir")
        if isinstance(user_data_dir, str) and user_data_dir:
            self._set_profile_manifest_paths(user_data_dir)
            should_install_native_host = True
        if should_install_native_host and self.bound_port is not None:
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
        with self._peer_condition:
            self._close_generation += 1
            self.peer_seen.clear()
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
        deadline = time.monotonic() + self.wait_timeout_ms / 1000
        with self._peer_condition:
            close_generation = self._close_generation
            while self.socket is None:
                if close_generation != self._close_generation:
                    raise RuntimeError(
                        f"Native messaging transport for {self.host_name} closed before a peer connected."
                    )
                remaining = deadline - time.monotonic()
                if remaining <= 0:
                    raise RuntimeError(
                        f"Timed out waiting {self.wait_timeout_ms}ms for native messaging host {self.host_name}."
                    )
                self._peer_condition.wait(remaining)

    def close(self) -> None:
        for sock in (self.socket, self.server):
            try:
                if sock is not None:
                    sock.close()
            except Exception:
                pass
        self.socket = None
        self.server = None
        self.peer_seen.clear()
        with self._peer_condition:
            self._close_generation += 1
            self._peer_condition.notify_all()

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
        with self._peer_condition:
            self.socket = conn
            self.peer_seen.set()
            self._peer_condition.notify_all()
        threading.Thread(target=self._read_loop, args=(conn,), daemon=True).start()

    def _read_loop(self, conn: socket.socket) -> None:
        try:
            for message in _read_length_prefixed_json_messages(conn):
                if message.get("type") == "modcdp.native.hello":
                    continue
                self._emit_recv(message)
        except Exception as error:
            if self.socket is conn:
                with self._peer_condition:
                    if self.socket is conn:
                        self.socket = None
                self._emit_close(error if isinstance(error, Exception) else Exception(str(error)))

    def _install_native_host(self, port: int) -> None:
        host_dir = Path.home() / ".modcdp" / "native-messaging"
        host_dir.mkdir(parents=True, exist_ok=True)
        config_path = host_dir / f"{self.host_name}.config.json"
        host_script_path = host_dir / f"{self.host_name}.py"
        host_executable_path = host_dir / f"{self.host_name}{'.cmd' if sys.platform.startswith('win') else '.sh'}"
        config_path.write_text(json.dumps({"host": "127.0.0.1", "port": port}, indent=2) + "\n")
        host_script_path.write_text(_native_host_script(str(config_path)))
        host_executable_path.write_text(_native_host_wrapper(sys.executable, str(host_script_path)))
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
        if sys.platform.startswith("win") and manifest_paths:
            _register_windows_native_messaging_host(self.host_name, manifest_paths[0])

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
    if sys.platform.startswith("win"):
        return [str(Path.home() / ".modcdp" / "native-messaging" / f"{host_name}.json")]
    raise RuntimeError("upstream nativemessaging manifest_path is required on this platform.")


def _native_host_wrapper(python_path: str, host_script_path: str) -> str:
    if sys.platform.startswith("win"):
        return f"@echo off\r\n{_cmd_quote(python_path)} {_cmd_quote(host_script_path)}\r\n"
    return f"#!/bin/sh\nexec {json.dumps(python_path)} {json.dumps(host_script_path)}\n"


def _cmd_quote(value: str) -> str:
    return f'"{value.replace(chr(34), chr(34) + chr(34))}"'


def _register_windows_native_messaging_host(host_name: str, manifest_path: str) -> None:
    subprocess.run(
        [
            "reg",
            "add",
            rf"HKCU\Software\Google\Chrome\NativeMessagingHosts\{host_name}",
            "/ve",
            "/t",
            "REG_SZ",
            "/d",
            manifest_path,
            "/f",
        ],
        check=True,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
    )


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
