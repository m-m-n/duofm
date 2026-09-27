# Implementation Plan: minibuffer-shrink-loop-linear

## Overview

Rework the range-selection part of Minibuffer.View (prompt truncation, input
window selection, line assembly) so that one View call finishes in time linear
in prompt length + input length on every path, while every existing
postcondition of the View result is kept. The feature is a single task
(task0001) confined to `internal/ui/minibuffer.go` and
`internal/ui/minibuffer_test.go`.

## Technology Stack

- **Language**: Go 1.25.5 (build/test/format run through the podman-wrapped
  commands in workflow.yaml `project.components.duofm`).
- **lipgloss v1.1.0** (existing dependency) — the only display-width basis
  (its grapheme-cluster-based width measurement) and the minibuffer line style.
- **go-runewidth** (existing dependency) — not used by this feature and never
  used as a width basis (SPEC a3).
- **New dependencies**: none. `go.mod` and `go.sum` stay unchanged, so there
  is no new license to check (project license: MIT).

## Layer Structure

- One UI component: Minibuffer (`internal/ui`). All production changes stay
  inside View and its private helper functions in `internal/ui/minibuffer.go`.
- Dependency direction is unchanged: View depends only on the Minibuffer's own
  state (prompt, input, cursor position, width, visibility), the width basis,
  and the existing line style. No new package, no new exported identifier.
- Tests stay in `internal/ui/minibuffer_test.go` (same package) and drive the
  component only through its existing setters and View.

## Shared Components

| Component | Responsibility | Contract (pre/postcondition) | Used by tasks |
|-----------|----------------|------------------------------|---------------|
| (none) | The feature has a single task; no component is built by one task and consumed by another. | — | task0001 |

The feature-wide invariants under "Cross-task Design Decisions" apply to any
later task in this feature that touches View (for example a rework task
appended by review or verify).

## Conventions

- **Width basis**: every display-width judgment uses the lipgloss width
  measurement. A rune's "own width" is that measurement applied to the rune
  alone. go-runewidth is never consulted.
- **Positions**: the cursor position stays a rune index into the input; every
  window or truncation boundary View computes is a rune index into the input
  or the prompt.
- **No test seam**: no injectable width function, no exported test hook, no
  build-tag variant (NFR1).
- **Error handling**: View introduces no error values; it never panics and
  never returns a line break (existing narrow-width behaviour kept).
- **Time-bound regression tests**: run View in a separate goroutine, fail when
  a 10-second limit elapses, and log the measured View duration. The test
  names for TS-1..TS-4 start with `TestMinibufferView_LinearTime`
  (VERIFICATION.md filters on this prefix).
- **Test helpers**: the existing ui test-package helpers (bounded single-line
  assertion, ANSI stripping, panic guard) are reused, not duplicated.

## Cross-task Design Decisions

### CD1: Linear-time invariant

Within one View call, a string whose length grows with the prompt or input
length is assembled or measured at most a constant number of times,
independent of that length. Every rune's own width is measured at most once,
and every scan over runes or units is a single pass. (FR1)

### CD2: Width units

- A width unit is a maximal run of runes that starts with a rune whose own
  width is positive and continues through every immediately following rune
  whose own width is zero. Zero-width runes at the very start of a segment
  (CD3) form one leading unit.
- A unit's estimated width is the width basis applied to the unit's text
  alone; each unit is measured once.
- Window and truncation boundaries fall only on unit boundaries. A unit that
  would cross the width budget is excluded whole, never split.
- Effect: runs of zero-width runes (combining marks, variation selectors, ZWJ)
  travel with their base, so no step processes them one rune at a time, and a
  base is never separated from its combining marks. Boundary content may
  differ from the pre-change output where the old per-rune boundary fell
  inside a unit (FR2).
- Full UAX #29 grapheme segmentation is not used: the segmentation library is
  only an indirect dependency, and making it direct changes `go.mod`, which is
  outside the NFR1 change scope.

### CD3: Segments split at the cursor

The cursor highlight wraps the cursor's rune in style sequences, which make
the width measurement treat the text on each side independently. Units are
therefore computed per segment: the prompt; the input before the cursor rune;
the cursor rune alone; the input after the cursor rune. With the cursor at the
end, the whole input is one segment followed by the 1-column cursor block.
Whether style sequences are actually emitted depends on the renderer's color
profile (tests normally run without them), so this model is an estimate; CD4
makes the result correct either way.

### CD4: Estimate, then verify once (bounded guard)

- The sum of unit estimates is exact for ASCII, East Asian wide characters,
  a base with combining marks, and a base with VS16. It can exceed the
  rendered width for sequences joined across positive-width runes (ZWJ emoji
  sequences, regional-indicator pairs, Hangul jamo) and can fall short in rare
  cases (for example a prepended format character joined to a following emoji
  with VS15, a prompt ending that joins with the input start, or a cursor rune
  joined with its neighbours when no style sequences are emitted).
- The assembled line is therefore measured once with the width basis. If it
  exceeds the content width, a guard trims whole units on the side away from
  the cursor — in one step, enough units whose estimates sum to at least the
  measured excess — and re-measures, at most 3 times. If the line still
  exceeds, the window collapses to the cursor alone (the cursor rune, or the
  cursor block at the end); if even that exceeds, no input is rendered.
- On the prompt-truncation path the guard trims trailing prompt units the same
  way and finally falls back to an empty prompt.

### CD5: Preserved behaviour

Unchanged: a hidden minibuffer renders the empty string; widths 0-4 render no
prompt or input, without panic or line break; prompt + input that fit the
content width are rendered untruncated (judged by measuring the fully
assembled line, cursor highlight or block included); a prompt at least as
wide as the content width is truncated and no input or cursor is rendered; a
cursor rune wider than the remaining width renders no input; the line style
(width m.width-2, 1-column horizontal padding, colors) is unchanged.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Unit estimate falls short of the rendered width for rare sequences, overflowing the line | Low | Medium (line wraps) | CD4 guard; TS-7 exercises it |
| ZWJ / regional-indicator sequences are overestimated, so less context is shown while scrolling | Medium | Low | Permitted by FR2; existing ZWJ tests pin the visible-content cases |
| Race-detector timing on a slow container approaches the 10-second limit | Low | Medium | CD1 single-pass design; TS-6 records durations against a 2-second guideline |
| Red-phase run of TS-1..TS-4 on the pre-change View leaves CPU-bound goroutines running | High (by design) | Low | Run only the new tests during the red phase when possible |
| TS-6 needs a race-enabled, filtered test command that differs from the approved test_command string | Medium | Low | Full command written in VERIFICATION.md; the verify phase obtains approval if the command hook denies it |

## Open Questions

- [ ] NFR2 ("10 秒の上限を十分に下回る") gives no numeric threshold. This plan
  uses "each TS-1..TS-4 duration under the race detector is at most 2 seconds
  (1/5 of the limit)" as the verification guideline.
- [ ] TS-6 and TS-9 use test commands that are not the approved
  `project.components.duofm.test_command` string.
