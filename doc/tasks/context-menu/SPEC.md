# Feature: Context Menu

## Overview

This feature adds a context menu functionality to duofm, accessible via the `@` key. The context menu displays available actions for the currently selected file or directory, allowing users to visually select and execute operations using keyboard navigation.

This enhancement improves discoverability of features and reduces the learning curve for new users who may not be familiar with keyboard shortcuts.

## Objectives

- Improve operation discoverability by displaying available actions in a menu
- Reduce learning cost for beginners by eliminating the need to memorize all keybindings
- Ensure extensibility for future feature additions
- Enable advanced symlink operations (logical path vs physical path navigation)
- Establish foundation for future keybinding customization support

## User Stories

- As a new user, I want to see available actions in a menu, so that I don't need to memorize all keyboard shortcuts
- As a user, I want to access the context menu with a single key press, so that I can quickly perform file operations
- As a user working with symlinks, I want to choose between following the link or opening the target's parent directory, so that I have full control over symlink navigation
- As a power user, I want to use both keyboard shortcuts and the context menu, so that I can choose the most efficient method for each situation

## Technical Requirements

### Functional Requirements

#### FR1: Context Menu Display

- FR1.1: Pressing `@` key displays the context menu
- FR1.2: Menu appears as a dialog **centered in the active pane**
- FR1.3: Background (file list) is overlaid with semi-transparent or dark background (similar to existing dialogs)
- FR1.4: Menu is enclosed with rounded border (consistent with existing dialog style)
- FR1.5: Menu is not displayed when parent directory (`..`) is selected

#### FR2: Menu Items

Initial version includes the following three actions:

- FR2.1: Copy (to opposite pane) - same behavior as `c` key
- FR2.2: Move (to opposite pane) - same behavior as `m` key
- FR2.3: Delete - same behavior as `d` key (shows confirmation dialog)

#### FR3: Symlink Special Handling

When a symlink is selected:

- FR3.1: In addition to normal actions (copy/move/delete), add the following:
  - "Enter as directory (logical path)": Follow symlink and enter target directory (default `Enter` key behavior)
  - "Open link target (physical path)": Open parent directory of the actual file/directory
- FR3.2: When link is broken, disable or gray out "Open link target" option

#### FR4: Menu Navigation

- FR4.1: `j`/`k` keys move selection up/down (consistent with existing cursor movement)
- FR4.2: `Enter` key executes the selected item
- FR4.3: `Esc` key cancels and closes the menu
- FR4.4: Numeric keys `1`-`9` directly select corresponding items (1st through 9th)
- FR4.5: `h`/`l` keys for page navigation (when more than 10 items exist)
  - `h`: Previous page
  - `l`: Next page

#### FR5: Pagination (Future Extension)

When more than 10 menu items exist:

- FR5.1: Display 9 items per page
- FR5.2: Show page number in header (e.g., "Context Menu (1/3)")
- FR5.3: Show navigation hints in footer (e.g., "h:prev l:next Esc:cancel")
- FR5.4: Numeric keys 1-9 apply only to items on current page

#### FR6: Keybinding Design

- FR6.1: Define `@` key binding as constant in `internal/ui/keys.go` (`KeyContextMenu = "@"`)
- FR6.2: Implement with future configuration file support in mind (avoid hardcoding)
- FR6.3: Follow same pattern as existing key handlers (c, m, d, etc.)

#### FR7: Error Handling

- FR7.1: Detect permission errors at operation execution time and notify via error dialog (no pre-checking)
- FR7.2: During loading (directory read), follow existing implementation (no special restrictions on menu display)

### Non-functional Requirements

#### NFR1: Performance

- NFR1.1: Menu display is instantaneous (within 100ms)
- NFR1.2: Key input response is immediate (no rendering delay)

#### NFR2: Usability

- NFR2.1: Menu width adjusts to item length, with minimum 40 columns and maximum 60 columns
- NFR2.2: Selected item is highlighted (background color change)
- NFR2.3: Items selectable by numeric keys display numbers (e.g., `1. Copy to other pane`)
- NFR2.4: Menu items use concise, clear wording (verb + object)

#### NFR3: Consistency

- NFR3.1: Use same UI pattern as existing dialogs (ConfirmDialog, HelpDialog, etc.)
- NFR3.2: Maintain consistency with existing keybindings (j/k/h/l/Enter/Esc)
- NFR3.3: Error messages follow same format as existing ErrorDialog

#### NFR4: Maintainability

- NFR4.1: Context menu implements the Dialog interface
- NFR4.2: Menu item definitions are easily extensible
- NFR4.3: Include unit tests (table-driven tests)

#### NFR5: Screen Size Support

- NFR5.1: Recommended minimum width: 80 columns (40 columns per pane)
- NFR5.2: Minimum operating width: 60 columns (30 columns per pane, degraded mode)
- NFR5.3: Centering in active pane ensures visibility even on narrow terminals

## Implementation Approach

### Architecture

The context menu follows the existing dialog pattern in duofm:

```
Model (internal/ui/model.go)
  ├── dialog (Dialog interface)
  │   ├── ConfirmDialog
  │   ├── HelpDialog
  │   ├── ErrorDialog
  │   └── ContextMenuDialog (NEW)
  └── Update() handles dialog lifecycle
```

**Key Design Decisions:**

1. **Dialog Interface Implementation**: ContextMenuDialog implements the existing `Dialog` interface for consistency
2. **Centered in Active Pane**: Menu is positioned in the center of the active pane (not screen center) for better context
3. **Action Delegation**: Menu items trigger existing operations (copy, move, delete) to maintain DRY principle
4. **Keybinding Constant**: Define `KeyContextMenu` in `keys.go` for future configurability

### Component Design

#### 1. ContextMenuDialog (internal/ui/context_menu_dialog.go)

```go
type ContextMenuDialog struct {
    items         []MenuItem
    cursor        int
    currentPage   int
    itemsPerPage  int
    active        bool
    width         int
}

type MenuItem struct {
    ID          string      // Unique identifier (e.g., "copy", "move", "delete")
    Label       string      // Display text (e.g., "Copy to other pane")
    Action      func() error // Action to execute
    Enabled     bool        // Whether item is selectable
}

func NewContextMenuDialog(entry *fs.FileEntry, sourcePath string, destPath string) *ContextMenuDialog
func (d *ContextMenuDialog) Update(msg tea.Msg) (Dialog, tea.Cmd)
func (d *ContextMenuDialog) View() string
func (d *ContextMenuDialog) IsActive() bool
func (d *ContextMenuDialog) buildMenuItems(entry *fs.FileEntry, sourcePath, destPath string) []MenuItem
```

#### 2. Keybinding Constant (internal/ui/keys.go)

```go
const (
    // Existing keys...
    KeyContextMenu = "@"
)
```

#### 3. Model Integration (internal/ui/model.go)

In the `Update` method, add handler for `@` key:

```go
case KeyContextMenu:
    // Don't show menu for parent directory
    entry := m.getActivePane().SelectedEntry()
    if entry != nil && !entry.IsParentDir() {
        m.dialog = NewContextMenuDialog(
            entry,
            m.getActivePane().Path(),
            m.getInactivePane().Path(),
        )
    }
    return m, nil
```

Handle context menu result:

```go
if result, ok := msg.(contextMenuResultMsg); ok {
    prevDialog := m.dialog
    m.dialog = nil

    if _, ok := prevDialog.(*ContextMenuDialog); ok {
        // Execute selected action
        if result.action != nil {
            if err := result.action(); err != nil {
                m.dialog = NewErrorDialog(fmt.Sprintf("Failed: %v", err))
            } else {
                // Reload directories if needed
                m.getActivePane().LoadDirectory()
                m.getInactivePane().LoadDirectory()
            }
        }
    }

    return m, nil
}
```

### Data Structures

#### MenuItem Structure

```go
type MenuItem struct {
    ID      string       // "copy", "move", "delete", "enter_logical", "enter_physical"
    Label   string       // User-facing text
    Action  func() error // Action closure
    Enabled bool         // Grayed out if false
}
```

#### Context Menu Result Message

```go
type contextMenuResultMsg struct {
    action func() error
    cancelled bool
}
```

### UI/UX Design

#### Visual Design

```
┌─ Context Menu (1/1) ─────────────────┐
│                                       │
│  1. Copy to other pane                │
│  2. Move to other pane                │
│  3. Delete                            │
│                                       │
│  j/k:select Enter:confirm Esc:cancel  │
└───────────────────────────────────────┘
```

**For Symlinks:**

```
┌─ Context Menu (1/1) ─────────────────┐
│                                       │
│  1. Copy to other pane                │
│  2. Move to other pane                │
│  3. Delete                            │
│  4. Enter as directory (logical)      │
│  5. Open link target (physical)       │
│                                       │
│  j/k:select Enter:confirm Esc:cancel  │
└───────────────────────────────────────┘
```

**With Pagination (Future):**

```
┌─ Context Menu (2/3) ─────────────────┐
│                                       │
│  1. Item 10                           │
│  2. Item 11                           │
│  ...                                  │
│  9. Item 18                           │
│                                       │
│  h:prev l:next Esc:cancel             │
└───────────────────────────────────────┘
```

#### Styling (lipgloss)

- **Border**: Rounded border with accent color (Color "39" - blue)
- **Title**: Bold, accent color
- **Selected Item**: Background highlight (Color "39"), white foreground
- **Unselected Item**: Default foreground
- **Disabled Item**: Gray foreground (Color "240")
- **Footer**: Muted gray foreground (Color "245")

#### Positioning

The menu is centered within the **active pane**, not the entire screen:

```go
// Calculate pane center position
paneWidth := m.width / 2
paneHeight := m.height - 2
paneX := 0
if m.activePane == RightPane {
    paneX = paneWidth
}

// Center dialog within pane
dialogView := lipgloss.Place(
    paneWidth,
    paneHeight,
    lipgloss.Center,
    lipgloss.Center,
    m.dialog.View(),
    lipgloss.WithWhitespaceChars("█"),
    lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
)
```

### Dependencies

#### Internal Dependencies

- `internal/ui/dialog.go`: Dialog interface
- `internal/ui/model.go`: Main model and update logic
- `internal/ui/pane.go`: Pane structure and selected entry
- `internal/ui/keys.go`: Keybinding constants
- `internal/fs/operations.go`: File operations (Copy, MoveFile, Delete)
- `internal/fs/types.go`: FileEntry structure

#### External Dependencies

- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling and layout

### Key Algorithms

#### Menu Item Generation

```go
func (d *ContextMenuDialog) buildMenuItems(entry *fs.FileEntry, sourcePath, destPath string) []MenuItem {
    items := []MenuItem{}
    fullPath := filepath.Join(sourcePath, entry.Name)

    // Basic operations (always available)
    items = append(items, MenuItem{
        ID:      "copy",
        Label:   "Copy to other pane",
        Action:  func() error { return fs.Copy(fullPath, destPath) },
        Enabled: true,
    })

    items = append(items, MenuItem{
        ID:      "move",
        Label:   "Move to other pane",
        Action:  func() error { return fs.MoveFile(fullPath, destPath) },
        Enabled: true,
    })

    items = append(items, MenuItem{
        ID:      "delete",
        Label:   "Delete",
        Action:  func() error { return fs.Delete(fullPath) },
        Enabled: true,
    })

    // Symlink-specific operations
    if entry.IsSymlink && entry.IsDir && !entry.LinkBroken {
        items = append(items, MenuItem{
            ID:      "enter_logical",
            Label:   "Enter as directory (logical path)",
            Action:  func() error { /* Navigate to entry.LinkTarget */ },
            Enabled: true,
        })

        if !entry.LinkBroken {
            items = append(items, MenuItem{
                ID:      "enter_physical",
                Label:   "Open link target (physical path)",
                Action:  func() error { /* Navigate to parent of LinkTarget */ },
                Enabled: true,
            })
        }
    }

    return items
}
```

#### Page Calculation (Future)

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
```

## Test Scenarios

### Unit Tests (context_menu_dialog_test.go)

- [ ] **Test: Menu creation with regular file**
  - Given: Regular file entry
  - When: NewContextMenuDialog is called
  - Then: Menu has 3 items (copy, move, delete)

- [ ] **Test: Menu creation with symlink**
  - Given: Symlink directory entry (not broken)
  - When: NewContextMenuDialog is called
  - Then: Menu has 5 items (copy, move, delete, enter_logical, enter_physical)

- [ ] **Test: Menu creation with broken symlink**
  - Given: Broken symlink entry
  - When: NewContextMenuDialog is called
  - Then: Menu has 3 items, "enter_physical" is disabled

- [ ] **Test: Keyboard navigation - j/k**
  - Given: Menu with 3 items, cursor at position 0
  - When: 'j' key is pressed
  - Then: Cursor moves to position 1
  - When: 'k' key is pressed twice
  - Then: Cursor moves to position 0 (with boundary check)

- [ ] **Test: Keyboard navigation - numeric keys**
  - Given: Menu with 5 items
  - When: '3' key is pressed
  - Then: Item 3 is executed immediately

- [ ] **Test: Keyboard navigation - Enter**
  - Given: Menu with cursor at position 1
  - When: Enter key is pressed
  - Then: Item 1's action is executed

- [ ] **Test: Keyboard navigation - Esc**
  - Given: Active menu
  - When: Esc key is pressed
  - Then: Menu is closed, no action executed

- [ ] **Test: Page navigation (Future)**
  - Given: Menu with 15 items (2 pages)
  - When: 'l' key is pressed
  - Then: Current page changes to 1
  - When: 'h' key is pressed
  - Then: Current page changes to 0

### Integration Tests (model_test.go)

- [ ] **Test: Context menu invocation**
  - Given: Active pane with file selected
  - When: '@' key is pressed
  - Then: Context menu dialog is displayed

- [ ] **Test: Parent directory protection**
  - Given: Active pane with '..' selected
  - When: '@' key is pressed
  - Then: No dialog is displayed

- [ ] **Test: Copy action from menu**
  - Given: Context menu open with cursor on "Copy"
  - When: Enter is pressed
  - Then: File is copied to opposite pane
  - And: Opposite pane is reloaded

- [ ] **Test: Delete action from menu**
  - Given: Context menu open with cursor on "Delete"
  - When: Enter is pressed
  - Then: Confirmation dialog is shown
  - When: 'y' is pressed
  - Then: File is deleted
  - And: Current pane is reloaded

- [ ] **Test: Error handling**
  - Given: Read-only file, context menu open
  - When: "Delete" is selected
  - Then: Error dialog is shown with permission error

### Manual Testing Checklist

- [ ] Menu displays correctly centered in active pane
- [ ] All menu items are readable and properly formatted
- [ ] Selected item is clearly highlighted
- [ ] j/k navigation works smoothly
- [ ] Numeric keys 1-9 work for direct selection
- [ ] Enter executes the correct action
- [ ] Esc closes menu without action
- [ ] Menu does not appear for parent directory
- [ ] Symlink menu shows additional options
- [ ] Error dialog appears on operation failure
- [ ] Menu layout looks good on 80-column terminal
- [ ] Menu layout is acceptable on 60-column terminal (degraded mode)

## Success Criteria

### Functional Success

- [ ] `@` key displays context menu
- [ ] Menu shows "Copy", "Move", "Delete" options
- [ ] Symlinks show additional "Enter logical/physical" options
- [ ] j/k/Enter/Esc navigation works
- [ ] Numeric keys 1-9 work for direct selection
- [ ] Parent directory selection prevents menu display
- [ ] All operations execute correctly

### Quality Success

- [ ] All existing unit tests pass
- [ ] New code has 80%+ test coverage
- [ ] Menu display and operation response is instantaneous (subjectively < 100ms)
- [ ] UI design is consistent with existing dialogs

### User Experience Success

- [ ] Beginners can discover actions via `@` key without knowing shortcuts
- [ ] Menu operation is intuitive (usable without help)
- [ ] Symlink logical/physical path distinction is clear

## Open Questions

- [x] **Q1**: Should we pre-check file permissions before showing menu items?
  - **A1**: No, detect permission errors at execution time and show error dialog

- [x] **Q2**: Should menu be disabled during directory loading?
  - **A2**: Follow existing implementation (no special restrictions)

- [x] **Q3**: What should be the menu width?
  - **A3**: Adaptive width, minimum 40 columns, maximum 60 columns

- [x] **Q4**: Should we support multiple file operations (marked files)?
  - **A4**: Not in initial version; add after mark feature is implemented

## Future Considerations

### Phase 2: Additional Actions

After initial implementation, consider adding:

- **Rename**: In-place file/directory rename
- **New Directory**: Create new subdirectory
- **New File**: Create new file
- **Properties**: Show detailed file information (size, permissions, timestamps)
- **Change Permissions**: Modify file/directory permissions
- **Archive**: Create tar/zip archive
- **Extract**: Extract archive contents

### Phase 3: Mark Feature Integration

When mark feature is implemented:

- Show "X files marked" in menu header
- Batch operations on marked files
- "Mark All", "Unmark All" menu items
- "Invert Selection" menu item

### Phase 4: Keybinding Customization

Implement configuration file support:

```toml
# ~/.config/duofm/config.toml
[keybindings]
context_menu = "@"
copy = "c"
move = "m"
delete = "d"
```

Load keybindings from config:

```go
type Config struct {
    Keybindings map[string]string
}

// In keys.go
var KeyContextMenu = config.Keybindings["context_menu"] // Default: "@"
```

### Phase 5: Plugin System

Allow users to add custom menu items:

```toml
# ~/.config/duofm/plugins.toml
[[plugins]]
name = "Open in VSCode"
command = "code {file}"
key = "v"
```

## Implementation Plan

### Step 1: Create Context Menu Dialog (2-3 hours)

1. Create `internal/ui/context_menu_dialog.go`
2. Implement `ContextMenuDialog` struct
3. Implement `Dialog` interface methods
4. Implement `buildMenuItems()` logic
5. Implement `View()` with lipgloss styling

**Deliverables:**
- `context_menu_dialog.go`
- Basic rendering working

### Step 2: Add Keybinding and Integration (1-2 hours)

1. Add `KeyContextMenu = "@"` to `keys.go`
2. Add key handler in `model.go` Update method
3. Integrate with parent directory check
4. Handle context menu result messages

**Deliverables:**
- Updated `keys.go`
- Updated `model.go`
- Menu can be opened and closed

### Step 3: Implement Actions (1-2 hours)

1. Wire up copy action
2. Wire up move action
3. Wire up delete action (with confirmation dialog)
4. Implement symlink special actions
5. Add error handling with error dialog

**Deliverables:**
- All actions functional
- Error handling in place

### Step 4: Testing (2-3 hours)

1. Write unit tests for `ContextMenuDialog`
2. Write integration tests for model integration
3. Manual testing on different terminal sizes
4. Edge case testing (permissions, broken symlinks, etc.)

**Deliverables:**
- `context_menu_dialog_test.go`
- Updated `model_test.go`
- Test coverage report

### Step 5: Documentation and Polish (1 hour)

1. Update help dialog with `@` key
2. Add code comments
3. Update README if needed
4. Final manual testing

**Deliverables:**
- Updated `help_dialog.go`
- Code documentation
- Feature complete and tested

**Total Estimated Time: 7-11 hours**

## References

### Existing Implementation

- `/home/sakura/go/src/duofm/internal/ui/dialog.go` - Dialog interface
- `/home/sakura/go/src/duofm/internal/ui/confirm_dialog.go` - Confirmation dialog example
- `/home/sakura/go/src/duofm/internal/ui/help_dialog.go` - Help dialog example
- `/home/sakura/go/src/duofm/internal/ui/keys.go` - Keybinding definitions
- `/home/sakura/go/src/duofm/internal/ui/model.go` - Main model and key handling
- `/home/sakura/go/src/duofm/internal/ui/pane.go` - Pane structure
- `/home/sakura/go/src/duofm/internal/fs/operations.go` - File operations

### External References

- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [The Elm Architecture](https://guide.elm-lang.org/architecture/)
