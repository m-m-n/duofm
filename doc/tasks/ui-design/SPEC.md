# Feature: UI Design (MVP)

## Overview

This specification defines the initial user interface design and implementation for duofm, a TUI (Text User Interface) dual-pane file manager written in Go.

The MVP (Minimum Viable Product) focuses on core functionality: displaying two panes, basic navigation, and fundamental file operations (copy, move, delete). The goal is to create a working dual-pane file manager with minimal but complete features.

## Objectives

- Create a functional dual-pane file manager interface
- Implement Vim-style keyboard navigation
- Provide basic file operations (copy, move, delete)
- Display help screen with keybinding list (lazygit-style)
- Handle errors gracefully with user-friendly messages
- Lay groundwork for future feature expansion

## User Stories

### MVP (Phase 1)

- As a user, I want to view two directory panes side-by-side
- As a user, I want to navigate directories using keyboard shortcuts
- As a user, I want to copy, move, and delete files
- As a user, I want confirmation before deleting files to prevent accidents
- As a user, I want to view available keybindings in a help screen

### Phase 2 (Future)

- As a user, I want to open files in editors or viewers
- As a user, I want to mark multiple files for batch operations
- As a user, I want to sort files by name, date, or size
- As a user, I want to toggle hidden file visibility
- As a user, I want to see detailed file information (size, date, permissions)

### Phase 3 (Future)

- As a user, I want to search and filter files by name
- As a user, I want advanced navigation (jump to top/bottom, page scroll)
- As a user, I want to see progress when copying large files
- As a user, I want to copy file paths to clipboard
- As a user, I want to remember the last opened directory

## Technical Requirements

### Architecture

#### Technology Stack

- **Language**: Go 1.21+
- **TUI Framework**: Bubble Tea (recommended)
  - Repository: github.com/charmbracelet/bubbletea
  - Reason: Modern, composable framework based on The Elm Architecture
  - Excellent for building complex TUI applications
- **Styling**: Lip Gloss (companion library for Bubble Tea)
  - Repository: github.com/charmbracelet/lipgloss
  - Provides layout and styling capabilities
- **File Operations**: Go standard library (os, filepath, io)

#### Component Structure

```
internal/ui/
├── model.go          # Main application model (Bubble Tea)
├── pane.go           # Pane component (left/right)
├── statusbar.go      # Status bar component
├── dialog.go         # Modal dialogs (confirmation, error, help)
├── keys.go           # Keybinding definitions
└── styles.go         # Visual styles (Lip Gloss)

internal/fs/
├── operations.go     # File operations (copy, move, delete)
├── navigation.go     # Directory traversal
└── sort.go           # File sorting logic
```

### UI Components

#### 1. Screen Layout

```
┌────────────────────────────────────────────────────────────────┐
│ duofm v0.1.0                                                    │ ← Title bar
├──────────────────────────────┬─────────────────────────────────┤
│ /home/user/Documents         │ /home/user                      │ ← Path display
├──────────────────────────────┼─────────────────────────────────┤
│ ../                          │ ../                             │
│ Projects/                    │ Documents/                      │
│ README.md                    │ Downloads/                      │
│ notes.txt                    │ Pictures/                       │
│ image.png                    │ .bashrc                         │
│                               │                                 │
│                               │                                 │
│                               │                                 │
├──────────────────────────────┴─────────────────────────────────┤
│ 3/5                                        ?:help q:quit       │ ← Status bar
└────────────────────────────────────────────────────────────────┘
```

**Components:**
- **Title Bar**: Application name and version
- **Path Display**: Current directory path for each pane (home directory shown as `~`)
- **Dual Panes**: Left and right file/directory lists
- **Status Bar**: Selection info and key hints

#### 2. File/Directory Display

**MVP (Minimal):**
- Directory names end with `/`
- Parent directory shown as `../`
- Cursor position indicated by highlight or `>` marker

**Default Sorting:**
- Directories listed first
- Then files
- Both sorted alphabetically (ascending)

**Phase 2 additions:**
- File size and modification date
- Icons (📁 📄 🔗)
- Permission information

#### 3. Initial Directories

On startup:
- **Left pane**: Current working directory (where duofm was launched)
- **Right pane**: Home directory (`~`)

**Phase 3 addition:**
- Right pane remembers last opened directory
- Falls back to home directory if last directory was deleted

### Keybindings

#### MVP Keybindings

| Key | Action | Context |
|-----|--------|---------|
| `j` | Move cursor down | File list |
| `k` | Move cursor up | File list |
| `h` | Move to left pane OR move to parent directory | Left pane: parent dir<br>Right pane: switch to left |
| `l` | Move to right pane OR move to parent directory | Left pane: switch to right<br>Right pane: parent dir |
| `Enter` | Enter directory | Directory selected |
| `c` | Copy to opposite pane | File/directory selected |
| `m` | Move to opposite pane | File/directory selected |
| `d` | Delete (with confirmation) | File/directory selected |
| `?` | Show help screen | Any |
| `Esc` | Close dialog/help | Dialog or help open |
| `q` | Quit application | Any |

#### Navigation Details

**Left Pane Focus:**
- `h`: Move left pane to parent directory
- `l`: Switch focus to right pane

**Right Pane Focus:**
- `h`: Switch focus to left pane
- `l`: Move right pane to parent directory

**Scrolling Behavior:**
- Cursor does not loop (stops at top/bottom of list)
- No wrapping when reaching list boundaries

#### Phase 2 Keybindings (Future)

| Key | Action |
|-----|--------|
| `e` | Open file in vim |
| `@` | Open with specified command |
| `n` | Create new file (opens in vim) |
| `Space` | Mark/unmark file |
| `V` | Select/deselect all |
| `s` | Toggle sort mode |
| `Ctrl+H` | Toggle hidden files |
| `p` | Toggle permission display |

#### Phase 3 Keybindings (Future)

| Key | Action |
|-----|--------|
| `/` | Search/filter files |
| `gg` | Jump to top |
| `G` | Jump to bottom |
| `Ctrl+D` | Scroll half page down |
| `Ctrl+U` | Scroll half page up |
| `Ctrl+C` | Copy path to clipboard |
| `=` | Sync opposite pane to current directory |
| `F5` or `Ctrl+R` | Refresh directory |

### File Operations

#### Copy Operation (`c`)

**Behavior:**
- Copies selected file/directory to opposite pane's directory
- Left pane `c` → copies to right pane's directory
- Right pane `c` → copies to left pane's directory
- Overwrites if file exists (Phase 2: add confirmation)

**Implementation:**
```go
func CopyFile(src, dst string) error
func CopyDirectory(src, dst string) error
```

**Error Handling:**
- Permission denied
- Insufficient disk space
- File not found

#### Move Operation (`m`)

**Behavior:**
- Moves selected file/directory to opposite pane's directory
- Left pane `m` → moves to right pane's directory
- Right pane `m` → moves to left pane's directory
- Overwrites if file exists (Phase 2: add confirmation)

**Implementation:**
```go
func MoveFile(src, dst string) error
```

**Error Handling:**
- Permission denied
- Cross-device move (falls back to copy + delete)
- File not found

#### Delete Operation (`d`)

**Behavior:**
- Shows confirmation dialog before deletion
- Deletes file or directory (recursive)

**Confirmation Dialog:**
```
┌───────────────────────────────┐
│ Delete file?                  │
│                               │
│ notes.txt                     │
│                               │
│ [y] Yes  [n] No               │
└───────────────────────────────┘
```

**Keys:**
- `y` or `Enter`: Confirm deletion
- `n` or `Esc`: Cancel

**Implementation:**
```go
func DeleteFile(path string) error
func DeleteDirectory(path string) error
```

**Error Handling:**
- Permission denied
- File in use
- Directory not empty (recursive delete)

### Help Screen

#### Lazygit-Style Help Display

Pressing `?` displays a full-screen help modal with keybinding list:

```
┌─────────────────────────────────────────────────────────────┐
│ Keybindings                                                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│ Navigation                                                   │
│   j/k      : move cursor down/up                            │
│   h/l      : move to left/right pane or parent directory    │
│   Enter    : enter directory                                │
│   q        : quit                                            │
│                                                              │
│ File Operations                                              │
│   c        : copy to opposite pane                          │
│   m        : move to opposite pane                          │
│   d        : delete (with confirmation)                     │
│                                                              │
│ Help                                                         │
│   ?        : show this help                                 │
│                                                              │
│                                                              │
│ Press Esc or ? to close                                     │
└─────────────────────────────────────────────────────────────┘
```

**Features:**
- Grouped by category (Navigation, File Operations, etc.)
- Clear key → action mapping
- Displays in modal overlay (blocks interaction with panes)
- Close with `Esc` or `?`

**Implementation:**
```go
type HelpScreen struct {
    keybindings []KeybindingHelp
}

type KeybindingHelp struct {
    category string
    key      string
    action   string
}
```

### Error Handling

#### Error Message Display

When file operations fail, display modal dialog:

```
┌───────────────────────────────┐
│ Error                         │
│                               │
│ Permission denied             │
│                               │
│ Press Esc to close            │
└───────────────────────────────┘
```

**Error Types:**
- Permission errors (read/write denied)
- File not found
- Disk space errors
- Filesystem errors

**User Actions:**
- `Esc` or `Enter`: Close dialog
- Application continues running

**Implementation:**
```go
type ErrorDialog struct {
    message string
}

func ShowError(err error) ErrorDialog
```

#### Symbolic Link Handling

**MVP:**
- Follow symbolic links transparently
- Broken links displayed as regular files (no special treatment)

**Phase 2:**
- Visual indicator for symlinks (🔗 icon or color)
- Display link target path

### Data Models

#### Application State (Bubble Tea Model)

```go
type Model struct {
    leftPane    *Pane
    rightPane   *Pane
    activPane   PanePosition // left or right
    dialog      Dialog       // nil or active dialog
    width       int
    height      int
}

type Pane struct {
    path        string
    entries     []fs.DirEntry
    cursor      int
    scrollOffset int
}

type Dialog interface {
    View() string
    Update(msg tea.Msg) (Dialog, tea.Cmd)
}

type PanePosition int
const (
    Left PanePosition = iota
    Right
)
```

#### Directory Entry

```go
type FileEntry struct {
    name        string
    isDir       bool
    size        int64
    modTime     time.Time
    permissions fs.FileMode
}
```

### Implementation Approach

#### Phase 1: Basic Structure (MVP)

1. **Setup Bubble Tea Application**
   - Initialize model with two panes
   - Implement basic Update/View cycle
   - Handle window resize

2. **Implement Pane Component**
   - Load directory contents
   - Display file/directory list
   - Handle cursor movement
   - Default sorting (directories first, then alphabetical)

3. **Implement Navigation**
   - hjkl cursor movement
   - Pane switching
   - Parent directory navigation
   - Enter directory

4. **Implement File Operations**
   - Copy to opposite pane
   - Move to opposite pane
   - Delete with confirmation dialog

5. **Implement Dialogs**
   - Confirmation dialog (delete)
   - Error dialog
   - Help screen (lazygit-style)

6. **Status Bar**
   - Show selection position / total
   - Show key hints

#### Phase 2: Enhanced Features (Future)

- File opening (less, vim, custom commands)
- Mark multiple files
- Sort toggle
- Hidden file toggle
- File information display
- Copy/move dialogs with name input
- Overwrite confirmation

#### Phase 3: Advanced Features (Future)

- Search/filter
- Advanced navigation (gg, G, Ctrl+D/U)
- Progress indicators
- Clipboard integration
- Directory sync
- Refresh command
- Visual mode for marking
- Remember last directory

### Dependencies

```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
)
```

## Test Scenarios

### MVP Test Cases

#### Navigation Tests
- [ ] Scenario 1: Launch duofm, verify left pane shows current directory, right pane shows home directory
- [ ] Scenario 2: Press `j` repeatedly, verify cursor moves down and stops at bottom
- [ ] Scenario 3: Press `k` repeatedly, verify cursor moves up and stops at top
- [ ] Scenario 4: In left pane, press `l`, verify focus switches to right pane
- [ ] Scenario 5: In right pane, press `h`, verify focus switches to left pane
- [ ] Scenario 6: In left pane, press `h`, verify pane moves to parent directory
- [ ] Scenario 7: In right pane, press `l`, verify pane moves to parent directory
- [ ] Scenario 8: Select directory and press `Enter`, verify pane enters the directory

#### File Operation Tests
- [ ] Scenario 9: Select file in left pane, press `c`, verify file is copied to right pane's directory
- [ ] Scenario 10: Select file in right pane, press `m`, verify file is moved to left pane's directory
- [ ] Scenario 11: Select file and press `d`, verify confirmation dialog appears
- [ ] Scenario 12: In delete confirmation, press `y`, verify file is deleted
- [ ] Scenario 13: In delete confirmation, press `n`, verify deletion is canceled
- [ ] Scenario 14: Try to copy file without permission, verify error dialog appears

#### Help Screen Tests
- [ ] Scenario 15: Press `?`, verify help screen appears with keybinding list
- [ ] Scenario 16: In help screen, press `Esc`, verify help closes and returns to file view
- [ ] Scenario 17: In help screen, press `?`, verify help closes

#### Error Handling Tests
- [ ] Scenario 18: Navigate to directory without read permission, verify error dialog
- [ ] Scenario 19: Try to delete read-only file, verify error dialog
- [ ] Scenario 20: Press `Esc` in error dialog, verify dialog closes

#### UI Display Tests
- [ ] Scenario 21: Verify directories are listed before files
- [ ] Scenario 22: Verify directories end with `/`
- [ ] Scenario 23: Verify parent directory shown as `../`
- [ ] Scenario 24: Verify home directory path shown as `~`
- [ ] Scenario 25: Verify status bar shows current position and total count
- [ ] Scenario 26: Verify cursor position is visually indicated

### Phase 2 Test Cases (Future)

- File opening tests
- Multi-file marking tests
- Sort toggle tests
- Hidden file toggle tests
- Rename/copy dialog tests

### Phase 3 Test Cases (Future)

- Search/filter tests
- Advanced navigation tests
- Progress indicator tests
- Clipboard tests
- Directory memory tests

## Success Criteria

### MVP Success Criteria

- [ ] Application displays two panes side-by-side
- [ ] Navigation works with hjkl keys
- [ ] Can enter directories with Enter key
- [ ] Can copy files between panes with `c`
- [ ] Can move files between panes with `m`
- [ ] Can delete files with `d` after confirmation
- [ ] Confirmation dialog appears before deletion
- [ ] Error dialogs appear when operations fail
- [ ] Help screen displays all keybindings
- [ ] Application quits with `q` key
- [ ] Left pane starts in current directory, right pane in home directory
- [ ] Files are sorted with directories first, then alphabetically

### Performance Criteria

- Directory loading completes within 1 second (for typical directories)
- No perceptible lag in keyboard input response
- Smooth cursor movement and pane switching

### Compatibility Criteria

- Works on Linux and macOS
- Minimum terminal size: 80x24
- Handles Unicode filenames correctly

## Open Questions

- [ ] Performance optimization strategy for very large directories (10,000+ files)?
- [ ] Color scheme preferences (use terminal default or custom theme)?
- [ ] Should MVP include basic file size display, or defer to Phase 2?
- [ ] Unit testing strategy for TUI components (how to test Bubble Tea views)?
- [ ] Integration testing approach (automated testing of key inputs)?

## Future Enhancements (Post-MVP)

### Phase 2 Priority
1. File opening capabilities (less, vim, custom commands)
2. Multi-file marking
3. Sort toggle functionality
4. Detailed file information display

### Phase 3 Priority
1. Search/filter functionality
2. Advanced navigation (gg, G, page scroll)
3. Progress indicators for long operations
4. Configuration file support for keybindings
5. Remember last directory across sessions

### Long-term Ideas
- Tabs for multiple directory pairs
- Preview pane
- File thumbnails (for images)
- Archive support (zip, tar.gz)
- Network file system support
- Plugin system for extensibility

## References

- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea)
- [Lip Gloss Documentation](https://github.com/charmbracelet/lipgloss)
- [lazygit](https://github.com/jesseduffield/lazygit) - Help screen reference
- [Midnight Commander](https://midnight-commander.org/) - Classic dual-pane file manager
- [ranger](https://github.com/ranger/ranger) - Vim-style file manager
- [The Elm Architecture](https://guide.elm-lang.org/architecture/) - Bubble Tea's architectural pattern
