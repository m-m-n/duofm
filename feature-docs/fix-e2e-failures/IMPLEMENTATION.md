# Implementation Plan: fix-e2e-failures

## Overview

Bring the E2E shell tests in line with current app behavior so the full run reports Failed: 0, and make the full-run runner detect run-list entries with no defined function and defined tests missing from the run list. Only E2E shell scripts change.

## Technology Stack

- **Language**: Bash (existing E2E scripts)
- **Test harness**: existing tmux-driven helpers in `test/e2e/scripts/helpers.sh` (unchanged)
- **Execution environment**: existing image built from `test/e2e/Dockerfile` (unchanged), entered through `test/e2e/scripts/run_tests.sh` (tracked symlink to `run_all_tests.sh`, unchanged)
- **New dependencies**: none (no license impact; project license MIT)

## Layer Structure

| Layer | Files | Responsibility | May depend on |
|-------|-------|----------------|---------------|
| Runner | `test/e2e/scripts/run_all_tests.sh` | Sources all test files, owns the explicit full-run list, the run-list guards, and the final summary / exit status | helpers, test function names (by contract below) |
| Helpers | `test/e2e/scripts/helpers.sh` | tmux session control, screen assertions, counters, `run_test`, `print_summary` | nothing |
| Test files | `test/e2e/scripts/tests/*.sh` | One function per test case | helpers only |

Test functions never depend on the runner, on each other, or on files another test leaves behind (NFR4).

## Shared Components

| Component | Responsibility | Contract (pre/postcondition) | Used by tasks |
|-----------|----------------|------------------------------|---------------|
| Test function name set | The entry points the runner calls | Pre: at the base commit, `tests/*.sh` defines 120 top-level functions whose names start with `test_` (per file: basic 9, directory 10, file_operation 13, copy_move 8, mark 8, sort 14, cursor 8, shell 6, config 5, bookmark 3, history 8, archive 9, background 5, cursor_preserve 14). Post: after every task, the set of top-level `test_`-prefixed function names in `tests/*.sh` is exactly the same 120 names. No task renames, removes, or adds a `test_`-prefixed function. Any helper function a test file needs uses a name that does not start with `test_`. | task0001 (builds the run list against this set), task0002, task0003, task0004, task0005 (must keep it) |
| Helpers API (existing, unchanged) | Screen assertions, session control, counters (`TESTS_RUN`, `TESTS_PASSED`, `TESTS_FAILED`), `run_test`, `print_summary` | Called as-is. No task modifies `helpers.sh`. | all tasks |
| Inline check convention | Filesystem checks that the screen helpers cannot express (file exists / removed / archive content) | Each check increments `TESTS_RUN` by 1 and exactly one of `TESTS_PASSED` / `TESTS_FAILED` by 1, and prints one line with the check description (pass mark or fail mark, same format as the helpers). Postcondition per test: the increase of `TESTS_RUN` equals the increase of `TESTS_PASSED` + `TESTS_FAILED`. No branch increments `TESTS_RUN` alone. | task0001 (guard failures), task0002, task0003, task0004, task0005 |
| Writable fixture area `/testdata/user_owned` | The only fixture directory the test user can write | Each edited test creates only names unique to that test, removes those exact names before it starts and again at its end whatever the assertion outcomes, never deletes with wildcards, and never touches the image-provided `deletable.txt`. | task0002, task0003 |
| Full-run outcome | Summary line counts and exit status of `run_all_tests.sh` with no argument | `Failed` includes every failing assertion plus every guard failure. Exit status is non-zero exactly when `Failed` > 0. | task0001 (producer); verification (consumer) |

## Conventions

- **Screen literals**: every string a test expects on screen must appear in the current `internal/ui` rendering source at the base commit. Tests are not run inside implementation worktrees (no Docker there), so this traceability is the implementer's check.
- **Locate by name, not position**: tests reach a file or directory by its name (search filter), never by counting rows, because `/testdata/user_owned` contents depend on what earlier tests did.
- **No always-pass checks in edited tests (FR9)**: a check must be able to fail when the behavior it names is wrong (for example, asserting a bare `!` or a word that appears elsewhere on screen does not count).
- **Out of bounds for every task**: Go sources, `test/e2e/Dockerfile`, `test/e2e/scripts/run_tests.sh` (must stay a symlink), `test/e2e/scripts/helpers.sh`, `Makefile`.
- **Static checks for every modified script**: bash syntax check passes. If shellcheck is installed, the lines the task touched introduce no new findings.

## Cross-task Design Decisions

### D1: Runner fixes live only in `run_all_tests.sh`
The undefined-name and missing-test guards are implemented in the runner, not by changing `run_test` in `helpers.sh` (NFR2). Test files keep using the helpers unchanged. Affected: task0001 (implements), all others (must not touch helpers).

### D2: Frozen test function names
Tasks run in parallel, and the runner's explicit list is written against the current names. Keeping the name set fixed (Shared Components) lets task0001 write the final list without seeing the other tasks' edits. Semantics may change while names stay (for example, the two sort "q_cancel" tests now assert that `q` does NOT close the dialog). Affected: all tasks.

### D3: Newly listed tests are audited by the task that owns their file
The 9 tests added to the run list have never run. The task that edits each file checks them against current app behavior and corrects mismatches without renaming: the 2 delete-confirmation tests in `file_operation_tests.sh` → task0003; the 7 unlisted sort tests in `sort_tests.sh` → task0004 (`test_sort_dialog_tab_navigation` has no assertions and stays that way under A5 unless the task has to edit it). task0001 only lists them.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| E2E cannot run in implementation worktrees, so key sequences and screen literals are checked only against source | High | Medium | Screen-literal convention; full E2E run in the verify phase; remaining failures become rework tasks |
| Newly listed tests fail on first run | Medium | Medium | D3 audit by file-owning task; verify-phase full run |
| A parallel task changes a `test_` name and the run list no longer matches | Low | High | D2 contract; runner guard and check-only mode flag any mismatch at verify |
| Files left in `/testdata/user_owned` shift row positions for later tests (mark, cursor-preserve) | Medium | Medium | Fixture-area contract; name-based navigation |

## Open Questions

- [ ] The E2E command (`make test-e2e-build && make test-e2e`) needs Docker, which is not available in the current environment. The TS1–TS5 and TS10–TS12 runs need an environment with Docker.
- [ ] `test_sort_dialog_tab_navigation` joins the run list with no assertions. It passes without checking anything. A5 keeps it out of scope.
