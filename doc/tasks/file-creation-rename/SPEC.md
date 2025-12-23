# Feature: File and Directory Creation and Renaming

## Overview

This feature adds the ability to create new files, create new directories, and rename existing files/directories in duofm. Users can perform these operations using simple keyboard shortcuts with a dialog-based input interface.

## Objectives

- Enable quick file and directory creation without leaving the file manager
- Provide intuitive renaming functionality for files and directories
- Maintain consistency with existing UI patterns and keyboard-driven workflow
- Handle errors gracefully with clear user feedback

## User Stories

- As a user, I want to press `n` and enter a filename, so that I can quickly create an empty file
- As a user, I want to press `Shift+n` and enter a directory name, so that I can organize my files into new directories
- As a user, I want to press `r` and enter a new name, so that I can rename files and directories to fix typos or reorganize

## Technical Requirements

### Key Bindings

| Key | Action | Unix Equivalent |
|-----|--------|----------------|
| `n` | Create new file | `touch` (when file does not exist) |
| `Shift+n` | Create new directory | `mkdir` (when directory does not exist) |
| `r` | Rename file/directory | `mv` (same directory) |

### UI Design

#### Dialog Specification

**Display Type**: Pane-local dialog (DialogDisplayPane)
- Active pane is dimmed
- Inactive pane remains normal (allows viewing file list)
- Dialog overlays on center of active pane

**Dialog Layout**:
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

**Prompt Messages**:
- New file: `New file:`
- New directory: `New directory:`
- Rename: `Rename to:`

**Input Field**:
- Single-line text input
- Initial value:
  - New creation: empty
  - Rename: empty (do not pre-fill current filename)
- Visible cursor
- Basic text editing (insert, backspace, delete, cursor movement)

#### Input Controls

| Key | Action |
|-----|--------|
| `Enter` | Confirm input and execute create/rename |
| `Esc` | Cancel and return to normal state |
| Character keys | Insert character at cursor position |
| `Backspace` | Delete character before cursor |
| `Delete` | Delete character at cursor position |
| `←` / `→` | Move cursor left/right |
| `Ctrl+A` | Move cursor to beginning |
| `Ctrl+E` | Move cursor to end |
| `Ctrl+U` | Delete from cursor to beginning |
| `Ctrl+K` | Delete from cursor to end |

### Operation Specifications

#### New File Creation (n key)

1. User presses `n` key
2. Dialog appears with "New file:" prompt
3. User enters filename
4. User presses `Enter` to confirm
5. Create empty file in active pane's current directory
6. Reload both panes
7. Move cursor to created file
   - **Exception**: If hidden files are OFF and dot file is created, cursor stays at original position

#### New Directory Creation (Shift+n key)

1. User presses `Shift+n` key
2. Dialog appears with "New directory:" prompt
3. User enters directory name
4. User presses `Enter` to confirm
5. Create new directory in active pane's current directory
6. Reload both panes
7. Move cursor to created directory
   - **Exception**: If hidden files are OFF and dot directory is created, cursor stays at original position

#### Rename (r key)

1. User presses `r` key
2. Check selected entry:
   - If parent directory (..): do nothing (ignore)
   - If regular file/directory: proceed
3. Dialog appears with "Rename to:" prompt
4. User enters new name (input field starts empty)
5. User presses `Enter` to confirm
6. Rename within same directory
7. Reload both panes
8. Move cursor to renamed item
   - **Exception**: If hidden files are OFF and renamed to dot file, move cursor to next file
     - If it was the last file, move to previous file

### Error Handling

All errors are displayed in the status bar and automatically cleared after 5 seconds.

| Error Case | Status Bar Message | Behavior |
|-----------|-------------------|----------|
| Empty string confirmation | `File name cannot be empty` | Keep dialog open, allow continued input |
| File/directory already exists | `File already exists: [filename]` | Close dialog, do not create/rename |
| Path separator in name (`/`) | `Invalid file name: path separator not allowed` | Close dialog, do not create/rename |
| Permission error | `Permission denied: [error details]` | Close dialog, do not create/rename |
| Other filesystem errors | `Failed to create/rename: [error details]` | Close dialog, do not create/rename |

### Hidden File Handling

#### During Creation

- No restrictions on creating files/directories starting with dot (`.`)
- When creating a dot file with hidden files OFF:
  - File/directory is created successfully
  - Not shown in list (invisible until user toggles with `Ctrl+H`)
  - Cursor remains at original position

#### During Rename

- No restrictions on renaming to names starting with dot (`.`)
- When renaming to dot file with hidden files OFF:
  - Rename executes successfully
  - Not shown in list after rename
  - Cursor moves to next file (or previous if it was the last file)

### Cursor Position Control

#### After Creation

| Condition | Cursor Position |
|-----------|----------------|
| Normal file/directory created | Move to created item |
| Dot file created with hidden files ON | Move to created item |
| Dot file created with hidden files OFF | Stay at original position |

#### After Rename

| Condition | Cursor Position |
|-----------|----------------|
| Normal rename | Move to renamed item (follows sort order change) |
| Rename to dot file (hidden files OFF) | Move to next file (previous if last) |
| Rename from dot file (hidden files OFF) | Move to renamed item |

### Pane Reload

- After successful creation/rename: reload **both panes**
- Rationale: When left and right panes show the same directory, changes should be reflected in both

## Implementation Approach

### Architecture

```
internal/
├── ui/
│   ├── dialog.go           # Add InputDialog type
│   ├── input_dialog.go     # NEW: Input dialog implementation
│   ├── input_dialog_test.go
│   ├── model.go            # Add n, N, r key handlers
│   ├── keys.go             # Add new key constants
│   └── pane.go             # Add cursor tracking helpers
├── fs/
│   ├── operations.go       # Add CreateFile, CreateDirectory, Rename
│   └── operations_test.go
```

### Data Structures

#### InputDialog

```go
// InputDialog represents a text input dialog
type InputDialog struct {
    title       string              // Dialog title/prompt
    input       string              // Current input text
    cursorPos   int                 // Cursor position in input
    active      bool                // Dialog is active
    width       int                 // Dialog width
    onConfirm   func(string) tea.Cmd // Callback on Enter
    errorMsg    string              // Validation error message (for empty input)
}

// NewInputDialog creates a new input dialog
func NewInputDialog(title string, onConfirm func(string) tea.Cmd) *InputDialog

// Update handles key input
func (d *InputDialog) Update(msg tea.Msg) (Dialog, tea.Cmd)

// View renders the dialog
func (d *InputDialog) View() string

// IsActive returns whether dialog is active
func (d *InputDialog) IsActive() bool

// DisplayType returns pane-local display
func (d *InputDialog) DisplayType() DialogDisplayType
```

#### File Operations API

```go
// CreateFile creates an empty file
func CreateFile(path string) error {
    file, err := os.Create(path)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()
    return nil
}

// CreateDirectory creates a new directory
func CreateDirectory(path string) error {
    if err := os.Mkdir(path, 0755); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    return nil
}

// Rename renames a file or directory within the same directory
func Rename(oldPath, newName string) error {
    dir := filepath.Dir(oldPath)
    newPath := filepath.Join(dir, newName)

    // Check if newPath already exists
    if _, err := os.Stat(newPath); err == nil {
        return fmt.Errorf("file already exists: %s", newName)
    }

    if err := os.Rename(oldPath, newPath); err != nil {
        return fmt.Errorf("failed to rename: %w", err)
    }
    return nil
}
```

### Model Integration

```go
// In model.go Update function
case tea.KeyMsg:
    // Handle dialog input
    if m.dialog != nil {
        var cmd tea.Cmd
        m.dialog, cmd = m.dialog.Update(msg)
        return m, cmd
    }

    switch msg.String() {
    case KeyNewFile:
        m.dialog = NewInputDialog("New file:", func(filename string) error {
            if filename == "" {
                return fmt.Errorf("file name cannot be empty")
            }
            if strings.Contains(filename, "/") {
                return fmt.Errorf("invalid file name: path separator not allowed")
            }

            fullPath := filepath.Join(m.getActivePane().Path(), filename)
            if err := fs.CreateFile(fullPath); err != nil {
                return err
            }

            // Reload both panes
            m.getActivePane().LoadDirectory()
            m.getInactivePane().LoadDirectory()

            // Move cursor to created file (if visible)
            m.moveCursorToFile(filename)
            return nil
        })
        return m, nil

    case KeyNewDirectory:
        m.dialog = NewInputDialog("New directory:", func(dirname string) error {
            if dirname == "" {
                return fmt.Errorf("directory name cannot be empty")
            }
            if strings.Contains(dirname, "/") {
                return fmt.Errorf("invalid directory name: path separator not allowed")
            }

            fullPath := filepath.Join(m.getActivePane().Path(), dirname)
            if err := fs.CreateDirectory(fullPath); err != nil {
                return err
            }

            // Reload both panes
            m.getActivePane().LoadDirectory()
            m.getInactivePane().LoadDirectory()

            // Move cursor to created directory (if visible)
            m.moveCursorToFile(dirname)
            return nil
        })
        return m, nil

    case KeyRename:
        entry := m.getActivePane().SelectedEntry()
        if entry == nil || entry.IsParentDir() {
            // Ignore if parent directory selected
            return m, nil
        }

        oldName := entry.Name
        m.dialog = NewInputDialog("Rename to:", func(newName string) error {
            if newName == "" {
                return fmt.Errorf("file name cannot be empty")
            }
            if strings.Contains(newName, "/") {
                return fmt.Errorf("invalid file name: path separator not allowed")
            }

            oldPath := filepath.Join(m.getActivePane().Path(), oldName)
            if err := fs.Rename(oldPath, newName); err != nil {
                return err
            }

            // Reload both panes
            m.getActivePane().LoadDirectory()
            m.getInactivePane().LoadDirectory()

            // Move cursor to renamed file (if visible)
            m.moveCursorToFileAfterRename(oldName, newName)
            return nil
        })
        return m, nil
    }
```

### Helper Functions

```go
// moveCursorToFile moves cursor to the specified filename
// If file is not visible (hidden file with showHidden=false), cursor stays at original position
func (m *Model) moveCursorToFile(filename string) {
    pane := m.getActivePane()

    // Check if file is hidden and hidden files are OFF
    if strings.HasPrefix(filename, ".") && !pane.showHidden {
        // Don't move cursor
        return
    }

    // Find the file in entries
    for i, entry := range pane.entries {
        if entry.Name == filename {
            pane.cursor = i
            pane.EnsureCursorVisible()
            return
        }
    }
}

// moveCursorToFileAfterRename handles cursor movement after rename
func (m *Model) moveCursorToFileAfterRename(oldName, newName string) {
    pane := m.getActivePane()

    // If renamed to hidden file and hidden files are OFF
    if strings.HasPrefix(newName, ".") && !pane.showHidden {
        // Move to next file, or previous if it was the last
        if pane.cursor >= len(pane.entries)-1 {
            // Was last file, move to new last file
            if pane.cursor > 0 {
                pane.cursor--
            }
        }
        // If not last file, cursor stays at same index (next file)
        pane.EnsureCursorVisible()
        return
    }

    // Normal case: move to renamed file
    m.moveCursorToFile(newName)
}
```

### Input Validation

```go
// validateFilename checks if filename is valid
func validateFilename(name string) error {
    if name == "" {
        return fmt.Errorf("file name cannot be empty")
    }

    if strings.Contains(name, "/") {
        return fmt.Errorf("invalid file name: path separator not allowed")
    }

    // Let the filesystem handle other invalid characters
    return nil
}
```

## Test Scenarios

### New File Creation Tests

- [ ] `n` key opens input dialog with "New file:" prompt
- [ ] Entering valid filename and pressing Enter creates file
- [ ] Pressing Esc cancels without creating file
- [ ] Empty filename shows error in status bar
- [ ] Filename with `/` shows error in status bar
- [ ] Existing filename shows error in status bar
- [ ] Created file appears in list and cursor moves to it
- [ ] Creating dot file with hidden files OFF: file created but not visible, cursor stays

### New Directory Creation Tests

- [ ] `Shift+n` key opens input dialog with "New directory:" prompt
- [ ] Entering valid directory name and pressing Enter creates directory
- [ ] Pressing Esc cancels without creating directory
- [ ] Empty directory name shows error in status bar
- [ ] Directory name with `/` shows error in status bar
- [ ] Existing directory name shows error in status bar
- [ ] Created directory appears in list and cursor moves to it
- [ ] Creating dot directory with hidden files OFF: directory created but not visible, cursor stays

### Rename Tests

- [ ] `r` key on normal file opens input dialog with "Rename to:" prompt
- [ ] `r` key on directory opens input dialog
- [ ] `r` key on parent directory (..) does nothing
- [ ] Entering valid new name and pressing Enter renames file/directory
- [ ] Pressing Esc cancels without renaming
- [ ] Empty new name shows error in status bar
- [ ] New name with `/` shows error in status bar
- [ ] Existing filename shows error in status bar
- [ ] Renamed file appears in correct sort position and cursor moves to it
- [ ] Renaming to dot file with hidden files OFF: file disappears, cursor moves to next

### Cursor Movement Tests

- [ ] After creating file: cursor moves to new file
- [ ] After creating directory: cursor moves to new directory
- [ ] After creating dot file (hidden OFF): cursor stays at original position
- [ ] After renaming file: cursor moves to renamed file
- [ ] After renaming to dot file (hidden OFF): cursor moves to next file
- [ ] After renaming last file to dot file (hidden OFF): cursor moves to previous file

### Pane Reload Tests

- [ ] After creation: both panes reload
- [ ] When both panes show same directory: both reflect changes
- [ ] When panes show different directories: only active pane changes

### Error Handling Tests

- [ ] Permission denied: error shown in status bar
- [ ] Disk full: error shown in status bar
- [ ] Invalid filesystem characters: handled by OS, error shown
- [ ] File already exists: error shown in status bar

### Edge Cases

- [ ] Very long filename (>255 characters): handled by OS
- [ ] Unicode filename: handled correctly
- [ ] Filename with leading/trailing spaces: preserved as-is
- [ ] Multiple consecutive operations: all work correctly
- [ ] Operation during directory load: handled gracefully

## Success Criteria

- [ ] All three operations (create file, create directory, rename) work correctly
- [ ] Dialog appears as pane-local overlay on active pane
- [ ] Input field supports all specified text editing operations
- [ ] Enter confirms, Esc cancels consistently
- [ ] Empty input is rejected with clear error message
- [ ] Path separators are rejected with clear error message
- [ ] Existing files are detected and error is shown
- [ ] Permission errors are handled gracefully
- [ ] Cursor moves to created/renamed item when visible
- [ ] Cursor behavior with hidden files follows specification
- [ ] Both panes reload after successful operation
- [ ] Status bar errors auto-clear after 5 seconds
- [ ] Parent directory (..) is not renameable
- [ ] All existing functionality remains unaffected
- [ ] All unit tests pass
- [ ] Manual testing confirms smooth user experience

## Dependencies

- Go standard library (`os`, `path/filepath`, `strings`)
- Bubble Tea framework
- lipgloss for styling
- Existing `internal/fs` package
- Existing `internal/ui` package

## Open Questions

None - all requirements have been clarified.

## Out of Scope

The following features are explicitly out of scope for this implementation:

- Batch file creation
- File creation from templates
- Rename with move (to different directory)
- Showing current filename in rename dialog
- Filename auto-completion
- Undo functionality for create/rename
- Rename preview
- Confirmation dialog for overwrite (instead, show error and prevent operation)
