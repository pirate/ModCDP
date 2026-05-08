const MODCDP_OFFSCREEN_KEEPALIVE_PORT = "ModCDPOffscreenKeepAlive";
const MODCDP_OFFSCREEN_KEEPALIVE_INTERVAL_MS = 1_000;
const MODCDP_OFFSCREEN_RECONNECT_DELAY_MS = 250;

let port: chrome.runtime.Port | null = null;

function sendKeepAlive() {
  port?.postMessage({ type: "ModCDPOffscreenKeepAlive", at: Date.now() });
}

function connectKeepAlivePort() {
  port = chrome.runtime.connect({ name: MODCDP_OFFSCREEN_KEEPALIVE_PORT });
  port.onDisconnect.addListener(() => {
    port = null;
    setTimeout(connectKeepAlivePort, MODCDP_OFFSCREEN_RECONNECT_DELAY_MS);
  });
  sendKeepAlive();
}

connectKeepAlivePort();
setInterval(sendKeepAlive, MODCDP_OFFSCREEN_KEEPALIVE_INTERVAL_MS);
