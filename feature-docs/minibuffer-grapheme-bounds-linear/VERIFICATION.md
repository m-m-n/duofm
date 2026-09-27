# Verification Document: minibuffer-grapheme-bounds-linear

## Overview
**Feature**: minibuffer-grapheme-bounds-linear / **SPEC.md**: `feature-docs/minibuffer-grapheme-bounds-linear/SPEC.md` / **IMPLEMENTATION.md**: not written (reduced tier; one task, no file shared between tasks) / **THREAT-MODEL.md**: `feature-docs/minibuffer-grapheme-bounds-linear/THREAT-MODEL.md`

## Build Verification
- Command: `make build`
- Expected: exit code 0, no errors

## Test Verification
- Command: `go test ./...`
- Coverage target: not specified by SPEC.md; the changed locator is exercised by TS1–TS4 and by task0001's direct locator tests (task0001 AC-5)

### Test Scenarios from SPEC.md
| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS1 | Input "." + 100000 spaces + "A", visible minibuffer, cursor at rune count − 1; View timed with the existing timed-View helper | Returns within `linearTimeLimit` (10 s) without panic; result contains no line break; display width at most `m.width-2` | Unit |
| TS2 | Same input, cursor at the end; one Left and, separately, one Ctrl+B, each run in its own goroutine on a fresh minibuffer | Each finishes within `linearTimeLimit` without panic (a panic is reported as a test failure); cursor position becomes rune count − 1 | Unit |
| TS3 | Same input, cursor at rune count − 1; one Right and, separately, one Ctrl+F, each run in its own goroutine on a fresh minibuffer | Each finishes within `linearTimeLimit` without panic; cursor position becomes the rune count | Unit |
| TS4 | Every test that existed in `internal/ui/minibuffer_test.go` before this feature (grapheme movement, whole-cluster reverse, linear-time TS1–TS4 and TS7 of the earlier feature, single-line width bound) runs unmodified | All pass; the diff does not modify or delete any pre-existing test | Unit |
| TS5 | `go test ./...`, `go vet ./...`, `gofmt -l .` | Tests pass; vet reports nothing; gofmt lists no file | Static |

## Code Quality Verification
- Format: `gofmt -l .` prints nothing (the component format command `gofmt -w .` must leave no diff)
- Static analysis: `go vet ./...` exits 0

## SPEC.md Compliance
### Success Criteria
| ID | Criterion | How to Verify |
|----|-----------|---------------|
| SC1 | All functional requirements are implemented and tested | Functional Requirements Coverage below: each requirement has task0001 and at least one passing scenario |
| SC2 | All test scenarios pass | TS1–TS5 pass |
| SC3 | The reproduction procedure (paste "." + about 8000 spaces + "A", move the cursor left and right near the end) shows no delay | Manual item M1; automated counterpart TS1–TS3 at 100000 spaces |
| SC4 | A test detecting recurrence exists | TS1–TS3 exist, and task0001's red-phase record shows they fail against the pre-change locator (task0001 AC-4) |

### Functional Requirements Coverage
| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS1, TS2, TS3; diff inspection: the locator uses the grapheme-cluster-only stepping function with carried state and no longer uses the combined-boundary iterator |
| FR2 | task0001 | TS2, TS3, TS4; the locator's signature and its callers are unchanged in the diff |
| NFR1 | task0001 | TS1, TS2, TS3; `guardFit` and View's windowing logic are unchanged in the diff |
| NFR2 | task0001 | TS5; `go.mod` / `go.sum` unchanged in the feature diff |
| NFR3 | task0001 | TS4, TS5 |

## E2E Testing
The project's E2E suite runs as a regression check only; no E2E scenario is added (SPEC.md A3: no existing E2E test covers the minibuffer; the recurrence check is met by TS1–TS3).
- [ ] `make test-e2e-build && make test-e2e` passes

## Manual Testing (E2E Not Possible)
- [ ] M1: Open the minibuffer, paste "." followed by about 8000 spaces and "A", then move the cursor left and right near the end with Left/Ctrl+B and Right/Ctrl+F — cursor movement and redraw show no noticeable delay.

## Performance / Security Verification
- NFR1: on "." + 100000 spaces + "A", View (cursor at rune count − 1), Left and Ctrl+B (cursor at the end), and Right and Ctrl+F (cursor at rune count − 1) each complete within `linearTimeLimit` (10 s) — checked by TS1, TS2, TS3; the `guardFit` re-measurement bound (at most 3) and View's windowing logic stay unchanged — checked by diff inspection.
- TM-1: the locator computes grapheme cluster boundaries only, in one state-carrying forward pass that stops at the containing cluster — checked by TS1, TS2, TS3 passing within `linearTimeLimit`, and by task0001's red-phase record showing the same tests fail against the pre-change locator.

## Verification Summary
| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Test scenarios (TS1–TS5) | 5 | 5 | 0 | 0 |
| Code quality (format, vet) | 2 | 2 | 0 | 0 |
| E2E regression | 1 | 0 | 1 | 0 |
| Manual (M1) | 1 | 0 | 0 | 1 |
| Performance / Security (NFR1, TM-1) | 2 | 2 | 0 | 0 |
| **Total** | 12 | 10 | 1 | 1 |
