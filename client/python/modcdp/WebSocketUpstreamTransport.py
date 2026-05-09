from __future__ import annotations

import json
import threading
from urllib.request import urlopen
from typing import Any

from websocket import create_connection

from .UpstreamTransport import UpstreamTransport


class WebSocketUpstreamTransport(UpstreamTransport):
    mode = "ws"
    endpoint_kind = "raw_cdp"

    def __init__(self, url: str | None = None) -> None:
        super().__init__()
        self.url = url or ""
        self.ws: Any | None = None
        self._reader_thread: threading.Thread | None = None
        self._closed = False

    def update(self, config: dict[str, Any] | None = None) -> "WebSocketUpstreamTransport":
        config = config or {}
        url = config.get("ws_url") or config.get("cdp_url") or config.get("url")
        if url:
            self.url = str(url)
        return self

    def getServerConfig(self) -> dict[str, Any]:
        return {"loopback_cdp_url": self.url} if self.url else {}

    def connect(self) -> None:
        if not self.url:
            raise RuntimeError("upstream.mode='ws' requires upstream.ws_url or launcher-provided ws_url.")
        self.url = _websocket_url_for(self.url)
        self.ws = create_connection(self.url, timeout=10)
        self._closed = False
        self._reader_thread = threading.Thread(target=self._read_loop, daemon=True)
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
        if self.ws is not None:
            self.ws.close()
        self.ws = None
        if self._reader_thread is not None and self._reader_thread.is_alive():
            self._reader_thread.join(timeout=1)
        self._reader_thread = None

    def _read_loop(self) -> None:
        ws = self.ws
        if ws is None:
            return
        try:
            while not self._closed:
                raw = ws.recv()
                if not raw:
                    break
                self._parse_and_emit_recv(raw)
        except Exception as error:
            if not self._closed:
                self._emit_close(error if isinstance(error, Exception) else RuntimeError(str(error)))


def _websocket_url_for(endpoint: str) -> str:
    if endpoint.startswith(("ws://", "wss://")):
        return endpoint
    http_endpoint = endpoint if "://" in endpoint else f"http://{endpoint}"
    with urlopen(http_endpoint.rstrip("/") + "/json/version", timeout=10) as response:
        version = json.loads(response.read().decode())
    ws_url = version.get("webSocketDebuggerUrl")
    if not isinstance(ws_url, str) or not ws_url:
        raise RuntimeError("upstream.ws_url HTTP discovery returned no webSocketDebuggerUrl")
    return ws_url
