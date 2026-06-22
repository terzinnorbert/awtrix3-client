#!/usr/bin/env node

const { readPromptFromStdin, runNotify, writeHookOutput } = require("./awtrix-runtime");

function mapPromptToEvent(prompt) {
  const text = String(prompt || "").trim();
  const lowered = text.toLowerCase();

  if (!/^[/@$]awtrix-notify\b/.test(lowered) && !/^[/@$]awtrix-notify:awtrix-notify\b/.test(lowered)) {
    return null;
  }

  const parts = text.split(/\s+/);
  const eventType = parts[1] || "attention";
  const message = parts.slice(2).join(" ").trim() || `Event: ${eventType}`;
  return { eventType, message };
}

readPromptFromStdin((prompt) => {
  const mapped = mapPromptToEvent(prompt);
  if (!mapped) {
    try {
      writeHookOutput("UserPromptSubmit", "idle", "");
    } catch (_) {
      // Keep hook non-blocking.
    }
    return;
  }

  const result = runNotify(mapped.eventType, mapped.message);
  const context = result.ok
    ? `AWTRIX notification sent (${mapped.eventType})`
    : `AWTRIX notification failed (${mapped.eventType})`;

  try {
    writeHookOutput("UserPromptSubmit", mapped.eventType, context);
  } catch (_) {
    // Keep hook non-blocking.
  }
});
