# Context Menu Feature - Implementation Plan

## Overview

This document outlines the detailed implementation plan for adding context menu functionality to duofm. The context menu will be accessible via the `@` key and will display available file operations in a user-friendly dialog format, improving discoverability and reducing the learning curve for new users.

**Key Features:**
- Display context menu with `@` key
- Show operations: Copy, Move, Delete
- Special handling for symlinks (logical/physical path navigation)
- Keyboard-driven navigation (j/k/Enter/Esc/1-9)
- Consistent with existing dialog UI patterns
- Extensible for future feature additions

## Objectives

1. **Improve Discoverability**: Users can discover available operations without memorizing shortcuts
2. **Reduce Learning Curve**: Visual menu reduces barrier to entry for new users
3. **Maintain Consistency**: Follow existing dialog patterns and TUI conventions
4. **Enable Extensibility**: Design for easy addition of future menu items
5. **Symlink Support**: Provide advanced symlink navigation options

## Prerequisites

Before starting implementation, ensure:

- [x] Existing dialog system is stable (`Dialog` interface, `ConfirmDialog`, `ErrorDialog`, `HelpDialog`)
- [x] File operations are functional (`Copy`, `MoveFile`, `Delete` in `internal/fs/operations.go`)
- [x] Symlink detection is implemented (`FileEntry.IsSymlink`, `LinkTarget`, `LinkBroken`)
- [x] Bubble Tea framework and lipgloss are available
- [x] Development environment is set up (Go 1.21+, dependencies installed)

## Architecture Overview

### Design Decisions

1. **Dialog Interface Implementation**
   - `ContextMenuDialog` implements the existing `Dialog` interface
   - Ensures consistency with `ConfirmDialog`, `ErrorDialog`, etc.
   - Leverages existing dialog lifecycle management in `Model.Update()`

2. **Centered in Active Pane**
   - Menu appears centered within the active pane (not screen center)
   - Provides better context by showing which pane is being operated on
   - Works well even on narrow terminals (minimum 60 columns)

3. **Action Delegation Pattern**
   - Menu items don't duplicate logic; they trigger existing operations
   - Maintains DRY principle
   - Reduces maintenance burden and bug surface area

4. **MenuItem Structure**
   - Flexible design supporting both simple actions and complex workflows
   - `Action` as closure captures necessary context (paths, entries)
   - `Enabled` flag allows for future permission-based disabling

5. **Message-Based Result Communication**
   - Custom `contextMenuResultMsg` type for result communication
   - Follows Bubble Tea's message-passing pattern
   - Allows for deferred action execution (after dialog closes)

### Component Hierarchy

```
Model (internal/ui/model.go)
  ├── Dialog interface
  │   ├── ConfirmDialog (existing)
  │   ├── HelpDialog (existing)
  │   ├── ErrorDialog (existing)
  │   └── ContextMenuDialog (NEW)
  │       ├── MenuItem[] (menu items)
  │       ├── cursor (selection state)
  │       ├── currentPage (pagination state)
  │       └── buildMenuItems() (item generation)
  └── Update() method
      ├── Key handler for "@" (NEW)
      └── contextMenuResultMsg handler (NEW)
```

### Data Flow

```
User presses "@"
  ↓
Model.Update() receives KeyMsg
  ↓
Create ContextMenuDialog with:
  - FileEntry (selected file/directory)
  - Source path (active pane)
  - Destination path (inactive pane)
  ↓
ContextMenuDialog.buildMenuItems()
  - Generates MenuItem[] based on file type
  - Adds symlink-specific items if applicable
  ↓
Dialog.View() renders menu
  ↓
User navigates with j/k/1-9 and presses Enter
  ↓
ContextMenuDialog.Update() receives KeyMsg
  ↓
Returns contextMenuResultMsg with action closure
  ↓
Model.Update() receives contextMenuResultMsg
  ↓
Execute action closure
  ↓
Handle result (error dialog or reload directories)
```

## Implementation Phases

### Phase 1: Core Context Menu Dialog (3-4 hours)

**Goal**: Create the basic `ContextMenuDialog` structure and rendering logic.

**Files to Create:**
- `/home/sakura/go/src/duofm/internal/ui/context_menu_dialog.go`
  - Main dialog implementation
  - `ContextMenuDialog` struct with cursor and menu items
  - `MenuItem` struct for menu item definition
  - `Dialog` interface implementation
  - `View()` method with lipgloss styling

**Files to Modify:**
- None in this phase

**Implementation Steps:**

1. **Create context_menu_dialog.go skeleton** (30 min)
   ```go
   package ui

   import (
       "strings"
       tea "github.com/charmbracelet/bubbletea"
       "github.com/charmbracelet/lipgloss"
       "duofm/internal/fs"
   )

   // ContextMenuDialog represents the context menu
   type ContextMenuDialog struct {
       items        []MenuItem
       cursor       int
       currentPage  int
       itemsPerPage int
       active       bool
       width        int
       minWidth     int
       maxWidth     int
   }

   // MenuItem represents a single menu item
   type MenuItem struct {
       ID      string       // Unique identifier
       Label   string       // Display text
       Action  func() error // Action to execute
       Enabled bool         // Whether item is selectable
   }

   // contextMenuResultMsg is the result message
   type contextMenuResultMsg struct {
       action    func() error
       cancelled bool
   }
   ```

2. **Implement NewContextMenuDialog constructor** (30 min)
   ```go
   func NewContextMenuDialog(entry *fs.FileEntry, sourcePath, destPath string) *ContextMenuDialog {
       d := &ContextMenuDialog{
           cursor:       0,
           currentPage:  0,
           itemsPerPage: 9,
           active:       true,
           minWidth:     40,
           maxWidth:     60,
       }

       d.items = d.buildMenuItems(entry, sourcePath, destPath)
       d.calculateWidth()

       return d
   }
   ```

3. **Implement buildMenuItems logic** (1-1.5 hours)
   - Add basic operations (Copy, Move, Delete)
   - Add symlink-specific operations (Enter logical, Enter physical)
   - Handle broken symlinks (disable physical path navigation)
   - Create action closures that capture necessary context

4. **Implement View() method** (1-1.5 hours)
   - Create menu border with rounded corners
   - Render title with page indicator
   - Render menu items with numbering (1-9)
   - Highlight selected item with background color
   - Add footer with keyboard hints
   - Apply lipgloss styles for consistency

**Dependencies:**
- `internal/fs/types.go` for `FileEntry`
- `internal/ui/dialog.go` for `Dialog` interface
- External: `bubbletea`, `lipgloss`

**Testing Approach:**
- Manual testing: Create dialog instance and call `View()` to verify rendering
- Verify menu items are generated correctly for different file types
- Check visual appearance in terminal (80 columns and 60 columns)

**Estimated Effort**: 3-4 hours

**Deliverables:**
- `context_menu_dialog.go` with basic structure
- Menu rendering works (no interaction yet)
- Visual design matches specification

---

### Phase 2: Keyboard Navigation and Interaction (2-3 hours)

**Goal**: Implement keyboard navigation (j/k/Enter/Esc/1-9) and integrate with Bubble Tea's Update cycle.

**Files to Modify:**
- `/home/sakura/go/src/duofm/internal/ui/context_menu_dialog.go`
  - Add `Update()` method
  - Add `IsActive()` method
  - Implement keyboard navigation logic

**Implementation Steps:**

1. **Implement IsActive() method** (5 min)
   ```go
   func (d *ContextMenuDialog) IsActive() bool {
       return d.active
   }
   ```

2. **Implement Update() method with j/k navigation** (1 hour)
   ```go
   func (d *ContextMenuDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
       if !d.active {
           return d, nil
       }

       switch msg := msg.(type) {
       case tea.KeyMsg:
           switch msg.String() {
           case "j", "down":
               // Move cursor down with boundary check
               d.cursor++
               visibleItems := len(d.getCurrentPageItems())
               if d.cursor >= visibleItems {
                   d.cursor = 0
               }
               return d, nil

           case "k", "up":
               // Move cursor up with boundary check
               d.cursor--
               if d.cursor < 0 {
                   visibleItems := len(d.getCurrentPageItems())
                   d.cursor = visibleItems - 1
               }
               return d, nil

           case "esc":
               // Cancel and close
               d.active = false
               return d, func() tea.Msg {
                   return contextMenuResultMsg{cancelled: true}
               }

           case "enter":
               // Execute selected action
               items := d.getCurrentPageItems()
               if d.cursor >= 0 && d.cursor < len(items) {
                   selectedItem := items[d.cursor]
                   if selectedItem.Enabled {
                       d.active = false
                       return d, func() tea.Msg {
                           return contextMenuResultMsg{action: selectedItem.Action}
                       }
                   }
               }
               return d, nil
           }
       }

       return d, nil
   }
   ```

3. **Add numeric key (1-9) direct selection** (30 min)
   ```go
   // In Update() method, add:
   case "1", "2", "3", "4", "5", "6", "7", "8", "9":
       num := int(msg.String()[0] - '0') - 1
       items := d.getCurrentPageItems()
       if num >= 0 && num < len(items) {
           selectedItem := items[num]
           if selectedItem.Enabled {
               d.active = false
               return d, func() tea.Msg {
                   return contextMenuResultMsg{action: selectedItem.Action}
               }
           }
       }
       return d, nil
   ```

4. **Implement helper methods** (30 min)
   ```go
   func (d *ContextMenuDialog) getCurrentPageItems() []MenuItem {
       start := d.currentPage * d.itemsPerPage
       end := start + d.itemsPerPage
       if end > len(d.items) {
           end = len(d.items)
       }
       return d.items[start:end]
   }

   func (d *ContextMenuDialog) getTotalPages() int {
       return (len(d.items) + d.itemsPerPage - 1) / d.itemsPerPage
   }

   func (d *ContextMenuDialog) calculateWidth() {
       maxLabelWidth := 0
       for _, item := range d.items {
           labelWidth := len(item.Label) + 3 // "1. " prefix
           if labelWidth > maxLabelWidth {
               maxLabelWidth = labelWidth
           }
       }

       d.width = maxLabelWidth + 4 // Padding
       if d.width < d.minWidth {
           d.width = d.minWidth
       }
       if d.width > d.maxWidth {
           d.width = d.maxWidth
       }
   }
   ```

**Dependencies:**
- Phase 1 completion

**Testing Approach:**
- Unit tests: Verify cursor movement logic
- Unit tests: Verify numeric key selection
- Unit tests: Verify Enter/Esc behavior
- Manual testing: Interactive keyboard navigation

**Estimated Effort**: 2-3 hours

**Deliverables:**
- Full keyboard navigation working
- Dialog responds to all specified keys
- Cursor wraps correctly at boundaries

---

### Phase 3: Model Integration and Key Binding (1-2 hours)

**Goal**: Integrate `ContextMenuDialog` into the main Model and add `@` key binding.

**Files to Create:**
- None

**Files to Modify:**
- `/home/sakura/go/src/duofm/internal/ui/keys.go`
  - Add `KeyContextMenu` constant

- `/home/sakura/go/src/duofm/internal/ui/model.go`
  - Add `@` key handler in `Update()` method
  - Add `contextMenuResultMsg` handler
  - Integrate context menu lifecycle

**Implementation Steps:**

1. **Add key constant** (5 min)
   ```go
   // In keys.go
   const (
       // ... existing keys
       KeyContextMenu = "@"
   )
   ```

2. **Add context menu invocation in Model.Update()** (30 min)
   ```go
   // In model.go Update() method, add key handler:
   case KeyContextMenu:
       // Don't show menu for parent directory
       activePane := m.getActivePane()
       entry := activePane.SelectedEntry()

       if entry != nil && !entry.IsParentDir() {
           m.dialog = NewContextMenuDialog(
               entry,
               activePane.Path(),
               m.getInactivePane().Path(),
           )
       }
       return m, nil
   ```

3. **Add contextMenuResultMsg handler** (30-45 min)
   ```go
   // In model.go Update() method, add message handler:
   if result, ok := msg.(contextMenuResultMsg); ok {
       prevDialog := m.dialog
       m.dialog = nil

       if _, ok := prevDialog.(*ContextMenuDialog); ok {
           if result.cancelled {
               // Menu was cancelled, do nothing
               return m, nil
           }

           if result.action != nil {
               // Execute the selected action
               if err := result.action(); err != nil {
                   m.dialog = NewErrorDialog(fmt.Sprintf("Operation failed: %v", err))
                   return m, nil
               }

               // Reload both panes to reflect changes
               m.getActivePane().LoadDirectory()
               m.getInactivePane().LoadDirectory()
           }
       }

       return m, nil
   }
   ```

4. **Handle delete confirmation flow** (15-30 min)
   - Modify buildMenuItems to create delete action that:
     1. Shows ConfirmDialog
     2. Waits for confirmation
     3. Executes delete only if confirmed

   Note: This may require refactoring the action closure to return a Dialog instead of directly executing.

**Dependencies:**
- Phase 1 and 2 completion
- Existing `Model` structure and update logic
- Existing `getActivePane()` and `getInactivePane()` methods

**Testing Approach:**
- Integration tests: Verify `@` key opens dialog
- Integration tests: Verify parent directory protection
- Manual testing: Open menu, navigate, and execute actions
- Manual testing: Verify directory reloading after operations

**Estimated Effort**: 1-2 hours

**Deliverables:**
- `@` key opens context menu
- Menu integrates seamlessly with existing UI
- Parent directory check works
- Actions execute and directories reload

---

### Phase 4: Action Implementation (2-3 hours)

**Goal**: Wire up all menu item actions (Copy, Move, Delete, Symlink navigation) and handle errors properly.

**Files to Create:**
- None

**Files to Modify:**
- `/home/sakura/go/src/duofm/internal/ui/context_menu_dialog.go`
  - Complete `buildMenuItems()` with all action implementations

**Implementation Steps:**

1. **Implement Copy action** (30 min)
   ```go
   // In buildMenuItems():
   items = append(items, MenuItem{
       ID:    "copy",
       Label: "Copy to other pane",
       Action: func() error {
           fullPath := filepath.Join(sourcePath, entry.Name)
           return fs.Copy(fullPath, destPath)
       },
       Enabled: true,
   })
   ```

2. **Implement Move action** (30 min)
   ```go
   items = append(items, MenuItem{
       ID:    "move",
       Label: "Move to other pane",
       Action: func() error {
           fullPath := filepath.Join(sourcePath, entry.Name)
           return fs.MoveFile(fullPath, destPath)
       },
       Enabled: true,
   })
   ```

3. **Implement Delete action with confirmation** (1 hour)

   This is more complex because we need to integrate with the existing ConfirmDialog. There are two approaches:

   **Approach A**: Direct deletion (no confirmation in menu)
   - Menu item executes delete directly
   - Requires separate confirmation handling in Model

   **Approach B**: Nested dialog (show ConfirmDialog from menu)
   - More complex, may require refactoring dialog lifecycle

   **Recommended: Approach A**
   ```go
   items = append(items, MenuItem{
       ID:    "delete",
       Label: "Delete",
       Action: func() error {
           // This action will be wrapped by confirmation dialog
           // in the Model's contextMenuResultMsg handler
           fullPath := filepath.Join(sourcePath, entry.Name)
           return fs.Delete(fullPath)
       },
       Enabled: true,
   })
   ```

   Then in Model.Update() contextMenuResultMsg handler:
   ```go
   if result.action != nil {
       // Check if this is a delete action by examining the selected item
       // If delete, show confirmation dialog first
       if shouldConfirmAction(result) {
           m.dialog = NewConfirmDialog("Delete", "Are you sure?")
           m.pendingAction = result.action // Store for later
           return m, nil
       }

       // Execute non-delete actions directly
       if err := result.action(); err != nil {
           m.dialog = NewErrorDialog(fmt.Sprintf("Operation failed: %v", err))
           return m, nil
       }
       // ... reload directories
   }
   ```

4. **Implement symlink navigation actions** (1-1.5 hours)

   This requires integration with pane navigation logic.

   ```go
   // Enter as directory (logical path - follow symlink)
   if entry.IsSymlink && entry.IsDir && !entry.LinkBroken {
       items = append(items, MenuItem{
           ID:    "enter_logical",
           Label: "Enter as directory (logical path)",
           Action: func() error {
               // Navigate to the symlink itself (follow it)
               fullPath := filepath.Join(sourcePath, entry.Name)
               // This needs access to pane.ChangeDirectory()
               // May need to pass pane reference to buildMenuItems()
               return nil // Implementation depends on pane structure
           },
           Enabled: true,
       })

       // Open link target (physical path)
       items = append(items, MenuItem{
           ID:    "enter_physical",
           Label: "Open link target (physical path)",
           Action: func() error {
               // Navigate to parent directory of actual target
               targetPath := entry.LinkTarget
               if filepath.IsAbs(targetPath) {
                   return nil // Navigate to filepath.Dir(targetPath)
               } else {
                   absTarget := filepath.Join(sourcePath, targetPath)
                   return nil // Navigate to filepath.Dir(absTarget)
               }
           },
           Enabled: !entry.LinkBroken,
       })
   }
   ```

   **Note**: Symlink navigation requires refactoring to pass pane reference or navigation callback to `buildMenuItems()`.

5. **Add error handling** (30 min)
   - Ensure all actions return descriptive errors
   - Test permission denied scenarios
   - Test broken symlink scenarios

**Dependencies:**
- Phase 1, 2, 3 completion
- `internal/fs/operations.go` (Copy, MoveFile, Delete functions)
- Understanding of pane navigation for symlink actions

**Testing Approach:**
- Unit tests: Verify each action executes correctly
- Integration tests: Test copy/move/delete with real files
- Manual testing: Permission errors, disk full, etc.
- Manual testing: Symlink navigation (logical vs physical)

**Estimated Effort**: 2-3 hours

**Deliverables:**
- All actions fully implemented and working
- Delete action integrates with confirmation dialog
- Symlink navigation works correctly
- Proper error handling for all scenarios

---

### Phase 5: Testing and Quality Assurance (3-4 hours)

**Goal**: Create comprehensive test suite and perform thorough manual testing.

**Files to Create:**
- `/home/sakura/go/src/duofm/internal/ui/context_menu_dialog_test.go`
  - Unit tests for ContextMenuDialog

**Files to Modify:**
- `/home/sakura/go/src/duofm/internal/ui/model_test.go`
  - Integration tests for context menu in Model

**Implementation Steps:**

1. **Create unit test file structure** (30 min)
   ```go
   package ui

   import (
       "testing"
       "duofm/internal/fs"
       tea "github.com/charmbracelet/bubbletea"
   )

   func TestNewContextMenuDialog(t *testing.T) {
       // Test cases
   }

   func TestContextMenuDialog_Update(t *testing.T) {
       // Test cases
   }

   func TestContextMenuDialog_View(t *testing.T) {
       // Test cases
   }

   func TestBuildMenuItems(t *testing.T) {
       // Test cases
   }
   ```

2. **Write unit tests for menu creation** (1 hour)
   - Test regular file: 3 items (copy, move, delete)
   - Test directory: 3 items
   - Test symlink directory: 5 items (+ enter_logical, enter_physical)
   - Test broken symlink: verify enter_physical is disabled
   - Use table-driven tests for multiple scenarios

3. **Write unit tests for keyboard navigation** (1-1.5 hours)
   - Test j/k cursor movement
   - Test cursor wrapping at boundaries
   - Test numeric key (1-9) direct selection
   - Test Enter key action execution
   - Test Esc key cancellation
   - Test page navigation (h/l) if pagination implemented

4. **Write integration tests** (1 hour)
   - Test `@` key opens menu
   - Test parent directory protection
   - Test copy action execution
   - Test move action execution
   - Test delete action with confirmation
   - Test error dialog on operation failure

5. **Manual testing checklist** (30-45 min)
   - [ ] Visual appearance (80-column terminal)
   - [ ] Visual appearance (60-column terminal)
   - [ ] Menu centering in active pane
   - [ ] Highlight selected item
   - [ ] All keyboard shortcuts work
   - [ ] Copy operation works
   - [ ] Move operation works
   - [ ] Delete operation shows confirmation
   - [ ] Error handling (permission denied, etc.)
   - [ ] Symlink navigation (logical and physical)
   - [ ] Broken symlink handling
   - [ ] Parent directory protection

**Dependencies:**
- Phase 1-4 completion
- Test fixtures (sample files and directories)

**Testing Approach:**
- Table-driven tests for multiple scenarios
- Mock file operations if needed
- Real file operations in integration tests
- Manual testing on different terminal sizes

**Estimated Effort**: 3-4 hours

**Deliverables:**
- Comprehensive unit test suite
- Integration tests for Model interaction
- All tests passing
- Manual testing checklist completed
- Test coverage report (aim for 80%+)

---

### Phase 6: Documentation and Polish (1-2 hours)

**Goal**: Update documentation, add code comments, and final polish.

**Files to Modify:**
- `/home/sakura/go/src/duofm/internal/ui/help_dialog.go`
  - Add `@` key to help menu

- `/home/sakura/go/src/duofm/internal/ui/context_menu_dialog.go`
  - Add comprehensive code comments
  - Add package documentation

- `/home/sakura/go/src/duofm/README.md` (if exists)
  - Document new feature

- `/home/sakura/go/src/duofm/doc/tasks/context-menu/SPEC.md`
  - Mark all test scenarios as completed

**Implementation Steps:**

1. **Update HelpDialog** (15-20 min)
   - Add entry: `@: Show context menu`
   - Ensure alphabetical/logical ordering

2. **Add code documentation** (30-45 min)
   - Package comment for context_menu_dialog.go
   - Function comments for all exported functions
   - Inline comments for complex logic
   - Example usage in comments

3. **Update README** (15-20 min)
   - Add context menu to feature list
   - Document keyboard shortcuts
   - Add screenshot or example if applicable

4. **Final code review** (30 min)
   - Check code style consistency (gofmt)
   - Run go vet for static analysis
   - Run golangci-lint if available
   - Check for TODO comments

5. **Performance verification** (15 min)
   - Verify menu opens instantaneously
   - Test on large directories
   - Profile if any performance concerns

**Dependencies:**
- Phase 1-5 completion
- All tests passing

**Testing Approach:**
- Final manual testing of complete feature
- Verify help menu is accurate
- Check documentation is clear and complete

**Estimated Effort**: 1-2 hours

**Deliverables:**
- Updated help dialog
- Comprehensive code documentation
- Updated project documentation
- Feature complete and production-ready

---

## File Structure

### New Files

```
internal/ui/
├── context_menu_dialog.go        # Context menu dialog implementation
└── context_menu_dialog_test.go   # Unit tests for context menu
```

### Modified Files

```
internal/ui/
├── keys.go                        # Add KeyContextMenu constant
├── model.go                       # Add @ key handler and contextMenuResultMsg handler
├── model_test.go                  # Add integration tests
└── help_dialog.go                 # Add @ key to help menu

doc/tasks/context-menu/
├── SPEC.md                        # Mark test scenarios as completed
└── IMPLEMENTATION.md              # This file
```

### File Descriptions

**context_menu_dialog.go** (Estimated: ~300-400 lines)
- `ContextMenuDialog` struct
- `MenuItem` struct
- `contextMenuResultMsg` type
- `NewContextMenuDialog()` constructor
- `Update()` method (keyboard handling)
- `View()` method (rendering with lipgloss)
- `IsActive()` method
- `buildMenuItems()` helper
- `getCurrentPageItems()` helper
- `getTotalPages()` helper
- `calculateWidth()` helper

**context_menu_dialog_test.go** (Estimated: ~400-500 lines)
- `TestNewContextMenuDialog()` - test dialog creation
- `TestBuildMenuItems_RegularFile()` - verify 3 items
- `TestBuildMenuItems_Directory()` - verify 3 items
- `TestBuildMenuItems_Symlink()` - verify 5 items
- `TestBuildMenuItems_BrokenSymlink()` - verify disabled items
- `TestUpdate_NavigationJK()` - test j/k cursor movement
- `TestUpdate_NavigationNumeric()` - test 1-9 keys
- `TestUpdate_Enter()` - test action execution
- `TestUpdate_Esc()` - test cancellation
- `TestView()` - test rendering output
- Table-driven test utilities

## Testing Strategy

### Unit Tests (context_menu_dialog_test.go)

**Menu Creation Tests:**
- [x] Regular file creates 3 items (copy, move, delete)
- [x] Directory creates 3 items
- [x] Symlink directory creates 5 items (+ enter_logical, enter_physical)
- [x] Broken symlink disables enter_physical option
- [x] Menu width is calculated correctly
- [x] Items have correct labels and IDs

**Navigation Tests:**
- [x] j key moves cursor down
- [x] k key moves cursor up
- [x] Cursor wraps from bottom to top
- [x] Cursor wraps from top to bottom
- [x] Numeric keys 1-9 select items directly
- [x] Numeric key > item count is ignored
- [x] Enter executes selected item's action
- [x] Esc cancels and returns cancelled message
- [x] Disabled items are not selectable

**Rendering Tests:**
- [x] View() returns non-empty string
- [x] View() includes border characters
- [x] View() includes item labels
- [x] View() highlights selected item
- [x] View() shows correct page number (when pagination added)

### Integration Tests (model_test.go)

**Model Integration Tests:**
- [x] @ key opens context menu dialog
- [x] @ key does nothing when parent directory selected
- [x] Context menu is centered in active pane
- [x] Copy action copies file to opposite pane
- [x] Move action moves file to opposite pane
- [x] Delete action shows confirmation dialog
- [x] Confirmed delete removes file
- [x] Cancelled delete preserves file
- [x] Error in action shows error dialog
- [x] Successful action reloads both panes

### Manual Testing

**Visual Testing:**
- [ ] Menu displays correctly on 80-column terminal
- [ ] Menu displays correctly on 60-column terminal
- [ ] Menu is centered in active pane (not screen center)
- [ ] Selected item is clearly highlighted
- [ ] Border and styling match other dialogs
- [ ] Items are numbered 1-9
- [ ] Footer shows keyboard hints

**Functional Testing:**
- [ ] @ key opens menu
- [ ] Parent directory (..) prevents menu
- [ ] j/k navigation works smoothly
- [ ] Numeric keys 1-9 work
- [ ] Enter executes selected action
- [ ] Esc closes menu
- [ ] Copy operation works correctly
- [ ] Move operation works correctly
- [ ] Delete shows confirmation
- [ ] Symlink shows additional options
- [ ] Logical path navigation works
- [ ] Physical path navigation works
- [ ] Broken symlink disables physical option

**Error Testing:**
- [ ] Permission denied shows error dialog
- [ ] Disk full shows error dialog
- [ ] Invalid path shows error dialog
- [ ] Broken symlink handled gracefully

**Edge Cases:**
- [ ] Empty directory
- [ ] Single file
- [ ] Many files (stress test)
- [ ] Very long file names
- [ ] Special characters in file names
- [ ] Symlink loops

## Dependencies

### Internal Dependencies

**Required Packages:**
- `internal/ui/dialog.go` - Dialog interface definition
- `internal/ui/model.go` - Main model structure and update logic
- `internal/ui/pane.go` - Pane structure, SelectedEntry(), Path() methods
- `internal/ui/keys.go` - Keybinding constants
- `internal/ui/confirm_dialog.go` - ConfirmDialog for delete confirmation
- `internal/ui/error_dialog.go` - ErrorDialog for operation errors
- `internal/fs/types.go` - FileEntry structure
- `internal/fs/operations.go` - Copy, MoveFile, Delete functions

**Required Methods/Functions:**
- `Model.getActivePane()` - Get active pane
- `Model.getInactivePane()` - Get inactive pane
- `Pane.SelectedEntry()` - Get selected FileEntry
- `Pane.Path()` - Get pane directory path
- `Pane.LoadDirectory()` - Reload pane contents
- `FileEntry.IsParentDir()` - Check if entry is ".."
- `FileEntry.IsSymlink` - Check if symlink
- `FileEntry.LinkTarget` - Get symlink target
- `FileEntry.LinkBroken` - Check if symlink is broken

### External Dependencies

**Go Standard Library:**
- `fmt` - String formatting
- `path/filepath` - Path manipulation
- `strings` - String utilities

**Third-party Packages:**
- `github.com/charmbracelet/bubbletea` - TUI framework
  - `tea.Msg` - Message interface
  - `tea.KeyMsg` - Keyboard input
  - `tea.Cmd` - Command type
- `github.com/charmbracelet/lipgloss` - Styling and layout
  - Border styles
  - Color support
  - Layout helpers (Place, Join)

**No new dependencies required** - all dependencies already exist in the project.

## Risk Assessment

### Technical Risks

**Risk 1: Dialog Lifecycle Complexity**
- **Impact**: High
- **Probability**: Medium
- **Description**: Managing nested dialogs (ConfirmDialog after ContextMenu) may introduce complexity in Model.Update() logic
- **Mitigation**:
  - Design clear state machine for dialog transitions
  - Store pending action in Model when confirmation is needed
  - Thorough integration testing of dialog chains
  - Consider refactoring dialog lifecycle if needed

**Risk 2: Action Closure Memory Leaks**
- **Impact**: Low
- **Probability**: Low
- **Description**: MenuItem actions capture local variables in closures, potential for memory leaks if not handled properly
- **Mitigation**:
  - Keep closures simple and short-lived
  - Dialog lifetime is brief (user closes quickly)
  - Go's garbage collector handles this well
  - Profile memory usage if concerns arise

**Risk 3: Symlink Navigation Integration**
- **Impact**: Medium
- **Probability**: Medium
- **Description**: Symlink navigation requires pane directory change, which may not be easily accessible from menu item actions
- **Mitigation**:
  - Pass pane reference or navigation callback to buildMenuItems()
  - Refactor action type to support navigation commands
  - Defer implementation to later phase if too complex
  - Document navigation API requirements

**Risk 4: Terminal Size Variability**
- **Impact**: Medium
- **Probability**: Low
- **Description**: Menu may not render correctly on very narrow terminals (< 60 columns)
- **Mitigation**:
  - Set minimum width requirements in documentation
  - Test on various terminal sizes (60, 80, 120 columns)
  - Implement adaptive width calculation
  - Graceful degradation for narrow terminals

**Risk 5: Performance on Large Item Lists**
- **Impact**: Low
- **Probability**: Low
- **Description**: If many menu items are added in future, rendering may slow down
- **Mitigation**:
  - Initial version has only 3-5 items (minimal impact)
  - Pagination designed for future expansion
  - Profile rendering if item count exceeds 10
  - Optimize View() method if needed

### Implementation Risks

**Risk 6: Test Coverage Gaps**
- **Impact**: Medium
- **Probability**: Medium
- **Description**: Complex dialog interactions may have edge cases not covered by tests
- **Mitigation**:
  - Aim for 80%+ test coverage
  - Table-driven tests for multiple scenarios
  - Integration tests for complete workflows
  - Thorough manual testing checklist

**Risk 7: Inconsistent Styling**
- **Impact**: Low
- **Probability**: Low
- **Description**: Menu styling may not match existing dialogs perfectly
- **Mitigation**:
  - Review existing dialog styles (ConfirmDialog, HelpDialog)
  - Use same lipgloss color constants
  - Copy border and padding patterns
  - Visual comparison testing

**Risk 8: Scope Creep**
- **Impact**: Medium
- **Probability**: Medium
- **Description**: Feature requests may expand scope beyond initial requirements (e.g., multi-select, custom actions)
- **Mitigation**:
  - Stick to SPEC.md requirements for initial version
  - Document future enhancements separately
  - Defer Phase 2+ features to later releases
  - Clear communication with stakeholders

## Performance Considerations

### Requirements

From SPEC.md NFR1 (Performance):
- Menu display must be instantaneous (within 100ms)
- Key input response must be immediate (no rendering delay)

### Optimization Strategies

**1. Menu Item Generation**
- `buildMenuItems()` is called once during dialog creation
- Pre-generate all items, not on every render
- Minimal overhead: 3-5 items for basic files, max ~10 items future
- **Expected performance**: < 1ms

**2. View Rendering**
- Use lipgloss caching where possible
- Avoid recreating styles on every render
- Pre-calculate static layout elements
- **Expected performance**: < 10ms for rendering

**3. Keyboard Input Handling**
- Direct key matching (no regex or complex parsing)
- Simple cursor arithmetic (no expensive calculations)
- Immediate state updates
- **Expected performance**: < 1ms for input handling

**4. Dialog Lifecycle**
- Dialog lifetime is very short (< 5 seconds typically)
- No background goroutines or async operations
- Memory usage is minimal (< 1 KB per dialog instance)
- Garbage collected immediately after close

### Profiling Plan

If performance issues are observed:
1. Use `go test -bench` for benchmark tests
2. Use `pprof` for CPU profiling
3. Measure View() rendering time
4. Measure buildMenuItems() execution time
5. Profile on large directories (1000+ files) to ensure no impact

### Performance Testing

- [ ] Menu opens in < 100ms (manual testing with stopwatch)
- [ ] No visible lag on key input
- [ ] Rendering smooth on 80-column terminal
- [ ] No memory leaks over repeated open/close cycles
- [ ] Works well on older hardware (low-spec VMs)

## Security Considerations

### Input Validation

**File Path Validation:**
- All file paths are constructed using `filepath.Join()` to prevent path traversal
- Use `filepath.Clean()` to normalize paths
- Never accept raw user input for paths (paths come from FileEntry)

**Action Execution:**
- Actions use `fs.Copy()`, `fs.MoveFile()`, `fs.Delete()` which already have validation
- No shell command execution (no `exec.Command()` calls)
- No string interpolation in file operations

### Permission Handling

**Operation Permissions:**
- Do not pre-check permissions (per SPEC.md FR7.1)
- Detect permission errors at execution time
- Display error dialog with informative message
- Never escalate privileges or use sudo

**File Access:**
- All file operations respect Unix file permissions
- Symlink following respects user permissions
- No bypassing of OS security controls

### Symlink Security

**Symlink Handling:**
- Detect broken symlinks before navigation
- Prevent infinite symlink loops (rely on OS limits)
- Use `filepath.EvalSymlinks()` for resolution
- Display clear error if symlink target is inaccessible

**Path Traversal Prevention:**
- Never construct paths from user string input
- Always use filepath package functions
- Validate resolved symlink paths are within expected bounds

### Error Information Disclosure

**Error Messages:**
- Show user-friendly errors, not internal paths
- Don't expose sensitive system information
- Sanitize error messages from OS calls
- Example: "Permission denied" not "/etc/shadow: permission denied"

### Denial of Service

**Resource Limits:**
- Menu item count is small (3-10 items)
- No unbounded loops or recursion
- Dialog lifetime is short and user-controlled
- No network operations or external process spawning

### Code Injection

**No Risk:**
- No dynamic code execution
- No eval() or reflection-based execution
- No template rendering with user input
- All operations are type-safe Go code

## Open Questions

### Q1: How should we handle delete confirmation?

**Options:**
1. Show ConfirmDialog immediately after menu selection
2. Execute delete directly (no confirmation) - not recommended
3. Add confirmation as part of the menu flow

**Decision**: Option 1 - Show ConfirmDialog after menu selection
- Consistent with existing delete operation (d key)
- User expects confirmation for destructive actions
- Requires storing pending action in Model

**Implementation Note:**
- In contextMenuResultMsg handler, detect delete action
- Show ConfirmDialog instead of executing immediately
- Store action in Model.pendingAction
- Execute on confirmation, discard on cancellation

**Status**: ✓ Resolved

---

### Q2: Should symlink navigation actions change active pane or inactive pane?

**Options:**
1. Change active pane directory (where symlink is located)
2. Change inactive pane directory
3. Let user choose via menu item

**Decision**: Option 1 - Change active pane directory
- More intuitive: user navigates "into" the selected item
- Consistent with Enter key behavior
- Inactive pane remains unchanged (useful for context)

**Implementation Note:**
- Action closure needs access to active pane's ChangeDirectory() method
- May require refactoring buildMenuItems() to accept pane reference
- Alternative: return navigation message instead of executing directly

**Status**: ✓ Resolved

---

### Q3: How to handle pagination for future menu expansion?

**Options:**
1. Implement pagination now (9 items per page, h/l keys)
2. Defer pagination until needed (when > 10 items exist)
3. Use scrolling instead of pagination

**Decision**: Option 2 - Defer pagination until needed
- Current menu has only 3-5 items
- Pagination adds complexity without immediate benefit
- Design supports pagination (currentPage, itemsPerPage fields exist)
- Can be added in Phase 2 when more actions are available

**Implementation Note:**
- Keep pagination-related fields in ContextMenuDialog struct
- Keep getTotalPages() and getCurrentPageItems() methods
- Don't implement h/l key handlers yet
- Document as future enhancement

**Status**: ✓ Resolved

---

### Q4: Should we add keyboard shortcuts in menu item labels?

**Options:**
1. Show shortcuts: "1. Copy to other pane (c)"
2. Don't show shortcuts: "1. Copy to other pane"
3. Show only for items with shortcuts

**Decision**: Option 2 - Don't show shortcuts in labels
- Keeps labels clean and concise
- Numeric keys (1-9) are already shortcuts
- Power users already know keyboard shortcuts
- Menu is for discoverability, not shortcut reference

**Implementation Note:**
- Help dialog (h key) shows all shortcuts separately
- Menu focuses on action description, not shortcuts
- Shorter labels reduce menu width

**Status**: ✓ Resolved

---

### Q5: How to handle action execution errors that need user intervention?

**Example**: Copy fails due to duplicate file name
**Options:**
1. Show error dialog only
2. Show error dialog with retry option
3. Cancel entire operation

**Decision**: Option 1 - Show error dialog only (for initial version)
- Simple and consistent with current error handling
- User can retry manually (reopen menu, try again)
- Retry logic can be added in Phase 2 if needed

**Implementation Note:**
- contextMenuResultMsg handler catches errors
- Shows ErrorDialog with error message
- User dismisses error and continues
- No partial state preservation

**Status**: ✓ Resolved

---

### Q6: Should menu remember last cursor position when reopened?

**Options:**
1. Always start at first item (cursor = 0)
2. Remember last cursor position per file
3. Remember last cursor position globally

**Decision**: Option 1 - Always start at first item
- Simpler implementation
- Predictable behavior
- Menu lifetime is very short (few seconds)
- Small menu size makes this a non-issue

**Implementation Note:**
- cursor initialized to 0 in NewContextMenuDialog()
- No persistence of cursor state
- Can be added later if user feedback requests it

**Status**: ✓ Resolved

---

## Future Enhancements

### Phase 2: Additional File Operations

**Rename Operation:**
- Menu item: "Rename"
- Show input dialog for new name
- Validate name (no slashes, no empty string)
- Execute rename and reload directory

**Create Directory:**
- Menu item: "New directory"
- Show input dialog for directory name
- Create directory with default permissions (0755)
- Reload directory and select new directory

**Create File:**
- Menu item: "New file"
- Show input dialog for file name
- Create empty file with default permissions (0644)
- Reload directory and select new file

**File Properties:**
- Menu item: "Properties"
- Show dialog with detailed information:
  - Full path
  - Size (human-readable)
  - Permissions (octal and symbolic)
  - Owner and group
  - Timestamps (created, modified, accessed)
  - Inode number
  - Symlink target (if applicable)

**Change Permissions:**
- Menu item: "Change permissions"
- Show dialog with permission editor
- Visual checkboxes for rwx permissions (owner, group, other)
- Execute chmod and reload directory

### Phase 3: Archive Operations

**Create Archive:**
- Menu items: "Create tar.gz", "Create zip"
- Show input dialog for archive name
- Execute tar/zip command
- Reload directory and select new archive

**Extract Archive:**
- Menu item: "Extract here" (when archive file selected)
- Auto-detect archive type (.tar.gz, .zip, .tar.bz2, etc.)
- Extract to current directory
- Reload directory

### Phase 4: Mark Feature Integration

When mark feature is implemented in duofm:

**Marked File Support:**
- Show "X files marked" in menu header
- All operations apply to marked files instead of single file
- Example: "Copy 5 marked files to other pane"

**Mark Management Menu Items:**
- "Mark all" - Mark all files in directory
- "Unmark all" - Clear all marks
- "Invert selection" - Toggle marks on all files
- "Mark by pattern" - Show input dialog for glob pattern

**Bulk Operations:**
- Copy multiple files
- Move multiple files
- Delete multiple files (single confirmation)
- Archive multiple files

### Phase 5: Pagination Implementation

When menu items exceed 9 items:

**Pagination Features:**
- 9 items per page (leaving 10th slot for navigation hint)
- Page indicator in title: "Context Menu (2/3)"
- Navigation hints in footer: "h:prev l:next"
- h key: Previous page
- l key: Next page
- Numeric keys 1-9 apply to current page only
- Cursor resets to 0 when changing pages

**Implementation:**
- Enable h/l key handlers in Update() method
- Update View() to show page indicator
- Use getCurrentPageItems() for rendering
- Add boundary checks for page navigation

### Phase 6: Keybinding Customization

**Configuration File Support:**
```toml
# ~/.config/duofm/config.toml
[keybindings]
context_menu = "@"
copy = "c"
move = "m"
delete = "d"
help = "h"
quit = "q"
```

**Implementation:**
- Add config package to load TOML file
- Replace constants in keys.go with config values
- Add config validation
- Fallback to defaults if config not found
- Document custom keybindings in help dialog

### Phase 7: Plugin System

Allow users to add custom menu items:

**Plugin Configuration:**
```toml
# ~/.config/duofm/plugins.toml
[[plugins]]
name = "Open in VSCode"
command = "code {file}"
key = "v"
enabled = true

[[plugins]]
name = "Git Status"
command = "git status"
key = "g"
enabled = true
requires_git_repo = true
```

**Implementation:**
- Load plugin definitions from config
- Execute external commands safely
- Capture command output
- Show output in scrollable dialog
- Implement plugin enable/disable
- Add plugin menu to context menu

### Phase 8: Advanced Symlink Features

**Symlink Creation:**
- Menu item: "Create symlink"
- Show input dialog for symlink name
- Create relative or absolute symlink
- Option to choose symbolic vs hard link

**Symlink Resolution:**
- Menu item: "Resolve symlink chain"
- Show dialog with full symlink chain
- Display each level of redirection
- Show final target

**Symlink Repair:**
- Detect broken symlinks
- Menu item: "Repair symlink"
- Browse for new target
- Update symlink target

## Implementation Timeline

### Week 1: Core Implementation

**Day 1-2**: Phase 1 (Core Context Menu Dialog)
- Create context_menu_dialog.go
- Implement basic structure and rendering
- Visual testing

**Day 3**: Phase 2 (Keyboard Navigation)
- Implement Update() method
- Add j/k/Enter/Esc/1-9 handling
- Manual testing

**Day 4**: Phase 3 (Model Integration)
- Add @ key binding
- Integrate with Model
- Parent directory protection

**Day 5**: Phase 4 (Action Implementation)
- Wire up Copy/Move/Delete actions
- Symlink navigation actions
- Error handling

### Week 2: Testing and Polish

**Day 6-7**: Phase 5 (Testing)
- Write unit tests
- Write integration tests
- Manual testing
- Fix bugs

**Day 8**: Phase 6 (Documentation)
- Update help dialog
- Code documentation
- README updates
- Final review

**Total Estimated Time**: 7-11 hours (as per SPEC.md)
**Actual Calendar Time**: 8 working days (with buffer for review and testing)

## Success Criteria Checklist

### Functional Success

- [ ] `@` key displays context menu
- [ ] Menu appears centered in active pane
- [ ] Menu shows "Copy to other pane" option
- [ ] Menu shows "Move to other pane" option
- [ ] Menu shows "Delete" option
- [ ] Symlinks show "Enter as directory (logical path)" option
- [ ] Symlinks show "Open link target (physical path)" option
- [ ] j/k keys navigate menu items
- [ ] Numeric keys 1-9 select items directly
- [ ] Enter key executes selected action
- [ ] Esc key closes menu without action
- [ ] Parent directory (..) prevents menu display
- [ ] Copy operation copies file to opposite pane
- [ ] Move operation moves file to opposite pane
- [ ] Delete operation shows confirmation dialog
- [ ] Delete confirmation works correctly
- [ ] Symlink logical navigation works
- [ ] Symlink physical navigation works
- [ ] Broken symlink disables physical option

### Quality Success

- [ ] All existing unit tests still pass
- [ ] New unit tests for ContextMenuDialog pass
- [ ] Integration tests for Model pass
- [ ] Test coverage is 80%+ for new code
- [ ] Code passes `go vet` with no warnings
- [ ] Code passes `gofmt` (is properly formatted)
- [ ] No TODO comments in production code
- [ ] Menu display is instantaneous (< 100ms subjectively)
- [ ] No rendering lag on key input

### User Experience Success

- [ ] Menu is visually consistent with existing dialogs
- [ ] Selected item is clearly highlighted
- [ ] Menu items are concise and understandable
- [ ] Keyboard shortcuts are intuitive
- [ ] Error messages are helpful and actionable
- [ ] Menu works well on 80-column terminal
- [ ] Menu works acceptably on 60-column terminal
- [ ] Beginners can discover features without knowing shortcuts
- [ ] Symlink logical vs physical distinction is clear

### Documentation Success

- [ ] Help dialog includes `@` key
- [ ] Code has comprehensive comments
- [ ] README documents new feature
- [ ] SPEC.md test scenarios marked complete
- [ ] Implementation plan is followed

## Conclusion

This implementation plan provides a comprehensive roadmap for adding context menu functionality to duofm. The phased approach ensures incremental progress with testable milestones, while the detailed specifications reduce ambiguity and implementation risk.

**Key Takeaways:**
- Total estimated effort: 12-18 hours (including testing and documentation)
- Phased implementation allows for early validation and iteration
- Extensive testing strategy ensures quality and reliability
- Clear success criteria provide objective completion metrics
- Future enhancements are documented but deferred to maintain scope

**Next Steps:**
1. Review this implementation plan
2. Set up development branch
3. Begin Phase 1 implementation
4. Follow test-driven development (TDD) approach where applicable
5. Regular progress reviews after each phase

**Questions or Concerns:**
- Raise any questions about technical approach before starting
- Flag any risks or dependencies not covered in this plan
- Suggest alternative approaches if better solutions exist
- Request clarification on any ambiguous requirements
