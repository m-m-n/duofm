# Implementation Plan: Path Jump Dialog

## Overview

Enable direct navigation to any directory via full path input with bash-style inline autocompletion. This feature provides efficient keyboard-driven navigation without requiring bookmarks.

## Objectives

- Implement a dialog that accepts absolute path input with real-time filesystem suggestions
- Provide Tab completion following bash shell mental model
- Integrate seamlessly with existing dialog system and keybinding infrastructure

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make (for build automation)

### Dependencies
- Existing `BaseDialog` and `DialogStyles` infrastructure
- Existing `TextInput` component for text editing
- `os` and `filepath` packages for filesystem operations

### Knowledge Requirements
- Bubble Tea message handling and Update pattern
- Existing dialog implementation patterns (InputDialog, BookmarkDialog)
- Keybinding system (Action, DefaultKeybindings, KeybindingMap)

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Styling**: Lip Gloss (github.com/charmbracelet/lipgloss)

### Design Approach
- Extend existing dialog pattern by embedding BaseDialog
- Separate suggestion logic into dedicated PathSuggester component
- Follow message-based communication (pathJumpResultMsg, pathJumpCancelMsg)
- Integrate via keybinding system using ActionPathJump

### Component Interaction

```
User Input (Ctrl+J)
    |
    v
KeybindingMap.GetAction -> ActionPathJump
    |
    v
Model.handleAction -> creates PathJumpDialog
    |
    v
PathJumpDialog (handles input, Tab, Enter, Esc)
    |
    +--> PathSuggester (filesystem lookup)
    |
    v
pathJumpResultMsg or pathJumpCancelMsg
    |
    v
Model.handlePathJumpMessages -> Pane.ChangeDirectoryAsync
```

## Implementation Phases

### Phase 1: Core Infrastructure

**Goal**: Add action constant, keybinding, and message types for path jump functionality.

**Files to Create**:
- None

**Files to Modify**:
- `internal/ui/actions.go`:
  - Add ActionPathJump constant to Action enum
  - Add mapping entries to actionNames and nameToAction maps
- `internal/config/defaults.go`:
  - Add "path_jump" action with "Ctrl+J" keybinding

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ActionPathJump | Represent path jump action in keybinding system | None | Action constant available for mapping |

**Processing Flow**:
```
1. Add ActionPathJump constant after existing actions
2. Register in both mapping tables (action->name, name->action)
3. Add default keybinding "Ctrl+J" for "path_jump" action
```

**Implementation Steps**:

1. **Add ActionPathJump constant**
   - Place in appropriate section (Navigation extended or new category)
   - Assign next available iota value

2. **Update action maps**
   - Add to actionNames: ActionPathJump -> "path_jump"
   - Add to nameToAction: "path_jump" -> ActionPathJump

3. **Add default keybinding**
   - Add entry in DefaultKeybindings function
   - Also add to AllActions list

**Dependencies**:
- Requires: None
- Blocks: Phase 2, Phase 3

**Testing Approach**:

*Unit Tests*:
- Test ActionFromName("path_jump") returns ActionPathJump
- Test ActionPathJump.String() returns "path_jump"

*Integration Tests*:
- Test KeybindingMap correctly maps Ctrl+J to ActionPathJump

**Acceptance Criteria**:
- [ ] ActionPathJump constant exists
- [ ] ActionFromName("path_jump") returns correct action
- [ ] DefaultKeybindings includes "path_jump" with "Ctrl+J"

**Estimated Effort**: Small (1-2 hours)

---

### Phase 2: Path Suggester Component

**Goal**: Implement filesystem lookup and suggestion algorithm for path autocompletion.

**Files to Create**:
- `internal/ui/path_suggester.go` - Suggestion logic
- `internal/ui/path_suggester_test.go` - Unit tests

**Files to Modify**:
- None

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| PathSuggester | Compute completion suffix from partial path | Input is non-empty string | Returns suffix string (may be empty) |
| Suggest method | Query filesystem for matching directories | Parent directory exists | Returns first alphabetical match suffix |

**Processing Flow**:
```
1. Receive partial path input (e.g., "/home/us")
   |
   +-- Empty or no "/" -> return empty suffix
   |
2. Split into parent directory and prefix
   |
   Parent: "/home", Prefix: "us"
   |
3. Read parent directory contents
   |
   +-- Error (not exist, permission) -> return empty suffix
   |
4. Filter entries: directories only, name starts with prefix
   |
5. Sort matches alphabetically
   |
6. Return suffix of first match
   |
   Match: "user" -> Suffix: "er"
```

**Implementation Steps**:

1. **Define PathSuggester struct**
   - Stateless component (no caching in MVP)
   - Constructor function

2. **Implement Suggest method**
   - Input: partial path string
   - Output: suffix string for completion
   - Handle edge cases: empty input, root path, trailing slash

3. **Implement directory filtering**
   - Use os.ReadDir for directory listing
   - Filter by IsDir() and prefix match
   - Case-sensitive matching

4. **Handle edge cases**
   - Root "/" returns first top-level directory
   - Path ending with "/" suggests children of that directory
   - Non-existent parent returns empty suffix

**Dependencies**:
- Requires: None (standalone component)
- Blocks: Phase 3

**Testing Approach**:

*Unit Tests*:
- Test basic completion: "/home/us" -> "er" (if "user" exists)
- Test no match: "/nonexistent/path" -> ""
- Test directories only: skip files in suggestions
- Test hidden directories: include in results
- Test root path: "/" suggests first top-level dir
- Test case sensitivity: "Us" does not match "user"
- Test trailing slash: "/home/" suggests first child
- Test permission denied: returns empty without crash

*Manual Testing*:
- [ ] Verify suggestions match filesystem state

**Acceptance Criteria**:
- [ ] PathSuggester returns correct suffix for valid partial paths
- [ ] Returns empty string when no match found
- [ ] Only suggests directories, never files
- [ ] Handles filesystem errors gracefully (no panic)
- [ ] Performance: completes within 100ms for typical directories

**Estimated Effort**: Small (2-4 hours)

---

### Phase 3: Path Jump Dialog

**Goal**: Implement the dialog UI with input handling, suggestion display, and validation.

**Files to Create**:
- `internal/ui/path_jump_dialog.go` - Dialog implementation
- `internal/ui/path_jump_dialog_test.go` - Unit tests

**Files to Modify**:
- None

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| PathJumpDialog | Manage dialog state and render UI | Dialog is active | Handles all keyboard input |
| pathJumpResultMsg | Communicate successful path selection | Valid path entered | Contains validated path |
| pathJumpCancelMsg | Communicate dialog cancellation | Esc pressed | No payload |

**Processing Flow**:
```
1. Dialog created and activated
   |
2. User input loop:
   |
   +-- Character input -> update TextInput, recalculate suggestion
   |
   +-- Tab key:
   |     |
   |     +-- Suggestion exists -> append suffix to input, recalculate
   |     +-- No suggestion -> no action
   |
   +-- Enter key:
   |     |
   |     +-- Validate path (exists, is directory)
   |     |     |
   |     |     +-- Valid -> Close dialog, return pathJumpResultMsg
   |     |     +-- Invalid -> Set errorMsg, remain open
   |
   +-- Esc key -> Close dialog, return pathJumpCancelMsg
   |
   +-- Backspace/Delete -> update TextInput, clear error, recalculate
```

**Implementation Steps**:

1. **Define PathJumpDialog struct**
   - Embed BaseDialog
   - TextInput for path editing
   - PathSuggester for completion
   - Fields: suggestion (current suffix), errorMsg
   - DialogStyles for rendering

2. **Implement NewPathJumpDialog**
   - Initialize with empty TextInput
   - Create PathSuggester instance
   - Set default dialog width

3. **Implement Update method**
   - Guard: return early if not active
   - Clear error on any keystroke
   - Handle Tab: confirm suggestion if present
   - Handle Enter: validate and submit or show error
   - Handle Esc: cancel dialog
   - Delegate other keys to TextInput
   - Recalculate suggestion after input changes

4. **Implement View method**
   - Render title "Jump to Directory"
   - Render input field with suggestion suffix (grayed)
   - Render error message if present
   - Render footer with keybinding hints

5. **Implement path validation**
   - Check for empty input
   - Check path exists using os.Stat
   - Check path is directory not file
   - Return appropriate error message

**Dependencies**:
- Requires: Phase 2 (PathSuggester)
- Blocks: Phase 4

**Testing Approach**:

*Unit Tests*:
- Test dialog initialization
- Test Tab completion updates input
- Test Tab with no suggestion does nothing
- Test Enter with valid path sends pathJumpResultMsg
- Test Enter with non-existent path shows error
- Test Enter with file path shows error
- Test Enter with empty input shows error
- Test Esc sends pathJumpCancelMsg
- Test error clears on subsequent keystroke
- Test inactive dialog ignores all input

*Manual Testing*:
- [ ] Dialog opens with cursor in input field
- [ ] Suggestion appears grayed after typed portion
- [ ] Tab completes visible suggestion
- [ ] Error message appears and clears correctly

**Acceptance Criteria**:
- [ ] Dialog renders with title, input, and footer
- [ ] Suggestion suffix displays in muted color
- [ ] Tab confirms suggestion and recalculates
- [ ] Enter validates and navigates or shows error
- [ ] Esc closes without side effects
- [ ] All keyboard shortcuts work (Ctrl+A, Ctrl+E, etc.)

**Estimated Effort**: Medium (4-6 hours)

---

### Phase 4: Model Integration

**Goal**: Wire up the dialog to Model, handle messages, and trigger navigation.

**Files to Create**:
- None

**Files to Modify**:
- `internal/ui/model_update_keyboard.go`:
  - Add ActionPathJump case in handleAction function
- `internal/ui/model_update.go`:
  - Add handlePathJumpMessages function
  - Call from main Update message switch

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| handleAction(ActionPathJump) | Create and display PathJumpDialog | No other dialog active | Dialog assigned to m.dialog |
| handlePathJumpMessages | Process result/cancel messages | Message received | Directory changed or dialog cleared |

**Processing Flow**:
```
1. Ctrl+J pressed
   |
   v
2. KeybindingMap returns ActionPathJump
   |
   v
3. handleAction creates PathJumpDialog
   |
   m.dialog = NewPathJumpDialog()
   |
4. Dialog handles user input (Phase 3 flow)
   |
5. On Enter (valid path):
   |
   pathJumpResultMsg{path: "/target/path"}
   |
   v
6. handlePathJumpMessages:
   |
   m.dialog = nil
   m.getActivePane().ChangeDirectoryAsync(result.path)
   |
7. On Esc:
   |
   pathJumpCancelMsg{}
   |
   v
8. handlePathJumpMessages:
   |
   m.dialog = nil
```

**Implementation Steps**:

1. **Add ActionPathJump handler**
   - In handleAction switch statement
   - Create PathJumpDialog and assign to m.dialog
   - Return m, nil (no command needed)

2. **Add message handler function**
   - handlePathJumpMessages receives tea.Msg
   - Type switch on pathJumpResultMsg and pathJumpCancelMsg
   - For result: clear dialog, call ChangeDirectoryAsync
   - For cancel: clear dialog only
   - Return handled flag

3. **Wire into Update**
   - Call handlePathJumpMessages from main message handler
   - Follow existing pattern (bookmarkJumpMsg handling)

**Dependencies**:
- Requires: Phase 3 (PathJumpDialog)
- Blocks: None

**Testing Approach**:

*Unit Tests*:
- Test handleAction(ActionPathJump) creates dialog
- Test pathJumpResultMsg triggers directory change
- Test pathJumpCancelMsg clears dialog without navigation

*Integration Tests*:
- Test full flow: Ctrl+J -> type path -> Enter -> directory changes
- Test cancel flow: Ctrl+J -> Esc -> no change

**Acceptance Criteria**:
- [ ] Ctrl+J opens PathJumpDialog when no other dialog active
- [ ] Successful path entry navigates active pane
- [ ] Cancel returns to normal state without navigation
- [ ] No regression in existing functionality

**Estimated Effort**: Small (2-3 hours)

---

## Complete File Structure

```
internal/ui/
├── actions.go                    # Add ActionPathJump constant
├── model_update_keyboard.go      # Add ActionPathJump handling
├── model_update.go               # Add pathJumpResultMsg/Cancel handling
├── path_jump_dialog.go           # NEW: Dialog implementation
├── path_jump_dialog_test.go      # NEW: Dialog unit tests
├── path_suggester.go             # NEW: Filesystem suggestion logic
├── path_suggester_test.go        # NEW: Suggestion unit tests
├── text_input.go                 # Reuse existing component
├── dialog_base.go                # Reuse existing BaseDialog
└── dialog.go                     # Reuse existing interface

internal/config/
└── defaults.go                   # Add path_jump keybinding
```

**File Descriptions**:
- `path_jump_dialog.go`: Dialog component managing input state, suggestion display, validation, and message emission
- `path_suggester.go`: Pure logic component for filesystem-based path completion
- `actions.go`: Extended with ActionPathJump constant
- `defaults.go`: Extended with "path_jump" -> ["Ctrl+J"] mapping
- `model_update_keyboard.go`: Extended handleAction to create dialog
- `model_update.go`: Extended with pathJumpResultMsg/Cancel handlers

## Testing Strategy

### Unit Testing

**Approach**:
- Table-driven tests for PathSuggester with various filesystem scenarios
- Mock filesystem using temp directories for predictable behavior
- Test PathJumpDialog using tea.Msg injection

**Test Coverage Goals**:
- PathSuggester: 90%+ coverage (core logic)
- PathJumpDialog: 80%+ coverage (keyboard handling)
- Integration handlers: 70%+ coverage

**Key Test Areas**:
1. **PathSuggester** (`internal/ui/path_suggester_test.go`)
   - Valid prefix completion
   - No match scenarios
   - Directory-only filtering
   - Edge cases: root, trailing slash, permissions

2. **PathJumpDialog** (`internal/ui/path_jump_dialog_test.go`)
   - Tab completion behavior
   - Enter validation (valid/invalid/empty)
   - Esc cancellation
   - Error message lifecycle

3. **Model Integration**
   - Action handling creates dialog
   - Message handling triggers navigation

### Integration Testing

**Scenarios**:
1. Full keyboard flow: Ctrl+J -> input -> Tab -> Enter
2. Error recovery: invalid path -> error shown -> correction -> success
3. Cancellation: Ctrl+J -> partial input -> Esc -> no change

### Manual Testing Checklist

From SPEC.md test scenarios:
- [ ] `TestPathJumpDialog_NewDialog` - Dialog initializes correctly
- [ ] `TestPathJumpDialog_TabCompletion` - Tab confirms suggestion
- [ ] `TestPathJumpDialog_TabNoSuggestion` - Tab does nothing without suggestion
- [ ] `TestPathJumpDialog_EnterValidPath` - Enter with valid path sends result message
- [ ] `TestPathJumpDialog_EnterInvalidPath` - Enter with invalid path shows error
- [ ] `TestPathJumpDialog_EnterEmptyPath` - Enter with empty path shows error
- [ ] `TestPathJumpDialog_EscCancel` - Esc sends cancel message
- [ ] `TestPathJumpDialog_ErrorClearsOnInput` - Error message clears on keystroke
- [ ] `TestPathJumpDialog_InactiveIgnoresInput` - Inactive dialog ignores all input

PathSuggester tests:
- [ ] `TestPathSuggester_BasicCompletion` - Returns correct suffix
- [ ] `TestPathSuggester_NoMatch` - Returns empty when no match
- [ ] `TestPathSuggester_DirectoriesOnly` - Ignores files
- [ ] `TestPathSuggester_HiddenDirs` - Includes hidden directories
- [ ] `TestPathSuggester_RootPath` - Handles "/" correctly
- [ ] `TestPathSuggester_NonExistentParent` - Handles missing parent gracefully
- [ ] `TestPathSuggester_CaseSensitive` - Case-sensitive matching

## Dependencies

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/charmbracelet/bubbletea | (existing) | TUI framework |
| github.com/charmbracelet/lipgloss | (existing) | Styling |
| os | stdlib | Filesystem operations |
| filepath | stdlib | Path manipulation |

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1: Infrastructure (no dependencies)
2. Phase 2: PathSuggester (no dependencies)
3. Phase 3: PathJumpDialog (depends on Phase 2)
4. Phase 4: Model Integration (depends on Phase 3)

**Component Dependencies**:
- `PathJumpDialog` depends on `PathSuggester`, `TextInput`, `BaseDialog`
- `model_update_keyboard.go` depends on `PathJumpDialog`
- `model_update.go` depends on message types defined in `path_jump_dialog.go`

## Risk Assessment

### Technical Risks

1. **Filesystem Performance on Large Directories**
   - **Risk**: Suggestion lookup slow for /usr/lib or similar
   - **Likelihood**: Low (single directory read)
   - **Impact**: Medium (laggy autocomplete)
   - **Mitigation**: NFR1 requires < 100ms; monitor and optimize if needed

2. **Path Edge Cases**
   - **Risk**: Symlinks, special paths, mounted filesystems
   - **Likelihood**: Medium
   - **Impact**: Low (graceful degradation to no suggestion)
   - **Mitigation**: Use os.Stat which handles these; document behavior

### Implementation Risks

1. **Suggestion Display Width**
   - **Risk**: Long paths may not fit in dialog
   - **Likelihood**: Medium
   - **Impact**: Low (visual only)
   - **Mitigation**: TextInput handles scrolling; adjust dialog width if needed

## Performance Considerations

1. **Suggestion Lookup**
   - Single os.ReadDir call per keystroke
   - No caching in MVP (simple, predictable)
   - Can add caching if performance issues arise

2. **Rendering**
   - Minimal re-rendering via Bubble Tea
   - No complex calculations in View

## Security Considerations

1. **Path Traversal**
   - No sanitization beyond what OS provides
   - User can navigate anywhere they have permission
   - Consistent with file manager's purpose

2. **Permission Handling**
   - os.Stat and os.ReadDir respect OS permissions
   - Errors displayed as user-friendly messages

## Open Questions

None - all requirements resolved in specification.

## Future Enhancements

Items not in current scope:
- Path history (remember recently jumped paths)
- Multiple suggestions (cycle through matches)
- Fuzzy matching
- Custom path validation rules

## Success Metrics

### Functional Completeness
- [ ] Ctrl+J opens dialog
- [ ] Tab completion works
- [ ] Enter navigates or shows error
- [ ] Esc cancels
- [ ] All unit tests pass

### Quality Metrics
- [ ] Test coverage > 80% for new code
- [ ] No crashes on invalid input
- [ ] gofmt clean
- [ ] go vet clean

### Performance Metrics
- [ ] Suggestion lookup < 100ms
- [ ] Dialog render < 50ms

## References

- **Specification**: `doc/tasks/path-jump-dialog/SPEC.md`
- **Reference Implementation**: `internal/ui/input_dialog.go`
- **Reference Implementation**: `internal/ui/bookmark_dialog.go`
- **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
- **Lip Gloss Documentation**: https://github.com/charmbracelet/lipgloss

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Verify alignment with specification
   - Confirm phase order

2. **Begin Implementation**
   - Start with Phase 1 (infrastructure)
   - Follow TDD approach
   - Commit after each phase

3. **Verification**
   - Run tests after each phase
   - Manual verification per checklist
