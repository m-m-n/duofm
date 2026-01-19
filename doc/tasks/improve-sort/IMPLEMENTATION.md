# Implementation Plan: Improve Sort Dialog with Dropdown Menus

## Overview

Replace the current horizontal option selection in the sort dialog with a dropdown menu interface. This provides a more intuitive and visually clear user experience while maintaining all existing sort functionality.

## Objectives

- Replace horizontal option selection (h/l navigation) with dropdown menus
- Provide clearer visual feedback for current selections
- Align with standard UI conventions for option selection
- Maintain backwards compatibility with existing SortConfig structure
- Follow dialog best practices for message-based state management

## Prerequisites

### Development Environment
- Go 1.21 or later
- Make (for build automation)

### Dependencies
- Bubble Tea framework (existing)
- lipgloss styling library (existing)

### Knowledge Requirements
- Bubble Tea message-based architecture
- Dialog lifecycle and cancellation patterns (see doc/development/DIALOG_BEST_PRACTICES.md)
- Existing SortDialog implementation structure

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (github.com/charmbracelet/bubbletea)
- **Styling**: Lip Gloss (github.com/charmbracelet/lipgloss)

### Design Approach

The implementation introduces a reusable Dropdown component that can be embedded in dialogs. The SortDialog will contain two Dropdown instances (Sort by, Order) and manage focus between them.

**Key Design Decisions**:
1. **Dropdown as Embedded Component**: Not a full Dialog, but a sub-component with its own state
2. **Focus Management**: Dialog tracks which dropdown has focus; only focused dropdown responds to input
3. **State Isolation**: Each dropdown manages its own expanded/collapsed state and option cursor
4. **Live Preview**: Selection changes emit messages for immediate file list update (existing behavior preserved)

### Component Interaction

```
SortDialog
  |-- Dropdown (Sort by field)
  |     |-- options: [Name, Size, Date]
  |     |-- selected: current field
  |     +-- expanded: bool
  +-- Dropdown (Order)
        |-- options: [Asc, Desc]
        |-- selected: current order
        +-- expanded: bool

Message Flow:
User Input -> SortDialog.Update()
           -> If dropdown expanded: delegate to Dropdown
           -> If dropdown closed: handle Tab/Enter/Escape
           -> Emit appropriate message (result, cancel, or config changed)
```

### State Ownership

**Dropdown owns (internal state)**:
- `expanded`: Whether the dropdown is currently open
- `highlightedIndex`: Temporary cursor position while navigating options (reset on open)

**SortDialog owns (confirmed values)**:
- `sortField`: Current confirmed sort field (Name/Size/Date)
- `sortOrder`: Current confirmed sort order (Asc/Desc)
- `focusedDropdown`: Which dropdown has focus (0 = Sort by, 1 = Order)

**State Flow on Selection**:
1. User presses Enter on highlighted option in Dropdown
2. Dropdown returns "selected" action with the highlighted value
3. SortDialog receives the action and updates its own `sortField` or `sortOrder`
4. SortDialog emits `sortDialogConfigChangedMsg` for live preview
5. Dropdown collapses (resets `highlightedIndex`)

**State Flow on Cancel**:
1. User presses Escape while Dropdown is expanded
2. Dropdown returns "cancelled" action
3. SortDialog collapses the Dropdown without updating `sortField`/`sortOrder`
4. `highlightedIndex` is reset (will re-initialize from confirmed value on next open)

## Implementation Phases

### Phase 1: Create Dropdown Component

**Goal**: Create a reusable dropdown component that handles expansion, option navigation, and selection.

**Files to Create**:
- `internal/ui/dropdown.go` - Dropdown component implementation
- `internal/ui/dropdown_test.go` - Unit tests for Dropdown

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Dropdown | Manage dropdown state (expanded, cursor, selection) | Valid options provided | Returns selected value on confirmation |
| Dropdown.HandleKey | Process key input when expanded | Dropdown is expanded | State updated, returns action (select/cancel/none) |
| Dropdown.View | Render dropdown in closed or expanded state | Valid state | Returns formatted string |

**Processing Flow**:

```
Dropdown Key Handling (Expanded State):
1. Receive key input
   |-- j/Down -> Move cursor down (if not at last)
   |-- k/Up -> Move cursor up (if not at first)
   |-- Enter -> Return "selected" action with current cursor value
   +-- Escape -> Return "cancelled" action
2. Return action and updated state
```

**DropdownAction Enum Type**:

The HandleKey method returns a DropdownAction enum to communicate the result of key handling:

```go
type DropdownAction int

const (
    DropdownActionNone      DropdownAction = iota  // No action (cursor moved or invalid key)
    DropdownActionSelected                          // User pressed Enter, option was selected
    DropdownActionCancelled                         // User pressed Escape, dropdown cancelled
)
```

This enum pattern ensures type safety and clear semantics for the three possible outcomes:
- `DropdownActionNone`: Dropdown handled the key internally (e.g., cursor movement), no further action needed by parent
- `DropdownActionSelected`: User confirmed selection, parent should update state and emit config change
- `DropdownActionCancelled`: User cancelled, parent should collapse dropdown without changing confirmed value

**Dropdown State**:
- Options list (labels and internal values)
- Current selection index
- Expanded state (bool)
- Cursor position (when expanded)

**Visual Rendering**:
- Closed: `[Value ▼]` with down arrow indicator
- Expanded: Bordered list below trigger, current option highlighted

**Symbol Conventions**:
- `▼` (U+25BC): Down-pointing triangle shown on closed dropdown to indicate it can be expanded
- `↑` (U+2191): Up arrow indicating ascending sort order (shown in Order dropdown option)
- `↓` (U+2193): Down arrow indicating descending sort order (shown in Order dropdown option)

**Implementation Steps**:

1. **Define Dropdown Structure**
   - Define struct with options, selected index, expanded flag, cursor position
   - Options should be generic (label + value pairs)

2. **Implement Key Handling**
   - Handle j/k/Down/Up for cursor movement (no cycling)
   - Handle Enter for selection
   - Handle Escape for cancel

3. **Implement View Rendering**
   - Closed state: format as `[Value ▼]`
   - Expanded state: bordered list with highlighted option
   - Use Unicode box drawing characters

**Dependencies**:
- Requires: lipgloss for styling
- Blocks: Phase 2

**Testing Approach**:

*Unit Tests*:
- Test cursor movement boundaries (first/last item)
- Test expansion and collapse
- Test selection returns correct value
- Test escape returns without changing selection

**Acceptance Criteria**:
- [ ] Dropdown renders correctly in closed state with down arrow
- [ ] Enter expands dropdown
- [ ] j/k navigate options without cycling past boundaries
- [ ] Enter selects option and collapses
- [ ] Escape collapses without selecting

**Estimated Effort**: Medium (3-5 days)

---

### Phase 2: Refactor SortDialog to Use Dropdowns

**Goal**: Replace horizontal option selection with two Dropdown components while maintaining all existing functionality.

**Files to Modify**:
- `internal/ui/sort_dialog.go`:
  - Replace focusedRow and field/order navigation with Dropdown components
  - Add focus tracking for dropdown selection
  - Update key handling for new navigation pattern
  - Update View to render dropdowns

**Files to Modify (Tests)**:
- `internal/ui/sort_dialog_test.go`:
  - Update tests to reflect new key bindings
  - Add tests for Tab/Shift+Tab navigation
  - Add tests for dropdown expansion/collapse

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| SortDialog (updated) | Manage two dropdowns and dialog-level actions | Valid SortConfig | Returns result message on confirm/cancel |
| Focus Management | Track which dropdown (or none) has focus | Dialog active | Focus state determines input routing |
| Dialog Confirmation | Handle Enter when no dropdown expanded | All dropdowns closed | Emit result message |

**Processing Flow**:

```
SortDialog Key Handling:
1. Check if any dropdown is expanded
   |-- Yes -> Delegate to expanded dropdown
   |        |-- Dropdown returns "selected" -> Close dropdown, emit config change
   |        |-- Dropdown returns "cancelled" -> Close dropdown only
   |        +-- Dropdown returns "none" -> Continue (cursor moved)
   +-- No -> Handle dialog-level keys
           |-- Enter/Space -> Expand focused dropdown
           |-- Tab -> Move focus to next dropdown
           |-- Shift+Tab -> Move focus to previous dropdown
           |-- Enter (no dropdown open) -> Confirm dialog
           +-- Escape/q -> Cancel dialog
```

**Escape vs q Key Semantics**:

The Escape and q keys have different behaviors depending on context:

| State | Escape | q |
|-------|--------|---|
| Dropdown expanded | Close dropdown (no change to confirmed value) | Close entire dialog (cancel) |
| Dropdowns closed | Cancel dialog | Cancel dialog |

Rationale:
- **Escape**: Context-aware. When a dropdown is expanded, Escape should close the dropdown only (standard UI pattern for "dismiss current overlay"). When no dropdown is expanded, Escape cancels the dialog.
- **q**: Always cancels the entire dialog immediately, regardless of dropdown state. This provides a quick exit shortcut consistent with other dialogs in duofm.

**State Transitions** (from SPEC.md):

```
Dialog States:
[*] -> SortByFocused (on dialog open)

SortByFocused:
  Enter/Space -> SortByExpanded
  Tab -> OrderFocused
  Esc/q -> [*] (cancel)
  Enter (dropdown closed) -> [*] (confirm)

SortByExpanded:
  Enter -> SortByFocused (with selection)
  Esc -> SortByFocused (no change)

OrderFocused:
  Enter/Space -> OrderExpanded
  Shift+Tab -> SortByFocused
  Esc/q -> [*] (cancel)
  Enter (dropdown closed) -> [*] (confirm)

OrderExpanded:
  Enter -> OrderFocused (with selection)
  Esc -> OrderFocused (no change)
```

**Implementation Steps**:

1. **Add Dropdown Instances to SortDialog**
   - Create Sort by dropdown with Name/Size/Date options
   - Create Order dropdown with Asc/Desc options
   - Add focusedDropdown field (0 = Sort by, 1 = Order)

2. **Update Key Handling**
   - Route keys to expanded dropdown when applicable
   - Handle Tab/Shift+Tab for focus movement
   - Handle Enter/Space for dropdown expansion
   - Handle dialog confirmation when dropdowns closed

3. **Update View Method**
   - Replace horizontal options with dropdown rendering
   - Show focused state for current dropdown
   - Update help text to reflect new key bindings

4. **Preserve Live Preview**
   - Emit sortDialogConfigChangedMsg on selection changes
   - Maintain existing message types for compatibility

**Dependencies**:
- Requires: Phase 1 (Dropdown component)
- Blocks: Phase 3

**Testing Approach**:

*Unit Tests*:
- Test Tab moves focus between dropdowns
- Test Shift+Tab moves focus in reverse
- Test Enter/Space expands focused dropdown
- Test Enter confirms when no dropdown expanded
- Test Escape cancels dialog when no dropdown expanded
- Test Escape closes dropdown when expanded (without cancelling dialog)
- Test q cancels dialog even when dropdown is expanded
- Test live preview still works

*Integration Tests*:
- Test complete workflow: open, navigate, select, confirm
- Test cancel workflow: change settings, cancel, verify revert

**Acceptance Criteria**:
- [ ] Tab/Shift+Tab navigation works between fields
- [ ] Enter/Space expands dropdown
- [ ] j/k navigate options within expanded dropdown
- [ ] Enter confirms dialog when dropdowns closed
- [ ] Escape cancels dialog or closes dropdown appropriately
- [ ] Live preview continues to work
- [ ] All existing message types preserved

**Estimated Effort**: Medium (3-5 days)

---

### Phase 3: Update E2E Tests

**Goal**: Update E2E tests to use new key bindings while maintaining full test coverage.

**Files to Modify**:
- `test/e2e/scripts/tests/sort_tests.sh`:
  - Update key sequences for dropdown workflow
  - Modify navigation tests for Tab-based navigation
  - Ensure all existing test scenarios still pass

**Key Changes**:

| Old Binding | New Binding | Affected Tests |
|-------------|-------------|----------------|
| h/l (change option) | Enter + j/k + Enter | test_sort_dialog_hl_navigation |
| j/k (move between rows) | Tab/Shift+Tab | test_sort_dialog_jk_navigation |
| l l j l Enter (Size Desc) | Enter l Enter Tab Enter j Enter Enter | test_sort_by_size_desc |

**Implementation Steps**:

1. **Update test_sort_dialog_hl_navigation**
   - Replace h/l sequence with dropdown open, navigate, select
   - Verify selection changes correctly

2. **Update test_sort_dialog_jk_navigation**
   - Replace j/k with Tab/Shift+Tab
   - Verify focus moves between dropdowns

3. **Update test_sort_by_size_desc**
   - Update key sequence for new dropdown workflow
   - Verify sort is applied correctly

4. **Update test_sort_persists_after_navigation**
   - Adjust key sequence to use dropdown selection

5. **Update test_sort_independent_panes**
   - Adjust key sequence for pane switching test

6. **Update test_sort_dialog_arrow_keys**
   - Test arrow keys work within expanded dropdown

**Testing Approach**:

*Manual Verification*:
- Run each updated test individually
- Verify assertions match new UI behavior

**Acceptance Criteria**:
- [ ] All E2E tests pass with new key bindings
- [ ] Test coverage maintained for all sort scenarios
- [ ] New dropdown interactions are properly tested

**Estimated Effort**: Small (1-2 days)

---

## Complete File Structure

```
internal/ui/
  dropdown.go           # NEW: Reusable dropdown component
  dropdown_test.go      # NEW: Dropdown unit tests
  sort_dialog.go        # MODIFIED: Use dropdowns instead of horizontal options
  sort_dialog_test.go   # MODIFIED: Update tests for new key bindings
  sort.go               # UNCHANGED: SortConfig, SortEntries
  sort_test.go          # UNCHANGED
  dialog.go             # UNCHANGED: Dialog interface
  dialog_base.go        # UNCHANGED: BaseDialog

test/e2e/scripts/tests/
  sort_tests.sh         # MODIFIED: Update key sequences
```

**File Descriptions**:
- `dropdown.go`: Generic dropdown component for selection from a list of options
- `dropdown_test.go`: Tests for dropdown expansion, navigation, selection
- `sort_dialog.go`: Refactored to use two Dropdown instances
- `sort_dialog_test.go`: Updated tests for Tab navigation and dropdown interaction
- `sort_tests.sh`: E2E tests with new key sequences

## Testing Strategy

### Unit Testing

**Approach**:
- Table-driven tests for navigation boundaries
- State transition tests for dropdown expansion/collapse
- Message emission tests for dialog result

**Test Coverage Goals**:
- Dropdown component: 90%+ coverage
- SortDialog: 85%+ coverage (UI rendering harder to test)

**Key Test Areas**:

1. **Dropdown Component** (`internal/ui/dropdown_test.go`)
   - Closed state rendering
   - Expanded state rendering with highlight
   - Cursor movement (j/k, with boundaries)
   - Selection on Enter
   - Cancel on Escape
   - Options don't cycle

2. **SortDialog** (`internal/ui/sort_dialog_test.go`)
   - Tab/Shift+Tab focus navigation
   - Enter/Space dropdown expansion
   - Dialog confirmation when dropdowns closed
   - Dialog cancellation (Escape/q)
   - Live preview on selection change
   - Original config restoration on cancel

### Integration Testing

**Scenarios**:
1. Full dialog workflow: open dropdown, select, confirm
2. Cancel workflow: open dropdown, change, cancel dialog
3. Mixed navigation: Tab between fields, select in each

### E2E Testing

Based on spec test scenarios:
- [ ] Sort dialog opens with 's' key
- [ ] Dropdown expands with Enter/Space
- [ ] j/k navigate options in expanded dropdown
- [ ] Tab moves focus between dropdowns
- [ ] Enter confirms dialog when dropdowns closed
- [ ] Escape cancels or closes dropdown appropriately
- [ ] Sort settings persist after navigation
- [ ] Sort settings independent per pane

## Dependencies

### External Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| github.com/charmbracelet/bubbletea | existing | TUI framework |
| github.com/charmbracelet/lipgloss | existing | Styling |

### Internal Dependencies

**Implementation Order**:
1. Phase 1: Dropdown component (no dependencies)
2. Phase 2: SortDialog refactor (depends on Phase 1)
3. Phase 3: E2E tests (depends on Phase 2)

**Component Dependencies**:
- `internal/ui/sort_dialog.go` depends on `internal/ui/dropdown.go`
- Tests depend on both components

## Risk Assessment

### Technical Risks

1. **Key Binding Conflicts**
   - **Risk**: Tab key might conflict with other UI elements
   - **Likelihood**: Low (dialog captures all input when active)
   - **Impact**: Low
   - **Mitigation**: Dialog has exclusive input focus when active

2. **Visual Layout Issues**
   - **Risk**: Expanded dropdown may overflow dialog bounds
   - **Likelihood**: Medium
   - **Impact**: Medium (visual glitch)
   - **Mitigation**: Test with all option combinations; adjust dialog size if needed

3. **Dialog Resize on Dropdown Expansion**
   - **Risk**: Dialog height changes when dropdown expands, causing jarring visual jumps
   - **Likelihood**: High (if not addressed)
   - **Impact**: Medium (poor UX)
   - **Mitigation**: Pre-allocate fixed height for dialog
     - Dialog should maintain constant height regardless of dropdown state
     - Reserve space for maximum dropdown expansion (3 options for Sort by)
     - Use consistent padding/margins to prevent content shift
     - Expanded dropdown options overlay within pre-allocated space

### Implementation Risks

1. **E2E Test Stability**
   - **Risk**: Timing-sensitive tests may fail with new key sequences
   - **Likelihood**: Medium
   - **Impact**: Low (tests need adjustment)
   - **Mitigation**: Increase sleep delays if needed; use explicit waits

## Performance Considerations

- Dropdown rendering is lightweight (small option lists)
- No performance concerns expected
- NFR1 (16ms rendering) should be easily met

## Security Considerations

- No security implications (UI-only change)
- No file operations affected
- No user input validation changes needed

## Open Questions

None - all requirements confirmed in specification.

## Success Criteria

- [ ] All dropdowns expand and collapse correctly
- [ ] All option selections work via j/k and Enter
- [ ] Tab/Shift+Tab navigation works between fields
- [ ] Enter confirms dialog when dropdowns closed
- [ ] Escape cancels dialog or closes dropdown appropriately
- [ ] Live preview continues to work
- [ ] All unit tests pass
- [ ] All E2E tests pass (after updates for new key bindings)
- [ ] Visual inspection confirms correct layout

## References

- **Specification**: `doc/tasks/improve-sort/SPEC.md`
- **Requirements (Japanese)**: `doc/tasks/improve-sort/要件定義書.md`
- **Dialog Best Practices**: `doc/development/DIALOG_BEST_PRACTICES.md`
- **Current Implementation**: `internal/ui/sort_dialog.go`
- **E2E Tests**: `test/e2e/scripts/tests/sort_tests.sh`

---

### Phase 4: Add j/k Navigation and OK Button

**Goal**: Add j/k navigation between major items (Sort by, Order, OK button) and an explicit OK button for confirmation. This improves consistency with other dialogs and provides Vim-style navigation.

**Files to Modify**:
- `internal/ui/sort_dialog.go`:
  - Change `focusedDropdown` to `focusedItem` (0-2)
  - Add OK button as major item 2
  - Implement j/k navigation between major items
  - Update Enter handling for OK button confirmation
  - Update Space handling to not affect OK button
  - Update View to render OK button
  - Update help text

- `internal/ui/sort_dialog_test.go`:
  - Add tests for j/k navigation between major items
  - Add tests for OK button confirmation
  - Add tests for Space key behavior on OK button
  - Update Tab/Shift+Tab tests for 3 major items

- `test/e2e/scripts/tests/sort_tests.sh`:
  - Add tests for j/k navigation between major items
  - Add tests for OK button confirmation
  - Update existing tests for 3 major items

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| focusedItem | Track focus across 3 major items (Sort by, Order, OK) | Dialog active | Focus index in range 0-2 |
| j/k Navigation | Move focus between major items when dropdowns closed | Dropdowns closed | Focus moves (no cycling) |
| OK Button | Confirm dialog with explicit action | Focus on OK button | Dialog closes, settings applied |

**FocusTarget Type Definition**:

To provide type safety and clear semantics for focus tracking, define a FocusTarget enum type:

```go
// FocusTarget represents which major item has focus in the dialog
type FocusTarget int

const (
    FocusTargetSortBy FocusTarget = iota  // 0: Sort by dropdown
    FocusTargetOrder                       // 1: Order dropdown
    FocusTargetOK                          // 2: OK button
)
```

This enum pattern ensures:
- `FocusTargetSortBy` (0): Sort by dropdown has focus
- `FocusTargetOrder` (1): Order dropdown has focus
- `FocusTargetOK` (2): OK button has focus

**Processing Flow**:

```
Major Item Navigation (Dropdowns Closed):
1. Receive j/down key
   |-- focusedItem < FocusTargetOK -> focusedItem++
   +-- focusedItem == FocusTargetOK -> no change (at last item)

2. Receive k/up key
   |-- focusedItem > FocusTargetSortBy -> focusedItem--
   +-- focusedItem == FocusTargetSortBy -> no change (at first item)

3. Receive Enter key
   |-- focusedItem == 0 -> Expand Sort by dropdown
   |-- focusedItem == 1 -> Expand Order dropdown
   +-- focusedItem == 2 -> Confirm dialog, close

4. Receive Space key
   |-- focusedItem == 0 -> Expand Sort by dropdown
   |-- focusedItem == 1 -> Expand Order dropdown
   +-- focusedItem == 2 -> No action (differentiate from dropdown)
```

**State Transitions** (Updated from SPEC.md):

```
Dialog States:
[*] -> SortByFocused (on dialog open)

SortByFocused (focusedItem = 0):
  Enter/Space -> SortByExpanded
  j/Tab -> OrderFocused
  Esc/q -> [*] (cancel)

SortByExpanded:
  Enter -> SortByFocused (with selection)
  Esc -> SortByFocused (no change)
  q -> [*] (cancel dialog)

OrderFocused (focusedItem = 1):
  Enter/Space -> OrderExpanded
  k/Shift+Tab -> SortByFocused
  j/Tab -> OKFocused
  Esc/q -> [*] (cancel)

OrderExpanded:
  Enter -> OrderFocused (with selection)
  Esc -> OrderFocused (no change)
  q -> [*] (cancel dialog)

OKFocused (focusedItem = 2):
  Enter -> [*] (confirm dialog)
  Space -> No action
  k/Shift+Tab -> OrderFocused
  Esc/q -> [*] (cancel)
```

**Visual Design** (from SPEC.md):

```
Closed State:
╭──────────────────────────────────╮
│                                  │
│   Sort                           │
│                                  │
│  Sort by    [Name ▼]             │  <- Major item 0
│                                  │
│  Order      [↑Asc ▼]             │  <- Major item 1
│                                  │
│            [OK]                  │  <- Major item 2 (NEW)
│                                  │
│  j/k:move  Enter:select  q:quit  │
│                                  │
╰──────────────────────────────────╯

OK Button Focused:
│           [ OK ]                 │  <- Highlighted
```

**Implementation Steps**:

1. **Rename focusedDropdown to focusedItem**
   - Change field name and update all references
   - Change range from 0-1 to 0-2

2. **Add j/k Navigation in handleDialogKey()**
   - j/down: increment focusedItem (if < 2)
   - k/up: decrement focusedItem (if > 0)
   - Navigation only when dropdowns are closed

3. **Update Enter Handling**
   - Check if focusedItem == 2 (OK button)
   - If OK button: close dialog and return confirmed
   - Otherwise: expand focused dropdown

4. **Update Space Handling**
   - Only expand dropdown if focusedItem < 2
   - Do nothing if focusedItem == 2 (OK button)

5. **Update Tab/Shift+Tab Handling**
   - Tab: move to next major item (0->1->2)
   - Shift+Tab: move to previous major item (2->1->0)
   - No cycling at boundaries

6. **Add OK Button Rendering in View()**
   - Add new row for OK button
   - Show focus indicator when focusedItem == 2

7. **Update Help Text**
   - Change to: "j/k:move  Enter:select  q:quit"

**Dependencies**:
- Requires: Phase 1-3 complete (Dropdown component, SortDialog refactor, E2E tests)
- Blocks: None

**Testing Approach**:

*Unit Tests*:
- j moves focus down (0->1->2), stops at 2
- k moves focus up (2->1->0), stops at 0
- Tab moves focus forward through all 3 items
- Shift+Tab moves focus backward through all 3 items
- Enter on OK button (focusedItem=2) confirms dialog
- Space on OK button does nothing
- j/k/Tab/Shift+Tab only work when dropdowns closed

*E2E Tests*:
- Full workflow using j/k navigation between major items
- OK button confirmation via Enter
- Navigation from Order to OK and back

**Acceptance Criteria**:
- [ ] j/down moves focus to next major item when dropdowns closed
- [ ] k/up moves focus to previous major item when dropdowns closed
- [ ] Tab/Shift+Tab navigation works between 3 major items
- [ ] Navigation does not cycle (stops at first/last item)
- [ ] OK button is displayed below Order dropdown
- [ ] Enter on OK button confirms dialog
- [ ] Space on OK button does nothing
- [ ] OK button has visual focus indication when focused
- [ ] Help text shows "j/k:move"
- [ ] All existing functionality preserved

**Estimated Effort**: Small (1-2 days)

**Risks and Mitigation**:
- **Risk**: OK button might overlap with dropdown expansion
  - **Mitigation**: Dialog height already accounts for dropdown expansion; OK button is below the space reserved for dropdowns

---

## Complete File Structure (Updated)

```
internal/ui/
  dropdown.go           # Reusable dropdown component (Phase 1)
  dropdown_test.go      # Dropdown unit tests (Phase 1)
  sort_dialog.go        # MODIFIED: Use dropdowns + j/k + OK button
  sort_dialog_test.go   # MODIFIED: Tests for all navigation modes
  sort.go               # UNCHANGED: SortConfig, SortEntries
  sort_test.go          # UNCHANGED
  dialog.go             # UNCHANGED: Dialog interface
  dialog_base.go        # UNCHANGED: BaseDialog

test/e2e/scripts/tests/
  sort_tests.sh         # MODIFIED: j/k navigation + OK button tests
```

## Success Criteria (Updated)

- [ ] All dropdowns expand and collapse correctly
- [ ] All option selections work via j/k and Enter
- [ ] **j/k navigation works between major items when dropdowns closed**
- [ ] Tab/Shift+Tab navigation works between **3** major items
- [ ] **Enter on OK button confirms dialog**
- [ ] **Space on OK button does nothing**
- [ ] Escape cancels dialog or closes dropdown appropriately
- [ ] q cancels dialog at any time
- [ ] Live preview continues to work
- [ ] All unit tests pass
- [ ] All E2E tests pass
- [ ] Visual inspection confirms correct layout **with OK button**

## Implementation Status

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | Create Dropdown Component | Completed |
| Phase 2 | Refactor SortDialog to Use Dropdowns | Completed |
| Phase 3 | Update E2E Tests | Completed |
| Phase 4 | Add j/k Navigation and OK Button | **Completed** |

**Note**: All phases have been implemented and verified. Phase 4 adds j/k navigation between major items (Sort by, Order, OK button) with explicit OK button confirmation.

## Implementation Complete

All implementation phases have been completed successfully:

1. **Phase 1-3**: Dropdown component, SortDialog refactor, E2E tests
2. **Phase 4**: j/k navigation + OK button (completed 2026-01-19)

### Phase 4 Key Changes
- Introduced `FocusTarget` type for type-safe focus tracking
- Renamed `focusedDropdown` to `focusedItem` (0-2 range)
- Added j/k/down/up navigation between 3 major items
- Added OK button as 3rd major item
- Enter on OK button confirms dialog
- Space on OK button does nothing
- Updated help text to "j/k:move Enter:select q:quit"
