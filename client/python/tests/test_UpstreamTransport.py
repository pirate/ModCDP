from __future__ import annotations

import unittest

from modcdp.UpstreamTransport import UpstreamTransport, endpoint_kind_for_upstream


class TestTransport(UpstreamTransport):
    mode = "ws"
    endpoint_kind = "raw_cdp"

    def emit(self, value: str) -> None:
        self._parse_and_emit_recv(value)


class UpstreamTransportTests(unittest.TestCase):
    def test_shared_transport_config_endpoint_classification_and_recv_callbacks(self) -> None:
        transport = UpstreamTransport()
        received = []
        stop = transport.on_recv(lambda message: received.append(message))

        self.assertEqual(endpoint_kind_for_upstream("ws"), "raw_cdp")
        self.assertEqual(endpoint_kind_for_upstream("pipe"), "raw_cdp")
        self.assertEqual(endpoint_kind_for_upstream("nativemessaging"), "modcdp_server")
        self.assertEqual(endpoint_kind_for_upstream("reversews"), "modcdp_server")
        self.assertEqual(endpoint_kind_for_upstream("nats"), "modcdp_server")
        self.assertIs(transport.update(), transport)
        self.assertEqual(transport.getLauncherConfig(), {})
        self.assertEqual(transport.getInjectorConfig(), {})
        self.assertEqual(transport.getServerConfig(), {})

        parsed = []
        test_transport = TestTransport()
        test_transport.on_recv(lambda message: parsed.append(message))
        test_transport.emit('{"id":1,"result":{"ok":true}}')
        test_transport.emit('{"method":"Runtime.executionContextCreated","params":{}}')
        self.assertEqual(
            parsed,
            [
                {"id": 1, "result": {"ok": True}},
                {"method": "Runtime.executionContextCreated", "params": {}},
            ],
        )

        stop()
        self.assertEqual(received, [])
        with self.assertRaisesRegex(NotImplementedError, "UpstreamTransport.connect is not implemented"):
            transport.connect()
        with self.assertRaisesRegex(NotImplementedError, "UpstreamTransport.send is not implemented"):
            transport.send({"id": 1, "method": "Browser.getVersion", "params": {}})


if __name__ == "__main__":
    unittest.main()
