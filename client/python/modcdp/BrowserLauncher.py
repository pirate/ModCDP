from __future__ import annotations

from collections.abc import Callable
from typing import Any, TypedDict, cast

from typing_extensions import NotRequired


class BrowserLaunchOptions(TypedDict, total=False):
    executable_path: str | None
    port: int | None
    user_data_dir: str | None
    headless: bool
    sandbox: bool
    args: list[str]
    extra_args: list[str]
    remote_debugging: str
    cleanup_user_data_dir: bool
    chrome_ready_timeout_ms: int
    chrome_ready_poll_interval_ms: int
    cdp_url: str | None
    ws_url: str | None
    browserbase_api_key: str | None
    project_id: str | None
    browserbase_project_id: str | None
    base_url: str | None
    browserbase_base_url: str | None
    session_id: str | None
    resume_session_id: str | None
    keep_alive: bool
    close_session_on_close: bool
    region: str | None
    timeout: int | None
    extension_id: str | None
    browser_settings: dict[str, Any] | None
    user_metadata: dict[str, Any] | None
    session_create_params: dict[str, Any] | None
    browserbase_session_create_params: dict[str, Any] | None


class LaunchedBrowser(TypedDict):
    cdp_url: str | None
    ws_url: str | None
    close: Callable[[], Any]
    profile_dir: NotRequired[str | None]
    pipe_read: NotRequired[Any]
    pipe_write: NotRequired[Any]
    browserbase_session_id: NotRequired[str | None]
    browserbase_session_url: NotRequired[str | None]
    browserbase_debug_url: NotRequired[str | None]


DEFAULT_CHROME_READY_TIMEOUT_MS = 45_000
DEFAULT_CHROME_READY_POLL_INTERVAL_MS = 100


class BrowserLauncher:
    launched: LaunchedBrowser | None

    def __init__(self, options: BrowserLaunchOptions | None = None) -> None:
        self.options = cast(BrowserLaunchOptions, dict(options or {}))
        self.launched = None

    def update(self, config: BrowserLaunchOptions | None = None) -> "BrowserLauncher":
        config = cast(BrowserLaunchOptions, dict(config or {}))
        self.options = cast(
            BrowserLaunchOptions,
            {
                **self.options,
                **config,
                **({"args": merge_chrome_args(self.options.get("args"), config["args"])} if "args" in config else {}),
                **(
                    {"extra_args": merge_chrome_args(self.options.get("extra_args"), config["extra_args"])}
                    if "extra_args" in config
                    else {}
                ),
            },
        )
        return self

    def getTransportConfig(self) -> dict[str, Any]:
        return {
            "cdp_url": (self.launched or {}).get("cdp_url") or self.options.get("cdp_url"),
            "ws_url": (self.launched or {}).get("ws_url") or self.options.get("ws_url"),
            "user_data_dir": (self.launched or {}).get("profile_dir") or self.options.get("user_data_dir"),
            "pipe_read": (self.launched or {}).get("pipe_read"),
            "pipe_write": (self.launched or {}).get("pipe_write"),
        }

    def getInjectorConfig(self) -> dict[str, Any]:
        return {
            "browserbase_api_key": self.options.get("browserbase_api_key"),
            "base_url": self.options.get("base_url"),
            "browserbase_base_url": self.options.get("browserbase_base_url"),
            "extension_id": self.options.get("extension_id"),
        }

    def launch(self, options: BrowserLaunchOptions | None = None) -> LaunchedBrowser:
        raise NotImplementedError(f"{type(self).__name__}.launch is not implemented.")


def merge_chrome_args(existing: list[str] | None = None, incoming: list[str] | None = None) -> list[str]:
    args = [*(existing or []), *(incoming or [])]
    load_extension_paths: list[str] = []
    merged: list[str] = []
    for arg in args:
        if not arg.startswith("--load-extension="):
            merged.append(arg)
            continue
        for extension_path in arg[len("--load-extension="):].split(","):
            if extension_path and extension_path not in load_extension_paths:
                load_extension_paths.append(extension_path)
    if load_extension_paths:
        first_url_index = next((index for index, arg in enumerate(merged) if not arg.startswith("-")), -1)
        load_extension_arg = f"--load-extension={','.join(load_extension_paths)}"
        if first_url_index == -1:
            merged.append(load_extension_arg)
        else:
            merged.insert(first_url_index, load_extension_arg)
    return merged
