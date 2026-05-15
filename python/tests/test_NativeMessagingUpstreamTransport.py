from __future__ import annotations

import threading
import time
import unittest
import glob
import os
import re
import socket
import sys
import tempfile
from pathlib import Path
from queue import Queue

from modcdp import ModCDPClient
from modcdp.transport.NativeMessagingUpstreamTransport import NativeMessagingUpstreamTransport

ROOT = Path(__file__).resolve().parents[2]
EXTENSION_PATH = ROOT / "dist" / "extension"
NATIVE_MESSAGING_TEST_BROWSER_PATH: str | None = None


class NativeMessagingUpstreamTransportTests(unittest.TestCase):
    def test_config_owns_manifest_host_wait_timeout_loopback_and_injector_config(self) -> None:
        transport = NativeMessagingUpstreamTransport(
            {
                "upstream_nativemessaging_manifest": "/tmp/modcdp-native-host.json",
                "upstream_nativemessaging_manifests": ["/tmp/modcdp-native-host-extra.json"],
                "upstream_nativemessaging_host_name": "com.modcdp.test",
                "injector_extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "upstream_nativemessaging_wait_timeout_ms": 10,
            }
        )
        self.assertEqual(transport.getInjectorConfig(), {"upstream_nativemessaging_host_name": "com.modcdp.test"})
        self.assertEqual(transport.getServerConfig(), {})
        self.assertIs(
            transport.update(
                {
                    "cdp_url": "ws://127.0.0.1:9222/devtools/browser/test",
                    "upstream_nativemessaging_manifests": [],
                    "upstream_nativemessaging_host_name": "com.modcdp.updated",
                    "upstream_nativemessaging_wait_timeout_ms": 5,
                }
            ),
            transport,
        )
        self.assertEqual(
            transport.getServerConfig(),
            {"server_loopback_cdp_url": "ws://127.0.0.1:9222/devtools/browser/test"},
        )
        self.assertEqual(transport.getInjectorConfig(), {"upstream_nativemessaging_host_name": "com.modcdp.updated"})
        self.assertFalse(transport.include_default_manifest_paths)
        transport.update({"upstream_nativemessaging_manifest": None})
        self.assertTrue(transport.include_default_manifest_paths)
        transport.update({"user_data_dir": "/tmp/modcdp-profile-one"})
        transport.update({"user_data_dir": "/tmp/modcdp-profile-one"})
        transport.update({"user_data_dir": "/tmp/modcdp-profile-two"})
        self.assertEqual(
            transport.upstream_nativemessaging_manifests,
            [
                str(Path("/tmp/modcdp-profile-two") / "NativeMessagingHosts" / "com.modcdp.updated.json"),
                str(Path("/tmp/modcdp-profile-two") / "Default" / "NativeMessagingHosts" / "com.modcdp.updated.json"),
            ],
        )
        with self.assertRaisesRegex(RuntimeError, r"Timed out waiting 5ms for native messaging host com\.modcdp\.updated"):
            transport.waitForPeer()

    def test_close_resets_peer_wait_state(self) -> None:
        upstream_nativemessaging_host_name = f"com.modcdp.close.reset.python.{os.getpid()}"
        transport = NativeMessagingUpstreamTransport({"upstream_nativemessaging_host_name": upstream_nativemessaging_host_name, "upstream_nativemessaging_wait_timeout_ms": 5})
        transport.connect()
        bound_port = transport.bound_port
        if bound_port is None:
            self.fail("native messaging transport did not bind a port")
        peer = socket.create_connection(("127.0.0.1", bound_port), timeout=10)

        try:
            transport.waitForPeer()
            transport.close()
            native_host_name_pattern = upstream_nativemessaging_host_name.replace(".", r"\.")
            with self.assertRaisesRegex(
                RuntimeError,
                rf"Timed out waiting 5ms for native messaging host {native_host_name_pattern}",
            ):
                transport.waitForPeer()
        finally:
            peer.close()
            transport.close()

    def test_waits_again_after_peer_disconnects(self) -> None:
        upstream_nativemessaging_host_name = f"com.modcdp.disconnect.reset.python.{os.getpid()}"
        transport = NativeMessagingUpstreamTransport({"upstream_nativemessaging_host_name": upstream_nativemessaging_host_name, "upstream_nativemessaging_wait_timeout_ms": 5})
        transport.connect()
        bound_port = transport.bound_port
        if bound_port is None:
            self.fail("native messaging transport did not bind a port")
        peer = socket.create_connection(("127.0.0.1", bound_port), timeout=10)

        try:
            transport.waitForPeer()
            peer.close()
            _wait_until(lambda: transport.socket is None)
            native_host_name_pattern = upstream_nativemessaging_host_name.replace(".", r"\.")
            with self.assertRaisesRegex(
                RuntimeError,
                rf"Timed out waiting 5ms for native messaging host {native_host_name_pattern}",
            ):
                transport.waitForPeer()
        finally:
            peer.close()
            transport.close()

    def test_accepts_replacement_peer_after_disconnect(self) -> None:
        upstream_nativemessaging_host_name = f"com.modcdp.replacement.python.{os.getpid()}"
        transport = NativeMessagingUpstreamTransport({"upstream_nativemessaging_host_name": upstream_nativemessaging_host_name, "upstream_nativemessaging_wait_timeout_ms": 500})
        transport.connect()
        bound_port = transport.bound_port
        if bound_port is None:
            self.fail("native messaging transport did not bind a port")
        first_peer = socket.create_connection(("127.0.0.1", bound_port), timeout=10)

        try:
            transport.waitForPeer()
            first_peer.close()
            _wait_until(lambda: transport.socket is None)

            second_peer = socket.create_connection(("127.0.0.1", bound_port), timeout=10)
            try:
                transport.waitForPeer()
            finally:
                second_peer.close()
        finally:
            first_peer.close()
            transport.close()

    def test_close_rejects_pending_peer_waits(self) -> None:
        transport = NativeMessagingUpstreamTransport(
            {
                "upstream_nativemessaging_host_name": "com.modcdp.close",
                "upstream_nativemessaging_wait_timeout_ms": 5_000,
            }
        )
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
            r"Native messaging transport for com\.modcdp\.close closed before a peer connected",
        )

    def test_installs_launch_profile_native_host_manifest_and_connects_to_real_extension(self) -> None:
        upstream_nativemessaging_host_name = "com.modcdp.bridge"
        temp_profile_dir = tempfile.TemporaryDirectory(prefix="modcdp.native.")
        cdp = ModCDPClient(
            launcher={
                "launcher_mode": "local",
                "launcher_options": {
                    "headless": True,
                    "user_data_dir": temp_profile_dir.name,
                    "cleanup_user_data_dir": True,
                    # Native messaging is browser -> client only. After explicit CHROME_PATH
                    # and CI /usr/bin/chromium, this test uses Chrome for Testing because
                    # Canary rejects --load-extension in this local test path.
                    "executable_path": extension_launch_flag_test_browser_path(),
                },
            },
            upstream={"upstream_mode": "nativemessaging", "upstream_nativemessaging_host_name": upstream_nativemessaging_host_name},
            injector={
                "injector_mode": "auto",
                "injector_extension_path": str(EXTENSION_PATH),
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
            server={"server_routes": {"*.*": "loopback_cdp"}},
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.transport.mode if cdp.transport else None, "nativemessaging")
            self.assertEqual(cdp.upstream_endpoint_kind, "modcdp_server")
            transport_url = cdp.transport.url if cdp.transport and cdp.transport.url else ""
            self.assertRegex(transport_url, rf"^native://{upstream_nativemessaging_host_name}@127\.0\.0\.1:\d+$")
            launched_profile_dir = cdp._launched_browser.get("profile_dir") if cdp._launched_browser else ""
            self.assertTrue((Path(launched_profile_dir) / "NativeMessagingHosts" / f"{upstream_nativemessaging_host_name}.json").exists())
            version = cdp.send("Browser.getVersion")
            self.assertIsInstance(version["product"], str)
            time.sleep(1.5)
            second_version = cdp.send("Browser.getVersion")
            self.assertIsInstance(second_version["product"], str)
        finally:
            cdp.close()
            temp_profile_dir.cleanup()


def _wait_until(predicate, timeout_s: float = 2.0) -> None:
    deadline = time.monotonic() + timeout_s
    while time.monotonic() < deadline:
        if predicate():
            return
        time.sleep(0.02)
    raise AssertionError("timed out waiting for condition")


def extension_launch_flag_test_browser_path() -> str:
    global NATIVE_MESSAGING_TEST_BROWSER_PATH
    if NATIVE_MESSAGING_TEST_BROWSER_PATH is not None:
        return NATIVE_MESSAGING_TEST_BROWSER_PATH
    explicit_candidates = [
        os.environ.get("CHROME_PATH"),
        "/usr/bin/chromium" if sys.platform.startswith("linux") else None,
    ]
    for candidate in explicit_candidates:
        if candidate and Path(candidate).exists():
            NATIVE_MESSAGING_TEST_BROWSER_PATH = candidate
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
    candidates = sorted(
        {candidate for pattern in patterns for candidate in glob.glob(pattern)},
        key=lambda candidate: (_path_version(candidate), Path(candidate).stat().st_mtime),
        reverse=True,
    )
    if candidates:
        NATIVE_MESSAGING_TEST_BROWSER_PATH = candidates[0]
        return candidates[0]
    raise RuntimeError("Native messaging tests require CHROME_PATH, /usr/bin/chromium, or Chrome for Testing.")


def _path_version(candidate: str) -> int:
    numbers = [int(value) for value in re.findall(r"\d+", candidate)]
    return max(numbers) if numbers else 0


if __name__ == "__main__":
    unittest.main()
