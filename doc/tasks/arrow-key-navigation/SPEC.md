# Specification: Arrow Key Navigation Support

## Overview

Add arrow key (↑↓←→) navigation support to duofm, providing identical functionality to the existing hjkl vim-style key bindings.

## Objectives

1. Enable arrow key navigation as an alternative to hjkl keys
2. Maintain full compatibility with existing key bindings
3. Update help documentation to reflect new key options

## User Stories

### US-1: Arrow Key Navigation in Main View
As a user who is not familiar with Vim key bindings,
I want to use arrow keys to navigate the file list,
So that I can use duofm without learning Vim-style navigation.

### US-2: Arrow Key Navigation in Dialogs
As a user,
I want arrow keys to work consistently in menus and dialogs,
So that I have a uniform experience throughout the application.

## Technical Requirements

### TR-1: Key Mapping

Add arrow key support with the following mappings:

| Arrow Key | Equivalent Vim Key | Action |
|-----------|-------------------|--------|
| Up (↑) | k | Move cursor up |
| Down (↓) | j | Move cursor down |
| Left (←) | h | Switch pane left / Move to parent directory |
| Right (→) | l | Switch pane right / Move to parent directory |

### TR-2: Affected Components

| File | Change Required |
|------|-----------------|
| `internal/ui/keys.go` | Add arrow key constants |
| `internal/ui/model.go` | Handle arrow keys in Update() |
| `internal/ui/context_menu_dialog.go` | Already supports ↑↓ (verify ←→ for pagination) |
| `internal/ui/help_dialog.go` | Update help text |

### TR-3: Bubble Tea Key Strings

Arrow keys are represented as strings in Bubble Tea:
- `"up"` - Up arrow
- `"down"` - Down arrow
- `"left"` - Left arrow
- `"right"` - Right arrow

## Implementation Approach

### Architecture

No architectural changes required. This is a straightforward addition to existing key handling.

### Code Changes

#### 1. keys.go - Add Arrow Key Constants

```go
const (
    // Existing keys...
    KeyMoveDown    = "j"
    KeyMoveUp      = "k"
    KeyMoveLeft    = "h"
    KeyMoveRight   = "l"

    // Arrow key alternatives
    KeyArrowDown   = "down"
    KeyArrowUp     = "up"
    KeyArrowLeft   = "left"
    KeyArrowRight  = "right"
    // ...
)
```

#### 2. model.go - Update Key Handling

```go
case KeyMoveDown, KeyArrowDown:
    m.getActivePane().MoveCursorDown()

case KeyMoveUp, KeyArrowUp:
    m.getActivePane().MoveCursorUp()

case KeyMoveLeft, KeyArrowLeft:
    // ... existing logic

case KeyMoveRight, KeyArrowRight:
    // ... existing logic
```

#### 3. help_dialog.go - Update Help Text

```go
content := []string{
    "Navigation",
    "  j/k/↑/↓  : move cursor down/up",
    "  h/l/←/→  : move to left/right pane or parent directory",
    // ...
}
```

#### 4. context_menu_dialog.go - Verify/Add Arrow Key Support

The context menu already handles `"up"` and `"down"`. Verify and add `"left"` and `"right"` for pagination if needed.

## Dependencies

- No external dependencies
- Relies on Bubble Tea's built-in key handling

## Test Scenarios

### Unit Tests

#### Test Arrow Key Navigation (model_test.go)

```go
func TestArrowKeyNavigation(t *testing.T) {
    tests := []struct {
        name     string
        key      string
        expected func(m Model) bool
    }{
        {"down arrow moves cursor down", "down", func(m Model) bool { return m.getActivePane().cursor == 1 }},
        {"up arrow moves cursor up", "up", func(m Model) bool { return m.getActivePane().cursor == 0 }},
        // ...
    }
    // ...
}
```

### Integration Tests

1. Test arrow key navigation in main view
2. Test arrow key navigation in context menu
3. Test arrow key pane switching
4. Test arrow key parent directory navigation

## Security Considerations

None - this is a UI enhancement with no security implications.

## Error Handling

No additional error handling required. Arrow keys use the same code paths as existing hjkl keys.

## Performance Optimization

No performance impact. Key handling is synchronous and immediate.

## Success Criteria

1. All arrow keys (↑↓←→) work identically to their hjkl counterparts
2. Existing hjkl key bindings remain functional
3. Help dialog displays updated key binding information
4. All existing tests pass
5. New tests for arrow key navigation pass

## Open Questions

None - all requirements have been clarified.
