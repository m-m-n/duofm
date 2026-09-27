# Implementation Plan: minibuffer-grapheme-cursor-highlight

## Overview

The minibuffer cursor moves one grapheme cluster per Left/Ctrl+B/Right/Ctrl+F press, and View renders the whole grapheme cluster containing the cursor as one reversed segment on the fit path and on every scroll path. The cursor position stays a rune offset and editing operations stay rune-wise.

Per-task detail (the View step-by-step changes, the line-builder contract, test fixtures) lives in `tasks/task0001.md`. This document holds only the feature-wide decisions every task touching the minibuffer (including later rework tasks) must follow.

## Technology Stack

- **Language**: Go (module and toolchain as declared in go.mod)
- **TUI framework**: Bubble Tea — key messages delivered to the minibuffer (existing)
- **Styling / width basis**: Lip Gloss — display-width measurement (the width basis) and the reverse attribute used for the cursor highlight (existing, unchanged)
- **Grapheme segmentation**: rivo/uniseg v0.4.7 — default extended grapheme cluster segmentation. Already required by go.mod as an indirect dependency (it is what the width basis itself segments with); this feature imports it directly from internal/ui, so its go.mod entry becomes a direct requirement at the same version.

### Dependency licenses

- github.com/rivo/uniseg v0.4.7 — MIT. Project license is MIT; permissive dependency, compatible. Promoted from indirect to direct; no new module is added (NFR4).

## Layer Structure

All changes stay inside the Minibuffer component (internal/ui/minibuffer.go). No other package or file in internal/ui changes; callers such as tab completion (which passes rune positions to SetCursorPos) are unaffected.

| Layer | Responsibility | Allowed dependencies |
|-------|----------------|----------------------|
| Key handling (HandleKey) | State transitions of the input text and cursorPos | Grapheme cluster locator |
| Rendering (View, line builder, width-unit helpers, fit guard) | Assemble one display line within the content width, with the cursor highlight | Grapheme cluster locator, width basis |
| Grapheme cluster locator (new, pure function) | Report grapheme cluster bounds in rune offsets | rivo/uniseg only |

Dependency direction: key handling and rendering call the locator; the locator depends on nothing else in the package and never uses the width basis. The locator holds no state.

## Shared Components

| Component | Responsibility | Contract (pre/postcondition) | Used by tasks |
|-----------|----------------|------------------------------|---------------|
| Grapheme cluster locator | Given the input's runes and a rune index p, return the rune-offset bounds [start, end) of the grapheme cluster that contains rune p | Pre: 0 <= p < rune count of the input (callers never ask about the end position; the end position has no cluster). Post: start <= p < end <= rune count; [start, end) is exactly one cluster of rivo/uniseg's default segmentation applied to the WHOLE input starting at its first rune (never to a sub-slice, because segmentation is context-dependent: regional-indicator pairing, ZWJ sequences); a leading isolated zero-width rune (a combining mark or ZWJ at the input start) is its own cluster exactly as that segmentation yields; running time is linear in the input length (one forward pass that stops at the containing cluster); inputs are not modified. | task0001 (cursor movement keys and View) |

Any later task that needs cluster bounds in the minibuffer reuses this locator instead of adding a second segmentation.

## Conventions

- **Position units**: cursorPos, SetCursorPos, CursorPos and every display-range index inside View are rune offsets into the input. Grapheme clusters are expressed as rune-offset ranges; byte offsets never leave the locator.
- **Width basis**: Lip Gloss display-width measurement, unchanged (NFR2). Every width compared against a budget is measured on actually assembled text through the existing estimate-then-verify fit guard; cluster widths are measured on the cluster's own text, once.
- **Segmentation basis**: exactly one segmentation for cluster purposes — rivo/uniseg default grapheme clusters, reached only through the grapheme cluster locator (NFR2). The existing width units (a positive-width rune plus following zero-width runes) remain in use solely for trimming the window edge away from the cursor.
- **Highlight**: the existing reverse attribute; at most one reversed segment per rendered line. The trailing block cursor (a reversed single space when the cursor is at the end) is unchanged.
- **Tests that inspect reversed output**: they fix the default renderer's color profile to 256 colors and restore the previous profile at test cleanup, and they do not run in parallel (the profile is process-global).
- **Error handling**: no new error paths. Positions passed to the locator are always in range because cursorPos is clamped by existing code and the end position is handled separately by callers.
- **Decision IDs**: this feature's decisions are numbered GD1–GD5. Existing code comments citing CD2–CD5 refer to earlier minibuffer features' documents; those citations stay valid and are not renumbered.

## Cross-task Design Decisions

### GD1: cursorPos stays a rune offset; movement snaps to cluster bounds

- **Decision**: the cursor is still stored and exposed as a rune offset. Left/Ctrl+B move to the start of the cluster containing the rune just before the cursor; Right/Ctrl+F move to the end of the cluster containing the rune at the cursor. A cursor placed inside a cluster by SetCursorPos (e.g. tab completion) is accepted as-is.
- **Rationale**: FR1, FR5, SPEC A1 — external callers and editing operations keep their rune semantics.
- **Affected tasks**: task0001

### GD2: the highlight unit is the cursor's whole grapheme cluster

- **Decision**: when the cursor is not at the end, the reversed segment is the whole cluster [start, end) containing the cursor rune, including when the cursor sits inside the cluster. Because highlight boundaries then coincide with cluster boundaries, the reverse escape sequences never split a cluster, so the measured width of the highlighted line equals that of the plain line.
- **Rationale**: FR2, FR3 — this is what removes both the split rendering and the width inflation that caused truncation.
- **Affected tasks**: task0001

### GD3: scroll paths treat the cursor's cluster as one indivisible unit; the far window edge is unchanged

- **Decision**: in every scroll path, the segments before and after the cursor are split at the cursor cluster's start and end, the cursor's width is the whole cluster's width, and the cluster is either rendered whole or the input is not rendered at all. The window edge away from the cursor keeps the existing width-unit trimming and is not aligned to cluster boundaries.
- **Rationale**: FR4, SPEC A4 (confirmed decision req.window-edge-split = cursor_grapheme_only).
- **Affected tasks**: task0001

### GD4: editing operations stay rune-wise

- **Decision**: character/space insertion, Backspace, Delete, Ctrl+K, Ctrl+U, Ctrl+A, Ctrl+E, SetInput, SetCursorPos and CursorPos keep their current rune-offset behavior. After an edit leaves the cursor inside a cluster, GD2 applies on the next View.
- **Rationale**: FR5, SPEC A5 (confirmed decision req.edit-ops-scope = keep_rune_wise_editing).
- **Affected tasks**: task0001

### GD5: linear time is preserved

- **Decision**: View locates the cursor's cluster once per call and measures its width once; the fit guard's bound (at most 3 re-measurements) is unchanged. Cursor movement locates one cluster per key press. All of these are linear in the input length.
- **Rationale**: NFR3; the existing time-bound regression tests must keep passing.
- **Affected tasks**: task0001

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Regression tests pass vacuously because the default renderer in a non-terminal test run emits no escape sequence for the reverse attribute (a split highlight is then invisible to width measurement and to reversed-run inspection) | High (if not addressed) | High | Tests that inspect reversed output or depend on highlight-induced width fix the 256-color profile, and first assert (fatally) that the reverse attribute actually emits an escape sequence under that profile |
| Cluster bounds from the locator disagree with the width basis's own segmentation | Low | Medium | Both resolve to the single rivo/uniseg version selected by go.mod; the locator segments the whole input, never a sub-slice |
| Global color-profile change leaks into other tests | Medium | Medium | Restore the previous profile at test cleanup; no parallel execution for those tests |
| Linear-time regression from per-rune or repeated cluster scans | Low | High | GD5; existing time-bound tests (TS1–TS4, TS7 in minibuffer_test.go) stay mandatory |
| Users perceive a movement change on plain text | Low | Low | Single-rune clusters move exactly as before; the existing movement tests on ASCII input remain unchanged |

## Open Questions

None.
