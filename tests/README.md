# Tests

Plugin manifest smoke tests for the AWTRIX3 Copilot and Codex integration.

## Overview

These tests validate:
- Plugin manifest JSON validity for both Copilot and Codex
- All referenced files and directories exist
- Hook maps reference valid hook scripts
- Required metadata is present in manifests
- Skills and commands are accessible

## Test Files

- `copilot-plugin.test.js` - Validates Copilot plugin manifest and configuration
- `codex-plugin.test.js` - Validates Codex plugin manifest and configuration
- `run.js` - Test suite runner (runs all tests)

## Running Tests

Run all tests:

```bash
node tests/run.js
```

Run individual tests:

```bash
node tests/copilot-plugin.test.js
node tests/codex-plugin.test.js
```

## Test Coverage

### Copilot Plugin Tests (17 tests)
- Manifest existence and JSON validity
- Skill, command, and hook file references
- Hook map structure and events
- Timeout and failure policy definitions
- Required metadata fields

### Codex Plugin Tests (16 tests)
- Manifest existence and JSON validity
- Skill and hook file references
- Interface section and capabilities
- Hook map structure and event matchers
- Hook command syntax (bash and Windows)
- Shared script dependencies

## CI Integration

To integrate these tests into CI/CD, add to your workflow:

```yaml
- name: Run plugin manifest tests
  run: node tests/run.js
```

## Next Steps

- [ ] Add hook compatibility tests
- [ ] Add Windows-specific hook syntax checks
- [ ] Add event parity validation tests
- [ ] Add manual integration tests for both hosts
