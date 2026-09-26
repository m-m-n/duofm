# Implementation Plan: press-scroll-click-mark

## Overview

A left press on an entry of an inactive pane whose activation shows the background-output split keeps that pane's pre-press scroll offset, so a press and release on the same row stays a click and adds no marks. The change is confined to the mouse entry-press path in `internal/ui/model_update_mouse.go`; the shared scroll logic is not modified.

## Technology Stack

- **Language**: Go (toolchain and container image as in workflow.yaml `project.components.duofm`)
- **Framework**: Bubble Tea (existing TUI event loop; no change)
- **New dependencies**: none. No license impact (project license: MIT).

## Layer Structure

| Layer | Members (all in `internal/ui`) | Change in this feature |
|-------|--------------------------------|------------------------|
| Mouse gesture layer | entry-press handler (handleEntryPress), press dispatcher, drag update (updateDrag) | entry-press handler only |
| Mouse read-only mapping | hit test, drag-row mapping (clampedEntryIndex), pane hit layout adapter | none |
| Double-click detection | click-key detector | none |
| Shared pane/scroll layer | Model.switchToPane, Model.syncPaneBgOutputState, Pane.SetBgOutputActive, Pane.adjustScroll, Pane.EnsureCursorVisible | none |

Allowed dependency direction: the mouse gesture layer calls into the shared pane/scroll layer and the read-only mapping layer. The shared layer never refers to the mouse layer. This feature adds no new dependency edge.

## Shared Components

This feature has a single task; no component is built by one task and consumed by another. The table pins the existing components the task relies on, whose contracts stay unchanged.

| Component | Responsibility | Contract (pre/postcondition) | Used by tasks |
|-----------|----------------|------------------------------|---------------|
| Model.switchToPane | Activate a pane and sync each pane's background-output split flag | Unchanged. Post: the newly active pane shows the split when a background command is running or closing on it; each pane's scroll offset is re-adjusted against that pane's current cursor and its new visible-line count | task0001 (caller) |
| Hit test / drag-row mapping | Convert a screen row to an entry index | Unchanged. Always uses the pane's scroll offset at the time of the call; drag-row mapping clamps into the currently displayed range | task0001 (relied on) |
| Double-click detector | Report a double click when the same click key (pane, directory, entry name) is pressed twice within the window | Unchanged | task0001 (relied on) |
| Pane.EnsureCursorVisible | Scroll so the cursor is inside the visible range | Unchanged. Post: when the cursor is below the visible range, it lands on the bottom visible row | task0001 (caller) |

## Conventions

- Tests drive the behavior through Model.Update with mouse messages at the 80x24 geometry used by `internal/ui/model_update_mouse_test.go`, reusing that file's existing helpers (test model builder, row helper, mark assertions, fake clock).
- No new exported identifiers.
- Code is gofmt-formatted and go vet clean.

## Cross-task Design Decisions

### D1: Restore the pre-press scroll offset in the entry-press path (keep_scroll)

- Decision: the entry-press handler records the clicked pane's scroll offset before pane activation, restores it after activation, then moves the cursor to the pressed entry and applies the existing ensure-visible adjustment only when the pressed entry is outside the visible range of the restored offset.
- Basis: SPEC.md "Implementation Approach" (keep_scroll) and REQUIREMENTS.md 14.1.
- Affected tasks: task0001.

### D2: Shared scroll logic and keyboard pane switching are not modified

- Decision: Pane.adjustScroll, Pane.SetBgOutputActive, Model.syncPaneBgOutputState and Model.switchToPane keep their current code and behavior. Keyboard pane switching (move_left / move_right actions) keeps its current scroll behavior on bg-split activation.
- Basis: NFR1, NFR2.
- Affected tasks: task0001, and any later rework task in this feature.

### D3: Drag mapping and double-click detection are not modified

- Decision: the drag update, drag-row mapping, mark application and double-click detector stay as they are. FR3 and FR4 are met through D1 alone.
- Basis: SPEC.md "Implementation Approach".
- Affected tasks: task0001.

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Restore runs before activation, or ensure-visible runs from the shifted offset instead of the restored one | Low | High | TS-1 asserts scroll offset 0; TS-4 asserts the bottom-row placement |
| Restore changes behavior when the pressed pane is already active | Low | Medium | Restore is a no-op when no activation happens; existing mouse tests (NFR3) |
| Fix leaks into the deactivated pane or the keyboard path | Low | Medium | D2; TS-10 pins keyboard activation scrolling; TS-11 diff check |

## Open Questions

- [ ] (non-blocking) SPEC.md NFR1 names "Tab" as the keyboard pane switch. The default key bindings switch panes with move_left / move_right (H / Left, L / Right); there is no Tab binding for pane switching. This plan treats NFR1 as keyboard pane switching through move_left / move_right.
