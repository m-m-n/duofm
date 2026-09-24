# Verification Document: Mouse Support (mouse-support)

## Overview

**Feature**: mouse-support (left-button click / drag / double-click on the file panes)
/ **SPEC.md**: `feature-docs/mouse-support/SPEC.md`
/ **IMPLEMENTATION.md**: `feature-docs/mouse-support/IMPLEMENTATION.md`

This document covers the integrated verification run on the integration branch
after both tasks have merged. Each task's own acceptance criteria are in
`tasks/taskNNNN.md`.

## Build Verification

- Command: `workflow.yaml` → `project.components.duofm.build_command`. It
  runs `make build` inside the `golang:1.25.5-trixie` container.
- Expected: exit code 0, no errors.

## Test Verification

- Command: `workflow.yaml` → `project.components.duofm.test_command`. It
  runs `go test ./...` inside the same container.
- Expected: exit code 0, and every test passes.
- Coverage target for the new source files (`internal/ui/mouse_hittest.go`,
  `internal/ui/pane_drag.go`, `internal/ui/mouse_doubleclick.go`,
  `internal/ui/model_update_mouse.go`): minimum 80%, target 90%. Coverage
  of the package `internal/ui` must not fall below its current level.

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type | Tasks |
|----|----------|-----------------|-----------|-------|
| TS-1 | Hit test of entry rows in the left and right panes (80×24, scroll offset 0) | (5,4) → left index 0; (39,6) → left index 2; (40,4) → right index 0 | Unit | task0001 |
| TS-2 | Hit test with a non-zero scroll offset (10) | Y=4 → index 10; Y=4+V−1 → 10+V−1; Y=4+V is not an entry row | Unit | task0001 |
| TS-3 | Hit test of non-entry areas (81×24, 3 entries, a pane showing `(No matches)`) | Title bar, status bar and X=80 → no pane; header rows, blank rows and the `(No matches)` row → that pane's non-entry area | Unit | task0001 |
| TS-4 | Hit test in the bg output split view | File-list rows map to entries by the file-list height; separator and output rows are non-entry | Unit | task0001 |
| TS-5 | A click moves the cursor without changing marks | Cursor 3; the same 2 marks; active pane unchanged | Unit | task0001 |
| TS-6 | Clicking an entry in the inactive pane | Right pane active with cursor 2; left pane's cursor and marks unchanged | Unit | task0001 |
| TS-7 | Clicking non-entry areas, the title bar and the status bar | The right pane becomes active and its cursor stays 1; title and status clicks change nothing | Unit | task0001 |
| TS-8 | A drag range is added to the existing marks | Indices 1–4 and 7 marked; cursor 4 | Unit | task0001 |
| TS-9 | The drag range follows the pointer and can shrink | Ends with 1–2 marked (2 was already marked) and 3–5 unmarked; an upward drag marks 3–5 | Unit | task0001 |
| TS-10 | A drag never marks `..` | Indices 1–3 marked; `..` not marked | Unit | task0001 |
| TS-11 | The drag range is clamped and never scrolls | The range stops at the start pane's first or last visible entry; scroll offset stays 5; the other pane is unchanged | Unit | task0001 |
| TS-12 | Press and release on the same row is a click | Cursor 2; no marks | Unit | task0001 |
| TS-13 | Double-click on a directory, a file and `..` (injected clock) | Directory → the same effect and command as Enter (async directory load); file → open command per EnterBehavior; `..` → load of the parent directory | Unit | task0001 |
| TS-14 | Presses outside the window, on another entry, or A→B→A are not double-clicks | No Enter at 600 ms apart, for different entries, or for A→B→A; a triple press within 500 ms runs Enter once | Unit | task0001 |
| TS-15 | Double-click on an entry of the inactive pane | Right pane active; cursor on the entry; Enter action executed | Unit | task0001 |
| TS-16 | Mouse input is ignored in modal states | Active pane, cursors, marks and modal state unchanged; no command | Unit | task0001 |
| TS-17 | bg split view with the file list focused | A file-list click moves the cursor; a click on the output area changes nothing | Unit | task0001 |
| TS-18 | Wheel, right-button and middle-button input is ignored | Cursor, marks, active pane and scroll offset unchanged | Unit | task0001 |
| TS-19 | Existing test suites pass unchanged (regression) | Every pre-existing unit test and E2E test passes | Integration / E2E | task0001, task0002 |
| TS-20 | E2E: a click moves the cursor | `assert_cursor_position` reports the third entry | E2E (conditional) | task0002 |
| TS-21 | E2E: a drag marks several files | The header shows "Marked N/" with N = entries A..B, excluding `..` | E2E (conditional) | task0002 |
| TS-22 | E2E: a double-click enters a directory | `another.txt` is displayed (inside dir2) | E2E (conditional) | task0002 |

## Code Quality Verification

- Format: `workflow.yaml` → `project.components.duofm.format_command`. It
  runs `gofmt -w .` and must produce no diff on the integrated branch.
- Static analysis: `workflow.yaml` declares no static-analysis command, so it
  is not a gate.

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC1 | A click moves the cursor to the file without changing marks; a click on the inactive pane activates it and moves its cursor | TS-5, TS-6, TS-12 (unit); TS-20 (E2E, conditional) |
| AC2 | A drag marks the entries from the start row to the release row (excluding `..`) on top of the existing marks; the cursor ends on the release row; the range stays within visible rows; nothing scrolls | TS-8, TS-9, TS-10, TS-11 (unit); TS-21 (E2E, conditional) |
| AC3 | A double-click within 500 ms performs the Enter action on the entry | TS-13, TS-14, TS-15 (unit); TS-22 (E2E, conditional) |
| AC4 | A click on blank or header rows only activates the pane; title and status bar clicks do nothing | TS-3, TS-7 |
| AC5 | While a dialog, the sort dialog, the minibuffer or the focused bg output is shown, mouse input changes nothing; wheel, right and middle input changes nothing | TS-16, TS-17, TS-18 |
| AC6 | Existing unit and E2E tests pass unchanged | TS-19 |
| — | FR1–FR8 and NFR1–NFR5 are implemented and tested | Coverage table below |
| — | TS-1..TS-19 pass; TS-20..TS-22 follow their conditions | This document |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1, TS-2, TS-3, TS-4 |
| FR2 | task0001, task0002 | TS-5, TS-6, TS-12, TS-15, TS-20 |
| FR3 | task0001 | TS-3, TS-7 |
| FR4 | task0001, task0002 | TS-8, TS-9, TS-10, TS-12, TS-21 |
| FR5 | task0001 | TS-11 |
| FR6 | task0001, task0002 | TS-13, TS-14, TS-15, TS-22 |
| FR7 | task0001 | TS-4, TS-16, TS-17 |
| FR8 | task0001 | TS-18 |
| NFR1 | task0001, task0002 | TS-19 |
| NFR2 | task0001 | TS-1 (the hit test is a pure function tested without a terminal); manual check that mouse code lives only under `internal/ui` |
| NFR3 | task0001 | TS-1, TS-2, TS-4 |
| NFR4 | task0001 | TS-13, TS-14 |
| NFR5 | task0001 | No SPEC test scenario. Manual check below: `cmd/duofm/main.go` is unchanged |

## E2E Testing

- Command: `workflow.yaml` → `project.components.duofm.e2e_test_command`
  (`make test-e2e-build && make test-e2e`).
  - The image's default command runs `/e2e/scripts/run_all_tests.sh`
    (set by task0002).
  - The full run includes the "Mouse Tests" section.
- Judge results from the runner's per-test ✓/✗ lines.
- [ ] TS-19: every pre-existing E2E test passes. That is every section
      except "Mouse Tests".
- [ ] TS-20, TS-21, TS-22: the three tests in the "Mouse Tests" section pass.
      - These three are conditional: sending SGR sequences through
        `tmux send-keys -l` is unverified in the E2E container.
      - If any of them fails, first check whether the injected sequence
        reaches duofm as a mouse event. For example, run the E2E image
        interactively, send one SGR press on an entry row, and watch the
        status-bar position.
      - If the sequence is not delivered, record TS-20, TS-21 and TS-22 as
        not applicable and rely on the unit fallbacks: TS-5 for TS-20, TS-8
        for TS-21, TS-13 for TS-22.
      - If the sequence is delivered, the failure is real.
      - TS-22 depends on send timing. If TS-22 alone fails while the
        sequence is delivered, re-run it once before judging.

## Manual Testing (E2E Not Possible)

- [ ] In a real terminal emulator with mouse reporting, run duofm and check:
      - a click moves the cursor;
      - a click on the other pane activates it;
      - a drag across entries marks them, and the header count matches;
      - a double-click on a directory enters it;
      - a double-click on a file opens it per EnterBehavior;
      - A→B→A clicked quickly does not run Enter;
      - the wheel and the right/middle buttons do nothing.
- [ ] Odd terminal width: a click on the rightmost leftover column does
      nothing.
- [ ] NFR5: `git diff` of `cmd/duofm/main.go` between
      `implement.base_commit` and the integration tip is empty. The program
      still uses only the existing cell-motion mouse reporting option.
- [ ] NFR2: `internal/fs` has no mouse-related change; the mouse code lives
      only in `internal/ui`.
- (No mockup comparison: the design step was skipped.)

## Performance / Security Verification (if applicable)

- Not applicable (SPEC: 該当なし).

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit test scenarios (TS-1..TS-18) | 18 | 18 | 0 | 0 |
| Regression (TS-19) | 1 | 1 | 1 | 0 |
| E2E mouse scenarios (TS-20..TS-22, conditional) | 3 | 0 | 3 | 0 |
| Format | 1 | 1 | 0 | 0 |
| Manual checks (terminal behavior, odd width, NFR5, NFR2) | 4 | 0 | 0 | 4 |
