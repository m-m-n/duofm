#!/bin/bash
# E2E Test Runner for duofm

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

echo "========================================"
echo "duofm E2E Tests"
echo "========================================"
echo "Working directory: $(pwd)"
echo "Test fixtures:"
ls -la
echo ""

# ===========================================
# Test: Basic Startup
# ===========================================
test_basic_startup() {
    start_duofm "$CURRENT_SESSION"

    assert_contains "$CURRENT_SESSION" "duofm" \
        "Title bar shows 'duofm'"

    assert_contains "$CURRENT_SESSION" "file1.txt" \
        "Shows file1.txt in listing"

    assert_contains "$CURRENT_SESSION" "dir1" \
        "Shows dir1 directory"

    assert_contains "$CURRENT_SESSION" "?:help" \
        "Status bar shows help hint"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Navigation with j/k
# ===========================================
test_jk_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Initial position should be 1
    assert_cursor_position "$CURRENT_SESSION" "1" \
        "Initial cursor position is 1"

    # Move down twice
    send_keys "$CURRENT_SESSION" "j" "j"

    assert_cursor_position "$CURRENT_SESSION" "3" \
        "Cursor moves to 3 after jj"

    # Move up once
    send_keys "$CURRENT_SESSION" "k"

    assert_cursor_position "$CURRENT_SESSION" "2" \
        "Cursor moves to 2 after k"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Enter directory
# ===========================================
test_enter_directory() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to dir1 and enter
    send_keys "$CURRENT_SESSION" "j"  # Move to dir1

    # Check we're on dir1
    local screen
    screen=$(capture_screen "$CURRENT_SESSION")

    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Should now be inside dir1
    assert_contains "$CURRENT_SESSION" "subdir" \
        "Shows subdir after entering dir1"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Parent directory navigation
# ===========================================
test_parent_directory() {
    start_duofm "$CURRENT_SESSION"

    # Enter a directory first
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.3

    # Go back to parent with .. or backspace
    send_keys "$CURRENT_SESSION" "Enter"  # Select ".." and enter
    sleep 0.3

    assert_contains "$CURRENT_SESSION" "file1.txt" \
        "Back to parent directory shows file1.txt"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Pane switching with h/l
# ===========================================
test_pane_switching() {
    start_duofm "$CURRENT_SESSION"

    # Initially left pane is active - move to right pane
    send_keys "$CURRENT_SESSION" "l"

    # Navigate in right pane
    send_keys "$CURRENT_SESSION" "j" "j"

    # Switch back to left pane
    send_keys "$CURRENT_SESSION" "h"

    # Should still show the files
    assert_contains "$CURRENT_SESSION" "file1.txt" \
        "Left pane still shows files after switch"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Help dialog
# ===========================================
test_help_dialog() {
    start_duofm "$CURRENT_SESSION"

    # Open help
    send_keys "$CURRENT_SESSION" "?"

    assert_contains "$CURRENT_SESSION" "Keybindings" \
        "Help dialog shows Keybindings"

    # Close help with Esc
    send_keys "$CURRENT_SESSION" "Escape"

    assert_not_contains "$CURRENT_SESSION" "Keybindings" \
        "Help dialog closed after Escape"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Search/Filter with /
# ===========================================
test_search_filter() {
    start_duofm "$CURRENT_SESSION"

    # Start search
    send_keys "$CURRENT_SESSION" "/"

    assert_contains "$CURRENT_SESSION" "/:" \
        "Search prompt appears"

    # Type search pattern
    send_keys "$CURRENT_SESSION" "f" "i" "l" "e"

    # Confirm search
    send_keys "$CURRENT_SESSION" "Enter"

    # Should show filtered results
    assert_contains "$CURRENT_SESSION" "file1.txt" \
        "Filtered list shows file1.txt"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Quit with q
# ===========================================
test_quit() {
    start_duofm "$CURRENT_SESSION"

    # Verify app is running
    assert_contains "$CURRENT_SESSION" "duofm" \
        "App is running"

    # Quit
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.5

    # Session should be gone
    if tmux has-session -t "${SESSION_PREFIX}_${CURRENT_SESSION}" 2>/dev/null; then
        echo -e "${RED}✗${NC} App should have quit"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    else
        echo -e "${GREEN}✓${NC} App quit successfully"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
}

# ===========================================
# Test: Symlink display
# ===========================================
test_symlink_display() {
    start_duofm "$CURRENT_SESSION"

    assert_contains "$CURRENT_SESSION" "link_to_file1.txt" \
        "Shows symlink in listing"

    # Symlinks should show arrow
    assert_contains "$CURRENT_SESSION" "->" \
        "Symlink shows arrow indicator"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Ctrl+C double press to quit
# ===========================================
test_ctrlc_quit() {
    start_duofm "$CURRENT_SESSION"

    # First Ctrl+C
    send_keys "$CURRENT_SESSION" "C-c"

    assert_contains "$CURRENT_SESSION" "Ctrl+C again" \
        "First Ctrl+C shows confirmation message"

    # Second Ctrl+C
    send_keys "$CURRENT_SESSION" "C-c"
    sleep 0.5

    # Session should be gone
    if tmux has-session -t "${SESSION_PREFIX}_${CURRENT_SESSION}" 2>/dev/null; then
        echo -e "${RED}✗${NC} App should have quit after double Ctrl+C"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    else
        echo -e "${GREEN}✓${NC} Double Ctrl+C quit works"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
}

# ===========================================
# Test: Permission denied on directory access
# ===========================================
test_permission_denied_directory() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to no_access directory
    # First find and select no_access using search
    send_keys "$CURRENT_SESSION" "/" "n" "o" "_" "a" "c" "c" "e" "s" "s" "Enter"
    sleep 0.3

    # Now the filter is applied, Enter again to enter the directory
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Should show error message in status bar (case-insensitive check)
    local screen
    screen=$(capture_screen "$CURRENT_SESSION")

    # Check for permission denied message (could be various formats)
    if echo "$screen" | grep -qiE "permission|denied|access|cannot"; then
        echo -e "${GREEN}✓${NC} Shows permission denied error for inaccessible directory"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        # If no error shown, check if we're still in the same directory (didn't enter)
        if echo "$screen" | grep -qF "/testdata"; then
            echo -e "${GREEN}✓${NC} Cannot enter inaccessible directory (stayed in parent)"
            TESTS_RUN=$((TESTS_RUN + 1))
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "${RED}✗${NC} Should show permission error or stay in parent directory"
            echo "  Screen content:"
            echo "$screen" | head -10 | sed 's/^/    /'
            TESTS_RUN=$((TESTS_RUN + 1))
            TESTS_FAILED=$((TESTS_FAILED + 1))
        fi
    fi

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Cannot delete root-owned file
# ===========================================
test_cannot_delete_root_file() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to root_owned directory
    send_keys "$CURRENT_SESSION" "/" "r" "o" "o" "t" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Select protected.txt (move past ..)
    send_keys "$CURRENT_SESSION" "j"

    # Try to delete
    send_keys "$CURRENT_SESSION" "d"
    sleep 0.3

    # Confirm deletion
    send_keys "$CURRENT_SESSION" "y"
    sleep 0.5

    # Should show error (permission denied or similar)
    local screen
    screen=$(capture_screen "$CURRENT_SESSION")

    # File should still exist (deletion failed) or error shown
    if echo "$screen" | grep -qiF "permission" || echo "$screen" | grep -qiF "error" || echo "$screen" | grep -qF "protected.txt"; then
        echo -e "${GREEN}✓${NC} Cannot delete root-owned file (permission check works)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Should show error or file should remain"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Can delete user-owned file
# ===========================================
test_can_delete_user_file() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Select deletable.txt (move past ..)
    send_keys "$CURRENT_SESSION" "j"

    # Verify file exists
    assert_contains "$CURRENT_SESSION" "deletable.txt" \
        "User-owned file exists before deletion"

    # Delete the file
    send_keys "$CURRENT_SESSION" "d"
    sleep 0.3

    # Confirm deletion
    send_keys "$CURRENT_SESSION" "y"
    sleep 0.5

    # File should be gone
    assert_not_contains "$CURRENT_SESSION" "deletable.txt" \
        "User-owned file deleted successfully"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Run all tests
# ===========================================
run_test test_basic_startup
run_test test_jk_navigation
run_test test_enter_directory
run_test test_parent_directory
run_test test_pane_switching
run_test test_help_dialog
run_test test_search_filter
run_test test_quit
run_test test_symlink_display
run_test test_ctrlc_quit

# Permission tests (require non-root user)
run_test test_permission_denied_directory
run_test test_cannot_delete_root_file
run_test test_can_delete_user_file

# Print summary and exit with appropriate code
print_summary
exit $?
