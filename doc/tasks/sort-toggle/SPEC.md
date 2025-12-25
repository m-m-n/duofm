# Feature: Sort Toggle

## Overview

Add sort toggle functionality to duofm. Pressing `s` opens a sort settings dialog where users can select the sort field (Name/Size/Date) and order (Ascending/Descending) using a two-row selection interface.

## Objectives

- Allow users to sort files by name, size, or modification date
- Support both ascending and descending order for each sort type
- Maintain independent sort settings per pane
- Preserve cursor position on the same file after sort change
- Provide immediate visual feedback with live preview

## Key Binding

| Key | Action |
|-----|--------|
| `s` | Open sort settings dialog |

## Dialog Design

```
┌─ Sort ─────────────────────────┐
│                                │
│  Sort by    [Name]  Size  Date │
│  Order       ↑Asc  [↓Desc]     │
│                                │
│  h/l:change  j/k:row           │
│  Enter:OK  Esc:cancel          │
└────────────────────────────────┘
```

### Dialog Structure

| Row | Options | Description |
|-----|---------|-------------|
| Sort by | Name, Size, Date | Field to sort by |
| Order | ↑Asc, ↓Desc | Sort direction |

### Key Bindings in Dialog

| Key | Action |
|-----|--------|
| `h` / `←` | Move to left option |
| `l` / `→` | Move to right option |
| `k` / `↑` | Move to upper row (Order → Sort by) |
| `j` / `↓` | Move to lower row (Sort by → Order) |
| `Enter` | Confirm and close dialog |
| `Esc` / `q` | Cancel and restore previous sort |

### Visual Indicators

- **Selected option**: Enclosed in `[ ]` brackets
- **Focused row**: Highlighted background
- **Help text**: Two lines at bottom
  - Line 1: Navigation controls (h/l:change j/k:row)
  - Line 2: Dialog actions (Enter:OK Esc:cancel)

## Sort Options

### Sort Fields

| Field | Description |
|-------|-------------|
| Name | Sort by file/directory name |
| Size | Sort by file size in bytes |
| Date | Sort by modification time |

### Sort Orders

| Order | Display | Description |
|-------|---------|-------------|
| Ascending | ↑Asc | A→Z, Small→Large, Old→New |
| Descending | ↓Desc | Z→A, Large→Small, New→Old |

**Default**: Name + ↑Asc (Name ascending)

## Technical Requirements

### TR-1: Sort Configuration Types

```go
// internal/ui/sort.go

type SortField int

const (
    SortByName SortField = iota
    SortBySize
    SortByDate
)

type SortOrder int

const (
    SortAsc SortOrder = iota
    SortDesc
)

type SortConfig struct {
    Field SortField
    Order SortOrder
}

func (c SortConfig) String() string {
    fields := []string{"Name", "Size", "Date"}
    orders := []string{"↑", "↓"}
    return fmt.Sprintf("%s %s", fields[c.Field], orders[c.Order])
}
```

### TR-2: Sort Dialog Component

```go
// internal/ui/sort_dialog.go

type SortDialog struct {
    config       SortConfig  // Current selection
    originalConfig SortConfig // For cancel/restore
    focusedRow   int         // 0: Sort by, 1: Order
    visible      bool
}

func NewSortDialog(current SortConfig) *SortDialog {
    return &SortDialog{
        config:         current,
        originalConfig: current,
        focusedRow:     0,
        visible:        true,
    }
}

func (d *SortDialog) HandleKey(key string) (confirmed bool, cancelled bool) {
    switch key {
    case "h", "left":
        d.moveLeft()
    case "l", "right":
        d.moveRight()
    case "j", "down":
        d.focusedRow = 1
    case "k", "up":
        d.focusedRow = 0
    case "enter":
        return true, false
    case "esc", "q":
        d.config = d.originalConfig
        return false, true
    }
    return false, false
}

func (d *SortDialog) moveLeft() {
    if d.focusedRow == 0 {
        // Sort by: Name <- Size <- Date
        if d.config.Field > SortByName {
            d.config.Field--
        }
    } else {
        // Order: Asc <- Desc
        d.config.Order = SortAsc
    }
}

func (d *SortDialog) moveRight() {
    if d.focusedRow == 0 {
        // Sort by: Name -> Size -> Date
        if d.config.Field < SortByDate {
            d.config.Field++
        }
    } else {
        // Order: Asc -> Desc
        d.config.Order = SortDesc
    }
}
```

### TR-3: Pane Sort State

```go
// internal/ui/pane.go

type Pane struct {
    // ... existing fields
    sortConfig SortConfig  // Default: {SortByName, SortAsc}
}

func (p *Pane) ApplySort() {
    p.entries = SortEntries(p.entries, p.sortConfig)
}
```

### TR-4: Sort Function with Directory-First Ordering

```go
// internal/ui/sort.go

func SortEntries(entries []fs.FileEntry, config SortConfig) []fs.FileEntry {
    // Separate parent dir, directories, and files
    var parentDir []fs.FileEntry
    var dirs []fs.FileEntry
    var files []fs.FileEntry

    for _, e := range entries {
        if e.Name == ".." {
            parentDir = append(parentDir, e)
        } else if e.IsDir {
            dirs = append(dirs, e)
        } else {
            files = append(files, e)
        }
    }

    // Get comparison function
    less := getLessFunc(config)

    // Sort directories and files separately
    sort.SliceStable(dirs, func(i, j int) bool {
        return less(dirs[i], dirs[j])
    })
    sort.SliceStable(files, func(i, j int) bool {
        return less(files[i], files[j])
    })

    // Combine: parent dir first, then dirs, then files
    result := make([]fs.FileEntry, 0, len(entries))
    result = append(result, parentDir...)
    result = append(result, dirs...)
    result = append(result, files...)
    return result
}

func getLessFunc(config SortConfig) func(a, b fs.FileEntry) bool {
    switch config.Field {
    case SortByName:
        if config.Order == SortAsc {
            return func(a, b fs.FileEntry) bool { return a.Name < b.Name }
        }
        return func(a, b fs.FileEntry) bool { return a.Name > b.Name }
    case SortBySize:
        if config.Order == SortAsc {
            return func(a, b fs.FileEntry) bool { return a.Size < b.Size }
        }
        return func(a, b fs.FileEntry) bool { return a.Size > b.Size }
    case SortByDate:
        if config.Order == SortAsc {
            return func(a, b fs.FileEntry) bool { return a.ModTime.Before(b.ModTime) }
        }
        return func(a, b fs.FileEntry) bool { return a.ModTime.After(b.ModTime) }
    default:
        return func(a, b fs.FileEntry) bool { return a.Name < b.Name }
    }
}
```

### TR-5: Cursor Position Preservation

```go
func (p *Pane) ApplySortAndPreserveCursor() {
    // Remember current file name
    currentName := ""
    if p.cursor >= 0 && p.cursor < len(p.entries) {
        currentName = p.entries[p.cursor].Name
    }

    // Apply sort
    p.entries = SortEntries(p.entries, p.sortConfig)

    // Find and restore cursor position
    if currentName != "" {
        for i, e := range p.entries {
            if e.Name == currentName {
                p.cursor = i
                p.adjustScroll()
                return
            }
        }
    }
    // If not found, keep current index (bounded)
    if p.cursor >= len(p.entries) {
        p.cursor = max(0, len(p.entries)-1)
    }
    p.adjustScroll()
}
```

### TR-6: Live Preview

When user changes selection in dialog, immediately apply sort to pane:

```go
// internal/ui/model.go

func (m *Model) handleSortDialogKey(key string) tea.Cmd {
    confirmed, cancelled := m.sortDialog.HandleKey(key)

    if confirmed || cancelled {
        m.sortDialog = nil
        if cancelled {
            // Restore original sort
            m.ActivePane().ApplySortAndPreserveCursor()
        }
        return nil
    }

    // Live preview: apply current selection
    m.ActivePane().sortConfig = m.sortDialog.config
    m.ActivePane().ApplySortAndPreserveCursor()

    return nil
}
```

## Implementation Files

| File | Changes |
|------|---------|
| `internal/ui/sort.go` | New: SortField, SortOrder, SortConfig types, SortEntries function |
| `internal/ui/sort_test.go` | New: Unit tests for sorting |
| `internal/ui/sort_dialog.go` | New: SortDialog component |
| `internal/ui/sort_dialog_test.go` | New: Unit tests for dialog |
| `internal/ui/pane.go` | Add sortConfig field, ApplySortAndPreserveCursor method |
| `internal/ui/keys.go` | Add KeySort constant |
| `internal/ui/model.go` | Add sort dialog handling, key handler for 's' |
| `internal/ui/view.go` | Add sort dialog rendering |

## Test Scenarios

### Unit Tests (sort.go)

- [ ] SortEntries sorts by name ascending correctly
- [ ] SortEntries sorts by name descending correctly
- [ ] SortEntries sorts by size ascending correctly
- [ ] SortEntries sorts by size descending correctly
- [ ] SortEntries sorts by date ascending correctly
- [ ] SortEntries sorts by date descending correctly
- [ ] SortEntries keeps parent directory (..) at top
- [ ] SortEntries keeps directories before files
- [ ] SortConfig.String() returns correct display string

### Unit Tests (sort_dialog.go)

- [ ] HandleKey 'h'/'l' changes field selection
- [ ] HandleKey 'j'/'k' changes focused row
- [ ] HandleKey 'enter' returns confirmed=true
- [ ] HandleKey 'esc' restores original config and returns cancelled=true
- [ ] HandleKey 'q' behaves same as 'esc'
- [ ] Arrow keys work same as hjkl
- [ ] Field selection wraps correctly (stays at bounds)

### Integration Tests

- [ ] Cursor position preserved after sort change
- [ ] Live preview updates file list during selection
- [ ] Cancel restores original sort order

### E2E Tests

- [ ] Press `s` opens sort dialog
- [ ] Dialog shows current sort settings
- [ ] Navigation with hjkl and arrow keys works
- [ ] Enter confirms and closes dialog
- [ ] Esc cancels and restores previous sort
- [ ] Sort is independent between left and right panes
- [ ] Sort persists after directory navigation
- [ ] Sort resets after app restart

## Success Criteria

- [ ] `s` key opens sort dialog
- [ ] Dialog displays two-row selection interface
- [ ] hjkl and arrow keys navigate dialog
- [ ] Enter confirms, Esc cancels
- [ ] Live preview shows sort changes immediately
- [ ] Directories always appear before files
- [ ] Parent directory (..) always at top
- [ ] Cursor position preserved on same file
- [ ] Each pane maintains independent sort setting
- [ ] All existing features work correctly
- [ ] Unit tests pass
- [ ] E2E tests pass

## Dependencies

- Existing dialog rendering infrastructure
- `fs.FileEntry` with Name, Size, ModTime, IsDir fields
- Existing pane structure

## Notes

- Sort settings are not persisted (reset on app restart)
- Dialog width should accommodate all options on one line
- Consider case-insensitive name sorting in future
