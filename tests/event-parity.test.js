#!/usr/bin/env node

/**
 * Event parity validation tests.
 *
 * Validates runtime event semantics against the legacy AWTRIX mapping contract:
 * - start/success/error/attention/build/test
 * - aliases: fail/failure -> error, input -> attention
 * - color/hold/wakeup behavior parity
 */

const assert = require('assert');
const { resolveEventProfile } = require('../hooks/awtrix-runtime');

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

console.log('\n🧪 Event Parity Validation Tests\n');

const expected = {
  start: { normalizedEvent: 'start', color: '#FFAA00', hold: false, wakeup: true },
  success: { normalizedEvent: 'success', color: '#00FF00', hold: false, wakeup: true },
  error: { normalizedEvent: 'error', color: '#FF0000', hold: true, wakeup: true },
  attention: { normalizedEvent: 'attention', color: '#FF8800', hold: true, wakeup: true },
  build: { normalizedEvent: 'build', color: '#00AAFF', hold: false, wakeup: true },
  test: { normalizedEvent: 'test', color: '#AA44FF', hold: false, wakeup: true },
};

Object.entries(expected).forEach(([eventType, profile]) => {
  test(`event parity: ${eventType} matches legacy color/hold/wakeup`, () => {
    const actual = resolveEventProfile(eventType);
    assert.deepStrictEqual(actual, profile);
  });
});

test('event parity: fail alias maps to error profile', () => {
  const actual = resolveEventProfile('fail');
  assert.deepStrictEqual(actual, expected.error);
});

test('event parity: failure alias maps to error profile', () => {
  const actual = resolveEventProfile('failure');
  assert.deepStrictEqual(actual, expected.error);
});

test('event parity: input alias maps to attention profile', () => {
  const actual = resolveEventProfile('input');
  assert.deepStrictEqual(actual, expected.attention);
});

test('event parity: unknown event falls back to attention profile', () => {
  const actual = resolveEventProfile('anything-else');
  assert.deepStrictEqual(actual, expected.attention);
});

console.log('\n' + '='.repeat(60));
if (TESTS_FAILED.length > 0) {
  console.log(`❌ ${TESTS_FAILED.length} test(s) failed, ${TESTS_PASSED.length} passed`);
  console.log('='.repeat(60) + '\n');
  process.exit(1);
} else {
  console.log(`✅ All ${TESTS_PASSED.length} event parity tests passed`);
  console.log('='.repeat(60) + '\n');
  process.exit(0);
}
