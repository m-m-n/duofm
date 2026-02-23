# Feature: Background Shell Command Execution

## Overview

Add background shell command execution to duofm. Users can run shell commands in the background while continuing to use all duofm file management features. Command output is displayed in real-time in the bottom third of the active pane.

## Objectives

- Enable background shell command execution without suspending the TUI
- Display real-time command output in a split pane area
- Allow full duofm operation during background command execution
- Provide a mechanism to cancel running background commands

## User Stories

### US1: Background Command Execution

As a user, I want to run a shell command in the background, so that I can continue file management while the command runs.

**Acceptance Criteria:**
- [ ] `!` in shell command mode switches to background mode (pink `!` prompt)
- [ ] Backspace on empty input returns to normal shell command mode
- [ ] Enter executes the command in the background
- [ ] Output displays in the bottom 1/3 of the active pane in real-time
- [ ] All duofm operations remain available during execution

### US2: Background Command Cancellation

As a user, I want to cancel a running background command, so that I can stop long-running or mistaken commands.

**Acceptance Criteria:**
- [ ] TAB focuses the output area when background command is running
- [ ] Ctrl+C in focused output area cancels the background command
- [ ] Output area closes after cancellation

### US3: Automatic Cleanup

As a user, I want the output area to close automatically after the command finishes, so that the pane returns to full file list display.

**Acceptance Criteria:**
- [ ] Output area remains for 2 seconds after command completion
- [ ] Output area closes automatically after the 2-second delay
- [ ] Both panes reload after the output area closes

## Technical Requirements

### Functional Requirements

#### Background Mode Activation

- **FR1:** In shell command mode, pressing `!` shall switch to background mode
- **FR2:** Background mode prompt shall display `!` in pink color (lipgloss color)
- **FR3:** Backspace on empty input in background mode shall return to normal shell command mode
- **FR4:** Escape in background mode shall cancel and return to normal mode
- **FR5:** All existing shell command mode features (history navigation, Ctrl+R, TAB completion) shall work in background mode

#### Background Command Execution

- **FR6:** Enter in background mode shall execute the command as a background process
- **FR7:** Command shall execute via `/bin/sh -c "<command>"`
- **FR8:** Working directory shall be the active pane's directory at the time of command input
- **FR9:** TUI shall NOT suspend during background command execution (no `tea.ExecProcess`)
- **FR10:** Both stdout and stderr shall be captured and displayed
- **FR11:** Only one background command can run at a time
- **FR12:** Attempting to start a new shell command while a background command is running shall display a warning in the status bar ("Background command running")

#### Output Display Area

- **FR13:** During execution, the active pane's bottom 1/3 shall display command output
- **FR14:** Output shall auto-scroll to the bottom as new lines arrive (tail -f behavior)
- **FR15:** The file list area above the output area shall remain fully interactive
- **FR16:** Output area shall display the running command at the top as a header

#### Auto-Close Behavior

- **FR17:** After command completion, output area shall remain visible for 2 seconds
- **FR18:** After the 2-second delay, output area shall close automatically
- **FR19:** Both panes shall reload when the output area closes

#### Output Area Focus

- **FR20:** TAB key shall focus the output area when a background command is running on that pane
- **FR21:** While output area is focused, Ctrl+C shall send SIGTERM/SIGKILL to the background process
- **FR22:** After cancellation or TAB/Esc, focus returns to the file list
- **FR23:** While output area is focused, all other keys are ignored (only Ctrl+C and TAB/Esc work)

#### Pane Interaction During Execution

- **FR24:** Pane switching (move_left/move_right) shall work normally during background execution
- **FR25:** File operations (copy, move, delete, rename, etc.) shall work normally during background execution
- **FR26:** The output area is tied to the pane that launched the command (not the currently active pane)
- **FR27:** When the launching pane is not active, the output area is hidden but the process continues

#### Shell Log Integration

- **FR28:** Background command output shall be recorded in the shell log (ShellLogger)
- **FR29:** Background commands shall appear in shell log viewer (Ctrl+L) alongside foreground commands

### Non-Functional Requirements

- **NFR1 - Performance:** Output display update latency shall be under 100ms
- **NFR2 - Performance:** Background process shall not cause noticeable lag in TUI responsiveness
- **NFR3 - Compatibility:** All existing shell command functionality (foreground execution, history, TAB completion) shall remain intact
- **NFR4 - Compatibility:** All existing keybindings shall remain unaffected
- **NFR5 - Reliability:** Background process must be properly cleaned up on duofm exit

## Implementation Approach

### Architecture

**Component Diagram:**
```
┌──────────────────────────────────────────────────────────────┐
│                         Model                                │
│                                                              │
│  ┌─────────────────────┐  ┌────────────────────────────────┐│
│  │  BackgroundRunner    │  │       OutputBuffer             ││
│  │                      │  │                                ││
│  │  - cmd *exec.Cmd     │  │  - lines []string             ││
│  │  - ctx context.Ctx   │  │  - maxLines int               ││
│  │  - cancel func()     │  │  + Append(line string)        ││
│  │  - running bool      │  │  + Lines() []string           ││
│  │  - pane PanePosition │  │  + Clear()                    ││
│  │  + Start(cmd,dir)    │  │                                ││
│  │  + Cancel()          │  │                                ││
│  │  + IsRunning() bool  │  │                                ││
│  └─────────────────────┘  └────────────────────────────────┘│
│                                                              │
│  State: bgMode bool (in shell command mode)                  │
│  State: bgOutputFocused bool (output area has focus)         │
│                                                              │
│  Messages:                                                   │
│    bgOutputMsg{line string}     - new output line            │
│    bgCommandDoneMsg{err error}  - command completed          │
│    bgAutoCloseMsg{}             - 2-sec timer fired          │
└──────────────────────────────────────────────────────────────┘
```

### State Machine

```
┌─────────┐    !     ┌──────────────┐    !     ┌──────────────┐
│ Normal  │────────>│ ShellCommand │────────>│ Background   │
│  Mode   │<────────│    Mode      │<────────│    Mode      │
└─────────┘   Esc    └──────┬───────┘  Bksp   └──────┬───────┘
     ^                      │ Enter(fg)               │ Enter(bg)
     │                      v                         v
     │               ┌──────────────┐         ┌──────────────┐
     │               │ FG Running   │         │ BG Running   │
     │               │(TUI suspend) │         │(TUI active)  │
     │               └──────┬───────┘         └──────┬───────┘
     │                      │                   TAB  │  │ Done
     │                      │                   ┌────v──v────┐
     │                      │                   │ BG Output  │
     │                      │                   │  Focused   │
     │                      │                   └──────┬─────┘
     │                      │                     Ctrl+C│TAB/Esc
     └──────────────────────┴──────────────────────────┘
                         (cleanup + reload)
```

### Data Flow

#### Background Command Execution Flow
```
User presses Enter in background mode
       │
       v
Create BackgroundRunner with context
       │
       v
Start goroutine: exec.CommandContext("/bin/sh", "-c", command)
       │
       ├── Pipe stdout+stderr to scanner goroutine
       │        │
       │        └── For each line: send bgOutputMsg via channel → tea.Cmd
       │
       v
ShellLogger.AppendHeader(command, workDir)
       │
       v
Command runs... (output streams in real-time)
       │
       v
Command exits → send bgCommandDoneMsg
       │
       v
ShellLogger.AppendFooter()
       │
       v
Start 2-second timer → send bgAutoCloseMsg
       │
       v
Close output area, reload both panes
```

#### Output Display Flow
```
bgOutputMsg received
       │
       v
Append line to OutputBuffer
       │
       v
View() re-renders:
  - Top 2/3: file list (via Pane.View)
  - Bottom 1/3: OutputBuffer.Lines() (auto-scrolled)
```

### Message Types

```go
// bgOutputMsg delivers a line of output from the background command
type bgOutputMsg struct {
    line string
}

// bgCommandDoneMsg signals that the background command has finished
type bgCommandDoneMsg struct {
    err     error
    command string
    workDir string
}

// bgAutoCloseMsg signals that the 2-second post-completion timer has fired
type bgAutoCloseMsg struct{}
```

### Model State Additions

```go
type Model struct {
    // ... existing fields ...

    // Background shell command
    bgMode           bool           // true when in background input mode (pink prompt)
    bgRunner         *BackgroundRunner // manages the background process
    bgOutputBuffer   *OutputBuffer  // circular buffer for output lines
    bgOutputFocused  bool           // true when output area has keyboard focus
    bgCommandDir     string         // working directory for the bg command
    bgClosing        bool           // true during 2-sec auto-close delay
}
```

### Key Handler Changes

In `handleShellCommandInput`:
```
!  key → if not bgMode: set bgMode=true
         if bgMode: append '!' to input (normal char)

Backspace → if bgMode and input empty: set bgMode=false
            else: normal backspace

Enter → if bgMode: start background execution
         else: start foreground execution (existing)
```

In `handleKeyInput` (normal mode):
```
TAB → if bgRunner.IsRunning() and activePane == bgRunner.pane:
        set bgOutputFocused=true
      else: normal TAB handling

! key → if bgRunner.IsRunning():
          show "Background command running" in status bar
        else: enter shell command mode (existing)
```

In new `handleBgOutputFocused`:
```
Ctrl+C → bgRunner.Cancel(), close output area
TAB/Esc → bgOutputFocused=false (return to file list)
other keys → ignore
```

### View Changes

In `View()` or pane rendering:
```
if bgRunner.IsRunning() or bgClosing:
    paneHeight = totalPaneHeight
    if this pane launched the bg command:
        fileListHeight = paneHeight * 2/3
        outputHeight = paneHeight * 1/3
        render file list in top portion
        render output buffer in bottom portion
    else:
        render full file list (normal)
```

### File Structure

```
internal/ui/
├── background_runner.go          # BackgroundRunner: process management
├── background_runner_test.go     # Unit tests
├── output_buffer.go              # OutputBuffer: circular line buffer
├── output_buffer_test.go         # Unit tests
├── model.go                      # Add bg* fields
├── model_update_keyboard.go      # Handle bg mode input, TAB focus, Ctrl+C
├── model_update.go               # Handle bgOutputMsg, bgCommandDoneMsg, bgAutoCloseMsg
├── model_view.go                 # Render split pane with output area
├── pane_render.go                # Adjust pane rendering for split view
├── exec.go                       # Add startBackgroundCommand function
└── help_dialog.go                # Update help text
```

### Dependencies

**Internal Dependencies:**
- Existing shell command mode: extended with background mode toggle
- Existing ShellLogger: reused for logging background command output
- Existing Minibuffer: reused for command input
- Existing shell history: reused for history navigation

**External Dependencies:**
- Go standard library `os/exec` with `context.Context` for cancellation
- Go standard library `io` for pipe handling
- Go standard library `bufio` for line scanning

## Test Scenarios

### Unit Tests

#### Background Mode Activation
- [ ] `!` in shell command mode sets bgMode=true
- [ ] `!` in bgMode appends `!` character to input
- [ ] Backspace on empty input in bgMode sets bgMode=false
- [ ] Backspace on non-empty input in bgMode deletes character normally
- [ ] Escape in bgMode cancels and returns to normal mode
- [ ] History navigation (Up/Down) works in bgMode
- [ ] Ctrl+R history search works in bgMode
- [ ] TAB completion works in bgMode

#### Background Command Execution
- [ ] Enter in bgMode starts background process
- [ ] Enter in bgMode with empty input does nothing
- [ ] Background command runs in specified working directory
- [ ] stdout is captured and sent as bgOutputMsg
- [ ] stderr is captured and sent as bgOutputMsg
- [ ] bgCommandDoneMsg is sent when command completes
- [ ] bgCommandDoneMsg includes error for failed commands
- [ ] Output is recorded in ShellLogger

#### Output Buffer
- [ ] Append adds lines to buffer
- [ ] Buffer respects maxLines limit (circular)
- [ ] Lines() returns all buffered lines in order
- [ ] Clear() empties the buffer
- [ ] Empty buffer returns empty slice
- [ ] Unicode characters are preserved

#### Output Display
- [ ] Output area renders in bottom 1/3 of pane
- [ ] File list renders in top 2/3 of pane
- [ ] Auto-scroll shows latest lines
- [ ] Output area shows command header
- [ ] Output area hidden when bg command not on this pane

#### Auto-Close
- [ ] 2-second timer starts after bgCommandDoneMsg
- [ ] bgAutoCloseMsg closes output area
- [ ] Both panes reload after close
- [ ] Output buffer is cleared after close

#### Output Area Focus
- [ ] TAB focuses output area when bg running on active pane
- [ ] TAB does nothing when no bg running
- [ ] Ctrl+C in focused mode cancels bg command
- [ ] TAB in focused mode returns to file list
- [ ] Esc in focused mode returns to file list
- [ ] Other keys are ignored in focused mode

#### Cancellation
- [ ] Cancel sends signal to background process
- [ ] Process group is terminated (child processes too)
- [ ] Output area closes after cancellation
- [ ] Both panes reload after cancellation
- [ ] ShellLogger records partial output

#### Concurrent Operation
- [ ] File operations work during bg execution
- [ ] Pane switching works during bg execution
- [ ] Search works during bg execution
- [ ] Dialog operations work during bg execution
- [ ] Starting new shell command while bg running shows warning

### E2E Tests
**Existing E2E tests**: `test/e2e/` (Docker + tmux)
**Run command**: `make test-e2e`
- [ ] Existing E2E tests pass without regression
- [ ] Execute background command and verify output display
- [ ] Cancel background command with Ctrl+C
- [ ] Verify pane operations during background execution

### Edge Cases
- [ ] Command that produces no output
- [ ] Command that produces very large output (>10000 lines)
- [ ] Command that exits immediately
- [ ] Command killed by signal
- [ ] duofm quit while background command is running
- [ ] Very long single line of output (>terminal width)
- [ ] Binary output (non-UTF8)
- [ ] Terminal resize during background execution

## Security Considerations

- Commands are executed via `/bin/sh -c`, same as existing foreground execution
- Working directory is controlled by duofm (active pane path at input time)
- Background process is killed on duofm exit to prevent orphans
- Process group kill ensures child processes are also terminated

## Performance Considerations

- Output lines are buffered in a circular buffer to limit memory usage
- View re-rendering is triggered by tea.Msg, following Bubble Tea's event loop
- Large output does not block TUI (output reading runs in a separate goroutine)
- Channel-based message passing avoids lock contention with the TUI thread

## Implementation Phases

### Phase 1: Core Background Execution
**Goals:** Basic background command execution and output capture
**Deliverables:**
- BackgroundRunner with process management
- OutputBuffer for line storage
- bgOutputMsg/bgCommandDoneMsg message handling
- Basic output display in pane

### Phase 2: UI Integration
**Goals:** Complete UI with focus, cancellation, and auto-close
**Deliverables:**
- Background mode toggle in shell command mode (`!` → pink prompt)
- Output area focus with TAB
- Ctrl+C cancellation
- 2-second auto-close timer
- Shell log integration

### Phase 3: Polish
**Goals:** Edge cases and robustness
**Deliverables:**
- Process group cleanup
- Graceful shutdown on duofm exit
- Help dialog updates
- Edge case handling (binary output, resize, etc.)

## References

- Existing shell command spec: `doc/tasks/shell-command-execution/SPEC.md`
- Shell command enhancement spec: `doc/tasks/shell-command-enhancement/SPEC.md`
- Current implementation: `internal/ui/exec.go`, `internal/ui/model_update_keyboard.go`
