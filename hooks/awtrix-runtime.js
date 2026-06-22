const path = require("path");
const { spawnSync } = require("child_process");

const isCopilot = Boolean(process.env.COPILOT_PLUGIN_DATA);
const isCodex = !isCopilot && Boolean(process.env.PLUGIN_DATA);

const EVENT_PROFILES = {
  start: { normalizedEvent: "start", color: "#FFAA00", hold: false, wakeup: true },
  success: { normalizedEvent: "success", color: "#00FF00", hold: false, wakeup: true },
  error: { normalizedEvent: "error", color: "#FF0000", hold: true, wakeup: true },
  fail: { normalizedEvent: "error", color: "#FF0000", hold: true, wakeup: true },
  failure: { normalizedEvent: "error", color: "#FF0000", hold: true, wakeup: true },
  attention: { normalizedEvent: "attention", color: "#FF8800", hold: true, wakeup: true },
  input: { normalizedEvent: "attention", color: "#FF8800", hold: true, wakeup: true },
  build: { normalizedEvent: "build", color: "#00AAFF", hold: false, wakeup: true },
  test: { normalizedEvent: "test", color: "#AA44FF", hold: false, wakeup: true },
};

function repoRoot() {
  return path.resolve(__dirname, "..");
}

function resolveEventProfile(eventType) {
  const key = String(eventType || "attention").trim().toLowerCase();
  return EVENT_PROFILES[key] || EVENT_PROFILES.attention;
}

function runNotify(eventType, message) {
  const profile = resolveEventProfile(eventType);
  const args = [
    "run",
    ".",
    "notify",
    "--text",
    String(message || "").trim() || `Event: ${profile.normalizedEvent}`,
    "--color",
    profile.color,
    "--wakeup",
  ];
  if (profile.hold) {
    args.push("--hold");
  }

  const result = spawnSync("go", args, {
    cwd: repoRoot(),
    env: process.env,
    encoding: "utf8",
    timeout: 15000,
  });

  return {
    ok: result.status === 0,
    status: result.status,
    stdout: (result.stdout || "").trim(),
    stderr: (result.stderr || "").trim(),
  };
}

function writeHookOutput(eventName, mode, additionalContext) {
  if (isCopilot) {
    const payload = eventName === "SessionStart" && additionalContext
      ? { additionalContext }
      : {};
    process.stdout.write(JSON.stringify(payload));
    return;
  }

  if (isCodex) {
    const payload = {
      systemMessage: `AWTRIX:${String(mode || "active").toUpperCase()}`,
    };
    if (additionalContext) {
      payload.hookSpecificOutput = {
        hookEventName: eventName,
        additionalContext,
      };
    }
    process.stdout.write(JSON.stringify(payload));
    return;
  }

  if (additionalContext) {
    process.stdout.write(additionalContext);
  }
}

function readPromptFromStdin(callback) {
  let input = "";
  process.stdin.on("data", (chunk) => {
    input += chunk;
  });
  process.stdin.on("end", () => {
    try {
      const parsed = JSON.parse(input.replace(/^\uFEFF/, ""));
      callback(parsed.prompt || "");
    } catch (_) {
      callback("");
    }
  });
}

module.exports = {
  isCopilot,
  isCodex,
  resolveEventProfile,
  runNotify,
  writeHookOutput,
  readPromptFromStdin,
};
