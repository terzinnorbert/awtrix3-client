# AWTRIX Plugin Portability Matrix

This document maps the AWTRIX3 plugin architecture across supported AI coding agents (Copilot and Codex), detailing exact file locations, capability tiers, and installation requirements for each host.

---

## Portability Overview

The AWTRIX plugin uses a **Ponytail-style multi-host adapter** pattern:
- **Plugin-level integration** (preferred): Host-specific manifests and hooks enable automatic event tracking and notifications.
- **Instruction-level fallback**: Standalone AGENTS.md instructions allow manual notification control when plugins are unavailable.

---

## Capability Matrix by Host

| Host | Plugin Support | Manifest Path | Capability Tier | Event Binding | Lifecycle Hooks |
|------|----------------|---------------|-----------------|---------------|-----------------|
| **GitHub Copilot** | ✅ Full | `.github/plugin/plugin.json` | Plugin + Instructions | Session start, user prompt submit | Node.js + bash/PowerShell |
| **Claude (Codex CLI)** | ✅ Full | `.codex-plugin/plugin.json` | Plugin + Instructions | Session start, prompt events | Node.js + bash/PowerShell |
| **AGENTS.md fallback** | ✅ Manual | `AGENTS.md` | Instructions only | None (on-demand) | Not applicable |

---

## File Map: Copilot Plugin

### Installation & Discovery

- **Manifest**: `.github/plugin/plugin.json`
- **Install entry point**: Defined in Copilot plugin marketplace / VS Code extensions UI
- **Auto-discovery**: Copilot scans the manifest on load; no manual registration needed

### Asset Locations

| Asset | Path | Purpose |
|-------|------|---------|
| Plugin metadata | `.github/plugin/plugin.json` | Host manifest with name, version, and asset references |
| Skills directory | `skills/` | Plugin-discoverable skill definitions |
| Awtrix skill | `skills/awtrix-notify/SKILL.md` | Skill interface for `/awtrix-notify` command |
| Commands directory | `commands/` | Command definitions (if exposed via plugin) |
| Awtrix command | `commands/awtrix-notify.toml` | Command metadata and argument schema |
| Hook manifest | `hooks/copilot-hooks.json` | Copilot-specific event-to-hook bindings |
| Hook scripts | `hooks/awtrix-*.js` | Shared Node.js implementations (activate, mode-tracker, runtime) |

### Lifecycle Hook Events

**Copilot hook map** (`hooks/copilot-hooks.json`):

| Event | Binding | Handler | Timeout |
|-------|---------|---------|---------|
| `sessionStart` | Plugin load / session init | `hooks/awtrix-activate.js` | 5 seconds |
| `userPromptSubmitted` | User submits a prompt | `hooks/awtrix-mode-tracker.js` | 5 seconds |

**Command execution**:
- **Linux/macOS**: `node "${PLUGIN_ROOT}/hooks/awtrix-activate.js"`
- **Windows**: `node "${PLUGIN_ROOT}\hooks\awtrix-activate.js"`

**Failure handling**: Non-blocking; hook timeout or errors do not interrupt the Copilot session.

---

## File Map: Claude (Codex CLI) Plugin

### Installation & Discovery

- **Manifest**: `.codex-plugin/plugin.json`
- **Install entry point**: Codex CLI plugin load path (typically `~/.claude/plugins/` or project-scoped `.claude/` folder)
- **Auto-discovery**: Codex reads the manifest on startup; no manual registration needed

### Asset Locations

| Asset | Path | Purpose |
|-------|------|---------|
| Plugin metadata | `.codex-plugin/plugin.json` | Host manifest with name, version, and asset references |
| Skills directory | `skills/` | Plugin-discoverable skill definitions (shared with Copilot) |
| Awtrix skill | `skills/awtrix-notify/SKILL.md` | Skill interface for `/awtrix-notify` command |
| Hook manifest | `hooks/claude-codex-hooks.json` | Codex-specific event-to-hook bindings |
| Hook scripts | `hooks/awtrix-*.js` | Shared Node.js implementations (activate, mode-tracker, runtime) |

### Lifecycle Hook Events

**Codex hook map** (`hooks/claude-codex-hooks.json`):

| Event | Matcher | Handler | Timeout | Status Message |
|-------|---------|---------|---------|-----------------|
| `SessionStart` | `startup`, `resume`, `clear`, `compact` | `hooks/awtrix-activate.js` | 5 seconds | "Loading AWTRIX plugin..." |
| `UserPromptSubmit` | (all prompts) | `hooks/awtrix-mode-tracker.js` | 5 seconds | "Tracking AWTRIX events..." |

**Command execution**:
- **Linux/macOS**: `node "${CLAUDE_PLUGIN_ROOT:-$PLUGIN_ROOT}/hooks/awtrix-activate.js"`
- **Windows**: `if (Get-Command node) { node "$env:CLAUDE_PLUGIN_ROOT\hooks\awtrix-activate.js" }`

**Failure handling**: Non-blocking; hook errors are silently caught (`|| exit 0`); missing Node.js does not prevent Codex from starting.

---

## File Map: AGENTS.md Fallback (Instruction-Only)

### Installation & Discovery

- **Source**: `AGENTS.md` in repository root
- **Activation**: Read by Copilot/Codex when instructions-only mode is active or plugins are unavailable
- **User interaction**: On-demand via explicit `/awtrix-notify <event-type> "<message>"` invocations

### Asset Locations

| Asset | Path | Purpose |
|-------|------|---------|
| Instructions | `AGENTS.md` | Agent-level directives and fallback command syntax |
| Notify scripts | `.github/skills/awtrix-notify/scripts/notify.sh` | Linux/macOS notification delivery |
| Notify scripts | `.github/skills/awtrix-notify/scripts/notify.ps1` | Windows notification delivery |
| Runtime dependencies | `main.go`, `go.mod` | AWTRIX3 client (invoked by notify scripts) |

### Command Interface

```
/awtrix-notify <event-type> "<message>"
```

**Supported event types**: `start`, `success`, `error`, `attention`, `build`, `test`

### Event Parity Reference

The plugin and instruction-only paths share the same AWTRIX event semantics.

| Event type | Color | `--hold` | `--wakeup` |
|------------|-------|:--------:|:----------:|
| `start` | `#FFAA00` yellow | No | Yes |
| `success` | `#00FF00` green | No | Yes |
| `error` / `fail` / `failure` | `#FF0000` red | Yes | Yes |
| `attention` / `input` | `#FF8800` orange | Yes | Yes |
| `build` | `#00AAFF` blue | No | Yes |
| `test` | `#AA44FF` purple | No | Yes |

**Execution model**:
1. User or agent issues the command manually.
2. Instruction redirects to helper script based on OS detection.
3. Helper script reads `AWTRIX_HOST` env var and calls `go run` client.
4. Go client sends HTTP notification to AWTRIX3 device at `${AWTRIX_HOST}:7070`.

**Failure handling**: Blocking; if `AWTRIX_HOST` is unset, script exits with code 2 and prompts user to configure.

---

## Cross-Platform Compatibility

### Node.js Requirement

All lifecycle hooks depend on **Node.js 14+**:

- **Check (Copilot)**: Assumed available by default in Copilot environments.
- **Check (Codex)**: Explicit runtime check in hook command: `command -v node >/dev/null 2>&1`
- **Check (AGENTS.md)**: Go runtime required; Node.js not needed.

### Go Runtime Requirement

AWTRIX notification delivery requires **Go 1.21+** (invoked via `go run`):

- **Plugin path**: Not required by hooks; hooks delegate to AWTRIX_HOST via existing client.
- **AGENTS.md path**: Required by `.github/skills/awtrix-notify/scripts/notify.sh` and `notify.ps1`.

### Environment Variables

| Variable | Scope | Required | Default | Purpose |
|----------|-------|----------|---------|---------|
| `AWTRIX_HOST` | All hooks & scripts | No | None | IP address of AWTRIX3 device (e.g., `192.168.1.100`) |
| `PLUGIN_ROOT` | Copilot hooks | Yes | Set by host | Path to plugin root directory |
| `CLAUDE_PLUGIN_ROOT` | Codex hooks | No | Falls back to `$PLUGIN_ROOT` | Codex-specific plugin root |

### Platform-Specific Paths

**Bash/sh (Linux/macOS)**:
```bash
"${PLUGIN_ROOT}/hooks/awtrix-activate.js"
```

**PowerShell (Windows)**:
```powershell
"${PLUGIN_ROOT}\hooks\awtrix-activate.js"
node "$env:PLUGIN_ROOT\hooks\awtrix-activate.js"
```

---

## Shared Hook Implementation

All lifecycle hooks delegate to a common Node.js runtime helper:

| File | Purpose |
|------|---------|
| `hooks/awtrix-activate.js` | Session-start handler; emits "start" event to AWTRIX |
| `hooks/awtrix-mode-tracker.js` | Prompt-submit handler; parses task type and emits corresponding event |
| `hooks/awtrix-runtime.js` | Shared runtime utilities (host detection, output formatting, AWTRIX_HOST resolution) |

**Runtime guarantee**: If `AWTRIX_HOST` is unset or unreachable, hooks exit silently without blocking the session.

---

## Installation & Trust Flow

### Copilot Plugin

1. User opens Copilot in VS Code.
2. Copilot discovers and installs plugin from `.github/plugin/plugin.json`.
3. On next session, `sessionStart` hook executes `awtrix-activate.js`.
4. Hook checks `AWTRIX_HOST` and sends startup notification (if set).

### Codex CLI Plugin

1. User places `.codex-plugin/plugin.json` in Codex plugins directory or repo `.claude/` folder.
2. On `codex` CLI startup, plugin is loaded and hooks are registered.
3. On next session or prompt, Codex executes lifecycle hooks.
4. Hooks check `AWTRIX_HOST` and send notifications (if set).

### AGENTS.md Fallback

1. Plugin installation fails or user prefers manual control.
2. User invokes `/awtrix-notify success "Done"` in chat.
3. Agent reads `AGENTS.md` instructions and executes helper script.
4. Script checks `AWTRIX_HOST` and sends notification via Go client.

---

## Deprecations & Legacy Paths

The following legacy paths are **removed** from the distribution:

- `.github/skills/awtrix-notify/` — Replaced by `skills/awtrix-notify/`

The `.github/skills/` directory is now **empty**; plugin manifests and instructions both reference the new `skills/` and `commands/` directories at the repository root.

---

## Validation Checklist

- [x] Copilot plugin manifest exists and references correct asset paths.
- [x] Codex plugin manifest exists and references correct asset paths.
- [x] Skill definition moved to plugin-discoverable location (`skills/awtrix-notify/`).
- [x] Hook maps define both `sessionStart` and `userPromptSubmitted` events.
- [x] Hooks use non-blocking error handling for missing `AWTRIX_HOST` or Node.js.
- [x] AGENTS.md instructions remain as on-demand fallback.
- [x] Cross-platform paths (bash vs PowerShell) are syntactically correct.
- [ ] Manual install and execution tests pass on Linux and Windows.
- [ ] CI validates manifest paths and referenced files exist.

---

## Related Documentation

- [Plugin Architecture & Contracts](../PLAN.md#phase-2---plugin-manifests-and-hook-maps)
- [AGENTS.md — Fallback Instructions](../AGENTS.md)
- [README — Installation & Usage](../README.md)
- [Hook Implementation Details](../hooks/)
