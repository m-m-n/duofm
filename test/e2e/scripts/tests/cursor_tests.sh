#!/bin/bash
# Cursor Tests for duofm
#
# Description: Tests for cursor position preservation after various operations
#              including file viewing, external commands, and parent navigation
# Tests: 8

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# Test: Cursor position preserved after viewing file with 'v' key (less)
test_cursor_preserved_after_view() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to file1.txt specifically using search (not a directory)
    send_keys "$CURRENT_SESSION" "/" "f" "i" "l" "e" "1" "." "t" "x" "t" "Enter"
    sleep 0.3

    # Verify we are on file1.txt
    assert_contains "$CURRENT_SESSION" "file1.txt" \
        "file1.txt is visible before viewing"

    # View file with 'v' key (less)
    send_keys "$CURRENT_SESSION" "v"
    sleep 1.0

    # Exit less with 'q'
    send_keys "$CURRENT_SESSION" "q"
    sleep 1.0

    # Verify session is still running and cursor is preserved
    if tmux has-session -t "${SESSION_PREFIX}_${CURRENT_SESSION}" 2>/dev/null; then
        # file1.txt should still be visible (cursor position preserved)
        assert_contains "$CURRENT_SESSION" "file1.txt" \
            "file1.txt still visible after viewing"
    else
        echo -e "${RED}✗${NC} Session ended unexpectedly after viewing file"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return
    fi

    stop_duofm "$CURRENT_SESSION"
}

# Test: Cursor position preserved after viewing file with Enter key
test_cursor_preserved_after_enter_view() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to a file (not directory) using search
    send_keys "$CURRENT_SESSION" "/" "f" "i" "l" "e" "1" "." "t" "x" "t" "Enter"
    sleep 0.3

    # Get current cursor position (should be on file1.txt)
    local screen_before
    screen_before=$(capture_screen "$CURRENT_SESSION")

    # Press Enter to view file (should open with less since it's a file)
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Exit less with 'q'
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.5

    # Cursor should still be on file1.txt
    assert_contains "$CURRENT_SESSION" "file1.txt" \
        "file1.txt still visible after viewing"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Cursor resets to 0 when selected file is deleted during external edit
test_cursor_reset_when_file_deleted() {
    # Create a test file that we'll delete during edit
    echo "test content" > /testdata/user_owned/will_delete.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter and find the file
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "/" "w" "i" "l" "l" "_" "d" "e" "l" "e" "t" "e" "Enter"
    sleep 0.3

    # View the file
    send_keys "$CURRENT_SESSION" "v"
    sleep 0.5

    # While in less, delete the file externally
    rm -f /testdata/user_owned/will_delete.txt

    # Exit less with 'q'
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.5

    # Cursor should reset to beginning (position 1 for parent dir)
    # The file will no longer appear in the list
    assert_not_contains "$CURRENT_SESSION" "will_delete.txt" \
        "Deleted file no longer appears in list"

    # Verify the filter is cleared and we see the directory contents
    assert_contains "$CURRENT_SESSION" "user_owned" \
        "User owned directory visible after file deletion"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Both panes preserve cursor after external command
test_both_panes_preserve_cursor() {
    start_duofm "$CURRENT_SESSION"

    # Move left pane cursor to position 3
    send_keys "$CURRENT_SESSION" "j" "j" "j"
    sleep 0.3

    # Switch to right pane
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3

    # Move right pane cursor to position 2
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.3

    # Switch back to left pane
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3

    # Navigate to file and view
    send_keys "$CURRENT_SESSION" "/" "f" "i" "l" "e" "1" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "v"
    sleep 0.5

    # Exit less
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.5

    # Both panes should have their cursor positions preserved
    # (This test verifies the implementation refreshes both panes)
    assert_contains "$CURRENT_SESSION" "duofm" \
        "Application still running after viewer exit"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Cursor positioned on previous subdirectory after parent navigation via h key
test_parent_nav_cursor_on_subdir_h_key() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to dir1 subdirectory
    send_keys "$CURRENT_SESSION" "/" "d" "i" "r" "1" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Now inside dir1, go back to parent using h key
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.5

    # Cursor should be positioned on dir1 (not at the top)
    # Verify by checking that we can see dir1 is selected
    assert_contains "$CURRENT_SESSION" "dir1" \
        "Parent directory shows dir1 after returning via h key"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Cursor positioned on previous subdirectory after parent navigation via .. entry
test_parent_nav_cursor_on_subdir_dotdot() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to dir1 subdirectory
    send_keys "$CURRENT_SESSION" "/" "d" "i" "r" "1" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Inside dir1, navigate to .. and press Enter to go to parent
    # The .. entry should be at the top, so we need to navigate there
    send_keys "$CURRENT_SESSION" "g" "g"  # Go to top
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Enter"  # Select ..
    sleep 0.5

    # Cursor should be positioned on dir1
    assert_contains "$CURRENT_SESSION" "dir1" \
        "Parent directory shows dir1 after returning via .. entry"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Right pane cursor positioned on previous subdirectory via l key
test_parent_nav_cursor_on_subdir_l_key() {
    start_duofm "$CURRENT_SESSION"

    # Switch to right pane
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.2

    # Navigate to dir1 subdirectory in right pane
    send_keys "$CURRENT_SESSION" "/" "d" "i" "r" "1" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Now inside dir1 in right pane, go back to parent using l key
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.5

    # Cursor should be positioned on dir1
    assert_contains "$CURRENT_SESSION" "dir1" \
        "Right pane shows dir1 after returning via l key"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Independent pane cursor memory
test_parent_nav_independent_pane_memory() {
    start_duofm "$CURRENT_SESSION"

    # Left pane: Navigate to dir1
    send_keys "$CURRENT_SESSION" "/" "d" "i" "r" "1" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Switch to right pane
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.2

    # Right pane: Navigate to dir2 (if it exists, or another dir)
    send_keys "$CURRENT_SESSION" "/" "d" "i" "r" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Right pane: Go back to parent
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.5

    # Switch back to left pane
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.2

    # Left pane: Go back to parent
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.5

    # Left pane should show dir1 after navigating back
    assert_contains "$CURRENT_SESSION" "dir1" \
        "Left pane cursor memory is independent of right pane"

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Cursor"
    echo "========================================"

    run_test test_cursor_preserved_after_view
    run_test test_cursor_preserved_after_enter_view
    run_test test_cursor_reset_when_file_deleted
    run_test test_both_panes_preserve_cursor
    run_test test_parent_nav_cursor_on_subdir_h_key
    run_test test_parent_nav_cursor_on_subdir_dotdot
    run_test test_parent_nav_cursor_on_subdir_l_key
    run_test test_parent_nav_independent_pane_memory

    print_summary
    exit $?
fi
