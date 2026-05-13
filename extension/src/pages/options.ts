const out = document.getElementById("out");
const load = () =>
  chrome.runtime.sendMessage({ type: "modcdp.options.status" }, (status) => {
    out.textContent = JSON.stringify(status || chrome.runtime.lastError, null, 2);
  });
document.getElementById("refresh").onclick = load;
load();
setInterval(load, 2000);
