#!/bin/bash
# History Navigation Tests for duofm
#
# Description: Tests for directory history navigation using [ and ] keys
# Tests: 8

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# ===========================================
# Test: Basic history navigation (back)
# ===========================================
test_history_back() {
    start_duofm "$CURRENT_SESSION"

    # Enter a subdirectory
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Capture current directory (should be in subdirectory)
    local screen_in_subdir
    screen_in_subdir=$(capture_screen "$CURRENT_SESSION")

    # Go back in history using [ key
    send_keys "$CURRENT_SESSION" "["
    sleep 0.5

    # Should be back in original directory (testdata)
    assert_contains "$CURRENT_SESSION" "testdata" \
        "[ key navigates back in history"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: History forward navigation
# ===========================================
test_history_forward() {
    start_duofm "$CURRENT_SESSION"

    # Enter a subdirectory
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Go back in history
    send_keys "$CURRENT_SESSION" "["
    sleep 0.5

    # Go forward in history using ] key
    send_keys "$CURRENT_SESSION" "]"
    sleep 0.5

    # Should be back in subdirectory (dir1)
    assert_contains "$CURRENT_SESSION" "subdir" \
        "] key navigates forward in history (back to dir1 with subdir)"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Multiple level navigation
# ===========================================
test_history_multiple_levels() {
    start_duofm "$CURRENT_SESSION"

    # Enter first subdirectory (dir1)
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Enter second subdirectory (subdir)
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Go back twice in history
    send_keys "$CURRENT_SESSION" "["
    sleep 0.5
    send_keys "$CURRENT_SESSION" "["
    sleep 0.5

    # Should be back in original directory (testdata)
    assert_contains "$CURRENT_SESSION" "testdata" \
        "[ key navigates back through multiple history levels"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: History navigation with no history
# ===========================================
test_history_no_history() {
    start_duofm "$CURRENT_SESSION"

    # Try to go back/forward with no history (should not crash)
    send_keys "$CURRENT_SESSION" "[" "]"
    sleep 0.3

    # Application should still be running and show testdata
    assert_contains "$CURRENT_SESSION" "testdata" \
        "History navigation with no history does not crash"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Independence from - key (previousPath)
# ===========================================
test_history_independent_from_previous() {
    start_duofm "$CURRENT_SESSION"

    # Enter first subdirectory (dir1)
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Enter second subdirectory (dir1/subdir)
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Use - key to toggle back to dir1
    send_keys "$CURRENT_SESSION" "-"
    sleep 0.5

    # Use - key again to toggle back to subdir
    send_keys "$CURRENT_SESSION" "-"
    sleep 0.5

    # Now use [ to go back in history
    send_keys "$CURRENT_SESSION" "["
    sleep 0.5

    # Should be able to navigate back in history
    # The - key toggles between last two, but [ navigates through full history
    local screen
    screen=$(capture_screen "$CURRENT_SESSION")

    if echo "$screen" | grep -qE "(dir1|testdata)"; then
        echo -e "${GREEN}✓${NC} History navigation independent from - key"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} History navigation should be independent from - key"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Forward history cleared on new navigation
# ===========================================
test_history_forward_cleared() {
    start_duofm "$CURRENT_SESSION"

    # Enter first subdirectory (dir1)
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Enter second subdirectory (subdir)
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Go back in history to dir1
    send_keys "$CURRENT_SESSION" "["
    sleep 0.5

    # Navigate to a different directory (dir2 via parent)
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "j" "j" "Enter"
    sleep 0.5

    # Try to go forward - should have no effect (forward history cleared)
    send_keys "$CURRENT_SESSION" "]"
    sleep 0.3

    # Should still be in dir2 (forward history was cleared)
    assert_contains "$CURRENT_SESSION" "another.txt" \
        "Forward history cleared after new navigation"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Parent directory navigation recorded in history
# ===========================================
test_history_parent_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Enter subdirectory
    send_keys "$CURRENT_SESSION" "j" "Enter"
    sleep 0.5

    # Go back to parent using h key
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.5

    # Go back in history - should return to subdirectory
    send_keys "$CURRENT_SESSION" "["
    sleep 0.5

    # Should be in dir1 (the subdirectory we were in before h)
    assert_contains "$CURRENT_SESSION" "subdir" \
        "Parent navigation (h key) recorded in history"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Home directory navigation recorded in history
# ===========================================
test_history_home_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Navigate to home directory
    send_keys "$CURRENT_SESSION" "~"
    sleep 0.5

    # Go back in history - should return to testdata
    send_keys "$CURRENT_SESSION" "["
    sleep 0.5

    # Should be back in testdata
    assert_contains "$CURRENT_SESSION" "testdata" \
        "Home navigation (~) recorded in history"

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - History Navigation"
    echo "========================================"

    run_test test_history_back
    run_test test_history_forward
    run_test test_history_multiple_levels
    run_test test_history_no_history
    run_test test_history_independent_from_previous
    run_test test_history_forward_cleared
    run_test test_history_parent_navigation
    run_test test_history_home_navigation

    print_summary
    exit $?
fi
