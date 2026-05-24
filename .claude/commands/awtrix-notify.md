---
description: Send a color-coded notification to the AWTRIX3 pixel display
allowed-tools: Bash
---

Send a notification to the AWTRIX3 pixel display using the `awtrix3-client notify` CLI.

**Arguments:** $ARGUMENTS
**Expected format:** `<event-type> "<message>"`

## Event types

| Event type | Color | `--hold` | Use when |
|------------|-------|:--------:|----------|
| `start` | yellow `#FFAA00` | — | Starting a long or complex task |
| `success` | green `#00FF00` | — | Task completed successfully |
| `error` / `fail` | red `#FF0000` | ✓ | Task failed or error encountered |
| `attention` / `input` | orange `#FF8800` | ✓ | User attention or input required |
| `build` | blue `#00AAFF` | — | Build finished |
| `test` | purple `#AA44FF` | — | Test run finished |

`--hold` means the message stays on screen until the button is pressed.
All calls use `--wakeup` so the display activates even if sleeping.

## Message guidelines

- Max 30 characters — text scrolls horizontally
- Be specific: `"Tests: 12 pass 1 fail"` beats `"Done"`
- Include counts when relevant
- Skip trailing punctuation

## Steps

1. Parse the event type and message from the arguments above.
2. Determine the OS, then run the matching script from the project root:

**Linux / macOS:**
```bash
bash .github/skills/awtrix-notify/scripts/notify.sh <event-type> "<message>"
```

**Windows:**
```powershell
pwsh .github/skills/awtrix-notify/scripts/notify.ps1 <event-type> "<message>"
```

3. If the helper script is missing, call the binary directly:

```bash
awtrix3-client notify --text "<message>" --color "<color>" --wakeup [--hold]
```

## Setup (first time only)

### 1. Set the device host

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

### 2. Install the binary

**Pre-built (recommended):** Download from [GitHub Releases](https://github.com/terzinnorbert/awtrix3-client/releases/latest).

```bash
tar -xzf awtrix3-client_*_linux_amd64.tar.gz
sudo mv awtrix3-client /usr/local/bin/
```

**Build from source (requires Go 1.21+):**

```bash
go install github.com/terzinnorbert/awtrix3-client@latest
# Add to PATH if needed:
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc && source ~/.bashrc
```

### 3. Verify

```bash
awtrix3-client notify --text "Skill ready" --color "#00FF00" --wakeup
```
