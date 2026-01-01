# Implementation Plan: Directory History Navigation

## Overview

Add browser-like directory history navigation to duofm. Each pane maintains an independent history stack (maximum 100 entries) allowing users to navigate backward and forward through previously visited directories using `Alt+←`/`Alt+→` or `[`/`]` keys.

## Objectives

- Implement a directory history stack with browser-like navigation (back/forward)
- Support independent history for left and right panes
- Preserve existing `-` key functionality (toggle to previous directory)
- Maintain session-only history with automatic cleanup
- Ensure navigation operations are efficient (O(1) time complexity)

## Prerequisites

### Development Environment
- Go 1.21 or later
- Bubble Tea framework (github.com/charmbracelet/bubbletea)
- duofm codebase (already set up)

### Dependencies
No external dependencies required beyond existing project dependencies.

### Knowledge Requirements
- Understanding of Bubble Tea's message-based architecture
- Familiarity with Go slices and index management
- Knowledge of duofm's Pane and navigation system

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (TUI framework)
- **Key Components**:
  - `DirectoryHistory` - History state management
  - `Pane` - Existing pane structure (to be extended)
  - `Action` - Existing action system (to be extended)

### Design Approach

**Bottom-up incremental implementation:**
1. Build history data structure first (isolated, testable)
2. Integrate into Pane structure
3. Add keybindings and actions
4. Wire up message handling

**Key design decisions:**
- History stored as simple slice (not circular buffer) for clarity
- Current position tracked by index for efficient back/forward
- Duplicate consecutive paths prevented automatically
- History navigation operations do NOT add to history (prevent infinite loops)

### Component Interaction

```
User Keypress → Keybinding Map → Action → Update Handler
                                            ↓
                                      DirectoryHistory
                                            ↓
                                      Path Change → Directory Load
```

## Implementation Phases

### Phase 1: History Data Structure

**Goal**: Create a self-contained, fully tested DirectoryHistory type that manages the history stack and current position according to specification.

**Files to Create**:
- `internal/ui/directory_history.go` - DirectoryHistory type and operations
- `internal/ui/directory_history_test.go` - Comprehensive unit tests

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| DirectoryHistory | Maintains history stack and current position | maxSize = 100 | Invariants preserved |
| AddToHistory | Add new path to history | Valid directory path | Forward history cleared, path appended |
| NavigateBack | Move position backward | None | Position decremented if possible |
| NavigateForward | Move position forward | None | Position incremented if possible |
| CanGoBack | Check backward availability | None | Returns true if currentIndex > 0 |
| CanGoForward | Check forward availability | None | Returns true if currentIndex < len-1 |

**Processing Flow**:
```
1. AddToHistory called with new path
   ├─ Duplicate consecutive path? → Ignore, return early
   ├─ Position in middle of history? → Truncate forward entries
   └─ History at max size? → Remove oldest entry (shift left)
2. Append new path to history
3. Set currentIndex to last position
```

**Implementation Steps**:

1. **Create DirectoryHistory struct**
   - Define fields: paths ([]string), currentIndex (int), maxSize (int)
   - Implement constructor: NewDirectoryHistory() with maxSize=100
   - Key considerations:
     - Initialize with empty slice and currentIndex=-1
     - Enforce invariant: -1 <= currentIndex < len(paths)

2. **Implement AddToHistory operation**
   - Describe behavior: truncate forward history, append path, maintain size limit
   - Key considerations:
     - Check for duplicate consecutive paths first
     - Truncate slice at currentIndex+1 before appending
     - If size exceeds maxSize, remove paths[0] and adjust index

3. **Implement NavigateBack operation**
   - Describe behavior: decrement currentIndex and return path at new position
   - Key considerations:
     - Return (path, true) if possible, ("", false) if already at beginning
     - Do not modify history stack

4. **Implement NavigateForward operation**
   - Describe behavior: increment currentIndex and return path at new position
   - Key considerations:
     - Return (path, true) if possible, ("", false) if already at end
     - Do not modify history stack

5. **Implement query operations**
   - CanGoBack: check if currentIndex > 0
   - CanGoForward: check if currentIndex < len(paths)-1

**Dependencies**:
- Requires: None (self-contained)
- Blocks: Phase 2 (Pane integration)

**Testing Approach**:

*Unit Tests*:
- Test AddToHistory with various scenarios (empty, middle position, max size)
- Test NavigateBack/Forward boundary conditions
- Test duplicate path prevention
- Test max size enforcement (101st entry removes 1st)
- Test state transitions (example: A→B→C, back twice, add D → [A,B,D])

*Integration Tests*:
- Not applicable (pure data structure)

*Manual Testing*:
- [ ] N/A for this phase

**Acceptance Criteria**:
- [ ] DirectoryHistory type compiles without errors
- [ ] All unit tests pass with 100% coverage
- [ ] AddToHistory correctly truncates forward history
- [ ] NavigateBack/Forward return correct path and ok value
- [ ] Max size limit (100) enforced correctly
- [ ] Duplicate consecutive paths prevented

**Estimated Effort**: 小 (1-2 days)

**Risks and Mitigation**:
- **Risk**: Edge cases in index management cause panics
  - **Mitigation**: Comprehensive unit tests covering all boundary conditions
  - **Mitigation**: Explicit invariant checks in code

---

### Phase 2: Pane Integration

**Goal**: Integrate DirectoryHistory into the Pane struct and ensure all directory navigation operations correctly add to history, while history navigation itself does NOT add to history.

**Files to Modify**:
- `internal/ui/pane.go`:
  - Add history field to Pane struct
  - Modify all directory navigation methods to call AddToHistory
  - Add new methods: NavigateHistoryBack, NavigateHistoryForward
  - Preserve existing previousPath field and behavior

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Pane.history | Stores directory history | Initialized on creation | Updated on all directory changes |
| recordPreviousPath | Save current path to previousPath | None | previousPath updated |
| addToHistory | Add current path to history | Valid directory | History updated |
| NavigateHistoryBack | Navigate backward in history | None | Path changed if history available |
| NavigateHistoryForward | Navigate forward in history | None | Path changed if history available |

**Processing Flow**:
```
Normal Directory Navigation (Enter, h, ~, -, =, etc.):
1. recordPreviousPath() - for existing - key behavior
2. addToHistory(currentPath) - BEFORE changing path
3. Change path
4. Load directory

History Navigation ([, ], Alt+←, Alt+→):
1. Call NavigateBack/Forward on history
2. Obtain destination path
3. Change path WITHOUT calling addToHistory
4. Load directory
```

**Implementation Steps**:

1. **Add history field to Pane struct**
   - Add `history DirectoryHistory` field in Pane struct definition
   - Initialize in NewPane constructor with NewDirectoryHistory()
   - Key considerations:
     - Preserve all existing fields including previousPath
     - No breaking changes to Pane creation

2. **Create helper method addToHistory**
   - Describe behavior: wraps history.AddToHistory with current path
   - Call from all directory navigation methods BEFORE path change
   - Key considerations:
     - Called before path changes, not after
     - Only called from non-history navigation methods

3. **Modify existing navigation methods**
   - Update EnterDirectory, MoveToParent, ChangeDirectory, NavigateToHome, NavigateToPrevious, SyncTo
   - Add addToHistory call after recordPreviousPath, before path change
   - Key considerations:
     - Do NOT modify async versions (handled in message processing)
     - Preserve all existing behavior
     - previousPath handling remains unchanged

4. **Add NavigateHistoryBack method**
   - Describe behavior: attempt backward navigation, load directory if successful
   - Return error if directory doesn't exist (do not change currentIndex)
   - Key considerations:
     - Do NOT call addToHistory
     - Do NOT call recordPreviousPath (history navigation is independent)
     - Handle deleted directories gracefully

5. **Add NavigateHistoryForward method**
   - Same behavior as NavigateHistoryBack but forward direction
   - Key considerations: same as NavigateHistoryBack

6. **Add async versions for Bubble Tea**
   - Create NavigateHistoryBackAsync and NavigateHistoryForwardAsync
   - Return tea.Cmd for directory loading
   - Key considerations:
     - Set loading state
     - Set pendingCursorTarget = "" (not parent navigation)

**Dependencies**:
- Requires: Phase 1 (DirectoryHistory implementation)
- Blocks: Phase 3 (Actions and keybindings)

**Testing Approach**:

*Unit Tests*:
- Test that normal navigation adds to history
- Test that history navigation does NOT add to history
- Test NavigateHistoryBack/Forward behavior
- Mock directory loading to focus on history logic

*Integration Tests*:
- Test sequence: A→B→C, back, back, verify at A
- Test sequence: A→B→C, back, navigate to D, verify forward history cleared
- Test left and right panes have independent histories

*Manual Testing*:
- [ ] Navigate through directories and verify history grows
- [ ] Use `[` to go back, verify previous directory shown
- [ ] Use `]` to go forward, verify next directory shown
- [ ] Verify `-` key still works (toggle behavior preserved)

**Acceptance Criteria**:
- [ ] All directory navigation methods add to history
- [ ] History navigation methods do NOT add to history
- [ ] NavigateHistoryBack/Forward work correctly
- [ ] previousPath field still functions for `-` key
- [ ] No regressions in existing navigation
- [ ] Unit tests pass with good coverage

**Estimated Effort**: 中 (3-4 days)

**Risks and Mitigation**:
- **Risk**: Missing a navigation method that should add to history
  - **Mitigation**: Code review of all path-changing methods in pane.go
  - **Mitigation**: Integration tests covering all navigation keys
- **Risk**: Breaking existing `-` key behavior
  - **Mitigation**: Preserve previousPath completely unchanged
  - **Mitigation**: Dedicated test for `-` key behavior

---

### Phase 3: Actions and Keybindings

**Goal**: Add new actions for history navigation and configure keybindings (Alt+←, Alt+→, [, ]) to trigger these actions.

**Files to Modify**:
- `internal/ui/actions.go`:
  - Add ActionHistoryBack and ActionHistoryForward constants
  - Add to actionNames and nameToAction maps
- `internal/config/defaults.go`:
  - Add default keybindings for history actions

**Files to Create**:
- None (modifications only)

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ActionHistoryBack | Represents backward navigation action | None | Action defined |
| ActionHistoryForward | Represents forward navigation action | None | Action defined |
| Keybinding config | Maps keys to actions | Valid key names | Actions triggered on keypress |

**Processing Flow**:
```
1. User presses [ or Alt+←
2. Bubble Tea captures key event
3. KeybindingMap looks up action (ActionHistoryBack)
4. Model.Update receives action
5. Calls activePane.NavigateHistoryBackAsync()
6. Returns Cmd for directory loading
```

**Implementation Steps**:

1. **Add action constants**
   - Add ActionHistoryBack and ActionHistoryForward to Action enum in actions.go
   - Place after ActionPrevDir for logical grouping
   - Key considerations:
     - Maintain iota sequence
     - Use consistent naming convention

2. **Add action name mappings**
   - Add "history_back" and "history_forward" to actionNames map
   - Add reverse mappings to nameToAction map
   - Key considerations:
     - Use snake_case for consistency
     - Ensure bidirectional mapping

3. **Configure default keybindings**
   - Add entries in defaults.go for history_back and history_forward
   - Map to ["Alt+Left", "["] and ["Alt+Right", "]"]
   - Key considerations:
     - Bubble Tea's key representation for Alt+arrow keys
     - Provide both Alt and bracket alternatives

**Dependencies**:
- Requires: Phase 2 (Pane methods implemented)
- Blocks: Phase 4 (Message handling)

**Testing Approach**:

*Unit Tests*:
- Test ActionFromName("history_back") returns ActionHistoryBack
- Test Action.String() for new actions
- Test default keybinding map contains history actions

*Integration Tests*:
- Not applicable (integration tested in Phase 4)

*Manual Testing*:
- [ ] N/A (no user-facing changes yet)

**Acceptance Criteria**:
- [ ] ActionHistoryBack and ActionHistoryForward defined
- [ ] Action name mappings bidirectional
- [ ] Default keybindings configured
- [ ] Unit tests pass
- [ ] No compilation errors

**Estimated Effort**: 小 (0.5-1 day)

**Risks and Mitigation**:
- **Risk**: Terminal doesn't recognize Alt+arrow keys
  - **Mitigation**: Already mitigated by providing [ and ] alternatives
  - **Mitigation**: Document in help screen

---

### Phase 4: Message Handling and UI Integration

**Goal**: Wire up the new actions in Model.Update to call history navigation methods and handle directory loading results, including error handling for deleted directories.

**Files to Modify**:
- `internal/ui/model.go`:
  - Add case for ActionHistoryBack in Update function
  - Add case for ActionHistoryForward in Update function
  - Handle directory load errors from history navigation
- `internal/ui/messages.go`:
  - Extend directoryLoadCompleteMsg if needed for history context

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Update handler | Dispatch actions to pane methods | Valid action | Cmd returned or state updated |
| Error handler | Display error for deleted directories | Load failed | Status message shown, position unchanged |
| Message handler | Apply loaded directory | Load succeeded | Pane updated with new entries |

**Processing Flow**:
```
History Navigation Flow:
1. ActionHistoryBack received in Update
2. Call activePane.NavigateHistoryBackAsync()
   ├─ history.NavigateBack() returns (path, true)
   └─ Returns LoadDirectoryAsync Cmd
3. directoryLoadCompleteMsg received
   ├─ Success? → Apply entries to pane, clear loading state
   └─ Failure? → Show error, do NOT change history position, restore path
```

**Implementation Steps**:

1. **Add ActionHistoryBack handler**
   - In Model.Update switch statement, add case for ActionHistoryBack
   - Call activePane.NavigateHistoryBackAsync()
   - Return resulting Cmd
   - Key considerations:
     - Only process when no dialog is open
     - Handle nil Cmd (no history available)

2. **Add ActionHistoryForward handler**
   - Same as ActionHistoryBack but for forward direction
   - Key considerations: same as above

3. **Extend directory load error handling**
   - In directoryLoadCompleteMsg handler, check if from history navigation
   - If error and from history: set status message, do NOT update pane path
   - Key considerations:
     - Distinguish history navigation from normal navigation
     - Keep history position unchanged on error (allow retry)
     - Display "Directory not found: /path" for 5 seconds

4. **Handle directory load success**
   - Existing logic should work without modification
   - Verify pendingCursorTarget is "" for history navigation (no cursor memory)
   - Key considerations:
     - Reuse existing directoryLoadCompleteMsg handling
     - No special cursor positioning for history navigation

**Dependencies**:
- Requires: Phase 3 (Actions defined)
- Blocks: None (final phase)

**Testing Approach**:

*Unit Tests*:
- Test Update function handles ActionHistoryBack/Forward
- Mock activePane.NavigateHistoryBackAsync to verify it's called
- Test error message setting on directory load failure

*Integration Tests*:
- Test full flow: key press → action → pane method → directory load → UI update
- Test error scenario: navigate to deleted directory → error shown

*Manual Testing*:
- [ ] Press `[` key and verify backward navigation
- [ ] Press `]` key and verify forward navigation
- [ ] Press `Alt+←` and verify backward navigation
- [ ] Press `Alt+→` and verify forward navigation
- [ ] Navigate to directory, delete it externally, press `[`, verify error message
- [ ] Verify left pane history independent from right pane
- [ ] Verify `-` key still works independently

**Acceptance Criteria**:
- [ ] ActionHistoryBack/Forward trigger pane methods
- [ ] Directory loading works correctly
- [ ] Error handling displays appropriate message
- [ ] History position unchanged on error
- [ ] All keybindings work as expected
- [ ] No regressions in existing navigation
- [ ] Integration tests pass

**Estimated Effort**: 中 (2-3 days)

**Risks and Mitigation**:
- **Risk**: Async directory loading complicates error handling
  - **Mitigation**: Reuse existing error handling patterns from async navigation
  - **Mitigation**: Clear distinction between normal and history navigation errors
- **Risk**: Bubble Tea key events vary by terminal
  - **Mitigation**: Provide [ and ] alternatives
  - **Mitigation**: Test on multiple terminals (documented in constraints)

---

## Complete File Structure

```
duofm/
├── cmd/duofm/
│   └── main.go                          # Entry point (no changes)
├── internal/
│   ├── ui/
│   │   ├── directory_history.go         # NEW: DirectoryHistory type
│   │   ├── directory_history_test.go    # NEW: Unit tests
│   │   ├── pane.go                      # MODIFIED: Add history field and methods
│   │   ├── pane_test.go                 # MODIFIED: Add history tests
│   │   ├── actions.go                   # MODIFIED: Add history actions
│   │   ├── actions_test.go              # MODIFIED: Test new actions
│   │   ├── model.go                     # MODIFIED: Handle history actions
│   │   ├── model_test.go                # MODIFIED: Test message handling
│   │   ├── keys.go                      # No changes (uses dynamic keybinding map)
│   │   ├── messages.go                  # Possibly extended for context
│   │   └── ... (other UI files unchanged)
│   ├── config/
│   │   ├── defaults.go                  # MODIFIED: Add default keybindings
│   │   └── ... (other config files unchanged)
│   └── fs/
│       └── ... (no changes)
├── doc/
│   └── tasks/
│       └── directory-history/
│           ├── SPEC.md                  # Specification (reference)
│           ├── 要件定義書.md             # Requirements (reference)
│           └── IMPLEMENTATION.md        # This file
└── tests/
    └── ... (E2E tests added separately)
```

**File Descriptions**:
- **directory_history.go**: Core history management logic (stack, position, operations)
- **directory_history_test.go**: Comprehensive unit tests for history data structure
- **pane.go**: Pane struct extended with history field and navigation methods
- **actions.go**: Action enum extended with history navigation actions
- **model.go**: Message handling for history actions, async directory loading
- **defaults.go**: Default keybindings configuration

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- Focus on state transitions and boundary conditions
- Mock directory loading where applicable

**Test Coverage Goals**:
- DirectoryHistory: 100% coverage (critical, pure logic)
- Pane history integration: 80%+ coverage
- Action mapping: 100% coverage (simple)
- Message handling: 70%+ coverage

**Key Test Areas**:

1. **DirectoryHistory (internal/ui/directory_history_test.go)**
   - AddToHistory: empty, middle position, max size, duplicates
   - NavigateBack/Forward: boundary conditions, empty history
   - State transitions: A→B→C, back, add D → [A,B,D]
   - Max size enforcement: 101st entry removes oldest

2. **Pane Integration (internal/ui/pane_test.go)**
   - Normal navigation adds to history
   - History navigation does NOT add to history
   - NavigateHistoryBack/Forward behavior
   - Independent histories for multiple panes
   - previousPath preservation

3. **Actions (internal/ui/actions_test.go)**
   - ActionFromName mapping correctness
   - Action.String() output

4. **Model Update (internal/ui/model_test.go)**
   - ActionHistoryBack/Forward dispatch
   - Directory load success handling
   - Directory load error handling
   - Status message setting

### Integration Testing

**Scenarios**:
1. Navigate A→B→C, press `[` twice, verify at A
2. Navigate A→B→C, press `[`, press `]`, verify at B
3. Navigate A→B→C, press `[` twice, navigate to D, verify history [A,B,D]
4. Verify left and right panes maintain independent histories
5. Navigate to directory, delete externally, press `[`, verify error message
6. Verify `-` key still works (independent from history)

**Approach**:
- Use existing test infrastructure in tests/ directory
- Create scenario-based E2E tests using programmatic key events
- Verify state after each navigation step

### Manual Testing Checklist

Based on spec test scenarios:

**Basic Behavior**:
- [ ] Navigate to new directory adds to history (Enter, h, ~, -, =, bookmark, symlink)
- [ ] `Alt+←` or `[` navigates backward in history
- [ ] `Alt+→` or `]` navigates forward in history
- [ ] Back/forward with no history does nothing (no error)
- [ ] Left and right panes have independent histories

**History State Management**:
- [ ] After A→B→C, back twice, navigate to D → history [A,B,D]
- [ ] After 100 entries, 101st removes oldest
- [ ] Consecutive navigation to same directory recorded only once

**Various Navigation Methods**:
- [ ] `Enter` to subdirectory recorded
- [ ] `h`/`←` to parent recorded
- [ ] `~` to home recorded
- [ ] `-` to previous recorded
- [ ] Bookmark navigation recorded
- [ ] `=` pane sync recorded
- [ ] Symlink following recorded

**History Navigation Itself**:
- [ ] After back with `Alt+←`, new navigation clears forward history
- [ ] History navigation (`[`, `]`, `Alt+←`, `Alt+→`) NOT recorded

**Integration with Existing Features**:
- [ ] `-` key works normally (toggle to previous)
- [ ] `-` navigation recorded in history
- [ ] After A→B→C, `-` goes to B, `[` goes to A

**Error Handling**:
- [ ] Navigate to deleted directory shows error "Directory not found: /path"
- [ ] On error, history position unchanged (can retry)
- [ ] Navigate to directory without permission shows permission error

**Terminal Compatibility**:
- [ ] Test on multiple terminals: xterm, iTerm2, GNOME Terminal, Alacritty
- [ ] Verify `Alt+←`/`Alt+→` work if terminal supports them
- [ ] Verify `[`/`]` work as fallback

## Dependencies

### External Dependencies

No new external dependencies required.

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1: DirectoryHistory (standalone)
2. Phase 2: Pane integration (depends on Phase 1)
3. Phase 3: Actions and keybindings (depends on Phase 2)
4. Phase 4: Message handling (depends on Phase 3)

**Component Dependencies**:
- `DirectoryHistory` has no dependencies (pure data structure)
- `Pane` depends on `DirectoryHistory`
- `Action` enum is independent (just adds constants)
- `Model.Update` depends on `Pane` methods and `Action` definitions

## Risk Assessment

### Technical Risks

1. **Index Management Complexity**
   - **Risk**: Off-by-one errors in currentIndex management cause panics
   - **Likelihood**: Medium (complex state transitions)
   - **Impact**: High (crashes application)
   - **Mitigation**:
     - Comprehensive unit tests for all state transitions
     - Explicit invariant checks in DirectoryHistory methods
     - Use defensive programming (range checks)

2. **Async Directory Loading Edge Cases**
   - **Risk**: Race conditions between history navigation and directory loading
   - **Likelihood**: Low (Bubble Tea is single-threaded)
   - **Impact**: Medium (incorrect state)
   - **Mitigation**:
     - Follow existing async patterns in codebase
     - Clear separation: history change vs directory load
     - Integration tests for async scenarios

3. **Terminal Compatibility**
   - **Risk**: Alt+arrow keys not recognized on some terminals
   - **Likelihood**: Medium (known limitation)
   - **Impact**: Low ([ and ] alternatives provided)
   - **Mitigation**:
     - Provide [ and ] as documented alternatives
     - Test on multiple terminals
     - Document terminal requirements

4. **Memory Usage with Large Histories**
   - **Risk**: 100 paths × 2 panes could use significant memory with very long paths
   - **Likelihood**: Low (max ~20KB as per spec)
   - **Impact**: Low (negligible)
   - **Mitigation**:
     - Already limited by spec to 100 entries
     - Monitor in practice, no action needed

### Implementation Risks

1. **Breaking Existing Navigation**
   - **Risk**: Adding history calls breaks previousPath behavior
   - **Likelihood**: Medium
   - **Impact**: High (regression in existing feature)
   - **Mitigation**:
     - Keep previousPath completely independent
     - Dedicated tests for `-` key behavior
     - Code review focusing on navigation methods

2. **Missing Navigation Methods**
   - **Risk**: Forgetting to add addToHistory to a navigation method
   - **Likelihood**: Medium
   - **Impact**: Medium (incomplete feature)
   - **Mitigation**:
     - Systematic code review of all path-changing methods
     - Integration tests covering all navigation keys
     - Checklist of all navigation operations from spec

3. **Scope Creep**
   - **Risk**: Adding features not in spec (e.g., history persistence, UI indicators)
   - **Likelihood**: Low
   - **Impact**: Low (wasted time)
   - **Mitigation**:
     - Strict adherence to SPEC.md requirements
     - Reject any features not explicitly requested

## Performance Considerations

### History Operations

**AddToHistory**:
- Slice truncation: O(1) when appending at end
- Slice append: O(1) amortized
- Shift removal at max size: O(n) where n=100 (negligible)
- Overall: Meets O(1) requirement in practice

**NavigateBack/Forward**:
- Index decrement/increment: O(1)
- Path retrieval: O(1)
- Overall: Meets O(1) requirement

### Memory Usage

- 100 paths per pane × 2 panes = 200 paths
- Average path length: ~100 bytes
- Total: ~20KB (as per spec NFR1.3)
- Negligible impact on modern systems

### Directory Loading

- History navigation reuses existing async loading mechanism
- No additional performance impact beyond normal navigation
- Loading indicators already implemented

## Security Considerations

### Path Validation

- Rely on existing directory loading error handling
- No special path validation needed for history
- Deleted/renamed directories handled gracefully (error message, no crash)

### Path Traversal

- History stores paths as-is (no normalization needed)
- Existing navigation already prevents path traversal attacks
- No new security risks introduced

## Open Questions

### From Specification:
None - all requirements are clarified in SPEC.md

### Implementation-Specific:
None - implementation approach is clear

### Resolved During Planning:
- **Q**: Should history be per-pane or global?
  - **A**: Per-pane (independent), as per spec DR1.1
- **Q**: Should previousPath be removed?
  - **A**: No, preserve for `-` key functionality (spec DR5.1)
- **Q**: How to distinguish history navigation from normal navigation in async loading?
  - **A**: Use context field in directoryLoadCompleteMsg if needed, or infer from pendingPath state

## Future Enhancements

Items explicitly deferred or out of scope:

### Not in Current Spec:
- **History persistence** (save/restore on app restart) - Spec states session-only
- **History UI indicator** (show position like "3/10") - Spec states no UI needed
- **Configurable max size** - Spec states fixed at 100
- **History dialog** (Midnight Commander style) - Not requested
- **History search** - Not requested
- **Global history across sessions** - Conflicts with session-only requirement

### Potential Phase 2 (if requested):
- Visual history indicator in status bar
- Configurable history size via config file
- History management commands (clear, view)

## Success Metrics

### Functional Completeness
- [ ] All navigation methods record history
- [ ] History navigation (4 keys) works correctly
- [ ] Max 100 entries enforced
- [ ] Duplicate prevention works
- [ ] Error handling for deleted directories
- [ ] previousPath and `-` key unaffected

### Quality Metrics
- [ ] Unit test coverage: DirectoryHistory 100%, Pane integration 80%+
- [ ] All integration tests pass
- [ ] No regressions in existing navigation
- [ ] Code follows Go conventions and duofm patterns

### Performance Metrics
- [ ] AddToHistory completes in O(1) time (measured)
- [ ] NavigateBack/Forward completes in O(1) time (measured)
- [ ] Memory usage ~20KB (100 paths × 2 panes)
- [ ] No perceivable performance degradation

### User Experience
- [ ] All keybindings work as documented
- [ ] Error messages are clear and helpful
- [ ] Navigation feels responsive
- [ ] Left and right panes work independently

## References

- **Specification**: `doc/tasks/directory-history/SPEC.md`
- **Requirements**: `doc/tasks/directory-history/要件定義書.md`
- **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
- **Go Testing**: https://go.dev/doc/tutorial/add-a-test
- **Existing Code**:
  - `internal/ui/pane.go` - Navigation and previousPath implementation
  - `internal/ui/actions.go` - Action definitions
  - `internal/config/defaults.go` - Default keybindings
- **Similar Implementations**:
  - Web browser history (Chrome, Firefox)
  - Midnight Commander Alt+O history dialog
  - ranger directory history

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Verify plan aligns with specification
   - Confirm phase breakdown is logical
   - Address any questions or concerns

2. **Environment Setup**
   - Ensure Go 1.21+ installed
   - Verify duofm builds successfully
   - Run existing tests to ensure clean baseline

3. **Begin Implementation**
   - Start with Phase 1 (DirectoryHistory)
   - Follow TDD approach:
     1. Write test cases from spec
     2. Implement to pass tests
     3. Refactor for clarity
   - Commit incrementally after each phase

4. **Testing and Validation**
   - Run unit tests after each phase
   - Run integration tests after Phase 4
   - Perform manual testing using checklist
   - Verify all acceptance criteria met

5. **Documentation and Cleanup**
   - Update help dialog with new keybindings
   - Document any deviations from plan (if any)
   - Clean up debug code
   - Final code review
