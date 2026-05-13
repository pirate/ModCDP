#!/usr/bin/env node

import { execFileSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";

const root = process.cwd();
const dist_extension = path.join(root, "dist", "extension");

if (fs.existsSync(dist_extension)) {
  for (const entry of fs.readdirSync(dist_extension, { withFileTypes: true })) {
    if (entry.name !== "src") fs.rmSync(path.join(dist_extension, entry.name), { recursive: true, force: true });
  }
}

const copy = (from, to) => {
  fs.mkdirSync(path.dirname(path.join(root, to)), { recursive: true });
  fs.copyFileSync(path.join(root, from), path.join(root, to));
};

const write = (to, contents) => {
  fs.mkdirSync(path.dirname(path.join(root, to)), { recursive: true });
  fs.writeFileSync(path.join(root, to), contents);
};

copy("extension/manifest.json", "dist/extension/manifest.json");
copy("extension/src/pages/options.html", "dist/extension/options.html");
copy("dist/extension/src/pages/options.js", "dist/extension/options.js");
copy("extension/src/pages/wake.html", "dist/extension/modcdp/wake.html");
copy("dist/extension/src/pages/wake.js", "dist/extension/modcdp/wake.js");
copy("extension/src/pages/offscreen_keepalive.html", "dist/extension/offscreen/keepalive.html");
copy("dist/extension/src/pages/offscreen_keepalive.js", "dist/extension/offscreen/offscreen_keepalive.js");
copy("dist/js/src/server/ModCDPServer.js", "dist/extension/js/src/server/ModCDPServer.js");
if (fs.existsSync(path.join(root, "dist/js/src/server/ModCDPServer.js.map"))) {
  copy("dist/js/src/server/ModCDPServer.js.map", "dist/extension/js/src/server/ModCDPServer.js.map");
}

const service_worker = fs
  .readFileSync(path.join(root, "dist/extension/src/service_worker.js"), "utf8")
  .replace("../../js/src/server/ModCDPServer.js", "../js/src/server/ModCDPServer.js");
write("dist/extension/modcdp/service_worker.js", service_worker);
copy("dist/extension/src/service_worker.js.map", "dist/extension/modcdp/service_worker.js.map");
fs.rmSync(path.join(dist_extension, "src"), { recursive: true, force: true });

const fixed_date = new Date("2000-01-01T00:00:00Z");
const touchTree = (dir) => {
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full_path = path.join(dir, entry.name);
    if (entry.isDirectory()) touchTree(full_path);
    fs.utimesSync(full_path, fixed_date, fixed_date);
  }
  fs.utimesSync(dir, fixed_date, fixed_date);
};
touchTree(dist_extension);

fs.rmSync(path.join(root, "dist", "extension.zip"), { force: true });
execFileSync("zip", ["-X", "-qr", "extension.zip", "extension"], {
  cwd: path.join(root, "dist"),
  stdio: "inherit",
});

copy("dist/extension.zip", "python/modcdp/extension.zip");
copy("dist/extension.zip", "go/modcdp/injector/extension.zip");
