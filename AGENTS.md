# AGENTS.md — awtrix3-client

Project-level instructions for AI coding agents (OpenAI Codex CLI, and other agents that read AGENTS.md).

---

## AWTRIX Notify

Push short, color-coded messages to the AWTRIX3 pixel display at key moments during a task using the `awtrix3-client notify` CLI.

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

Run the helper script from the project root:

**Linux / macOS:**
```bash
bash .github/skills/awtrix-notify/scripts/notify.sh <event-type> "<message>"
```

**Windows:**
```powershell
pwsh .github/skills/awtrix-notify/scripts/notify.ps1 <event-type> "<message>"
```

If the script is unavailable, call the binary directly:

```bash
awtrix3-client notify --text "<message>" --color "<color>" --wakeup [--hold]
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

#### 1. Set the device host

```bash
# Linux / macOS
export AWTRIX_HOST=192.168.x.x
echo 'export AWTRIX_HOST=192.168.x.x' >> ~/.bashrc
```

```powershell
# Windows
$env:AWTRIX_HOST = "192.168.x.x"
[Environment]::SetEnvironmentVariable("AWTRIX_HOST", "192.168.x.x", "User")
```

#### 2. Install the binary

**Pre-built (recommended):** Download from [GitHub Releases](https://github.com/terzinnorbert/awtrix3-client/releases/latest).

```bash
# Linux / macOS
tar -xzf awtrix3-client_*_linux_amd64.tar.gz
sudo mv awtrix3-client /usr/local/bin/
```

```powershell
# Windows
Expand-Archive awtrix3-client_*_windows_amd64.zip .
Move-Item awtrix3-client.exe "$env:USERPROFILE\bin\"
```

**Build from source (requires Go 1.21+):**

```bash
go install github.com/terzinnorbert/awtrix3-client@latest
# Add to PATH if needed:
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc && source ~/.bashrc
```

#### 3. Verify

```bash
awtrix3-client notify --text "Skill ready" --color "#00FF00" --wakeup
```
