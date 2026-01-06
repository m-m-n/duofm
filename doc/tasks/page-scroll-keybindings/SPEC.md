# Feature: Page Scroll Keybindings

## Overview

This feature adds Vim-like page scrolling functionality to duofm's file list panes and scrollable dialogs. Users will be able to quickly navigate through large file lists using `Ctrl+U` (page up) and `Ctrl+D` (page down), as well as the standard `PageUp` and `PageDown` keys.

**Key Benefits:**
- Faster navigation in directories with many files
- Familiar keybindings for Vim users
- Consistent scrolling behavior across the application
- Customizable keybindings via configuration file

## Objectives

- Implement page-down functionality with `Ctrl+D` and `PageDown` keys
- Implement page-up functionality with `Ctrl+U` and `PageUp` keys
- Handle boundary conditions gracefully (top/bottom of list)
- Extend scrolling support to scrollable dialogs
- Make keybindings customizable through config file
- Maintain consistency with existing navigation (j/k keys)
- Achieve < 50ms response time for smooth user experience

## User Stories

### US1: Quick Navigation in Large Directories
As a user navigating a directory with hundreds of files, I want to scroll by full pages, so that I can reach distant files quickly without repeatedly pressing j/k.

**Acceptance Criteria:**
- [x] Ctrl+D moves cursor down by one page (visible lines)
- [x] Ctrl+U moves cursor up by one page (visible lines)
- [x] PageDown key works the same as Ctrl+D
- [x] PageUp key works the same as Ctrl+U
- [x] Cursor stops at list boundaries (top/bottom)
- [x] Screen updates smoothly without flicker

### US2: Vim-like Workflow
As a Vim user, I want to use Ctrl+D and Ctrl+U for page scrolling, so that I can use muscle memory from my text editor.

**Acceptance Criteria:**
- [x] Ctrl+D scrolls down (matches Vim behavior)
- [x] Ctrl+U scrolls up (matches Vim behavior)
- [x] Keybindings work in active pane only
- [x] Behavior is consistent across application

### US3: Dialog Scrolling
As a user viewing long dialog content (e.g., error reports), I want to use the same keys to scroll, so that I have a consistent experience.

**Acceptance Criteria:**
- [x] Ctrl+D/PageDown scroll dialog content down
- [x] Ctrl+U/PageUp scroll dialog content up
- [x] Works for all scrollable dialogs
- [x] Does not affect non-scrollable dialogs

### US4: Customizable Keybindings
As a user with custom workflow, I want to change page scroll keybindings, so that they don't conflict with my other tools.

**Acceptance Criteria:**
- [x] Can define custom keys in config.toml
- [x] Multiple keys can be assigned to same action
- [x] Invalid key configurations are handled gracefully
- [x] Default keybindings work without config file

## Technical Requirements

### Functional Requirements

- **FR1.1:** System SHALL move cursor down by visible line count when user presses Ctrl+D
- **FR1.2:** System SHALL move cursor up by visible line count when user presses Ctrl+U
- **FR1.3:** System SHALL support PageDown key as alias for Ctrl+D
- **FR1.4:** System SHALL support PageUp key as alias for Ctrl+U
- **FR1.5:** System SHALL stop cursor at list bottom when scrolling down
- **FR1.6:** System SHALL stop cursor at list top when scrolling up
- **FR1.7:** System SHALL calculate visible lines as: pane height - header lines (4)
- **FR1.8:** System SHALL maintain minimum movement of 1 line even in small panes
- **FR1.9:** System SHALL update scroll offset to keep cursor visible
- **FR1.10:** System SHALL redraw screen after cursor movement
- **FR1.11:** System SHALL apply same behavior to scrollable dialogs
- **FR1.12:** System SHALL allow keybinding customization via config file
- **FR1.13:** System SHALL use action names "page_down" and "page_up"

### Non-Functional Requirements

- **NFR1.1 - Performance:** Key press to screen update SHALL complete in < 50ms
- **NFR1.2 - Performance:** SHALL work efficiently with 10,000+ files
- **NFR1.3 - Usability:** SHALL work without documentation for Vim users
- **NFR1.4 - Compatibility:** SHALL not break existing keybindings (j/k)
- **NFR1.5 - Compatibility:** SHALL work on all common terminal emulators
- **NFR1.6 - Maintainability:** SHALL follow existing code patterns in pane.go
- **NFR1.7 - Testability:** SHALL be covered by unit and E2E tests

## Implementation Approach

### Architecture

The implementation follows duofm's existing action-based architecture:

```
User Input (Ctrl+D/U)
    ↓
KeybindingMap (maps key → Action)
    ↓
handleAction() (dispatches Action)
    ↓
Pane.MoveCursorPageDown() or Pane.MoveCursorPageUp()
    ↓
Pane.adjustScroll() (adjust scroll offset)
    ↓
Screen Redraw
```

**Components Involved:**
1. **Action Definition** (`internal/ui/actions.go`): Define `ActionPageDown` / `ActionPageUp`
2. **Keybinding Map** (`internal/ui/keybinding_map.go`): Map keys to actions
3. **Default Keybindings** (`internal/config/defaults.go`): Set default key assignments
4. **Pane Methods** (`internal/ui/pane.go`): Implement cursor movement logic
5. **Action Handler** (`internal/ui/model_update_keyboard.go`): Dispatch to pane methods
6. **Dialog Handlers**: Extend existing dialogs with page scroll support

### Data Flow

#### Page Down Flow
```
User: Ctrl+D
    ↓
KeyMsg{String: "ctrl+d"}
    ↓
GetAction("ctrl+d") → ActionPageDown
    ↓
handleAction(ActionPageDown)
    ↓
pane.MoveCursorPageDown()
    ↓
Calculate: newCursor = cursor + visibleLines
    ↓
Clamp: newCursor = min(newCursor, len(entries)-1)
    ↓
cursor = newCursor
    ↓
adjustScroll()
    ↓
Return updated model
```

#### Page Up Flow
```
User: Ctrl+U
    ↓
KeyMsg{String: "ctrl+u"}
    ↓
GetAction("ctrl+u") → ActionPageUp
    ↓
handleAction(ActionPageUp)
    ↓
pane.MoveCursorPageUp()
    ↓
Calculate: newCursor = cursor - visibleLines
    ↓
Clamp: newCursor = max(newCursor, 0)
    ↓
cursor = newCursor
    ↓
adjustScroll()
    ↓
Return updated model
```

### API Design

#### New Actions

```go
// In internal/ui/actions.go
const (
    // ... existing actions
    ActionPageDown    // Ctrl+D, PageDown
    ActionPageUp      // Ctrl+U, PageUp
)

var actionNames = map[Action]string{
    // ... existing mappings
    ActionPageDown:  "page_down",
    ActionPageUp:    "page_up",
}

var nameToAction = map[string]Action{
    // ... existing mappings
    "page_down": ActionPageDown,
    "page_up":   ActionPageUp,
}
```

#### Pane Methods

```go
// In internal/ui/pane.go

// MoveCursorPageDown moves cursor down by one page (visible lines)
func (p *Pane) MoveCursorPageDown() {
    visibleLines := p.getVisibleLines()
    if visibleLines < 1 {
        visibleLines = 1  // Minimum 1 line
    }

    newCursor := p.cursor + visibleLines
    if newCursor >= len(p.entries) {
        newCursor = len(p.entries) - 1
    }

    if newCursor != p.cursor && newCursor >= 0 {
        p.cursor = newCursor
        p.adjustScroll()
    }
}

// MoveCursorPageUp moves cursor up by one page (visible lines)
func (p *Pane) MoveCursorPageUp() {
    visibleLines := p.getVisibleLines()
    if visibleLines < 1 {
        visibleLines = 1  // Minimum 1 line
    }

    newCursor := p.cursor - visibleLines
    if newCursor < 0 {
        newCursor = 0
    }

    if newCursor != p.cursor {
        p.cursor = newCursor
        p.adjustScroll()
    }
}

// getVisibleLines returns the number of lines visible in the pane
func (p *Pane) getVisibleLines() int {
    // height - header(2) - border(1) - status(1) = height - 4
    return p.height - 4
}
```

#### Default Keybindings

```go
// In internal/config/defaults.go
func DefaultKeybindings() map[string][]string {
    return map[string][]string{
        // ... existing bindings
        "page_down": {"Ctrl+D", "PageDown"},
        "page_up":   {"Ctrl+U", "PageUp"},
    }
}
```

#### Action Handler

```go
// In internal/ui/model_update_keyboard.go
func (m Model) handleAction(action Action) (tea.Model, tea.Cmd) {
    switch action {
    // ... existing actions

    case ActionPageDown:
        m.getActivePane().MoveCursorPageDown()
        return m, nil

    case ActionPageUp:
        m.getActivePane().MoveCursorPageUp()
        return m, nil

    // ... rest of actions
    }
}
```

### File Structure

Files to modify:

```
internal/
├── ui/
│   ├── actions.go                    # Add ActionPageDown/ActionPageUp
│   ├── pane.go                       # Add MoveCursorPage{Down,Up}()
│   ├── pane_test.go                  # Add unit tests
│   ├── model_update_keyboard.go      # Add action handlers
│   ├── model_keyboard_test.go        # Add integration tests
│   └── permission_error_report_dialog.go  # Already has page scroll
└── config/
    └── defaults.go                   # Add default keybindings
```

Files already supporting page scroll (reference implementation):
```
internal/ui/permission_error_report_dialog.go
internal/ui/permission_error_report_dialog_test.go
```

## Test Scenarios

### Unit Tests

#### Test: MoveCursorPageDown - Normal Case
- **Setup:** Pane with 100 entries, cursor at position 0, height 24 (20 visible lines)
- **Action:** Call `MoveCursorPageDown()`
- **Expected:** Cursor moves to position 20, scroll offset adjusted

#### Test: MoveCursorPageDown - Near Bottom
- **Setup:** Pane with 50 entries, cursor at position 40, 20 visible lines
- **Action:** Call `MoveCursorPageDown()`
- **Expected:** Cursor moves to position 49 (last entry)

#### Test: MoveCursorPageDown - At Bottom
- **Setup:** Pane with 50 entries, cursor at position 49
- **Action:** Call `MoveCursorPageDown()`
- **Expected:** Cursor stays at position 49

#### Test: MoveCursorPageUp - Normal Case
- **Setup:** Pane with 100 entries, cursor at position 50, 20 visible lines
- **Action:** Call `MoveCursorPageUp()`
- **Expected:** Cursor moves to position 30

#### Test: MoveCursorPageUp - Near Top
- **Setup:** Pane with 50 entries, cursor at position 10, 20 visible lines
- **Action:** Call `MoveCursorPageUp()`
- **Expected:** Cursor moves to position 0 (first entry)

#### Test: MoveCursorPageUp - At Top
- **Setup:** Pane with 50 entries, cursor at position 0
- **Action:** Call `MoveCursorPageUp()`
- **Expected:** Cursor stays at position 0

#### Test: Small Pane - Minimum Movement
- **Setup:** Pane with height 5 (1 visible line)
- **Action:** Call `MoveCursorPageDown()`
- **Expected:** Cursor moves by 1 line (minimum)

#### Test: Empty Directory
- **Setup:** Pane with 0 entries
- **Action:** Call `MoveCursorPageDown()`
- **Expected:** No error, cursor stays at 0

### Integration Tests

#### Test: Ctrl+D Key Binding
- **Setup:** Model with active pane, cursor at position 0
- **Action:** Send `tea.KeyMsg{String: "ctrl+d"}`
- **Expected:** `ActionPageDown` triggered, cursor moves down one page

#### Test: PageDown Key Binding
- **Setup:** Model with active pane, cursor at position 0
- **Action:** Send `tea.KeyMsg{Type: tea.KeyPgDown}`
- **Expected:** Same behavior as Ctrl+D

#### Test: Ctrl+U Key Binding
- **Setup:** Model with active pane, cursor at position 50
- **Action:** Send `tea.KeyMsg{String: "ctrl+u"}`
- **Expected:** `ActionPageUp` triggered, cursor moves up one page

#### Test: PageUp Key Binding
- **Setup:** Model with active pane, cursor at position 50
- **Action:** Send `tea.KeyMsg{Type: tea.KeyPgUp}`
- **Expected:** Same behavior as Ctrl+U

#### Test: Mixed Navigation
- **Setup:** Pane with 100 entries
- **Actions:**
  1. Press `j` 5 times (cursor at 5)
  2. Press `Ctrl+D` (cursor at 25)
  3. Press `k` 2 times (cursor at 23)
  4. Press `Ctrl+U` (cursor at 3)
- **Expected:** All movements work correctly, cursor at expected positions

### E2E Tests

#### Test: Page Down in File List
```bash
test_page_down() {
    start_duofm "$CURRENT_SESSION"

    # Verify starting position
    assert_cursor_position "$CURRENT_SESSION" "1" "Start at first entry"

    # Press Ctrl+D
    send_keys "$CURRENT_SESSION" "C-d"
    sleep 0.2

    # Verify cursor moved down by visible lines
    # Assuming ~20 visible lines in test environment
    assert_cursor_position "$CURRENT_SESSION" "20" "Cursor moved one page down"

    stop_duofm "$CURRENT_SESSION"
}
```

#### Test: Page Up in File List
```bash
test_page_up() {
    start_duofm "$CURRENT_SESSION"

    # Move to middle of list
    send_keys "$CURRENT_SESSION" "C-d" "C-d"
    sleep 0.3

    # Press Ctrl+U
    send_keys "$CURRENT_SESSION" "C-u"
    sleep 0.2

    # Verify cursor moved up by visible lines
    assert_cursor_position "$CURRENT_SESSION" "20" "Cursor moved one page up"

    stop_duofm "$CURRENT_SESSION"
}
```

#### Test: Boundary at Bottom
```bash
test_page_down_boundary() {
    start_duofm "$CURRENT_SESSION"

    # Press Ctrl+D multiple times to reach bottom
    send_keys "$CURRENT_SESSION" "C-d" "C-d" "C-d" "C-d" "C-d"
    sleep 0.5

    # Get final cursor position
    capture_screen "$CURRENT_SESSION"

    # Press Ctrl+D again - should not move
    send_keys "$CURRENT_SESSION" "C-d"
    sleep 0.2

    # Verify cursor stayed at bottom
    assert_contains "$CURRENT_SESSION" "cursor_at_last_entry" "Cursor at bottom"

    stop_duofm "$CURRENT_SESSION"
}
```

#### Test: Boundary at Top
```bash
test_page_up_boundary() {
    start_duofm "$CURRENT_SESSION"

    # Move down first
    send_keys "$CURRENT_SESSION" "C-d"
    sleep 0.2

    # Move back up to top
    send_keys "$CURRENT_SESSION" "C-u"
    sleep 0.2

    # Try to go above top
    send_keys "$CURRENT_SESSION" "C-u"
    sleep 0.2

    # Verify cursor stayed at top
    assert_cursor_position "$CURRENT_SESSION" "1" "Cursor stayed at top"

    stop_duofm "$CURRENT_SESSION"
}
```

### Edge Cases

- **Edge Case 1: Extremely Small Pane**
  - Height = 5 (1 visible line after headers)
  - Expected: Page scroll moves by 1 line
  - Validation: Cursor advances correctly

- **Edge Case 2: File Count Less Than Page Size**
  - 10 files, 20 visible lines
  - Expected: Page down goes to bottom, page up goes to top
  - Validation: No out-of-bounds errors

- **Edge Case 3: Rapid Key Presses**
  - User holds Ctrl+D
  - Expected: Multiple page scrolls until bottom
  - Validation: Stops at bottom, no crash

- **Edge Case 4: Dialog with No Scrollable Content**
  - Dialog content fits in window
  - Expected: Page scroll keys ignored or no-op
  - Validation: Dialog remains stable

### Performance Tests

- **Load Test: Large Directory (10,000 files)**
  - Measure time from key press to screen update
  - Target: < 50ms
  - Method: Use Go benchmark tests

- **Stress Test: Continuous Scrolling**
  - Scroll from top to bottom and back 100 times
  - Target: Consistent performance, no memory leaks
  - Method: Monitor with pprof

## Security Considerations

- **Input Validation:** No special security concerns - cursor position is always clamped to valid range
- **Resource Usage:** No additional memory allocation during scroll operations
- **Terminal Security:** Standard key handling, no raw terminal manipulation

This feature does not introduce any new security risks as it only manipulates in-memory cursor positions using existing boundary checks.

## Error Handling

### Error Codes

No new error codes required. The feature handles all edge cases through boundary checking.

### Error Flow

```
Invalid State → Check Boundaries → Clamp to Valid Range → Continue
```

**Error Prevention:**
- Cursor position always clamped: `0 <= cursor < len(entries)`
- Scroll offset always valid: adjusted by `adjustScroll()`
- Empty directory check: operations no-op when `len(entries) == 0`

## Performance Optimization

### Performance Goals
- Response time: < 50ms for 99% of operations
- Memory overhead: 0 bytes (pure computation)
- CPU usage: Negligible (simple arithmetic)

### Optimization Strategies
- **No Allocations:** Reuse existing `adjustScroll()` mechanism
- **Early Exit:** Check boundaries before calculation
- **Minimal Redraw:** Only changed regions repainted (Bubble Tea handles this)

### Caching Strategy
No caching needed - `getVisibleLines()` is a simple subtraction:
```go
return p.height - 4
```

## Success Criteria

- [x] All functional requirements are implemented
- [x] All unit tests pass (> 80% coverage)
- [x] All E2E tests pass
- [x] Performance < 50ms response time
- [x] Keybindings work in config file
- [x] Documentation is complete
- [x] Code review completed
- [x] No regression in existing functionality
- [x] Works on common terminals (xterm, kitty, alacritty)

## Open Questions

None - all requirements have been clarified with the user.

## Implementation Phases

### Phase 1: Core Implementation
**Goals:** Implement basic page scroll for panes
**Deliverables:**
- Add `ActionPageDown` and `ActionPageUp` to actions.go
- Implement `MoveCursorPageDown()` and `MoveCursorPageUp()` in pane.go
- Add default keybindings to config/defaults.go
- Add action handlers to model_update_keyboard.go
- Write unit tests for pane methods

### Phase 2: Integration & Testing
**Goals:** Integrate with keybinding system and test thoroughly
**Deliverables:**
- Add integration tests to model_keyboard_test.go
- Add E2E tests to test/e2e/scripts/run_all_tests.sh
- Test with large directories (10,000+ files)
- Test boundary conditions
- Verify config file customization

### Phase 3: Dialog Support
**Goals:** Extend to scrollable dialogs
**Deliverables:**
- Review existing dialog implementations
- Add page scroll to HelpDialog (if needed)
- Ensure PermissionErrorReportDialog continues working
- Test dialog scrolling
- Document dialog scrolling behavior

### Phase 4: Polish & Documentation
**Goals:** Finalize and document
**Deliverables:**
- Update help dialog with new keybindings
- Add to README or user documentation
- Performance profiling and optimization
- Final code review
- Merge to main branch

## References

- **Vim Documentation:** `:help CTRL-D` and `:help CTRL-U`
- **Bubble Tea Framework:** https://github.com/charmbracelet/bubbletea
- **Existing Implementation:**
  - `internal/ui/pane.go` - `MoveCursorUp()` / `MoveCursorDown()`
  - `internal/ui/permission_error_report_dialog.go` - Dialog page scroll reference
- **Related Tasks:**
  - `doc/tasks/bookmark/SPEC.md` - Similar navigation patterns
