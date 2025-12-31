#!/bin/bash
# Shell Command Tests for duofm
#
# Description: Tests for shell command execution mode including
#              entering mode, input handling, and mode restrictions
# Tests: 6

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# Test: '!' key enters shell command mode
test_shell_command_mode_enter() {
    start_duofm "$CURRENT_SESSION"

    # Press '!' to enter shell command mode
    send_keys "$CURRENT_SESSION" "!"
    sleep 0.3

    # Should show shell command prompt with "!"
    assert_contains "$CURRENT_SESSION" "!:" \
        "Shell command prompt appears"

    # Cancel with Escape
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    # Prompt should be gone
    assert_not_contains "$CURRENT_SESSION" "!:" \
        "Shell command prompt closes after Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Shell command mode shows typed characters
test_shell_command_input() {
    start_duofm "$CURRENT_SESSION"

    # Press '!' to enter shell command mode
    send_keys "$CURRENT_SESSION" "!"
    sleep 0.3

    # Type a simple command (without spaces, as tmux Space key handling can be tricky)
    send_keys "$CURRENT_SESSION" "l" "s"
    sleep 0.3

    # Should show typed command
    assert_contains "$CURRENT_SESSION" "ls" \
        "Typed command is visible"

    # Cancel with Escape
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Empty Enter exits shell command mode without execution
test_shell_command_empty_enter() {
    start_duofm "$CURRENT_SESSION"

    # Press '!' to enter shell command mode
    send_keys "$CURRENT_SESSION" "!"
    sleep 0.3

    # Should show prompt
    assert_contains "$CURRENT_SESSION" "!:" \
        "Shell command prompt appears"

    # Press Enter with empty input
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Should exit mode without prompt
    assert_not_contains "$CURRENT_SESSION" "!:" \
        "Empty Enter exits shell command mode"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Shell command ignored when dialog is open
test_shell_command_ignored_with_dialog() {
    start_duofm "$CURRENT_SESSION"

    # Open help dialog
    send_keys "$CURRENT_SESSION" "?"
    sleep 0.3

    # Verify dialog is open
    assert_contains "$CURRENT_SESSION" "Keybindings" \
        "Help dialog is open"

    # Try to enter shell command mode
    send_keys "$CURRENT_SESSION" "!"
    sleep 0.3

    # Help dialog should still be visible, no shell prompt
    assert_contains "$CURRENT_SESSION" "Keybindings" \
        "Help dialog still visible after ! key"

    assert_not_contains "$CURRENT_SESSION" "!:" \
        "Shell command prompt does not appear when dialog is open"

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Shell command ignored during search mode
test_shell_command_ignored_during_search() {
    start_duofm "$CURRENT_SESSION"

    # Enter search mode
    send_keys "$CURRENT_SESSION" "/"
    sleep 0.3

    # Verify search mode is active
    assert_contains "$CURRENT_SESSION" "/:" \
        "Search mode is active"

    # Try to enter shell command mode with '!'
    send_keys "$CURRENT_SESSION" "!"
    sleep 0.3

    # Search should still be active and '!' added to search string
    # The '!' character should be part of the search input, not trigger shell command mode
    local screen
    screen=$(capture_screen "$CURRENT_SESSION")

    # Should NOT show shell command prompt
    if echo "$screen" | grep -q "!:" | grep -v "/:"; then
        echo -e "${RED}✗${NC} Shell command mode should not activate during search"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    else
        echo -e "${GREEN}✓${NC} Shell command ignored during search mode"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi

    # Cancel search
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Help dialog shows shell command key
test_help_shows_shell_command() {
    start_duofm "$CURRENT_SESSION"

    # Open help dialog
    send_keys "$CURRENT_SESSION" "?"
    sleep 0.3

    # Help should show shell command entry
    assert_contains "$CURRENT_SESSION" "!" \
        "Help dialog shows shell command key"

    # Close help
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Shell"
    echo "========================================"

    run_test test_shell_command_mode_enter
    run_test test_shell_command_input
    run_test test_shell_command_empty_enter
    run_test test_shell_command_ignored_with_dialog
    run_test test_shell_command_ignored_during_search
    run_test test_help_shows_shell_command

    print_summary
    exit $?
fi
