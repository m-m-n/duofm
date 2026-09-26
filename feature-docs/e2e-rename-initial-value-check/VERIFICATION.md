# Verification Document: e2e-rename-initial-value-check

## Overview

**Feature**: e2e-rename-initial-value-check / **SPEC.md**: `feature-docs/e2e-rename-initial-value-check/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/e2e-rename-initial-value-check/IMPLEMENTATION.md` / **REQUIREMENTS.md**: `feature-docs/e2e-rename-initial-value-check/REQUIREMENTS.md`

The change is limited to two E2E test scripts
(`test/e2e/scripts/tests/file_operation_tests.sh`,
`test/e2e/scripts/tests/history_tests.sh`). Build, unit test, and format
checks are regression checks only; they confirm that nothing outside the
scripts changed.

## Build Verification

- Command: `project.components.duofm.build_command` in `workflow.yaml` (runs `make build` in the `golang:1.25.5-trixie` container)
- Expected: exit code 0, no errors

## Test Verification

- Command: `project.components.duofm.test_command` in `workflow.yaml` (runs `go test ./...` in the same container)
- Expected: exit code 0
- Coverage target: N/A (no Go source changes)

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | Run the fixed rename tests: `make test-e2e` (or `/e2e/scripts/run_all_tests.sh file-ops` inside the E2E container) | Every check in `test_rename_file` and `test_navigation_after_rename` passes, including both initial-value checks | E2E |
| TS-2 | With a local, uncommitted change, one variant per run: (a) send `C-u` immediately before each initial-value check to clear the input field; (b) make the input field hold the full name with extension immediately before each check (for example send `End` `.` `t` `x` `t`). Run the file-ops tests for each variant | In each variant, the initial-value checks in both tests fail, even though `before_rename.txt` / `navren_before.txt` are shown in the file list | E2E (mutation, manual setup) |
| TS-3 | Run the history tests: `make test-e2e` (or `/e2e/scripts/run_all_tests.sh history` inside the E2E container) | All history tests pass, including `test_history_forward_cleared`. The `assert_contains "testdata"` call ("[ key navigates back toward /testdata") is absent from that test, and `assert_not_contains "/testdata/dir1"` remains | E2E + source inspection |
| TS-4 | Check the diff against the base (`workflow.implement.base_commit`) and run `make test` (via `test_command`) | No `*.go` file in the diff. `test/e2e/scripts/helpers.sh` has no diff. Unit tests pass | Automated (diff check + unit regression) |
| TS-5 | With a local, uncommitted change, immediately before the initial-value check in each test, send keys that add one extra character to the pre-filled value, one variant per run: (a) `End` `.` → `before_rename.` / `navren_before.`; (b) `Home` `x` → `xbefore_rename` / `xnavren_before`; (c) `End` `x` → `before_renamex` / `navren_beforex`. Run the file-ops tests for each variant | In every variant, the initial-value check fails in both `test_rename_file` and `test_navigation_after_rename`. The results of later checks in the mutated runs are not evaluated | E2E (mutation, manual setup) |

### Edge Cases (from SPEC.md)

| Edge case | How it is exercised | Expected |
|-----------|---------------------|----------|
| Input initial value is empty | TS-2 (a) | Check fails |
| Input initial value is the full name with extension (`before_rename.txt` / `navren_before.txt`) | TS-2 (b) | Check fails |
| One extra non-whitespace character before or after the base name (`before_rename.`, `xbefore_rename`, `before_renamex`, and the same for `navren_before`) | TS-5 (a), (b), (c) | Check fails |
| File-list rows behind the dialog visible or hidden | TS-1 and TS-2: the file-list row with the full file name stays on screen in both | The result depends only on the input field's content (passes under TS-1, fails under TS-2) |
| Search-filter display (`/before_ren`, `/navren_before`) still on screen | TS-2 and TS-5 in `test_navigation_after_rename`, whose filter text contains the full base name | Check fails under every mutation. The filter display never satisfies the check |
| The cursor at the end of the value is a reverse-video space | TS-1: the unmodified run checks the pre-filled value as the dialog shows it | Check passes; the end-of-value cursor cell does not make it fail |

## Code Quality Verification

- Format: `project.components.duofm.format_command` in `workflow.yaml` (`gofmt -w .`). Expected: no file changes (no Go source is touched)
- Shell syntax: `bash -n` on both modified scripts. Expected: exit code 0
- Run-list guard: `test/e2e/scripts/run_all_tests.sh --check-list`. Expected: `OK: run list matches defined tests` (no test function added, removed, or renamed; the file-local assertion is not detected as a test)

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC-1 | The initial-value checks in `test_rename_file` and `test_navigation_after_rename` are performed in a form that can only match the input field's row | Source inspection: both tests call the input-field value assertion, and no `assert_contains` on the bare base name remains; TS-1 passes |
| AC-2 | When the input field's initial value differs from the expected base name, the initial-value checks in both tests fail | TS-2 (a) and (b) |
| AC-3 | The always-passing `assert_contains "testdata"` in `history_tests.sh` is gone and `assert_not_contains "/testdata/dir1"` remains | TS-3 (source inspection part) |
| AC-4 | All file-ops and history tests pass under `make test-e2e` | TS-1, TS-3 |
| AC-5 | The change contains no `*.go` files | TS-4 |
| AC-6 | The initial-value match allows no extra non-whitespace character before or after the base name inside the input field | TS-5 (a), (b), (c); source inspection: between the two border characters, the pass condition tolerates only whitespace besides the base name |
| SC-1 | All functional requirements are implemented and tested | Functional Requirements Coverage table below |
| SC-2 | All test scenarios pass | TS-1 to TS-5 |
| SC-3 | Code review is completed | Review phase result in `workflow.yaml` |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1 (passes on the correct value), TS-2 (fails on an empty or full-name value), TS-5 (fails on one extra leading or trailing character) |
| FR2 | task0001 | TS-1 (passes on the correct value), TS-2 (fails on an empty or full-name value; the filter display does not satisfy the check), TS-5 (fails on one extra leading or trailing character) |
| FR3 | task0001 | TS-3 |
| NFR1 | task0001 | TS-4 (no `*.go` in the diff; unit tests pass) |
| NFR2 | task0001 | TS-4 (no diff in `helpers.sh`). TS-1 and TS-3 also run the existing assertions unchanged |

## E2E Testing

Framework: tmux-driven shell tests under `test/e2e/scripts/`, run in the E2E
container. Command: `project.components.duofm.e2e_test_command` in
`workflow.yaml` (`make test-e2e-build && make test-e2e`).

- [ ] TS-1: file-ops tests pass (0 failed), including both rename tests
- [ ] TS-3: history tests pass (0 failed), including `test_history_forward_cleared`
- [ ] Full E2E run passes with no regression in other categories

## Manual Testing (E2E Not Possible)

These need a temporary, uncommitted edit to the test script, so they are not
part of the automated suite. Each variant is a separate run. Revert every edit
after the run. In a mutated run, only the initial-value checks' results are
evaluated.

- [ ] TS-2 (a): `C-u` sent immediately before each initial-value check. Both initial-value checks fail
- [ ] TS-2 (b): the input field made to hold `before_rename.txt` / `navren_before.txt` immediately before each check. Both checks fail
- [ ] TS-5 (a): `End` `.` sent immediately before each check. Both checks fail
- [ ] TS-5 (b): `Home` `x` sent immediately before each check. Both checks fail
- [ ] TS-5 (c): `End` `x` sent immediately before each check. Both checks fail
- [ ] TS-3 source inspection: `test_history_forward_cleared` no longer contains the `testdata` `assert_contains`, and still contains `assert_not_contains "/testdata/dir1"`
- [ ] AC-1 / AC-6 source inspection: both rename tests call the input-field value assertion, and its pass condition tolerates only whitespace besides the base name between the two border characters

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit test regression | 1 | 1 | 0 | 0 |
| Code quality (format, shell syntax, run-list guard) | 3 | 3 | 0 | 0 |
| Test scenarios (TS-1 to TS-5) | 5 | 1 (TS-4) | 2 (TS-1, TS-3) | 2 (TS-2, TS-5) |
| Edge cases | 6 | 0 | 1 (end-of-value cursor, via TS-1) | 5 (via TS-2, TS-5, and TS-1 alongside) |
