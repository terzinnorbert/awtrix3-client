#!/usr/bin/env node

/**
 * Copilot Plugin Manifest Smoke Tests
 * 
 * Validates:
 * - Plugin manifest JSON validity
 * - Referenced files and directories exist
 * - Hook map references are correct
 * - Skills and commands paths are accessible
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

const PROJECT_ROOT = path.join(__dirname, '..');
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

function fileExists(filePath, description) {
  const fullPath = path.join(PROJECT_ROOT, filePath);
  assert.ok(fs.existsSync(fullPath), `${description}: ${filePath} does not exist`);
}

function directoryExists(dirPath, description) {
  const fullPath = path.join(PROJECT_ROOT, dirPath);
  assert.ok(
    fs.existsSync(fullPath) && fs.statSync(fullPath).isDirectory(),
    `${description}: ${dirPath} is not a directory`
  );
}

function isValidJSON(filePath, description) {
  const fullPath = path.join(PROJECT_ROOT, filePath);
  const content = fs.readFileSync(fullPath, 'utf8');
  try {
    JSON.parse(content);
  } catch (err) {
    throw new Error(`${description}: ${filePath} contains invalid JSON: ${err.message}`);
  }
}

// ============================================================
// Copilot Plugin Tests
// ============================================================

console.log('\n🧪 Copilot Plugin Manifest Tests\n');

test('Copilot manifest exists', () => {
  fileExists('.github/plugin/plugin.json', 'Copilot manifest');
});

test('Copilot manifest is valid JSON', () => {
  isValidJSON('.github/plugin/plugin.json', 'Copilot manifest');
});

test('Copilot manifest references skills directory', () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.github/plugin/plugin.json'), 'utf8')
  );
  assert.ok(manifest.skills, 'Copilot manifest missing "skills" field');
  directoryExists(manifest.skills, 'Copilot manifest skills ref');
});

test('Copilot manifest references commands directory', () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.github/plugin/plugin.json'), 'utf8')
  );
  assert.ok(manifest.commands, 'Copilot manifest missing "commands" field');
  directoryExists(manifest.commands, 'Copilot manifest commands ref');
});

test('Copilot manifest references hook map', () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.github/plugin/plugin.json'), 'utf8')
  );
  assert.ok(manifest.hooks, 'Copilot manifest missing "hooks" field');
  fileExists(manifest.hooks, 'Copilot manifest hooks ref');
});

test('Copilot manifest has required metadata', () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.github/plugin/plugin.json'), 'utf8')
  );
  assert.ok(manifest.name, 'Missing "name" field');
  assert.ok(manifest.version, 'Missing "version" field');
  assert.ok(manifest.description, 'Missing "description" field');
  assert.ok(manifest.author, 'Missing "author" field');
  assert.ok(manifest.license, 'Missing "license" field');
});

test('Canonical plugin identity is reused across manifests/docs/tests', () => {
  const copilotManifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.github/plugin/plugin.json'), 'utf8')
  );
  const codexManifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.codex-plugin/plugin.json'), 'utf8')
  );
  const commandDef = fs.readFileSync(
    path.join(PROJECT_ROOT, 'commands/awtrix-notify.toml'),
    'utf8'
  );
  const readme = fs.readFileSync(path.join(PROJECT_ROOT, 'README.md'), 'utf8');

  const canonicalName = 'awtrix-notify';
  const canonicalDisplayName = 'AWTRIX Notify';
  const canonicalPrefix = '/awtrix-notify';

  assert.strictEqual(copilotManifest.name, canonicalName, 'Copilot manifest.name mismatch');
  assert.strictEqual(codexManifest.name, canonicalName, 'Codex manifest.name mismatch');
  assert.strictEqual(
    codexManifest.interface.displayName,
    canonicalDisplayName,
    'Codex interface.displayName mismatch'
  );
  assert.ok(commandDef.includes(canonicalPrefix), 'Command definition missing canonical prefix');
  assert.ok(readme.includes(canonicalPrefix), 'README missing canonical prefix');
  assert.ok(readme.includes(canonicalDisplayName), 'README missing canonical display name');
});

test('Copilot hook map exists', () => {
  fileExists('hooks/copilot-hooks.json', 'Copilot hook map');
});

test('Copilot hook map is valid JSON', () => {
  isValidJSON('hooks/copilot-hooks.json', 'Copilot hook map');
});

test('Copilot hook map has sessionStart event', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/copilot-hooks.json'), 'utf8')
  );
  assert.ok(hookMap.hooks.sessionStart, 'Missing "sessionStart" hook');
  assert.ok(Array.isArray(hookMap.hooks.sessionStart), '"sessionStart" must be an array');
});

test('Copilot hook map has userPromptSubmitted event', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/copilot-hooks.json'), 'utf8')
  );
  assert.ok(hookMap.hooks.userPromptSubmitted, 'Missing "userPromptSubmitted" hook');
  assert.ok(Array.isArray(hookMap.hooks.userPromptSubmitted), '"userPromptSubmitted" must be an array');
});

test('Copilot sessionStart hook references valid script', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/copilot-hooks.json'), 'utf8')
  );
  const hook = hookMap.hooks.sessionStart[0];
  assert.ok(hook.bash, 'Missing "bash" command in sessionStart hook');
  // Extract script name from bash command
  const bashMatch = hook.bash.match(/hooks\/([a-z-]+\.js)/);
  assert.ok(bashMatch, 'Could not parse script name from bash command');
  fileExists(`hooks/${bashMatch[1]}`, 'Copilot sessionStart script');
});

test('Copilot userPromptSubmitted hook references valid script', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/copilot-hooks.json'), 'utf8')
  );
  const hook = hookMap.hooks.userPromptSubmitted[0];
  assert.ok(hook.bash, 'Missing "bash" command in userPromptSubmitted hook');
  // Extract script name from bash command
  const bashMatch = hook.bash.match(/hooks\/([a-z-]+\.js)/);
  assert.ok(bashMatch, 'Could not parse script name from bash command');
  fileExists(`hooks/${bashMatch[1]}`, 'Copilot userPromptSubmitted script');
});

test('Copilot hook commands have timeout defined', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/copilot-hooks.json'), 'utf8')
  );
  const sessionStartHook = hookMap.hooks.sessionStart[0];
  const promptHook = hookMap.hooks.userPromptSubmitted[0];
  assert.ok(sessionStartHook.timeoutSec !== undefined, 'sessionStart hook missing timeoutSec');
  assert.ok(promptHook.timeoutSec !== undefined, 'userPromptSubmitted hook missing timeoutSec');
  assert.ok(sessionStartHook.timeoutSec > 0, 'sessionStart timeout must be positive');
  assert.ok(promptHook.timeoutSec > 0, 'userPromptSubmitted timeout must be positive');
});

test('awtrix-notify skill exists', () => {
  fileExists('skills/awtrix-notify/SKILL.md', 'awtrix-notify skill');
});

test('awtrix-notify command exists', () => {
  fileExists('commands/awtrix-notify.toml', 'awtrix-notify command');
});

test('Hook scripts directory exists', () => {
  directoryExists('hooks', 'Hooks directory');
});

test('Required hook scripts exist', () => {
  fileExists('hooks/awtrix-activate.js', 'awtrix-activate.js');
  fileExists('hooks/awtrix-mode-tracker.js', 'awtrix-mode-tracker.js');
  fileExists('hooks/awtrix-runtime.js', 'awtrix-runtime.js');
});

// ============================================================
// Summary
// ============================================================

console.log('\n' + '='.repeat(60));
console.log(`Tests passed: ${TESTS_PASSED.length}`);
console.log(`Tests failed: ${TESTS_FAILED.length}`);
console.log('='.repeat(60) + '\n');

if (TESTS_FAILED.length > 0) {
  console.log('Failed tests:\n');
  TESTS_FAILED.forEach(({ description, error }) => {
    console.log(`  • ${description}`);
    console.log(`    → ${error}\n`);
  });
  process.exit(1);
} else {
  console.log('✅ All Copilot plugin manifest tests passed!\n');
  process.exit(0);
}
