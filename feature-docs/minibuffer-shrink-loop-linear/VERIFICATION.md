# Verification Document: minibuffer-shrink-loop-linear

## Overview

**Feature**: minibuffer-shrink-loop-linear / **SPEC.md**: `feature-docs/minibuffer-shrink-loop-linear/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/minibuffer-shrink-loop-linear/IMPLEMENTATION.md`

Scenario IDs TS-1..TS-6 correspond to SPEC.md TS1..TS6. TS-7..TS-9 are
verification items derived from FR3, NFR1 and SPEC AC5.

## Build Verification

- Command: workflow.yaml `project.components.duofm.build_command` (podman-wrapped `make build`)
- Expected: exit code 0, no errors

## Test Verification

- Command: workflow.yaml `project.components.duofm.test_command` (podman-wrapped `go test ./...`)
- Expected: exit code 0; every package passes
- Coverage target: no coverage gate for this feature; each View path touched by
  the change is exercised by TS-1..TS-5 and TS-7

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | m.width 80, prompt "!: ", input U+0301 x100000 + 'a' x71 + U+00A9 U+FE0F, cursor at the end; View run in a goroutine | Returns within 10 s; no line break; display width at most 78; ANSI-stripped result contains U+00A9 U+FE0F | Unit |
| TS-2 | m.width 40, prompt "!: ", input U+0301 x100000 + 'a' x31 + U+00A9 U+FE0F, cursor at the end | Returns within 10 s; no line break; display width at most 38; ANSI-stripped result contains U+00A9 U+FE0F | Unit |
| TS-3 | m.width 40, prompt "!: ", input U+00A9 U+FE0F + 'a' x32 + U+0301 x100000, cursor at rune index 2 | Returns within 10 s; no line break; display width at most 38; ANSI-stripped result contains U+00A9 U+FE0F immediately followed by 'a' | Unit |
| TS-4 | m.width 40, prompt "(reverse-i-search)'" + U+00A9 U+FE0F + 'a' x16 + U+0301 x100000 + "': ", empty input | Returns within 10 s; no line break; display width at most 38; ANSI-stripped result contains "(reverse-i-search)'" immediately followed by U+00A9 U+FE0F | Unit |
| TS-5 | Existing tests in `internal/ui/minibuffer_test.go` and `internal/ui/pane_render_test.go` (widths 0-4, wide prompt and input, ZWJ and combining characters, cursor-position tables, width 40 untruncated) | All pass with their assertions unchanged | Unit |
| TS-6 | TS-1..TS-4 run under the race detector with verbose output (command in Performance Verification) | Each test passes and logs its View duration; every duration is at most 2 s; the four values are recorded in the verify report | Performance |
| TS-7 | m.width 40, prompt "!: ", input U+0600 U+1F600 U+FE0E repeated 30 times, cursor at the end (rendered width exceeds the unit-width estimate) | Fixture assumption holds (1 + 1 < 3); returns within 10 s; no line break; display width at most 38; ANSI-stripped result contains U+1F600 | Unit |
| TS-8 | Change scope (NFR1): diff from workflow.yaml `workflow.implement.base_commit` to the integration branch tip | Outside `feature-docs/` and `test-docs/`, only `internal/ui/minibuffer.go` and `internal/ui/minibuffer_test.go` change; in `minibuffer.go` only View and its private helpers change; no width-function injection or other test seam; no new non-standard-library import; `go.mod` and `go.sum` unchanged; HandleKey tests pass | Review (diff) |
| TS-9 | Pre-change failure (SPEC AC5): the TS-1..TS-4 tests run against the pre-change View | Each of TS-1..TS-4 fails by reaching the 10-second limit | Evidence / semi-automated |

## Code Quality Verification

- Format: workflow.yaml `project.components.duofm.format_command` (podman-wrapped `gofmt -w .`); expected: no file changes afterwards (`git status` clean)
- Static analysis: none declared in workflow.yaml

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC1 | m.width 80, TS1 input, cursor at the end: View within 10 s | TS-1 |
| AC2 | m.width 40, TS2 input, cursor at the end: View within 10 s | TS-2 |
| AC3 | Cursor-not-at-end path (TS3) and prompt-truncation path (TS4): View within 10 s | TS-3, TS-4 |
| AC4 | AC1-AC3 inputs: no line break, display width at most m.width-2 | TS-1..TS-4 assertions |
| AC5 | TS1-TS4 tests fail by exceeding the limit on the pre-change implementation | TS-9 |
| AC6 | Race-detector durations of TS1-TS4 are well below 10 s | TS-6 |
| AC7 | Existing minibuffer and pane_render tests pass | TS-5 |
| AC8 | `go test ./...` passes | Test Verification command |
| DoD-1 | Goal: the reproduction steps no longer freeze duofm | M-1 (manual), TS-1 and TS-2 (same inputs at View level) |
| DoD-2 | Goal: a test detects recurrence | TS-1..TS-4 together with TS-9 |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1, TS-2, TS-3, TS-4 |
| FR2 | task0001 | TS-5 |
| FR3 | task0001 | TS-1, TS-2, TS-3, TS-4, TS-5, TS-7 |
| FR4 | task0001 | TS-1, TS-2, TS-3, TS-4, TS-9 |
| NFR1 | task0001 | TS-8 |
| NFR2 | task0001 | TS-6 |

## E2E Testing

No E2E scenario is added (SPEC a4). As a regression check, the existing suite
runs with workflow.yaml `project.components.duofm.e2e_test_command`.

- [ ] Existing E2E suite passes

## Manual Testing (E2E Not Possible)

- [ ] M-1: With terminal width 80 and again with terminal width 160, start duofm, open the shell command input (prompt "!: "), keep the cursor at the end, and paste U+0301 x100000 + 71 ASCII characters + U+00A9 U+FE0F. The minibuffer redraws promptly and duofm stays operable (the line stays one row; the pasted text's tail and the cursor block are visible).

No mockup comparison applies (the design step was skipped).

## Performance / Security Verification (if applicable)

- TS-6 command (test_command's container invocation with a race-enabled, filtered test run; not the approved test_command string):
  `podman run --rm --userns=keep-id -v "$PWD":"$PWD" -w "$PWD" -v /home/sakura/go/src/duofm/.git:/home/sakura/go/src/duofm/.git:ro -v duofm-gomod:/gomod:U -v duofm-gocache:/gocache:U -e GOMODCACHE=/gomod -e GOCACHE=/gocache -e GOTOOLCHAIN=local -e DISPLAY=:0 -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=init.defaultBranch -e GIT_CONFIG_VALUE_0=main docker.io/library/golang:1.25.5-trixie go test -race -v -run TestMinibufferView_LinearTime ./internal/ui/`
- TS-6 threshold: every logged View duration is at most 2 s (1/5 of the 10-second limit; planner guideline for NFR2's "十分に下回る"). Record all four values.
- TS-9 evidence: the implementer's red-phase record for task0001 (AC-3). To re-confirm, create a temporary worktree at `workflow.implement.base_commit`, copy the post-change `internal/ui/minibuffer_test.go` into it, and run the same container invocation there with `go test -v -run TestMinibufferView_LinearTime ./internal/ui/` (no race detector); each of the four tests must fail at the 10-second limit.
- Security: no dedicated check beyond input robustness (TS-1..TS-4, TS-7: pathological pasted input cannot freeze or overflow the minibuffer).

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit tests | 6 (TS-1..TS-5, TS-7) | 6 | 0 | 0 |
| Performance | 1 (TS-6) | 1 | 0 | 0 |
| Change scope | 1 (TS-8) | 1 (diff review) | 0 | 0 |
| Pre-change failure | 1 (TS-9) | 1 (recorded red run / re-run) | 0 | 0 |
| Code quality | 1 (format) | 1 | 0 | 0 |
| E2E regression | 1 | 0 | 1 | 0 |
| Manual | 1 (M-1) | 0 | 0 | 1 |
