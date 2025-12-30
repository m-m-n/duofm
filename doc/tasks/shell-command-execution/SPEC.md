# Feature: Shell Command Execution

## Overview

Add the ability to execute arbitrary shell commands from within duofm. Users press `!` key to enter a command, which is executed in an external shell with the current working directory set to the active pane's directory.

## Objectives

- Enable users to run shell commands without leaving duofm
- Provide seamless integration with the existing TUI workflow
- Allow users to review command output before returning to duofm

## User Stories

- As a user, I want to press `!` to enter a command and execute it, so that I can run shell commands from within duofm
- As a user, I want the command to run in the active pane's directory, so that I can operate on files in that location
- As a user, I want to see "Press Enter to continue..." after command completion, so that I can review the output before returning

## Technical Requirements

### Key Binding

| Key | Action |
|-----|--------|
| `!` | Enter command input mode |

### Command Input Mode

| Key | Action |
|-----|--------|
| Any character | Append to command string |
| `Space` | Append space to command string (Note: handled separately due to Bubble Tea's `tea.KeySpace` type) |
| `Backspace` | Delete last character |
| `Enter` | Execute command (if non-empty) |
| `Escape` | Cancel and return to file list |

**Note:** In Bubble Tea, the space key is reported as `tea.KeySpace` type, not as `tea.KeyRunes`. This requires explicit handling in the minibuffer's key handler to ensure spaces can be entered in commands (e.g., `ls -la`, `echo hello world`).

### Command Execution

| Aspect | Specification |
|--------|---------------|
| Shell | `/bin/sh -c "<command>"` |
| Working directory | Active pane's current directory |
| Screen handling | Suspend duofm TUI using `tea.ExecProcess` |

### Post-Execution Behavior

| Aspect | Specification |
|--------|---------------|
| Wait message | Display "Press Enter to continue..." after command completes |
| Return action | User presses `Enter` to return to duofm |
| Pane refresh | Reload both panes upon return |

### Error Handling

| Condition | Behavior |
|-----------|----------|
| Empty command | Exit input mode without executing |
| Command execution error | Shell displays error; duofm returns normally |

## Implementation Approach

### Architecture

```
internal/ui/
├── keys.go              # Add KeyShellCommand = "!"
├── model.go             # Add shell command mode and key handler
├── exec.go              # Add executeShellCommand function
└── minibuffer.go        # Reuse existing minibuffer for command input
```

### Key Constant

```go
// keys.go
const (
    // ... existing keys
    KeyShellCommand = "!" // Execute shell command
)
```

### Shell Command Execution

```go
// exec.go

// shellCommandFinishedMsg is sent when shell command completes
type shellCommandFinishedMsg struct {
    err error
}

// executeShellCommand executes a shell command in the specified directory
func executeShellCommand(command, workDir string) tea.Cmd {
    shellCmd := exec.Command("/bin/sh", "-c", command+"; echo; echo 'Press Enter to continue...'; read _")
    shellCmd.Dir = workDir
    return tea.ExecProcess(shellCmd, func(err error) tea.Msg {
        return shellCommandFinishedMsg{err: err}
    })
}
```

### Model State

```go
// model.go

type Model struct {
    // ... existing fields
    shellCommandMode   bool   // true when entering shell command
    shellCommandBuffer string // command being entered
}
```

### Key Handler

```go
// model.go Update function

case KeyShellCommand:
    if !m.isAnyDialogActive() && !m.searchMode {
        m.shellCommandMode = true
        m.shellCommandBuffer = ""
        return m, nil
    }
```

### Input Handling in Shell Command Mode

```go
// model.go Update function (within shell command mode)

if m.shellCommandMode {
    switch msg.String() {
    case KeyEscape:
        m.shellCommandMode = false
        m.shellCommandBuffer = ""
        return m, nil
    case KeyEnter:
        if m.shellCommandBuffer != "" {
            cmd := m.shellCommandBuffer
            workDir := m.getActivePane().Path()
            m.shellCommandMode = false
            m.shellCommandBuffer = ""
            return m, executeShellCommand(cmd, workDir)
        }
        m.shellCommandMode = false
        return m, nil
    case "backspace":
        if len(m.shellCommandBuffer) > 0 {
            m.shellCommandBuffer = m.shellCommandBuffer[:len(m.shellCommandBuffer)-1]
        }
        return m, nil
    default:
        // Append character (handle runes properly)
        if len(msg.String()) == 1 {
            m.shellCommandBuffer += msg.String()
        }
        return m, nil
    }
}
```

### Message Handler for Command Completion

```go
// model.go Update function

case shellCommandFinishedMsg:
    // Reload both panes to reflect any changes
    m.getActivePane().LoadDirectory()
    m.getInactivePane().LoadDirectory()
    return m, nil
```

### View Rendering

```go
// model.go View function

if m.shellCommandMode {
    // Render minibuffer with "!" prompt
    minibuffer := fmt.Sprintf("!%s", m.shellCommandBuffer)
    // ... render at bottom of screen
}
```

### State Diagram

```
┌─────────┐    !     ┌──────────────┐  Enter    ┌───────────────┐
│ Normal  │────────>│ CommandInput │─────────>│ ShellRunning  │
│  Mode   │<────────│     Mode     │           │ (TUI suspended)│
└─────────┘   Esc    └──────────────┘           └───────┬───────┘
     ^                                                   │
     │                    Command exits                  │
     └───────────────────────────────────────────────────┘
```

### Data Flow

```
User presses "!"
       │
       v
Enter shellCommandMode (buffer = "")
       │
       v
User types command
       │
       v
User presses Enter
       │
       v
executeShellCommand(cmd, activePane.Path())
       │
       v
TUI suspends, shell runs
       │
       v
"Press Enter to continue..." displayed
       │
       v
User presses Enter (handled by shell)
       │
       v
shellCommandFinishedMsg received
       │
       v
Both panes reload
       │
       v
Normal mode restored
```

## Dependencies

- Go standard library `os/exec`
- Bubble Tea's `tea.ExecProcess` for TUI suspension
- Existing minibuffer UI component (for consistent input experience)

## Test Scenarios

### Command Input

- [x] `!` key activates command input mode
- [x] Characters typed are displayed in minibuffer
- [x] Space key can be entered (handled via `tea.KeySpace`)
- [x] `Backspace` deletes last character
- [x] `Escape` cancels input and returns to normal mode
- [x] `Enter` on empty command exits input mode without execution

### Command Execution

- [ ] Command executes in active pane's directory
- [ ] "Press Enter to continue..." is displayed after command
- [ ] Pressing Enter returns to duofm

### Post-Execution

- [ ] Active pane reloads after return
- [ ] Inactive pane reloads after return
- [ ] Cursor position preserved in both panes

### Error Conditions

- [ ] Failed command shows error in shell
- [ ] duofm returns normally after failed command

### Mode Interaction

- [ ] `!` key ignored when dialog is open
- [ ] `!` key ignored when search mode is active

## Success Criteria

- [ ] `!` key enters command input mode
- [ ] Command input displayed with `!` prompt
- [ ] `Escape` cancels command input
- [ ] `Enter` executes non-empty command
- [ ] Empty command does not execute
- [ ] Command runs in active pane's directory
- [ ] "Press Enter to continue..." displayed after execution
- [ ] Both panes reload upon return
- [ ] All existing functionality unaffected
- [ ] Unit tests pass

## Security Considerations

- Commands are executed via `/bin/sh -c`, allowing full shell functionality
- Working directory is controlled by duofm (active pane path)
- No input sanitization required as user explicitly enters commands

## Performance Considerations

- No performance impact during normal operation
- Command execution handled by OS
- Pane reload after return uses existing async mechanism

## Open Questions

None - all requirements have been clarified.
