# duofm - Unifies Orthodox File Manipulation

A terminal-based dual-pane file manager written in Go, inspired by classic file managers with vim-style keybindings.

## Features

### Core
- **Dual-pane interface**: Navigate two directories side-by-side
- **Vim-style keybindings**: Familiar hjkl navigation
- **File operations**: Copy, move, delete, rename files and directories
- **Multi-file marking**: Select multiple files with Space for batch operations
- **Symbolic link support**: Display targets, detect broken links

### Navigation
- **Search & Filter**: Incremental (`/`) and regex (`Ctrl+F`) search
- **Hidden files**: Toggle visibility with `Ctrl+H`
- **Quick navigation**: Home (`~`), previous directory (`-`), sync panes (`=`)
- **Sort options**: By name, size, or date (ascending/descending)

### Display
- **Three display modes**: Minimal, Basic (size+date), Detail (permissions+owner)
- **Context menu**: Press `@` for visual action selection
- **Help system**: Press `?` for scrollable keybinding reference

### Integration
- **External viewer**: Open files with $PAGER (`v` key)
- **External editor**: Edit files with $EDITOR (`e` key)
- **Shell commands**: Execute commands with `!` key

### Customization
- **Configuration file**: `~/.config/duofm/config.toml`
- **Custom keybindings**: Remap any key in config
- **Color theme**: Full 256-color customization

## Screenshots

```
┌─────────────────────────────────────────────────────────────┐
│ duofm v0.1.0                                                │
├──────────────────────────┬──────────────────────────────────┤
│ ~/projects/duofm         │ ~                                │
│──────────────────────────│──────────────────────────────────│
│ ../                      │ ../                              │
│ cmd/                     │ Documents/                       │
│ internal/                │ Downloads/                       │
│ test/                    │ Pictures/                        │
│ go.mod                   │ .bashrc                          │
│ go.sum                   │ .profile                         │
│ Makefile                 │                                  │
└──────────────────────────┴──────────────────────────────────┘
│ 1/7                                      ?:help q:quit      │
└─────────────────────────────────────────────────────────────┘
```

## Installation

### Prerequisites

- Go 1.21 or later

### Build from source

```bash
# Clone the repository
git clone https://github.com/sakura/duofm.git
cd duofm

# Install dependencies
go mod download

# Build the binary
make build

# Run
./duofm
```

### Install to system

```bash
# Install to $GOPATH/bin
make install

# Run from anywhere
duofm
```

## Usage

### Navigation

| Key     | Action                                    |
|---------|-------------------------------------------|
| `j`     | Move cursor down                          |
| `k`     | Move cursor up                            |
| `h`     | Move to left pane or parent directory     |
| `l`     | Move to right pane or parent directory    |
| `Enter` | Enter directory                           |

### File Operations

| Key | Action                              |
|-----|-------------------------------------|
| `c` | Copy to opposite pane               |
| `m` | Move to opposite pane               |
| `d` | Delete (with confirmation)          |

### Other

| Key       | Action         |
|-----------|----------------|
| `?`       | Show help      |
| `q`       | Quit           |
| `Ctrl+C`  | Quit           |

### Tips

- Use `h` and `l` to quickly switch between panes
- The active pane is highlighted with a blue border
- Press `?` anytime to see all available keybindings
- Confirmation dialogs appear for destructive operations (delete)
- Error messages are shown in red dialog boxes

## Development

### Project Structure

```
duofm/
├── cmd/duofm/           # Application entry point
├── internal/
│   ├── ui/             # TUI components (Bubble Tea)
│   │   ├── model.go    # Main application model
│   │   ├── pane.go     # File pane component
│   │   ├── dialog.go   # Dialog interface
│   │   └── ...
│   └── fs/             # File system operations
│       ├── reader.go   # Directory reading
│       ├── operations.go # File operations
│       └── ...
├── test/               # Integration tests
└── Makefile            # Build automation
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test -v ./internal/fs
go test -v ./internal/ui
go test -v ./test
```

### Code Quality

```bash
# Format code
make fmt

# Run static analysis
make vet

# Run linter (requires golangci-lint)
make lint
```

### Building

```bash
# Development build
make build

# Run directly
make run

# Clean build artifacts
make clean
```

## Technology Stack

- **Language**: Go 1.21+
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Composable TUI framework based on The Elm Architecture
- **Styling**: [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling library
- **Testing**: Go's built-in testing package

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](doc/CONTRIBUTING.md) for guidelines.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Make your changes
4. Write tests for your changes
5. Run tests (`make test`)
6. Commit your changes (`git commit -am 'Add new feature'`)
7. Push to the branch (`git push origin feature/my-feature`)
8. Create a Pull Request

### Code Style

- Follow standard Go conventions (see [Effective Go](https://go.dev/doc/effective_go))
- Use `gofmt` for formatting
- Write tests for new functionality
- Keep business logic separate from UI logic
- Document public APIs with comments

## Acknowledgments

- Inspired by [ranger](https://github.com/ranger/ranger) and [nnn](https://github.com/jarun/nnn)
- Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and [Lip Gloss](https://github.com/charmbracelet/lipgloss)
- Thanks to the Go community for excellent tooling and libraries

## Support

- Report issues: [GitHub Issues](https://github.com/sakura/duofm/issues)
- Documentation: See [doc/](doc/) directory
- Specification: [doc/specification/SPEC.md](doc/specification/SPEC.md)

---

Made with ❤️ by the duofm team
