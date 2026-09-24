# Implementation Plan: Mouse Support (mouse-support)

## Overview

Adds left-button mouse interaction to the two file panes: a click moves the
cursor (activating the pane when needed), a drag marks a contiguous range on
top of the existing marks, and a double-click performs the Enter action. All
logic lives in `internal/ui`. Three small components (hit test, pane drag
marking, double-click detection) are consumed by one Model-level mouse
controller.

## Technology Stack

- **Language**: Go (toolchain as declared in `go.mod`)
- **Framework**: Bubble Tea (existing dependency) — delivers mouse events to
  the Model's update loop through the cell-motion mouse reporting option that
  `cmd/duofm/main.go` already enables (kept unchanged, NFR5)
- **New dependencies**: none. Project license (MIT) is unaffected; there is
  no new dependency license to record.

## Layer Structure

| Layer | Location | Responsibility | May depend on |
|-------|----------|----------------|---------------|
| Event routing | `internal/ui/model_update.go` | Hands mouse events to the controller; every other message path is untouched | Mouse controller |
| Mouse controller | `internal/ui/model_update_mouse.go` (Model methods) + gesture state fields in `internal/ui/model.go` | Input gating, input filtering, the press/motion/release gesture state machine, invoking existing Model/Pane operations | Hit test, Pane drag marking, Double-click detection, existing Model/Pane methods |
| Hit test | `internal/ui/mouse_hittest.go` | Pure mapping of screen coordinates to region / pane / entry; clamped drag-row mapping; read-only adapters from current Pane/Model layout state | Existing Pane layout accessors (read-only) |
| Pane drag marking | `internal/ui/pane_drag.go` (Pane methods) | Mark-set computation for a drag range on one pane | Existing Pane state |
| Double-click detection | `internal/ui/mouse_doubleclick.go` | Press-sequence timing with an injectable time source | `PanePosition` only |

Dependency rules:

- Lower layers never call the controller and never read the Model's gesture
  state.
- Nothing is added to or changed in `internal/fs`.
- Rendering code (`pane_render.go`, `model_view.go`) is not modified.

## Shared Components

All identifiers below are unexported members of package `ui`. Names and
parameter order are fixed so that producer and consumer tasks can be written
in parallel against them.

| Component | Responsibility | Contract (pre/postcondition) | Built by | Used by |
|-----------|----------------|------------------------------|----------|---------|
| `paneHitLayout` | Layout snapshot of one pane | A value carrying three integers: scroll offset, visible entry lines, entry count | task0001 | task0004 |
| `hitKind` with values `hitNone`, `hitNonEntry`, `hitEntry`; `mouseHit` | Classification result | `mouseHit` carries: kind; pane (`LeftPane` / `RightPane`, meaningful unless kind is `hitNone`); index (meaningful only for `hitEntry`) | task0001 | task0004 |
| `hitTest(x, y, width, height, left, right) → mouseHit` | Pure screen classification (`left` / `right` are `paneHitLayout`) | Pre: none (any integers accepted). Post: rules H1–H5 below; the result depends on the arguments only | task0001 | task0004 |
| `clampedEntryIndex(y, layout) → (index, ok)` | Pure drag-row mapping | Post: `ok` is false iff the pane displays no entry (entry count ≤ scroll offset, or visible lines ≤ 0). Otherwise `index` = scroll offset + (y − 4), clamped into [scroll offset, min(scroll offset + visible lines, entry count) − 1]. Any integer y is accepted; no x input exists | task0001 | task0004 |
| Pane method `hitLayout() → paneHitLayout` | Adapter | Scroll offset = the pane's current scroll offset; visible lines = `getVisibleLines()` (bg split already reflected); entry count = number of displayed (post-filter) entries. Read-only | task0001 | task0004 |
| Model method `mouseHitAt(x, y) → mouseHit` | Adapter | Uses the Model's current width and height and both panes' `hitLayout()`. Read-only. If either pane is absent, returns kind `hitNone` | task0001 | task0004 |
| Model method `mouseDragIndexAt(pane, y) → (index, ok)` | Adapter | `clampedEntryIndex` with the given pane's current `hitLayout()`. Read-only. Absent pane → ok is false | task0001 | task0004 |
| Pane method `snapshotMarks() → name set` | Baseline capture | Returns an independent copy of the pane's current mark set, in the same name→marked representation the pane uses internally; later changes to either copy do not affect the other | task0002 | task0004 |
| Pane method `applyDragMarks(anchor, target, baseline) → bool` | Range marking | Pre: `baseline` came from `snapshotMarks()`. If anchor or target is outside [0, entry count) → returns false and changes nothing. Otherwise: mark set := baseline ∪ {names of displayed entries whose index lies between anchor and target inclusive, excluding the parent entry `..`}; cursor := target; scroll offset unchanged; `baseline` itself is not mutated; returns true | task0002 | task0004 |
| `doubleClickWindow` | Detection window | 500 ms; the comparison is inclusive (elapsed ≤ 500 ms qualifies) | task0003 | task0004 |
| `clickKey` | Press target identity | A value carrying pane position, directory path, entry name; two keys are equal iff all three are equal | task0003 | task0004 |
| `newDoubleClickDetector(now) → detector reference` | Construction | `now` is a function returning the current time; when absent, the system clock is used. The returned reference is the `doubleClickDetector` type | task0003 | task0004 |
| detector method `press(key) → bool` | Double-click decision | Returns true iff a previous press is recorded AND its key equals `key` AND 0 ≤ elapsed ≤ `doubleClickWindow`. On true the record is cleared; on false this press becomes the record | task0003 | task0004 |
| detector method `reset()` | Sequence restart | Clears the record | task0003 | task0004 |

Hit test rules (`hitTest`), with 0-based coordinates and integer division:

- **H1**: kind `hitNone` when y ≤ 0 (title bar), y ≥ height − 1 (status bar
  or beyond), x < 0, or x ≥ 2 × (width / 2) (the leftover column of an odd
  width, or beyond).
- **H2**: otherwise the pane is `LeftPane` when x < width / 2, else
  `RightPane`; the layout used is that pane's.
- **H3**: rows 1–3 (path header, info header, border) → `hitNonEntry`.
- **H4**: for y ≥ 4 with r = y − 4: when r < visible lines AND scroll offset
  + r < entry count → `hitEntry` with index = scroll offset + r.
- **H5**: every other row inside the pane (blank rows below the last entry,
  the `(No matches)` row, bg split separator and output rows, the pane's
  trailing row) → `hitNonEntry`.

## Conventions

- **Identifier prefixes (collision prevention inside one package)**: new
  package-level identifiers carry their component's naming (`hit…` /
  `mouseHit…` for the hit test, `…DragMarks` / `drag…` for drag marking,
  `doubleClick…` / `clickKey` for detection, `mouse…` / `handleMouse` for the
  controller). Test-only identifiers use a per-task prefix — task0001
  `htTest`, task0002 `dragTest`, task0003 `dcTest`, task0004 `mouseTest` —
  and test function names start with `TestHitTest`, `TestPaneDragMarks`,
  `TestDoubleClickDetector`, `TestHandleMouse` respectively.
- **Contract-only references**: code outside a component's own file uses only
  the identifiers pinned in Shared Components.
- **Coordinates**: 0-based screen cells exactly as delivered by the event; no
  other layer converts them.
- **Ignored input**: mouse handling never creates dialogs or status messages
  of its own. An event that is not applicable is ignored: no change to the
  active pane, cursors, marks, scroll offsets, modal state or status message,
  and no command is returned. Messages produced by the reused Enter path are
  that path's existing behavior.
- **Formatting**: gofmt; comment language follows the surrounding files.

## Cross-task Design Decisions

### D1: Parallel implementation of shared components

- All tasks start from the same base, so the consumer (task0004) cannot rely
  on producer code (task0001–task0003) being present in its worktree.
- When a producer component is absent from its worktree, task0004 implements
  it at the producer's own file path (`mouse_hittest.go`, `pane_drag.go`,
  `mouse_doubleclick.go`), strictly to the Shared Components contract: the
  pinned identifiers plus private helpers inside that same file. It does not
  create the producer's test file.
- Using the same path means any overlap surfaces as a textual merge conflict
  on one file, never as duplicate declarations spread over different files.
  Such overlap is resolved by the implementer's parent-side-adoption
  protocol.
- Both sides implement the identical contract, so the adopted version must
  pass both tasks' tests. The later-merging task re-runs its own tests after
  adoption and corrects any contract deviation inside that shared file.
- Affected tasks: task0001, task0002, task0003, task0004.

### D2: Layout values and evaluation moment

- The hit test uses the renderer's layout values: pane width = width / 2,
  one title row, three pane header rows (first entry row = 4), and
  `getVisibleLines()`, which already folds in `bgSplitHeights()`.
- These constants are defined once, in `mouse_hittest.go`. Rendering files
  are not refactored.
- Every mouse event is evaluated against the Model state at the moment it is
  processed, before any state change that the same event causes (for
  example, the pane activation performed by that press).
- Affected tasks: task0001, task0004.

### D3: Press-driven gesture model

- **Press on an entry**:
  - The click effect is applied immediately: the pane is activated if
    inactive, its cursor moves to the entry, and marks are unchanged.
  - The detector is then consulted with key (pane, pane directory, entry
    name).
  - On true, the existing Enter action runs for that entry, and no drag is
    armed.
  - On false, a drag session is armed with anchor = entry index and
    baseline = `snapshotMarks()` taken at that press.
- **Motion with the left button held**, while a session is armed:
  - The target is `mouseDragIndexAt(session pane, y)`; x is never used.
  - The drag becomes active the first time target ≠ anchor. From then on,
    every update applies `applyDragMarks(anchor, target, baseline)`, even
    when the target returns to the anchor.
  - When the drag becomes active, the detector is reset.
- **Release**: processed as a final motion update, so a release on a
  different row completes the range. The session then ends. A release on the
  anchor row with no activation is a pure click.
- **Left press not on an entry** (a pane's non-entry area, the title bar, the
  status bar, or outside both panes): resets the detector and never arms a
  session.
- Affected tasks: task0002, task0003, task0004.

### D4: Double-click sequence semantics

- The window is 500 ms, inclusive.
- The "previous press" is the immediately preceding left press that the
  controller processed: a press on another entry or outside the entries
  starts a new sequence.
- A successful double-click clears the record, so a quick third press starts
  a new sequence and does not repeat Enter.
- The time source is injectable: unit tests supply a fake source through
  `newDoubleClickDetector`.
- Affected tasks: task0003, task0004.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Hit test drifts from rendering when the layout changes later | Low | Medium | Single set of named layout constants plus `getVisibleLines()`; hit-test unit tests pin the row→entry mapping (TS-1..TS-4) |
| Stand-in and canonical versions of a shared component diverge at merge (D1) | Medium | Medium | Identical pinned contract; the later-merging task re-runs its tests after parent-side adoption |
| Identifier collisions between tasks writing into the same package in parallel | Medium | Low | Prefix convention (Conventions) |
| Directory contents change during a drag (auto-refresh), so the anchor index points at a different entry | Low | Low | Sessions are index-based; `applyDragMarks` rejects out-of-range indices instead of failing |

## Open Questions

- [ ] D4 reads FR6's "そのエントリへの前回の押下" as "the immediately
      preceding left press". Under this reading, pressing entry A, then B,
      then A within 500 ms does not trigger Enter.
- [ ] NFR5 has no SPEC test scenario. VERIFICATION.md covers it with a manual
      check: no diff in `cmd/duofm/main.go`.
