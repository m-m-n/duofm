# Implementation Plan: Open File with External Application

## Overview

Implement the ability to open files with external applications (less for viewing, vim for editing) from duofm. This feature adds `v`, `e`, and modifies `Enter` key behavior for files.

## Objectives

- Add `v` key to open files with `less` (view)
- Add `e` key to open files with `vim` (edit)
- Modify `Enter` key to open files with `less` (when file is selected)
- Maintain existing `Enter` behavior for directories
- Reload both panes after external app exits

## Prerequisites

- Existing Bubble Tea TUI framework
- Understanding of `tea.ExecProcess` for external command execution
- Familiarity with existing key handling in `model.go`

## Architecture Overview

The implementation uses Bubble Tea's `tea.ExecProcess` to suspend the TUI and run external commands. This approach:
- Temporarily hides duofm's screen
- Gives full terminal control to the external app (less/vim)
- Restores duofm when the external app exits
- Sends a message to trigger pane reload

```
internal/ui/
├── keys.go           # Add KeyView, KeyEdit constants
├── model.go          # Add key handlers and message processing
├── exec.go           # NEW: External command execution functions
└── exec_test.go      # NEW: Unit tests for exec functions
```

## Implementation Phases

### Phase 1: Add Key Constants

**Goal**: Define new key bindings for view and edit actions

**Files to Modify**:
- `internal/ui/keys.go` - Add KeyView and KeyEdit constants

**Implementation Steps**:
1. Add `KeyView = "v"` constant for viewing files
2. Add `KeyEdit = "e"` constant for editing files
3. Keep constants organized with existing navigation keys

**Code Changes**:
```go
// keys.go - Add after existing key constants
const (
    // ... existing keys ...

    // File operations
    KeyView = "v"  // View file with less
    KeyEdit = "e"  // Edit file with vim
)
```

**Testing**:
- Verify constants are accessible from model.go
- No functional tests needed for this phase

**Estimated Effort**: Small

---

### Phase 2: Create External Command Execution Module

**Goal**: Implement functions to execute external applications

**Files to Create**:
- `internal/ui/exec.go` - External command execution

**Implementation Steps**:
1. Create `execFinishedMsg` message type for command completion
2. Implement `openWithViewer(path string) tea.Cmd` function
3. Implement `openWithEditor(path string) tea.Cmd` function
4. Implement `checkReadPermission(path string) error` helper function

**Code Structure**:
```go
// exec.go
package ui

import (
    "os"
    "os/exec"

    tea "github.com/charmbracelet/bubbletea"
)

// execFinishedMsg is sent when external command completes
type execFinishedMsg struct {
    err error
}

// openWithViewer opens the file with less
func openWithViewer(path string) tea.Cmd {
    c := exec.Command("less", path)
    return tea.ExecProcess(c, func(err error) tea.Msg {
        return execFinishedMsg{err: err}
    })
}

// openWithEditor opens the file with vim
func openWithEditor(path string) tea.Cmd {
    c := exec.Command("vim", path)
    return tea.ExecProcess(c, func(err error) tea.Msg {
        return execFinishedMsg{err: err}
    })
}

// checkReadPermission verifies the file can be read
func checkReadPermission(path string) error {
    f, err := os.Open(path)
    if err != nil {
        return err
    }
    f.Close()
    return nil
}
```

**Dependencies**:
- `os` package for file permission check
- `os/exec` package for command execution
- `github.com/charmbracelet/bubbletea` for tea.ExecProcess

**Testing**:
- Test `checkReadPermission` with readable and non-readable files
- Test message types are properly defined

**Estimated Effort**: Small

---

### Phase 3: Add Message Handler for Command Completion

**Goal**: Process `execFinishedMsg` and reload panes

**Files to Modify**:
- `internal/ui/model.go` - Add message handler

**Implementation Steps**:
1. Add `execFinishedMsg` handler at the beginning of `Update()` function
2. Reload both active and inactive panes on command completion
3. Display error in status bar if command failed
4. Use existing `statusMessageClearCmd` for auto-clearing errors

**Code Location**: Add after `directoryLoadCompleteMsg` handler in `Update()`

```go
// Handle external command completion
if result, ok := msg.(execFinishedMsg); ok {
    // Reload both panes to reflect any changes
    m.getActivePane().LoadDirectory()
    m.getInactivePane().LoadDirectory()

    if result.err != nil {
        m.statusMessage = fmt.Sprintf("Command failed: %v", result.err)
        m.isStatusError = true
        return m, statusMessageClearCmd(5 * time.Second)
    }
    return m, nil
}
```

**Testing**:
- Verify panes reload after command completion
- Verify error message displays on failure
- Verify error message auto-clears

**Estimated Effort**: Small

---

### Phase 4: Add Key Handlers for View and Edit

**Goal**: Implement `v` and `e` key handlers

**Files to Modify**:
- `internal/ui/model.go` - Add key handlers in switch statement

**Implementation Steps**:
1. Add handler for `KeyView` ("v") key
2. Add handler for `KeyEdit` ("e") key
3. Both handlers should:
   - Get selected entry
   - Check if it's a file (not directory, not parent dir)
   - Check read permission
   - Execute appropriate external command

**Code Location**: Add in the key switch statement after existing handlers

```go
case KeyView:
    // View file with less
    entry := m.getActivePane().SelectedEntry()
    if entry != nil && !entry.IsParentDir() && !entry.IsDir {
        fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
        if err := checkReadPermission(fullPath); err != nil {
            m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
            m.isStatusError = true
            return m, statusMessageClearCmd(5 * time.Second)
        }
        return m, openWithViewer(fullPath)
    }
    return m, nil

case KeyEdit:
    // Edit file with vim
    entry := m.getActivePane().SelectedEntry()
    if entry != nil && !entry.IsParentDir() && !entry.IsDir {
        fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
        if err := checkReadPermission(fullPath); err != nil {
            m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
            m.isStatusError = true
            return m, statusMessageClearCmd(5 * time.Second)
        }
        return m, openWithEditor(fullPath)
    }
    return m, nil
```

**Testing**:
- `v` on file opens less
- `e` on file opens vim
- `v` on directory does nothing
- `e` on directory does nothing
- `v` on parent dir (..) does nothing
- `e` on parent dir (..) does nothing
- Permission error shows in status bar

**Estimated Effort**: Small

---

### Phase 5: Modify Enter Key Behavior for Files

**Goal**: Make Enter key open files with less while maintaining directory behavior

**Files to Modify**:
- `internal/ui/model.go` - Modify KeyEnter handler

**Implementation Steps**:
1. Locate existing `KeyEnter` handler
2. Add check for file vs directory at the beginning
3. If file: open with viewer (same as `v` key)
4. If directory or parent: use existing `EnterDirectoryAsync()` behavior

**Code Changes**:
```go
case KeyEnter:
    entry := m.getActivePane().SelectedEntry()
    if entry != nil && !entry.IsParentDir() && !entry.IsDir {
        // File selected: open with viewer (same as v key)
        fullPath := filepath.Join(m.getActivePane().Path(), entry.Name)
        if err := checkReadPermission(fullPath); err != nil {
            m.statusMessage = fmt.Sprintf("Cannot read file: %v", err)
            m.isStatusError = true
            return m, statusMessageClearCmd(5 * time.Second)
        }
        return m, openWithViewer(fullPath)
    }
    // Directory or parent dir: existing behavior
    cmd := m.getActivePane().EnterDirectoryAsync()
    return m, cmd
```

**Testing**:
- `Enter` on file opens less
- `Enter` on directory enters it (existing behavior)
- `Enter` on `..` goes to parent (existing behavior)
- Permission error shows in status bar

**Estimated Effort**: Small

---

### Phase 6: Unit Tests

**Goal**: Create comprehensive unit tests

**Files to Create**:
- `internal/ui/exec_test.go` - Tests for exec functions

**Test Cases**:

```go
// exec_test.go
package ui

import (
    "os"
    "path/filepath"
    "testing"
)

func TestCheckReadPermission(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() string // returns file path
        cleanup func(string)
        wantErr bool
    }{
        {
            name: "readable file",
            setup: func() string {
                f, _ := os.CreateTemp("", "test")
                f.Close()
                return f.Name()
            },
            cleanup: func(path string) {
                os.Remove(path)
            },
            wantErr: false,
        },
        {
            name: "non-existent file",
            setup: func() string {
                return "/nonexistent/file/path"
            },
            cleanup: func(string) {},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            path := tt.setup()
            defer tt.cleanup(path)

            err := checkReadPermission(path)
            if (err != nil) != tt.wantErr {
                t.Errorf("checkReadPermission() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestExecFinishedMsg(t *testing.T) {
    // Test that execFinishedMsg can carry error information
    msg := execFinishedMsg{err: nil}
    if msg.err != nil {
        t.Error("expected nil error")
    }

    msg = execFinishedMsg{err: os.ErrNotExist}
    if msg.err != os.ErrNotExist {
        t.Error("expected ErrNotExist")
    }
}
```

**Estimated Effort**: Small

---

## File Structure

```
internal/ui/
├── keys.go           # Add KeyView = "v", KeyEdit = "e"
├── model.go          # Add execFinishedMsg handler, v/e/Enter key handlers
├── exec.go           # NEW: openWithViewer, openWithEditor, checkReadPermission
├── exec_test.go      # NEW: Unit tests
└── messages.go       # Existing message types (no changes)
```

## Testing Strategy

### Unit Tests
- `checkReadPermission()` function
- Message type definitions
- Key constant definitions

### Integration Tests (Manual)
- External app execution with tea.ExecProcess
- Screen restoration after app exit
- Pane reload functionality

### Manual Testing Checklist
- [ ] `v` on text file opens less
- [ ] `e` on text file opens vim
- [ ] `Enter` on text file opens less
- [ ] `Enter` on directory enters it
- [ ] `Enter` on `..` goes to parent
- [ ] `v` on directory: no action
- [ ] `e` on directory: no action
- [ ] `v` on `..`: no action
- [ ] `e` on `..`: no action
- [ ] After exiting less: duofm screen restored
- [ ] After exiting vim: duofm screen restored
- [ ] After vim edit: file list reflects changes
- [ ] Permission error: status bar shows error
- [ ] less not found: error displayed
- [ ] vim not found: error displayed

## Dependencies

### External Libraries
- `github.com/charmbracelet/bubbletea` - Already in use, provides `tea.ExecProcess`

### Go Standard Library
- `os` - File permission checking
- `os/exec` - Command execution
- `path/filepath` - Path construction

### Internal Dependencies
- Phase 1 (keys.go) must be completed before Phase 4
- Phase 2 (exec.go) must be completed before Phase 3 and 4
- Phase 3 (message handler) must be completed before Phase 4 and 5

## Risk Assessment

### Technical Risks
- **tea.ExecProcess behavior**: The function should handle terminal state correctly
  - Mitigation: tea.ExecProcess is well-tested in Bubble Tea; follow documented patterns

- **Terminal restoration issues**: Potential for corrupted terminal state
  - Mitigation: Bubble Tea handles this automatically; test thoroughly

### Implementation Risks
- **Breaking existing Enter behavior**: Must preserve directory navigation
  - Mitigation: Careful condition ordering; test existing functionality first

## Performance Considerations

- No performance impact during normal operation
- External command execution is handled by the OS
- Pane reload after command exit is same as existing directory change

## Security Considerations

- File paths constructed using `filepath.Join` to prevent path traversal
- Read permission checked before executing external command
- Only predetermined commands (`less`, `vim`) are executed - no user input in command
- No shell interpretation - commands executed directly via exec.Command

## Open Questions

None - all requirements have been clarified.

## Future Enhancements

Not in scope for this implementation:
- Configuration file for viewer/editor selection
- MIME type-based application selection
- File extension-based application selection
- Custom command templates

## References

- [Specification Document](./SPEC.md)
- [Requirements Document](./要件定義書.md)
- [Bubble Tea ExecProcess documentation](https://github.com/charmbracelet/bubbletea)
