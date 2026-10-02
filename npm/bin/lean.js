#!/usr/bin/env node

const { spawn } = require("node:child_process");
const path = require("node:path");

const binary = path.join(
  __dirname,
  "..",
  "bin",
  process.platform === "win32" ? "lean.exe" : "lean",
);

const child = spawn(binary, process.argv.slice(2), {
  stdio: "inherit",
  windowsHide: false,
});

child.on("error", (err) => {
  console.error("lean: unable to start the installed binary:", err.message);
  process.exit(1);
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code ?? 1);
  }
});
