# duofm - Unifies Orthodox File Manipulation

A terminal-based dual-pane file manager written in Go, inspired by classic file managers with vim-style keybindings.

## Features

- **Core File Management**
  - Dual-pane interface: Navigate two directories side-by-side
  - Vim-style keybindings: Familiar hjkl navigation with arrow key support
  - File operations: Copy, move, delete, rename files and directories
  - Smart rename: Extension-preserving rename (R) and full filename rename (Shift+R)
  - Permission editing: Change permissions (chmod) with Shift+P, including recursive mode
  - File creation: Create new files (N) and directories (Shift+N)
  - Multi-file marking: Select multiple files with Space for batch operations
  - Symbolic link support: Display targets, detect broken links, navigate to physical/logical paths
  - Overwrite handling: Smart conflict resolution with overwrite, skip, or rename options

- **Navigation & Search**
  - Search & Filter: Incremental (`/`) and regex (`Ctrl+F`) search with smart case
  - Directory history: Browser-like forward/back navigation (`Alt+←`/`Alt+→` or `[`/`]`)
  - Path jump: Direct navigation to any directory with `Ctrl+J` and Tab completion
  - Page scrolling: Ctrl+D/U and PageUp/PageDown for fast navigation through large lists
  - Hidden files: Toggle visibility with `Ctrl+H`
  - Quick navigation: Home (`~`), previous directory (`-`), sync panes (`=`)
  - Sort options: By name, size, or date with live preview
  - Bookmarks: Save and jump to frequently used directories (`B`/`Shift+B`)
  - Smart cursor: Remember position when navigating to parent directory
  - Refresh: Reload directory contents with F5 or Ctrl+R

- **Archive Operations** (Linux only)
  - Create archives: tar, tar.gz, tar.bz2, tar.xz, zip, 7z
  - Extract archives: tar, tar.gz, tar.bz2, tar.xz, zip, 7z
  - Smart extraction: Adapts to archive structure
  - Compression levels: Choose from 0-9 for supported formats
  - Security checks: Zip bomb detection, disk space validation

- **Display & UI**
  - Three display modes: Minimal, Basic (size+date), Detail (permissions+owner)
  - Unicode support: Proper display for Japanese, Chinese, Korean and emoji
  - East Asian Width: Configurable width for ambiguous characters
  - Context menu: Press `@` for visual action selection with number key shortcuts
  - Help system: Press `?` for scrollable keybinding reference with color palette
  - Dialog overlays: Dimmed background keeps file list visible during dialogs
  - Git branch display: Shows active branch in status bar when in Git repository

- **External Integration**
  - External viewer: Configurable Enter key behavior (pager, xdg-open, or custom app)
  - External editor: Edit files with $EDITOR (`E` key)
  - Shell commands: Execute commands with `!` key in current directory
  - Shell history: Ctrl+R incremental search and up/down arrow navigation
  - Context menu actions: Open and "Open with" for custom applications
  - Working directory: External apps open in file's directory

- **Customization**
  - Configuration file: `~/.config/duofm/config.toml` (auto-generated with auto-merge)
  - Custom keybindings: Remap any key with modifier support (Ctrl, Shift, Alt)
  - Color theme: Full 256-color customization for all UI elements
  - Enter key behavior: Choose between pager, xdg-open, or custom application
  - Bookmarks: Persisted in configuration file with edit/delete support

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

#### External Dependencies (for archive operations)

Archive operations require the following external tools to be installed:

| Format | Required Tools |
|--------|----------------|
| tar | `tar` |
| tar.gz | `tar`, `gzip` |
| tar.bz2 | `tar`, `bzip2` |
| tar.xz | `tar`, `xz` |
| zip | `zip`, `unzip` |
| 7z | `7z` (p7zip-full) |

On Debian/Ubuntu:
```bash
sudo apt install tar gzip bzip2 xz-utils zip unzip p7zip-full
```

On macOS:
```bash
brew install gnu-tar gzip bzip2 xz zip p7zip
```

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
| `P` (Shift+P) | Change permissions (chmod)    |
| `o` | Open context menu (includes Compress/Extract) |

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
