from __future__ import annotations

import json
from collections.abc import Callable
from typing import Any, Literal, TypedDict


UpstreamMode = Literal["ws", "pipe", "nativemessaging", "reversews", "nats"]
UpstreamEndpointKind = Literal["raw_cdp", "modcdp_server"]


class UpstreamTransport:
    mode: UpstreamMode
    endpoint_kind: UpstreamEndpointKind
    url: str | None = None

    def __init__(self) -> None:
        self._recv_listeners: list[Callable[[dict[str, Any]], None]] = []
        self._close_listeners: list[Callable[[Exception], None]] = []

    def connect(self) -> None:
        raise NotImplementedError(f"{type(self).__name__}.connect is not implemented.")

    def update(self, config: dict[str, Any] | None = None) -> "UpstreamTransport":
        return self

    def getLauncherConfig(self) -> dict[str, Any]:
        return {}

    def getInjectorConfig(self) -> dict[str, Any]:
        return {}

    def getServerConfig(self) -> dict[str, Any]:
        return {}

    def close(self) -> None:
        return None

    def send(self, message: dict[str, Any]) -> None:
        raise NotImplementedError(f"{type(self).__name__}.send is not implemented.")

    def onRecv(self, listener: Callable[[dict[str, Any]], None]) -> Callable[[], None]:
        self._recv_listeners.append(listener)

        removed = False

        def stop() -> None:
            nonlocal removed
            if removed:
                return
            removed = True
            try:
                self._recv_listeners.remove(listener)
            except ValueError:
                return

        return stop

    def onClose(self, listener: Callable[[Exception], None]) -> Callable[[], None]:
        self._close_listeners.append(listener)

        removed = False

        def stop() -> None:
            nonlocal removed
            if removed:
                return
            removed = True
            try:
                self._close_listeners.remove(listener)
            except ValueError:
                return

        return stop

    def waitForPeer(self) -> None:
        return None

    def _emit_recv(self, message: dict[str, Any]) -> None:
        for listener in list(self._recv_listeners):
            listener(message)

    def _emit_close(self, error: Exception) -> None:
        for listener in list(self._close_listeners):
            listener(error)

    def _parse_and_emit_recv(self, data: str | bytes) -> None:
        try:
            raw = data.decode() if isinstance(data, bytes) else data
            parsed = json.loads(raw)
            if isinstance(parsed, dict):
                self._emit_recv(parsed)
        except Exception:
            return


def endpoint_kind_for_upstream(mode: str) -> UpstreamEndpointKind:
    return "raw_cdp" if mode in ("ws", "pipe") else "modcdp_server"
