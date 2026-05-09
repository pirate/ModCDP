import { describe, expect, it } from "vitest";

import { BrowserLauncher } from "../bridge/BrowserLauncher.js";

describe("BrowserLauncher", () => {
  it("merges launch config and exposes transport/injector config without launching a browser", async () => {
    const launcher = new BrowserLauncher({
      cdp_url: "http://127.0.0.1:9222",
      ws_url: "ws://127.0.0.1:9222/devtools/browser/initial",
      user_data_dir: "/tmp/modcdp-browser-launcher",
      browserbase_api_key: "test-key",
      extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      extra_args: ["--load-extension=/tmp/one"],
    });
    launcher.update({
      ws_url: "ws://127.0.0.1:9222/devtools/browser/updated",
      extra_args: ["--load-extension=/tmp/two", "--window-size=900,700"],
    });

    expect(launcher.options.extra_args).toEqual(["--window-size=900,700", "--load-extension=/tmp/one,/tmp/two"]);
    expect(launcher.getTransportConfig()).toMatchObject({
      cdp_url: "http://127.0.0.1:9222",
      ws_url: "ws://127.0.0.1:9222/devtools/browser/updated",
      user_data_dir: "/tmp/modcdp-browser-launcher",
    });
    expect(launcher.getInjectorConfig()).toMatchObject({
      browserbase_api_key: "test-key",
      extension_id: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
    });
    await expect(launcher.launch()).rejects.toThrow("BrowserLauncher.launch is not implemented.");
  });
});
