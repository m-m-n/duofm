# Verification Document: narrow-pane-render-panic

## Overview

**Feature**: narrow-pane-render-panic / **SPEC.md**: `feature-docs/narrow-pane-render-panic/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/narrow-pane-render-panic/IMPLEMENTATION.md`

## Build Verification

- Command: workflow.yaml `project.components.duofm.build_command` (containerized `make build`)
- Expected: exit code 0, no errors

## Test Verification

- Command: workflow.yaml `project.components.duofm.test_command` (containerized `go test ./...`)
- Expected: exit code 0, every package passes
- Coverage target: no numeric target for this feature. Every new branch must be exercised by TS-1, TS-5 or TS-6:
  - the zero-clamped separator count
  - the minibuffer path taken when the input display width is ≤ 0

### Test Scenarios from SPEC.md

TS-1 to TS-5 correspond one-to-one to SPEC.md TS1 to TS5. TS-6 is derived from FR1 / NFR1 (separator glyph count).

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | Narrow pane widths, all pane views (`internal/ui/pane_render_test.go`). For each pane width 0–4, render `View`, `ViewWithBgOutput` (focused false and true, non-empty command, output buffer with lines), `ViewDimmedWithDiskSpace`, and `ViewWithMinibuffer` (minibuffer shown, prompt "/: ", width = pane width, input empty and non-empty). Each call is wrapped so a panic becomes a test failure. | Every case returns a string without panicking | Unit |
| TS-2 | Row count kept at narrow widths (`internal/ui/pane_render_test.go`). For pane widths 0 and 1, compare each of the four views' row count (ignoring the trailing line break) with the same view at width 40 and the same height. `ViewWithMinibuffer` is measured with the minibuffer shown. | Row counts are equal for every view | Unit |
| TS-3 | Whole model view at terminal widths 0–3 (`internal/ui/model_basic_test.go`). Initialize the model with a window-size message of width 0, 1, 2 or 3 at a normal height. Call `Model.View` in three states: no minibuffer; after starting incremental search; after starting shell command mode. | No panic in any of the 12 cases | Unit |
| TS-4 | Full test suite (`./...`), including the existing width-40 row-alignment / hit-test tests, `TestMinibufferView` and `TestMinibufferViewTruncation` | All tests pass | Unit |
| TS-5 | `Minibuffer.View` at narrow widths (`internal/ui/minibuffer_test.go`). Shown minibuffer, every combination of: width 0–4; prompt "/: " or "!: "; input "", "a" or "abc"; cursor at 0, half the rune count (rounded down), or the rune count. Plus: prompt "(reverse-i-search)'': ", width 24, input "a", cursor 1. | No panic; output contains no line break | Unit |
| TS-6 | Separator glyph count (`internal/ui/pane_render_test.go`). For pane widths 0, 1, 2, 3, 4 and 40, count the "─" glyphs in the third output row of `View`, `ViewWithBgOutput` and `ViewDimmedWithDiskSpace`. | Count equals max(0, width − 2); zero at widths 0 and 1 | Unit |

### Regression-detection check (SPEC AC2, AC7)

- Primary evidence: task0001's TDD red phase, recorded in `test-docs/narrow-pane-render-panic/task0001.tests.yaml`. It must show:
  - TS-1 failing at pane widths 0 and 1 for the `View`, `ViewWithBgOutput` and `ViewDimmedWithDiskSpace` cases
  - TS-5 failing, including the case width 4, prompt "/: ", input "a", cursor 1, before the production fix
- Optional re-check:
  1. In a scratch checkout of the implement base commit, add only the new test code from `internal/ui/pane_render_test.go` and `internal/ui/minibuffer_test.go`.
  2. Run the TS-1 and TS-5 tests.
  3. Expect the same failures as in the primary evidence.
  4. Discard the scratch checkout.

## Code Quality Verification

- Format: workflow.yaml `project.components.duofm.format_command` (containerized `gofmt -w .`). Expected: no file changes afterwards.
- Static analysis: `go vet ./...` (project convention in CLAUDE.md), run in the same container form as `test_command`. Expected: no findings.

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| SC-1 | All functional requirements (FR1–FR6) are implemented and tested | Functional Requirements Coverage table below; TS-1, TS-2, TS-3, TS-5, TS-6 pass |
| SC-2 | All test scenarios (SPEC TS1–TS5) pass | TS-1 to TS-5 pass (TS-6 as well) |
| SC-3 | Acceptance criteria AC1–AC8 are met | AC1 → TS-1; AC2 → regression-detection check; AC3 → TS-3; AC4 → TS-2; AC5 → TS-4; AC6 → TS-5; AC7 → regression-detection check; AC8 → TS-1 |
| SC-4 | Code review is complete | workflow.yaml review step completed |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-6 (separator count); TS-1 and TS-2 (no panic, row kept) |
| FR2 | task0001 | TS-1 |
| FR3 | task0001 | TS-3 |
| FR4 | task0001 | TS-2 |
| FR5 | task0001 | TS-1, TS-3, TS-5 exist and pass; regression-detection check |
| FR6 | task0001 | TS-5 |
| NFR1 | task0001 | TS-4 (existing width-40 alignment / hit-test tests); TS-6 at widths 2, 3, 4, 40 |
| NFR2 | task0001 | TS-4 |
| NFR3 | task0001 | TS-4 (existing `TestMinibufferView`, `TestMinibufferViewTruncation`) |

## E2E Testing

No E2E scenario is added (SPEC A5). The existing E2E suite is run as a regression check only.

- [ ] Existing E2E suite passes: workflow.yaml `project.components.duofm.e2e_test_command`

## Manual Testing (E2E Not Possible)

- [ ] Reproduction steps from the goal: start duofm in a terminal 3 columns wide or narrower (pane width 0–1). It starts and draws without a panic.
- [ ] In the same narrow terminal, start incremental search and then, after leaving it, shell command mode. Neither mode panics while its minibuffer is shown.

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit test scenarios (TS-1 to TS-6) | 6 | 6 | 0 | 0 |
| Regression-detection check | 1 | 1 | 0 | 0 |
| Code quality (format, vet) | 2 | 2 | 0 | 0 |
| E2E regression (existing suite) | 1 | 0 | 1 | 0 |
| Manual reproduction | 2 | 0 | 0 | 2 |
