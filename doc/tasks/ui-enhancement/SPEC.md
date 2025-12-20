# Feature: UI Enhancement

## Overview

This specification enhances the duofm user interface to provide a more practical and informative experience similar to Total Commander. The enhancement includes adding header information, detailed file list display with multiple view modes, and improved symbolic link handling.

## Objectives

- Display essential file operation information (marked files, capacity, free space) at all times
- Allow users to toggle display detail levels based on their needs
- Properly handle symbolic links and display target information
- Provide a practical file manager UI similar to Total Commander

## User Stories

- As a user, I want to see the number of marked files and their total size in the header
- As a user, I want to see the partition's free space in the header
- As a user, I want to toggle between basic and detail display modes using the `i` key when terminal is wide
- As a user, I want the display to automatically switch to minimal mode when the terminal is narrow
- As a user, I want to see file sizes, timestamps, permissions, owners, and groups when needed
- As a user, I want to see symbolic link targets and navigate to them
- As a user, I want to visually identify broken symbolic links
- As a user, I want loading feedback when opening directories with many files

## Technical Requirements

### 1. Header Information Enhancement

#### 1.1 Current Layout

```
/home/user/Documents
────────────────────
File list...
```

#### 1.2 New Layout (2-line header)

**Normal state:**
```
/home/user/Documents
Marked 3/15 12.5 MiB    150 GiB Free
────────────────────
File list...
```

**Loading state:**
```
/home/user/Documents
Loading directory... (1234 files loaded)
────────────────────
(still loading)
```

**Line 1**: Directory path (existing implementation)
- Display absolute path of current directory
- When under home directory, abbreviate home directory part with `~` (e.g., `/home/user/Documents` → `~/Documents`)

**Line 2**: Mark information and free space, or loading state
- Normal state: `Marked X/N SIZE` + `SIZE Free`
- Loading state: `Loading directory...` + progress info (optional)

#### 1.3 Detailed Specification

**Mark Information:**
- `X`: Number of marked files/directories
- `N`: Total number of displayed files/directories (excluding `..`)
- `SIZE`: Total size of marked files (base-1024, auto unit conversion)
- When no files marked: `Marked 0/N 0 B`

**Free Space:**
- Display partition free space
- Base-1024 unit auto conversion (B, KiB, MiB, GiB, TiB)
- Refresh every 5 seconds (future: configurable via `refresh_rate` in config file)

**Layout:**
- Left side: Mark information
- Right side: Free space
- Center: Filled with spaces

### 2. Detailed File List Display

#### 2.1 Display Modes

Display mode is determined by **terminal width** and **user selection**.

##### 2.1.1 Narrow Terminal (Automatic)

When terminal width is insufficient, **automatically** display only file/directory names.
In this case, the `i` key is **disabled** (does nothing).

**Minimal Mode (automatic fallback):**
```
../
Projects/
README.md
notes.txt
config.conf -> /etc/app/config.conf
```

##### 2.1.2 Wide Terminal (Manual Toggle)

When terminal width is sufficient, the `i` key toggles between **two modes**.

**Mode A: Basic Info Mode (Default)**
```
../
Projects/         -        2024-12-10 09:00
README.md         1.2 KiB  2024-12-01 10:30
notes.txt         450 B    2024-11-28 15:45
config.conf -> /etc/app/config.conf  512 B  2024-12-01 10:30
```

Display columns:
- File name
- File size
- Timestamp

**Mode B: Detail Info Mode**
```
../
Projects/         rwxr-xr-x  user  staff
README.md         rw-r--r--  user  staff
notes.txt         rw-r--r--  user  staff
config.conf -> /etc/app/config.conf  rw-r--r--  user  staff
```

Display columns:
- File name
- Permissions
- Owner
- Group

**Toggle Behavior:**
- `i` key: Mode A ⇔ Mode B
- When terminal becomes narrow: Automatically switch to minimal mode, `i` key disabled
- When terminal becomes wide: Return to last selected mode (A or B), `i` key enabled

#### 2.2 Column Specifications

**File Name:**
- Left-aligned
- Directories suffixed with `/`
- Parent directory shown as `../`
- Symbolic links shown as `name -> target`
- Variable width (adjusted to terminal width)

**File Size:**
- Right-aligned
- Base-1024 auto unit conversion (B, KiB, MiB, GiB, TiB)
- Directories shown as `-`
- Fixed width (auto-determined)

**Timestamp:**
- Format: `2024-12-17 22:28` (ISO 8601 compliant)
- Use Go's `time.Format("2006-01-02 15:04")`
- Year-month-day hour:minute format
- 24-hour format
- Fixed width (16 characters: `2024-12-17 22:28`)
- Not customizable (prioritizes simplicity)

**Permissions:**
- Fixed width (10 characters: `rwxrwxrwx` format)
- Unix-style permission display

**Owner:**
- Left-aligned
- Display username
- Fixed width (auto-determined)

**Group:**
- Left-aligned
- Display group name
- Fixed width (auto-determined)

#### 2.3 Terminal Width Threshold

Terminal width thresholds for display mode switching:
- **Force minimal mode**: Terminal width < 60 columns (guideline, adjust during implementation)
- **Allow Mode A/B toggle**: Terminal width >= 60 columns

This threshold is dynamically determined by calculating total column widths.

#### 2.4 Display Mode Management

- Each pane can toggle display mode independently (Mode A and Mode B)
- When terminal is narrow, automatically switch to minimal mode and disable `i` key
- Settings reset on application restart (not persisted)
- Current mode not displayed in status bar (may be added later if needed)

### 3. Symbolic Link Handling Improvement

#### 3.1 Current Issues

- Cannot follow symbolic links even when target is a directory
- Link target information not displayed
- Cannot identify broken links

#### 3.2 Improvements

**Display Format:**
```
config.conf -> /etc/app/config.conf    512 B    2024-12-01 10:30
logs -> /var/log/app/                    -      2024-12-10 09:00
broken -> /missing/file                  ?      2024-11-15 08:00
```

**Link Target Display:**
- Format: `filename -> absolute path to target`
- Truncate with `...` if target path is too long

**Broken Link Visual Representation:**
- Change color (e.g., red, grayed out)
- Display `?` in size column
- Add broken link marker (color, icon, etc.)

**Enter Key Behavior:**
- Target exists and is directory → Navigate to target directory
- Target exists and is file → Do nothing (file open feature planned for Phase 2)
- Target does not exist → Do nothing (no error dialog)

**Symbolic Link Information Retrieval:**
- Use `os.Readlink()` to get target path
- Use `os.Stat()` to check target existence
- Determine if target is directory or file

### 4. Loading Display

#### 4.1 Purpose

Provide feedback to users when opening directories with many files (thousands to tens of thousands).

#### 4.2 Display Timing

- When directory loading takes time (threshold: 0.5+ seconds)
- During directory navigation, refresh

#### 4.3 Display Location and Content

**Display in header line 2:**
```
/home/user/large-directory
Loading directory... (1234 files loaded)
────────────────────
(still loading)
```

Or with percentage:
```
/home/user/large-directory
Loading... 45%
────────────────────
(still loading)
```

**Display content:**
- `Loading directory...` - Basic loading message
- `(N files loaded)` - Number of files loaded so far (optional)
- `N%` - Progress percentage (if calculable)

**Normal header line 2 (mark info and free space) is hidden during loading.**

#### 4.4 Implementation Approach

- Utilize Bubble Tea's async messaging
- Execute directory reading in goroutine
- Display loading state in header line 2 during loading
- Switch header line 2 to normal display (mark info and free space) after loading complete

## Implementation Approach

### Architecture

#### Data Structure Extensions

**FileEntry struct (existing):**
```go
type FileEntry struct {
    Name        string
    IsDir       bool
    Size        int64
    ModTime     time.Time
    Permissions fs.FileMode
}
```

**Required extensions:**
```go
type FileEntry struct {
    Name        string
    IsDir       bool
    Size        int64
    ModTime     time.Time
    Permissions fs.FileMode
    Owner       string      // Owner name (new)
    Group       string      // Group name (new)
    IsSymlink   bool        // Is symbolic link (new)
    LinkTarget  string      // Link target path (new)
    LinkBroken  bool        // Is link broken (new)
}
```

#### Pane Struct Extensions

```go
type Pane struct {
    path         string
    entries      []fs.FileEntry
    cursor       int
    scrollOffset int
    width        int
    height       int
    isActive     bool
    displayMode  DisplayMode  // Display mode (new)
    markedFiles  map[int]bool // Marked file indices (future)
    loading      bool         // Is loading (new)
}

type DisplayMode int
const (
    DisplayMinimal DisplayMode = iota  // Name only (automatic when terminal narrow)
    DisplayBasic                        // Name + size + timestamp (Mode A, default)
    DisplayDetail                       // Name + permissions + owner + group (Mode B)
)
```

#### Model Struct Extensions

```go
type Model struct {
    leftPane          *Pane
    rightPane         *Pane
    leftPath          string
    rightPath         string
    activePane        PanePosition
    dialog            Dialog
    width             int
    height            int
    ready             bool
    lastDiskSpaceCheck time.Time      // Last disk space check time (new)
    leftDiskSpace     uint64          // Left pane free space (new)
    rightDiskSpace    uint64          // Right pane free space (new)
}
```

### API Design

#### File System Operations

**Get file owner and group:**
```go
// GetFileOwnerGroup returns owner and group names for a file
func GetFileOwnerGroup(path string) (owner, group string, err error)
```

**Get symbolic link information:**
```go
// GetSymlinkInfo returns symbolic link target and status
func GetSymlinkInfo(path string) (target string, isBroken bool, err error)
```

**Get disk space:**
```go
// GetDiskSpace returns free space for the partition containing path
func GetDiskSpace(path string) (freeBytes uint64, err error)
```

#### UI Rendering

**Format file size:**
```go
// FormatSize formats bytes to human-readable string (base-1024)
// Examples: 512 B, 1.5 KiB, 2.3 MiB, 1.8 GiB, 3.2 TiB
func FormatSize(bytes int64) string
```

**Format timestamp:**
```go
// FormatTimestamp formats time.Time in ISO 8601 format "2006-01-02 15:04"
func FormatTimestamp(t time.Time) string {
    return t.Format("2006-01-02 15:04")
}
```

**Format permissions:**
```go
// FormatPermissions formats fs.FileMode to Unix-style string
// Example: rwxr-xr-x
func FormatPermissions(mode fs.FileMode) string
```

#### Display Mode Management

**Toggle display mode:**
```go
// ToggleDisplayMode toggles between Mode A and Mode B
// Only called when terminal is wide enough
func (p *Pane) ToggleDisplayMode() {
    if p.displayMode == DisplayBasic {
        p.displayMode = DisplayDetail
    } else if p.displayMode == DisplayDetail {
        p.displayMode = DisplayBasic
    }
    // DisplayMinimal is never set manually, only by width check
}

// ShouldUseMinimalMode checks if terminal width requires minimal mode
func (p *Pane) ShouldUseMinimalMode() bool {
    requiredWidth := p.calculateRequiredWidth()
    return p.width < requiredWidth
}
```

**Render file entry:**
```go
// RenderEntry renders a file entry according to current display mode
func (p *Pane) RenderEntry(entry fs.FileEntry, isCursor bool) string
```

### Dependencies

**Existing dependencies:**
```go
require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.9.1
)
```

**Additional dependencies (for disk space):**
```go
require (
    golang.org/x/sys v0.15.0  // For syscall.Statfs (Unix) or similar
)
```

### Platform Considerations

**Unix/Linux:**
- Use `syscall.Statfs` for disk space
- Use `os/user` package for owner/group lookup
- Use `os.Readlink` for symbolic links

**Windows:**
- Use appropriate Windows API for disk space
- Owner/group handling may differ
- Symbolic link handling uses same `os.Readlink`

**macOS:**
- Same as Unix/Linux

## Test Scenarios

### Header Display Tests

- [ ] Test 1: Directory path displays absolute path correctly
- [ ] Test 2: Home directory path abbreviates with `~` (e.g., `/home/user/Documents` → `~/Documents`)
- [ ] Test 3: File count is correct (excluding `..`)
- [ ] Test 4: Free space displays with appropriate unit
- [ ] Test 5: Free space updates every 5 seconds
- [ ] Test 6: Mark count shows 0/N when no files marked
- [ ] Test 7: Header layout adjusts to pane width

### Display Mode Toggle Tests

- [ ] Test 8: When terminal is narrow, display automatically switches to minimal mode (name only)
- [ ] Test 9: When terminal is narrow, `i` key is disabled (does nothing)
- [ ] Test 10: When terminal is wide, press `i` toggles Mode A ⇔ Mode B
- [ ] Test 11: Left and right panes toggle independently
- [ ] Test 12: Mode A shows name + size + timestamp
- [ ] Test 13: Mode B shows name + permissions + owner + group
- [ ] Test 14: When terminal becomes narrow, automatically switch to minimal mode
- [ ] Test 15: When terminal becomes wide, return to last selected mode (A or B)

### File Information Display Tests

- [ ] Test 16: File size displays with appropriate unit (B, KiB, MiB, GiB, TiB)
- [ ] Test 17: Directory size displays as `-`
- [ ] Test 18: Timestamp displays in `2024-12-17 22:28` format (ISO 8601)
- [ ] Test 19: Timestamp includes year
- [ ] Test 20: Timestamp uses 24-hour format
- [ ] Test 21: Permissions display in Unix format (rwxrwxrwx)
- [ ] Test 22: Owner and group display correctly
- [ ] Test 23: Very large file sizes display correctly (TiB range)
- [ ] Test 24: Column widths auto-adjust appropriately

### Symbolic Link Tests

- [ ] Test 25: Symbolic link displays as `name -> target`
- [ ] Test 26: Enter on directory link navigates to target
- [ ] Test 27: Enter on file link does nothing
- [ ] Test 28: Broken link visually distinguished
- [ ] Test 29: Broken link shows `?` in size column
- [ ] Test 30: Enter on broken link does nothing
- [ ] Test 31: Long link target path truncates with `...`
- [ ] Test 32: Relative link paths resolve correctly

### Loading Display Tests

- [ ] Test 33: Loading display shows in header line 2 for large directories
- [ ] Test 34: Mark info and free space hidden during loading
- [ ] Test 35: Application remains responsive during loading
- [ ] Test 36: Loading display disappears after loading complete
- [ ] Test 37: Header line 2 returns to normal display after loading
- [ ] Test 38: File count displayed during loading (optional feature)
- [ ] Test 39: Can cancel loading with quit command

### Layout and Responsive Tests

- [ ] Test 40: Column widths auto-adjust to content
- [ ] Test 41: Layout responds to terminal resize
- [ ] Test 42: Long filenames truncate appropriately
- [ ] Test 43: Long link targets truncate appropriately
- [ ] Test 44: Very narrow terminal (< 60 columns) shows name only
- [ ] Test 45: Very wide terminal uses space effectively
- [ ] Test 46: Terminal resize triggers mode recalculation

### Edge Cases

- [ ] Test 47: Empty directory displays correctly
- [ ] Test 48: Directory with only `..` displays correctly
- [ ] Test 49: Files with no owner/group (deleted user) display placeholder
- [ ] Test 50: Files with special characters in name display correctly
- [ ] Test 51: Symlink pointing to symlink (chain) handled correctly
- [ ] Test 52: Circular symlink (link to ancestor) handled correctly
- [ ] Test 53: Very large directory (10,000+ files) loads successfully

### Integration Tests

- [ ] Test 54: Display mode persists during navigation within session
- [ ] Test 55: Display mode resets on app restart
- [ ] Test 56: Header updates after file operations
- [ ] Test 57: Free space updates after large file operations
- [ ] Test 58: Free space updates every 5 seconds automatically
- [ ] Test 59: All features work together without conflicts
- [ ] Test 60: `i` key behavior correct when switching between narrow/wide terminals

## Success Criteria

- [ ] Header displays mark information and free space
- [ ] When terminal is wide, `i` key toggles Mode A ⇔ Mode B
- [ ] When terminal is narrow, display automatically switches to minimal mode
- [ ] When terminal is narrow, `i` key is disabled
- [ ] Each pane toggles display mode independently
- [ ] Mode A displays file size and timestamp appropriately
- [ ] Mode B displays permissions, owner, and group
- [ ] Can navigate to symbolic link targets
- [ ] Broken links are visually identifiable
- [ ] Loading display appears in header line 2 for large directories
- [ ] Application remains usable with narrow terminal width
- [ ] All test scenarios pass
- [ ] No performance degradation with typical directories (< 1000 files)
- [ ] Loading large directories (10,000+ files) shows feedback within 0.5 seconds
- [ ] Timestamp displays correctly in `2024-12-17 22:28` format (ISO 8601)

## Implementation Phases

### Phase 1: Core Features (This Specification)

**Step 1: Data Structure Extensions**
1. Extend `FileEntry` struct with owner, group, symlink fields
2. Extend `Pane` struct with display mode and loading state
3. Extend `Model` struct with disk space fields

**Step 2: File Information Gathering**
1. Implement `GetFileOwnerGroup()`
2. Implement `GetSymlinkInfo()`
3. Implement `GetDiskSpace()`
4. Update `ReadDirectory()` to gather all file information

**Step 3: Formatting Functions**
1. Implement `FormatSize()` with base-1024 conversion
2. Implement `FormatTimestamp()` with fixed `"2006-01-02 15:04"` format (ISO 8601)
3. Implement `FormatPermissions()` with Unix-style output

**Step 4: Display Mode Implementation**
1. Add `DisplayMode` enum with three states (Minimal, Basic, Detail)
2. Implement terminal width calculation logic
3. Implement `ToggleDisplayMode()` to switch between Mode A and Mode B
4. Update `Pane.RenderEntry()` to respect display mode
5. Add `i` key handler to toggle mode (only when terminal is wide)
6. Implement automatic minimal mode when terminal is narrow
7. Implement mode memory when terminal width changes

**Step 5: Header Enhancement**
1. Add second header line with mark info and free space
2. Implement disk space refresh (5-second interval)
3. Update `Pane.View()` to render new header
4. Implement header line 2 switching between normal and loading states

**Step 6: Symbolic Link Handling**
1. Update entry display to show link targets
2. Implement broken link visual distinction
3. Update `EnterDirectory()` to follow symlinks appropriately

**Step 7: Loading Display**
1. Add loading state to Pane
2. Implement async directory loading with goroutine
3. Add loading message display in header line 2
4. Implement progress tracking (file count or percentage)
5. Switch header line 2 back to normal after loading complete
6. Add loading timeout (if needed)

**Step 8: Testing and Refinement**
1. Unit tests for formatting functions
2. Integration tests for display modes
3. Manual testing with various directory types
4. Performance testing with large directories
5. Cross-platform testing (Linux, macOS, Windows)

### Phase 2: Mark Feature (Future)

1. Implement file marking with `Space` key
2. Calculate marked file count and total size
3. Update header to display marked file information
4. Implement batch operations on marked files

### Phase 3: Configuration File (Future)

1. Add free space refresh rate configuration
2. Add default display mode configuration
3. Add color scheme configuration

## Open Questions

- [ ] Should we display total directory size (calculated) or just `-`?
- [ ] Should broken symlinks be clickable to show error details?
- [ ] Should we add icons/symbols for file types (not just colors)?
- [ ] Should display mode be saved per-directory (like `ls` views)?
- [ ] Should we implement `du`-like directory size calculation?
- [ ] How to handle network file systems (NFS, SMB) with slow stat calls?
- [ ] Should we implement parallel file info gathering for large directories?
- [ ] What exact threshold (in columns) should trigger minimal mode? (current: 60)

## Future Enhancements (Post-Implementation)

### Short-term (Phase 2)
1. File marking and batch operations
2. Status bar mode indicator (if needed after user feedback)
3. Customizable colors for different file types
4. Display total selected file size in status bar

### Medium-term (Phase 3)
1. Configuration file support
2. Customizable refresh rates
3. Remember last display mode per session
4. Sorting by size, date, name, etc.

### Long-term
1. Directory size calculation option
2. Network file system optimization
3. Extended attributes display
4. File type icons/symbols
5. Custom color schemes
6. Plugin system for custom formatters

## References

- [Total Commander](https://www.ghisler.com/) - UI reference
- [Midnight Commander](https://midnight-commander.org/) - Feature reference
- [Go os package](https://pkg.go.dev/os) - File operations
- [Go syscall package](https://pkg.go.dev/syscall) - Disk space
- [Go os/user package](https://pkg.go.dev/os/user) - Owner/group lookup
- [Bubble Tea Documentation](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss Documentation](https://github.com/charmbracelet/lipgloss) - Styling
- Existing specification: `doc/tasks/ui-design/SPEC.md`
