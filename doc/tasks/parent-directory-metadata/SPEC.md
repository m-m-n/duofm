# SPEC: Parent Directory Metadata Display

## Overview

Fix the bug where parent directory entry (..) displays timestamp as `0001-01-01 00:00` by fetching and displaying actual metadata (modification time, permissions, owner, group) from the parent directory.

## Objectives

1. Display accurate modification time for parent directory entry
2. Display correct permissions, owner, and group for parent directory
3. Maintain consistency with other file entries in the list
4. Preserve existing performance characteristics

## User Stories

### US-1: View Parent Directory Information
As a user, I want to see the actual modification time of the parent directory so that I can understand when the parent directory was last modified.

**Acceptance Criteria:**
- Parent directory (..) shows real modification timestamp
- Timestamp format matches other entries (YYYY-MM-DD HH:MM)
- Permissions, owner, and group are displayed correctly

## Technical Requirements

### TR-1: Metadata Retrieval
- Use `filepath.Dir(absPath)` to get parent directory path
- Use `os.Stat()` to retrieve:
  - `ModTime` - modification time
  - `Mode` - file permissions
- Use existing `GetFileOwnerGroup()` to retrieve:
  - Owner name
  - Group name

### TR-2: Error Handling
- If `os.Stat()` fails, continue with default values
- If `GetFileOwnerGroup()` fails, use "unknown" for owner/group
- Parent directory entry must always be displayed (when not at root)

### TR-3: Root Directory Behavior
- No parent directory entry when `absPath == "/"` (existing behavior)

## Implementation Approach

### Architecture

No architectural changes required. This is a localized fix in `ReadDirectory` function.

### Code Changes

**File:** `internal/fs/reader.go`

**Current Implementation (lines 26-32):**
```go
if absPath != "/" {
    fileEntries = append(fileEntries, FileEntry{
        Name:  "..",
        IsDir: true,
    })
}
```

**Proposed Implementation:**
```go
if absPath != "/" {
    parentPath := filepath.Dir(absPath)
    parentEntry := FileEntry{
        Name:  "..",
        IsDir: true,
    }

    // Get parent directory info
    if info, err := os.Stat(parentPath); err == nil {
        parentEntry.ModTime = info.ModTime()
        parentEntry.Permissions = info.Mode()
    }

    // Get owner and group info
    if owner, group, err := GetFileOwnerGroup(parentPath); err == nil {
        parentEntry.Owner = owner
        parentEntry.Group = group
    } else {
        parentEntry.Owner = "unknown"
        parentEntry.Group = "unknown"
    }

    fileEntries = append(fileEntries, parentEntry)
}
```

## Dependencies

- No new dependencies required
- Uses existing standard library: `os`, `filepath`
- Reuses existing function: `GetFileOwnerGroup()`

## Test Scenarios

### TS-1: Normal Directory
```go
func TestReadDirectory_ParentDirMetadata(t *testing.T) {
    // Create temp directory structure
    // Read directory
    // Verify parent entry has non-zero ModTime
    // Verify permissions are set
    // Verify owner/group are not empty
}
```

### TS-2: Root Directory
```go
func TestReadDirectory_RootNoParent(t *testing.T) {
    entries, err := ReadDirectory("/")
    // Verify no ".." entry exists
}
```

### TS-3: Permission Denied
```go
func TestReadDirectory_ParentPermissionDenied(t *testing.T) {
    // Create directory with restricted parent
    // Verify ".." entry still exists with default values
}
```

## Security Considerations

- No new security concerns
- Uses same permission checks as existing code
- Follows principle of least privilege (only reads metadata)

## Error Handling

| Error Condition | Handling |
|-----------------|----------|
| `os.Stat()` fails | Use zero values for ModTime/Permissions |
| `GetFileOwnerGroup()` fails | Use "unknown" for Owner/Group |
| Parent path resolution fails | Should not occur (filepath.Dir is deterministic) |

## Performance Optimization

### Analysis
- Adds exactly 2 system calls per directory read:
  - 1x `os.Stat()` for parent directory
  - 1x `GetFileOwnerGroup()` (internally uses syscall)
- Fixed overhead regardless of directory size
- Negligible impact: ~0.07% for 1000-file directories

### Benchmarks to Verify
```go
func BenchmarkReadDirectory(b *testing.B) {
    for i := 0; i < b.N; i++ {
        ReadDirectory("/tmp/test-dir")
    }
}
```

## Success Criteria

1. [ ] Parent directory timestamp displays actual modification time
2. [ ] Parent directory permissions display correctly
3. [ ] Parent directory owner/group display correctly
4. [ ] All existing tests pass
5. [ ] New unit tests added and passing
6. [ ] No performance regression (benchmark within 5% of baseline)

## Open Questions

None - all requirements clarified during specification phase.

## References

- Go `os.Stat`: https://pkg.go.dev/os#Stat
- Go `filepath.Dir`: https://pkg.go.dev/path/filepath#Dir
