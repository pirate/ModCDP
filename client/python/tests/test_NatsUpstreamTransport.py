from __future__ import annotations

import socket
import subprocess
import tempfile
import time
import unittest
from pathlib import Path

from websocket import create_connection

from modcdp import ModCDPClient
from modcdp.NatsUpstreamTransport import NatsUpstreamTransport


ROOT = Path(__file__).resolve().parents[3]


class NatsUpstreamTransportTests(unittest.TestCase):
    def test_relays_cdp_through_real_extension_over_real_nats_server(self) -> None:
        nats = _start_nats_server()
        subject_prefix = f"modcdp.test.{int(time.time() * 1000)}"
        cdp = ModCDPClient(
            launch={"mode": "local", "options": {"headless": True, "sandbox": False}},
            upstream={"mode": "nats", "nats_url": nats["url"], "nats_subject_prefix": subject_prefix},
            extension={
                "mode": "auto",
                "service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "trust_service_worker_target": True,
            },
            server={"routes": {"*.*": "loopback_cdp"}},
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.transport.mode if cdp.transport else None, "nats")
            self.assertEqual(cdp.upstream_endpoint_kind, "modcdp_server")
            transport = cdp.transport
            if not isinstance(transport, NatsUpstreamTransport):
                self.fail(f"transport = {type(transport).__name__}")
            self.assertEqual(transport.url, f"{nats['url']}/")
            self.assertEqual(transport.subject_prefix, subject_prefix)
            version = cdp.send("Browser.getVersion")
            self.assertIsInstance(version["product"], str)
        finally:
            cdp.close()
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
