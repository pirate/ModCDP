import { describe, expect, it } from "vitest";

import { UpstreamTransport, endpointKindForUpstream, parseHostPort } from "../bridge/UpstreamTransport.js";

describe("UpstreamTransport", () => {
  it("owns shared transport config, endpoint classification, and recv callbacks", async () => {
    const transport = new UpstreamTransport();
    const received: unknown[] = [];
    const stop = transport.onRecv((message) => received.push(message));

    expect(endpointKindForUpstream("ws")).toBe("raw_cdp");
    expect(endpointKindForUpstream("pipe")).toBe("raw_cdp");
    expect(endpointKindForUpstream("nativemessaging")).toBe("modcdp_server");
    expect(endpointKindForUpstream("reversews")).toBe("modcdp_server");
    expect(endpointKindForUpstream("nats")).toBe("modcdp_server");
    expect(parseHostPort("127.0.0.1:29292", "0.0.0.0", 80)).toEqual({ host: "127.0.0.1", port: 29292 });
    expect(transport.update()).toBe(transport);
    expect(transport.getLauncherConfig()).toEqual({});
    expect(transport.getInjectorConfig()).toEqual({});
    expect(transport.getServerConfig()).toEqual({});

    class TestTransport extends UpstreamTransport {
      emit(value: unknown) {
        this.parseAndEmitRecv(value);
      }
    }
    const test_transport = new TestTransport();
    const parsed: unknown[] = [];
    test_transport.onRecv((message) => parsed.push(message));
    test_transport.emit(JSON.stringify({ id: 1, result: { ok: true } }));
    test_transport.emit(JSON.stringify({ method: "Runtime.executionContextCreated", params: {} }));
    expect(parsed).toEqual([
      { id: 1, result: { ok: true } },
      { method: "Runtime.executionContextCreated", params: {} },
    ]);

    stop();
    expect(received).toEqual([]);
    await expect(transport.connect()).rejects.toThrow("UpstreamTransport.connect is not implemented.");
    expect(() => transport.send({ id: 1, method: "Browser.getVersion", params: {} })).toThrow(
      "UpstreamTransport.send is not implemented.",
    );
  });
});
