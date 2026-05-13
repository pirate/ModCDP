import { BrowserLauncher, type BrowserLaunchOptions, type LaunchedBrowser } from "./BrowserLauncher.js";

export class NoopBrowserLauncher extends BrowserLauncher {
  async launch(_options: BrowserLaunchOptions = {}): Promise<LaunchedBrowser> {
    this.launched = { cdp_url: null, close: async () => {} };
    return this.launched;
  }
}
