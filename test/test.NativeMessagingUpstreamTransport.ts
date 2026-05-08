import assert from "node:assert/strict";
import { existsSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { DEFAULT_NATIVE_MESSAGING_HOST_NAME } from "../bridge/NativeMessagingUpstreamTransport.js";
import { ModCDPClient } from "../client/js/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "dist", "extension");

test.skipIf(process.platform === "win32")(
  "nativemessaging upstream installs the launch-profile native host manifest and connects to a real extension",
  async () => {
    const native_client = new ModCDPClient({
      launch: {
        mode: "local",
        options: { headless: process.platform === "linux", sandbox: process.platform !== "linux" },
      },
      upstream: { mode: "nativemessaging" },
      extension: {
        mode: "auto",
        path: EXTENSION_PATH,
        service_worker_url_suffixes: ["/modcdp/service_worker.js"],
        trust_service_worker_target: true,
      },
      server: {
        routes: { "*.*": "loopback_cdp" },
      },
    });

    try {
      await native_client.connect();
      assert.equal(native_client.transport?.mode, "nativemessaging");
      assert.equal(native_client.upstream_endpoint_kind, "modcdp_server");
      assert.match(native_client.transport?.url ?? "", /^native:\/\/com\.modcdp\.bridge@127\.0\.0\.1:\d+$/);
      assert.equal(
        existsSync(
          path.join(
            native_client._launched?.profile_dir ?? "",
            "NativeMessagingHosts",
            `${DEFAULT_NATIVE_MESSAGING_HOST_NAME}.json`,
          ),
        ),
        true,
      );
      const version = (await native_client.send("Browser.getVersion")) as Record<string, unknown>;
      assert.equal(typeof version.product, "string");
    } finally {
      await native_client.close();
    }
  },
  90_000,
);
