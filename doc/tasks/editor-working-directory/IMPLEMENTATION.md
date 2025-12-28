# Implementation Plan: Editor/Viewer Working Directory

## Overview

Modify the external application launch functions (`openWithViewer`, `openWithEditor`) to:
1. Set the working directory to the active pane's directory
2. Use environment variables ($EDITOR, $PAGER) for application selection

## Objectives

- Set working directory for external applications (consistent with `!` key behavior)
- Respect $EDITOR and $PAGER environment variables
- Maintain backward compatibility with existing behavior

## Prerequisites

- Understanding of `tea.ExecProcess` and `exec.Command`
- Existing implementation in `internal/ui/exec.go`
- Reference implementation: `executeShellCommand` function (already uses `cmd.Dir`)

## Architecture Overview

The change is localized to `internal/ui/exec.go` and its callers in `internal/ui/model.go`. The modification follows the existing pattern used by `executeShellCommand`.

```
internal/ui/
├── exec.go      # Modify openWithViewer, openWithEditor
├── exec_test.go # Add tests for workDir and env vars
└── model.go     # Update call sites to pass workDir
```

## Implementation Phases

### Phase 1: Modify exec.go Functions

**Goal**: Add working directory and environment variable support to `openWithViewer` and `openWithEditor`

**Files to Modify**:
- `internal/ui/exec.go` - Modify function signatures and implementation

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| `openWithViewer` | Launch file viewer | File path and workDir provided | Viewer runs with correct workDir |
| `openWithEditor` | Launch file editor | File path and workDir provided | Editor runs with correct workDir |
| `getEditor` | Get editor command | None | Returns $EDITOR or "vim" |
| `getPager` | Get pager command | None | Returns $PAGER or "less" |

**Processing Flow**:
```
User presses v/e/Enter on file
       ↓
model.go calls openWithViewer/openWithEditor(filePath, workDir)
       ↓
exec.go: Get command from $EDITOR/$PAGER or fallback
       ↓
exec.go: Create exec.Command with command and filePath
       ↓
exec.go: Set cmd.Dir = workDir
       ↓
exec.go: Return tea.ExecProcess(cmd, callback)
```

**Implementation Steps**:

1. Add helper functions for environment variable lookup
   - `getEditor()` returns $EDITOR or "vim"
   - `getPager()` returns $PAGER or "less"
   - Handle empty string case (treat as unset)

2. Modify `openWithViewer` signature
   - Add `workDir string` parameter
   - Use `getPager()` instead of hardcoded "less"
   - Set `cmd.Dir = workDir`

3. Modify `openWithEditor` signature
   - Add `workDir string` parameter
   - Use `getEditor()` instead of hardcoded "vim"
   - Set `cmd.Dir = workDir`

**Dependencies**:
- Go `os` package for `os.Getenv`

**Testing**:
- Unit tests for `getEditor()` and `getPager()` with various env var states
- Verify command has correct `Dir` field set

**Estimated Effort**: Small

---

### Phase 2: Update model.go Call Sites

**Goal**: Pass the active pane's directory to the modified functions

**Files to Modify**:
- `internal/ui/model.go` - Update all call sites

**Key Components**:

| Component | Responsibility | Precondition | Postcondition |
|-----------|----------------|--------------|---------------|
| KeyView handler | Handle `v` key | File selected | Viewer launched with correct workDir |
| KeyEdit handler | Handle `e` key | File selected | Editor launched with correct workDir |
| KeyEnter handler | Handle `Enter` on file | File selected | Viewer launched with correct workDir |

**Call Sites to Update** (3 locations):

1. `KeyView` handler (~line 777)
   - Currently: `openWithViewer(fullPath)`
   - Change to: `openWithViewer(fullPath, m.getActivePane().Path())`

2. `KeyEdit` handler (~line 791)
   - Currently: `openWithEditor(fullPath)`
   - Change to: `openWithEditor(fullPath, m.getActivePane().Path())`

3. `KeyEnter` handler for files (~line 653)
   - Currently: `openWithViewer(fullPath)`
   - Change to: `openWithViewer(fullPath, m.getActivePane().Path())`

**Implementation Steps**:

1. Locate all `openWithViewer` calls in model.go
2. Add second argument: `m.getActivePane().Path()`
3. Locate all `openWithEditor` calls in model.go
4. Add second argument: `m.getActivePane().Path()`

**Dependencies**:
- Phase 1 must be completed first

**Testing**:
- Integration test: verify external app receives correct working directory
- Manual testing: vim `:!pwd`, `:e .`

**Estimated Effort**: Small

---

### Phase 3: Add Comprehensive Tests

**Goal**: Ensure all functionality is properly tested

**Files to Modify**:
- `internal/ui/exec_test.go` - Add new test cases

**Test Cases**:

1. Environment variable tests:
   - `TestGetEditor_WithEnvVar` - $EDITOR set
   - `TestGetEditor_WithoutEnvVar` - $EDITOR not set
   - `TestGetEditor_EmptyEnvVar` - $EDITOR=""
   - `TestGetPager_WithEnvVar` - $PAGER set
   - `TestGetPager_WithoutEnvVar` - $PAGER not set
   - `TestGetPager_EmptyEnvVar` - $PAGER=""

2. Working directory tests:
   - `TestOpenWithViewer_WorkDir` - Verify workDir is passed
   - `TestOpenWithEditor_WorkDir` - Verify workDir is passed

**Implementation Steps**:

1. Write tests for `getEditor()` helper
   - Test with env var set
   - Test with env var unset
   - Test with empty env var

2. Write tests for `getPager()` helper
   - Same pattern as getEditor tests

3. Update existing `TestOpenWithViewerReturnsCmd` and `TestOpenWithEditorReturnsCmd`
   - Add workDir parameter to calls

**Dependencies**:
- Phase 1 must be completed first

**Testing**:
- Run `go test ./internal/ui/... -v`

**Estimated Effort**: Small

---

## File Structure

```
internal/ui/
├── exec.go           # Modified: add workDir param, env var lookup
├── exec_test.go      # Modified: add tests for new functionality
├── model.go          # Modified: pass workDir to exec functions
└── model_test.go     # Existing tests should still pass
```

## Testing Strategy

### Unit Tests

| Test | Description |
|------|-------------|
| `TestGetEditor_*` | Verify correct editor is returned based on $EDITOR |
| `TestGetPager_*` | Verify correct pager is returned based on $PAGER |
| `TestOpenWithViewer_WorkDir` | Verify working directory is set correctly |
| `TestOpenWithEditor_WorkDir` | Verify working directory is set correctly |

### Integration Tests

- Existing integration tests should pass without modification
- New manual testing scenarios for working directory verification

### Manual Testing Checklist

- [ ] Press `v` on a file, run `!pwd` in less - shows file's directory
- [ ] Press `e` on a file, run `:!pwd` in vim - shows file's directory
- [ ] Press `e` on a file, run `:e .` in vim - shows file's directory contents
- [ ] Press `Enter` on a file, run `!pwd` in less - shows file's directory
- [ ] Set $EDITOR=nano, press `e` - opens nano
- [ ] Set $PAGER=cat, press `v` - runs cat
- [ ] Unset $EDITOR, press `e` - opens vim (fallback)
- [ ] Unset $PAGER, press `v` - opens less (fallback)
- [ ] Set $EDITOR="", press `e` - opens vim (empty treated as unset)
- [ ] Set $PAGER="", press `v` - opens less (empty treated as unset)

## Dependencies

### External Libraries

None - uses only Go standard library and existing Bubble Tea framework

### Internal Dependencies

| Dependency | Reason |
|------------|--------|
| Phase 1 → Phase 2 | model.go needs updated function signatures |
| Phase 1 → Phase 3 | Tests need to test new functionality |

## Risk Assessment

### Technical Risks

- **$EDITOR/$PAGER with arguments**: Some users may have `EDITOR="vim -u NONE"`. This requires shell expansion.
  - Mitigation: Use `/bin/sh -c "$EDITOR $filepath"` pattern if needed, or document limitation
  - Alternative: Parse first word only as command (current approach)

### Implementation Risks

- **Breaking existing tests**: Function signature changes may break tests
  - Mitigation: Update tests in Phase 3

## Performance Considerations

- `os.Getenv` is called once per external app launch - negligible overhead
- No additional file I/O or network calls

## Security Considerations

- Environment variables are user-controlled - acceptable as users explicitly set them
- File paths are already validated before reaching these functions
- Working directory is always the pane's current path (user-navigated)

## Open Questions

None - all requirements are clear from the specification.

## References

- Specification: `doc/tasks/editor-working-directory/SPEC.md`
- Reference implementation: `executeShellCommand` in `internal/ui/exec.go`
- Bubble Tea ExecProcess: https://github.com/charmbracelet/bubbletea
