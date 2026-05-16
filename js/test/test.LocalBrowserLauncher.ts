import { mkdtemp, rm, stat } from "node:fs/promises";
import { tmpdir } from "node:os";
import path from "node:path";
import { describe, expect, it } from "vitest";

import { LocalBrowserLauncher } from "../src/launcher/LocalBrowserLauncher.js";
import {
  CdpSocket,
  expectCdpBrowserSurface,
  expectHttpEndpointDown,
  PipeCdpSocket,
} from "./helpers.BrowserLauncher.js";

const LIVE_BROWSER_TIMEOUT_MS = 60_000;

describe("LocalBrowserLauncher", () => {
  it("class helpers match the local launcher surface", async () => {
    expect(LocalBrowserLauncher.findChromeBinary()).toEqual(expect.any(String));
    expect(await LocalBrowserLauncher.freePort()).toEqual(expect.any(Number));
  });

  it(
    "launches a real browser over a chosen CDP port and honors launch options",
    { timeout: LIVE_BROWSER_TIMEOUT_MS },
    async () => {
      const userDataDir = await mkdtemp(path.join(tmpdir(), "modcdp-local-profile-"));
      const port = await LocalBrowserLauncher.freePort();
      const chrome = await new LocalBrowserLauncher({
        headless: true,
        chrome_ready_timeout_ms: 45_000,
        chrome_ready_poll_interval_ms: 50,
      }).launch({
        port,
        user_data_dir: userDataDir,
        extra_args: ["--window-size=900,700"],
      });
      let cdp: CdpSocket | null = null;

      try {
        expect(chrome.port).toBe(port);
        expect(chrome.cdp_url).toEqual(expect.stringMatching(new RegExp(`^ws://127\\.0\\.0\\.1:${port}/`)));
        expect(chrome.profile_dir).toBe(userDataDir);
        expect((chrome.proc as { spawnargs?: string[] }).spawnargs ?? []).toEqual(
          expect.arrayContaining([
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
            "--window-size=900,700",
          ]),
        );
        if (process.platform === "linux") {
          expect((chrome.proc as { spawnargs?: string[] }).spawnargs ?? []).toContain("--no-sandbox");
        } else {
          expect((chrome.proc as { spawnargs?: string[] }).spawnargs ?? []).not.toContain("--no-sandbox");
        }
        await expect(stat(userDataDir)).resolves.toBeTruthy();
        cdp = await CdpSocket.connect(chrome.cdp_url!);
        const systemInfo = await cdp.send("SystemInfo.getInfo");
        const commandLine = systemInfo.commandLine;
        expect(commandLine).toEqual(expect.any(String));
        expect(commandLine).toContain("--window-size=900,700");
        if (process.platform === "linux") {
          expect(commandLine).toContain("--no-sandbox");
        } else {
          expect(commandLine).not.toContain("--no-sandbox");
        }
        await expectCdpBrowserSurface(cdp);
      } finally {
        await cdp?.close();
        await chrome.close();
        await chrome.close();
        await expectHttpEndpointDown(`http://127.0.0.1:${port}`);
        await expect(stat(userDataDir)).resolves.toBeTruthy();
        await rm(userDataDir, { recursive: true, force: true });
      }
    },
  );

  it(
    "launches a real browser over remote-debugging-pipe and speaks CDP over the returned pipes",
    { timeout: LIVE_BROWSER_TIMEOUT_MS },
    async () => {
      const chrome = await new LocalBrowserLauncher().launch({
        headless: true,
        remote_debugging: "pipe",
        chrome_ready_timeout_ms: 45_000,
      });
      const profile_dir = chrome.profile_dir;

      try {
        expect(chrome.port).toBeUndefined();
        expect(chrome.cdp_url).toEqual(expect.stringMatching(/^pipe:\/\/\d+/));
        expect(chrome.loopback_cdp_url).toBeUndefined();
        expect(chrome.pipe_read).toBeTruthy();
        expect(chrome.pipe_write).toBeTruthy();
        const pipeCdp = new PipeCdpSocket(chrome.pipe_read!, chrome.pipe_write!);
        await expectCdpBrowserSurface(pipeCdp);
      } finally {
        await chrome.close();
        await chrome.close();
      }

      if (profile_dir) {
        await expect(stat(profile_dir)).rejects.toMatchObject({
          code: "ENOENT",
        });
      }
    },
  );

  it(
    "launches a pipe browser with an auxiliary loopback CDP endpoint only when requested",
    { timeout: LIVE_BROWSER_TIMEOUT_MS },
    async () => {
      const chrome = await new LocalBrowserLauncher().launch({
        headless: true,
        remote_debugging: "pipe",
        loopback_cdp: true,
        chrome_ready_timeout_ms: 45_000,
      });
      let cdp: CdpSocket | null = null;

      try {
        expect(chrome.cdp_url).toEqual(expect.stringMatching(/^pipe:\/\/\d+/));
        expect(chrome.port).toEqual(expect.any(Number));
        expect(chrome.loopback_cdp_url).toEqual(expect.stringMatching(/^ws:\/\/127\.0\.0\.1:\d+\//));
        cdp = await CdpSocket.connect(chrome.loopback_cdp_url!);
        await expectCdpBrowserSurface(cdp);
      } finally {
        await cdp?.close();
        await chrome.close();
      }
    },
  );

  it(
    "removes an explicit user data dir when cleanup_user_data_dir is set",
    { timeout: LIVE_BROWSER_TIMEOUT_MS },
    async () => {
      const userDataDir = await mkdtemp(path.join(tmpdir(), "modcdp-local-profile-"));
      const chrome = await new LocalBrowserLauncher({
        headless: true,
        chrome_ready_timeout_ms: 45_000,
      }).launch({
        user_data_dir: userDataDir,
        cleanup_user_data_dir: true,
      });

      try {
        expect(chrome.profile_dir).toBe(userDataDir);
        await expect(stat(userDataDir)).resolves.toBeTruthy();
      } finally {
        await chrome.close();
      }

      await expect(stat(userDataDir)).rejects.toMatchObject({ code: "ENOENT" });
    },
  );
});
