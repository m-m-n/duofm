# Implementation Plan: Search and Filter

## Overview

Implement incremental search (`/`) and regex search (`Ctrl+F`) functionality for filtering files in the active pane. Both features use a minibuffer UI component displayed at the bottom of the active pane.

## Objectives

- Add minibuffer component for text input
- Implement incremental search with real-time filtering
- Implement regex search with pattern matching
- Manage search state transitions correctly
- Maintain filter state between searches

## Prerequisites

- Existing pane and model infrastructure
- Understanding of Bubble Tea update/view cycle
- Existing key binding system in `keys.go`

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                           Model                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────────┐ │
│  │  leftPane   │  │  rightPane  │  │      searchState         │ │
│  │  (Pane)     │  │  (Pane)     │  │  - Mode                  │ │
│  │             │  │             │  │  - Pattern               │ │
│  │  allEntries │  │  allEntries │  │  - PreviousResult        │ │
│  │  filtered   │  │  filtered   │  │  - IsActive              │ │
│  └─────────────┘  └─────────────┘  └──────────────────────────┘ │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────────┤
│  │                    Minibuffer                                 │
│  │  - prompt ("/: " or "(search): ")                            │
│  │  - input                                                      │
│  │  - cursorPos                                                  │
│  │  - visible                                                    │
│  └──────────────────────────────────────────────────────────────┤
└─────────────────────────────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: Search Types and State

**Goal**: Define search mode types and state management structures

**Files to Create/Modify**:
- `internal/ui/search.go` - NEW: Search types and filter functions

**Implementation Steps**:

1. Create `search.go` with type definitions:
   ```go
   type SearchMode int
   const (
       SearchModeNone SearchMode = iota
       SearchModeIncremental
       SearchModeRegex
   )

   type SearchState struct {
       Mode           SearchMode
       Pattern        string
       PreviousResult *SearchResult
       IsActive       bool
   }

   type SearchResult struct {
       Mode    SearchMode
       Pattern string
   }
   ```

2. Implement filter functions:
   - `isSmartCaseSensitive(pattern string) bool`
   - `filterIncremental(entries []fs.FileEntry, pattern string) []fs.FileEntry`
   - `filterRegex(entries []fs.FileEntry, pattern string) ([]fs.FileEntry, error)`

**Testing**:
- Unit tests for `isSmartCaseSensitive`
- Unit tests for `filterIncremental` with various patterns
- Unit tests for `filterRegex` with valid/invalid patterns
- Smart case behavior verification

**Estimated Effort**: Small

---

### Phase 2: Minibuffer Component

**Goal**: Create reusable minibuffer input component

**Files to Create/Modify**:
- `internal/ui/minibuffer.go` - NEW: Minibuffer component

**Implementation Steps**:

1. Create `Minibuffer` struct:
   ```go
   type Minibuffer struct {
       prompt    string
       input     string
       cursorPos int
       visible   bool
       width     int
   }
   ```

2. Implement methods:
   - `NewMinibuffer() *Minibuffer`
   - `SetPrompt(prompt string)`
   - `SetWidth(width int)`
   - `Clear()`
   - `Show()` / `Hide()`
   - `Input() string`
   - `IsVisible() bool`
   - `HandleKey(msg tea.KeyMsg) (handled bool)`
   - `View() string`

3. Key handling in `HandleKey`:
   - Normal characters: insert at cursor
   - Backspace: delete before cursor
   - Delete: delete at cursor
   - Left/Right arrows: move cursor (Ctrl+B/F alternative)
   - Ctrl+A: move to beginning
   - Ctrl+E: move to end
   - Ctrl+K: kill to end of line
   - Ctrl+U: kill to beginning of line

4. View rendering:
   - Prompt + input text
   - Visual cursor indication
   - Truncate if exceeds width

**Dependencies**:
- lipgloss for styling

**Testing**:
- Key input handling tests
- Cursor movement tests
- View rendering tests
- Edge cases (empty input, long input)

**Estimated Effort**: Medium

---

### Phase 3: Pane Filter Integration

**Goal**: Add filtering capability to Pane

**Files to Create/Modify**:
- `internal/ui/pane.go` - Add filter fields and methods

**Implementation Steps**:

1. Add fields to `Pane` struct:
   ```go
   allEntries      []fs.FileEntry  // Unfiltered entries (full list)
   // existing entries field becomes the filtered/displayed list
   ```

2. Modify `LoadDirectory()`:
   - Store entries in `allEntries`
   - Copy to `entries` (which is displayed)
   - Hidden file filtering applies to `allEntries` first

3. Add new methods:
   - `ApplyFilter(pattern string, mode SearchMode)` - Filter allEntries to entries
   - `ClearFilter()` - Reset entries to allEntries
   - `ResetToFullList()` - Reload allEntries and clear filter

4. Update entry count display:
   - Show filtered count vs total count when filter active
   - Format: "Marked 0/5 (15) 0 B" where (15) is total

**Dependencies**:
- Phase 1 (filter functions)

**Testing**:
- Filter application tests
- Filter clearing tests
- Cursor position after filter
- Integration with hidden file toggle

**Estimated Effort**: Medium

---

### Phase 4: Model Search Integration

**Goal**: Integrate search state and minibuffer into Model

**Files to Create/Modify**:
- `internal/ui/model.go` - Add search handling
- `internal/ui/keys.go` - Add search key constants

**Implementation Steps**:

1. Add key constants to `keys.go`:
   ```go
   KeySearch      = "/"
   KeyRegexSearch = "ctrl+f"
   ```

2. Add fields to `Model`:
   ```go
   searchState SearchState
   minibuffer  *Minibuffer
   ```

3. Initialize minibuffer in `NewModel()` or on first use

4. Add helper method:
   ```go
   func (m *Model) startSearch(mode SearchMode)
   ```

5. Modify `Update()` to handle search:
   - When minibuffer is visible, route keys to minibuffer first
   - Handle Enter (confirm) and Esc (cancel)
   - For incremental mode, apply filter on each keystroke
   - For regex mode, apply filter only on Enter

6. State transition handling (see SPEC.md state machine):
   - Track `PreviousResult` for Esc restoration
   - Clear filter on empty Enter
   - Switch modes correctly

**Dependencies**:
- Phase 2 (Minibuffer)
- Phase 3 (Pane filter)

**Testing**:
- Key binding activation tests
- State transition tests
- Incremental filter update tests
- Regex filter on confirm tests
- Cancel/restore behavior tests

**Estimated Effort**: Medium

---

### Phase 5: View Integration

**Goal**: Render minibuffer in pane view

**Files to Create/Modify**:
- `internal/ui/pane.go` - Modify View methods
- `internal/ui/model.go` - Pass minibuffer to pane view

**Implementation Steps**:

1. Option A: Pane renders minibuffer
   - Add method `ViewWithMinibuffer(minibuffer *Minibuffer, diskSpace uint64) string`
   - Reduce file list height by 1 when minibuffer visible
   - Render minibuffer at bottom of pane

2. Option B: Model overlays minibuffer
   - Calculate position based on active pane
   - Overlay minibuffer line at pane bottom

3. Recommended: Option A (cleaner encapsulation)
   - Modify `ViewWithDiskSpace` to accept optional minibuffer
   - Or add separate method

4. Adjust visible lines calculation:
   ```go
   visibleLines := p.height - 4  // Normal
   if minibuffer.IsVisible() {
       visibleLines--  // One less line for minibuffer
   }
   ```

5. Style minibuffer consistently with pane:
   - Same width as pane
   - Background/foreground matching pane style

**Dependencies**:
- Phase 4 (Model integration)

**Testing**:
- Visual layout tests
- Minibuffer visibility toggle
- Scroll adjustment when minibuffer appears

**Estimated Effort**: Small

---

### Phase 6: Edge Cases and Polish

**Goal**: Handle edge cases and refine user experience

**Files to Create/Modify**:
- `internal/ui/search.go` - Error handling
- `internal/ui/model.go` - Edge case handling
- `internal/ui/pane.go` - Filter indicator

**Implementation Steps**:

1. Invalid regex handling:
   - Show error in status bar (not minibuffer)
   - Keep minibuffer open for correction
   - Use existing `statusMessage` mechanism

2. Empty result handling:
   - Show "(No matches)" in file list area
   - Keep minibuffer functional

3. Filter state indicator:
   - Show current filter pattern in header or status
   - Format: `[/pattern]` or `[re/pattern]`

4. Pane switch during search:
   - Cancel current search
   - Clear minibuffer
   - Or: keep search state per-pane (future enhancement)

5. Directory change during filter:
   - Clear filter automatically
   - Reset to normal state

6. Hidden file toggle during filter:
   - Reapply filter to new entry set

**Dependencies**:
- All previous phases

**Testing**:
- Invalid regex error display
- Empty result display
- Filter indicator visibility
- Interaction with other features

**Estimated Effort**: Small

---

## File Structure

```
internal/ui/
├── keys.go              # Add KeySearch, KeyRegexSearch
├── search.go            # NEW: SearchMode, SearchState, filter functions
├── search_test.go       # NEW: Filter function tests
├── minibuffer.go        # NEW: Minibuffer component
├── minibuffer_test.go   # NEW: Minibuffer tests
├── pane.go              # Add allEntries, filter methods, view with minibuffer
├── pane_test.go         # Add filter integration tests
├── model.go             # Add searchState, minibuffer, key handling
└── model_test.go        # Add search integration tests
```

## Testing Strategy

### Unit Tests

**search_test.go**:
- `TestIsSmartCaseSensitive` - lowercase patterns vs mixed case
- `TestFilterIncremental` - substring matching, case sensitivity
- `TestFilterRegex` - valid patterns, invalid patterns, smart case

**minibuffer_test.go**:
- `TestMinibufferKeyHandling` - character input, deletion, cursor movement
- `TestMinibufferView` - rendering with various input lengths

**pane_test.go** (additions):
- `TestApplyFilter` - filter application
- `TestClearFilter` - filter removal
- `TestFilterWithHiddenToggle` - interaction with hidden files

### Integration Tests

**model_test.go** (additions):
- `TestSearchModeActivation` - `/` and `Ctrl+F` key handling
- `TestIncrementalSearchFlow` - type, see results, confirm
- `TestRegexSearchFlow` - type, confirm, see results
- `TestSearchCancel` - Esc behavior
- `TestSearchStateTransitions` - all state transitions from SPEC

### Manual Testing Checklist

- [ ] Press `/`, type pattern, see filtering in real-time
- [ ] Press `Enter` to confirm search, filter persists
- [ ] Press `Esc` to cancel, returns to previous state
- [ ] Press `Ctrl+F`, type regex, press `Enter`, see filtered results
- [ ] Invalid regex shows error, minibuffer stays open
- [ ] Empty `Enter` clears filter
- [ ] Nested search: `/` after confirmed `/`, `Esc` restores previous
- [ ] Switch search modes: `/` then `Ctrl+F`
- [ ] Large directory (1000+ files) - performance acceptable
- [ ] Unicode file names filter correctly
- [ ] Hidden file toggle during active filter

## Dependencies

### External Libraries

- `regexp` (Go standard library) - Regex pattern matching

### Internal Dependencies

1. Phase 1 must complete before Phase 3 (filter functions needed)
2. Phase 2 must complete before Phase 4 (minibuffer needed)
3. Phases 1-4 must complete before Phase 5 (view integration)
4. Phase 5 should complete before Phase 6 (polish)

```
Phase 1 (Search Types) ──┬──→ Phase 3 (Pane Filter) ──┐
                         │                             │
Phase 2 (Minibuffer) ────┴──→ Phase 4 (Model) ────────┼──→ Phase 5 (View) ──→ Phase 6 (Polish)
```

## Risk Assessment

### Technical Risks

- **Regex Performance**: Complex patterns on large directories
  - Mitigation: Add timeout or pattern complexity limit (future)

- **Cursor Position After Filter**: Filter may hide current selection
  - Mitigation: Reset cursor to 0 or find nearest match

### Implementation Risks

- **State Machine Complexity**: Many state transitions
  - Mitigation: Thorough unit tests for each transition

- **View Layout**: Minibuffer may overlap with content
  - Mitigation: Reduce visible lines when minibuffer active

## Performance Considerations

- Filter functions operate on in-memory slice (fast for typical directories)
- Incremental search recalculates filter on each keystroke
  - For 1000 files, string comparison is negligible (<1ms)
- Regex compilation on each keystroke (Ctrl+F mode does not apply until Enter)
- Consider caching compiled regex for repeated patterns (future)

## Security Considerations

- Regex patterns from user input: Use `regexp.Compile` which is safe
- No file system operations during search (read-only filter)
- Pattern injection not applicable (filtering local file names only)

## Open Questions

None - all requirements clarified in specification.

## Future Enhancements

- Search history (up/down to recall previous patterns)
- Highlight matched portion of file names
- Per-pane search state (currently global to model)
- Fuzzy matching mode
- Recursive search (search in subdirectories)

## References

- [SPEC.md](./SPEC.md) - Feature specification
- [要件定義書.md](./要件定義書.md) - Japanese requirements
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
