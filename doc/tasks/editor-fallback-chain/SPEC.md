# Feature: Editor Fallback Chain

## Overview

Enhance the `getEditor()` function to support a three-level fallback chain: `$EDITOR` → `vim` → `vi`. Currently, the function falls back to `vim` when `$EDITOR` is not set, but does not handle the case where `vim` is unavailable. This change uses `exec.LookPath` to verify command existence before selecting it.

## Objectives

- Support environments where `vim` is not installed but `vi` is available
- Use `exec.LookPath` to verify editor command existence at runtime
- Maintain backward compatibility with existing `$EDITOR` behavior

## User Stories

### US1: Editor Fallback
As a user without `$EDITOR` set, I want duofm to find an available editor (`vim` first, then `vi`), so that the `e` key always works if any vi-compatible editor is installed.

**Acceptance Criteria:**
- [ ] When `$EDITOR` is set, use it directly (no LookPath check)
- [ ] When `$EDITOR` is not set, try `vim` via `exec.LookPath`
- [ ] If `vim` is not found, try `vi` via `exec.LookPath`
- [ ] If neither `vim` nor `vi` is found, return `"vi"` as last resort (let the OS report the error)

## Technical Requirements

### Functional Requirements
- **FR1:** `getEditor()` returns `$EDITOR` value when the environment variable is set and non-empty
- **FR2:** `getEditor()` returns `"vim"` when `$EDITOR` is unset/empty and `vim` is found via `exec.LookPath`
- **FR3:** `getEditor()` returns `"vi"` when `$EDITOR` is unset/empty and `vim` is NOT found but `vi` is found via `exec.LookPath`
- **FR4:** `getEditor()` returns `"vi"` as final fallback when neither `vim` nor `vi` is found (error will surface at execution time)

### Non-Functional Requirements
- **NFR1 - Compatibility:** No behavior change for users with `$EDITOR` set
- **NFR2 - Performance:** `exec.LookPath` is called only when `$EDITOR` is not set; negligible overhead

## Implementation Approach

### File Structure

```
internal/ui/
├── exec.go           # Modify getEditor() function
└── exec_test.go      # Update/add test cases for fallback chain
```

### Current Implementation

```go
func getEditor() string {
    editor := os.Getenv("EDITOR")
    if editor == "" {
        return "vim"
    }
    return editor
}
```

### New Implementation

```go
func getEditor() string {
    editor := os.Getenv("EDITOR")
    if editor != "" {
        return editor
    }
    if _, err := exec.LookPath("vim"); err == nil {
        return "vim"
    }
    return "vi"
}
```

### Dependencies

**Internal Dependencies:**
- `openWithEditor()` in `exec.go` - calls `getEditor()`, no changes needed

**External Dependencies:**
- `os/exec` (already imported) - `exec.LookPath` for command lookup

## Test Scenarios

### Unit Tests
- [ ] `$EDITOR` set to "nano" → returns "nano"
- [ ] `$EDITOR` set to "emacs" → returns "emacs"
- [ ] `$EDITOR` not set, `vim` available → returns "vim"
- [ ] `$EDITOR` not set, `vim` unavailable, `vi` available → returns "vi"
- [ ] `$EDITOR` set to empty string, `vim` available → returns "vim"
- [ ] `$EDITOR` set to empty string, `vim` unavailable → returns "vi"

### Edge Cases
- [ ] `$EDITOR` with spaces (e.g., "vim -u NONE") → returns as-is (existing behavior)
- [ ] Neither `vim` nor `vi` found → returns "vi" (deferred error)

## Error Handling

| Condition | Behavior |
|-----------|----------|
| `$EDITOR` set | Use directly, no validation |
| `vim` not in PATH | Fall back to `vi` |
| `vi` not in PATH | Return `"vi"`, error at exec time |

## Success Criteria

- [ ] All existing `getEditor` tests pass
- [ ] New fallback test cases pass
- [ ] `$EDITOR` users experience no behavior change
- [ ] `e` key works on systems with only `vi` installed

## Open Questions

None - all requirements have been clarified.
