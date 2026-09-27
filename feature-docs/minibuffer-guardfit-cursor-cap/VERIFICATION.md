# Verification Document: minibuffer-guardfit-cursor-cap

## Overview
**Feature**: minibuffer-guardfit-cursor-cap / **SPEC.md**: `feature-docs/minibuffer-guardfit-cursor-cap/SPEC.md` / **IMPLEMENTATION.md**: not produced (reduced tier, single task) / **THREAT-MODEL.md**: `feature-docs/minibuffer-guardfit-cursor-cap/THREAT-MODEL.md`

## Build Verification
- Command: `project.components.duofm.build_command` in workflow.yaml (runs `make build` in the `docker.io/library/golang:1.25.5-trixie` container)
- Expected: exit code 0, no errors

## Test Verification
- Command: `project.components.duofm.test_command` in workflow.yaml (runs `go test ./...` in the same container)
- Expected: exit code 0; every test passes
- Coverage target: no numeric threshold is set for this feature. TS-1 to TS-4 must each run the cursor-alone fallback in View.

### Test Scenarios from SPEC.md
| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-1 | SPEC TS1 (AC1). Cursor-at-end path: prompt U+0600, m.width 10, input g1+g2+g3+g4, cursor 20, View under the 256-color profile | Exactly one reversed run, a single space. U+0600 is present, neither U+1F600 nor U+2764 is present, and assertBoundedSingleLine passes | Unit |
| TS-2 | SPEC TS2 (AC2). Right-then-left trimming path: same prompt and width, input g2+g3+g4+"xyz", cursor 18 (on "x") | Exactly one reversed run, "x". U+0600 is present, neither U+1F600 nor U+2764 is present, and assertBoundedSingleLine passes | Unit |
| TS-3 | SPEC TS3 (AC3). Left-scroll path: same prompt and width, input "z"+g1+g2+g3+g4+"x", cursor 21 (on "x") | Same as TS-2 | Unit |
| TS-4 | SPEC TS4 (AC4). The prompt plus the cursor-alone line is still wider than content width 6, e.g. g5 = U+2764 U+200D + g4 at the start of the input, cursor 0, under a profile that emits no style sequences | No input rune and no cursor are shown (no reversed run), and assertBoundedSingleLine passes | Unit |
| TS-5 | SPEC TS5 (AC5). Existing tests in internal/ui/minibuffer_test.go and internal/ui/pane_render_test.go (width 0-4, full-width, ZWJ / VS16 / grapheme, linear-time LinearTime_TS1-TS4, TS7 joined-sequence guard, scroll boundaries) | All pass with the test files' existing cases unmodified | Unit |
| TS-6 | SPEC AC6. Changed-file set of the integration branch compared with its base | Only internal/ui/minibuffer.go and internal/ui/minibuffer_test.go change outside the workflow-generated feature-docs/ and test-docs/ entries | Inspection |

TS-n here are this document's scenario IDs. TS-1 to TS-5 correspond to
SPEC.md's TS1 to TS5. The existing Go test names
TestMinibufferView_LinearTime_TS1 to TS4 are a separate numbering.

Fixture graphemes (SPEC FR4): g1 = U+1F600 U+FE0E; g2 = U+2764 U+200D + g1;
g3 = U+2764 U+200D + g2; g4 = U+2764 U+200D + g3.

## Code Quality Verification
- Format: `project.components.duofm.format_command` in workflow.yaml (`gofmt -w .` in the container). Expected: no diff afterwards.
- Static analysis: `go vet ./...` in the same container. Expected: no findings in internal/ui.

## SPEC.md Compliance
### Success Criteria
| ID | Criterion | How to Verify |
|----|-----------|---------------|
| SC-1 | Every functional requirement is implemented and tested | Functional Requirements Coverage below: every row has a task and a passing scenario |
| SC-2 | Every test scenario passes | TS-1 to TS-5 pass under the test command; TS-6 inspection passes |
| SC-3 | SPEC AC1 to AC6 are met | AC1 to TS-1, AC2 to TS-2, AC3 to TS-3, AC4 to TS-4, AC5 to TS-5, AC6 to TS-6 |

### Functional Requirements Coverage
| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS-1, TS-2, TS-3 |
| FR2 | task0001 | TS-4 |
| FR3 | task0001 | TS-1, TS-2, TS-3, TS-4, TS-5 |
| FR4 | task0001 | TS-1, TS-2, TS-3, TS-4 (each test exists and goes through View) |
| NFR1 | task0001 | TS-6 |
| NFR2 | task0001 | TS-5 (linear-time tests within 10 s); TM-1 check below |
| NFR3 | task0001 | TS-5 |

## E2E Testing
Existing framework: test/e2e (Docker + tmux bash scripts). This feature adds no E2E test (SPEC a4).
- [ ] Existing E2E tests pass without regression: `project.components.duofm.e2e_test_command` in workflow.yaml (`make test-e2e-build && make test-e2e`)

## Manual Testing (E2E Not Possible)
None. Every scenario is covered by automated unit tests or by inspecting the changed files.

## Performance / Security Verification (if applicable)
- NFR2: TestMinibufferView_LinearTime_TS1 to TS4 pass unchanged, each within its 10-second limit.
- TM-1: the cursor-alone fallback makes at most one width measurement per View call beyond guardFit's own, so redraw cost does not grow with input length. Check: TS-5's linear-time tests pass within their limits, and reading the fallback in internal/ui/minibuffer.go shows exactly one measurement and no loop.
- TM-2: the cursor-alone line is displayed only after it has been measured within the content width. Otherwise neither input nor cursor is rendered, and the output stays one line no wider than m.width-2. Check: TS-4 passes (no input, no cursor, assertBoundedSingleLine), and TS-1 to TS-3 pass assertBoundedSingleLine.

## Verification Summary
| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit test scenarios (TS-1 to TS-5) | 5 | 5 | 0 | 0 |
| Change-scope inspection (TS-6) | 1 | 0 | 0 | 1 |
| Code quality (format, vet) | 2 | 2 | 0 | 0 |
| E2E regression | 1 | 0 | 1 | 0 |
| Performance (NFR2) | 1 | 1 | 0 | 0 |
| Security (TM-1, TM-2) | 2 | 1 | 0 | 1 |
| **Total** | 13 | 10 | 1 | 2 |
