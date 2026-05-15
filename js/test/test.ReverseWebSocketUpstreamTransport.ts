import assert from "node:assert/strict";
import { once } from "node:events";
import { existsSync, readdirSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { homedir, platform } from "node:os";
import WebSocket from "ws";
import { test } from "vitest";

import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import { ReverseWebSocketUpstreamTransport } from "../src/transport/ReverseWebSocketUpstreamTransport.js";
import { ModCDPClient } from "../src/client/ModCDPClient.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");
const REVERSEWS_TEST_BROWSER_PATH = reversewsTestBrowserPath();

test("reversews upstream config owns bind updates and wait timeout", async () => {
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: "127.0.0.1:29292",
    upstream_reversews_wait_timeout_ms: 10,
  });
  assert.equal(transport.url, "ws://127.0.0.1:29292");
  assert.deepEqual(transport.getInjectorConfig(), {});
  assert.equal(
    transport.update({
      upstream_reversews_bind: "127.0.0.1:29293",
      upstream_reversews_wait_timeout_ms: 5,
    }),
    transport,
  );
  assert.equal(transport.url, "ws://127.0.0.1:29293");
  assert.deepEqual(transport.getInjectorConfig(), {});
  assert.throws(
    () => transport.send({ id: 1, method: "Browser.getVersion" }),
    /No reverse ModCDP extension peer is connected/,
  );
  await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms/);
});

test("reversews upstream close rejects pending peer waits", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: `127.0.0.1:${reverse_port}`,
    upstream_reversews_wait_timeout_ms: 5_000,
  });
  const pending = transport.waitForPeer();

  await transport.close();

  await assert.rejects(
    () => pending,
    new RegExp(`Reverse websocket transport at ws://127\\.0\\.0\\.1:${reverse_port} closed before a peer connected`),
  );
});

test("reversews upstream close resets peer wait state", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: `127.0.0.1:${reverse_port}`,
    upstream_reversews_wait_timeout_ms: 5,
  });
  await transport.connect();
  const peer = new WebSocket(transport.url);
  await once(peer, "open");
  peer.send(JSON.stringify({ type: "modcdp.reverse.hello", role: "test-peer", version: 1 }));

  try {
    await transport.waitForPeer();
    assert.deepEqual(transport.peer_info, { type: "modcdp.reverse.hello", role: "test-peer", version: 1 });
    await transport.close();

    await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms/);
    assert.equal(transport.peer_info, null);
  } finally {
    peer.close();
    await transport.close();
  }
});

test("reversews upstream waits again after a peer disconnects", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: `127.0.0.1:${reverse_port}`,
    upstream_reversews_wait_timeout_ms: 5,
  });
  await transport.connect();
  const peer = new WebSocket(transport.url);
  await once(peer, "open");
  peer.send(JSON.stringify({ type: "modcdp.reverse.hello", role: "test-peer", version: 1 }));

  try {
    await transport.waitForPeer();
    peer.close();
    await waitFor(() => (transport as unknown as { socket: unknown | null }).socket === null);

    await assert.rejects(() => transport.waitForPeer(), /Timed out waiting 5ms/);
  } finally {
    peer.close();
    await transport.close();
  }
});

test("reversews upstream accepts a replacement peer after disconnect", async () => {
  const reverse_port = await LocalBrowserLauncher.freePort();
  const transport = new ReverseWebSocketUpstreamTransport({
    upstream_reversews_bind: `127.0.0.1:${reverse_port}`,
    upstream_reversews_wait_timeout_ms: 500,
  });
  await transport.connect();
  const first_peer = new WebSocket(transport.url);
  await once(first_peer, "open");
  first_peer.send(JSON.stringify({ type: "modcdp.reverse.hello", role: "first-peer", version: 1 }));

  try {
    await transport.waitForPeer();
    first_peer.close();
    await waitFor(() => (transport as unknown as { socket: unknown | null }).socket === null);

    const second_peer = new WebSocket(transport.url);
    await once(second_peer, "open");
    second_peer.send(JSON.stringify({ type: "modcdp.reverse.hello", role: "second-peer", version: 1 }));
    try {
      await transport.waitForPeer();
      assert.equal(transport.peer_info?.role, "second-peer");
    } finally {
      second_peer.close();
    }
  } finally {
    first_peer.close();
    await transport.close();
  }
});

test("reversews upstream accepts a real extension reverse connection and routes CDP through chrome debugger", async () => {
  const cdp = new ModCDPClient({
    launcher: {
      launcher_mode: "local",
      launcher_options: {
        headless: process.platform === "linux" && !process.env.DISPLAY,
        // Reversews is browser -> client only. After explicit CHROME_PATH and
        // CI /usr/bin/chromium, these tests use Chrome for Testing because
        // Canary rejects --load-extension in this local test path.
        executable_path: REVERSEWS_TEST_BROWSER_PATH,
      },
    },
    upstream: { upstream_mode: "reversews" },
    injector: {
      injector_mode: "auto",
      injector_extension_path: EXTENSION_PATH,
      injector_service_worker_url_suffixes: ["/modcdp/service_worker.js"],
      injector_trust_service_worker_target: true,
      injector_service_worker_probe_timeout_ms: 1_000,
    },
  });

  try {
    await cdp.connect();
    assert.equal(cdp.transport?.mode, "reversews");
    assert.equal(cdp.upstream_endpoint_kind, "modcdp_server");
    assert.equal(cdp.transport?.url, "ws://127.0.0.1:29292");
    assert.equal(
      (cdp.transport as ReverseWebSocketUpstreamTransport).peer_info?.extension_id,
      "mdedooklbnfejodmnhmkdpkaedafkehf",
    );
    const evaluated = (await cdp.send("Runtime.evaluate", {
      expression: "location.href",
      returnByValue: true,
    })) as { result?: { value?: unknown } };
    assert.equal(evaluated.result?.value, "about:blank");
    await new Promise((resolve) => setTimeout(resolve, 1_500));
    const second_evaluated = (await cdp.send("Runtime.evaluate", {
      expression: "document.readyState",
      returnByValue: true,
    })) as { result?: { value?: unknown } };
    assert.equal(second_evaluated.result?.value, "complete");
  } finally {
    await cdp.close();
  }
}, 60_000);

function reversewsTestBrowserPath() {
  const explicit_candidates = [process.env.CHROME_PATH, platform() === "linux" ? "/usr/bin/chromium" : null].filter(
    (candidate): candidate is string => Boolean(candidate),
  );
  for (const candidate of explicit_candidates) {
    if (existsSync(candidate)) return candidate;
  }
  const home = homedir();
  const patterns =
    platform() === "darwin"
      ? [
          path.join(
            home,
            "Library/Caches/ms-playwright/chromium-*/chrome-mac*/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing",
          ),
          path.join(home, "Library/Caches/ms-playwright/chromium-*/chrome-mac*/Chromium.app/Contents/MacOS/Chromium"),
          path.join(
            home,
            "Library/Caches/puppeteer/chrome/mac*-*/chrome-mac*/Google Chrome for Testing.app/Contents/MacOS/Google Chrome for Testing",
          ),
        ]
      : platform() === "win32"
        ? [
            path.join(
              process.env.LOCALAPPDATA || path.join(home, "AppData/Local"),
              "ms-playwright/chromium-*/chrome-win*/chrome.exe",
            ),
            path.join(home, ".cache/puppeteer/chrome/win*-*/chrome-win*/chrome.exe"),
          ]
        : [
            path.join(home, ".cache/ms-playwright/chromium-*/chrome-linux*/chrome"),
            "/opt/pw-browsers/chromium-*/chrome-linux*/chrome",
            path.join(home, ".cache/puppeteer/chrome/linux-*/chrome-linux*/chrome"),
          ];
  const candidates = newestFirst(patterns.flatMap(expandGlob));
  if (candidates[0]) return candidates[0];
  throw new Error("Reversews tests require CHROME_PATH, /usr/bin/chromium, or Chrome for Testing.");
}

function expandGlob(pattern: string) {
  const normalized = path.normalize(pattern);
  const { root } = path.parse(normalized);
  const parts = normalized.slice(root.length).split(path.sep).filter(Boolean);
  let candidates = [root || "."];
  for (const part of parts) {
    const has_wildcard = part.includes("*");
    const matcher = has_wildcard ? wildcardToRegExp(part) : null;
    const next: string[] = [];
    for (const base of candidates) {
      if (!existsSync(base)) continue;
      if (!has_wildcard) {
        const candidate = path.join(base, part);
        if (existsSync(candidate)) next.push(candidate);
        continue;
      }
      for (const child of readdirSync(base)) {
        if (matcher!.test(child)) next.push(path.join(base, child));
      }
    }
    candidates = next;
  }
  return candidates.filter((candidate) => existsSync(candidate));
}

function wildcardToRegExp(value: string) {
  return new RegExp(`^${value.replace(/[.+^${}()|[\]\\]/g, "\\$&").replace(/\*/g, ".*")}$`);
}

function newestFirst(candidates: string[]) {
  return [...new Set(candidates)].sort((a, b) => {
    const left = scorePath(a);
    const right = scorePath(b);
    return right.version - left.version || right.mtime - left.mtime || a.localeCompare(b);
  });
}

function scorePath(candidate: string) {
  const numbers = candidate.match(/\d+/g)?.map(Number) ?? [];
  const version = numbers.length > 0 ? Math.max(...numbers) : 0;
  let mtime = 0;
  try {
    mtime = statSync(candidate).mtimeMs;
  } catch {}
  return { version, mtime };
}

async function waitFor(predicate: () => boolean, timeout_ms = 2_000) {
  const deadline = Date.now() + timeout_ms;
  while (Date.now() < deadline) {
    if (predicate()) return;
    await new Promise((resolve) => setTimeout(resolve, 20));
  }
  throw new Error("Timed out waiting for condition");
}
