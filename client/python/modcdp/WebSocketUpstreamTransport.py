from __future__ import annotations

import json
from urllib.request import urlopen
from typing import Any

from websocket import create_connection

from .UpstreamTransport import UpstreamTransport


class WebSocketUpstreamTransport(UpstreamTransport):
    mode = "ws"
    endpoint_kind = "raw_cdp"

    def __init__(self, url: str | None = None, timeout_s: float = 10) -> None:
        super().__init__()
        self.url = url or ""
        self.timeout_s = timeout_s
        self.ws: Any | None = None

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
        self.ws = create_connection(self.url, timeout=self.timeout_s)

    def send(self, message: dict[str, Any]) -> None:
        if self.ws is None:
            raise RuntimeError("CDP websocket is not connected.")
        self.ws.send(json.dumps(message))

    def recv(self) -> Any:
        if self.ws is None:
            raise RuntimeError("CDP websocket is not connected.")
        return self.ws.recv()

    def close(self) -> None:
        if self.ws is not None:
            self.ws.close()
        self.ws = None


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
