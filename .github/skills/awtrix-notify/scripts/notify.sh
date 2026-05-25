#!/usr/bin/env bash
# notify.sh — send a notification to the AWTRIX3 pixel display
#
# No installation required — uses "go run" to fetch and execute the client.
# Requires: Go 1.21+  (https://go.dev/dl/)
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

set -euo pipefail

EVENT_TYPE="${1:-}"
MESSAGE="${2:-}"

if [[ -z "$EVENT_TYPE" || -z "$MESSAGE" ]]; then
  echo "Usage: $(basename "$0") <event-type> \"<message>\"" >&2
  echo "Event types: start, success, error, attention, build, test" >&2
  exit 1
fi

# Verify Go is available
if ! command -v go &>/dev/null; then
  echo "Error: 'go' not found on PATH." >&2
  echo "Install Go 1.21+ from https://go.dev/dl/" >&2
  exit 1
fi

# Resolve AWTRIX_HOST — prompt interactively and persist if not set
if [[ -z "${AWTRIX_HOST:-}" ]]; then
  if [[ -t 0 ]]; then
    read -rp "AWTRIX_HOST not set. Enter device IP address: " AWTRIX_HOST
    if [[ -z "$AWTRIX_HOST" ]]; then
      echo "Error: no IP address provided." >&2
      exit 1
    fi
    export AWTRIX_HOST
    # Persist to the appropriate shell RC file
    RC_FILE="${HOME}/.bashrc"
    if [[ "${SHELL:-}" == */zsh ]]; then
      RC_FILE="${HOME}/.zshrc"
    fi
    echo "export AWTRIX_HOST=${AWTRIX_HOST}" >> "$RC_FILE"
    echo "Saved AWTRIX_HOST=${AWTRIX_HOST} to ${RC_FILE}" >&2
  else
    # Non-interactive (e.g. called from an AI agent): exit with a distinct code
    # so the caller knows to ask the user for the IP first.
    echo "Error: AWTRIX_HOST is not set." >&2
    echo "Set it with:" >&2
    echo "  export AWTRIX_HOST=192.168.x.x" >&2
    echo "  echo 'export AWTRIX_HOST=192.168.x.x' >> ~/.bashrc  # or ~/.zshrc" >&2
    exit 2
  fi
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
CMD=(go run github.com/terzinnorbert/awtrix3-client@latest notify --text "$MESSAGE" --color "$COLOR")

if [[ "$WAKEUP" == true ]]; then
  CMD+=(--wakeup)
fi

if [[ "$HOLD" == true ]]; then
  CMD+=(--hold)
fi

# Echo the resolved command (properly quoted) so the agent can log what was sent
echo "+ $(printf '%q ' "${CMD[@]}")" >&2

exec "${CMD[@]}"
