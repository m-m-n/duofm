# Verification Document: e2e-rename-initial-value-check

## Overview

**Feature**: e2e-rename-initial-value-check / **SPEC.md**: `feature-docs/e2e-rename-initial-value-check/SPEC.md` / **IMPLEMENTATION.md**: not written (reduced tier; single task, no files shared between tasks) / **Task plan**: `feature-docs/e2e-rename-initial-value-check/tasks/task0001.md`

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
| TS-1 | Run the fixed rename tests: `make test-e2e` (or `/e2e/scripts/run_all_tests.sh file-ops` inside the E2E container) | Every check in `test_rename_file` and `test_navigation_after_rename` passes, including the new input-line initial-value checks | E2E |
| TS-2 | With a local, uncommitted change that sends `C-u` immediately before each initial-value check, run the file-ops tests | The initial-value checks in both tests fail, even though `before_rename.txt` / `navren_before.txt` and the search-filter display are on screen | E2E (mutation, manual setup) |
| TS-3 | Run the history tests: `make test-e2e` (or `/e2e/scripts/run_all_tests.sh history` inside the E2E container) | All history tests pass, including `test_history_forward_cleared`. The `assert_contains "testdata"` call ("[ key navigates back toward /testdata") is absent from that test, and `assert_not_contains "/testdata/dir1"` remains | E2E + source inspection |
| TS-4 | Check the diff against the base (`workflow.implement.base_commit`) and run `make test` (via `test_command`) | No `*.go` file in the diff. `test/e2e/scripts/helpers.sh` has no diff. Unit tests pass | Automated (diff check + unit regression) |

### Edge Cases (from SPEC.md, covered under TS-2)

| Edge case | How it is exercised | Expected |
|-----------|---------------------|----------|
| Input initial value is empty | TS-2 mutation (`C-u` before the check) | Check fails |
| Input initial value is the full name with extension | TS-2 variant: a local, uncommitted change that clears the input and types `before_rename.txt` / `navren_before.txt` before the check | Check fails |
| File-list rows behind the dialog visible or hidden | TS-2: the file-list row with the full file name stays on screen during the mutated run | Result depends only on the input field's content (check fails under the mutation, passes under TS-1) |
| Search-filter display (`/before_ren`, `/navren_before`) still on screen | TS-2 in `test_navigation_after_rename`, whose filter text contains the full base name | Check fails under the mutation. The filter display never satisfies the check |

## Code Quality Verification

- Format: `project.components.duofm.format_command` in `workflow.yaml` (`gofmt -w .`). Expected: no file changes (no Go source is touched)
- Shell syntax: `bash -n` on both modified scripts. Expected: exit code 0
- Run-list guard: `test/e2e/scripts/run_all_tests.sh --check-list`. Expected: `OK: run list matches defined tests` (no test function added, removed, or renamed; the new helper is not detected as a test)

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC-1 | The initial-value checks in `test_rename_file` and `test_navigation_after_rename` are performed in a form that can only appear on the input field row | Source inspection: both tests call the input-line helper, and no bare base-name `assert_contains` remains; TS-1 passes |
| AC-2 | When the input field's initial value differs from the expected base name, both initial-value checks fail | TS-2 and its full-name variant |
| AC-3 | The always-passing `assert_contains "testdata"` in `history_tests.sh` is gone and `assert_not_contains "/testdata/dir1"` remains | TS-3 (source inspection part) |
| AC-4 | All file-ops and history tests pass under `make test-e2e` | TS-1, TS-3 |
| AC-5 | The change contains no `*.go` files | TS-4 |
| SC-1 | All functional requirements are implemented and tested | Functional Requirements Coverage table below |
| SC-2 | All test scenarios pass | TS-1 to TS-4 |
| SC-3 | Code review is completed | Review phase result in `workflow.yaml` |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1 (passes on the correct value), TS-2 (fails on an empty or full-name value) |
| FR2 | task0001 | TS-1 (passes on the correct value), TS-2 (fails on an empty or full-name value; filter display does not satisfy the check) |
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
part of the automated suite. Revert every edit after the run.

- [ ] TS-2: `C-u` sent immediately before each initial-value check. Both initial-value checks fail
- [ ] TS-2 full-name variant: input cleared and `before_rename.txt` / `navren_before.txt` typed immediately before each check. Both checks fail
- [ ] TS-3 source inspection: `test_history_forward_cleared` no longer contains the `testdata` `assert_contains`, and still contains `assert_not_contains "/testdata/dir1"`

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit test regression | 1 | 1 | 0 | 0 |
| Code quality (format, shell syntax, run-list guard) | 3 | 3 | 0 | 0 |
| Test scenarios (TS-1 to TS-4) | 4 | 1 (TS-4) | 2 (TS-1, TS-3) | 1 (TS-2) |
| Edge cases | 4 | 0 | 0 | 4 (via TS-2 and its variant) |
