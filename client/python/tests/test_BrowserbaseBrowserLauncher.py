from __future__ import annotations

import json
import os
import time
import unittest
import urllib.request

from websocket import create_connection

from modcdp.BrowserbaseBrowserLauncher import BrowserbaseBrowserLauncher


LIVE_BROWSERBASE_TIMEOUT_S = 120


@unittest.skipUnless(os.environ.get("BROWSERBASE_API_KEY", "").strip(), "BROWSERBASE_API_KEY is required for live Browserbase tests")
class BrowserbaseBrowserLauncherTests(unittest.TestCase):
    def test_creates_verifies_resumes_and_releases_real_browserbase_session(self) -> None:
        launcher = BrowserbaseBrowserLauncher(
            {
                "project_id": os.environ.get("BROWSERBASE_PROJECT_ID"),
                "timeout": 120,
                **({"region": os.environ["BROWSERBASE_REGION"]} if os.environ.get("BROWSERBASE_REGION") else {}),
                "browser_settings": {
                    "viewport": {"width": 900, "height": 700},
                    "recordSession": False,
                },
                "user_metadata": {
                    "modcdp_launcher_test": "BrowserbaseBrowserLauncher",
                },
            }
        )
        browser = launcher.launch()
        resumed = None
        ws = None
        session_id = browser.get("browserbase_session_id")
        try:
            if not isinstance(session_id, str):
                self.fail(f"browserbase_session_id = {session_id!r}")
            self.assertIn(session_id, browser.get("browserbase_session_url") or "")
            ws_url = browser.get("ws_url")
            if not isinstance(ws_url, str):
                self.fail(f"ws_url = {ws_url!r}")
            self.assertRegex(ws_url, r"^wss://")
            self.assertEqual(browser.get("cdp_url"), browser.get("ws_url"))
            ws = create_connection(ws_url, timeout=LIVE_BROWSERBASE_TIMEOUT_S)
            _expect_cdp_browser_surface(ws)

            retrieved = _retrieve_browserbase_session(session_id)
            self.assertEqual(retrieved.get("id"), session_id)
            self.assertEqual(retrieved.get("status"), "RUNNING")
            if os.environ.get("BROWSERBASE_PROJECT_ID"):
                self.assertEqual(retrieved.get("projectId"), os.environ["BROWSERBASE_PROJECT_ID"])

            resumed = BrowserbaseBrowserLauncher(
                {
                    "session_id": session_id,
                    "close_session_on_close": False,
                }
            ).launch()
            self.assertEqual(resumed.get("browserbase_session_id"), session_id)
            self.assertRegex(resumed.get("ws_url") or "", r"^wss://")
            _expect_cdp_browser_surface(ws)
        finally:
            if ws is not None:
                ws.close()
            if resumed is not None:
                resumed["close"]()
            browser["close"]()
            browser["close"]()

        deadline = time.time() + 30
        while time.time() < deadline:
            if _retrieve_browserbase_session(session_id).get("status") != "RUNNING":
                return
            time.sleep(1)
        self.fail("Browserbase session did not leave RUNNING status after release")


def _retrieve_browserbase_session(session_id: str) -> dict:
    request = urllib.request.Request(
        _browserbase_api_url(f"/v1/sessions/{session_id}"),
        headers={"x-bb-api-key": os.environ["BROWSERBASE_API_KEY"]},
    )
    with urllib.request.urlopen(request, timeout=60) as response:
        if response.status < 200 or response.status >= 300:
            raise AssertionError(f"Browserbase session fetch returned {response.status}")
        return json.loads(response.read())


def _browserbase_api_url(pathname: str) -> str:
    base_url = os.environ.get("BROWSERBASE_BASE_URL", "https://api.browserbase.com").rstrip("/")
    return f"{base_url}/{pathname.lstrip('/')}"


def _expect_cdp_browser_surface(ws) -> None:
    ws.send(json.dumps({"id": 1, "method": "Browser.getVersion", "params": {}}))
    message = json.loads(ws.recv())
    if not isinstance(message.get("result", {}).get("product"), str):
        raise AssertionError(f"Browser.getVersion result = {message!r}")


if __name__ == "__main__":
    unittest.main()
