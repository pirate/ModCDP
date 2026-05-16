import assert from "node:assert/strict";
import { existsSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "vitest";

import { DEFAULT_MODCDP_EXTENSION_ID } from "../src/injector/ExtensionInjector.js";
import { LocalBrowserLaunchExtensionInjector } from "../src/injector/LocalBrowserLaunchExtensionInjector.js";

const HERE = path.dirname(fileURLToPath(import.meta.url));
const EXTENSION_PATH = path.resolve(HERE, "..", "..", "dist", "extension");

function crc32(data: Buffer) {
  let crc = 0xffffffff;
  for (const byte of data) {
    crc ^= byte;
    for (let bit = 0; bit < 8; bit++)
      crc = (crc >>> 1) ^ (0xedb88320 & -(crc & 1));
  }
  return (crc ^ 0xffffffff) >>> 0;
}

function storedZipEntry(name: string, data: Buffer) {
  const name_buffer = Buffer.from(name);
  const checksum = crc32(data);
  const local = Buffer.alloc(30);
  local.writeUInt32LE(0x04034b50, 0);
  local.writeUInt16LE(20, 4);
  local.writeUInt32LE(checksum, 14);
  local.writeUInt32LE(data.length, 18);
  local.writeUInt32LE(data.length, 22);
  local.writeUInt16LE(name_buffer.length, 26);
  const central = Buffer.alloc(46);
  central.writeUInt32LE(0x02014b50, 0);
  central.writeUInt16LE(20, 4);
  central.writeUInt16LE(20, 6);
  central.writeUInt32LE(checksum, 16);
  central.writeUInt32LE(data.length, 20);
  central.writeUInt32LE(data.length, 24);
  central.writeUInt16LE(name_buffer.length, 28);
  const central_offset = local.length + name_buffer.length + data.length;
  const end = Buffer.alloc(22);
  end.writeUInt32LE(0x06054b50, 0);
  end.writeUInt16LE(1, 8);
  end.writeUInt16LE(1, 10);
  end.writeUInt32LE(central.length + name_buffer.length, 12);
  end.writeUInt32LE(central_offset, 16);
  return Buffer.concat([local, name_buffer, data, central, name_buffer, end]);
}

test("LocalBrowserLaunchExtensionInjector rejects zip entries outside extraction directory", async () => {
  const temp_dir = mkdtempSync(path.join(os.tmpdir(), "modcdp-bad-zip-"));
  const zip_path = path.join(temp_dir, "extension.zip");
  writeFileSync(zip_path, storedZipEntry("../evil.txt", Buffer.from("evil")));
  const injector = new LocalBrowserLaunchExtensionInjector({
    injector_extension_path: zip_path,
  });

  try {
    await assert.rejects(
      () => injector.prepare(),
      /escapes extension extraction directory/,
    );
    assert.equal(existsSync(path.join(temp_dir, "evil.txt")), false);
  } finally {
    await injector.close();
    rmSync(temp_dir, { recursive: true, force: true });
  }
});

test("LocalBrowserLaunchExtensionInjector prepares an unpacked extension directory for --load-extension", async () => {
  const injector = new LocalBrowserLaunchExtensionInjector({
    injector_extension_path: EXTENSION_PATH,
  });

  try {
    await injector.prepare();
    const unpacked_extension_path = (
      injector as unknown as { unpacked_extension_path?: string | null }
    ).unpacked_extension_path;
    assert.equal(typeof unpacked_extension_path, "string");
    assert.notEqual(unpacked_extension_path, EXTENSION_PATH);
    assert.equal(
      existsSync(path.join(unpacked_extension_path!, "manifest.json")),
      true,
    );
    assert.deepEqual(injector.getLauncherConfig(), {
      extra_args: [`--load-extension=${unpacked_extension_path}`],
    });
    assert.equal(
      injector.options.injector_extension_id,
      DEFAULT_MODCDP_EXTENSION_ID,
    );
  } finally {
    await injector.close();
  }
});

test("LocalBrowserLaunchExtensionInjector prepares the default extension zip for --load-extension", async () => {
  const injector = new LocalBrowserLaunchExtensionInjector();

  try {
    await injector.prepare();
    const unpacked_extension_path = (
      injector as unknown as { unpacked_extension_path?: string | null }
    ).unpacked_extension_path;
    assert.equal(typeof unpacked_extension_path, "string");
    assert.match(unpacked_extension_path!, /modcdp-extension-/);
    assert.equal(
      existsSync(path.join(unpacked_extension_path!, "manifest.json")),
      true,
    );
    assert.deepEqual(injector.getLauncherConfig(), {
      extra_args: [`--load-extension=${unpacked_extension_path}`],
    });
    assert.equal(
      injector.options.injector_extension_id,
      DEFAULT_MODCDP_EXTENSION_ID,
    );
  } finally {
    await injector.close();
  }
});

test("LocalBrowserLaunchExtensionInjector returns immediately when the launched extension target is absent", async () => {
  const methods: string[] = [];
  const injector = new LocalBrowserLaunchExtensionInjector({
    injector_extension_path: EXTENSION_PATH,
    injector_trust_service_worker_target: true,
    send: async (method) => {
      methods.push(method);
      if (method === "Target.getTargets") return { targetInfos: [] };
      throw new Error(`unexpected ${method}`);
    },
  });

  try {
    await injector.prepare();
    const started_at = performance.now();
    const result = await injector.inject();
    const elapsed_ms = performance.now() - started_at;
    assert.equal(result, null);
    assert.deepEqual(methods, ["Target.getTargets"]);
    assert.equal(elapsed_ms < 200, true, `inject() took ${elapsed_ms}ms`);
  } finally {
    await injector.close();
  }
});
