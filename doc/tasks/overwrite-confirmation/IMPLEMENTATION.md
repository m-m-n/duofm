# Implementation Plan: Overwrite Confirmation Dialog

## Overview

This implementation plan details the steps to add an overwrite confirmation dialog for copy/move operations in duofm. When a file with the same name exists at the destination, users can choose to overwrite, cancel, or save with a different name.

## Objectives

- Add OverwriteDialog implementing the Dialog interface
- Integrate overwrite checking into copy/move operations
- Create RenameInputDialog with real-time validation
- Handle directory conflicts with error dialogs
- Maintain consistency with existing dialog patterns

## Prerequisites

- Understanding of existing Dialog interface and implementations
- Familiarity with Bubble Tea message passing
- Knowledge of existing copy/move implementation in `model.go`

## Architecture Overview

The implementation follows the existing dialog pattern:

```
Model (internal/ui/model.go)
  │
  ├── KeyCopy ('c') / KeyMove ('m')
  │     │
  │     └── checkFileConflict()
  │           │
  │           ├── No conflict → Execute operation
  │           │
  │           └── Conflict detected
  │                 │
  │                 ├── Dir vs Dir → ErrorDialog
  │                 │
  │                 └── Otherwise → OverwriteDialog
  │                                   │
  │                                   ├── Overwrite → Delete + Copy/Move
  │                                   ├── Cancel → Close
  │                                   └── Rename → RenameInputDialog
  │
  └── Context Menu (@)
        └── Same flow for copy/move actions
```

## Implementation Phases

### Phase 1: OverwriteDialog Component

**Goal**: Create the core dialog component with UI and navigation

**Files to Create/Modify**:
- `internal/ui/overwrite_dialog.go` - New dialog component
- `internal/ui/overwrite_dialog_test.go` - Unit tests

**Implementation Steps**:

1. **Define types and constants**
   ```go
   // OverwriteChoice represents user's selection
   type OverwriteChoice int

   const (
       OverwriteChoiceOverwrite OverwriteChoice = iota
       OverwriteChoiceCancel
       OverwriteChoiceRename
   )

   // OverwriteFileInfo holds metadata for display
   type OverwriteFileInfo struct {
       Size    int64
       ModTime time.Time
   }
   ```

2. **Implement OverwriteDialog struct**
   - Fields: filename, destPath, srcInfo, destInfo, cursor, active, operation, srcPath
   - Reference existing ContextMenuDialog for patterns

3. **Implement Dialog interface methods**
   - `Update(msg tea.Msg) (Dialog, tea.Cmd)` - Handle j/k/arrows, 1-3, Enter, Esc
   - `View() string` - Render dialog with file info and options
   - `IsActive() bool`
   - `DisplayType() DialogDisplayType` - Return `DialogDisplayPane`

4. **Implement helper functions**
   - `formatFileSize(bytes int64) string` - Human-readable sizes
   - `formatModTime(t time.Time) string` - Formatted date/time

5. **Create result message type**
   ```go
   type overwriteDialogResultMsg struct {
       choice    OverwriteChoice
       srcPath   string
       destPath  string
       filename  string
       operation string // "copy" or "move"
   }
   ```

**Testing**:
- Test dialog creation with correct initial state
- Test cursor navigation (j/k, up/down, wrap-around)
- Test number key selection (1, 2, 3)
- Test Enter key confirms current selection
- Test Esc key cancels
- Test file size formatting (B, KB, MB, GB)

**Estimated Effort**: Medium

---

### Phase 2: Model Integration for Copy/Move

**Goal**: Integrate overwrite checking into existing copy/move flow

**Files to Create/Modify**:
- `internal/ui/model.go` - Add conflict checking and result handling
- `internal/ui/messages.go` - New message types (if separate file exists)

**Implementation Steps**:

1. **Add conflict checking function**
   ```go
   // checkFileConflict checks if destination file exists and returns appropriate action
   func (m *Model) checkFileConflict(srcPath, destDir, operation string) tea.Cmd {
       filename := filepath.Base(srcPath)
       destPath := filepath.Join(destDir, filename)

       // Check destination
       destInfo, err := os.Stat(destPath)
       if os.IsNotExist(err) {
           // No conflict - execute immediately
           return m.executeOperation(srcPath, destDir, operation)
       }

       srcInfo, _ := os.Stat(srcPath)

       // Directory conflict
       if srcInfo.IsDir() && destInfo.IsDir() {
           return showErrorDialogCmd(fmt.Sprintf(
               "Directory \"%s\" already exists in\n%s", filename, destDir))
       }

       // File conflict - show dialog
       return showOverwriteDialogCmd(srcPath, destDir, srcInfo, destInfo, operation)
   }
   ```

2. **Modify KeyCopy handler**
   - Replace direct `fs.Copy()` call with `checkFileConflict()`
   - Handle async flow

3. **Modify KeyMove handler**
   - Replace direct `fs.MoveFile()` call with `checkFileConflict()`

4. **Add overwriteDialogResultMsg handler**
   ```go
   case overwriteDialogResultMsg:
       m.dialog = nil
       switch msg.choice {
       case OverwriteChoiceOverwrite:
           // Delete existing file, then copy/move
           destFile := filepath.Join(msg.destPath, msg.filename)
           if err := os.Remove(destFile); err != nil {
               m.dialog = NewErrorDialog(fmt.Sprintf("Failed to remove: %v", err))
               return m, nil
           }
           return m, m.executeOperation(msg.srcPath, msg.destPath, msg.operation)

       case OverwriteChoiceCancel:
           return m, nil

       case OverwriteChoiceRename:
           // Show rename dialog (Phase 3)
           m.dialog = NewRenameInputDialog(msg.destPath, msg.srcPath, msg.operation)
           return m, nil
       }
   ```

5. **Create helper command functions**
   ```go
   func showOverwriteDialogCmd(srcPath, destDir string, srcInfo, destInfo os.FileInfo, operation string) tea.Cmd
   func showErrorDialogCmd(message string) tea.Cmd
   ```

**Dependencies**:
- Phase 1 must be complete

**Testing**:
- Test copy without conflict (immediate execution)
- Test copy with file conflict (dialog shown)
- Test copy with directory conflict (error dialog)
- Test move without conflict
- Test move with file conflict
- Test overwrite action executes correctly
- Test cancel action closes dialog

**Estimated Effort**: Medium

---

### Phase 3: RenameInputDialog with Validation

**Goal**: Create input dialog for rename with real-time validation

**Files to Create/Modify**:
- `internal/ui/rename_input_dialog.go` - New specialized input dialog
- `internal/ui/rename_input_dialog_test.go` - Unit tests

**Implementation Steps**:

1. **Create RenameInputDialog struct**
   ```go
   type RenameInputDialog struct {
       title         string
       input         string
       cursorPos     int
       active        bool
       width         int
       destPath      string          // Destination directory
       srcPath       string          // Source file path
       operation     string          // "copy" or "move"
       existingFiles map[string]bool // Cached filenames in dest
       hasError      bool
       errorMessage  string
       suggestedName string          // Pre-filled suggestion
   }
   ```

2. **Implement constructor with suggested name**
   ```go
   func NewRenameInputDialog(destPath, srcPath, operation string) *RenameInputDialog {
       // Read destination directory
       existingFiles := loadExistingFiles(destPath)

       // Generate suggested name
       filename := filepath.Base(srcPath)
       suggested := suggestRename(filename, existingFiles)

       return &RenameInputDialog{
           title:         "New name:",
           input:         suggested,
           cursorPos:     len(suggested),
           active:        true,
           width:         50,
           destPath:      destPath,
           srcPath:       srcPath,
           operation:     operation,
           existingFiles: existingFiles,
           hasError:      false,
       }
   }
   ```

3. **Implement real-time validation in Update()**
   ```go
   // After any input change:
   d.validateInput()

   func (d *RenameInputDialog) validateInput() {
       if d.input == "" {
           d.hasError = true
           d.errorMessage = "File name cannot be empty"
           return
       }

       if d.existingFiles[d.input] {
           d.hasError = true
           d.errorMessage = "File already exists"
           return
       }

       if err := fs.ValidateFilename(d.input); err != nil {
           d.hasError = true
           d.errorMessage = err.Error()
           return
       }

       d.hasError = false
       d.errorMessage = ""
   }
   ```

4. **Modify Enter key handling**
   ```go
   case tea.KeyEnter:
       if d.hasError {
           return d, nil // Do nothing if error
       }
       d.active = false
       return d, func() tea.Msg {
           return renameInputResultMsg{
               newName:   d.input,
               srcPath:   d.srcPath,
               destPath:  d.destPath,
               operation: d.operation,
           }
       }
   ```

5. **Implement View() with error display**
   - Show error message in red when hasError is true
   - Remove "Enter: Confirm" hint when error exists
   - Show only "Esc: Cancel" in footer when error

6. **Implement suggestRename() helper**
   ```go
   func suggestRename(filename string, existing map[string]bool) string {
       ext := filepath.Ext(filename)
       base := strings.TrimSuffix(filename, ext)

       candidate := base + "_copy" + ext
       if !existing[candidate] {
           return candidate
       }

       for i := 2; i <= 100; i++ {
           candidate = fmt.Sprintf("%s_copy_%d%s", base, i, ext)
           if !existing[candidate] {
               return candidate
           }
       }
       return filename
   }
   ```

7. **Add result handler in model.go**
   ```go
   case renameInputResultMsg:
       m.dialog = nil
       newDestPath := filepath.Join(msg.destPath, msg.newName)

       if msg.operation == "copy" {
           if err := fs.Copy(msg.srcPath, newDestPath); err != nil {
               m.dialog = NewErrorDialog(fmt.Sprintf("Failed to copy: %v", err))
           }
       } else {
           if err := fs.MoveFile(msg.srcPath, newDestPath); err != nil {
               m.dialog = NewErrorDialog(fmt.Sprintf("Failed to move: %v", err))
           }
       }

       m.getActivePane().LoadDirectory()
       m.getInactivePane().LoadDirectory()
       return m, nil
   ```

**Dependencies**:
- Phase 1 and 2 must be complete
- Reference existing InputDialog implementation

**Testing**:
- Test suggested name generation
- Test validation: empty input shows error
- Test validation: existing file shows error
- Test validation: invalid filename shows error
- Test Enter is disabled when error exists
- Test successful rename executes operation
- Test cursor navigation and editing

**Estimated Effort**: Medium

---

### Phase 4: Context Menu Integration

**Goal**: Apply overwrite checking to context menu copy/move actions

**Files to Create/Modify**:
- `internal/ui/context_menu_dialog.go` - Modify action handlers
- `internal/ui/model.go` - Handle context menu results with conflict check

**Implementation Steps**:

1. **Modify context menu result handling**

   Current flow:
   ```go
   case contextMenuResultMsg:
       if result.action != nil {
           // Direct execution
           if err := result.action(); err != nil { ... }
       }
   ```

   New flow:
   ```go
   case contextMenuResultMsg:
       if result.actionID == "copy" || result.actionID == "move" {
           // Use conflict checking flow
           entry := m.getActivePane().SelectedEntry()
           srcPath := filepath.Join(m.getActivePane().Path(), entry.Name)
           destPath := m.getInactivePane().Path()
           return m, m.checkFileConflict(srcPath, destPath, result.actionID)
       }
       // Other actions (delete, enter_logical, etc.) remain unchanged
       if result.action != nil {
           if err := result.action(); err != nil { ... }
       }
   ```

2. **Ensure consistent behavior**
   - Context menu copy should show same dialog as 'c' key
   - Context menu move should show same dialog as 'm' key

**Dependencies**:
- Phase 2 must be complete

**Testing**:
- Test context menu copy with conflict
- Test context menu move with conflict
- Test context menu copy without conflict
- Test context menu delete still works (no change)

**Estimated Effort**: Small

---

### Phase 5: Edge Cases and Polish

**Goal**: Handle edge cases and improve user experience

**Files to Create/Modify**:
- `internal/ui/overwrite_dialog.go` - Edge case handling
- `internal/ui/model.go` - Error handling improvements

**Implementation Steps**:

1. **Handle symlink edge cases**
   - Use `os.Lstat()` for checking symlink existence
   - Treat symlinks same as files for confirmation

2. **Handle permission errors**
   ```go
   // In overwrite action
   destFile := filepath.Join(msg.destPath, msg.filename)
   if err := os.Remove(destFile); err != nil {
       if os.IsPermission(err) {
           m.dialog = NewErrorDialog("Permission denied: cannot remove existing file")
       } else {
           m.dialog = NewErrorDialog(fmt.Sprintf("Failed to remove: %v", err))
       }
       return m, nil
   }
   ```

3. **Improve file info display**
   - Handle very large files gracefully
   - Handle files with unknown modification time
   - Truncate long paths in display

4. **Visual polish**
   - Ensure consistent spacing
   - Test on various terminal widths
   - Verify colors work on light/dark themes

**Dependencies**:
- All previous phases complete

**Testing**:
- Test with read-only destination files
- Test with symlinks
- Test with very long filenames
- Test on narrow terminals (60 columns)
- Test on wide terminals (120+ columns)

**Estimated Effort**: Small

---

## File Structure

```
internal/ui/
├── dialog.go                    # Dialog interface (no change)
├── model.go                     # Main model - add conflict handling
├── overwrite_dialog.go          # NEW: Overwrite confirmation dialog
├── overwrite_dialog_test.go     # NEW: Unit tests
├── rename_input_dialog.go       # NEW: Rename input with validation
├── rename_input_dialog_test.go  # NEW: Unit tests
├── context_menu_dialog.go       # Modify result handling
├── input_dialog.go              # Reference only (no change)
├── confirm_dialog.go            # Reference only (no change)
└── error_dialog.go              # Reference only (no change)

internal/fs/
├── operations.go                # Copy, MoveFile, Delete (no change)
└── fileinfo.go                  # NEW (optional): File info utilities
```

## Testing Strategy

### Unit Tests

**overwrite_dialog_test.go**:
- TestNewOverwriteDialog - Creation with correct initial state
- TestOverwriteDialogNavigation - j/k/arrow cursor movement
- TestOverwriteDialogNumberKeys - 1/2/3 direct selection
- TestOverwriteDialogEnter - Enter confirms selection
- TestOverwriteDialogEsc - Esc cancels
- TestFormatFileSize - Human-readable size formatting
- TestOverwriteDialogView - Rendering verification

**rename_input_dialog_test.go**:
- TestNewRenameInputDialog - Creation with suggested name
- TestRenameInputDialogValidation - Real-time validation
- TestRenameInputDialogExistingFile - Error when file exists
- TestRenameInputDialogEmptyInput - Error when empty
- TestRenameInputDialogEnterDisabled - Enter disabled on error
- TestSuggestRename - Name generation algorithm

### Integration Tests

Add to existing test files:
- Test copy operation with file conflict
- Test move operation with file conflict
- Test directory conflict handling
- Test context menu integration
- Test full rename flow

### E2E Tests

Leverage existing Docker-based E2E testing infrastructure:
- Test copy operation with file conflict (dialog appears, overwrite works)
- Test move operation with file conflict
- Test rename flow end-to-end
- Test directory conflict shows error dialog
- Test context menu copy/move with conflict

### Manual Testing Checklist

- [ ] Copy file with same name - dialog appears
- [ ] Move file with same name - dialog appears
- [ ] Overwrite selection replaces file
- [ ] Cancel selection closes without action
- [ ] Rename selection opens input dialog
- [ ] Invalid rename shows error
- [ ] Valid rename copies with new name
- [ ] Directory conflict shows error
- [ ] Context menu copy with conflict works
- [ ] Context menu move with conflict works
- [ ] Navigation with j/k works
- [ ] Navigation with arrow keys works
- [ ] Number keys 1-3 select options
- [ ] File info (size, date) displays correctly
- [ ] Dialog centers in pane correctly

## Dependencies

### External Libraries

- `github.com/charmbracelet/bubbletea` - TUI framework (existing)
- `github.com/charmbracelet/lipgloss` - Styling (existing)

### Internal Dependencies

Implementation order (by dependency):
1. Phase 1: OverwriteDialog (no dependencies)
2. Phase 2: Model integration (depends on Phase 1)
3. Phase 3: RenameInputDialog (depends on Phase 2)
4. Phase 4: Context menu (depends on Phase 2)
5. Phase 5: Polish (depends on all)

## Risk Assessment

### Technical Risks

- **Race conditions in file operations**
  - Risk: File state changes between check and operation
  - Mitigation: Keep dialog open during operation, handle errors gracefully

- **Large file metadata fetching**
  - Risk: Slow response for network drives
  - Mitigation: Set timeout, show loading indicator if needed

### Implementation Risks

- **Breaking existing copy/move behavior**
  - Risk: Regression in no-conflict case
  - Mitigation: Comprehensive testing, keep existing path fast

- **InputDialog code duplication**
  - Risk: Duplicating InputDialog logic in RenameInputDialog
  - Mitigation: Consider embedding or extracting common base

## Performance Considerations

- File existence check uses `os.Stat()` - typically < 1ms
- Destination directory listing for validation - cache on dialog creation
- Dialog rendering is pure string manipulation - negligible overhead

## Security Considerations

- Validate filenames to prevent path traversal attacks
- Use `fs.ValidateFilename()` for input validation
- Never execute shell commands with user input
- Handle permission errors gracefully without exposing sensitive info

## Open Questions

All questions resolved in specification phase.

## Future Enhancements

- **Phase 2 (Multiple files)**: "Apply to all" option for batch operations
- **Phase 3 (Advanced)**: File comparison, auto-rename, skip option
- **Configuration**: Option to disable confirmation for power users

## References

- [SPEC.md](./SPEC.md) - Feature specification
- [要件定義書.md](./要件定義書.md) - Japanese requirements document
- [Dialog interface](../../internal/ui/dialog.go)
- [InputDialog](../../internal/ui/input_dialog.go) - Reference implementation
- [ContextMenuDialog](../../internal/ui/context_menu_dialog.go) - Reference implementation
