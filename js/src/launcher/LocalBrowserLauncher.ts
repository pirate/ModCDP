import { spawn, type ChildProcess } from "node:child_process";
import { once } from "node:events";
import { existsSync, readdirSync, statSync } from "node:fs";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import net from "node:net";
import type { AddressInfo } from "node:net";
import { homedir, platform, tmpdir } from "node:os";
import path from "node:path";
import {
  BrowserLauncher,
  DEFAULT_CHROME_READY_POLL_INTERVAL_MS,
  DEFAULT_CHROME_READY_TIMEOUT_MS,
  resolveCdpWebSocketUrl,
  type BrowserLaunchOptions,
  type LaunchedBrowser,
} from "./BrowserLauncher.js";

function wildcardToRegExp(value: string) {
  return new RegExp(
    `^${value.replace(/[.+^${}()|[\]\\]/g, "\\$&").replace(/\*/g, ".*")}$`,
  );
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
    return (
      right.version - left.version ||
      right.mtime - left.mtime ||
      a.localeCompare(b)
    );
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
          path.join(
            home,
            "Library/Caches/ms-playwright/chromium-*/chrome-mac*/Chromium.app/Contents/MacOS/Chromium",
          ),
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
            path.join(
              home,
              ".cache/puppeteer/chrome/win*-*/chrome-win*/chrome.exe",
            ),
          ]
        : [
            path.join(
              home,
              ".cache/ms-playwright/chromium-*/chrome-linux*/chrome",
            ),
            "/opt/pw-browsers/chromium-*/chrome-linux*/chrome",
            path.join(
              home,
              ".cache/puppeteer/chrome/linux-*/chrome-linux*/chrome",
            ),
          ];
  return newestFirst(patterns.flatMap(expandGlob));
}

function candidatePaths() {
  const home = homedir();
  const programFiles = [
    process.env.PROGRAMFILES,
    process.env["PROGRAMFILES(X86)"],
  ].filter(Boolean) as string[];
  const canary =
    platform() === "darwin"
      ? [
          "/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
        ]
      : platform() === "win32"
        ? [
            path.join(
              process.env.LOCALAPPDATA || path.join(home, "AppData/Local"),
              "Google/Chrome SxS/Application/chrome.exe",
            ),
            ...programFiles.map((base) =>
              path.join(base, "Google/Chrome SxS/Application/chrome.exe"),
            ),
          ]
        : [
            "/usr/bin/google-chrome-canary",
            "/usr/bin/google-chrome-unstable",
            "/opt/google/chrome-unstable/chrome",
          ];
  const stock =
    platform() === "darwin"
      ? ["/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"]
      : platform() === "win32"
        ? [
            ...programFiles.map((base) =>
              path.join(base, "Google/Chrome/Application/chrome.exe"),
            ),
            path.join(
              process.env.LOCALAPPDATA || path.join(home, "AppData/Local"),
              "Google/Chrome/Application/chrome.exe",
            ),
          ]
        : [
            "/usr/bin/google-chrome-stable",
            "/usr/bin/google-chrome",
            "/opt/google/chrome/chrome",
          ];
  const chromium =
    platform() === "linux"
      ? ["/usr/bin/chromium", "/usr/bin/chromium-browser"]
      : [];
  return [
    process.env.CHROME_PATH,
    ...chromium,
    ...canary,
    ...chromeForTestingCandidates(),
    ...stock,
  ].filter((candidate): candidate is string => Boolean(candidate));
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
  const signalProcess = (signal: NodeJS.Signals) => {
    if (process.platform !== "win32" && proc.pid) {
      try {
        process.kill(-proc.pid, signal);
        return;
      } catch {}
    }
    try {
      proc.kill(signal);
    } catch {}
  };
  signalProcess("SIGTERM");
  await Promise.race([once(proc, "exit"), delay(timeoutMs)]);
  if (proc.exitCode !== null || proc.signalCode !== null) return;
  signalProcess("SIGKILL");
  await Promise.race([once(proc, "exit"), delay(timeoutMs)]);
}

async function removeProfileDir(profile_dir: string) {
  for (let attempt = 0; attempt < 5; attempt++) {
    await rm(profile_dir, { recursive: true, force: true }).catch(() => {});
    if (!existsSync(profile_dir)) return;
    await delay(100 * (attempt + 1));
  }
  try {
    await rm(profile_dir, { recursive: true, force: true });
  } catch {}
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
        if (message.error)
          reject(
            new Error(
              message.error.message ?? "Browser.getVersion failed over pipe",
            ),
          );
        else resolve();
      }
    };
    const timeout = setTimeout(() => {
      cleanup();
      reject(
        new Error(
          `Chrome remote-debugging pipe did not respond within ${timeoutMs}ms`,
        ),
      );
    }, timeoutMs);
    pipe_read.on("data", onData);
    pipe_read.on("error", onError);
    pipe_write.on("error", onError);
    pipe_write.write(
      `${JSON.stringify({ id: readyId, method: "Browser.getVersion" })}\0`,
    );
  });
}

async function waitForCdpWebSocketUrl(
  cdp_url: string,
  timeout_ms: number,
  poll_interval_ms: number,
) {
  const deadline = Date.now() + timeout_ms;
  let lastError: unknown = null;
  while (Date.now() < deadline) {
    try {
      return await resolveCdpWebSocketUrl(cdp_url);
    } catch (error) {
      lastError = error;
      await delay(poll_interval_ms);
    }
  }
  if (lastError instanceof Error) {
    throw new Error(
      `Chrome at ${cdp_url} did not expose a WebSocket CDP URL within ${timeout_ms}ms: ${lastError.message}`,
    );
  }
  throw new Error(
    `Chrome at ${cdp_url} did not expose a WebSocket CDP URL within ${timeout_ms}ms`,
  );
}

async function readDevToolsActivePort(profile_dir: string) {
  const activePortPath = path.join(profile_dir, "DevToolsActivePort");
  let body: string;
  try {
    body = await readFile(activePortPath, "utf8");
  } catch (error) {
    if ((error as NodeJS.ErrnoException).code === "ENOENT") return null;
    throw error;
  }
  const [rawPort, websocketPath] = body.trim().split(/\r?\n/);
  if (!rawPort || !websocketPath) return null;
  const port = Number(rawPort);
  if (!Number.isInteger(port) || port <= 0)
    throw new Error(`Invalid DevToolsActivePort port: ${rawPort}`);
  return { port, cdp_url: `http://127.0.0.1:${port}`, websocketPath };
}

async function waitForBrowserSelectedCdpWebSocketUrl(
  profile_dir: string,
  timeout_ms: number,
  poll_interval_ms: number,
  assertChromeRunning: () => void,
) {
  const deadline = Date.now() + timeout_ms;
  let lastError: unknown = null;
  while (Date.now() < deadline) {
    assertChromeRunning();
    const activePort = await readDevToolsActivePort(profile_dir);
    if (activePort) {
      try {
        return {
          port: activePort.port,
          cdp_url: await resolveCdpWebSocketUrl(activePort.cdp_url),
        };
      } catch (error) {
        lastError = error;
      }
    }
    await delay(poll_interval_ms);
  }
  if (lastError instanceof Error) {
    throw new Error(
      `Chrome did not expose DevToolsActivePort from ${profile_dir} within ${timeout_ms}ms: ${lastError.message}`,
    );
  }
  throw new Error(
    `Chrome did not expose DevToolsActivePort from ${profile_dir} within ${timeout_ms}ms`,
  );
}

export class LocalBrowserLauncher extends BrowserLauncher {
  static findChromeBinary(explicit?: string | null) {
    const candidates = [explicit, ...candidatePaths()].filter(
      (candidate): candidate is string => Boolean(candidate),
    );
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
      sandbox = process.platform !== "linux",
      args = [],
      extra_args = [],
      remote_debugging = "port",
      loopback_cdp = false,
      cleanup_user_data_dir = false,
      chrome_ready_timeout_ms = DEFAULT_CHROME_READY_TIMEOUT_MS,
      chrome_ready_poll_interval_ms = DEFAULT_CHROME_READY_POLL_INTERVAL_MS,
    } = { ...this.options, ...options };
    const exe = LocalBrowserLauncher.findChromeBinary(executable_path);
    const usePipe = remote_debugging === "pipe";
    const useLoopbackCdp = !usePipe || loopback_cdp || port != null;
    const usePort = useLoopbackCdp ? (port ?? 0) : null;
    const profile_dir =
      user_data_dir || (await mkdtemp(path.join(tmpdir(), "modcdp.")));
    const flags = [
      ...DEFAULT_FLAGS,
      headless ? "--headless=new" : null,
      "--disable-gpu",
      sandbox === false ? "--no-sandbox" : null,
      `--user-data-dir=${profile_dir}`,
      useLoopbackCdp ? "--remote-debugging-address=127.0.0.1" : null,
      useLoopbackCdp ? `--remote-debugging-port=${usePort}` : null,
      usePipe ? "--remote-debugging-pipe" : null,
      ...args,
      ...extra_args,
      "about:blank",
    ].filter(Boolean);

    const useStdio = (
      usePipe ? ["ignore", "ignore", "ignore", "pipe", "pipe"] : "ignore"
    ) as
      | "ignore"
      | "inherit"
      | "pipe"
      | import("node:child_process").StdioOptions;
    const proc = spawn(exe, flags, {
      stdio: useStdio,
      detached: process.platform !== "win32",
    });
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
        await removeProfileDir(profile_dir);
    };
    const assertChromeRunning = () => {
      if (spawnError) throw spawnError;
      if (proc.exitCode !== null || proc.signalCode !== null) {
        throw new Error(
          `Chrome exited before CDP became ready (exit=${proc.exitCode}, signal=${proc.signalCode}).`,
        );
      }
    };

    if (usePipe) {
      const pipe_write = proc.stdio[3] as NodeJS.WritableStream | null;
      const pipe_read = proc.stdio[4] as NodeJS.ReadableStream | null;
      if (!pipe_write || !pipe_read) {
        await close();
        throw new Error(
          "Chrome remote-debugging pipe stdio handles were not created.",
        );
      }
      assertChromeRunning();
      await waitForPipeReady(pipe_read, pipe_write, chrome_ready_timeout_ms);
      const loopback =
        usePort == null
          ? null
          : usePort === 0
            ? await waitForBrowserSelectedCdpWebSocketUrl(
                profile_dir,
                chrome_ready_timeout_ms,
                chrome_ready_poll_interval_ms,
                assertChromeRunning,
              )
            : {
                port: usePort,
                cdp_url: await waitForCdpWebSocketUrl(
                  `http://127.0.0.1:${usePort}`,
                  chrome_ready_timeout_ms,
                  chrome_ready_poll_interval_ms,
                ),
              };
      this.launched = {
        proc,
        ...(loopback == null ? {} : { port: loopback.port }),
        cdp_url: `pipe://${proc.pid}`,
        ...(loopback == null ? {} : { loopback_cdp_url: loopback.cdp_url }),
        pipe_read,
        pipe_write,
        profile_dir,
        close,
      };
      return this.launched;
    }

    const deadline = Date.now() + chrome_ready_timeout_ms;
    while (Date.now() < deadline) {
      try {
        assertChromeRunning();
      } catch (error) {
        await close();
        throw error;
      }
      const activePort =
        usePort === 0
          ? await readDevToolsActivePort(profile_dir)
          : { port: usePort as number, cdp_url: `http://127.0.0.1:${usePort}` };
      if (!activePort) {
        await delay(chrome_ready_poll_interval_ms);
        continue;
      }
      try {
        const response = await fetch(`${activePort.cdp_url}/json/version`);
        if (response.ok) {
          const version = await response.json();
          // cdp_url is resolved from the HTTP discovery endpoint before returning.
          this.launched = {
            proc,
            port: activePort.port,
            cdp_url: version.webSocketDebuggerUrl ?? activePort.cdp_url,
            loopback_cdp_url:
              version.webSocketDebuggerUrl ?? activePort.cdp_url,
            profile_dir,
            close,
          };
          return this.launched;
        }
      } catch {}
      await new Promise((resolve) =>
        setTimeout(resolve, chrome_ready_poll_interval_ms),
      );
    }
    await close();
    throw new Error(
      `Chrome did not become ready within ${chrome_ready_timeout_ms}ms`,
    );
  }
}
