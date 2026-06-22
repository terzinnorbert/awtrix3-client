---
name: awtrix-notify
description: "Send color-coded AWTRIX3 notifications during coding tasks from plugin-installed skills."
argument-hint: "<event-type> \"<message>\""
---

# AWTRIX Notify

Send a short color-coded notification to AWTRIX3.

## Usage

Use:

```text
/awtrix-notify <event-type> "<message>"
```

Supported event types:
- `start`
- `success`
- `error`
- `attention`
- `build`
- `test`

## Event Mapping

- `start` -> yellow
- `success` -> green
- `error` -> red (held)
- `attention` -> orange (held)
- `build` -> blue
- `test` -> purple

## Behavior

The plugin hooks call the existing notify scripts from this repository:
- `.github/skills/awtrix-notify/scripts/notify.sh`
- `.github/skills/awtrix-notify/scripts/notify.ps1`

If `AWTRIX_HOST` is missing, script execution fails non-blocking and the host session continues.
