// Extension service worker entry point. Importing ModCDPServer installs it on
// globalThis and starts best-effort keepalive setup.

import "./ModCDPServer.js";

const DEFAULT_REVERSE_PROXY_URL = "ws://127.0.0.1:29292";

void (
  globalThis as typeof globalThis & {
    ModCDP?: {
      startReverseBridge?: (endpoint: string, options?: { reconnect_interval_ms?: number }) => unknown;
    };
  }
).ModCDP?.startReverseBridge?.(DEFAULT_REVERSE_PROXY_URL, { reconnect_interval_ms: 2_000 });
