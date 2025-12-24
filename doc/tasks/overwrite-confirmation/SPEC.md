# Feature: Overwrite Confirmation Dialog

## Overview

This feature adds an overwrite confirmation dialog when copying or moving files to a destination where a file with the same name already exists. Users can choose to overwrite, cancel, or save with a different name.

## Objectives

- Prevent accidental data loss during copy/move operations
- Provide clear information about source and destination files
- Allow users to rename files instead of overwriting
- Handle directory conflicts with appropriate error messages
- Maintain consistency with existing dialog patterns

## User Stories

- As a user, I want to be warned when copying a file would overwrite an existing file, so that I don't accidentally lose data
- As a user, I want to see file information (size, date) to help me decide whether to overwrite, so that I can make an informed decision
- As a user, I want to save with a different name instead of overwriting, so that I can keep both files
- As a user, I want directory conflicts to show an error, so that I understand why the operation cannot proceed

## Technical Requirements

### Functional Requirements

#### FR1: Overwrite Detection

- FR1.1: Before copy/move operation, check if destination file exists
- FR1.2: Use `os.Stat()` to check file existence (follows symlinks)
- FR1.3: For symlinks, check the symlink itself, not the target
- FR1.4: If no conflict, proceed with operation immediately

#### FR2: Overwrite Confirmation Dialog

- FR2.1: Display dialog when destination file exists
- FR2.2: Show filename and destination path
- FR2.3: Show file metadata (size in human-readable format, modification date)
- FR2.4: Provide three options: Overwrite, Cancel, Rename

#### FR3: Dialog Navigation

- FR3.1: Number keys (`1`, `2`, `3`) for direct selection
- FR3.2: `j`/`k` or arrow keys for cursor movement
- FR3.3: `Enter` to confirm current selection
- FR3.4: `Esc` to cancel (same as option 2)
- FR3.5: Cursor wraps around at boundaries

#### FR4: Overwrite Action

- FR4.1: Remove existing file before copy/move
- FR4.2: Proceed with original copy/move operation
- FR4.3: Reload both panes after operation

#### FR5: Rename Action

- FR5.1: Show input dialog for new filename
- FR5.2: Pre-populate with suggested name (e.g., `filename_copy.ext`)
- FR5.3: Validate input in real-time
- FR5.4: Show error if entered name already exists in destination
- FR5.5: Disable confirmation when name conflict exists
- FR5.6: On success, copy/move with new name

#### FR6: Directory Handling

- FR6.1: If source is directory and same-name directory exists at destination, show error dialog
- FR6.2: Do not attempt merge operation
- FR6.3: Error message: "Directory already exists"

#### FR7: Symlink Handling

- FR7.1: Treat symlinks same as regular files for confirmation
- FR7.2: Check symlink name, not target
- FR7.3: Broken symlinks: show error on copy attempt (existing behavior)

### Non-functional Requirements

#### NFR1: Performance

- NFR1.1: File existence check completes within 10ms
- NFR1.2: Dialog renders within 50ms

#### NFR2: Usability

- NFR2.1: Selected option is visually highlighted
- NFR2.2: File sizes displayed in human-readable format (KB, MB, GB)
- NFR2.3: Dates displayed in locale-appropriate format

#### NFR3: Consistency

- NFR3.1: Follow existing dialog styling (border, colors)
- NFR3.2: Use same key bindings as other dialogs (j/k navigation)
- NFR3.3: Implement Dialog interface

## Implementation Approach

### Architecture

```
Model (internal/ui/model.go)
  ├── dialog (Dialog interface)
  │   ├── ConfirmDialog (existing)
  │   ├── ErrorDialog (existing)
  │   ├── InputDialog (existing)
  │   ├── ContextMenuDialog (existing)
  │   └── OverwriteDialog (NEW)
  └── Update() handles dialog lifecycle
```

### Component Design

#### 1. OverwriteDialog (internal/ui/overwrite_dialog.go)

```go
// OverwriteChoice represents the user's choice in the overwrite dialog
type OverwriteChoice int

const (
    OverwriteChoiceOverwrite OverwriteChoice = iota
    OverwriteChoiceCancel
    OverwriteChoiceRename
)

// OverwriteDialog displays overwrite confirmation options
type OverwriteDialog struct {
    filename    string          // Name of the file being copied/moved
    destPath    string          // Destination directory path
    srcInfo     FileInfo        // Source file information
    destInfo    FileInfo        // Destination file information
    cursor      int             // Current selection (0-2)
    active      bool
    operation   string          // "copy" or "move"
}

// FileInfo holds file metadata for display
type FileInfo struct {
    Size    int64
    ModTime time.Time
}

func NewOverwriteDialog(filename, destPath string, srcInfo, destInfo FileInfo, operation string) *OverwriteDialog
func (d *OverwriteDialog) Update(msg tea.Msg) (Dialog, tea.Cmd)
func (d *OverwriteDialog) View() string
func (d *OverwriteDialog) IsActive() bool
func (d *OverwriteDialog) DisplayType() DialogDisplayType
```

#### 2. Overwrite Dialog Result Message

```go
type overwriteDialogResultMsg struct {
    choice   OverwriteChoice
    filename string  // Original filename
    newName  string  // New name (only for Rename choice)
    srcPath  string  // Full source path
    destPath string  // Destination directory
    operation string // "copy" or "move"
}
```

#### 3. Model Integration

Update `model.go` to:
1. Check for file existence before copy/move
2. Show OverwriteDialog if conflict exists
3. Handle overwriteDialogResultMsg
4. Execute appropriate action based on choice

```go
// In handleCopy/handleMove:
func (m *Model) checkAndCopy(srcPath, destDir string) tea.Cmd {
    filename := filepath.Base(srcPath)
    destPath := filepath.Join(destDir, filename)

    // Check if destination exists
    destInfo, err := os.Stat(destPath)
    if err == nil {
        // File exists - check if it's a directory
        srcInfo, _ := os.Stat(srcPath)
        if srcInfo.IsDir() && destInfo.IsDir() {
            // Directory conflict - show error
            return func() tea.Msg {
                return showErrorMsg{
                    message: fmt.Sprintf("Directory \"%s\" already exists", filename),
                }
            }
        }

        // File conflict - show overwrite dialog
        return func() tea.Msg {
            return showOverwriteDialogMsg{
                filename:  filename,
                srcPath:   srcPath,
                destPath:  destDir,
                srcInfo:   FileInfo{Size: srcInfo.Size(), ModTime: srcInfo.ModTime()},
                destInfo:  FileInfo{Size: destInfo.Size(), ModTime: destInfo.ModTime()},
                operation: "copy",
            }
        }
    }

    // No conflict - proceed with copy
    return m.executeCopy(srcPath, destDir)
}
```

#### 4. Rename Input Dialog Enhancement

Enhance existing `InputDialog` or create `RenameInputDialog`:

```go
// RenameInputDialog validates input against existing files
type RenameInputDialog struct {
    *InputDialog
    destPath      string   // Destination directory to check against
    existingFiles []string // Cached list of existing filenames
    hasError      bool     // Whether current input has validation error
    errorMessage  string   // Current error message
}

func NewRenameInputDialog(destPath string) *RenameInputDialog
func (d *RenameInputDialog) validateInput(name string) error
```

### Data Flow

```
User presses 'c' (copy)
    │
    ▼
Check destination for same filename
    │
    ├── Not exists ──────► Execute copy ──► Reload panes
    │
    └── Exists
        │
        ├── Both are directories ──► Show ErrorDialog
        │
        └── Otherwise ──► Show OverwriteDialog
                              │
                              ├── 1. Overwrite ──► Delete dest ──► Execute copy
                              │
                              ├── 2. Cancel ──► Close dialog
                              │
                              └── 3. Rename ──► Show RenameInputDialog
                                                    │
                                                    ├── Valid name ──► Execute copy with new name
                                                    │
                                                    └── Invalid ──► Show error, disable confirm
```

### UI/UX Design

#### Overwrite Confirmation Dialog

```
┌─ File already exists ────────────────┐
│                                       │
│  "example.txt" already exists in      │
│  /home/user/destination/              │
│                                       │
│  Source: 1.2 KB    2024-01-15 10:30   │
│  Dest:   2.5 KB    2024-01-10 15:45   │
│                                       │
│  1. Overwrite                    ←    │
│  2. Cancel                            │
│  3. Rename                            │
│                                       │
│  1-3/j/k:select Enter:confirm         │
└───────────────────────────────────────┘
```

#### Directory Error Dialog

```
┌─ Cannot copy directory ──────────────┐
│                                       │
│  Directory "src" already exists in    │
│  /home/user/destination/              │
│                                       │
│  [Enter] OK                           │
└───────────────────────────────────────┘
```

#### Rename Input with Validation Error

```
┌─ Rename ─────────────────────────────┐
│                                       │
│  New name: existing_file.txt          │
│            ~~~~~~~~~~~~~~~            │
│                                       │
│  File already exists                  │
│                                       │
│  Esc:cancel                           │
└───────────────────────────────────────┘
```

(Note: Enter hint removed when error exists)

### Styling (lipgloss)

- **Border**: Rounded, accent color (Color "39" - blue)
- **Title**: Bold, accent color
- **Filename**: Bold
- **Path**: Muted (Color "245")
- **File info labels**: Normal
- **File info values**: Muted (Color "245")
- **Selected option**: Background highlight (Color "39"), white foreground
- **Unselected option**: Default foreground
- **Error message**: Red foreground (Color "196")
- **Footer**: Muted (Color "245")

### Dependencies

#### Internal Dependencies

- `internal/ui/dialog.go`: Dialog interface
- `internal/ui/model.go`: Main model
- `internal/ui/input_dialog.go`: Base input dialog (for rename)
- `internal/fs/operations.go`: Copy, MoveFile, Delete functions

#### External Dependencies

- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/lipgloss`: Styling

### Key Algorithms

#### File Size Formatting

```go
func formatFileSize(bytes int64) string {
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
```

#### Suggested Rename

```go
func suggestRename(filename string, existingFiles map[string]bool) string {
    ext := filepath.Ext(filename)
    base := strings.TrimSuffix(filename, ext)

    // Try "name_copy.ext"
    candidate := base + "_copy" + ext
    if !existingFiles[candidate] {
        return candidate
    }

    // Try "name_copy_2.ext", "name_copy_3.ext", etc.
    for i := 2; i <= 100; i++ {
        candidate = fmt.Sprintf("%s_copy_%d%s", base, i, ext)
        if !existingFiles[candidate] {
            return candidate
        }
    }

    return filename // Fallback
}
```

## Test Scenarios

### Unit Tests (overwrite_dialog_test.go)

- [ ] **Test: Dialog creation**
  - Given: Filename, paths, and file info
  - When: NewOverwriteDialog is called
  - Then: Dialog is created with correct initial state (cursor at 0)

- [ ] **Test: Keyboard navigation - j/k**
  - Given: Dialog with cursor at position 0
  - When: 'j' key is pressed
  - Then: Cursor moves to position 1
  - When: 'j' pressed again
  - Then: Cursor moves to position 2
  - When: 'j' pressed again
  - Then: Cursor wraps to position 0

- [ ] **Test: Keyboard navigation - arrow keys**
  - Given: Dialog with cursor at position 1
  - When: Up arrow is pressed
  - Then: Cursor moves to position 0
  - When: Down arrow is pressed
  - Then: Cursor moves to position 1

- [ ] **Test: Number key selection**
  - Given: Active dialog
  - When: '1' key is pressed
  - Then: Returns overwriteDialogResultMsg with OverwriteChoiceOverwrite

- [ ] **Test: Enter key selection**
  - Given: Dialog with cursor at position 2 (Rename)
  - When: Enter key is pressed
  - Then: Returns overwriteDialogResultMsg with OverwriteChoiceRename

- [ ] **Test: Escape key**
  - Given: Active dialog
  - When: Esc key is pressed
  - Then: Returns overwriteDialogResultMsg with OverwriteChoiceCancel

- [ ] **Test: File size formatting**
  - Given: Various byte values
  - When: formatFileSize is called
  - Then: Returns human-readable sizes (e.g., "1.2 KB", "3.5 MB")

### Integration Tests

- [ ] **Test: Copy with overwrite dialog**
  - Given: Source file "test.txt", destination has "test.txt"
  - When: Copy operation is initiated
  - Then: OverwriteDialog is displayed

- [ ] **Test: Copy without conflict**
  - Given: Source file "new.txt", destination has no "new.txt"
  - When: Copy operation is initiated
  - Then: File is copied immediately, no dialog shown

- [ ] **Test: Directory conflict shows error**
  - Given: Source directory "src", destination has directory "src"
  - When: Copy operation is initiated
  - Then: ErrorDialog is displayed with appropriate message

- [ ] **Test: Overwrite action**
  - Given: OverwriteDialog displayed, user selects "Overwrite"
  - When: Enter is pressed
  - Then: Destination file is replaced with source file

- [ ] **Test: Cancel action**
  - Given: OverwriteDialog displayed
  - When: User presses Esc or selects Cancel
  - Then: Dialog closes, no file operation performed

- [ ] **Test: Rename action - valid name**
  - Given: OverwriteDialog displayed, user selects "Rename"
  - When: User enters valid new name and confirms
  - Then: File is copied with new name

- [ ] **Test: Rename action - invalid name**
  - Given: Rename input dialog displayed
  - When: User enters name that already exists
  - Then: Error message shown, Enter key does not confirm

- [ ] **Test: Move with overwrite**
  - Given: Source file, destination has same-name file
  - When: Move + Overwrite
  - Then: Source file moved, destination file replaced

- [ ] **Test: Context menu copy with overwrite**
  - Given: Context menu opened, Copy selected, destination has conflict
  - When: Copy action executed
  - Then: OverwriteDialog is displayed

### Manual Testing Checklist

- [ ] Overwrite dialog displays correctly centered in pane
- [ ] File information (size, date) is accurate and readable
- [ ] j/k and arrow key navigation works smoothly
- [ ] Number keys 1-3 work for direct selection
- [ ] Selected option is clearly highlighted
- [ ] Overwrite action successfully replaces file
- [ ] Cancel action closes dialog without changes
- [ ] Rename action opens input dialog
- [ ] Invalid rename shows error and prevents confirmation
- [ ] Valid rename successfully copies/moves with new name
- [ ] Directory conflict shows appropriate error
- [ ] Symlinks are handled correctly
- [ ] Both copy (c) and move (m) operations work
- [ ] Context menu operations work

## Success Criteria

### Functional Success

- [ ] Copy operation shows overwrite dialog when destination file exists
- [ ] Move operation shows overwrite dialog when destination file exists
- [ ] Overwrite action replaces destination file
- [ ] Cancel action aborts operation
- [ ] Rename action allows saving with different name
- [ ] Rename validates against existing files in destination
- [ ] Directory conflicts show error dialog
- [ ] Symlinks handled same as regular files

### Quality Success

- [ ] All existing tests pass
- [ ] New code has 80%+ test coverage
- [ ] Dialog response is immediate (< 100ms)
- [ ] UI design is consistent with existing dialogs

### User Experience Success

- [ ] File information helps users make informed decisions
- [ ] Navigation is intuitive (consistent with existing dialogs)
- [ ] Error messages are clear and actionable

## Open Questions

- [x] **Q1**: Should we support "Apply to all" for future batch operations?
  - **A1**: No, YAGNI. Add when multiple file selection is implemented.

- [x] **Q2**: How to handle directory conflicts?
  - **A2**: Show error dialog. No merge functionality (safety and simplicity).

- [x] **Q3**: How to handle symlinks?
  - **A3**: Treat same as regular files. Show overwrite confirmation.

- [x] **Q4**: What if renamed file also exists?
  - **A4**: Show inline error, disable confirmation until valid name entered.

## Future Considerations

### Phase 2: Multiple File Selection

When mark feature is implemented:
- Add "Apply to all" checkbox
- "Yes to all" / "No to all" options
- Progress indicator for batch operations

### Phase 3: Advanced Options

- "Compare files" option to show diff
- "Keep both" option (auto-rename)
- "Skip" option for batch operations

## Implementation Plan

### Step 1: Create OverwriteDialog (2-3 hours)

1. Create `internal/ui/overwrite_dialog.go`
2. Implement `OverwriteDialog` struct
3. Implement `Dialog` interface methods
4. Implement `View()` with file info display
5. Add unit tests

**Deliverables:**
- `overwrite_dialog.go`
- `overwrite_dialog_test.go`

### Step 2: Integrate with Copy/Move (2-3 hours)

1. Add file existence check before copy/move
2. Show OverwriteDialog on conflict
3. Handle overwriteDialogResultMsg
4. Implement Overwrite action
5. Implement Cancel action

**Deliverables:**
- Updated `model.go`
- Updated `keys.go` (if needed)

### Step 3: Implement Rename Flow (2-3 hours)

1. Create or enhance rename input dialog with validation
2. Implement real-time validation
3. Show error message for invalid names
4. Disable confirmation on error
5. Execute copy/move with new name on success

**Deliverables:**
- `rename_input_dialog.go` or updated `input_dialog.go`
- Unit tests

### Step 4: Directory Handling (1 hour)

1. Detect directory-to-directory conflict
2. Show error dialog for directory conflicts
3. Add tests

**Deliverables:**
- Updated conflict detection logic
- Integration tests

### Step 5: Context Menu Integration (1 hour)

1. Update context menu copy/move actions
2. Ensure overwrite dialog appears for context menu operations
3. Test all paths

**Deliverables:**
- Updated `context_menu_dialog.go`
- Integration tests

### Step 6: Testing and Polish (2 hours)

1. Manual testing on various scenarios
2. Edge case testing (permissions, symlinks)
3. UI polish (spacing, alignment)
4. Code documentation

**Deliverables:**
- Complete test coverage
- Documentation

## References

### Existing Implementation

- `/home/sakura/go/src/duofm/internal/ui/dialog.go` - Dialog interface
- `/home/sakura/go/src/duofm/internal/ui/confirm_dialog.go` - Confirm dialog pattern
- `/home/sakura/go/src/duofm/internal/ui/input_dialog.go` - Input dialog pattern
- `/home/sakura/go/src/duofm/internal/ui/context_menu_dialog.go` - Context menu
- `/home/sakura/go/src/duofm/internal/fs/operations.go` - File operations

### External References

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lipgloss Documentation](https://github.com/charmbracelet/lipgloss)
