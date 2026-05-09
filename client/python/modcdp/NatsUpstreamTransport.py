from __future__ import annotations

import json
import socket
import ssl
import threading
import time
from collections.abc import Mapping
from typing import Any, cast
from urllib.parse import parse_qs, urlencode, urlparse, urlunparse

from websocket import WebSocket

from .UpstreamTransport import UpstreamTransport


DEFAULT_NATS_URL = "ws://127.0.0.1:4223"
DEFAULT_NATS_SUBJECT_PREFIX = "modcdp.default"
DEFAULT_NATS_WAIT_TIMEOUT_MS = 10_000


class NatsUpstreamTransport(UpstreamTransport):
    mode = "nats"
    endpoint_kind = "modcdp_server"

    def __init__(
        self,
        options: str | Mapping[str, Any] | None = None,
    ) -> None:
        super().__init__()
        normalized_options = {"url": options} if isinstance(options, str) else dict(options or {})
        normalized_url, normalized_subject_prefix = _normalize_nats_url(
            cast(str | None, normalized_options.get("url")) or DEFAULT_NATS_URL,
            cast(str | None, normalized_options.get("subject_prefix")),
        )
        self.url = normalized_url
        self.subject_prefix = normalized_subject_prefix
        self.role = str(normalized_options.get("role") or "client")
        self.wait_timeout_ms = int(normalized_options.get("wait_timeout_ms") or DEFAULT_NATS_WAIT_TIMEOUT_MS)
        self.socket: WebSocket | socket.socket | None = None
        self.connected = False
        self.closed = False
        self.peer_seen = threading.Event()
        self._peer_condition = threading.Condition()
        self._close_generation = 0
        self.write_lock = threading.Lock()
        self.buffer = ""

    def update(self, config: dict[str, Any] | None = None) -> "NatsUpstreamTransport":
        config = config or {}
        nats_url = config.get("nats_url")
        nats_subject_prefix = config.get("nats_subject_prefix")
        if isinstance(nats_url, str) and nats_url or isinstance(nats_subject_prefix, str) and nats_subject_prefix:
            current_url = self.url or DEFAULT_NATS_URL
            self.url, self.subject_prefix = _normalize_nats_url(
                nats_url if isinstance(nats_url, str) and nats_url else current_url,
                nats_subject_prefix if isinstance(nats_subject_prefix, str) and nats_subject_prefix else self.subject_prefix,
            )
        role = config.get("role")
        if role in ("client", "browser"):
            self.role = str(role)
        wait_timeout_ms = config.get("wait_timeout_ms")
        if isinstance(wait_timeout_ms, int | float):
            self.wait_timeout_ms = int(wait_timeout_ms)
        return self

    def getInjectorConfig(self) -> dict[str, Any]:
        return {"nats_url": self.url, "nats_subject_prefix": self.subject_prefix}

    def connect(self) -> None:
        if self.connected:
            return
        self.closed = False
        with self._peer_condition:
            self._close_generation += 1
            close_generation = self._close_generation
            self.peer_seen.clear()
        parsed = urlparse(self.url)
        if parsed.scheme in ("ws", "wss"):
            ws = WebSocket()
            ws.connect(self.url)
            self.socket = ws
            self._write_protocol(f"CONNECT {json.dumps(_connect_options())}\r\nPING\r\n")
            threading.Thread(target=self._read_websocket_loop, args=(ws, close_generation), daemon=True).start()
        elif parsed.scheme in ("nats", "tls"):
            port = parsed.port or 4222
            host = cast(str, parsed.hostname or "127.0.0.1")
            raw_socket = socket.create_connection((host, port))
            tcp_socket = ssl.create_default_context().wrap_socket(raw_socket, server_hostname=host) if parsed.scheme == "tls" else raw_socket
            self.socket = tcp_socket
            self._write_protocol(f"CONNECT {json.dumps(_connect_options())}\r\nPING\r\n")
            threading.Thread(target=self._read_tcp_loop, args=(tcp_socket, close_generation), daemon=True).start()
        else:
            raise RuntimeError(f"upstream.mode=nats requires ws://, wss://, nats://, or tls:// URL, got {self.url}.")
        self.connected = True
        self._subscribe()
        self._publish(self._outgoing_subject(), {"type": "modcdp.nats.hello", "role": self.role, "version": 1})

    def send(self, message: dict[str, Any]) -> None:
        if not self.connected or self.socket is None:
            raise RuntimeError("NATS transport is not connected.")
        self._publish(self._outgoing_subject(), {"type": "modcdp.nats.message", "message": message})

    def waitForPeer(self) -> None:
        deadline = time.monotonic() + self.wait_timeout_ms / 1000
        with self._peer_condition:
            close_generation = self._close_generation
            while not self.peer_seen.is_set():
                if close_generation != self._close_generation:
                    raise RuntimeError(f"NATS transport for {self.subject_prefix} closed before a peer connected.")
                remaining = deadline - time.monotonic()
                if remaining <= 0:
                    raise RuntimeError(f"Timed out waiting {self.wait_timeout_ms}ms for NATS ModCDP peer.")
                self._peer_condition.wait(remaining)

    def close(self) -> None:
        self.closed = True
        try:
            if isinstance(self.socket, WebSocket):
                self.socket.close()
            elif self.socket is not None:
                self.socket.close()
        except Exception:
            pass
        self.socket = None
        self.connected = False
        self.peer_seen.clear()
        with self._peer_condition:
            self._close_generation += 1
            self._peer_condition.notify_all()

    def _subscribe(self) -> None:
        self._write_protocol(f"SUB {self._incoming_subject()} 1\r\n")

    def _publish(self, subject: str, message: dict[str, Any]) -> None:
        body = json.dumps(message, separators=(",", ":"))
        self._write_protocol(f"PUB {subject} {len(body.encode())}\r\n{body}\r\n")

    def _write_protocol(self, data: str) -> None:
        if self.socket is None:
            raise RuntimeError("NATS transport is not connected.")
        with self.write_lock:
            if isinstance(self.socket, WebSocket):
                self.socket.send(data)
            else:
                self.socket.sendall(data.encode())

    def _incoming_subject(self) -> str:
        return f"{self.subject_prefix}.{'browser_to_client' if self.role == 'client' else 'client_to_browser'}"

    def _outgoing_subject(self) -> str:
        return f"{self.subject_prefix}.{'client_to_browser' if self.role == 'client' else 'browser_to_client'}"

    def _generation_closed(self, close_generation: int) -> bool:
        with self._peer_condition:
            return self.closed or close_generation != self._close_generation

    def _read_websocket_loop(self, ws: WebSocket, close_generation: int) -> None:
        try:
            while not self._generation_closed(close_generation):
                data = ws.recv()
                if isinstance(data, bytes):
                    self.buffer += data.decode()
                else:
                    self.buffer += str(data)
                self.buffer = self._consume_protocol(self.buffer)
        except Exception as error:
            if not self._generation_closed(close_generation):
                self._emit_close(error if isinstance(error, Exception) else RuntimeError(str(error)))

    def _read_tcp_loop(self, tcp_socket: socket.socket, close_generation: int) -> None:
        try:
            while not self._generation_closed(close_generation):
                chunk = tcp_socket.recv(65536)
                if not chunk:
                    break
                self.buffer += chunk.decode()
                self.buffer = self._consume_protocol(self.buffer)
        except Exception as error:
            if not self._generation_closed(close_generation):
                self._emit_close(error if isinstance(error, Exception) else RuntimeError(str(error)))

    def _consume_protocol(self, buffer: str) -> str:
        while True:
            line_end = buffer.find("\r\n")
            if line_end < 0:
                return buffer
            line = buffer[:line_end]
            upper = line.upper()
            if upper.startswith("MSG "):
                parts = line.split()
                size = int(parts[-1]) if parts and parts[-1].isdigit() else -1
                payload_start = line_end + 2
                payload_end = payload_start + size
                if size < 0 or len(buffer) < payload_end + 2:
                    return buffer
                payload = buffer[payload_start:payload_end]
                buffer = buffer[payload_end + 2 :]
                self._handle_payload(payload)
                continue
            buffer = buffer[line_end + 2 :]
            if upper == "PING":
                self._write_protocol("PONG\r\n")
            elif upper.startswith("-ERR"):
                self._emit_close(RuntimeError(f"NATS error: {line}"))

    def _handle_payload(self, payload: str) -> None:
        try:
            parsed = json.loads(payload)
        except Exception:
            return
        if isinstance(parsed, dict) and parsed.get("type") == "modcdp.nats.hello":
            with self._peer_condition:
                self.peer_seen.set()
                self._peer_condition.notify_all()
            return
        message = parsed.get("message") if isinstance(parsed, dict) and parsed.get("type") == "modcdp.nats.message" else parsed
        if isinstance(message, dict):
            self._emit_recv(message)


def _connect_options() -> dict[str, Any]:
    return {"verbose": False, "pedantic": False, "lang": "modcdp", "version": "1", "protocol": 1}


def _normalize_nats_url(url: str, subject_prefix: str | None = None) -> tuple[str, str]:
    parsed = urlparse(url)
    query = parse_qs(parsed.query)
    subject = subject_prefix or (query.get("subject") or query.get("subject_prefix") or [None])[0]
    query.pop("subject", None)
    query.pop("subject_prefix", None)
    normalized_query = urlencode(query, doseq=True)
    normalized_path = parsed.path or ("/" if parsed.scheme in ("ws", "wss") else "")
    normalized_url = urlunparse((parsed.scheme, parsed.netloc, normalized_path, parsed.params, normalized_query, parsed.fragment))
    return normalized_url, _sanitize_subject_prefix(subject or DEFAULT_NATS_SUBJECT_PREFIX)


def _sanitize_subject_prefix(value: str) -> str:
    subject = value.strip()
    if not subject or any(char.isspace() or char in "*>" for char in subject):
        raise ValueError(f"Invalid NATS subject prefix {value}")
    return subject
