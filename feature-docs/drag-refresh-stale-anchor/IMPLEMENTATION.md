# Implementation Plan: drag-refresh-stale-anchor

## Overview

An armed mouse drag session checks, at every motion and release, whether its
pane's displayed entry list (post-filter) still matches the list at press
time, name for name and in order. On a mismatch the session is cancelled
without touching marks; on a match the drag continues from the original
anchor, and the press-time baseline never revives marks of files that no
longer exist. The feature is delivered by a single task (task0001).

## Technology Stack

- **Language**: Go (existing module)
- **Framework**: Bubble Tea (existing TUI framework; used as-is)
- **New dependencies**: none. No license record is required; `project.license`
  (MIT) is unaffected.

## Layer Structure

| Layer | Location | Responsibility in this feature |
|-------|----------|--------------------------------|
| Mouse gesture state machine (Model) | `internal/ui/model_update_mouse.go` | Owns the drag session lifecycle: arm on press, verify on motion/release, apply drag marks, cancel. |
| Pane drag helpers (Pane) | `internal/ui/pane_drag.go` | Read-only snapshots and comparisons of the pane's displayed and full entry lists; drag mark application (existing). |
| Directory refresh paths | `internal/ui/model_update.go`, `internal/ui/model_update_dialog.go`, `internal/ui/model.go`, `internal/ui/pane_filter.go` | Unchanged by this feature. |

Allowed dependency direction: the gesture state machine calls Pane helpers.
Pane code never refers to the drag session. Refresh paths never refer to the
drag session.

## Shared Components

None. The feature has one task (task0001); the Pane helper contracts it
introduces are defined in its task plan (`tasks/task0001.md`).

## Conventions

- **Cancellation form**: a cancelled drag session is always reset to the
  zero-value session, the same form used by the existing modal disarm and
  the existing directory-load-completion cancellation. Fields are never
  cleared individually.
- **Silent cancellation**: cancellation produces no status message, dialog,
  or visual change.
- **Tests**: standard Go tests in package `ui`, reusing the existing mouse
  test geometry, model builders and mark assertions; files are created in
  per-test temporary directories. No E2E test is added.

## Cross-task Design Decisions

### D1: Detect at the point of consumption, not at the refresh sites

- **Decision**: the stale-list check runs inside the drag session's motion
  and release handling, comparing a press-time snapshot of the drag pane's
  displayed names against the pane's current displayed names. No refresh
  path (auto refresh, exec / shell command / batch completion, dialog
  results and file operation completion, F5/Ctrl+R, delete, background
  command completion) is modified.
- **Rationale**: one check covers every route that replaces the displayed
  list (FR2), leaves unchanged refreshes transparent to the drag (FR3), and
  keeps refresh behavior byte-for-byte the same (NFR2).
- **Affected tasks**: task0001. Any later rework keeps drag-session logic
  out of the refresh paths.

### D2: The existing directory-load-completion cancellation is kept

- **Decision**: the explicit cancellation on a successful directory load
  completion for the drag pane (drag-dir-load-stale-marks) stays as it is;
  the new check is additive. Both reset the session to the zero value.
- **Affected tasks**: task0001 (must keep that behavior and its tests,
  NFR3).

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Per-motion cost of comparing the whole displayed name sequence in very large directories | Low | Low | The comparison is linear and stops at the first length or name difference; it runs only while a session is armed. |
| A different directory with an identical name sequence replaces the list during a drag and is not detected | Low | Low | Directory navigation during a drag is already cancelled by the load-completion path (D2); FR3 defines "unchanged" by names and order only. |
| Operations that clear marks without changing the displayed list (for example batch operation completion) let a continuing drag re-apply press-time baseline marks of files that still exist | Low | Low | Pre-existing behavior outside FR6 (A3); not changed by this feature. |

## Open Questions

- None.
