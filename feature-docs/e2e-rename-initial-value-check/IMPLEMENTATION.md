# Implementation Plan: e2e-rename-initial-value-check

## Overview

The feature changes only E2E test shell scripts. The two rename tests check the
rename dialog's input-field initial value through one strict assertion that
only the input field's row can satisfy, and the always-passing `testdata`
check is removed from `history_tests.sh`. This document pins the decisions
shared by every task that touches these checks.

## Technology Stack

- **Language**: Bash (E2E test scripts under `test/e2e/scripts/`)
- **Runtime**: the E2E container, which drives duofm inside tmux and reads the
  screen as plain text (colors and reverse video are dropped from the capture)
- **Key libraries**: none added. The feature introduces no new dependency, so
  there is no dependency license to record (project license: MIT).

## Layer Structure

| Layer | Files | Responsibility | Change policy |
|-------|-------|----------------|---------------|
| Runner | `test/e2e/scripts/run_all_tests.sh`, `test/e2e/scripts/run_tests.sh` | Treats every line starting with `test_` in `tests/*.sh` as a test definition; holds the run lists and the `--check-list` guard | Not changed (SPEC A-4) |
| Shared assertions | `test/e2e/scripts/helpers.sh` | Session control, screen capture, the run / passed / failed counters, `assert_contains` / `assert_not_contains` and the other shared assertions | Not changed (NFR2) |
| Test categories | `test/e2e/scripts/tests/*.sh` | Test functions per category; may define file-local helpers | Only `file_operation_tests.sh` and `history_tests.sh` change |
| Product | `*.go` | Renders the rename dialog that the capture observes | Not changed (NFR1) |

Dependency direction: runner → test categories → shared assertions. The
product is observed only through the screen capture. A full run sources every
test-category file into one shell, so every name defined in any of them
shares one namespace.

## Shared Components

| Component | Responsibility | Contract (pre/postcondition) | Used by tasks |
|-----------|----------------|------------------------------|---------------|
| Input-field value assertion `assert_input_field_value`, file-local in `test/e2e/scripts/tests/file_operation_tests.sh` | Checks that the extension-preserving rename dialog's input field holds exactly the expected base name | See "Input-field value assertion contract" below | task0001 (built it); any later task that changes the initial-value checks |

### Input-field value assertion contract

| Item | Contract |
|------|----------|
| Inputs (positional, in this order) | session identifier; expected base name; fixed extension; human-readable description |
| Precondition | The extension-preserving rename dialog is open in the session |
| Capture | The screen is captured once. Each captured row is evaluated on its own; a match never spans two rows |
| Pass condition | At least one row contains, in this order and with nothing else in between: (1) a vertical border character `│`; (2) zero or more whitespace characters; (3) the expected base name; (4) zero or more whitespace characters; (5) a vertical border character `│`; (6) zero or more whitespace characters; (7) the extension |
| Strictness | Between the two border characters, the only characters besides the base name are whitespace. No non-whitespace character next to the base name is tolerated, including a single one on either side (SPEC FR1, FR2, AC-6). There is no cursor-glyph allowance: the cursor is drawn by reverse video only, and the capture drops attributes (SPEC A-2) |
| Whitespace scope | Extra whitespace-only characters around the base name are outside the check's scope; they cannot be told apart from the field's padding or the end-of-value cursor cell (SPEC A-6) |
| Literal matching | The base name and the extension are matched as literal text. No character in them carries pattern meaning (the `.` of `.txt` matches only a period) |
| Locale independence | The result is the same under the E2E container's default locale and under a UTF-8 locale. The border character is multi-byte and is matched as literal text; the pass condition has no "any single character" element |
| Required outcomes | Passes when the input field holds exactly the base name (the end-of-value cursor appears as a space). Fails when the input field is empty, holds the full name with extension, or has one or more extra non-whitespace characters before or after the base name (for example `before_rename.`, `xbefore_rename`, `before_renamex`). Never satisfied by the file-list row (`before_rename.txt`), the dialog title row, or the search-filter display (`/before_ren`, `/navren_before`) |
| Counting | Exactly one check per call: the run counter increases by one, and exactly one of the passed / failed counters increases by one |
| Output | One pass / fail line with the description, in the same style as the assertions in `helpers.sh`. On failure it also prints the expected base name, the extension, and the whole captured screen |
| Return status | Success on pass, failure on fail |
| Call sites | `test_rename_file` with base name `before_rename` and extension `.txt`; `test_navigation_after_rename` with base name `navren_before` and extension `.txt`. Both use this one assertion (FR2 requires the same method as FR1) |

## Conventions

- **File-local helper names**: a helper defined in `tests/*.sh` never starts
  with `test_` (the runner would treat it as a test) and never reuses a name
  defined in `helpers.sh` or in another `tests/*.sh` file.
- **Frozen files**: no task in this feature edits `helpers.sh`, the runner
  scripts, or any `*.go` file. No test function is added, removed, or renamed,
  so the run lists and the `# Tests:` count headers stay as they are.
- **Assertion behavior**: every check updates the shared counters exactly once
  and prints one pass / fail line in the `helpers.sh` style.
- **Mutation verification**: for a change to test code, the red evidence is a
  mutation run — a temporary, uncommitted edit that makes the input field hold
  a wrong value immediately before the initial-value check — and the green
  evidence is the unmodified run. Mutations are reverted after the run and
  never committed. Their results (which checks failed) are recorded in the
  implementation report. In a mutated run, only the initial-value checks'
  results are evaluated.

## Cross-task Design Decisions

### D1: The input-field assertion is file-local

- **Decision**: the assertion lives in `file_operation_tests.sh`, not in
  `helpers.sh`.
- **Rationale**: NFR2.
- **Affected tasks**: task0001 and any later task that changes the
  initial-value checks.

### D2: Strict, row-anchored match with a whitespace-only allowance

- **Decision**: the pass condition in the contract above. Whitespace is the
  only character tolerated between the borders besides the base name.
- **Rationale**: SPEC FR1, FR2, AC-2, AC-6, A-2, A-6.
- **Affected tasks**: task0001 built the assertion with at most one arbitrary
  non-whitespace character tolerated on each side of the base name (a
  cursor-glyph allowance). This contract supersedes that allowance. The merged
  assertion still carries it, which is the verify failure AC-2; bringing it in
  line with this contract is outside this plan.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| The border character does not appear in the container capture | Low | High: the initial-value checks always fail | Established by the previous verify run (SPEC A-5); TS-1 confirms it again |
| Another row reproduces the anchored structure (pane borders, a truncated file-list name) | Low | Medium: the check passes on a wrong value | The structure requires the extension right after the right border, which only the dialog's input row produces; the TS-2 and TS-5 mutation runs confirm it |
| The match result depends on the container locale because the border character is multi-byte | Low | Medium | Literal matching only, with no "any single character" element (contract: Locale independence) |
| The merged assertion accepts one extra non-whitespace character (verify failure AC-2) | Occurred | High: AC-2 and AC-6 unmet | Brought in line with the contract above outside this plan; TS-5 verifies it |

## Open Questions

- [ ] The change that makes the merged assertion satisfy AC-2 / AC-6 has no task in this plan; it is synthesized separately.
