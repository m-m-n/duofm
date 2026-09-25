# Implementation Plan: header-wrap-hittest

## Overview

Keep pane header line 1 on exactly one screen row in every pane view, and make
the pane header row count a single definition that both the renderer's
visible-line computations and the mouse hit test reference. The work is one
task (task0001); this document pins the cross-component contracts that task —
and any rework task appended later — must honor.

## Technology Stack

- **Language**: Go (module go version 1.25)
- **TUI framework**: Bubble Tea + lipgloss (existing) — pane rendering
- **Display width**: go-runewidth (existing dependency) — every display-width
  measurement and truncation decision (NFR2)
- **New dependencies**: none. No license check is triggered
  (`project.license: MIT` unaffected).

## Layer Structure

All changes stay inside `internal/ui`. Three roles are involved:

| Role | Location | Responsibility |
|------|----------|----------------|
| Pane layout definition | `internal/ui/pane.go` | Owns the shared pane header row count and the visible-line / bg-split height computations |
| Pane renderer | `internal/ui/pane_render.go` | Renders header rows and entry rows for the normal, bg output split and dimmed views |
| Mouse hit test | `internal/ui/mouse_hittest.go` | Maps screen coordinates to panes / entries; stays a pure function of its arguments |

Allowed dependency directions: the renderer and the hit test both depend on
the pane layout definition. The hit test never depends on rendered output, and
the renderer never depends on the hit test.

## Shared Components

| Component | Responsibility | Contract (pre/postcondition) | Used by tasks |
|-----------|----------------|------------------------------|---------------|
| Pane header row count (`paneHeaderRows`) | The single definition of how many screen rows a pane header occupies: header line 1, header line 2, border row | Value: 3. Precondition it relies on: every header row renders as exactly one screen row (header line 1 by the contract below; header line 2 and the border row are already bounded by existing code). Postconditions: hit-test first entry row = title row + 1 + this value; normal-view visible lines, bg-split content height and dimmed-view entry rows = pane height − this value − 1 trailing row (existing minimum clamps unchanged). No other literal for the header row count remains in `internal/ui` layout, rendering or hit-test code. | task0001 |
| Header line 1 renderer (`renderHeaderLine1`, pane method, no arguments → string) | Composes filter indicator, `[H]` indicator, path and git branch into header line 1 | Content width W = max(pane width − 4, 0). Postconditions for every pane state: display width (go-runewidth) ≤ W; contains no line break; never panics; output is byte-identical to the current output whenever the current output already fits in W (NFR1). | task0001 |
| Ellipsis truncation (`truncateStringWithEllipsis(s string, maxWidth int) string`) | Shortens a string to a display width, keeping the head | Precondition: none (any integer width accepted). Postconditions: a negative maxWidth is treated as 0; result display width ≤ max(maxWidth, 0); s is returned unchanged when it fits; otherwise the head of s is kept and the result ends with "..." (only the part of "..." that fits when maxWidth ≤ 3); never panics. Behavior for maxWidth ≥ 0 is unchanged. | task0001 (also used by existing trash-entry formatting) |

## Conventions

- Display width is always measured with go-runewidth — never byte length or
  rune count (NFR2).
- A pane row's content width is pane width − 4 (row style width = pane width −
  2, plus one column of padding on each side). It is clamped to 0 before being
  passed to any truncation (FR5).
- Layout numbers used by both rendering and hit testing are named definitions
  owned by the pane layout role, never repeated literals.
- Tests are table-driven Go tests in `internal/ui` `*_test.go`, next to the
  existing ones. Tests that check row positions find rows in the rendered
  output (with ANSI escape sequences removed). They do not take row positions
  from hit-test constants.

## Cross-task Design Decisions

### D1: Fixed header row count, guaranteed by bounding header line 1

- **Decision**: The header row count stays a fixed value (3). Its truth is
  guaranteed by making header line 1 never wrap. The rendered header height is
  not measured at runtime.
- **Rationale**: The hit test is a pure function of its arguments and cannot
  see rendered output (existing design, A4). Bounding header line 1's width
  removes the only wrap source in all three views at once (NFR3).
- **Affected tasks**: task0001.

### D2: Truncation keeps the head (A2)

- **Decision**: When header line 1 is too wide, it is cut at the end and ends
  with "...". The filter indicator and `[H]` indicator come before the path,
  so they stay visible. The tail of the path (the current directory name) can
  disappear. This is the accepted result of A2.
- **Affected tasks**: task0001.

### D3: Numeric layout preserved (NFR1)

- **Decision**: The number of visible lines, bg-split heights and the first
  entry screen row must not change at any pane height. Only the source of the
  header row number changes. Existing hit-test rules (H1–H5, drag-row
  clamping) are not modified (A4).
- **Affected tasks**: task0001.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| lipgloss measures some characters (ambiguous-width, emoji sequences) wider than go-runewidth does, so a line within W by go-runewidth still wraps | Low | Medium (hit test misaligned for such paths) | Truncate by go-runewidth as NFR2 requires; full-width characters are covered by TS-3; other character classes are a known limit |
| A new test becomes vacuous or flaky because the temp directory path length or its git state differs by environment | Medium | Low | Tests build a nested path long enough to exceed the pane width regardless of the temp root, and force the git branch to empty |
| A test locating the parent-dir entry ".." matches the "..." at the end of the truncated header instead | Medium | Low | Tests locate entries by unique file names |

## Open Questions

- [ ] FR5 says "ヘッダ描画は panic せず". This plan scopes FR5 to header line 1
  and the ellipsis truncation, as TS5 names them. The border row (third header
  row) repeats a character (pane width − 2) times, and that count is negative
  at pane widths 0–1. SPEC.md does not say whether the border row is part of
  FR5, so it is out of scope for task0001.
