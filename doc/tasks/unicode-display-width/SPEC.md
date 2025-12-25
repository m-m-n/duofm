# Feature: Unicode Display Width Support

## Overview

Fix display corruption when file/directory names contain multibyte characters (Japanese, Chinese, Korean, emoji, etc.) by implementing proper display width calculation.

## Objectives

- Correct file name truncation using rune-based slicing instead of byte-based slicing
- Fix column alignment in file list by using display width instead of byte length
- Ensure all truncated strings remain valid UTF-8

## User Stories

- As a user with Japanese file names, I want to see file lists properly aligned so that I can read file information clearly
- As a user, I want file names to be truncated correctly without garbled characters when they exceed the column width

## Technical Requirements

### TR-1: Display Width Calculation

Use `github.com/mattn/go-runewidth` library (already in dependencies) for accurate display width calculation.

Key functions:
- `runewidth.StringWidth(s string) int` - Returns display width of string
- `runewidth.Truncate(s string, w int, tail string) string` - Truncates string to fit display width

### TR-2: Files to Modify

| File | Function | Current Issue | Fix |
|------|----------|---------------|-----|
| `internal/fs/types.go` | `DisplayNameWithLimit()` | Byte-based slicing | Use `runewidth.Truncate()` |
| `internal/ui/pane.go` | `formatBasicEntry()` | `len()` for padding | Use `runewidth.StringWidth()` |
| `internal/ui/pane.go` | `formatDetailEntry()` | `len()` for padding | Use `runewidth.StringWidth()` |
| `internal/ui/pane.go` | `renderHeaderLine2()` | `len()` for layout | Use `runewidth.StringWidth()` |
| `internal/ui/model.go` | `renderStatusBar()` | Byte-based truncation | Use `runewidth.Truncate()` |

## Implementation Approach

### Architecture

No architectural changes required. This is a targeted fix to specific functions.

### Code Changes

#### 1. internal/fs/types.go - DisplayNameWithLimit()

```go
import "github.com/mattn/go-runewidth"

func (e FileEntry) DisplayNameWithLimit(maxWidth int) string {
    fullName := e.DisplayName()
    if runewidth.StringWidth(fullName) <= maxWidth {
        return fullName
    }

    // Handle symlink truncation
    if e.IsSymlink && e.LinkTarget != "" {
        prefix := e.Name + " -> "
        prefixWidth := runewidth.StringWidth(prefix)
        if prefixWidth+3 <= maxWidth {
            return prefix + "..."
        }
    }

    // Use runewidth.Truncate for proper truncation
    if maxWidth > 3 {
        return runewidth.Truncate(fullName, maxWidth-3, "") + "..."
    }
    return runewidth.Truncate(fullName, maxWidth, "")
}
```

#### 2. internal/ui/pane.go - formatBasicEntry()

Replace:
```go
namePadding := nameWidth - len(name)
```

With:
```go
namePadding := nameWidth - runewidth.StringWidth(name)
```

#### 3. internal/ui/pane.go - formatDetailEntry()

Same pattern as formatBasicEntry().

#### 4. internal/ui/pane.go - renderHeaderLine2()

Replace all `len()` calls for layout calculation with `runewidth.StringWidth()`.

#### 5. internal/ui/model.go - renderStatusBar()

Replace byte-based truncation with:
```go
if runewidth.StringWidth(msg) > maxLen {
    msg = runewidth.Truncate(msg, maxLen-3, "") + "..."
}
```

### Dependencies

- `github.com/mattn/go-runewidth` - Already in go.mod, no changes needed

## Test Scenarios

- [ ] Japanese file name displays correctly without truncation
- [ ] Japanese file name truncates correctly with "..." suffix
- [ ] Mixed ASCII and Japanese file names align properly in columns
- [ ] Symlink with Japanese target displays correctly
- [ ] Status bar with Japanese error message truncates correctly
- [ ] Performance: Directory with 1000+ files renders without noticeable delay
- [ ] Existing ASCII-only file name tests continue to pass

## Success Criteria

- [ ] All existing tests pass
- [ ] New tests for Japanese file names pass
- [ ] Visual verification: File list columns are properly aligned with Japanese file names
- [ ] Visual verification: Truncated Japanese file names end with "..." and no garbled characters
- [ ] No performance regression in file list rendering

## Open Questions

None - the approach is clear and the library is already available.
