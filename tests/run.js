#!/usr/bin/env node

/**
 * Test Suite Runner
 * 
 * Runs all plugin manifest smoke tests and reports results.
 */

const { spawn } = require('child_process');
const path = require('path');

const tests = [
  'copilot-plugin.test.js',
  'codex-plugin.test.js',
  'hooks.test.js',
  'hooks-windows.test.js',
  'event-parity.test.js',
  'release-readiness.test.js'
];

const PROJECT_ROOT = path.join(__dirname, '..');
let failed = false;

console.log('🚀 Running plugin manifest smoke tests...\n');

function runTest(testFile, index) {
  return new Promise((resolve) => {
    const testPath = path.join(__dirname, testFile);
    const child = spawn('node', [testPath], {
      cwd: PROJECT_ROOT,
      stdio: 'inherit'
    });

    child.on('exit', (code) => {
      if (code !== 0) {
        failed = true;
      }
      resolve();
    });
  });
}

async function main() {
  for (let i = 0; i < tests.length; i++) {
    await runTest(tests[i], i);
  }

  console.log('\n' + '='.repeat(60));
  if (failed) {
    console.log('❌ Some tests failed!');
    console.log('='.repeat(60) + '\n');
    process.exit(1);
  } else {
    console.log('✅ All plugin manifest smoke tests passed!');
    console.log('='.repeat(60) + '\n');
    process.exit(0);
  }
}

main();
