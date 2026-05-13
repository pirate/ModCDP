from __future__ import annotations

import os
import unittest

from modcdp import ModCDPClient
from modcdp.transport.PipeUpstreamTransport import PipeUpstreamTransport


class PipeUpstreamTransportTests(unittest.TestCase):
    def test_constructor_update_launcher_config_and_unconnected_errors_match_transport_surface(self) -> None:
        transport = PipeUpstreamTransport({"cdp_url": "pipe://constructor"})
        self.assertEqual(transport.mode, "pipe")
        self.assertEqual(transport.endpoint_kind, "raw_cdp")
        self.assertEqual(transport.url, "pipe://constructor")
        self.assertEqual(transport.getLauncherConfig(), {"remote_debugging": "pipe"})
        self.assertIs(transport.update({"cdp_url": "pipe://1234"}), transport)
        self.assertEqual(transport.url, "pipe://1234")
        with self.assertRaisesRegex(RuntimeError, r"upstream\.upstream_mode=pipe requires"):
            transport.connect()
        with self.assertRaisesRegex(RuntimeError, "CDP pipe is not connected"):
            transport.send({"id": 1, "method": "Browser.getVersion"})

    def test_resets_connection_state_after_pipe_closes(self) -> None:
        read_fd, read_writer_fd = os.pipe()
        write_reader_fd, write_fd = os.pipe()
        pipe_read = os.fdopen(read_fd, "rb", buffering=0)
        pipe_read_writer = os.fdopen(read_writer_fd, "wb", buffering=0)
        pipe_write_reader = os.fdopen(write_reader_fd, "rb", buffering=0)
        pipe_write = os.fdopen(write_fd, "wb", buffering=0)
        transport = PipeUpstreamTransport({"pipe_read": pipe_read, "pipe_write": pipe_write, "cdp_url": "pipe://test"})
        closed: list[Exception] = []
        transport.onClose(lambda error: closed.append(error))

        try:
            transport.connect()
            transport.send({"id": 1, "method": "Browser.getVersion", "params": {}})
            pipe_read_writer.close()
            self.assertTrue(_wait_for(lambda: len(closed) == 1))
            with self.assertRaisesRegex(RuntimeError, "CDP pipe is not connected"):
                transport.send({"id": 2, "method": "Browser.getVersion", "params": {}})
        finally:
            transport.close()
            pipe_write_reader.close()

    def test_launches_real_browser_and_uses_pid_scoped_pipe_url(self) -> None:
        cdp = ModCDPClient(
            launcher={"launcher_mode": "local", "launcher_options": {"headless": True, "sandbox": False}},
            upstream={"upstream_mode": "pipe"},
            injector={
                "injector_mode": "auto",
                "injector_service_worker_url_suffixes": ["/modcdp/service_worker.js"],
                "injector_trust_service_worker_target": True,
            },
        )

        try:
            cdp.connect()
            self.assertEqual(cdp.transport.mode if cdp.transport else None, "pipe")
            self.assertEqual(cdp.upstream_endpoint_kind, "raw_cdp")
            self.assertRegex(cdp.cdp_url or "", r"^pipe://\d+$")
            self.assertEqual(cdp.transport.url if cdp.transport else None, cdp.cdp_url)
            version = cdp.sendRaw("Browser.getVersion")
            self.assertIsInstance(version["product"], str)
        finally:
            cdp.close()

def _wait_for(fn, timeout_s: float = 2) -> bool:
    import time

    deadline = time.monotonic() + timeout_s
    while time.monotonic() < deadline:
        if fn():
            return True
        time.sleep(0.02)
    return False


if __name__ == "__main__":
    unittest.main()
