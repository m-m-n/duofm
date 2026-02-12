# Implementation Plan: Shell Command Enhancement

## Overview

This plan implements three features in dependency order:
1. **Phase 1**: ShellLogger + modified shell execution (output capture, 2-sec auto-return)
2. **Phase 2**: Shell log viewer (ActionShellLog, keybinding, pager integration)
3. **Phase 3**: TAB completion (TabCompleter, PATH cache, file path completion)

## Phase 1: ShellLogger and Modified Shell Execution

### Step 1.1: Add `shell_log_dir` to Config system

**Files to modify:**

1. `internal/config/config.go`
   - Add `ShellLogDir string` to `Config` struct
   - Add `ShellLogDir *string` to `rawConfig` struct (pointer for nil-check)
   - Update `defaultConfig()`: set `ShellLogDir: DefaultShellLogDir`
   - Add constant: `DefaultShellLogDir = "/tmp"`
   - Update `LoadConfig()`: load `shell_log_dir` from raw (same pattern as `history_limit`)
   - Update `LoadConfigDetailed()` in `reload.go`: handle `shell_log_dir` field

2. `internal/config/merger.go`
   - Add `ShellLogDir *string` to `mergeResult` struct
   - Add `IsMissingShellLogDir()` function
   - Update `hasContent()` to include `ShellLogDir`
   - Update `MergeConfig()`: check and merge `shell_log_dir`
   - Update `generateMergedFile()`: insert `shell_log_dir` as root-level key

3. `internal/config/defaults.go`
   - Add `"shell_log"` to `DefaultKeybindings()` with value `{"Ctrl+L"}`
   - Add `"shell_log"` to `AllActions()` list

**Tests:**
- `internal/config/config_test.go`: Test loading `shell_log_dir` from TOML
- `internal/config/merger_test.go`: Test merging missing `shell_log_dir`

### Step 1.2: Create ShellLogger

**New file: `internal/ui/shell_logger.go`**

```go
package ui

import (
    "fmt"
    "os"
    "path/filepath"
    "time"
)

type ShellLogger struct {
    logPath string
    logFile *os.File
}

func NewShellLogger(logDir string) *ShellLogger {
    if logDir == "" {
        logDir = "/tmp"
    }
    // Ensure directory exists; fallback to /tmp on failure
    if err := os.MkdirAll(logDir, 0755); err != nil {
        logDir = "/tmp"
    }
    pid := os.Getpid()
    logPath := filepath.Join(logDir, fmt.Sprintf("duofm-shell-%d.log", pid))
    return &ShellLogger{logPath: logPath}
}

// AppendHeader writes the command header to the log file (before execution).
// Creates the file with 0600 permissions on first call.
func (sl *ShellLogger) AppendHeader(command, workDir string) error {
    // Create file on first call (0600 permissions)
    // Write formatted header:
    // ════════════════════════════════════════
    // [timestamp] $ command
    // Directory: workDir
    // ════════════════════════════════════════
}

// AppendFooter writes a blank line separator after command output
func (sl *ShellLogger) AppendFooter() error {
    // Write blank line separator
}

func (sl *ShellLogger) LogPath() string { return sl.logPath }

func (sl *ShellLogger) HasLog() bool { return sl.logFile != nil }

func (sl *ShellLogger) Close() error {
    // Close file handle
    // Delete log file (os.Remove)
    // Ignore error if file doesn't exist
}
```

**New file: `internal/ui/shell_logger_test.go`**

Tests:
- `TestNewShellLogger_DefaultDir`: path is `/tmp/duofm-shell-<PID>.log`
- `TestNewShellLogger_CustomDir`: custom directory used
- `TestNewShellLogger_NonExistentDir`: attempts MkdirAll, fallback to /tmp
- `TestShellLogger_AppendHeader_CreatesFile`: file created on first call
- `TestShellLogger_AppendHeader_Format`: header formatted correctly
- `TestShellLogger_AppendFooter`: writes blank line separator
- `TestShellLogger_Multiple_HeaderFooter`: entries accumulate
- `TestShellLogger_HasLog_BeforeWrite`: returns false
- `TestShellLogger_HasLog_AfterWrite`: returns true
- `TestShellLogger_Close_DeletesFile`: file removed
- `TestShellLogger_Close_NoFile`: no error
- `TestShellLogger_FilePermissions`: 0600
- `TestShellLogger_Unicode`: unicode preserved in output

### Step 1.3: Modify `executeShellCommand` for output capture and 2-sec wait

**File to modify: `internal/ui/exec.go`**

1. Update `shellCommandFinishedMsg` (remove output field, add command/workDir):
   ```go
   type shellCommandFinishedMsg struct {
       err     error
       command string
       workDir string
   }
   ```

2. Replace `executeShellCommand`:
   - Remove `"; echo; echo 'Press Enter to continue...'; read _"` suffix
   - Add `logFile` parameter for tee output
   - Use `set -o pipefail` to preserve command exit status through pipe
   - Save exit code before sleep to return correct error status

   ```go
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

   **Execution sequence:**
   1. `ShellLogger.AppendHeader(command, workDir)` writes header to log file
   2. `tee -a` appends command output directly to the same file
   3. `ShellLogger.AppendFooter()` writes blank line separator after return

### Step 1.4: Integrate ShellLogger into Model

**Files to modify:**

1. `internal/ui/model.go`
   - Add `shellLogger *ShellLogger` field to `Model`
   - Add `ShellLogDir string` to `ModelOptions`
   - In `NewModelWithConfig()`: create `ShellLogger` with `opts.ShellLogDir`

2. `internal/ui/model_update_keyboard.go`
   - In `handleShellCommandInput` case `tea.KeyEnter`:
     - Call `m.shellLogger.AppendHeader(command, workDir)` before `executeShellCommand`
     - Pass `m.shellLogger.LogPath()` to `executeShellCommand`
   - Same for `handleHistorySearchInput` case `tea.KeyEnter`

3. `internal/ui/model_update.go`
   - In `handleShellCommandFinished`:
     - Call `m.shellLogger.AppendFooter()` after command completes

4. `internal/ui/model_update_keyboard.go`
   - In `handleAction` case `ActionQuit`:
     - Add `m.shellLogger.Close()` before `tea.Quit`
   - In `handleCtrlC`:
     - Add `m.shellLogger.Close()` before `tea.Quit`

5. `cmd/duofm/main.go`
   - Pass `ShellLogDir: cfg.ShellLogDir` in `ModelOptions`
   - Also add `ShellLogDir: config.DefaultShellLogDir` in the fallback path

**Tests:**
- Test that `handleShellCommandFinished` calls AppendFooter
- Test that quit/Ctrl+C calls Close on logger

## Phase 2: Shell Log Viewer

### Step 2.1: Add ActionShellLog to action system

**File to modify: `internal/ui/actions.go`**

1. Add `ActionShellLog` constant after `ActionEmptyTrash`
2. Add to `actionNames`: `ActionShellLog: "shell_log"`
3. Add to `nameToAction`: `"shell_log": ActionShellLog`

### Step 2.2: Add shell_log handler

**File to modify: `internal/ui/model_update_keyboard.go`**

In `handleAction`:
```go
case ActionShellLog:
    return m.handleShellLog()
```

**File to modify: `internal/ui/model_update_keyboard.go` or new function:**

```go
func (m Model) handleShellLog() (tea.Model, tea.Cmd) {
    if m.shellLogger == nil || !m.shellLogger.HasLog() {
        m.statusMessage = "No shell log"
        return m, statusMessageClearCmd(3 * time.Second)
    }
    return m, openWithViewer(m.shellLogger.LogPath(), m.getActivePane().Path())
}
```

**Note:** `openWithViewer` uses `getPager()` which returns `$PAGER` or `"less"`.
If `$PAGER` contains arguments (e.g., `"less -R"`), `exec.Command` treats it as a
single executable name and fails. Fix `getPager()` to split on spaces:
```go
func getPager() (string, []string) {
    pager := os.Getenv("PAGER")
    if pager == "" {
        return "less", nil
    }
    parts := strings.Fields(pager)
    return parts[0], parts[1:]
}
```
Apply this fix to both `openWithViewer` and `openWithEditor` if applicable.

### Step 2.3: Add shell_log to help dialog

**File to modify: `internal/ui/help_dialog.go`**

In `buildContent()`, after "External Apps" section, add:
```go
lines = append(lines, "  !              : execute shell command")
lines = append(lines, "  Ctrl+L         : view shell command log")
```

(Note: shell command `!` may already be listed. Add `Ctrl+L` line in appropriate section.)

### Step 2.4: Update config defaults and merger for shell_log keybinding

Already done in Step 1.1 (added to `DefaultKeybindings()` and `AllActions()`).

**Tests:**
- `TestActionShellLog_Registered`: verify action exists
- `TestHandleShellLog_NoLog`: shows status message
- `TestHandleShellLog_WithLog`: returns pager command
- `TestHandleShellLog_IgnoredInDialog`: action filtered by normal mode check
- `TestHelpDialogContainsShellLog`: help text includes Ctrl+L

## Phase 3: TAB Completion

### Step 3.1: Create TabCompleter

**New file: `internal/ui/tab_completer.go`**

```go
package ui

type TabCompleter struct {
    pathCache []string // cached PATH executables
    pathEnv   string   // PATH value when cache was built
}

func NewTabCompleter() *TabCompleter

// Complete returns updated input and cursor position after completion
func (tc *TabCompleter) Complete(input string, cursorPos int, cwd string) (string, int)

// isCommandPosition returns true if cursor is in the first word
func isCommandPosition(input string, cursorPos int) bool

// extractWordAtCursor extracts the word being completed and its start position
func extractWordAtCursor(input string, cursorPos int) (word string, wordStart int)

// completeCommand finds matching executables from PATH
func (tc *TabCompleter) completeCommand(prefix string) []string

// completePath finds matching files/directories
func (tc *TabCompleter) completePath(prefix string, cwd string) []string

// buildPathCache scans PATH directories for executables
func (tc *TabCompleter) buildPathCache()

// commonPrefix returns the longest common prefix of candidates
func commonPrefix(candidates []string) string
```

**Algorithm for `Complete`:**
1. Extract word at cursor position and its start index
2. If empty word and not after space, return unchanged
3. Determine if command position or file path position
4. Get candidates (command names or file paths)
5. If 0 candidates: return unchanged
6. If 1 candidate: replace word with candidate + suffix (`/` for dir, ` ` for command)
7. If N candidates: replace word with `commonPrefix(candidates)`
8. Reconstruct input string, calculate new cursor position

**New file: `internal/ui/tab_completer_test.go`**

Tests:
- `TestTabComplete_EmptyInput`: no change
- `TestTabComplete_SingleCommand`: auto-complete + space
- `TestTabComplete_MultipleCommands`: common prefix
- `TestTabComplete_SingleFile`: auto-complete
- `TestTabComplete_MultipleFiles`: common prefix
- `TestTabComplete_DirectoryAppendSlash`: appends `/`
- `TestTabComplete_AbsolutePath`: works with `/etc/`
- `TestTabComplete_RelativePath`: works with `./`
- `TestTabComplete_ParentPath`: works with `../`
- `TestTabComplete_NoMatches`: no change
- `TestTabComplete_MiddleOfWord`: completes from cursor
- `TestTabComplete_PathCacheInvalidation`: rebuilds on PATH change
- `TestTabComplete_HiddenFiles`: dot files included
- `TestTabComplete_SpecialChars`: spaces/quotes in filenames
- `TestIsCommandPosition_FirstWord`: true
- `TestIsCommandPosition_AfterSpace`: false
- `TestExtractWordAtCursor_Beginning`: correct extraction
- `TestExtractWordAtCursor_Middle`: correct extraction
- `TestCommonPrefix_Empty`: empty string
- `TestCommonPrefix_Single`: full string
- `TestCommonPrefix_Multiple`: common part

### Step 3.2: Integrate TAB into shell command mode

**File to modify: `internal/ui/model.go`**
- Add `tabCompleter *TabCompleter` field
- In `NewModelWithConfig()`: create `TabCompleter`

**File to modify: `internal/ui/model_update_keyboard.go`**

In `handleShellCommandInput`, add before the `switch msg.Type`:
```go
if msg.Type == tea.KeyTab {
    return m.handleShellCommandTab()
}
```

New function:
```go
func (m Model) handleShellCommandTab() (tea.Model, tea.Cmd) {
    input := m.minibuffer.Input()
    cursorPos := m.minibuffer.CursorPos() // Need to add this getter to Minibuffer
    cwd := m.getActivePane().Path()

    newInput, newCursorPos := m.tabCompleter.Complete(input, cursorPos, cwd)
    if newInput != input {
        m.minibuffer.SetInput(newInput)
        m.minibuffer.SetCursorPos(newCursorPos) // Need to add this setter to Minibuffer
    }
    return m, nil
}
```

**File to modify: `internal/ui/minibuffer.go`**
- Add `CursorPos() int` getter method
- Add `SetCursorPos(pos int)` setter method

**Tests:**
- Integration test: TAB key in shell command mode triggers completion
- Verify minibuffer updates after TAB

## File Change Summary

### New Files
| File | Phase | Description |
|------|-------|-------------|
| `internal/ui/shell_logger.go` | 1 | ShellLogger struct and methods |
| `internal/ui/shell_logger_test.go` | 1 | ShellLogger unit tests |
| `internal/ui/tab_completer.go` | 3 | TabCompleter struct and methods |
| `internal/ui/tab_completer_test.go` | 3 | TabCompleter unit tests |

### Modified Files
| File | Phase | Changes |
|------|-------|---------|
| `internal/config/config.go` | 1 | Add `ShellLogDir` to Config/rawConfig, loading logic |
| `internal/config/defaults.go` | 1,2 | Add `shell_log` keybinding, AllActions |
| `internal/config/merger.go` | 1 | Add `ShellLogDir` merge support |
| `internal/config/reload.go` | 1 | Handle `shell_log_dir` in detailed load |
| `internal/ui/exec.go` | 1 | Modify `shellCommandFinishedMsg`, `executeShellCommand` |
| `internal/ui/model.go` | 1,3 | Add `shellLogger`, `tabCompleter`, `ShellLogDir` to ModelOptions |
| `internal/ui/model_update_keyboard.go` | 1,2,3 | Shell log action, TAB handler, logger integration |
| `internal/ui/model_update.go` | 1 | Update `handleShellCommandFinished` for logger |
| `internal/ui/actions.go` | 2 | Add `ActionShellLog` |
| `internal/ui/help_dialog.go` | 2 | Add shell log entry |
| `internal/ui/minibuffer.go` | 3 | Add `CursorPos()`, `SetCursorPos()` |
| `cmd/duofm/main.go` | 1 | Pass `ShellLogDir` in ModelOptions |

### Config Merger Test Files
| File | Phase | Changes |
|------|-------|---------|
| `internal/config/merger_test.go` | 1 | Add `shell_log_dir` to idempotency tests |
| `internal/config/config_test.go` | 1 | Test `shell_log_dir` loading |

## Dependency Order

```
Step 1.1 (Config: shell_log_dir)
    │
    v
Step 1.2 (ShellLogger)
    │
    v
Step 1.3 (Modified executeShellCommand) ─── depends on ShellLogger
    │
    v
Step 1.4 (Model integration) ─── depends on 1.1, 1.2, 1.3
    │
    v
Step 2.1 (ActionShellLog) ─── depends on 1.4
    │
    v
Step 2.2 (Shell log handler) ─── depends on 2.1
    │
    v
Step 2.3 (Help dialog) ─── independent, can parallel with 2.2
    │
    v
Step 3.1 (TabCompleter) ─── independent of Phase 2
    │
    v
Step 3.2 (TAB integration) ─── depends on 3.1
```

## Risk Areas

1. **`executeShellCommand` tee approach**: Using shell `tee -a` to capture output is simple but means output capture happens at the shell level, not Go level. If `tee` is not available (unlikely on Linux), this fails. Mitigation: `tee` is part of coreutils, always present.

2. **Terminal state during sleep**: The 2-second `sleep` happens inside `tea.ExecProcess`, so terminal raw mode is off. This is correct behavior - the user sees the output in normal terminal mode during the sleep.

3. **Existing test compatibility**: Many tests use `NewModel()` which calls `NewModelWithConfig` with defaults. Since `ShellLogDir` defaults to empty (which ShellLogger treats as `/tmp`), existing tests should not be affected. The ShellLogger is always created but only writes files when commands are executed.

4. **Config merger idempotency**: The `merger_test.go` idempotency tests require all config fields to be present. Must update these tests to include `shell_log_dir`.

5. **TAB completion performance (NFR1)**: TAB completion must respond within 200ms for directories with up to 10000 entries. PATH cache helps for command completion. File completion relies on `os.ReadDir` which is fast for single directories.

6. **tee pipe and isatty()**: Commands piped through `tee` see `isatty()=false`, so some commands may suppress color output. This is an acceptable trade-off for log capture.

7. **`$PAGER` with arguments**: If `$PAGER` is set to e.g. `"less -R"`, the existing `exec.Command(getPager(), path)` approach will fail. Fix `getPager()` to split on spaces (see Step 2.2).
