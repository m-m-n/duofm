# Verification Document: minibuffer-vs16-regression-test

## Overview
**Feature**: minibuffer-vs16-regression-test / **SPEC.md**: `feature-docs/minibuffer-vs16-regression-test/SPEC.md` / **IMPLEMENTATION.md**: not produced (reduced tier, single task, no cross-task decisions)

Scope: regression tests added to `internal/ui/minibuffer_test.go` (task0001).
The implementation `internal/ui/minibuffer.go` is not changed.

Fixture names used below:
- P-JA: prompt `(reverse-i-search)'日本語': ` (display width 28)
- PAIR: U+2764 U+FE0F (display width 2; sum of per-rune widths 1)
- IN-4: PAIR repeated 4 times (8 runes)
- P-VS: prompt `(reverse-i-search)'` + PAIR repeated 4 times + `': ` (display width 30)

## Build Verification
- Command: `make build`
- Expected: exit code 0, no errors

## Test Verification
- Command: `go test ./...`
- Expected: exit code 0; every test passes, including the tests added by task0001
- Focused command: `go test ./internal/ui/...` (SPEC AC1)
- Coverage target: not applicable (test-only change; no production code is added or changed)

### Test Scenarios from SPEC.md
| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | (SPEC TS1) Reproduction condition, cursor at end: P-JA, m.width 37, input IN-4, cursor at end, minibuffer shown, View rendered | Premise values hold (P-JA 28, PAIR 2, per-rune sum 1); no line break; display width <= 35; ANSI-stripped text contains P-JA immediately followed by two PAIRs; no U+2764 without a following U+FE0F and no U+FE0F without a preceding U+2764 | Unit |
| TS-2 | (SPEC TS2) Width x cursor table: P-JA, input IN-4, every m.width 24-41 x every cursor position 0-8 (including positions on U+FE0F) | Every combination: no panic, no line break, display width <= m.width-2 | Unit (table-driven) |
| TS-3 | (SPEC TS3) Prompt-truncation path width bound: P-VS, empty input, every m.width 24-34 | Premise P-VS width 30 holds; every width: no line break, display width <= m.width-2 | Unit (table-driven) |
| TS-4 | (SPEC TS4) Prompt-truncation path content: P-VS, empty input, m.width 29 | ANSI-stripped text contains `(reverse-i-search)'` immediately followed by three PAIRs; no split pair | Unit |
| TS-5 | (SPEC AC3) Regression detection, unit width estimate: temporarily change the width-unit partitioning (partitionUnits) to estimate a unit's width as the sum of its runes' own widths, then run the TS-1 and TS-4 tests | TS-1's content check and TS-4's content check fail; the change is reverted and not committed | Mutation check (temporary, uncommitted) |
| TS-6 | (SPEC AC4) Regression detection, guard removed: TS-5's change plus the fit guard (guardFit) always returning its first rendering, then run the TS-1 test | TS-1's single-line check fails (line break, or display width > 35); both changes are reverted and not committed | Mutation check (temporary, uncommitted) |
| TS-7 | (SPEC AC5) Full suite and formatting: `go test ./...`, `gofmt -l internal/ui` | Test suite exit code 0; gofmt lists no file | Build / format |
| TS-8 | (SPEC NFR1, NFR2) Change scope and conventions: diff of the integration branch against its base | Outside `feature-docs/` and `test-docs/`, only `internal/ui/minibuffer_test.go` changed; `internal/ui/minibuffer.go`, `go.mod`, `go.sum` unchanged; new test functions are named `TestMinibufferView_{Scenario}` and use assertBoundedSingleLine, mustNotPanic and stripANSI | Static check (diff + review) |

## Code Quality Verification
- Format: `gofmt -w .` (workflow.yaml) — expected to leave no diff; check-only equivalent `gofmt -l internal/ui` prints nothing
- Static analysis: `go vet ./...` — expected exit code 0

## SPEC.md Compliance
### Success Criteria
| ID | Criterion | How to Verify |
|----|-----------|---------------|
| SC-1 | SPEC AC1: the added tests pass against the current implementation with `go test ./internal/ui/...` | TS-1, TS-2, TS-3, TS-4 |
| SC-2 | SPEC AC2: a test checks that the reproduction condition renders on one line | TS-1 |
| SC-3 | SPEC AC3: the unit-width regression makes the input-side (FR2) and prompt-side (FR5, m.width 29) content checks fail | TS-5 |
| SC-4 | SPEC AC4: the unit-width regression plus guard removal makes the FR1 single-line check fail | TS-6 |
| SC-5 | SPEC AC5: `go test ./...` passes in full and gofmt produces no diff | TS-7 |
| SC-6 | Change limited to the test file, no new dependency, existing conventions followed (NFR1, NFR2) | TS-8 |

### Functional Requirements Coverage
| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1 (single line, width <= 35); TS-6 (detects guard regression) |
| FR2 | task0001 | TS-1 (two pairs after the prompt, no split pair); TS-5 (detects unit-width regression) |
| FR3 | task0001 | TS-1 (premise checks run first and stop the test fatally on drift) |
| FR4 | task0001 | TS-2 |
| FR5 | task0001 | TS-3, TS-4; TS-5 (detects unit-width regression on the prompt side) |
| NFR1 | task0001 | TS-8 |
| NFR2 | task0001 | TS-7 (gofmt), TS-8 (naming, helpers, no new dependency) |
| NFR3 | task0001 | TS-7 |

## E2E Testing
None for this feature (SPEC A5): the display-width checks are unit tests.

## Manual Testing (E2E Not Possible)
- [ ] TS-5: apply the temporary partitioning change, run the TS-1 and TS-4 tests, confirm the content checks fail, revert (`internal/ui/minibuffer.go` identical to base)
- [ ] TS-6: apply the TS-5 change plus the temporary guard change, run the TS-1 test, confirm the single-line check fails, revert both

## Verification Summary
| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit test scenarios (TS-1 to TS-4) | 4 | 4 | 0 | 0 |
| Regression detection (TS-5, TS-6) | 2 | 0 | 0 | 2 |
| Full suite and format (TS-7) | 1 | 1 | 0 | 0 |
| Change scope and conventions (TS-8) | 1 | 1 | 0 | 0 |
| Static analysis (go vet) | 1 | 1 | 0 | 0 |
