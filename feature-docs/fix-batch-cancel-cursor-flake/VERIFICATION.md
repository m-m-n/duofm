# Verification Document: fix-batch-cancel-cursor-flake

## Overview

**Feature**: fix-batch-cancel-cursor-flake / **SPEC.md**: `feature-docs/fix-batch-cancel-cursor-flake/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/fix-batch-cancel-cursor-flake/IMPLEMENTATION.md`

Scenario IDs T1–T8 are the SPEC.md test scenario IDs. T9 and T10 are verification-only IDs added here to cover NFR1 and NFR4.

## Build Verification

- Command: `project.components.duofm.build_command` in `workflow.yaml` (containerized `make build`)
- Expected: exit code 0, no errors

## Test Verification

- Command: `project.components.duofm.test_command` in `workflow.yaml` (containerized `go test ./...`)
- Expected: exit code 0; every test in `internal/ui` passes, including the pre-existing mark tests
- Coverage target: no package-wide percentage target. Both changed queries (`GetMarkedFiles`, `GetMarkedFilePaths`) are exercised by T1–T6, including the filter-hidden tail and the zero / one mark cases.

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| T1 | Mark at least three names in the reverse of display order and call `GetMarkedFiles`, repeating the check many times in one run | Every repetition returns the names in display order | Unit |
| T2 | Mark directories and files mixed, in an order different from display order | Result equals the display list's order restricted to marked names; directories precede files | Unit |
| T3 | Hide some marked names with an active filter (at least two hidden, including a directory) | Visible marked names in display order, then hidden marked names in ascending name order; length equals `MarkCount()`; elements equal the mark set | Unit |
| T4 | Change the display order after marking (sort changed and re-applied, marks kept) | Result follows the display order at call time | Unit |
| T5 | Compare `GetMarkedFilePaths` with `GetMarkedFiles` for several marks, including a filter-hidden one | Same length; i-th path equals the pane path joined with the i-th name | Unit |
| T6 | Zero marks and exactly one mark | Empty lists for zero marks; exactly that name / that full path for one mark | Unit |
| T7 | `test_batch_cancel_cursor`: mark `01_aaa` and `02_bbb`, press `m`, wait with an upper bound for `already exists`, send `2` | Dialog observed within the bound; application still running; `dst/01_aaa.txt` exists; `src/02_bbb.txt` exists; `dst/02_bbb.txt` content is `existing` | E2E |
| T8 | Run the full E2E suite (`make test-e2e`) repeatedly | Every run passes, including `test_batch_cancel_cursor` | E2E |
| T9 | Signature compatibility of `GetMarkedFiles` and `GetMarkedFilePaths` (verification-only ID) | The build succeeds with every existing caller unchanged; the integrated diff shows no change to either method's signature | Build / diff check |
| T10 | Formatting and static analysis (verification-only ID) | Running the format command leaves no change in tracked Go files; go vet over all packages reports nothing | Static analysis |

## Code Quality Verification

- Format: `project.components.duofm.format_command` in `workflow.yaml` (containerized `gofmt -w .`); afterwards the working tree shows no change to Go files
- Static analysis: go vet over all packages (`go vet ./...`), run in the same container image as the test command; expected: no findings

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC1 | `GetMarkedFiles` returns entries in display order regardless of marking order; repeated checks all match | T1, T2, T4 |
| AC2 | Marked names absent from the display list follow the visible ones in ascending name order; count equals `MarkCount()` | T3, T6 |
| AC3 | The i-th element of `GetMarkedFilePaths` equals the pane path joined with the i-th element of `GetMarkedFiles` | T5 |
| AC4 | Hardened `test_batch_cancel_cursor` passes: dialog observed, `dst/01_aaa.txt` exists, `src/02_bbb.txt` remains, `dst/02_bbb.txt` content is `existing` | T7 |
| AC5 | Repeated `make test-e2e` runs all pass `test_batch_cancel_cursor`; all existing unit and E2E tests pass | T8, plus Test Verification above |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | T1, T2, T4, T6 |
| FR2 | task0001 | T3, T6 |
| FR3 | task0001 | T5 |
| FR4 | task0001 | T1, T2, T4, T5, T6 |
| FR5 | task0001 | T7 |
| NFR1 | task0001 | T9 |
| NFR2 | task0001 | T3, T6 |
| NFR3 | task0001 | T8 |
| NFR4 | task0001 | T10 |

## E2E Testing

- Command: `project.components.duofm.e2e_test_command` in `workflow.yaml` (`make test-e2e-build && make test-e2e`)
- [ ] T7: `test_batch_cancel_cursor` passes with all of its assertions, including the bounded-wait assertion for `already exists`
- [ ] T7: the runner's run-list check (`test/e2e/scripts/run_all_tests.sh --check-list`) passes, i.e. no new helper function was picked up as a test
- [ ] T8: repeat the full E2E run; every run passes. The number of runs is decided in this phase (SPEC 14.1 A-repeat-e2e); record the count and each run's result
- [ ] Existing E2E tests pass without regression

## Manual Testing (E2E Not Possible)

- None. The feature has no visual or interaction change, and the design step was skipped.

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit tests (T1–T6) | 6 | 6 | 0 | 0 |
| E2E (T7, T8) | 2 | 0 | 2 | 0 |
| Compatibility / quality (T9, T10) | 2 | 2 | 0 | 0 |
| Manual | 0 | 0 | 0 | 0 |
