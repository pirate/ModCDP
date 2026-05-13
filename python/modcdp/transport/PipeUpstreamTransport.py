from __future__ import annotations

import json
import threading
from typing import Any

from ..transport.UpstreamTransport import UpstreamTransport


class PipeUpstreamTransport(UpstreamTransport):
    mode = "pipe"
    endpoint_kind = "raw_cdp"

    def __init__(self, options: dict[str, Any] | None = None) -> None:
        super().__init__()
        options = options or {}
        self.pipe_read = options.get("pipe_read")
        self.pipe_write = options.get("pipe_write")
        self.url = str(options.get("cdp_url") or "pipe://unknown")
        self._thread: threading.Thread | None = None
        self._connected = False
        self._closed = False

    def update(self, config: dict[str, Any] | None = None) -> "PipeUpstreamTransport":
        config = config or {}
        self.pipe_read = config.get("pipe_read") or self.pipe_read
        self.pipe_write = config.get("pipe_write") or self.pipe_write
        self.url = str(config.get("cdp_url") or self.url)
        return self

    def getLauncherConfig(self) -> dict[str, Any]:
        return {"remote_debugging": "pipe"}

    def connect(self) -> None:
        if self.pipe_read is None or self.pipe_write is None:
            raise RuntimeError("upstream.upstream_mode=pipe requires launcher-provided remote-debugging pipe handles.")
        if self._connected:
            return
        self._connected = True
        self._closed = False
        self._thread = threading.Thread(target=self._read_loop, daemon=True)
        self._thread.start()

    def send(self, message: dict[str, Any]) -> None:
        if not self._connected or self.pipe_write is None:
            raise RuntimeError("CDP pipe is not connected.")
        self.pipe_write.write(json.dumps(message).encode() + b"\0")
        self.pipe_write.flush()

    def close(self) -> None:
        self._closed = True
        for pipe in (self.pipe_read, self.pipe_write):
            try:
                if pipe is not None:
                    pipe.close()
            except Exception:
                pass
        self._connected = False

    def _read_loop(self) -> None:
        buffer = b""
        try:
            while not self._closed and self.pipe_read is not None:
                chunk = self.pipe_read.read(1)
                if not chunk:
                    if not self._closed:
                        self._handle_close(RuntimeError("CDP pipe closed"))
                    break
                buffer += chunk
                if b"\0" not in buffer:
                    continue
                raw, buffer = buffer.split(b"\0", 1)
                if raw:
                    self._parse_and_emit_recv(raw)
        except Exception as error:
            if not self._closed:
                self._handle_close(error if isinstance(error, Exception) else Exception(str(error)))

    def _handle_close(self, error: Exception) -> None:
        self._connected = False
        self._emit_close(error)
