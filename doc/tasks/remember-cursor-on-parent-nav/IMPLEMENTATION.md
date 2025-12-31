# Implementation Plan: Remember Cursor Position on Parent Directory Navigation

## Overview

This document describes the implementation plan for positioning the cursor on the previous subdirectory when navigating to a parent directory.

## Current Implementation Analysis

### Relevant Files and Methods

1. **`internal/ui/pane.go`**
   - `MoveToParent()` - Synchronous parent navigation (line 361-369)
   - `MoveToParentAsync()` - Asynchronous parent navigation (line 372-382)
   - `EnterDirectory()` - Directory entry via `..` (synchronous) (line 313-358)
   - `EnterDirectoryAsync()` - Directory entry via `..` (asynchronous) (line 266-310)
   - `LoadDirectory()` - Loads directory contents and resets cursor to 0 (line 93-118)
   - `LoadDirectoryAsync()` - Asynchronous directory loading (line 187-209)

2. **`internal/ui/model.go`**
   - `directoryLoadCompleteMsg` handling (line 374-414) - Processes async load completion
   - Key handling for `h`/`l`/`Enter` keys that trigger navigation

3. **`internal/ui/messages.go`**
   - `directoryLoadCompleteMsg` struct (line 36-42) - Message for async load completion

### Current Navigation Flow

```
User Action → Navigation Method → LoadDirectory → cursor = 0
                                               → scrollOffset = 0
```

**Synchronous flow:**
1. User presses `h`/`l` or selects `..` + Enter
2. `MoveToParent()` or `EnterDirectory()` is called
3. `LoadDirectory()` resets cursor to 0

**Asynchronous flow:**
1. User triggers navigation
2. `MoveToParentAsync()` or `EnterDirectoryAsync()` sets `pendingPath` and returns `tea.Cmd`
3. `LoadDirectoryAsync()` executes in background
4. `directoryLoadCompleteMsg` is sent to `Update()`
5. `Update()` sets `targetPane.cursor = 0`

## Design

### Approach

Store the subdirectory name before navigation and use it to find the cursor position after loading. This is a minimal change that follows the YAGNI principle.

### Data Flow

```
1. Before navigation: Extract subdirectory name from current path
2. Store subdirectory name temporarily
3. Load parent directory
4. After loading: Search for subdirectory by name
5. Set cursor to found index (or 0 if not found)
```

### Implementation Strategy

**Option A: Store in Pane struct field (Selected)**
- Add `pendingCursorTarget string` field to `Pane`
- Set before navigation, use after load completion
- Clear after use
- Pros: Clean, explicit state management
- Cons: Additional field in Pane struct

**Option B: Include in directoryLoadCompleteMsg**
- Add `cursorTarget string` field to message
- Pass through async flow
- Pros: No struct change
- Cons: Message modification, harder to track state

**Selected: Option A** - More maintainable and works for both sync and async flows.

## Detailed Implementation

### Phase 1: Add Pane Field

**File: `internal/ui/pane.go`**

Add new field to `Pane` struct:

```go
// Pane struct (around line 41)
type Pane struct {
    // ... existing fields ...
    pendingCursorTarget string // Target entry name for cursor positioning after parent navigation
}
```

**Contract:**
- `pendingCursorTarget` is set before parent navigation
- `pendingCursorTarget` is used after directory load to find cursor position
- `pendingCursorTarget` is cleared after use or on any other navigation

### Phase 2: Helper Function for Subdirectory Extraction

**File: `internal/ui/pane.go`**

Add helper function:

```go
// extractSubdirName extracts the subdirectory name from the current path
// relative to the parent directory.
//
// Precondition: p.path is not root "/"
// Postcondition: Returns the base name of p.path (e.g., "/home/user/docs" -> "docs")
//
// Example:
//   Current path: /home/user/documents
//   Returns: "documents"
func (p *Pane) extractSubdirName() string
```

**Pseudocode:**
```
func extractSubdirName():
    return filepath.Base(p.path)
```

### Phase 3: Helper Function for Cursor Positioning

**File: `internal/ui/pane.go`**

Add helper function:

```go
// findEntryIndex finds the index of an entry by name in the current entries list.
//
// Precondition: p.entries is populated
// Postcondition: Returns index if found, -1 if not found
//
// Parameters:
//   name: The entry name to search for
//
// Returns:
//   Index of the entry if found, -1 otherwise
func (p *Pane) findEntryIndex(name string) int
```

**Pseudocode:**
```
func findEntryIndex(name):
    for i, entry in p.entries:
        if entry.Name == name:
            return i
    return -1
```

### Phase 4: Modify MoveToParent (Synchronous)

**File: `internal/ui/pane.go`**

Modify `MoveToParent()` (line 361-369):

```go
// MoveToParent moves to parent directory with cursor positioning.
//
// Precondition: p.path is a valid directory
// Postcondition:
//   - If at root: No change
//   - Otherwise: p.path is parent directory
//   - Cursor is on previous subdirectory if found, else at index 0
func (p *Pane) MoveToParent() error
```

**Pseudocode:**
```
func MoveToParent():
    if p.path == "/":
        return nil

    // Store subdirectory name before navigation
    subdirName := p.extractSubdirName()

    p.recordPreviousPath()
    p.path = filepath.Dir(p.path)

    err := p.LoadDirectory()
    if err != nil:
        return err

    // Position cursor on previous subdirectory
    index := p.findEntryIndex(subdirName)
    if index >= 0:
        p.cursor = index
        p.adjustScroll()
    // else: cursor already at 0 from LoadDirectory()

    return nil
```

### Phase 5: Modify MoveToParentAsync (Asynchronous)

**File: `internal/ui/pane.go`**

Modify `MoveToParentAsync()` (line 372-382):

```go
// MoveToParentAsync initiates async move to parent directory.
//
// Precondition: p.path is a valid directory
// Postcondition:
//   - If at root: Returns nil
//   - Otherwise: Returns tea.Cmd for async load
//   - p.pendingCursorTarget is set to subdirectory name
func (p *Pane) MoveToParentAsync() tea.Cmd
```

**Pseudocode:**
```
func MoveToParentAsync():
    if p.path == "/":
        return nil

    // Store subdirectory name for cursor positioning after load
    p.pendingCursorTarget = p.extractSubdirName()

    newPath := filepath.Dir(p.path)
    p.recordPreviousPath()
    p.pendingPath = newPath
    p.path = newPath
    p.StartLoadingDirectory()
    return LoadDirectoryAsync(p.paneID, newPath, p.sortConfig)
```

### Phase 6: Modify EnterDirectory for `..` (Synchronous)

**File: `internal/ui/pane.go`**

Modify `EnterDirectory()` (line 313-358) for the `..` case:

**Pseudocode:**
```
func EnterDirectory():
    entry := p.SelectedEntry()
    if entry == nil:
        return nil

    // ... existing symlink handling ...

    if !entry.IsDir && !entry.IsSymlink:
        return nil

    var newPath string
    var subdirName string  // NEW: track subdirectory name for parent nav

    if entry.IsParentDir():
        subdirName = p.extractSubdirName()  // NEW: store before navigation
        newPath = filepath.Dir(p.path)
    else:
        newPath = filepath.Join(p.path, entry.Name)

    p.recordPreviousPath()
    p.path = newPath
    err := p.LoadDirectory()
    if err != nil:
        return err

    // NEW: Position cursor on previous subdirectory for parent navigation
    if subdirName != "":
        index := p.findEntryIndex(subdirName)
        if index >= 0:
            p.cursor = index
            p.adjustScroll()

    return nil
```

### Phase 7: Modify EnterDirectoryAsync for `..` (Asynchronous)

**File: `internal/ui/pane.go`**

Modify `EnterDirectoryAsync()` (line 266-310) for the `..` case:

**Pseudocode:**
```
func EnterDirectoryAsync():
    entry := p.SelectedEntry()
    if entry == nil:
        return nil

    // ... existing symlink handling ...

    if !entry.IsDir && !entry.IsSymlink:
        return nil

    var newPath string
    if entry.IsParentDir():
        p.pendingCursorTarget = p.extractSubdirName()  // NEW
        newPath = filepath.Dir(p.path)
    else:
        p.pendingCursorTarget = ""  // NEW: clear for subdirectory navigation
        newPath = filepath.Join(p.path, entry.Name)

    p.recordPreviousPath()
    p.pendingPath = newPath
    p.path = newPath
    p.StartLoadingDirectory()
    return LoadDirectoryAsync(p.paneID, newPath, p.sortConfig)
```

### Phase 8: Modify directoryLoadCompleteMsg Handling

**File: `internal/ui/model.go`**

Modify the `directoryLoadCompleteMsg` handling (around line 374-414):

**Pseudocode:**
```
case directoryLoadCompleteMsg:
    // ... existing pane identification ...

    if targetPane != nil:
        targetPane.loading = false
        targetPane.loadingProgress = ""

        if msg.err != nil:
            // ... existing error handling ...
            targetPane.pendingCursorTarget = ""  // NEW: clear on error
            return m, ...

        // ... existing entry processing ...
        entries := msg.entries
        if !targetPane.showHidden:
            entries = filterHiddenFiles(entries)

        targetPane.allEntries = entries
        targetPane.entries = entries
        targetPane.filterPattern = ""
        targetPane.filterMode = SearchModeNone

        // NEW: Position cursor based on pendingCursorTarget
        if targetPane.pendingCursorTarget != "":
            index := targetPane.findEntryIndex(targetPane.pendingCursorTarget)
            if index >= 0:
                targetPane.cursor = index
            else:
                targetPane.cursor = 0
            targetPane.pendingCursorTarget = ""  // Clear after use
        else:
            targetPane.cursor = 0

        targetPane.scrollOffset = 0
        targetPane.adjustScroll()  // NEW: ensure cursor is visible
        targetPane.pendingPath = ""

        m.updateDiskSpace()
    return m, nil
```

### Phase 9: Clear pendingCursorTarget on Other Navigations

Ensure `pendingCursorTarget` is cleared when navigating to non-parent directories:

**File: `internal/ui/pane.go`**

In `ChangeDirectoryAsync()` (line 392-398):
```go
func (p *Pane) ChangeDirectoryAsync(path string) tea.Cmd {
    p.pendingCursorTarget = ""  // NEW: clear for non-parent navigation
    // ... rest of existing code ...
}
```

In `NavigateToHomeAsync()` (line 917-930):
```go
func (p *Pane) NavigateToHomeAsync() tea.Cmd {
    // ...
    p.pendingCursorTarget = ""  // NEW: clear for non-parent navigation
    // ... rest of existing code ...
}
```

In `NavigateToPreviousAsync()` (line 947-957):
```go
func (p *Pane) NavigateToPreviousAsync() tea.Cmd {
    // ...
    p.pendingCursorTarget = ""  // NEW: clear for non-parent navigation
    // ... rest of existing code ...
}
```

## Test Strategy

### Unit Tests

**File: `internal/ui/pane_test.go`**

1. **Test `extractSubdirName()`**
   ```go
   func TestExtractSubdirName(t *testing.T)
   // Cases:
   // - Normal path: /home/user/docs -> "docs"
   // - Root subdirectory: /home -> "home"
   // - Deep path: /a/b/c/d -> "d"
   ```

2. **Test `findEntryIndex()`**
   ```go
   func TestFindEntryIndex(t *testing.T)
   // Cases:
   // - Entry exists: returns correct index
   // - Entry does not exist: returns -1
   // - Empty entries: returns -1
   ```

3. **Test `MoveToParent()` cursor positioning**
   ```go
   func TestMoveToParentCursorPositioning(t *testing.T)
   // Cases:
   // - Normal case: cursor on previous subdirectory
   // - Subdirectory deleted: cursor at 0
   // - Hidden subdirectory with showHidden=false: cursor at 0
   ```

4. **Test `MoveToParentAsync()` sets pendingCursorTarget**
   ```go
   func TestMoveToParentAsyncSetsPendingCursorTarget(t *testing.T)
   // Verify pendingCursorTarget is set correctly
   ```

5. **Test `EnterDirectory()` via `..`**
   ```go
   func TestEnterDirectoryParentCursorPositioning(t *testing.T)
   // Same cases as MoveToParent
   ```

6. **Test `EnterDirectoryAsync()` via `..`**
   ```go
   func TestEnterDirectoryAsyncParentSetsPendingCursorTarget(t *testing.T)
   // Verify pendingCursorTarget is set for parent, cleared for subdirectory
   ```

### Integration Tests

**File: `internal/ui/model_test.go`**

1. **Test async parent navigation cursor positioning**
   ```go
   func TestModelParentNavigationCursorPositioning(t *testing.T)
   // Simulate full flow: key press -> async load -> cursor positioned
   ```

2. **Test left and right panes operate independently**
   ```go
   func TestPaneIndependentCursorMemory(t *testing.T)
   // Navigate left pane, verify right pane unaffected
   ```

### Test Scenarios from SPEC.md

- [ ] Navigate to parent via `..` entry: cursor on previous subdirectory
- [ ] Navigate to parent via `h` key (left pane): cursor on previous subdirectory
- [ ] Navigate to parent via `l` key (right pane): cursor on previous subdirectory
- [ ] Navigate to parent when subdirectory was deleted: cursor at index 0
- [ ] Navigate to parent from hidden subdirectory with hidden files OFF: cursor at index 0
- [ ] Left pane and right pane maintain independent cursor memory
- [ ] Multiple up/down navigations maintain correct cursor positions

## Implementation Order

1. Add `pendingCursorTarget` field to `Pane` struct
2. Implement `extractSubdirName()` helper
3. Implement `findEntryIndex()` helper
4. Write unit tests for helpers
5. Modify `MoveToParent()` (sync)
6. Write tests for `MoveToParent()`
7. Modify `EnterDirectory()` for `..` case
8. Write tests for `EnterDirectory()` parent case
9. Modify `MoveToParentAsync()` (async)
10. Modify `EnterDirectoryAsync()` for `..` case
11. Modify `directoryLoadCompleteMsg` handling in `model.go`
12. Clear `pendingCursorTarget` in other async navigation methods
13. Write integration tests
14. Run full test suite and verify

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Race condition with async loads | Low | Medium | Use paneID-based targeting (already implemented) |
| Breaking existing navigation | Medium | High | Comprehensive test coverage |
| Performance impact | Low | Low | Name-based search is O(n), negligible for typical directories |

## Verification Checklist

- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Manual testing with various directory structures
- [ ] Verify hidden file handling
- [ ] Verify both sync and async paths work
- [ ] Verify pane independence
- [ ] No regression in existing functionality
