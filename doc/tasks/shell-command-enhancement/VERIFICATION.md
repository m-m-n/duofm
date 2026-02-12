# Verification Plan: Shell Command Enhancement

## Phase 1: ShellLogger and Modified Shell Execution

### V1.1: Config - shell_log_dir

```bash
# Build succeeds
make build
```

```bash
# Unit tests pass
go test ./internal/config/... -run "ShellLogDir|shellLogDir|shell_log_dir" -v
```

- [ ] `shell_log_dir` loads from config.toml correctly
- [ ] Missing `shell_log_dir` defaults to "/tmp"
- [ ] Config merger adds `shell_log_dir` to existing config files
- [ ] Merger idempotency test passes (all fields present)
- [ ] `LoadConfigDetailed` handles `shell_log_dir` correctly
- [ ] DefaultKeybindings includes `shell_log` with `["Ctrl+L"]`
- [ ] AllActions includes `"shell_log"`

### V1.2: ShellLogger

```bash
go test ./internal/ui/... -run "ShellLogger" -v
```

- [ ] `NewShellLogger("/tmp")` creates logger with path `/tmp/duofm-shell-<PID>.log`
- [ ] `NewShellLogger("")` defaults to `/tmp`
- [ ] `NewShellLogger("/custom/dir")` uses custom dir
- [ ] `NewShellLogger("/non/existent/dir")` falls back to `/tmp` after MkdirAll failure
- [ ] `NewShellLogger` with non-writable dir falls back to `/tmp`
- [ ] `AppendHeader` does not create file before first call
- [ ] `AppendHeader` creates file with 0600 permissions on first call
- [ ] `AppendHeader` writes formatted header with timestamp, command, directory
- [ ] `AppendFooter` writes blank line separator
- [ ] Multiple headers/footers accumulate in same file
- [ ] `HasLog()` returns false before first write
- [ ] `HasLog()` returns true after first write
- [ ] `LogPath()` returns correct path
- [ ] `Close()` deletes the log file
- [ ] `Close()` on never-written logger does not error
- [ ] Unicode characters preserved in log output

### V1.3: Modified executeShellCommand

```bash
go test ./internal/ui/... -run "ExecuteShellCommand|ShellCommandFinished" -v
```

- [ ] `executeShellCommand` no longer includes "Press Enter to continue..."
- [ ] `executeShellCommand` includes `sleep 2` in wrapped command
- [ ] `executeShellCommand` uses `tee -a` to append to log file
- [ ] `executeShellCommand` uses `set -o pipefail` to preserve exit status
- [ ] `executeShellCommand` preserves exit code through sleep (`_exit=$?; sleep 2; exit $_exit`)
- [ ] `shellCommandFinishedMsg` contains command and workDir fields (no output field)
- [ ] Failed command reports error via `shellCommandFinishedMsg.err`
- [ ] Command output visible on terminal during execution
- [ ] TUI resumes after ~2 seconds

### V1.4: Model Integration

```bash
go test ./internal/ui/... -run "ShellCommand" -v
go test ./internal/ui/... -v
```

- [ ] `ModelOptions` includes `ShellLogDir` field
- [ ] `NewModelWithConfig` creates `ShellLogger`
- [ ] Shell command Enter handler calls `AppendHeader` before execution
- [ ] History search Enter handler calls `AppendHeader` before execution
- [ ] `handleShellCommandFinished` calls `AppendFooter`
- [ ] `ActionQuit` calls `shellLogger.Close()`
- [ ] `handleCtrlC` (double-press) calls `shellLogger.Close()`
- [ ] `main.go` passes `ShellLogDir` from config
- [ ] All existing tests still pass (no regressions)

### V1.5: Full Test Suite (Phase 1 Gate)

```bash
make test
```

- [ ] All unit tests pass
- [ ] No regressions in existing shell command tests
- [ ] No regressions in config tests
- [ ] No regressions in merger tests
- [ ] Build succeeds: `make build`

## Phase 2: Shell Log Viewer

### V2.1: ActionShellLog

```bash
go test ./internal/ui/... -run "ActionShellLog|ShellLog" -v
```

- [ ] `ActionShellLog` constant exists in actions.go
- [ ] `actionNames[ActionShellLog]` == `"shell_log"`
- [ ] `nameToAction["shell_log"]` == `ActionShellLog`
- [ ] `ActionFromName("shell_log")` returns `ActionShellLog`

### V2.2: Shell Log Handler

```bash
go test ./internal/ui/... -run "HandleShellLog|ShellLogViewer" -v
```

- [ ] `Ctrl+L` (shell_log action) with no log shows "No shell log" status
- [ ] `Ctrl+L` with log entries opens pager with log file
- [ ] `Ctrl+L` uses `$PAGER` (or less as fallback)
- [ ] `$PAGER` with arguments (e.g., `"less -R"`) works correctly
- [ ] Pager close returns to normal mode
- [ ] Action ignored during dialog (handled by normal action dispatch)
- [ ] Action ignored during search mode (handled by search input handler)
- [ ] Action ignored during shell command mode (handled by shell input handler)

### V2.3: Help Dialog

```bash
go test ./internal/ui/... -run "HelpDialog" -v
```

- [ ] Help dialog contains "Ctrl+L" entry for shell log
- [ ] Help dialog contains "TAB" entry for completion (placeholder for Phase 3)

### V2.4: Full Test Suite (Phase 2 Gate)

```bash
make test
```

- [ ] All unit tests pass
- [ ] All Phase 1 verifications still pass
- [ ] Build succeeds: `make build`

## Phase 3: TAB Completion

### V3.1: Minibuffer Extensions

```bash
go test ./internal/ui/... -run "Minibuffer" -v
```

- [ ] `CursorPos()` returns current cursor position
- [ ] `SetCursorPos(pos)` sets cursor position within bounds

### V3.2: TabCompleter

```bash
go test ./internal/ui/... -run "TabComplete|TabCompleter" -v
```

- [ ] Empty input: no change
- [ ] Single matching command: auto-complete + trailing space
- [ ] Multiple matching commands: common prefix only
- [ ] Single matching file: auto-complete
- [ ] Multiple matching files: common prefix
- [ ] Directory match appends `/`
- [ ] Absolute path completion (`/etc/ho` -> `/etc/hosts`)
- [ ] Relative path completion (`./ma` -> `./main.go`)
- [ ] Parent path completion (`../`)
- [ ] No matches: no change
- [ ] Cursor at middle of word: completes from cursor position
- [ ] Hidden files (dot files) included in completion
- [ ] PATH cache builds correctly
- [ ] PATH cache invalidates on PATH env change
- [ ] Non-existent PATH directory skipped gracefully
- [ ] Duplicate PATH directories handled
- [ ] Empty PATH handled
- [ ] Special characters in filenames (spaces)
- [ ] Symlinks included in completion

### V3.3: TAB Integration

```bash
go test ./internal/ui/... -run "ShellCommandTab|TabKey" -v
```

- [ ] TAB key in shell command mode triggers completion
- [ ] Minibuffer input updated after completion
- [ ] Cursor position updated after completion
- [ ] TAB during history search: no effect (or ignored)
- [ ] TAB with history-recalled command: works correctly
- [ ] Completion uses active pane's directory as cwd

### V3.4: Full Test Suite (Phase 3 Gate)

```bash
make test
```

- [ ] All unit tests pass
- [ ] All Phase 1 and Phase 2 verifications still pass
- [ ] Build succeeds: `make build`

## Final Verification

### Build and Test

```bash
make build && make test
```

- [ ] Build succeeds with no errors
- [ ] All tests pass

### Manual E2E Testing

1. **Shell command auto-return:**
   - Press `!`, type `ls`, press Enter
   - Verify output shows for ~2 seconds, then TUI resumes
   - Verify "Press Enter to continue..." does NOT appear

2. **Shell log viewer:**
   - Execute a command via `!`
   - Press `Ctrl+L`
   - Verify log opens in pager with command header and output
   - Close pager, verify return to normal mode

3. **Multiple commands in log:**
   - Execute 3 different commands via `!`
   - Press `Ctrl+L`
   - Verify all 3 commands appear in order with headers

4. **TAB completion - command:**
   - Press `!`, type `gi`, press TAB
   - Verify completes to common prefix (e.g., `git`)

5. **TAB completion - file path:**
   - Press `!`, type `ls Ma`, press TAB
   - Verify completes to `ls Makefile` (or common prefix if multiple matches)

6. **TAB completion - directory:**
   - Press `!`, type `cd int`, press TAB
   - Verify completes to `cd internal/`

7. **Log cleanup on exit:**
   - Execute a command, verify log file exists in /tmp
   - Quit duofm normally
   - Verify log file is deleted

8. **Ctrl+L with no log:**
   - Start fresh duofm, press `Ctrl+L` without executing any command
   - Verify "No shell log" status message

9. **Existing features preserved:**
   - History (Up/Down in shell mode): works
   - Ctrl+R search: works
   - Esc to cancel: works
   - All other keybindings: unaffected

### Config Verification

10. **shell_log_dir in config:**
    - Edit config.toml, set `shell_log_dir = "/var/tmp"`
    - Execute a command, verify log file in `/var/tmp/duofm-shell-<PID>.log`

11. **Auto-merge:**
    - Delete `shell_log_dir` and `shell_log` keybinding from config.toml
    - Start duofm
    - Verify they are auto-merged into config file

### Regression Checklist

- [ ] Editor opening (`E` key) still works
- [ ] Pager viewing (`V` key) still works
- [ ] File copy/move/delete still works
- [ ] Search (`/`, `Ctrl+F`, `Ctrl+G`) still works
- [ ] Bookmarks still work
- [ ] Config hot-reload still works
- [ ] Per-directory sort settings still work
