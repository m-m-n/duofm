# Feature: Refresh and Pane Synchronization

## Overview

Add two new features to duofm:

1. **Refresh Feature (F5/Ctrl+R)**: Reload both panes to reflect filesystem changes
2. **Pane Sync Feature (=)**: Synchronize the opposite pane to the active pane's directory

These features enable users to immediately see external file changes and to display the same directory in both panes with different settings (e.g., hidden file visibility).

## Objectives

- Enable instant reflection of filesystem changes
- Improve file operation efficiency by viewing the same directory in both panes
- Allow comparison of the same directory with different display settings (hidden file visibility)

## User Stories

### US-1: Refresh View
**As a** user who works with external programs,
**I want to** press F5 or Ctrl+R to refresh the display to the latest state,
**So that** I can immediately see file changes made by other terminals or programs.

### US-2: Pane Synchronization
**As a** user,
**I want to** press = to change the opposite pane to the same directory as the current pane,
**So that** I can select files within the same directory or compare different display settings.

## Technical Requirements

### TR-1: Refresh Feature (F5/Ctrl+R)

#### Key Bindings

| Key | Action |
|-----|--------|
| `F5` | Reload both panes |
| `Ctrl+R` | Reload both panes (same as F5) |

#### Behavior Specification

1. **Refresh Scope**
   - Refresh both left and right panes simultaneously
   - Refresh both panes regardless of which is active

2. **Cursor Position Preservation**
   - Remember the currently selected filename in each pane
   - After reload, select the same file if it still exists
   - If the file doesn't exist, maintain the previous cursor position (index)
   - If the index is out of range, select the last entry

3. **Refresh Content**
   - Reload directory contents (reflect file/directory additions, deletions, and modifications)
   - Recalculate disk free space

4. **Error Handling**
   - If the displayed directory has been deleted:
     - Navigate up to an existing parent directory
     - If no parent exists up to root, navigate to home directory
     - If home directory doesn't exist, navigate to root (/)
   - If directory read permissions are lost:
     - Display error dialog
     - Maintain previous pane state

5. **Visual Feedback**
   - No special indicator or message required
   - Processing typically completes instantly

### TR-2: Pane Sync Feature (=)

#### Key Bindings

| Key | Action |
|-----|--------|
| `=` | Change the opposite pane to the active pane's directory |

#### Behavior Specification

1. **Basic Operation**
   - Get the current directory path of the active pane
   - Change the opposite pane to the same directory

2. **Cursor Position**
   - Move the opposite pane's cursor to the top (index 0)
   - Reset scroll offset to 0

3. **Same Directory Case**
   - If both panes already display the same directory, do nothing
   - Skip processing silently (no error message)

4. **Display Settings Preservation**
   - Preserve the opposite pane's hidden file visibility setting (`showHidden`)
   - Preserve the opposite pane's display mode (`displayMode`)
   - This allows viewing the same directory with different settings

5. **History Update**
   - Update the opposite pane's "previous directory" (`previousPath`)
   - This allows returning to the original directory with the `-` key

## Implementation Approach

### Architecture

```
internal/ui/
├── keys.go          # Add new key definitions
├── pane.go          # Add Refresh(), SyncTo() methods
├── pane_test.go     # Add tests
├── model.go         # Add key handlers
└── model_test.go    # Add tests
```

### Data Structures

Use existing structures. No new fields required.

### Key Definitions (keys.go)

```go
const (
    // ... existing key definitions
    KeyRefresh    = "f5"      // Refresh view
    KeyRefreshAlt = "ctrl+r"  // Refresh view (alternative)
    KeySyncPane   = "="       // Pane synchronization
)
```

### Pane Methods (pane.go)

#### Refresh() - Reload the pane

```go
// Refresh reloads the current directory, preserving cursor position
func (p *Pane) Refresh() error {
    // Save currently selected filename
    var selectedName string
    if p.cursor < len(p.entries) {
        selectedName = p.entries[p.cursor].Name
    }
    savedCursor := p.cursor

    // Reload directory with existence check
    currentPath := p.currentPath
    for {
        if fs.DirectoryExists(currentPath) {
            break
        }
        // Navigate up to parent directory
        parent := filepath.Dir(currentPath)
        if parent == currentPath {
            // Reached root but it doesn't exist
            home, err := fs.HomeDirectory()
            if err == nil && fs.DirectoryExists(home) {
                currentPath = home
                break
            }
            currentPath = "/"
            break
        }
        currentPath = parent
    }

    if currentPath != p.currentPath {
        // Directory was changed
        p.currentPath = currentPath
    }

    err := p.LoadDirectory()
    if err != nil {
        return err
    }

    // Restore cursor position
    if selectedName != "" {
        // Search for the same filename
        for i, e := range p.entries {
            if e.Name == selectedName {
                p.cursor = i
                p.ensureCursorVisible()
                return nil
            }
        }
    }

    // If file not found, use previous index
    if savedCursor < len(p.entries) {
        p.cursor = savedCursor
    } else if len(p.entries) > 0 {
        p.cursor = len(p.entries) - 1
    } else {
        p.cursor = 0
    }
    p.ensureCursorVisible()

    return nil
}
```

#### SyncTo() - Synchronize to specified directory

```go
// SyncTo synchronizes this pane to the specified directory
// Preserves display settings but resets cursor to top
func (p *Pane) SyncTo(path string) error {
    // Do nothing if already in the same directory
    if p.currentPath == path {
        return nil
    }

    // Update previousPath
    p.previousPath = p.currentPath

    // Change directory
    p.currentPath = path
    err := p.LoadDirectory()
    if err != nil {
        return err
    }

    // Reset cursor and scroll to top
    p.cursor = 0
    p.scrollOffset = 0

    return nil
}
```

### Model Methods (model.go)

#### RefreshBothPanes() - Refresh both panes

```go
// RefreshBothPanes refreshes both panes
func (m *Model) RefreshBothPanes() tea.Cmd {
    var cmds []tea.Cmd

    // Refresh left pane
    if err := m.leftPane.Refresh(); err != nil {
        m.dialog = NewErrorDialog(fmt.Sprintf("Failed to refresh left pane: %v", err))
    }

    // Refresh right pane
    if err := m.rightPane.Refresh(); err != nil {
        m.dialog = NewErrorDialog(fmt.Sprintf("Failed to refresh right pane: %v", err))
    }

    // Update disk space
    m.updateDiskSpace()

    return tea.Batch(cmds...)
}
```

#### SyncOppositePane() - Synchronize the opposite pane

```go
// SyncOppositePane synchronizes the opposite pane to the active pane's directory
func (m *Model) SyncOppositePane() {
    activePane := m.getActivePane()
    oppositePane := m.getOppositePane()

    if err := oppositePane.SyncTo(activePane.currentPath); err != nil {
        m.dialog = NewErrorDialog(fmt.Sprintf("Failed to sync pane: %v", err))
    }
}
```

#### Integration into Update method

```go
// In Update() method
case tea.KeyMsg:
    switch msg.String() {
    case KeyRefresh, KeyRefreshAlt:
        return m, m.RefreshBothPanes()

    case KeySyncPane:
        m.SyncOppositePane()
        return m, nil

    // ... existing cases
    }
```

## Helper Functions

### DirectoryExists() - Check directory existence

If this function doesn't exist in the `fs` package, it needs to be implemented:

```go
// DirectoryExists checks if a directory exists and is accessible
func DirectoryExists(path string) bool {
    info, err := os.Stat(path)
    if err != nil {
        return false
    }
    return info.IsDir()
}
```

## Test Scenarios

### Refresh Feature Tests

#### Basic Operation
- [ ] F5 key reloads both panes
- [ ] Ctrl+R key reloads both panes
- [ ] After reload, the same file is selected in each pane (if the file exists)
- [ ] If a file is deleted, the previous index is maintained
- [ ] Disk free space is recalculated

#### Cursor Position Preservation
- [ ] If the selected file exists, the same file is selected
- [ ] If the selected file is deleted, the same index is maintained
- [ ] If the index is out of range, the last entry is selected
- [ ] If entries are empty, cursor becomes 0

#### Error Handling
- [ ] If the directory is deleted, navigate to parent directory
- [ ] If no parent exists up to root, navigate to home directory
- [ ] If home directory doesn't exist, navigate to root (/)
- [ ] If read permission is lost, display error dialog

#### Filesystem Change Reflection
- [ ] Externally added files are displayed
- [ ] Externally deleted files are hidden
- [ ] Externally renamed files are correctly reflected
- [ ] File size changes are reflected
- [ ] File modification time changes are reflected

### Pane Sync Feature Tests

#### Basic Operation
- [ ] = key changes the opposite pane to the active pane's directory
- [ ] When left pane is active, right pane is synchronized
- [ ] When right pane is active, left pane is synchronized
- [ ] After sync, opposite pane's cursor becomes top (0)
- [ ] After sync, opposite pane's scroll offset becomes 0

#### Display Settings Preservation
- [ ] Opposite pane's hidden file visibility setting (showHidden) is preserved
- [ ] Opposite pane's display mode (displayMode) is preserved
- [ ] Settings don't change even when syncing with left pane showing hidden files and right pane hiding them

#### History Update
- [ ] After sync, opposite pane's previousPath is correctly updated
- [ ] After sync, can return to original directory with - key

#### Edge Cases
- [ ] If both panes already display the same directory, do nothing (no error)
- [ ] If sync target directory is unreadable, display error dialog
- [ ] If sync target directory doesn't exist, display error dialog

### Integration Tests

- [ ] Refresh and pane sync work correctly when executed consecutively
- [ ] F5 works correctly during search mode
- [ ] Keys are ignored when dialog is displayed
- [ ] No conflicts with existing key bindings

## Success Criteria

- [ ] Both panes refresh correctly with F5 and Ctrl+R
- [ ] Cursor position is appropriately preserved
- [ ] Navigate to appropriate parent directory when directory is deleted
- [ ] Opposite pane syncs correctly with = key
- [ ] Display settings are preserved during sync
- [ ] Same directory case is ignored
- [ ] No impact on existing functionality
- [ ] All unit tests pass
- [ ] Performance within acceptable range (refresh processing < 100ms)

## Dependencies

- Existing `fs.DirectoryExists()` function (implementation required if it doesn't exist)
- Existing `fs.HomeDirectory()` function
- Bubble Tea framework
- No new external libraries required

## Open Questions

None - all requirements have been clarified.

## Out of Scope

The following features are not included in this implementation:

- Auto-refresh (filesystem watcher)
- Configurable refresh interval
- Selective refresh (only one pane)
- Refresh history persistence
- Multi-directory synchronization
