#!/bin/bash
# Bookmark Tests for duofm
#
# Description: Tests for bookmark functionality including
#              dialog display, adding bookmarks, and empty state
# Tests: 3

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# Test: Bookmark Dialog Opens (B key)
test_bookmark_dialog_opens() {
    start_duofm "$CURRENT_SESSION"

    # Open bookmark manager with B key
    send_keys "$CURRENT_SESSION" "b"
    sleep 0.3

    # Should show Bookmarks dialog
    assert_contains "$CURRENT_SESSION" "Bookmarks" \
        "Bookmark dialog title is shown"

    # Should show key hints
    assert_contains "$CURRENT_SESSION" "Enter:Jump" \
        "Bookmark dialog shows Enter:Jump hint"

    assert_contains "$CURRENT_SESSION" "D:Delete" \
        "Bookmark dialog shows D:Delete hint"

    assert_contains "$CURRENT_SESSION" "Esc:Close" \
        "Bookmark dialog shows Esc:Close hint"

    # Close with Esc
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    # Dialog should be closed, file list should be visible
    assert_contains "$CURRENT_SESSION" "file1.txt" \
        "File list is visible after closing bookmark dialog"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Add Bookmark Dialog (Shift+B)
test_add_bookmark_dialog() {
    start_duofm "$CURRENT_SESSION"

    # Open add bookmark dialog with Shift+B
    send_keys "$CURRENT_SESSION" "B"
    sleep 0.3

    # Should show input dialog with "Bookmark name:" or similar
    assert_contains "$CURRENT_SESSION" "Bookmark" \
        "Add bookmark dialog is shown"

    # Cancel with Esc
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    # Dialog should be closed
    assert_contains "$CURRENT_SESSION" "file1.txt" \
        "File list is visible after canceling add bookmark"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Bookmark Empty State
test_bookmark_empty_state() {
    start_duofm "$CURRENT_SESSION"

    # Open bookmark manager (should be empty initially)
    send_keys "$CURRENT_SESSION" "b"
    sleep 0.3

    # Should show empty message
    assert_contains "$CURRENT_SESSION" "No bookmarks" \
        "Empty bookmark list shows 'No bookmarks'"

    # Close with Esc
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Bookmark"
    echo "========================================"

    run_test test_bookmark_dialog_opens
    run_test test_add_bookmark_dialog
    run_test test_bookmark_empty_state

    print_summary
    exit $?
fi
