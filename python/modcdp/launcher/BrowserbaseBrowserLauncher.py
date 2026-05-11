from __future__ import annotations

import json
import os
import urllib.error
import urllib.request
from typing import Any
from websocket import create_connection

from ..launcher.BrowserLauncher import BrowserLaunchOptions, BrowserLauncher, LaunchedBrowser


DEFAULT_BROWSERBASE_BASE_URL = "https://api.browserbase.com"
DEFAULT_BROWSERBASE_VIEWPORT = {"width": 1288, "height": 711}


class BrowserbaseBrowserLauncher(BrowserLauncher):
    def launch(self, options: BrowserLaunchOptions | None = None) -> LaunchedBrowser:
        merged = {**self.options, **dict(options or {})}
        browserbase_api_key = _first_string(merged.get("browserbase_api_key"), os.environ.get("BROWSERBASE_API_KEY"))
        if not browserbase_api_key:
            raise RuntimeError("launcher.launcher_mode=bb requires BROWSERBASE_API_KEY or launcher.launcher_options.browserbase_api_key.")

        base_url = _first_string(
            merged.get("browserbase_base_url"),
            os.environ.get("BROWSERBASE_BASE_URL"),
        ) or DEFAULT_BROWSERBASE_BASE_URL
        resume_session_id = _first_string(merged.get("browserbase_session_id"))
        keep_alive = _first_bool(merged.get("browserbase_keep_alive")) or False
        close_session_on_close = _first_bool(merged.get("browserbase_close_session_on_close"))
        if close_session_on_close is None:
            close_session_on_close = not keep_alive

        created_session = False
        if resume_session_id:
            session = _browserbase_request(
                base_url=base_url,
                browserbase_api_key=browserbase_api_key,
                method="GET",
                pathname=f"/v1/sessions/{resume_session_id}",
            )
        else:
            session_create_params = _object_value(merged.get("browserbase_session_create_params"))
            browser_settings = {
                **_object_value(session_create_params.get("browserSettings")),
                **_object_value(merged.get("browserbase_browser_settings")),
            }
            user_metadata = {
                **_object_value(session_create_params.get("userMetadata")),
                **_object_value(merged.get("browserbase_user_metadata")),
            }
            extension_id = _first_string(
                merged.get("injector_extension_id"),
                session_create_params.get("extensionId"),
                _object_value(session_create_params.get("browserSettings")).get("extensionId"),
            )
            region = _first_string(merged.get("region"), session_create_params.get("region"))
            viewport = _object_value(browser_settings.get("viewport"))
            body: dict[str, Any] = {
                **session_create_params,
                **({"keepAlive": True} if keep_alive else {}),
                **({"region": region} if region else {}),
                **({"timeout": merged.get("timeout")} if isinstance(merged.get("timeout"), int) else {}),
                **({"extensionId": extension_id} if extension_id else {}),
                "browserSettings": {
                    **browser_settings,
                    **({"extensionId": extension_id} if extension_id else {}),
                    "viewport": viewport if viewport.get("width") else DEFAULT_BROWSERBASE_VIEWPORT,
                },
                "userMetadata": {
                    **user_metadata,
                    "modcdp": "true",
                },
            }
            session = _browserbase_request(
                base_url=base_url,
                browserbase_api_key=browserbase_api_key,
                method="POST",
                pathname="/v1/sessions",
                body=body,
            )
            created_session = True

        session_id = _first_string(session.get("id"))
        connect_url = _first_string(session.get("connectUrl"))
        if not session_id or not connect_url:
            raise RuntimeError("Browserbase session creation returned an unexpected shape.")

        closed = False

        def close() -> None:
            nonlocal closed
            if closed:
                return
            closed = True
            if not created_session or not close_session_on_close:
                return
            _close_browser_cdp(connect_url)
            try:
                _browserbase_request(
                    base_url=base_url,
                    browserbase_api_key=browserbase_api_key,
                    method="POST",
                    pathname=f"/v1/sessions/{session_id}",
                    body={"status": "REQUEST_RELEASE"},
                )
            except Exception:
                pass

        self.launched = {
            # Browserbase connectUrl is already a WebSocket CDP endpoint.
            "cdp_url": connect_url,
            "browserbase_session_id": session_id,
            "browserbase_session_url": f"https://www.browserbase.com/sessions/{session_id}",
            "browserbase_debug_url": _first_string(session.get("debuggerUrl"), session.get("debuggerFullscreenUrl")),
            "close": close,
        }
        return self.launched


def _browserbase_request(
    *,
    base_url: str,
    browserbase_api_key: str,
    method: str,
    pathname: str,
    body: dict[str, Any] | None = None,
) -> dict[str, Any]:
    request = urllib.request.Request(
        _browserbase_url(base_url, pathname),
        method=method,
        headers={
            "content-type": "application/json",
            "x-bb-api-key": browserbase_api_key,
        },
        data=json.dumps(body).encode() if body is not None else None,
    )
    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            return json.loads(response.read())
    except urllib.error.HTTPError as error:
        text = error.read().decode(errors="replace")
        raise RuntimeError(f"Browserbase {method} {pathname} -> {error.code}{f': {text}' if text else ''}") from error


def _browserbase_url(base_url: str, pathname: str) -> str:
    return f"{base_url.rstrip('/')}/{pathname.lstrip('/')}"


def _close_browser_cdp(cdp_url: str | None) -> None:
    if not cdp_url or not cdp_url.startswith(("ws://", "wss://")):
        return
    try:
        ws = create_connection(cdp_url, timeout=2)
        try:
            ws.send(json.dumps({"id": 1, "method": "Browser.close", "params": {}}))
            ws.settimeout(0.5)
            try:
                ws.recv()
            except Exception:
                pass
        finally:
            ws.close()
    except Exception:
        pass


def _first_string(*values: Any) -> str | None:
    for value in values:
        if isinstance(value, str) and value.strip():
            return value.strip()
    return None


def _first_bool(*values: Any) -> bool | None:
    for value in values:
        if isinstance(value, bool):
            return value
    return None


def _object_value(value: Any) -> dict[str, Any]:
    return value if isinstance(value, dict) else {}
