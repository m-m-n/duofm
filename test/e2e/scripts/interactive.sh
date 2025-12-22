#!/bin/bash
# Interactive E2E testing helper for Claude Code
# Usage: docker run --rm -it duofm-e2e-test /e2e/scripts/interactive.sh "<commands>"
#
# Example:
#   docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "j j Enter"
#   docker run --rm duofm-e2e-test /e2e/scripts/interactive.sh "/ t e s t Enter"

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

SESSION="interactive_$$"
COMMANDS="$1"

echo "=== duofm Interactive Test ==="
echo "Commands: $COMMANDS"
echo ""

# Start duofm
start_duofm "$SESSION"

# Execute commands if provided
if [ -n "$COMMANDS" ]; then
    for cmd in $COMMANDS; do
        case "$cmd" in
            Enter)   send_keys "$SESSION" "Enter" ;;
            Escape)  send_keys "$SESSION" "Escape" ;;
            Tab)     send_keys "$SESSION" "Tab" ;;
            Space)   send_keys "$SESSION" "Space" ;;
            C-c)     send_keys "$SESSION" "C-c" ;;
            C-f)     send_keys "$SESSION" "C-f" ;;
            Up)      send_keys "$SESSION" "Up" ;;
            Down)    send_keys "$SESSION" "Down" ;;
            Left)    send_keys "$SESSION" "Left" ;;
            Right)   send_keys "$SESSION" "Right" ;;
            WAIT)    sleep 1 ;;
            *)       send_keys "$SESSION" "$cmd" ;;
        esac
    done
fi

# Capture and display screen
echo "=== Screen Capture ==="
capture_screen "$SESSION"

# Cleanup
stop_duofm "$SESSION"
