# Implementation Plan: Parent Directory Metadata Display

## Overview

Fix the bug where parent directory entry (..) displays timestamp as `0001-01-01 00:00` by fetching and displaying actual metadata from the parent directory. This is a localized fix in the `ReadDirectory` function.

## Objectives

- Display accurate modification time for parent directory entry
- Display correct permissions, owner, and group for parent directory
- Maintain consistency with other file entries
- Preserve existing performance characteristics

## Prerequisites

- Go 1.21 or later installed
- Existing codebase builds and tests pass

## Architecture Overview

No architectural changes required. This is a simple enhancement to the existing `ReadDirectory` function in `internal/fs/reader.go`.

## Implementation Phases

### Phase 1: Modify ReadDirectory Function

**Goal**: Add metadata retrieval for parent directory entry

**Files to Modify**:
- `internal/fs/reader.go` - Add parent directory metadata retrieval

**Implementation Steps**:

1. Get parent directory path using `filepath.Dir(absPath)`
2. Retrieve file info using `os.Stat(parentPath)`
3. Set `ModTime` and `Permissions` from the stat result
4. Retrieve owner/group using existing `GetFileOwnerGroup()` function
5. Handle errors gracefully (continue with defaults on failure)

**Code Change**:

Replace lines 26-32 in `internal/fs/reader.go`:

```go
// 親ディレクトリエントリを追加（ルートディレクトリ以外）
if absPath != "/" {
    parentPath := filepath.Dir(absPath)
    parentEntry := FileEntry{
        Name:  "..",
        IsDir: true,
    }

    // 親ディレクトリの情報を取得
    if info, err := os.Stat(parentPath); err == nil {
        parentEntry.ModTime = info.ModTime()
        parentEntry.Permissions = info.Mode()
    }

    // 所有者・グループ情報を取得
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

**Dependencies**: None (uses existing imports and functions)

**Testing**: Run existing tests + add new test case

**Estimated Effort**: Small

---

### Phase 2: Add Unit Tests

**Goal**: Verify parent directory metadata is correctly retrieved

**Files to Modify**:
- `internal/fs/reader_test.go` - Add test for parent directory metadata

**Implementation Steps**:

1. Add `TestReadDirectory_ParentDirMetadata` test function
2. Create temp directory structure
3. Verify parent entry has non-zero ModTime
4. Verify permissions are set
5. Verify owner/group are populated

**Test Code**:

```go
func TestReadDirectory_ParentDirMetadata(t *testing.T) {
    // Create a subdirectory to test parent metadata
    tmpDir := t.TempDir()
    subDir := filepath.Join(tmpDir, "subdir")
    if err := os.Mkdir(subDir, 0755); err != nil {
        t.Fatalf("Failed to create subdirectory: %v", err)
    }

    entries, err := ReadDirectory(subDir)
    if err != nil {
        t.Fatalf("ReadDirectory() failed: %v", err)
    }

    // Find parent directory entry
    var parentEntry *FileEntry
    for i := range entries {
        if entries[i].IsParentDir() {
            parentEntry = &entries[i]
            break
        }
    }

    if parentEntry == nil {
        t.Fatal("Parent directory entry not found")
    }

    // Verify ModTime is not zero value
    if parentEntry.ModTime.IsZero() {
        t.Error("Parent directory ModTime should not be zero")
    }

    // Verify Permissions is set
    if parentEntry.Permissions == 0 {
        t.Error("Parent directory Permissions should be set")
    }

    // Verify Owner/Group are populated
    if parentEntry.Owner == "" {
        t.Error("Parent directory Owner should not be empty")
    }
    if parentEntry.Group == "" {
        t.Error("Parent directory Group should not be empty")
    }
}
```

**Dependencies**: Phase 1 must be completed

**Testing**: Run `go test ./internal/fs/...`

**Estimated Effort**: Small

---

## File Structure

```
duofm/
└── internal/
    └── fs/
        ├── reader.go       # Modified: Add parent directory metadata
        └── reader_test.go  # Modified: Add test for parent metadata
```

## Testing Strategy

### Unit Tests

- `TestReadDirectory_ParentDirMetadata`: Verify parent entry has correct metadata
- Existing tests should continue to pass

### Manual Testing Checklist

- [ ] Navigate to any non-root directory
- [ ] Verify ".." entry shows actual modification timestamp (not 0001-01-01)
- [ ] Verify permissions column shows correct permissions for parent
- [ ] Verify owner/group columns show correct values
- [ ] Navigate to root directory (/) - verify no ".." entry

### Commands

```bash
# Run all tests
make test

# Run specific fs tests
go test -v ./internal/fs/...

# Build and run manually
make build && ./duofm
```

## Dependencies

### External Libraries
- None required (uses standard library only)

### Internal Dependencies
- `GetFileOwnerGroup()` function in `internal/fs/owner.go` - already exists

## Risk Assessment

### Technical Risks

- **Permission denied on parent directory**: Low risk
  - Mitigation: Error handling already in place, will use default values

### Implementation Risks

- **None identified**: This is a straightforward change

## Performance Considerations

- Adds 2 system calls per directory read (fixed overhead)
- Negligible impact: ~0.07% increase for 1000-file directories
- No caching needed for this simple case

## Security Considerations

- No new security concerns
- Uses same permission model as existing code
- Only reads metadata, no file content access

## Open Questions

None - all requirements clarified during specification phase.

## Success Criteria

1. [ ] Parent directory timestamp displays actual modification time
2. [ ] Parent directory permissions display correctly
3. [ ] Parent directory owner/group display correctly
4. [ ] All existing tests pass
5. [ ] New unit test added and passing
6. [ ] Manual verification completed

## References

- Specification: `doc/tasks/parent-directory-metadata/SPEC.md`
- Requirements: `doc/tasks/parent-directory-metadata/要件定義書.md`
