#!/bin/bash
# Copy/Move Tests for duofm
#
# Description: Tests for copy and move operations including overwrite dialogs,
#              conflict handling, and dialog navigation
# Tests: 8

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

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

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Copy/Move"
    echo "========================================"

    run_test test_copy_overwrite_cancel
    run_test test_copy_overwrite_confirm
    run_test test_copy_overwrite_rename
    run_test test_move_overwrite
    run_test test_directory_conflict_error
    run_test test_overwrite_dialog_navigation
    run_test test_rename_dialog_validation
    run_test test_copy_no_conflict

    print_summary
    exit $?
fi
