# Implementation Plan: Shell Command History Enhancement

## Overview

This implementation enhances the existing shell command history functionality with bash-style Up/Down arrow key navigation and visual feedback of search patterns during Ctrl+R incremental search.

## Objectives

- Implement bash-style Up/Down arrow key history navigation in shell command mode
- Display search pattern inline during Ctrl+R incremental search (bash-style format)
- Maintain backward compatibility with existing functionality

## Prerequisites

### Development Environment

- Go 1.21 or later
- Bubble Tea framework already integrated
- Existing shell history infrastructure (`shell_history.go`, `history_searcher.go`)

### Dependencies

- No new external dependencies required
- Builds on existing `ShellHistory` and `HistorySearcher` components

### Knowledge Requirements

- Understanding of Bubble Tea message/update cycle
- Existing shell command mode architecture (`handleShellCommandInput`)
- Minibuffer rendering and state management

## Architecture Overview

### Technology Stack

- **Language**: Go 1.21+
- **Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Styling**: Lip Gloss (github.com/charmbracelet/lipgloss)

### Design Approach

The implementation extends the existing Model struct with two new fields for history navigation state, and modifies key handling to intercept Up/Down arrows in shell command mode. The search pattern display enhancement requires adding a getter method to HistorySearcher and updating the prompt format.

### Component Interaction

```
User Input (Up/Down/Ctrl+R)
    |
    v
handleShellCommandInput
    |
    +-- Up Key --> historyIndex++, recall command from ShellHistory
    |
    +-- Down Key --> historyIndex--, restore edit buffer or recall command
    |
    +-- Ctrl+R --> HistorySearcher with Pattern() for prompt display
    |
    v
Minibuffer.SetInput / SetPrompt
    |
    v
View renders updated prompt and command
```

## Implementation Phases

### Phase 1: Up/Down History Navigation

**Goal**: Enable bash-style arrow key navigation through command history in shell command mode

**Files to Modify**:

- `internal/ui/model.go`:
  - Add `historyIndex` field (navigation position: -1=at input, 0+=history positions)
  - Add `historyEditBuf` field (preserve original input before navigation)

- `internal/ui/model_update_keyboard.go`:
  - Extend `handleShellCommandInput` to handle Up/Down keys
  - Reset `historyIndex` when entering shell command mode
  - Reset `historyIndex` when typing characters (non-navigation)

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Model.historyIndex | Track current position in history | Shell command mode active | Valid index: -1 (edit buffer) to len(history)-1 |
| Model.historyEditBuf | Preserve user input before navigation | First Up key press | Original input stored, restorable |
| handleShellCommandInput | Route Up/Down keys to navigation logic | Key event received | Command displayed in minibuffer |

**Processing Flow**:

```
Up Key Pressed:
1. Check if in history search mode
   |-- Yes --> Do nothing (handled by search)
   |-- No --> Continue
2. Check if history is enabled and non-empty
   |-- No --> Do nothing
   |-- Yes --> Continue
3. If historyIndex == -1 (first navigation)
   |-- Save current input to historyEditBuf
4. If historyIndex < len(commands) - 1
   |-- Increment historyIndex
   |-- Display command at new index
```

```
Down Key Pressed:
1. Check if in history search mode
   |-- Yes --> Do nothing
   |-- No --> Continue
2. If historyIndex > -1
   |-- Decrement historyIndex
   |-- If historyIndex == -1 --> Restore historyEditBuf
   |-- Else --> Display command at new index
```

**Implementation Steps**:

1. **Add navigation state fields to Model**
   - Add integer field for history navigation position
   - Add string field for edit buffer preservation
   - Key consideration: Initialize historyIndex to -1 (before history)

2. **Handle Up key in shell command mode**
   - Save edit buffer on first navigation
   - Bounds check against history length
   - Display recalled command via minibuffer

3. **Handle Down key in shell command mode**
   - Navigate to newer commands or restore edit buffer
   - Bounds check: cannot go below -1

4. **Reset navigation state on mode entry**
   - Reset historyIndex to -1 when entering shell command mode
   - Note: Typing characters does NOT reset historyIndex (bash-style: edit in place, continue navigation)

**Dependencies**:

- Requires: Existing ShellHistory.Commands() method
- Blocks: None (independent of Phase 2)

**Testing Approach**:

*Unit Tests*:
- Up on empty history returns without change
- Up on first press shows most recent command
- Up at oldest command does not advance further
- Down after Up shows newer command
- Down at most recent restores original input
- Edit buffer is saved only on first Up press
- Entering shell mode resets historyIndex to -1

*Integration Tests*:
- Navigate up 3 commands, then down 2, correct command shown
- Type characters after navigation resets position

**Acceptance Criteria**:

- [ ] Up arrow shows previous command from history
- [ ] Down arrow shows newer command
- [ ] Down at newest command restores original input
- [ ] Empty history: Up key does nothing
- [ ] Navigation is bounded (no crash on edge cases)

**Estimated Effort**: Small (1-2 days)

**Risks and Mitigation**:

- **Risk**: Index out of bounds on rapidly changing history
  - **Mitigation**: Always bounds-check before accessing Commands() slice

---

### Phase 2: Search Pattern Display

**Goal**: Show the typed search pattern during Ctrl+R incremental search in bash-style format

**Files to Modify**:

- `internal/ui/history_searcher.go`:
  - Add `Pattern()` getter method

- `internal/ui/model_update_keyboard.go`:
  - Update prompt format during history search mode
  - Use bash-style format: `(reverse-i-search)'pattern': command`

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| HistorySearcher.Pattern() | Return current search pattern | HistorySearcher initialized | Pattern string returned |
| Prompt formatting | Construct bash-style prompt | In history search mode | Prompt shows pattern and match |

**Processing Flow**:

```
Ctrl+R Mode Active, Character Typed:
1. Update historySearchPattern
2. Call historySearcher.SetPattern()
3. Get current match
4. Format prompt: "(reverse-i-search)'{pattern}': {match}"
5. Display in minibuffer
```

```
Prompt Construction:
1. If historySearcher is nil
   |-- Return "(reverse-i-search)'': "
2. Get pattern from historySearcher.Pattern()
3. Format: "(reverse-i-search)'{pattern}': "
```

**Implementation Steps**:

1. **Add Pattern() getter to HistorySearcher**
   - Return the current search pattern string
   - Key consideration: Pattern field already exists, just needs getter

2. **Update prompt format during search**
   - Change from "(bck-i-search): " to "(reverse-i-search)'pattern': "
   - Key consideration: Pattern displayed even when empty

3. **Update prompt dynamically as pattern changes**
   - Refresh prompt after each character typed/deleted
   - Key consideration: Backspace removes characters from pattern

**Dependencies**:

- Requires: None
- Blocks: None (independent of Phase 1)

**Testing Approach**:

*Unit Tests*:
- Pattern() returns current search pattern
- Initial Ctrl+R shows empty pattern in prompt
- Typing updates pattern in prompt
- Backspace removes last character from pattern
- Pattern persists across Ctrl+R (next match)
- Esc clears pattern and returns to shell mode

*Integration Tests*:
- Full workflow: type pattern, see matches, continue typing

**Acceptance Criteria**:

- [ ] Prompt shows `(reverse-i-search)'': ` initially
- [ ] Prompt shows `(reverse-i-search)'g': ` after typing 'g'
- [ ] Backspace removes characters from pattern display
- [ ] Matched command displays after the pattern section
- [ ] No match: command section is empty

**Estimated Effort**: Small (1 day)

**Risks and Mitigation**:

- **Risk**: Long patterns may overflow display
  - **Mitigation**: Minibuffer already handles long input truncation

---

### Phase 3: Polish and Testing

**Goal**: Comprehensive testing, edge case handling, and documentation updates

**Files to Modify**:

- Test files:
  - `internal/ui/model_keyboard_test.go` (add tests)
  - `internal/ui/history_searcher_test.go` (add Pattern() test)

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Unit tests | Verify individual behaviors | Implementation complete | All tests pass |
| Integration tests | Verify end-to-end workflows | Unit tests pass | Workflows verified |
| Edge case tests | Verify boundary conditions | Core tests pass | Edge cases handled |

**Processing Flow**:

```
Test Execution:
1. Run unit tests for navigation
2. Run unit tests for pattern display
3. Run integration tests for workflows
4. Run edge case tests
5. Verify performance requirements
```

**Implementation Steps**:

1. **Add unit tests for history navigation**
   - Test all acceptance criteria from Phase 1
   - Key consideration: Use table-driven tests

2. **Add unit tests for pattern display**
   - Test Pattern() getter
   - Test prompt format variations

3. **Add integration tests**
   - Test combined navigation and search workflows
   - Test mode transitions

4. **Add edge case tests**
   - Unicode characters in search pattern
   - Very long patterns
   - Rapid key presses
   - Single entry history

**Dependencies**:

- Requires: Phase 1 and Phase 2 complete
- Blocks: None

**Testing Approach**:

See detailed test scenarios in SPEC.md

**Acceptance Criteria**:

- [ ] All unit tests pass with 80%+ coverage
- [ ] All integration tests pass
- [ ] Existing functionality remains intact (regression tests)
- [ ] Performance requirements met (50ms navigation, 100ms search update)

**Estimated Effort**: Small (1 day)

**Risks and Mitigation**:

- **Risk**: Existing tests may fail after changes
  - **Mitigation**: Run existing tests after each phase

---

## Complete File Structure

```
internal/ui/
|-- model.go                      # Add historyIndex, historyEditBuf fields
|-- model_update_keyboard.go      # Add Up/Down handling, update search prompt
|-- history_searcher.go           # Add Pattern() getter method
|-- history_searcher_test.go      # Add Pattern() tests
|-- model_keyboard_test.go        # Add navigation tests
|-- shell_history.go              # No changes expected
|-- shell_history_test.go         # No changes expected
|-- minibuffer.go                 # No changes expected
```

**File Descriptions**:

- `model.go`: Contains Model struct definition; adds navigation state fields
- `model_update_keyboard.go`: Contains key input handling; adds Up/Down navigation and prompt formatting
- `history_searcher.go`: Contains HistorySearcher for Ctrl+R search; adds Pattern() getter
- Test files: Contain unit and integration tests for new functionality

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- Temporary directories for history file tests

**Test Coverage Goals**:
- Navigation logic: 90%+ coverage
- Pattern display: 90%+ coverage
- Edge cases: 80%+ coverage

**Key Test Areas**:

1. **History Navigation** (`internal/ui/`)
   - Up key behavior (first press, repeated, at boundary)
   - Down key behavior (after navigation, at boundary)
   - Edit buffer preservation and restoration
   - Mode entry/exit state reset

2. **Pattern Display** (`internal/ui/`)
   - Pattern() getter returns correct value
   - Prompt format matches specification
   - Pattern updates on character input
   - Pattern clears on backspace/reset

### Integration Testing

**Scenarios**:
1. Navigate history, select command, execute
2. Search with pattern, navigate matches, execute
3. Mode transitions: shell -> search -> shell -> normal

### Manual Testing Checklist

Based on SPEC.md test scenarios:
- [ ] Up on empty history returns without change
- [ ] Up on first press shows most recent command
- [ ] Up on subsequent press shows older commands
- [ ] Down after Up shows newer command
- [ ] Down at most recent restores original input
- [ ] Initial Ctrl+R shows `(reverse-i-search)'': `
- [ ] Typing updates pattern display
- [ ] Backspace removes last char from pattern
- [ ] Unicode characters display correctly

## Dependencies

### External Dependencies

No new external dependencies required.

### Internal Dependencies

**Implementation Order**:
1. Phase 1 (no dependencies)
2. Phase 2 (no dependencies, can be done in parallel)
3. Phase 3 (depends on Phases 1 and 2)

**Component Dependencies**:
- `historyIndex`/`historyEditBuf` depend on `ShellHistory.Commands()`
- `Pattern()` getter depends on `HistorySearcher.pattern` field
- Prompt formatting depends on `Pattern()` getter

## Risk Assessment

### Technical Risks

1. **Index Out of Bounds**
   - **Risk**: History may change during navigation
   - **Likelihood**: Low
   - **Impact**: High (crash)
   - **Mitigation**: Always bounds-check before accessing history

2. **State Inconsistency**
   - **Risk**: Navigation state not properly reset on mode transitions
   - **Likelihood**: Medium
   - **Impact**: Medium (confusing UX)
   - **Mitigation**: Reset state explicitly on mode entry/exit

### Implementation Risks

1. **Regression in Existing Functionality**
   - **Risk**: Changes may break existing Ctrl+R or shell command behavior
   - **Mitigation**: Run existing tests after each change

## Performance Considerations

1. **Navigation Response Time**
   - Target: < 50ms per Up/Down press
   - History access is O(1) via slice index
   - No performance concerns expected

2. **Search Pattern Update**
   - Target: < 100ms per character
   - Existing search is already fast
   - No performance concerns expected

## Security Considerations

None specific to this feature. Command history is already stored securely with appropriate permissions.

## Open Questions

### From Specification

None - specification is complete.

### Implementation-Specific

None - implementation approach is clear from existing codebase.

## Future Enhancements

Items not in current specification:

- Forward search (Ctrl+S) - common in bash
- Edit and re-execute (Ctrl+X Ctrl+E style)
- History expansion (!!, !$, etc.)

## Success Metrics

### Functional Completeness

- [ ] All acceptance criteria from user stories met
- [ ] Up/Down navigation works as specified
- [ ] Search pattern display works as specified
- [ ] Existing functionality intact

### Quality Metrics

- [ ] Test coverage 80%+
- [ ] No critical bugs in manual testing
- [ ] Code follows Go best practices

### Performance Metrics

- [ ] Navigation response < 50ms
- [ ] Search update response < 100ms

### User Experience

- [ ] Intuitive bash-like navigation
- [ ] Clear visual feedback during search
- [ ] No confusion between modes

## References

- **Specification**: `doc/tasks/shell-command-history-enhancement/SPEC.md`
- **Base Feature**: `doc/tasks/shell-command-history/SPEC.md`
- **Existing Implementation**: `internal/ui/model_update_keyboard.go`, `internal/ui/shell_history.go`
- **bash readline**: https://www.gnu.org/software/bash/manual/html_node/Commands-For-History.html

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Confirm approach with stakeholder
   - Verify understanding of existing codebase

2. **Begin Implementation**
   - Start with Phase 1 or Phase 2 (can be parallel)
   - Follow TDD approach
   - Commit incrementally

3. **Testing**
   - Run tests after each phase
   - Manual testing per checklist
   - Performance verification
