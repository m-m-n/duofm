#!/bin/bash
# Mark Tests for duofm
#
# Description: Tests for multi-file marking functionality including
#              marking, unmarking, batch operations, and context menu
# Tests: 8

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

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

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Mark"
    echo "========================================"

    run_test test_mark_file
    run_test test_mark_cursor_movement
    run_test test_unmark_file
    run_test test_mark_parent_dir_ignored
    run_test test_mark_multiple_files
    run_test test_marks_cleared_on_directory_change
    run_test test_batch_delete_marked_files
    run_test test_context_menu_mark_count

    print_summary
    exit $?
fi
