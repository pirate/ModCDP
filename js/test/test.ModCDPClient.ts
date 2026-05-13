import assert from "node:assert/strict";
import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { ModCDPClient } from "../src/client/ModCDPClient.js";
import { CdpSocket } from "./helpers.BrowserLauncher.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

test("ModCDPClient normalizes nested config owners", () => {
  const cdp = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_executable_path: "/tmp/chrome",
      launcher_user_data_dir: "/tmp/profile",
      launcher_options: { headless: true },
    },
    upstream: {
      upstream_mode: "ws",
      upstream_cdp_url: "http://127.0.0.1:9222",
      upstream_nats_wait_timeout_ms: 345,
      upstream_reversews_wait_timeout_ms: 456,
      upstream_nativemessaging_manifest: "/tmp/native-host.json",
      upstream_nativemessaging_manifests: ["/tmp/native-host-extra.json"],
      upstream_nativemessaging_host_name: "com.modcdp.custom",
      upstream_nativemessaging_wait_timeout_ms: 567,
      upstream_ws_connect_error_settle_timeout_ms: 321,
    },
    injector: {
      injector_mode: "discover",
      injector_extension_path: "/tmp/ext",
      injector_extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      injector_service_worker_url_includes: ["modcdp"],
      injector_service_worker_url_suffixes: ["/custom/service_worker.js"],
      injector_trust_service_worker_target: true,
      injector_require_service_worker_target: true,
      injector_execution_context_timeout_ms: 4321,
      injector_service_worker_probe_timeout_ms: 5432,
      injector_service_worker_ready_timeout_ms: 6543,
      injector_service_worker_poll_interval_ms: 76,
      injector_target_session_poll_interval_ms: 87,
    },
    client: {
      client_routes: { "*.*": "direct_cdp" },
      client_hydrate_aliases: false,
      client_mirror_upstream_events: false,
      client_cdp_send_timeout_ms: 1234,
      client_event_wait_timeout_ms: 2345,
    },
    server: {
      server_routes: { "*.*": "loopback_cdp" },
      server_browser_token: "token-1",
      server_cdp_send_timeout_ms: 9876,
      server_loopback_execution_context_timeout_ms: 8765,
      server_ws_connect_error_settle_timeout_ms: 7654,
    },
  });

  assert.deepEqual(cdp.launcher.launcher_options, { headless: true });
  assert.equal(cdp._launcherOptions().executable_path, "/tmp/chrome");
  assert.equal(cdp._launcherOptions().user_data_dir, "/tmp/profile");
  assert.equal(cdp.upstream.upstream_nats_wait_timeout_ms, 345);
  assert.equal(cdp.upstream.upstream_reversews_wait_timeout_ms, 456);
  assert.equal(cdp.upstream.upstream_nativemessaging_manifest, "/tmp/native-host.json");
  assert.deepEqual(cdp.upstream.upstream_nativemessaging_manifests, ["/tmp/native-host-extra.json"]);
  assert.equal(cdp.upstream.upstream_nativemessaging_host_name, "com.modcdp.custom");
  assert.equal(cdp.upstream.upstream_nativemessaging_wait_timeout_ms, 567);
  assert.equal(cdp.upstream.upstream_ws_connect_error_settle_timeout_ms, 321);
  assert.equal(cdp.injector.injector_execution_context_timeout_ms, 4321);
  assert.equal(cdp.injector.injector_service_worker_probe_timeout_ms, 5432);
  assert.equal(cdp.injector.injector_service_worker_ready_timeout_ms, 6543);
  assert.equal(cdp.injector.injector_service_worker_poll_interval_ms, 76);
  assert.equal(cdp.injector.injector_target_session_poll_interval_ms, 87);
  assert.equal(cdp.client.client_routes["*.*"], "direct_cdp");
  assert.equal(cdp.client.client_hydrate_aliases, false);
  assert.equal(cdp.client.client_mirror_upstream_events, false);
  assert.equal(cdp.client.client_cdp_send_timeout_ms, 1234);
  assert.equal(cdp.client.client_event_wait_timeout_ms, 2345);
  assert.equal("routes" in cdp, false);
  assert.equal("cdp_send_timeout_ms" in cdp, false);
  assert.equal("service_worker_probe_timeout_ms" in cdp, false);

  const params = cdp._serverConfigureParams();
  assert.equal(params.client.client_routes["*.*"], "direct_cdp");
  assert.equal(params.server.server_browser_token, "token-1");
  assert.equal(params.server.server_cdp_send_timeout_ms, 9876);
  assert.equal(params.server.server_loopback_execution_context_timeout_ms, 8765);
  assert.equal(params.server.server_ws_connect_error_settle_timeout_ms, 7654);
});

test("ModCDPClient dispatches root events before extension session is attached", () => {
  const cdp = new ModCDPClient();
  const seen: string[] = [];
  cdp.on("Target.targetCreated", (payload: { targetInfo?: { targetId?: string } }) => {
    seen.push(String(payload.targetInfo?.targetId));
  });

  cdp._onRecv({
    method: "Target.targetCreated",
    params: {
      targetInfo: { targetId: "target-1", type: "page", url: "about:blank" },
    },
  });

  assert.deepEqual(seen, ["target-1"]);
});

test("ModCDPClient event dispatch snapshots handlers when once removes itself", () => {
  const cdp = new ModCDPClient();
  const seen: string[] = [];
  cdp.once("Target.targetCreated", () => {
    seen.push("once");
  });
  cdp.on("Target.targetCreated", () => {
    seen.push("persistent");
  });

  cdp._onRecv({
    method: "Target.targetCreated",
    params: {
      targetInfo: { targetId: "target-1", type: "page", url: "about:blank" },
    },
  });
  assert.deepEqual(seen, ["once", "persistent"]);

  seen.length = 0;
  cdp._onRecv({
    method: "Target.targetCreated",
    params: {
      targetInfo: { targetId: "target-2", type: "page", url: "about:blank" },
    },
  });
  assert.deepEqual(seen, ["persistent"]);
});

test("ModCDPClient connects with nested launch/upstream/extension/client/server config", async () => {
  const cdp = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
    client: {
      client_routes: {
        "Mod.*": "service_worker",
        "Custom.*": "service_worker",
        "*.*": "direct_cdp",
      },
      client_hydrate_aliases: true,
      client_mirror_upstream_events: true,
      client_cdp_send_timeout_ms: 10_000,
      client_event_wait_timeout_ms: 10_000,
    },
    server: {
      server_routes: { "*.*": "loopback_cdp" },
      server_cdp_send_timeout_ms: 10_000,
      server_loopback_execution_context_timeout_ms: 10_000,
      server_ws_connect_error_settle_timeout_ms: 250,
    },
  });

  let direct_session_target_id: string | null = null;
  try {
    await cdp.connect();
    assert.equal(cdp.launcher.launcher_mode, "local");
    assert.equal(cdp.upstream.upstream_mode, "ws");
    assert.equal(cdp.upstream.upstream_reversews_wait_timeout_ms, 10_000);
    assert.equal(cdp.injector.injector_mode, "auto");
    assert.equal(cdp.client.client_routes["*.*"], "direct_cdp");
    assert.equal(cdp.upstream_endpoint_kind, "raw_cdp");
    assert.match(cdp.cdp_url ?? "", /^ws:\/\//);
    await delay(2_000);
    const targets = (await cdp.sendRaw("Target.getTargets")) as {
      targetInfos: { type?: string; url?: string }[];
    };
    assert.equal(
      targets.targetInfos.some(
        (target) =>
          target.type === "service_worker" &&
          target.url === `chrome-extension://${cdp.extension_id}/modcdp/service_worker.js`,
      ),
      true,
    );
    assert.equal(
      targets.targetInfos.some(
        (target) =>
          target.type === "background_page" &&
          target.url === `chrome-extension://${cdp.extension_id}/offscreen/keepalive.html`,
      ),
      true,
    );
    assert.equal(typeof (await cdp.Browser.getVersion()).product, "string");
    const direct_target = (await cdp.send("Target.createTarget", {
      url: "about:blank#direct-session-routing",
    })) as Record<string, unknown>;
    direct_session_target_id = String(direct_target.targetId);
    const direct_session = (await cdp.send("Target.attachToTarget", {
      targetId: direct_session_target_id,
      flatten: true,
    })) as Record<string, unknown>;
    const direct_eval = (await cdp.send(
      "Runtime.evaluate",
      { expression: "1 + 1", returnByValue: true },
      String(direct_session.sessionId),
    )) as { result?: { value?: unknown } };
    assert.equal(direct_eval.result?.value, 2);
    const sent_at = Date.now();
    const pong = new Promise<Record<string, unknown>>((resolve) => {
      const listener = (payload: Record<string, unknown>) => {
        if (payload.sent_at !== sent_at) return;
        cdp.off("Mod.pong", listener);
        resolve(payload);
      };
      cdp.on("Mod.pong", listener);
    });
    const ping_result = (await cdp.Mod.ping({ sent_at })) as Record<string, unknown>;
    const pong_payload = await pong;
    assert.equal(ping_result.ok, true);
    assert.equal(pong_payload.sent_at, sent_at);
    assert.equal(typeof pong_payload.received_at, "number");
    assert.equal(pong_payload.from, "extension-service-worker");
  } finally {
    if (direct_session_target_id) {
      await cdp.send("Target.closeTarget", { targetId: direct_session_target_id }).catch(() => ({}));
    }
    await cdp.close();
  }
}, 60_000);

test("ModCDPClient preserves explicit empty service worker suffix config", async () => {
  const cdp = new ModCDPClient({
    injector: {
      injector_mode: "borrow",
      injector_service_worker_url_suffixes: [],
    },
  });

  assert.deepEqual(cdp.injector.injector_service_worker_url_suffixes, []);
  assert.deepEqual((await cdp._baseInjectorConfig()).injector_service_worker_url_suffixes, []);
});

test("ModCDPClient defaults service worker suffix config to the ModCDP worker", async () => {
  const cdp = new ModCDPClient();

  assert.deepEqual(cdp.injector.injector_service_worker_url_suffixes, ["/modcdp/service_worker.js"]);
  assert.deepEqual((await cdp._baseInjectorConfig()).injector_service_worker_url_suffixes, [
    "/modcdp/service_worker.js",
  ]);
});

test("ModCDPClient preserves explicit null server config", () => {
  const cdp = new ModCDPClient({ server: null });

  assert.equal(cdp.server, null);
});

test("ModCDPClient only exposes injector attach after CDP send is available", () => {
  const cdp = new ModCDPClient();
  const disconnected_config = cdp._baseInjectorConfig(null);
  assert.equal(disconnected_config.send, null);
  assert.equal(disconnected_config.attachToTarget, null);

  const connected_config = cdp._baseInjectorConfig(async () => ({}));
  assert.equal(typeof connected_config.send, "function");
  assert.equal(typeof connected_config.attachToTarget, "function");
});

test("ModCDPClient defaults launched ModCDP-server upstreams to extension auto", () => {
  for (const mode of ["nativemessaging", "reversews", "nats"] as const) {
    const launched = new ModCDPClient({ launcher: { launcher_mode: "local" }, upstream: { upstream_mode: mode } });
    assert.equal(launched.launcher.launcher_mode, "local");
    assert.equal(launched.upstream_endpoint_kind, "modcdp_server");
    assert.equal(launched.injector.injector_mode, "auto");

    const attach_only = new ModCDPClient({ upstream: { upstream_mode: mode } });
    assert.equal(attach_only.launcher.launcher_mode, "none");
    assert.equal(attach_only.upstream_endpoint_kind, "modcdp_server");
    assert.equal(attach_only.injector.injector_mode, "none");
  }
});

test("ModCDPClient rejects unknown component modes at their owning factory boundary", () => {
  assert.throws(
    () =>
      new ModCDPClient({
        upstream: { upstream_mode: "bogus" as any },
      })._upstreamTransport(),
    /unknown upstream\.upstream_mode=bogus/,
  );
  assert.throws(
    () => new ModCDPClient({ launcher: { launcher_mode: "bogus" as any } })._browserLauncher(),
    /unknown launcher\.launcher_mode=bogus/,
  );
  assert.throws(
    () => new ModCDPClient({ injector: { injector_mode: "bogus" as any } })._injectorsForConfig(),
    /unknown injector\.injector_mode=bogus/,
  );
});

test("ModCDPClient.close does not close a remote browser it did not launch", async () => {
  const chrome = await new LocalBrowserLauncher({
    headless: true,
    sandbox: process.platform !== "linux",
    extra_args: [`--load-extension=${EXTENSION_PATH}`],
  }).launch();
  const raw_cdp = await CdpSocket.connect(chrome.cdp_url!);
  const cdp = new ModCDPClient({
    launcher: { launcher_mode: "remote" },
    upstream: { upstream_mode: "ws", upstream_cdp_url: chrome.cdp_url },
    injector: {
      injector_mode: "discover",
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
  });

  try {
    await cdp.connect();
    await cdp.close();
    await delay(500);
    const version = await raw_cdp.send("Browser.getVersion");
    assert.match(String(version.product), /Chrome|Chromium/);
  } finally {
    await raw_cdp.close();
    await cdp.close();
    await chrome.close();
  }
}, 60_000);

test("ModCDPClient.close keeps injector files until after launched browser shutdown", async () => {
  const cdp = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: {
      upstream_mode: "reversews",
      upstream_reversews_wait_timeout_ms: 30_000,
    },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
    server: {
      server_routes: { "*.*": "loopback_cdp" },
    },
  });

  try {
    await cdp.connect();
    const injector = cdp._injectors.find(
      (candidate) => candidate.constructor.name === "LocalBrowserLaunchExtensionInjector",
    ) as unknown as { unpacked_extension_path?: string | null } | undefined;
    const unpacked_extension_path = injector?.unpacked_extension_path;
    assert.equal(typeof unpacked_extension_path, "string");
    assert.notEqual(unpacked_extension_path, EXTENSION_PATH);

    const launched = cdp._launched;
    assert.ok(launched);
    const close_browser = launched.close;
    let browser_close_saw_extension = false;
    launched.close = async () => {
      browser_close_saw_extension = existsSync(unpacked_extension_path!);
      await close_browser();
    };

    await cdp.close();

    assert.equal(browser_close_saw_extension, true);
    assert.equal(existsSync(unpacked_extension_path!), false);
  } finally {
    await cdp.close();
  }
  assert.equal(cdp.transport, null);
  assert.equal(cdp._launched, null);
  assert.deepEqual(cdp._injectors, []);
}, 90_000);

test("ModCDPClient.close clears top-level connection state", async () => {
  const cdp = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: { headless: true, sandbox: process.platform !== "linux" },
    },
    upstream: { upstream_mode: "ws" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
    },
  });

  await cdp.connect();
  assert.ok(cdp.transport);
  await cdp.close();

  assert.equal(cdp.transport, null);
  await assert.rejects(() => cdp.sendRaw("Browser.getVersion"), /ModCDP upstream is not connected/);
});
