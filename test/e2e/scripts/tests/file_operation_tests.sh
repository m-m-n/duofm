#!/bin/bash
# File Operation Tests for duofm
#
# Description: Tests for file operations including create, delete, rename,
#              and their interactions with navigation
# Tests: 11

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

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

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - File Operations"
    echo "========================================"

    run_test test_cannot_delete_root_file
    run_test test_can_delete_user_file
    run_test test_create_new_file
    run_test test_create_new_directory
    run_test test_rename_file
    run_test test_cancel_file_creation
    run_test test_empty_filename_error
    run_test test_navigation_after_file_creation
    run_test test_navigation_after_dir_creation
    run_test test_navigation_after_rename
    run_test test_rename_parent_dir_ignored

    print_summary
    exit $?
fi
