from __future__ import annotations

import json
import tempfile
import unittest
from pathlib import Path

from websocket import create_connection

from modcdp.launcher.LocalBrowserLauncher import LocalBrowserLauncher


class LocalBrowserLauncherTests(unittest.TestCase):
    def test_class_helpers_match_ts_surface(self) -> None:
        self.assertIsInstance(LocalBrowserLauncher.findChromeBinary(), str)
        self.assertIsInstance(LocalBrowserLauncher.freePort(), int)

    def test_launches_real_browser_and_speaks_cdp(self) -> None:
        with tempfile.TemporaryDirectory(prefix="modcdp-python-local-profile-") as user_data_dir:
            chrome = LocalBrowserLauncher(
                {
                    "headless": True,
                    "chrome_ready_timeout_ms": 45_000,
                    "chrome_ready_poll_interval_ms": 50,
                }
            ).launch({"user_data_dir": user_data_dir, "args": ["--window-size=900,700"]})
            cdp_url = chrome["cdp_url"]
            if cdp_url is None:
                raise AssertionError("expected launcher to return cdp_url")
            ws = create_connection(cdp_url, timeout=10)

            try:
                self.assertEqual(chrome.get("profile_dir"), user_data_dir)
                ws.send(json.dumps({"id": 1, "method": "Browser.getVersion", "params": {}}))
                version = json.loads(ws.recv())
                self.assertEqual(version["id"], 1)
                self.assertIn("Chrome", version["result"]["product"])
                self.assertIsInstance(version["result"]["protocolVersion"], str)
            finally:
                ws.close()
                chrome["close"]()

            self.assertTrue(Path(user_data_dir).exists())

    def test_cleanup_user_data_dir_removes_explicit_profile(self) -> None:
        user_data_dir = tempfile.mkdtemp(prefix="modcdp-python-local-profile-")
        chrome = LocalBrowserLauncher(
            {
                "headless": True,
                "chrome_ready_timeout_ms": 45_000,
            }
        ).launch({"user_data_dir": user_data_dir, "cleanup_user_data_dir": True})

        try:
            self.assertEqual(chrome.get("profile_dir"), user_data_dir)
        finally:
            chrome["close"]()
        self.assertFalse(Path(user_data_dir).exists())

    def test_launches_real_browser_over_remote_debugging_pipe(self) -> None:
        chrome = LocalBrowserLauncher(
            {
                "headless": True,
                "remote_debugging": "pipe",
                "chrome_ready_timeout_ms": 45_000,
            }
        ).launch()
        pipe_read = chrome.get("pipe_read")
        pipe_write = chrome.get("pipe_write")
        if pipe_read is None or pipe_write is None:
            raise AssertionError("expected launcher to return pipe handles")

        try:
            self.assertRegex(chrome["cdp_url"] or "", r"^pipe://\d+$")
            self.assertNotIn("loopback_cdp_url", chrome)
            pipe_write.write(json.dumps({"id": 10, "method": "Browser.getVersion", "params": {}}).encode() + b"\0")
            pipe_write.flush()
            response = _read_pipe_message(pipe_read)
            self.assertEqual(response["id"], 10)
            self.assertIn("Chrome", response["result"]["product"])
        finally:
            chrome["close"]()

    def test_launches_pipe_browser_with_auxiliary_loopback_only_when_requested(self) -> None:
        chrome = LocalBrowserLauncher(
            {
                "headless": True,
                "remote_debugging": "pipe",
                "loopback_cdp": True,
                "chrome_ready_timeout_ms": 45_000,
            }
        ).launch()
        loopback_cdp_url = chrome.get("loopback_cdp_url")
        if not isinstance(loopback_cdp_url, str):
            raise AssertionError("expected launcher to return loopback_cdp_url")
        ws = create_connection(loopback_cdp_url, timeout=10)

        try:
            self.assertRegex(chrome["cdp_url"] or "", r"^pipe://\d+$")
            self.assertRegex(loopback_cdp_url, r"^ws://127\.0\.0\.1:\d+/")
            ws.send(json.dumps({"id": 1, "method": "Browser.getVersion", "params": {}}))
            version = json.loads(ws.recv())
            self.assertEqual(version["id"], 1)
            self.assertIn("Chrome", version["result"]["product"])
        finally:
            ws.close()
            chrome["close"]()


def _read_pipe_message(pipe_read) -> dict:
    buffer = b""
    while True:
        chunk = pipe_read.read(1)
        if not chunk:
            raise AssertionError("pipe closed before CDP response")
        buffer += chunk
        if b"\0" not in buffer:
            continue
        raw, _ = buffer.split(b"\0", 1)
        return json.loads(raw.decode())


if __name__ == "__main__":
    unittest.main()
