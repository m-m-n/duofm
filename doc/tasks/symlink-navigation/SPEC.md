# Symlink Navigation Specification

## Overview

Clarify and correct the symlink navigation behavior in duofm. This specification addresses ambiguity in existing documentation and fixes incorrect behavior description for the "Open link target" context menu action.

## Objectives

1. Define clear terminology for "logical path" and "physical path" navigation
2. Correct the "Open link target" specification to navigate to the link target itself (not parent directory)
3. Ensure consistency between documentation and implementation
4. Support chained symlinks with single-step resolution

## Terminology

| Term | Definition | Example (`/bin -> /usr/bin`) |
|------|------------|------------------------------|
| **Logical Path** | Navigation that preserves the symlink path | Enter `/bin`, `..` returns to `/` |
| **Physical Path** | Navigation to the actual target location | Navigate to `/usr/bin`, `..` returns to `/usr` |

## User Stories

### US1: Developer Navigating Project Symlinks
As a developer working with symlinked project directories, I want to navigate using logical paths so that I maintain my mental model of the directory structure.

### US2: System Administrator Checking Link Targets
As a system administrator, I want to quickly navigate to the physical location of a symlink so that I can inspect or modify the actual files.

### US3: User Understanding Symlink Chains
As a user encountering chained symlinks, I want to follow them step by step so that I can understand the link structure.

## Technical Requirements

### FR1: Enter Key - Logical Path Navigation

When pressing Enter on a symlink directory:

- **FR1.1**: Store the symlink path as the current directory
  - Example: `/bin -> /usr/bin`, Enter → current directory is `/bin`
- **FR1.2**: Navigate to logical parent when selecting `..`
  - Example: From `/bin`, `..` → navigate to `/`
- **FR1.3**: Display contents of the link target directory
  - Example: Inside `/bin`, display contents of `/usr/bin`

```
Navigation Flow:
/bin -> /usr/bin

1. Select /bin and press Enter
2. Current directory: /bin (displaying /usr/bin contents)
3. Select .. and press Enter
4. Current directory: / (logical parent of /bin)
```

**Current Implementation Status**: Fixed - `pane.go:EnterDirectory` now uses logical path (`filepath.Join(p.path, entry.Name)`) instead of physical path (`entry.LinkTarget`).

### FR2: Context Menu "Enter as directory (logical)"

- **FR2.1**: Identical behavior to Enter key
- **FR2.2**: Display only for symlink directories

**Current Implementation Status**: Working correctly - no changes needed.

### FR3: Context Menu "Open link target (physical)"

When selecting "Open link target" on a symlink:

- **FR3.1**: Navigate directly to the link target path
  - Example: `/bin -> /usr/bin`, execute → current directory is `/usr/bin`
- **FR3.2**: After navigation, treat as regular directory
  - Example: From `/usr/bin`, `..` → navigate to `/usr`
- **FR3.3**: Disable (gray out) when link is broken
- **FR3.4**: Display only for symlinks

```
Navigation Flow:
/bin -> /usr/bin

1. Select /bin and press @ for context menu
2. Select "Open link target (physical)"
3. Current directory: /usr/bin (the link target itself)
4. Select .. and press Enter
5. Current directory: /usr (physical parent)
```

**Current Implementation Status**: Fixed - `context_menu_dialog.go:enter_physical` now navigates directly to the link target itself.

### FR4: Chained Symlinks

For multi-level symlinks (e.g., `link1 -> link2 -> actual_dir`):

- **FR4.1**: Display only first-level target (`link2`)
- **FR4.2**: Enter (logical path) preserves current link's path
- **FR4.3**: Open link target (physical path) follows one level only (to `link2`)
- **FR4.4**: User repeats operation to reach final target

### FR5: Link Target Display Format

- **FR5.1**: Display link targets as absolute paths
  - Example: `link -> ../foo` displayed as `link -> /absolute/path/to/foo`
- **FR5.2**: Convert relative paths to absolute (current behavior)

**Current Implementation Status**: Working correctly - no changes needed.

### FR6: File Symlinks (Future)

- **FR6.1**: Open link target opens the file (editor, etc.)
- **FR6.2**: Out of scope for this specification

## Implementation Approach

### Changes Required

#### 1. Documentation Updates

**File: `doc/tasks/context-menu/SPEC.md`**

Current (line 50):
```markdown
- "Open link target (physical path)": Open parent directory of the actual file/directory
```

Change to:
```markdown
- "Open link target (physical path)": Navigate to the link target directory/file location
```

**File: `doc/tasks/context-menu/要件定義書.md`**

Current (line 54):
```markdown
- **「リンク先を開く（物理パス）」**: リンク先の実体が存在するディレクトリを開く
```

Change to:
```markdown
- **「リンク先を開く（物理パス）」**: リンク先のディレクトリ/ファイルに直接移動する
```

#### 2. Implementation Updates

**File: `internal/ui/context_menu_dialog.go`**

Verify and fix the "enter_physical" action handler to:
1. Get the symlink's `LinkTarget` field
2. Navigate directly to that path (not its parent)

```go
// Expected behavior for enter_physical action
case "enter_physical":
    if entry.IsSymlink && !entry.LinkBroken {
        // Navigate to the link target itself
        pane.ChangeDirectory(entry.LinkTarget)
    }
```

### Files to Modify

| File | Change Type | Description |
|------|-------------|-------------|
| `doc/tasks/context-menu/SPEC.md` | Documentation | Fix FR3.1 description |
| `doc/tasks/context-menu/要件定義書.md` | Documentation | Fix FR3.1 description |
| `internal/ui/context_menu_dialog.go` | Code | Fix enter_physical action |

### Files Not Changed

| File | Reason |
|------|--------|
| `internal/ui/pane.go` | Enter key logic is correct |
| `internal/fs/symlink.go` | Symlink resolution is correct |

## Test Scenarios

### TS1: Logical Path Navigation

```
Precondition: /tmp/test-link -> /usr/share/doc
Steps:
  1. Navigate to /tmp
  2. Select test-link and press Enter
  3. Verify current directory is /tmp/test-link
  4. Select .. and press Enter
  5. Verify current directory is /tmp
Expected: Returns to logical parent directory
```

### TS2: Physical Path Navigation

```
Precondition: /tmp/test-link -> /usr/share/doc
Steps:
  1. Navigate to /tmp
  2. Select test-link and press @ for context menu
  3. Select "Open link target (physical)"
  4. Verify current directory is /usr/share/doc
  5. Select .. and press Enter
  6. Verify current directory is /usr/share
Expected: Navigates to link target, returns to physical parent
```

### TS3: Broken Link Handling

```
Precondition: /tmp/broken-link -> /nonexistent
Steps:
  1. Navigate to /tmp
  2. Select broken-link and press @ for context menu
Expected: "Open link target (physical)" is disabled/grayed out
```

### TS4: Chained Symlinks

```
Precondition:
  /tmp/link1 -> /tmp/link2
  /tmp/link2 -> /usr/share
Steps:
  1. Navigate to /tmp
  2. Select link1 and press @ for context menu
  3. Select "Open link target (physical)"
Expected: Navigates to /tmp/link2 (not /usr/share)
```

### TS5: Absolute Path Display

```
Precondition:
  cd /tmp && ln -s ../usr/share rel-link
Steps:
  1. Navigate to /tmp
  2. View rel-link entry
Expected: Displays "rel-link -> /usr/share" (absolute path)
```

## Unit Tests

### Test Cases for context_menu_dialog.go

```go
func TestEnterPhysical_NavigatesToLinkTarget(t *testing.T) {
    // Setup: symlink /tmp/link -> /usr/share
    // Action: Execute enter_physical
    // Assert: Current directory is /usr/share (not /usr)
}

func TestEnterPhysical_ChainedSymlink(t *testing.T) {
    // Setup: /tmp/link1 -> /tmp/link2 -> /usr/share
    // Action: Execute enter_physical on link1
    // Assert: Current directory is /tmp/link2
}

func TestEnterPhysical_BrokenLink(t *testing.T) {
    // Setup: /tmp/broken -> /nonexistent
    // Action: Try to execute enter_physical
    // Assert: Action is disabled/no-op
}
```

## Security Considerations

- **Path Traversal**: Use `filepath.Clean()` on resolved paths
- **Permission Checks**: Handle permission denied gracefully
- **Symlink Loops**: Rely on OS limits (typically 40 levels)

## Performance Optimization

- No caching needed for symlink resolution (instant operation)
- Single `os.Readlink()` call per symlink display
- No recursive resolution for chained symlinks

## Success Criteria

1. [ ] Enter key performs logical path navigation correctly
2. [ ] "Open link target" navigates to link target (not parent)
3. [ ] Chained symlinks follow one level at a time
4. [ ] Broken symlinks disable physical navigation option
5. [ ] All existing tests continue to pass
6. [ ] Documentation matches implementation

## Open Questions

All questions resolved through user dialogue:

- ~~Should we resolve chained symlinks to final target?~~ → No, follow one level
- ~~Display relative or absolute paths?~~ → Absolute paths (current behavior)
- ~~File symlinks behavior?~~ → Future implementation (out of scope)

## Dependencies

- Existing symlink infrastructure in `internal/fs/symlink.go`
- Context menu framework in `internal/ui/context_menu_dialog.go`
- Pane navigation in `internal/ui/pane.go`

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-21 | - | Initial specification |
