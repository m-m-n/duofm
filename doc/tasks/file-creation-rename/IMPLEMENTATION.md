# Implementation Plan: File and Directory Creation and Renaming

## Overview

This implementation adds three file operations to duofm:
- New file creation (`n` key)
- New directory creation (`Shift+n` key)
- Rename (`r` key)

All operations use a text input dialog with consistent UI patterns following existing dialog implementations.

## Objectives

- Implement InputDialog component for text input
- Add file system operations: CreateFile, CreateDirectory, Rename
- Integrate with existing model key handling
- Handle cursor positioning after operations
- Implement proper error handling with status bar messages

## Prerequisites

- Existing dialog infrastructure (ConfirmDialog, ErrorDialog patterns)
- Existing Minibuffer text editing implementation
- Existing fs.operations package

## Architecture Overview

The implementation follows existing patterns:
1. **InputDialog**: New dialog type for text input (similar to ConfirmDialog but with editable text field)
2. **File Operations**: Add CreateFile, CreateDirectory, Rename to `internal/fs/operations.go`
3. **Key Handling**: Add handlers for `n`, `N`, `r` keys in model.go
4. **Cursor Management**: Add helper methods for cursor positioning after operations

## Implementation Phases

### Phase 1: File System Operations

**Goal**: Add low-level file operations to the fs package

**Files to Create/Modify**:
- `internal/fs/operations.go` - Add CreateFile, CreateDirectory, Rename functions
- `internal/fs/operations_test.go` - Add tests for new operations

**Implementation Steps**:

1. Add `CreateFile` function
   ```go
   func CreateFile(path string) error {
       // Check if file already exists
       if _, err := os.Stat(path); err == nil {
           return fmt.Errorf("file already exists: %s", filepath.Base(path))
       }
       file, err := os.Create(path)
       if err != nil {
           return fmt.Errorf("failed to create file: %w", err)
       }
       return file.Close()
   }
   ```

2. Add `CreateDirectory` function
   ```go
   func CreateDirectory(path string) error {
       // Check if directory already exists
       if _, err := os.Stat(path); err == nil {
           return fmt.Errorf("directory already exists: %s", filepath.Base(path))
       }
       if err := os.Mkdir(path, 0755); err != nil {
           return fmt.Errorf("failed to create directory: %w", err)
       }
       return nil
   }
   ```

3. Add `Rename` function
   ```go
   func Rename(oldPath, newName string) error {
       dir := filepath.Dir(oldPath)
       newPath := filepath.Join(dir, newName)

       // Check if target already exists
       if _, err := os.Stat(newPath); err == nil {
           return fmt.Errorf("file already exists: %s", newName)
       }

       if err := os.Rename(oldPath, newPath); err != nil {
           return fmt.Errorf("failed to rename: %w", err)
       }
       return nil
   }
   ```

4. Add input validation helper
   ```go
   func ValidateFilename(name string) error {
       if name == "" {
           return fmt.Errorf("file name cannot be empty")
       }
       if strings.Contains(name, "/") {
           return fmt.Errorf("invalid file name: path separator not allowed")
       }
       return nil
   }
   ```

**Testing**:
- Unit tests for CreateFile with various scenarios
- Unit tests for CreateDirectory with various scenarios
- Unit tests for Rename with various scenarios
- Tests for validation function

**Estimated Effort**: Small

---

### Phase 2: Input Dialog Implementation

**Goal**: Create InputDialog component with text editing capabilities

**Files to Create/Modify**:
- `internal/ui/input_dialog.go` - NEW: Input dialog implementation
- `internal/ui/input_dialog_test.go` - NEW: Tests for input dialog

**Implementation Steps**:

1. Define InputDialog struct
   ```go
   type InputDialog struct {
       title       string        // Dialog title/prompt
       input       string        // Current input text
       cursorPos   int           // Cursor position in input
       active      bool          // Dialog is active
       width       int           // Dialog width
       onConfirm   func(string) tea.Cmd  // Callback on Enter
       errorMsg    string        // Validation error message
   }
   ```

2. Implement NewInputDialog constructor
   - Accept title and onConfirm callback
   - Initialize with empty input

3. Implement Update method
   - Reuse Minibuffer text editing logic for key handling:
     - Character input
     - Backspace/Delete
     - Left/Right arrow movement
     - Ctrl+A (beginning), Ctrl+E (end)
     - Ctrl+U (delete to beginning), Ctrl+K (delete to end)
   - Handle Enter key:
     - Validate input (not empty, no path separator)
     - If validation fails, set errorMsg and keep dialog open
     - If validation passes, call onConfirm and return result message
   - Handle Esc key:
     - Close dialog and return cancelled result

4. Implement View method
   - Layout matching specification:
     ```
     ┌─────────────────────────────┐
     │ New file:                   │
     │ ┌─────────────────────────┐ │
     │ │ [input field]_          │ │
     │ └─────────────────────────┘ │
     │                             │
     │ Enter: Create  Esc: Cancel  │
     └─────────────────────────────┘
     ```
   - Show error message if validation failed
   - Display cursor with reverse video style

5. Implement IsActive and DisplayType methods
   - DisplayType returns DialogDisplayPane

**Testing**:
- Test dialog creation
- Test text input handling
- Test cursor movement
- Test Enter/Esc handling
- Test validation error display

**Estimated Effort**: Medium

---

### Phase 3: Key Bindings and Constants

**Goal**: Add key constants for new operations

**Files to Create/Modify**:
- `internal/ui/keys.go` - Add new key constants

**Implementation Steps**:

1. Add key constants
   ```go
   const (
       // Existing keys...

       // File creation and renaming
       KeyNewFile      = "n"      // Create new file
       KeyNewDirectory = "N"      // Create new directory (Shift+n)
       KeyRename       = "r"      // Rename file/directory
   )
   ```

**Testing**:
- No specific tests needed (compile-time verification)

**Estimated Effort**: Small

---

### Phase 4: Model Integration

**Goal**: Integrate InputDialog with Model and handle operations

**Files to Create/Modify**:
- `internal/ui/model.go` - Add key handlers and operation logic
- `internal/ui/messages.go` - Add new message types if needed

**Implementation Steps**:

1. Add inputDialogResultMsg type
   ```go
   type inputDialogResultMsg struct {
       confirmed bool
       input     string
       operation string // "create_file", "create_dir", "rename"
   }
   ```

2. Add key handler for `n` (new file)
   ```go
   case KeyNewFile:
       pane := m.getActivePane()
       m.dialog = NewInputDialog("New file:", func(filename string) tea.Cmd {
           // Validation is done in dialog
           fullPath := filepath.Join(pane.Path(), filename)
           if err := fs.CreateFile(fullPath); err != nil {
               return func() tea.Msg {
                   return inputDialogResultMsg{
                       confirmed: false,
                       operation: "create_file",
                       error:     err,
                   }
               }
           }
           return func() tea.Msg {
               return inputDialogResultMsg{
                   confirmed: true,
                   input:     filename,
                   operation: "create_file",
               }
           }
       })
       return m, nil
   ```

3. Add key handler for `N` (new directory)
   - Similar to new file handler with CreateDirectory

4. Add key handler for `r` (rename)
   - Check if parent directory (..) is selected - if so, ignore
   - Get current entry name for reference
   - Create dialog with callback

5. Handle inputDialogResultMsg
   - On success:
     - Reload both panes
     - Move cursor to created/renamed item
   - On failure:
     - Show error in status bar

**Testing**:
- Integration tests for key handling
- Tests for dialog creation and result handling

**Estimated Effort**: Medium

---

### Phase 5: Cursor Position Control

**Goal**: Implement cursor positioning after create/rename operations

**Files to Create/Modify**:
- `internal/ui/pane.go` - Add cursor movement helpers
- `internal/ui/model.go` - Use helpers after operations

**Implementation Steps**:

1. Add `MoveCursorToEntry` method to Pane
   ```go
   func (p *Pane) MoveCursorToEntry(name string) bool {
       for i, entry := range p.entries {
           if entry.Name == name {
               p.cursor = i
               p.adjustScroll()
               return true
           }
       }
       return false
   }
   ```

2. Add cursor handling after create operation
   ```go
   func (m *Model) handleCreateSuccess(filename string) {
       pane := m.getActivePane()

       // Reload both panes
       pane.LoadDirectory()
       m.getInactivePane().LoadDirectory()

       // Check if created file is visible
       if strings.HasPrefix(filename, ".") && !pane.showHidden {
           // Dot file created with hidden files OFF - stay at original position
           return
       }

       // Move cursor to created file
       pane.MoveCursorToEntry(filename)
   }
   ```

3. Add cursor handling after rename operation
   ```go
   func (m *Model) handleRenameSuccess(oldName, newName string) {
       pane := m.getActivePane()
       savedCursor := pane.cursor

       // Reload both panes
       pane.LoadDirectory()
       m.getInactivePane().LoadDirectory()

       // Check if renamed to hidden file with hidden files OFF
       if strings.HasPrefix(newName, ".") && !pane.showHidden {
           // File is now hidden - move to next/previous
           if savedCursor >= len(pane.entries) {
               if len(pane.entries) > 0 {
                   pane.cursor = len(pane.entries) - 1
               }
           }
           pane.adjustScroll()
           return
       }

       // Move cursor to renamed file
       pane.MoveCursorToEntry(newName)
   }
   ```

**Testing**:
- Test cursor movement to normal files
- Test cursor behavior with hidden files
- Test cursor behavior when renaming to hidden file

**Estimated Effort**: Small

---

### Phase 6: Error Handling and Status Messages

**Goal**: Implement proper error display in status bar

**Files to Create/Modify**:
- `internal/ui/model.go` - Error handling integration

**Implementation Steps**:

1. Use existing status message infrastructure
   - `m.statusMessage` for displaying errors
   - `m.isStatusError = true` for error styling
   - `statusMessageClearCmd(5 * time.Second)` for auto-clear

2. Error message formats (as per spec):
   - Empty string: `File name cannot be empty` (dialog stays open)
   - File exists: `File already exists: [filename]`
   - Path separator: `Invalid file name: path separator not allowed`
   - Permission denied: `Permission denied: [error details]`
   - Other errors: `Failed to create/rename: [error details]`

3. Special handling for empty string error
   - Keep dialog open
   - Show error message in dialog itself

**Testing**:
- Test error message display
- Test auto-clear after 5 seconds
- Test empty string keeps dialog open

**Estimated Effort**: Small

---

## File Structure

```
internal/
├── ui/
│   ├── dialog.go           # (existing) Dialog interface
│   ├── input_dialog.go     # NEW: Input dialog implementation
│   ├── input_dialog_test.go # NEW: Tests for input dialog
│   ├── model.go            # Add n, N, r key handlers
│   ├── keys.go             # Add new key constants
│   └── pane.go             # Add cursor movement helpers
├── fs/
│   ├── operations.go       # Add CreateFile, CreateDirectory, Rename
│   └── operations_test.go  # Add tests for new operations
```

## Testing Strategy

### Unit Tests

**fs/operations_test.go**:
- `TestCreateFile_Success`
- `TestCreateFile_AlreadyExists`
- `TestCreateFile_PermissionDenied`
- `TestCreateDirectory_Success`
- `TestCreateDirectory_AlreadyExists`
- `TestRename_Success`
- `TestRename_TargetExists`
- `TestValidateFilename_Valid`
- `TestValidateFilename_Empty`
- `TestValidateFilename_PathSeparator`

**ui/input_dialog_test.go**:
- `TestInputDialog_New`
- `TestInputDialog_CharacterInput`
- `TestInputDialog_CursorMovement`
- `TestInputDialog_Backspace`
- `TestInputDialog_Delete`
- `TestInputDialog_CtrlA_CtrlE`
- `TestInputDialog_CtrlU_CtrlK`
- `TestInputDialog_EnterConfirm`
- `TestInputDialog_EscCancel`
- `TestInputDialog_EmptyInputError`

### Integration Tests

- Test full flow: press `n` -> type filename -> Enter -> file created
- Test full flow: press `N` -> type dirname -> Enter -> directory created
- Test full flow: press `r` -> type new name -> Enter -> file renamed

### Manual Testing Checklist

- [ ] `n` key opens "New file:" dialog
- [ ] `Shift+n` key opens "New directory:" dialog
- [ ] `r` key opens "Rename to:" dialog (normal file)
- [ ] `r` key on parent directory (..) does nothing
- [ ] Enter confirms and executes operation
- [ ] Esc cancels and closes dialog
- [ ] Empty input shows error but keeps dialog open
- [ ] Path separator shows error and closes dialog
- [ ] Existing file shows error and closes dialog
- [ ] Created file appears and cursor moves to it
- [ ] Created dot file (hidden OFF): file created, cursor stays
- [ ] Renamed file: cursor follows to new name
- [ ] Renamed to dot file (hidden OFF): cursor moves to next
- [ ] Both panes reload after operation

## Dependencies

### External Libraries
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- Go standard library (`os`, `path/filepath`, `strings`)

### Internal Dependencies
- Phase 1 (fs operations) - No dependencies
- Phase 2 (input dialog) - No dependencies
- Phase 3 (key constants) - No dependencies
- Phase 4 (model integration) - Depends on Phase 1, 2, 3
- Phase 5 (cursor control) - Depends on Phase 4
- Phase 6 (error handling) - Depends on Phase 4

## Risk Assessment

### Technical Risks

- **Text editing in dialog**: The Minibuffer already implements all required text editing; we can reuse that logic
  - Mitigation: Extract text editing logic into shared helper or inline copy

- **Cursor position after reload**: LoadDirectory resets cursor to 0
  - Mitigation: Implement dedicated methods that preserve/find cursor position

### Implementation Risks

- **Dialog focus handling**: Need to ensure InputDialog properly captures input
  - Mitigation: Follow existing ConfirmDialog pattern exactly

- **Hidden file edge cases**: Complex cursor behavior when hidden files are toggled
  - Mitigation: Clearly defined spec; implement with explicit conditionals

## Performance Considerations

- File operations are synchronous but fast (single file/directory)
- Both panes reload after operations (necessary for same-directory case)
- No performance concerns expected

## Security Considerations

- Input validation prevents path traversal (no `/` in filenames)
- File operations use standard Go os package (follows system permissions)
- No arbitrary command execution

## Open Questions

None - all requirements are clearly specified in SPEC.md.

## Future Enhancements

Features explicitly out of scope (may be added later):
- Batch file creation
- File creation from templates
- Rename with move (to different directory)
- Showing current filename in rename dialog
- Filename auto-completion
- Undo functionality for create/rename

## References

- [SPEC.md](./SPEC.md) - Feature specification
- [要件定義書.md](./要件定義書.md) - Japanese requirements document
- `internal/ui/confirm_dialog.go` - Dialog pattern reference
- `internal/ui/minibuffer.go` - Text editing reference
