from __future__ import annotations

from typing import cast

from ..launcher.BrowserLauncher import BrowserLaunchOptions, BrowserLauncher, LaunchedBrowser, resolveCdpWebSocketUrl


class RemoteBrowserLauncher(BrowserLauncher):
    def __init__(self, options: BrowserLaunchOptions | None = None, cdp_url: str | None = None) -> None:
        super().__init__(cast(BrowserLaunchOptions, {**dict(options or {}), **({"cdp_url": cdp_url} if cdp_url is not None else {})}))

    def launch(self, options: BrowserLaunchOptions | None = None) -> LaunchedBrowser:
        merged = {**self.options, **dict(options or {})}
        cdp_url = cast(str | None, merged.get("cdp_url"))
        if not cdp_url:
            raise RuntimeError("launcher.launcher_mode=remote requires upstream.upstream_cdp_url.")
        # cdp_url is resolved here so downstream transports can dial it directly.
        cdp_url = resolveCdpWebSocketUrl(cdp_url, "remote cdp_url")
        self.launched = {"cdp_url": cdp_url, "close": lambda: None}
        return self.launched
