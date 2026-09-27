# Verification Document: drag-refresh-stale-anchor

## Overview

**Feature**: drag-refresh-stale-anchor / **SPEC.md**: `feature-docs/drag-refresh-stale-anchor/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/drag-refresh-stale-anchor/IMPLEMENTATION.md`

## Build Verification

- Command: `make build`
- Expected: exit code 0, no errors

## Test Verification

- Command: `go test ./...`
- Coverage target: no numeric target; TS-1..TS-8 exercise every new branch (cancel on mismatch, continue on match, baseline restriction)

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | Temp-directory model: left-pane press (armed), add a file to the left pane's directory, send `autoRefreshMsg`, then move and release on another row | The session is back to its initial value after the move; the left pane's mark set equals the set right after the refresh | Unit |
| TS-2 | Left-pane press and move mark a range; delete an unmarked file; send `autoRefreshMsg`; then move and release on another row | Marks right after the refresh are kept; the move and release change no marks | Unit |
| TS-3 | Left-pane press; send `autoRefreshMsg` without changing any file; then move | The session stays armed; the range from the original anchor is marked | Unit |
| TS-4 | Left-pane press; add a file only to the right pane's directory; send `autoRefreshMsg`; then move | The left session is unchanged; the range from the original anchor is marked | Unit |
| TS-5 | Left-pane press; add a file to the left pane's directory; send `execFinishedMsg` (and, separately, `shellCommandFinishedMsg`); then move and release | The mark set equals the set right after the refresh | Unit |
| TS-6 | Left-pane press; change only the content of an existing file (names and order unchanged); send `autoRefreshMsg`; then move | The range from the original anchor is marked | Unit |
| TS-7 | Filter applied to the left pane; press; add a file that does not match the filter; send `autoRefreshMsg`; then move | The range from the original anchor is marked | Unit |
| TS-8 | A left-pane file hidden by the filter is marked; press while filtered; delete that file; send `autoRefreshMsg` (displayed list unchanged); then move and release | The deleted file's name is not in the mark set | Unit |
| TS-9 | `go test ./...`, `gofmt -w .`, `go vet ./...` | All tests pass; gofmt produces no diff; go vet reports nothing | Static / Unit |

## Code Quality Verification

- Format: `gofmt -w .` — expected: no diff in the working tree afterwards
- Static analysis: `go vet ./...` — expected: no findings

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC1 | After an armed press and a list-changing `autoRefreshMsg`, a move and release on another row leave the post-refresh mark set and the session is not armed | TS-1 |
| AC2 | After an active range drag and a list-changing `autoRefreshMsg` (file deleted), a move and release on another row leave the post-refresh mark set | TS-2 |
| AC3 | After an `autoRefreshMsg` with no file change, a move marks the range from the original anchor | TS-3 |
| AC4 | A file added only to the right pane's directory leaves the left session unchanged; a move marks the range from the original anchor | TS-4 |
| AC5 | A list change through `execFinishedMsg` / `shellCommandFinishedMsg` leaves marks unchanged on move and release | TS-5 |
| AC6 | After a content-only change, or an addition that does not match the filter, a move marks the range from the original anchor | TS-6, TS-7 |
| AC7 | A file deleted by the refresh does not reappear in the mark set when the drag continues | TS-8 |
| AC8 | The AC1 and AC2 tests fail on the pre-fix code and pass after the fix | TS-1 and TS-2 pass on the integrated branch; with only the new model test file (`internal/ui/model_update_mouse_refresh_test.go`) placed on the base revision, TS-1 and TS-2 fail |
| AC9 | Existing unit and E2E tests pass; gofmt and go vet are clean | TS-9 and the E2E run below |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1, TS-2 |
| FR2 | task0001 | TS-5 |
| FR3 | task0001 | TS-3, TS-6, TS-7 |
| FR4 | task0001 | TS-4 |
| FR5 | task0001 | TS-2 |
| FR6 | task0001 | TS-8 |
| FR7 | task0001 | TS-1, TS-2 (added unit tests; fail before the fix, pass after) |
| NFR1 | task0001 | TS-9 (existing click, drag and double-click unit tests pass unchanged) and the E2E run (`mouse_tests.sh`) |
| NFR2 | task0001 | TS-9 (existing `RefreshDirectoryPreserveCursor` unit tests pass unchanged) |
| NFR3 | task0001 | TS-9 (drag-dir-load-stale-marks TS-1..TS-5 in `model_update_mouse_test.go` pass unchanged) |
| NFR4 | task0001 | TS-9 and the E2E run |

## E2E Testing

- Command: `make test-e2e-build && make test-e2e`
- [ ] Existing E2E tests, including `mouse_tests.sh` and `mark_tests.sh`, pass without regression (no E2E test is added)

## Manual Testing (E2E Not Possible)

- None. The reproduction steps are reproduced at the model level by TS-1 and TS-2.

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit tests (TS-1..TS-8) | 8 | 8 | 0 | 0 |
| Static checks (TS-9) | 1 | 1 | 0 | 0 |
| Existing E2E regression | 1 | 0 | 1 | 0 |
| Total | 11 | 10 | 1 | 0 |
