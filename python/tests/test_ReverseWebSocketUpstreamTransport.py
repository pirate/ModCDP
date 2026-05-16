from __future__ import annotations

import json
import glob
import os
import re
import socket
import sys
import threading
import time
import unittest
from pathlib import Path
from queue import Queue
from typing import cast

from websocket import create_connection

from modcdp import ModCDPClient
from modcdp.transport.ReverseWebSocketUpstreamTransport import ReverseWebSocketUpstreamTransport

ROOT = Path(__file__).resolve().parents[2]
EXTENSION_PATH = ROOT / "dist" / "extension"
REVERSEWS_TEST_BROWSER_PATH = None


class ReverseWebSocketUpstreamTransportTests(unittest.TestCase):
    def test_config_owns_bind_updates_and_wait_timeout(self) -> None:
        transport = ReverseWebSocketUpstreamTransport({
            "upstream_reversews_bind": "127.0.0.1:29292",
            "upstream_reversews_wait_timeout_ms": 10,
        })
        self.assertEqual(transport.url, "ws://127.0.0.1:29292")
        self.assertEqual(transport.getInjectorConfig(), {})
        self.assertIs(
            transport.update({
                "upstream_reversews_bind": "127.0.0.1:29293",
                "upstream_reversews_wait_timeout_ms": 5,
            }),
            transport,
        )
        self.assertEqual(transport.url, "ws://127.0.0.1:29293")
        self.assertEqual(transport.getInjectorConfig(), {})
        with self.assertRaisesRegex(RuntimeError, "Timed out waiting 5ms"):
            transport.waitForPeer()

    def test_close_rejects_pending_peer_waits(self) -> None:
        reverse_port = _free_port()
        transport = ReverseWebSocketUpstreamTransport({"upstream_reversews_bind": f"127.0.0.1:{reverse_port}", "upstream_reversews_wait_timeout_ms": 5_000})
        result: Queue[BaseException | None] = Queue()

        def wait_for_peer() -> None:
            try:
                transport.waitForPeer()
            except BaseException as error:
                result.put(error)
                return
            result.put(None)

        thread = threading.Thread(target=wait_for_peer, daemon=True)
        thread.start()
        time.sleep(0.05)
        transport.close()
        thread.join(timeout=1)

        error = result.get(timeout=1)
        self.assertIsInstance(error, RuntimeError)
        self.assertRegex(
            str(error),
            rf"Reverse websocket transport at ws://127\.0\.0\.1:{reverse_port} closed before a peer connected",
        )

    def test_close_resets_peer_wait_state(self) -> None:
        reverse_port = _free_port()
        transport = ReverseWebSocketUpstreamTransport({"upstream_reversews_bind": f"127.0.0.1:{reverse_port}", "upstream_reversews_wait_timeout_ms": 5})
        transport.connect()
        url = transport.url
        if url is None:
            self.fail("reverse transport url was not set")
        peer = create_connection(url, timeout=10)
        peer.send(json.dumps({"type": "modcdp.reverse.hello", "role": "test-peer", "version": 1}))

        try:
            transport.waitForPeer()
            self.assertEqual(
                transport.peer_info,
                {"type": "modcdp.reverse.hello", "role": "test-peer", "version": 1},
            )
            transport.close()

            with self.assertRaisesRegex(RuntimeError, "Timed out waiting 5ms"):
                transport.waitForPeer()
            self.assertIsNone(transport.peer_info)
        finally:
            peer.close()
            transport.close()

    def test_waits_again_after_peer_disconnects(self) -> None:
        reverse_port = _free_port()
        transport = ReverseWebSocketUpstreamTransport({"upstream_reversews_bind": f"127.0.0.1:{reverse_port}", "upstream_reversews_wait_timeout_ms": 5})
        transport.connect()
        url = transport.url
        if url is None:
            self.fail("reverse transport url was not set")
        peer = create_connection(url, timeout=10)
        peer.send(json.dumps({"type": "modcdp.reverse.hello", "role": "test-peer", "version": 1}))

        try:
            transport.waitForPeer()
            peer.close()
            _wait_until(lambda: transport.socket is None)

            with self.assertRaisesRegex(RuntimeError, "Timed out waiting 5ms"):
                transport.waitForPeer()
        finally:
            peer.close()
            transport.close()

    def test_accepts_replacement_peer_after_disconnect(self) -> None:
        reverse_port = _free_port()
        transport = ReverseWebSocketUpstreamTransport({"upstream_reversews_bind": f"127.0.0.1:{reverse_port}", "upstream_reversews_wait_timeout_ms": 500})
        transport.connect()
        url = transport.url
        if url is None:
            self.fail("reverse transport url was not set")
        first_peer = create_connection(url, timeout=10)
        first_peer.send(json.dumps({"type": "modcdp.reverse.hello", "role": "first-peer", "version": 1}))

        try:
            transport.waitForPeer()
            first_peer.close()
            _wait_until(lambda: transport.socket is None)

            second_peer = create_connection(url, timeout=10)
            second_peer.send(json.dumps({"type": "modcdp.reverse.hello", "role": "second-peer", "version": 1}))
            try:
                transport.waitForPeer()
                self.assertEqual((transport.peer_info or {}).get("role"), "second-peer")
            finally:
                second_peer.close()
        finally:
            first_peer.close()
            transport.close()

    def test_accepts_real_extension_reverse_connection_and_routes_cdp_through_chrome_debugger(self) -> None:
        cdp = ModCDPClient(
            launcher={
                "launcher_mode": "local",
                "launcher_options": {
                    "headless": sys.platform.startswith("linux") and not os.environ.get("DISPLAY"),
                    # Reversews is browser -> client only. After explicit CHROME_PATH and
                    # CI /usr/bin/chromium, these tests use Chrome for Testing because
                    # Canary rejects --load-extension in this local test path.
                    "executable_path": reversews_test_browser_path(),
                },
            },
            upstream={"upstream_mode": "reversews"},
            injector={
                "injector_mode": "auto",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
                "injector_service_worker_probe_timeout_ms": 1_000,
            },
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.transport.mode if cdp.transport else None, "reversews")
            self.assertEqual(cdp.upstream_endpoint_kind, "modcdp_server")
            self.assertIsInstance(cdp.transport, ReverseWebSocketUpstreamTransport)
            transport = cast(ReverseWebSocketUpstreamTransport, cdp.transport)
            self.assertEqual(transport.url, "ws://127.0.0.1:29292")
            self.assertEqual(
                transport.peer_info.get("extension_id") if transport.peer_info else None,
                "mdedooklbnfejodmnhmkdpkaedafkehf",
            )
            evaluated = cdp.send("Runtime.evaluate", {"expression": "location.href", "returnByValue": True})
            self.assertEqual(evaluated["result"]["value"], "about:blank")
            time.sleep(1.5)
            second_evaluated = cdp.send("Runtime.evaluate", {"expression": "document.readyState", "returnByValue": True})
            self.assertEqual(second_evaluated["result"]["value"], "complete")
        finally:
            cdp.close()


def _free_port() -> int:
    sock = socket.socket()
    sock.bind(("127.0.0.1", 0))
    try:
        return int(sock.getsockname()[1])
    finally:
        sock.close()


def _wait_until(predicate, timeout_s: float = 2.0) -> None:
    deadline = time.monotonic() + timeout_s
    while time.monotonic() < deadline:
        if predicate():
            return
        time.sleep(0.02)
    raise AssertionError("timed out waiting for condition")


def reversews_test_browser_path() -> str:
    global REVERSEWS_TEST_BROWSER_PATH
    if REVERSEWS_TEST_BROWSER_PATH is not None:
        return REVERSEWS_TEST_BROWSER_PATH
    explicit_candidates = [
        os.environ.get("CHROME_PATH"),
        "/usr/bin/chromium" if sys.platform.startswith("linux") else None,
    ]
    for candidate in explicit_candidates:
        if candidate and Path(candidate).exists():
            REVERSEWS_TEST_BROWSER_PATH = candidate
            return candidate
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
    candidates = newest_first([match for pattern in patterns for match in glob.glob(pattern)])
    if candidates:
        REVERSEWS_TEST_BROWSER_PATH = candidates[0]
        return candidates[0]
    raise RuntimeError("Reversews tests require CHROME_PATH, /usr/bin/chromium, or Chrome for Testing.")


def newest_first(candidates: list[str]) -> list[str]:
    return sorted(dict.fromkeys(candidates), key=path_score)


def path_score(candidate: str) -> tuple[int, float, str]:
    numbers = [int(part) for part in re.findall(r"\d+", candidate)]
    version = max(numbers) if numbers else 0
    try:
        mtime = Path(candidate).stat().st_mtime
    except OSError:
        mtime = 0.0
    return (-version, -mtime, candidate)


if __name__ == "__main__":
    unittest.main()
