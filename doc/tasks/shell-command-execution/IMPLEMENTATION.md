# Implementation Plan: Shell Command Execution

## Overview

Add the ability to execute arbitrary shell commands from within duofm using the `!` key. This feature leverages the existing minibuffer component for command input and the `tea.ExecProcess` mechanism for TUI suspension.

## Objectives

- Enable users to run shell commands without leaving duofm
- Provide familiar command input using the existing minibuffer
- Execute commands in the active pane's directory
- Display "Press Enter to continue..." after command completion

## Prerequisites

- Existing minibuffer component (internal/ui/minibuffer.go)
- Existing external command execution pattern (internal/ui/exec.go)
- Understanding of Bubble Tea's `tea.ExecProcess` for TUI suspension

## Architecture Overview

The implementation follows the existing pattern used for search functionality:

```
User Input → Model State Change → Minibuffer Display → Command Execution → Pane Reload
```

### Component Interaction

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   keys.go   │────>│  model.go   │────>│   exec.go   │
│ (key const) │     │ (state mgmt)│     │ (execution) │
└─────────────┘     └──────┬──────┘     └─────────────┘
                          │
                          v
                   ┌─────────────┐
                   │minibuffer.go│
                   │  (reuse)    │
                   └─────────────┘
```

## Implementation Phases

### Phase 1: Add Key Constant and Shell Command Function

**Goal**: Define the key binding and create the shell command execution function

**Files to Modify**:
- `internal/ui/keys.go` - Add key constant for shell command
- `internal/ui/exec.go` - Add shell command execution function

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| KeyShellCommand | Define `!` as shell command trigger | None | Constant available for use in model |
| shellCommandFinishedMsg | Signal shell command completion | None | Contains execution result (success/error) |
| executeShellCommand | Execute command via shell | Valid command string, valid directory path | TUI suspended, shell process started, completion message sent |

**Processing Flow** (must be convertible to flowchart):
```
executeShellCommand(command, workDir):
    1. Create shell process with /bin/sh -c
    2. Set working directory to workDir
    3. Append "Press Enter to continue..." prompt
    4. Execute via tea.ExecProcess (suspends TUI)
    5. On completion → send shellCommandFinishedMsg
```

**Implementation Steps**:
1. Add `KeyShellCommand = "!"` constant to keys.go
2. Add `shellCommandFinishedMsg` type to exec.go
3. Add `executeShellCommand` function following existing `openWithViewer` pattern

**Testing**:
- Verify key constant is defined and accessible
- Verify message type can carry error information

**Estimated Effort**: Small

---

### Phase 2: Add Model State for Shell Command Mode

**Goal**: Track shell command mode in the model

**Files to Modify**:
- `internal/ui/model.go` - Add state field

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| shellCommandMode (field) | Track whether user is entering shell command | Model initialized | Boolean state available |

**Processing Flow**:
```
Model State Transitions:
    Normal Mode ──[! key]──> Shell Command Mode
    Shell Command Mode ──[Enter]──> Executing (TUI suspended)
    Shell Command Mode ──[Escape]──> Normal Mode
    Executing ──[command exits]──> Normal Mode
```

**Implementation Steps**:
1. Add `shellCommandMode bool` field to Model struct
2. Initialize to `false` in NewModel

**Testing**:
- Verify model initializes with shellCommandMode = false

**Estimated Effort**: Small

---

### Phase 3: Implement Key Handler and Input Processing

**Goal**: Handle `!` key to enter shell command mode and process input

**Files to Modify**:
- `internal/ui/model.go` - Add key handler and input processing

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `!` key handler | Enter shell command mode | No dialog active, not in search mode | shellCommandMode=true, minibuffer visible with "!" prompt |
| Input handler | Process keystrokes in shell command mode | shellCommandMode=true | Command buffer updated or mode exited |
| Completion handler | Handle shellCommandFinishedMsg | Message received | Both panes reloaded, error shown if any |

**Processing Flow** (must be convertible to flowchart):
```
On key press:
    1. Check current mode
       ├─ Dialog active → ignore `!`
       ├─ Search active → ignore `!`
       └─ Normal mode → enter shell command mode

    2. In shell command mode:
       ├─ Enter key
       │   ├─ Command empty → exit mode (no execution)
       │   └─ Command not empty → execute command
       ├─ Escape key → exit mode (cancel)
       └─ Other keys → delegate to minibuffer

On shellCommandFinishedMsg:
    1. Reload both panes (preserve cursor)
    2. If error → display in status bar
```

**Error Handling** (behavioral level):
- Command execution error → Display error message in status bar for 5 seconds
- Empty command → Exit mode without execution (not an error)

**Implementation Steps**:
1. Add `!` key case to Update function's key switch
2. Add shell command mode input handling block (similar to search mode)
3. Add shellCommandFinishedMsg case to message handling

**Testing**:
- `!` key activates shell command mode
- Typing characters appends to minibuffer
- Enter executes non-empty command
- Enter on empty command exits without execution
- Escape cancels and returns to normal mode
- `!` ignored when dialog is open
- `!` ignored when search is active

**Estimated Effort**: Medium

---

### Phase 4: Update View Rendering

**Goal**: Display minibuffer when in shell command mode

**Files to Modify**:
- `internal/ui/model.go` - Update View function

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| View rendering logic | Show minibuffer in active pane | shellCommandMode=true | Minibuffer visible with "!" prompt |

**Processing Flow**:
```
View():
    If shellCommandMode OR searchState.IsActive:
        Render active pane with minibuffer
        Render inactive pane normally
    Else:
        Render both panes normally
```

**Implementation Steps**:
1. Extend existing search mode condition to include shellCommandMode
2. Reuse ViewWithMinibuffer for rendering

**Testing**:
- Minibuffer displays with "!" prompt in shell command mode
- Input visible as user types
- Minibuffer disappears after Enter or Escape

**Estimated Effort**: Small

---

### Phase 5: Update Help Dialog

**Goal**: Add shell command key to help screen

**Files to Modify**:
- `internal/ui/help_dialog.go` - Add help entry

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Help text entry | Document `!` key functionality | None | User can see shell command in help |

**Implementation Steps**:
1. Add `!` key entry to help dialog content

**Testing**:
- Help dialog shows shell command key binding

**Estimated Effort**: Small

---

## File Structure

```
internal/ui/
├── keys.go         # Add KeyShellCommand = "!"
├── exec.go         # Add executeShellCommand(), shellCommandFinishedMsg
├── model.go        # Add shellCommandMode field, key handler, message handler
├── minibuffer.go   # (no changes - reuse existing)
└── help_dialog.go  # Add shell command to help text
```

## Testing Strategy

### Unit Tests

- `internal/ui/exec_test.go`: Test shellCommandFinishedMsg type
- `internal/ui/model_test.go`: Test shell command mode state transitions

### Manual Testing Checklist

- [ ] Press `!` to enter shell command mode
- [ ] Minibuffer shows with "!" prompt
- [ ] Type a command (e.g., `ls -la`)
- [ ] Press `Enter` - command executes in active pane's directory
- [ ] "Press Enter to continue..." is displayed
- [ ] Press `Enter` - returns to duofm
- [ ] Both panes are refreshed
- [ ] Press `!` then `Escape` - cancels without executing
- [ ] Press `!` then `Enter` with empty input - exits without executing
- [ ] `!` is ignored when a dialog is open
- [ ] `!` is ignored during search mode
- [ ] Command with error (e.g., `nonexistent_command`) - shell shows error, returns normally
- [ ] Complex commands work (e.g., `ls -la | head -5`)

## Dependencies

### External Libraries

| Library | Purpose |
|---------|---------|
| `os/exec` | Standard library for command execution |
| `github.com/charmbracelet/bubbletea` | TUI framework (tea.ExecProcess) |

### Internal Dependencies

| Component | Dependency Reason |
|-----------|-------------------|
| `internal/ui/minibuffer.go` | Text input component (reuse) |
| `internal/ui/exec.go` | External command execution pattern (extend) |

## Risk Assessment

### Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Shell portability (`/bin/sh`) | Low | Medium | Standard approach, works on all Unix-like systems |
| Mode conflicts | Low | Low | Check all active modes before entering shell command mode |

## Performance Considerations

- No performance impact during normal operation
- Command execution handled by OS
- Pane reload after return uses existing async mechanism

## Security Considerations

- Commands executed via `/bin/sh -c` (full shell functionality)
- Working directory controlled by duofm (active pane path)
- No input sanitization required (user explicitly enters commands)
- Intentional behavior for a file manager

## Open Questions

None - all requirements have been clarified.

## References

- Specification: `doc/tasks/shell-command-execution/SPEC.md`
- Existing external command pattern: `internal/ui/exec.go` (openWithViewer, openWithEditor)
- Existing minibuffer usage: search functionality in `internal/ui/model.go`
