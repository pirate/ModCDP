from __future__ import annotations

import json
import threading
from typing import Any

from websocket import create_connection

from ..launcher.BrowserLauncher import resolveCdpWebSocketUrl
from ..transport.UpstreamTransport import UpstreamTransport


class WebSocketUpstreamTransport(UpstreamTransport):
    mode = "ws"
    endpoint_kind = "raw_cdp"

    def __init__(self, options: dict[str, Any] | None = None) -> None:
        super().__init__()
        options = options or {}
        self.url = str(options.get("cdp_url") or "")
        self.ws: Any | None = None
        self._reader_thread: threading.Thread | None = None
        self._closed = False
        self._generation = 0

    def update(self, config: dict[str, Any] | None = None) -> "WebSocketUpstreamTransport":
        config = config or {}
        cdp_url = config.get("cdp_url")
        if cdp_url:
            self.url = str(cdp_url)
        return self

    def getServerConfig(self) -> dict[str, Any]:
        return {"server_loopback_cdp_url": self.url} if self.url else {}

    def connect(self) -> None:
        if not self.url:
            raise RuntimeError("upstream.upstream_mode=ws requires upstream.upstream_cdp_url or launcher-provided cdp_url.")
        # cdp_url may start as an HTTP discovery endpoint; from here on it is the resolved WebSocket CDP endpoint.
        self.url = resolveCdpWebSocketUrl(self.url, "upstream_cdp_url")
        previous_ws = self.ws
        if previous_ws is not None:
            previous_ws.close()
        self.ws = create_connection(self.url, timeout=10)
        self._closed = False
        self._generation += 1
        generation = self._generation
        self._reader_thread = threading.Thread(target=lambda: self._read_loop(generation), daemon=True)
        self._reader_thread.start()

    def send(self, message: dict[str, Any]) -> None:
        if self.ws is None:
            raise RuntimeError("CDP websocket is not connected.")
        self.ws.send(json.dumps(message))

    def _recv(self) -> Any:
        if self.ws is None:
            raise RuntimeError("CDP websocket is not connected.")
        return self.ws.recv()

    def close(self) -> None:
        self._closed = True
        self._generation += 1
        if self.ws is not None:
            self.ws.close()
        self.ws = None
        if self._reader_thread is not None and self._reader_thread.is_alive():
            self._reader_thread.join(timeout=1)
        self._reader_thread = None

    def _read_loop(self, generation: int | None = None) -> None:
        generation = self._generation if generation is None else generation
        ws = self.ws
        if ws is None:
            return
        try:
            while not self._closed and self.ws is ws and self._generation == generation:
                raw = ws.recv()
                if not raw:
                    break
                self._parse_and_emit_recv(raw)
        except Exception as error:
            if not self._closed and self.ws is ws and self._generation == generation:
                self._emit_close(error if isinstance(error, Exception) else RuntimeError(str(error)))
