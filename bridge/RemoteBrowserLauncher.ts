import {
  BrowserLauncher,
  resolveCdpWebSocketUrl,
  type BrowserLaunchOptions,
  type LaunchedBrowser,
} from "./BrowserLauncher.js";

export class RemoteBrowserLauncher extends BrowserLauncher {
  constructor(options: BrowserLaunchOptions = {}, cdp_url: string | null = null) {
    super({ ...options, ...(cdp_url == null ? {} : { cdp_url }) });
  }

  async launch(options: BrowserLaunchOptions = {}): Promise<LaunchedBrowser> {
    const endpoint = options.ws_url ?? options.cdp_url ?? this.options.ws_url ?? this.options.cdp_url;
    if (!endpoint) throw new Error("launch.mode=remote requires upstream.ws_url or cdp_url.");
    const ws_url = await resolveCdpWebSocketUrl(endpoint, "remote cdp_url");
    this.launched = { cdp_url: endpoint, ws_url, close: async () => {} };
    return this.launched;
  }
}
