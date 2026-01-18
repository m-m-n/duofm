# Feature: Adjust Search UI

## Overview

This feature adjusts the search functionality UI in duofm by migrating regex search (Ctrl+F) and query search (Ctrl+G) from the minibuffer to dedicated dialog components. Incremental search (/) remains unchanged, using the minibuffer for real-time filtering. The new dialogs include syntax hints and history navigation features.

## Objectives

- Improve UX by using dialogs for regex and query searches that don't require real-time feedback
- Provide syntax hints to help users learn search patterns
- Add history navigation (Up/Down keys) to recall previous search patterns
- Remove unused minibuffer code after migration to reduce complexity

## User Stories

### US1: Regex Search via Dialog
As a user, I want to use a dialog for regex search so that I can see syntax hints and access my search history.

**Acceptance Criteria:**
- [ ] Ctrl+F opens RegexSearchDialog instead of minibuffer
- [ ] Dialog displays syntax hints (e.g., `^prefix`, `suffix$`, `\.txt$`)
- [ ] Up/Down keys navigate through search history
- [ ] Enter confirms search, Esc cancels
- [ ] Empty input clears the filter
- [ ] Invalid regex shows error message in dialog

### US2: Query Search via Dialog
As a user, I want to use a dialog for query search so that I can see syntax hints and access my query history.

**Acceptance Criteria:**
- [ ] Ctrl+G opens QuerySearchDialog instead of minibuffer
- [ ] Dialog displays syntax hints (e.g., `size > 1MB`, `ext = ".go"`)
- [ ] Up/Down keys navigate through search history
- [ ] Enter confirms search, Esc cancels
- [ ] Empty input clears the filter
- [ ] Invalid query shows error message in dialog

### US3: Incremental Search Unchanged
As a user, I want incremental search (/) to continue working as before.

**Acceptance Criteria:**
- [ ] `/` key still opens minibuffer for incremental search
- [ ] Real-time filtering works as before
- [ ] No regression in existing functionality

## Technical Requirements

### Functional Requirements
- **FR1:** Create `RegexSearchDialog` component with title "Regex Search"
- **FR2:** Create `QuerySearchDialog` component with title "Query Filter"
- **FR3:** Both dialogs display in pane center (DialogDisplayPane type)
- **FR4:** Both dialogs show syntax hint examples
- **FR5:** Both dialogs support history navigation via Up/Down keys
- **FR6:** Both dialogs show validation errors inline
- **FR7:** Empty input + Enter clears the active filter
- **FR8:** Remove SearchModeRegex/SearchModeSQLLike handling from minibuffer code
- **FR9:** Update keybinding handlers to open dialogs instead of minibuffer

### Non-Functional Requirements
- **NFR1 - Performance:** Dialog display must be instant (no perceptible delay)
- **NFR2 - Maintainability:** Reuse existing BaseDialog, DialogStyles, and TextInput components
- **NFR3 - Consistency:** Follow existing dialog styling patterns (InputDialog reference)

## Implementation Approach

### Architecture

**Component Structure:**
```
internal/ui/
├── regex_search_dialog.go      # New: RegexSearchDialog component
├── regex_search_dialog_test.go # New: Tests for RegexSearchDialog
├── query_search_dialog.go      # New: QuerySearchDialog component
├── query_search_dialog_test.go # New: Tests for QuerySearchDialog
├── search_history.go           # New: SearchHistory helper component
├── search_history_test.go      # New: Tests for SearchHistory
├── model.go                    # Modified: Add dialog fields, update search handling
├── model_update_keyboard.go    # Modified: Update keybinding handlers
├── model_update.go             # Modified: Handle new dialog result messages
└── search.go                   # Modified: Remove unused code
```

**Reused Components:**
- `BaseDialog` - Common dialog state and behavior
- `DialogStyles` - Standard dialog styling
- `TextInput` - Text input with cursor management

### Component Design

#### SearchHistory Component

A helper component for managing search history that can be embedded in dialog structs.

```go
// SearchHistory manages a list of past search patterns with navigation.
type SearchHistory struct {
    patterns  []string // History entries (newest at index 0)
    index     int      // Current position (-1 = at input, 0+ = in history)
    editBuf   string   // Original input before navigation started
    maxSize   int      // Maximum number of entries to keep
}

// NewSearchHistory creates a new SearchHistory with the given max size.
func NewSearchHistory(maxSize int) *SearchHistory

// Add adds a pattern to history (moves to front if exists).
func (h *SearchHistory) Add(pattern string)

// NavigateUp moves to an older entry, returns the pattern to display.
func (h *SearchHistory) NavigateUp(currentInput string) string

// NavigateDown moves to a newer entry, returns the pattern to display.
func (h *SearchHistory) NavigateDown() string

// Reset resets navigation state (call when dialog opens).
func (h *SearchHistory) Reset()
```

#### RegexSearchDialog Component

```go
type RegexSearchDialog struct {
    BaseDialog
    textInput *TextInput
    history   *SearchHistory
    errorMsg  string
    styles    DialogStyles
}

// Result message
type regexSearchResultMsg struct {
    pattern   string
    cancelled bool
}
```

**View Layout:**
```
╭─────────────────────────────────────────╮
│  Regex Search                           │  <- Title (bold, cyan)
│                                         │
│  ┌─────────────────────────────────┐    │  <- Input field with border
│  │ pattern input with cursor       │    │
│  └─────────────────────────────────┘    │
│                                         │
│  [Error message if any - red]           │  <- Error (conditional)
│                                         │
│  Examples: ^prefix  suffix$  \.txt$     │  <- Hint (muted gray)
│                                         │
│  Enter: Search  Esc: Cancel  ↑↓: History│  <- Footer (muted)
╰─────────────────────────────────────────╯
```

#### QuerySearchDialog Component

```go
type QuerySearchDialog struct {
    BaseDialog
    textInput *TextInput
    history   *SearchHistory
    errorMsg  string
    styles    DialogStyles
}

// Result message
type querySearchResultMsg struct {
    query     string
    cancelled bool
}
```

**View Layout:**
```
╭─────────────────────────────────────────╮
│  Query Filter                           │  <- Title (bold, cyan)
│                                         │
│  ┌─────────────────────────────────┐    │  <- Input field with border
│  │ query input with cursor         │    │
│  └─────────────────────────────────┘    │
│                                         │
│  [Error message if any - red]           │  <- Error (conditional)
│                                         │
│  Examples: size > 1MB  ext = ".go"      │  <- Hint line 1 (muted gray)
│            name LIKE "test%"            │  <- Hint line 2 (muted gray)
│                                         │
│  Enter: Filter  Esc: Cancel  ↑↓: History│  <- Footer (muted)
╰─────────────────────────────────────────╯
```

### Data Flow

```
User presses Ctrl+F
    │
    ▼
handleAction(ActionRegexSearch)
    │
    ▼
Create and show RegexSearchDialog
    │
    ▼
Dialog.Update() handles keypresses
    │
    ├── Enter → validate regex
    │   ├── valid → return regexSearchResultMsg{pattern: "..."}
    │   └── invalid → set errorMsg, stay open
    │
    ├── Esc → return regexSearchResultMsg{cancelled: true}
    │
    ├── Up/Down → history.NavigateUp/Down()
    │   └── update textInput.Value
    │
    └── Other keys → textInput.HandleKey()

Model.Update() receives regexSearchResultMsg
    │
    ├── cancelled → do nothing
    │
    ├── pattern == "" → pane.ClearFilter()
    │
    └── pattern != "" → pane.ApplyFilter(pattern, SearchModeRegex)
```

### Key Handling in Dialog

```go
func (d *RegexSearchDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        d.errorMsg = "" // Clear error on any key

        switch msg.Type {
        case tea.KeyEnter:
            pattern := d.textInput.Value
            if pattern == "" {
                d.Close()
                return d, func() tea.Msg {
                    return regexSearchResultMsg{pattern: ""}
                }
            }
            // Validate regex
            if _, err := regexp.Compile(pattern); err != nil {
                d.errorMsg = "Invalid regex: " + err.Error()
                return d, nil
            }
            d.history.Add(pattern)
            d.Close()
            return d, func() tea.Msg {
                return regexSearchResultMsg{pattern: pattern}
            }

        case tea.KeyEsc:
            d.Close()
            return d, func() tea.Msg {
                return regexSearchResultMsg{cancelled: true}
            }

        case tea.KeyUp:
            newValue := d.history.NavigateUp(d.textInput.Value)
            d.textInput.SetValue(newValue)
            d.textInput.MoveCursorToEnd()
            return d, nil

        case tea.KeyDown:
            newValue := d.history.NavigateDown()
            d.textInput.SetValue(newValue)
            d.textInput.MoveCursorToEnd()
            return d, nil

        default:
            d.textInput.HandleKey(msg)
        }
    }
    return d, nil
}
```

### Files to Modify

#### internal/ui/model.go

Add dialog state:
```go
type Model struct {
    // ... existing fields ...

    // Search dialog history (shared across sessions)
    regexHistory *SearchHistory
    queryHistory *SearchHistory
}
```

Initialize in `NewModelWithConfig`:
```go
return Model{
    // ... existing fields ...
    regexHistory: NewSearchHistory(50),
    queryHistory: NewSearchHistory(50),
}
```

Remove from `startSearch()`:
- Case handling for `SearchModeRegex`
- Case handling for `SearchModeSQLLike`

#### internal/ui/model_update_keyboard.go

Update `handleAction()`:
```go
case ActionRegexSearch:
    m.dialog = NewRegexSearchDialog(m.regexHistory)
    return m, nil

case ActionSQLFilter:
    m.dialog = NewQuerySearchDialog(m.queryHistory)
    return m, nil
```

#### internal/ui/model_update.go

Add message handlers:
```go
case regexSearchResultMsg:
    if msg.cancelled {
        return m, nil
    }
    pane := m.getActivePane()
    if msg.pattern == "" {
        pane.ClearFilter()
    } else {
        if err := pane.ApplyFilter(msg.pattern, SearchModeRegex); err != nil {
            m.statusMessage = fmt.Sprintf("Regex error: %v", err)
            m.isStatusError = true
        }
    }
    return m, nil

case querySearchResultMsg:
    if msg.cancelled {
        return m, nil
    }
    pane := m.getActivePane()
    if msg.query == "" {
        pane.ClearFilter()
    } else {
        if err := pane.ApplyFilter(msg.query, SearchModeSQLLike); err != nil {
            m.statusMessage = fmt.Sprintf("Query error: %v", err)
            m.isStatusError = true
        }
    }
    return m, nil
```

#### internal/ui/search.go

Keep:
- `SearchModeNone`
- `SearchModeIncremental`
- `SearchModeRegex`
- `SearchModeSQLLike`
- `filterIncremental()`
- `filterRegex()`
- `filterSQLLike()`

Remove or simplify:
- `SearchState` struct (may simplify - only needed for incremental now)

### Files to Create

#### internal/ui/search_history.go

```go
package ui

// SearchHistory manages search pattern history with navigation.
type SearchHistory struct {
    patterns []string
    index    int
    editBuf  string
    maxSize  int
}

const DefaultSearchHistorySize = 50

func NewSearchHistory(maxSize int) *SearchHistory {
    return &SearchHistory{
        patterns: make([]string, 0),
        index:    -1,
        editBuf:  "",
        maxSize:  maxSize,
    }
}

func (h *SearchHistory) Add(pattern string) {
    if pattern == "" {
        return
    }
    // Remove existing occurrence
    for i, p := range h.patterns {
        if p == pattern {
            h.patterns = append(h.patterns[:i], h.patterns[i+1:]...)
            break
        }
    }
    // Add to front
    h.patterns = append([]string{pattern}, h.patterns...)
    // Trim to max size
    if len(h.patterns) > h.maxSize {
        h.patterns = h.patterns[:h.maxSize]
    }
}

func (h *SearchHistory) NavigateUp(currentInput string) string {
    if len(h.patterns) == 0 {
        return currentInput
    }
    // First navigation - save current input
    if h.index == -1 {
        h.editBuf = currentInput
    }
    // Move to older entry
    if h.index < len(h.patterns)-1 {
        h.index++
    }
    return h.patterns[h.index]
}

func (h *SearchHistory) NavigateDown() string {
    if h.index < 0 {
        return h.editBuf
    }
    h.index--
    if h.index == -1 {
        return h.editBuf
    }
    return h.patterns[h.index]
}

func (h *SearchHistory) Reset() {
    h.index = -1
    h.editBuf = ""
}
```

#### internal/ui/regex_search_dialog.go

```go
package ui

import (
    "regexp"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
)

type RegexSearchDialog struct {
    BaseDialog
    textInput *TextInput
    history   *SearchHistory
    errorMsg  string
    styles    DialogStyles
}

type regexSearchResultMsg struct {
    pattern   string
    cancelled bool
}

func NewRegexSearchDialog(history *SearchHistory) *RegexSearchDialog {
    base := NewBaseDialog(DialogDisplayPane)
    history.Reset()
    return &RegexSearchDialog{
        BaseDialog: base,
        textInput:  NewTextInput(""),
        history:    history,
        errorMsg:   "",
        styles:     DefaultDialogStyles(base.Width()),
    }
}

func (d *RegexSearchDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
    // ... implementation as shown above ...
}

func (d *RegexSearchDialog) View() string {
    if !d.IsActive() {
        return ""
    }

    var b strings.Builder
    width := d.Width()

    // Title
    b.WriteString(d.styles.Title.Render("Regex Search"))
    b.WriteString("\n\n")

    // Input field
    inputWidth := width - 8
    inputStyle := d.styles.Input.Width(inputWidth)
    b.WriteString(inputStyle.Render(d.textInput.RenderWithCursor(inputWidth - 2)))
    b.WriteString("\n")

    // Error message
    if d.errorMsg != "" {
        b.WriteString("\n")
        b.WriteString(d.styles.Error.Render(d.errorMsg))
    }

    b.WriteString("\n")

    // Syntax hints
    b.WriteString(d.styles.Footer.Render("Examples: ^prefix  suffix$  \\.txt$"))
    b.WriteString("\n\n")

    // Footer
    b.WriteString(d.styles.Footer.Render("Enter: Search  Esc: Cancel  ↑↓: History"))

    return d.styles.Box.Render(b.String())
}
```

#### internal/ui/query_search_dialog.go

Similar structure to RegexSearchDialog with:
- Title: "Query Filter"
- Hints: `size > 1MB  ext = ".go"` and `name LIKE "test%"`
- Validation using `filter.ValidateQuery()`

### Minibuffer Code to Remove/Modify

After implementing dialogs, simplify `startSearch()` in `model.go`:

```go
func (m *Model) startSearch(mode SearchMode) {
    // Only handle incremental search now
    if mode != SearchModeIncremental {
        return // Should not happen - dialogs handle other modes
    }

    // ... existing incremental search setup ...
    m.searchState.Mode = mode
    m.searchState.Pattern = ""
    m.searchState.IsActive = true
    m.minibuffer.SetPrompt("/: ")
    m.minibuffer.Clear()
    m.minibuffer.SetWidth(m.getActivePane().width)
    m.minibuffer.Show()
}
```

## Test Scenarios

### Unit Tests

#### SearchHistory Tests
- [ ] Add pattern to empty history
- [ ] Add duplicate pattern moves to front
- [ ] NavigateUp returns patterns in order
- [ ] NavigateDown returns to original input
- [ ] NavigateUp at end stays at last entry
- [ ] NavigateDown at beginning stays at original
- [ ] Reset clears navigation state
- [ ] History respects maxSize limit

#### RegexSearchDialog Tests
- [ ] New dialog is active and shows correct title
- [ ] Enter with valid regex returns success message
- [ ] Enter with invalid regex shows error, stays open
- [ ] Enter with empty input returns empty pattern
- [ ] Esc returns cancelled message
- [ ] Up/Down updates input from history
- [ ] Text input handles regular characters
- [ ] DisplayType is DialogDisplayPane

#### QuerySearchDialog Tests
- [ ] New dialog is active and shows correct title
- [ ] Enter with valid query returns success message
- [ ] Enter with invalid query shows error, stays open
- [ ] Enter with empty input returns empty pattern
- [ ] Esc returns cancelled message
- [ ] Up/Down updates input from history
- [ ] Text input handles regular characters
- [ ] DisplayType is DialogDisplayPane

### Integration Tests
- [ ] Ctrl+F opens RegexSearchDialog
- [ ] Ctrl+G opens QuerySearchDialog
- [ ] Dialog result applies filter to pane
- [ ] Empty result clears filter
- [ ] Cancelled result doesn't change filter

### Regression Tests
- [ ] `/` still opens minibuffer for incremental search
- [ ] Incremental search filters in real-time
- [ ] Incremental search Enter confirms filter
- [ ] Incremental search Esc cancels

## Security Considerations

- **Input Validation:** Regex patterns are validated using `regexp.Compile` before use
- **Query Validation:** SQL-like queries are validated using `filter.ValidateQuery` before use
- **No File System Access:** Dialogs only handle user input, no direct file operations

## Error Handling

### Error Cases

| Error | Condition | Handling |
|-------|-----------|----------|
| Invalid regex | `regexp.Compile` fails | Show error in dialog, keep open |
| Invalid query | `filter.ValidateQuery` fails | Show error in dialog, keep open |

### Error Display

Errors are displayed inline within the dialog using the `Error` style (red text):
```
┌─────────────────────────────────────┐
│ pattern with error                   │
└─────────────────────────────────────┘

error parsing regexp: missing closing ): `^(test`
```

## Success Criteria

- [ ] All functional requirements (FR1-FR9) are implemented
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] All regression tests pass
- [ ] Code follows existing patterns and style
- [ ] No unused minibuffer code for regex/query search remains

## File Structure

```
internal/ui/
├── regex_search_dialog.go           # New: Regex search dialog
├── regex_search_dialog_test.go      # New: Tests
├── query_search_dialog.go           # New: Query search dialog
├── query_search_dialog_test.go      # New: Tests
├── search_history.go                # New: History helper
├── search_history_test.go           # New: Tests
├── model.go                         # Modified: Add history fields
├── model_update_keyboard.go         # Modified: Open dialogs
├── model_update.go                  # Modified: Handle result messages
└── search.go                        # Modified: Simplify (keep incremental)
```

## References

- Requirements Document: `doc/tasks/adjust-search-ui/要件定義書.md`
- InputDialog Implementation: `internal/ui/input_dialog.go`
- BaseDialog Implementation: `internal/ui/dialog_base.go`
- TextInput Implementation: `internal/ui/text_input.go`
- Current Minibuffer: `internal/ui/minibuffer.go`
- Current Search: `internal/ui/search.go`
- Query Filter: `internal/filter/filter.go`
