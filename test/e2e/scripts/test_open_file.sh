#!/bin/bash
# E2E Test: Open file with external application
# Tests for v, e, and Enter key functionality

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"

echo "=== E2E Tests: Open File with External Application ==="
echo ""

# Test 1: View file with v key
test_view_file_with_v() {
    local session="$CURRENT_SESSION"

    start_duofm "$session"

    # Navigate down to file1.txt
    # Order: .. -> dir1 -> dir2 -> empty_dir -> no_access -> root_owned -> user_owned -> file1.txt
    # So we need 7 j presses to reach file1.txt
    send_keys "$session" "j" "j" "j" "j" "j" "j" "j"
    sleep 0.2

    # Press v to view - less should open and show the file content
    send_keys "$session" "v"
    sleep 0.5

    # Verify less is showing the file content
    assert_contains "$session" "file1 content" "less should display file1 content"

    # Exit less with 'q'
    send_keys "$session" "q"
    sleep 0.5

    # After less exits, screen should show duofm again
    assert_contains "$session" "duofm" "duofm title bar should be visible after less exits"

    stop_duofm "$session"
}

# Test 2: v key on directory should do nothing
test_view_directory_ignored() {
    local session="$CURRENT_SESSION"

    start_duofm "$session"

    # Navigate to a directory (dir1 - first after ..)
    send_keys "$session" "j"  # Move to dir1
    sleep 0.2

    # Press v - should be ignored for directories
    send_keys "$session" "v"
    sleep 0.3

    # Screen should still show duofm (no external app opened)
    assert_contains "$session" "duofm" "duofm title should remain visible (v on dir ignored)"

    stop_duofm "$session"
}

# Test 3: Enter on directory should enter it (existing behavior)
test_enter_directory() {
    local session="$CURRENT_SESSION"

    start_duofm "$session"

    # Navigate to dir1 directory (first directory after ..)
    send_keys "$session" "j"  # Move to dir1
    sleep 0.2

    # Press Enter to enter directory
    send_keys "$session" "Enter"
    sleep 0.5

    # Should now be in dir1 (path should show dir1 and contain subdir)
    assert_contains "$session" "subdir" "Should see subdir inside dir1 directory"

    stop_duofm "$session"
}

# Test 4: Enter on parent (..) should navigate to parent
test_enter_parent_dir() {
    local session="$CURRENT_SESSION"

    # Start in dir1 subdirectory
    start_duofm "$session" "/testdata/dir1"

    # First entry should be ..
    # Press Enter to go to parent
    send_keys "$session" "Enter"
    sleep 0.5

    # Should be back in /testdata
    assert_contains "$session" "file1.txt" "Should see file1.txt after navigating to parent"

    stop_duofm "$session"
}

# Test 5: v key on parent dir (..) should be ignored
test_view_parent_dir_ignored() {
    local session="$CURRENT_SESSION"

    start_duofm "$session"

    # First entry is .. (parent dir)
    # Press v - should be ignored
    send_keys "$session" "v"
    sleep 0.3

    # Should still show duofm, not less
    assert_contains "$session" "duofm" "duofm title should remain visible (v on .. ignored)"

    stop_duofm "$session"
}

# Test 6: e key on parent dir (..) should be ignored
test_edit_parent_dir_ignored() {
    local session="$CURRENT_SESSION"

    start_duofm "$session"

    # First entry is .. (parent dir)
    # Press e - should be ignored
    send_keys "$session" "e"
    sleep 0.3

    # Should still show duofm, not vim
    assert_contains "$session" "duofm" "duofm title should remain visible (e on .. ignored)"

    stop_duofm "$session"
}

# Test 7: e key on directory should be ignored
test_edit_directory_ignored() {
    local session="$CURRENT_SESSION"

    start_duofm "$session"

    # Navigate to dir1 (first directory after ..)
    send_keys "$session" "j"  # Move to dir1
    sleep 0.2

    # Press e - should be ignored for directories
    send_keys "$session" "e"
    sleep 0.3

    # Should still show duofm
    assert_contains "$session" "duofm" "duofm title should remain visible (e on dir ignored)"

    stop_duofm "$session"
}

# Test 8: Edit file with e key
test_edit_file_with_e() {
    local session="$CURRENT_SESSION"

    start_duofm "$session"

    # Navigate to file1.txt (7 j presses to reach the first file)
    send_keys "$session" "j" "j" "j" "j" "j" "j" "j"
    sleep 0.2

    # Press e to edit - vim should open and show the file content
    send_keys "$session" "e"
    sleep 0.5

    # Verify vim is showing the file content
    assert_contains "$session" "file1 content" "vim should display file1 content"

    # Exit vim with :q
    send_keys "$session" "Escape" ":" "q" "Enter"
    sleep 0.5

    # After vim exits, duofm should restore
    assert_contains "$session" "duofm" "duofm title bar should be visible after vim exits"

    stop_duofm "$session"
}

# Test 9: Enter on file should open viewer (same as v key)
test_enter_file_opens_viewer() {
    local session="$CURRENT_SESSION"

    start_duofm "$session"

    # Navigate to file1.txt (7 j presses to reach the first file)
    send_keys "$session" "j" "j" "j" "j" "j" "j" "j"
    sleep 0.2

    # Press Enter on file - should open less
    send_keys "$session" "Enter"
    sleep 0.5

    # Verify less is showing the file content
    assert_contains "$session" "file1 content" "less should display file1 content via Enter"

    # Exit less with 'q'
    send_keys "$session" "q"
    sleep 0.5

    # After less exits, duofm should restore
    assert_contains "$session" "duofm" "duofm title bar should be visible after Enter on file"

    stop_duofm "$session"
}

# Run all tests
run_test test_view_parent_dir_ignored
run_test test_edit_parent_dir_ignored
run_test test_view_directory_ignored
run_test test_edit_directory_ignored
run_test test_enter_directory
run_test test_enter_parent_dir
run_test test_view_file_with_v
run_test test_edit_file_with_e
run_test test_enter_file_opens_viewer

# Print summary
print_summary
exit $?
