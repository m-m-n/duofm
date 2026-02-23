# Verification Document: Background Shell Command Execution

## Overview
**Feature**: Background Shell Command Execution
**SPEC.md**: `doc/tasks/background-shell-command/SPEC.md`
**IMPLEMENTATION.md**: `doc/tasks/background-shell-command/IMPLEMENTATION.md`

## Build Verification
- Command: `make build`
- Expected: exit code 0, no errors, no warnings

## Test Verification
- Command: `go test ./...`
- Coverage target: minimum 80%, target 90% for critical paths (BackgroundRunner, OutputBuffer)

### Test Scenarios from SPEC.md

| ID | Scenario | Expected Result | Test Type |
|----|----------|-----------------|-----------|
| TS-01 | `!` in shell command mode sets bgMode=true | Prompt changes to pink indicator | Unit |
| TS-02 | `!` in bgMode appends `!` character | Input field contains `!` | Unit |
| TS-03 | Backspace on empty input in bgMode | bgMode=false, normal prompt restored | Unit |
| TS-04 | Backspace on non-empty input in bgMode | Last character deleted, bgMode remains | Unit |
| TS-05 | Escape in bgMode | bgMode=false, shell command mode exits | Unit |
| TS-06 | History navigation (Up/Down) in bgMode | History entries cycle normally | Unit |
| TS-07 | Ctrl+R history search in bgMode | History search activates | Unit |
| TS-08 | TAB completion in bgMode | Completion candidates shown | Unit |
| TS-09 | Enter in bgMode starts bg process | bgRunner.IsRunning() == true | Unit |
| TS-10 | Enter in bgMode with empty input | No process started | Unit |
| TS-11 | Background command runs in correct directory | Process working dir matches active pane | Unit |
| TS-12 | stdout captured as bgOutputMsg | OutputBuffer receives stdout lines | Integration |
| TS-13 | stderr captured as bgOutputMsg | OutputBuffer receives stderr lines | Integration |
| TS-14 | bgCommandDoneMsg sent on completion | bgClosing flag set | Unit |
| TS-15 | Failed command sends error in done msg | Error available in bgCommandDoneMsg | Unit |
| TS-16 | Output recorded in ShellLogger | Log file contains command output | Integration |
| TS-17 | OutputBuffer append adds lines | Lines() returns appended lines | Unit |
| TS-18 | OutputBuffer respects maxLines | Oldest lines evicted at capacity | Unit |
| TS-19 | OutputBuffer Lines() returns in order | Insertion order preserved | Unit |
| TS-20 | OutputBuffer Clear() empties buffer | Lines() returns empty slice | Unit |
| TS-21 | Empty buffer returns empty slice | No panic, empty result | Unit |
| TS-22 | Unicode characters preserved in buffer | Multi-byte chars intact | Unit |
| TS-23 | Output area renders in bottom 1/3 | Pane split at correct height ratio | Unit |
| TS-24 | File list renders in top 2/3 | File list visible with reduced height | Unit |
| TS-25 | Auto-scroll shows latest lines | Most recent lines visible | Unit |
| TS-26 | Output area shows command header | Header contains running command | Unit |
| TS-27 | Output area hidden when bg not on pane | Normal rendering when pane inactive | Unit |
| TS-28 | 2-second timer starts after done msg | Timer command returned | Unit |
| TS-29 | bgAutoCloseMsg closes output area | bgClosing reset, buffer cleared | Unit |
| TS-30 | Both panes reload after close | Directory refresh triggered for both | Unit |
| TS-31 | TAB focuses output when bg on active pane | bgOutputFocused = true | Unit |
| TS-32 | TAB no-op when no bg running | Normal TAB behavior | Unit |
| TS-33 | Ctrl+C in focused mode cancels process | Process terminated, cleanup triggered | Unit |
| TS-34 | TAB in focused mode unfocuses | bgOutputFocused = false | Unit |
| TS-35 | Esc in focused mode unfocuses | bgOutputFocused = false | Unit |
| TS-36 | Other keys ignored in focused mode | No state change | Unit |
| TS-37 | Cancel terminates process group | Parent + child processes killed | Integration |
| TS-38 | Output area closes after cancellation | Cleanup triggered, panes reload | Unit |
| TS-39 | ShellLogger records partial output | Log contains lines up to cancellation | Integration |
| TS-40 | File operations during bg execution | Copy/move/delete work normally | Unit |
| TS-41 | Pane switching during bg execution | Active pane changes normally | Unit |
| TS-42 | Search during bg execution | Search mode works normally | Unit |
| TS-43 | New shell command while bg running | Warning shown in status bar | Unit |
| TS-44 | Command with no output | Output area shows header, auto-closes | Unit |
| TS-45 | Large output (>10000 lines) | Buffer evicts old lines, no crash | Unit |
| TS-46 | Immediate exit command | Auto-close timer fires normally | Unit |
| TS-47 | duofm quit while bg running | Process terminated, no orphans | Unit |
| TS-48 | Binary/non-UTF8 output | Displayed safely without crash | Unit |
| TS-49 | Terminal resize during execution | Output area re-renders at new size | Unit |
| TS-50 | Pane switching hides/shows output | Output visible only on launching pane | Unit |

## Code Quality Verification
- Format: `gofmt -w .`
- Static analysis: `go vet ./...`

## File Structure Verification

### Files to Create
- `internal/ui/output_buffer.go` - Circular line buffer
- `internal/ui/output_buffer_test.go` - OutputBuffer unit tests
- `internal/ui/background_runner.go` - Background process lifecycle
- `internal/ui/background_runner_test.go` - BackgroundRunner unit tests

### Files to Modify
- `internal/ui/model.go` - Add bg* state fields, cleanup helpers, init
- `internal/ui/model_update.go` - Handle bgOutputMsg, bgCommandDoneMsg, bgAutoCloseMsg
- `internal/ui/model_update_keyboard.go` - bg mode toggle, focus handler, Ctrl+C routing
- `internal/ui/model_view.go` - Route to split-view rendering
- `internal/ui/pane_render.go` - Split pane renderer with output area
- `internal/ui/exec.go` - Background command start function
- `internal/ui/messages.go` - Add bg message types
- `internal/ui/shell_logger.go` - Line-level append method
- `internal/ui/help_dialog.go` - Background mode documentation

## SPEC.md Compliance

### Success Criteria

| ID | Criterion | How to Verify |
|----|-----------|---------------|
| SC-01 | Background mode activates via double-`!` | Unit test: `!` in shell mode → bgMode, pink prompt |
| SC-02 | TUI remains interactive during bg execution | Unit test: pane switch, file ops, search all work |
| SC-03 | Output displays in real-time | Integration test: output lines appear <100ms |
| SC-04 | Ctrl+C cancels bg command | Unit test: focused Ctrl+C → process terminated |
| SC-05 | Auto-close after 2 seconds | Unit test: done msg → 2-sec timer → cleanup |
| SC-06 | No orphan processes | Unit test: quit during bg → process killed |
| SC-07 | Shell log records bg output | Integration test: log file contains output lines |

### Functional Requirements Coverage

| Requirement | Phase | Verification |
|-------------|-------|--------------|
| FR1: `!` switches to background mode | Phase 2 | TS-01: unit test for bgMode toggle |
| FR2: Pink prompt in background mode | Phase 2 | TS-01: unit test for prompt style |
| FR3: Backspace returns to normal mode | Phase 2 | TS-03: unit test for backspace on empty |
| FR4: Escape cancels background mode | Phase 2 | TS-05: unit test for escape handling |
| FR5: Existing shell features in bg mode | Phase 2 | TS-06, TS-07, TS-08: history, search, completion |
| FR6: Enter executes as background | Phase 3 | TS-09: unit test for bg start |
| FR7: Command via /bin/sh -c | Phase 1 | BackgroundRunner unit test |
| FR8: Working dir from active pane | Phase 3 | TS-11: unit test for working directory |
| FR9: TUI not suspended | Phase 3 | TS-40, TS-41, TS-42: operations during bg |
| FR10: Capture stdout and stderr | Phase 3 | TS-12, TS-13: integration tests |
| FR11: Single bg command limit | Phase 3 | TS-43: unit test for rejection |
| FR12: Warning when bg running | Phase 3 | TS-43: status bar warning |
| FR13: Output in bottom 1/3 | Phase 4 | TS-23: unit test for layout |
| FR14: Auto-scroll output | Phase 4 | TS-25: unit test for tail behavior |
| FR15: File list remains interactive | Phase 4 | TS-24: unit test for file list |
| FR16: Output area header | Phase 4 | TS-26: unit test for header |
| FR17: 2-second post-completion display | Phase 5 | TS-28: unit test for timer |
| FR18: Auto-close after delay | Phase 5 | TS-29: unit test for cleanup |
| FR19: Both panes reload on close | Phase 5 | TS-30: unit test for reload |
| FR20: TAB focuses output area | Phase 5 | TS-31: unit test for focus |
| FR21: Ctrl+C cancels bg when focused | Phase 5 | TS-33: unit test for cancel |
| FR22: Focus returns after cancel/TAB/Esc | Phase 5 | TS-34, TS-35: unit test for unfocus |
| FR23: Only Ctrl+C and TAB/Esc in focus | Phase 5 | TS-36: unit test for ignored keys |
| FR24: Pane switching during execution | Phase 5 | TS-41: unit test for pane switch |
| FR25: File operations during execution | Phase 5 | TS-40: unit test for file ops |
| FR26: Output tied to launching pane | Phase 4 | TS-50: unit test for pane tracking |
| FR27: Output hidden when pane inactive | Phase 4 | TS-27: unit test for visibility |
| FR28: Output in shell log | Phase 3 | TS-16: integration test for logging |
| FR29: Background visible in log viewer | Phase 3 | TS-16: verify log format |

## E2E Testing (Docker)
- [ ] Execute background command (`!!sleep 5 && echo done`) and verify output displays
- [ ] Cancel background command with TAB + Ctrl+C
- [ ] Verify pane operations (switching, navigation) during bg execution
- [ ] Verify auto-close after command completion
- [ ] All existing E2E tests pass without regression

## Manual Testing (E2E Not Possible)
- [ ] Visual appearance of pink prompt indicator
- [ ] Output area proportions look correct (1/3 of pane height)
- [ ] Subjective TUI responsiveness during background execution
- [ ] Output auto-scroll feels smooth with rapid output

## Performance Verification
- NFR1: Output display latency under 100ms (measure with high-frequency output command)
- NFR2: No noticeable TUI lag during bg execution (subjective assessment)

## Security Verification
- [ ] Commands execute via `/bin/sh -c` (same as foreground, no privilege escalation)
- [ ] Background process killed on duofm exit (no orphan processes)
- [ ] Process group kill ensures child processes terminated

## Verification Summary

| Category | Items | Automated | E2E (Docker) | Manual |
|----------|-------|-----------|--------------|--------|
| Background Mode Input | 8 | 8 | 0 | 0 |
| Background Execution | 8 | 6 | 2 | 0 |
| Output Buffer | 6 | 6 | 0 | 0 |
| Output Display | 5 | 5 | 0 | 2 |
| Auto-Close | 3 | 3 | 1 | 0 |
| Focus/Cancellation | 6 | 6 | 1 | 0 |
| Concurrent Operation | 4 | 4 | 1 | 1 |
| Edge Cases | 7 | 6 | 0 | 1 |
| Performance | 2 | 0 | 0 | 2 |
| Security | 3 | 2 | 0 | 1 |
| **Total** | **52** | **46** | **5** | **7** |
