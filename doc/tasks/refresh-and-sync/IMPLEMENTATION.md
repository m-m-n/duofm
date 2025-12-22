# Implementation Plan: Refresh and Pane Synchronization

## Overview

Implement two new features for duofm:
1. **Refresh Feature (F5/Ctrl+R)**: Reload both panes to reflect filesystem changes
2. **Pane Sync Feature (=)**: Synchronize the opposite pane to the active pane's directory

These features enable users to immediately see external file changes and to display the same directory in both panes with different settings.

## Objectives

- Enable instant reflection of filesystem changes with F5/Ctrl+R
- Improve file operation efficiency by viewing the same directory in both panes
- Allow comparison of the same directory with different display settings
- Preserve cursor position intelligently after refresh
- Handle edge cases (deleted directories, permission changes)

## Prerequisites

- Go 1.21 or later
- Understanding of Bubble Tea update/view cycle
- Existing pane and model infrastructure
- Familiarity with existing key binding system

## Architecture Overview

```
┌────────────────────────────────────────────────────────────┐
│                         Model                               │
│  ┌──────────────┐  ┌──────────────┐                        │
│  │  leftPane    │  │  rightPane   │                        │
│  │              │  │              │                        │
│  │  Refresh()   │  │  Refresh()   │                        │
│  │  SyncTo()    │  │  SyncTo()    │                        │
│  └──────────────┘  └──────────────┘                        │
│                                                             │
│  RefreshBothPanes()  - Refresh both panes simultaneously   │
│  SyncOppositePane()  - Sync opposite pane to active dir    │
└────────────────────────────────────────────────────────────┘

Key Bindings:
  F5, Ctrl+R  → RefreshBothPanes()
  =           → SyncOppositePane()
```

## Implementation Phases

### Phase 1: Helper Functions in fs Package

**Goal**: Add directory existence check function to fs package

**Files to Create/Modify**:
- `internal/fs/reader.go` - Add `DirectoryExists()` function

**Implementation Steps**:

1. Add `DirectoryExists()` function to check if a directory exists and is accessible:
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

**Dependencies**:
- None

**Testing**:
- Unit test for existing directory (returns true)
- Unit test for non-existent directory (returns false)
- Unit test for file path (returns false)
- Unit test for permission denied directory (returns false)

**Estimated Effort**: Small

---

### Phase 2: Add Key Definitions

**Goal**: Define new key bindings for refresh and sync operations

**Files to Create/Modify**:
- `internal/ui/keys.go` - Add new key constants

**Implementation Steps**:

1. Add key constants to `keys.go`:
   ```go
   const (
       // ... existing keys ...

       // Refresh and sync
       KeyRefresh    = "f5"      // Refresh view
       KeyRefreshAlt = "ctrl+r"  // Refresh view (alternative)
       KeySyncPane   = "="       // Pane synchronization
   )
   ```

**Dependencies**:
- None

**Testing**:
- Verify constants are defined (compile check)

**Estimated Effort**: Small

---

### Phase 3: Implement Pane.Refresh() Method

**Goal**: Add refresh functionality to Pane with cursor position preservation

**Files to Create/Modify**:
- `internal/ui/pane.go` - Add `Refresh()` method

**Implementation Steps**:

1. Add `Refresh()` method to Pane:
   ```go
   // Refresh reloads the current directory, preserving cursor position
   func (p *Pane) Refresh() error {
       // Save currently selected filename
       var selectedName string
       if p.cursor >= 0 && p.cursor < len(p.entries) {
           selectedName = p.entries[p.cursor].Name
       }
       savedCursor := p.cursor

       // Reload directory with existence check
       currentPath := p.path
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

       if currentPath != p.path {
           // Directory was changed, update previousPath for navigation history
           p.previousPath = p.path
           p.path = currentPath
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
                   p.adjustScroll()
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
       p.adjustScroll()

       return nil
   }
   ```

**Dependencies**:
- Phase 1 (DirectoryExists function)

**Testing**:
- Test cursor position preservation when file still exists
- Test cursor index preservation when file deleted
- Test cursor adjustment when index out of range
- Test directory deletion handling (navigate to parent)
- Test cascade to root when parents deleted
- Test fallback to home directory
- Test filter state preservation after refresh

**Estimated Effort**: Medium

---

### Phase 4: Implement Pane.SyncTo() Method

**Goal**: Add synchronization functionality to Pane

**Files to Create/Modify**:
- `internal/ui/pane.go` - Add `SyncTo()` method

**Implementation Steps**:

1. Add `SyncTo()` method to Pane:
   ```go
   // SyncTo synchronizes this pane to the specified directory
   // Preserves display settings but resets cursor to top
   func (p *Pane) SyncTo(path string) error {
       // Do nothing if already in the same directory
       if p.path == path {
           return nil
       }

       // Update previousPath for navigation history
       p.previousPath = p.path

       // Change directory
       p.path = path
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

**Dependencies**:
- None (uses existing LoadDirectory)

**Testing**:
- Test basic sync operation
- Test same directory skip (no-op)
- Test previousPath update
- Test cursor reset to 0
- Test scroll offset reset
- Test showHidden preservation
- Test displayMode preservation
- Test filter state cleared (by LoadDirectory)

**Estimated Effort**: Small

---

### Phase 5: Implement Model Methods

**Goal**: Add refresh and sync operations to Model

**Files to Create/Modify**:
- `internal/ui/model.go` - Add `RefreshBothPanes()` and `SyncOppositePane()` methods

**Implementation Steps**:

1. Add `RefreshBothPanes()` method:
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

2. Add `SyncOppositePane()` method:
   ```go
   // SyncOppositePane synchronizes the opposite pane to the active pane's directory
   func (m *Model) SyncOppositePane() {
       activePane := m.getActivePane()
       oppositePane := m.getInactivePane() // Use existing helper method

       if err := oppositePane.SyncTo(activePane.path); err != nil {
           m.dialog = NewErrorDialog(fmt.Sprintf("Failed to sync pane: %v", err))
       }
   }
   ```

   Note: Use existing `getInactivePane()` method instead of creating `getOppositePane()`

**Dependencies**:
- Phase 3 (Pane.Refresh)
- Phase 4 (Pane.SyncTo)

**Testing**:
- Test both panes refresh successfully
- Test disk space update after refresh
- Test error dialog on refresh failure
- Test sync to opposite pane
- Test error dialog on sync failure
- Test sync preserves opposite pane settings

**Estimated Effort**: Small

---

### Phase 6: Integrate Key Handlers

**Goal**: Add key bindings to Model.Update()

**Files to Create/Modify**:
- `internal/ui/model.go` - Add key handlers in Update method

**Implementation Steps**:

1. Add key handlers in the `Update()` method's switch statement:
   ```go
   case tea.KeyMsg:
       // ... dialog and search handling ...

       switch msg.String() {

       case KeyRefresh, KeyRefreshAlt:
           return m, m.RefreshBothPanes()

       case KeySyncPane:
           m.SyncOppositePane()
           return m, nil

       // ... existing cases ...
       }
   ```

   Insert after Ctrl+C handling but before other key handlers

**Dependencies**:
- Phase 2 (Key constants)
- Phase 5 (Model methods)

**Testing**:
- Test F5 key triggers refresh
- Test Ctrl+R key triggers refresh
- Test = key triggers sync
- Test keys work in left pane
- Test keys work in right pane
- Test keys ignored during dialog display
- Test keys work during search mode (verify if needed or should cancel search)

**Estimated Effort**: Small

---

### Phase 7: Unit Tests

**Goal**: Add comprehensive unit tests for new functionality

**Files to Create/Modify**:
- `internal/fs/reader_test.go` - Add DirectoryExists tests
- `internal/ui/pane_test.go` - Add Refresh and SyncTo tests
- `internal/ui/model_test.go` - Add integration tests

**Implementation Steps**:

1. Add tests to `internal/fs/reader_test.go`:
   ```go
   func TestDirectoryExists(t *testing.T) {
       // Test cases:
       // - Existing directory
       // - Non-existent directory
       // - File path (not directory)
       // - Permission denied (if testable)
   }
   ```

2. Add tests to `internal/ui/pane_test.go`:
   ```go
   func TestPaneRefresh(t *testing.T) {
       // Test cursor position preservation
       // Test directory deletion handling
       // Test filter preservation
   }

   func TestPaneSyncTo(t *testing.T) {
       // Test basic sync
       // Test same directory skip
       // Test cursor reset
       // Test settings preservation
   }
   ```

3. Add tests to `internal/ui/model_test.go`:
   ```go
   func TestRefreshBothPanes(t *testing.T) {
       // Test both panes refresh
       // Test disk space update
   }

   func TestSyncOppositePane(t *testing.T) {
       // Test sync from left to right
       // Test sync from right to left
   }

   func TestRefreshKeys(t *testing.T) {
       // Test F5 key
       // Test Ctrl+R key
   }

   func TestSyncKey(t *testing.T) {
       // Test = key
   }
   ```

**Dependencies**:
- Phase 1-6 (All implementation)

**Testing**:
- Run `go test ./internal/fs/...`
- Run `go test ./internal/ui/...`
- Achieve >80% code coverage for new code

**Estimated Effort**: Medium

---

## File Structure

```
internal/
├── fs/
│   ├── reader.go          # Add DirectoryExists()
│   └── reader_test.go     # Add DirectoryExists tests
└── ui/
    ├── keys.go            # Add KeyRefresh, KeyRefreshAlt, KeySyncPane
    ├── pane.go            # Add Refresh(), SyncTo()
    ├── pane_test.go       # Add Refresh and SyncTo tests
    ├── model.go           # Add RefreshBothPanes(), SyncOppositePane(), key handlers
    └── model_test.go      # Add integration tests
```

## Testing Strategy

### Unit Tests

**fs/reader_test.go**:
- `TestDirectoryExists` - Directory existence check with various inputs

**ui/pane_test.go**:
- `TestPaneRefresh` - Cursor position preservation, directory deletion handling
- `TestPaneSyncTo` - Basic sync, settings preservation, cursor reset

**ui/model_test.go**:
- `TestRefreshBothPanes` - Both panes refresh, disk space update
- `TestSyncOppositePane` - Sync operation, error handling
- `TestRefreshKeys` - F5 and Ctrl+R key handling
- `TestSyncKey` - = key handling

### Integration Tests

- Test refresh with filter active
- Test refresh during search mode
- Test sync with different showHidden settings
- Test sync with different displayMode settings
- Test refresh after external file changes
- Test error recovery paths

### Manual Testing Checklist

- [ ] F5 refreshes both panes
- [ ] Ctrl+R refreshes both panes (same as F5)
- [ ] After refresh, same file is selected (if exists)
- [ ] After refresh, same index is maintained (if file deleted)
- [ ] Deleted directory navigates to parent
- [ ] Cascade navigation to existing parent works
- [ ] Fallback to home directory works
- [ ] = key syncs opposite pane to active pane's directory
- [ ] Sync resets cursor to top
- [ ] Sync preserves showHidden setting
- [ ] Sync preserves displayMode setting
- [ ] Sync same directory is no-op
- [ ] - key returns to original directory after sync
- [ ] Filter state preserved after refresh
- [ ] Disk space updated after refresh
- [ ] Error dialogs shown on failures
- [ ] Keys ignored during dialog display
- [ ] Performance acceptable (<100ms for typical directories)

## Dependencies

### External Libraries

- No new external libraries required
- Uses existing Go standard library (`os`, `path/filepath`)
- Uses existing Bubble Tea framework

### Internal Dependencies

```
Phase 1 (DirectoryExists) ──→ Phase 3 (Pane.Refresh)
                                      │
Phase 2 (Key constants) ────→ Phase 6 (Key handlers)
                                      ▲
Phase 4 (Pane.SyncTo) ──→ Phase 5 (Model methods) ──┘
                                      │
                                      ▼
                              Phase 7 (Testing)
```

Phases 1-2 can be implemented in parallel
Phases 3-4 can be implemented in parallel
Phase 5 requires 3 and 4
Phase 6 requires 2 and 5
Phase 7 requires all previous phases

## Risk Assessment

### Technical Risks

- **Directory Deletion During Refresh**: User's current directory might be deleted externally
  - Mitigation: Implemented robust parent navigation with fallbacks

- **Permission Changes**: Directory permissions might change between operations
  - Mitigation: Error handling with error dialogs, maintain previous state

- **Performance**: Refresh on very large directories might be slow
  - Mitigation: LoadDirectory is already optimized, no additional overhead

### Implementation Risks

- **Cursor Position Logic**: Complex logic for cursor restoration might have edge cases
  - Mitigation: Comprehensive unit tests covering all scenarios

- **Filter State**: Refresh should preserve filter state correctly
  - Mitigation: LoadDirectory already handles allEntries correctly

## Performance Considerations

- Refresh operation reuses existing `LoadDirectory()` which is already optimized
- No additional overhead beyond normal directory loading
- Cursor position restoration is O(n) where n = number of entries (acceptable)
- Directory existence checks are fast (single stat call per directory in chain)
- Expected performance: <100ms for directories with <1000 entries

## Security Considerations

- No security implications (read-only operations)
- Error handling prevents crashes on permission denied
- No user input validation required (keys are predefined)
- Directory existence checks prevent invalid state

## Open Questions

None - all requirements have been clarified in the specification.

## Future Enhancements

The following features are explicitly out of scope but could be considered later:

- Auto-refresh with filesystem watcher (fsnotify)
- Configurable refresh interval
- Selective refresh (only active pane)
- Refresh history persistence
- Multi-directory synchronization
- Visual feedback during refresh (loading indicator)

## Success Criteria

- [ ] F5 and Ctrl+R refresh both panes correctly
- [ ] Cursor position is intelligently preserved after refresh
- [ ] Deleted directories navigate to appropriate parent
- [ ] = key syncs opposite pane correctly
- [ ] Display settings (showHidden, displayMode) preserved during sync
- [ ] Same directory case handled correctly (no-op)
- [ ] Filter state preserved after refresh
- [ ] Disk space updated after refresh
- [ ] Error dialogs shown appropriately
- [ ] All unit tests pass
- [ ] Manual testing checklist completed
- [ ] Performance within acceptable range (<100ms)
- [ ] No regressions in existing functionality

## References

- [SPEC.md](./SPEC.md) - Feature specification
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- Existing implementation patterns:
  - [search-filter/IMPLEMENTATION.md](../search-filter/IMPLEMENTATION.md) - Reference for phased approach
  - [ctrl-c-cancel/IMPLEMENTATION.md](../ctrl-c-cancel/IMPLEMENTATION.md) - Reference for key handling
