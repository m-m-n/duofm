# Test Instructions for AI Agents

This document provides guidelines for AI agents when writing and executing tests for duofm.

## Test Strategy Overview

### Current Coverage Status

| Package | Coverage | Target | Priority |
|---------|----------|--------|----------|
| `internal/fs` | 87.9% | 95%+ | High |
| `internal/archive` | 80.8% | 90%+ | High |
| `internal/ui` | 79.4% | 85%+ | Medium |
| `internal/config` | 74.6% | 80%+ | Low |
| **Total** | **80.0%** | 85%+ | - |

### Quality Principles

**Coverage numbers are secondary to test depth on critical paths.**

For a file manager that performs destructive operations (delete, move, overwrite), the priority is:
1. **Error handling coverage** over happy path coverage
2. **Edge case coverage** over percentage metrics
3. **Security tests** for path traversal, symlinks, and archive extraction

## Priority Areas for 100% Coverage

### 1. Critical Path Functions (MUST be 100%)

These functions perform destructive or security-sensitive operations:

```
internal/fs/
├── operations.go
│   ├── CopyFile()      - Currently tested, verify overwrite cases
│   ├── MoveFile()      - Add EXDEV (cross-device) fallback test
│   ├── DeleteFile()    - Add permission denied cases
│   └── Delete()        - Recursive delete edge cases
├── symlink.go
│   └── All functions   - Symlink escape prevention
└── permissions.go
    └── All functions   - Permission modification safety

internal/archive/
├── smart_extractor.go
│   ├── Extract()       - Path traversal, zip bombs
│   └── GetArchiveMetadata() - Currently 44.4%, needs improvement
└── security.go
    └── All functions   - Currently good, extend to hardlinks
```

### 2. High-Value Missing Tests

| Function | File | Current | Issue |
|----------|------|---------|-------|
| `handleContextMenuResult` | model_update.go | 35.9% | Menu branch coverage |
| `extract` | smart_extractor.go | 53.1% | Error handling |
| `GetArchiveMetadata` | smart_extractor.go | 44.4% | Edge cases |
| `Delete` | bookmark_manager.go | 42.9% | Error propagation |

### 3. Commonly Overlooked Test Cases

Based on security review and second opinion analysis:

#### File Operations
- [ ] Cross-device move (`rename` fails with EXDEV) and copy+delete fallback
- [ ] Destination already exists (file vs directory, case sensitivity)
- [ ] Permission denied with clean error propagation (no partial corruption)
- [ ] Partial failure during recursive copy (cleanup behavior)
- [ ] Long paths, Unicode filenames, reserved names

#### Symlink Handling
- [ ] Copy/move symlink without following target
- [ ] Delete symlink without deleting target
- [ ] Symlink pointing outside destination during extraction

#### Archive Extraction
- [ ] Windows-style backslash paths in archives
- [ ] Hardlink entries in tar archives
- [ ] Zip bombs with high compression ratio
- [ ] Interrupted extraction cleanup

#### UI Operations
- [ ] Confirmation dialogs for destructive operations
- [ ] Overwrite/skip/rename conflict resolution
- [ ] Error message display and recovery

### 4. Test Depth Requirements

For each critical function, ensure tests cover:

```
┌─────────────────────────────────────────────────────────┐
│                    Test Coverage Depth                   │
├─────────────────────────────────────────────────────────┤
│ Level 1: Happy Path                                      │
│   └── Normal operation with valid input                  │
│                                                          │
│ Level 2: Input Validation                                │
│   ├── Empty input                                        │
│   ├── Nil/null values                                    │
│   └── Invalid format                                     │
│                                                          │
│ Level 3: Edge Cases                                      │
│   ├── Boundary values                                    │
│   ├── Unicode/special characters                         │
│   └── Maximum sizes                                      │
│                                                          │
│ Level 4: Error Conditions                                │
│   ├── Permission denied                                  │
│   ├── Disk full                                          │
│   ├── File not found                                     │
│   └── Concurrent access                                  │
│                                                          │
│ Level 5: Security                                        │
│   ├── Path traversal attempts                            │
│   ├── Symlink attacks                                    │
│   └── Resource exhaustion                                │
└─────────────────────────────────────────────────────────┘
```

## Test Framework

- **Unit Tests**: Go standard `testing` package
- **E2E Tests**: Custom bash scripts using tmux for terminal automation

## Test Execution

### Unit Tests

```bash
# Run all unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test -v ./internal/ui/...
```

### E2E Tests

```bash
# Build E2E test environment (Docker)
make test-e2e-build

# Run all E2E tests
make test-e2e

# Run interactive E2E test (for debugging)
docker run --rm -it duofm-e2e-test /e2e/scripts/interactive.sh "j j Enter"
```

## Test File Organization

```
duofm/
├── internal/
│   ├── ui/
│   │   ├── model.go
│   │   ├── model_test.go      # Unit tests next to source
│   │   ├── pane.go
│   │   └── pane_test.go
│   └── ...
└── test/
    └── e2e/
        ├── Dockerfile          # E2E test environment
        ├── testdata/           # Test fixtures
        └── scripts/
            ├── helpers.sh      # Shared helper functions
            ├── run_all_tests.sh # Main test runner
            └── tests/          # Individual test scripts
```

## Writing Unit Tests

### Test Naming Conventions

- Test file: `{source}_test.go` (e.g., `pane_test.go`)
- Test function: `Test{FunctionName}_{Scenario}` (e.g., `TestNavigateBack_EmptyHistory`)

### Test Structure

Use table-driven tests for Go:

```go
func TestAddToHistory(t *testing.T) {
    tests := []struct {
        name     string
        paths    []string
        expected []string
    }{
        {
            name:     "empty history",
            paths:    []string{"/home"},
            expected: []string{"/home"},
        },
        {
            name:     "multiple paths",
            paths:    []string{"/home", "/tmp", "/var"},
            expected: []string{"/home", "/tmp", "/var"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            h := NewDirectoryHistory(100)
            for _, p := range tt.paths {
                h.AddToHistory(p)
            }
            // assertions...
        })
    }
}
```

## Writing E2E Tests

### Adding New E2E Test Cases

Add new test functions to `test/e2e/scripts/run_all_tests.sh`:

```bash
# ===========================================
# Test: Feature Name
# ===========================================
test_feature_name() {
    start_duofm "$CURRENT_SESSION"

    # Execute test actions
    send_keys "$CURRENT_SESSION" "key1" "key2"
    sleep 0.3

    # Verify expected behavior
    assert_contains "$CURRENT_SESSION" "expected_text" \
        "Description of what should happen"

    stop_duofm "$CURRENT_SESSION"
}

# Register the test at the bottom of the file
run_test test_feature_name
```

### Available Helper Functions

| Function | Usage | Description |
|----------|-------|-------------|
| `start_duofm` | `start_duofm "$CURRENT_SESSION"` | Start duofm in tmux session |
| `send_keys` | `send_keys "$CURRENT_SESSION" "j" "k"` | Send keystrokes |
| `capture_screen` | `capture_screen "$CURRENT_SESSION"` | Get current screen content |
| `stop_duofm` | `stop_duofm "$CURRENT_SESSION"` | Stop duofm session |
| `assert_contains` | `assert_contains "$CURRENT_SESSION" "text" "desc"` | Verify text exists |
| `assert_not_contains` | `assert_not_contains "$CURRENT_SESSION" "text" "desc"` | Verify text doesn't exist |
| `assert_cursor_position` | `assert_cursor_position "$CURRENT_SESSION" "3" "desc"` | Verify cursor at line N |

### Key Mappings for E2E Tests

| Key | tmux send-keys | Description |
|-----|----------------|-------------|
| Arrow keys | `Up`, `Down`, `Left`, `Right` | Navigation |
| Enter | `Enter` | Confirm/open |
| Escape | `Escape` | Cancel/close |
| Tab | `Tab` | Switch pane |
| Ctrl+C | `C-c` | Quit |
| Alt+Left | `M-Left` | History back |
| Alt+Right | `M-Right` | History forward |
| Space | `Space` | Select item |
| Letters | `a`, `b`, etc. | Direct key |

### E2E Test Guidelines

1. **Always call `stop_duofm`** at the end of each test
2. **Add `sleep` after key sequences** that trigger async operations
3. **Use descriptive assertion messages** for debugging
4. **Keep tests independent** - each test starts with fresh duofm instance

## Common Patterns

### Testing Keyboard Navigation

```bash
test_navigation() {
    start_duofm "$CURRENT_SESSION"

    # Move down 3 times
    send_keys "$CURRENT_SESSION" "j" "j" "j"
    assert_cursor_position "$CURRENT_SESSION" "4" "Cursor moved to line 4"

    # Enter directory
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    assert_contains "$CURRENT_SESSION" "expected_dir_content" "Entered directory"

    stop_duofm "$CURRENT_SESSION"
}
```

### Testing Dialog Interactions

```bash
test_dialog() {
    start_duofm "$CURRENT_SESSION"

    # Open dialog
    send_keys "$CURRENT_SESSION" "b"
    sleep 0.3
    assert_contains "$CURRENT_SESSION" "Bookmarks" "Dialog opened"

    # Close dialog
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3
    assert_not_contains "$CURRENT_SESSION" "Bookmarks" "Dialog closed"

    stop_duofm "$CURRENT_SESSION"
}
```

### Testing Error Conditions

```bash
test_error_handling() {
    start_duofm "$CURRENT_SESSION"

    # Trigger error condition
    send_keys "$CURRENT_SESSION" "some_action"
    sleep 0.5

    # Verify error message appears
    assert_contains "$CURRENT_SESSION" "Error:" "Error message displayed"

    # Verify application is still functional
    send_keys "$CURRENT_SESSION" "j"
    assert_cursor_position "$CURRENT_SESSION" "2" "Navigation still works"

    stop_duofm "$CURRENT_SESSION"
}
```

## Important Notes

1. **Do NOT create symlinks** like `run_tests.sh -> run_all_tests.sh`. The main test runner is `run_all_tests.sh`.

2. **Test data location**: E2E tests run in Docker with `/testdata` as the working directory.

3. **Terminal compatibility**: Some key combinations (like `Alt+Arrow`) may not work in all terminals. Always test with alternative keys when available.

4. **Cleanup**: The `run_test` helper automatically cleans up tmux sessions, but individual tests should still call `stop_duofm`.

## Test Quality Checklist

Before considering a function fully tested, verify:

### Unit Tests
- [ ] Happy path with valid input
- [ ] Empty/nil input handling
- [ ] Boundary values (0, max, negative)
- [ ] Error return values are tested
- [ ] Error messages are meaningful

### File Operation Tests
- [ ] Source file doesn't exist
- [ ] Destination already exists
- [ ] Permission denied (read/write/execute)
- [ ] Cross-device operation (if applicable)
- [ ] Symlink handling

### UI Tests
- [ ] Initial state verification
- [ ] State after user action
- [ ] Emitted commands are correct (not just non-nil)
- [ ] Error state display
- [ ] Recovery from error state

### Security Tests
- [ ] Path traversal prevention
- [ ] Symlink escape prevention
- [ ] Resource limits enforced
- [ ] Input sanitization

## Metrics and Goals

### Coverage Targets by Priority

| Priority | Package | Current | Target | Gap |
|----------|---------|---------|--------|-----|
| P0 | fs/operations.go | ~85% | 100% | Critical functions |
| P0 | archive/security.go | ~90% | 100% | Security functions |
| P1 | fs/symlink.go | ~80% | 95% | Symlink safety |
| P1 | archive/smart_extractor.go | ~60% | 90% | Extraction |
| P2 | ui/model_operations.go | ~70% | 85% | File operations UI |
| P2 | config/bookmark.go | ~75% | 85% | Bookmark persistence |
| P3 | Other UI components | ~80% | 80% | Maintain |

### Quality Metrics

Beyond coverage percentage, track:
- **Assertion density**: Tests should have meaningful assertions, not just "runs without panic"
- **Edge case ratio**: At least 40% of tests should cover edge cases or errors
- **Mock isolation**: File system tests should use `t.TempDir()` consistently
