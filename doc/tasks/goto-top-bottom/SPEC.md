# Feature: Go-to-Top and Go-to-Bottom Navigation

## Overview

Add keybindings for jumping to the first and last entry in the file list, following `less` command conventions. `g` moves the cursor to the first entry (top) and `G` (Shift+G) moves the cursor to the last entry (bottom).

## Objectives

- Provide fast navigation to the beginning and end of the file list
- Follow `less` keybinding conventions for consistency with terminal workflows
- Integrate with the existing keybinding and action system

## User Stories

### US1: Jump to First Entry
As a user, I want to press `g` to jump to the first entry in the file list, so that I can quickly navigate to the top without repeatedly pressing `k`.

**Acceptance Criteria:**
- [ ] Pressing `g` moves the cursor to the first entry (index 0)
- [ ] The viewport scrolls to show the first entry
- [ ] Works in both left and right panes

### US2: Jump to Last Entry
As a user, I want to press `G` (Shift+G) to jump to the last entry in the file list, so that I can quickly navigate to the bottom without repeatedly pressing `j`.

**Acceptance Criteria:**
- [ ] Pressing `G` moves the cursor to the last entry
- [ ] The viewport scrolls to show the last entry
- [ ] Works in both left and right panes

## Technical Requirements

### Functional Requirements
- **FR1:** `ActionGotoTop` action moves the cursor to index 0 and adjusts scroll offset
- **FR2:** `ActionGotoBottom` action moves the cursor to the last entry index and adjusts scroll offset
- **FR3:** Default keybindings: `g` for goto_top, `Shift+G` for goto_bottom
- **FR4:** Keybindings are configurable via `config.toml` (`goto_top`, `goto_bottom` action names)
- **FR5:** Actions are no-ops when the file list is empty

### Non-Functional Requirements
- **NFR1 - Performance:** Cursor jump completes in < 1ms (same as existing cursor movement)
- **NFR2 - Consistency:** Follow existing patterns from `ActionPageDown`/`ActionPageUp` implementation

## Implementation Approach

### Files to Modify

```
internal/ui/
├── actions.go              # Add ActionGotoTop, ActionGotoBottom constants
├── pane.go                 # Add GotoTop(), GotoBottom() methods
├── pane_test.go            # Add unit tests for new methods
├── model_update_keyboard.go # Add action dispatch cases
└── keybinding_map_test.go  # Add keybinding mapping tests (if needed)

internal/config/
├── defaults.go             # Add goto_top, goto_bottom default keybindings
└── merger.go               # Auto-merge adds new keys to existing config
```

### Pane Methods

```go
func (p *Pane) GotoTop() {
    if len(p.entries) == 0 {
        return
    }
    p.cursor = 0
    p.adjustScroll()
}

func (p *Pane) GotoBottom() {
    if len(p.entries) == 0 {
        return
    }
    p.cursor = len(p.entries) - 1
    p.adjustScroll()
}
```

### Default Keybindings

```toml
[keybindings]
goto_top = ["G"]           # Normalizes to "g" (lowercase)
goto_bottom = ["Shift+G"]  # Normalizes to "G" (uppercase)
```

## Test Scenarios

### Unit Tests
- [ ] GotoTop moves cursor to 0 from any position
- [ ] GotoTop adjusts scroll offset to show first entry
- [ ] GotoTop is no-op on empty list
- [ ] GotoTop is no-op when already at top
- [ ] GotoBottom moves cursor to last entry from any position
- [ ] GotoBottom adjusts scroll offset to show last entry
- [ ] GotoBottom is no-op on empty list
- [ ] GotoBottom is no-op when already at bottom
- [ ] Action dispatch returns correct model state for both actions
- [ ] Default keybindings map correctly to actions

### Edge Cases
- [ ] Single-entry list: both GotoTop and GotoBottom set cursor to 0
- [ ] List shorter than viewport: scroll offset remains 0
- [ ] List longer than viewport: scroll adjusts correctly

## Success Criteria

- [ ] All functional requirements are implemented and tested
- [ ] All unit tests pass
- [ ] Keybindings are configurable via config.toml
- [ ] Config auto-merge adds new keybindings to existing user configs
