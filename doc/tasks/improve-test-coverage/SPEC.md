# Feature: Improve Test Coverage

## Overview

duofmプロジェクトのテストカバレッジを向上させる。現在の全体カバレッジ77.4%を目標の80%以上に引き上げ、各パッケージで80%以上のカバレッジを達成する。

## Current Coverage Status

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/fs` | 87.9% | Good |
| `internal/archive` | 80.8% | Good |
| `internal/ui` | 76.0% | Needs Improvement |
| `internal/config` | 73.6% | Needs Improvement |
| `cmd/duofm` | 0.0% | No Tests |
| `internal/version` | - | No Test File |

**Overall Coverage: 77.4%**

## Objectives

- Achieve 80%+ overall test coverage
- Achieve 80%+ coverage for each package
- Add tests for all core functionality without test coverage
- Prioritize security-related code (permission handling)

## Scope

### Priority 1: High (Core Functionality / Security)

#### Files requiring new test files:
- `internal/ui/archive_operation_manager.go` - Archive compression/extraction core
- `internal/ui/batch_operation_manager.go` - Batch file operations
- `internal/ui/bookmark_manager.go` - Bookmark management
- `internal/config/defaults.go` - Default settings
- `internal/config/generator.go` - Config file generation
- `internal/version/version.go` - Version information

#### Files requiring additional tests:
- `internal/ui/model_permission.go` - Permission change functionality (security)
- `internal/config/path.go` - `GetConfigDir` function

### Priority 2: Medium (Usability)

- `internal/ui/model_update.go` - Archive-related handlers (11-35% coverage)
- `internal/ui/pane.go` - Untested methods

### Priority 3: Low (Edge Cases)

- `cmd/duofm/main.go` - Entry point (integration tests)
- `internal/ui/exec.go` - `openWithXDG`, `openWithCustom`

## Functional Requirements

### FR-1: Archive Operation Manager Tests

Test coverage for `archive_operation_manager.go`:

- FR-1.1: Test `NewArchiveOperationManager` initialization
- FR-1.2: Test `IsActive` state detection
- FR-1.3: Test `State` returns correct operation state
- FR-1.4: Test `PrepareCompression` sets up compression task
- FR-1.5: Test `StartCompression` initiates compression
- FR-1.6: Test `PrepareExtraction` sets up extraction task
- FR-1.7: Test `StartExtraction` initiates extraction
- FR-1.8: Test `CheckSecurity` validates archive safety
- FR-1.9: Test `PollProgress` returns progress updates
- FR-1.10: Test `CancelTask` properly cancels operations
- FR-1.11: Test `Clear` resets manager state

### FR-2: Batch Operation Manager Tests

Test coverage for `batch_operation_manager.go`:

- FR-2.1: Test `NewBatchOperationManager` initialization
- FR-2.2: Test `Start` begins batch operation
- FR-2.3: Test `Current` returns current batch state
- FR-2.4: Test `CurrentFile` returns file being processed
- FR-2.5: Test `DestPath` returns correct destination
- FR-2.6: Test `Operation` returns operation type
- FR-2.7: Test `Advance` moves to next file
- FR-2.8: Test `Cancel` aborts batch operation
- FR-2.9: Test `ExecuteCurrentFile` performs file operation

### FR-3: Bookmark Manager Tests

Test coverage for `bookmark_manager.go`:

- FR-3.1: Test `NewBookmarkManager` initialization
- FR-3.2: Test `SetBookmarks` updates bookmark list
- FR-3.3: Test `Add` creates new bookmark
- FR-3.4: Test `Edit` modifies existing bookmark
- FR-3.5: Test `Delete` removes bookmark
- FR-3.6: Test `EditIndex` and `SetEditIndex` for editing state
- FR-3.7: Test `save` persists bookmarks to file

### FR-4: Config Package Tests

Test coverage for config files:

- FR-4.1: Test `GetConfigDir` returns correct path
- FR-4.2: Test `GetConfigDir` respects XDG_CONFIG_HOME
- FR-4.3: Test `GenerateDefaultConfig` creates valid TOML
- FR-4.4: Test default values are correctly applied
- FR-4.5: Test `GetHistoryPath` returns correct path

### FR-5: Version Package Tests

Test coverage for version.go:

- FR-5.1: Test `Version` returns correct version string
- FR-5.2: Test `BuildInfo` returns build information

### FR-6: Permission Handler Tests

Additional test coverage for model_permission.go:

- FR-6.1: Test `handlePermission` processes permission dialog
- FR-6.2: Test `executePermissionChange` modifies file permissions
- FR-6.3: Test `executeRecursivePermissionChange` handles directories
- FR-6.4: Test `handleBatchPermission` processes multiple files
- FR-6.5: Test error handling for permission denied scenarios

## Non-Functional Requirements

- NFR-1: All tests must pass with `go test ./...`
- NFR-2: No test should take longer than 5 seconds
- NFR-3: Tests must be isolated (no side effects between tests)
- NFR-4: Use `t.TempDir()` for file system tests
- NFR-5: Follow existing test patterns (table-driven tests)

## Test Patterns

### Existing Patterns to Follow

```go
// Table-driven tests
tests := []struct {
    name    string
    setup   func(string) (string, string)
    wantErr bool
}{...}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test implementation
    })
}
```

### File System Test Pattern

```go
func TestSomeOperation(t *testing.T) {
    tmpDir := t.TempDir()
    // Create test files in tmpDir
    // Run test
    // Assertions
}
```

## Files to Create

| File | Priority | Purpose |
|------|----------|---------|
| `internal/ui/archive_operation_manager_test.go` | High | Archive operations |
| `internal/ui/batch_operation_manager_test.go` | High | Batch operations |
| `internal/ui/bookmark_manager_test.go` | High | Bookmark management |
| `internal/config/defaults_test.go` | Medium | Default settings |
| `internal/config/generator_test.go` | Medium | Config generation |
| `internal/version/version_test.go` | Low | Version info |

## Files to Extend

| File | Priority | Additional Tests Needed |
|------|----------|------------------------|
| `internal/ui/model_permission_test.go` | High | Permission handlers |
| `internal/config/config_test.go` | Medium | GetConfigDir, GetHistoryPath |

## Success Criteria

- [ ] Overall coverage >= 80%
- [ ] `internal/ui` coverage >= 80%
- [ ] `internal/config` coverage >= 80%
- [ ] All new test files created
- [ ] All tests pass with `go test ./...`
- [ ] No race conditions (`go test -race ./...`)

## Dependencies

- Standard `testing` package
- `t.TempDir()` for isolated file operations
- Existing test utilities in the project

## Constraints

- Follow existing code style and test patterns
- No external test dependencies (testify not used in project)
- Tests must work on Linux (primary platform)
