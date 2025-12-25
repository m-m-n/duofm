# Feature: Multi-file Marking

## Overview

This feature adds the ability to mark (select) multiple files and perform batch file operations (copy, move, delete) on them. Users can mark individual files with the Space key and see visual feedback indicating which files are selected.

## Objectives

- Enable efficient batch operations on multiple files
- Provide clear visual feedback for marked files
- Maintain consistency with existing single-file operations

## User Stories

- As a user, I want to mark files one by one with Space, so that I can select multiple files for batch operations
- As a user, I want to see which files are marked, so that I can verify my selection before performing operations
- As a user, I want to copy/move/delete all marked files at once, so that I can work more efficiently
- As a user, I want to see the count and total size of marked files in the header, so that I can understand my selection

## Technical Requirements

### Functional Requirements

#### FR1: Mark Operations

- FR1.1: Pressing Space on an unmarked file marks it
- FR1.2: Pressing Space on a marked file removes the mark
- FR1.3: After marking or unmarking, cursor moves down by one position
- FR1.4: Parent directory (..) cannot be marked (Space does nothing)
- FR1.5: Mark state is managed independently per pane
- FR1.6: Marks are cleared when changing directories

#### FR2: Visual Display

- FR2.1: Marked files are displayed with a different background color
- FR2.2: Marked files on cursor position are visually distinguishable
- FR2.3: Active and inactive panes use different mark colors

#### FR3: Header Display

- FR3.1: Display mark info as "Marked X/Y Z MiB"
- FR3.2: X = marked count, Y = total file count (excluding parent dir)
- FR3.3: Z = total size of marked files
- FR3.4: Directories are counted as 0 bytes for size calculation

#### FR4: File Operation Integration

- FR4.1: When marks exist, c/m/d operations apply to marked files
- FR4.2: When no marks, c/m/d operations apply to cursor position (existing behavior)
- FR4.3: Marks are cleared after operation completes
- FR4.4: Handle overwrite confirmation for multiple files

#### FR5: Multi-file Overwrite Confirmation

- FR5.1: Use existing overwrite confirmation dialog
- FR5.2: Confirm per file
- FR5.3: Cancel aborts remaining files

### Non-functional Requirements

#### NFR1: Performance

- NFR1.1: Mark operation completes within 50ms
- NFR1.2: No rendering delay with 1000+ marked files
- NFR1.3: Mark info calculation is efficient

#### NFR2: Consistency

- NFR2.1: No conflicts with existing key bindings
- NFR2.2: Consistent with other dialog-based features
- NFR2.3: Context menu operations support marked files

## Implementation Approach

### Architecture

```
Pane (internal/ui/pane.go)
  ├── markedFiles map[string]bool    // Mark state per file
  ├── ToggleMark()                   // Toggle current file
  ├── ClearMarks()                   // Clear all marks
  ├── CalculateMarkInfo()            // Get count and size
  └── View() with mark highlighting

Model (internal/ui/model.go)
  ├── handleSpaceKey()               // Mark toggle handler
  └── executeFileOperationBatch()    // Batch operation handler
```

### Component Design

#### 1. Pane Mark Management

```go
// Add to Pane struct
type Pane struct {
    // ... existing fields
    markedFiles map[string]bool  // key: filename, value: marked state
}

// MarkInfo holds mark statistics
type MarkInfo struct {
    Count     int   // Number of marked files
    TotalSize int64 // Total size in bytes
}

// ToggleMark toggles mark on current file
// Returns false if current entry is parent directory
func (p *Pane) ToggleMark() bool

// ClearMarks removes all marks
func (p *Pane) ClearMarks()

// CalculateMarkInfo returns mark statistics
func (p *Pane) CalculateMarkInfo() MarkInfo

// IsMarked returns whether a file is marked
func (p *Pane) IsMarked(filename string) bool

// GetMarkedFiles returns list of marked filenames
func (p *Pane) GetMarkedFiles() []string
```

#### 2. Visual Styling

Background colors for marked files:

| State | Active Pane | Inactive Pane |
|-------|-------------|---------------|
| Normal | Default | Default |
| Cursor | Color 39 (Blue) | Color 240 (Gray) |
| Marked | Color 136 (Yellow) | Color 94 (Dark Yellow) |
| Cursor + Marked | Color 30 (Cyan) | Color 23 (Dark Cyan) |

```go
// Mark colors
var (
    markBgColorActive   = lipgloss.Color("136") // Yellow
    markBgColorInactive = lipgloss.Color("94")  // Dark yellow

    cursorMarkBgColorActive   = lipgloss.Color("30") // Cyan
    cursorMarkBgColorInactive = lipgloss.Color("23") // Dark cyan
)
```

#### 3. Header Display Update

Update `renderHeaderLine2()` to show actual mark info:

```go
func (p *Pane) renderHeaderLine2(diskSpace uint64) string {
    markInfo := p.CalculateMarkInfo()

    var markedInfo string
    if p.IsFiltered() {
        filteredCount := p.FilteredEntryCount()
        totalCount := p.TotalEntryCount()
        markedInfo = fmt.Sprintf("Marked %d/%d (%d) %s",
            markInfo.Count, filteredCount, totalCount,
            FormatSize(markInfo.TotalSize))
    } else {
        totalCount := p.TotalEntryCount()
        markedInfo = fmt.Sprintf("Marked %d/%d %s",
            markInfo.Count, totalCount,
            FormatSize(markInfo.TotalSize))
    }

    // ... rest of the function
}
```

#### 4. Key Handling

Add Space key constant and handler:

```go
// In keys.go
const KeyMark = " " // Space key

// In model.go Update()
case KeyMark:
    return m.handleMarkToggle()
```

#### 5. Batch File Operations

Modify copy/move/delete handlers:

```go
func (m *Model) handleCopy() tea.Cmd {
    activePane := m.ActivePane()
    markedFiles := activePane.GetMarkedFiles()

    if len(markedFiles) > 0 {
        // Batch operation on marked files
        return m.executeBatchCopy(markedFiles)
    }

    // Single file operation (existing behavior)
    return m.executeSingleCopy()
}
```

### Data Flow

```
User presses Space
    │
    ├── On parent directory (..) → Do nothing
    │
    └── On regular file/directory
        │
        ├── If unmarked → Add to markedFiles map
        │
        ├── If marked → Remove from markedFiles map
        │
        ├── Move cursor down (if not at last entry)
        │
        └── Re-render pane with updated mark highlighting

User presses 'c' (copy)
    │
    ├── Check for marked files
    │   │
    │   ├── Has marks → Batch copy all marked files
    │   │   │
    │   │   └── For each file:
    │   │       ├── Check for conflicts
    │   │       ├── Show overwrite dialog if needed
    │   │       ├── Execute copy
    │   │       └── Continue to next file (or abort on cancel)
    │   │
    │   └── No marks → Single file copy (existing behavior)
    │
    └── Clear marks after completion

Directory change
    │
    └── Clear all marks
```

### UI/UX Design

#### Marked File Display

```
Before marking:
  README.md          1.2 KiB  2024-12-01
  notes.txt          450 B    2024-11-28
  image.png          2.3 MiB  2024-12-10

After marking (background color highlight):
  README.md          1.2 KiB  2024-12-01    ← Yellow background
  notes.txt          450 B    2024-11-28
  image.png          2.3 MiB  2024-12-10    ← Yellow background
```

#### Header Examples

```
No marks:
  Marked 0/15 0 B                         10.5 GiB Free

3 files marked (total 1.5 MiB):
  Marked 3/15 1.5 MiB                     10.5 GiB Free

With filter (5 shown, 2 marked):
  Marked 2/5 (15) 500 KiB                 10.5 GiB Free
```

### Dependencies

#### Internal Dependencies

- `internal/ui/pane.go`: Mark state management, display
- `internal/ui/model.go`: Key handling, file operations
- `internal/ui/keys.go`: Space key constant
- `internal/fs/operations.go`: Copy, MoveFile, Delete functions
- `internal/ui/overwrite_dialog.go`: Conflict handling

#### External Dependencies

- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling

### Edge Cases

| Case | Behavior |
|------|----------|
| Space on parent dir (..) | Ignored (no mark, no cursor move) |
| Space on last file | Mark toggles, cursor stays at last position |
| Toggle hidden files | Marks on hidden files are cleared when hidden |
| Apply filter | Marks preserved (including non-visible files) |
| Directory change | All marks cleared |
| Refresh (F5) | Marks preserved (if files still exist) |
| File deleted externally | Mark auto-cleared on refresh |
| Symlink | Treated same as regular file/directory |

## Test Scenarios

### Unit Tests (pane_mark_test.go)

- [ ] **Test: Mark unmarked file**
  - Given: Cursor on unmarked file
  - When: Space is pressed
  - Then: File is marked

- [ ] **Test: Unmark marked file**
  - Given: Cursor on marked file
  - When: Space is pressed
  - Then: Mark is removed

- [ ] **Test: Space on parent directory**
  - Given: Cursor on parent directory (..)
  - When: Space is pressed
  - Then: No mark, cursor does not move

- [ ] **Test: Cursor moves after mark**
  - Given: Cursor on file (not last)
  - When: Space is pressed
  - Then: Cursor moves down by one

- [ ] **Test: Cursor stays on last file**
  - Given: Cursor on last file
  - When: Space is pressed
  - Then: File is marked, cursor stays at last position

- [ ] **Test: ClearMarks**
  - Given: Multiple files are marked
  - When: ClearMarks is called
  - Then: All marks are cleared

- [ ] **Test: CalculateMarkInfo count**
  - Given: 3 files marked
  - When: CalculateMarkInfo is called
  - Then: Count is 3

- [ ] **Test: CalculateMarkInfo size**
  - Given: Files of 100, 200, 300 bytes marked
  - When: CalculateMarkInfo is called
  - Then: TotalSize is 600

- [ ] **Test: CalculateMarkInfo with directory**
  - Given: Directory and file marked
  - When: CalculateMarkInfo is called
  - Then: Directory counted as 0 bytes

- [ ] **Test: GetMarkedFiles returns correct list**
  - Given: Files A, B, C marked
  - When: GetMarkedFiles is called
  - Then: Returns [A, B, C]

- [ ] **Test: IsMarked returns correct state**
  - Given: File A marked, File B not marked
  - When: IsMarked is called
  - Then: Returns true for A, false for B

### Integration Tests

- [ ] **Test: Space key toggles mark**
  - Given: Model in normal state
  - When: Space key is pressed on file
  - Then: File is marked, cursor moves down

- [ ] **Test: Marks cleared on directory change**
  - Given: Files marked in current directory
  - When: Enter directory
  - Then: All marks are cleared

- [ ] **Test: Batch copy with marks**
  - Given: 2 files marked
  - When: 'c' key pressed
  - Then: Both files are copied to opposite pane

- [ ] **Test: Batch delete with marks**
  - Given: 2 files marked
  - When: 'd' key pressed, confirm
  - Then: Both files are deleted, marks cleared

- [ ] **Test: Single operation when no marks**
  - Given: No files marked
  - When: 'c' key pressed
  - Then: Only cursor file is copied

- [ ] **Test: Overwrite confirmation per file**
  - Given: 2 files marked, 1 conflicts
  - When: 'c' key pressed
  - Then: Overwrite dialog shown for conflicting file

- [ ] **Test: Cancel aborts batch**
  - Given: 3 files marked, 2nd file conflicts
  - When: Cancel in overwrite dialog
  - Then: 1st file copied, 2nd and 3rd skipped

### Visual Tests (Manual)

- [ ] Marked files show yellow background in active pane
- [ ] Marked files show dark yellow in inactive pane
- [ ] Cursor on marked file shows cyan background
- [ ] Header shows correct mark count and size
- [ ] Mark indicator updates immediately on Space

## Success Criteria

### Functional Success

- [ ] Space marks unmarked files
- [ ] Space unmarks marked files
- [ ] Cursor moves down after marking (except on last file)
- [ ] Parent directory cannot be marked
- [ ] Marked files visually highlighted with background color
- [ ] Header shows mark count and total size
- [ ] c/m/d apply to marked files when marks exist
- [ ] c/m/d apply to cursor file when no marks
- [ ] Marks cleared after operations
- [ ] Marks cleared on directory change
- [ ] Overwrite confirmation works for batch operations

### Quality Success

- [ ] All existing tests pass
- [ ] New code has 80%+ test coverage
- [ ] Mark toggle is immediate (< 50ms)
- [ ] No performance degradation with many marks

### User Experience Success

- [ ] Mark state is immediately visible
- [ ] Header info helps user understand selection
- [ ] Batch operations feel intuitive
- [ ] No confusion with single-file operations

## Implementation Plan

### Step 1: Pane Mark Management (1-2 hours)

1. Add `markedFiles` map to Pane struct
2. Implement `ToggleMark()`, `ClearMarks()`, `IsMarked()`
3. Implement `CalculateMarkInfo()`
4. Implement `GetMarkedFiles()`
5. Add unit tests

**Deliverables:**
- Updated `pane.go`
- New `pane_mark_test.go`

### Step 2: Visual Display (1-2 hours)

1. Define mark color constants
2. Update `formatEntry()` to apply mark styling
3. Update `formatEntryDimmed()` for dialog overlay
4. Update `renderHeaderLine2()` with real mark info
5. Manual visual testing

**Deliverables:**
- Updated `pane.go` with styling

### Step 3: Key Handling (1 hour)

1. Add `KeyMark` constant to keys.go
2. Add Space key handler in model.go
3. Clear marks on directory change
4. Add integration tests

**Deliverables:**
- Updated `keys.go`
- Updated `model.go`
- Integration tests

### Step 4: Batch Operations (2-3 hours)

1. Modify `handleCopy()` for batch operations
2. Modify `handleMove()` for batch operations
3. Modify `handleDelete()` for batch operations
4. Handle overwrite confirmation for multiple files
5. Clear marks after operation
6. Add integration tests

**Deliverables:**
- Updated `model.go`
- Integration tests

### Step 5: Context Menu Integration (1 hour)

1. Update context menu copy/move/delete actions
2. Ensure marked files are used when available
3. Test all paths

**Deliverables:**
- Updated `context_menu_dialog.go`

### Step 6: Testing and Polish (1-2 hours)

1. Manual testing on various scenarios
2. Edge case testing
3. Performance testing with many files
4. Code documentation

**Deliverables:**
- Complete test coverage
- Documentation

## References

### Existing Implementation

- `/home/sakura/go/src/duofm/internal/ui/pane.go` - Pane component
- `/home/sakura/go/src/duofm/internal/ui/model.go` - Main model
- `/home/sakura/go/src/duofm/internal/ui/keys.go` - Key constants
- `/home/sakura/go/src/duofm/internal/fs/operations.go` - File operations
- `/home/sakura/go/src/duofm/internal/ui/overwrite_dialog.go` - Overwrite dialog

### External References

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
- [Midnight Commander](https://midnight-commander.org/) - Reference for mark behavior
