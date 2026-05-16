from __future__ import annotations

import glob
import json
import os
import re
import select
import signal
import shutil
import subprocess
import sys
import tempfile
import time
import urllib.request
from pathlib import Path
from typing import Protocol, cast

from ..launcher.BrowserLauncher import (
    DEFAULT_CHROME_READY_POLL_INTERVAL_MS,
    DEFAULT_CHROME_READY_TIMEOUT_MS,
    BrowserLaunchOptions,
    BrowserLauncher,
    LaunchedBrowser,
)


class LocalBrowserLauncher(BrowserLauncher):
    @staticmethod
    def findChromeBinary(explicit: str | None = None) -> str:
        candidates = [explicit, os.environ.get("CHROME_PATH"), *_candidate_paths()]
        for candidate in candidates:
            if candidate and Path(candidate).exists():
                return str(candidate)
        tried = ", ".join(str(candidate) for candidate in candidates if candidate)
        raise RuntimeError(f"No Chrome/Chromium binary found. Tried: {tried}. Set CHROME_PATH or pass executable_path.")

    @staticmethod
    def freePort() -> int:
        return _free_port()

    def launch(self, options: BrowserLaunchOptions | None = None) -> LaunchedBrowser:
        merged = cast(BrowserLaunchOptions, {**self.options, **dict(options or {})})
        executable_path = self.findChromeBinary(merged.get("executable_path"))
        use_pipe = merged.get("remote_debugging") == "pipe"
        use_loopback_cdp = (not use_pipe) or bool(merged.get("loopback_cdp")) or merged.get("port") is not None
        port = int(merged.get("port") or LocalBrowserLauncher.freePort()) if use_loopback_cdp else None
        temp_profile_dir: tempfile.TemporaryDirectory[str] | None = None
        profile_dir = merged.get("user_data_dir")
        if not profile_dir:
            temp_profile_dir = tempfile.TemporaryDirectory(prefix="modcdp.")
            profile_dir = temp_profile_dir.name
        cleanup_profile_dir = str(profile_dir) if merged.get("cleanup_user_data_dir") else None
        args = [
            "--enable-unsafe-extension-debugging",
            "--remote-allow-origins=*",
            "--no-first-run",
            "--no-default-browser-check",
            "--disable-default-apps",
            "--disable-dev-shm-usage",
            "--disable-background-networking",
            "--disable-backgrounding-occluded-windows",
            "--disable-renderer-backgrounding",
            "--disable-background-timer-throttling",
            "--disable-sync",
            "--disable-features=DisableLoadExtensionCommandLineSwitch",
            "--password-store=basic",
            "--use-mock-keychain",
            "--disable-gpu",
            f"--user-data-dir={profile_dir}",
            "--remote-debugging-address=127.0.0.1" if use_loopback_cdp else None,
            f"--remote-debugging-port={port}" if use_loopback_cdp else None,
            "--remote-debugging-pipe" if use_pipe else None,
        ]
        args = [arg for arg in args if arg is not None]
        default_headless = sys.platform.startswith("linux") and not os.environ.get("DISPLAY")
        if merged.get("headless", default_headless):
            args.append("--headless=new")
        default_sandbox = not sys.platform.startswith("linux")
        if merged.get("sandbox", default_sandbox) is False:
            args.append("--no-sandbox")
        args.extend(list(merged.get("args") or []))
        args.extend(list(merged.get("extra_args") or []))
        args.append("about:blank")
        if use_pipe:
            parent_read, child_write = os.pipe()
            child_read, parent_write = os.pipe()
            parent_read = _move_fd_if_needed(parent_read, {3, 4})
            parent_write = _move_fd_if_needed(parent_write, {3, 4})
            child_read = _move_fd_if_needed(child_read, {3, 4})
            child_write = _move_fd_if_needed(child_write, {3, 4})
            process = _spawn_chrome_with_pipe_fds(executable_path, args, child_read, child_write)
            os.close(child_read)
            os.close(child_write)
            pipe_read = os.fdopen(parent_read, "rb", buffering=0)
            pipe_write = os.fdopen(parent_write, "wb", buffering=0)
            try:
                _wait_for_pipe_ready(pipe_read, pipe_write, int(merged.get("chrome_ready_timeout_ms") or DEFAULT_CHROME_READY_TIMEOUT_MS))
                loopback_cdp_url = (
                    _wait_for_cdp_websocket_url(
                        f"http://127.0.0.1:{port}",
                        int(merged.get("chrome_ready_timeout_ms") or DEFAULT_CHROME_READY_TIMEOUT_MS),
                        int(merged.get("chrome_ready_poll_interval_ms") or DEFAULT_CHROME_READY_POLL_INTERVAL_MS),
                    )
                    if port is not None
                    else None
                )
            except Exception:
                pipe_read.close()
                pipe_write.close()
                _close(process, temp_profile_dir, cleanup_profile_dir=cleanup_profile_dir)
                raise
            launched: LaunchedBrowser = {
                "cdp_url": f"pipe://{process.pid}",
                "profile_dir": profile_dir,
                "pipe_read": pipe_read,
                "pipe_write": pipe_write,
                "close": lambda: _close(
                    process,
                    temp_profile_dir,
                    pipe_read,
                    pipe_write,
                    cleanup_profile_dir=cleanup_profile_dir,
                ),
            }
            if loopback_cdp_url:
                launched["loopback_cdp_url"] = loopback_cdp_url
            self.launched = launched
            return self.launched

        process = subprocess.Popen(
            [executable_path, *args],
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            start_new_session=not sys.platform.startswith("win"),
        )
        cdp_url = f"http://127.0.0.1:{port}"
        timeout_s = int(merged.get("chrome_ready_timeout_ms") or DEFAULT_CHROME_READY_TIMEOUT_MS) / 1000
        poll_s = int(merged.get("chrome_ready_poll_interval_ms") or DEFAULT_CHROME_READY_POLL_INTERVAL_MS) / 1000
        deadline = time.time() + timeout_s
        while time.time() < deadline:
            try:
                with urllib.request.urlopen(f"{cdp_url}/json/version", timeout=0.5) as response:
                    version = json.loads(response.read())
                    self.launched = {
                        # cdp_url is resolved from the HTTP discovery endpoint before returning.
                        "cdp_url": version.get("webSocketDebuggerUrl") or cdp_url,
                        "loopback_cdp_url": version.get("webSocketDebuggerUrl") or cdp_url,
                        "profile_dir": profile_dir,
                        "close": lambda: _close(process, temp_profile_dir, cleanup_profile_dir=cleanup_profile_dir),
                    }
                    return self.launched
            except Exception:
                time.sleep(poll_s)
        _close(process, temp_profile_dir, cleanup_profile_dir=cleanup_profile_dir)
        raise RuntimeError(f"Chrome at {cdp_url} did not become ready within {timeout_s}s")


def _newest_first(candidates: list[str]) -> list[str]:
    def score(candidate: str) -> tuple[int, float, str]:
        numbers = [int(part) for part in re.findall(r"\d+", candidate)]
        version = max(numbers) if numbers else 0
        try:
            mtime = Path(candidate).stat().st_mtime
        except OSError:
            mtime = 0.0
        return (-version, -mtime, candidate)

    return sorted(dict.fromkeys(candidates), key=score)


def _chrome_for_testing_candidates() -> list[str]:
    home = Path.home()
    if sys.platform == "darwin":
        patterns = [
            str(home / "Library/Caches/ms-playwright/chromium-*/chrome-mac*/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing"),
            str(home / "Library/Caches/ms-playwright/chromium-*/chrome-mac*/Chromium.app/Contents/MacOS/Chromium"),
            str(home / "Library/Caches/puppeteer/chrome/mac*-*/chrome-mac*/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing"),
        ]
    elif sys.platform.startswith("win"):
        local_app_data = Path(os.environ.get("LOCALAPPDATA") or home / "AppData/Local")
        patterns = [
            str(local_app_data / "ms-playwright/chromium-*/chrome-win*/chrome.exe"),
            str(home / ".cache/puppeteer/chrome/win*-*/chrome-win*/chrome.exe"),
        ]
    else:
        patterns = [
            str(home / ".cache/ms-playwright/chromium-*/chrome-linux*/chrome"),
            "/opt/pw-browsers/chromium-*/chrome-linux*/chrome",
            str(home / ".cache/puppeteer/chrome/linux-*/chrome-linux*/chrome"),
        ]
    return _newest_first([match for pattern in patterns for match in glob.glob(pattern)])


def _candidate_paths() -> list[str]:
    home = Path.home()
    program_files = [value for value in [os.environ.get("PROGRAMFILES"), os.environ.get("PROGRAMFILES(X86)")] if value]
    if sys.platform == "darwin":
        canary = ["/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary"]
        stock = ["/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"]
    elif sys.platform.startswith("win"):
        local_app_data = Path(os.environ.get("LOCALAPPDATA") or home / "AppData/Local")
        canary = [
            str(local_app_data / "Google/Chrome SxS/Application/chrome.exe"),
            *[str(Path(base) / "Google/Chrome SxS/Application/chrome.exe") for base in program_files],
        ]
        stock = [
            *[str(Path(base) / "Google/Chrome/Application/chrome.exe") for base in program_files],
            str(local_app_data / "Google/Chrome/Application/chrome.exe"),
        ]
    else:
        chromium = ["/usr/bin/chromium", "/usr/bin/chromium-browser"]
        canary = ["/usr/bin/google-chrome-canary", "/usr/bin/google-chrome-unstable", "/opt/google/chrome-unstable/chrome"]
        stock = ["/usr/bin/google-chrome-stable", "/usr/bin/google-chrome", "/opt/google/chrome/chrome"]
        return [*chromium, *canary, *_chrome_for_testing_candidates(), *stock]
    return [*canary, *_chrome_for_testing_candidates(), *stock]


def _move_fd_if_needed(fd: int, reserved: set[int]) -> int:
    if fd not in reserved:
        return fd
    moved = os.dup(fd)
    while moved in reserved:
        next_fd = os.dup(fd)
        os.close(moved)
        moved = next_fd
    os.close(fd)
    return moved


def _spawn_chrome_with_pipe_fds(executable_path: str, args: list[str], child_read: int, child_write: int) -> _ChromeProcess:
    def map_pipe_fds() -> None:
        os.dup2(child_read, 3)
        os.dup2(child_write, 4)
        os.close(child_read)
        os.close(child_write)

    return subprocess.Popen(
        [executable_path, *args],
        stdin=subprocess.DEVNULL,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        close_fds=sys.platform.startswith("win"),
        preexec_fn=None if sys.platform.startswith("win") else map_pipe_fds,
        start_new_session=not sys.platform.startswith("win"),
    )


class _ChromeProcess(Protocol):
    pid: int

    def terminate(self) -> None: ...

    def kill(self) -> None: ...

    def wait(self, timeout: float | None = None) -> int | None: ...


def _wait_for_pipe_ready(pipe_read, pipe_write, timeout_ms: int) -> None:
    ready_id = 1
    pipe_write.write(json.dumps({"id": ready_id, "method": "Browser.getVersion", "params": {}}).encode() + b"\0")
    pipe_write.flush()
    deadline = time.time() + timeout_ms / 1000
    buffer = b""
    while time.time() < deadline:
        ready, _, _ = select.select([pipe_read], [], [], max(0.0, min(0.1, deadline - time.time())))
        if not ready:
            continue
        chunk = pipe_read.read(1)
        if not chunk:
            time.sleep(0.01)
            continue
        buffer += chunk
        if b"\0" not in buffer:
            continue
        raw, buffer = buffer.split(b"\0", 1)
        if not raw:
            continue
        message = json.loads(raw.decode())
        if message.get("id") != ready_id:
            continue
        if message.get("error"):
            raise RuntimeError(message["error"].get("message") or "Browser.getVersion failed over pipe")
        return
    raise RuntimeError(f"Chrome remote-debugging pipe did not respond within {timeout_ms}ms")


def _wait_for_cdp_websocket_url(cdp_url: str, timeout_ms: int, poll_interval_ms: int) -> str:
    deadline = time.time() + timeout_ms / 1000
    poll_s = poll_interval_ms / 1000
    while time.time() < deadline:
        try:
            with urllib.request.urlopen(f"{cdp_url}/json/version", timeout=0.5) as response:
                version = json.loads(response.read())
                websocket_url = version.get("webSocketDebuggerUrl")
                if websocket_url:
                    return str(websocket_url)
        except Exception:
            pass
        time.sleep(poll_s)
    raise RuntimeError(f"Chrome at {cdp_url} did not expose a WebSocket CDP URL within {timeout_ms}ms")


def _close(
    process: _ChromeProcess,
    temp_profile_dir: tempfile.TemporaryDirectory[str] | None,
    pipe_read=None,
    pipe_write=None,
    cleanup_profile_dir: str | None = None,
) -> None:
    for pipe in (pipe_read, pipe_write):
        try:
            if pipe is not None:
                pipe.close()
        except Exception:
            pass
    _signal_process(process, signal.SIGTERM)
    try:
        process.wait(timeout=2)
    except subprocess.TimeoutExpired:
        _signal_process(process, signal.SIGKILL)
        process.wait(timeout=2)
    if temp_profile_dir:
        temp_profile_dir.cleanup()
    if cleanup_profile_dir:
        _remove_profile_dir(cleanup_profile_dir)


def _signal_process(process: _ChromeProcess, sig: signal.Signals) -> None:
    if not sys.platform.startswith("win"):
        try:
            os.killpg(process.pid, sig)
            return
        except ProcessLookupError:
            return
        except Exception:
            pass
    if sig == signal.SIGKILL:
        process.kill()
    else:
        process.terminate()


def _remove_profile_dir(profile_dir: str) -> None:
    for attempt in range(5):
        shutil.rmtree(profile_dir, ignore_errors=True)
        if not Path(profile_dir).exists():
            return
        time.sleep(0.1 * (attempt + 1))
    shutil.rmtree(profile_dir, ignore_errors=True)


def _free_port() -> int:
    import socket

    sock = socket.socket()
    sock.bind(("127.0.0.1", 0))
    try:
        return int(sock.getsockname()[1])
    finally:
        sock.close()
