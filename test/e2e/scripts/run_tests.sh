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

# Print summary and exit with appropriate code
print_summary
exit $?
