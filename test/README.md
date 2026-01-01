# Test Instructions for AI Agents

This document provides guidelines for AI agents when writing and executing tests for duofm.

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
