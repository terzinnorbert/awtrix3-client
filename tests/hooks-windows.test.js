#!/usr/bin/env node

/**
 * Windows Hook Command Syntax Tests
 *
 * Validates Windows-specific command strings in both hook maps without
 * executing them.  Checks:
 * - Every hook entry has a Windows command variant (powershell / commandWindows)
 * - No bash-only syntax appears in Windows commands
 * - Environment variable references use PowerShell $env: prefix
 * - Path separators use backslash in Windows commands
 * - Script file paths referenced by Windows commands resolve to files that exist
 * - Timeout fields are present and are positive integers
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

const PROJECT_ROOT = path.join(__dirname, '..');
const COPILOT_HOOKS = path.join(PROJECT_ROOT, 'hooks', 'copilot-hooks.json');
const CODEX_HOOKS = path.join(PROJECT_ROOT, 'hooks', 'claude-codex-hooks.json');

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

// ── Helpers ──────────────────────────────────────────────────────────────────

/**
 * Patterns that are valid in bash but invalid (or dangerous) in PowerShell.
 */
const BASH_ONLY_PATTERNS = [
  { re: /\$\{[^}]+:-[^}]+\}/,  label: 'bash default-value substitution ${VAR:-default}' },
  { re: />\s*\/dev\/null/,       label: 'bash null redirect >/dev/null' },
  { re: /\bcommand\s+-v\b/,     label: 'bash command -v availability check' },
  { re: /\|\|\s*exit\s+0\b/,    label: 'bash ||exit 0 fallback' },
  { re: /&&/,                    label: 'bash && chaining (not PowerShell-safe bare &&)' },
];

/**
 * Extract all node script file paths referenced inside a Windows command string.
 * Looks for patterns like:  node "...\hooks\something.js"
 */
function extractNodeScriptPaths(cmd) {
  const matches = [];
  // Match: node "<path>" or node '<path>'
  const re = /\bnode\s+["']([^"']+\.js)["']/g;
  let m;
  while ((m = re.exec(cmd)) !== null) {
    matches.push(m[1]);
  }
  return matches;
}

/**
 * Convert a Windows-style path fragment containing a trailing hooks\<file>
 * into a workspace-relative path for existence checking.
 * Handles both Copilot-style ${PLUGIN_ROOT}\ and Codex-style $env:VAR\ prefixes.
 */
function resolveScriptPath(rawPath) {
  const stripped = rawPath
    .replace(/^\$\{[A-Z_]+\}\\?/i, '')   // ${PLUGIN_ROOT}\
    .replace(/^\$env:[A-Z_]+\\?/i, '');   // $env:PLUGIN_ROOT\
  // Normalise separators for the current OS
  return path.join(PROJECT_ROOT, stripped.replace(/\\/g, path.sep));
}

/**
 * Platform-agnostic basename that splits on both / and \ so Windows paths
 * parsed on Linux yield the correct filename component.
 */
function winBasename(p) {
  const parts = p.split(/[\/\\]/);
  return parts[parts.length - 1];
}

// ── Load and validate hook maps ───────────────────────────────────────────────

console.log('\n🧪 Windows Hook Command Syntax Tests\n');

let copilotHooks, codexHooks;

test('copilot-hooks.json can be loaded and parsed', () => {
  copilotHooks = JSON.parse(fs.readFileSync(COPILOT_HOOKS, 'utf8'));
});

test('claude-codex-hooks.json can be loaded and parsed', () => {
  codexHooks = JSON.parse(fs.readFileSync(CODEX_HOOKS, 'utf8'));
});

// ── Copilot hook map ─────────────────────────────────────────────────────────

console.log('\n--- Copilot hook map (powershell field) ---\n');

function collectCopilotEntries() {
  const entries = [];
  for (const [event, list] of Object.entries(copilotHooks.hooks || {})) {
    for (const entry of list) {
      entries.push({ event, entry });
    }
  }
  return entries;
}

test('every Copilot hook entry has a powershell field', () => {
  for (const { event, entry } of collectCopilotEntries()) {
    assert.ok(
      typeof entry.powershell === 'string' && entry.powershell.trim().length > 0,
      `Hook "${event}" is missing a non-empty powershell field`
    );
  }
});

test('Copilot powershell commands contain no bash-only syntax', () => {
  for (const { event, entry } of collectCopilotEntries()) {
    if (typeof entry.powershell !== 'string') continue;
    for (const { re, label } of BASH_ONLY_PATTERNS) {
      assert.ok(
        !re.test(entry.powershell),
        `Copilot hook "${event}" powershell field contains ${label}: ${entry.powershell}`
      );
    }
  }
});

test('Copilot powershell commands use backslash path separators for hook scripts', () => {
  for (const { event, entry } of collectCopilotEntries()) {
    if (typeof entry.powershell !== 'string') continue;
    // If the command references a .js file it must use backslash
    if (entry.powershell.includes('.js')) {
      assert.ok(
        entry.powershell.includes('\\'),
        `Copilot hook "${event}" powershell field references .js file but uses forward slash: ${entry.powershell}`
      );
    }
  }
});

test('Copilot powershell commands reference .js scripts that exist on disk', () => {
  for (const { event, entry } of collectCopilotEntries()) {
    if (typeof entry.powershell !== 'string') continue;
    const scriptPaths = extractNodeScriptPaths(entry.powershell);
    for (const raw of scriptPaths) {
      const resolved = resolveScriptPath(raw);
      assert.ok(
        fs.existsSync(resolved),
        `Copilot hook "${event}" powershell references missing file: ${raw} (resolved: ${resolved})`
      );
    }
  }
});

test('every Copilot hook entry has a positive-integer timeoutSec', () => {
  for (const { event, entry } of collectCopilotEntries()) {
    assert.ok(
      Number.isInteger(entry.timeoutSec) && entry.timeoutSec > 0,
      `Copilot hook "${event}" must have timeoutSec > 0, got: ${entry.timeoutSec}`
    );
  }
});

// ── Codex hook map ───────────────────────────────────────────────────────────

console.log('\n--- Codex hook map (commandWindows field) ---\n');

function collectCodexEntries() {
  const entries = [];
  for (const [event, list] of Object.entries(codexHooks.hooks || {})) {
    for (const group of list) {
      for (const hook of (group.hooks || [])) {
        entries.push({ event, hook });
      }
    }
  }
  return entries;
}

test('every Codex hook entry has a commandWindows field', () => {
  for (const { event, hook } of collectCodexEntries()) {
    assert.ok(
      typeof hook.commandWindows === 'string' && hook.commandWindows.trim().length > 0,
      `Codex hook "${event}" is missing a non-empty commandWindows field`
    );
  }
});

test('Codex commandWindows fields contain no bash-only syntax', () => {
  for (const { event, hook } of collectCodexEntries()) {
    if (typeof hook.commandWindows !== 'string') continue;
    for (const { re, label } of BASH_ONLY_PATTERNS) {
      assert.ok(
        !re.test(hook.commandWindows),
        `Codex hook "${event}" commandWindows contains ${label}: ${hook.commandWindows}`
      );
    }
  }
});

test('Codex commandWindows fields use $env: prefix for environment variables', () => {
  for (const { event, hook } of collectCodexEntries()) {
    if (typeof hook.commandWindows !== 'string') continue;
    // Any $UPPER_SNAKE reference that is NOT $env: is a bare bash-style variable
    const bareVarRe = /\$(?!env:)[A-Z_]{2,}/g;
    const matches = hook.commandWindows.match(bareVarRe);
    assert.ok(
      !matches,
      `Codex hook "${event}" commandWindows uses bare bash variable(s) ${JSON.stringify(matches)} instead of $env:: ${hook.commandWindows}`
    );
  }
});

test('Codex commandWindows uses backslash path separators for hook scripts', () => {
  for (const { event, hook } of collectCodexEntries()) {
    if (typeof hook.commandWindows !== 'string') continue;
    if (hook.commandWindows.includes('.js')) {
      assert.ok(
        hook.commandWindows.includes('\\'),
        `Codex hook "${event}" commandWindows references .js file but uses forward slash: ${hook.commandWindows}`
      );
    }
  }
});

test('Codex commandWindows references .js scripts that exist on disk', () => {
  for (const { event, hook } of collectCodexEntries()) {
    if (typeof hook.commandWindows !== 'string') continue;
    const scriptPaths = extractNodeScriptPaths(hook.commandWindows);
    for (const raw of scriptPaths) {
      const resolved = resolveScriptPath(raw);
      assert.ok(
        fs.existsSync(resolved),
        `Codex hook "${event}" commandWindows references missing file: ${raw} (resolved: ${resolved})`
      );
    }
  }
});

test('every Codex hook entry has a positive-integer timeout', () => {
  for (const { event, hook } of collectCodexEntries()) {
    assert.ok(
      Number.isInteger(hook.timeout) && hook.timeout > 0,
      `Codex hook "${event}" must have timeout > 0, got: ${hook.timeout}`
    );
  }
});

test('Codex commandWindows uses Get-Command for node availability check', () => {
  for (const { event, hook } of collectCodexEntries()) {
    if (typeof hook.commandWindows !== 'string') continue;
    assert.ok(
      hook.commandWindows.includes('Get-Command'),
      `Codex hook "${event}" commandWindows should use Get-Command to check for node, got: ${hook.commandWindows}`
    );
  }
});

// ── Cross-map consistency ────────────────────────────────────────────────────

console.log('\n--- Cross-map consistency ---\n');

test('both hook maps reference the same set of .js hook scripts', () => {
  const copilotScripts = new Set();
  for (const { entry } of collectCopilotEntries()) {
    if (typeof entry.powershell === 'string') {
      extractNodeScriptPaths(entry.powershell).forEach(p => copilotScripts.add(winBasename(p)));
    }
  }
  const codexScripts = new Set();
  for (const { hook } of collectCodexEntries()) {
    if (typeof hook.commandWindows === 'string') {
      extractNodeScriptPaths(hook.commandWindows).forEach(p => codexScripts.add(winBasename(p)));
    }
  }
  for (const s of copilotScripts) {
    assert.ok(codexScripts.has(s), `Script "${s}" is in Copilot hooks but missing from Codex hooks`);
  }
  for (const s of codexScripts) {
    assert.ok(copilotScripts.has(s), `Script "${s}" is in Codex hooks but missing from Copilot hooks`);
  }
});

// ── Summary ──────────────────────────────────────────────────────────────────

console.log('\n' + '='.repeat(60));
if (TESTS_FAILED.length > 0) {
  console.log(`❌ ${TESTS_FAILED.length} test(s) failed, ${TESTS_PASSED.length} passed`);
  console.log('='.repeat(60) + '\n');
  process.exit(1);
} else {
  console.log(`✅ All ${TESTS_PASSED.length} Windows hook syntax tests passed`);
  console.log('='.repeat(60) + '\n');
  process.exit(0);
}
