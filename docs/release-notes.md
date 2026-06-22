# Release Notes v0.2.2

## Summary

Version 0.2.2 finalizes the AWTRIX skill-to-plugin migration for Copilot and Codex.
The plugin metadata now matches the release version, and CI validates the plugin
surface alongside the Go build with explicit plugin integrity checks.

## Highlights

- Bumped both plugin manifests to `0.2.2`
- Added CI coverage for plugin integrity and release metadata alignment
- Published plugin-first release notes and kept the README migration guidance current
- Retained the shared hook runtime and host-specific manifests introduced during migration

## Migration impact

- Legacy `.github/skills/awtrix-notify/` distribution remains removed from active use
- New installs should use the plugin entrypoints documented in [README.md](../README.md)
- The `skills/` and `commands/` directories remain the canonical package surface

## Validation

- `node tests/run.js`
- `go vet ./...`
- `go build -ldflags "-s -w" ./...`