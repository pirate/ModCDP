import { spawn, type ChildProcess } from "node:child_process";
import { once } from "node:events";
import { existsSync, readdirSync, statSync } from "node:fs";
import { mkdtemp, rm } from "node:fs/promises";
import net from "node:net";
import type { AddressInfo } from "node:net";
import { homedir, platform, tmpdir } from "node:os";
import path from "node:path";
import {
  BrowserLauncher,
  DEFAULT_CHROME_READY_POLL_INTERVAL_MS,
  DEFAULT_CHROME_READY_TIMEOUT_MS,
  type BrowserLaunchOptions,
  type LaunchedBrowser,
} from "./BrowserLauncher.js";

function wildcardToRegExp(value: string) {
  return new RegExp(`^${value.replace(/[.+^${}()|[\]\\]/g, "\\$&").replace(/\*/g, ".*")}$`);
}

function expandGlob(pattern: string) {
  const normalized = path.normalize(pattern);
  const { root } = path.parse(normalized);
  const parts = normalized.slice(root.length).split(path.sep).filter(Boolean);
  let candidates = [root || "."];
  for (const part of parts) {
    const hasWildcard = part.includes("*");
    const matcher = hasWildcard ? wildcardToRegExp(part) : null;
    const next: string[] = [];
    for (const base of candidates) {
      if (!existsSync(base)) continue;
      if (!hasWildcard) {
        const candidate = path.join(base, part);
        if (existsSync(candidate)) next.push(candidate);
        continue;
      }
      try {
        for (const child of readdirSync(base)) {
          if (matcher!.test(child)) next.push(path.join(base, child));
        }
      } catch {}
    }
    candidates = next;
  }
  return candidates.filter((candidate) => existsSync(candidate));
}

function newestFirst(candidates: string[]) {
  const score = (candidate: string) => {
    const numbers = candidate.match(/\d+/g)?.map(Number) ?? [];
    const version = numbers.length > 0 ? Math.max(...numbers) : 0;
    let mtime = 0;
    try {
      mtime = statSync(candidate).mtimeMs;
    } catch {}
    return { version, mtime };
  };
  return [...new Set(candidates)].sort((a, b) => {
    const left = score(a);
    const right = score(b);
    return right.version - left.version || right.mtime - left.mtime || a.localeCompare(b);
  });
}

function chromeForTestingCandidates() {
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
  return newestFirst(patterns.flatMap(expandGlob));
}

function candidatePaths() {
  const home = homedir();
  const programFiles = [process.env.PROGRAMFILES, process.env["PROGRAMFILES(X86)"]].filter(Boolean) as string[];
  const canary =
    platform() === "darwin"
      ? ["/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary"]
      : platform() === "win32"
        ? [
            path.join(
              process.env.LOCALAPPDATA || path.join(home, "AppData/Local"),
              "Google/Chrome SxS/Application/chrome.exe",
            ),
            ...programFiles.map((base) => path.join(base, "Google/Chrome SxS/Application/chrome.exe")),
          ]
        : ["/usr/bin/google-chrome-canary", "/usr/bin/google-chrome-unstable", "/opt/google/chrome-unstable/chrome"];
  const stock =
    platform() === "darwin"
      ? ["/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"]
      : platform() === "win32"
        ? [
            ...programFiles.map((base) => path.join(base, "Google/Chrome/Application/chrome.exe")),
            path.join(
              process.env.LOCALAPPDATA || path.join(home, "AppData/Local"),
              "Google/Chrome/Application/chrome.exe",
            ),
          ]
        : ["/usr/bin/google-chrome-stable", "/usr/bin/google-chrome", "/opt/google/chrome/chrome"];
  return [process.env.CHROME_PATH, ...chromeForTestingCandidates(), ...canary, ...stock].filter(
    (candidate): candidate is string => Boolean(candidate),
  );
}

const DEFAULT_FLAGS = [
  "--enable-unsafe-extension-debugging",
  "--remote-allow-origins=*",
  "--no-first-run",
  "--no-default-browser-check",
  "--disable-default-apps",
  "--disable-dev-shm-usage",
  "--disable-background-networking",
  "--disable-backgrounding-occluded-windows",
  "--disable-renderer-backgrounding",
  "--disable-background-timer-throttling",
  "--disable-sync",
  "--disable-features=DisableLoadExtensionCommandLineSwitch",
  "--password-store=basic",
  "--use-mock-keychain",
];

function delay(ms: number) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function terminateProcess(proc: ChildProcess, timeoutMs = 2_000) {
  if (proc.exitCode !== null || proc.signalCode !== null) return;
  try {
    proc.kill("SIGTERM");
  } catch {}
  await Promise.race([once(proc, "exit"), delay(timeoutMs)]);
  if (proc.exitCode !== null || proc.signalCode !== null) return;
  try {
    proc.kill("SIGKILL");
  } catch {}
  await Promise.race([once(proc, "exit"), delay(timeoutMs)]);
}

async function waitForPipeReady(
  pipe_read: NodeJS.ReadableStream,
  pipe_write: NodeJS.WritableStream,
  timeoutMs: number,
) {
  let buffer = "";
  const readyId = 1;
  await new Promise<void>((resolve, reject) => {
    const cleanup = () => {
      clearTimeout(timeout);
      pipe_read.off("data", onData);
      pipe_read.off("error", onError);
      pipe_write.off("error", onError);
    };
    const onError = (error: Error) => {
      cleanup();
      reject(error);
    };
    const onData = (chunk: Buffer | string) => {
      buffer += chunk.toString();
      while (buffer.includes("\0")) {
        const [raw, ...rest] = buffer.split("\0");
        buffer = rest.join("\0");
        if (!raw) continue;
        const message = JSON.parse(raw);
        if (message.id !== readyId) continue;
        cleanup();
        if (message.error) reject(new Error(message.error.message ?? "Browser.getVersion failed over pipe"));
        else resolve();
      }
    };
    const timeout = setTimeout(() => {
      cleanup();
      reject(new Error(`Chrome remote-debugging pipe did not respond within ${timeoutMs}ms`));
    }, timeoutMs);
    pipe_read.on("data", onData);
    pipe_read.on("error", onError);
    pipe_write.on("error", onError);
    pipe_write.write(`${JSON.stringify({ id: readyId, method: "Browser.getVersion" })}\0`);
  });
}

export class LocalBrowserLauncher extends BrowserLauncher {
  static findChromeBinary(explicit?: string | null) {
    const candidates = [explicit, ...candidatePaths()].filter((candidate): candidate is string => Boolean(candidate));
    for (const candidate of candidates) {
      if (candidate && existsSync(candidate)) return candidate;
    }
    throw new Error(
      `No Chrome/Chromium binary found. Tried: ${candidates.join(", ")}. Set CHROME_PATH or pass executable_path.`,
    );
  }

  static async freePort() {
    const server = net.createServer();
    await new Promise<void>((resolve, reject) => {
      server.listen(0, "127.0.0.1", () => resolve());
      server.once("error", reject);
    });
    const { port } = server.address() as AddressInfo;
    await new Promise<void>((resolve) => server.close(() => resolve()));
    return port;
  }

  async launch(options: BrowserLaunchOptions = {}): Promise<LaunchedBrowser> {
    const {
      executable_path,
      port,
      user_data_dir,
      headless = process.platform === "linux" && !process.env.DISPLAY,
      sandbox = false,
      args = [],
      extra_args = [],
      stdio = "ignore",
      remote_debugging = "port",
      cleanup_user_data_dir = false,
      chrome_ready_timeout_ms = DEFAULT_CHROME_READY_TIMEOUT_MS,
      chrome_ready_poll_interval_ms = DEFAULT_CHROME_READY_POLL_INTERVAL_MS,
    } = { ...this.options, ...options };
    const exe = LocalBrowserLauncher.findChromeBinary(executable_path);
    const usePipe = remote_debugging === "pipe";
    const usePort = usePipe ? null : port || (await LocalBrowserLauncher.freePort());
    const profile_dir = user_data_dir || (await mkdtemp(path.join(tmpdir(), "modcdp.")));
    const needsNoSandbox = headless && process.platform === "linux" && !process.env.DISPLAY && sandbox !== true;
    const flags = [
      ...DEFAULT_FLAGS,
      headless ? "--headless=new" : null,
      "--disable-gpu",
      needsNoSandbox ? "--no-sandbox" : null,
      `--user-data-dir=${profile_dir}`,
      usePipe ? "--remote-debugging-pipe" : "--remote-debugging-address=127.0.0.1",
      usePipe ? null : `--remote-debugging-port=${usePort}`,
      ...args,
      ...extra_args,
      "about:blank",
    ].filter(Boolean);

    const useStdio = (usePipe && stdio === "ignore" ? ["ignore", "ignore", "ignore", "pipe", "pipe"] : stdio) as
      | "ignore"
      | "inherit"
      | "pipe"
      | import("node:child_process").StdioOptions;
    const proc = spawn(exe, flags, { stdio: useStdio, detached: false });
    let spawnError: Error | null = null;
    proc.once("error", (error) => {
      spawnError = error;
    });
    let closed = false;
    const close = async () => {
      if (closed) return;
      closed = true;
      await terminateProcess(proc);
      if (!user_data_dir || cleanup_user_data_dir)
        await rm(profile_dir, { recursive: true, force: true }).catch(() => {});
    };

    if (usePipe) {
      const pipe_write = proc.stdio[3] as NodeJS.WritableStream | null;
      const pipe_read = proc.stdio[4] as NodeJS.ReadableStream | null;
      if (!pipe_write || !pipe_read) {
        await close();
        throw new Error("Chrome remote-debugging pipe stdio handles were not created.");
      }
      if (spawnError) {
        await close();
        throw spawnError;
      }
      await waitForPipeReady(pipe_read, pipe_write, chrome_ready_timeout_ms);
      this.launched = {
        proc,
        port: undefined,
        cdp_url: `pipe://${proc.pid}`,
        ws_url: null,
        pipe_read,
        pipe_write,
        profile_dir,
        close,
      };
      return this.launched;
    }

    const cdp_url = `http://127.0.0.1:${usePort}`;
    const deadline = Date.now() + chrome_ready_timeout_ms;
    while (Date.now() < deadline) {
      if (spawnError) {
        await close();
        throw spawnError;
      }
      if (proc.exitCode !== null || proc.signalCode !== null) {
        await close();
        throw new Error(`Chrome exited before CDP became ready (exit=${proc.exitCode}, signal=${proc.signalCode}).`);
      }
      try {
        const response = await fetch(`${cdp_url}/json/version`);
        if (response.ok) {
          const version = await response.json();
          this.launched = { proc, port: usePort, cdp_url, ws_url: version.webSocketDebuggerUrl, profile_dir, close };
          return this.launched;
        }
      } catch {}
      await new Promise((resolve) => setTimeout(resolve, chrome_ready_poll_interval_ms));
    }
    await close();
    throw new Error(`Chrome at ${cdp_url} did not become ready within ${chrome_ready_timeout_ms}ms`);
  }
}
