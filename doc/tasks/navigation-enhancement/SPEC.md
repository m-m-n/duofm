# Feature: Navigation Enhancement

## Overview

Enhance duofm's navigation capabilities with three new features:
1. Hidden file visibility toggle (Ctrl+H)
2. Home directory navigation (~)
3. Previous directory navigation (-)

These features align with common shell conventions, providing a familiar experience for terminal users.

## Objectives

- Improve navigation efficiency within the file manager
- Provide shell-like directory navigation (cd ~, cd -)
- Allow per-pane hidden file visibility control
- Maintain intuitive keyboard-driven workflow

## User Stories

- As a user, I want to quickly toggle hidden file visibility so that I can access dotfiles when needed
- As a user, I want to press ~ to jump to my home directory so that I can navigate there quickly
- As a user, I want to press - to return to the previous directory so that I can toggle between two directories like `cd -`

## Technical Requirements

### Key Bindings

| Key | Action | Scope |
|-----|--------|-------|
| `Ctrl+H` | Toggle hidden file visibility | Active pane only |
| `~` | Navigate to home directory | Active pane only |
| `-` | Navigate to previous directory | Active pane only |

### Feature 1: Hidden File Toggle (Ctrl+H)

**Behavior:**
- Toggle visibility of files/directories starting with `.`
- Default state: hidden (not shown)
- Each pane maintains independent visibility setting

**State Management:**
```go
// Add to Pane struct
type Pane struct {
    // ... existing fields
    showHidden bool  // Default: false
}
```

**Implementation:**
1. Add `showHidden` field to `Pane` struct
2. Modify `LoadDirectory()` to filter entries based on `showHidden`
3. Add `ToggleHidden()` method to `Pane`
4. Preserve cursor position when toggling (keep same file selected if possible)

**Visual Indicator:**
- Display `[H]` in pane header when hidden files are visible
- Position: Next to the path display

### Feature 2: Home Directory Navigation (~)

**Behavior:**
- Navigate active pane to user's home directory
- Update previous directory before navigation

**Implementation:**
```go
func (p *Pane) NavigateToHome() error {
    home, err := fs.HomeDirectory()
    if err != nil {
        return err
    }
    p.previousPath = p.currentPath
    return p.NavigateTo(home)
}
```

**Error Handling:**
- Show error dialog if home directory cannot be determined
- Show error dialog if home directory is not accessible

### Feature 3: Previous Directory Navigation (-)

**Behavior:**
- Navigate to the last visited directory (toggle behavior like `cd -`)
- Each pane maintains its own previous directory
- History depth: 1 (single previous directory only)

**State Management:**
```go
// Add to Pane struct
type Pane struct {
    // ... existing fields
    previousPath string  // Empty string when no history
}
```

**Implementation:**
```go
func (p *Pane) NavigateToPrevious() {
    if p.previousPath == "" {
        return  // No previous directory, do nothing
    }
    current := p.currentPath
    p.NavigateTo(p.previousPath)
    p.previousPath = current  // Enable toggle behavior
}
```

**History Update Points:**
Update `previousPath` before navigation in:
- `EnterDirectory()` - when entering a subdirectory
- `MoveToParent()` - when going to parent directory
- `NavigateToHome()` - when going to home directory
- `NavigateToPrevious()` - swap current and previous

## Implementation Approach

### Architecture

```
internal/ui/
├── keys.go          # Add new key constants
├── pane.go          # Add showHidden, previousPath, related methods
└── model.go         # Add key handlers in Update()
```

### Key Definitions (keys.go)

```go
const (
    // ... existing keys
    KeyToggleHidden = "ctrl+h"
    KeyHome         = "~"
    KeyPrevDir      = "-"
)
```

### Pane Modifications (pane.go)

```go
// New fields
type Pane struct {
    // ... existing
    showHidden   bool
    previousPath string
}

// New methods
func (p *Pane) ToggleHidden()
func (p *Pane) IsShowingHidden() bool
func (p *Pane) NavigateToHome() error
func (p *Pane) NavigateToPrevious()
func (p *Pane) recordPreviousPath()  // Internal helper
```

### Model Key Handlers (model.go)

```go
case KeyToggleHidden:
    m.getActivePane().ToggleHidden()
    return m, nil

case KeyHome:
    if err := m.getActivePane().NavigateToHome(); err != nil {
        m.dialog = NewErrorDialog(fmt.Sprintf("Cannot navigate to home: %v", err))
    }
    return m, nil

case KeyPrevDir:
    m.getActivePane().NavigateToPrevious()
    return m, nil
```

### Hidden File Filtering

Modify the directory loading to filter hidden files:

```go
func (p *Pane) LoadDirectory() {
    entries, err := fs.ReadDirectory(p.currentPath)
    if err != nil {
        // handle error
        return
    }

    if !p.showHidden {
        entries = filterHiddenFiles(entries)
    }

    p.entries = entries
    // ... rest of loading logic
}

func filterHiddenFiles(entries []fs.Entry) []fs.Entry {
    result := make([]fs.Entry, 0, len(entries))
    for _, e := range entries {
        if !strings.HasPrefix(e.Name, ".") {
            result = append(result, e)
        }
    }
    return result
}
```

### Cursor Preservation on Toggle

When toggling hidden files, preserve the cursor on the same file if possible:

```go
func (p *Pane) ToggleHidden() {
    // Remember current selection
    var selectedName string
    if p.cursor < len(p.entries) {
        selectedName = p.entries[p.cursor].Name
    }

    p.showHidden = !p.showHidden
    p.LoadDirectory()

    // Try to restore cursor position
    if selectedName != "" {
        for i, e := range p.entries {
            if e.Name == selectedName {
                p.cursor = i
                p.ensureCursorVisible()
                return
            }
        }
    }
    // If not found, reset to top
    p.cursor = 0
    p.scrollOffset = 0
}
```

## File Structure

```
internal/ui/
├── keys.go              # Add KeyToggleHidden, KeyHome, KeyPrevDir
├── pane.go              # Add showHidden, previousPath, new methods
├── pane_test.go         # Add tests for new functionality
├── model.go             # Add key handlers
└── model_test.go        # Add tests for key handling
```

## Test Scenarios

### Hidden File Toggle Tests

- [ ] `ToggleHidden()` changes `showHidden` from false to true
- [ ] `ToggleHidden()` changes `showHidden` from true to false
- [ ] Directory listing excludes hidden files when `showHidden` is false
- [ ] Directory listing includes hidden files when `showHidden` is true
- [ ] Cursor position preserved when toggling (if current file still visible)
- [ ] Cursor resets to 0 when toggled and current file becomes hidden
- [ ] Left and right panes have independent `showHidden` state

### Home Directory Navigation Tests

- [ ] `NavigateToHome()` changes current path to home directory
- [ ] `NavigateToHome()` updates `previousPath` before navigation
- [ ] `NavigateToHome()` returns error when home directory unavailable
- [ ] Error dialog shown when navigation fails

### Previous Directory Navigation Tests

- [ ] `NavigateToPrevious()` does nothing when `previousPath` is empty
- [ ] `NavigateToPrevious()` navigates to previous path when available
- [ ] Previous path becomes current path after toggle (A→B→A pattern)
- [ ] `previousPath` updated when entering directory
- [ ] `previousPath` updated when moving to parent
- [ ] `previousPath` updated when navigating to home
- [ ] Left and right panes have independent `previousPath`

### Integration Tests

- [ ] Ctrl+H key correctly triggers `ToggleHidden()`
- [ ] ~ key correctly triggers `NavigateToHome()`
- [ ] - key correctly triggers `NavigateToPrevious()`
- [ ] No key binding conflicts with existing keys
- [ ] Hidden indicator [H] appears/disappears correctly

## Success Criteria

- [ ] All three key bindings work correctly
- [ ] Hidden file toggle works independently per pane
- [ ] Home navigation works correctly
- [ ] Previous directory navigation provides toggle behavior (like cd -)
- [ ] Visual indicator shows hidden file visibility state
- [ ] All existing functionality remains unaffected
- [ ] All unit tests pass
- [ ] Performance within specified limits (50ms key response)

## Dependencies

- Existing `fs.HomeDirectory()` function
- No new external libraries required

## Open Questions

None - all requirements have been clarified.

## Out of Scope

- Stack-based directory history (future feature)
- Persistent configuration across sessions
- Multiple previous directory memory
