#!/bin/bash
# E2E Test Helper Functions for duofm

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Session name prefix
SESSION_PREFIX="duofm_e2e"

# Start duofm in a tmux session
# Usage: start_duofm <session_name> [working_dir] [width] [height]
start_duofm() {
    local session_name="${SESSION_PREFIX}_$1"
    local work_dir="${2:-/testdata}"
    local width="${3:-120}"
    local height="${4:-40}"

    # Kill existing session if any
    tmux kill-session -t "$session_name" 2>/dev/null || true

    # Create new session and run duofm
    tmux new-session -d -s "$session_name" -x "$width" -y "$height" -c "$work_dir" "duofm"

    # Wait for startup
    sleep 0.5
}

# Send keys to duofm session
# Usage: send_keys <session_name> <keys...>
send_keys() {
    local session_name="${SESSION_PREFIX}_$1"
    shift

    for key in "$@"; do
        tmux send-keys -t "$session_name" "$key"
        sleep 0.1
    done

    # Wait for UI to update
    sleep 0.3
}

# Capture current screen content
# Usage: capture_screen <session_name>
capture_screen() {
    local session_name="${SESSION_PREFIX}_$1"
    tmux capture-pane -t "$session_name" -p
}

# Stop duofm session
# Usage: stop_duofm <session_name>
stop_duofm() {
    local session_name="${SESSION_PREFIX}_$1"
    tmux send-keys -t "$session_name" 'q'
    sleep 0.2
    tmux kill-session -t "$session_name" 2>/dev/null || true
}

# Assert screen contains text
# Usage: assert_contains <session_name> <expected_text> <test_description>
assert_contains() {
    local session_name="$1"
    local expected="$2"
    local description="$3"

    TESTS_RUN=$((TESTS_RUN + 1))

    local screen
    screen=$(capture_screen "$session_name")

    if echo "$screen" | grep -qF -- "$expected"; then
        echo -e "${GREEN}✓${NC} $description"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $description"
        echo -e "  ${YELLOW}Expected to find:${NC} $expected"
        echo -e "  ${YELLOW}Screen content:${NC}"
        echo "$screen" | head -10 | sed 's/^/    /'
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Assert screen does NOT contain text
# Usage: assert_not_contains <session_name> <unexpected_text> <test_description>
assert_not_contains() {
    local session_name="$1"
    local unexpected="$2"
    local description="$3"

    TESTS_RUN=$((TESTS_RUN + 1))

    local screen
    screen=$(capture_screen "$session_name")

    if ! echo "$screen" | grep -qF -- "$unexpected"; then
        echo -e "${GREEN}✓${NC} $description"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $description"
        echo -e "  ${YELLOW}Did not expect to find:${NC} $unexpected"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Assert cursor position (by checking status bar)
# Usage: assert_cursor_position <session_name> <expected_position> <test_description>
assert_cursor_position() {
    local session_name="$1"
    local expected="$2"
    local description="$3"

    TESTS_RUN=$((TESTS_RUN + 1))

    local screen
    screen=$(capture_screen "$session_name")

    # Status bar shows position like "3/11"
    if echo "$screen" | grep -q " ${expected}/"; then
        echo -e "${GREEN}✓${NC} $description"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $description"
        echo -e "  ${YELLOW}Expected cursor at:${NC} $expected"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Compare screen with golden file
# Usage: assert_golden <session_name> <golden_file> <test_description>
assert_golden() {
    local session_name="$1"
    local golden_file="$2"
    local description="$3"

    TESTS_RUN=$((TESTS_RUN + 1))

    local screen
    screen=$(capture_screen "$session_name")

    if [ ! -f "$golden_file" ]; then
        # Create golden file if it doesn't exist
        echo "$screen" > "$golden_file"
        echo -e "${YELLOW}⚠${NC} $description (golden file created)"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    fi

    if diff -q <(echo "$screen") "$golden_file" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $description"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $description"
        echo -e "  ${YELLOW}Diff:${NC}"
        diff <(echo "$screen") "$golden_file" | head -20 | sed 's/^/    /'
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Print test summary
print_summary() {
    echo ""
    echo "========================================"
    echo "Test Summary"
    echo "========================================"
    echo -e "Total:  $TESTS_RUN"
    echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
    echo "========================================"

    if [ $TESTS_FAILED -gt 0 ]; then
        return 1
    fi
    return 0
}

# Run a test function with setup/teardown
# Usage: run_test <test_function_name>
run_test() {
    local test_name="$1"
    local session_name="test_$$_$RANDOM"

    echo ""
    echo "--- Running: $test_name ---"

    # Export session name for test function
    export CURRENT_SESSION="$session_name"

    # Run the test
    if "$test_name"; then
        : # Test passed
    else
        : # Test failed (already logged)
    fi

    # Cleanup
    tmux kill-session -t "${SESSION_PREFIX}_${session_name}" 2>/dev/null || true
}
