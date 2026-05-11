import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { execFileSync } from "node:child_process";
import type { CdpCommandMessage } from "../types/modcdp.js";
import { DEFAULT_MODCDP_EXTENSION_ID } from "../injector/ExtensionInjector.js";
import { UpstreamTransport, type UpstreamTransportConfig } from "./UpstreamTransport.js";

export const DEFAULT_UPSTREAM_NATIVEMESSAGING_HOST_NAME = "com.modcdp.bridge";
export const DEFAULT_UPSTREAM_NATIVEMESSAGING_WAIT_TIMEOUT_MS = 10_000;

type NativeMessagingOptions = {
  upstream_nativemessaging_manifest?: string | null;
  upstream_nativemessaging_manifests?: string[] | null;
  upstream_nativemessaging_host_name?: string | null;
  injector_extension_id?: string | null;
  upstream_nativemessaging_wait_timeout_ms?: number;
};

export class NativeMessagingUpstreamTransport extends UpstreamTransport {
  readonly mode = "nativemessaging" as const;
  readonly endpoint_kind = "modcdp_server" as const;
  url = "";
  private server: any = null;
  private socket: any = null;
  private peer_waiters = new Set<{
    resolve: () => void;
    reject: (error: Error) => void;
    timeout: ReturnType<typeof setTimeout>;
  }>();
  private wait_timeout_ms: number;
  private upstream_nativemessaging_manifest: string | null;
  private upstream_nativemessaging_manifests: string[];
  private include_default_manifest_paths: boolean;
  private upstream_nativemessaging_host_name: string;
  private extension_id: string;
  private user_data_dir: string | null = null;
  private bound_port: number | null = null;
  private cdp_url: string | null = null;

  constructor({
    upstream_nativemessaging_manifest = null,
    upstream_nativemessaging_manifests = null,
    upstream_nativemessaging_host_name = DEFAULT_UPSTREAM_NATIVEMESSAGING_HOST_NAME,
    injector_extension_id = DEFAULT_MODCDP_EXTENSION_ID,
    upstream_nativemessaging_wait_timeout_ms = DEFAULT_UPSTREAM_NATIVEMESSAGING_WAIT_TIMEOUT_MS,
  }: NativeMessagingOptions = {}) {
    super();
    this.upstream_nativemessaging_manifest = upstream_nativemessaging_manifest;
    this.upstream_nativemessaging_manifests = upstream_nativemessaging_manifests ?? [];
    this.include_default_manifest_paths =
      !upstream_nativemessaging_manifest && !upstream_nativemessaging_manifests?.length;
    this.upstream_nativemessaging_host_name =
      upstream_nativemessaging_host_name || DEFAULT_UPSTREAM_NATIVEMESSAGING_HOST_NAME;
    this.extension_id = injector_extension_id || DEFAULT_MODCDP_EXTENSION_ID;
    this.wait_timeout_ms = upstream_nativemessaging_wait_timeout_ms;
  }

  update(config: UpstreamTransportConfig = {}) {
    let should_install_native_host = false;
    if (config.upstream_nativemessaging_manifest !== undefined) {
      this.upstream_nativemessaging_manifest = config.upstream_nativemessaging_manifest;
      should_install_native_host = true;
    }
    if (config.upstream_nativemessaging_manifests !== undefined) {
      this.upstream_nativemessaging_manifests = config.upstream_nativemessaging_manifests ?? [];
      should_install_native_host = true;
    }
    this.include_default_manifest_paths =
      !this.upstream_nativemessaging_manifest && this.upstream_nativemessaging_manifests.length === 0;
    if (config.upstream_nativemessaging_host_name) {
      this.upstream_nativemessaging_host_name = config.upstream_nativemessaging_host_name;
      should_install_native_host = true;
    }
    if (typeof config.upstream_nativemessaging_wait_timeout_ms === "number")
      this.wait_timeout_ms = config.upstream_nativemessaging_wait_timeout_ms;
    if (config.injector_extension_id) {
      this.extension_id = config.injector_extension_id;
      should_install_native_host = true;
    }
    if (config.user_data_dir && config.user_data_dir !== this.user_data_dir) {
      this.setProfileManifestPaths(config.user_data_dir);
      this.user_data_dir = config.user_data_dir;
      should_install_native_host = true;
    }
    if (should_install_native_host && this.bound_port != null) this.installNativeHost(this.bound_port);
    this.cdp_url = config.cdp_url ?? this.cdp_url;
    return this;
  }

  getServerConfig() {
    return this.cdp_url ? { server_loopback_cdp_url: this.cdp_url } : {};
  }

  getInjectorConfig() {
    return {
      upstream_nativemessaging_host_name: this.upstream_nativemessaging_host_name,
    };
  }

  async connect() {
    if (typeof process !== "object" || !process?.versions?.node) {
      throw new Error("upstream.upstream_mode=nativemessaging requires Node.");
    }
    const net = await import("node:net");
    const server = net.createServer((socket) => this.accept(socket));
    this.server = server;
    await new Promise<void>((resolve, reject) => {
      server.once("error", reject);
      server.listen(0, "127.0.0.1", () => resolve());
    });
    const address = server.address();
    if (!address || typeof address === "string") throw new Error("Native messaging bridge did not bind a TCP port.");
    this.url = `native://${this.upstream_nativemessaging_host_name}@127.0.0.1:${address.port}`;
    this.bound_port = address.port;
    this.installNativeHost(address.port);
  }

  send(message: CdpCommandMessage) {
    if (!this.socket || this.socket.destroyed)
      throw new Error(`No native messaging peer is connected for ${this.upstream_nativemessaging_host_name}.`);
    writeLengthPrefixedJSON(this.socket, message);
  }

  async waitForPeer() {
    if (this.socket && !this.socket.destroyed) return;
    await new Promise<void>((resolve, reject) => {
      let waiter!: {
        resolve: () => void;
        reject: (error: Error) => void;
        timeout: ReturnType<typeof setTimeout>;
      };
      const timeout = setTimeout(() => {
        this.peer_waiters.delete(waiter);
        reject(
          new Error(
            `Timed out waiting ${this.wait_timeout_ms}ms for native messaging host ${this.upstream_nativemessaging_host_name}.`,
          ),
        );
      }, this.wait_timeout_ms);
      waiter = { resolve, reject, timeout };
      this.peer_waiters.add(waiter);
    });
  }

  async close() {
    try {
      this.socket?.destroy();
    } catch {}
    this.socket = null;
    if (this.server) await new Promise<void>((resolve) => this.server.close(() => resolve()));
    this.server = null;
    for (const waiter of this.peer_waiters) {
      clearTimeout(waiter.timeout);
      waiter.reject(
        new Error(
          `Native messaging transport for ${this.upstream_nativemessaging_host_name} closed before a peer connected.`,
        ),
      );
    }
    this.peer_waiters.clear();
  }

  private accept(socket: any) {
    if (this.socket && this.socket !== socket) {
      try {
        this.socket.destroy();
      } catch {}
    }
    this.socket = socket;
    const readMessage = createLengthPrefixedJSONReader((message) => {
      if (message?.type === "modcdp.native.hello") return;
      this.parseAndEmitRecv(JSON.stringify(message));
    });
    socket.on("data", readMessage);
    socket.on("close", () => {
      if (this.socket !== socket) return;
      this.socket = null;
      this.emitClose(new Error("Native messaging host disconnected"));
    });
    socket.on("error", () => {
      if (this.socket !== socket) return;
      this.socket = null;
      this.emitClose(new Error("Native messaging host error"));
    });
    for (const waiter of this.peer_waiters) {
      clearTimeout(waiter.timeout);
      waiter.resolve();
    }
    this.peer_waiters.clear();
  }

  private installNativeHost(port: number) {
    const hostDir = path.join(os.homedir(), ".modcdp", "native-messaging");
    fs.mkdirSync(hostDir, { recursive: true });

    const configPath = path.join(hostDir, `${this.upstream_nativemessaging_host_name}.config.json`);
    const hostScriptPath = path.join(hostDir, `${this.upstream_nativemessaging_host_name}.mjs`);
    const hostExecutablePath = path.join(
      hostDir,
      `${this.upstream_nativemessaging_host_name}${process.platform === "win32" ? ".cmd" : ".sh"}`,
    );
    fs.writeFileSync(configPath, JSON.stringify({ host: "127.0.0.1", port }, null, 2));
    fs.writeFileSync(hostScriptPath, nativeHostScript(configPath));
    fs.writeFileSync(hostExecutablePath, nativeHostWrapper(process.execPath, hostScriptPath));
    fs.chmodSync(hostExecutablePath, 0o755);

    const manifestPaths =
      this.upstream_nativemessaging_manifest || this.upstream_nativemessaging_manifests.length > 0
        ? [
            ...(this.upstream_nativemessaging_manifest ? [this.upstream_nativemessaging_manifest] : []),
            ...this.upstream_nativemessaging_manifests,
            ...(this.include_default_manifest_paths
              ? defaultNativeMessagingManifestPaths(this.upstream_nativemessaging_host_name, os.homedir())
              : []),
          ]
        : defaultNativeMessagingManifestPaths(this.upstream_nativemessaging_host_name, os.homedir());
    const manifest = JSON.stringify(
      {
        name: this.upstream_nativemessaging_host_name,
        description: "ModCDP Native Messaging bridge",
        path: hostExecutablePath,
        type: "stdio",
        allowed_origins: [`chrome-extension://${this.extension_id}/`],
      },
      null,
      2,
    );
    for (const manifestPath of manifestPaths) {
      fs.mkdirSync(path.dirname(manifestPath), { recursive: true });
      fs.writeFileSync(manifestPath, manifest);
    }
    if (process.platform === "win32" && manifestPaths[0]) {
      registerWindowsNativeMessagingHost(this.upstream_nativemessaging_host_name, manifestPaths[0]);
    }
  }

  private setProfileManifestPaths(user_data_dir: string) {
    const previous_profile_manifest_paths = this.user_data_dir
      ? [
          path.join(this.user_data_dir, "NativeMessagingHosts", `${this.upstream_nativemessaging_host_name}.json`),
          path.join(
            this.user_data_dir,
            "Default",
            "NativeMessagingHosts",
            `${this.upstream_nativemessaging_host_name}.json`,
          ),
        ]
      : [];
    const profile_manifest_paths = [
      path.join(user_data_dir, "NativeMessagingHosts", `${this.upstream_nativemessaging_host_name}.json`),
      path.join(user_data_dir, "Default", "NativeMessagingHosts", `${this.upstream_nativemessaging_host_name}.json`),
    ];
    this.upstream_nativemessaging_manifests = [
      ...profile_manifest_paths,
      ...this.upstream_nativemessaging_manifests.filter(
        (upstream_nativemessaging_manifest) =>
          !previous_profile_manifest_paths.includes(upstream_nativemessaging_manifest) &&
          !profile_manifest_paths.includes(upstream_nativemessaging_manifest),
      ),
    ];
  }
}

function defaultNativeMessagingManifestPaths(upstream_nativemessaging_host_name: string, home: string) {
  if (process.platform === "darwin") {
    return [
      `${home}/Library/Application Support/Google/Chrome/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
      `${home}/Library/Application Support/Google/Chrome Canary/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
      `${home}/Library/Application Support/Google/ChromeForTesting/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
      `${home}/Library/Application Support/Google/Chrome for Testing/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
      `${home}/Library/Application Support/Google/Chrome SxS/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
      `${home}/Library/Application Support/Chromium/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
    ];
  }
  if (process.platform === "linux") {
    return [
      `${home}/.config/google-chrome/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
      `${home}/.config/google-chrome-for-testing/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
      `${home}/.config/chromium/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
      `${home}/.config/chromium-browser/NativeMessagingHosts/${upstream_nativemessaging_host_name}.json`,
    ];
  }
  if (process.platform === "win32") {
    return [path.join(home, ".modcdp", "native-messaging", `${upstream_nativemessaging_host_name}.json`)];
  }
  throw new Error("upstream_nativemessaging_manifest is required on this platform.");
}

function nativeHostWrapper(node_path: string, host_script_path: string) {
  if (process.platform === "win32") {
    return `@echo off\r\n${cmdQuote(node_path)} ${cmdQuote(host_script_path)}\r\n`;
  }
  return `#!/bin/sh\nexec ${JSON.stringify(node_path)} ${JSON.stringify(host_script_path)}\n`;
}

function cmdQuote(value: string) {
  return `"${value.replace(/"/g, '""')}"`;
}

function registerWindowsNativeMessagingHost(
  upstream_nativemessaging_host_name: string,
  upstream_nativemessaging_manifest: string,
) {
  execFileSync(
    "reg",
    [
      "add",
      `HKCU\\Software\\Google\\Chrome\\NativeMessagingHosts\\${upstream_nativemessaging_host_name}`,
      "/ve",
      "/t",
      "REG_SZ",
      "/d",
      upstream_nativemessaging_manifest,
      "/f",
    ],
    { stdio: "ignore" },
  );
}

function writeLengthPrefixedJSON(stream: { write: (chunk: Buffer) => void }, message: unknown) {
  const body = Buffer.from(JSON.stringify(message), "utf8");
  const header = Buffer.alloc(4);
  header.writeUInt32LE(body.length, 0);
  stream.write(Buffer.concat([header, body]));
}

function createLengthPrefixedJSONReader(onMessage: (message: any) => void) {
  let buffer = Buffer.alloc(0);
  return (chunk: Buffer) => {
    buffer = Buffer.concat([buffer, chunk]);
    while (buffer.length >= 4) {
      const length = buffer.readUInt32LE(0);
      if (buffer.length < length + 4) return;
      const body = buffer.subarray(4, 4 + length);
      buffer = buffer.subarray(4 + length);
      try {
        onMessage(JSON.parse(body.toString("utf8")));
      } catch {}
    }
  };
}

function nativeHostScript(configPath: string) {
  return `
import fs from "node:fs";
import net from "node:net";

const config = JSON.parse(fs.readFileSync(${JSON.stringify(configPath)}, "utf8"));
const socket = net.createConnection({ host: config.host, port: config.port }, () => {
  writeTCP({ type: "modcdp.native.hello", role: "native-host", version: 1 });
});
let stdinBuffer = Buffer.alloc(0);
let socketBuffer = Buffer.alloc(0);

process.stdin.on("data", (chunk) => {
  stdinBuffer = Buffer.concat([stdinBuffer, chunk]);
  stdinBuffer = readMessages(stdinBuffer, (message) => {
    writeTCP(message);
  });
});
socket.on("data", (chunk) => {
  socketBuffer = Buffer.concat([socketBuffer, chunk]);
  socketBuffer = readMessages(socketBuffer, (message) => {
    writeNative(message);
  });
});
socket.on("close", () => process.exit(0));
socket.on("error", () => process.exit(1));

function readMessages(buffer, onRecv) {
  while (buffer.length >= 4) {
    const length = buffer.readUInt32LE(0);
    if (buffer.length < length + 4) return buffer;
    const body = buffer.subarray(4, 4 + length);
    buffer = buffer.subarray(4 + length);
    try {
      onRecv(JSON.parse(body.toString("utf8")));
    } catch {}
  }
  return buffer;
}

function writeNative(message) {
  const body = Buffer.from(JSON.stringify(message), "utf8");
  const header = Buffer.alloc(4);
  header.writeUInt32LE(body.length, 0);
  process.stdout.write(Buffer.concat([header, body]));
}

function writeTCP(message) {
  const body = Buffer.from(JSON.stringify(message), "utf8");
  const header = Buffer.alloc(4);
  header.writeUInt32LE(body.length, 0);
  socket.write(Buffer.concat([header, body]));
}
`;
}
