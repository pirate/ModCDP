chrome.runtime.sendMessage({ type: "modcdp.wake", at: Date.now() }, () => {
  setTimeout(() => window.close(), 250);
});
