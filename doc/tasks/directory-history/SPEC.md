# Feature: Directory History Navigation

## Overview

Add browser-like directory history navigation to duofm. Each pane maintains an independent history stack of visited directories, allowing users to navigate backward and forward using `Alt+←`/`Alt+→` or `[`/`]` keys.

### Background

The current implementation provides the `-` key to toggle between the last two directories via the `previousPath` field. While useful for simple back-and-forth navigation, this is limited to only two locations. To enable more flexible directory navigation, we need a complete history stack similar to web browser history.

### Objectives

- Improve directory navigation efficiency
- Enable easy return to previous locations after deep directory exploration
- Enhance overall usability of the TUI file manager

## Domain Rules

### History Fundamentals

- **DR1.1**: Each pane (left and right) maintains an independent history stack
- **DR1.2**: Maximum history size is 100 entries
- **DR1.3**: When exceeding 100 entries, oldest entries are removed (FIFO)
- **DR1.4**: History is session-only (cleared on application exit)
- **DR1.5**: All directory transitions are recorded in history, including parent directory navigation

### History State Management

- **DR2.1**: History has a "current position" concept (same as browser history)
- **DR2.2**: When navigating to a new directory, all history after the current position is deleted, and the new entry is added
- **DR2.3**: "Back" operation moves the current position backward in the stack without deleting history
- **DR2.4**: "Forward" operation moves the current position forward in the stack without deleting history

**Example:**
```
Initial state: [A, B, C, D*, E]  (* = current position)
Back:          [A, B, C*, D, E]
Back:          [A, B*, C, D, E]
New nav:       [A, B, F*]  (C, D, E are deleted)
```

### History Recording Conditions

The following directory transitions are recorded in history:

- **DR3.1**: `Enter` key to enter subdirectory
- **DR3.2**: `h`/`←` key or `..` to move to parent directory
- **DR3.3**: `~` key to navigate to home directory
- **DR3.4**: `-` key to toggle to previous directory
- **DR3.5**: Bookmark navigation
- **DR3.6**: `=` key for pane synchronization
- **DR3.7**: Following symlinks to directories
- **DR3.8**: All other directory change operations

**Exceptions (NOT recorded):**
- History navigation itself (`Alt+←`/`Alt+→`/`[`/`]`)

### Deleted/Renamed Directories

- **DR4.1**: Directories remain in history even if deleted or renamed
- **DR4.2**: Attempting to navigate to a non-existent directory displays an error message and stays in the current directory
- **DR4.3**: Error message format: "Directory not found: /path/to/dir"

### Coexistence with Existing Features

- **DR5.1**: The `-` key's existing functionality (toggle to previous directory) is **retained**
- **DR5.2**: Navigation via `-` key is also recorded in history (DR3.4)
- **DR5.3**: History feature and `-` key feature operate independently

## Functional Requirements

### History Storage

- **FR1.1**: Add a history stack field to each `Pane` struct
- **FR1.2**: History stack contains:
  - History list (max 100 entries)
  - Current position index
- **FR1.3**: Add new path to history on directory transition
- **FR1.4**: Remove oldest entry when history exceeds 100 entries
- **FR1.5**: The `previousPath` field is **retained** for `-` key implementation

### History Navigation

- **FR2.1**: `Alt+←` key navigates one step backward in history
- **FR2.2**: `Alt+→` key navigates one step forward in history
- **FR2.3**: `[` key navigates backward (same as `Alt+←`)
- **FR2.4**: `]` key navigates forward (same as `Alt+→`)
- **FR2.5**: When no backward history exists, operation is ignored (no error)
- **FR2.6**: When no forward history exists, operation is ignored (no error)

### Relationship with Existing Features

- **FR3.1**: The `-` key's existing behavior (toggle to previous directory) is **retained**
- **FR3.2**: The `previousPath` field is **retained** and continues to be used for `-` key implementation
- **FR3.3**: Navigation via `-` key is recorded in history

### Error Handling

- **FR4.1**: When history navigation target directory does not exist:
  - Display error message in status bar for 5 seconds
  - Stay in current directory
  - Do not change history current position (allowing retry)
- **FR4.2**: When history navigation target directory lacks read permission:
  - Use existing directory error handling
  - Display error message in status bar

## Non-Functional Requirements

### Performance

- **NFR1.1**: History append operation completes in O(1)
- **NFR1.2**: History navigation operation completes in O(1)
- **NFR1.3**: Memory usage is negligible (approximately 100 bytes per path × 100 entries × 2 panes = ~20KB)

### Reliability

- **NFR2.1**: Adding history feature does not affect existing directory navigation functionality
- **NFR2.2**: Application does not crash even if history becomes corrupted (should not normally occur)

### Maintainability

- **NFR3.1**: History management logic is encapsulated within the `Pane` struct
- **NFR3.2**: Changing history internal implementation does not affect external interface

## Interface Contract

### Data Structures

#### Directory History

```go
type DirectoryHistory struct {
    paths        []string  // History path list (max 100 entries)
    currentIndex int       // Current position index (-1 = no history)
    maxSize      int       // Maximum size (100)
}
```

**Invariants:**
- `0 <= currentIndex < len(paths)` or `currentIndex == -1` (when history is empty)
- `len(paths) <= maxSize`

#### Pane Struct Addition

```go
type Pane struct {
    // ... existing fields ...
    previousPath string          // Previous directory path (for - key, retained)
    history      DirectoryHistory // Directory history (newly added)
}
```

### Operations

#### Add to History

```go
Function: AddToHistory(path string)
Precondition: path is a valid directory path
Postcondition:
  - All entries after current position are deleted
  - New path is appended to history
  - currentIndex points to the end
  - If history exceeds 100 entries, remove first entry
  - Do not add duplicate consecutive paths
```

**Example:**
```
Initial: paths=[A, B, C], currentIndex=2
AddToHistory(D) → paths=[A, B, C, D], currentIndex=3

Mid-state: paths=[A, B, C, D], currentIndex=1 (pointing to B)
AddToHistory(E) → paths=[A, B, E], currentIndex=2
```

#### Navigate Backward

```go
Function: NavigateBack() (path string, ok bool)
Precondition: none
Returns:
  - path: Destination path
  - ok: true=can navigate, false=no backward history
Postcondition:
  - If ok is true, currentIndex decreases by 1
  - If ok is false, state unchanged
```

#### Navigate Forward

```go
Function: NavigateForward() (path string, ok bool)
Precondition: none
Returns:
  - path: Destination path
  - ok: true=can navigate, false=no forward history
Postcondition:
  - If ok is true, currentIndex increases by 1
  - If ok is false, state unchanged
```

#### State Check

```go
Function: CanGoBack() bool
Returns: Whether backward history exists

Function: CanGoForward() bool
Returns: Whether forward history exists
```

### Keybinding Definitions

#### New Actions

```go
ActionHistoryBack    // Navigate backward in history
ActionHistoryForward // Navigate forward in history
```

#### Key Mapping

```
"history_back":    ["Alt+Left", "["]
"history_forward": ["Alt+Right", "]"]
"prev_dir":        ["-"]  // Existing, retained
```

## Test Scenarios

### Basic Behavior

- [ ] Scenario 1: Moving to a new directory adds it to history
- [ ] Scenario 2: `Alt+←` or `[` navigates backward in history
- [ ] Scenario 3: `Alt+→` or `]` navigates forward in history
- [ ] Scenario 4: Back/forward operations with no history do nothing
- [ ] Scenario 5: Left and right panes have independent histories

### History State Management

- [ ] Scenario 6: After A→B→C, going back twice then navigating to D results in history [A, B, D]
- [ ] Scenario 7: When exceeding 100 entries, oldest entries are removed
- [ ] Scenario 8: Consecutive navigation to same directory is recorded only once

### Various Navigation Methods

- [ ] Scenario 9: `Enter` to subdirectory is recorded in history
- [ ] Scenario 10: `h`/`←` to parent directory is recorded in history
- [ ] Scenario 11: `~` to home directory is recorded in history
- [ ] Scenario 12: `-` to previous directory is recorded in history
- [ ] Scenario 13: Bookmark navigation is recorded in history
- [ ] Scenario 14: `=` pane sync is recorded in history
- [ ] Scenario 15: Following symlinks is recorded in history

### History Navigation Itself

- [ ] Scenario 16: After going back with `Alt+←`, navigating to a new directory clears forward history
- [ ] Scenario 17: History navigation (`Alt+←`/`Alt+→`/`[`/`]`) is not recorded in history

### Integration with Existing Features

- [ ] Scenario 18: `-` key continues to work normally (toggle to previous directory)
- [ ] Scenario 19: Navigation via `-` key is recorded in history
- [ ] Scenario 20: After A→B→C, `-` goes to B, `[` goes to A

### Error Handling

- [ ] Scenario 21: Attempting to navigate to a deleted directory displays error message
- [ ] Scenario 22: On error, history current position is unchanged (retry possible)
- [ ] Scenario 23: Navigating to directory without permission displays appropriate error

## Success Criteria

- [ ] All test scenarios pass
- [ ] Existing directory navigation features (including `-` key) are unaffected
- [ ] Maximum history size (100 entries) works correctly
- [ ] Left and right panes have independent functioning histories
- [ ] Keybindings (`Alt+←`/`Alt+→`/`[`/`]`) work correctly
- [ ] Error handling functions appropriately
- [ ] Performance impact is negligible

## Constraints

### Technical Constraints

- Bubble Tea (TUI framework) limitations may cause `Alt+key` combinations to vary by terminal emulator
- Some terminals may not recognize `Alt+←`/`Alt+→` correctly (that's why `[`/`]` are also provided)

### Business Constraints

- Session-only history (not persisted)
- Maximum history size is fixed at 100 entries (not configurable)

## Open Questions

None (all requirements are clarified)

## References

### Similar Implementations

- Web browser history (Chrome, Firefox, etc.)
- Midnight Commander (`Alt+O` for history dialog)
- ranger (parent-child directory navigation history)

### Existing Code

- `internal/ui/pane.go`: `previousPath` field implementation (retained)
- `internal/ui/keys.go`: Key definitions
- `internal/ui/actions.go`: Action definitions
- `internal/config/defaults.go`: Default keybindings
