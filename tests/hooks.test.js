#!/usr/bin/env node

/**
 * Hook Compatibility Tests
 *
 * Validates:
 * - Copilot-host output shape (SessionStart and UserPromptSubmit)
 * - Codex-host output shape (SessionStart and UserPromptSubmit)
 * - Non-blocking behavior when AWTRIX_HOST is missing
 * - State isolation between sequential hook invocations
 */

const { spawnSync } = require('child_process');
const assert = require('assert');
const path = require('path');

const PROJECT_ROOT = path.join(__dirname, '..');
const HOOK_ACTIVATE = path.join(PROJECT_ROOT, 'hooks', 'awtrix-activate.js');
const HOOK_TRACKER = path.join(PROJECT_ROOT, 'hooks', 'awtrix-mode-tracker.js');

const TESTS_PASSED = [];
const TESTS_FAILED = [];

function test(description, fn) {
  try {
    fn();
    TESTS_PASSED.push(description);
    console.log(`✓ ${description}`);
  } catch (err) {
    TESTS_FAILED.push({ description, error: err.message });
    console.error(`✗ ${description}`);
    console.error(`  ${err.message}`);
  }
}

/**
 * Spawn a hook script and return its result.
 * @param {string} scriptPath - Absolute path to the hook script.
 * @param {object} extraEnv - Additional environment variables to set.
 * @param {string|null} stdin - Optional stdin input for the process.
 */
function runHook(scriptPath, extraEnv = {}, stdin = null) {
  // Strip any real AWTRIX_HOST so notify calls don't actually fire unless
  // a test explicitly needs them to.  Tests that want a live host pass it
  // via extraEnv.
  const env = Object.assign({}, process.env, {
    AWTRIX_HOST: undefined,
    COPILOT_PLUGIN_DATA: undefined,
    PLUGIN_DATA: undefined,
  }, extraEnv);

  // Remove keys set to undefined (Object.assign keeps them as own keys)
  for (const key of Object.keys(env)) {
    if (env[key] === undefined) delete env[key];
  }

  const opts = {
    cwd: PROJECT_ROOT,
    env,
    encoding: 'utf8',
    timeout: 10000,
  };

  if (stdin !== null) {
    opts.input = stdin;
  }

  return spawnSync('node', [scriptPath], opts);
}

function parseJSON(raw, label) {
  try {
    return JSON.parse(raw);
  } catch (err) {
    throw new Error(`${label}: output is not valid JSON: ${err.message}\n  Raw: ${JSON.stringify(raw)}`);
  }
}

// ============================================================
// Copilot host — awtrix-activate.js (SessionStart)
// ============================================================

console.log('\n🧪 Hook Compatibility Tests\n');

console.log('--- Copilot host output shape ---\n');

test('activate: Copilot host exits with code 0', () => {
  const result = runHook(HOOK_ACTIVATE, { COPILOT_PLUGIN_DATA: '1' });
  assert.strictEqual(result.status, 0, `Expected exit 0, got ${result.status}\n  stderr: ${result.stderr}`);
});

test('activate: Copilot host writes valid JSON to stdout', () => {
  const result = runHook(HOOK_ACTIVATE, { COPILOT_PLUGIN_DATA: '1' });
  parseJSON(result.stdout, 'Copilot SessionStart');
});

test('activate: Copilot host output contains additionalContext field', () => {
  const result = runHook(HOOK_ACTIVATE, { COPILOT_PLUGIN_DATA: '1' });
  const output = parseJSON(result.stdout, 'Copilot SessionStart');
  assert.ok(
    typeof output.additionalContext === 'string',
    `Expected additionalContext string, got ${JSON.stringify(output)}`
  );
});

test('activate: Copilot host does not emit systemMessage field', () => {
  const result = runHook(HOOK_ACTIVATE, { COPILOT_PLUGIN_DATA: '1' });
  const output = parseJSON(result.stdout, 'Copilot SessionStart');
  assert.ok(
    !('systemMessage' in output),
    `Copilot output must not contain systemMessage: ${JSON.stringify(output)}`
  );
});

// ============================================================
// Copilot host — awtrix-mode-tracker.js (UserPromptSubmit)
// ============================================================

test('tracker: Copilot host exits with code 0 on non-awtrix prompt', () => {
  const stdin = JSON.stringify({ prompt: 'What is the capital of France?' });
  const result = runHook(HOOK_TRACKER, { COPILOT_PLUGIN_DATA: '1' }, stdin);
  assert.strictEqual(result.status, 0, `Expected exit 0, got ${result.status}\n  stderr: ${result.stderr}`);
});

test('tracker: Copilot host writes valid JSON to stdout on non-awtrix prompt', () => {
  const stdin = JSON.stringify({ prompt: 'Hello world' });
  const result = runHook(HOOK_TRACKER, { COPILOT_PLUGIN_DATA: '1' }, stdin);
  parseJSON(result.stdout, 'Copilot UserPromptSubmit non-awtrix');
});

test('tracker: Copilot host output is empty object when no awtrix command in prompt', () => {
  const stdin = JSON.stringify({ prompt: 'Hello world' });
  const result = runHook(HOOK_TRACKER, { COPILOT_PLUGIN_DATA: '1' }, stdin);
  const output = parseJSON(result.stdout, 'Copilot UserPromptSubmit non-awtrix');
  assert.deepStrictEqual(output, {}, `Expected {}, got ${JSON.stringify(output)}`);
});

// ============================================================
// Codex host — awtrix-activate.js (SessionStart)
// ============================================================

console.log('\n--- Codex host output shape ---\n');

test('activate: Codex host exits with code 0', () => {
  const result = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  assert.strictEqual(result.status, 0, `Expected exit 0, got ${result.status}\n  stderr: ${result.stderr}`);
});

test('activate: Codex host writes valid JSON to stdout', () => {
  const result = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  parseJSON(result.stdout, 'Codex SessionStart');
});

test('activate: Codex host output contains systemMessage field', () => {
  const result = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  const output = parseJSON(result.stdout, 'Codex SessionStart');
  assert.ok(
    typeof output.systemMessage === 'string',
    `Expected systemMessage string, got ${JSON.stringify(output)}`
  );
});

test('activate: Codex host systemMessage follows AWTRIX:<MODE> format', () => {
  const result = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  const output = parseJSON(result.stdout, 'Codex SessionStart');
  assert.match(
    output.systemMessage,
    /^AWTRIX:[A-Z]+$/,
    `Expected AWTRIX:<MODE> format, got ${output.systemMessage}`
  );
});

test('activate: Codex host output contains hookSpecificOutput field', () => {
  const result = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  const output = parseJSON(result.stdout, 'Codex SessionStart');
  assert.ok(
    output.hookSpecificOutput && typeof output.hookSpecificOutput === 'object',
    `Expected hookSpecificOutput object, got ${JSON.stringify(output)}`
  );
});

test('activate: Codex hookSpecificOutput.hookEventName is SessionStart', () => {
  const result = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  const output = parseJSON(result.stdout, 'Codex SessionStart');
  assert.strictEqual(
    output.hookSpecificOutput.hookEventName,
    'SessionStart',
    `Expected hookEventName "SessionStart", got ${output.hookSpecificOutput.hookEventName}`
  );
});

// ============================================================
// Codex host — awtrix-mode-tracker.js (UserPromptSubmit)
// ============================================================

test('tracker: Codex host exits with code 0 on non-awtrix prompt', () => {
  const stdin = JSON.stringify({ prompt: 'Explain recursion' });
  const result = runHook(HOOK_TRACKER, { PLUGIN_DATA: '1' }, stdin);
  assert.strictEqual(result.status, 0, `Expected exit 0, got ${result.status}\n  stderr: ${result.stderr}`);
});

test('tracker: Codex host writes valid JSON to stdout', () => {
  const stdin = JSON.stringify({ prompt: 'Explain recursion' });
  const result = runHook(HOOK_TRACKER, { PLUGIN_DATA: '1' }, stdin);
  parseJSON(result.stdout, 'Codex UserPromptSubmit');
});

test('tracker: Codex host output contains systemMessage field', () => {
  const stdin = JSON.stringify({ prompt: 'Explain recursion' });
  const result = runHook(HOOK_TRACKER, { PLUGIN_DATA: '1' }, stdin);
  const output = parseJSON(result.stdout, 'Codex UserPromptSubmit');
  assert.ok(
    typeof output.systemMessage === 'string',
    `Expected systemMessage string, got ${JSON.stringify(output)}`
  );
});

// ============================================================
// Non-blocking behavior — missing AWTRIX_HOST
// ============================================================

console.log('\n--- Non-blocking behavior (no AWTRIX_HOST) ---\n');

test('activate: exits 0 and produces JSON even when AWTRIX_HOST is absent (Copilot)', () => {
  // AWTRIX_HOST is intentionally absent (stripped by runHook default)
  const result = runHook(HOOK_ACTIVATE, { COPILOT_PLUGIN_DATA: '1' });
  assert.strictEqual(result.status, 0, `Expected exit 0, got ${result.status}`);
  parseJSON(result.stdout, 'Copilot SessionStart no-host');
});

test('activate: exits 0 and produces JSON even when AWTRIX_HOST is absent (Codex)', () => {
  const result = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  assert.strictEqual(result.status, 0, `Expected exit 0, got ${result.status}`);
  parseJSON(result.stdout, 'Codex SessionStart no-host');
});

test('tracker: exits 0 and produces JSON even when AWTRIX_HOST is absent (Copilot)', () => {
  const stdin = JSON.stringify({ prompt: 'Hello' });
  const result = runHook(HOOK_TRACKER, { COPILOT_PLUGIN_DATA: '1' }, stdin);
  assert.strictEqual(result.status, 0, `Expected exit 0, got ${result.status}`);
  parseJSON(result.stdout, 'Copilot UserPromptSubmit no-host');
});

test('tracker: exits 0 and produces JSON even when AWTRIX_HOST is absent (Codex)', () => {
  const stdin = JSON.stringify({ prompt: 'Hello' });
  const result = runHook(HOOK_TRACKER, { PLUGIN_DATA: '1' }, stdin);
  assert.strictEqual(result.status, 0, `Expected exit 0, got ${result.status}`);
  parseJSON(result.stdout, 'Codex UserPromptSubmit no-host');
});

// ============================================================
// State isolation — sequential invocations do not share state
// ============================================================

console.log('\n--- State isolation ---\n');

test('isolation: two Codex SessionStart invocations produce independent outputs', () => {
  const r1 = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  const r2 = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  const o1 = parseJSON(r1.stdout, 'Codex run-1');
  const o2 = parseJSON(r2.stdout, 'Codex run-2');
  // Both must have systemMessage — independent, not accumulated
  assert.ok(typeof o1.systemMessage === 'string', 'run-1 missing systemMessage');
  assert.ok(typeof o2.systemMessage === 'string', 'run-2 missing systemMessage');
});

test('isolation: Copilot env does not bleed into Codex invocation', () => {
  const rCopilot = runHook(HOOK_ACTIVATE, { COPILOT_PLUGIN_DATA: '1' });
  const rCodex = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  const oCopilot = parseJSON(rCopilot.stdout, 'Copilot run');
  const oCodex = parseJSON(rCodex.stdout, 'Codex run');
  assert.ok(!('systemMessage' in oCopilot), 'Copilot output must not have systemMessage');
  assert.ok('systemMessage' in oCodex, 'Codex output must have systemMessage');
});

test('isolation: Codex env does not bleed into Copilot invocation', () => {
  const rCodex = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  const rCopilot = runHook(HOOK_ACTIVATE, { COPILOT_PLUGIN_DATA: '1' });
  const oCodex = parseJSON(rCodex.stdout, 'Codex run');
  const oCopilot = parseJSON(rCopilot.stdout, 'Copilot run');
  assert.ok('systemMessage' in oCodex, 'Codex output must have systemMessage');
  assert.ok(!('systemMessage' in oCopilot), 'Copilot output must not have systemMessage');
});

test('isolation: no-host invocation does not pollute subsequent Codex invocation', () => {
  // First run with no known host, second with Codex
  runHook(HOOK_ACTIVATE, {});
  const r2 = runHook(HOOK_ACTIVATE, { PLUGIN_DATA: '1' });
  const o2 = parseJSON(r2.stdout, 'Codex after no-host');
  assert.ok(typeof o2.systemMessage === 'string', 'Codex run after no-host run must have systemMessage');
});

// ============================================================
// Summary
// ============================================================

console.log('\n' + '='.repeat(60));
if (TESTS_FAILED.length > 0) {
  console.log(`❌ ${TESTS_FAILED.length} test(s) failed, ${TESTS_PASSED.length} passed`);
  console.log('='.repeat(60) + '\n');
  process.exit(1);
} else {
  console.log(`✅ All ${TESTS_PASSED.length} hook compatibility tests passed`);
  console.log('='.repeat(60) + '\n');
  process.exit(0);
}
