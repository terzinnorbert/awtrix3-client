---
name: awtrix-notify
description: "Send real-time notifications to the AWTRIX3 pixel display during agent conversations. Use when: starting a long or complex multi-step task, task completed successfully, task failed or build failed, tests passed or failed, user attention or input required, error encountered, progress update needed. Sends color-coded text to the pixel clock via awtrix3-client notify CLI."
argument-hint: "<event-type> \"<message>\"  — e.g.  success \"Build complete\""
---

# AWTRIX Notify

Push short, color-coded messages to the AWTRIX3 pixel display at key moments during a conversation.
No installation required — the client is fetched and run directly via `go run`.

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

### 2. Ensure AWTRIX_HOST is set

Check whether `AWTRIX_HOST` is available in the current environment:

**Linux / macOS:**
```bash
echo "${AWTRIX_HOST:-NOT SET}"
```

**Windows (PowerShell):**
```powershell
[System.Environment]::GetEnvironmentVariable("AWTRIX_HOST")
```

**If it is not set**, ask the user for their device IP using `vscode_askQuestions`, then persist it:

**Linux / macOS** — append to the correct RC file and export for this session:
```bash
# Replace 192.168.x.x with the IP the user provided
echo 'export AWTRIX_HOST=192.168.x.x' >> ~/.bashrc   # use ~/.zshrc if the shell is zsh
export AWTRIX_HOST=192.168.x.x
```

**Windows (PowerShell)** — persist to user environment and set for this session:
```powershell
# Replace 192.168.x.x with the IP the user provided
[Environment]::SetEnvironmentVariable("AWTRIX_HOST", "192.168.x.x", "User")
$env:AWTRIX_HOST = "192.168.x.x"
```

### 3. Run the helper script

Detect the OS and run the matching script from the project root.
No binary installation needed — just Go 1.21+.

**Linux / macOS:**
```bash
bash .github/skills/awtrix-notify/scripts/notify.sh <event-type> "<message>"
```

**Windows (PowerShell):**
```powershell
pwsh .github/skills/awtrix-notify/scripts/notify.ps1 <event-type> "<message>"
```

The scripts read `AWTRIX_HOST` from the environment and run the client via `go run`.
If the script exits with code `2`, it means `AWTRIX_HOST` was not set — go back to step 2.

### 4. Fallback: call the client directly (no script needed)

```bash
go run github.com/terzinnorbert/awtrix3-client@latest notify \
  --text "<message>" --color "<color>" --wakeup [--hold]
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

## Setup (first time only)

### 1. Install Go 1.21+

Check whether Go is already available:
```bash
go version
```

If not installed, download from [https://go.dev/dl/](https://go.dev/dl/).
On first use, `go run` will download and cache the client automatically — no manual install step needed.

### 2. Set AWTRIX_HOST

Follow step 2 in the Procedure above. The value only needs to be set once; it persists across sessions.

### 3. Verify

Run a test notification — the display should show "Skill ready" in green:

```bash
go run github.com/terzinnorbert/awtrix3-client@latest notify \
  --text "Skill ready" --color "#00FF00" --wakeup
```
