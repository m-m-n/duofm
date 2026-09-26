#!/bin/bash
# Runner Self-Test for duofm E2E test runner
#
# Description: Exercises fixture copies of run_all_tests.sh and helpers.sh
#              to prove the run-list guard (unregistered tests/*.sh files),
#              run_test's undefined-function handling, and the existing
#              undefined/missing run-list checks all still work together.
#              Scenarios TS-1 through TS-5 are documented in the task plan.
#
# Usage:
#   ./runner_self_test.sh
#
# Exit status: 0 if and only if every scenario passed. Non-zero if any
# scenario failed, including when a scenario's fixture lists could not be
# applied to the runner copy (fail closed).
#
# Writes nothing outside a temporary directory this script creates itself,
# and starts no tmux session or duofm process (the only tmux invocation
# that can occur is the session-cleanup call inside run_test, which is a
# harmless no-op for the fixture tests here since none of them start a
# session).

set -u

SELF_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

SELF_TMPDIR="$(mktemp -d "${TMPDIR:-/tmp}/runner_self_test.XXXXXX")"
cleanup_self_tmpdir() {
    rm -rf "$SELF_TMPDIR"
}
trap cleanup_self_tmpdir EXIT

OVERALL_STATUS=0

# ---------------------------------------------------------------------------
# Shared fixture helpers
# ---------------------------------------------------------------------------

# Strip ANSI color escape sequences so summary counts can be parsed exactly.
strip_colors() {
    sed -E 's/\x1b\[[0-9;]*m//g'
}

# Extract "Total: N" / "Passed: N" / "Failed: N" from normalized summary
# text. Prints "Total=<n> Passed=<n> Failed=<n>" on success, or
# "UNREADABLE" (and a non-zero return) when any count cannot be parsed.
extract_counts() {
    local text="$1"
    local total passed failed
    total="$(printf '%s\n' "$text" | grep -oE '^Total:[[:space:]]+[0-9]+' | grep -oE '[0-9]+' | head -1)"
    passed="$(printf '%s\n' "$text" | grep -oE '^Passed:[[:space:]]+[0-9]+' | grep -oE '[0-9]+' | head -1)"
    failed="$(printf '%s\n' "$text" | grep -oE '^Failed:[[:space:]]+[0-9]+' | grep -oE '[0-9]+' | head -1)"
    if [ -z "$total" ] || [ -z "$passed" ] || [ -z "$failed" ]; then
        echo "UNREADABLE"
        return 1
    fi
    echo "Total=$total Passed=$passed Failed=$failed"
    return 0
}

# Create a fresh fixture directory at $1 containing copies of the runner and
# helpers sitting next to this self-test, plus an empty tests/ subdirectory
# for the caller to populate.
new_fixture_dir() {
    local dir="$1"
    mkdir -p "$dir/tests"
    cp "${SELF_DIR}/run_all_tests.sh" "$dir/run_all_tests.sh"
    cp "${SELF_DIR}/helpers.sh" "$dir/helpers.sh"
    chmod +x "$dir/run_all_tests.sh"
}

# Write a fixture test file at <dir>/tests/<file_basename> defining one
# passing test function named <fn_name>. It only updates the tally counters
# and prints a line; it never starts tmux or duofm. Loads helpers.sh
# relative to its own location, like the real tests/*.sh files.
write_passing_test_file() {
    local dir="$1" file_basename="$2" fn_name="$3"
    cat > "${dir}/tests/${file_basename}" <<EOF
#!/bin/bash
SCRIPT_DIR="\$(cd "\$(dirname "\${BASH_SOURCE[0]}")" && pwd)"
source "\${SCRIPT_DIR}/../helpers.sh"

${fn_name}() {
    TESTS_RUN=\$((TESTS_RUN + 1))
    TESTS_PASSED=\$((TESTS_PASSED + 1))
    echo "fixture: ${fn_name} passed"
}
EOF
}

# Replace the TEST_FILES and FULL_RUN_LIST array literals inside a runner
# copy with the given body text (already-indented bash source lines), then
# confirm the replacement actually took effect by checking that a
# production-only entry is no longer present. Returns 1 (does not exit) on
# any failure to apply, so callers can fail the scenario closed.
apply_fixture_lists() {
    local runner_copy="$1" test_files_body="$2" full_run_list_body="$3"
    local tmp

    tmp="$(mktemp)"
    awk -v body="$test_files_body" '
        /^declare -A TEST_FILES=\($/ { print; print body; skip = 1; next }
        skip == 1 { if ($0 ~ /^\)$/) { print; skip = 0 }; next }
        { print }
    ' "$runner_copy" > "$tmp" && mv "$tmp" "$runner_copy"

    tmp="$(mktemp)"
    awk -v body="$full_run_list_body" '
        /^FULL_RUN_LIST=\($/ { print; print body; skip = 1; next }
        skip == 1 { if ($0 ~ /^\)$/) { print; skip = 0 }; next }
        { print }
    ' "$runner_copy" > "$tmp" && mv "$tmp" "$runner_copy"

    # mv replaced the file with a mktemp-created one (mode 600); restore the
    # executable bit the fixture copy needs to be run as a program.
    chmod +x "$runner_copy"

    # Fail closed: confirm the copy no longer carries production entries.
    if grep -q 'basic_tests\.sh' "$runner_copy"; then
        return 1
    fi
    if grep -q 'test_basic_startup' "$runner_copy"; then
        return 1
    fi
    return 0
}

# ---------------------------------------------------------------------------
# Scenarios
# ---------------------------------------------------------------------------

# TS-1: fixture A (registered) and B (unregistered) each with one passing
# test; TEST_FILES lists A only; FULL_RUN_LIST lists both tests. Full run
# (no argument) must exit non-zero with Total=3 Passed=1 Failed=2, and
# report both an unregistered-file line for B and an undefined-function
# line for B's test.
scenario_ts1() {
    local dir="${SELF_TMPDIR}/ts1"
    new_fixture_dir "$dir"
    write_passing_test_file "$dir" "a_tests.sh" "test_pass_a"
    write_passing_test_file "$dir" "b_tests.sh" "test_pass_b"

    local test_files_body='    ["a"]="a_tests.sh"'
    local full_run_list_body='    "Fixture Tests|test_pass_a"
    "Fixture Tests|test_pass_b"'

    if ! apply_fixture_lists "${dir}/run_all_tests.sh" "$test_files_body" "$full_run_list_body"; then
        echo "  fixture lists were not applied to the runner copy"
        return 1
    fi

    local output status normalized counts ok=1
    output="$(cd "$dir" && ./run_all_tests.sh 2>&1)"
    status=$?
    normalized="$(printf '%s\n' "$output" | strip_colors)"

    if [ "$status" -eq 0 ]; then
        echo "  expected non-zero exit status, got 0"
        ok=0
    fi

    counts="$(extract_counts "$normalized")"
    if [ "$counts" != "Total=3 Passed=1 Failed=2" ]; then
        echo "  expected Total=3 Passed=1 Failed=2, got: $counts"
        ok=0
    fi

    if ! printf '%s\n' "$normalized" | grep -q "Test file not registered in TEST_FILES: b_tests.sh"; then
        echo "  missing unregistered-file line for b_tests.sh"
        ok=0
    fi

    if ! printf '%s\n' "$normalized" | grep -q "No test function named 'test_pass_b' is defined"; then
        echo "  missing undefined-function line for test_pass_b"
        ok=0
    fi

    [ "$ok" -eq 1 ]
}

# TS-2: same fixture as TS-1. --check-list must exit 1, print the
# unregistered-file line for B, and never print the OK line.
scenario_ts2() {
    local dir="${SELF_TMPDIR}/ts2"
    new_fixture_dir "$dir"
    write_passing_test_file "$dir" "a_tests.sh" "test_pass_a"
    write_passing_test_file "$dir" "b_tests.sh" "test_pass_b"

    local test_files_body='    ["a"]="a_tests.sh"'
    local full_run_list_body='    "Fixture Tests|test_pass_a"
    "Fixture Tests|test_pass_b"'

    if ! apply_fixture_lists "${dir}/run_all_tests.sh" "$test_files_body" "$full_run_list_body"; then
        echo "  fixture lists were not applied to the runner copy"
        return 1
    fi

    local output status normalized ok=1
    output="$(cd "$dir" && ./run_all_tests.sh --check-list 2>&1)"
    status=$?
    normalized="$(printf '%s\n' "$output" | strip_colors)"

    if [ "$status" -ne 1 ]; then
        echo "  expected exit status 1, got $status"
        ok=0
    fi
    if ! printf '%s\n' "$normalized" | grep -q "Test file not registered in TEST_FILES: b_tests.sh"; then
        echo "  missing unregistered-file line for b_tests.sh"
        ok=0
    fi
    if printf '%s\n' "$normalized" | grep -q "^OK:"; then
        echo "  unexpected OK line present"
        ok=0
    fi

    [ "$ok" -eq 1 ]
}

# TS-3: every fixture file registered, every defined test listed.
# --check-list must print the OK line and exit 0; the full run must report
# Failed=0 and exit 0.
scenario_ts3() {
    local dir="${SELF_TMPDIR}/ts3"
    new_fixture_dir "$dir"
    write_passing_test_file "$dir" "a_tests.sh" "test_pass_a"
    write_passing_test_file "$dir" "b_tests.sh" "test_pass_b"

    local test_files_body='    ["a"]="a_tests.sh"
    ["b"]="b_tests.sh"'
    local full_run_list_body='    "Fixture Tests|test_pass_a"
    "Fixture Tests|test_pass_b"'

    if ! apply_fixture_lists "${dir}/run_all_tests.sh" "$test_files_body" "$full_run_list_body"; then
        echo "  fixture lists were not applied to the runner copy"
        return 1
    fi

    local ok=1

    local check_output check_status
    check_output="$(cd "$dir" && ./run_all_tests.sh --check-list 2>&1)"
    check_status=$?
    if [ "$check_status" -ne 0 ]; then
        echo "  --check-list: expected exit status 0, got $check_status"
        ok=0
    fi
    if ! printf '%s\n' "$check_output" | strip_colors | grep -q "^OK:"; then
        echo "  --check-list: missing OK line"
        ok=0
    fi

    local run_output run_status counts
    run_output="$(cd "$dir" && ./run_all_tests.sh 2>&1)"
    run_status=$?
    if [ "$run_status" -ne 0 ]; then
        echo "  full run: expected exit status 0, got $run_status"
        ok=0
    fi
    counts="$(extract_counts "$(printf '%s\n' "$run_output" | strip_colors)")"
    if [ "${counts#*Failed=}" != "0" ] || [ "$counts" = "UNREADABLE" ]; then
        echo "  full run: expected Failed=0, got: $counts"
        ok=0
    fi

    [ "$ok" -eq 1 ]
}

# TS-4: same as TS-1 but B's test is not in FULL_RUN_LIST. --check-list must
# exit 1, print the unregistered-file line for B, and also print the
# existing "Missing from run list" line for B's test.
scenario_ts4() {
    local dir="${SELF_TMPDIR}/ts4"
    new_fixture_dir "$dir"
    write_passing_test_file "$dir" "a_tests.sh" "test_pass_a"
    write_passing_test_file "$dir" "b_tests.sh" "test_pass_b"

    local test_files_body='    ["a"]="a_tests.sh"'
    local full_run_list_body='    "Fixture Tests|test_pass_a"'

    if ! apply_fixture_lists "${dir}/run_all_tests.sh" "$test_files_body" "$full_run_list_body"; then
        echo "  fixture lists were not applied to the runner copy"
        return 1
    fi

    local output status normalized ok=1
    output="$(cd "$dir" && ./run_all_tests.sh --check-list 2>&1)"
    status=$?
    normalized="$(printf '%s\n' "$output" | strip_colors)"

    if [ "$status" -ne 1 ]; then
        echo "  expected exit status 1, got $status"
        ok=0
    fi
    if ! printf '%s\n' "$normalized" | grep -q "Test file not registered in TEST_FILES: b_tests.sh"; then
        echo "  missing unregistered-file line for b_tests.sh"
        ok=0
    fi
    if ! printf '%s\n' "$normalized" | grep -q "^Missing from run list: test_pass_b (b_tests.sh)$"; then
        echo "  missing 'Missing from run list' line for test_pass_b"
        ok=0
    fi

    [ "$ok" -eq 1 ]
}

# TS-5: a copy of helpers.sh loaded into a child shell. run_test with a name
# that is not defined must not invoke it, must increase TESTS_RUN and
# TESTS_FAILED by exactly one each while leaving TESTS_PASSED unchanged,
# must print an undefined-function line naming it, and must produce no
# "command not found" diagnostic. Runs the child shell with set -e, since
# real callers run with exit-on-error enabled.
scenario_ts5() {
    local dir="${SELF_TMPDIR}/ts5"
    mkdir -p "$dir"
    cp "${SELF_DIR}/helpers.sh" "${dir}/helpers.sh"

    local script="${dir}/run.sh"
    cat > "$script" <<'EOF'
#!/bin/bash
set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/helpers.sh"
run_test "this_function_is_not_defined_anywhere"
echo "RUN=${TESTS_RUN} PASSED=${TESTS_PASSED} FAILED=${TESTS_FAILED}"
EOF
    chmod +x "$script"

    local output status normalized ok=1
    output="$(bash "$script" 2>&1)"
    status=$?
    normalized="$(printf '%s\n' "$output" | strip_colors)"

    if [ "$status" -ne 0 ]; then
        echo "  expected exit status 0, got $status"
        ok=0
    fi
    if ! printf '%s\n' "$normalized" | grep -q "^RUN=1 PASSED=0 FAILED=1$"; then
        echo "  expected RUN=1 PASSED=0 FAILED=1, got:"
        printf '%s\n' "$normalized" | sed 's/^/    /'
        ok=0
    fi
    if ! printf '%s\n' "$normalized" | grep -q "No test function named 'this_function_is_not_defined_anywhere' is defined"; then
        echo "  missing undefined-function line"
        ok=0
    fi
    if printf '%s\n' "$normalized" | grep -qi "command not found"; then
        echo "  unexpected 'command not found' diagnostic"
        ok=0
    fi

    [ "$ok" -eq 1 ]
}

# ---------------------------------------------------------------------------
# Runner
# ---------------------------------------------------------------------------

run_scenario() {
    local id="$1" desc="$2" fn="$3"
    local output status
    output="$("$fn" 2>&1)"
    status=$?
    if [ "$status" -eq 0 ]; then
        echo "${id} ${desc}: PASS"
    else
        echo "${id} ${desc}: FAIL"
        printf '%s\n' "$output" | sed 's/^/    /'
        OVERALL_STATUS=1
    fi
}

main() {
    run_scenario "TS-1" "full run counts an unregistered file and its undefined test" scenario_ts1
    run_scenario "TS-2" "check-list reports an unregistered file" scenario_ts2
    run_scenario "TS-3" "check-list OK and full run Failed=0 when fully registered" scenario_ts3
    run_scenario "TS-4" "check-list reports an unregistered file together with a missing entry" scenario_ts4
    run_scenario "TS-5" "run_test given an undefined name never invokes it" scenario_ts5

    exit "$OVERALL_STATUS"
}

main "$@"
