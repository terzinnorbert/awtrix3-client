# AGENTS.md — awtrix3-client

Project-level instructions for AI coding agents (OpenAI Codex CLI, and other agents that read AGENTS.md).

---

## AWTRIX Notify

Push short, color-coded messages to the AWTRIX3 pixel display at key moments during a task.
No binary installation needed — the client runs directly via `go run`.

### When to send a notification

Send one at each of these moments without being asked:

| Situation | Event type | Example message |
|-----------|------------|-----------------|
| Starting a long or complex task | `start` | `"Planning migration"` |
| Task completed successfully | `success` | `"Done: 14 files updated"` |
| Task failed or error encountered | `error` | `"Build failed: see logs"` |
| User attention or input required | `attention` | `"Waiting for input"` |
| Build finished | `build` | `"Build: 0 errors 2 warn"` |
| Test run finished | `test` | `"Tests: 42 pass 0 fail"` |

### Message guidelines

- **Max 30 characters** — text scrolls horizontally; truncate if needed
- Be specific: `"Tests: 12 pass 1 fail"` beats `"Done"`
- Include counts when relevant
- Skip trailing punctuation (`.` `!`)

### How to send

Before calling the script, verify `AWTRIX_HOST` is set in the environment.
If it is not set, ask the user for their device IP address, then persist it:

**Linux / macOS:**
```bash
echo 'export AWTRIX_HOST=192.168.x.x' >> ~/.bashrc   # use ~/.zshrc for zsh
export AWTRIX_HOST=192.168.x.x
```

**Windows:**
```powershell
[Environment]::SetEnvironmentVariable("AWTRIX_HOST", "192.168.x.x", "User")
$env:AWTRIX_HOST = "192.168.x.x"
```

Then run the helper script from the project root:

**Linux / macOS:**
```bash
bash .github/skills/awtrix-notify/scripts/notify.sh <event-type> "<message>"
```

**Windows:**
```powershell
pwsh .github/skills/awtrix-notify/scripts/notify.ps1 <event-type> "<message>"
```

If the script exits with code `2`, `AWTRIX_HOST` was not set — ask the user and persist it first (see above).

If the script is unavailable, call the client directly via `go run`:

```bash
go run github.com/terzinnorbert/awtrix3-client@latest notify \
  --text "<message>" --color "<color>" --wakeup [--hold]
```

### Color and flags reference

| Event type | Color | `--hold` |
|------------|-------|:--------:|
| `start` | `#FFAA00` yellow | — |
| `success` | `#00FF00` green | — |
| `error` / `fail` / `failure` | `#FF0000` red | ✓ |
| `attention` / `input` | `#FF8800` orange | ✓ |
| `build` | `#00AAFF` blue | — |
| `test` | `#AA44FF` purple | — |

`--hold` keeps the message on screen until the hardware button is pressed.
Always include `--wakeup` so the display activates even when sleeping.

### Setup (first time only)

#### 1. Install Go 1.21+

```bash
go version   # check if already installed
```

If not installed, download from [https://go.dev/dl/](https://go.dev/dl/).
On first use, `go run` downloads and caches the client automatically.

#### 2. Set the device host

```bash
# Linux / macOS
export AWTRIX_HOST=192.168.x.x
echo 'export AWTRIX_HOST=192.168.x.x' >> ~/.bashrc   # or ~/.zshrc
```

```powershell
# Windows
$env:AWTRIX_HOST = "192.168.x.x"
[Environment]::SetEnvironmentVariable("AWTRIX_HOST", "192.168.x.x", "User")
```

#### 3. Verify

```bash
go run github.com/terzinnorbert/awtrix3-client@latest notify \
  --text "Skill ready" --color "#00FF00" --wakeup
```
