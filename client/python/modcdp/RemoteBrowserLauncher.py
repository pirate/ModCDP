from __future__ import annotations

import json
import urllib.request

from .BrowserLauncher import BrowserLaunchOptions, BrowserLauncher, LaunchedBrowser


class RemoteBrowserLauncher(BrowserLauncher):
    def __init__(self, options: BrowserLaunchOptions | None = None, cdp_url: str | None = None) -> None:
        super().__init__(options)
        self.cdp_url = cdp_url

    def launch(self, options: BrowserLaunchOptions | None = None) -> LaunchedBrowser:
        merged = {**self.options, **dict(options or {})}
        cdp_url = self.cdp_url or merged.get("ws_url") or merged.get("cdp_url")
        if not cdp_url:
            raise RuntimeError("launch.mode='remote' requires upstream.ws_url.")
        ws_url = _websocket_url_for(cdp_url)
        self.launched = {"cdp_url": cdp_url, "ws_url": ws_url, "close": lambda: None}
        return self.launched


def _websocket_url_for(endpoint: str) -> str:
    if endpoint.startswith("ws://") or endpoint.startswith("wss://"):
        return endpoint
    with urllib.request.urlopen(f"{endpoint.rstrip('/')}/json/version", timeout=5) as response:
        version = json.loads(response.read())
    ws_url = version.get("webSocketDebuggerUrl") if isinstance(version, dict) else None
    if not isinstance(ws_url, str) or not ws_url:
        raise RuntimeError(f"HTTP discovery for {endpoint} returned no webSocketDebuggerUrl")
    return ws_url
