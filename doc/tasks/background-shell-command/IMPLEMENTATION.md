# Implementation Plan: Background Shell Command Execution

## Overview

Add background shell command execution to duofm. Users enter background mode via double-`!` keystroke in shell command mode, execute commands that run without suspending the TUI, and view real-time output in the bottom third of the launching pane.

## Objectives

- Enable background shell command execution without TUI suspension
- Display real-time command output in a split pane area
- Allow full duofm file management during background execution
- Provide cancellation, auto-close, and shell log integration

## Prerequisites

### Development Environment
- Go 1.21+
- Make

### Dependencies
- Go standard library: `os/exec`, `context`, `bufio`, `io`, `sync`
- Existing: Bubble Tea framework, lipgloss styles
- Existing: ShellLogger, Minibuffer, ShellHistory, TabCompleter

## Architecture Overview

### Technology Stack
- **Language**: Go
- **Framework**: Bubble Tea (Elm Architecture)
- **Key Libraries**: lipgloss (styling), os/exec (process management)

### Design Approach

Background commands use a goroutine-based execution model. Unlike foreground commands which use TUI suspension, background commands pipe stdout/stderr through a scanner goroutine that feeds output lines back to the Bubble Tea event loop as messages. A context with cancellation supports process termination.

### Component Interaction

```
User Input → Model.Update() → BackgroundRunner.Start()
                                    │
                                    ├─ goroutine: scanner → bgOutputMsg → Model.Update()
                                    │                                        │
                                    │                                        └─ OutputBuffer.Append()
                                    │                                              │
                                    └─ process exit → bgCommandDoneMsg             └─ View() re-renders
                                                        │
                                                        └─ 2-sec timer → bgAutoCloseMsg → cleanup
```

**Key constraint**: Cannot use TUI suspension (existing foreground approach) because the TUI must remain interactive. Cannot use PTY allocation via `script(1)` while TUI owns the terminal.

---

## Implementation Phases

### Phase 1: Core Data Structures

**Goal**: Create the foundational OutputBuffer and BackgroundRunner components with full test coverage.

**Files to Create**:
- `internal/ui/output_buffer.go` - Circular line buffer for command output storage
- `internal/ui/output_buffer_test.go` - Unit tests for OutputBuffer
- `internal/ui/background_runner.go` - Background process lifecycle management
- `internal/ui/background_runner_test.go` - Unit tests for BackgroundRunner

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| OutputBuffer | Store output lines in a fixed-size circular buffer | maxLines > 0 | Lines() returns at most maxLines entries in insertion order |
| BackgroundRunner | Manage lifecycle of a single background process | Not already running | Process starts, output streams via callback, completion notified |

**OutputBuffer contract**:
- `NewOutputBuffer(maxLines)` → buffer
  - Precondition: maxLines > 0
  - Postcondition: Empty buffer with capacity maxLines
- `Append(line)` → void
  - Precondition: none
  - Postcondition: Line added; oldest line evicted if at capacity
- `Lines()` → ordered slice
  - Postcondition: Returns lines in insertion order, len <= maxLines
- `Clear()` → void
  - Postcondition: Buffer empty, Lines() returns empty slice

**BackgroundRunner contract**:
- `NewBackgroundRunner()` → runner
  - Postcondition: runner.IsRunning() == false
- `Start(command, workDir, onOutput, onDone)` → error
  - Precondition: IsRunning() == false
  - Postcondition: Process started, onOutput called per line, onDone called on exit
  - Error: Returns error if already running or process fails to start
- `Cancel()` → void
  - Precondition: IsRunning() == true
  - Postcondition: Process group terminated, onDone eventually called
- `IsRunning()` → bool
- `Pane()` → which pane launched the command
- `Command()` → the command string being executed

**Processing Flow**:
1. Start receives command string and working directory
2. Create cancellable context
3. Launch process via shell interpreter with combined stdout+stderr pipe
4. Scanner goroutine reads lines → calls onOutput callback per line
5. Process exits → goroutine calls onDone callback with exit error
6. Cancel triggers context cancellation → process group receives signal

**Implementation Steps**:
1. **Create OutputBuffer** - Circular buffer with append, lines, clear operations
2. **Create BackgroundRunner** - Process management with context cancellation
3. **Add process group handling** - Ensure child processes are terminated on cancel
4. **Write comprehensive unit tests** - Cover buffer overflow, cancellation, concurrent access

**Dependencies**: None (standalone components)

**Testing Approach**:
- Unit: Buffer capacity overflow, empty buffer, clear, unicode preservation
- Unit: Runner start/cancel/completion, concurrent start rejection, working directory
- Integration: Actual process execution with output capture

**Acceptance Criteria**:
- [ ] OutputBuffer correctly handles capacity overflow with circular eviction
- [ ] BackgroundRunner starts process and streams output via callbacks
- [ ] Cancel terminates the process group (parent + children)
- [ ] Concurrent Start calls are rejected when already running

**Estimated Effort**: Small

---

### Phase 2: Background Mode Input

**Goal**: Extend shell command mode to support background mode toggle via `!` keystroke, with pink prompt indicator.

**Files to Modify**:
- `internal/ui/model.go` - Add background-related state fields to Model struct
- `internal/ui/model_update_keyboard.go` - Extend shell command input handler for background mode
- `internal/ui/messages.go` - Add background message types

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Model (bg state) | Track background mode, runner, output buffer, focus state | Model initialized | bg* fields available for all handlers |
| Shell command handler (bg mode) | Toggle background mode on `!`, handle backspace exit | shellCommandMode == true | bgMode toggles, prompt updates |
| Message types | Define bgOutputMsg, bgCommandDoneMsg, bgAutoCloseMsg | N/A | Messages available for event loop |

**Model State Additions**:
- bgMode flag - true when in background input mode (pink prompt visible)
- bgRunner reference - manages the background process
- bgOutputBuffer reference - circular buffer for output lines
- bgOutputFocused flag - true when output area has keyboard focus
- bgClosing flag - true during 2-second post-completion delay

**Processing Flow** (shell command `!` key):
1. Key `!` received in shell command mode
   - If bgMode == false → set bgMode = true, change prompt to pink-colored indicator
   - If bgMode == true → append `!` character to input (normal character input)
2. Backspace on empty input in bgMode → set bgMode = false, restore normal prompt
3. Escape in bgMode → cancel background mode, exit shell command mode
4. All other shell features (history, completion, search) → work unchanged

**Implementation Steps**:
1. **Add bg state fields to Model** - bgMode, bgRunner, bgOutputBuffer, bgOutputFocused, bgClosing
2. **Define message types** - bgOutputMsg, bgCommandDoneMsg, bgAutoCloseMsg
3. **Extend shell command input handler** - `!` toggle, backspace exit, escape cancel
4. **Update prompt rendering** - Pink-colored prompt when bgMode is active, using existing highlightColor
5. **Initialize bg components** - OutputBuffer and BackgroundRunner creation in model constructor

**Dependencies**: Requires Phase 1 (OutputBuffer, BackgroundRunner)

**Testing Approach**:
- Unit: `!` sets bgMode=true, second `!` appends character
- Unit: Backspace on empty exits bgMode, backspace on non-empty deletes normally
- Unit: Escape cancels bgMode and shell command mode
- Unit: History/completion/search work in bgMode

**Acceptance Criteria**:
- [ ] `!` in shell command mode toggles bgMode and shows pink prompt
- [ ] Backspace on empty input in bgMode returns to normal shell command mode
- [ ] Escape in bgMode cancels and returns to normal mode
- [ ] All existing shell command features (history, Ctrl+R, TAB) work in bgMode

**Estimated Effort**: Small

---

### Phase 3: Background Execution and Output Streaming

**Goal**: Execute commands in background mode, stream output to buffer, and handle completion.

**Files to Modify**:
- `internal/ui/exec.go` - Add background command start function
- `internal/ui/model_update_keyboard.go` - Handle Enter in bgMode
- `internal/ui/model_update.go` - Handle bgOutputMsg, bgCommandDoneMsg
- `internal/ui/shell_logger.go` - Add line-level append for background output

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| startBackgroundCommand | Bridge between Model and BackgroundRunner, create Bubble Tea commands | bgMode == true, input non-empty | BackgroundRunner started, output flows as messages |
| Output message handler | Append output lines to buffer, trigger re-render | bgRunner.IsRunning() | OutputBuffer updated, view refreshes |
| Completion handler | Handle command exit, start auto-close timer | bgCommandDoneMsg received | bgClosing set, 2-sec timer started |
| ShellLogger (extended) | Record individual output lines to log file | Log file open | Each line appended to session log |

**Processing Flow** (Enter in bgMode):
1. Validate input is non-empty
2. Capture working directory from active pane's current path
3. Record active pane position (left/right) as launching pane
4. Exit shell command mode and bgMode
5. Log command header via ShellLogger
6. Start BackgroundRunner with command and working directory
7. Return Bubble Tea command that waits for first output message

**Processing Flow** (output streaming):
1. bgOutputMsg received → append line to OutputBuffer
2. Log line to ShellLogger
3. Return Bubble Tea command that waits for next output message

**Processing Flow** (command completion):
1. bgCommandDoneMsg received → set bgClosing = true
2. Log footer via ShellLogger
3. Start 2-second delay timer
4. Return timer command

**Processing Flow** (single command limit):
1. User presses `!` while bgRunner.IsRunning() → show warning in status bar
2. Warning message: "Background command running"
3. Do not enter shell command mode

**Implementation Steps**:
1. **Add ShellLogger line append** - Method to write individual output lines
2. **Create background start function** - Bridge BackgroundRunner with Bubble Tea message loop
3. **Handle Enter in bgMode** - Validate, capture context, start execution
4. **Handle output messages** - Buffer append, log, request next message
5. **Handle completion message** - Set closing state, start timer
6. **Block new shell commands during execution** - Warning in status bar

**Dependencies**: Requires Phase 1 (BackgroundRunner, OutputBuffer), Phase 2 (bgMode state)

**Testing Approach**:
- Unit: Enter with empty input does nothing
- Unit: Enter with input starts background process
- Unit: bgOutputMsg appends to buffer
- Unit: bgCommandDoneMsg sets bgClosing
- Unit: Shell command `!` while running shows warning
- Integration: Full command execution with output capture

**Acceptance Criteria**:
- [ ] Enter in bgMode starts background process without suspending TUI
- [ ] Output lines stream to OutputBuffer in real-time
- [ ] Command completion triggers 2-second close timer
- [ ] Attempting shell command while bg running shows status bar warning
- [ ] Output is recorded in ShellLogger

**Estimated Effort**: Medium

---

### Phase 4: Output Display Rendering

**Goal**: Render the output area in the bottom third of the launching pane, with auto-scroll and command header.

**Files to Modify**:
- `internal/ui/pane_render.go` - Add split-view rendering with output area
- `internal/ui/model_view.go` - Route to split-view rendering when bg active

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Split pane renderer | Render file list in top 2/3 and output in bottom 1/3 | bgRunner active on this pane | Pane displays both file list and output |
| Output area renderer | Render scrolled output lines with header | OutputBuffer has content | Output visible with auto-scroll to bottom |
| View router | Select normal or split rendering per pane | Model state available | Correct renderer called per pane |

**Processing Flow** (view rendering):
1. For each pane, check if background command is active on that pane
   - Active (running or closing) → render split view
   - Not active → render normal view
2. Split view layout:
   - Calculate total available height for file content area
   - Top 2/3: file list (with adjusted visible lines)
   - Separator line: command header with running indicator
   - Bottom 1/3: output lines (auto-scrolled to show latest)
3. When launching pane is not the active pane → output area hidden, render normal

**Output area header format**:
- Display the running command text
- Visual separator from file list above

**Auto-scroll behavior**:
- Always show the most recent lines (tail behavior)
- If output exceeds display area, show last N lines where N = output area height

**Implementation Steps**:
1. **Create output area renderer** - Render output lines with header, auto-scroll
2. **Create split pane renderer** - Divide pane height 2/3 + 1/3, compose renderers
3. **Update view routing** - Check bg state per pane, select appropriate renderer
4. **Handle pane visibility** - Hide output when launching pane is inactive
5. **Adjust file list height** - Reduce visible lines for file list in split mode

**Dependencies**: Requires Phase 3 (output streaming populates buffer)

**Testing Approach**:
- Unit: Output area renders in bottom 1/3 of pane height
- Unit: File list renders in top 2/3 with reduced visible lines
- Unit: Auto-scroll shows latest lines
- Unit: Header shows running command
- Unit: Output area hidden when launching pane not active

**Acceptance Criteria**:
- [ ] Output area visible in bottom 1/3 of launching pane
- [ ] File list remains interactive in top 2/3
- [ ] Output auto-scrolls to show latest lines
- [ ] Command header displayed at output area top
- [ ] Output hidden when switching away from launching pane

**Estimated Effort**: Medium

---

### Phase 5: Focus, Cancellation, and Auto-Close

**Goal**: Implement output area focus via TAB, Ctrl+C cancellation, and auto-close behavior.

**Files to Modify**:
- `internal/ui/model_update_keyboard.go` - Add focus handler, Ctrl+C, TAB routing
- `internal/ui/model_update.go` - Handle bgAutoCloseMsg
- `internal/ui/model.go` - Add cleanup helper methods

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Focus handler | Route TAB to focus output area | bg running on active pane | bgOutputFocused = true |
| Bg output focused handler | Handle keys when output area focused | bgOutputFocused == true | Ctrl+C cancels, TAB/Esc unfocuses, others ignored |
| Auto-close handler | Close output area after timer | bgAutoCloseMsg received | Output cleared, both panes reload |
| Cleanup helper | Reset all bg state and reload panes | bg operation ending | All bg fields reset, directory refresh triggered |

**Processing Flow** (TAB focus):
1. TAB key in normal mode
   - If bgRunner running AND active pane == launching pane → set bgOutputFocused = true
   - Otherwise → normal TAB behavior (existing)

**Processing Flow** (focused output area keys):
1. Ctrl+C → cancel background process via BackgroundRunner.Cancel()
   - Process group receives signal
   - Cleanup and reload both panes
2. TAB or Escape → set bgOutputFocused = false (return focus to file list)
3. All other keys → ignored (no-op)

**Processing Flow** (auto-close):
1. bgAutoCloseMsg received (2 seconds after completion)
2. Reset bgClosing flag
3. Clear OutputBuffer
4. Reset BackgroundRunner
5. Trigger both pane directory reload

**Processing Flow** (pane interaction during execution):
1. Pane switching (move_left/move_right) → works normally
2. File operations → work normally
3. Output area tracks launching pane, not active pane
4. Search, dialogs → work normally

**Implementation Steps**:
1. **Add TAB routing for bg focus** - Check bg state before normal TAB handling
2. **Create bg output focused handler** - Ctrl+C, TAB/Esc, ignore others
3. **Implement auto-close handler** - Clear state, reload panes on timer message
4. **Create bg cleanup helper** - Central method to reset all bg state
5. **Ensure pane operations work during execution** - Verify no interference

**Dependencies**: Requires Phase 3 (execution), Phase 4 (output display)

**Testing Approach**:
- Unit: TAB focuses output area when bg running on active pane
- Unit: TAB does nothing when no bg running
- Unit: Ctrl+C in focused mode cancels process
- Unit: TAB/Esc in focused mode returns to file list
- Unit: Other keys ignored in focused mode
- Unit: Auto-close clears state and reloads panes
- Unit: File operations work during bg execution
- Unit: Pane switching works during bg execution

**Acceptance Criteria**:
- [ ] TAB focuses output area when bg command running on active pane
- [ ] Ctrl+C in focused mode terminates background process group
- [ ] TAB/Esc in focused mode returns focus to file list
- [ ] All other keys ignored when output area focused
- [ ] Auto-close fires 2 seconds after completion, clears state, reloads both panes
- [ ] All existing operations (pane switch, file ops, search, dialogs) work during bg execution

**Estimated Effort**: Medium

---

### Phase 6: Polish and Edge Cases

**Goal**: Handle edge cases, update help dialog, ensure graceful shutdown.

**Files to Modify**:
- `internal/ui/model.go` - Graceful shutdown on quit
- `internal/ui/background_runner.go` - Robust process cleanup
- `internal/ui/help_dialog.go` - Update help text with background mode info
- `internal/ui/pane_render.go` - Handle edge cases in output display

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Graceful shutdown | Kill bg process on duofm exit | duofm quitting | No orphan processes |
| Output sanitizer | Handle binary/non-UTF8 output | Raw output received | Safe displayable string |
| Help dialog update | Document background mode keys | Help dialog rendered | Background mode info visible |

**Processing Flow** (graceful shutdown):
1. User quits duofm (Ctrl+Q or similar)
2. If bgRunner.IsRunning() → Cancel() to terminate process group
3. Wait briefly for cleanup
4. Proceed with normal quit

**Edge cases to handle**:
- Command with no output → output area shows header only, auto-closes normally
- Very large output (>10000 lines) → OutputBuffer circular eviction handles this
- Immediate exit command → bgCommandDoneMsg follows quickly, 2-sec timer fires
- Binary/non-UTF8 output → replace invalid bytes with replacement character
- Terminal resize during execution → output area re-renders with new dimensions
- Long single line → truncate or wrap based on pane width

**Implementation Steps**:
1. **Add graceful shutdown** - Cancel bg process on quit message
2. **Handle binary output** - Sanitize non-UTF8 bytes in output lines
3. **Handle terminal resize** - Re-calculate split dimensions on resize
4. **Update help dialog** - Add background mode keybinding documentation
5. **Review and fix edge cases** - No-output commands, immediate exit, long lines

**Dependencies**: Requires all previous phases

**Testing Approach**:
- Unit: Quit with running bg command terminates process
- Unit: Binary output is sanitized
- Unit: Terminal resize recalculates layout
- Unit: No-output command auto-closes correctly
- E2E: Background command execution end-to-end
- E2E: Background command cancellation

**Acceptance Criteria**:
- [ ] No orphan processes on duofm exit
- [ ] Binary output does not crash the display
- [ ] Terminal resize re-renders output area correctly
- [ ] Help dialog includes background mode documentation
- [ ] All existing E2E tests pass without regression

**Estimated Effort**: Small

---

## Complete File Structure

```
internal/ui/
├── background_runner.go          # NEW: BackgroundRunner process lifecycle
├── background_runner_test.go     # NEW: BackgroundRunner unit tests
├── output_buffer.go              # NEW: Circular line buffer
├── output_buffer_test.go         # NEW: OutputBuffer unit tests
├── model.go                      # MODIFY: Add bg* state fields, cleanup helpers
├── model_update.go               # MODIFY: Handle bg messages (output, done, auto-close)
├── model_update_keyboard.go      # MODIFY: bg mode toggle, focus handler, Ctrl+C
├── model_view.go                 # MODIFY: Route to split-view rendering
├── pane_render.go                # MODIFY: Split pane renderer with output area
├── exec.go                       # MODIFY: Add background command start function
├── messages.go                   # MODIFY: Add bg message types
├── shell_logger.go               # MODIFY: Add line-level append method
└── help_dialog.go                # MODIFY: Add background mode documentation
```

## Testing Strategy

- **Unit**: Core logic coverage 80%+, critical paths (cancellation, cleanup) 90%+
  - OutputBuffer: capacity, overflow, clear, unicode
  - BackgroundRunner: start, cancel, completion, concurrent rejection
  - Model state: bgMode toggle, focus transitions, auto-close
- **Integration**: End-to-end message flow through Bubble Tea event loop
  - Command execution → output streaming → completion → auto-close
- **E2E (Docker)**: Automated via existing `make test-e2e` infrastructure
  - Background command execution and output display
  - Cancellation via Ctrl+C
  - No regression in existing tests
- **Manual**: Items requiring human judgment
  - Visual appearance of pink prompt
  - Output area layout proportions
  - Subjective responsiveness during bg execution

## Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| os/exec | stdlib | Process execution |
| context | stdlib | Cancellation propagation |
| bufio | stdlib | Line-by-line output scanning |
| io | stdlib | Pipe handling |
| sync | stdlib | Concurrent access safety |
| syscall | stdlib | Process group signal delivery |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Orphan processes on crash | Medium | High | Process group kill, defer cleanup |
| Output flooding freezes TUI | Low | Medium | Circular buffer limits memory; Bubble Tea batches re-renders |
| Race conditions in runner state | Medium | Medium | Mutex protection on runner state |
| Goroutine leak on cancel | Low | Medium | Context cancellation closes pipes, unblocks scanner |

## Open Questions

- None (all requirements resolved in SPEC.md)

## Success Metrics

- [ ] Background commands execute without suspending TUI
- [ ] Output displays in real-time with <100ms latency
- [ ] Ctrl+C reliably terminates background processes including children
- [ ] No orphan processes under any exit scenario
- [ ] All existing tests pass without modification
- [ ] No noticeable TUI lag during background execution
