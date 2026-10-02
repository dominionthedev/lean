#!/usr/bin/env node

const fs = require("node:fs");
const path = require("node:path");
const https = require("node:https");
const crypto = require("node:crypto");
const { spawnSync } = require("node:child_process");

const pkg = require("../package.json");
const root = path.join(__dirname, "..");
const binDir = path.join(root, "bin");
const isWindows = process.platform === "win32";

const platform = {
  linux: "linux",
  darwin: "darwin",
  win32: "windows",
}[process.platform];

const arch = {
  x64: "amd64",
  arm64: "arm64",
}[process.arch];

if (!platform || !arch) {
  console.error("lean: unsupported platform or architecture:", process.platform, process.arch);
  process.exit(1);
}

const version = pkg.version;
const tag = "v" + version;
const ext = isWindows ? "zip" : "tar.gz";
const archive = "lean_" + version + "_" + platform + "_" + arch + "." + ext;
const baseUrl = "https://github.com/dominionthedev/lean/releases/download/" + tag + "/";
const url = baseUrl + archive;
const checksumsUrl = baseUrl + "checksums.txt";
const destination = path.join(root, archive);

fs.mkdirSync(binDir, { recursive: true });

function request(location) {
  return new Promise((resolve, reject) => {
    https.get(location, { headers: { "User-Agent": "lean-npm-installer" } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        res.resume();
        request(new URL(res.headers.location, location).toString()).then(resolve, reject);
        return;
      }

      if (res.statusCode !== 200) {
        res.resume();
        reject(new Error("GitHub returned HTTP " + res.statusCode));
        return;
      }

      const chunks = [];
      res.on("data", (chunk) => chunks.push(chunk));
      res.on("end", () => resolve(Buffer.concat(chunks)));
      res.on("error", reject);
    }).on("error", reject);
  });
}

function download(location) {
  return request(location).then((data) => fs.writeFileSync(destination, data));
}

function run(command, args) {
  const result = spawnSync(command, args, { cwd: root, stdio: "inherit" });
  if (result.error || result.status !== 0) {
    throw result.error || new Error(command + " exited with status " + result.status);
  }
}

function verifyChecksum(checksums) {
  const line = checksums
    .toString("utf8")
    .split(/\r?\n/)
    .find((entry) => entry.trim().endsWith("  " + archive) || entry.trim().endsWith(" *" + archive));

  if (!line) {
    throw new Error("checksum for " + archive + " was not found");
  }

  const expected = line.trim().split(/\s+/)[0].toLowerCase();
  const actual = crypto.createHash("sha256").update(fs.readFileSync(destination)).digest("hex");

  if (expected !== actual) {
    throw new Error("checksum mismatch for " + archive);
  }
}

async function main() {
  console.log("lean: downloading " + tag + " for " + platform + "/" + arch + "...");
  await download(url);
  const checksums = await request(checksumsUrl);
  verifyChecksum(checksums);

  try {
    if (isWindows) {
      run("tar", ["-xf", destination, "-C", binDir]);
    } else {
      run("tar", ["-xzf", destination, "-C", binDir]);
      fs.chmodSync(path.join(binDir, "lean"), 0o755);
    }
  } finally {
    fs.rmSync(destination, { force: true });
  }

  console.log("lean: installed successfully.");
}

main().catch((err) => {
  console.error("lean: installation failed:", err.message);
  console.error("You can install a release manually from https://github.com/dominionthedev/lean/releases");
  process.exit(1);
});
