#!/usr/bin/env bash
# notify.sh — send a notification to the AWTRIX3 pixel display
#
# Usage: notify.sh <event-type> "<message>"
#
# Event types:
#   start      — starting a long/complex task        (yellow)
#   success    — task completed successfully          (green)
#   error      — task failed or error encountered     (red, held)
#   attention  — user input or attention required     (orange, held)
#   build      — build result                         (blue)
#   test       — test run result                      (purple)
#
# The binary is resolved from PATH; set AWTRIX_HOST or --host to target device.

set -euo pipefail

EVENT_TYPE="${1:-}"
MESSAGE="${2:-}"

if [[ -z "$EVENT_TYPE" || -z "$MESSAGE" ]]; then
  echo "Usage: $(basename "$0") <event-type> \"<message>\"" >&2
  echo "Event types: start, success, error, attention, build, test" >&2
  exit 1
fi

# Verify the binary is available
if ! command -v awtrix3-client &>/dev/null; then
  echo "Error: awtrix3-client not found on PATH." >&2
  echo "Install with: go install github.com/terzi/awtrix3-client@latest" >&2
  exit 1
fi

# Truncate message to 30 characters — keeps the scrolling pixel display readable
MESSAGE="${MESSAGE:0:30}"

# Map event type → color and flags
COLOR="#FFFFFF"
HOLD=false
WAKEUP=true

case "$EVENT_TYPE" in
  start)
    COLOR="#FFAA00"
    ;;
  success)
    COLOR="#00FF00"
    ;;
  error|fail|failure)
    COLOR="#FF0000"
    HOLD=true
    ;;
  attention|input)
    COLOR="#FF8800"
    HOLD=true
    ;;
  build)
    COLOR="#00AAFF"
    ;;
  test)
    COLOR="#AA44FF"
    ;;
  *)
    echo "Warning: unknown event type '${EVENT_TYPE}', using white." >&2
    ;;
esac

# Build the command array
CMD=(awtrix3-client notify --text "$MESSAGE" --color "$COLOR")

if [[ "$WAKEUP" == true ]]; then
  CMD+=(--wakeup)
fi

if [[ "$HOLD" == true ]]; then
  CMD+=(--hold)
fi

# Echo the resolved command so the agent can log what was sent
echo "+ ${CMD[*]}" >&2

exec "${CMD[@]}"
