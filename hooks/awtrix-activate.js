#!/usr/bin/env node

const { runNotify, writeHookOutput } = require("./awtrix-runtime");

const result = runNotify("start", "Session started");

let context = "AWTRIX plugin active";
if (!result.ok) {
  context += `; startup notification skipped (exit ${result.status ?? "unknown"})`;
}

try {
  writeHookOutput("SessionStart", "start", context);
} catch (_) {
  // Hook output failures should never block the session.
}
