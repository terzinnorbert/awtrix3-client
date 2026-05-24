---
name: awtrix-notify
description: "Send real-time notifications to the AWTRIX3 pixel display during agent conversations. Use when: starting a long or complex multi-step task, task completed successfully, task failed or build failed, tests passed or failed, user attention or input required, error encountered, progress update needed. Sends color-coded text to the pixel clock via awtrix3-client notify CLI."
argument-hint: "<event-type> \"<message>\"  — e.g.  success \"Build complete\""
---

# AWTRIX Notify

Push short, color-coded messages to the AWTRIX3 pixel display at key moments during a conversation using this project's `awtrix3-client notify` CLI.

## When to Use

Send a notification at each of these moments:

| Situation | Event Type | Example Message |
|-----------|------------|-----------------|
| Starting a long / complex task | `start` | `"Planning migration"` |
| Task completed successfully | `success` | `"Done: 14 files updated"` |
| Task failed or error encountered | `error` | `"Build failed: see logs"` |
| User attention or input required | `attention` | `"Waiting for input"` |
| Build finished | `build` | `"Build: 0 errors 2 warn"` |
| Test run finished | `test` | `"Tests: 42 pass 0 fail"` |

## Message Crafting Guidelines

- **Max 30 characters** — text scrolls horizontally; longer messages are truncated by the script
- Be specific: `"Tests: 12 pass 1 fail"` beats `"Done"`
- Include numbers when relevant: counts, file totals, error counts
- Skip trailing punctuation (`.` `!`) — it wastes pixels
- Avoid abbreviations the user won't recognise

## Procedure

### 1. Choose event type and draft message

Pick one event type from the table above. Keep the message under 30 characters.

### 2. Run the helper script

Detect the OS and run the matching script from the project root:

**Linux / macOS:**
```bash
bash .github/skills/awtrix-notify/scripts/notify.sh <event-type> "<message>"
```

**Windows (PowerShell):**
```powershell
pwsh .github/skills/awtrix-notify/scripts/notify.ps1 <event-type> "<message>"
```

The scripts inherit `AWTRIX_HOST` from the environment — the same variable this project already uses.

### 3. Fallback: direct CLI (no script needed)

If the script is unavailable, call the binary directly:

```bash
awtrix3-client notify --text "<message>" --color "<color>" --wakeup [--hold]
```

## Color & Flags Reference

| Event | Color | `--hold` | Notes |
|-------|-------|:--------:|-------|
| `start` | `#FFAA00` yellow | — | Wakes display if off |
| `success` | `#00FF00` green | — | Wakes display if off |
| `error` | `#FF0000` red | ✓ | Held until button-dismissed |
| `attention` | `#FF8800` orange | ✓ | Held until button-dismissed |
| `build` | `#00AAFF` blue | — | Wakes display if off |
| `test` | `#AA44FF` purple | — | Wakes display if off |

All calls use `--wakeup` so the display activates even if it was sleeping.

## Binary Setup

### Option A — Pre-built binary (recommended)

Download the archive for your platform from the [GitHub Releases](https://github.com/terzinnorbert/awtrix3-client/releases/latest) page, extract it, and place the binary on your `PATH`.

| Platform | Archive | Binary name |
|----------|---------|-------------|
| Linux x86-64 | `awtrix3-client_*_linux_amd64.tar.gz` | `awtrix3-client` |
| Linux ARM64 | `awtrix3-client_*_linux_arm64.tar.gz` | `awtrix3-client` |
| macOS (Intel) | `awtrix3-client_*_darwin_amd64.tar.gz` | `awtrix3-client` |
| macOS (Apple Silicon) | `awtrix3-client_*_darwin_arm64.tar.gz` | `awtrix3-client` |
| Windows x86-64 | `awtrix3-client_*_windows_amd64.zip` | `awtrix3-client.exe` |

**Linux / macOS — extract and install:**
```bash
tar -xzf awtrix3-client_*_linux_amd64.tar.gz
sudo mv awtrix3-client /usr/local/bin/
```

**Windows — extract and add to PATH:**
```powershell
Expand-Archive awtrix3-client_*_windows_amd64.zip .
Move-Item awtrix3-client.exe "$env:USERPROFILE\bin\"
# Ensure $env:USERPROFILE\bin is in your PATH
```

### Option B — Build from source

```bash
go install github.com/terzinnorbert/awtrix3-client@latest
```

Requires Go 1.21+.

### Configure the target device

```bash
export AWTRIX_HOST=192.168.1.100   # Linux / macOS (add to ~/.bashrc or ~/.zshrc)
$env:AWTRIX_HOST = "192.168.1.100" # Windows PowerShell (add to $PROFILE)
```

Alternatively pass `--host 192.168.1.100` directly to every `awtrix3-client` call.
