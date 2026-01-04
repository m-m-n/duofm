# Implementation Plan: Context Menu Open File

## Overview

Add "Open" and "Open with ..." menu items to duofm's context menu, enabling users to launch files and directories with system default applications (via xdg-open) or custom applications. This bridges duofm's terminal interface with desktop GUI applications.

## Objectives

- Integrate Open and Open with menu items into existing context menu
- Enable single-file opening with system default applications via xdg-open
- Support custom application selection with user-editable options
- Handle multiple marked files for batch operations with custom applications
- Maintain duofm's responsiveness through background process execution
- Implement proper error handling for missing tools and failed launches

## Prerequisites

### Development Environment

- Go 1.21 or later
- xdg-utils package installed (for xdg-open and xdg-mime commands)
- Linux desktop environment with X11 or Wayland

### Dependencies

**External Runtime Dependencies:**
- xdg-open (from xdg-utils package) - required for "Open" functionality
- xdg-mime (from xdg-utils package) - optional for default app detection

**Internal Dependencies:**
- `internal/ui/context_menu_dialog.go` - existing context menu implementation
- `internal/ui/dialog.go` - Dialog interface
- `internal/ui/dialog_base.go` - BaseDialog embedding
- `internal/ui/text_input.go` - TextInput component for application field
- `internal/ui/messages.go` - message type definitions
- `internal/ui/exec.go` - existing process execution patterns
- `internal/ui/model_update.go` - message routing

### Knowledge Requirements

- Bubble Tea architecture (Model-Update-View pattern)
- Go's os/exec package for process management
- xdg-utils command-line interface
- Dialog pattern used throughout duofm codebase
- Message-based communication in Bubble Tea

## Architecture Overview

### Technology Stack

- **Language**: Go 1.21+
- **TUI Framework**: Bubble Tea v0.25.0
- **Styling**: Lip Gloss v0.9.1
- **External Tools**: xdg-open, xdg-mime (Linux standard)

### Design Approach

**Pattern Consistency**: Follow established duofm patterns:
- Dialog implementation extends BaseDialog
- Message-based communication between components
- Background process execution using tea.Cmd
- Error reporting through status messages

**Responsibility Separation**:
- Context menu manages item display and selection
- Open with dialog handles user input for custom applications
- Exec functions launch processes in background
- Model orchestrates message flow and error handling

**Security-First Design**:
- Use exec.Command with separate arguments (no shell interpolation)
- Sanitize file paths with filepath.Clean
- Pass filenames as separate arguments (exec.Command handles escaping automatically)
- Quote filenames for display purposes only (not for execution)
- No privilege escalation

### Component Interaction

```
User Input (@) → ContextMenuDialog
                      ↓
                 Select "Open" (1)
                      ↓
                 openWithXDGCmd → Background Process
                      ↓
                 openWithFinishedMsg → Model → Status Update

User Input (@) → ContextMenuDialog
                      ↓
                 Select "Open with ..." (2)
                      ↓
                 OpenWithDialog (query default app)
                      ↓
                 User edits application field
                      ↓
                 openWithDialogResultMsg → Model
                      ↓
                 openWithCustomCmd → Background Process
                      ↓
                 openWithFinishedMsg → Model → Status Update
```

## Implementation Phases

### Phase 1: Basic Open with xdg-open

**Goal**: Implement "Open" menu item that launches xdg-open for single files and directories.

**Files to Create**:
- None (modifications only)

**Files to Modify**:
- `internal/ui/context_menu_dialog.go`:
  - Add "Open" menu item at position 1 in buildMenuItems()
  - Enable only when no files are marked (markCount == 0)
  - Set action to trigger openWithXDGMsg

- `internal/ui/messages.go`:
  - Add openWithXDGMsg type with file path and working directory
  - Add openWithFinishedMsg type with error field

- `internal/ui/exec.go`:
  - Add openWithXDG() function to launch xdg-open in background
  - Use exec.Command("xdg-open", filename) with cmd.Start()
  - Set working directory to current pane path
  - Return tea.Cmd that sends openWithFinishedMsg

- `internal/ui/model_update.go`:
  - Add handler for openWithXDGMsg in handleCustomMessages()
  - Add handler for openWithFinishedMsg to display status or error

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ContextMenuDialog.buildMenuItems() | Add "Open" menu item to items list | Entry and paths provided | "Open" item at position 1, enabled if markCount == 0 |
| openWithXDG() | Launch xdg-open with file in background | Valid file path and workDir | Process started, cmd returned |
| openWithXDGMsg handler | Invoke openWithXDG and return cmd | Active pane has selected entry | Background process launched |
| openWithFinishedMsg handler | Update status with success or error | openWithFinishedMsg received | Status message displayed for 5 seconds |

**Processing Flow**:

```
1. User presses @ → Context menu displays with items
2. Context menu includes "Open" at position 1
   ├─ If markCount > 0 → "Open" is disabled
   └─ If markCount == 0 → "Open" is enabled
3. User selects "Open" (key 1 or Enter)
4. Context menu closes and sends openWithXDGMsg
5. Model receives openWithXDGMsg
   ├─ Extract file path from active pane selected entry
   ├─ Get working directory from active pane path
   └─ Return openWithXDG() cmd
6. openWithXDG() executes
   ├─ Build command: xdg-open with filename as separate argument
   ├─ Set working directory
   ├─ Call cmd.Start() (non-blocking)
   └─ Return openWithFinishedMsg with error or nil
7. Model receives openWithFinishedMsg
   ├─ If error → Display error in status bar
   └─ If success → Display "Opened with xdg-open" in status bar
8. Status message clears after 5 seconds
```

**Implementation Steps**:

1. **Add "Open" menu item to context menu**
   - Modify buildMenuItems() to insert "Open" at position 1
   - Determine enabled state based on markCount
   - Set action ID to "open" for message routing

2. **Define message types**
   - Create openWithXDGMsg struct with file and workDir fields
   - Create openWithFinishedMsg struct with error field

3. **Implement background execution function**
   - Create openWithXDG() in exec.go
   - Use exec.Command with separate arguments
   - Set working directory
   - Start process without waiting
   - Return tea.Cmd wrapping result message

4. **Wire message handlers in Model**
   - Add openWithXDGMsg handler in handleContextMenuResult()
   - Add openWithFinishedMsg handler in handleCustomMessages()
   - Display appropriate status messages

**Dependencies**:
- Requires: Existing context menu and message infrastructure
- Blocks: Phase 2 (Open with custom app depends on this pattern)

**Testing Approach**:

*Unit Tests*:
- Test "Open" menu item appears at position 1
- Test "Open" enabled when markCount == 0
- Test "Open" disabled when markCount > 0
- Test openWithXDG() builds correct command
- Test openWithXDG() sets working directory
- Mock exec.Command to verify arguments

*Integration Tests*:
- Test end-to-end: Select "Open" → xdg-open invoked
- Test error handling: xdg-open not found
- Test with file path
- Test with directory path

*Manual Testing*:
- [ ] Open file with xdg-open launches default application
- [ ] Open directory launches file manager
- [ ] duofm remains responsive after launch
- [ ] Error displays if xdg-open missing
- [ ] "Open" is disabled when multiple files marked

**Acceptance Criteria**:
- [ ] "Open" menu item appears at position 1
- [ ] Single file opens with xdg-open
- [ ] Single directory opens with xdg-open
- [ ] Process runs in background (duofm not blocked)
- [ ] Error message shown if xdg-open unavailable
- [ ] "Open" disabled when multiple files marked

**Estimated Effort**: 小 (1-2 days)

**Risks and Mitigation**:
- **Risk**: xdg-open may not be available on all systems
  - **Mitigation**: Clear error message, graceful degradation, document requirement
- **Risk**: Background process may fail silently
  - **Mitigation**: Check cmd.Start() error, report to user

---

### Phase 2: Open With Dialog Implementation

**Goal**: Create OpenWithDialog to accept custom application input, display file list, and launch custom applications.

**Files to Create**:
- `internal/ui/open_with_dialog.go` - OpenWithDialog implementation
- `internal/ui/open_with_dialog_test.go` - Unit tests for OpenWithDialog

**Files to Modify**:
- `internal/ui/context_menu_dialog.go`:
  - Add "Open with ..." menu item at position 2
  - Always enabled (works with single or multiple files)
  - Set action to trigger openWithDialog creation

- `internal/ui/messages.go`:
  - Add openWithDialogResultMsg with application, files, workDir, cancelled

- `internal/ui/exec.go`:
  - Add openWithCustom() function to launch custom application
  - Parse application string (may contain options)
  - Pass multiple files as separate arguments

- `internal/ui/model_update.go`:
  - Add handler to create OpenWithDialog
  - Add handler for openWithDialogResultMsg to launch custom command

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| OpenWithDialog | Display application input and file list | Files and workDir provided | User input or cancellation |
| OpenWithDialog.Init() | Populate file list for display | Marked files or selected entry | filesDisplay string formatted |
| OpenWithDialog.Update() | Handle keyboard input for editing | Active dialog | Input updated or dialog closed with result |
| OpenWithDialog.View() | Render dialog with input field and file list | Active dialog | Dialog UI string returned |
| openWithCustom() | Launch application with files as arguments | Application string and file list | Process started, cmd returned |
| openWithDialogResultMsg handler | Execute openWithCustom if not cancelled | Result message received | Background process launched or no-op |

**Processing Flow**:

```
1. User selects "Open with ..." from context menu
2. Model creates OpenWithDialog
   ├─ Collect marked files or use selected entry
   ├─ Get working directory from active pane
   ├─ Format file list for display (quote for display only, truncate if long)
   └─ Set default application to empty (Phase 3 adds detection)
3. Dialog displays with:
   ├─ Title: "Open with Application"
   ├─ Application input field (editable with cursor)
   ├─ Files field (read-only, shows filenames quoted for display only)
   └─ Footer: keyboard hints
4. User interacts with dialog
   ├─ Edit application field (Emacs keybindings, horizontal scroll)
   ├─ Press Enter → Confirm
   └─ Press Esc → Cancel
5. On Enter with non-empty application:
   ├─ Close dialog
   ├─ Send openWithDialogResultMsg(application, files, workDir, false)
   └─ Model receives message
6. Model handler executes:
   ├─ Parse application string (command + options)
   ├─ Call openWithCustom(application, files, workDir)
   └─ Return tea.Cmd for background execution
7. openWithCustom() executes:
   ├─ Split application into command and options
   ├─ Build argument list: options + files (as separate arguments)
   ├─ exec.Command(command, args...)
   ├─ Set working directory
   ├─ cmd.Start()
   └─ Return openWithFinishedMsg
8. Status message displays result
```

**Implementation Steps**:

1. **Create OpenWithDialog structure**
   - Embed BaseDialog
   - Add TextInput for application field
   - Add fileList []string and filesDisplay string
   - Add workDir string
   - Add DialogStyles

2. **Implement dialog lifecycle methods**
   - Update() handles keyboard input (Enter, Esc, editing)
   - View() renders title, input field, file list, footer
   - Init logic populates filesDisplay with formatted file list

3. **Implement file list formatting for display**
   - Quote each filename with double quotes (for display purposes only)
   - Join with spaces
   - Truncate with "..." if exceeds width limit
   - Note: Actual execution uses unquoted arguments passed to exec.Command

4. **Add "Open with ..." menu item**
   - Insert at position 2 in buildMenuItems()
   - Always enabled (Enabled: true)
   - Action sends message to create OpenWithDialog

5. **Implement openWithCustom() function**
   - Parse application string with strings.Fields()
   - Check `len(parts) > 0` after parsing
   - If empty, return openWithFinishedMsg with error "Application field cannot be empty"
   - First element is command, rest are options
   - Append file list to arguments
   - Use exec.Command(command, args...)
   - Set working directory
   - Start process and return result message

6. **Wire message handlers**
   - Add OpenWithDialog creation in handleCustomMessages()
   - Add openWithDialogResultMsg handler to invoke openWithCustom()

**Dependencies**:
- Requires: Phase 1 (message types, exec patterns)
- Blocks: Phase 3 (default app detection enhances this dialog)

**Testing Approach**:

*Unit Tests*:
- Test OpenWithDialog initializes correctly
- Test file list formatting (single file, multiple files)
- Test file list truncation for long lists
- Test application input editing (insert, delete, cursor movement)
- Test horizontal scrolling in application field
- Test Enter with non-empty input sends result message
- Test Enter with empty input does nothing
- Test Esc sends cancelled result
- Test openWithCustom() parses application correctly
- Test openWithCustom() builds correct command with options

*Integration Tests*:
- Test complete flow: Context menu → "Open with ..." → Dialog → Command launch
- Test with single file
- Test with multiple marked files
- Test custom application with options (e.g., "mpv --loop")
- Test error handling: command not found

*Manual Testing*:
- [ ] Dialog displays with empty application field
- [ ] File list shows all marked files (or selected file)
- [ ] Application field accepts editing with cursor visible
- [ ] Long application names scroll horizontally
- [ ] Enter launches application with files
- [ ] Esc cancels and returns to file manager
- [ ] Multiple files passed correctly to application

**Acceptance Criteria**:
- [ ] "Open with ..." menu item appears at position 2
- [ ] Dialog displays application input field and file list
- [ ] Application field supports editing with Emacs keybindings
- [ ] File list shows filenames quoted for display
- [ ] Enter with application launches command in background
- [ ] Multiple files passed as separate arguments (unquoted)
- [ ] Esc cancels dialog without action
- [ ] Error message shown if command not found

**Estimated Effort**: 中 (3-5 days)

**Risks and Mitigation**:
- **Risk**: File list truncation may hide important files
  - **Mitigation**: Show file count in dialog, truncate intelligently
- **Risk**: Parsing application string may fail for complex commands
  - **Mitigation**: Use strings.Fields(), document limitations, user can work around with shell

---

### Phase 3: Default Application Detection

**Goal**: Enhance OpenWithDialog to automatically detect and pre-fill the default application using xdg-mime.

**Files to Create**:
- None (modifications only)

**Files to Modify**:
- `internal/ui/open_with_dialog.go`:
  - Add getDefaultApplication() function
  - Call xdg-mime query filetype and xdg-mime query default
  - Parse .desktop filename to extract application name
  - Pre-populate application field if detection succeeds

- `internal/ui/open_with_dialog_test.go`:
  - Add tests for getDefaultApplication()
  - Mock xdg-mime commands

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| getDefaultApplication() | Query MIME type and default app with timeout | Single file path provided | Application name or empty string |
| OpenWithDialog.Init() | Call getDefaultApplication and populate field | File path available | Application field pre-filled or empty |
| xdg-mime query filetype | Determine MIME type of file | File exists | MIME type string or error |
| xdg-mime query default | Find default app for MIME type | MIME type known | .desktop filename or error |

**Processing Flow**:

```
1. OpenWithDialog created with file path
2. If single file (not directory, not multiple files):
   ├─ Call getDefaultApplication(filePath)
   └─ Otherwise: skip (leave application field empty)
3. getDefaultApplication() executes:
   ├─ Run: xdg-mime query filetype <file>
   ├─ Parse output to get MIME type (e.g., "video/mp4")
   ├─ If error → Return empty string
   ├─ Run: xdg-mime query default <mimetype>
   ├─ Parse output to get .desktop file (e.g., "mpv.desktop")
   ├─ If error → Return empty string
   ├─ Remove ".desktop" extension
   └─ Return application name (e.g., "mpv")
4. OpenWithDialog sets application field to result
5. User sees pre-filled application (or empty if detection failed)
6. User can edit or accept default
```

**Implementation Steps**:

1. **Implement getDefaultApplication() function**
   - Execute xdg-mime query filetype with file path using `exec.CommandContext(ctx, "xdg-mime", "query", "filetype", filePath)`
     - Create individual context with 500ms timeout: `ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond); defer cancel()`
     - Capture stdout with `cmd.Output()`
     - Apply `strings.TrimSpace()` to stdout to remove trailing newlines
     - Handle errors (context deadline exceeded, command failure) by returning empty string
   - Execute xdg-mime query default with MIME type using `exec.CommandContext(ctx, "xdg-mime", "query", "default", mimeType)`
     - Create individual context with 500ms timeout: `ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond); defer cancel()`
     - Capture stdout with `cmd.Output()`
     - Apply `strings.TrimSpace()` to stdout to remove trailing newlines
     - Handle errors (context deadline exceeded, command failure) by returning empty string
   - Parse .desktop filename using `strings.TrimSuffix(desktop, ".desktop")`
   - Return application name or empty string on any failure

2. **Integrate into OpenWithDialog initialization**
   - Determine if single file (not multiple, not directory)
   - Call getDefaultApplication() if applicable
   - Set TextInput value to result
   - Position cursor at end of application name

3. **Add error handling**
   - xdg-mime not available → empty string (no error to user)
   - MIME type unknown → empty string
   - No default app set → empty string
   - User sees empty field and can type manually

**Dependencies**:
- Requires: Phase 2 (OpenWithDialog exists)
- Blocks: Phase 4 (polish builds on this foundation)

**Testing Approach**:

*Unit Tests*:
- Test getDefaultApplication() with mock xdg-mime commands
- Test returns application name when successful
- Test returns empty string when xdg-mime fails
- Test removes .desktop extension correctly
- Test handles MIME type detection failure
- Test handles no default app case

*Integration Tests*:
- Test OpenWithDialog initializes with default app for known file type
- Test OpenWithDialog initializes with empty app for unknown file type
- Test OpenWithDialog skips detection for multiple files
- Test OpenWithDialog skips detection for directories

*Manual Testing*:
- [ ] Open .mp4 file → "mpv" or similar pre-filled
- [ ] Open .pdf file → PDF reader pre-filled
- [ ] Open unknown file type → empty field
- [ ] Open multiple files → empty field
- [ ] User can edit pre-filled application
- [ ] User can clear and type different application

**Acceptance Criteria**:
- [ ] Default application detected for single files
- [ ] Application field pre-filled when default app found
- [ ] Empty field when default app not found (no error)
- [ ] Detection skipped for multiple files
- [ ] Detection skipped for directories
- [ ] User can edit pre-filled application

**Estimated Effort**: 小 (1-2 days)

**Risks and Mitigation**:
- **Risk**: xdg-mime may be slow on some systems
  - **Mitigation**: Use timeout (500ms), skip if too slow
- **Risk**: Detection may fail for many file types
  - **Mitigation**: Fallback to empty is acceptable UX

---

### Phase 4: Polish and Edge Cases

**Goal**: Handle edge cases, improve error messages, add comprehensive tests, and ensure robustness.

**Files to Create**:
- None (test files and modifications only)

**Files to Modify**:
- `internal/ui/open_with_dialog.go`:
  - Improve file list truncation algorithm
  - Add validation for empty application field (already implemented in Phase 2)
  - Improve error messages for specific failure modes

- `internal/ui/exec.go`:
  - Add better error detection (command not found vs. other errors)
  - Sanitize file paths with filepath.Clean

- `internal/ui/model_update.go`:
  - Differentiate error messages (xdg-open not found vs. launch failed)

- Test files:
  - Add comprehensive unit tests for all components
  - Add edge case tests (special characters, long filenames, etc.)
  - Add integration tests

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| File path sanitization | Ensure safe paths before execution | Raw file path | Cleaned, normalized path |
| Error classification | Determine error type for user message | exec error | Specific error message |
| File list truncation | Smart truncation to show most files | Long file list | Truncated with count indicator |
| Special character handling | Handle filenames with quotes, spaces, etc. | Any filename | Safely passed as separate arguments (exec.Command handles escaping) |

**Processing Flow**:

```
1. File path sanitization:
   ├─ Use filepath.Join(workDir, filename)
   ├─ Apply filepath.Clean() to normalize
   └─ Prevent path traversal attacks

2. Error classification:
   ├─ Check error message for "executable file not found"
   ├─ Map to user-friendly message
   └─ Display specific guidance

3. File list truncation:
   ├─ Calculate available width
   ├─ Add files until width exceeded
   ├─ Show "... and N more" if truncated
   └─ Always show at least 1 file

4. Special character handling:
   ├─ Filenames passed as separate arguments (unquoted)
   ├─ exec.Command handles escaping automatically
   ├─ Filenames quoted only for UI display purposes
   └─ No shell interpolation (safe)
```

**Implementation Steps**:

1. **Add path sanitization**
   - Use filepath.Join for all path construction
   - Apply filepath.Clean to remove .. and normalize
   - Review all path handling in exec functions

2. **Improve error messages**
   - Check error strings for common patterns
   - Map to user-friendly messages
   - Provide actionable guidance (e.g., "Install xdg-utils")

3. **Enhance file list truncation**
   - Calculate available width based on dialog size
   - Build file list incrementally for display:
     - Start with empty string
     - For each file, add quoted filename + space (quotes for display only)
     - Check if total width exceeds limit
     - If exceeded, stop adding files and append "... and N more" where N = remaining file count
   - Ensure first file always visible (minimum display)
   - Example: `"file1.txt" "file2.txt" ... and 3 more`
   - Note: Display quotes are cosmetic; execution uses unquoted arguments

4. **Add comprehensive tests**
   - Unit tests for all functions (target 90%+ coverage)
   - Edge cases: empty filenames, special characters, symlinks
   - Error cases: missing commands, permission denied
   - Integration tests for complete flows

5. **Test with special characters**
   - Test filenames with spaces: "test file.txt"
   - Test filenames with quotes: "file's name.txt"
   - Test filenames with shell special chars: "file;rm.txt"
   - Verify no shell injection vulnerabilities

6. **Add performance checks**
   - Ensure default app detection < 500ms
   - Ensure command launch < 100ms
   - Ensure dialog display < 50ms

**Dependencies**:
- Requires: Phases 1-3 (all core functionality complete)
- Blocks: None (final phase)

**Testing Approach**:

*Unit Tests*:
- Test path sanitization removes .. and normalizes
- Test error classification for different error types
- Test file list truncation with various lengths
- Test special character handling (spaces, quotes, etc.)
- Achieve 90%+ code coverage

*Integration Tests*:
- Test complete flow with special character filenames
- Test rapid consecutive opens
- Test with very long file lists (100+ files)
- Test error recovery (missing command, permission denied)

*Manual Testing*:
- [ ] Open file with spaces in name
- [ ] Open file with quotes in name
- [ ] Open file with shell special characters
- [ ] Open symlink (follows link)
- [ ] Open broken symlink (error message)
- [ ] Open with very long application command
- [ ] Open with 100 marked files (truncation)
- [ ] xdg-open missing (clear error)
- [ ] Custom command missing (clear error)
- [ ] Rapid consecutive opens (no freeze)

**Acceptance Criteria**:
- [ ] All special character filenames handled correctly
- [ ] Path traversal prevented
- [ ] Error messages are specific and helpful
- [ ] File list truncation works for large selections
- [ ] Unit test coverage > 90%
- [ ] All integration tests pass
- [ ] All manual tests pass
- [ ] No regressions in existing functionality

**Estimated Effort**: 中 (3-5 days)

**Risks and Mitigation**:
- **Risk**: Edge cases may be difficult to reproduce
  - **Mitigation**: Automated tests for all edge cases
- **Risk**: Performance regressions with large file lists
  - **Mitigation**: Performance benchmarks, optimize truncation

---

## Complete File Structure

```
duofm/
├── cmd/duofm/
│   └── main.go                           # Entry point
├── internal/
│   ├── ui/
│   │   ├── context_menu_dialog.go        # MODIFY: Add Open and Open with items
│   │   ├── open_with_dialog.go           # NEW: Open with dialog
│   │   ├── open_with_dialog_test.go      # NEW: Tests for open with dialog
│   │   ├── exec.go                       # MODIFY: Add openWithXDG, openWithCustom
│   │   ├── model_update.go               # MODIFY: Add message handlers
│   │   ├── messages.go                   # MODIFY: Add message types
│   │   ├── dialog.go                     # EXISTING: Dialog interface
│   │   ├── dialog_base.go                # EXISTING: BaseDialog
│   │   ├── text_input.go                 # EXISTING: TextInput component
│   │   └── ...                           # Other existing files
│   ├── fs/
│   │   └── ...                           # File system operations
│   └── config/
│       └── ...                           # Configuration
├── doc/
│   └── tasks/
│       └── context-menu-open-file/
│           ├── SPEC.md                   # Feature specification
│           ├── IMPLEMENTATION.md         # This file
│           └── VERIFICATION.md           # Verification document
├── tests/
│   └── e2e/
│       └── context_menu_open_test.sh     # E2E tests
├── go.mod
├── go.sum
└── Makefile
```

**File Descriptions**:

- **context_menu_dialog.go**: Extends buildMenuItems() to include "Open" and "Open with ..." menu items at positions 1 and 2. Determines enabled state based on marked file count.

- **open_with_dialog.go**: Implements OpenWithDialog with application input field, read-only file list display, and default application detection via xdg-mime. Handles keyboard input for editing and confirmation.

- **open_with_dialog_test.go**: Unit tests for OpenWithDialog, covering initialization, input handling, file list formatting, default app detection, and edge cases.

- **exec.go**: Adds openWithXDG() for launching xdg-open and openWithCustom() for launching user-specified applications. Both functions execute processes in the background using exec.Command and cmd.Start().

- **model_update.go**: Routes messages for open operations. Handles openWithXDGMsg, openWithDialogResultMsg, and openWithFinishedMsg. Updates status bar with results or errors.

- **messages.go**: Defines message types: openWithXDGMsg (file, workDir), openWithDialogResultMsg (application, files, workDir, cancelled), openWithFinishedMsg (error).

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in testing package
- Table-driven tests for multiple scenarios
- Mock exec.Command using interfaces or test helpers
- Validate message flows and state transitions

**Test Coverage Goals**:
- Core logic: 90%+ coverage
- Dialog components: 85%+ coverage
- Error handling paths: 100% coverage

**Key Test Areas**:

1. **Context Menu** (`context_menu_dialog_test.go`)
   - "Open" item appears at position 1
   - "Open" enabled when markCount == 0
   - "Open" disabled when markCount > 0
   - "Open with ..." appears at position 2
   - "Open with ..." always enabled
   - Menu item callbacks return correct action IDs

2. **Open With Dialog** (`open_with_dialog_test.go`)
   - Dialog initializes with empty application field
   - Dialog initializes with default app (when detected)
   - File list formatting: single file, multiple files
   - File list truncation for long lists
   - Application input accepts editing (insert, delete, cursor)
   - Horizontal scrolling in application field
   - Enter with non-empty app sends result message
   - Enter with empty app does nothing
   - Esc sends cancelled result

3. **Default App Detection** (`open_with_dialog_test.go`)
   - getDefaultApplication() returns app name
   - Removes .desktop extension correctly
   - Returns empty string if xdg-mime fails
   - Returns empty string if MIME type unknown
   - Handles directories gracefully (returns empty)

4. **Background Execution** (`exec_test.go`)
   - openWithXDG() builds correct command
   - openWithCustom() parses application string
   - openWithCustom() builds command with options
   - Working directory set correctly
   - Returns error if command not found
   - Process runs in background (non-blocking)

### Integration Testing

**Scenarios**:

1. **Complete Open Flow**
   - User opens context menu
   - Selects "Open"
   - xdg-open invoked with correct file
   - Status message displayed

2. **Complete Open With Flow**
   - User opens context menu
   - Selects "Open with ..."
   - Dialog displays with file list
   - User types application name
   - Presses Enter
   - Application launched with files
   - Status message displayed

3. **Multiple Files Flow**
   - User marks multiple files
   - Opens context menu
   - "Open" is disabled
   - Selects "Open with ..."
   - Dialog shows all files
   - Application receives all files as arguments

4. **Error Handling**
   - xdg-open not found → Error message
   - Custom command not found → Error message
   - duofm remains responsive after errors

### Manual Testing Checklist

**Basic Functionality**:
- [ ] "Open" menu item launches xdg-open
- [ ] File opens in default application
- [ ] Directory opens in file manager
- [ ] "Open with ..." displays dialog
- [ ] Application field accepts input
- [ ] Custom application launches with file
- [ ] Multiple files passed to custom application
- [ ] duofm remains responsive

**Edge Cases**:
- [ ] Very long application name (scrolling)
- [ ] Very long file list (truncation)
- [ ] Filename with spaces: "test file.txt"
- [ ] Filename with quotes: "file's name.txt"
- [ ] Filename with shell chars: "file;rm.txt"
- [ ] Parent directory (..)
- [ ] Symlink (follows link)
- [ ] Broken symlink (error)
- [ ] Rapid consecutive opens

**Error Handling**:
- [ ] xdg-open not installed
- [ ] Custom command not found
- [ ] Permission denied on file
- [ ] Empty application field (no action)
- [ ] Default app detection timeout

**Performance**:
- [ ] Default app detection < 500ms
- [ ] Command launch < 100ms
- [ ] Dialog display < 50ms
- [ ] 100 marked files (smooth)

## Dependencies

### External Dependencies

| Package | Version | Purpose | Installation |
|---------|---------|---------|--------------|
| xdg-utils | - | xdg-open, xdg-mime commands | `apt install xdg-utils` (Debian/Ubuntu) |
| github.com/charmbracelet/bubbletea | v0.25.0 | TUI framework | `go get` |
| github.com/charmbracelet/lipgloss | v0.9.1 | Styling | `go get` |

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1 (no dependencies on new code)
2. Phase 2 (depends on Phase 1 message types and exec patterns)
3. Phase 3 (depends on Phase 2 dialog structure)
4. Phase 4 (depends on Phases 1-3 being complete)

**Component Dependencies**:
- `open_with_dialog.go` depends on `text_input.go`, `dialog_base.go`, `messages.go`
- `exec.go` additions depend on Go stdlib `os/exec`
- `model_update.go` handlers depend on all message types being defined
- Context menu additions depend on existing menu infrastructure

## Risk Assessment

### Technical Risks

1. **xdg-open Availability**
   - **Risk**: xdg-open may not be installed on user's system
   - **Likelihood**: Low (standard on most desktop Linux)
   - **Impact**: High (core feature unusable)
   - **Mitigation**:
     - Clear error message: "Cannot open file: xdg-open not found. Install xdg-utils package."
     - Document requirement in README
     - Graceful degradation (feature disabled, no crash)

2. **Default App Detection Performance**
   - **Risk**: xdg-mime queries may be slow on some systems
   - **Likelihood**: Low
   - **Impact**: Medium (dialog appears slowly)
   - **Mitigation**:
     - Implement timeout (500ms max)
     - Skip detection if timeout exceeded
     - User can still type manually

3. **Shell Injection Vulnerability**
   - **Risk**: Improper handling of filenames could allow command injection
   - **Likelihood**: Very Low (mitigated by design)
   - **Impact**: Critical (security vulnerability)
   - **Mitigation**:
     - Never use shell execution (no sh -c)
     - Always use exec.Command with separate arguments
     - Quote filenames at display only
     - Comprehensive security testing

4. **Background Process Management**
   - **Risk**: Child processes may not terminate cleanly
   - **Likelihood**: Low
   - **Impact**: Medium (resource leak)
   - **Mitigation**:
     - Use cmd.Start() without Wait() (child runs independently)
     - Document that duofm doesn't manage child process lifecycle
     - Rely on OS to clean up orphaned processes

### Implementation Risks

1. **Scope Creep**
   - **Risk**: Adding features beyond spec (e.g., history, favorites)
   - **Mitigation**: Stick to spec phases, document future enhancements separately

2. **Test Coverage Gaps**
   - **Risk**: Missing edge cases in testing
   - **Mitigation**: Comprehensive test plan, code review, manual testing checklist

3. **Integration with Existing Code**
   - **Risk**: Breaking existing context menu functionality
   - **Mitigation**: Regression tests, careful message routing, code review

## Performance Considerations

### Performance Goals

- **Dialog Display**: < 50ms
- **Default App Detection**: < 500ms (acceptable, one-time)
- **Command Launch**: < 100ms (start only, not including app launch)

### Optimization Strategies

**Dialog Rendering**:
- Pre-calculate file list string at dialog creation
- Cache formatted file list (no recalculation per frame)
- Use strings.Builder for efficient string concatenation

**Default App Detection**:
- Run xdg-mime commands serially (simple, adequate performance)
- Implement timeout to prevent blocking UI
- Cache result during dialog lifetime (no re-query)
- Skip detection for multiple files and directories (not applicable)

**Command Execution**:
- Use cmd.Start() instead of cmd.Run() (non-blocking)
- Do not wait for process completion
- Discard stdout/stderr (no pipe overhead)

**File List Truncation**:
- Calculate truncation once during initialization
- Use efficient string building
- Limit complexity to O(n) where n = number of files

**Memory Management**:
- No caching across operations (stateless, low memory footprint)
- File list held temporarily during dialog display
- Release resources when dialog closes

## Security Considerations

### Path Sanitization

**Strategy**:
- Use `filepath.Join(workDir, filename)` to construct absolute paths
- Apply `filepath.Clean()` to normalize and remove `..` components
- Prevent path traversal attacks

**Example**:
```
// Safe path construction
fullPath := filepath.Clean(filepath.Join(workDir, filename))
```

### Shell Injection Prevention

**CRITICAL**: Never use shell execution.

**Safe Approach**:
```
// SAFE: Direct execution with separate arguments (no quotes needed)
cmd := exec.Command("mpv", "--loop", "video1.mp4", "video2.mp4")
cmd.Dir = workDir
cmd.Start()
```

**Unsafe Approach (DO NOT USE)**:
```
// UNSAFE: Shell injection vulnerability
cmd := exec.Command("sh", "-c", "mpv --loop video1.mp4 video2.mp4")
```

**Rationale**:
- exec.Command automatically handles argument escaping
- No manual quoting needed - pass filenames as-is
- No shell interpolation occurs
- Special characters in filenames are safe
- Quotes are only for UI display purposes

### Input Validation

**Application Field**:
- Accept any string (validation happens at execution)
- Do not attempt to parse complex quoting/escaping
- Let exec.Command handle argument separation

**File List**:
- Only use files from active pane (not user input)
- Filenames come from file system, not user typing
- Trust file system layer to provide valid paths

**Working Directory**:
- Only use current pane directory
- No user input for workDir
- Already validated by pane navigation

### Process Isolation

**Approach**:
- Child processes run with same user permissions as duofm
- No privilege escalation attempted
- Stdout/stderr discarded (no information leak to terminal)
- Child process lifecycle independent of duofm

**Security Properties**:
- If user can run duofm, they can run applications via Open
- No additional permissions granted
- Sandboxing delegated to external application (if any)

## Error Handling

### Error Categories

**xdg-open Errors**:

| Error | Condition | User Message | Action |
|-------|-----------|--------------|--------|
| Command not found | xdg-open not installed | "Cannot open file: xdg-open not found. Install xdg-utils package." | Display in status bar for 5s |
| Execution failed | xdg-open returns error | "Failed to open file: [error]" | Display in status bar for 5s |

**Custom Command Errors**:

| Error | Condition | User Message | Action |
|-------|-----------|--------------|--------|
| Command not found | Application not in PATH | "Command not found: [command]" | Display in status bar for 5s |
| Launch failed | exec.Start() fails | "Failed to execute: [error]" | Display in status bar for 5s |
| Empty command | User clears field and presses Enter | (no action) | Ignore Enter key |

**Default App Detection Errors**:

| Error | Condition | Behavior |
|-------|-----------|----------|
| xdg-mime not found | Command unavailable | Empty application field (no error message) |
| MIME type unknown | File type unrecognized | Empty application field |
| No default app | No association set | Empty application field |
| Detection timeout | > 500ms | Empty application field |

### Error Flow

```
Launch Command → cmd.Start()
    ├─ Success → Return openWithFinishedMsg{err: nil}
    │           → Status: "Opened with [app]" (5 seconds)
    │
    └─ Failure → Inspect error
                ├─ "executable file not found" → "Command not found: [cmd]"
                └─ Other error → "Failed to execute: [error]"
                → Status message (5 seconds, red text)
                → Auto-clear after timeout
```

### Error Message Design

**Principles**:
- Specific: Identify the exact problem
- Actionable: Provide guidance on fixing the issue
- Non-intrusive: Status bar message, auto-clear after 5 seconds
- No modal errors: User can continue working

**Examples**:
- Good: "Cannot open file: xdg-open not found. Install xdg-utils package."
- Bad: "Error" (too vague)
- Good: "Command not found: vlc"
- Bad: "exec: \"vlc\": executable file not found in $PATH" (too technical)

## Open Questions

None - all requirements clarified through specification.

## Future Enhancements

**Not in Current Spec** (deferred to future releases):

1. **Open with History**
   - Remember recently used applications
   - Quick selection from history in dialog
   - Stored in configuration file

2. **Application Favorites**
   - Pin frequently used applications
   - Display pinned apps in dialog or submenu
   - User-configurable shortcuts

3. **Terminal vs. GUI App Detection**
   - Detect if application is terminal-based
   - Use tea.ExecProcess for terminal apps (foreground)
   - Use cmd.Start() for GUI apps (background)
   - Improves UX for mixed workflows

4. **Progress Feedback for Slow Apps**
   - Show "Launching..." message while app starts
   - Detect when app window appears
   - Requires external dependencies (wmctrl, xdotool)

5. **macOS and Windows Support**
   - Use `open` command on macOS
   - Use `start` command on Windows
   - Abstract platform differences

## Success Metrics

### Functional Completeness
- [ ] "Open" launches xdg-open successfully
- [ ] "Open with ..." dialog accepts input and launches applications
- [ ] Default application detection works for common file types
- [ ] Multiple files passed correctly to custom applications
- [ ] Filenames with special characters handled safely
- [ ] Applications run in background without blocking duofm

### Quality Metrics
- [ ] Unit test coverage > 90% for new code
- [ ] All integration tests pass
- [ ] All manual test checklist items pass
- [ ] No regressions in existing context menu functionality
- [ ] Code follows duofm conventions and patterns

### Performance Metrics
- [ ] Dialog display < 50ms
- [ ] Default app detection < 500ms
- [ ] Command launch < 100ms
- [ ] UI responsive during background processes

### User Experience
- [ ] Clear error messages for all failure modes
- [ ] Application field supports standard editing keybindings
- [ ] File list display is clear and informative
- [ ] "Open" disabled state is visually clear
- [ ] No unexpected behavior or freezes

### Security
- [ ] No shell injection vulnerabilities
- [ ] Path traversal prevented
- [ ] Special characters in filenames handled safely
- [ ] Security review passed

## References

- **Specification**: `doc/tasks/context-menu-open-file/SPEC.md`
- **xdg-utils Documentation**: https://www.freedesktop.org/wiki/Software/xdg-utils/
- **Existing Implementation Patterns**:
  - Context menu: `internal/ui/context_menu_dialog.go`
  - Input dialog: `internal/ui/input_dialog.go`
  - Text input: `internal/ui/text_input.go`
  - Exec patterns: `internal/ui/exec.go`
- **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
- **Go exec Package**: https://pkg.go.dev/os/exec
- **Related Features**:
  - Open file with viewer/editor: `doc/tasks/open-file/SPEC.md`
  - Context menu implementation: existing codebase

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Review this document for accuracy and completeness
   - Address any questions or concerns
   - Confirm implementation approach

2. **Environment Setup**
   - Ensure Go 1.21+ installed
   - Install xdg-utils package for testing
   - Verify development environment

3. **Begin Implementation**
   - Start with Phase 1 (Basic Open with xdg-open)
   - Follow TDD approach (write tests first when practical)
   - Commit incrementally with clear messages
   - Review after each phase completion

4. **Testing and Validation**
   - Run unit tests after each component
   - Run integration tests after each phase
   - Manual testing throughout development
   - Use `/sdd.6-verify` to check SPEC compliance

5. **Code Review**
   - Use `/sdd.7-review` for automated code review
   - Address any issues found
   - Ensure code quality standards met

6. **Completion**
   - Verify all acceptance criteria met
   - Run full test suite
   - Update documentation if needed
   - Create release notes summarizing changes
