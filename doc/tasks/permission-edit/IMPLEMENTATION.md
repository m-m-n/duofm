# Implementation Plan: Permission Edit

## Overview

Implement in-app permission editing (chmod functionality) with support for single files, recursive directory operations, and batch changes. The feature integrates with duofm's existing Bubble Tea dialog system and follows established patterns for progress tracking and error reporting.

## Objectives

- Enable users to change file/directory permissions without leaving duofm
- Provide intuitive UI with real-time symbolic notation feedback
- Support recursive permission changes with separate directory/file modes
- Handle batch operations for multiple selected files
- Ensure robust error handling with detailed failure reports

## Prerequisites

### Development Environment

- Go 1.21 or later
- Bubble Tea framework (github.com/charmbracelet/bubbletea)
- Lip Gloss styling library (github.com/charmbracelet/lipgloss)
- Unix-like system with standard chmod support

### Dependencies

**External Dependencies:**
- Bubble Tea v0.25.0+ (already in project)
- Lip Gloss v0.9.1+ (already in project)

**Internal Dependencies:**
- Dialog interface (`internal/ui/dialog.go`)
- Message architecture (`internal/ui/messages.go`)
- File operations module (`internal/fs/operations.go`)
- UI styling patterns (`internal/ui/styles.go`)

### Knowledge Requirements

- Bubble Tea's Elm Architecture (Model-Update-View pattern)
- Unix file permission model (octal notation, symbolic representation)
- Go's `os.Chmod` and `filepath.Walk` APIs
- duofm's dialog lifecycle and message handling

## Architecture Overview

### Technology Stack

- **Language**: Go 1.21+
- **TUI Framework**: Bubble Tea (Elm Architecture)
- **Styling**: Lip Gloss (declarative terminal styling)
- **File System**: Go standard library (`os`, `filepath`)

### Design Approach

**Separation of Concerns:**
- UI components in `internal/ui/` (dialogs, rendering, interaction)
- File operations in `internal/fs/` (permission changes, validation, recursive traversal)
- Messages in `internal/ui/messages.go` (operation lifecycle communication)

**Dialog-Based Workflow:**
- Follow existing InputDialog and ArchiveProgressDialog patterns
- Use multi-step dialog flow for recursive operations
- Implement progress tracking for long-running operations
- Provide comprehensive error reporting after completion

**Progressive Enhancement:**
- Phase 1: Basic single-file permission change
- Phase 2: Add recursive directory support
- Phase 3: Add progress tracking and error reporting
- Phase 4: Add batch operation support

### Component Interaction

```
User Input (Shift+P)
    ↓
Model Update Handler
    ↓
PermissionDialog Creation
    ↓
User enters permission → Real-time symbolic update
    ↓
[If recursive selected] → RecursivePermDialog (2 steps)
    ↓
Execute Operation → Send messages via Bubble Tea Cmd
    ↓
[If long operation] → PermissionProgressDialog
    ↓
Operation Complete → Collect errors
    ↓
[If errors exist] → PermissionErrorReportDialog
    ↓
Refresh Directory Listing → Return to normal mode
```

**Message Flow:**
1. `permissionOperationStartMsg` - Triggers operation execution
2. `permissionProgressUpdateMsg` - Periodic progress updates
3. `permissionOperationCompleteMsg` - Operation finished (success/failure)
4. `directoryLoadCompleteMsg` - Refresh pane after completion

## Implementation Phases

### Phase 1: Permission Core and Basic Dialog

**Goal**: Implement single file/directory permission changes with validation and symbolic notation display.

**Files to Create:**

- `internal/fs/permissions.go` - Permission change operations and validation
- `internal/fs/permissions_test.go` - Unit tests for permission operations
- `internal/ui/permission_dialog.go` - Basic permission change dialog
- `internal/ui/permission_dialog_test.go` - Dialog unit tests

**Files to Modify:**

- `internal/ui/messages.go`:
  - Add permission operation messages (start, complete, error)
- `internal/ui/keys.go`:
  - Add Shift+P keybinding definition
- `internal/ui/model_update.go`:
  - Add Shift+P handler to open PermissionDialog
  - Add message handlers for permission operations

**Key Components:**

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ValidatePermissionMode | Validate octal permission string format | Input is non-empty string | Returns nil if valid (000-777), error otherwise |
| ParsePermissionMode | Convert string to FileMode | Input is valid octal string | Returns FileMode value or error |
| FormatSymbolic | Convert FileMode to symbolic string | Valid FileMode value | Returns symbolic string (e.g., "-rw-r--r--" or "drwxr-xr-x") |
| ChangePermission | Execute chmod on single path | Path exists and is accessible | File permission changed or error returned |
| PermissionDialog | Display and handle permission input | Target file/directory selected | User enters permission or cancels |

**Processing Flow:**

```
1. User presses Shift+P on selected entry
   ├─ Entry is ".." → Show error "Cannot change parent directory permissions"
   └─ Entry is valid → Continue

2. Create PermissionDialog
   ├─ Retrieve current permission from file stat
   ├─ Determine if target is directory
   └─ Load appropriate presets (file vs directory)

3. User inputs permission digits (0-7)
   ├─ On each keystroke → Validate digit
   ├─ Valid digit → Update input field
   ├─ Invalid digit → Show inline error
   └─ Update symbolic notation in real-time

4. User presses preset key (1-4)
   └─ Fill input field with preset value

5. User confirms (Enter)
   ├─ Validate complete permission (3 digits, all 0-7)
   ├─ Valid → Parse to FileMode → Execute ChangePermission
   ├─ Invalid → Show error, remain in dialog
   └─ Success → Close dialog, show status message, refresh pane

6. User cancels (Esc)
   └─ Close dialog, no changes
```

**Implementation Steps:**

1. **Create Permission Operations Module (`internal/fs/permissions.go`)**
   - Implement validation function (3-digit octal check)
   - Implement parsing function (string to FileMode conversion)
   - Implement symbolic formatter (FileMode to readable string)
   - Implement single-file permission change function
   - Key considerations:
     - Validate each digit is in range 0-7
     - Handle directory prefix 'd' in symbolic notation
     - Use `os.Chmod` for actual permission changes
     - Return descriptive errors for common failure cases

2. **Create PermissionDialog Component (`internal/ui/permission_dialog.go`)**
   - Implement Dialog interface (Update, View, IsActive, DisplayType)
   - Handle keyboard input (digits 0-7, presets 1-4, Backspace)
   - Implement real-time symbolic notation update
   - Render dialog with current permission, input field, presets
   - Key considerations:
     - Cursor management for input field (reuse InputDialog patterns)
     - Separate preset sets for files vs directories
     - Error message display for validation failures
     - Width: 50 columns (consistent with InputDialog)

3. **Add Keybinding and Message Handlers**
   - Define Shift+P in keys.go
   - Create permission operation messages in messages.go
   - Add Shift+P handler in model_update.go to create dialog
   - Add completion handler to refresh pane and show status
   - Key considerations:
     - Check for ".." entry and show error if selected
     - Pass current FileMode to dialog for display
     - Handle success/failure messages appropriately

4. **Implement Unit Tests**
   - Test validation (valid/invalid inputs, boundary cases)
   - Test symbolic formatting (various permission combinations)
   - Test permission parsing (octal to FileMode)
   - Test dialog behavior (preset selection, input validation)
   - Key considerations:
     - Table-driven tests for validation (cover 000-777 range)
     - Test edge cases (empty input, non-numeric, too short/long)
     - Test symbolic notation for files vs directories

**Dependencies:**
- Requires: None (foundation phase)
- Blocks: Phase 2 (recursive mode), Phase 3 (progress), Phase 4 (batch)

**Testing Approach:**

*Unit Tests:*
- Test ValidatePermissionMode with valid inputs (644, 755, 000, 777)
- Test ValidatePermissionMode with invalid inputs (888, 12, 1234, abc)
- Test FormatSymbolic for files (0644 → "-rw-r--r--")
- Test FormatSymbolic for directories (0755 → "drwxr-xr-x")
- Test ParsePermissionMode conversion accuracy
- Test preset selection updates input field correctly

*Integration Tests:*
- Create test file with 644, change to 755, verify file mode
- Create test directory with 755, change to 700, verify directory mode
- Test error handling for non-existent files
- Test error handling for permission denied scenarios

*Manual Testing:*
- [ ] Open permission dialog on file, verify current permission displayed
- [ ] Enter valid permission (755), verify symbolic updates in real-time
- [ ] Apply permission, verify file mode changed via `ls -l`
- [ ] Try invalid input (888), verify error message shown
- [ ] Press preset key (2), verify input auto-filled
- [ ] Cancel dialog (Esc), verify no changes made

**Acceptance Criteria:**
- [ ] Shift+P opens permission dialog with current permission displayed
- [ ] Dialog shows both numeric and symbolic notation
- [ ] Real-time symbolic update as user types
- [ ] Preset keys (1-4) auto-fill appropriate values
- [ ] Enter applies permission change successfully
- [ ] Invalid input shows clear error message
- [ ] ".." entry shows error when Shift+P pressed
- [ ] Single file/directory permission changes complete within 50ms
- [ ] All unit tests pass with 80%+ coverage

**Estimated Effort**: 中 (3-5 days)

**Risks and Mitigation:**
- **Risk**: Permission change fails on read-only filesystems
  - **Mitigation**: Detect EROFS error and show user-friendly message
- **Risk**: Symbolic notation formatting is complex
  - **Mitigation**: Use bit masking to extract owner/group/other permissions separately

---

### Phase 2: Recursive Permission Changes

**Goal**: Enable recursive permission changes for directory trees with separate directory and file modes.

**Files to Create:**

- `internal/ui/recursive_perm_dialog.go` - Two-step recursive permission dialog
- `internal/ui/recursive_perm_dialog_test.go` - Recursive dialog tests

**Files to Modify:**

- `internal/fs/permissions.go`:
  - Add ChangePermissionRecursive function
  - Add symlink detection and skipping logic
- `internal/ui/permission_dialog.go`:
  - Add "Apply to" section for directories
  - Add recursive option toggle (Tab, j/k, Space keys)
  - Transition to RecursivePermDialog when recursive selected

**Key Components:**

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ChangePermissionRecursive | Recursively change permissions in directory tree | Valid directory path, valid dir/file modes | All directories/files changed, errors collected |
| RecursivePermDialog | Two-step permission input for dirs and files | User selected recursive option | Returns dir mode and file mode, or cancellation |
| RecursiveOptionToggle | Toggle between "this only" and "recursive" | Directory selected in PermissionDialog | Option state updated, visible in UI |
| SymlinkSkipper | Detect and skip symlinks during traversal | Entry in directory tree | Returns true if symlink (skip), false otherwise |

**Processing Flow:**

```
1. User opens PermissionDialog on directory
   └─ Show "Apply to" section with two options

2. User toggles recursive option (Tab or j/k)
   ├─ "This directory only" (default)
   └─ "Recursively (all contents)"

3. User selects "Recursively" and presses Enter
   └─ Transition to RecursivePermDialog

4. RecursivePermDialog Step 1: Directory Permissions
   ├─ Show input field for directory permission
   ├─ Show directory-specific presets
   ├─ User enters permission or uses preset
   └─ User confirms → Store dir mode, proceed to Step 2

5. RecursivePermDialog Step 2: File Permissions
   ├─ Show previously entered dir permission
   ├─ Show input field for file permission
   ├─ Show file-specific presets
   ├─ User enters permission or uses preset
   └─ User confirms → Execute recursive operation

6. Execute ChangePermissionRecursive
   ├─ Walk directory tree using filepath.Walk
   ├─ For each entry:
   │   ├─ Is symlink? → Skip (no error)
   │   ├─ Is directory? → Apply dir mode
   │   └─ Is file? → Apply file mode
   ├─ Collect errors (continue on individual failures)
   └─ Return success count and error list

7. Show result
   ├─ All succeeded → Show success status
   └─ Some failed → Proceed to Phase 3 error reporting
```

**Implementation Steps:**

1. **Add Recursive Option to PermissionDialog**
   - Add "Apply to" section rendering for directories
   - Implement Tab key toggle between options
   - Implement j/k navigation and Space selection
   - Visual indicator for selected option (radio button style)
   - Key considerations:
     - Only show for directories (isDir flag)
     - Default to "This directory only"
     - Transition to RecursivePermDialog when recursive + Enter

2. **Create RecursivePermDialog Component**
   - Implement two-step dialog (step 0: dir input, step 1: file input)
   - Handle Enter to advance steps
   - Handle Esc to cancel entire operation
   - Display previously entered value in step 2
   - Key considerations:
     - Reuse permission input logic from PermissionDialog
     - Show different presets per step (dir vs file)
     - Clear step indicator (1/2, 2/2)

3. **Implement ChangePermissionRecursive Function**
   - Use filepath.Walk for directory traversal
   - Detect symlinks and skip them
   - Apply appropriate mode based on entry type
   - Collect errors without stopping operation
   - Return success count and error slice
   - Key considerations:
     - Use os.Lstat to detect symlinks (not os.Stat)
     - Don't follow symlinks (could cause infinite loops)
     - Continue processing even if individual files fail
     - Efficient error collection (pre-allocate slice)

4. **Implement Unit Tests**
   - Test recursive traversal with mixed files/directories
   - Test symlink skipping behavior
   - Test partial success scenarios (some files fail)
   - Test RecursivePermDialog step progression
   - Key considerations:
     - Create temporary directory trees for testing
     - Include symlinks in test data
     - Verify symlinks are untouched
     - Test cancellation at each step

**Dependencies:**
- Requires: Phase 1 (basic permission dialog and operations)
- Blocks: Phase 3 (progress display), Phase 4 (batch operations)

**Testing Approach:**

*Unit Tests:*
- Test ChangePermissionRecursive on directory tree (verify all dirs/files changed)
- Test symlink skipping (create symlink, verify not changed)
- Test error collection (create unwritable file, verify error collected)
- Test RecursivePermDialog step advancement
- Test RecursivePermDialog cancellation at each step

*Integration Tests:*
- Create directory tree with subdirectories and files
- Apply recursive change (dir: 755, file: 644)
- Verify all directories have 755, all files have 644
- Verify symlinks are untouched
- Test empty directory (no errors)

*Manual Testing:*
- [ ] Open PermissionDialog on directory
- [ ] Verify "Apply to" section displayed
- [ ] Toggle recursive option with Tab
- [ ] Select "Recursively" and press Enter
- [ ] Enter directory permission (755)
- [ ] Enter file permission (644)
- [ ] Verify all directories changed to 755
- [ ] Verify all files changed to 644
- [ ] Verify symlinks unchanged

**Acceptance Criteria:**
- [ ] Directories show "Apply to" section in PermissionDialog
- [ ] Tab key toggles between "this only" and "recursive" options
- [ ] j/k and Space keys navigate and select options
- [ ] Selecting "Recursively" shows two-step dialog
- [ ] First step accepts directory permission
- [ ] Second step accepts file permission
- [ ] Recursive operation changes all directories and files
- [ ] Symlinks are skipped without errors
- [ ] Esc at any step cancels entire operation
- [ ] Recursive operation processes at least 500 files/second

**Estimated Effort**: 中 (3-5 days)

**Risks and Mitigation:**
- **Risk**: Infinite loops with symlinks
  - **Mitigation**: Use os.Lstat to detect symlinks, never follow them
- **Risk**: Very large directory trees cause UI freeze
  - **Mitigation**: Phase 3 will add progress dialog for long operations

---

### Phase 3: Progress Display and Error Reporting

**Goal**: Provide visual progress feedback for long operations and comprehensive error reporting for failures.

**Files to Create:**

- `internal/ui/permission_progress_dialog.go` - Progress display dialog
- `internal/ui/permission_progress_dialog_test.go` - Progress dialog tests
- `internal/ui/permission_error_report_dialog.go` - Error report dialog
- `internal/ui/permission_error_report_dialog_test.go` - Error report tests

**Files to Modify:**

- `internal/fs/permissions.go`:
  - Add progress callback parameter to ChangePermissionRecursive
  - Call callback periodically during traversal
- `internal/ui/messages.go`:
  - Add permissionProgressUpdateMsg
  - Add permissionOperationCompleteMsg with error collection

**Key Components:**

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| PermissionProgressDialog | Display operation progress with percentage and current file | Operation running in background | Updates shown every 100ms, user can cancel |
| PermissionErrorReportDialog | Display complete error list with full scrolling support | Operation completed with errors | User views all errors, navigates with j/k/PgUp/PgDn, closes dialog |
| ScrollManager | Manage scroll offset and visible window | Error list exceeds visible area | Scroll position updated, scroll indicators shown |
| ProgressCallback | Report progress during recursive operation | Called from ChangePermissionRecursive | Progress message sent to UI |
| ErrorCollector | Collect failures during operation | Errors occur during permission changes | Error slice with paths and reasons |

**Processing Flow:**

```
1. User initiates recursive or batch operation
   └─ Show PermissionProgressDialog

2. Execute operation in background (goroutine)
   └─ Periodically call progress callback

3. Progress callback sends permissionProgressUpdateMsg
   ├─ Update processed count
   ├─ Update current file path
   ├─ Calculate percentage
   ├─ Calculate elapsed time
   └─ Estimate remaining time

4. PermissionProgressDialog receives updates
   ├─ Render progress bar (filled vs empty portions)
   ├─ Show current file (truncate if too long)
   ├─ Show elapsed time (MM:SS format)
   └─ Show remaining time estimate

5. User can cancel (Ctrl+C or Esc)
   ├─ Set cancellation flag
   ├─ Background operation checks flag
   └─ Stop early, report partial completion

6. Operation completes
   ├─ Send permissionOperationCompleteMsg
   ├─ Include success count and error list
   └─ Close PermissionProgressDialog

7. If errors occurred
   ├─ Show PermissionErrorReportDialog
   ├─ Display success/failure counts
   ├─ List ALL failed files with error reasons (no limit)
   ├─ Support full scrolling with j/k keys
   └─ Show scroll indicators if content exceeds visible area

8. User navigates error list
   ├─ j/k keys scroll through all errors
   ├─ Page Up/Down for faster scrolling
   └─ Scroll position indicator (e.g., "Line 15/234")

9. User closes error report (Enter or Esc)
   └─ Return to file list, refresh pane
```

**Implementation Steps:**

1. **Create PermissionProgressDialog Component**
   - Implement progress bar rendering (filled/empty blocks)
   - Display processed/total file counts
   - Show current file path (truncate if needed)
   - Display elapsed time and estimated remaining
   - Handle Esc/Ctrl+C for cancellation
   - Key considerations:
     - Reuse ArchiveProgressDialog patterns
     - Progress bar width: 50 characters
     - Update frequency: 100ms
     - Time estimation: linear extrapolation from current rate

2. **Create PermissionErrorReportDialog Component**
   - Display success and failure counts
   - Render fully scrollable error list (no artificial limits)
   - Show file path and error reason for each failure
   - Handle j/k for line-by-line scrolling
   - Handle Page Up/Down for page-by-page scrolling
   - Show scroll position indicator
   - Handle Enter or Esc to close
   - Key considerations:
     - Scroll offset management for arbitrarily long error lists
     - Calculate visible window size dynamically based on terminal height
     - Truncate paths to fit dialog width (show "..." prefix for long paths)
     - Categorize common errors (permission denied, not found, read-only filesystem)
     - Show scroll indicators (↑/↓ arrows) when more content available
     - Display current position (e.g., "Showing 10-20 of 234 errors")

3. **Add Progress Callback to Recursive Operation**
   - Modify ChangePermissionRecursive to accept callback function
   - Call callback every N files (e.g., every 10 files)
   - Pass current count, total count, and current file path
   - Check cancellation flag periodically
   - Key considerations:
     - Don't call callback too frequently (performance)
     - Use buffered channel for callback communication
     - Graceful cancellation (clean up before returning)

4. **Implement Background Operation Execution**
   - Execute ChangePermissionRecursive in goroutine
   - Send progress updates via Bubble Tea Cmd
   - Send completion message when done
   - Handle cancellation requests
   - Key considerations:
     - Use context.Context for cancellation
     - Ensure goroutine always completes (no leaks)
     - Error collection thread-safe (mutex if needed)

**Dependencies:**
- Requires: Phase 2 (recursive operations)
- Blocks: Phase 4 (batch operations need progress too)

**Testing Approach:**

*Unit Tests:*
- Test progress bar rendering at various percentages (0%, 50%, 100%)
- Test time formatting (seconds to MM:SS)
- Test remaining time calculation
- Test error report rendering with different error counts
- Test scrolling behavior in error dialog

*Integration Tests:*
- Run recursive operation on large directory (1000+ files)
- Verify progress updates received periodically
- Verify progress dialog displays correctly
- Cancel operation mid-way, verify early termination
- Create permission errors, verify error report shows all failures

*Manual Testing:*
- [ ] Start recursive operation on large directory
- [ ] Verify progress dialog appears
- [ ] Verify progress bar updates smoothly
- [ ] Verify current file path displayed
- [ ] Verify elapsed time increments
- [ ] Press Esc to cancel, verify operation stops
- [ ] Create permission errors (read-only file)
- [ ] Verify error report shows all failed files (no truncation)
- [ ] Test scrolling with j/k keys through entire error list
- [ ] Test Page Up/Down scrolling
- [ ] Verify scroll position indicator shows correct position
- [ ] Verify scroll indicators (↑/↓) appear when needed
- [ ] Create 100+ permission errors, verify all displayed
- [ ] Verify long file paths truncated with "..." prefix

**Acceptance Criteria:**
- [ ] Progress dialog shown for operations with 10+ files
- [ ] Progress bar updates every 100ms
- [ ] Percentage calculated correctly
- [ ] Current file path displayed (truncated if needed)
- [ ] Elapsed time shown in MM:SS format
- [ ] Remaining time estimated and displayed
- [ ] Esc or Ctrl+C cancels operation gracefully
- [ ] Error report dialog shown if any failures occurred
- [ ] Error report shows success and failure counts
- [ ] Error report lists ALL failed files with reasons (no artificial limit)
- [ ] j/k keys scroll through entire error list line-by-line
- [ ] Page Up/Down keys scroll by page (10-15 lines)
- [ ] Scroll position indicator shows current position (e.g., "10-20 of 234")
- [ ] Scroll indicators (↑/↓) appear when more content available above/below
- [ ] Long file paths truncated with "..." prefix to fit dialog width
- [ ] Error list handles 100+ errors smoothly without performance issues

**Estimated Effort**: 中 (3-5 days)

**Risks and Mitigation:**
- **Risk**: Goroutine leaks if operation doesn't complete
  - **Mitigation**: Use context.Context with timeout, ensure goroutine always returns
- **Risk**: Progress updates too frequent, causing UI lag
  - **Mitigation**: Throttle updates to 100ms intervals, batch file counts

---

### Phase 4: Batch Operations

**Goal**: Support permission changes on multiple selected files simultaneously.

**Files to Modify:**

- `internal/ui/permission_dialog.go`:
  - Detect when multiple files are marked
  - Show batch-specific dialog title and layout
  - Apply same permission to all marked items
- `internal/fs/permissions.go`:
  - Add ChangeBatchPermissions function
  - Handle mixed file/directory batches
- `internal/ui/model_update.go`:
  - Clear marks after successful batch operation

**Key Components:**

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ChangeBatchPermissions | Apply same permission to multiple paths | List of valid paths, valid permission mode | All paths changed, errors collected |
| BatchDetection | Determine if batch mode should be used | Check marked items in pane | Returns true if multiple items marked |
| MarkClearer | Clear all marks after successful operation | Batch operation completed successfully | All marked flags reset |
| BatchProgressTracker | Track progress across multiple files | Batch operation running | Updates sent for each file processed |

**Processing Flow:**

```
1. User marks multiple files (Space key)
   └─ Files added to marked set in pane

2. User presses Shift+P with marked files
   ├─ Detect batch mode (marked items > 0)
   └─ Show batch-specific PermissionDialog

3. PermissionDialog in batch mode
   ├─ Title shows count: "Permissions: N items"
   ├─ Current permission display omitted (multiple values)
   ├─ Show presets (file-specific, no recursive option)
   └─ User enters permission

4. User confirms batch operation
   ├─ Collect all marked paths
   ├─ Execute ChangeBatchPermissions
   └─ Show progress if count > 10

5. Execute ChangeBatchPermissions
   ├─ Iterate through each marked path
   ├─ Apply same permission to each
   ├─ Skip symlinks (no error)
   ├─ Skip directories in non-recursive mode
   ├─ Collect errors for failures
   └─ Report progress via callback

6. Show progress (if count > 10)
   └─ Reuse PermissionProgressDialog from Phase 3

7. Operation completes
   ├─ Clear all marks (if successful)
   ├─ Show error report if failures occurred
   └─ Refresh pane display

8. User views results
   └─ Return to file list with updated permissions
```

**Implementation Steps:**

1. **Add Batch Detection to Model Update Handler**
   - Check if marked items exist when Shift+P pressed
   - Determine appropriate dialog mode (single vs batch)
   - Pass marked item list to dialog creation
   - Key considerations:
     - Batch mode overrides recursive mode (no recursive for batch)
     - Count marked items to show in dialog title
     - Collect paths from marked items

2. **Modify PermissionDialog for Batch Mode**
   - Add batch-specific title rendering
   - Hide current permission display (multiple sources)
   - Hide recursive option entirely (batch is intentionally non-recursive only)
   - Use file-specific presets (even if directories included)
   - Key considerations:
     - Clear UI indication that this is batch mode
     - No recursive option shown (intentional design decision)
     - Show count of items being changed
     - User should use single-directory recursive mode for recursive changes

3. **Implement ChangeBatchPermissions Function**
   - Accept list of paths and single permission mode
   - Iterate through all paths
   - Apply permission to each file (skip symlinks, handle errors)
   - For directories: apply permission to directory itself only (non-recursive)
   - Call progress callback for each item
   - Collect and return errors
   - Key considerations:
     - Don't stop on first error, continue batch
     - Apply permission to directory itself, not contents (non-recursive by design)
     - Skip symlinks silently (use os.Lstat to detect)
     - Efficient iteration (no unnecessary stat calls)

4. **Add Mark Clearing Logic**
   - After successful batch operation, clear all marks
   - Send message to update pane state
   - Refresh pane display to show unmarked items
   - Key considerations:
     - Only clear marks on success (not on cancel)
     - Clear marks even if some items failed
     - Update pane rendering immediately

**Dependencies:**
- Requires: Phase 1 (basic dialog), Phase 3 (progress display for large batches)
- Blocks: None (final phase)

**Testing Approach:**

*Unit Tests:*
- Test batch mode detection (0 marked, 1 marked, multiple marked)
- Test ChangeBatchPermissions with mixed files/directories
- Test mark clearing after successful operation
- Test mark preservation after cancellation

*Integration Tests:*
- Mark 5 files, apply permission 644, verify all changed
- Mark 20 files, verify progress dialog shown
- Mark files with permission errors, verify error report
- Cancel batch operation, verify marks preserved

*Manual Testing:*
- [ ] Mark multiple files with Space
- [ ] Press Shift+P, verify batch dialog shown
- [ ] Verify title shows correct count ("Permissions: N items")
- [ ] Apply permission 644, verify all files changed
- [ ] Verify marks cleared after success
- [ ] Mark files including read-only, verify error report
- [ ] Cancel batch, verify marks preserved

**Acceptance Criteria:**
- [ ] Shift+P on marked items shows batch-specific dialog
- [ ] Dialog title shows item count ("Permissions: N items")
- [ ] Current permission display omitted in batch mode
- [ ] Recursive option completely hidden in batch mode (intentional)
- [ ] Same permission applied to all marked items
- [ ] Directories in batch: permission applied to directory itself only (non-recursive)
- [ ] Symlinks in batch are skipped silently
- [ ] Progress shown for batches with 10+ items
- [ ] Marks cleared after successful batch operation (even if some items failed)
- [ ] Marks preserved after cancellation
- [ ] Error report shown for partial failures with complete error list

**Estimated Effort**: 小 (1-2 days)

**Risks and Mitigation:**
- **Risk**: User expects recursive change for directories in batch
  - **Mitigation**: Clear UI indication that batch is non-recursive, suggest using recursive mode instead

---

## Complete File Structure

```
duofm/
├── cmd/duofm/
│   └── main.go                              # Entry point (no changes)
├── internal/
│   ├── fs/
│   │   ├── operations.go                    # Existing operations
│   │   ├── permissions.go                   # NEW: Permission operations
│   │   │   - ValidatePermissionMode         # Validate octal string
│   │   │   - ParsePermissionMode            # Convert string to FileMode
│   │   │   - FormatSymbolic                 # FileMode to symbolic string
│   │   │   - ChangePermission               # Single file/dir chmod
│   │   │   - ChangePermissionRecursive      # Recursive chmod with callback
│   │   │   - ChangeBatchPermissions         # Batch chmod
│   │   │   - PermissionError type           # Error collection structure
│   │   └── permissions_test.go              # NEW: Permission operation tests
│   └── ui/
│       ├── dialog.go                        # Existing Dialog interface
│       ├── keys.go                          # MODIFY: Add Shift+P keybinding
│       ├── messages.go                      # MODIFY: Add permission messages
│       │   - permissionOperationStartMsg
│       │   - permissionProgressUpdateMsg
│       │   - permissionOperationCompleteMsg
│       ├── model_update.go                  # MODIFY: Add Shift+P handler
│       ├── permission_dialog.go             # NEW: Basic permission dialog
│       │   - Handles single file/directory
│       │   - Shows current permission
│       │   - Real-time symbolic update
│       │   - Preset selection (1-4)
│       │   - Recursive option toggle (directories)
│       ├── permission_dialog_test.go        # NEW: Permission dialog tests
│       ├── recursive_perm_dialog.go         # NEW: Two-step recursive dialog
│       │   - Step 1: Directory permission input
│       │   - Step 2: File permission input
│       │   - Shows previously entered values
│       ├── recursive_perm_dialog_test.go    # NEW: Recursive dialog tests
│       ├── permission_progress_dialog.go    # NEW: Progress display
│       │   - Progress bar rendering
│       │   - Current file display
│       │   - Time tracking (elapsed/remaining)
│       │   - Cancellation support
│       ├── permission_progress_dialog_test.go # NEW: Progress dialog tests
│       ├── permission_error_report_dialog.go # NEW: Error report dialog
│       │   - Success/failure counts
│       │   - Fully scrollable error list (no limits)
│       │   - j/k and Page Up/Down scrolling
│       │   - Scroll position indicator
│       │   - File paths and error reasons
│       └── permission_error_report_dialog_test.go # NEW: Error report tests
├── tests/
│   ├── e2e/
│   │   └── permission_test.sh               # NEW: E2E tests for permission feature
│   ├── fs/
│   │   └── permissions_test.go              # Integration tests for fs operations
│   └── ui/
│       └── permission_dialogs_test.go       # Integration tests for dialogs
└── doc/
    └── tasks/
        └── permission-edit/
            ├── SPEC.md                      # Feature specification
            └── IMPLEMENTATION.md            # This document
```

**File Descriptions:**

**Core Permission Operations (`internal/fs/permissions.go`):**
- Validation and parsing of octal permission strings
- Symbolic notation formatting for display
- Single-file permission changes
- Recursive directory tree permission changes
- Batch permission changes for multiple files
- Error collection and reporting structures

**Permission Dialog (`internal/ui/permission_dialog.go`):**
- Basic permission change UI for files and directories
- Real-time symbolic notation updates
- Preset selection and management
- Recursive option toggle (directories only)
- Input validation and error display

**Recursive Dialog (`internal/ui/recursive_perm_dialog.go`):**
- Two-step input flow (directories then files)
- Context-sensitive presets per step
- Step progression and cancellation handling

**Progress Dialog (`internal/ui/permission_progress_dialog.go`):**
- Visual progress bar rendering
- File count and percentage display
- Time tracking and estimation
- Cancellation support

**Error Report Dialog (`internal/ui/permission_error_report_dialog.go`):**
- Success and failure statistics
- Complete error listing with file paths and reasons (no artificial limits)
- Full scrolling support with j/k (line-by-line) and Page Up/Down (page-by-page)
- Dynamic scroll position indicator (e.g., "Showing 10-20 of 234 errors")
- Scroll indicators (↑/↓) when more content available
- Efficient rendering for large error lists (100+ errors)

**Relationships Between Files:**
- `permission_dialog.go` depends on `permissions.go` for validation and formatting
- `recursive_perm_dialog.go` depends on `permission_dialog.go` for shared input logic
- `permission_progress_dialog.go` receives updates from background operations
- `permission_error_report_dialog.go` displays results from `permissions.go` error collection
- `model_update.go` orchestrates dialog lifecycle and operation execution
- `messages.go` defines communication protocol between components

## Testing Strategy

### Unit Testing

**Approach:**
- Use Go's built-in `testing` package with table-driven tests
- Test business logic separately from UI rendering
- Mock file system operations where appropriate
- Focus on edge cases and error conditions

**Test Coverage Goals:**
- Permission validation and parsing: 90%+ (critical for security)
- File operations: 85%+ (critical for data integrity)
- Dialog logic: 70%+ (UI has some visual-only code)
- Overall project: 80%+

**Key Test Areas:**

1. **Permission Validation (`internal/fs/permissions_test.go`)**

| Test Case | Input | Expected Output |
|-----------|-------|-----------------|
| Valid three-digit octal | "644" | nil (no error) |
| Valid boundary low | "000" | nil |
| Valid boundary high | "777" | nil |
| Invalid digit | "888" | error: "digit out of range" |
| Too short | "64" | error: "must be 3 digits" |
| Too long | "6440" | error: "must be 3 digits" |
| Non-numeric | "abc" | error: "must be numeric" |
| Empty string | "" | error: "cannot be empty" |

2. **Symbolic Formatting (`internal/fs/permissions_test.go`)**

| FileMode | IsDir | Expected Output |
|----------|-------|-----------------|
| 0644 | false | "-rw-r--r--" |
| 0755 | true | "drwxr-xr-x" |
| 0777 | false | "-rwxrwxrwx" |
| 0000 | false | "----------" |
| 0700 | true | "drwx------" |

3. **Recursive Operations (`internal/fs/permissions_test.go`)**
- Create temporary directory tree with known structure
- Apply recursive permission change (dirMode: 0755, fileMode: 0644)
- Verify all directories changed to 0755
- Verify all files changed to 0644
- Verify symlinks unchanged
- Test error collection when some files fail

4. **Dialog Behavior (`internal/ui/permission_dialog_test.go`)**
- Test preset selection updates input field
- Test real-time symbolic update on keystroke
- Test validation error display
- Test recursive option toggle (Tab key)
- Test dialog cancellation (Esc)

### Integration Testing

**Scenarios:**

1. **End-to-End Single File Change**
   - Setup: Create test file with 0644 permission
   - Action: Open dialog, enter 0755, confirm
   - Verify: File permission is 0755
   - Verify: Dialog closes, status message shown
   - Cleanup: Remove test file

2. **End-to-End Recursive Change**
   - Setup: Create directory tree (3 dirs, 5 files, 1 symlink)
   - Action: Open dialog on root dir, select recursive, enter dir=0755 file=0644
   - Verify: All directories are 0755
   - Verify: All files are 0644
   - Verify: Symlink unchanged
   - Cleanup: Remove test tree

3. **Error Handling Flow**
   - Setup: Create read-only file (chmod 0444), change owner to root
   - Action: Attempt to change permission to 0644
   - Verify: Error dialog shown with "Permission denied"
   - Verify: Application remains functional
   - Cleanup: Restore ownership, remove file

4. **Batch Operation Flow**
   - Setup: Create 5 test files with mixed permissions
   - Action: Mark all files, press Shift+P, enter 0644
   - Verify: All files changed to 0644
   - Verify: Marks cleared
   - Cleanup: Remove test files

### Manual Testing Checklist

Based on spec test scenarios:

**Basic Functionality:**
- [ ] Press Shift+P on file, verify dialog opens with current permission
- [ ] Enter 3-digit permission, verify symbolic notation updates in real-time
- [ ] Press Enter, verify permission changed and dialog closes
- [ ] Press Esc in dialog, verify cancellation (no changes)
- [ ] Try Shift+P on ".." entry, verify error shown

**Preset Functionality:**
- [ ] Press preset key 1 on file, verify 644 auto-filled
- [ ] Press preset key 2 on file, verify 755 auto-filled
- [ ] Press preset key 1 on directory, verify 755 auto-filled (dir preset)
- [ ] Modify preset value before applying, verify change accepted

**Recursive Functionality:**
- [ ] Open dialog on directory, verify "Apply to" section shown
- [ ] Press Tab, verify toggle between "This only" and "Recursively"
- [ ] Select "Recursively", press Enter, verify two-step dialog
- [ ] Enter dir permission 755 in step 1, verify step 2 shown
- [ ] Enter file permission 644 in step 2, verify operation executes
- [ ] Verify all subdirectories changed to 755
- [ ] Verify all files in tree changed to 644

**Progress and Error Reporting:**
- [ ] Start recursive operation on large tree (100+ files)
- [ ] Verify progress dialog shown
- [ ] Verify progress bar updates smoothly
- [ ] Press Esc during progress, verify cancellation
- [ ] Create permission-denied scenario, verify error report shown
- [ ] Scroll through error list with j/k keys

**Batch Operations:**
- [ ] Mark 3 files with Space key
- [ ] Press Shift+P, verify batch dialog shown with count
- [ ] Verify recursive option is NOT shown (batch is non-recursive only)
- [ ] Apply permission 644, verify all marked files changed
- [ ] Verify marks cleared after success
- [ ] Cancel batch operation, verify marks preserved
- [ ] Mark mix of files and directories, verify only files changed (dirs skipped in non-recursive mode)

**Edge Cases:**
- [ ] Try permission "000" (all disabled), verify accepted
- [ ] Try permission "777" (all enabled), verify accepted
- [ ] Try permission "888", verify error shown
- [ ] Try permission "12" (too short), verify error shown
- [ ] Create symlink, verify skipped in recursive mode
- [ ] Empty directory recursive change, verify no errors

## Dependencies

### External Dependencies

| Package | Version | Purpose | Installation |
|---------|---------|---------|--------------|
| github.com/charmbracelet/bubbletea | v0.25.0+ | TUI framework | Already in go.mod |
| github.com/charmbracelet/lipgloss | v0.9.1+ | Terminal styling | Already in go.mod |

### Internal Dependencies

**Implementation Order (respecting dependencies):**

1. **Phase 1** - No dependencies
   - `internal/fs/permissions.go` (core operations)
   - `internal/ui/permission_dialog.go` (basic dialog)
   - `internal/ui/messages.go` modifications
   - `internal/ui/model_update.go` handlers

2. **Phase 2** - Depends on Phase 1
   - `internal/ui/recursive_perm_dialog.go`
   - `internal/fs/permissions.go` (add recursive function)
   - `internal/ui/permission_dialog.go` (add recursive option)

3. **Phase 3** - Depends on Phase 2
   - `internal/ui/permission_progress_dialog.go`
   - `internal/ui/permission_error_report_dialog.go`
   - `internal/fs/permissions.go` (add callback support)

4. **Phase 4** - Depends on Phase 1 and Phase 3
   - `internal/ui/permission_dialog.go` (add batch mode)
   - `internal/fs/permissions.go` (add batch function)
   - `internal/ui/model_update.go` (mark clearing)

**Component Dependencies:**

- `permission_dialog.go` → `permissions.go` (validation, formatting)
- `recursive_perm_dialog.go` → `permission_dialog.go` (shared input patterns)
- `permission_progress_dialog.go` → `messages.go` (progress updates)
- `permission_error_report_dialog.go` → `permissions.go` (error structures)
- `model_update.go` → All dialogs (lifecycle management)

## Risk Assessment

### Technical Risks

1. **Performance with Large Directory Trees**
   - **Risk**: Recursive operations on 10,000+ files may cause UI lag or timeouts
   - **Likelihood**: Medium (users may have large project directories)
   - **Impact**: High (poor user experience, potential application hang)
   - **Mitigation**:
     - Execute recursive operations in background goroutines
     - Throttle progress updates to 100ms intervals (not per-file)
     - Implement cancellation via context.Context
     - Test with large directory trees (10,000+ files) during development
     - Add timeout detection (warn user if operation takes > 1 minute)

2. **Permission Denied Errors**
   - **Risk**: Users may attempt to change permissions on files they don't own
   - **Likelihood**: High (common scenario, especially on shared systems)
   - **Impact**: Medium (operation fails, user frustrated)
   - **Mitigation**:
     - Clear error messages explaining permission denied
     - Continue operation on other files (don't fail entire batch)
     - Show detailed error report with failed files
     - Suggest solutions (e.g., "Use sudo" or "Contact administrator")

3. **Symlink Handling Complexity**
   - **Risk**: Following symlinks could cause infinite loops or unexpected behavior
   - **Likelihood**: Low (symlinks are less common, but do exist)
   - **Impact**: High (infinite loop, application hang)
   - **Mitigation**:
     - Always use `os.Lstat` (not `os.Stat`) to detect symlinks
     - Never follow symlinks in recursive operations
     - Skip symlinks silently (not counted as errors)
     - Add unit tests with circular symlink scenarios

4. **Cross-Platform Compatibility**
   - **Risk**: Permission model differs on non-Unix systems (Windows)
   - **Likelihood**: Low (duofm likely targets Unix-like systems)
   - **Impact**: Medium (feature broken on Windows)
   - **Mitigation**:
     - Document Unix-only behavior in feature documentation
     - Consider adding platform check and disabling feature on Windows
     - Investigate Windows ACL support for future enhancement

### Implementation Risks

1. **Scope Creep**
   - **Risk**: Adding features beyond spec (e.g., symbolic mode, setuid support)
   - **Likelihood**: Medium (permission features have many possibilities)
   - **Impact**: Medium (delayed timeline, increased complexity)
   - **Mitigation**:
     - Strictly follow SPEC.md requirements
     - Document "Future Enhancements" separately
     - Review implementation plan against spec before coding

2. **Underestimated Complexity**
   - **Risk**: Dialog state management more complex than expected
   - **Likelihood**: Medium (two-step flow, progress tracking can be tricky)
   - **Impact**: Medium (delayed timeline, potential bugs)
   - **Mitigation**:
     - Study existing ArchiveProgressDialog implementation closely
     - Start with simplest case (single file) and add complexity gradually
     - Add buffer time to estimates (3-5 days per phase)

3. **Testing Challenges**
   - **Risk**: Hard to test permission errors without root access or special setup
   - **Likelihood**: Medium (CI/CD may not support permission scenarios)
   - **Impact**: Low (manual testing can cover gaps)
   - **Mitigation**:
     - Use Docker containers for permission testing
     - Create test files with specific owners/permissions in setup
     - Document manual testing procedures clearly

## Performance Considerations

### Performance Goals

- **Dialog Display**: < 100ms from Shift+P to visible dialog
- **Single File Change**: < 50ms to complete chmod
- **Recursive Operation**: > 500 files/second on HDD (target: 1000 files/second on SSD)
- **Progress Update**: Every 100ms (not more frequent to avoid UI lag)
- **Real-time Symbolic Update**: < 10ms from keystroke to symbolic notation update

### Optimization Strategies

1. **Efficient Directory Traversal**
   - Use `filepath.Walk` with early termination support
   - Avoid unnecessary `os.Stat` calls (reuse `WalkFunc` FileInfo)
   - Use `os.Lstat` only when symlink detection needed
   - Pre-allocate error slice with reasonable capacity (e.g., 100)

2. **Throttled Progress Updates**
   - Update UI every 100ms, not per file
   - Use time-based throttling, not count-based
   - Batch progress updates in goroutine communication
   - Reduce string allocations in progress messages

3. **Input Validation Optimization**
   - Validate each digit on keystroke (prevent invalid state)
   - Cache symbolic notation calculation (only recalculate on input change)
   - Avoid regex for simple octal validation (use simple range checks)

4. **UI Rendering Optimization**
   - Reuse lipgloss.Style objects (don't recreate per render)
   - Minimize string concatenation (use strings.Builder)
   - Truncate long file paths efficiently (substring, not repeated calculations)

### Profiling Plan

If performance issues arise:
1. Use `go test -bench` for micro-benchmarks (validation, formatting)
2. Use `pprof` for CPU profiling during recursive operations
3. Test on real-world large directories (e.g., node_modules, Linux kernel source)
4. Monitor memory allocations (look for unnecessary slice growth)

## Security Considerations

### Permission Validation

- **Octal Range Enforcement**: Validate each digit is 0-7 before parsing
- **Length Validation**: Reject inputs shorter or longer than 3 digits
- **No Code Injection**: Simple integer validation, no eval or shell execution
- **Error Message Safety**: Don't leak system paths in error messages

### Path Handling

- **Path Traversal Prevention**: Use `filepath.Clean` to normalize paths
- **Symlink Safety**: Never follow symlinks in recursive operations
- **Directory Boundary Enforcement**: Prevent operations outside current pane directory
- **Parent Directory Protection**: Explicitly block Shift+P on ".." entry

### Operation Safety

- **No Privilege Escalation**: Operations run with user's current permissions
- **No Sudo Automation**: Never automatically invoke sudo or elevated privileges
- **Atomic Operations**: Use `os.Chmod` (atomic) not multi-step permission changes
- **Error Isolation**: Individual file failures don't affect other files in batch

### User Input Sanitization

- **Numeric-Only Input**: Only accept digits 0-7 and control keys
- **Length Limits**: Input field limited to 3 characters
- **No Special Characters**: Reject any non-numeric input immediately
- **Validation Before Execution**: Always validate before calling `os.Chmod`

## Design Decisions

### Confirmed Decisions

The following design decisions have been confirmed and will be implemented as specified:

1. **Symbolic Mode Support**
   - **Decision**: Initial version supports numeric mode only (000-777 octal notation)
   - **Rationale**: Numeric mode is simpler to implement and validate; symbolic mode ("u+x", "go-w") deferred to future enhancement
   - **Note**: Symbolic *display* notation (e.g., "-rw-r--r--") is implemented for real-time feedback

2. **Symlink Handling**
   - **Decision**: Symlinks are always skipped (not changed) in all operations
   - **Rationale**: Changing symlink permissions is rarely needed and can be complex; skipping is safer
   - **Implementation**: Use `os.Lstat` to detect symlinks, skip without error

3. **Batch Operation Recursive Mode**
   - **Decision**: Batch operations are intentionally non-recursive only
   - **Rationale**: Batch mode is for quick changes to multiple files; recursive changes should use single-directory recursive mode
   - **UI Behavior**: Recursive option not shown in batch mode dialog

4. **Progress Display Threshold**
   - **Decision**: Show progress dialog for operations with 10+ files or expected duration > 1 second
   - **Implementation**: Define constant `ProgressThreshold = 10` in permissions.go
   - **Rationale**: Balance between showing progress for long operations and avoiding dialog flicker for quick operations

5. **Cancellation Behavior**
   - **Decision**: Cancelled operations keep partial changes (no rollback)
   - **Rationale**: chmod is idempotent, rollback adds complexity and may fail on permission errors
   - **User Impact**: User can manually revert changes if needed

6. **Error Report Display**
   - **Decision**: Display ALL errors in scrollable dialog (no artificial limit)
   - **Implementation**: PermissionErrorReportDialog supports full scrolling with j/k keys
   - **UI Design**: Show scroll indicators if content exceeds visible area
   - **Rationale**: Complete error visibility more important than dialog size; users need to see all failures

7. **Special Permissions UI**
   - **Decision**: Numeric input only (no dedicated UI for setuid/setgid/sticky bit)
   - **Rationale**: Special permissions are less commonly used; numeric input supports them (e.g., "4755" for setuid)

8. **Hidden File Handling**
   - **Decision**: Recursive operations change all files regardless of hidden file filter settings
   - **Rationale**: Permission changes should be exhaustive; filter is for display only, not operations

## Future Enhancements

Items deferred to later phases or releases:

### Phase 2 Features (from spec "Open Questions")

- **Undo Functionality**: Save previous permissions and allow undo
  - Complexity: Medium (requires state tracking)
  - Value: High (safety feature)
  - Implementation: Store old permissions in memory, add Ctrl+Z handler

- **Symbolic Mode Input**: Support "u+x", "go-w", "a=rx" notation
  - Complexity: High (complex parsing, edge cases)
  - Value: Medium (power users prefer symbolic)
  - Implementation: Add parser for symbolic notation, convert to numeric

- **Ownership Change**: Add chown functionality (change owner/group)
  - Complexity: Medium (similar to chmod, but requires root often)
  - Value: Medium (useful but less common)
  - Implementation: New dialog with user/group selection

### Phase 3 Features

- **Permission Preview in File List**: Show permissions in status bar or column
  - Complexity: Low (already available in FileEntry)
  - Value: Medium (informational)
  - Implementation: Add permission column to file list view

- **Permission Presets Configuration**: Allow users to customize preset values
  - Complexity: Medium (requires config file changes)
  - Value: Low (defaults cover most cases)
  - Implementation: Add presets section to config.toml

### Not in Current Spec

- **ACL Support**: Extended Access Control Lists (setfacl/getfacl)
  - Complexity: High (complex permission model)
  - Value: Low (advanced feature, rarely used)

- **SELinux Context**: Change SELinux security contexts
  - Complexity: Very High (SELinux-specific)
  - Value: Very Low (niche use case)

- **Permission Templates**: Save/load permission sets
  - Complexity: Medium
  - Value: Low (manual application is fast enough)

## Success Metrics

### Functional Completeness

- [ ] All FR1-FR10 requirements implemented
- [ ] All user stories (US1-US4) fulfilled
- [ ] All test scenarios pass (unit, integration, E2E)
- [ ] Error handling works for all identified error cases
- [ ] Keyboard shortcuts documented and functional

### Quality Metrics

- [ ] Unit test coverage ≥ 80% (overall), ≥ 90% (permissions.go)
- [ ] All integration tests pass
- [ ] All manual test checklist items verified
- [ ] No critical or high-priority bugs in issue tracker
- [ ] Code review completed and approved

### Performance Metrics

- [ ] Dialog display time < 100ms (measured with `time` command)
- [ ] Single file change < 50ms (measured in unit tests)
- [ ] Recursive operation ≥ 500 files/second on HDD (measured with test tree)
- [ ] Progress updates every 100ms ± 10ms (measured with instrumentation)
- [ ] Real-time symbolic update < 10ms (measured in UI tests)

### User Experience

- [ ] Intuitive keyboard navigation (no confusion in user testing)
- [ ] Clear error messages (users understand what went wrong)
- [ ] Help text is accurate and comprehensive
- [ ] No unexpected behavior or surprises
- [ ] Feature feels consistent with rest of duofm

### Documentation

- [ ] SPEC.md accurately reflects implemented behavior
- [ ] Code comments explain complex logic (especially recursive traversal)
- [ ] Test documentation describes what each test verifies
- [ ] README.md or user guide mentions permission feature
- [ ] Known limitations documented (e.g., Unix-only)

## References

- **Specification**: `doc/tasks/permission-edit/SPEC.md`
- **Go os.Chmod Documentation**: https://pkg.go.dev/os#Chmod
- **Go filepath.Walk Documentation**: https://pkg.go.dev/path/filepath#Walk
- **Unix chmod Manual**: `man 1 chmod`, `man 2 chmod`
- **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
- **Bubble Tea Examples**: https://github.com/charmbracelet/bubbletea/tree/master/examples
- **Lip Gloss Documentation**: https://github.com/charmbracelet/lipgloss
- **Existing Dialog Implementations**:
  - `internal/ui/input_dialog.go` (basic input pattern)
  - `internal/ui/archive_progress_dialog.go` (progress tracking pattern)
  - `internal/ui/confirm_dialog.go` (simple dialog pattern)
  - `internal/ui/context_menu_dialog.go` (option selection pattern)
- **File Operation Patterns**: `internal/fs/operations.go`
- **Message Architecture**: `internal/ui/messages.go`
- **duofm Project README**: `/home/sakura/cache/worktrees/feat-permission-edit/README.md`
- **duofm Contributing Guide**: `/home/sakura/cache/worktrees/feat-permission-edit/doc/CONTRIBUTING.md`

## Next Steps

**✅ Design Decisions Confirmed**

All open questions have been resolved and design decisions confirmed (see "Design Decisions" section above). Implementation can proceed with the following plan:

1. **Implementation Ready**
   - All design decisions confirmed and documented
   - Plan reviewed against SPEC.md for completeness
   - Phased approach and timeline estimates confirmed
   - Ready to begin Phase 1 implementation

2. **Environment Setup**
   - Verify Go 1.21+ installed
   - Ensure Bubble Tea and Lip Gloss dependencies up to date
   - Prepare test environment (ability to create temporary files/dirs)
   - Setup test data (large directory trees for performance testing)

3. **Begin Phase 1 Implementation**
   - Create `internal/fs/permissions.go` with validation functions
   - Write unit tests first (TDD approach)
   - Implement basic PermissionDialog
   - Add Shift+P keybinding and message handlers
   - Test single file permission changes end-to-end

4. **Continuous Integration**
   - Run tests after each component completion
   - Commit incrementally with descriptive messages
   - Perform code review at end of each phase
   - Update documentation as implementation progresses

5. **Phase Completion Checklist**
   - Complete all acceptance criteria for phase
   - All tests passing (unit + integration)
   - Code reviewed and refactored if needed
   - Documentation updated (comments, README)
   - Demo phase to stakeholders before proceeding

**Recommended Implementation Order:**
Phase 1 → Phase 2 → Phase 3 → Phase 4

**Estimated Total Timeline:**
- Phase 1: 3-5 days
- Phase 2: 3-5 days
- Phase 3: 3-5 days
- Phase 4: 1-2 days
- **Total**: 10-17 days (2-3.5 weeks)

Buffer for testing, code review, and refinement: +3-5 days

**Final Estimated Timeline**: 3-4 weeks
