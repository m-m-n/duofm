# Feature: Fix Cursor Position After File Deletion

## Overview

Currently, when a file is deleted in duofm, the cursor always jumps back to the top of the file list (position 0). This creates a poor user experience, especially when deleting files in the middle of large directories, as users lose their working context and must scroll back to continue their work. This feature fixes the cursor positioning behavior to move to the next file after deletion, or to the previous file if deleting the last file in the list.

## Objectives

- Improve user experience by maintaining work context after file deletion
- Implement cursor positioning that matches standard file manager behavior (like Midnight Commander, ranger)
- Ensure smooth workflow when performing multiple consecutive delete operations
- Handle edge cases gracefully (last file, multiple files, empty directory)

## User Stories

### US1: Delete File in Middle of List
As a user, I want the cursor to move to the next file after deletion, so that I can continue working without losing my position in the file list.

**Acceptance Criteria:**
- [ ] When deleting a file that is not the last file, cursor moves to the next file
- [ ] Cursor position is visually maintained (same relative screen position if possible)
- [ ] Scroll position automatically adjusts to keep cursor visible

### US2: Delete Last File in List
As a user, I want the cursor to move to the previous file when I delete the last file, so that I remain on a valid file entry.

**Acceptance Criteria:**
- [ ] When deleting the last file in the list, cursor moves to the new last file
- [ ] Cursor does not go out of bounds
- [ ] Scroll position adjusts appropriately

### US3: Delete Multiple Marked Files
As a user, I want the cursor to remain near my previous position after deleting multiple marked files, so that I can continue working efficiently.

**Acceptance Criteria:**
- [ ] After deleting multiple files, cursor moves to the nearest remaining file
- [ ] If the cursor was on a deleted file, it moves to the next available file
- [ ] If all subsequent files were deleted, cursor moves to the last remaining file

### US4: Delete All Files in Directory
As a user, I want the cursor to move to the parent directory entry (..) when all files are deleted, so that the application remains in a valid state.

**Acceptance Criteria:**
- [ ] When all files/directories are deleted (only .. remains), cursor moves to position 0
- [ ] Application does not crash or show errors
- [ ] User can navigate back to parent directory

## Technical Requirements

### Functional Requirements
- **FR1:** Preserve cursor context by remembering the index before deletion
- **FR2:** Calculate optimal cursor position based on deletion type (single, multiple, last)
- **FR3:** Automatically adjust scroll position to keep cursor visible
- **FR4:** Handle all edge cases without errors or crashes

### Non-Functional Requirements
- **NFR1 - Performance:** Cursor position calculation must complete within 1ms
- **NFR2 - Security:** Maintain existing deletion confirmation flow
- **NFR3 - Usability:** Cursor movement should feel natural and predictable
- **NFR4 - Maintainability:** Code should follow existing Pane patterns and be well-tested

## Implementation Approach

### Architecture

**Current Flow:**
```
User presses 'd' key
  ↓
handleDelete() - Show confirmation dialog
  ↓
User confirms (y key)
  ↓
handleConfirmDialogResult()
  ↓
executeDeleteOperation()
  ↓
deleteFile() for each file
  ↓
LoadDirectory() - RESETS CURSOR TO 0 ← PROBLEM
  ↓
Cursor is at position 0
```

**New Flow:**
```
User presses 'd' key
  ↓
handleDelete() - Show confirmation dialog
  ↓
User confirms (y key)
  ↓
handleConfirmDialogResult()
  ↓
executeDeleteOperation()
  ↓
[NEW] Save current cursor position and file name
  ↓
deleteFile() for each file
  ↓
LoadDirectory() - Loads entries, cursor at 0
  ↓
[NEW] Calculate optimal cursor position
  ↓
[NEW] Set cursor and adjust scroll
  ↓
Cursor is at appropriate position
```

**Component Diagram:**
```
┌─────────────────────────────────────────┐
│  model_update_keyboard.go               │
│  - handleDelete()                       │
└───────────┬─────────────────────────────┘
            │
            ↓
┌─────────────────────────────────────────┐
│  model_update.go                        │
│  - handleConfirmDialogResult()          │
│  - executeDeleteOperation() ← MODIFY    │
└───────────┬─────────────────────────────┘
            │
            ↓
┌─────────────────────────────────────────┐
│  pane.go                                │
│  - LoadDirectory()                      │
│  - SetCursor()                          │
│  - adjustScroll() ← EXISTING           │
│  [NEW] - calculateCursorAfterDeletion() │
└─────────────────────────────────────────┘
```

### Data Flow

```
Current State (before deletion):
- cursor: int (e.g., 5)
- entries: []FileEntry (e.g., 20 files)

Deletion Event:
- deletedIndex: 5
- deletedFileName: "file5.txt"
- isLastFile: false

After LoadDirectory():
- entries: []FileEntry (19 files)

Calculate New Cursor:
- if deletedIndex < len(entries):
    newCursor = deletedIndex (points to next file)
- else:
    newCursor = len(entries) - 1 (last file)

Final State:
- cursor: 5 (now pointing to what was "file6.txt")
- scrollOffset: adjusted to keep cursor visible
```

### API Design

#### Method 1: calculateCursorAfterDeletion (New)

**Purpose:** Calculate the optimal cursor position after file deletion

**Signature:**
```go
func (p *Pane) calculateCursorAfterDeletion(deletedIndex int, deletedCount int) int
```

**Parameters:**
- `deletedIndex`: The reference position for cursor calculation
  - For single file deletion: The cursor position before deletion
  - For multiple file deletion: The smallest index among marked files (first marked file)
- `deletedCount`: Number of files deleted (1 for single, N for multiple)

**Returns:**
- `int`: The new cursor position (0-indexed)

**Logic:**
```go
// If deletedIndex is still valid (there are files after the deleted position)
if deletedIndex < len(p.entries) {
    return deletedIndex
}

// If we're beyond the end, go to last entry
if len(p.entries) > 0 {
    return len(p.entries) - 1
}

// Empty directory (only ..)
return 0
```

**Behavior by Deletion Type:**
- **Single file deletion**: Uses the cursor position as deletedIndex
- **Multiple file deletion**: Uses the smallest index of marked files as deletedIndex
  - This ensures the cursor moves to the file at the position where the first marked file was deleted

#### Method 2: Modified executeDeleteOperation

**Current Implementation:**
```go
func (m Model) executeDeleteOperation() Model {
    activePane := m.getActivePane()
    markedFiles := activePane.GetMarkedFiles()

    if len(markedFiles) > 0 {
        // Delete multiple files
        for _, name := range markedFiles {
            fullPath := filepath.Join(activePane.Path(), name)
            deleteFile(fullPath)
        }
        activePane.ClearMarks()
        activePane.LoadDirectory() // ← Cursor resets to 0 here
    } else {
        // Delete single file
        entry := activePane.SelectedEntry()
        if entry != nil && !entry.IsParentDir() {
            fullPath := filepath.Join(activePane.Path(), entry.Name)
            deleteFile(fullPath)
            activePane.LoadDirectory() // ← Cursor resets to 0 here
        }
    }
    return m
}
```

**New Implementation:**
```go
func (m Model) executeDeleteOperation() Model {
    activePane := m.getActivePane()
    markedFiles := activePane.GetMarkedFiles()

    if len(markedFiles) > 0 {
        // Find the smallest index among marked files (first marked file position)
        minMarkedIndex := -1
        for i, entry := range activePane.entries {
            if entry.IsMarked() {
                if minMarkedIndex == -1 || i < minMarkedIndex {
                    minMarkedIndex = i
                }
            }
        }

        deletedCount := len(markedFiles)

        // Delete multiple files
        var deleteErr error
        for _, name := range markedFiles {
            fullPath := filepath.Join(activePane.Path(), name)
            if err := deleteFile(fullPath); err != nil {
                deleteErr = err
                break
            }
        }

        if deleteErr != nil {
            m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", deleteErr))
        }

        activePane.ClearMarks()
        activePane.LoadDirectory()

        // Calculate and set new cursor position based on first marked file position
        newCursor := activePane.calculateCursorAfterDeletion(minMarkedIndex, deletedCount)
        activePane.SetCursor(newCursor)
        activePane.EnsureCursorVisible()

    } else {
        entry := activePane.SelectedEntry()
        if entry != nil && !entry.IsParentDir() {
            // Remember cursor position before deletion
            cursorBeforeDeletion := activePane.cursor

            fullPath := filepath.Join(activePane.Path(), entry.Name)
            if err := deleteFile(fullPath); err != nil {
                m.dialog = NewErrorDialog(fmt.Sprintf("Failed to delete: %v", err))
            } else {
                activePane.LoadDirectory()

                // Calculate and set new cursor position
                newCursor := activePane.calculateCursorAfterDeletion(cursorBeforeDeletion, 1)
                activePane.SetCursor(newCursor)
                activePane.EnsureCursorVisible()
            }
        }
    }

    return m
}
```

### Database Schema

Not applicable - this feature does not require database changes.

### Dependencies

**Internal Dependencies:**
- `internal/ui/pane.go`: Cursor and scroll management
- `internal/ui/model_update.go`: Delete operation execution
- `internal/fs/operations.go`: File deletion functionality (unchanged)

**External Dependencies:**
None - uses existing Go standard library and Bubble Tea framework

### File Structure

```
internal/
├── ui/
│   ├── model_update.go           # Modify executeDeleteOperation()
│   ├── model_update_test.go      # Add tests for cursor positioning
│   ├── pane.go                   # Add calculateCursorAfterDeletion()
│   └── pane_test.go              # Add unit tests
test/
└── e2e/
    └── scripts/
        └── run_all_tests.sh      # Add E2E test for cursor behavior
```

## Test Scenarios

### Unit Tests

#### Test 1: Delete Middle File
**Description:** Delete a file in the middle of the list
**Expected:** Cursor moves to next file (same index)

```go
func TestCalculateCursorAfterDeletion_MiddleFile(t *testing.T) {
    // Setup: Directory with 5 files (+ parent dir = 6 entries)
    // Cursor at index 2 (third file)
    // After deleting: 5 entries remain
    // Expected cursor: 2 (now pointing to what was index 3)

    deletedIndex := 2
    deletedCount := 1
    // After deletion, we have 5 entries

    result := calculateCursorAfterDeletion(deletedIndex, deletedCount, 5)

    if result != 2 {
        t.Errorf("Expected cursor at 2, got %d", result)
    }
}
```

#### Test 2: Delete Last File
**Description:** Delete the last file in the list
**Expected:** Cursor moves to previous file (new last file)

```go
func TestCalculateCursorAfterDeletion_LastFile(t *testing.T) {
    // Setup: Directory with 5 files (+ parent dir = 6 entries)
    // Cursor at index 5 (last file)
    // After deleting: 5 entries remain
    // Expected cursor: 4 (new last file)

    deletedIndex := 5
    deletedCount := 1

    result := calculateCursorAfterDeletion(deletedIndex, deletedCount, 5)

    if result != 4 {
        t.Errorf("Expected cursor at 4, got %d", result)
    }
}
```

#### Test 3: Delete Multiple Files
**Description:** Delete multiple marked files
**Expected:** Cursor at nearest remaining file

```go
func TestCalculateCursorAfterDeletion_MultipleFiles(t *testing.T) {
    tests := []struct {
        name          string
        deletedIndex  int
        deletedCount  int
        remainingCount int
        expectedCursor int
    }{
        {
            name:           "Delete 3 files from middle",
            deletedIndex:   3,
            deletedCount:   3,
            remainingCount: 7,
            expectedCursor: 3,
        },
        {
            name:           "Delete all files after cursor",
            deletedIndex:   5,
            deletedCount:   5,
            remainingCount: 1, // Only .. remains
            expectedCursor: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := calculateCursorAfterDeletion(
                tt.deletedIndex,
                tt.deletedCount,
                tt.remainingCount,
            )
            if result != tt.expectedCursor {
                t.Errorf("Expected %d, got %d", tt.expectedCursor, result)
            }
        })
    }
}
```

#### Test 4: Delete All Files
**Description:** Delete all files (only parent directory remains)
**Expected:** Cursor at position 0 (parent directory)

```go
func TestCalculateCursorAfterDeletion_AllFiles(t *testing.T) {
    deletedIndex := 2
    deletedCount := 5
    remainingCount := 1 // Only .. remains

    result := calculateCursorAfterDeletion(deletedIndex, deletedCount, remainingCount)

    if result != 0 {
        t.Errorf("Expected cursor at 0 (parent dir), got %d", result)
    }
}
```

### Integration Tests

#### Test 5: executeDeleteOperation Integration
**Description:** Test the full delete operation flow with cursor positioning
**Expected:** Cursor positioned correctly after actual deletion and reload

```go
func TestExecuteDeleteOperation_CursorPosition(t *testing.T) {
    tmpDir := t.TempDir()

    // Create test files
    for i := 1; i <= 5; i++ {
        os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", i)), []byte("test"), 0644)
    }

    model := NewModel()
    model.leftPath = tmpDir
    // Initialize model...

    // Move cursor to middle file (index 3)
    for i := 0; i < 3; i++ {
        model.getActivePane().MoveCursorDown()
    }

    initialCursor := model.getActivePane().cursor

    // Execute delete
    model = model.executeDeleteOperation().(Model)

    finalCursor := model.getActivePane().cursor

    // Cursor should be at same index (pointing to next file)
    if finalCursor != initialCursor {
        t.Errorf("Expected cursor at %d, got %d", initialCursor, finalCursor)
    }
}
```

### E2E Tests

#### Test 6: Delete File with Cursor Movement (E2E)
**Description:** End-to-end test of file deletion with cursor positioning
**Expected:** Cursor visually moves to next file

```bash
test_delete_cursor_position() {
    start_duofm "$CURRENT_SESSION"

    # Move to third file
    send_keys "$CURRENT_SESSION" "j" "j"
    sleep 0.2

    # Capture current file name
    screen=$(capture_screen "$CURRENT_SESSION")
    current_file=$(echo "$screen" | grep ">" | head -1 | awk '{print $2}')

    # Delete file
    send_keys "$CURRENT_SESSION" "d"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "y"
    sleep 0.5

    # Verify cursor moved to next file (not back to top)
    assert_not_contains "$CURRENT_SESSION" "file1.txt" \
        "Cursor should not be on first file after deletion"

    stop_duofm "$CURRENT_SESSION"
}
```

#### Test 7: Delete Last File (E2E)
**Description:** Delete the last file and verify cursor on previous file
**Expected:** Cursor on second-to-last file after deletion

```bash
test_delete_last_file() {
    start_duofm "$CURRENT_SESSION"

    # Move to last file
    send_keys "$CURRENT_SESSION" "G"
    sleep 0.2

    # Delete last file
    send_keys "$CURRENT_SESSION" "d" "y"
    sleep 0.5

    # Verify not on parent directory (..)
    screen=$(capture_screen "$CURRENT_SESSION")
    current_line=$(echo "$screen" | grep ">" | head -1)

    if echo "$current_line" | grep -q "\.\.$"; then
        fail "Cursor should not be on parent directory after deleting last file"
    fi

    stop_duofm "$CURRENT_SESSION"
}
```

### Edge Cases

#### Edge Case 1: Two Files Only
**Description:** Directory with parent (..) and one file, delete that file
**Expected:** Cursor moves to parent directory (index 0)

```go
func TestCalculateCursorAfterDeletion_TwoEntriesOnly(t *testing.T) {
    // Before: [.., file.txt]  cursor=1
    // After:  [..]            cursor=0

    result := calculateCursorAfterDeletion(1, 1, 1)

    if result != 0 {
        t.Errorf("Expected cursor at 0, got %d", result)
    }
}
```

#### Edge Case 2: Delete First Real File
**Description:** Delete the first file after parent directory
**Expected:** Cursor stays at same index (now pointing to next file)

```go
func TestCalculateCursorAfterDeletion_FirstRealFile(t *testing.T) {
    // Before: [.., file1, file2, file3]  cursor=1
    // After:  [.., file2, file3]         cursor=1

    result := calculateCursorAfterDeletion(1, 1, 3)

    if result != 1 {
        t.Errorf("Expected cursor at 1, got %d", result)
    }
}
```

### Performance Tests

#### Test 8: Large Directory Performance
**Description:** Test cursor calculation performance with 1000+ files
**Expected:** Calculation completes within 1ms

```go
func BenchmarkCalculateCursorAfterDeletion(b *testing.B) {
    // Simulate large directory
    entriesCount := 1000
    deletedIndex := 500

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        calculateCursorAfterDeletion(deletedIndex, 1, entriesCount-1)
    }
}
```

## Security Considerations

- **Authentication:** Not applicable - local file operation
- **Authorization:** Inherits existing file system permissions
- **Input Validation:** Cursor position is bounds-checked before setting
- **Data Protection:** No sensitive data involved
- **XSS Prevention:** Not applicable - TUI application
- **SQL Injection Prevention:** Not applicable - no database
- **CSRF Protection:** Not applicable - local application

## Error Handling

### Error Codes

| Code | Description | HTTP Status | User Message |
|------|-------------|-------------|--------------|
| ERR_DEL_001 | Failed to delete file | N/A | "Failed to delete: [error details]" |
| ERR_DEL_002 | Invalid cursor position | N/A | (Silent fallback to position 0) |

### Error Flow

```
Delete Operation
  ↓
deleteFile() fails
  ↓
Show ErrorDialog with error message
  ↓
LoadDirectory() still called (to refresh state)
  ↓
Cursor positioned as if delete succeeded
  ↓
User can retry or cancel
```

**Error Recovery:**
- If deletion fails, error dialog is shown but directory is still reloaded
- Cursor positioning proceeds normally even after partial failures
- Invalid cursor positions are clamped to valid range [0, len(entries)-1]

## Performance Optimization

### Performance Goals
- Cursor position calculation: < 1ms
- Directory reload + cursor positioning: < 100ms (user-imperceptible)
- No performance degradation for large directories (1000+ files)

### Optimization Strategies
- **Simple Arithmetic:** Use index-based calculation (O(1) time complexity)
- **No File System Calls:** Cursor calculation based only on in-memory data
- **Minimal Overhead:** Add only 3-4 integer operations to existing flow

### Caching Strategy
Not applicable - cursor position is ephemeral state, no caching needed

## Success Criteria

- [x] All functional requirements are implemented and tested
- [ ] All test scenarios pass (unit, integration, E2E)
- [ ] Performance meets specified goals (< 1ms calculation time)
- [ ] Security requirements are satisfied (inherits existing permissions)
- [ ] Documentation is complete
- [ ] Code review is completed
- [ ] No regression in existing delete functionality
- [ ] E2E tests demonstrate expected user experience

## Open Questions

None - all requirements have been clarified with the user.

## Implementation Phases

### Phase 1: Core Implementation
**Goals:** Implement basic cursor positioning after deletion
**Deliverables:**
- Add `calculateCursorAfterDeletion()` method to Pane
- Modify `executeDeleteOperation()` to preserve and restore cursor
- Unit tests for cursor calculation logic

### Phase 2: Testing and Refinement
**Goals:** Ensure all edge cases are handled correctly
**Deliverables:**
- Comprehensive unit test suite
- Integration tests for delete operations
- E2E tests for user-visible behavior
- Performance benchmarks

### Phase 3: Documentation and Release
**Goals:** Complete documentation and prepare for release
**Deliverables:**
- Update CHANGELOG.md
- Code review and merge
- Release notes

## References

- User request: Original bug report describing cursor reset issue
- `internal/ui/pane.go`: Existing cursor management methods
- `internal/ui/model_update.go`: Delete operation implementation
- `test/README.md`: Testing guidelines
- Midnight Commander behavior: Industry standard reference
