from __future__ import annotations

import threading
import time
import unittest
import os
import socket
from pathlib import Path
from queue import Queue

from modcdp import ModCDPClient
from modcdp.NativeMessagingUpstreamTransport import NativeMessagingUpstreamTransport


class NativeMessagingUpstreamTransportTests(unittest.TestCase):
    def test_config_owns_manifest_host_wait_timeout_loopback_and_injector_config(self) -> None:
        transport = NativeMessagingUpstreamTransport(
            {
                "manifest_path": "/tmp/modcdp-native-host.json",
                "manifest_paths": ["/tmp/modcdp-native-host-extra.json"],
                "host_name": "com.modcdp.test",
                "extension_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
                "wait_timeout_ms": 10,
            }
        )
        self.assertEqual(transport.getInjectorConfig(), {"native_host_name": "com.modcdp.test"})
        self.assertEqual(transport.getServerConfig(), {})
        self.assertIs(
            transport.update(
                {
                    "ws_url": "ws://127.0.0.1:9222/devtools/browser/test",
                    "manifest_paths": [],
                    "native_host_name": "com.modcdp.updated",
                    "wait_timeout_ms": 5,
                }
            ),
            transport,
        )
        self.assertEqual(
            transport.getServerConfig(),
            {"loopback_cdp_url": "ws://127.0.0.1:9222/devtools/browser/test"},
        )
        self.assertEqual(transport.getInjectorConfig(), {"native_host_name": "com.modcdp.updated"})
        self.assertFalse(transport.include_default_manifest_paths)
        transport.update({"manifest_path": None})
        self.assertTrue(transport.include_default_manifest_paths)
        transport.update({"user_data_dir": "/tmp/modcdp-profile-one"})
        transport.update({"user_data_dir": "/tmp/modcdp-profile-one"})
        transport.update({"user_data_dir": "/tmp/modcdp-profile-two"})
        self.assertEqual(
            transport.manifest_paths,
            [
                str(Path("/tmp/modcdp-profile-two") / "NativeMessagingHosts" / "com.modcdp.updated.json"),
                str(Path("/tmp/modcdp-profile-two") / "Default" / "NativeMessagingHosts" / "com.modcdp.updated.json"),
            ],
        )
        with self.assertRaisesRegex(RuntimeError, r"Timed out waiting 5ms for native messaging host com\.modcdp\.updated"):
            transport.waitForPeer()

    def test_close_resets_peer_wait_state(self) -> None:
        host_name = f"com.modcdp.close.reset.python.{os.getpid()}"
        transport = NativeMessagingUpstreamTransport({"host_name": host_name, "wait_timeout_ms": 5})
        transport.connect()
        bound_port = transport.bound_port
        if bound_port is None:
            self.fail("native messaging transport did not bind a port")
        peer = socket.create_connection(("127.0.0.1", bound_port), timeout=10)

        try:
            transport.waitForPeer()
            transport.close()
            host_name_pattern = host_name.replace(".", r"\.")
            with self.assertRaisesRegex(
                RuntimeError,
                rf"Timed out waiting 5ms for native messaging host {host_name_pattern}",
            ):
                transport.waitForPeer()
        finally:
            peer.close()
            transport.close()

    def test_waits_again_after_peer_disconnects(self) -> None:
        host_name = f"com.modcdp.disconnect.reset.python.{os.getpid()}"
        transport = NativeMessagingUpstreamTransport({"host_name": host_name, "wait_timeout_ms": 5})
        transport.connect()
        bound_port = transport.bound_port
        if bound_port is None:
            self.fail("native messaging transport did not bind a port")
        peer = socket.create_connection(("127.0.0.1", bound_port), timeout=10)

        try:
            transport.waitForPeer()
            peer.close()
            _wait_until(lambda: transport.socket is None)
            host_name_pattern = host_name.replace(".", r"\.")
            with self.assertRaisesRegex(
                RuntimeError,
                rf"Timed out waiting 5ms for native messaging host {host_name_pattern}",
            ):
                transport.waitForPeer()
        finally:
            peer.close()
            transport.close()

    def test_accepts_replacement_peer_after_disconnect(self) -> None:
        host_name = f"com.modcdp.replacement.python.{os.getpid()}"
        transport = NativeMessagingUpstreamTransport({"host_name": host_name, "wait_timeout_ms": 500})
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
                "host_name": "com.modcdp.close",
                "wait_timeout_ms": 5_000,
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
        host_name = f"com.modcdp.test.python.{os.getpid()}"
        cdp = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "nativemessaging", "nativemessaging_host_name": host_name},
            extension={
                "mode": "auto",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
            server={"routes": {"*.*": "loopback_cdp"}},
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.transport.mode if cdp.transport else None, "nativemessaging")
            self.assertEqual(cdp.upstream_endpoint_kind, "modcdp_server")
            transport_url = cdp.transport.url if cdp.transport and cdp.transport.url else ""
            self.assertRegex(transport_url, rf"^native://{host_name}@127\.0\.0\.1:\d+$")
            profile_dir = cdp._launched_browser.get("profile_dir") if cdp._launched_browser else ""
            self.assertTrue((Path(profile_dir) / "NativeMessagingHosts" / f"{host_name}.json").exists())
            version = cdp.send("Browser.getVersion")
            self.assertIsInstance(version["product"], str)
        finally:
            cdp.close()


def _wait_until(predicate, timeout_s: float = 2.0) -> None:
    deadline = time.monotonic() + timeout_s
    while time.monotonic() < deadline:
        if predicate():
            return
        time.sleep(0.02)
    raise AssertionError("timed out waiting for condition")


if __name__ == "__main__":
    unittest.main()
