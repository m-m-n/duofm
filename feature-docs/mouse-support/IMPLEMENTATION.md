# Implementation Plan: Mouse Support (mouse-support)

## Overview

Adds left-button mouse interaction to the two file panes:

- A click moves the cursor, activating the pane when needed.
- A drag marks a contiguous range on top of the existing marks.
- A double-click performs the Enter action.

All Go logic lives in `internal/ui` and is built by task0001. task0002 adds
the E2E scenarios for the same behavior, plus the E2E runner entry point.

## Technology Stack

- **Language**: Go (toolchain as declared in `go.mod`)
- **Framework**: Bubble Tea (existing dependency). It delivers mouse events
  to the Model's update loop through the cell-motion mouse reporting option
  that `cmd/duofm/main.go` already enables (kept unchanged, NFR5).
- **E2E**: the existing tmux-based bash suite under `test/e2e`
- **New dependencies**: none. The project license (MIT) is unaffected, and
  no new dependency license needs recording.

## Layer Structure

| Layer | Location | Responsibility | May depend on |
|-------|----------|----------------|---------------|
| Event routing | `internal/ui/model_update.go` | Hands mouse events to the controller; every other message path is untouched | Mouse controller |
| Mouse controller | `internal/ui/model_update_mouse.go` (Model methods), plus gesture state fields in `internal/ui/model.go` | Input gating, input filtering, the press/motion/release gesture state machine, invoking existing Model/Pane operations | Hit test, Pane drag marking, Double-click detection, existing Model/Pane methods |
| Hit test | `internal/ui/mouse_hittest.go` | Pure mapping of screen coordinates to region / pane / entry; clamped drag-row mapping; read-only adapters from current Pane/Model layout state | Existing Pane layout accessors (read-only) |
| Pane drag marking | `internal/ui/pane_drag.go` (Pane methods) | Mark-set computation for a drag range on one pane | Existing Pane state |
| Double-click detection | `internal/ui/mouse_doubleclick.go` | Press-sequence timing with an injectable time source | `PanePosition` only |
| E2E scenarios | `test/e2e/scripts/tests/mouse_tests.sh`, runner registration, image default command | Drive the built binary through tmux and observe screen output | Runtime behavior only (D5) |

Dependency rules:

- Lower layers never call the controller and never read the Model's gesture
  state.
- Nothing is added to or changed in `internal/fs`.
- Rendering code (`pane_render.go`, `model_view.go`) is not modified.

## Shared Components

No file is shared between the two tasks. Every component below is built and
consumed inside task0001, and these contracts are its internal design. All
identifiers are unexported members of package `ui`, and their names are
fixed.

| Component | Responsibility | Contract (pre/postcondition) |
|-----------|----------------|------------------------------|
| `paneHitLayout` | Layout snapshot of one pane | A value carrying three integers: scroll offset, visible entry lines, entry count |
| `hitKind` with values `hitNone`, `hitNonEntry`, `hitEntry`; `mouseHit` | Classification result | `mouseHit` carries: kind; pane (`LeftPane` / `RightPane`, meaningful unless kind is `hitNone`); index (meaningful only for `hitEntry`) |
| `hitTest(x, y, width, height, left, right) → mouseHit` | Pure screen classification (`left` / `right` are `paneHitLayout`) | Pre: none (any integers accepted). Post: rules H1–H5 below; the result depends only on the arguments |
| `clampedEntryIndex(y, layout) → (index, ok)` | Pure drag-row mapping | Post: `ok` is false iff the pane displays no entry (entry count ≤ scroll offset, or visible lines ≤ 0). Otherwise `index` = scroll offset + (y − 4), clamped into [scroll offset, min(scroll offset + visible lines, entry count) − 1]. Any integer y is accepted, and there is no x input |
| Pane method `hitLayout() → paneHitLayout` | Adapter | Scroll offset = the pane's current scroll offset; visible lines = `getVisibleLines()` (bg split already reflected); entry count = number of displayed (post-filter) entries. Read-only |
| Model method `mouseHitAt(x, y) → mouseHit` | Adapter | Uses the Model's current width and height and both panes' `hitLayout()`. Read-only. If either pane is absent, the kind is `hitNone` |
| Model method `mouseDragIndexAt(pane, y) → (index, ok)` | Adapter | Applies `clampedEntryIndex` to the given pane's current `hitLayout()`. Read-only. Absent pane → ok is false |
| Pane method `snapshotMarks() → name set` | Baseline capture | Returns an independent copy of the pane's current mark set, in the same name→marked representation the pane uses internally; later changes to either copy do not affect the other |
| Pane method `applyDragMarks(anchor, target, baseline) → bool` | Range marking | Pre: `baseline` came from `snapshotMarks()`. If anchor or target is outside [0, entry count), returns false and changes nothing. Otherwise: mark set := baseline ∪ {names of displayed entries whose index lies between anchor and target inclusive, excluding the parent entry `..`}; cursor := target; scroll offset unchanged; `baseline` itself is not mutated; returns true |
| `doubleClickWindow` | Detection window | 500 ms; the comparison is inclusive (elapsed ≤ 500 ms qualifies) |
| `clickKey` | Press target identity | A value carrying pane position, directory path and entry name; two keys are equal iff all three are equal |
| `newDoubleClickDetector(now) → detector reference` | Construction | `now` is a function returning the current time; when absent, the system clock is used. The returned reference is of type `doubleClickDetector` |
| detector method `press(key) → bool` | Double-click decision | Returns true iff a previous press is recorded AND its key equals `key` AND 0 ≤ elapsed ≤ `doubleClickWindow`. On true the record is cleared; on false this press becomes the record |
| detector method `reset()` | Sequence restart | Clears the record |

Hit test rules (`hitTest`), with 0-based coordinates and integer division:

- **H1**: `hitNone` when y ≤ 0 (title bar), y ≥ height − 1 (status bar or
  beyond), x < 0, or x ≥ 2 × (width / 2) (the leftover column of an odd
  width, or beyond).
- **H2**: otherwise the pane is `LeftPane` when x < width / 2, and
  `RightPane` otherwise. The layout used is that pane's.
- **H3**: rows 1–3 (path header, info header, border) are `hitNonEntry`.
- **H4**: for y ≥ 4, let r = y − 4. When r < visible lines AND scroll
  offset + r < entry count, the result is `hitEntry` with index = scroll
  offset + r.
- **H5**: every other row inside the pane is `hitNonEntry`. This covers
  blank rows below the last entry, the `(No matches)` row, the bg split
  separator and output rows, and the pane's trailing row.

## Conventions

- **Coordinates**: 0-based screen cells, exactly as the event delivers them.
  No other layer converts them.
- **Ignored input**: mouse handling never creates dialogs or status messages
  of its own. An event that does not apply is ignored:
  - nothing changes in the active pane, cursors, marks, scroll offsets,
    modal state or status message;
  - no command is returned.

  Messages produced by the reused Enter path are that path's existing
  behavior.
- **Formatting**: gofmt. Comment language follows the surrounding files.

## Cross-task Design Decisions

### D1: Task structure

- task0001 owns every Go change. That covers all components in Shared
  Components, the Model state, and the event routing.
- task0002 owns the E2E scenario file, the runner registration and the E2E
  image's default command.
- The two tasks share no file and run in parallel. task0002 depends on
  task0001 only through the runtime behavior pinned in D5. Its mouse
  scenarios pass only on the integrated build.
- Affected tasks: task0001, task0002.

### D2: Layout values and evaluation moment

- The hit test uses the renderer's layout values:
  - pane width = width / 2;
  - one title row;
  - three pane header rows (first entry row = 4);
  - `getVisibleLines()`, which already includes `bgSplitHeights()`.
- These constants are defined once, in `mouse_hittest.go`. Rendering files
  are not refactored.
- Every mouse event is evaluated against the Model state at the moment it is
  processed, before any state change the same event causes. An example is
  the pane activation performed by that press.
- Affected tasks: task0001 (implementation), task0002 (row arithmetic in its
  scenarios, D5).

### D3: Press-driven gesture model

**Press on an entry**:

1. Apply the click effect at once:
   - activate the pane if it is inactive;
   - move its cursor to the entry;
   - leave the marks unchanged.
2. Consult the detector with key (pane, pane directory, entry name).
   - On true: run the existing Enter action for that entry, and arm no drag.
   - On false: arm a drag session with anchor = entry index and
     baseline = `snapshotMarks()`, taken at this press.

**Motion with the left button held, while a session is armed**:

- The target is `mouseDragIndexAt(session pane, y)`. x is never used.
- The drag becomes active the first time the target differs from the anchor.
- From then on, every update applies `applyDragMarks(anchor, target,
  baseline)`, even when the target returns to the anchor.
- When the drag becomes active, the detector is reset.

**Release**:

- Processed as a final motion update, so a release on a different row
  completes the range.
- The session then ends.
- A release on the anchor row with no activation is a pure click.

**Left press not on an entry**: covers a pane's non-entry area, the title
bar, the status bar, and the space outside both panes.

- The detector is reset.
- No session is armed.

Affected tasks: task0001 (implementation), task0002 (the scenarios observe
these effects).

### D4: Double-click sequence semantics (decided)

- The window is 500 ms, inclusive.
- **Decided**: FR6's "そのエントリへの前回の押下" means the immediately
  preceding left press, and that press must be on the same entry. "Same
  entry" means the same pane, the same directory and the same entry name.
- A press on another entry, or outside the entries, starts a new sequence.
  Pressing A, then B, then A within 500 ms does not trigger Enter.
- A successful double-click clears the record. A quick third press therefore
  starts a new sequence and does not repeat Enter.
- The time source is injectable. Unit tests supply a fake source through
  `newDoubleClickDetector`.
- Affected tasks: task0001 (implementation), task0002 (the double-click
  scenario sends one press/release pair on the same row twice, well within
  500 ms).

### D5: Observable contract relied on by the E2E scenarios

task0002's scenarios rely only on these observable effects of task0001:

- **Input**: SGR-encoded left-button events (press, motion with the button
  held, release) are handled as described in D3. Coordinates are 1-based on
  the wire and 0-based once they arrive in the Model.
- **Screen geometry**: row 0 is the title bar. Rows 1–3 are the pane
  headers. Entry index i of a pane with scroll offset s appears on row
  4 + i − s. The left pane spans columns [0, width / 2).
- **Cursor**: the status bar's "current/total" position reports the cursor
  after a click. Only the existing rendering is used.
- **Marks**: the pane header's existing "Marked N/…" count reports the marks
  after a drag.
- **Double-click on a directory**: the pane enters that directory, and its
  entries are displayed.

Affected tasks: task0001, task0002.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Hit test drifts from rendering when the layout changes later | Low | Medium | One set of named layout constants plus `getVisibleLines()`; hit-test unit tests pin the row→entry mapping (TS-1..TS-4) |
| SGR sequences sent through tmux do not reach duofm as mouse events in the E2E container, so the E2E scenarios cannot pass | Medium | Low | The scenarios are conditional per the SPEC. VERIFICATION.md defines how to judge them and the unit-test fallbacks (TS-5 / TS-8 / TS-13) |
| Directory contents change during a drag (auto-refresh), so the anchor index points at a different entry | Low | Low | Sessions are index-based; `applyDragMarks` rejects out-of-range indices instead of failing |

## Open Questions

- [x] FR6 "previous press" interpretation: decided (D4). It is the
      immediately preceding left press, on the same entry.
- [ ] NFR5 has no SPEC test scenario. VERIFICATION.md covers it with a
      manual check: `cmd/duofm/main.go` has no diff.
