from __future__ import annotations

import socket
import subprocess
import tempfile
import threading
import time
import unittest
from pathlib import Path
from queue import Queue

from websocket import create_connection

from modcdp.transport.NatsUpstreamTransport import NatsUpstreamTransport


ROOT = Path(__file__).resolve().parents[2]


class NatsUpstreamTransportTests(unittest.TestCase):
    def test_config_owns_url_nats_subject_prefix_wait_timeout_and_injector_config(self) -> None:
        transport = NatsUpstreamTransport({"upstream_nats_url": "ws://127.0.0.1:4223", "upstream_nats_subject_prefix": "modcdp.one"})
        self.assertEqual(transport.url, "ws://127.0.0.1:4223/")
        self.assertEqual(transport.upstream_nats_subject_prefix, "modcdp.one")
        self.assertEqual(
            transport.getInjectorConfig(),
            {"upstream_nats_url": "ws://127.0.0.1:4223/", "upstream_nats_subject_prefix": "modcdp.one"},
        )
        self.assertIs(
            transport.update(
                {
                    "upstream_nats_url": "nats://127.0.0.1:4222",
                    "upstream_nats_subject_prefix": "modcdp.two",
                    "upstream_nats_role": "browser",
                    "upstream_nats_wait_timeout_ms": 5,
                }
            ),
            transport,
        )
        self.assertEqual(transport.url, "nats://127.0.0.1:4222")
        self.assertEqual(transport.upstream_nats_subject_prefix, "modcdp.two")
        with self.assertRaisesRegex(RuntimeError, "Timed out waiting 5ms for NATS ModCDP peer"):
            transport.waitForPeer()

    def test_close_rejects_pending_peer_waits(self) -> None:
        transport = NatsUpstreamTransport(
            {
                "upstream_nats_url": "ws://127.0.0.1:4223",
                "upstream_nats_subject_prefix": "modcdp.close",
                "upstream_nats_wait_timeout_ms": 5_000,
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
        self.assertRegex(str(error), r"NATS transport for modcdp\.close closed before a peer connected")

    def test_close_resets_peer_wait_state(self) -> None:
        transport = NatsUpstreamTransport({"upstream_nats_wait_timeout_ms": 5})
        transport._handle_payload('{"type":"modcdp.nats.hello","role":"browser","version":1}')

        transport.waitForPeer()
        transport.close()

        with self.assertRaisesRegex(RuntimeError, "Timed out waiting 5ms for NATS ModCDP peer"):
            transport.waitForPeer()
        self.assertTrue(transport.closed)

    def test_reconnect_after_close_resets_closed_state_with_real_nats_server(self) -> None:
        nats = _start_nats_server()
        transport = NatsUpstreamTransport({"upstream_nats_url": nats["url"], "upstream_nats_subject_prefix": f"modcdp.reconnect.{int(time.time() * 1000)}"})

        try:
            transport.connect()
            self.assertTrue(transport.connected)
            transport.close()
            self.assertFalse(transport.connected)
            self.assertTrue(transport.closed)
            transport.connect()
            self.assertTrue(transport.connected)
            self.assertFalse(transport.closed)
        finally:
            transport.close()
            nats["close"]()

def _start_nats_server():
    websocket_port = _free_port()
    client_port = _free_port()
    temp_dir = tempfile.TemporaryDirectory(prefix="modcdp-nats-")
    config_path = Path(temp_dir.name) / "nats.conf"
    config_path.write_text(
        "\n".join(
            [
                'host: "127.0.0.1"',
                f"port: {client_port}",
                "websocket {",
                '  host: "127.0.0.1"',
                f"  port: {websocket_port}",
                "  no_tls: true",
                "}",
                "",
            ]
        )
    )
    binary_path = subprocess.check_output(
        [
            "pnpm",
            "exec",
            "node",
            "--input-type=module",
            "-e",
            "import { getBinaryPath } from '@eplightning/nats-server'; console.log(await getBinaryPath())",
        ],
        cwd=ROOT,
        text=True,
    ).strip()
    proc = subprocess.Popen([binary_path, "-c", str(config_path)], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    url = f"ws://127.0.0.1:{websocket_port}"
    try:
        _wait_for_websocket(url)
    except Exception:
        _close_process(proc)
        temp_dir.cleanup()
        raise

    def close() -> None:
        _close_process(proc)
        temp_dir.cleanup()

    return {"url": url, "close": close}


def _wait_for_websocket(url: str, timeout_s: float = 10) -> None:
    deadline = time.time() + timeout_s
    last_error: Exception | None = None
    while time.time() < deadline:
        try:
            ws = create_connection(url, timeout=1)
            ws.close()
            return
        except Exception as error:
            last_error = error
            time.sleep(0.05)
    raise last_error or TimeoutError(f"Timed out waiting for {url}")


def _close_process(proc: subprocess.Popen) -> None:
    if proc.poll() is not None:
        return
    proc.terminate()
    try:
        proc.wait(timeout=2)
    except subprocess.TimeoutExpired:
        proc.kill()
        proc.wait(timeout=2)


def _free_port() -> int:
    sock = socket.socket()
    sock.bind(("127.0.0.1", 0))
    try:
        return int(sock.getsockname()[1])
    finally:
        sock.close()


if __name__ == "__main__":
    unittest.main()
