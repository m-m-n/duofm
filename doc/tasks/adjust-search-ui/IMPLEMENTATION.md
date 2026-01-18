# Implementation Plan: Adjust Search UI

## Overview

This feature migrates regex search (Ctrl+F) and query search (Ctrl+G) from the minibuffer to dedicated dialog components with syntax hints and history navigation. Incremental search (/) remains unchanged, using the minibuffer for real-time filtering.

## Objectives

- Improve UX by using dialogs for regex and query searches that don't require real-time feedback
- Provide syntax hints to help users learn search patterns
- Add history navigation (Up/Down keys) to recall previous search patterns
- Remove unused minibuffer code after migration to reduce complexity

## Prerequisites

### Development Environment
- Go 1.21 or later
- make (for build automation)

### Dependencies
- Bubble Tea (already in use)
- Lip Gloss (already in use)
- internal/filter package (for query validation)

### Knowledge Requirements
- Bubble Tea message/update pattern
- Existing Dialog interface and BaseDialog implementation
- TextInput component usage
- Pane filter application flow

## Architecture Overview

### Technology Stack
- **Language**: Go 1.21+
- **Framework**: Bubble Tea (TUI framework)
- **Key Libraries**:
  - bubbletea - Event loop and message handling
  - lipgloss - Terminal styling

### Design Approach

Reuse existing patterns from InputDialog:
- Embed `BaseDialog` for common dialog state
- Embed `TextInput` for text editing
- Use `DialogStyles` for consistent styling
- Return result messages for Model to process

New addition: `SearchHistory` helper component for history navigation, shared between both dialogs.

### Component Interaction

```
User presses Ctrl+F/Ctrl+G
    |
    v
handleAction() in model_update_keyboard.go
    |
    v
Create RegexSearchDialog/QuerySearchDialog
    |-- history reference passed from Model
    v
Dialog.Update() handles keypresses
    |-- Enter: validate, return result message
    |-- Esc: return cancelled message
    |-- Up/Down: navigate history
    |-- Other: delegate to TextInput
    v
Model.Update() receives result message
    |-- Apply filter to pane OR clear filter
    v
Dialog closed, pane displays filtered results
```

## Implementation Phases

### Phase 1: SearchHistory Component

**Goal**: Create a reusable history navigation component that both dialogs can use.

**Files to Create**:
- `internal/ui/search_history.go` - SearchHistory component
- `internal/ui/search_history_test.go` - Unit tests

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| SearchHistory | Manage history entries and navigation state | History list initialized | Returns appropriate pattern for navigation |
| NewSearchHistory | Create new history with max size | maxSize > 0 | Empty history ready for use |
| Add | Add pattern to history (deduplicate) | Valid pattern | Pattern at front of history |
| NavigateUp | Move to older entry | History has entries | Returns older pattern |
| NavigateDown | Move to newer entry | In navigation mode | Returns newer pattern or original input |
| Reset | Reset navigation state | - | Ready for new navigation session |

**Behavior Contract**:

1. **Add(pattern)**
   - Precondition: pattern is non-empty string
   - Postcondition:
     - Pattern is at index 0 of patterns slice
     - If pattern existed before, old occurrence is removed (no duplicates)
     - If len(patterns) > maxSize, oldest entry is truncated
   - Side effects: None

2. **NavigateUp(currentInput)**
   - Precondition: History instance exists
   - Postcondition:
     - If first call (index == -1), currentInput is saved to editBuf
     - index is incremented (capped at len(patterns)-1)
     - Returns pattern at new index, or empty string if no history
   - Side effects: Updates internal index and editBuf state

3. **NavigateDown()**
   - Precondition: History instance exists
   - Postcondition:
     - index is decremented (capped at -1)
     - If index == -1, returns editBuf (original input)
     - Else returns pattern at current index
   - Side effects: Updates internal index state

4. **Reset()**
   - Precondition: None
   - Postcondition:
     - index is set to -1
     - editBuf is cleared
     - patterns slice is unchanged
   - Side effects: None

**Processing Flow**:
```
NavigateUp:
1. If first navigation, save current input to editBuf
2. If not at oldest entry, increment index
3. Return pattern at current index

NavigateDown:
1. If not at input position, decrement index
2. If at input position (index = -1), return editBuf
3. Else return pattern at current index
```

**Implementation Steps**:

1. **Define SearchHistory struct**
   - patterns slice for history entries (newest at index 0)
   - index for current navigation position (-1 = at input)
   - editBuf for preserving original input during navigation
   - maxSize for limiting history length

2. **Implement history management**
   - Add: deduplicate, prepend, truncate to maxSize
   - Reset: clear navigation state for new dialog session

3. **Implement navigation**
   - NavigateUp/Down with boundary checking
   - Preserve user input when entering navigation mode

**Dependencies**:
- Requires: None
- Blocks: Phase 2 (RegexSearchDialog), Phase 3 (QuerySearchDialog)

**Testing Approach**:

*Unit Tests*:
- Add pattern to empty history
- Add duplicate pattern moves to front
- NavigateUp returns patterns in order
- NavigateDown returns to original input
- NavigateUp at end stays at last entry
- NavigateDown at beginning stays at original
- Reset clears navigation state
- History respects maxSize limit

**Acceptance Criteria**:
- [ ] SearchHistory can be created with configurable max size
- [ ] Patterns can be added with deduplication
- [ ] Navigation works correctly with boundary conditions
- [ ] Original input is preserved during navigation

**Estimated Effort**: Small (1-2 days)

---

### Phase 2: RegexSearchDialog Component

**Goal**: Create a dialog for regex search with syntax hints and history navigation.

**Files to Create**:
- `internal/ui/regex_search_dialog.go` - Dialog component
- `internal/ui/regex_search_dialog_test.go` - Unit tests

**Files to Modify**:
- `internal/ui/messages.go`:
  - Add `regexSearchResultMsg` type

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| RegexSearchDialog | Display regex search dialog with hints | History reference provided | Returns result message on close |
| NewRegexSearchDialog | Create dialog instance | Valid history | Active dialog with reset history |
| Update | Handle key events | Dialog is active | Updated state or result message |
| View | Render dialog content | Dialog is active | Styled dialog string |
| regexSearchResultMsg | Carry search result | - | Contains pattern or cancelled flag |

**Processing Flow**:
```
Dialog Update Flow:
1. Receive key message
   |-- Enter pressed
   |   |-- Empty input -> return result with empty pattern (clear filter)
   |   |-- Non-empty input -> validate regex
   |       |-- Valid -> add to history, return result with pattern
   |       |-- Invalid -> set errorMsg, stay open
   |
   |-- Esc pressed -> return result with cancelled=true
   |
   |-- Up pressed -> history.NavigateUp(), update textInput
   |
   |-- Down pressed -> history.NavigateDown(), update textInput
   |
   |-- Other key -> delegate to textInput.HandleKey()

2. Clear errorMsg on any key press
```

**View Layout**:
```
+---------------------------------------------+
|  Regex Search                               |  <- Title (bold, cyan)
|                                             |
|  +---------------------------------------+  |  <- Input field with border
|  | pattern input with cursor             |  |
|  +---------------------------------------+  |
|                                             |
|  [Error message if any - red]               |  <- Error (conditional)
|                                             |
|  Examples: ^prefix  suffix$  \.txt$         |  <- Hint (muted gray)
|                                             |
|  Enter: Search  Esc: Cancel  Up/Down: History|  <- Footer (muted)
+---------------------------------------------+
```

**Implementation Steps**:

1. **Define RegexSearchDialog struct**
   - Embed BaseDialog for common state
   - TextInput for text editing
   - SearchHistory reference (shared with Model)
   - errorMsg for validation errors
   - DialogStyles for rendering

2. **Implement constructor**
   - Create BaseDialog with DialogDisplayPane type
   - Create TextInput
   - Reset history navigation state

3. **Implement Update method**
   - Handle Enter, Esc, Up, Down, and text input
   - Validate regex pattern before confirming
   - Clear error on any key press

4. **Implement View method**
   - Render title, input, error (if any), hints, footer
   - Follow existing InputDialog layout pattern

**Smart Case Matching**:

Regex search uses smart case matching (implemented in `search.go`):
- If pattern contains any uppercase letters -> case-sensitive search
- If pattern is all lowercase -> case-insensitive search (adds `(?i)` prefix)

This is handled by `isSmartCaseSensitive()` and `filterRegex()` in `search.go`:
```go
// isSmartCaseSensitive returns true if pattern contains uppercase letters
func isSmartCaseSensitive(pattern string) bool {
    return pattern != strings.ToLower(pattern)
}

// In filterRegex: add (?i) prefix for case-insensitive matching
if !isSmartCaseSensitive(pattern) {
    regexPattern = "(?i)" + pattern
}
```

No changes needed in dialog - smart case is applied at filter level.

**Dependencies**:
- Requires: Phase 1 (SearchHistory)
- Blocks: Phase 4 (Integration)

**Testing Approach**:

*Unit Tests*:
- New dialog is active and shows correct title
- Enter with valid regex returns success message
- Enter with invalid regex shows error, stays open
- Enter with empty input returns empty pattern
- Esc returns cancelled message
- Up/Down updates input from history
- Text input handles regular characters
- DisplayType is DialogDisplayPane

**Acceptance Criteria**:
- [ ] Dialog displays with correct title "Regex Search"
- [ ] Syntax hints are visible
- [ ] Valid regex pattern returns success result
- [ ] Invalid regex shows inline error
- [ ] Empty input returns empty pattern (for clearing filter)
- [ ] Esc cancels without changing filter
- [ ] History navigation works with Up/Down

**Estimated Effort**: Small (1-2 days)

---

### Phase 3: QuerySearchDialog Component

**Goal**: Create a dialog for query search with syntax hints and history navigation.

**Files to Create**:
- `internal/ui/query_search_dialog.go` - Dialog component
- `internal/ui/query_search_dialog_test.go` - Unit tests

**Files to Modify**:
- `internal/ui/messages.go`:
  - Add `querySearchResultMsg` type

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| QuerySearchDialog | Display query search dialog with hints | History reference provided | Returns result message on close |
| NewQuerySearchDialog | Create dialog instance | Valid history | Active dialog with reset history |
| Update | Handle key events | Dialog is active | Updated state or result message |
| View | Render dialog content | Dialog is active | Styled dialog string |
| querySearchResultMsg | Carry search result | - | Contains query or cancelled flag |

**Processing Flow**:
```
Dialog Update Flow (same structure as RegexSearchDialog):
1. Receive key message
   |-- Enter pressed
   |   |-- Empty input -> return result with empty query (clear filter)
   |   |-- Non-empty input -> validate query using filter.ValidateQuery()
   |       |-- Valid -> add to history, return result with query
   |       |-- Invalid -> set errorMsg, stay open
   |
   |-- Esc pressed -> return result with cancelled=true
   |
   |-- Up/Down pressed -> navigate history, update textInput
   |
   |-- Other key -> delegate to textInput.HandleKey()
```

**View Layout**:
```
+---------------------------------------------+
|  Query Filter                               |  <- Title (bold, cyan)
|                                             |
|  +---------------------------------------+  |  <- Input field with border
|  | query input with cursor               |  |
|  +---------------------------------------+  |
|                                             |
|  [Error message if any - red]               |  <- Error (conditional)
|                                             |
|  Examples: size > 1MB  ext = ".go"          |  <- Hint line 1 (muted gray)
|            name LIKE "test%"                |  <- Hint line 2 (muted gray)
|                                             |
|  Enter: Filter  Esc: Cancel  Up/Down: History|  <- Footer (muted)
+---------------------------------------------+
```

**Implementation Steps**:

1. **Define QuerySearchDialog struct**
   - Same structure as RegexSearchDialog
   - Uses filter.ValidateQuery() for validation instead of regexp.Compile()

2. **Implement constructor, Update, View**
   - Follow same pattern as RegexSearchDialog
   - Different title, hints, and footer text

**Dependencies**:
- Requires: Phase 1 (SearchHistory)
- Blocks: Phase 4 (Integration)

**Testing Approach**:

*Unit Tests*:
- New dialog is active and shows correct title
- Enter with valid query returns success message
- Enter with invalid query shows error, stays open
- Enter with empty input returns empty pattern
- Esc returns cancelled message
- Up/Down updates input from history
- Text input handles regular characters
- DisplayType is DialogDisplayPane

**Acceptance Criteria**:
- [ ] Dialog displays with correct title "Query Filter"
- [ ] Syntax hints are visible (2 lines)
- [ ] Valid query pattern returns success result
- [ ] Invalid query shows inline error
- [ ] Empty input returns empty query (for clearing filter)
- [ ] Esc cancels without changing filter
- [ ] History navigation works with Up/Down

**Estimated Effort**: Small (1-2 days)

---

### Phase 4: Integration with Model

**Goal**: Wire up the new dialogs to Model and remove unused minibuffer code.

**Files to Modify**:
- `internal/ui/model.go`:
  - Add `regexHistory *SearchHistory` field
  - Add `queryHistory *SearchHistory` field
  - Initialize histories in `NewModelWithConfig()`

- `internal/ui/model_update_keyboard.go`:
  - Modify `ActionRegexSearch` case to open RegexSearchDialog
  - Modify `ActionSQLFilter` case to open QuerySearchDialog

- `internal/ui/model_update.go`:
  - Add handler for `regexSearchResultMsg`
  - Add handler for `querySearchResultMsg`

- `internal/ui/model.go`:
  - Simplify `startSearch()` to only handle incremental mode

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| Model.regexHistory | Store regex search history | Model initialized | History persists across dialog opens |
| Model.queryHistory | Store query search history | Model initialized | History persists across dialog opens |
| handleAction (ActionRegexSearch) | Open RegexSearchDialog | Model ready | Dialog displayed with history |
| handleAction (ActionSQLFilter) | Open QuerySearchDialog | Model ready | Dialog displayed with history |
| regexSearchResultMsg handler | Apply regex filter to pane | Result received | Pane filtered or filter cleared |
| querySearchResultMsg handler | Apply query filter to pane | Result received | Pane filtered or filter cleared |

**Processing Flow**:
```
ActionRegexSearch:
1. Create RegexSearchDialog with m.regexHistory
2. Set m.dialog = created dialog
3. Return (dialog will handle input)

ActionSQLFilter:
1. Create QuerySearchDialog with m.queryHistory
2. Set m.dialog = created dialog
3. Return (dialog will handle input)

regexSearchResultMsg received:
1. If cancelled -> do nothing
2. If pattern == "" -> pane.ClearFilter()
3. If pattern != "" -> pane.ApplyFilter(pattern, SearchModeRegex)
   |-- If error -> show in status message
   Note: History is already added in dialog (on valid Enter, before returning result)

querySearchResultMsg received:
1. If cancelled -> do nothing
2. If query == "" -> pane.ClearFilter()
3. If query != "" -> pane.ApplyFilter(query, SearchModeSQLLike)
   |-- If error -> show in status message
   Note: History is already added in dialog (on valid Enter, before returning result)

History Addition Timing (in dialogs):
- Pattern/query is added to history ONLY when:
  1. User presses Enter
  2. Input is non-empty
  3. Validation passes (valid regex or valid query)
- History is NOT added when:
  - User presses Esc (cancel)
  - Input is empty (clear filter)
  - Validation fails (invalid pattern)
```

**Implementation Steps**:

1. **Add history fields to Model**
   - Define regexHistory and queryHistory fields
   - Initialize in NewModelWithConfig with size 50

2. **Update handleAction for search actions**
   - ActionRegexSearch: create and set RegexSearchDialog
   - ActionSQLFilter: create and set QuerySearchDialog
   - Remove minibuffer-based search initiation

3. **Add result message handlers**
   - Handle regexSearchResultMsg in handleDialogMessages
   - Handle querySearchResultMsg in handleDialogMessages
   - Apply filter or show error on invalid pattern

4. **Clean up startSearch()**
   - Keep only SearchModeIncremental case
   - Remove SearchModeRegex and SearchModeSQLLike cases

5. **Clean up handleSearchInput()**
   - Verify only incremental search uses this path
   - No changes needed if only incremental uses minibuffer

**Dependencies**:
- Requires: Phase 2 (RegexSearchDialog), Phase 3 (QuerySearchDialog)
- Blocks: None

**Testing Approach**:

*Integration Tests*:
- Ctrl+F opens RegexSearchDialog
- Ctrl+G opens QuerySearchDialog
- Dialog result applies filter to pane
- Empty result clears filter
- Cancelled result doesn't change filter

*Regression Tests*:
- `/` still opens minibuffer for incremental search
- Incremental search filters in real-time
- Incremental search Enter confirms filter
- Incremental search Esc cancels

**Acceptance Criteria**:
- [ ] Ctrl+F opens RegexSearchDialog
- [ ] Ctrl+G opens QuerySearchDialog
- [ ] Regex search applies filter correctly
- [ ] Query search applies filter correctly
- [ ] Empty input clears filter
- [ ] Cancelled dialog preserves previous filter
- [ ] Incremental search (/) works unchanged
- [ ] No unused minibuffer code for regex/query search

**Estimated Effort**: Medium (2-3 days)

---

## Complete File Structure

```
internal/ui/
|-- search_history.go           # New: SearchHistory component
|-- search_history_test.go      # New: Tests
|-- regex_search_dialog.go      # New: RegexSearchDialog component
|-- regex_search_dialog_test.go # New: Tests
|-- query_search_dialog.go      # New: QuerySearchDialog component
|-- query_search_dialog_test.go # New: Tests
|-- messages.go                 # Modified: Add result message types
|-- model.go                    # Modified: Add history fields, simplify startSearch
|-- model_update_keyboard.go    # Modified: Open dialogs instead of minibuffer
|-- model_update.go             # Modified: Handle result messages
|-- search.go                   # Unchanged: Keep filter functions
|-- minibuffer.go               # Unchanged: Still used for incremental search
|-- dialog_base.go              # Unchanged: Reused
|-- text_input.go               # Unchanged: Reused
```

**File Descriptions**:
- `search_history.go`: Reusable history navigation for both dialogs
- `regex_search_dialog.go`: Regex search dialog with validation and hints
- `query_search_dialog.go`: Query search dialog with validation and hints
- `messages.go`: Result message types for dialog communication
- `model.go`: History storage and simplified search initiation
- `model_update_keyboard.go`: Action handlers for opening dialogs
- `model_update.go`: Result message handlers for applying filters

## Testing Strategy

### Unit Testing

**Approach**:
- Use Go's built-in `testing` package
- Table-driven tests for multiple scenarios
- Test each component in isolation

**Test Coverage Goals**:
- Core logic: 80%+ coverage
- SearchHistory: 90%+ (critical for UX)
- Dialog components: 80%+ (key flows)

**Key Test Areas**:

1. **SearchHistory** (`internal/ui/search_history_test.go`)
   - Add patterns (empty, duplicate, overflow)
   - Navigation (boundaries, preserve input)
   - Reset behavior

2. **RegexSearchDialog** (`internal/ui/regex_search_dialog_test.go`)
   - Valid/invalid regex handling
   - Empty input handling
   - History navigation
   - Key delegation to TextInput

3. **QuerySearchDialog** (`internal/ui/query_search_dialog_test.go`)
   - Valid/invalid query handling
   - Empty input handling
   - History navigation
   - Key delegation to TextInput

### Integration Testing

**Scenarios**:
1. Ctrl+F opens dialog, enter valid regex, filter applies
2. Ctrl+G opens dialog, enter valid query, filter applies
3. Empty input clears active filter
4. Esc preserves previous filter state
5. Invalid patterns show inline errors

### Regression Testing

**Scenarios**:
1. `/` opens minibuffer for incremental search
2. Incremental search updates in real-time
3. Enter confirms incremental filter
4. Esc cancels incremental search
5. Pane switching cancels active search

### Manual Testing Checklist

Based on spec test scenarios:
- [ ] Ctrl+F shows RegexSearchDialog with hints
- [ ] Enter valid regex -> filter applies
- [ ] Enter invalid regex -> error shown in dialog
- [ ] Empty input + Enter -> filter clears
- [ ] Esc -> dialog closes, filter unchanged
- [ ] Up/Down -> navigates through history
- [ ] History entry selected and executed
- [ ] Ctrl+G shows QuerySearchDialog with hints
- [ ] Query filter works correctly
- [ ] Incremental search (/) works as before

## Dependencies

### External Dependencies

| Package | Version | Purpose | Installation |
|---------|---------|---------|--------------|
| github.com/charmbracelet/bubbletea | existing | TUI framework | already installed |
| github.com/charmbracelet/lipgloss | existing | Styling | already installed |

### Internal Dependencies

**Implementation Order** (respecting dependencies):
1. Phase 1: SearchHistory (no dependencies)
2. Phase 2: RegexSearchDialog (depends on Phase 1)
3. Phase 3: QuerySearchDialog (depends on Phase 1)
4. Phase 4: Integration (depends on Phases 2 & 3)

**Component Dependencies**:
- `RegexSearchDialog` depends on `SearchHistory`, `BaseDialog`, `TextInput`, `DialogStyles`
- `QuerySearchDialog` depends on `SearchHistory`, `BaseDialog`, `TextInput`, `DialogStyles`, `filter.ValidateQuery`
- `Model` history fields depend on `SearchHistory`
- Message handlers depend on `Pane.ApplyFilter`, `Pane.ClearFilter`

## Risk Assessment

### Technical Risks

1. **History State Management**
   - **Risk**: History state corruption if dialog closed unexpectedly
   - **Likelihood**: Low
   - **Impact**: Medium (user loses history)
   - **Mitigation**: Always call Reset() when creating new dialog

2. **Minibuffer Code Removal**
   - **Risk**: Breaking incremental search while cleaning up
   - **Likelihood**: Medium
   - **Impact**: High (core feature broken)
   - **Mitigation**:
     - Keep startSearch() for incremental mode
     - Thorough regression testing
     - Incremental cleanup with tests at each step

### Implementation Risks

1. **Dialog Styling Consistency**
   - **Risk**: New dialogs look different from existing ones
   - **Likelihood**: Low
   - **Impact**: Low (cosmetic)
   - **Mitigation**: Reuse DialogStyles, follow InputDialog pattern

## Performance Considerations

1. **History Operations**
   - Add operation is O(n) for deduplication
   - Acceptable for history size <= 50

2. **Regex Compilation**
   - Compile on Enter only, not on every keystroke
   - No performance concern

3. **Query Validation**
   - Uses existing filter.ValidateQuery()
   - Already optimized

## Security Considerations

1. **Regex Pattern Validation**
   - Patterns are validated using regexp.Compile before use
   - No risk of regex injection (patterns match filenames only)

2. **Query Validation**
   - Queries are validated using filter.ValidateQuery before use
   - SQL-like syntax is sandboxed (no actual SQL execution)

3. **No File System Access**
   - Dialogs only handle user input
   - File operations delegated to existing Pane methods

## Open Questions

### From Specification:
- None (all questions confirmed in requirements document)

### Implementation-Specific:
- None identified

## Future Enhancements

Items not in current scope:

### Phase 2 Features (from spec):
- History persistence across sessions (currently memory-only)
- More syntax hint examples
- Regex syntax highlighting in input

### Not in Current Spec:
- Fuzzy search mode
- Search result highlighting
- Search and replace functionality

## Success Metrics

### Functional Completeness
- [ ] All functional requirements (FR1-FR9) implemented
- [ ] All user stories (US1-US3) satisfied
- [ ] All test scenarios pass

### Quality Metrics
- [ ] Test coverage meets goals (80%+ core logic)
- [ ] No critical bugs in manual testing
- [ ] Code follows existing patterns

### User Experience
- [ ] Syntax hints are helpful
- [ ] History navigation is intuitive
- [ ] Error messages are clear
- [ ] Dialog styling is consistent

## References

- **Specification**: `doc/tasks/adjust-search-ui/SPEC.md`
- **Requirements**: `doc/tasks/adjust-search-ui/要件定義書.md`
- **InputDialog Reference**: `internal/ui/input_dialog.go`
- **BaseDialog Reference**: `internal/ui/dialog_base.go`
- **TextInput Reference**: `internal/ui/text_input.go`
- **Current Minibuffer**: `internal/ui/minibuffer.go`
- **Current Search**: `internal/ui/search.go`
- **Query Filter**: `internal/filter/filter.go`

## Next Steps

After reviewing this implementation plan:

1. **Review and Approval**
   - Review implementation approach
   - Confirm phase breakdown
   - Verify no missing requirements

2. **Begin Implementation**
   - Start with Phase 1 (SearchHistory)
   - Write tests before implementation (TDD)
   - Commit after each phase

3. **Verification**
   - Run `/sdd.5-check` to verify implementation matches plan
   - Run `/sdd.6-verify` to verify all requirements met
