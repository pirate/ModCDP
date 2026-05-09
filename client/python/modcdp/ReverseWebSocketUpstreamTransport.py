from __future__ import annotations

import base64
import hashlib
import json
import socket
import struct
import threading
import time
from typing import Any
from urllib.parse import urlparse

from .UpstreamTransport import UpstreamTransport


DEFAULT_REVERSEWS_BIND = "127.0.0.1:29292"
DEFAULT_REVERSEWS_WAIT_TIMEOUT_MS = 10_000
_WS_GUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"


class ReverseWebSocketUpstreamTransport(UpstreamTransport):
    mode = "reversews"
    endpoint_kind = "modcdp_server"

    def __init__(self, bind: str = DEFAULT_REVERSEWS_BIND, wait_timeout_ms: int = DEFAULT_REVERSEWS_WAIT_TIMEOUT_MS) -> None:
        super().__init__()
        self.wait_timeout_ms = wait_timeout_ms
        self.server_socket: socket.socket | None = None
        self.socket: socket.socket | None = None
        self.accept_thread: threading.Thread | None = None
        self.reader_thread: threading.Thread | None = None
        self.peer_info: dict[str, Any] | None = None
        self.peer_event = threading.Event()
        self._peer_condition = threading.Condition()
        self._close_generation = 0
        self.closed = False
        self.write_lock = threading.Lock()
        self._setBind(bind)

    def update(self, config: dict[str, Any] | None = None) -> "ReverseWebSocketUpstreamTransport":
        config = config or {}
        bind = config.get("reversews_bind") or config.get("url")
        if isinstance(bind, str) and bind:
            self._setBind(bind)
        wait_timeout_ms = config.get("reversews_wait_timeout_ms")
        if isinstance(wait_timeout_ms, int | float):
            self.wait_timeout_ms = int(wait_timeout_ms)
        return self

    def getInjectorConfig(self) -> dict[str, Any]:
        return {"reverse_proxy_url": self.url}

    def _setBind(self, bind: str) -> None:
        parsed = urlparse(bind if "://" in bind else f"ws://{bind}")
        host = parsed.hostname or "127.0.0.1"
        port = parsed.port or 29292
        if port <= 0 or port > 65535:
            raise ValueError(f"Invalid host:port {bind}")
        self.url = f"ws://{host}:{port}"

    def connect(self) -> None:
        parsed = urlparse(self.url or "")
        host = parsed.hostname or "127.0.0.1"
        port = parsed.port or 29292
        server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        server_socket.bind((host, port))
        server_socket.listen(1)
        self.server_socket = server_socket
        with self._peer_condition:
            self._close_generation += 1
            self.peer_event.clear()
        self.closed = False
        self.accept_thread = threading.Thread(target=self._accept_loop, daemon=True)
        self.accept_thread.start()

    def send(self, message: dict[str, Any]) -> None:
        if self.socket is None:
            raise RuntimeError(f"No reverse ModCDP extension peer is connected at {self.url}.")
        payload = json.dumps(message).encode()
        with self.write_lock:
            self.socket.sendall(_encode_server_text_frame(payload))

    def waitForPeer(self) -> None:
        if self.socket is not None:
            return
        deadline = time.monotonic() + self.wait_timeout_ms / 1000
        with self._peer_condition:
            close_generation = self._close_generation
            while self.socket is None:
                if close_generation != self._close_generation:
                    raise RuntimeError(f"Reverse websocket transport at {self.url} closed before a peer connected.")
                remaining = deadline - time.monotonic()
                if remaining <= 0:
                    raise RuntimeError(f"Timed out waiting {self.wait_timeout_ms}ms for reverse ModCDP extension connection.")
                self._peer_condition.wait(remaining)

    def close(self) -> None:
        self.closed = True
        for sock in (self.socket, self.server_socket):
            try:
                sock.close() if sock is not None else None
            except Exception:
                pass
        self.socket = None
        self.server_socket = None
        self.peer_info = None
        self.peer_event.clear()
        with self._peer_condition:
            self._close_generation += 1
            self._peer_condition.notify_all()

    def _accept_loop(self) -> None:
        while not self.closed and self.server_socket is not None:
            try:
                sock, _ = self.server_socket.accept()
                self._accept(sock)
            except OSError:
                return
            except Exception as error:
                self._emit_close(error if isinstance(error, Exception) else RuntimeError(str(error)))

    def _accept(self, sock: socket.socket) -> None:
        try:
            _perform_server_handshake(sock)
            hello_raw = _read_client_text_frame(sock, self.wait_timeout_ms / 1000)
            if hello_raw is None:
                raise RuntimeError("reverse hello socket closed")
            hello = json.loads(hello_raw)
            if not isinstance(hello, dict) or hello.get("type") != "modcdp.reverse.hello":
                raise RuntimeError("invalid reverse hello")
            old_socket = self.socket
            if old_socket is not None and old_socket is not sock:
                try:
                    old_socket.close()
                except Exception:
                    pass
            with self._peer_condition:
                self.socket = sock
                self.peer_info = hello
                self.peer_event.set()
                self._peer_condition.notify_all()
            self.reader_thread = threading.Thread(target=self._read_loop, args=(sock,), daemon=True)
            self.reader_thread.start()
        except Exception as error:
            try:
                sock.close()
            except Exception:
                pass
            self._emit_close(error if isinstance(error, Exception) else RuntimeError(str(error)))

    def _read_loop(self, sock: socket.socket) -> None:
        try:
            while not self.closed and self.socket is sock:
                data = _read_client_text_frame(sock, None)
                if data is None:
                    break
                self._parse_and_emit_recv(data)
        except Exception as error:
            if not self.closed:
                self._emit_close(error if isinstance(error, Exception) else RuntimeError(str(error)))
        finally:
            if self.socket is sock:
                with self._peer_condition:
                    if self.socket is sock:
                        self.socket = None
                        self.peer_info = None
                        self.peer_event.clear()


def _perform_server_handshake(sock: socket.socket) -> None:
    request = b""
    while b"\r\n\r\n" not in request:
        chunk = sock.recv(4096)
        if not chunk:
            raise RuntimeError("websocket handshake closed")
        request += chunk
    headers: dict[str, str] = {}
    for line in request.decode(errors="replace").split("\r\n")[1:]:
        if ":" in line:
            key, value = line.split(":", 1)
            headers[key.strip().lower()] = value.strip()
    ws_key = headers.get("sec-websocket-key")
    if not ws_key:
        raise RuntimeError("websocket handshake missing Sec-WebSocket-Key")
    accept = base64.b64encode(hashlib.sha1((ws_key + _WS_GUID).encode()).digest()).decode()
    sock.sendall(
        (
            "HTTP/1.1 101 Switching Protocols\r\n"
            "Upgrade: websocket\r\n"
            "Connection: Upgrade\r\n"
            f"Sec-WebSocket-Accept: {accept}\r\n"
            "\r\n"
        ).encode()
    )


def _read_exact(sock: socket.socket, length: int) -> bytes:
    chunks: list[bytes] = []
    remaining = length
    while remaining > 0:
        chunk = sock.recv(remaining)
        if not chunk:
            raise EOFError("websocket closed")
        chunks.append(chunk)
        remaining -= len(chunk)
    return b"".join(chunks)


def _read_client_text_frame(sock: socket.socket, timeout_s: float | None) -> str | None:
    old_timeout = sock.gettimeout()
    if timeout_s is not None:
        sock.settimeout(timeout_s)
    try:
        header = _read_exact(sock, 2)
        opcode = header[0] & 0x0F
        masked = bool(header[1] & 0x80)
        length = header[1] & 0x7F
        if length == 126:
            length = struct.unpack("!H", _read_exact(sock, 2))[0]
        elif length == 127:
            length = struct.unpack("!Q", _read_exact(sock, 8))[0]
        mask = _read_exact(sock, 4) if masked else b""
        payload = bytearray(_read_exact(sock, length))
        if masked:
            for index, value in enumerate(payload):
                payload[index] = value ^ mask[index % 4]
        if opcode == 0x8:
            return None
        if opcode != 0x1:
            return ""
        return payload.decode()
    finally:
        sock.settimeout(old_timeout)


def _encode_server_text_frame(payload: bytes) -> bytes:
    if len(payload) < 126:
        header = bytes([0x81, len(payload)])
    elif len(payload) < 65536:
        header = bytes([0x81, 126]) + struct.pack("!H", len(payload))
    else:
        header = bytes([0x81, 127]) + struct.pack("!Q", len(payload))
    return header + payload
