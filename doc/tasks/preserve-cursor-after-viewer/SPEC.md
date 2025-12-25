# Feature: Preserve Cursor Position After External Viewer

## Overview

When returning from an external application (less/vim), the cursor position in the file list should be preserved instead of resetting to the first item.

## Objectives

- Preserve cursor position after external viewer (less) exits
- Preserve cursor position after external editor (vim) exits
- Maintain correct directory reload behavior for file changes

## User Stories

- As a user, I want to return to the same file position after viewing a file with less, so that I can continue browsing from where I left off

## Technical Requirements

### New Method: RefreshDirectoryPreserveCursor

Add a new method `RefreshDirectoryPreserveCursor()` to `Pane` that reloads directory contents while preserving cursor position if the selected file still exists.

```go
// RefreshDirectoryPreserveCursor reloads directory contents while preserving cursor position
// If the previously selected file no longer exists, cursor resets to the beginning
func (p *Pane) RefreshDirectoryPreserveCursor() error {
    // Store current selected file name
    var selectedName string
    if entry := p.SelectedEntry(); entry != nil {
        selectedName = entry.Name
    }

    // Reload directory entries
    entries, err := fs.ReadDirectory(p.path)
    if err != nil {
        return err
    }

    entries = SortEntries(entries, p.sortConfig)

    if !p.showHidden {
        entries = filterHiddenFiles(entries)
    }

    p.allEntries = entries
    p.entries = entries
    p.filterPattern = ""
    p.filterMode = SearchModeNone

    // Find the previously selected file in new entries
    newCursor := 0 // Default to beginning if file not found
    if selectedName != "" {
        for i, e := range entries {
            if e.Name == selectedName {
                newCursor = i
                break
            }
        }
    }

    p.cursor = newCursor

    // Adjust scroll offset to keep cursor visible
    p.adjustScrollOffset()

    // Clear marks on refresh (same as LoadDirectory)
    p.markedFiles = make(map[string]bool)

    return nil
}
```

### Update execFinishedMsg Handler

Modify the `execFinishedMsg` handler in `model.go` to use `RefreshDirectoryPreserveCursor()` instead of `LoadDirectory()`:

```go
case execFinishedMsg:
    // External command completed
    // Refresh both panes while preserving cursor position
    m.getActivePane().RefreshDirectoryPreserveCursor()
    m.getInactivePane().RefreshDirectoryPreserveCursor()

    if msg.err != nil {
        m.statusMessage = fmt.Sprintf("Command failed: %v", msg.err)
        m.isStatusError = true
        return m, statusMessageClearCmd(5 * time.Second)
    }
    return m, nil
```

### Helper Method: adjustScrollOffset

Add a helper method to ensure cursor remains visible after refresh:

```go
// adjustScrollOffset ensures cursor is visible in the pane
func (p *Pane) adjustScrollOffset() {
    visibleLines := p.height - 3 // Account for header and borders

    if p.cursor < p.scrollOffset {
        p.scrollOffset = p.cursor
    } else if p.cursor >= p.scrollOffset + visibleLines {
        p.scrollOffset = p.cursor - visibleLines + 1
    }

    if p.scrollOffset < 0 {
        p.scrollOffset = 0
    }
}
```

## Implementation Approach

### Files to Modify

| File | Changes |
|------|---------|
| `internal/ui/pane.go` | Add `RefreshDirectoryPreserveCursor()` and `adjustScrollOffset()` methods |
| `internal/ui/model.go` | Update `execFinishedMsg` handler to use `RefreshDirectoryPreserveCursor()` |
| `internal/ui/pane_test.go` | Add tests for `RefreshDirectoryPreserveCursor()` |

### Behavior Comparison

| Scenario | LoadDirectory() | RefreshDirectoryPreserveCursor() |
|----------|-----------------|----------------------------------|
| Cursor position | Reset to 0 | Preserved (by file name) |
| Selected file deleted | N/A | Reset to 0 |
| Scroll offset | Reset to 0 | Adjusted to keep cursor visible |
| Filter cleared | Yes | Yes |
| Marks cleared | Yes | Yes |

## Test Scenarios

### Unit Tests

- [ ] RefreshDirectoryPreserveCursor preserves cursor on same file
- [ ] RefreshDirectoryPreserveCursor resets cursor to 0 when selected file is deleted
- [ ] RefreshDirectoryPreserveCursor handles empty directory
- [ ] RefreshDirectoryPreserveCursor clears filter pattern
- [ ] adjustScrollOffset keeps cursor in view

### Integration Tests (Manual)

- [ ] View file with less, exit, cursor stays on same file
- [ ] Edit file with vim, exit, cursor stays on same file
- [ ] Delete selected file externally, exit, cursor resets to beginning
- [ ] Both panes maintain their cursor positions

## Success Criteria

- [ ] Cursor position preserved after less exits
- [ ] Cursor position preserved after vim exits
- [ ] Cursor resets to beginning when selected file is deleted
- [ ] Existing LoadDirectory behavior unchanged
- [ ] All existing tests pass
- [ ] New unit tests pass

## Dependencies

None - uses existing internal packages only.

## Error Handling

| Condition | Behavior |
|-----------|----------|
| Directory read error | Return error, keep current state |
| Selected file no longer exists | Reset cursor to 0 |
| Empty directory | Set cursor to 0 |

## Security Considerations

None - no new external inputs or commands.

## Performance Considerations

Minimal impact - same directory read operation as LoadDirectory.
