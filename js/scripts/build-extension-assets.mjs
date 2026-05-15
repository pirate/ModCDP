#!/usr/bin/env node

import { execFileSync } from "node:child_process";
import fs from "node:fs";
import path from "node:path";
import { build } from "esbuild";

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

copy("extension/manifest.json", "dist/extension/manifest.json");
copy("extension/src/pages/options.html", "dist/extension/options.html");
copy("dist/extension/src/pages/options.js", "dist/extension/options.js");
copy("extension/src/pages/offscreen_keepalive.html", "dist/extension/offscreen/keepalive.html");
copy("dist/extension/src/pages/offscreen_keepalive.js", "dist/extension/offscreen/offscreen_keepalive.js");
await build({
  entryPoints: [path.join(root, "extension/src/service_worker.ts")],
  outfile: path.join(root, "dist/extension/modcdp/service_worker.js"),
  bundle: true,
  format: "esm",
  platform: "browser",
  target: ["chrome116"],
  sourcemap: true,
  logLevel: "silent",
});
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
