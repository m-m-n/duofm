# Feature: Permission Edit

## Overview

This feature adds the ability to change file and directory permissions (chmod functionality) directly within duofm. Users can modify permissions using numeric mode (octal notation), with support for recursive directory operations and separate permission settings for directories and files during recursive execution.

The feature is designed to integrate seamlessly with duofm's existing TUI interface, using the familiar dialog pattern and Bubble Tea architecture.

## Objectives

- Enable in-app permission changes without requiring external chmod commands
- Provide visual feedback with real-time symbolic notation updates
- Support recursive permission changes for directory trees
- Allow separate permission settings for directories and files during recursive operations
- Maintain consistency with existing duofm UI patterns and keyboard shortcuts
- Ensure safe operation with proper validation and error reporting

## User Stories

### US1: Quick Permission Change
As a user, I want to change file permissions with a simple keyboard shortcut, so that I can quickly adjust access rights without leaving duofm.

**Acceptance Criteria:**
- [ ] Press Shift+P to open permission dialog
- [ ] See current permission in both numeric and symbolic format
- [ ] Enter new permission as 3-digit octal number
- [ ] See symbolic notation update in real-time
- [ ] Press Enter to apply changes

### US2: Preset Shortcuts
As a user, I want to use quick presets for common permissions, so that I don't need to remember numeric values.

**Acceptance Criteria:**
- [ ] Press number keys 1-4 to select preset permissions
- [ ] See preset values for both files and directories
- [ ] Presets auto-fill the input field

### US3: Recursive Directory Permissions
As a developer, I want to recursively change permissions for an entire directory tree, so that I can quickly configure project permissions.

**Acceptance Criteria:**
- [ ] Select directory and press Shift+P
- [ ] Choose "Recursively" option via Tab or j/k keys
- [ ] Enter separate permissions for directories and files
- [ ] See progress bar during execution
- [ ] Receive error report if any files fail

### US4: Batch Permission Changes
As a system administrator, I want to apply the same permissions to multiple files at once, so that I can efficiently manage file access rights.

**Acceptance Criteria:**
- [ ] Mark multiple files with Space key
- [ ] Press Shift+P to open batch permission dialog
- [ ] Apply same permission to all marked items
- [ ] See progress for large batches
- [ ] Marked items are cleared after successful operation

## Technical Requirements

### Functional Requirements

#### FR1: Permission Dialog UI

- **FR1.1**: Pressing Shift+P opens the permission change dialog
- **FR1.2**: Dialog is centered in the active pane
- **FR1.3**: Dialog displays current permission in both numeric (e.g., 644) and symbolic (e.g., -rw-r--r--) format
- **FR1.4**: Numeric input field accepts 3-digit octal numbers (000-777)
- **FR1.5**: Symbolic notation updates in real-time as user types
- **FR1.6**: Dialog is not shown when parent directory (..) is selected
- **FR1.7**: For files, show file-specific presets
- **FR1.8**: For directories, show directory-specific presets and recursive option

#### FR2: Quick Presets

**File Presets:**
| Key | Mode | Symbolic | Description |
|-----|------|----------|-------------|
| 1 | 644 | -rw-r--r-- | Default file |
| 2 | 755 | -rwxr-xr-x | Executable |
| 3 | 600 | -rw------- | Private |
| 4 | 777 | -rwxrwxrwx | Full access |

**Directory Presets:**
| Key | Mode | Symbolic | Description |
|-----|------|----------|-------------|
| 1 | 755 | drwxr-xr-x | Default directory |
| 2 | 700 | drwx------ | Private directory |
| 3 | 775 | drwxrwxr-x | Group writable |
| 4 | 777 | drwxrwxrwx | Full access |

- **FR2.1**: Pressing number keys 1-4 auto-fills corresponding preset value
- **FR2.2**: User can edit preset value before applying
- **FR2.3**: Presets are context-sensitive (different for files vs directories)

#### FR3: Recursive Permission Changes

- **FR3.1**: For directories, show "Apply to:" section with two options:
  - "This directory only" (default)
  - "Recursively (all contents)"
- **FR3.2**: Tab key toggles between options
- **FR3.3**: j/k keys navigate options
- **FR3.4**: Space key selects option
- **FR3.5**: When "Recursively" is selected, show two-step input process

#### FR4: Two-Step Recursive Input

When recursive mode is selected:

- **FR4.1**: First dialog: "Enter permissions for directories:"
  - Shows input field for directory permission
  - Shows presets for directories
  - Shows real-time symbolic notation
- **FR4.2**: After Enter, second dialog: "Enter permissions for files:"
  - Shows input field for file permission
  - Shows presets for files
  - Shows real-time symbolic notation
- **FR4.3**: After second Enter, start recursive operation
- **FR4.4**: Esc at any step cancels the entire operation

#### FR5: Progress Display

For recursive operations and batch changes:

- **FR5.1**: Show progress dialog with:
  - Processed count / Total count
  - Progress bar with percentage
  - Current file path being processed
  - Elapsed time
- **FR5.2**: Progress updates every 100ms
- **FR5.3**: Ctrl+C cancels operation (with confirmation)

#### FR6: Error Reporting

- **FR6.1**: Continue processing even if individual files fail
- **FR6.2**: Collect all failures during operation
- **FR6.3**: After completion, show error report dialog if any failures occurred
- **FR6.4**: Error report includes:
  - Success count
  - Failure count
  - List of failed files with error reasons
  - Scrollable list if more than 10 failures
- **FR6.5**: Common errors to display:
  - "Permission denied" (EPERM)
  - "No such file or directory" (ENOENT)
  - "Read-only file system" (EROFS)

#### FR7: Batch Operations (Multiple Selection)

- **FR7.1**: When files are marked (Space key), Shift+P operates on all marked items
- **FR7.2**: Dialog title shows count: "Permissions: N items"
- **FR7.3**: Current permission display is omitted (multiple items have different permissions)
- **FR7.4**: Same permission is applied to all marked items
- **FR7.5**: Directories in batch are changed in non-recursive mode
- **FR7.6**: After successful operation, marks are cleared
- **FR7.7**: Show progress dialog for batches with more than 10 items

#### FR8: Symlink Handling

- **FR8.1**: Symlinks are skipped (not changed)
- **FR8.2**: Skipped symlinks are not counted as errors
- **FR8.3**: If only symlinks are selected, show informative message

#### FR9: Input Validation

- **FR9.1**: Accept only 3-digit octal numbers (000-777)
- **FR9.2**: Each digit must be 0-7
- **FR9.3**: Show inline error for invalid input
- **FR9.4**: Prevent submission of invalid values
- **FR9.5**: Error messages:
  - "Invalid permission: must be 3 digits (0-7)"
  - "Permission must be exactly 3 digits"

#### FR10: Keyboard Navigation

In permission dialog:
- **FR10.1**: Numbers 0-7: Enter digits
- **FR10.2**: Numbers 1-4: Apply presets
- **FR10.3**: Backspace: Delete digit
- **FR10.4**: Tab: Toggle recursive option (directories only)
- **FR10.5**: j/k: Navigate recursive options
- **FR10.6**: Space: Select recursive option
- **FR10.7**: Enter: Confirm and apply
- **FR10.8**: Esc: Cancel

### Non-Functional Requirements

#### NFR1: Performance

- **NFR1.1**: Dialog displays within 100ms of Shift+P press
- **NFR1.2**: Single file permission change completes within 50ms
- **NFR1.3**: Recursive operation processes at least 500 files/second on HDD
- **NFR1.4**: Progress bar updates every 100ms
- **NFR1.5**: Real-time symbolic notation update within 10ms of input

#### NFR2: Usability

- **NFR2.1**: Dialog width: 50 columns (fixed)
- **NFR2.2**: Dialog height: Auto-adjusts to content (max 20 rows)
- **NFR2.3**: Input field shows cursor position
- **NFR2.4**: Selected preset is visually highlighted
- **NFR2.5**: Clear indication of which option is selected in recursive mode
- **NFR2.6**: Error messages are concise and actionable

#### NFR3: Consistency

- **NFR3.1**: Dialog follows existing duofm dialog pattern
- **NFR3.2**: Uses same border style and colors as other dialogs
- **NFR3.3**: Keyboard shortcuts consistent with existing navigation
- **NFR3.4**: Error dialog uses same format as ErrorDialog
- **NFR3.5**: Progress dialog uses same format as ArchiveProgressDialog

#### NFR4: Reliability

- **NFR4.1**: Failed permission changes don't crash the application
- **NFR4.2**: Partial success is acceptable (some files changed, some failed)
- **NFR4.3**: Operation state is consistent even after errors
- **NFR4.4**: All file system errors are properly caught and reported

#### NFR5: Maintainability

- **NFR5.1**: Dialog implements the Dialog interface
- **NFR5.2**: Permission change logic is separate from UI
- **NFR5.3**: Preset configurations are easily modifiable
- **NFR5.4**: Comprehensive unit tests for permission validation
- **NFR5.5**: E2E tests for all user scenarios

## Implementation Approach

### Architecture

The permission edit feature follows duofm's existing architecture pattern:

```
Model (internal/ui/model.go)
  ├── dialog (Dialog interface)
  │   ├── ConfirmDialog
  │   ├── InputDialog
  │   ├── ArchiveProgressDialog
  │   ├── PermissionDialog (NEW)
  │   ├── RecursivePermDialog (NEW)
  │   └── PermissionErrorReportDialog (NEW)
  └── Update() handles dialog lifecycle

FileSystem (internal/fs/)
  ├── operations.go
  └── permissions.go (NEW)
      ├── ChangePermission()
      ├── ChangePermissionRecursive()
      └── ValidatePermissionMode()
```

**Key Design Decisions:**

1. **Dialog Pattern**: Implement permission UI as dialogs following existing patterns
2. **Two-Step Flow**: Recursive mode uses two sequential dialogs for clarity
3. **Separation of Concerns**: UI logic in internal/ui, file operations in internal/fs
4. **Progressive Enhancement**: Basic permission change first, then recursive, then batch
5. **Error Tolerance**: Continue on errors, collect and report at the end

### Data Flow

```
User presses Shift+P
    ↓
Determine target (file/dir/multiple)
    ↓
Show appropriate PermissionDialog
    ↓
User inputs permission (or selects preset)
    ↓
[If directory with recursive option]
    ↓
Show RecursivePermDialog (2 steps)
    ↓
Execute permission change
    ↓
[If recursive/batch] Show progress
    ↓
[If errors] Show error report
    ↓
Refresh pane display
    ↓
Return to file list
```

### Component Design

#### 1. PermissionDialog (internal/ui/permission_dialog.go)

```go
type PermissionDialog struct {
    targetName      string
    isDir           bool
    currentMode     fs.FileMode
    inputValue      string
    cursorPos       int
    recursiveOption int  // 0: this only, 1: recursive
    showRecursive   bool // true for directories
    presets         []PermissionPreset
    errorMsg        string
    active          bool
}

type PermissionPreset struct {
    Number      int
    Mode        string    // e.g., "755"
    Symbolic    string    // e.g., "drwxr-xr-x"
    Description string    // e.g., "Default directory"
}

// Interface methods
func (d *PermissionDialog) Update(msg tea.Msg) (Dialog, tea.Cmd)
func (d *PermissionDialog) View() string

// Helper methods
func (d *PermissionDialog) formatSymbolic(mode string) string
func (d *PermissionDialog) validateInput(input string) error
func (d *PermissionDialog) applyPreset(num int)
```

**Dialog Layout (File):**
```
┌─ Permissions: example.txt ─────────────────────┐
│                                                │
│   Mode: [644]  →   -rw-r--r--                  │
│         ↑                                      │
│                                                │
│   ── Quick Presets ────────────────────────    │
│   [1] 644  -rw-r--r--  (default file)          │
│   [2] 755  -rwxr-xr-x  (executable)            │
│   [3] 600  -rw-------  (private)               │
│   [4] 777  -rwxrwxrwx  (full access)           │
│                                                │
│   [Enter] Apply  [Esc] Cancel                  │
└────────────────────────────────────────────────┘
```

**Dialog Layout (Directory):**
```
┌─ Permissions: mydir/ ──────────────────────────┐
│                                                │
│   Mode: [755]  →   drwxr-xr-x                  │
│                                                │
│   ── Quick Presets ────────────────────────    │
│   [1] 755  drwxr-xr-x  (default directory)     │
│   [2] 700  drwx------  (private directory)     │
│   [3] 775  drwxrwxr-x  (group writable)        │
│   [4] 777  drwxrwxrwx  (full access)           │
│                                                │
│   Apply to:                                    │
│     (●) This directory only                    │
│     ( ) Recursively (all contents)             │
│                                                │
│   [Tab] Toggle  [Enter] Apply  [Esc] Cancel    │
└────────────────────────────────────────────────┘
```

#### 2. RecursivePermDialog (internal/ui/recursive_perm_dialog.go)

```go
type RecursivePermDialog struct {
    step          int  // 0: dir input, 1: file input
    dirMode       string
    fileMode      string
    currentInput  string
    cursorPos     int
    presets       []PermissionPreset
    errorMsg      string
    active        bool
}
```

**Dialog Layout (Step 1 - Directories):**
```
┌─ Recursive Permissions (1/2) ──────────────────┐
│                                                │
│   Permissions for DIRECTORIES:                 │
│                                                │
│   Mode: [755]  →   drwxr-xr-x                  │
│                                                │
│   ── Quick Presets ────────────────────────    │
│   [1] 755  drwxr-xr-x  (default directory)     │
│   [2] 700  drwx------  (private directory)     │
│   [3] 775  drwxrwxr-x  (group writable)        │
│                                                │
│   [Enter] Next  [Esc] Cancel                   │
└────────────────────────────────────────────────┘
```

**Dialog Layout (Step 2 - Files):**
```
┌─ Recursive Permissions (2/2) ──────────────────┐
│                                                │
│   Permissions for FILES:                       │
│   (Directories will use: 755)                  │
│                                                │
│   Mode: [644]  →   -rw-r--r--                  │
│                                                │
│   ── Quick Presets ────────────────────────    │
│   [1] 644  -rw-r--r--  (default file)          │
│   [2] 755  -rwxr-xr-x  (executable)            │
│   [3] 600  -rw-------  (private)               │
│                                                │
│   [Enter] Apply  [Esc] Cancel                  │
└────────────────────────────────────────────────┘
```

#### 3. PermissionProgressDialog (internal/ui/permission_progress_dialog.go)

```go
type PermissionProgressDialog struct {
    totalFiles     int
    processedFiles int
    currentFile    string
    startTime      time.Time
    errors         []PermissionError
    active         bool
}

type PermissionError struct {
    Path  string
    Error error
}
```

**Dialog Layout:**
```
┌─ Changing Permissions ─────────────────────────┐
│                                                │
│   Progress: 127 / 543 files                    │
│   ▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░  23%               │
│                                                │
│   Current: /home/user/project/src/main.go      │
│                                                │
│   Elapsed: 00:02                               │
│                                                │
│   [Ctrl+C] Cancel                              │
└────────────────────────────────────────────────┘
```

#### 4. PermissionErrorReportDialog (internal/ui/permission_error_report_dialog.go)

```go
type PermissionErrorReportDialog struct {
    successCount int
    failureCount int
    errors       []PermissionError
    scrollOffset int
    active       bool
}
```

**Dialog Layout:**
```
┌─ Permission Change Report ─────────────────────┐
│                                                │
│   Success: 520 files                           │
│   Failed:  23 files                            │
│                                                │
│   Failed files:                                │
│   - /path/to/file1.txt                         │
│     Error: Permission denied                   │
│   - /path/to/file2.txt                         │
│     Error: Permission denied                   │
│   - /path/to/dir/file3.sh                      │
│     Error: No such file or directory           │
│   ... (20 more)                                │
│                                                │
│   [j/k] Scroll  [Enter] Close                  │
└────────────────────────────────────────────────┘
```

#### 5. Permission Operations (internal/fs/permissions.go)

```go
// ChangePermission changes permission of a single file/directory
func ChangePermission(path string, mode fs.FileMode) error

// ChangePermissionRecursive recursively changes permissions
// dirMode: permission for directories
// fileMode: permission for files
// Returns: success count, errors
func ChangePermissionRecursive(
    rootPath string,
    dirMode fs.FileMode,
    fileMode fs.FileMode,
    progressCallback func(current, total int, path string),
) (successCount int, errors []PermissionError, err error)

// ValidatePermissionMode validates octal permission string (000-777)
func ValidatePermissionMode(mode string) error

// ParsePermissionMode converts string to fs.FileMode
func ParsePermissionMode(mode string) (fs.FileMode, error)

// FormatSymbolic converts fs.FileMode to symbolic string
// isDir: true for directories (shows 'd' prefix)
func FormatSymbolic(mode fs.FileMode, isDir bool) string

// PermissionError represents a permission change failure
type PermissionError struct {
    Path  string
    Error error
}
```

### State Machine

```
[File List] --Shift+P--> [Permission Dialog]
    |                           |
    |                    [User inputs mode]
    |                           |
    |                    [Enter pressed]
    |                           |
    +--(Simple)--> [Apply Permission]
    |                           |
    +--(Recursive)--> [Recursive Dialog Step 1]
                               |
                        [Dir mode entered]
                               |
                        [Recursive Dialog Step 2]
                               |
                        [File mode entered]
                               |
                        [Progress Dialog]
                               |
                        [Apply Recursive]
                               |
                        [Error Report (if errors)]
                               |
                        [Back to File List]
```

### File Structure

```
internal/
├── fs/
│   ├── permissions.go           # Permission change operations
│   └── permissions_test.go      # Unit tests
└── ui/
    ├── keys.go                  # Add KeyPermission = "P"
    ├── permission_dialog.go     # Basic permission dialog
    ├── permission_dialog_test.go
    ├── recursive_perm_dialog.go # Two-step recursive dialog
    ├── recursive_perm_dialog_test.go
    ├── permission_progress_dialog.go
    ├── permission_progress_dialog_test.go
    ├── permission_error_report_dialog.go
    ├── permission_error_report_dialog_test.go
    ├── model_update.go          # Add permission message handlers
    └── messages.go              # Add permission-related messages
```

## Test Scenarios

### Unit Tests

#### Permission Validation
- [ ] Test 1: ValidatePermissionMode("644") - Should pass
- [ ] Test 2: ValidatePermissionMode("755") - Should pass
- [ ] Test 3: ValidatePermissionMode("000") - Should pass (boundary)
- [ ] Test 4: ValidatePermissionMode("777") - Should pass (boundary)
- [ ] Test 5: ValidatePermissionMode("888") - Should fail (invalid digit)
- [ ] Test 6: ValidatePermissionMode("12") - Should fail (too short)
- [ ] Test 7: ValidatePermissionMode("1234") - Should fail (too long)
- [ ] Test 8: ValidatePermissionMode("abc") - Should fail (non-numeric)

#### Symbolic Formatting
- [ ] Test 1: FormatSymbolic(0644, false) → "-rw-r--r--"
- [ ] Test 2: FormatSymbolic(0755, true) → "drwxr-xr-x"
- [ ] Test 3: FormatSymbolic(0777, false) → "-rwxrwxrwx"
- [ ] Test 4: FormatSymbolic(0000, false) → "----------"

#### Dialog Behavior
- [ ] Test 1: Preset selection updates input field
- [ ] Test 2: Input validation shows error for invalid values
- [ ] Test 3: Real-time symbolic update on input change
- [ ] Test 4: Recursive option toggling (Tab key)
- [ ] Test 5: Two-step dialog progression

### Integration Tests

- [ ] Test 1: Change single file permission (644 → 755)
  - Create test file with 644
  - Apply 755 via dialog
  - Verify file mode is 755
- [ ] Test 2: Change directory permission (755 → 700)
  - Create test directory with 755
  - Apply 700 via dialog
  - Verify directory mode is 700
- [ ] Test 3: Recursive permission change
  - Create directory tree with mixed permissions
  - Apply recursive change (dir:755, file:644)
  - Verify all directories are 755
  - Verify all files are 644
  - Verify symlinks are untouched
- [ ] Test 4: Batch permission change
  - Mark multiple files
  - Apply same permission
  - Verify all marked files have new permission
- [ ] Test 5: Error handling - permission denied
  - Create file owned by root
  - Attempt to change permission
  - Verify error dialog is shown
  - Verify application continues to function

### E2E Tests

#### Test 1: Basic Permission Change
```bash
test_basic_permission_change() {
    start_duofm "$CURRENT_SESSION"

    # Create test file
    echo "test" > /testdata/file.txt
    chmod 644 /testdata/file.txt

    # Navigate to file and open permission dialog
    send_keys "$CURRENT_SESSION" "P"
    sleep 0.3
    assert_contains "$CURRENT_SESSION" "Permissions: file.txt" \
        "Permission dialog opened"

    # Enter new permission
    send_keys "$CURRENT_SESSION" "7" "5" "5"
    assert_contains "$CURRENT_SESSION" "-rwxr-xr-x" \
        "Symbolic notation updated"

    # Apply
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Verify permission changed
    PERM=$(stat -c "%a" /testdata/file.txt)
    if [ "$PERM" = "755" ]; then
        echo "✓ Permission changed successfully"
    else
        echo "✗ Permission change failed: expected 755, got $PERM"
        exit 1
    fi

    stop_duofm "$CURRENT_SESSION"
}
```

#### Test 2: Preset Selection
```bash
test_preset_selection() {
    start_duofm "$CURRENT_SESSION"

    # Open permission dialog
    send_keys "$CURRENT_SESSION" "P"
    sleep 0.3

    # Press preset 2 (755)
    send_keys "$CURRENT_SESSION" "2"
    assert_contains "$CURRENT_SESSION" "755" "Preset applied"
    assert_contains "$CURRENT_SESSION" "-rwxr-xr-x" "Symbolic updated"

    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    stop_duofm "$CURRENT_SESSION"
}
```

#### Test 3: Recursive Permission Change
```bash
test_recursive_permission() {
    start_duofm "$CURRENT_SESSION"

    # Create directory tree
    mkdir -p /testdata/recursive_test/subdir
    touch /testdata/recursive_test/file1.txt
    touch /testdata/recursive_test/subdir/file2.txt

    # Navigate to directory
    send_keys "$CURRENT_SESSION" "j" "j" "Enter"
    sleep 0.3

    # Open permission dialog
    send_keys "$CURRENT_SESSION" "P"
    sleep 0.3

    # Select recursive option
    send_keys "$CURRENT_SESSION" "Tab"
    assert_contains "$CURRENT_SESSION" "(●) Recursively" \
        "Recursive option selected"

    # Confirm
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Enter directory permission
    send_keys "$CURRENT_SESSION" "7" "5" "5"
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Enter file permission
    send_keys "$CURRENT_SESSION" "6" "4" "4"
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Verify permissions
    DIR_PERM=$(stat -c "%a" /testdata/recursive_test)
    FILE_PERM=$(stat -c "%a" /testdata/recursive_test/file1.txt)

    if [ "$DIR_PERM" = "755" ] && [ "$FILE_PERM" = "644" ]; then
        echo "✓ Recursive permission change successful"
    else
        echo "✗ Failed: dir=$DIR_PERM (expected 755), file=$FILE_PERM (expected 644)"
        exit 1
    fi

    stop_duofm "$CURRENT_SESSION"
}
```

#### Test 4: Error Handling
```bash
test_permission_denied() {
    start_duofm "$CURRENT_SESSION"

    # Create file owned by root (in Docker)
    sudo touch /testdata/root_file.txt
    sudo chmod 644 /testdata/root_file.txt

    # Try to change permission
    send_keys "$CURRENT_SESSION" "P"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "7" "5" "5"
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Verify error is shown
    assert_contains "$CURRENT_SESSION" "Permission denied" \
        "Error dialog shown"

    # Close error dialog
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Verify duofm still works
    send_keys "$CURRENT_SESSION" "j"
    assert_cursor_position "$CURRENT_SESSION" "2" \
        "Navigation still functional after error"

    stop_duofm "$CURRENT_SESSION"
}
```

### Edge Cases

- [ ] Edge case 1: Permission 000 (no permissions)
- [ ] Edge case 2: Permission 777 (full permissions)
- [ ] Edge case 3: Symlink in directory tree (should be skipped)
- [ ] Edge case 4: Parent directory (..) selected (should be disabled)
- [ ] Edge case 5: Empty directory (recursive change)
- [ ] Edge case 6: Very large directory tree (10000+ files)
- [ ] Edge case 7: File deleted during operation
- [ ] Edge case 8: Permission denied on some but not all files

### Performance Tests

- [ ] Performance 1: Single file change completes within 50ms
- [ ] Performance 2: Dialog displays within 100ms
- [ ] Performance 3: 1000 files recursive change completes within 2 seconds
- [ ] Performance 4: Progress updates at least 10 times per second

## Security Considerations

- **Permission Validation**: All input is validated before execution (000-777 range)
- **No Privilege Escalation**: Operations respect OS-level permissions; no sudo or elevated privileges
- **Symlink Safety**: Symlinks are skipped to prevent following malicious links
- **Parent Directory Protection**: Cannot change permissions of parent directory (..)
- **Input Sanitization**: Path inputs are sanitized to prevent path traversal
- **Error Information**: Error messages don't leak sensitive system information

## Error Handling

### Error Codes

| Error | Condition | User Message |
|-------|-----------|--------------|
| EINVAL | Invalid permission value | "Invalid permission: must be 3 digits (0-7)" |
| EPERM | Permission denied | "Permission denied: cannot change this file" |
| ENOENT | File not found | "File not found: {path}" |
| EROFS | Read-only filesystem | "Cannot modify: read-only filesystem" |
| EIO | I/O error | "I/O error: {path}" |

### Error Flow

```
Operation Starts
    ↓
Execute chmod
    ↓
[Error?] --Yes--> Collect error
    |                ↓
    |          Continue to next file
    ↓
[More files?] --Yes--> Loop
    |
    No
    ↓
[Any errors?] --Yes--> Show Error Report Dialog
    |
    No
    ↓
Show Success Status
```

## Performance Optimization

### Performance Goals
- Dialog display: < 100ms
- Single file change: < 50ms
- Recursive operation: > 500 files/second (HDD)
- Progress update interval: 100ms

### Optimization Strategies
- **Batch System Calls**: Group multiple chmod calls when possible
- **Buffered Progress Updates**: Update UI every 100ms, not per file
- **Goroutine for Long Operations**: Run recursive operations in background
- **Early Validation**: Validate input before starting operation
- **Minimize Allocations**: Reuse error slice, avoid unnecessary string allocations

### Caching Strategy
- No caching needed (permissions are direct system calls)

## Success Criteria

- [ ] All functional requirements (FR1-FR10) are implemented and tested
- [ ] All test scenarios pass (unit, integration, E2E)
- [ ] Performance meets specified goals (100ms dialog, 50ms single change)
- [ ] Security requirements are satisfied (validation, no privilege escalation)
- [ ] Error handling is robust (continues on errors, reports failures)
- [ ] Documentation is complete (code comments, test documentation)
- [ ] Code review is completed
- [ ] E2E tests added to test suite
- [ ] Feature is integrated with main branch

## Open Questions

- [ ] Should we add undo functionality for permission changes? (Out of scope for initial version)
- [ ] Should we support symbolic mode (u+x, go-w) in addition to numeric mode? (Future enhancement)
- [ ] Should we display current permission in the file list view? (UI enhancement consideration)
- [ ] Should we support special permissions (setuid/setgid/sticky bit) with dedicated UI? (Currently only via numeric input)

## Implementation Phases

### Phase 1: Basic Permission Dialog
**Goals:** Single file/directory permission change with presets
**Deliverables:**
- PermissionDialog component
- Basic permission change in internal/fs/permissions.go
- Input validation and symbolic formatting
- Preset functionality
- Unit tests for validation and formatting
- E2E test for basic permission change

### Phase 2: Recursive Mode
**Goals:** Recursive directory permission changes
**Deliverables:**
- RecursivePermDialog component (two-step)
- Recursive option in PermissionDialog
- ChangePermissionRecursive function
- Symlink skipping logic
- Unit tests for recursive operations
- E2E test for recursive permission change

### Phase 3: Progress and Error Reporting
**Goals:** Progress display and comprehensive error handling
**Deliverables:**
- PermissionProgressDialog component
- PermissionErrorReportDialog component
- Progress callback mechanism
- Error collection and reporting
- E2E test for error handling

### Phase 4: Batch Operations
**Goals:** Multiple selection support
**Deliverables:**
- Batch mode detection in PermissionDialog
- Batch permission change logic
- Progress display for batch operations
- Mark clearing after successful batch
- E2E test for batch operations

## References

- Go os.Chmod documentation: https://pkg.go.dev/os#Chmod
- Unix chmod manual: `man 1 chmod`, `man 2 chmod`
- Existing dialog implementations: `internal/ui/confirm_dialog.go`, `internal/ui/input_dialog.go`
- Archive progress dialog: `internal/ui/archive_progress_dialog.go`
- Bubble Tea framework: https://github.com/charmbracelet/bubbletea
- duofm CONTRIBUTING.md: `/home/sakura/cache/worktrees/feat-permission-edit/doc/CONTRIBUTING.md`
