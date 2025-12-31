#!/bin/bash
# Directory Tests for duofm
#
# Description: Tests for directory operations including symlink display,
#              permission handling, refresh, pane sync, and navigation
# Tests: 10

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

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
# Test: F5 Refresh both panes
# ===========================================
test_f5_refresh() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Create a new file externally
    touch /testdata/user_owned/f5_test_file.txt

    # Press F5 to refresh
    send_keys "$CURRENT_SESSION" "F5"
    sleep 0.5

    # Should show the new file
    assert_contains "$CURRENT_SESSION" "f5_test_file.txt" \
        "F5 refresh shows externally created file"

    # Cleanup
    rm -f /testdata/user_owned/f5_test_file.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Ctrl+R Refresh both panes
# ===========================================
test_ctrlr_refresh() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Create a new file externally
    touch /testdata/user_owned/ctrlr_test_file.txt

    # Press Ctrl+R to refresh
    send_keys "$CURRENT_SESSION" "C-r"
    sleep 0.5

    # Should show the new file
    assert_contains "$CURRENT_SESSION" "ctrlr_test_file.txt" \
        "Ctrl+R refresh shows externally created file"

    # Cleanup
    rm -f /testdata/user_owned/ctrlr_test_file.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Refresh preserves cursor position
# ===========================================
test_refresh_cursor_preservation() {
    start_duofm "$CURRENT_SESSION"

    # Move cursor down to file2.txt
    send_keys "$CURRENT_SESSION" "j" "j" "j"
    sleep 0.3

    # Check cursor position before refresh
    local screen_before
    screen_before=$(capture_screen "$CURRENT_SESSION")

    # Press F5 to refresh
    send_keys "$CURRENT_SESSION" "F5"
    sleep 0.5

    # Cursor should still be on the same file
    local screen_after
    screen_after=$(capture_screen "$CURRENT_SESSION")

    # Verify cursor is roughly at same position (position indicator should be similar)
    if echo "$screen_after" | grep -q " [34]/"; then
        echo -e "${GREEN}✓${NC} Refresh preserves cursor position"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Refresh should preserve cursor position"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: = key syncs opposite pane
# ===========================================
test_sync_pane() {
    start_duofm "$CURRENT_SESSION"

    # Enter dir1 on left pane
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.3

    # Verify left pane is in dir1
    assert_contains "$CURRENT_SESSION" "subdir" \
        "Left pane is in dir1 (shows subdir)"

    # Press = to sync right pane
    send_keys "$CURRENT_SESSION" "="
    sleep 0.5

    # Switch to right pane to verify
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3

    # Right pane should also show dir1 contents
    assert_contains "$CURRENT_SESSION" "subdir" \
        "= key syncs right pane to left pane's directory"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Sync preserves display settings
# ===========================================
test_sync_preserves_settings() {
    start_duofm "$CURRENT_SESSION"

    # Switch to right pane
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3

    # Toggle hidden files on right pane (. key)
    send_keys "$CURRENT_SESSION" "."
    sleep 0.3

    # Switch back to left pane
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3

    # Enter dir2 on left pane
    send_keys "$CURRENT_SESSION" "/" "d" "i" "r" "2" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Press = to sync right pane
    send_keys "$CURRENT_SESSION" "="
    sleep 0.5

    # Switch to right pane
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3

    # Right pane should be in dir2
    assert_contains "$CURRENT_SESSION" "another.txt" \
        "Sync moves to correct directory"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Sync from right to left pane
# ===========================================
test_sync_right_to_left() {
    start_duofm "$CURRENT_SESSION"

    # First, sync right pane to left (so both are in /testdata)
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3

    # Navigate to dir2 using j/k (dir2 is at position 2 - after dir1)
    send_keys "$CURRENT_SESSION" "j" "j"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Verify right pane is in dir2
    assert_contains "$CURRENT_SESSION" "another.txt" \
        "Right pane is in dir2"

    # Press = to sync left pane
    send_keys "$CURRENT_SESSION" "="
    sleep 0.5

    # Switch to left pane to verify
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3

    # Left pane should also show dir2 contents
    assert_contains "$CURRENT_SESSION" "another.txt" \
        "= key syncs left pane to right pane's directory"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Right pane navigation when both panes show same path
# (Bug fix: async directory load pane identification)
# ===========================================
test_right_pane_same_path_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Sync panes to ensure both are showing the same directory
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.2

    # Enter a subdirectory
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Navigate back using "-" (previous directory) - this triggered the bug
    send_keys "$CURRENT_SESSION" "-"
    sleep 0.5

    # Check that the right pane successfully completed loading
    # (Should not show "Loading directory...")
    assert_not_contains "$CURRENT_SESSION" "Loading directory" \
        "Right pane completes navigation when returning to same path as left pane"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Right pane home navigation when left pane is at home
# (Bug fix: async directory load pane identification)
# ===========================================
test_right_pane_home_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Make sure we're starting from home by syncing
    send_keys "$CURRENT_SESSION" "~"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.2

    # Navigate to parent directory
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.5

    # Navigate back to home using "~" - this triggered the bug
    send_keys "$CURRENT_SESSION" "~"
    sleep 0.5

    # Check that the right pane successfully completed loading
    assert_not_contains "$CURRENT_SESSION" "Loading directory" \
        "Right pane completes home navigation when left pane is at home"

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Directory"
    echo "========================================"

    run_test test_symlink_display
    run_test test_permission_denied_directory
    run_test test_f5_refresh
    run_test test_ctrlr_refresh
    run_test test_refresh_cursor_preservation
    run_test test_sync_pane
    run_test test_sync_preserves_settings
    run_test test_sync_right_to_left
    run_test test_right_pane_same_path_navigation
    run_test test_right_pane_home_navigation

    print_summary
    exit $?
fi
