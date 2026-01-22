# Feature: Disable Open/Open with Menu Without Desktop Environment

## Overview

When duofm runs on machines without a desktop environment (e.g., SSH sessions, headless servers), the "Open" and "Open with ..." context menu items should be grayed out and non-selectable, as `xdg-open` and `xdg-mime` will not function in such environments.

## Objectives

- Detect desktop environment presence using environment variables
- Gray out "Open" and "Open with ..." menu items when no desktop environment is detected
- Maintain normal functionality when desktop environment is available

## User Stories

### US1: Menu Disabled in Headless Environment
As a user connecting via SSH, I want to see "Open" and "Open with ..." grayed out, so that I know these features are unavailable in my current environment.

**Acceptance Criteria:**
- [ ] Menu items are displayed in gray/dimmed color
- [ ] Menu items cannot be selected via keyboard navigation
- [ ] Other menu items remain functional

### US2: Menu Enabled in Desktop Environment
As a desktop user, I want "Open" and "Open with ..." to work normally, so that I can open files with applications.

**Acceptance Criteria:**
- [ ] Menu items are displayed normally
- [ ] Menu items are selectable and functional

## Technical Requirements

### Functional Requirements
- **FR1:** Detect desktop environment by checking `DISPLAY` and `WAYLAND_DISPLAY` environment variables
- **FR2:** Gray out "Open" and "Open with ..." menu items when both environment variables are unset
- **FR3:** Skip disabled menu items during keyboard navigation

### Non-Functional Requirements
- **NFR1 - Performance:** Environment detection should happen once at startup and be cached
- **NFR2 - Maintainability:** Detection logic should be in a separate, reusable function

## Implementation Approach

### Architecture

```
┌─────────────────────────────────────┐
│     context_menu_dialog.go          │
│  (uses HasDesktopEnvironment())     │
├─────────────────────────────────────┤
│         internal/ui/env.go          │
│    HasDesktopEnvironment() bool     │
└─────────────────────────────────────┘
```

### Desktop Environment Detection

Create a new function in `internal/ui/` to detect desktop environment:

```go
// env.go

var hasDesktop = detectDesktopEnvironment()

func detectDesktopEnvironment() bool {
    if os.Getenv("DISPLAY") != "" {
        return true
    }
    if os.Getenv("WAYLAND_DISPLAY") != "" {
        return true
    }
    return false
}

// HasDesktopEnvironment returns true if a desktop environment is available
func HasDesktopEnvironment() bool {
    return hasDesktop
}
```

### Menu Item Modification

Modify `context_menu_dialog.go` to use desktop environment detection:

1. Use existing `Enabled` field in MenuItem struct
2. Gray out disabled items in rendering (already implemented)
3. Skip disabled items in navigation (already implemented)

```go
// Use existing MenuItem struct with Enabled field
// When creating menu items
items := []MenuItem{
    {Label: "Open", Action: openAction, Enabled: HasDesktopEnvironment()},
    {Label: "Open with ...", Action: openWithAction, Enabled: HasDesktopEnvironment()},
    // ... other items
}
```

### Rendering Disabled Items

```go
// In render function (existing behavior)
style := normalStyle
if !item.Enabled {
    style = disabledStyle  // Gray/dimmed color
}
```

### Navigation Skip Logic

```go
// In key handling, skip disabled items (existing behavior with guard)
func (m *contextMenuModel) moveDown() {
    attempts := 0
    for {
        m.cursor++
        if m.cursor >= len(m.items) {
            m.cursor = 0
        }
        attempts++
        if m.items[m.cursor].Enabled || attempts >= len(m.items) {
            break  // Stop if enabled item found or all items checked
        }
    }
}
```

### File Structure

```
internal/ui/
├── env.go                  # Desktop environment detection (new)
├── env_test.go             # Tests for env.go (new)
├── context_menu_dialog.go  # Modified to support disabled items
└── ...
```

## Test Scenarios

### Unit Tests
- [ ] `HasDesktopEnvironment()` returns `true` when `DISPLAY` is set
- [ ] `HasDesktopEnvironment()` returns `true` when `WAYLAND_DISPLAY` is set
- [ ] `HasDesktopEnvironment()` returns `false` when both are unset
- [ ] `HasDesktopEnvironment()` returns `false` when variables are empty strings

### Integration Tests
- [ ] Context menu renders "Open" as disabled when no desktop environment
- [ ] Context menu renders "Open" as enabled when desktop environment exists
- [ ] Keyboard navigation skips disabled items

### Edge Cases
- [ ] Both `DISPLAY` and `WAYLAND_DISPLAY` are set (should return true)
- [ ] Environment variable set to empty string (should return false)

## Success Criteria

- [ ] Desktop environment detection works correctly
- [ ] Menu items are grayed out in headless environment
- [ ] Menu items cannot be selected when disabled
- [ ] Normal operation in desktop environment
- [ ] All unit tests pass
- [ ] Code follows project conventions

## Open Questions

- None

## References

- `internal/ui/context_menu_dialog.go` - Current context menu implementation
- `internal/ui/open_with_dialog.go` - Open with dialog using xdg-mime
- `internal/ui/exec.go` - File opening with xdg-open
