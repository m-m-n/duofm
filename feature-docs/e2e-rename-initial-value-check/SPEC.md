# Feature: e2e-rename-initial-value-check

## Overview

The E2E rename tests check the initial value of the rename dialog's input field. This feature makes those checks match only the input field's row and require the input field's content to be exactly the expected base name, so that they fail whenever the input field does not hold exactly that base name. It also removes an always-passing `testdata` check from `history_tests.sh`. Requirements document: `feature-docs/e2e-rename-initial-value-check/REQUIREMENTS.md`.

## Objectives

- Make the E2E rename tests actually verify the initial value of the rename dialog's input field, so that the checks fail whenever the input field does not hold exactly the expected base name.
- Remove the always-passing `testdata` check from `history_tests.sh`.

## User Stories

N/A

## Technical Requirements

### Functional Requirements

- **FR1:** Restrict the initial-value check in `test_rename_file` to the input field, with strict matching.
  The initial-value check in `test_rename_file` in `test/e2e/scripts/tests/file_operation_tests.sh` (originally `assert_contains "before_rename"` at line 195) is performed in a form that can only match the input field's row, and it requires the input field's content to be exactly the base name `before_rename`. Between the input field's left and right border characters (`│`) there may be only whitespace besides the base name, and the fixed extension `.txt` follows the right border. The check fails when the input field's content is empty, when it is the full name with extension (`before_rename.txt`), or when there is any extra character before or after the base name, including a single non-whitespace character (e.g. `before_rename.`, `xbefore_rename`, `before_renamex`). `before_rename.txt` in the file list behind the dialog, or `/before_ren` in the search filter display, is not by itself enough for the check to pass.
- **FR2:** Restrict the initial-value check in `test_navigation_after_rename` to the input field, with strict matching.
  The initial-value check in `test_navigation_after_rename` in the same file (originally `assert_contains "navren_before"` at line 439) is performed with the same method as FR1, and requires the input field's content to be exactly the base name `navren_before`. The check fails when the input field's content is empty, is the full name `navren_before.txt`, or has any extra character before or after the base name, including a single non-whitespace character (e.g. `navren_before.`, `xnavren_before`, `navren_beforex`). `navren_before.txt` in the file list or `/navren_before` in the search filter display is not by itself enough for the check to pass.
- **FR3:** Remove the always-passing `testdata` check from `history_tests.sh`.
  Remove `assert_contains "testdata"` at lines 167-168 of `test_history_forward_cleared` in `test/e2e/scripts/tests/history_tests.sh`. The `assert_not_contains "/testdata/dir1"` that immediately follows is kept.

### Non-Functional Requirements

- **NFR1:** Go source files (`*.go`) are not changed.
- **NFR2:** The existing behavior of `assert_contains` / `assert_not_contains` in `test/e2e/scripts/helpers.sh` is not changed (all E2E tests use them).

## Acceptance Criteria

- [ ] AC-1 (FR1, FR2): The initial-value checks in `test_rename_file` and `test_navigation_after_rename` are performed in a form that can only match the input field's row.
- [ ] AC-2 (FR1, FR2): When the input field's initial value differs from the expected base name (e.g. the input field is cleared before the check, or holds the full name with extension), the initial-value checks in both tests fail.
- [ ] AC-3 (FR3): The always-passing `assert_contains "testdata"` (line 167) in `history_tests.sh` is gone, and `assert_not_contains "/testdata/dir1"` remains.
- [ ] AC-4 (FR1, FR2, FR3): All file-ops and history tests pass under `make test-e2e`.
- [ ] AC-5 (NFR1): The change contains no `*.go` files.
- [ ] AC-6 (FR1, FR2): The initial-value match is strict: it allows no extra character (including a single non-whitespace character) before or after the base name inside the input field. The check in `test_rename_file` fails when the input field shows `before_rename.`, `xbefore_rename`, or `before_renamex`, and the check in `test_navigation_after_rename` fails when the input field shows `navren_before.`, `xnavren_before`, or `navren_beforex`.

## Implementation Approach

### Architecture

N/A (changes are limited to E2E test shell scripts).

### Assumptions

- **A-1:** The extension-preserving rename dialog draws the input field surrounded by a rounded border (`│` on the input row) and shows the extension (`.txt`) separately to the right of the input field (`renderInputFieldWithExtension` in `internal/ui/extension_rename_dialog.go`). The input field contains only the base name, with one column of horizontal padding.
- **A-2:** `TextInput.RenderWithCursor` (`internal/ui/text_input.go`) renders the cursor only through the reverse-video attribute: a cursor on a character shows that character itself, and a cursor at the end of the value is a reverse-video space. No cursor glyph character is inserted. `tmux capture-pane -p` (without `-e`) drops attributes, so the strict match needs no allowance for a cursor glyph next to the base name.
- **A-3:** For line 167 of `history_tests.sh`, "delete" is adopted rather than "narrow to the path display line", because the subsequent name search for `dir2` and `assert_contains "/testdata/dir2"` confirm that the view had returned to `/testdata`.
- **A-4:** No test functions are added, removed, or renamed, so the run lists in `run_all_tests.sh` / `run_tests.sh` do not change.
- **A-5:** The `│` border character appears in the E2E container's screen capture of the rename dialog. This is established by the previous verify run, in which the border-anchored check passed under `make test-e2e`.
- **A-6:** Extra whitespace-only characters around the base name inside the input field are outside the strict match's scope: in the text capture they cannot be distinguished from the field's padding and the end-of-value cursor cell. The strict match rejects every extra non-whitespace character.

### API Design

N/A

### Database Schema

N/A

### Dependencies

N/A

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

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

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/e2e-rename-initial-value-check/` directory at all; the declared
`test-docs/e2e-rename-initial-value-check/**` entry is still correct in that
case — a declared path that never materializes is not a violation.

## Test Scenarios

### Unit Tests

N/A (covered by the TS-4 regression run of `make test`).

### Integration Tests

N/A

### E2E Tests

**Existing E2E tests**: `test/e2e/scripts/tests/file_operation_tests.sh`, `test/e2e/scripts/tests/history_tests.sh`
**Run command**: `make test-e2e`

- [ ] Existing E2E tests pass without regression
- [ ] **TS-1** (AC-1, AC-4): The fixed rename tests pass.
  - Steps: Run `make test-e2e` (running `/e2e/scripts/run_all_tests.sh file-ops` inside the container is also acceptable).
  - Expected: All checks in `test_rename_file` and `test_navigation_after_rename` pass.
- [ ] **TS-2** (AC-2): The new check depends on the input field's content.
  - Steps: With a local, uncommitted change, (a) send `C-u` immediately before the initial-value check to clear the input field, and (b) separately, make the input field hold the full name with extension (e.g. send `End` `.` `t` `x` `t`), then run the file-ops tests for each variant.
  - Expected: Even though `before_rename.txt` / `navren_before.txt` are shown in the file list, the initial-value checks in both tests fail in each variant.
- [ ] **TS-3** (AC-3, AC-4): The history tests pass.
  - Steps: Run `make test-e2e` (running `/e2e/scripts/run_all_tests.sh history` inside the container is also acceptable).
  - Expected: All history tests, including `test_history_forward_cleared`, pass, and the `testdata` check at line 167 is absent.
- [ ] **TS-4** (AC-5, regression): Go sources are unchanged.
  - Steps: Check whether the diff against the base contains any `*.go` file, and run `make test`.
  - Expected: No `*.go` diff, and unit tests pass.
- [ ] **TS-5** (AC-6): Strict match rejects one extra leading or trailing character.
  - Steps: With a local, uncommitted change, immediately before the initial-value check in each test, send keys that add one extra character to the pre-filled value, one variant per run: (a) `End` `.` -> `before_rename.` / `navren_before.`; (b) `Home` `x` -> `xbefore_rename` / `xnavren_before`; (c) `End` `x` -> `before_renamex` / `navren_beforex`. Run the file-ops tests for each variant.
  - Expected: In every variant, the initial-value check fails in both `test_rename_file` and `test_navigation_after_rename` (the results of later checks in the mutated runs are not evaluated).

### Edge Cases

- [ ] The input field's initial value is empty: the check fails.
- [ ] The input field's initial value is the full name with extension (`before_rename.txt` / `navren_before.txt`): the check fails.
- [ ] The input field's value has one extra non-whitespace character before or after the base name (`before_rename.`, `xbefore_rename`, `before_renamex`, and the same for `navren_before`): the check fails.
- [ ] Whether the file list rows behind the dialog are visible or hidden, the check result is determined only by the input field's content.
- [ ] Even if the search filter display (`/before_ren`, `/navren_before`) remains on screen while the dialog is shown, the check must not pass by matching it.
- [ ] The cursor at the end of the value is rendered as a reverse-video space; in the text capture it is indistinguishable from padding whitespace and must not cause the check to fail.

### Performance Tests

N/A

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

- Requirements: `feature-docs/e2e-rename-initial-value-check/REQUIREMENTS.md`
- `test/e2e/scripts/tests/file_operation_tests.sh`
- `test/e2e/scripts/tests/history_tests.sh`
- `test/e2e/scripts/helpers.sh`
- `internal/ui/extension_rename_dialog.go`
- `internal/ui/text_input.go`
