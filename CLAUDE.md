# duofm - Dual Pane File Manager

## Project Overview

duofm is a TUI (Text User Interface) dual-pane file manager written in Go. It provides efficient file management through a terminal interface with two side-by-side panes for easy navigation and file operations between directories.

**Key Features:**
- Dual-pane interface for intuitive file operations
- Terminal-based (runs anywhere with a terminal)
- Fast and lightweight
- Keyboard-driven workflow

## Documentation Structure

- `doc/specification/SPEC.md` - Overall project specification and architecture
- `doc/tasks/FEATURE_NAME/SPEC.md` - Feature-specific specifications and requirements
- `doc/CONTRIBUTING.md` - Detailed documentation guidelines and contribution rules

See `doc/CONTRIBUTING.md` for complete documentation guidelines.

## Development Setup

### Prerequisites
- Go 1.21 or later
- Make (for build automation)

### Quick Start
```bash
# Install dependencies
go mod download

# Build
make build

# Run
./duofm

# Run tests
make test
```

## Architecture

### Project Structure
```
duofm/
├── cmd/duofm/        # Application entry point
├── internal/
│   ├── ui/          # TUI components and rendering
│   ├── fs/          # File system operations
│   └── config/      # Configuration handling
├── pkg/             # Public libraries (if any)
└── Makefile         # Build automation
```

### Recommended TUI Libraries

**Bubble Tea** (recommended for beginners):
- Modern, composable framework based on The Elm Architecture
- Great for building complex TUI apps
- Repository: github.com/charmbracelet/bubbletea

**tview**:
- High-level TUI framework with many widgets
- Good for rapid development
- Repository: github.com/rivo/tview

**tcell**:
- Lower-level terminal manipulation
- More control but requires more code
- Repository: github.com/gdamore/tcell

### File System Operations
- Use `filepath.Walk` or `os.ReadDir` for directory traversal
- Handle symlinks carefully with `filepath.EvalSymlinks`
- Consider using `fsnotify` for watching directory changes
- Implement proper error handling for permission denied scenarios

## Development Workflow

### Building
```bash
make build          # Build binary
make install        # Install to $GOPATH/bin
make clean          # Clean build artifacts
```

### Testing
Go provides built-in testing with the `testing` package:

```bash
go test ./...                    # Run all tests
go test -v ./...                 # Verbose output
go test -cover ./...             # With coverage
go test -race ./...              # Race condition detection
```

**Testing Strategies:**
- Unit tests: Test file operations, configuration parsing
- Integration tests: Test UI components with mocked file systems
- Table-driven tests: Common Go pattern for testing multiple scenarios
- Use `testify/assert` for readable assertions (optional)

### Code Quality
```bash
gofmt -w .                      # Format code
go vet ./...                    # Static analysis
golangci-lint run               # Comprehensive linting (optional)
```

## Conventions

- Follow standard Go code conventions (Effective Go)
- Use `gofmt` for consistent formatting
- Write tests alongside code (`*_test.go` files)
- Keep UI logic separate from business logic
- Use context for cancellation in long-running operations

## Performance Considerations

- Cache directory listings for large directories
- Use goroutines for async operations (copying, moving files)
- Implement proper cancellation for long operations
- Consider lazy loading for very large directories
- Profile with `pprof` if performance issues arise

## Useful Resources

- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
