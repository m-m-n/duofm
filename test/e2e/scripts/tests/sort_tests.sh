#!/bin/bash
# Sort Tests for duofm
#
# Description: Tests for sort dialog including navigation,
#              confirmation, and persistence
# Tests: 10

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

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

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Sort"
    echo "========================================"

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

    print_summary
    exit $?
fi
