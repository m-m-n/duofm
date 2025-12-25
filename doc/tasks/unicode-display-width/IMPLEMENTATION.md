# Implementation Plan: Unicode Display Width Support

## Overview

Fix display corruption for multibyte character file/directory names by replacing byte-based string operations with display-width-aware operations using `go-runewidth` library.

## Objectives

- Replace `len()` calls with `runewidth.StringWidth()` for padding calculations
- Replace byte-based string slicing with `runewidth.Truncate()` for truncation
- Ensure all strings remain valid UTF-8 after truncation

## Prerequisites

- `github.com/mattn/go-runewidth` is already in go.mod (no action needed)

## Architecture Overview

No architectural changes. This is a targeted fix to 5 specific functions across 3 files.

## Implementation Phases

### Phase 1: Core Truncation Function

**Goal**: Fix file name truncation in the fs package

**Files to Modify**:
- `internal/fs/types.go` - `DisplayNameWithLimit()` function

**Implementation Steps**:

1. Add import for `github.com/mattn/go-runewidth`

2. Modify `DisplayNameWithLimit()` (lines 44-64):
   ```go
   func (e FileEntry) DisplayNameWithLimit(maxWidth int) string {
       fullName := e.DisplayName()
       if runewidth.StringWidth(fullName) <= maxWidth {
           return fullName
       }

       // Symlink: "name -> ..." format
       if e.IsSymlink && e.LinkTarget != "" {
           prefix := e.Name + " -> "
           prefixWidth := runewidth.StringWidth(prefix)
           if prefixWidth+3 < maxWidth {
               return prefix + "..."
           }
       }

       // Normal file name truncation
       if maxWidth > 3 {
           return runewidth.Truncate(fullName, maxWidth-3, "") + "..."
       }
       return runewidth.Truncate(fullName, maxWidth, "")
   }
   ```

**Testing**:
- Add test cases in `internal/fs/types_test.go` for Japanese file names
- Test symlink with Japanese link target

**Estimated Effort**: Small

---

### Phase 2: Pane Formatting Functions

**Goal**: Fix column alignment in file list display

**Files to Modify**:
- `internal/ui/pane.go` - `formatBasicEntry()` and `formatDetailEntry()` functions

**Implementation Steps**:

1. Add import for `github.com/mattn/go-runewidth` (if not already present)

2. Modify `formatBasicEntry()` (line 664):
   ```go
   // Before:
   namePadding := nameWidth - len(name)

   // After:
   namePadding := nameWidth - runewidth.StringWidth(name)
   ```

3. Modify `formatDetailEntry()` (line 696):
   ```go
   // Before:
   namePadding := nameWidth - len(name)

   // After:
   namePadding := nameWidth - runewidth.StringWidth(name)
   ```

**Dependencies**:
- Phase 1 should be completed first (DisplayNameWithLimit fix)

**Testing**:
- Add test cases in `internal/ui/pane_test.go`
- Verify column alignment with mixed ASCII/Japanese names

**Estimated Effort**: Small

---

### Phase 3: Header Layout

**Goal**: Fix header line layout calculation

**Files to Modify**:
- `internal/ui/pane.go` - `renderHeaderLine2()` function

**Implementation Steps**:

1. Modify `renderHeaderLine2()` (lines 557-558):
   ```go
   // Before:
   markedLen := len(markedInfo)
   freeLen := len(freeInfo)

   // After:
   markedLen := runewidth.StringWidth(markedInfo)
   freeLen := runewidth.StringWidth(freeInfo)
   ```

**Note**: Currently `markedInfo` and `freeInfo` only contain ASCII characters, but this fix ensures future compatibility if localization is added.

**Testing**:
- Verify header layout is correct

**Estimated Effort**: Small

---

### Phase 4: Status Bar

**Goal**: Fix status message truncation

**Files to Modify**:
- `internal/ui/model.go` - `renderStatusBar()` function

**Implementation Steps**:

1. Add import for `github.com/mattn/go-runewidth` (if not already present)

2. Modify message truncation (lines 1067-1069):
   ```go
   // Before:
   if len(msg) > maxLen {
       msg = msg[:maxLen-3] + "..."
   }

   // After:
   if runewidth.StringWidth(msg) > maxLen {
       msg = runewidth.Truncate(msg, maxLen-3, "") + "..."
   }
   ```

3. Also fix the padding calculation (line 1086):
   ```go
   // Before:
   padding := m.width - len(posInfo) - len(hints) - 4

   // After:
   padding := m.width - runewidth.StringWidth(posInfo) - runewidth.StringWidth(hints) - 4
   ```

**Testing**:
- Test with Japanese error messages

**Estimated Effort**: Small

---

## File Structure

No new files. Modifications only:

```
internal/
├── fs/
│   └── types.go          # DisplayNameWithLimit() fix
└── ui/
    ├── pane.go           # formatBasicEntry(), formatDetailEntry(), renderHeaderLine2() fixes
    └── model.go          # renderStatusBar() fix
```

## Testing Strategy

### Unit Tests

Add test cases to existing test files:

**internal/fs/types_test.go**:
```go
func TestDisplayNameWithLimit_Japanese(t *testing.T) {
    tests := []struct {
        name     string
        entry    FileEntry
        maxWidth int
        want     string
    }{
        {
            name:     "Japanese name fits",
            entry:    FileEntry{Name: "テスト.txt"},
            maxWidth: 20,
            want:     "テスト.txt",
        },
        {
            name:     "Japanese name truncated",
            entry:    FileEntry{Name: "日本語ファイル名.txt"},
            maxWidth: 10,
            want:     "日本...",  // Truncated to fit
        },
        {
            name:     "Mixed ASCII and Japanese",
            entry:    FileEntry{Name: "file_テスト.txt"},
            maxWidth: 15,
            want:     "file_テス...",
        },
    }
    // ... test implementation
}
```

**internal/ui/pane_test.go**:
```go
func TestFormatBasicEntry_Japanese(t *testing.T) {
    // Test that columns align correctly with Japanese names
}
```

### Manual Testing Checklist

- [ ] Open directory with Japanese file names
- [ ] Verify columns are properly aligned
- [ ] Resize terminal and verify layout adapts correctly
- [ ] Create symlink with Japanese target, verify display
- [ ] Trigger error with Japanese message, verify status bar

## Dependencies

### External Libraries
- `github.com/mattn/go-runewidth` v0.0.16 - Already in go.mod

### Internal Dependencies
- Phase 2 depends on Phase 1 (uses DisplayNameWithLimit)
- Phases 3 and 4 are independent

## Risk Assessment

### Technical Risks

- **Risk: runewidth.Truncate behavior with edge cases**
  - Mitigation: Add comprehensive test cases for edge cases (empty string, all ASCII, all Japanese, mixed)

- **Risk: Performance impact for large directories**
  - Mitigation: runewidth is highly optimized; benchmark if needed

### Implementation Risks

- **Risk: Missing a `len()` call that should be replaced**
  - Mitigation: Search entire codebase for `len()` calls on strings that might contain Unicode

## Performance Considerations

- `runewidth.StringWidth()` is O(n) where n is string length, but highly optimized
- No caching needed for typical file name lengths (<256 chars)
- Performance impact should be negligible

## Security Considerations

- No security implications (display-only change)
- Truncated strings remain valid UTF-8

## Verification Steps

After implementation:

1. Run all existing tests:
   ```bash
   make test
   ```

2. Run the application and verify visually:
   ```bash
   ./duofm
   ```
   - Navigate to a directory with Japanese file names
   - Verify proper alignment and truncation

3. Test edge cases:
   - Very long Japanese file names
   - Mixed ASCII/Japanese names
   - Symlinks with Japanese targets
   - Narrow terminal width

## References

- [Specification](./SPEC.md)
- [Requirements (Japanese)](./要件定義書.md)
- [go-runewidth documentation](https://pkg.go.dev/github.com/mattn/go-runewidth)
