import { describe, expect, it } from "vitest";

import { BrowserbaseBrowserLauncher } from "../src/launcher/BrowserbaseBrowserLauncher.js";
import { CdpSocket, expectCdpBrowserSurface } from "./helpers.BrowserLauncher.js";

const LIVE_BROWSERBASE_TIMEOUT_MS = 120_000;

function browserbaseApiUrl(pathname: string) {
  return new URL(
    pathname,
    `${(process.env.BROWSERBASE_BASE_URL ?? "https://api.browserbase.com").replace(/\/$/, "")}/`,
  );
}

async function retrieveBrowserbaseSession(session_id: string) {
  const response = await fetch(browserbaseApiUrl(`/v1/sessions/${session_id}`), {
    headers: { "x-bb-api-key": process.env.BROWSERBASE_API_KEY! },
  });
  expect(response.status).toBeGreaterThanOrEqual(200);
  expect(response.status).toBeLessThan(300);
  return (await response.json()) as Record<string, unknown>;
}

describe("BrowserbaseBrowserLauncher", () => {
  it(
    "creates, verifies, resumes, and releases a real Browserbase browser session",
    { timeout: LIVE_BROWSERBASE_TIMEOUT_MS },
    async () => {
      expect(
        process.env.BROWSERBASE_API_KEY?.trim(),
        "BROWSERBASE_API_KEY is required for live Browserbase tests",
      ).toBeTruthy();
      const launcher = new BrowserbaseBrowserLauncher({
        timeout: 120,
        ...(process.env.BROWSERBASE_REGION ? { region: process.env.BROWSERBASE_REGION } : {}),
        browserbase_browser_settings: {
          viewport: { width: 900, height: 700 },
          recordSession: false,
        },
        browserbase_user_metadata: {
          modcdp_launcher_test: "BrowserbaseBrowserLauncher",
        },
      });
      const browser = await launcher.launch();
      let resumed: Awaited<ReturnType<BrowserbaseBrowserLauncher["launch"]>> | null = null;
      let cdp: CdpSocket | null = null;
      const session_id = browser.browserbase_session_id;

      try {
        expect(session_id).toEqual(expect.any(String));
        expect(browser.browserbase_session_url).toContain(session_id);
        expect(browser.cdp_url).toEqual(expect.stringMatching(/^wss:\/\//));
        cdp = await CdpSocket.connect(browser.cdp_url!);
        await expectCdpBrowserSurface(cdp);

        const retrieved = await retrieveBrowserbaseSession(session_id!);
        expect(retrieved.id).toBe(session_id);
        expect(retrieved.status).toBe("RUNNING");

        resumed = await new BrowserbaseBrowserLauncher({
          browserbase_session_id: session_id,
          browserbase_close_session_on_close: false,
        }).launch();
        expect(resumed.browserbase_session_id).toBe(session_id);
        expect(resumed.cdp_url).toEqual(expect.stringMatching(/^wss:\/\//));
        await expectCdpBrowserSurface(cdp);
      } finally {
        await cdp?.close();
        await resumed?.close();
        await browser.close();
        await browser.close();
      }

      await expect
        .poll(async () => (await retrieveBrowserbaseSession(session_id!)).status, { timeout: 30_000, interval: 1_000 })
        .not.toBe("RUNNING");
    },
  );
});
