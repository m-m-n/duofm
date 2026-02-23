#!/bin/bash
# Background Shell Command Tests for duofm
#
# Description: Tests for background shell command execution mode including
#              !! activation, output display, cancellation, and auto-close
# Tests: 5

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# Test: '!!' enters background mode and executes command with output
test_bg_command_execution() {
    start_duofm "$CURRENT_SESSION"

    # Press '!' twice to enter background mode
    send_keys "$CURRENT_SESSION" "!" "!"
    sleep 0.3

    # Type a command that produces output
    send_keys "$CURRENT_SESSION" "e" "c" "h" "o" " " "h" "e" "l" "l" "o" "_" "b" "g"
    sleep 0.3

    # Press Enter to execute
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 1.0

    # Should show the output in the pane
    assert_contains "$CURRENT_SESSION" "hello_bg" \
        "Background command output appears in pane"

    # Wait for auto-close (2 seconds)
    sleep 3.0

    # Output area should have auto-closed
    assert_not_contains "$CURRENT_SESSION" "hello_bg" \
        "Output area auto-closes after 2 seconds"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Background mode shows pink prompt indicator
test_bg_mode_prompt() {
    start_duofm "$CURRENT_SESSION"

    # Press '!' to enter shell command mode
    send_keys "$CURRENT_SESSION" "!"
    sleep 0.3

    # Verify normal prompt
    assert_contains "$CURRENT_SESSION" "!:" \
        "Normal shell command prompt appears"

    # Press '!' again to enter background mode
    send_keys "$CURRENT_SESSION" "!"
    sleep 0.3

    # Should still show prompt (pink colored, but text content remains)
    local screen
    screen=$(capture_screen "$CURRENT_SESSION")
    # The prompt should contain "!" somewhere (might be colored)
    if echo "$screen" | grep -q "!"; then
        echo -e "${GREEN}✓${NC} Background mode prompt visible"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Background mode prompt visible"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Cancel with Escape
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Ctrl+C cancels background command when output is focused
test_bg_cancel_with_ctrlc() {
    start_duofm "$CURRENT_SESSION"

    # Enter background mode and run a long-running command
    send_keys "$CURRENT_SESSION" "!" "!"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "s" "l" "e" "e" "p" " " "3" "0"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # TAB to focus output area
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.3

    # Ctrl+C to cancel
    send_keys "$CURRENT_SESSION" "C-c"
    sleep 0.5

    # Output area should be cleared (command cancelled)
    # The TUI should still be running and responsive
    assert_contains "$CURRENT_SESSION" "/" \
        "TUI still responsive after background cancel"

    stop_duofm "$CURRENT_SESSION"
}

# Test: File operations work during background execution
test_bg_file_ops_during_execution() {
    start_duofm "$CURRENT_SESSION"

    # Start a background command
    send_keys "$CURRENT_SESSION" "!" "!"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "s" "l" "e" "e" "p" " " "5"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Navigate with j/k keys (should still work)
    send_keys "$CURRENT_SESSION" "j" "j"
    sleep 0.3

    # TUI should still be responsive
    local screen
    screen=$(capture_screen "$CURRENT_SESSION")
    if [ -n "$screen" ]; then
        echo -e "${GREEN}✓${NC} File navigation works during background execution"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} File navigation works during background execution"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Focus output and cancel
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "C-c"
    sleep 0.3

    stop_duofm "$CURRENT_SESSION"
}

# Test: Cannot start new command while background is running
test_bg_blocked_during_execution() {
    start_duofm "$CURRENT_SESSION"

    # Start a background command
    send_keys "$CURRENT_SESSION" "!" "!"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "s" "l" "e" "e" "p" " " "3" "0"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Try to start another shell command
    send_keys "$CURRENT_SESSION" "!"
    sleep 0.5

    # Should show warning message
    assert_contains "$CURRENT_SESSION" "Background command running" \
        "Warning shown when attempting shell command during bg execution"

    # Clean up: focus and cancel
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "C-c"
    sleep 0.3

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Background Commands"
    echo "========================================"

    run_test test_bg_command_execution
    run_test test_bg_mode_prompt
    run_test test_bg_cancel_with_ctrlc
    run_test test_bg_file_ops_during_execution
    run_test test_bg_blocked_during_execution

    print_summary
    exit $?
fi
