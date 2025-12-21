# Verification Checklist: Search and Filter

## Overview

This document contains verification items for the Search and Filter feature implementation.

## Automated Verification

### Build Status

- [ ] `go build ./...` completes without errors
- [ ] `go build -o /tmp/duofm ./cmd/duofm` produces binary

### Test Execution

- [ ] `go test ./...` all tests pass
- [ ] `go test -cover ./internal/ui/...` coverage meets threshold

### Code Quality

- [ ] `gofmt -l .` returns no unformatted files
- [ ] `go vet ./...` reports no issues

### File Structure

#### Created Files
- [ ] `internal/ui/search.go` - Search types and filter functions
- [ ] `internal/ui/search_test.go` - Search unit tests
- [ ] `internal/ui/minibuffer.go` - Minibuffer component
- [ ] `internal/ui/minibuffer_test.go` - Minibuffer tests

#### Modified Files
- [ ] `internal/ui/keys.go` - Added KeySearch, KeyRegexSearch
- [ ] `internal/ui/pane.go` - Added filter integration, ViewWithMinibuffer
- [ ] `internal/ui/pane_test.go` - Added filter tests
- [ ] `internal/ui/model.go` - Added search state, key handling
- [ ] `internal/ui/model_test.go` - Added search integration tests

## SPEC.md Compliance

### Success Criteria

- [ ] Both search modes (`/` and `Ctrl+F`) work correctly
- [ ] State transitions follow specification exactly
- [ ] Smart case sensitivity works for both modes
- [ ] Minibuffer displays correctly at bottom of pane
- [ ] Filter results persist until explicitly cleared
- [ ] `Esc` restores previous search state
- [ ] Empty `Enter` clears all filters
- [ ] No performance degradation with typical directory sizes
- [ ] All existing functionality remains unaffected
- [ ] All unit tests pass

## Manual Testing Checklist

### Incremental Search (`/`)

1. [ ] Press `/` to open minibuffer with prompt `/: `
2. [ ] Type characters and see list filter in real-time
3. [ ] Press `Enter` to confirm - minibuffer closes, filter persists
4. [ ] Press `Enter` with empty input - restores full list
5. [ ] Press `Esc` to cancel - restores previous state
6. [ ] Test smart case: "abc" matches "ABC", "Abc" only matches exact case

### Regex Search (`Ctrl+F`)

7. [ ] Press `Ctrl+F` to open minibuffer with prompt `(search): `
8. [ ] Type pattern, filter NOT applied until Enter
9. [ ] Press `Enter` - regex filter applied
10. [ ] Test pattern `.*\.go$` matches .go files
11. [ ] Invalid regex (e.g., `[invalid`) shows error, does not crash
12. [ ] Press `Esc` to cancel - restores previous state

### State Transitions

13. [ ] Normal -> `/` -> type -> `Enter` -> filtered list shown
14. [ ] Filtered state -> `/` -> type new pattern -> `Enter` -> new filter
15. [ ] Filtered state -> `Esc` -> previous filter restored
16. [ ] Filtered state -> empty `Enter` -> clear all filters
17. [ ] Switch between `/` and `Ctrl+F` modes correctly

### Edge Cases

18. [ ] Filter with no matches: "(No matches)" displayed
19. [ ] Very long pattern: handled without UI break
20. [ ] Unicode file names: filtered correctly
21. [ ] Pane switch during search: search cancelled
22. [ ] Directory change: filter cleared automatically

### Visual Verification

23. [ ] Minibuffer appears at bottom of active pane
24. [ ] Minibuffer has correct prompt (`/: ` or `(search): `)
25. [ ] Cursor visible in minibuffer
26. [ ] Filter indicator `[/pattern]` shown in header when filtered
27. [ ] File list adjusts height when minibuffer visible

### Keyboard Navigation in Minibuffer

28. [ ] Type characters: inserted at cursor
29. [ ] Backspace: delete before cursor
30. [ ] Left/Right arrows: move cursor
31. [ ] Ctrl+A: move to beginning
32. [ ] Ctrl+E: move to end
33. [ ] Ctrl+K: kill to end of line
34. [ ] Ctrl+U: kill to beginning of line

## Performance Verification

35. [ ] Directory with 100 files: instant filtering
36. [ ] Directory with 1000+ files: filter response < 100ms
37. [ ] No visible lag during typing

## References

- [SPEC.md](./SPEC.md) - Feature specification
- [IMPLEMENTATION.md](./IMPLEMENTATION.md) - Implementation plan
