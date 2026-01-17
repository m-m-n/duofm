#!/bin/bash
# Sort Tests for duofm
#
# Description: Tests for sort dialog with dropdown menus including navigation,
#              selection, confirmation, and persistence
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

# Test: Sort dialog Tab navigation between dropdowns
test_sort_dialog_tab_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Initial focus should be on Sort by dropdown
    # Press Tab to move to Order dropdown
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.2

    # Press Shift+Tab to go back to Sort by
    send_keys "$CURRENT_SESSION" "S-Tab"
    sleep 0.2

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort dialog dropdown expansion with Enter
test_sort_dialog_dropdown_expansion() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Press Enter to expand Sort by dropdown
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.2

    # Should show dropdown options (Name, Size, Date)
    assert_contains "$CURRENT_SESSION" "Name" \
        "Dropdown shows Name option"

    # Press Escape to close dropdown
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort dialog dropdown navigation with j/k
test_sort_dialog_dropdown_jk_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Press Enter to expand dropdown
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.2

    # Navigate down with j
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2

    # Navigate down again to Date
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2

    # Navigate up with k
    send_keys "$CURRENT_SESSION" "k"
    sleep 0.2

    # Select Size with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.2

    # Verify Size is selected
    assert_contains "$CURRENT_SESSION" "Size" \
        "Size option is selected"

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort dialog cancels with Escape
test_sort_dialog_cancel() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Expand dropdown and select Size
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "j"  # Size
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Enter"  # Select
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

# Test: Sort dialog q cancels even when dropdown is expanded
test_sort_dialog_q_cancel_with_dropdown() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Expand dropdown
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.2

    # Cancel with q (should close entire dialog, not just dropdown)
    send_keys "$CURRENT_SESSION" "q"
    sleep 0.3

    # Dialog should close
    assert_not_contains "$CURRENT_SESSION" "Sort by" \
        "Sort dialog closes after q even with dropdown expanded"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort by Size with Desc order
test_sort_by_size_desc() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Expand Sort by dropdown and select Size
    send_keys "$CURRENT_SESSION" "Enter"  # Expand
    sleep 0.2
    send_keys "$CURRENT_SESSION" "j"  # Move to Size
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Enter"  # Select Size
    sleep 0.2

    # Tab to Order dropdown
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.2

    # Expand Order dropdown and select Desc
    send_keys "$CURRENT_SESSION" "Enter"  # Expand
    sleep 0.2
    send_keys "$CURRENT_SESSION" "j"  # Move to Desc
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Enter"  # Select Desc
    sleep 0.2

    # Close dialog (Escape cancels, so we need to use q to apply changes)
    # Actually, selection changes are applied live, Escape just closes
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    # Verify sort is applied (dialog closed)
    assert_not_contains "$CURRENT_SESSION" "Sort by" \
        "Sort dialog closes"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort persists after directory navigation
test_sort_persists_after_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Set sort to Size via dropdown
    send_keys "$CURRENT_SESSION" "s"  # Open dialog
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"  # Expand Sort by
    sleep 0.2
    send_keys "$CURRENT_SESSION" "j"  # Move to Size
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Enter"  # Select Size
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Escape"  # Close dialog
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
    assert_contains "$CURRENT_SESSION" "Size" \
        "Sort setting persisted after navigation"

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Sort settings independent per pane
test_sort_independent_panes() {
    start_duofm "$CURRENT_SESSION"

    # Set left pane to Size via dropdown
    send_keys "$CURRENT_SESSION" "s"  # Open dialog
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"  # Expand Sort by
    sleep 0.2
    send_keys "$CURRENT_SESSION" "j"  # Move to Size
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Enter"  # Select Size
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Escape"  # Close dialog
    sleep 0.3

    # Switch to right pane (l switches panes in file list mode)
    send_keys "$CURRENT_SESSION" "Tab"
    sleep 0.3

    # Open sort dialog in right pane
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Should show Name (default for right pane, unchanged)
    assert_contains "$CURRENT_SESSION" "Name" \
        "Right pane has independent sort setting"

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Arrow keys work in dropdown
test_sort_dialog_arrow_keys() {
    start_duofm "$CURRENT_SESSION"

    # Press 's' to open sort dialog
    send_keys "$CURRENT_SESSION" "s"
    sleep 0.3

    # Expand dropdown
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.2

    # Use Down arrow to navigate
    send_keys "$CURRENT_SESSION" "Down"  # Move to Size
    sleep 0.2

    send_keys "$CURRENT_SESSION" "Down"  # Move to Date
    sleep 0.2

    # Use Up arrow to navigate back
    send_keys "$CURRENT_SESSION" "Up"  # Back to Size
    sleep 0.2

    # Select with Enter
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.2

    assert_contains "$CURRENT_SESSION" "Size" \
        "Arrow keys work for navigation"

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
    run_test test_sort_dialog_tab_navigation
    run_test test_sort_dialog_dropdown_expansion
    run_test test_sort_dialog_dropdown_jk_navigation
    run_test test_sort_dialog_cancel
    run_test test_sort_dialog_q_cancel
    run_test test_sort_dialog_q_cancel_with_dropdown
    run_test test_sort_by_size_desc
    run_test test_sort_persists_after_navigation
    run_test test_sort_independent_panes
    run_test test_sort_dialog_arrow_keys

    print_summary
    exit $?
fi
