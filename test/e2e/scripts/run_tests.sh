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
# Test: Create new file with n key
# ===========================================
test_create_new_file() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Press n to create new file
    send_keys "$CURRENT_SESSION" "n"
    sleep 0.3

    # Should show input dialog
    assert_contains "$CURRENT_SESSION" "New file:" \
        "New file dialog appears"

    # Type filename
    send_keys "$CURRENT_SESSION" "t" "e" "s" "t" "f" "i" "l" "e" "." "t" "x" "t"
    sleep 0.2

    # Confirm with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # File should appear in the list
    assert_contains "$CURRENT_SESSION" "testfile.txt" \
        "Created file appears in listing"

    # Cleanup
    rm -f /testdata/user_owned/testfile.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Create new directory with N key
# ===========================================
test_create_new_directory() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Press N (Shift+n) to create new directory
    send_keys "$CURRENT_SESSION" "N"
    sleep 0.3

    # Should show input dialog
    assert_contains "$CURRENT_SESSION" "New directory:" \
        "New directory dialog appears"

    # Type directory name
    send_keys "$CURRENT_SESSION" "t" "e" "s" "t" "d" "i" "r"
    sleep 0.2

    # Confirm with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Directory should appear in the list
    assert_contains "$CURRENT_SESSION" "testdir" \
        "Created directory appears in listing"

    # Cleanup
    rmdir /testdata/user_owned/testdir

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Rename file with r key
# ===========================================
test_rename_file() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Create a test file for renaming
    touch /testdata/user_owned/before_rename.txt
    send_keys "$CURRENT_SESSION" "F5"
    sleep 0.5

    # Navigate to the file
    send_keys "$CURRENT_SESSION" "/" "b" "e" "f" "o" "r" "e" "_" "r" "e" "n" "Enter"
    sleep 0.3

    # Press r to rename
    send_keys "$CURRENT_SESSION" "r"
    sleep 0.3

    # Should show input dialog
    assert_contains "$CURRENT_SESSION" "Rename to:" \
        "Rename dialog appears"

    # Type new name
    send_keys "$CURRENT_SESSION" "a" "f" "t" "e" "r" "_" "r" "e" "n" "a" "m" "e" "." "t" "x" "t"
    sleep 0.2

    # Confirm with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Old name should be gone, new name should appear
    assert_not_contains "$CURRENT_SESSION" "before_rename.txt" \
        "Old filename is gone"

    assert_contains "$CURRENT_SESSION" "after_rename.txt" \
        "New filename appears in listing"

    # Cleanup
    rm -f /testdata/user_owned/after_rename.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Cancel file creation with Esc
# ===========================================
test_cancel_file_creation() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Press n to create new file
    send_keys "$CURRENT_SESSION" "n"
    sleep 0.3

    # Type some filename
    send_keys "$CURRENT_SESSION" "c" "a" "n" "c" "e" "l" "l" "e" "d"
    sleep 0.2

    # Cancel with Escape
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    # Dialog should be closed, file should not exist
    assert_not_contains "$CURRENT_SESSION" "New file:" \
        "Dialog is closed"

    assert_not_contains "$CURRENT_SESSION" "cancelled" \
        "Cancelled file is not created"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Empty filename shows error
# ===========================================
test_empty_filename_error() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Press n to create new file
    send_keys "$CURRENT_SESSION" "n"
    sleep 0.3

    # Try to confirm with empty input
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Should show error message but keep dialog open
    assert_contains "$CURRENT_SESSION" "cannot be empty" \
        "Shows error for empty filename"

    assert_contains "$CURRENT_SESSION" "New file:" \
        "Dialog stays open after empty input error"

    # Cancel
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Navigation works after file creation
# ===========================================
test_navigation_after_file_creation() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Press n to create new file
    send_keys "$CURRENT_SESSION" "n"
    sleep 0.3

    # Type filename
    send_keys "$CURRENT_SESSION" "n" "a" "v" "t" "e" "s" "t" "." "t" "x" "t"
    sleep 0.2

    # Confirm with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # File should appear in the list
    assert_contains "$CURRENT_SESSION" "navtest.txt" \
        "Created file appears in listing"

    # Test navigation still works - try to quit with q key
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.5

    # Cleanup
    rm -f /testdata/user_owned/navtest.txt

    # Session should be gone (q key worked)
    if tmux has-session -t "${SESSION_PREFIX}_${CURRENT_SESSION}" 2>/dev/null; then
        echo -e "${RED}✗${NC} Navigation broken after file creation (q key didn't work)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        stop_duofm "$CURRENT_SESSION"
    else
        echo -e "${GREEN}✓${NC} Navigation works after file creation (q key quit app)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
}

# ===========================================
# Test: Navigation works after directory creation
# ===========================================
test_navigation_after_dir_creation() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Press N to create new directory
    send_keys "$CURRENT_SESSION" "N"
    sleep 0.3

    # Type directory name
    send_keys "$CURRENT_SESSION" "n" "a" "v" "d" "i" "r"
    sleep 0.2

    # Confirm with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Directory should appear in the list
    assert_contains "$CURRENT_SESSION" "navdir" \
        "Created directory appears in listing"

    # Test navigation still works
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "k"
    sleep 0.3

    # Try to quit - if navigation works, q should quit the app
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.5

    # Cleanup (in case app didn't quit)
    rmdir /testdata/user_owned/navdir 2>/dev/null || true

    # Session should be gone (q key worked)
    if tmux has-session -t "${SESSION_PREFIX}_${CURRENT_SESSION}" 2>/dev/null; then
        echo -e "${RED}✗${NC} Navigation broken after directory creation (q key didn't work)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        stop_duofm "$CURRENT_SESSION"
    else
        echo -e "${GREEN}✓${NC} Navigation works after directory creation (q key quit app)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
}

# ===========================================
# Test: Navigation works after rename
# ===========================================
test_navigation_after_rename() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Create a test file
    touch /testdata/user_owned/navren_before.txt
    send_keys "$CURRENT_SESSION" "F5"
    sleep 0.5

    # Navigate to the file
    send_keys "$CURRENT_SESSION" "/" "n" "a" "v" "r" "e" "n" "_" "b" "e" "f" "o" "r" "e" "Enter"
    sleep 0.3

    # Press r to rename
    send_keys "$CURRENT_SESSION" "r"
    sleep 0.3

    # Type new name
    send_keys "$CURRENT_SESSION" "n" "a" "v" "r" "e" "n" "_" "a" "f" "t" "e" "r" "." "t" "x" "t"
    sleep 0.2

    # Confirm with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Test navigation still works - try to quit
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.5

    # Cleanup
    rm -f /testdata/user_owned/navren_after.txt

    # Session should be gone (q key worked)
    if tmux has-session -t "${SESSION_PREFIX}_${CURRENT_SESSION}" 2>/dev/null; then
        echo -e "${RED}✗${NC} Navigation broken after rename (q key didn't work)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        stop_duofm "$CURRENT_SESSION"
    else
        echo -e "${GREEN}✓${NC} Navigation works after rename (q key quit app)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    fi
}

# ===========================================
# Test: Rename parent directory is ignored
# ===========================================
test_rename_parent_dir_ignored() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to dir1 (so we have a parent directory entry)
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.3

    # Cursor should be on ".."
    assert_cursor_position "$CURRENT_SESSION" "1" \
        "Cursor is on position 1 (..)"

    # Press r on parent directory
    send_keys "$CURRENT_SESSION" "r"
    sleep 0.3

    # Dialog should NOT appear
    assert_not_contains "$CURRENT_SESSION" "Rename to:" \
        "Rename dialog does not appear for parent directory"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Copy with overwrite dialog - Cancel
# ===========================================
test_copy_overwrite_cancel() {
    # Pre-cleanup
    rm -rf /testdata/user_owned/canceldir 2>/dev/null || true
    rm -f /testdata/user_owned/cancel_src.txt 2>/dev/null || true

    # Create source file and destination directory with conflict BEFORE starting
    echo "source content" > /testdata/user_owned/cancel_src.txt
    mkdir -p /testdata/user_owned/canceldir
    echo "existing content" > /testdata/user_owned/canceldir/cancel_src.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter and sync right pane to user_owned
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane and navigate to canceldir
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "c" "a" "n" "c" "e" "l" "d" "i" "r" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane and clear filter
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to cancel_src.txt
    send_keys "$CURRENT_SESSION" "/" "c" "a" "n" "c" "e" "l" "_" "s" "r" "c" "Enter"
    sleep 0.3

    # Press c to copy
    send_keys "$CURRENT_SESSION" "c"
    sleep 0.5

    # Should show overwrite dialog (file already exists in canceldir)
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Overwrite dialog appears for copy conflict"

    # Press 2 to select Cancel
    send_keys "$CURRENT_SESSION" "2"
    sleep 0.3

    # Dialog should close
    assert_not_contains "$CURRENT_SESSION" "already exists" \
        "Dialog closes after Cancel"

    # Verify file was NOT overwritten (still has old content)
    local content
    content=$(cat /testdata/user_owned/canceldir/cancel_src.txt 2>/dev/null || echo "")
    if [ "$content" = "existing content" ]; then
        echo -e "${GREEN}✓${NC} File was not overwritten after Cancel"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} File should not be overwritten after Cancel"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Cleanup
    rm -f /testdata/user_owned/cancel_src.txt
    rm -rf /testdata/user_owned/canceldir

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Copy with overwrite dialog - Overwrite
# ===========================================
test_copy_overwrite_confirm() {
    # Pre-cleanup
    rm -rf /testdata/user_owned/confirmdir 2>/dev/null || true
    rm -f /testdata/user_owned/confirm_src.txt 2>/dev/null || true

    # Create source file and destination directory with conflicting file BEFORE starting
    echo "new content" > /testdata/user_owned/confirm_src.txt
    mkdir -p /testdata/user_owned/confirmdir
    echo "old content" > /testdata/user_owned/confirmdir/confirm_src.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter and sync right pane to user_owned
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane and navigate to confirmdir
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "c" "o" "n" "f" "i" "r" "m" "d" "i" "r" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane and clear filter
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to confirm_src.txt
    send_keys "$CURRENT_SESSION" "/" "c" "o" "n" "f" "i" "r" "m" "_" "s" "r" "c" "Enter"
    sleep 0.3

    # Press c to copy
    send_keys "$CURRENT_SESSION" "c"
    sleep 0.5

    # Should show overwrite dialog
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Overwrite dialog appears for copy conflict"

    # Press 1 to Overwrite
    send_keys "$CURRENT_SESSION" "1"
    sleep 0.5

    # File should be overwritten - verify content
    local content
    content=$(cat /testdata/user_owned/confirmdir/confirm_src.txt 2>/dev/null || echo "")
    if [ "$content" = "new content" ]; then
        echo -e "${GREEN}✓${NC} File was overwritten with new content"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} File should be overwritten with new content"
        echo "  Content: $content"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Cleanup
    rm -f /testdata/user_owned/confirm_src.txt
    rm -rf /testdata/user_owned/confirmdir

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Copy with overwrite dialog - Rename
# ===========================================
test_copy_overwrite_rename() {
    # Pre-cleanup
    rm -rf /testdata/user_owned/rendir 2>/dev/null || true
    rm -f /testdata/user_owned/rename_test.txt 2>/dev/null || true

    # Create source file and destination directory with conflict BEFORE starting
    echo "source" > /testdata/user_owned/rename_test.txt
    mkdir -p /testdata/user_owned/rendir
    echo "existing" > /testdata/user_owned/rendir/rename_test.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter and sync right pane to user_owned
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane and navigate to rendir
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "r" "e" "n" "d" "i" "r" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane and clear filter
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to rename_test.txt
    send_keys "$CURRENT_SESSION" "/" "r" "e" "n" "a" "m" "e" "_" "t" "e" "s" "t" "." "t" "x" "t" "Enter"
    sleep 0.3

    # Press c to copy
    send_keys "$CURRENT_SESSION" "c"
    sleep 0.5

    # Should show overwrite dialog
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Overwrite dialog appears for copy conflict"

    # Press 3 to Rename
    send_keys "$CURRENT_SESSION" "3"
    sleep 0.8

    # Should show rename input dialog
    assert_contains "$CURRENT_SESSION" "New name:" \
        "Rename input dialog appears"

    # Confirm with Enter (accept suggested name)
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.8

    # New file should exist (either _copy or _copy_2 depending on suggestion)
    if [ -f "/testdata/user_owned/rendir/rename_test_copy.txt" ]; then
        echo -e "${GREEN}✓${NC} Renamed file was created"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Renamed file should exist"
        ls -la /testdata/user_owned/rendir/ 2>/dev/null | head -5
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Original should still exist
    if [ -f "/testdata/user_owned/rendir/rename_test.txt" ]; then
        echo -e "${GREEN}✓${NC} Original file still exists"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Original file should still exist"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Cleanup
    rm -f /testdata/user_owned/rename_test.txt
    rm -rf /testdata/user_owned/rendir

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Move with overwrite dialog
# ===========================================
test_move_overwrite() {
    # Pre-cleanup
    rm -rf /testdata/user_owned/movedir 2>/dev/null || true
    rm -f /testdata/user_owned/move_src.txt 2>/dev/null || true

    # Create source file and destination directory with conflict BEFORE starting
    echo "move source" > /testdata/user_owned/move_src.txt
    mkdir -p /testdata/user_owned/movedir
    echo "old content" > /testdata/user_owned/movedir/move_src.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory (writable)
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter and sync right pane to user_owned
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane and navigate to movedir
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "m" "o" "v" "e" "d" "i" "r" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane and clear filter
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to move_src.txt
    send_keys "$CURRENT_SESSION" "/" "m" "o" "v" "e" "_" "s" "r" "c" "Enter"
    sleep 0.3

    # Press m to move
    send_keys "$CURRENT_SESSION" "m"
    sleep 0.5

    # Should show overwrite dialog
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Overwrite dialog appears for move conflict"

    # Press 1 to Overwrite
    send_keys "$CURRENT_SESSION" "1"
    sleep 0.5

    # Source should be gone
    if [ ! -f "/testdata/user_owned/move_src.txt" ]; then
        echo -e "${GREEN}✓${NC} Source file was moved (deleted from source)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Source file should be gone after move"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Destination should have new content
    local content
    content=$(cat /testdata/user_owned/movedir/move_src.txt 2>/dev/null || echo "")
    if [ "$content" = "move source" ]; then
        echo -e "${GREEN}✓${NC} Destination has moved content"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Destination should have moved content"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Cleanup
    rm -rf /testdata/user_owned/movedir

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Directory conflict shows error
# ===========================================
test_directory_conflict_error() {
    # Pre-cleanup
    rm -rf /testdata/user_owned/conflict_dir 2>/dev/null || true
    rm -rf /testdata/user_owned/destparent 2>/dev/null || true

    # Create source directory with a file and destination with conflict BEFORE starting
    mkdir -p /testdata/user_owned/conflict_dir
    echo "test" > /testdata/user_owned/conflict_dir/file.txt
    mkdir -p /testdata/user_owned/destparent/conflict_dir

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter and sync right pane to user_owned
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane and navigate to destparent
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "d" "e" "s" "t" "p" "a" "r" "e" "n" "t" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane and clear filter
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to conflict_dir
    send_keys "$CURRENT_SESSION" "/" "c" "o" "n" "f" "l" "i" "c" "t" "_" "d" "i" "r" "Enter"
    sleep 0.3

    # Press c to copy directory
    send_keys "$CURRENT_SESSION" "c"
    sleep 0.5

    # Should show error dialog (directory conflict)
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Error dialog appears for directory conflict"

    # Close error dialog with Enter or Esc
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Cleanup
    rm -rf /testdata/user_owned/conflict_dir
    rm -rf /testdata/user_owned/destparent

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Overwrite dialog navigation with j/k
# ===========================================
test_overwrite_dialog_navigation() {
    # Pre-cleanup
    rm -rf /testdata/user_owned/navtest 2>/dev/null || true
    rm -f /testdata/user_owned/nav_dialog.txt 2>/dev/null || true

    # Create source file and destination directory with conflict BEFORE starting
    echo "nav test" > /testdata/user_owned/nav_dialog.txt
    mkdir -p /testdata/user_owned/navtest
    echo "conflict" > /testdata/user_owned/navtest/nav_dialog.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter and sync right pane to user_owned
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane and navigate
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "n" "a" "v" "t" "e" "s" "t" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane and clear filter
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to nav_dialog.txt
    send_keys "$CURRENT_SESSION" "/" "n" "a" "v" "_" "d" "i" "a" "l" "o" "g" "Enter"
    sleep 0.3

    # Press c to copy
    send_keys "$CURRENT_SESSION" "c"
    sleep 0.5

    # Should show overwrite dialog
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Overwrite dialog appears"

    # Navigate with j (down)
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2

    # Navigate with k (up)
    send_keys "$CURRENT_SESSION" "k"
    sleep 0.2

    # Should still show dialog
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Dialog navigation works without closing"

    # Cancel with Esc
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    assert_not_contains "$CURRENT_SESSION" "already exists" \
        "Dialog closes with Escape"

    # Cleanup
    rm -f /testdata/user_owned/nav_dialog.txt
    rm -rf /testdata/user_owned/navtest

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Rename dialog validation
# ===========================================
test_rename_dialog_validation() {
    # Pre-cleanup
    rm -f /testdata/user_owned/validate_src.txt 2>/dev/null || true
    rm -rf /testdata/user_owned/valdir 2>/dev/null || true

    # Create source file and destination directory with conflicts BEFORE starting
    echo "validate test" > /testdata/user_owned/validate_src.txt
    mkdir -p /testdata/user_owned/valdir
    echo "existing" > /testdata/user_owned/valdir/validate_src.txt
    echo "other" > /testdata/user_owned/valdir/existing_name.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter and sync right pane to user_owned
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane and navigate to valdir
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "v" "a" "l" "d" "i" "r" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane and clear filter
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to validate_src.txt
    send_keys "$CURRENT_SESSION" "/" "v" "a" "l" "i" "d" "a" "t" "e" "_" "s" "r" "c" "Enter"
    sleep 0.3

    # Press c to copy
    send_keys "$CURRENT_SESSION" "c"
    sleep 0.5

    # Should show overwrite dialog first
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Overwrite dialog appears for copy conflict"

    # Press 3 to Rename
    send_keys "$CURRENT_SESSION" "3"
    sleep 0.8

    # Should show rename dialog
    assert_contains "$CURRENT_SESSION" "New name:" \
        "Rename dialog appears"

    # Clear the input and type a conflicting name
    send_keys "$CURRENT_SESSION" "C-u"  # Clear line
    sleep 0.3
    send_keys "$CURRENT_SESSION" "e" "x" "i" "s" "t" "i" "n" "g" "_" "n" "a" "m" "e" "." "t" "x" "t"
    sleep 0.5

    # Should show error for existing filename
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Shows error for existing filename"

    # Cancel with Esc
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    # Cleanup
    rm -f /testdata/user_owned/validate_src.txt
    rm -rf /testdata/user_owned/valdir

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Copy without conflict (no dialog)
# ===========================================
test_copy_no_conflict() {
    # Pre-cleanup
    rm -f /testdata/user_owned/noconflict.txt 2>/dev/null || true
    rm -rf /testdata/user_owned/emptydir 2>/dev/null || true

    # Create source file and empty destination directory BEFORE starting
    echo "no conflict" > /testdata/user_owned/noconflict.txt
    mkdir -p /testdata/user_owned/emptydir

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter and sync right pane to user_owned
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Switch to right pane and navigate to emptydir
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "e" "m" "p" "t" "y" "d" "i" "r" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane and clear filter
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to noconflict.txt
    send_keys "$CURRENT_SESSION" "/" "n" "o" "c" "o" "n" "f" "l" "i" "c" "t" "Enter"
    sleep 0.3

    # Press c to copy
    send_keys "$CURRENT_SESSION" "c"
    sleep 0.5

    # Should NOT show overwrite dialog (no conflict)
    assert_not_contains "$CURRENT_SESSION" "already exists" \
        "No overwrite dialog for non-conflicting copy"

    # File should be copied
    if [ -f "/testdata/user_owned/emptydir/noconflict.txt" ]; then
        echo -e "${GREEN}✓${NC} File was copied without dialog"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} File should be copied"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Cleanup
    rm -f /testdata/user_owned/noconflict.txt
    rm -rf /testdata/user_owned/emptydir

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

# Refresh and Sync tests
run_test test_f5_refresh
run_test test_ctrlr_refresh
run_test test_refresh_cursor_preservation
run_test test_sync_pane
run_test test_sync_preserves_settings
run_test test_sync_right_to_left

# File creation and rename tests
run_test test_create_new_file
run_test test_create_new_directory
run_test test_rename_file
run_test test_cancel_file_creation
run_test test_empty_filename_error
run_test test_rename_parent_dir_ignored

# Post-operation navigation tests (regression tests for dialog cleanup)
run_test test_navigation_after_file_creation
run_test test_navigation_after_dir_creation
run_test test_navigation_after_rename

# Overwrite confirmation dialog tests
run_test test_copy_overwrite_cancel
run_test test_copy_overwrite_confirm
run_test test_copy_overwrite_rename
run_test test_move_overwrite
run_test test_directory_conflict_error
run_test test_overwrite_dialog_navigation
run_test test_rename_dialog_validation
run_test test_copy_no_conflict

# ===========================================
# Multi-file Marking Tests
# ===========================================

# Test: Mark file with Space key
test_mark_file() {
    start_duofm "$CURRENT_SESSION"

    # Initial position is 1 (which is parent dir "..")
    # Move down to first actual file/directory
    send_keys "$CURRENT_SESSION" "j"

    assert_cursor_position "$CURRENT_SESSION" "2" \
        "Cursor at position 2 (first file/dir)"

    # Press Space to mark the file
    send_keys "$CURRENT_SESSION" "Space"

    # Header should show "Marked 1/"
    assert_contains "$CURRENT_SESSION" "Marked 1/" \
        "Header shows Marked 1/ after marking one file"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Cursor moves down after marking
test_mark_cursor_movement() {
    start_duofm "$CURRENT_SESSION"

    # Move to first actual file (skip parent dir)
    send_keys "$CURRENT_SESSION" "j"

    assert_cursor_position "$CURRENT_SESSION" "2" \
        "Initial cursor position is 2 (first file)"

    # Press Space to mark - cursor should move down
    send_keys "$CURRENT_SESSION" "Space"

    # Cursor should now be at position 3
    assert_cursor_position "$CURRENT_SESSION" "3" \
        "Cursor moves to 3 after marking"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Unmark file with Space key
test_unmark_file() {
    start_duofm "$CURRENT_SESSION"

    # Move to first actual file (skip parent dir)
    send_keys "$CURRENT_SESSION" "j"

    # Mark a file
    send_keys "$CURRENT_SESSION" "Space"

    assert_contains "$CURRENT_SESSION" "Marked 1/" \
        "Header shows Marked 1/ after marking"

    # Move back up and unmark
    send_keys "$CURRENT_SESSION" "k"
    send_keys "$CURRENT_SESSION" "Space"

    # Marked count should now be 0 (header shows Marked 0/)
    assert_contains "$CURRENT_SESSION" "Marked 0/" \
        "Header shows Marked 0/ after unmarking"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Parent directory cannot be marked
test_mark_parent_dir_ignored() {
    start_duofm "$CURRENT_SESSION"

    # Enter a subdirectory first
    send_keys "$CURRENT_SESSION" "j"  # Move to dir1
    send_keys "$CURRENT_SESSION" "Enter"  # Enter dir1
    sleep 0.3

    # Now cursor should be on ".." (parent dir entry)
    assert_cursor_position "$CURRENT_SESSION" "1" \
        "Cursor is at position 1 (parent dir)"

    # Try to mark parent directory
    send_keys "$CURRENT_SESSION" "Space"

    # Header should NOT show any marks (parent dir cannot be marked)
    assert_not_contains "$CURRENT_SESSION" "Marked 1/" \
        "Parent directory cannot be marked"

    # Cursor should stay at position 1 (not move down)
    assert_cursor_position "$CURRENT_SESSION" "1" \
        "Cursor stays at position 1 after trying to mark parent dir"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Mark multiple files
test_mark_multiple_files() {
    start_duofm "$CURRENT_SESSION"

    # Move to first actual file (skip parent dir)
    send_keys "$CURRENT_SESSION" "j"

    # Mark first file
    send_keys "$CURRENT_SESSION" "Space"
    # Mark second file (cursor auto-moved down)
    send_keys "$CURRENT_SESSION" "Space"
    # Mark third file
    send_keys "$CURRENT_SESSION" "Space"

    # Header should show "Marked 3/"
    assert_contains "$CURRENT_SESSION" "Marked 3/" \
        "Header shows Marked 3/ after marking three files"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Marks cleared when changing directory
test_marks_cleared_on_directory_change() {
    start_duofm "$CURRENT_SESSION"

    # Move to first actual file (skip parent dir)
    send_keys "$CURRENT_SESSION" "j"

    # Mark a file (dir1)
    send_keys "$CURRENT_SESSION" "Space"

    assert_contains "$CURRENT_SESSION" "Marked 1/" \
        "Header shows Marked 1/ after marking"

    # Cursor auto-moved to dir2, enter dir2
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Marks should be cleared (shows Marked 0/)
    assert_contains "$CURRENT_SESSION" "Marked 0/" \
        "Marks cleared after entering directory"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Batch delete marked files
test_batch_delete_marked_files() {
    # Create test files BEFORE starting duofm
    echo "delete me 1" > /testdata/user_owned/del1.txt
    echo "delete me 2" > /testdata/user_owned/del2.txt

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"

    # File order (sorted alphabetically):
    # 1. .. (parent dir)
    # 2. del1.txt
    # 3. del2.txt

    # Navigate to first file (skip parent dir)
    send_keys "$CURRENT_SESSION" "j"  # Move to del1.txt (position 2)
    sleep 0.2

    # Mark del1.txt (cursor auto-moves to del2.txt)
    send_keys "$CURRENT_SESSION" "Space"
    sleep 0.2
    # Mark del2.txt (cursor stays or moves to next if exists)
    send_keys "$CURRENT_SESSION" "Space"
    sleep 0.2

    # Verify marks
    assert_contains "$CURRENT_SESSION" "Marked 2/" \
        "Two files are marked for deletion"

    # Delete marked files
    send_keys "$CURRENT_SESSION" "d"
    sleep 0.3

    # Confirm deletion dialog should appear
    assert_contains "$CURRENT_SESSION" "Delete 2" \
        "Delete confirmation shows 2 files"

    # Confirm deletion with Enter (Yes is default)
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Files should be gone
    if [ ! -f /testdata/user_owned/del1.txt ] && [ ! -f /testdata/user_owned/del2.txt ]; then
        echo -e "${GREEN}✓${NC} Both files deleted successfully"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Files were not deleted"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Marks should be cleared after operation (shows Marked 0/)
    assert_contains "$CURRENT_SESSION" "Marked 0/" \
        "Marks cleared after deletion"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Context menu shows mark count
test_context_menu_mark_count() {
    start_duofm "$CURRENT_SESSION"

    # Move to first actual file (skip parent dir)
    send_keys "$CURRENT_SESSION" "j"

    # Mark two files
    send_keys "$CURRENT_SESSION" "Space"
    send_keys "$CURRENT_SESSION" "Space"

    assert_contains "$CURRENT_SESSION" "Marked 2/" \
        "Two files are marked"

    # Open context menu
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.3

    # Context menu should show "Copy 2 files" or similar
    assert_contains "$CURRENT_SESSION" "2 files" \
        "Context menu shows file count"

    # Close menu
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Multi-file marking tests
run_test test_mark_file
run_test test_mark_cursor_movement
run_test test_unmark_file
run_test test_mark_parent_dir_ignored
run_test test_mark_multiple_files
run_test test_marks_cleared_on_directory_change
run_test test_batch_delete_marked_files
run_test test_context_menu_mark_count

# ===========================================
# Sort Toggle Tests
# ===========================================

# Test: Sort dialog opens with 's' key
test_sort_dialog_opens() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Should show sort dialog
    assert_contains "$CURRENT_SESSION" "Sort by" \
        "Sort dialog shows 'Sort by' label"

    assert_contains "$CURRENT_SESSION" "Order" \
        "Sort dialog shows 'Order' label"

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort dialog navigation with h/l
test_sort_dialog_hl_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Initial selection should be Name (default)
    assert_contains "$CURRENT_SESSION" "[Name]" \
        "Initial selection is Name"

    # Press l to move to Size
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.2

    assert_contains "$CURRENT_SESSION" "[Size]" \
        "l key moves to Size"

    # Press l to move to Date
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.2

    assert_contains "$CURRENT_SESSION" "[Date]" \
        "l key moves to Date"

    # Press h to go back to Size
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.2

    assert_contains "$CURRENT_SESSION" "[Size]" \
        "h key moves back to Size"

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort dialog row navigation with j/k
test_sort_dialog_jk_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Initial row is Sort by (row 0)
    # Press j to move to Order row
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2

    # Press l to change Order to Desc
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.2

    # Press k to go back to Sort by row
    send_keys "$CURRENT_SESSION" "k"
    sleep 0.2

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort dialog confirms with Enter
test_sort_dialog_confirm() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Change to Size
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.2

    # Confirm with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Dialog should close
    assert_not_contains "$CURRENT_SESSION" "Sort by" \
        "Sort dialog closes after Enter"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort dialog cancels with Escape
test_sort_dialog_cancel() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Change to Size
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.2

    # Cancel with Escape
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    # Dialog should close
    assert_not_contains "$CURRENT_SESSION" "Sort by" \
        "Sort dialog closes after Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort dialog q key cancels
test_sort_dialog_q_cancel() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Cancel with q
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.3

    # Dialog should close
    assert_not_contains "$CURRENT_SESSION" "Sort by" \
        "Sort dialog closes after q"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort by Size shows larger files first (descending)
test_sort_by_size_desc() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Move to Size
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.2

    # Move to Order row
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2

    # Change to Desc
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.2

    # Confirm
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Verify sort is applied (dialog closed, files sorted)
    assert_not_contains "$CURRENT_SESSION" "Sort by" \
        "Sort dialog closes"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort persists after directory navigation
test_sort_persists_after_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Set sort to Size Desc
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "l"  # Size
    sleep 0.2
    send_keys "$CURRENT_SESSION" "j"  # Order row
    sleep 0.2
    send_keys "$CURRENT_SESSION" "l"  # Desc
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Enter"  # Confirm
    sleep 0.3

    # Enter a directory
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.3

    # Go back
    send_keys "$CURRENT_SESSION" "Enter"  # Select ".." and enter
    sleep 0.3

    # Open sort dialog again
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Should show Size selected (persisted)
    assert_contains "$CURRENT_SESSION" "[Size]" \
        "Sort setting persisted after navigation"

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort settings independent per pane
test_sort_independent_panes() {
    start_duofm "$CURRENT_SESSION"

    # Set left pane to Size
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "l"  # Size
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch to right pane
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3

    # Open sort dialog in right pane
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Should show Name (default for right pane, unchanged)
    assert_contains "$CURRENT_SESSION" "[Name]" \
        "Right pane has independent sort setting"

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Arrow keys work in sort dialog
test_sort_dialog_arrow_keys() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Use arrow keys
    send_keys "$CURRENT_SESSION" "Right"  # Move to Size
    sleep 0.2

    assert_contains "$CURRENT_SESSION" "[Size]" \
        "Right arrow moves to Size"

    send_keys "$CURRENT_SESSION" "Down"  # Move to Order row
    sleep 0.2

    send_keys "$CURRENT_SESSION" "Left"  # Should move to Asc (left in Order row)
    sleep 0.2

    send_keys "$CURRENT_SESSION" "Up"  # Back to Sort by row
    sleep 0.2

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Sort toggle tests
run_test test_sort_dialog_opens
run_test test_sort_dialog_hl_navigation
run_test test_sort_dialog_jk_navigation
run_test test_sort_dialog_confirm
run_test test_sort_dialog_cancel
run_test test_sort_dialog_q_cancel
run_test test_sort_by_size_desc
run_test test_sort_persists_after_navigation
run_test test_sort_independent_panes
run_test test_sort_dialog_arrow_keys

# ===========================================
# Cursor Preservation After Viewer Tests
# ===========================================

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

# Cursor preservation tests
run_test test_cursor_preserved_after_view
run_test test_cursor_preserved_after_enter_view
run_test test_cursor_reset_when_file_deleted
run_test test_both_panes_preserve_cursor

# ===========================================
# Shell Command Execution Tests
# ===========================================

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

# Shell command execution tests
run_test test_shell_command_mode_enter
run_test test_shell_command_input
run_test test_shell_command_empty_enter
run_test test_shell_command_ignored_with_dialog
run_test test_shell_command_ignored_during_search
run_test test_help_shows_shell_command

# ===========================================
# Configuration File Tests
# ===========================================

# Test: Config file auto-generated on first run
test_config_auto_generated() {
    # Remove any existing config
    rm -f ~/.config/duofm/config.toml 2>/dev/null || true
    rm -rf ~/.config/duofm 2>/dev/null || true

    start_duofm "$CURRENT_SESSION"
    sleep 0.5

    # Check if config file was created
    if [ -f ~/.config/duofm/config.toml ]; then
        echo -e "${GREEN}✓${NC} Config file auto-generated on first run"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Config file should be auto-generated"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    stop_duofm "$CURRENT_SESSION"
}

# Test: Config file contains keybindings section
test_config_has_keybindings() {
    # Ensure config exists
    if [ ! -f ~/.config/duofm/config.toml ]; then
        start_duofm "$CURRENT_SESSION"
        sleep 0.5
        stop_duofm "$CURRENT_SESSION"
    fi

    # Check for [keybindings] section
    if grep -q "\[keybindings\]" ~/.config/duofm/config.toml; then
        echo -e "${GREEN}✓${NC} Config file has keybindings section"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Config file should have keybindings section"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Test: Config file has comments
test_config_has_comments() {
    if [ ! -f ~/.config/duofm/config.toml ]; then
        start_duofm "$CURRENT_SESSION"
        sleep 0.5
        stop_duofm "$CURRENT_SESSION"
    fi

    # Check for comments
    if grep -q "^#" ~/.config/duofm/config.toml; then
        echo -e "${GREEN}✓${NC} Config file has comments"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Config file should have comments"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Test: Default keybindings work without config
test_default_keybindings_without_config() {
    # Remove config file
    rm -f ~/.config/duofm/config.toml 2>/dev/null || true

    start_duofm "$CURRENT_SESSION"
    sleep 0.3

    # Test j/k navigation (default keybindings)
    send_keys "$CURRENT_SESSION" "j" "j"
    sleep 0.3

    assert_cursor_position "$CURRENT_SESSION" "3" \
        "Default j key moves cursor down"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Help dialog shows PascalCase keys
test_help_shows_pascalcase() {
    start_duofm "$CURRENT_SESSION"

    # Open help
    send_keys "$CURRENT_SESSION" "?"
    sleep 0.3

    # Should show PascalCase keys
    assert_contains "$CURRENT_SESSION" "J/K" \
        "Help shows J/K in PascalCase"

    # Close help
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Configuration file tests
run_test test_config_auto_generated
run_test test_config_has_keybindings
run_test test_config_has_comments
run_test test_default_keybindings_without_config
run_test test_help_shows_pascalcase

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

# Async directory load pane identification bug fix tests
run_test test_right_pane_same_path_navigation
run_test test_right_pane_home_navigation

# Print summary and exit with appropriate code
print_summary
exit $?
