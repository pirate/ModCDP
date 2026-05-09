from __future__ import annotations

import json
import socket
import threading
import time
import unittest
from queue import Queue
from typing import cast

from websocket import create_connection

from modcdp import ModCDPClient
from modcdp.ReverseWebSocketUpstreamTransport import ReverseWebSocketUpstreamTransport


class ReverseWebSocketUpstreamTransportTests(unittest.TestCase):
    def test_config_owns_bind_updates_wait_timeout_and_injector_config(self) -> None:
        transport = ReverseWebSocketUpstreamTransport("127.0.0.1:29292", 10)
        self.assertEqual(transport.url, "ws://127.0.0.1:29292")
        self.assertEqual(transport.getInjectorConfig(), {"reverse_proxy_url": "ws://127.0.0.1:29292"})
        self.assertIs(transport.update({"reversews_bind": "127.0.0.1:29293", "reversews_wait_timeout_ms": 5}), transport)
        self.assertEqual(transport.url, "ws://127.0.0.1:29293")
        self.assertEqual(transport.getInjectorConfig(), {"reverse_proxy_url": "ws://127.0.0.1:29293"})
        with self.assertRaisesRegex(RuntimeError, "Timed out waiting 5ms"):
            transport.waitForPeer()

    def test_close_rejects_pending_peer_waits(self) -> None:
        reverse_port = _free_port()
        transport = ReverseWebSocketUpstreamTransport(f"127.0.0.1:{reverse_port}", 5_000)
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
        transport = ReverseWebSocketUpstreamTransport(f"127.0.0.1:{reverse_port}", 5)
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

    def test_accepts_real_extension_reverse_connection_and_routes_cdp_through_loopback(self) -> None:
        reverse_port = _free_port()
        reverse_bind = f"127.0.0.1:{reverse_port}"
        cdp = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "reversews", "reversews_bind": reverse_bind},
            extension={
                "mode": "auto",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
            server={"routes": {"*.*": "loopback_cdp"}},
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.transport.mode if cdp.transport else None, "reversews")
            self.assertEqual(cdp.upstream_endpoint_kind, "modcdp_server")
            self.assertIsInstance(cdp.transport, ReverseWebSocketUpstreamTransport)
            transport = cast(ReverseWebSocketUpstreamTransport, cdp.transport)
            self.assertEqual(
                transport.peer_info.get("extension_id") if transport.peer_info else None,
                "mdedooklbnfejodmnhmkdpkaedafkehf",
            )
            version = cdp.send("Browser.getVersion")
            self.assertIsInstance(version["product"], str)
        finally:
            cdp.close()


def _free_port() -> int:
    sock = socket.socket()
    sock.bind(("127.0.0.1", 0))
    try:
        return int(sock.getsockname()[1])
    finally:
        sock.close()


if __name__ == "__main__":
    unittest.main()
