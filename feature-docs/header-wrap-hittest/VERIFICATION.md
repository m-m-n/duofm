# Verification Document: header-wrap-hittest

## Overview

**Feature**: header-wrap-hittest / **SPEC.md**: `feature-docs/header-wrap-hittest/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/header-wrap-hittest/IMPLEMENTATION.md`

This is the integrated verification run by the verify phase. Task-level
acceptance criteria live in `tasks/task0001.md`.

## Build Verification

- Command: `project.components.duofm.build_command` in workflow.yaml (runs `make build` in the golang 1.25.5 trixie container)
- Expected: exit code 0, no errors

## Test Verification

- Command: `project.components.duofm.test_command` in workflow.yaml (runs `go test ./...` in the golang 1.25.5 trixie container)
- Expected: exit code 0; every scenario below passes
- Coverage target: the project defines no numeric coverage threshold. Every
  scenario below must be covered by at least one passing test.

### Test Scenarios from SPEC.md

| ID | SPEC ID | Scenario | Expected Result | Test Type |
|----|---------|----------|-----------------|-----------|
| TS-1 | TS1 | Header line 1 with no git branch, pane width 40, path wider than the pane | Display width (go-runewidth) ≤ pane width − 4; ends with "..." | Unit |
| TS-2 | TS2 | Header line 1 with no git branch, `[H]` indicator and filter indicator, total wider than the pane | Display width ≤ pane width − 4 | Unit |
| TS-3 | TS3 | Header line 1 with no git branch, path containing full-width characters, wider than the pane | Display width ≤ pane width − 4 | Unit |
| TS-4 | TS4 | Header line 1 at a narrow width (15) with a long branch name (branch-only case) | Display width ≤ pane width − 4 | Unit |
| TS-5 | TS5 | Header line 1 and ellipsis truncation at pane widths 0–4 (content width ≤ 0), with and without a git branch | No panic; no negative width reaches the truncation | Unit |
| TS-6 | TS6 | Pane rendered in the normal, bg output split and dimmed views with no git branch and a long path | The row where the first entry appears equals the shared header row count and matches the hit test's first entry row (title row excluded); fails on the renderer as it was before the fix | Unit |
| TS-7 | TS7 | Model mouse handling on a long-path, no-branch pane: click, drag and double-click on the rendered row of entry i | Click moves the cursor to i; drag marks the dragged row range; double-click runs Enter on the clicked entry | Unit |
| TS-8 | TS8 | Regression: existing `mouse_hittest_test.go`, `model_update_mouse_test.go`, `pane_render_test.go` and `go test ./...` | All pass | Integration |
| TS-9 | E2E Tests | Existing E2E mouse tests (`test/e2e/scripts/tests/mouse_tests.sh`) | Pass without regression | E2E |

## Code Quality Verification

- Format: `project.components.duofm.format_command` in workflow.yaml (`gofmt -w .` in the container). Expected: no diff produced.
- Static analysis: no command is declared in workflow.yaml `project.components`, so none runs.

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC1 | Outside git, with a path longer than the pane width, clicking an entry row selects the clicked entry | TS-7; manual check M-1 |
| AC2 | Under the same condition, drag marks the dragged row range and double-click runs Enter on the clicked entry | TS-7; manual check M-2 |
| AC3 | With no branch and a long path, header line 1 fits the content width and ends with "..." | TS-1, TS-2, TS-3; manual check M-1 |
| AC4 | In the branch-only case, header line 1 fits the content width even when the branch name is wider | TS-4; manual check M-3 |
| AC5 | Rendering the header at pane widths 0–4 does not panic | TS-5 |
| AC6 | A test checks that the rendered first-entry row equals the hit test's first-entry row for a long path with no branch, and that test fails on the code before the fix | TS-6 (red-phase failure confirmed at implementation time) |
| AC7 | Existing unit tests with short paths and `go test ./...` all pass | TS-8 |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-6, TS-7 |
| FR2 | task0001 | TS-7 |
| FR3 | task0001 | TS-1, TS-2, TS-3, TS-7 |
| FR4 | task0001 | TS-4 |
| FR5 | task0001 | TS-5 |
| FR6 | task0001 | TS-6 |
| NFR1 | task0001 | TS-8, TS-9 |
| NFR2 | task0001 | TS-1, TS-2, TS-3 |
| NFR3 | task0001 | TS-6 |

## E2E Testing

- Command: `project.components.duofm.e2e_test_command` in workflow.yaml (`make test-e2e-build && make test-e2e`)
- [ ] TS-9: existing `test/e2e/scripts/tests/mouse_tests.sh` passes without regression. The spec does not require new E2E tests (A3).

## Manual Testing (E2E Not Possible)

- [ ] M-1: Outside git, open a directory whose path is longer than the pane width. Header line 1 stays on one row and ends with "...". Clicking an entry row selects the entry on that row, for the first, a middle and the last visible entry.
- [ ] M-2: Under the same condition, drag across several entry rows and check that exactly those entries are marked. Double-click a subdirectory row and check that that subdirectory opens.
- [ ] M-3: Inside a git repository with a long branch name, narrow the terminal until only the branch is shown in header line 1. The header stays on one row and a click still selects the entry on the clicked row.
- [ ] M-4: With a short path (fits the pane), header display and mouse click / drag / double-click behave as before this change.

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit / Integration tests | 8 (TS-1 to TS-8) | 8 | 0 | 0 |
| E2E | 1 (TS-9) | 0 | 1 | 0 |
| Code quality | 1 (format) | 1 | 0 | 0 |
| Manual | 4 (M-1 to M-4) | 0 | 0 | 4 |
