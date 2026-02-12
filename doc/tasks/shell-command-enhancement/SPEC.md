# Feature: Shell Command Enhancement

## Overview

Enhance duofm's shell command mode with three improvements:
1. TAB completion for file paths and command names
2. Auto-return after shell command execution (replacing "Press Enter to continue...")
3. Shell log viewer for reviewing command output history

## Objectives

- Provide TAB completion for file paths and PATH commands in shell command mode
- Remove "Press Enter to continue..." and auto-return after 2 seconds
- Capture command output and provide a keybinding to review shell logs via pager

## User Stories

### US1: TAB Completion in Shell Command Mode

As a user, I want to press TAB while typing a shell command to auto-complete file paths and command names, so that I can type commands faster.

**Acceptance Criteria:**
- [ ] TAB at command position completes executable names from PATH
- [ ] TAB at argument position completes file/directory names from active pane directory
- [ ] Single candidate: auto-complete the full name
- [ ] Multiple candidates: complete common prefix
- [ ] Directory completion appends `/`

### US2: Auto-Return After Shell Command

As a user, I want duofm to auto-return after executing a shell command, so that I don't have to press Enter to continue.

**Acceptance Criteria:**
- [ ] Command output displays for 2 seconds after execution
- [ ] TUI resumes automatically after 2 seconds
- [ ] Both panes refresh upon return
- [ ] Command output (stdout + stderr) is captured and stored in session log

### US3: Shell Log Viewer

As a user, I want to view the output of previously executed shell commands, so that I can review results without re-running commands.

**Acceptance Criteria:**
- [ ] Configurable keybinding (default: Ctrl+L) opens shell log in $PAGER
- [ ] All commands executed during the session are shown
- [ ] Each log entry includes: command, working directory, output
- [ ] Log is displayed with $PAGER (default: less) in full-screen

## Technical Requirements

### Functional Requirements

#### TAB Completion

- **FR1:** TAB key in shell command mode shall trigger completion
- **FR2:** If cursor is at the first word position (before any space), complete from PATH executables
- **FR3:** If cursor is after a space, complete file/directory paths relative to active pane directory
- **FR4:** Path completion shall support relative paths (e.g., `./`, `../`, partial names)
- **FR5:** Path completion shall support absolute paths (e.g., `/etc/`)
- **FR6:** Single match shall auto-complete the full name
- **FR7:** Multiple matches shall complete the common prefix
- **FR8:** Directory completion shall append trailing `/`
- **FR9:** Executable completion shall append trailing space
- **FR10:** Completion shall be case-sensitive (matching filesystem behavior)

#### Auto-Return After Execution

- **FR11:** Shell command execution shall no longer append `echo 'Press Enter to continue...'; read _`
- **FR12:** After command finishes, wait 2 seconds before returning to TUI
- **FR13:** Both stdout and stderr of the command shall be captured to a buffer
- **FR14:** The captured output shall also be displayed on the terminal during execution
- **FR15:** After the 2-second wait, TUI resumes and both panes reload

#### Shell Log Viewer

- **FR16:** A new action `shell_log` shall be added to the action system
- **FR17:** Default keybinding for `shell_log` shall be `ctrl+l`
- **FR18:** Each log entry shall contain: executed command, working directory, combined output (stdout + stderr), timestamp
- **FR19:** Shell log shall be displayed via `$PAGER` (fallback: `less`) using `tea.ExecProcess`
- **FR20:** Log entries shall be formatted with headers for each command
- **FR21:** `shell_log` action shall be available only in normal mode (not during dialogs/search)

#### Log File Management

- **FR22:** Log file path shall be `<dir>/duofm-shell-<PID>.log` where PID is duofm's process ID
- **FR23:** Default log directory shall be `/tmp`
- **FR24:** Log directory shall be configurable via `shell_log_dir` in config.toml
- **FR25:** Log file shall be created on first shell command execution (not on startup)
- **FR26:** Each command execution shall append output to the log file
- **FR27:** Shell log viewer (`Ctrl+L`) shall open the log file directly with pager
- **FR28:** Log file shall be deleted on normal program exit (quit action, Ctrl+C double-press)
- **FR29:** If program terminates abnormally (crash, SIGKILL), log file may remain (acceptable)

### Non-Functional Requirements

- **NFR1 - Performance:** TAB completion shall respond within 200ms for directories with up to 10000 entries
- **NFR2 - Performance:** PATH scanning for command completion shall cache results (invalidated on PATH change)
- **NFR3 - Disk:** Shell log file has no explicit size limit (relies on /tmp cleanup policies)
- **NFR4 - Compatibility:** All existing shell command functionality (history, Ctrl+R, Up/Down) shall remain intact
- **NFR5 - Compatibility:** All existing keybindings shall remain unaffected

## Implementation Approach

### Architecture

**Component Diagram:**
```
┌──────────────────────────────────────────────────────────────┐
│                         Model                                │
│                                                              │
│  ┌─────────────────────┐  ┌────────────────────────────────┐│
│  │   TabCompleter      │  │       ShellLogger              ││
│  │                     │  │                                ││
│  │  - pathCache        │  │  - logFile *os.File            ││
│  │  - pathEnv          │  │  - logPath string              ││
│  │  + Complete(input,  │  │  + AppendHeader(cmd, dir)      ││
│  │    cursorPos, cwd)  │  │  + AppendFooter()              ││
│  │    -> (string, int) │  │  + LogPath() string            ││
│  │                     │  │  + Close() error (delete file) ││
│  └─────────────────────┘  └────────────────────────────────┘│
│                                                              │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              Modified executeShellCommand                ││
│  │                                                          ││
│  │  - Output via tee -a to log file + terminal              ││
│  │  - Wait 2 seconds after command exits                    ││
│  │  - Return shellCommandFinishedMsg (command, workDir)     ││
│  └─────────────────────────────────────────────────────────┘│
│                                                              │
│           Log file: /tmp/duofm-shell-<PID>.log              │
│           (configurable dir via shell_log_dir)               │
└──────────────────────────────────────────────────────────────┘
```

### State Machine (Shell Command Mode with TAB)

```
┌─────────┐    !     ┌──────────────────┐  TAB    ┌───────────────┐
│ Normal  │────────>│ CommandInput     │────────>│ Complete      │
│  Mode   │<────────│     Mode         │<────────│ (same mode)   │
└─────────┘   Esc    └───────┬──────────┘         └───────────────┘
     ^                       │ Enter
     │                       v
     │               ┌──────────────────┐
     │               │ Shell Running    │
     │               │ (output captured)│
     │               └───────┬──────────┘
     │                       │ Command exits
     │                       v
     │               ┌──────────────────┐
     │               │ Wait 2 seconds   │
     │               └───────┬──────────┘
     │                       │ Timer expires
     └───────────────────────┘
```

### Data Flow

#### TAB Completion Flow
```
User presses TAB in shell command mode
       │
       v
Parse input: extract word under cursor
       │
       v
Determine completion type (command name or file path)
       │
       ├── Command position (first word)
       │   └── Scan PATH directories for matching executables
       │
       └── Argument position (after space)
           └── Scan filesystem for matching files/directories
       │
       v
Calculate candidates
       │
       ├── 0 candidates → No change
       ├── 1 candidate → Auto-complete full name
       └── N candidates → Complete common prefix
       │
       v
Update minibuffer input and cursor position
```

#### Shell Command Execution Flow (Modified)
```
User presses Enter (command non-empty)
       │
       v
Add command to history
       │
       v
ShellLogger.AppendHeader(command, workDir)
       │
       v
Suspend TUI via tea.ExecProcess
       │
       v
Execute /bin/sh -c "{ <command>; } 2>&1 | tee -a <logFile>; sleep 2"
 (stdout/stderr → terminal + log file via tee)
       │
       v
Return shellCommandFinishedMsg (command, workDir)
       │
       v
ShellLogger.AppendFooter()
       │
       v
Reload both panes
```

#### Shell Log Viewer Flow
```
User presses Ctrl+L (shell_log action)
       │
       v
Check if log file exists (logPath)
       │
       ├── No log file → Show "No shell log" in status bar
       │
       └── Log file exists
           │
           v
       Open log file with $PAGER via tea.ExecProcess
           │
           v
       User closes pager
           │
           v
       Return to normal mode
```

#### Log File Lifecycle
```
duofm starts
       │
       v
ShellLogger created (logPath = /tmp/duofm-shell-<PID>.log, file not created yet)
       │
       v
First shell command executed
       │
       v
Log file created, entry appended
       │
       v
Subsequent commands → entries appended to same file
       │
       v
duofm exits normally → ShellLogger.Close() → delete log file
```

### File Structure

```
internal/ui/
├── tab_completer.go           # TabCompleter: PATH and file path completion
├── tab_completer_test.go      # Unit tests
├── shell_logger.go            # ShellLogger: session log storage
├── shell_logger_test.go       # Unit tests
├── exec.go                    # Modified: capture output, 2-sec wait
├── model.go                   # Add tabCompleter, shellLogger fields
├── model_update_keyboard.go   # Handle TAB key in shell command mode
├── model_update.go            # Handle shellCommandFinishedMsg with log
├── actions.go                 # Add ActionShellLog
├── defaults.go                # Add default keybinding for shell_log
└── help_dialog.go             # Update help text
internal/config/
├── config.go                  # Add shell_log keybinding, shell_log_dir setting
```

### New Types

#### TabCompleter (internal/ui/tab_completer.go)

```go
// TabCompleter provides TAB completion for shell command mode
type TabCompleter struct {
    pathCache     []string  // Cached PATH executable names
    pathEnv       string    // PATH value when cache was built
}

// NewTabCompleter creates a new TabCompleter
func NewTabCompleter() *TabCompleter

// Complete takes the current input, cursor position, and working directory,
// returns the completed input and new cursor position.
// If no completion is possible, returns the original input unchanged.
func (tc *TabCompleter) Complete(input string, cursorPos int, cwd string) (string, int)

// completeCommand completes executable names from PATH
func (tc *TabCompleter) completeCommand(prefix string) []string

// completePath completes file/directory paths relative to cwd
func (tc *TabCompleter) completePath(prefix string, cwd string) []string

// commonPrefix returns the longest common prefix among candidates
func commonPrefix(candidates []string) string
```

#### ShellLogger (internal/ui/shell_logger.go)

```go
// ShellLogger manages the session log file at <dir>/duofm-shell-<PID>.log
type ShellLogger struct {
    logPath string    // Full path to log file
    logFile *os.File  // Open file handle (nil until first write)
}

// NewShellLogger creates a new ShellLogger with the given directory.
// The log file is not created until the first AppendHeader call.
// logDir defaults to "/tmp" if empty.
// If logDir does not exist, attempts os.MkdirAll; on failure falls back to "/tmp".
func NewShellLogger(logDir string) *ShellLogger

// AppendHeader writes the command header to the log file (before execution).
// Creates the file with 0600 permissions on first call.
func (sl *ShellLogger) AppendHeader(command, workDir string) error

// AppendFooter writes a blank line separator after command output.
func (sl *ShellLogger) AppendFooter() error

// LogPath returns the log file path (used by tee -a)
func (sl *ShellLogger) LogPath() string

// HasLog returns true if log file has been created (first AppendHeader call)
func (sl *ShellLogger) HasLog() bool

// Close closes the file handle and deletes the log file.
// Called on normal program exit.
func (sl *ShellLogger) Close() error
```

#### Modified executeShellCommand (internal/ui/exec.go)

```go
type shellCommandFinishedMsg struct {
    err     error
    command string
    workDir string
}

// executeShellCommand executes a shell command with output captured via tee,
// waits 2 seconds, then returns to TUI.
// logFile is the path for tee -a to append output.
func executeShellCommand(command, workDir, logFile string) tea.Cmd {
    wrapped := fmt.Sprintf(
        "set -o pipefail; { %s; } 2>&1 | tee -a %q; _exit=$?; sleep 2; exit $_exit",
        command, logFile)
    shellCmd := exec.Command("/bin/sh", "-c", wrapped)
    shellCmd.Dir = workDir
    return tea.ExecProcess(shellCmd, func(err error) tea.Msg {
        return shellCommandFinishedMsg{err: err, command: command, workDir: workDir}
    })
}
```

### Config Changes

```toml
# config.toml
# Directory for shell log file (default: "/tmp")
shell_log_dir = "/tmp"
```

```go
// config.go
type Config struct {
    // ... existing fields ...
    ShellLogDir string `toml:"shell_log_dir"` // Default: "/tmp"
}
```

### Model Changes

```go
type Model struct {
    // ... existing fields ...

    // TAB completion
    tabCompleter *TabCompleter

    // Shell log
    shellLogger *ShellLogger
}
```

### Action System Changes

```go
// actions.go
const (
    // ... existing actions ...
    ActionShellLog  // View shell command log
)

// actionNames
ActionShellLog: "shell_log",

// nameToAction
"shell_log": ActionShellLog,
```

### Default Keybinding

```go
// defaults.go - DefaultKeybindings()
"ctrl+l": ["shell_log"]
```

### Log Output Format

```
════════════════════════════════════════════════════════════════
[2024-01-15 14:30:05] $ ls -la
Directory: /home/user/projects
════════════════════════════════════════════════════════════════
total 48
drwxr-xr-x  5 user user 4096 Jan 15 14:30 .
drwxr-xr-x 20 user user 4096 Jan 15 10:00 ..
-rw-r--r--  1 user user  256 Jan 15 14:30 main.go

════════════════════════════════════════════════════════════════
[2024-01-15 14:30:10] $ go build ./...
Directory: /home/user/projects
════════════════════════════════════════════════════════════════

```

### Help Dialog Updates

Add entries:
- `TAB` → Complete command/path (in shell command mode)
- `Ctrl+L` → View shell log

## Test Scenarios

### Unit Tests

#### TAB Completion

- [ ] TAB on empty input does nothing
- [ ] TAB completes single matching command from PATH
- [ ] TAB completes common prefix of multiple matching commands
- [ ] TAB completes single matching file path
- [ ] TAB completes common prefix of multiple matching file paths
- [ ] TAB appends `/` for directory completion
- [ ] TAB appends space for single command completion
- [ ] TAB handles absolute paths (`/etc/hos` → `/etc/hosts`)
- [ ] TAB handles relative paths (`./ma` → `./main.go`)
- [ ] TAB handles `../` paths
- [ ] TAB with no matches does nothing
- [ ] TAB at middle of word completes from cursor position
- [ ] PATH cache invalidates when PATH env changes
- [ ] Hidden files (dot files) are included in completion

#### Shell Command Execution (Modified)

- [ ] Command executes and output is captured
- [ ] Both stdout and stderr are captured
- [ ] Command output is displayed on terminal during execution
- [ ] TUI resumes after 2 seconds
- [ ] Captured output is stored in ShellLogEntry
- [ ] Both panes refresh on return
- [ ] Failed command output is also captured
- [ ] Empty output command works correctly

#### Shell Logger

- [ ] NewShellLogger creates logger with correct path (`/tmp/duofm-shell-<PID>.log`)
- [ ] NewShellLogger with custom dir uses specified directory
- [ ] Append creates log file on first call
- [ ] Append writes formatted entry to file
- [ ] Multiple Append calls accumulate in same file
- [ ] HasLog returns false before first Append
- [ ] HasLog returns true after Append
- [ ] LogPath returns correct file path
- [ ] Close deletes the log file
- [ ] Close on logger with no writes does not error
- [ ] Unicode characters in output are preserved
- [ ] Log file has 0600 permissions

#### Shell Log Viewer

- [ ] Ctrl+L with no log entries shows "No shell log" status message
- [ ] Ctrl+L with log entries opens pager with log file
- [ ] Ctrl+L ignored when dialog is active
- [ ] Ctrl+L ignored when search mode is active
- [ ] Ctrl+L ignored when shell command mode is active
- [ ] Pager closes and returns to normal mode
- [ ] Log file is deleted on normal quit
- [ ] Config shell_log_dir changes log file directory

### Integration Tests

- [ ] Type `!ls`, press TAB, verify completion
- [ ] Execute command, verify auto-return after 2 seconds
- [ ] Execute command, press Ctrl+L, verify log in pager
- [ ] Execute multiple commands, verify all appear in log
- [ ] TAB completion works with command history (Up then TAB)

### Edge Cases

- [ ] TAB with special characters in file names (spaces, quotes)
- [ ] TAB with symlinks
- [ ] TAB on non-existent PATH directory (skip gracefully)
- [ ] Very long command output (>1MB) is captured
- [ ] Command that produces no output
- [ ] Command interrupted by signal
- [ ] PATH with duplicate directories
- [ ] Empty PATH
- [ ] Ctrl+L with Ctrl+L already bound to another action (config override)
- [ ] shell_log_dir set to non-existent directory (create or error)
- [ ] shell_log_dir with no write permission (fallback to /tmp)

## Security Considerations

- PATH scanning only reads directory listings, no execution of discovered commands
- Log file created with 0600 permissions (owner read/write only)
- Log file deleted on normal program exit
- Log file may persist after abnormal exit (acceptable; located in /tmp which has periodic cleanup)
- No sensitive data filtering on command output (user responsibility)

## Performance Considerations

- PATH executable list is cached; cache invalidated only when PATH changes
- File path completion uses `os.ReadDir` (does not recurse)
- Shell log appended directly to file (no in-memory accumulation)
- Pager opens log file directly (no temp file copy needed)

## Known Limitations

- tee pipe causes `isatty()` to return false for command output, so some commands may suppress color output (acceptable trade-off for log capture)

## Implementation Phases

### Phase 1: Auto-Return and Output Capture

**Deliverables:**
- Modified `executeShellCommand` with tee and 2-second wait
- `ShellLogger` struct and methods (AppendHeader/AppendFooter API)
- Unit tests

### Phase 2: Shell Log Viewer

**Deliverables:**
- `ActionShellLog` action and default keybinding
- Log viewer via pager (temp file approach)
- Help dialog update
- Config default keybinding
- Unit tests

### Phase 3: TAB Completion

**Deliverables:**
- `TabCompleter` struct with PATH and file completion
- TAB key handling in shell command mode
- PATH caching
- Unit tests

## References

- Existing shell command spec: `doc/tasks/shell-command-execution/SPEC.md`
- Shell history spec: `doc/tasks/shell-command-history/SPEC.md`
- History enhancement spec: `doc/tasks/shell-command-history-enhancement/SPEC.md`
- Current implementation: `internal/ui/exec.go`, `internal/ui/model_update_keyboard.go`
