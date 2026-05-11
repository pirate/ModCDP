from __future__ import annotations

import json
import urllib.request

from typing import cast

from ..launcher.BrowserLauncher import BrowserLaunchOptions, BrowserLauncher, LaunchedBrowser


class RemoteBrowserLauncher(BrowserLauncher):
    def __init__(self, options: BrowserLaunchOptions | None = None, cdp_url: str | None = None) -> None:
        super().__init__(cast(BrowserLaunchOptions, {**dict(options or {}), **({"cdp_url": cdp_url} if cdp_url is not None else {})}))

    def launch(self, options: BrowserLaunchOptions | None = None) -> LaunchedBrowser:
        merged = {**self.options, **dict(options or {})}
        cdp_url = cast(str | None, merged.get("cdp_url"))
        if not cdp_url:
            raise RuntimeError("launcher.launcher_mode=remote requires upstream.upstream_cdp_url.")
        # cdp_url is resolved here so downstream transports can dial it directly.
        cdp_url = _websocket_url_for(cdp_url)
        self.launched = {"cdp_url": cdp_url, "close": lambda: None}
        return self.launched


def _websocket_url_for(endpoint: str) -> str:
    if endpoint.startswith("ws://") or endpoint.startswith("wss://"):
        return endpoint
    with urllib.request.urlopen(f"{endpoint.rstrip('/')}/json/version", timeout=5) as response:
        version = json.loads(response.read())
    cdp_url = version.get("webSocketDebuggerUrl") if isinstance(version, dict) else None
    if not isinstance(cdp_url, str) or not cdp_url:
        raise RuntimeError(f"HTTP discovery for {endpoint} returned no webSocketDebuggerUrl")
    return cdp_url
