# Feature: e2e-rename-initial-value-check

## Overview

The E2E rename tests check the initial value of the rename dialog's input field. This feature makes those checks match only the input field line, so that they fail when the input field's initial value is not the expected base name. It also removes an always-passing `testdata` check from `history_tests.sh`.

## Objectives

- Make the E2E rename tests actually verify the initial value of the dialog's input field.

## User Stories

N/A

## Technical Requirements

### Functional Requirements

- **FR1:** Restrict the initial-value check in `test_rename_file` to the input field.
  The initial-value check in `test_rename_file` in `test/e2e/scripts/tests/file_operation_tests.sh` (line 195, `assert_contains "before_rename"`) is performed in a form that can only appear on the input field line. When the input field's initial value is not the base name `before_rename` (empty, or the full name with extension such as `before_rename.txt`), the check fails. `before_rename.txt` being shown in the file list behind the dialog is not, by itself, enough for the check to pass.
- **FR2:** Restrict the initial-value check in `test_navigation_after_rename` to the input field.
  The initial-value check in `test_navigation_after_rename` in the same file (line 439, `assert_contains "navren_before"`) is performed, using the same method as FR1, in a form that can only appear on the input field line. When the input field's initial value is not the base name `navren_before`, the check fails.
- **FR3:** Remove the always-passing `testdata` check from `history_tests.sh`.
  Remove `assert_contains "testdata"` at lines 167-168 of `test_history_forward_cleared` in `test/e2e/scripts/tests/history_tests.sh`. The `assert_not_contains "/testdata/dir1"` that immediately follows is kept.

### Non-Functional Requirements

- **NFR1:** Go source files (`*.go`) are not changed.
- **NFR2:** The existing behavior of `assert_contains` / `assert_not_contains` in `helpers.sh` is not changed (all E2E tests use them).

## Acceptance Criteria

- [ ] AC-1 (FR1, FR2): The initial-value checks in `test_rename_file` and `test_navigation_after_rename` are performed in a form that can only appear on the input field line.
- [ ] AC-2 (FR1, FR2): When the input field's initial value differs from the expected base name (e.g. when the input field is cleared before the check), the initial-value checks in both tests fail.
- [ ] AC-3 (FR3): The always-passing `assert_contains "testdata"` (line 167) in `history_tests.sh` is gone, and `assert_not_contains "/testdata/dir1"` remains.
- [ ] AC-4 (FR1, FR2, FR3): All file-ops and history tests pass under `make test-e2e`.
- [ ] AC-5 (NFR1): The change contains no `*.go` files.

## Implementation Approach

### Architecture

N/A (changes are limited to E2E test shell scripts).

### Assumptions

- **A-1:** The extension-preserving rename dialog draws the input field surrounded by a border and shows the extension (`.txt`) separately to the right of the input field (`renderInputFieldWithExtension` in `internal/ui/extension_rename_dialog.go`). The input field contains only the base name.
- **A-2:** How the text input component's cursor rendering (`RenderWithCursor`) appears in the screen capture has not been confirmed. The match pattern is made independent of how the cursor appears.
- **A-3:** For line 167 of `history_tests.sh`, "delete" is adopted rather than "narrow to the path display line", because the subsequent name search for `dir2` and `assert_contains "/testdata/dir2"` confirm that the view had returned to `/testdata`.
- **A-4:** No test functions are added, removed, or renamed, so the run lists in `run_all_tests.sh` / `run_tests.sh` do not change.

### API Design

N/A

### Database Schema

N/A

### Dependencies

N/A

## Declared Change Set

Feature-specific paths:

- `test/e2e/scripts/tests/file_operation_tests.sh`
- `test/e2e/scripts/tests/history_tests.sh`

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/e2e-rename-initial-value-check/**`
- `test-docs/e2e-rename-initial-value-check/**`

`feature-docs/e2e-rename-initial-value-check/**` covers `REQUIREMENTS.md`,
`SPEC.md`, `IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/e2e-rename-initial-value-check/**` covers
`test-docs/e2e-rename-initial-value-check/{T}.tests.yaml`, the per-task test
record. It is generated and owned by `implement-phase.md`; this section cites
it and restates none of its rules.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A declared path that never materializes is not a violation.

## Test Scenarios

### Unit Tests

N/A (covered by TS-4 regression run of `make test`).

### Integration Tests

N/A

### E2E Tests

**Existing E2E tests**: `test/e2e/scripts/tests/file_operation_tests.sh`, `test/e2e/scripts/tests/history_tests.sh`
**Run command**: `make test-e2e`

- [ ] Existing E2E tests pass without regression
- [ ] **TS-1** (AC-1, AC-4): The fixed rename tests pass.
  - Steps: Run `make test-e2e` (running `/e2e/scripts/run_all_tests.sh file-ops` inside the container is also acceptable).
  - Expected: All checks in `test_rename_file` and `test_navigation_after_rename` pass.
- [ ] **TS-2** (AC-2): Confirm the new check depends on the input field's content.
  - Steps: With a local, uncommitted change, send `C-u` immediately before the initial-value check to clear the input field, then run the file-ops tests.
  - Expected: Even though `before_rename.txt` / `navren_before.txt` are shown in the file list, the initial-value checks in both tests fail.
- [ ] **TS-3** (AC-3, AC-4): The history tests pass.
  - Steps: Run `make test-e2e` (running `/e2e/scripts/run_all_tests.sh history` inside the container is also acceptable).
  - Expected: All history tests, including `test_history_forward_cleared`, pass, and the `testdata` check at line 167 is absent.
- [ ] **TS-4** (AC-5): Confirm Go sources are unchanged.
  - Steps: Check whether the diff against the base contains any `*.go` file, and run `make test`.
  - Expected: No `*.go` diff, and unit tests pass.

### Edge Cases

- [ ] The input field's initial value is empty: the check fails.
- [ ] The input field's initial value is the full name with extension (`before_rename.txt` / `navren_before.txt`): the check fails.
- [ ] Whether the file list rows behind the dialog are visible or hidden, the check result is determined only by the input field's content.
- [ ] Even if the search filter display (`/before_ren`, `/navren_before`) remains on screen while the dialog is shown, the check must not pass by matching it.

## Security Considerations

N/A

## Error Handling

N/A

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All test scenarios pass
- [ ] Code review is completed

## Open Questions

None.

## References

- `test/e2e/scripts/tests/file_operation_tests.sh`
- `test/e2e/scripts/tests/history_tests.sh`
- `internal/ui/extension_rename_dialog.go`
