# Feature: Background Output Area UI Fix

## Overview

Fix two UI issues in the background shell command output area: (1) the separator line between the file list and output area should use gray (BorderFg) color instead of pink, and (2) the file list cursor should be constrained to the visible area above the output, preventing the cursor from scrolling into the output region.

## Objectives

- Improve visual separation between file list and background output area
- Prevent cursor from moving beyond the visible file list area during background execution

## User Stories

### US1: Visual Separator

As a user, I want a gray separator line between the file list and the background output area, so that the boundary is clearly visible and visually consistent with other borders in the UI.

**Acceptance Criteria:**
- [ ] Separator line uses BorderFg (gray) color by default
- [ ] Separator line shows the running command text (truncated if necessary)
- [ ] When output area is focused (TAB), separator changes to pink + bold
- [ ] When not focused, separator remains gray

### US2: Cursor Constraint

As a user, I want the file list cursor to stay within the visible file list area during background command execution, so that the cursor does not move into the output area.

**Acceptance Criteria:**
- [ ] File list visible area is reduced to top 2/3 when background output is active
- [ ] Cursor movement (j/k, Up/Down) is constrained to the visible file list height
- [ ] Page scrolling (Ctrl+D/U, PageDown/PageUp) respects the reduced height
- [ ] Scroll offset adjusts correctly for the reduced height
- [ ] After output area closes (auto-close or cancel), cursor behavior returns to full height

## Technical Requirements

### Functional Requirements

- **FR1:** Separator line color shall use `theme.BorderFg` (gray) when output area is not focused
- **FR2:** Separator line color shall use `highlightColor` (pink) with bold when output area is focused via TAB
- **FR3:** Separator line shall display the running command text, truncated with `...` if it exceeds available width
- **FR4:** `Pane.getVisibleLines()` shall return reduced height when background output area is active on that pane
- **FR5:** Cursor movement shall be constrained to the reduced visible area during background execution
- **FR6:** After background output area closes, visible lines shall return to normal full height

### Non-Functional Requirements

- **NFR1 - Compatibility:** All existing cursor navigation, scrolling, and display behavior shall remain unaffected when no background command is active
- **NFR2 - Compatibility:** All existing background shell command features shall remain intact

## Implementation Approach

### Changes Required

#### pane.go
- Add a field or method to indicate reduced visible height during background output
- Modify `getVisibleLines()` to account for background output area height

#### pane_render.go
- Change separator line styling from `highlightColor` to `theme.BorderFg` (default)
- Keep `highlightColor` + bold for focused state

### File Structure

```
internal/ui/
├── pane.go             # Add bg output height awareness to getVisibleLines()
├── pane_render.go      # Change separator color
├── model_view.go       # Pass bg active state to pane (if needed)
```

### Dependencies

**Internal Dependencies:**
- Existing `ViewWithBgOutput` in pane_render.go
- Existing `getVisibleLines()` and `adjustScroll()` in pane.go
- Existing `isBgActive()` in model.go

## Test Scenarios

### Unit Tests
- [ ] `getVisibleLines()` returns full height when no background output active
- [ ] `getVisibleLines()` returns reduced height when background output active
- [ ] Cursor stays within reduced visible area during background execution
- [ ] Scroll offset adjusts correctly with reduced height
- [ ] After bg output close, `getVisibleLines()` returns full height

### E2E Tests
**Existing E2E tests**: `test/e2e/` (Docker + tmux)
**Run command**: `make test-e2e`
- [ ] Existing E2E tests pass without regression
- [ ] Background output area shows gray separator line
- [ ] Cursor navigation works correctly during background execution

### Edge Cases
- [ ] Terminal resize during background execution recalculates heights
- [ ] Very small terminal height (minimum viable display)
- [ ] Background command on inactive pane has no effect on active pane cursor

## References

- Background shell command spec: `doc/tasks/background-shell-command/SPEC.md`
- Current implementation: `internal/ui/pane_render.go` (ViewWithBgOutput)
- Current cursor management: `internal/ui/pane.go` (getVisibleLines, adjustScroll)
