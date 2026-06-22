#!/usr/bin/env node

/**
 * Codex Plugin Manifest Smoke Tests
 * 
 * Validates:
 * - Plugin manifest JSON validity
 * - Referenced files and directories exist
 * - Hook map references are correct
 * - Skills paths are accessible
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
// Codex Plugin Tests
// ============================================================

console.log('\n🧪 Codex Plugin Manifest Tests\n');

test('Codex manifest exists', () => {
  fileExists('.codex-plugin/plugin.json', 'Codex manifest');
});

test('Codex manifest is valid JSON', () => {
  isValidJSON('.codex-plugin/plugin.json', 'Codex manifest');
});

test('Codex manifest references skills directory', () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.codex-plugin/plugin.json'), 'utf8')
  );
  assert.ok(manifest.skills, 'Codex manifest missing "skills" field');
  // Handle both relative (./skills) and absolute (skills) paths
  const skillsPath = manifest.skills.replace(/^\.\//, '');
  directoryExists(skillsPath, 'Codex manifest skills ref');
});

test('Codex manifest references hook map', () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.codex-plugin/plugin.json'), 'utf8')
  );
  assert.ok(manifest.hooks, 'Codex manifest missing "hooks" field');
  const hooksPath = manifest.hooks.replace(/^\.\//, '');
  fileExists(hooksPath, 'Codex manifest hooks ref');
});

test('Codex manifest has required metadata', () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.codex-plugin/plugin.json'), 'utf8')
  );
  assert.ok(manifest.name, 'Missing "name" field');
  assert.ok(manifest.version, 'Missing "version" field');
  assert.ok(manifest.description, 'Missing "description" field');
  assert.ok(manifest.author, 'Missing "author" field');
  assert.ok(manifest.license, 'Missing "license" field');
});

test('Codex manifest has interface section', () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.codex-plugin/plugin.json'), 'utf8')
  );
  assert.ok(manifest.interface, 'Codex manifest missing "interface" section');
  assert.ok(manifest.interface.displayName, 'Interface missing "displayName"');
  assert.ok(manifest.interface.capabilities, 'Interface missing "capabilities"');
});

test('Codex hook map exists', () => {
  fileExists('hooks/claude-codex-hooks.json', 'Codex hook map');
});

test('Codex hook map is valid JSON', () => {
  isValidJSON('hooks/claude-codex-hooks.json', 'Codex hook map');
});

test('Codex hook map has SessionStart event', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/claude-codex-hooks.json'), 'utf8')
  );
  assert.ok(hookMap.hooks.SessionStart, 'Missing "SessionStart" hook');
  assert.ok(Array.isArray(hookMap.hooks.SessionStart), '"SessionStart" must be an array');
});

test('Codex hook map has UserPromptSubmit event', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/claude-codex-hooks.json'), 'utf8')
  );
  assert.ok(hookMap.hooks.UserPromptSubmit, 'Missing "UserPromptSubmit" hook');
  assert.ok(Array.isArray(hookMap.hooks.UserPromptSubmit), '"UserPromptSubmit" must be an array');
});

test('Codex SessionStart hooks have timeout defined', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/claude-codex-hooks.json'), 'utf8')
  );
  const sessionStartHooks = hookMap.hooks.SessionStart;
  assert.ok(sessionStartHooks.length > 0, 'SessionStart hooks array is empty');
  sessionStartHooks.forEach((hookGroup) => {
    assert.ok(Array.isArray(hookGroup.hooks), 'SessionStart hooks must be an array');
    hookGroup.hooks.forEach((hook) => {
      assert.ok(hook.timeout !== undefined, 'Hook missing timeout property');
      assert.ok(hook.timeout > 0, 'Hook timeout must be positive');
    });
  });
});

test('Codex UserPromptSubmit hooks have timeout defined', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/claude-codex-hooks.json'), 'utf8')
  );
  const promptHooks = hookMap.hooks.UserPromptSubmit;
  assert.ok(promptHooks.length > 0, 'UserPromptSubmit hooks array is empty');
  promptHooks.forEach((hookGroup) => {
    assert.ok(Array.isArray(hookGroup.hooks), 'UserPromptSubmit hooks must be an array');
    hookGroup.hooks.forEach((hook) => {
      assert.ok(hook.timeout !== undefined, 'Hook missing timeout property');
      assert.ok(hook.timeout > 0, 'Hook timeout must be positive');
    });
  });
});

test('Codex hooks reference valid scripts', () => {
  const hookMap = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, 'hooks/claude-codex-hooks.json'), 'utf8')
  );
  
  // Check SessionStart hooks
  const sessionHooks = hookMap.hooks.SessionStart[0].hooks;
  sessionHooks.forEach((hook) => {
    const command = hook.command || hook.commandWindows;
    assert.ok(command, 'Hook missing command or commandWindows');
    // Check if script path exists in command
    const scriptMatch = command.match(/hooks\/([a-z-]+\.js)/);
    if (scriptMatch) {
      fileExists(`hooks/${scriptMatch[1]}`, 'Codex SessionStart script');
    }
  });

  // Check UserPromptSubmit hooks
  const promptHooks = hookMap.hooks.UserPromptSubmit[0].hooks;
  promptHooks.forEach((hook) => {
    const command = hook.command || hook.commandWindows;
    assert.ok(command, 'Hook missing command or commandWindows');
    // Check if script path exists in command
    const scriptMatch = command.match(/hooks\/([a-z-]+\.js)/);
    if (scriptMatch) {
      fileExists(`hooks/${scriptMatch[1]}`, 'Codex UserPromptSubmit script');
    }
  });
});

test('awtrix-notify skill exists (shared with Copilot)', () => {
  fileExists('skills/awtrix-notify/SKILL.md', 'awtrix-notify skill');
});

test('Required hook scripts exist', () => {
  fileExists('hooks/awtrix-activate.js', 'awtrix-activate.js');
  fileExists('hooks/awtrix-mode-tracker.js', 'awtrix-mode-tracker.js');
  fileExists('hooks/awtrix-runtime.js', 'awtrix-runtime.js');
});

test('Codex manifest references valid capabilities', () => {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(PROJECT_ROOT, '.codex-plugin/plugin.json'), 'utf8')
  );
  const capabilities = manifest.interface.capabilities || [];
  // Should include Instructions and/or Lifecycle hooks capability
  const validCapabilities = ['Instructions', 'Lifecycle hooks', 'Commands'];
  const hasValidCapability = capabilities.some(c => validCapabilities.includes(c));
  assert.ok(hasValidCapability, `No valid capabilities found. Expected one of: ${validCapabilities.join(', ')}`);
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
  console.log('✅ All Codex plugin manifest tests passed!\n');
  process.exit(0);
}
