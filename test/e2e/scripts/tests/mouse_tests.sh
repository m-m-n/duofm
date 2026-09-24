#!/bin/bash
# Mouse Tests for duofm
#
# Description: Tests for left-button mouse interaction: click moves the
#              cursor, drag marks a contiguous range, and double-click
#              enters a directory. Mouse input is injected as SGR-encoded
#              escape sequences sent as literal terminal input.
# Tests: 3

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# ===========================================
# Mouse sequence helpers (SGR encoding)
#
# All coordinates are 1-based column;row, matching the terminal wire
# format. These helpers are local to this test file; helpers.sh stays
# unchanged.
# ===========================================

# Build an SGR left-button press sequence.
# Usage: mouse_press_seq <col> <row>
mouse_press_seq() {
    printf '\x1b[<0;%d;%dM' "$1" "$2"
}

# Build an SGR motion-with-left-button-held sequence.
# Usage: mouse_motion_seq <col> <row>
mouse_motion_seq() {
    printf '\x1b[<32;%d;%dM' "$1" "$2"
}

# Build an SGR left-button release sequence.
# Usage: mouse_release_seq <col> <row>
mouse_release_seq() {
    printf '\x1b[<0;%d;%dm' "$1" "$2"
}

# Send one or more raw SGR mouse sequences to the session as literal input
# (tmux's literal send mode, so the escape byte is not reinterpreted as a
# key name).
# Usage: send_mouse_sequences <session_name> <sequence...>
send_mouse_sequences() {
    local session_name="${SESSION_PREFIX}_$1"
    shift

    for seq in "$@"; do
        tmux send-keys -t "$session_name" -l "$seq"
        sleep 0.1
    done

    # Wait for UI to update
    sleep 0.3
}

# Find the 1-based screen row of the line containing an entry name.
# Usage: find_entry_row <screen_text> <entry_name>
find_entry_row() {
    local screen="$1"
    local name="$2"
    echo "$screen" | grep -n -F -- "$name" | head -1 | cut -d: -f1
}

# A column safely inside the left pane (which spans columns 1-60).
LEFT_PANE_COL=10

# ===========================================
# Test: Mouse click moves cursor (TS-20)
# ===========================================
test_mouse_click_moves_cursor() {
    start_duofm "$CURRENT_SESSION"

    # With scroll offset 0, the third entry (cursor position 3) is always
    # on row 7: row 1 title, rows 2-4 pane headers, entry index i on row
    # 5 + i, so index 2 (the third entry, position 3) is on row 5 + 2 = 7.
    local target_row=7

    send_mouse_sequences "$CURRENT_SESSION" \
        "$(mouse_press_seq "$LEFT_PANE_COL" "$target_row")" \
        "$(mouse_release_seq "$LEFT_PANE_COL" "$target_row")"

    assert_cursor_position "$CURRENT_SESSION" "3" \
        "Mouse click on the third entry row moves cursor to position 3"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Mouse drag marks a contiguous range (TS-21)
# ===========================================
test_mouse_drag_marks_range() {
    start_duofm "$CURRENT_SESSION"

    local screen
    screen=$(capture_screen "$CURRENT_SESSION")

    # Row A: the first real entry (index 1, right after "..") is always on
    # row 6 at scroll offset 0 (row 5 + 1), regardless of what it is.
    local row_a=6
    # Row B: dir2, located by reading the captured screen so the sort
    # order is never assumed.
    local row_b
    row_b=$(find_entry_row "$screen" "dir2")

    if [ -z "$row_b" ]; then
        echo -e "${RED}✗${NC} Could not locate dir2 on screen for drag test"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        stop_duofm "$CURRENT_SESSION"
        return
    fi

    # Entries between row A and row B inclusive, excluding "..": row A is
    # already the first non-".." row, so the count is the number of rows
    # spanned regardless of direction.
    local row_diff=$((row_b - row_a))
    if [ "$row_diff" -lt 0 ]; then
        row_diff=$((-row_diff))
    fi
    local expected_count=$((row_diff + 1))

    send_mouse_sequences "$CURRENT_SESSION" \
        "$(mouse_press_seq "$LEFT_PANE_COL" "$row_a")"
    send_mouse_sequences "$CURRENT_SESSION" \
        "$(mouse_motion_seq "$LEFT_PANE_COL" "$row_b")"
    send_mouse_sequences "$CURRENT_SESSION" \
        "$(mouse_release_seq "$LEFT_PANE_COL" "$row_b")"

    assert_contains "$CURRENT_SESSION" "Marked ${expected_count}/" \
        "Header shows Marked ${expected_count}/ after dragging from row ${row_a} to row ${row_b}"

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Mouse double-click enters directory (TS-22)
# ===========================================
test_mouse_double_click_enters_directory() {
    start_duofm "$CURRENT_SESSION"

    local screen
    screen=$(capture_screen "$CURRENT_SESSION")

    local row_dir2
    row_dir2=$(find_entry_row "$screen" "dir2")

    if [ -z "$row_dir2" ]; then
        echo -e "${RED}✗${NC} Could not locate dir2 on screen for double-click test"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        stop_duofm "$CURRENT_SESSION"
        return
    fi

    # Two press/release pairs on the same row, sent in a single call so
    # they land well under the 500ms double-click window.
    send_mouse_sequences "$CURRENT_SESSION" \
        "$(mouse_press_seq "$LEFT_PANE_COL" "$row_dir2")" \
        "$(mouse_release_seq "$LEFT_PANE_COL" "$row_dir2")" \
        "$(mouse_press_seq "$LEFT_PANE_COL" "$row_dir2")" \
        "$(mouse_release_seq "$LEFT_PANE_COL" "$row_dir2")"

    assert_contains "$CURRENT_SESSION" "another.txt" \
        "Double-click on dir2 enters the directory"

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Mouse"
    echo "========================================"

    run_test test_mouse_click_moves_cursor
    run_test test_mouse_drag_marks_range
    run_test test_mouse_double_click_enters_directory

    print_summary
    exit $?
fi
