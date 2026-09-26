#!/bin/bash
# File Operation Tests for duofm
#
# Description: Tests for file operations including create, delete, rename,
#              and their interactions with navigation
# Tests: 13

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# ===========================================
# Helper: Check the extension-preserving rename dialog's input field for an
# exact base-name value.
#
# A bare substring check (assert_contains "$base_name") also matches the
# file-list row behind the dialog (e.g. "before_rename.txt") and the
# search-filter display (e.g. "/before_ren"), so it passes even when the
# input field itself is empty or holds the wrong value. This helper instead
# requires the match to occur on a single captured row, in this order: a
# vertical border character, only whitespace, the base name, only
# whitespace, another vertical border character, whitespace, and the fixed
# extension. No non-whitespace character next to the base name is
# tolerated, including a single one on either side: the cursor is drawn by
# reverse video only, and the capture drops attributes. That structure is
# unique to the input field's own row and cannot be produced by the
# file-list row, the title row, or the search-filter display.
#
# Usage: assert_input_field_value <session_name> <expected_base_name> <extension> <description>
# ===========================================
assert_input_field_value() {
    local session_name="$1"
    local expected_base_name="$2"
    local extension="$3"
    local description="$4"

    TESTS_RUN=$((TESTS_RUN + 1))

    local screen
    screen=$(capture_screen "$session_name")

    # Escape ERE metacharacters so the base name and extension are matched
    # as literal text (in particular, the "." in the extension must not
    # match an arbitrary character).
    local escaped_base escaped_ext
    escaped_base=$(printf '%s' "$expected_base_name" | sed -e 's/[][\.^$*+?(){}|]/\\&/g')
    escaped_ext=$(printf '%s' "$extension" | sed -e 's/[][\.^$*+?(){}|]/\\&/g')

    # │ = U+2502 BOX DRAWINGS LIGHT VERTICAL: the border character drawn by
    # lipgloss.RoundedBorder() around the input field.
    local border='│'
    local pattern="${border}[[:space:]]*${escaped_base}[[:space:]]*${border}[[:space:]]*${escaped_ext}"

    # Force a UTF-8 locale for this match only (independent of the E2E
    # container's ambient locale, which is POSIX/C): the border character
    # is multi-byte and is matched as literal text here, with no "any
    # single character" element in the pattern. grep without -z already
    # evaluates each line independently, so a match never spans two
    # captured rows.
    if echo "$screen" | LC_ALL=C.utf8 grep -qE -- "$pattern"; then
        echo -e "${GREEN}✓${NC} $description"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $description"
        echo -e "  ${YELLOW}Expected input field base name:${NC} $expected_base_name"
        echo -e "  ${YELLOW}Expected extension:${NC} $extension"
        echo -e "  ${YELLOW}Screen content:${NC}"
        echo "$screen" | sed 's/^/    /'
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
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
    # Remove leftover fixture names before starting
    rm -f /testdata/user_owned/before_rename.txt /testdata/user_owned/after_rename.txt

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

    # Navigate to the file by name
    send_keys "$CURRENT_SESSION" "/" "b" "e" "f" "o" "r" "e" "_" "r" "e" "n" "Enter"
    sleep 0.3

    # Press r to rename
    send_keys "$CURRENT_SESSION" "r"
    sleep 0.3

    # Should show the extension-preserving rename dialog, pre-filled with
    # the base name only
    assert_contains "$CURRENT_SESSION" "Rename (extension: .txt):" \
        "Extension-preserving rename dialog appears with extension in title"

    # Limited to the input field row: a bare substring check would also
    # match "before_rename.txt" in the file list behind the dialog.
    assert_input_field_value "$CURRENT_SESSION" "before_rename" ".txt" \
        "Rename dialog is pre-filled with the base name"

    # Clear the input and type only the new base name (extension is fixed)
    send_keys "$CURRENT_SESSION" "C-u"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "a" "f" "t" "e" "r" "_" "r" "e" "n" "a" "m" "e"
    sleep 0.2

    # Confirm with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Clear the search filter (still "before_ren") so the full listing shows:
    # reopen search with "/" and confirm an empty pattern with Enter
    send_keys "$CURRENT_SESSION" "/" "Enter"
    sleep 0.3

    # Old name should be gone, new name should appear
    assert_not_contains "$CURRENT_SESSION" "before_rename.txt" \
        "Old filename is gone"

    assert_contains "$CURRENT_SESSION" "after_rename.txt" \
        "New filename appears in listing"

    # On disk, the target exists and the source does not
    if [ -f /testdata/user_owned/after_rename.txt ] && [ ! -f /testdata/user_owned/before_rename.txt ]; then
        echo -e "${GREEN}✓${NC} after_rename.txt exists on disk and before_rename.txt does not"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Rename did not update the file on disk as expected"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    stop_duofm "$CURRENT_SESSION"

    # Cleanup regardless of outcome
    rm -f /testdata/user_owned/before_rename.txt /testdata/user_owned/after_rename.txt
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
    # Remove leftover fixture names before starting
    rm -f /testdata/user_owned/navren_before.txt /testdata/user_owned/navren_after.txt

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

    # Navigate to the file by name
    send_keys "$CURRENT_SESSION" "/" "n" "a" "v" "r" "e" "n" "_" "b" "e" "f" "o" "r" "e" "Enter"
    sleep 0.3

    # Press r to rename
    send_keys "$CURRENT_SESSION" "r"
    sleep 0.3

    # Should show the extension-preserving rename dialog, pre-filled with
    # the base name only
    assert_contains "$CURRENT_SESSION" "Rename (extension: .txt):" \
        "Extension-preserving rename dialog appears with extension in title"

    # Limited to the input field row: a bare substring check would also
    # match "navren_before.txt" in the file list and "/navren_before" in
    # the search-filter display, both still on screen behind the dialog.
    assert_input_field_value "$CURRENT_SESSION" "navren_before" ".txt" \
        "Rename dialog is pre-filled with the base name"

    # Clear the input and type only the new base name (extension is fixed)
    send_keys "$CURRENT_SESSION" "C-u"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "n" "a" "v" "r" "e" "n" "_" "a" "f" "t" "e" "r"
    sleep 0.2

    # Confirm with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # On disk, the target exists and the source does not
    if [ -f /testdata/user_owned/navren_after.txt ] && [ ! -f /testdata/user_owned/navren_before.txt ]; then
        echo -e "${GREEN}✓${NC} navren_after.txt exists on disk and navren_before.txt does not"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Rename did not update the file on disk as expected"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Test navigation still works - try to quit
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.5

    # Cleanup
    rm -f /testdata/user_owned/navren_after.txt /testdata/user_owned/navren_before.txt

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
# Test: Delete confirmation - Enter key is ignored
# ===========================================
test_delete_confirmation_enter_ignored() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Create a test file
    touch /testdata/user_owned/test_enter_delete.txt
    send_keys "$CURRENT_SESSION" "F5"
    sleep 0.5

    # Navigate to the file
    send_keys "$CURRENT_SESSION" "/" "t" "e" "s" "t" "_" "e" "n" "t" "e" "r" "_" "d" "e" "l" "e" "t" "e" "Enter"
    sleep 0.3

    # Press d to delete
    send_keys "$CURRENT_SESSION" "d"
    sleep 0.3

    # Confirm dialog appears
    assert_contains "$CURRENT_SESSION" "Delete" \
        "Delete confirmation dialog appears"

    # Press Enter (should be ignored, dialog stays open)
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # File should still exist and dialog should still be visible
    assert_contains "$CURRENT_SESSION" "Delete" \
        "Delete dialog still open after Enter key"

    if [ -f /testdata/user_owned/test_enter_delete.txt ]; then
        echo -e "${GREEN}✓${NC} Delete confirmation - Enter key is ignored (file still exists)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Enter key should not delete file"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Cancel the dialog with n
    send_keys "$CURRENT_SESSION" "n"
    sleep 0.3

    # Cleanup
    rm -f /testdata/user_owned/test_enter_delete.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Delete confirmation - Y key confirms deletion
# ===========================================
test_delete_confirmation_y_key_works() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Create a test file
    touch /testdata/user_owned/test_y_delete.txt
    send_keys "$CURRENT_SESSION" "F5"
    sleep 0.5

    # Navigate to the file
    send_keys "$CURRENT_SESSION" "/" "t" "e" "s" "t" "_" "y" "_" "d" "e" "l" "e" "t" "e" "Enter"
    sleep 0.3

    # Press d to delete
    send_keys "$CURRENT_SESSION" "d"
    sleep 0.3

    # Confirm deletion with y
    send_keys "$CURRENT_SESSION" "y"
    sleep 0.5

    # File should be deleted
    if [ ! -f /testdata/user_owned/test_y_delete.txt ]; then
        echo -e "${GREEN}✓${NC} Delete confirmation - Y key confirms deletion (file deleted)"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Y key should delete file"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    stop_duofm "$CURRENT_SESSION"

    # Cleanup regardless of outcome
    rm -f /testdata/user_owned/test_y_delete.txt
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
    run_test test_delete_confirmation_enter_ignored
    run_test test_delete_confirmation_y_key_works

    print_summary
    exit $?
fi
