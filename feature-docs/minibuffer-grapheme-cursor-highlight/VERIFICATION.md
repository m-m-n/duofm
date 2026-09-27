# Verification Document: minibuffer-grapheme-cursor-highlight

## Overview

**Feature**: minibuffer-grapheme-cursor-highlight / **SPEC.md**: `feature-docs/minibuffer-grapheme-cursor-highlight/SPEC.md` / **IMPLEMENTATION.md**: `feature-docs/minibuffer-grapheme-cursor-highlight/IMPLEMENTATION.md`

Scenario IDs below are the SPEC.md IDs verbatim (TS1–TS10, hyphen-less) so that they string-match the `tests` lists in workflow.yaml `requirements`.

## Build Verification

- Command: make build
- Expected: exit code 0, no errors

## Test Verification

- Command: go test ./...
- Expected: exit code 0; every test in internal/ui/minibuffer_test.go passes, including the pre-existing linear-time tests (TS1–TS4, TS7 of the earlier features) within their 10-second limit
- Coverage target: no project-wide threshold. Every branch of internal/ui/minibuffer.go changed by this feature (the four movement keys, the step 1 fit check, steps 3, 4 and 5 of View, the line builder's range highlight) is exercised by at least one scenario below

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS1 | Left/Ctrl+B on the family emoji: familyEmojiLsInput, cursor 9, three presses | cursorPos 8, 7, 0 for Left and for Ctrl+B | Unit |
| TS2 | Right/Ctrl+F on the family emoji: familyEmojiLsInput, cursor 0, three presses | cursorPos 7, 8, 9 for Right and for Ctrl+F | Unit |
| TS3 | Movement from inside a cluster: SetCursorPos(1..6), table-driven | Left gives 0, Right gives 7 | Unit |
| TS4 | Movement over other cluster kinds: e + U+0301, U+2764 U+FE0F, flag U+1F1EF U+1F1F5, leading isolated zero-width rune | Each Left/Right press crosses exactly one grapheme cluster | Unit |
| TS5 | Whole-cluster highlight: 256-color profile fixed (restored at cleanup), searchPrompt, familyEmojiLsInput, fitting width, cursor 0..9 | Cursor 0..6: one reversed run holding all 7 emoji runes, none outside it; 7: "l"; 8: "s"; 9: block cursor (reversed space); reproduction sequence never splits the emoji | Unit |
| TS6 | No truncation with the cursor inside the cluster: 256-color profile fixed, searchPrompt, familyEmojiLsInput, m.width 19..24, cursor 1..6 | ANSI-stripped output contains prompt immediately followed by the whole input | Unit |
| TS7 | Cursor cluster on scroll paths: 256-color profile fixed, searchPrompt, familyEmojiLsInput, m.width 14..18, cursor 0..6 | Family emoji either whole inside one reversed run or entirely absent; at m.width 15 no input rendered | Unit |
| TS8 | Single line and width bound: existing ZWJ/VS16/combining tables plus new flag and combining-mark tables, all widths and cursor positions, default and 256-color profiles | assertBoundedSingleLine passes everywhere (no line break, width <= m.width-2) | Unit |
| TS9 | Editing stays rune-wise: familyEmojiLsInput, SetCursorPos(3), then Backspace / Delete / insertion / Ctrl+K / Ctrl+U | Backspace and Delete each remove exactly one rune, insertion adds exactly the typed rune, Ctrl+K and Ctrl+U cut at rune position 3; cursorPos 2 after Backspace; existing editing tests pass | Unit |
| TS10 | Existing tests and static checks | go test ./..., go vet ./... pass; gofmt produces no difference; go.mod has rivo/uniseg v0.4.7 as a direct requirement and no other change | Unit / static check |

## Code Quality Verification

- Format: gofmt -w . (from workflow.yaml) — expected: produces no change to any tracked file
- Static analysis: go vet ./... — expected: exit code 0, no findings
- Dependency check: go.mod diff against the base branch is limited to rivo/uniseg v0.4.7 moving from indirect to direct; go.sum unchanged or changed only as the toolchain requires for that promotion

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| AC1 | From cursor 9, Left/Ctrl+B go 8, 7, 0 | TS1 |
| AC2 | From cursor 0, Right/Ctrl+F go 7, 8, 9 | TS2 |
| AC3 | From SetCursorPos(1..6), Left goes to 0 and Right to 7 | TS3 |
| AC4 | Combining mark, VS16 and flag clusters are crossed one cluster per press | TS4 |
| AC5 | With 256-color profile, cursor 0..6 reverses the whole emoji as one run; 7 and 8 reverse only "l" / "s" | TS5 |
| AC6 | Reproduction steps never show the emoji split | TS5 (reproduction sequence) |
| AC7 | m.width 19..24, cursor 1..6: prompt and whole input shown untruncated | TS6 |
| AC8 | Scroll widths: emoji whole in one reversed run or not rendered | TS7 |
| AC9 | Every width/cursor combination stays one line within m.width-2 | TS8 |
| AC10 | Editing inside a cluster stays rune-wise | TS9 |
| AC11 | Existing tests (including linear-time tests) pass; go vet and gofmt clean | TS9, TS10 |
| AC12 | Regression unit tests for cluster movement and whole-cluster highlight exist in internal/ui/minibuffer_test.go | Review of internal/ui/minibuffer_test.go against TS1–TS5, TS7 |
| SC-all | All functional requirements implemented and tested; all scenarios pass; performance goals met; code review completed | This document's scenarios, the linear-time tests, and the review phase |

### Functional Requirements Coverage

| Requirement | Tasks | Verification |
|-------------|-------|--------------|
| FR1 | task0001 | TS1, TS2, TS3, TS4, TS5, TS10 |
| FR2 | task0001 | TS5, TS10 |
| FR3 | task0001 | TS6 |
| FR4 | task0001 | TS7, TS10 |
| FR5 | task0001 | TS9, TS10 |
| NFR1 | task0001 | TS8, TS9, TS10 |
| NFR2 | task0001 | TS9, TS10 (plus review: segmentation only through rivo/uniseg, width basis unchanged) |
| NFR3 | task0001 | TS9, TS10 (pre-existing linear-time tests) |
| NFR4 | task0001 | TS9, TS10 (go.mod dependency check) |

## E2E Testing

No feature-specific E2E scenario (SPEC A6: recurrence detection is satisfied by unit tests). The project E2E command from workflow.yaml is run as a regression check only.

- [ ] Project E2E suite passes unchanged: make test-e2e-build && make test-e2e (expected exit code 0)

## Manual Testing (E2E Not Possible)

Requires a terminal whose font renders the ZWJ family emoji as a single glyph; rendering quality itself is terminal-dependent and not asserted.

- [ ] Open the minibuffer (for example in search), enter the family emoji (U+1F468 U+200D U+1F469 U+200D U+1F467 U+200D U+1F466) followed by "ls", then press Left repeatedly from the end: the cursor moves onto "s", then "l", then onto the whole emoji in a single press; the emoji is shown reversed as one glyph and is never drawn split
- [ ] Narrow the terminal until the minibuffer has to scroll, with the cursor on the emoji: the emoji is either shown whole and reversed or not shown at all, and the minibuffer line never wraps

## Performance / Security Verification

- NFR3 linear time: the pre-existing time-bound tests in internal/ui/minibuffer_test.go (linear-time TS1–TS4 and TS7 of the earlier features) pass within 10 seconds each
- Security: not applicable (no security-relevant surface; SPEC "Security Considerations": none)

## Verification Summary

| Category | Items | Automated | E2E | Manual |
|----------|-------|-----------|-----|--------|
| Build | 1 | 1 | 0 | 0 |
| Unit test scenarios (TS1–TS10) | 10 | 10 | 0 | 0 |
| Code quality (format, vet, dependency check) | 3 | 3 | 0 | 0 |
| E2E regression | 1 | 0 | 1 | 0 |
| Manual terminal checks | 2 | 0 | 0 | 2 |
| Performance (linear-time tests) | 1 | 1 | 0 | 0 |
