# Feature: fix-e2e-failures

## Overview

Update the E2E scripts so that the full E2E run passes against current app behavior, and make the runner detect test functions that never run and run-list entries with no defined function. Only E2E scripts change; no Go source changes. Requirements document: `feature-docs/fix-e2e-failures/REQUIREMENTS.md`.

## Objectives

- `make test-e2e-build && make test-e2e` (full run through run_all_tests.sh) reports Failed: 0.
- The E2E run catches this problem if it happens again: test functions that never run, and run-list entries with no defined function.

## User Stories

N/A

## Technical Requirements

### Functional Requirements

- **FR1:** Archive tests follow the current context-menu behavior. The 9 archive tests (test_compress_format_dialog_opens, test_compress_format_navigation, test_compression_level_dialog, test_archive_name_dialog, test_archive_conflict_dialog, test_compress_cancel_workflow, test_compress_complete_workflow, test_extract_complete_workflow, test_multifile_compress) open the context menu with '@' and choose Compress or Extract archive with its number key, with no extra Enter. Tests that check output in /testdata/user_owned first point the destination pane at user_owned.
- **FR2:** Rename tests follow the extension-preserving dialog. test_rename_file and test_navigation_after_rename work with the pre-filled rename dialog that keeps the extension, and check that after_rename.txt / navren_after.txt exist.
- **FR3:** Batch delete confirms with 'y'. test_batch_delete_marked_files confirms the deletion with 'y' and checks that the marked files are gone.
- **FR4:** Sort dialog ignores q. test_sort_dialog_q_cancel and test_sort_dialog_q_cancel_with_dropdown assert that pressing 'q' does NOT close the sort dialog: once in the normal state, and once with the dropdown open. The app code does not change.
- **FR5:** Help shell-command entry check. test_help_shows_shell_command scrolls or pages to the page with the '!' entry and checks the entry's text ('execute shell command').
- **FR6:** Bookmark hint text. test_bookmark_dialog_opens checks for 'd:Delete'.
- **FR7:** History forward-cleared path. test_history_forward_cleared moves only through directories it can enter when testing that forward history is cleared.
- **FR8:** Runner run list and coverage guards. run_all_tests.sh keeps its explicit run list. The list drops the 3 undefined names (test_sort_dialog_hl_navigation, test_sort_dialog_jk_navigation, test_sort_dialog_confirm) and adds the 9 defined tests that were never run. Calling an undefined test function counts as a failure. The run fails when a test_ function defined in test/e2e/scripts/tests/*.sh is missing from the run list.
- **FR9:** Assertion and cleanup hardening in edited tests. In the tests this feature edits, assertions that always pass are made strict, and each edited test deletes the files it creates.

### Non-Functional Requirements

- **NFR1 - No Go source changes:** The fix changes only the E2E scripts. No Go source changes.
- **NFR2 - Dockerfile and run_tests.sh untouched:** test/e2e/Dockerfile and test/e2e/scripts/run_tests.sh (a tracked symlink to run_all_tests.sh) are not changed. Runner fixes go into run_all_tests.sh only.
- **NFR3 - Unit tests keep passing:** `go test ./...` keeps passing.
- **NFR4 - Test isolation:** Edited E2E tests do not rely on files that other tests leave behind.

## Implementation Approach

### Architecture

N/A (changes are limited to E2E test scripts and the runner; no UI change, design step skipped)

### Data Flow

N/A

### API Design

N/A

### Database Schema

N/A

### Dependencies

**Internal Dependencies:**
- doc/tasks/cancel-key-unification/SPEC.md FR2: removes q from SortDialog (FR4)

**External Dependencies:**
- None

### File Structure

```
test/e2e/scripts/
├── run_all_tests.sh     # FR8 (run list, undefined-function and missing-test guards)
├── run_tests.sh         # tracked symlink to run_all_tests.sh; not changed (NFR2)
└── tests/*.sh           # FR1-FR7, FR9 (edited tests)
test/e2e/Dockerfile      # not changed (NFR2)
```

## Declared Change Set

This section states the create-plan derivation instead of a hand-authored
list: the feature-specific paths above are derived at create-plan from
every task's `files` entries in `workflow.yaml`
(`references/phases/create-plan-phase.md`).

Every SPEC declares, by default, the following two workflow-generated
entries in addition to the feature-specific paths above:

- `feature-docs/fix-e2e-failures/**`
- `test-docs/fix-e2e-failures/**`

`feature-docs/fix-e2e-failures/**` covers `REQUIREMENTS.md`, `SPEC.md`,
`IMPLEMENTATION.md`, `workflow.yaml`, `phase-state/`, `tasks/`,
`reviews/roundN.yaml`, `VERIFICATION.md`, `retrospect.yaml`, and the design
artifacts the design step produces. These are generated and owned by the
phase documents and by `references/phase-state.md`; this section cites them
and restates none of their rules.

`test-docs/fix-e2e-failures/**` covers `test-docs/fix-e2e-failures/{T}.tests.yaml`, the
per-task test record. It is generated and owned by `implement-phase.md`;
this section cites it and restates none of its rules.

These two default entries are part of the declaration unless the SPEC
author explicitly removes them; their absence is never assumed by
silence — removal is a deliberate, explicit narrowing.

This declaration is a SUPERSET assertion: the actual change set observed
at verification time must be CONTAINED IN the declared set, not equal to
it. A feature that produces no implement tasks generates no
`test-docs/fix-e2e-failures/` directory at all; the declared
`test-docs/fix-e2e-failures/**` entry is still correct in that case — a declared
path that never materializes is not a violation.

## Test Scenarios

### Unit Tests
- [ ] `go test ./...` passes (NFR3)

### Integration Tests
N/A

### E2E Tests
**Existing E2E tests**: test/e2e/scripts/run_all_tests.sh, test/e2e/scripts/tests/*.sh
**Run command**: `make test-e2e-build && make test-e2e`
- [ ] Existing E2E tests pass without regression
- [ ] TS1 (FR1, FR2, FR3, FR4, FR5, FR6, FR7): Run the full E2E suite. Every previously failing test in the Archive, File Operation, Mark, Sort, Shell, Bookmark and History groups passes.
- [ ] TS2 (FR8): Temporarily add an undefined name to the run list. The run counts it as a failure.
- [ ] TS3 (FR8): Temporarily remove a defined test from the run list. The run fails and names the missing test.
- [ ] TS4 (FR4): Sort dialog open, press 'q': dialog still shows. Dropdown open, press 'q': dialog still shows.
- [ ] TS5 (FR9, NFR4): After the full run, /testdata/user_owned has none of the archive or rename output files.

### Edge Cases
- [ ] The compress format dialog's numbering comes from archive.GetAvailableFormats(), which depends on the tools installed in the image. The tests assume 2 = tar.gz and 5 = zip. This was not checked against internal/archive.
- [ ] If zip is unavailable, test_multifile_compress increments TESTS_RUN without counting a pass or a failure, so Total/Passed/Failed do not add up.
- [ ] Context-menu numbers shift if menu items are added. Disabled items (Open, Open with in a container without a desktop) still take a number.
- [ ] Tests that were not run before may fail once added to the list. test_sort_dialog_tab_navigation has no assertions.
- [ ] Leftover rename files sort before del1.txt and break the position-based steps in test_batch_delete_marked_files.
- [ ] If the panes are not synced, compress output goes to /testdata (not writable) and the result is an error, not an archive.

### Performance Tests
N/A

## Security Considerations

N/A

## Error Handling

N/A

## Performance Optimization

N/A

## Success Criteria

- [ ] AC1 (FR1, FR2, FR3, FR4, FR5, FR6, FR7, FR8): The full run `make test-e2e-build && make test-e2e` reports Failed: 0.
- [ ] AC2 (FR8): The test summary includes the 9 tests that did not run before.
- [ ] AC3 (FR8): When an undefined test function name is in the run list, the run reports a failure.
- [ ] AC4 (FR8): When a defined test_ function in tests/*.sh is missing from the run list, the run fails.
- [ ] AC5 (FR4): Both q tests pass only while the sort dialog stays open after 'q', in the normal state and with the dropdown open.
- [ ] AC6 (FR1, FR2, FR9): After the full run, the files the archive and rename tests created are gone from /testdata/user_owned.
- [ ] AC7 (NFR1, NFR2, NFR3): The diff has no Go source changes and no changes to test/e2e/Dockerfile or test/e2e/scripts/run_tests.sh, and `go test ./...` passes.

## Assumptions

- **A1** (requirement.fix-direction): The fix updates the E2E scripts to match current app behavior. No Go changes (update_tests_only).
- **A2** (requirement.sort-q-cancel): Both q tests are rewritten to assert that 'q' does not close the sort dialog (invert_q_tests). doc/tasks/cancel-key-unification/SPEC.md FR2 removes q from SortDialog on purpose.
- **A3** (requirement.runner-coverage): The explicit run list stays and is corrected. Undefined-function calls count as failures. Defined tests missing from the list fail the run (fix_list_and_guard).
- **A4** (requirement.entrypoint): test/e2e/Dockerfile and test/e2e/scripts/run_tests.sh are not changed (leave_to_mouse_support). run_tests.sh is a tracked symlink (mode 120000) to run_all_tests.sh, so the current CMD already runs run_all_tests.sh.
- **A5** (requirement.assertion-scope): Assertion hardening and file cleanup cover only the tests this feature edits (touched_tests_only).

A1-A5 are batch answers recorded as assumptions.

## Open Questions

> **Note**: 未解決の要件は workflow.yaml で `status: tbd` として管理されています。
> plan フェーズの実行前に解決してください。

- None

## References

- Requirements: feature-docs/fix-e2e-failures/REQUIREMENTS.md
- doc/tasks/cancel-key-unification/SPEC.md
