# Implementation Plan: Page Scroll Keybindings

## Overview

Add Vim-like page scrolling functionality (Ctrl+D/Ctrl+U, PageDown/PageUp) to duofm's file list panes and scrollable dialogs, enabling faster navigation through large file lists.

## Objectives

- Implement page-down and page-up functionality for file list panes
- Support both Vim-style (Ctrl+D/U) and standard (PageDown/Up) keybindings
- Handle boundary conditions gracefully (top/bottom of list)
- Extend scrolling support to scrollable dialogs
- Make keybindings customizable through configuration file
- Achieve < 50ms response time for smooth user experience

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make (for build automation)
- Terminal emulator for testing (xterm, kitty, alacritty)

### Dependencies
- github.com/charmbracelet/bubbletea - TUI framework (already installed)
- github.com/charmbracelet/lipgloss - Styling (already installed)
- No new external dependencies required

### Knowledge Requirements
- Understanding of Bubble Tea's Elm Architecture (Model-Update-View)
- Familiarity with duofm's action-based keybinding system
- Basic knowledge of terminal key event handling

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea v0.25.0
- **Key Libraries**:
  - bubbletea - TUI framework and event handling
  - lipgloss - Terminal styling

### Design Approach

The implementation follows duofm's existing **action-based architecture**:

1. **Action Layer**: Define semantic actions (ActionPageDown, ActionPageUp)
2. **Keybinding Layer**: Map keyboard input to actions
3. **Handler Layer**: Dispatch actions to appropriate pane methods
4. **Component Layer**: Implement cursor movement logic in Pane

This layered approach ensures:
- Separation of concerns (input → intent → behavior)
- Easy customization (users can remap keys without changing logic)
- Consistent behavior across different input methods
- Testability at each layer

### Component Interaction

```
User Input (Ctrl+D)
    ↓
KeybindingMap.GetAction("ctrl+d") → ActionPageDown
    ↓
handleAction(ActionPageDown)
    ↓
Pane.MoveCursorPageDown()
    ↓
Pane.adjustScroll() (ensure cursor visible)
    ↓
Model returns updated state → View re-renders
```

**Key Contracts**:
- KeybindingMap: Maps key strings to Action constants
- handleAction: Dispatches actions to appropriate components
- Pane methods: Maintain cursor invariant (0 ≤ cursor < len(entries))
- adjustScroll: Maintains scroll invariant (cursor always visible)

## Implementation Phases

### Phase 1: Core Action Infrastructure

**Goal**: Define page scroll actions and integrate them into the action system so that all layers recognize the new actions.

**Files to Create**:
None (all modifications to existing files)

**Files to Modify**:
- `internal/ui/actions.go`:
  - Add ActionPageDown and ActionPageUp constants
  - Add action name mappings ("page_down", "page_up")
- `internal/config/defaults.go`:
  - Add default keybindings for page_down and page_up actions
- `internal/config/parser.go`:
  - Fix specialKeyMap normalization: "pageup" → "pgup", "pagedown" → "pgdown"
  - Ensures config file keys match Bubble Tea's KeyType constants

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| ActionPageDown | Semantic action constant for page-down operation | Defined in Action enum | Can be used in switch statements and maps |
| ActionPageUp | Semantic action constant for page-up operation | Defined in Action enum | Can be used in switch statements and maps |
| actionNames map | Maps Action constants to configuration names | Action constants exist | String "page_down"/"page_up" map to actions |
| DefaultKeybindings | Provides default key-to-action mappings | Action names exist | Keys "Ctrl+D", "PageDown" map to "page_down" |

**Processing Flow**:

```
1. Add action constants to Action enum
   └─ Position: After existing navigation actions

2. Update actionNames map
   └─ Add entries: ActionPageDown → "page_down", ActionPageUp → "page_up"

3. Update nameToAction map
   └─ Add reverse entries: "page_down" → ActionPageDown, etc.

4. Add default keybindings
   └─ "page_down": ["Ctrl+D", "PageDown"]
   └─ "page_up": ["Ctrl+U", "PageUp"]
```

**Implementation Steps**:

1. **Add Action Constants**
   - Add ActionPageDown and ActionPageUp to the Action enum in actions.go
   - Position them in the "Navigation" section for logical grouping
   - Update comment to reflect new action count

2. **Update Action Mapping Tables**
   - Add entries to actionNames map for string representation
   - Add reverse entries to nameToAction map for config parsing
   - Ensure consistency between forward and reverse mappings

3. **Define Default Keybindings**
   - Add keybinding entries to DefaultKeybindings() in defaults.go
   - Map both Vim-style (Ctrl+D/U) and standard (PageDown/Up) keys
   - Update AllActions() function to include new action names

4. **Fix Key Normalization in parser.go**
   - Update specialKeyMap to normalize "pageup" → "pgup"
   - Update specialKeyMap to normalize "pagedown" → "pgdown"
   - Ensures user config file keys match Bubble Tea's KeyType constants
   - This prevents keybinding mismatches when users configure page scroll keys

**Dependencies**:
- Requires: None (foundational phase)
- Blocks: Phase 2, 3, 4 (all depend on action definitions)

**Testing Approach**:

*Unit Tests*:
- Test Action.String() returns correct names for new actions
- Test ActionFromName() parses "page_down" and "page_up" correctly
- Test DefaultKeybindings() includes all four keybindings
- Test AllActions() includes new action names

*Integration Tests*:
- Verify KeybindingMap can look up actions for "Ctrl+D", "Ctrl+U", "PageDown", "PageUp"
- Verify config file parsing recognizes new action names

*Manual Testing*:
- [ ] Build completes without errors
- [ ] No duplicate action constants
- [ ] Action names follow naming convention (lowercase with underscores)

**Acceptance Criteria**:
- [ ] ActionPageDown and ActionPageUp constants added to actions.go
- [ ] actionNames and nameToAction maps include new actions
- [ ] DefaultKeybindings() includes all four keys (Ctrl+D/U, PageDown/Up)
- [ ] AllActions() includes "page_down" and "page_up"
- [ ] specialKeyMap in parser.go normalizes "pageup" → "pgup" and "pagedown" → "pgdown"
- [ ] Unit tests pass for action name conversion
- [ ] No compilation errors

**Estimated Effort**: 小 (< 1 hour)

**Risks and Mitigation**:
- **Risk**: Action constant numbering conflicts with future additions
  - **Mitigation**: Use iota auto-increment, add constants sequentially

---

### Phase 2: Pane Cursor Movement Logic

**Goal**: Implement the core cursor movement methods in Pane that move the cursor by one page (visible lines) while respecting list boundaries.

**Files to Create**:
None (modifications only)

**Files to Modify**:
- `internal/ui/pane.go`:
  - Add MoveCursorPageDown() method
  - Add MoveCursorPageUp() method
  - Add getVisibleLines() helper method

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| MoveCursorPageDown() | Move cursor down by visible lines | Pane initialized, entries loaded | Cursor moved or at bottom boundary |
| MoveCursorPageUp() | Move cursor up by visible lines | Pane initialized, entries loaded | Cursor moved or at top boundary |
| getVisibleLines() | Calculate number of visible lines in pane | Pane has valid height | Returns positive integer (min 1) |
| adjustScroll() | Ensure cursor is within visible viewport | Cursor updated | Scroll offset adjusted |

**Processing Flow**:

```
MoveCursorPageDown():
1. Calculate visible lines (height - header space)
   └─ Minimum: 1 line (even in tiny panes)
2. Calculate new cursor position (cursor + visibleLines)
3. Clamp to list bounds
   ├─ If newCursor >= len(entries) → newCursor = len(entries) - 1
   └─ If entries empty → no-op
4. Update cursor field
5. Call adjustScroll() to update viewport

MoveCursorPageUp():
1. Calculate visible lines (same as PageDown)
2. Calculate new cursor position (cursor - visibleLines)
3. Clamp to list bounds
   └─ If newCursor < 0 → newCursor = 0
4. Update cursor field
5. Call adjustScroll() to update viewport
```

**Implementation Steps**:

1. **Add getVisibleLines() Helper**
   - Calculate available lines: pane height minus header lines (4 lines total)
   - Return minimum of 1 to handle extremely small panes
   - Follow existing pattern from adjustScroll() calculation

2. **Implement MoveCursorPageDown()**
   - Get visible line count
   - Calculate target cursor position (current + visible)
   - Clamp to valid range [0, len(entries)-1]
   - Update cursor only if position changed
   - Call adjustScroll() to maintain visibility

3. **Implement MoveCursorPageUp()**
   - Same structure as PageDown but subtract visible lines
   - Clamp to minimum of 0
   - Update cursor and adjust scroll

**Dependencies**:
- Requires: Phase 1 (action definitions exist)
- Blocks: Phase 3 (action handlers need these methods)

**Testing Approach**:

*Unit Tests*:
- Test MoveCursorPageDown with normal case (100 entries, cursor at 0)
- Test MoveCursorPageDown near bottom (50 entries, cursor at 40)
- Test MoveCursorPageDown at bottom (cursor at last entry)
- Test MoveCursorPageUp with normal case (100 entries, cursor at 50)
- Test MoveCursorPageUp near top (50 entries, cursor at 10)
- Test MoveCursorPageUp at top (cursor at 0)
- Test with small pane (height 5, 1 visible line)
- Test with empty directory (0 entries)

*Integration Tests*:
- Test cursor position after multiple page scrolls
- Test scroll offset updates correctly
- Test interaction with existing MoveCursorUp/Down methods

*Manual Testing*:
- [ ] Cursor moves by approximately one screen height
- [ ] Cursor stops at top/bottom boundaries
- [ ] No crash with empty directories
- [ ] Works with various terminal sizes

**Acceptance Criteria**:
- [ ] MoveCursorPageDown() moves cursor by visible lines
- [ ] MoveCursorPageUp() moves cursor by visible lines
- [ ] Both methods handle boundary conditions (top/bottom)
- [ ] Minimum movement is 1 line (small pane case)
- [ ] Empty directory case handled gracefully
- [ ] adjustScroll() called to maintain cursor visibility
- [ ] Unit tests pass with > 90% coverage for new methods
- [ ] No regression in existing cursor movement tests

**Estimated Effort**: 小 (2-3 hours)

**Risks and Mitigation**:
- **Risk**: Scroll offset calculation error causes cursor to disappear
  - **Mitigation**: Leverage existing adjustScroll() which is battle-tested
- **Risk**: Edge case with 1-line visible area
  - **Mitigation**: Explicit minimum of 1 line, add specific test case

---

### Phase 3: Action Handler Integration

**Goal**: Connect the action layer to the pane methods by adding handler cases that dispatch page scroll actions to the active pane.

**Files to Create**:
None (modifications only)

**Files to Modify**:
- `internal/ui/model_update_keyboard.go`:
  - Add ActionPageDown case in handleAction()
  - Add ActionPageUp case in handleAction()

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| handleAction() | Dispatch actions to appropriate handlers | Action constant valid | Correct method called on active pane |
| getActivePane() | Return the currently active pane | Model initialized | Returns left or right pane |

**Processing Flow**:

```
handleAction(action):
1. Switch on action value
2. Case ActionPageDown:
   ├─ Get active pane reference
   ├─ Call pane.MoveCursorPageDown()
   └─ Return updated model with no command
3. Case ActionPageUp:
   ├─ Get active pane reference
   ├─ Call pane.MoveCursorPageUp()
   └─ Return updated model with no command
```

**Implementation Steps**:

1. **Add ActionPageDown Handler**
   - Add case for ActionPageDown in handleAction() switch
   - Call getActivePane().MoveCursorPageDown()
   - Return (m, nil) for synchronous update
   - Position near other navigation actions (ActionMoveDown/Up)

2. **Add ActionPageUp Handler**
   - Add case for ActionPageUp in handleAction() switch
   - Call getActivePane().MoveCursorPageUp()
   - Return (m, nil) for synchronous update
   - Mirror structure of ActionPageDown for consistency

**Dependencies**:
- Requires: Phase 1 (action constants), Phase 2 (pane methods)
- Blocks: Phase 4 (testing needs complete integration)

**Testing Approach**:

*Unit Tests*:
- Test handleAction(ActionPageDown) calls MoveCursorPageDown on active pane
- Test handleAction(ActionPageUp) calls MoveCursorPageUp on active pane
- Test action dispatches to correct pane (left vs. right)

*Integration Tests*:
- Test key press flow: KeyMsg → GetAction → handleAction → pane method
- Test Ctrl+D key binding triggers ActionPageDown
- Test Ctrl+U key binding triggers ActionPageUp
- Test PageDown key binding (tea.KeyPgDown) triggers ActionPageDown
- Test PageUp key binding (tea.KeyPgUp) triggers ActionPageUp
- Test mixed navigation (j, Ctrl+D, k, Ctrl+U)

*Manual Testing*:
- [ ] Ctrl+D scrolls down in active pane
- [ ] Ctrl+U scrolls up in active pane
- [ ] PageDown key works the same as Ctrl+D
- [ ] PageUp key works the same as Ctrl+U
- [ ] Only active pane responds to keys

**Acceptance Criteria**:
- [ ] ActionPageDown and ActionPageUp cases added to handleAction()
- [ ] Both cases call correct pane methods
- [ ] Both cases return updated model with nil command
- [ ] Integration tests pass for all four keybindings
- [ ] Active pane detection works correctly
- [ ] No regression in existing action handling tests

**Estimated Effort**: 小 (1-2 hours)

**Risks and Mitigation**:
- **Risk**: Forgetting to return model after mutation
  - **Mitigation**: Follow existing pattern (ActionMoveDown/Up as template)
- **Risk**: Dispatching to wrong pane
  - **Mitigation**: Use getActivePane() like all other pane operations

---

### Phase 4: Testing and Validation

**Goal**: Ensure comprehensive test coverage, validate behavior across edge cases, and verify performance requirements are met.

**Files to Create**:
- `internal/ui/pane_page_scroll_test.go` - Unit tests for page scroll methods

**Files to Modify**:
- `internal/ui/model_keyboard_test.go`:
  - Add integration tests for page scroll key bindings

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Unit Test Suite | Verify pane methods work correctly in isolation | Methods implemented | All boundary cases pass |
| Integration Test Suite | Verify end-to-end key-to-cursor flow | Full stack integrated | All keybindings work |
| E2E Test Suite | Verify behavior in real terminal | duofm binary built | User-visible behavior correct |

**Processing Flow**:

```
Testing Strategy:
1. Unit Tests (Pane methods)
   ├─ Normal cases (middle of list)
   ├─ Boundary cases (top/bottom)
   ├─ Edge cases (empty, small pane)
   └─ Scroll offset verification

2. Integration Tests (Key bindings)
   ├─ Ctrl+D → ActionPageDown → MoveCursorPageDown
   ├─ Ctrl+U → ActionPageUp → MoveCursorPageUp
   ├─ PageDown → same as Ctrl+D
   ├─ PageUp → same as Ctrl+U
   └─ Mixed navigation sequences

3. E2E Tests (Terminal behavior)
   ├─ Visual verification of scroll
   ├─ Boundary handling at top/bottom
   └─ Performance measurement
```

**Implementation Steps**:

1. **Create Unit Test File**
   - Create pane_page_scroll_test.go
   - Add table-driven tests for MoveCursorPageDown()
   - Add table-driven tests for MoveCursorPageUp()
   - Cover all scenarios from spec test cases (TS-1 through TS-8)

2. **Add Integration Tests**
   - Add TestPageDownKeyBinding() to model_keyboard_test.go
   - Add TestPageUpKeyBinding() to model_keyboard_test.go
   - Add TestPageScrollAlternateKeys() for PageDown/PageUp keys
   - Add TestMixedNavigation() for combined j/k/Ctrl+D/U usage

3. **Add E2E Tests** (if E2E framework exists)
   - Add test_page_down() shell function
   - Add test_page_up() shell function
   - Add test_page_scroll_boundary() for top/bottom limits

4. **Performance Validation**
   - Add benchmark test for page scroll operations
   - Verify < 50ms response time with large directory (10,000+ files)

**Dependencies**:
- Requires: Phase 1, 2, 3 (complete implementation)
- Blocks: None (final phase)

**Testing Approach**:

*Unit Tests*:
Test cases from SPEC.md (see VERIFICATION.md for complete list):
- TS-1: MoveCursorPageDown - Normal case
- TS-2: MoveCursorPageDown - Near bottom
- TS-3: MoveCursorPageDown - At bottom
- TS-4: MoveCursorPageUp - Normal case
- TS-5: MoveCursorPageUp - Near top
- TS-6: MoveCursorPageUp - At top
- TS-7: Small pane - Minimum movement
- TS-8: Empty directory

*Integration Tests*:
- IT-1: Ctrl+D keybinding triggers ActionPageDown
- IT-2: PageDown keybinding triggers ActionPageDown
- IT-3: Ctrl+U keybinding triggers ActionPageUp
- IT-4: PageUp keybinding triggers ActionPageUp
- IT-5: Mixed navigation sequence

*E2E Tests*:
- E2E-1: Visual page down in terminal
- E2E-2: Visual page up in terminal
- E2E-3: Boundary at bottom
- E2E-4: Boundary at top

*Manual Testing*:
- [ ] Test with 100 file directory
- [ ] Test with 10,000+ file directory
- [ ] Test with various terminal sizes (80x24, 120x40, 200x60)
- [ ] Test with hidden files on/off
- [ ] Test with different sort orders
- [ ] Test in both left and right panes
- [ ] Test pane switching + page scroll

**Acceptance Criteria**:
- [ ] All unit tests pass (> 90% coverage on new code)
- [ ] All integration tests pass
- [ ] E2E tests pass (if framework available)
- [ ] Performance benchmark shows < 50ms response time
- [ ] No regression in existing tests
- [ ] Code coverage report shows adequate coverage
- [ ] Manual testing checklist completed

**Estimated Effort**: 中 (4-6 hours)

**Risks and Mitigation**:
- **Risk**: Tests fail due to terminal size variations
  - **Mitigation**: Use fixed test dimensions (e.g., height=24)
- **Risk**: E2E tests flaky due to timing
  - **Mitigation**: Add explicit waits, increase timeouts
- **Risk**: Performance varies by machine
  - **Mitigation**: Use relative benchmarks, test on slow hardware

---

### Phase 5: Dialog Support (Required)

**Goal**: Extend page scroll functionality to scrollable dialogs (HelpDialog, PermissionErrorReportDialog) for consistent user experience across all scrollable UI elements.

**Note**: PermissionErrorReportDialog already has page scroll support. This phase focuses on HelpDialog and ensures all scrollable dialogs have consistent keybindings (FR1.11). This is a required phase to satisfy requirement FR1.11.

**Files to Create**:
None (modifications only)

**Files to Modify**:
- `internal/ui/help_dialog.go`:
  - Add PageDown/PageUp handling in Update() method (if not present)
- Any other scrollable dialog files identified during implementation

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Dialog.Update() | Handle keyboard input in dialog | Dialog active | Scroll position updated |
| Dialog scroll state | Track current scroll offset | Dialog has scrollable content | Offset within bounds |

**Processing Flow**:

```
Dialog Page Scroll:
1. Detect if dialog is active
2. Check if content is scrollable
   ├─ If content fits in viewport → ignore page keys
   └─ If content exceeds viewport → handle page scroll
3. On Ctrl+D/PageDown:
   └─ Increment scroll offset by viewport height
4. On Ctrl+U/PageUp:
   └─ Decrement scroll offset by viewport height
5. Clamp scroll offset to valid range
```

**Implementation Steps**:

1. **Review Existing Dialog Implementations**
   - Examine permission_error_report_dialog.go (reference implementation)
   - Identify which dialogs have scrollable content
   - List dialogs that need page scroll support

2. **Add Page Scroll to HelpDialog** (if needed)
   - Check if HelpDialog.Update() handles page scroll keys
   - Add key handling for Ctrl+D/U and PageDown/Up
   - Test with long help content

3. **Ensure Consistency**
   - Verify all scrollable dialogs use same keybindings
   - Test non-scrollable dialogs ignore page scroll keys

**Dependencies**:
- Requires: Phase 3 (action system working)
- Blocks: None (final implementation phase)

**Testing Approach**:

*Unit Tests*:
- Test dialog page scroll with scrollable content
- Test dialog ignores page keys when content fits
- Test scroll offset clamping at top/bottom

*Integration Tests*:
- Test page scroll in HelpDialog
- Test page scroll in PermissionErrorReportDialog (regression test)

*Manual Testing*:
- [ ] Open HelpDialog, press Ctrl+D, verify content scrolls
- [ ] Press Ctrl+U at top of dialog, verify no error
- [ ] Press Ctrl+D at bottom of dialog, verify no error
- [ ] Test with short dialog content (< viewport), verify no scroll

**Acceptance Criteria**:
- [ ] HelpDialog supports page scroll (if content is scrollable)
- [ ] PermissionErrorReportDialog still works (no regression)
- [ ] Non-scrollable dialogs gracefully ignore page keys
- [ ] Keybindings consistent across all dialogs
- [ ] Unit tests pass for dialog page scroll
- [ ] Manual testing checklist completed

**Estimated Effort**: 小 (2-3 hours)

**Risks and Mitigation**:
- **Risk**: Breaking existing dialog behavior
  - **Mitigation**: Start with tests for existing behavior (regression tests)
- **Risk**: Different dialogs have different scroll implementations
  - **Mitigation**: Extract common scroll logic if needed

---

## Complete File Structure

```
duofm/
├── cmd/duofm/
│   └── main.go                             # Entry point (no changes)
├── internal/
│   ├── ui/
│   │   ├── actions.go                      # ✏️ Add ActionPageDown/PageUp
│   │   ├── pane.go                         # ✏️ Add MoveCursorPageDown/Up
│   │   ├── pane_page_scroll_test.go        # ➕ NEW: Unit tests for page scroll
│   │   ├── model_update_keyboard.go        # ✏️ Add action handlers
│   │   ├── model_keyboard_test.go          # ✏️ Add integration tests
│   │   ├── help_dialog.go                  # ✏️ Add page scroll (Phase 5)
│   │   └── permission_error_report_dialog.go # Reference implementation
│   ├── config/
│   │   └── defaults.go                     # ✏️ Add default keybindings
│   └── fs/
│       └── (no changes)
├── go.mod
├── go.sum
├── Makefile
└── README.md

Legend:
✏️ = Modified
➕ = New file
```

**File Descriptions**:

- **actions.go**: Defines all action constants and name mappings. Contains the semantic action layer that decouples input from behavior.

- **pane.go**: Contains Pane struct and all pane-related methods. Handles file list display, cursor movement, and scroll management.

- **pane_page_scroll_test.go**: Unit tests for MoveCursorPageDown/Up methods. Tests boundary conditions, edge cases, and scroll offset updates.

- **model_update_keyboard.go**: Keyboard input handler. Dispatches actions to appropriate components based on application state.

- **model_keyboard_test.go**: Integration tests for keyboard handling. Tests end-to-end flow from key press to UI update.

- **defaults.go**: Default configuration values including keybindings. Users can override these in config.toml.

- **help_dialog.go**: Help screen dialog. May need page scroll support if content exceeds viewport.

- **permission_error_report_dialog.go**: Already has page scroll implementation, serves as reference for other dialogs.

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- Create temporary directories for file system tests
- Mock terminal dimensions in test environment

**Test Coverage Goals**:
- Pane methods (MoveCursorPageDown/Up): 95%+ coverage
- Action handlers: 90%+ coverage
- Edge cases: 100% coverage (empty dir, small pane, boundaries)

**Key Test Areas**:

1. **Pane Methods** (`internal/ui/pane_page_scroll_test.go`)
   - Normal operation (middle of list)
   - Boundary conditions (top/bottom)
   - Edge cases (empty directory, 1-line pane)
   - Scroll offset correctness
   - Cursor visibility maintenance

2. **Action System** (`internal/ui/actions_test.go` - existing file)
   - Action name conversion (ActionPageDown ↔ "page_down")
   - DefaultKeybindings includes new actions
   - AllActions list completeness

3. **Keyboard Integration** (`internal/ui/model_keyboard_test.go`)
   - Ctrl+D triggers ActionPageDown
   - Ctrl+U triggers ActionPageUp
   - PageDown key triggers ActionPageDown
   - PageUp key triggers ActionPageUp
   - Mixed navigation sequences (j, k, Ctrl+D, Ctrl+U)

### Integration Testing

**Scenarios**:
1. End-to-end keybinding flow (key press → action → cursor movement)
2. Multi-step navigation (page down × 3, then page up × 2)
3. Pane switching + page scroll
4. Large directory performance (10,000+ files)

**Approach**:
- Use Model integration tests with simulated key messages
- Test with realistic file counts and terminal dimensions
- Verify cursor position and scroll offset after each operation

### Manual Testing Checklist

Based on spec test scenarios:

**Basic Functionality**:
- [ ] Ctrl+D moves cursor down by approximately one screen
- [ ] Ctrl+U moves cursor up by approximately one screen
- [ ] PageDown works the same as Ctrl+D
- [ ] PageUp works the same as Ctrl+U
- [ ] Cursor stops at bottom when scrolling down
- [ ] Cursor stops at top when scrolling up

**Edge Cases**:
- [ ] Empty directory: no crash, cursor stays at 0
- [ ] Single file: page scroll moves to same position (no error)
- [ ] Very small pane (5 lines): moves by minimum 1 line
- [ ] Very large directory (10,000+ files): smooth and fast (< 50ms)

**User Experience**:
- [ ] Screen updates smoothly without flicker
- [ ] Cursor position is visible after scroll
- [ ] Works in both left and right panes
- [ ] Works with hidden files on/off
- [ ] Works with different sort orders

**Dialog Support** (Phase 5):
- [ ] HelpDialog scrolls with Ctrl+D/U
- [ ] Dialog boundaries respected
- [ ] Short dialogs ignore page keys gracefully

### E2E Testing

**Test Environment**:
- Terminal: xterm, kitty, alacritty
- Directory: 100+ files for realistic testing
- Screen sizes: 80×24, 120×40, 200×60

**Test Cases**:
- Visual verification of page scroll behavior
- Boundary handling at list ends
- Performance measurement with large directories
- Keybinding consistency across terminals

## Dependencies

### External Dependencies

| Package | Version | Purpose | Installation |
|---------|---------|---------|--------------|
| github.com/charmbracelet/bubbletea | v0.25.0 | TUI framework | Already installed |
| github.com/charmbracelet/lipgloss | v0.9.1 | Terminal styling | Already installed |

No new dependencies required.

### Internal Dependencies

**Implementation Order** (respecting dependencies):

```
Phase 1: Core Action Infrastructure
  ↓ (actions must exist before use)
Phase 2: Pane Cursor Movement Logic
  ↓ (methods must exist before calling)
Phase 3: Action Handler Integration
  ↓ (integration must work before testing)
Phase 4: Testing and Validation
  ↓ (validation must complete before dialog extension)
Phase 5: Dialog Support
```

**Component Dependencies**:
- `model_update_keyboard.go` depends on `actions.go` (action constants)
- `model_update_keyboard.go` depends on `pane.go` (cursor methods)
- `defaults.go` depends on `actions.go` (action names)
- All tests depend on complete implementation

## Risk Assessment

### Technical Risks

1. **Terminal Compatibility Issues**
   - **Risk**: Different terminals send different key codes for Ctrl+D/U
   - **Likelihood**: Low (Bubble Tea normalizes key events)
   - **Impact**: Medium (feature unusable in some terminals)
   - **Mitigation**:
     - Test on multiple terminal emulators (xterm, kitty, alacritty, GNOME Terminal)
     - Rely on Bubble Tea's key normalization
     - Document minimum terminal requirements if issues found

2. **Performance with Large Directories**
   - **Risk**: Page scroll feels sluggish with 10,000+ files
   - **Likelihood**: Low (simple arithmetic operation)
   - **Impact**: Medium (poor UX in large dirs)
   - **Mitigation**:
     - Profile with pprof if performance issues arise
     - Page scroll is O(1) operation (no iteration)
     - Existing adjustScroll() already handles large lists efficiently

3. **Scroll Offset Calculation Bugs**
   - **Risk**: Cursor becomes invisible or scroll position incorrect
   - **Likelihood**: Low (reusing existing adjustScroll())
   - **Impact**: High (breaks navigation)
   - **Mitigation**:
     - Leverage battle-tested adjustScroll() method
     - Comprehensive unit tests for boundary cases
     - Manual testing across various scenarios

### Implementation Risks

1. **Scope Creep**
   - **Risk**: Adding features beyond spec (e.g., half-page scroll, custom page sizes)
   - **Mitigation**: Stick strictly to spec requirements, document future enhancements separately

2. **Breaking Existing Keybindings**
   - **Risk**: Ctrl+D/U conflicts with other functionality
   - **Mitigation**: Review existing keybindings before implementation, test for regressions

3. **Dialog Integration Complexity**
   - **Risk**: Each dialog has different scroll implementation
   - **Mitigation**: Start with reference implementation (PermissionErrorReportDialog), extract common pattern if needed

## Performance Considerations

1. **Cursor Movement Efficiency**
   - Operation: O(1) arithmetic (cursor + visibleLines)
   - No iteration over entries
   - No memory allocation

2. **Scroll Offset Calculation**
   - Reuses existing adjustScroll() method
   - Simple comparison and assignment
   - Already optimized for large directories

3. **Rendering Performance**
   - Bubble Tea handles efficient re-rendering
   - Only changed regions repainted
   - No full screen redraw needed

4. **Expected Performance**:
   - Target: < 50ms from key press to screen update
   - Reality: < 10ms for typical cases (simple arithmetic)
   - Large directories (10,000+ files): Still < 20ms (no file I/O in scroll)

## Security Considerations

1. **Input Validation**
   - No user input to validate (keyboard events only)
   - Cursor position always clamped: `0 ≤ cursor < len(entries)`
   - No risk of out-of-bounds access

2. **Resource Usage**
   - No additional memory allocation during scroll
   - No file system access during scroll
   - No privilege escalation concerns

3. **Terminal Security**
   - Standard Bubble Tea key handling (safe)
   - No raw terminal manipulation
   - No escape sequence injection risk

**Conclusion**: This feature introduces no new security risks. All operations are in-memory cursor position updates with proper bounds checking.

## Open Questions

### From Specification:
None - all requirements have been clarified with the user.

### Implementation-Specific:
None - all decisions resolved during planning:
- ✅ Use existing adjustScroll() mechanism
- ✅ Follow existing action-based architecture
- ✅ Minimum movement is 1 line for small panes
- ✅ Empty directory is handled gracefully (no-op)

### To Clarify with User:
None at this time. Proceed with implementation.

## Future Enhancements

Items deferred to later phases or releases:

### Not in Current Spec:
- **Half-page scroll** (Ctrl+D/U in some editors moves by half page, not full page)
- **Configurable page size** (e.g., allow user to set page size to 10 lines instead of viewport height)
- **Page scroll with jump to first/last** (e.g., gg/G in Vim)
- **Smooth scrolling animation** (gradually move cursor instead of instant jump)
- **Page scroll in other UI contexts** (e.g., file preview pane if added in future)

### Dialog Enhancements (if Phase 5 reveals patterns):
- **Unified dialog scroll interface** (abstract base class for scrollable dialogs)
- **Scroll indicators** (show scroll position in dialog, e.g., "Page 2 of 5")
- **Mouse wheel support** (scroll with mouse in addition to keyboard)

## Success Criteria

- [ ] All functional requirements implemented (FR1.1 - FR1.13)
- [ ] All non-functional requirements met (NFR1.1 - NFR1.7)
- [ ] Unit test coverage > 90% for new code
- [ ] All integration tests pass
- [ ] Performance < 50ms response time (verified with benchmark)
- [ ] Keybindings work in config file
- [ ] Manual testing checklist completed
- [ ] Code review completed with no major issues
- [ ] No regression in existing functionality
- [ ] Works on common terminals (xterm, kitty, alacritty)
- [ ] Documentation updated (if help dialog changes)

## References

- **Specification**: `doc/tasks/page-scroll-keybindings/SPEC.md`
- **Bubble Tea Documentation**: https://github.com/charmbracelet/bubbletea
- **Bubble Tea Key Events**: https://github.com/charmbracelet/bubbletea/blob/master/key.go
- **Go Testing**: https://go.dev/doc/tutorial/add-a-test
- **Vim Documentation**: `:help CTRL-D` and `:help CTRL-U` for expected behavior
- **Existing Implementation References**:
  - `internal/ui/pane.go` - MoveCursorUp/Down pattern
  - `internal/ui/permission_error_report_dialog.go` - Dialog page scroll reference
  - `internal/ui/model_update_keyboard.go` - Action handler pattern

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Review plan for completeness and accuracy
   - Confirm approach and timeline
   - Resolve any remaining questions

2. **Run Verification Check**
   - Execute `/sdd.3-verify-plan` to perform consistency check
   - Review feedback from second opinion
   - Address any issues found

3. **Begin Implementation**
   - Start with Phase 1 (Core Action Infrastructure)
   - Follow phases sequentially
   - Commit incrementally with clear messages

4. **Continuous Testing**
   - Write tests before or alongside implementation (TDD)
   - Run tests frequently (`go test ./...`)
   - Monitor coverage (`go test -cover ./...`)

5. **Final Validation**
   - Complete manual testing checklist
   - Run E2E tests (if available)
   - Performance benchmarking
   - Code review before merge
