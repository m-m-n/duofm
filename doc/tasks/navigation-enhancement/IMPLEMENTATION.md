# Implementation Plan: Navigation Enhancement

## Overview

Implement three navigation features for duofm:
1. Hidden file visibility toggle (Ctrl+H)
2. Home directory navigation (~)
3. Previous directory navigation (-)

All features operate on a per-pane basis and align with common shell conventions.

## Objectives

- Add `showHidden` and `previousPath` fields to `Pane` struct
- Implement filtering logic for hidden files
- Add key handlers for new navigation commands
- Display visual indicator for hidden file visibility state
- Maintain cursor position when toggling hidden files

## Prerequisites

- Existing `fs.HomeDirectory()` function (already available)
- Understanding of current `Pane` and `Model` structure
- No external library additions required

## Architecture Overview

```
internal/ui/
├── keys.go    # Add KeyToggleHidden, KeyHome, KeyPrevDir
├── pane.go    # Add showHidden, previousPath, navigation methods
└── model.go   # Add key handlers in Update()
```

Changes are localized to the UI layer. No modifications to `internal/fs/` required.

## Implementation Phases

### Phase 1: Add Pane State Fields

**Goal**: Add necessary state fields to `Pane` struct

**Files to Modify**:
- `internal/ui/pane.go` - Add struct fields

**Implementation Steps**:

1. Add fields to `Pane` struct:
   ```go
   type Pane struct {
       // ... existing fields
       showHidden   bool   // Default: false (hidden files not shown)
       previousPath string // Empty string when no history
   }
   ```

2. No initialization changes needed (`showHidden` defaults to `false`, `previousPath` to `""`)

**Testing**:
- Verify struct compiles correctly
- Verify existing tests still pass

**Estimated Effort**: Small

---

### Phase 2: Implement Hidden File Filtering

**Goal**: Filter hidden files in directory listing based on `showHidden` flag

**Files to Modify**:
- `internal/ui/pane.go` - Modify `LoadDirectory()`, add helper function

**Implementation Steps**:

1. Add filter helper function:
   ```go
   // filterHiddenFiles removes entries starting with "." (except "..")
   func filterHiddenFiles(entries []fs.FileEntry) []fs.FileEntry {
       result := make([]fs.FileEntry, 0, len(entries))
       for _, e := range entries {
           // Always keep parent directory entry ".."
           if e.IsParentDir() || !strings.HasPrefix(e.Name, ".") {
               result = append(result, e)
           }
       }
       return result
   }
   ```

2. Modify `LoadDirectory()` to apply filter:
   ```go
   func (p *Pane) LoadDirectory() error {
       entries, err := fs.ReadDirectory(p.path)
       if err != nil {
           return err
       }

       fs.SortEntries(entries)

       // Filter hidden files if not showing
       if !p.showHidden {
           entries = filterHiddenFiles(entries)
       }

       p.entries = entries
       p.cursor = 0
       p.scrollOffset = 0

       return nil
   }
   ```

3. Update `LoadDirectoryAsync()` similarly (add filtering after sort)

**Testing**:
- Create test directory with hidden files
- Verify hidden files excluded when `showHidden = false`
- Verify hidden files included when `showHidden = true`
- Verify `..` entry always visible

**Estimated Effort**: Small

---

### Phase 3: Implement Toggle Hidden Method

**Goal**: Implement `ToggleHidden()` with cursor position preservation

**Files to Modify**:
- `internal/ui/pane.go` - Add `ToggleHidden()` and `IsShowingHidden()` methods

**Implementation Steps**:

1. Add `ToggleHidden()` method:
   ```go
   // ToggleHidden toggles visibility of hidden files
   func (p *Pane) ToggleHidden() {
       // Remember current selection
       var selectedName string
       if p.cursor >= 0 && p.cursor < len(p.entries) {
           selectedName = p.entries[p.cursor].Name
       }

       p.showHidden = !p.showHidden
       p.LoadDirectory()

       // Try to restore cursor position
       if selectedName != "" {
           for i, e := range p.entries {
               if e.Name == selectedName {
                   p.cursor = i
                   p.adjustScroll()
                   return
               }
           }
       }
       // If not found (file was hidden), reset to top
       p.cursor = 0
       p.scrollOffset = 0
   }
   ```

2. Add `IsShowingHidden()` accessor:
   ```go
   // IsShowingHidden returns whether hidden files are visible
   func (p *Pane) IsShowingHidden() bool {
       return p.showHidden
   }
   ```

**Testing**:
- Toggle from false to true: hidden files appear
- Toggle from true to false: hidden files disappear
- Cursor stays on same file if visible after toggle
- Cursor resets to 0 if current file becomes hidden

**Estimated Effort**: Small

---

### Phase 4: Add Visual Indicator for Hidden Files

**Goal**: Display `[H]` indicator in pane header when showing hidden files

**Files to Modify**:
- `internal/ui/pane.go` - Modify `formatPath()` or header rendering

**Implementation Steps**:

1. Modify `ViewWithDiskSpace()` to include indicator in path line:
   ```go
   // In ViewWithDiskSpace(), modify path display section:
   displayPath := p.formatPath()
   if p.showHidden {
       displayPath = "[H] " + displayPath
   }
   ```

2. Apply same change to `ViewDimmedWithDiskSpace()` for consistency

**Testing**:
- Verify `[H]` appears when `showHidden = true`
- Verify `[H]` disappears when `showHidden = false`
- Verify indicator visible in both normal and dimmed pane views

**Estimated Effort**: Small

---

### Phase 5: Implement Previous Directory Tracking

**Goal**: Track previous directory on navigation and implement swap behavior

**Files to Modify**:
- `internal/ui/pane.go` - Add helper method, modify navigation methods

**Implementation Steps**:

1. Add helper method to record previous path:
   ```go
   // recordPreviousPath saves current path before navigation
   func (p *Pane) recordPreviousPath() {
       p.previousPath = p.path
   }
   ```

2. Modify `EnterDirectory()` to record previous:
   ```go
   func (p *Pane) EnterDirectory() error {
       entry := p.SelectedEntry()
       if entry == nil {
           return nil
       }

       // ... existing symlink handling ...

       // Record before changing
       p.recordPreviousPath()

       // ... rest of navigation logic ...
   }
   ```

3. Modify `MoveToParent()` to record previous:
   ```go
   func (p *Pane) MoveToParent() error {
       if p.path == "/" {
           return nil
       }

       p.recordPreviousPath()
       p.path = filepath.Dir(p.path)
       return p.LoadDirectory()
   }
   ```

4. Modify `ChangeDirectory()` to record previous:
   ```go
   func (p *Pane) ChangeDirectory(path string) error {
       p.recordPreviousPath()
       p.path = path
       return p.LoadDirectory()
   }
   ```

**Testing**:
- Navigate to subdirectory: `previousPath` updated
- Navigate to parent: `previousPath` updated
- `previousPath` contains correct previous location

**Estimated Effort**: Small

---

### Phase 6: Implement Home and Previous Directory Navigation

**Goal**: Add `NavigateToHome()` and `NavigateToPrevious()` methods

**Files to Modify**:
- `internal/ui/pane.go` - Add new navigation methods

**Implementation Steps**:

1. Add `NavigateToHome()`:
   ```go
   // NavigateToHome navigates to user's home directory
   func (p *Pane) NavigateToHome() error {
       home, err := fs.HomeDirectory()
       if err != nil {
           return err
       }

       // Don't navigate if already at home
       if p.path == home {
           return nil
       }

       p.recordPreviousPath()
       p.path = home
       return p.LoadDirectory()
   }
   ```

2. Add `NavigateToPrevious()`:
   ```go
   // NavigateToPrevious navigates to previous directory (toggle behavior)
   func (p *Pane) NavigateToPrevious() error {
       if p.previousPath == "" {
           return nil // No previous directory
       }

       // Swap current and previous (toggle behavior)
       current := p.path
       p.path = p.previousPath
       p.previousPath = current

       return p.LoadDirectory()
   }
   ```

**Testing**:
- `NavigateToHome()` goes to home directory
- `NavigateToHome()` updates `previousPath`
- `NavigateToPrevious()` does nothing when no history
- `NavigateToPrevious()` toggles between two directories (A→B→A→B)

**Estimated Effort**: Small

---

### Phase 7: Add Key Bindings

**Goal**: Define new key constants

**Files to Modify**:
- `internal/ui/keys.go` - Add key constants

**Implementation Steps**:

1. Add key constants:
   ```go
   const (
       // ... existing keys ...
       KeyToggleHidden = "ctrl+h"
       KeyHome         = "~"
       KeyPrevDir      = "-"
   )
   ```

**Testing**:
- Verify no conflicts with existing key bindings
- Verify constants are accessible from model.go

**Estimated Effort**: Small

---

### Phase 8: Add Key Handlers in Model

**Goal**: Handle new key events in `Model.Update()`

**Files to Modify**:
- `internal/ui/model.go` - Add case handlers in keyboard switch

**Implementation Steps**:

1. Add handlers in the `tea.KeyMsg` switch statement:
   ```go
   case KeyToggleHidden:
       m.getActivePane().ToggleHidden()
       return m, nil

   case KeyHome:
       if err := m.getActivePane().NavigateToHome(); err != nil {
           m.dialog = NewErrorDialog(fmt.Sprintf("Cannot navigate to home: %v", err))
       }
       // Update disk space after navigation
       m.updateDiskSpace()
       return m, nil

   case KeyPrevDir:
       if err := m.getActivePane().NavigateToPrevious(); err != nil {
           m.dialog = NewErrorDialog(fmt.Sprintf("Cannot navigate: %v", err))
       }
       // Update disk space after navigation
       m.updateDiskSpace()
       return m, nil
   ```

**Testing**:
- Ctrl+H toggles hidden files in active pane
- ~ navigates to home directory
- - toggles between current and previous directory
- Error dialog shown on navigation failure

**Estimated Effort**: Small

---

### Phase 9: Update Help Dialog

**Goal**: Add new key bindings to help dialog

**Files to Modify**:
- `internal/ui/help_dialog.go` - Add entries for new keys

**Implementation Steps**:

1. Locate help content and add entries:
   ```
   Ctrl+H  Toggle hidden files
   ~       Go to home directory
   -       Go to previous directory
   ```

**Testing**:
- Verify help dialog shows new key bindings
- Verify descriptions are clear

**Estimated Effort**: Small

---

### Phase 10: Add Unit Tests

**Goal**: Comprehensive test coverage for new functionality

**Files to Modify**:
- `internal/ui/pane_test.go` - Add test functions

**Implementation Steps**:

1. Add tests for hidden file toggle:
   - `TestToggleHidden_TogglesState`
   - `TestToggleHidden_FilterHiddenFiles`
   - `TestToggleHidden_PreservesCursor`
   - `TestToggleHidden_IndependentPerPane`

2. Add tests for home navigation:
   - `TestNavigateToHome_Success`
   - `TestNavigateToHome_UpdatesPreviousPath`
   - `TestNavigateToHome_AlreadyAtHome`

3. Add tests for previous directory:
   - `TestNavigateToPrevious_NoHistory`
   - `TestNavigateToPrevious_ToggleBehavior`
   - `TestNavigateToPrevious_IndependentPerPane`

4. Add filter function test:
   - `TestFilterHiddenFiles`

**Testing**:
- All tests pass
- Coverage includes edge cases

**Estimated Effort**: Medium

---

## File Structure

```
internal/ui/
├── keys.go              # Add KeyToggleHidden, KeyHome, KeyPrevDir
├── pane.go              # Add showHidden, previousPath, new methods
├── pane_test.go         # Add tests for new functionality
├── model.go             # Add key handlers
└── help_dialog.go       # Update help content
```

## Testing Strategy

### Unit Tests
- `filterHiddenFiles()` function
- `ToggleHidden()` state changes and cursor preservation
- `NavigateToHome()` navigation and error handling
- `NavigateToPrevious()` toggle behavior
- Independent pane state (showHidden, previousPath)

### Integration Tests
- Key binding recognition (Ctrl+H, ~, -)
- Full navigation flow with history tracking
- Visual indicator display

### Manual Testing Checklist
- [ ] Ctrl+H toggles hidden files in left pane only
- [ ] Ctrl+H toggles hidden files in right pane only (after switching)
- [ ] [H] indicator appears when hidden files visible
- [ ] [H] indicator disappears when hidden files hidden
- [ ] ~ navigates to home directory
- [ ] ~ updates previous directory
- [ ] - returns to previous directory
- [ ] - toggles back and forth (A→B→A→B)
- [ ] - does nothing when no previous directory
- [ ] Error dialog on failed navigation
- [ ] Help dialog shows new key bindings

## Dependencies

### External Libraries
- None required

### Internal Dependencies
- Phase 2 depends on Phase 1 (struct fields)
- Phase 3 depends on Phase 2 (LoadDirectory changes)
- Phases 5-6 depend on Phase 1 (previousPath field)
- Phase 8 depends on Phase 7 (key constants)
- Phase 10 depends on all other phases

## Risk Assessment

### Technical Risks
- **Key conflict with terminal**: `Ctrl+H` may be interpreted as backspace in some terminals
  - Mitigation: Bubble Tea handles this correctly; test on various terminals

### Implementation Risks
- **Cursor position edge cases**: After toggling, current file might be hidden
  - Mitigation: Reset to position 0 if current file not found

## Performance Considerations

- `filterHiddenFiles()` is O(n) - acceptable for directory sizes
- No additional file system calls for filtering (done in memory)
- `previousPath` is a single string - negligible memory impact

## Security Considerations

- None specific to this feature
- Follows existing patterns for file system access

## Open Questions

None - all requirements clarified.

## Future Enhancements

- Stack-based directory history (multiple levels)
- Persistent configuration for default `showHidden` state
- Keyboard shortcut customization

## References

- [SPEC.md](./SPEC.md) - Feature specification
- [要件定義書.md](./要件定義書.md) - Requirements document (Japanese)
