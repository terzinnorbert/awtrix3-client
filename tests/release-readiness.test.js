#!/usr/bin/env node

/**
 * Release readiness checks.
 *
 * Validates that the release metadata stays aligned across plugin manifests,
 * release docs, and CI coverage.
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

const PROJECT_ROOT = path.join(__dirname, '..');
const EXPECTED_VERSION = '0.2.2';
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

function readJSON(relativePath) {
  const fullPath = path.join(PROJECT_ROOT, relativePath);
  return JSON.parse(fs.readFileSync(fullPath, 'utf8'));
}

function readText(relativePath) {
  return fs.readFileSync(path.join(PROJECT_ROOT, relativePath), 'utf8');
}

console.log('\n🧪 Release Readiness Checks\n');

test('Copilot manifest version matches release version', () => {
  const manifest = readJSON('.github/plugin/plugin.json');
  assert.strictEqual(manifest.version, EXPECTED_VERSION);
});

test('Codex manifest version matches release version', () => {
  const manifest = readJSON('.codex-plugin/plugin.json');
  assert.strictEqual(manifest.version, EXPECTED_VERSION);
});

test('README release notes mention the release version', () => {
  const readme = readText('README.md');
  assert.ok(readme.includes(`v${EXPECTED_VERSION}`), 'README missing release version reference');
  assert.ok(readme.includes('docs/release-notes.md'), 'README missing release notes link');
});

test('Release notes document exists and matches the release version', () => {
  const releaseNotes = readText('docs/release-notes.md');
  assert.ok(releaseNotes.includes(`v${EXPECTED_VERSION}`), 'Release notes missing version heading');
  assert.ok(releaseNotes.includes('plugin integrity checks'), 'Release notes missing CI summary');
});

test('CI workflow runs plugin integrity tests', () => {
  const workflow = readText('.github/workflows/ci.yml');
  assert.ok(workflow.includes('actions/setup-node@v4'), 'CI missing Node setup');
  assert.ok(workflow.includes('node tests/run.js'), 'CI missing plugin test run');
});

console.log('\n' + '='.repeat(60));
if (TESTS_FAILED.length > 0) {
  console.log(`❌ ${TESTS_FAILED.length} test(s) failed, ${TESTS_PASSED.length} passed`);
  console.log('='.repeat(60) + '\n');
  process.exit(1);
} else {
  console.log(`✅ All ${TESTS_PASSED.length} release readiness checks passed`);
  console.log('='.repeat(60) + '\n');
  process.exit(0);
}