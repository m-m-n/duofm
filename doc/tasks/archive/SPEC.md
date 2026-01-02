# Feature: Archive Operations

## Overview

Add comprehensive archive compression and extraction capabilities to duofm, enabling users to create and extract archives in various formats (tar, tar.gz, tar.bz2, tar.xz, zip, 7z) directly from the TUI interface. The feature integrates seamlessly with the existing dual-pane architecture and context menu system, allowing users to compress files to the opposite pane and extract archives efficiently.

This feature follows the UNIX philosophy by leveraging external CLI tools (tar, gzip, bzip2, xz, zip, 7z) for all archive operations. This approach ensures reliability, simplicity, and compatibility with standard tools that Linux users are familiar with.

**Target Platform:** Linux only (Windows/macOS are not supported)

## Objectives

- Provide seamless archive creation (tar, tar.gz, tar.bz2, tar.xz, zip, 7z) and extraction (tar, tar.gz, tar.bz2, tar.xz, zip, 7z) within the TUI
- Integrate archive operations into the existing context menu (`@` key)
- Support both single file/directory and multi-file batch compression via mark selection
- Implement smart extraction logic that adapts to archive structure
- Display real-time progress for long-running operations
- Execute archive operations asynchronously to maintain UI responsiveness
- Ensure proper error handling and user feedback

## User Stories

### US1: Compress a Directory
As a developer, I want to compress a project directory into a tar.xz archive, so that I can share it efficiently or archive it for backup.

**Acceptance Criteria:**
- [ ] User can select a directory and open context menu with `@` key
- [ ] "Compress" option appears in the context menu
- [ ] Submenu allows selection of tar, tar.xz, zip, or 7z format
- [ ] For tar.xz/zip/7z, compression level (0-9) can be selected
- [ ] Default archive name is presented and editable
- [ ] Archive is created in the opposite pane
- [ ] Progress is displayed during compression
- [ ] User receives notification upon completion

### US2: Compress Multiple Files
As a system administrator, I want to compress multiple log files into a single archive, so that I can organize and compress related files together.

**Acceptance Criteria:**
- [ ] User can mark multiple files using Space key
- [ ] Context menu shows "Compress N files" where N is the count
- [ ] All marked files are included in the archive at root level
- [ ] Default archive name is based on parent directory or timestamp
- [ ] Progress shows "X/N files (Y%)" during compression

### US3: Extract an Archive
As a user, I want to extract a tar.xz archive to the opposite pane, so that I can access the archived files.

**Acceptance Criteria:**
- [ ] User can select an archive file and open context menu
- [ ] "Extract archive" option appears for supported formats
- [ ] Smart extraction logic is applied (single directory vs. multiple files)
- [ ] Progress is displayed during extraction
- [ ] Extracted files appear in the opposite pane
- [ ] Symlinks are preserved correctly

### US4: Handle Existing Files
As a user, when creating an archive that already exists, I want to be prompted for action, so that I don't accidentally overwrite important files.

**Acceptance Criteria:**
- [ ] Dialog appears when archive name conflicts
- [ ] Options: Overwrite, Rename, Cancel
- [ ] Rename option allows entering new name with suggested default
- [ ] File information (size, modified date) is displayed

### US5: Cancel Long Operations
As a user, I want to cancel a long-running compression operation, so that I can abort if I realize I selected wrong files.

**Acceptance Criteria:**
- [ ] Esc key cancels operation during progress display
- [ ] Partial archive file is automatically deleted
- [ ] Cancellation notification is displayed
- [ ] UI returns to normal state immediately

## Functional Requirements

### FR1: Archive Creation

**FR1.1:** System shall support compression in the following formats using external CLI tools:
- tar (uncompressed archive) - using `tar -cvf`
- tar.gz (gzip compression) - using `tar -czvf`
- tar.bz2 (bzip2 compression) - using `tar -cjvf`
- tar.xz (LZMA2 compression) - using `tar -cJvf`
- zip (deflate compression) - using `zip -r`
- 7z (LZMA2 compression) - using `7z a`

**FR1.2:** System shall allow compression of:
- Single files
- Single directories (recursive)
- Multiple files/directories selected via mark selection

**FR1.3:** Compressed archives shall be created in the opposite pane's current directory.

**FR1.4:** System shall preserve:
- File permissions (Unix permission bits)
- File timestamps (modification time)
- Symlinks (as symlinks, not their targets)
- Directory structure

**FR1.5:** For multi-file compression, all files shall be placed at the root level of the archive.

**FR1.6:** System shall validate:
- Source files/directories exist and are readable
- Destination directory is writable
- Sufficient disk space is available
- Archive name is valid (non-empty, no invalid characters)

### FR2: Archive Extraction

**FR2.1:** System shall support extraction of the following formats using external CLI tools:
- tar (uncompressed) - using `tar -xvf`
- tar.gz (gzip compression) - using `tar -xzvf`
- tar.bz2 (bzip2 compression) - using `tar -xjvf`
- tar.xz (LZMA2 compression) - using `tar -xJvf`
- zip (deflate compression) - using `unzip`
- 7z (LZMA2 compression) - using `7z x`

**FR2.2:** System shall implement smart extraction logic:
- If archive root contains a single directory: extract directory contents directly to destination
- If archive root contains multiple files/directories: create directory named after archive (without extension) and extract into it

**FR2.3:** System shall detect archive format by:
1. File extension (primary method)
2. Magic number/signature (fallback for ambiguous cases)

**FR2.4:** System shall preserve during extraction:
- File permissions (except setuid/setgid bits for security)
- File timestamps
- Symlinks (as symlinks)
- Directory structure

**FR2.5:** System shall validate:
- Archive file exists and is readable
- Archive format is supported
- Archive is not corrupted (basic integrity check)
- Destination directory is writable
- Sufficient disk space is available (estimated from archive size)

**FR2.6:** System shall implement security measures:
- Reject paths containing ".." (path traversal prevention)
- Warn about absolute path symlinks
- Check compression ratio and display warning dialog if > 1:1000 (potential zip bomb detection)
- Check available disk space and display warning dialog if insufficient
- Ignore setuid/setgid bits during extraction

**FR2.7:** System shall perform pre-extraction safety checks:
- Parse archive metadata using format-specific commands (tar -tvf, unzip -l, 7z l)
- Calculate total extracted size from metadata
- Compare extracted size with archive size for compression bomb detection
- Compare extracted size with available disk space on destination

### FR3: Compression Level Selection

**FR3.1:** System shall allow compression level selection (0-9) for:
- tar.gz format (via gzip options)
- tar.bz2 format (via bzip2 options)
- tar.xz format (via xz options)
- zip format (via zip -N option)
- 7z format (via 7z -mx=N option)

**FR3.2:** tar format (uncompressed) shall not have compression level selection (not applicable).

**FR3.3:** Default compression level shall be 6 (balanced).

**FR3.4:** System shall provide level descriptions:
- 0: No compression (fastest)
- 1-3: Fast compression
- 4-6: Normal compression (recommended)
- 7-9: Best compression (slowest)

**FR3.5:** User shall be able to skip level selection by pressing Esc (defaults to level 6).

### FR4: Archive Naming

**FR4.1:** System shall generate default archive names:
- Single file/directory: `{original_name}.{extension}`
- Multiple files: `{parent_directory_name}.{extension}` or `archive_YYYY-MM-DD.{extension}` if parent name is not descriptive

**FR4.2:** System shall present default name in editable input field.

**FR4.3:** System shall allow user to:
- Edit name using standard text editing keys
- Confirm with Enter
- Cancel with Esc

**FR4.4:** System shall validate archive name:
- Must not be empty
- Must not contain invalid characters (NUL, control characters, OS-specific invalid chars)
- Must not conflict with existing files (or prompt for overwrite)

### FR5: Conflict Resolution

**FR5.1:** When target archive file already exists, system shall display dialog with:
- File information (name, size, modification date)
- Three options: Overwrite, Rename, Cancel

**FR5.2:** Overwrite option shall replace existing file after user confirmation.

**FR5.3:** Rename option shall:
- Re-display name input dialog
- Suggest name with sequential number suffix (e.g., `archive_1.tar.xz`)
- Re-check for conflicts after new name is entered

**FR5.4:** Cancel option shall abort the entire operation with no changes.

### FR6: Progress Display

**FR6.1:** System shall display progress dialog for operations that:
- Process more than 10 files, OR
- Process files larger than 10 MB total

**FR6.2:** Progress dialog shall show:
- Operation type (Compressing/Extracting)
- Target archive name
- Progress bar with percentage (0-100%)
- Current file being processed (truncated if path is too long)
- File count: "X/N files" where X is processed and N is total
- Elapsed time in MM:SS format
- Estimated remaining time (if calculable)

**FR6.3:** Progress shall update at most 10 times per second (100ms minimum interval).

**FR6.4:** Progress dialog shall show "[Esc] Cancel" option.

**FR6.5:** Small files (< 1 MB) may be processed without individual progress updates for performance.

**FR6.6:** Progress information fallback behavior:
- If progress information cannot be obtained (e.g., external command output cannot be parsed), the archive operation shall continue without interruption
- In such cases, progress display shall fall back to an indeterminate display showing "Processing..." instead of a percentage
- The progress dialog shall indicate that the operation is in progress even without detailed progress information
- This ensures that variations in external tool versions or output formats do not prevent archive operations from completing successfully

### FR7: Background Processing

**FR7.1:** Archive operations shall execute asynchronously in a goroutine.

**FR7.2:** UI shall remain responsive during archive operations (key input response < 100ms).

**FR7.3:** User shall be able to:
- Navigate through panes
- Browse directories
- View file information
- NOT start another archive operation until current one completes

**FR7.4:** System shall use channels for progress communication between goroutine and UI.

**FR7.5:** On completion, system shall:
- Display notification for 5 seconds (or until next action)
- Refresh file list in the destination pane
- Clear any marks in the source pane

### FR8: Operation Cancellation

**FR8.1:** User shall be able to cancel operation at any time by pressing Esc.

**FR8.2:** On cancellation, system shall:
- Stop the archive operation immediately
- Delete partially created archive file
- Display cancellation notification
- Return to normal state

**FR8.3:** Cancellation shall be acknowledged within 1 second.

### FR9: Error Handling

**FR9.1:** System shall handle and report errors for:
- File not found
- Permission denied (read or write)
- Disk space insufficient
- Corrupted archive file
- Unsupported archive format
- I/O errors
- Invalid file names
- Path traversal attempts

**FR9.2:** Error messages shall be:
- User-friendly (avoid technical jargon when possible)
- Specific (state exactly what went wrong)
- Actionable (suggest what user can do)

**FR9.3:** On error during compression/extraction:
- Display error dialog with details
- Delete partial archive/extracted files
- Log detailed error to log file
- Allow user to acknowledge and return to normal state

**FR9.4:** Transient errors (e.g., temporary network drive disconnection) shall automatically retry up to 3 times with 1-second delays.

### FR10: Context Menu Integration

**FR10.1:** "Compress" menu item shall appear in context menu for:
- Any file or directory
- When multiple files/directories are marked

**FR10.2:** "Compress" shall open a submenu with format options:
1. as tar (no compression)
2. as tar.gz (gzip compression)
3. as tar.bz2 (bzip2 compression)
4. as tar.xz (LZMA compression)
5. as zip (deflate compression) - only if zip command is available
6. as 7z (LZMA2 compression) - only if 7z command is available

**FR10.3:** "Extract archive" menu item shall appear ONLY when:
- Selected file has supported archive extension (.tar, .tar.gz, .tar.bz2, .tar.xz, .zip, .7z)
- File is readable

**FR10.4:** Menu items shall be numbered and selectable via:
- j/k keys for navigation
- 1-9 numeric keys for direct selection
- Enter key for confirmation
- Esc key for cancellation

## Non-Functional Requirements

### NFR1: Performance

**NFR1.1:** Small files (< 10 MB) shall compress within 3 seconds on typical hardware.

**NFR1.2:** UI shall respond to keyboard input within 100ms even during active archive operations.

**NFR1.3:** Progress updates shall occur at most 10 times per second to avoid UI flicker.

**NFR1.4:** Memory usage during compression/extraction shall not exceed:
- 64 MB buffer for streaming operations
- Additional memory proportional to number of files (estimate: 1 KB per file for metadata)

**NFR1.5:** Archive processing shall use streaming where possible to minimize memory footprint.

### NFR2: Security

**NFR2.1:** Path traversal prevention:
- Reject extraction of paths containing ".." segments
- Normalize all paths before extraction

**NFR2.2:** Symlink safety:
- Preserve symlinks as-is (don't follow)
- Warn if archive contains absolute path symlinks
- Ensure symlinks point within extraction directory

**NFR2.3:** Zip bomb protection:
- Check compression ratio before extraction using metadata commands
- Display warning dialog for archives with ratio > 1:1000 (does not block, user can continue)
- No fixed maximum extraction size limit (rely on disk space check instead)

**NFR2.3.1:** Disk space protection:
- Check available disk space on destination before extraction
- Display warning dialog if extracted size exceeds available space (does not block, user can continue)

**NFR2.4:** Permission handling:
- Ignore setuid and setgid bits during extraction
- Apply umask to extracted file permissions
- Never create world-writable files

**NFR2.5:** Input validation:
- Validate all file names for invalid characters
- Reject control characters and NUL bytes
- Enforce OS-specific path constraints

### NFR3: Reliability

**NFR3.1:** Atomic operations:
- Archive creation shall be atomic (temp file renamed on success)
- Failed operations shall leave no partial files

**NFR3.2:** Error recovery:
- All errors shall be caught and handled gracefully
- No operation shall cause application crash
- Panic recovery shall log error and notify user

**NFR3.3:** Data integrity:
- Verify archive integrity before extraction when possible
- Preserve all file attributes (permissions, timestamps)
- Handle symlinks consistently

**NFR3.4:** Retry logic:
- Transient errors shall auto-retry up to 3 times
- Retry delay: 1 second between attempts
- Permanent errors shall fail immediately

### NFR4: Usability

**NFR4.1:** Progress feedback:
- All operations > 2 seconds shall show progress
- Progress shall include time estimates when possible
- User shall always know current state

**NFR4.2:** Cancellability:
- All long operations shall be cancellable via Esc key
- Cancellation shall respond within 1 second

**NFR4.3:** Error messages:
- Clear, concise, non-technical language
- Include specific cause and suggested action
- Display in consistent dialog format

**NFR4.4:** Default values:
- All dialogs shall provide sensible defaults
- User can accept defaults quickly with minimal input

### NFR5: Compatibility

**NFR5.1:** Archive format compliance:
- tar: POSIX.1-2001 (ustar) format
- zip: PKZIP 2.0+ compatible
- Character encoding: UTF-8 for file names

**NFR5.2:** Platform compatibility:
- Linux: full feature support (only supported platform)
- macOS/Windows: not supported

**NFR5.3:** Archive portability:
- Archives created by duofm shall be extractable by standard tools (tar, unzip, 7z)
- Archives created by standard tools shall be extractable by duofm

## Implementation Approach

### Architecture

**Layered Architecture (Simplified with External CLI Tools):**
```
┌─────────────────────────────────────────────┐
│           UI Layer (Bubble Tea)             │
│  - ContextMenuDialog                        │
│  - ProgressDialog                           │
│  - ConfirmDialog                            │
│  - InputDialog                              │
├─────────────────────────────────────────────┤
│        Application Layer (Handlers)         │
│  - ArchiveController                        │
│  - TaskManager                              │
│  - ProgressTracker                          │
├─────────────────────────────────────────────┤
│       Archive Service (CLI Wrapper)         │
│  - CommandExecutor (exec.Command wrapper)   │
│  - FormatDetector                           │
│  - SmartExtractor                           │
│  - CommandAvailability                      │
├─────────────────────────────────────────────┤
│      External CLI Tools                     │
│  - tar, gzip, bzip2, xz                    │
│  - zip, unzip                              │
│  - 7z (p7zip)                              │
└─────────────────────────────────────────────┘
```

**Design Philosophy (UNIX Philosophy):**
- Each tool does one thing well
- Leverage existing, battle-tested CLI tools
- Simple wrapper layer for Go integration
- No external Go libraries for archive handling

### Component Diagram

```
┌──────────────┐         ┌──────────────┐
│   Context    │         │   Progress   │
│     Menu     │─────────│    Dialog    │
└──────┬───────┘         └──────────────┘
       │
       │ trigger
       ▼
┌──────────────────────────────────────┐
│       ArchiveController              │
│  + CreateArchive()                   │
│  + ExtractArchive()                  │
│  + SelectFormat()                    │
│  + SelectCompressionLevel()          │
│  + CheckAvailableFormats()           │
└──────┬───────────────────────────────┘
       │
       │ delegates
       ▼
┌──────────────────────────────────────┐
│         TaskManager                  │
│  + StartTask(task)                   │
│  + CancelTask(taskID)                │
│  + GetProgress(taskID)               │
└──────┬───────────────────────────────┘
       │
       │ executes via exec.Command
       ▼
┌──────────────────────────────────────┐
│       CommandExecutor                │
│  + ExecuteTar(args)                  │
│  + ExecuteZip(args)                  │
│  + Execute7z(args)                   │
│  + ParseOutput(stdout, stderr)       │
└──────────────────────────────────────┘
       │
       │ uses
       ▼
┌───────────────────────────────────────┐
│  External CLI Tools                   │
│  tar | gzip | bzip2 | xz | zip | 7z  │
└───────────────────────────────────────┘
       │
       │ detected by
       ▼
┌──────────────────┐    ┌──────────────────┐
│ CommandAvailability   │  FormatDetector  │
│  + CheckCommand()     │  + DetectFormat()│
│  + GetAvailable()     │  + IsSupportedExt│
└──────────────────┘    └──────────────────┘
```

### Data Flow

#### Compression Flow
```
User Action (@+Compress)
  → Context Menu (Format Selection - based on available commands)
  → Compression Level Dialog (if tar.gz/tar.bz2/tar.xz/zip/7z)
  → Archive Name Input Dialog
  → Conflict Check
    ├─ Exists → Overwrite/Rename/Cancel Dialog
    └─ Not Exists → Continue
  → ArchiveController.CreateArchive()
  → TaskManager.StartTask()
    → Spawn goroutine
      → CommandExecutor.Execute{Tar|Zip|7z}()
        → exec.Command() with appropriate args
        → Parse stdout/stderr for progress
        → Report progress via channel
      → On completion:
        → Refresh file list
        → Show notification
      → On error:
        → Delete partial file
        → Show error dialog
```

#### Extraction Flow
```
User Action (@+Extract on archive file)
  → ArchiveController.ExtractArchive()
  → FormatDetector.DetectFormat()
  → SmartExtractor.AnalyzeStructure() (using tar -tvf or similar)
    ├─ Single root directory → Extract directly
    └─ Multiple items → Create directory
  → TaskManager.StartTask()
    → Spawn goroutine
      → CommandExecutor.Execute{Tar|Unzip|7z}()
        → exec.Command() with appropriate args
        → Parse stdout/stderr for progress
        → Report progress via channel
      → On completion:
        → Refresh file list
        → Show notification
      → On error:
        → Delete partial extraction
        → Show error dialog
```

### API Design

#### ArchiveController Interface

```go
type ArchiveController interface {
    // CreateArchive initiates archive creation process
    CreateArchive(sources []string, destDir string, format ArchiveFormat, level int) (taskID string, err error)

    // ExtractArchive initiates archive extraction process
    ExtractArchive(archivePath string, destDir string) (taskID string, err error)

    // CancelTask cancels a running archive task
    CancelTask(taskID string) error

    // GetTaskProgress returns current progress of a task
    GetTaskProgress(taskID string) (*TaskProgress, error)
}
```

#### CommandExecutor Interface

```go
type CommandExecutor interface {
    // ExecuteCompress runs compression command
    ExecuteCompress(ctx context.Context, sources []string, output string, opts CompressOptions) error

    // ExecuteExtract runs extraction command
    ExecuteExtract(ctx context.Context, archivePath string, destDir string, opts ExtractOptions) error

    // ListArchiveContents lists archive contents (for smart extraction)
    ListArchiveContents(archivePath string, format ArchiveFormat) ([]string, error)
}

type CompressOptions struct {
    Format           ArchiveFormat
    CompressionLevel int
    ProgressChan     chan<- ProgressUpdate
}

type ExtractOptions struct {
    SmartExtraction bool // Use smart extraction logic
    ProgressChan    chan<- ProgressUpdate
}
```

#### CommandAvailability Interface

```go
type CommandAvailability interface {
    // CheckCommand checks if a command is available
    CheckCommand(cmd string) bool

    // GetAvailableFormats returns formats available for compression/extraction
    GetAvailableFormats(operation Operation) []ArchiveFormat

    // IsFormatAvailable checks if a specific format is available
    IsFormatAvailable(format ArchiveFormat, operation Operation) bool
}

// Required commands for each format
var formatCommands = map[ArchiveFormat][]string{
    FormatTar:    {"tar"},
    FormatTarGz:  {"tar", "gzip"},
    FormatTarBz2: {"tar", "bzip2"},
    FormatTarXz:  {"tar", "xz"},
    FormatZip:    {"zip", "unzip"},
    Format7z:     {"7z"},
}
```

#### FormatDetector Interface

```go
type FormatDetector interface {
    // DetectFormat determines archive format from file
    DetectFormat(filePath string) (ArchiveFormat, error)

    // IsSupportedFormat checks if format is supported for operation
    IsSupportedFormat(format ArchiveFormat, operation Operation) bool
}

type Operation int
const (
    OperationCompress Operation = iota
    OperationExtract
)
```

#### SmartExtractor Interface

```go
type SmartExtractor interface {
    // AnalyzeStructure examines archive and determines extraction strategy
    // Uses: tar -tvf, unzip -l, 7z l
    AnalyzeStructure(archivePath string, format ArchiveFormat) (*ExtractionStrategy, error)
}

type ExtractionStrategy struct {
    Method        ExtractionMethod // Direct or CreateDirectory
    DirectoryName string           // Directory name if CreateDirectory
}

type ExtractionMethod int
const (
    ExtractDirect ExtractionMethod = iota // Single root dir
    ExtractToDirectory                     // Multiple items
)
```

### Database Schema

No database is required. All state is managed in memory during operation lifetime.

### Dependencies

**Internal Dependencies:**
- `internal/fs`: File system operations (Copy, Move, Delete, file metadata)
- `internal/ui`: UI components (Dialog interface, Bubble Tea models)
- `internal/config`: Configuration (if adding archive preferences)

**External CLI Tools (Required):**

| Command | Package (Debian/Ubuntu) | Package (Fedora/RHEL) | Package (Arch) | Purpose |
|---------|-------------------------|------------------------|----------------|---------|
| tar | tar | tar | tar | tar archive creation/extraction |
| gzip | gzip | gzip | gzip | gzip compression |
| bzip2 | bzip2 | bzip2 | bzip2 | bzip2 compression |
| xz | xz-utils | xz | xz | LZMA compression |
| zip | zip | zip | zip | zip compression (optional) |
| unzip | unzip | unzip | unzip | zip extraction (optional) |
| 7z | p7zip-full | p7zip p7zip-plugins | p7zip | 7z handling (optional) |

**Go Dependencies (Minimal):**

| Package | Version | Purpose | License |
|---------|---------|---------|---------|
| `os/exec` | stdlib | Execute external commands | BSD |
| `context` | stdlib | Cancellation handling | BSD |

**Design Justification (UNIX Philosophy):**
- Leverage battle-tested CLI tools instead of reimplementing in Go
- Simpler codebase with fewer external Go library dependencies
- Better compatibility with existing archive tools
- Easier maintenance and security updates (via system packages)
- No need to handle complex archive format parsing in Go

### File Structure

```
internal/
├── archive/
│   ├── archive.go              # Main archive controller
│   ├── archive_test.go
│   ├── command_executor.go     # CLI command execution wrapper
│   ├── command_executor_test.go
│   ├── command_availability.go # Check available commands
│   ├── command_availability_test.go
│   ├── format.go               # Format detection and constants
│   ├── format_test.go
│   ├── smart_extractor.go      # Smart extraction logic
│   ├── smart_extractor_test.go
│   ├── task_manager.go         # Background task management
│   ├── task_manager_test.go
│   ├── progress.go             # Progress tracking
│   ├── progress_test.go
│   └── errors.go               # Error definitions
├── ui/
│   ├── archive_progress_dialog.go      # Progress display
│   ├── archive_progress_dialog_test.go
│   ├── compression_level_dialog.go     # Level selection
│   ├── compression_level_dialog_test.go
│   ├── archive_name_dialog.go          # Name input
│   ├── archive_name_dialog_test.go
│   ├── overwrite_dialog.go             # Conflict resolution
│   ├── overwrite_dialog_test.go
│   └── context_menu_dialog.go          # Updated with archive items
└── fs/
    └── operations.go           # May add helper functions if needed
```

## Test Scenarios

### Unit Tests

**Compression (via CLI):**
- [ ] Test tar creation from single file
- [ ] Test tar creation from single directory
- [ ] Test tar creation from multiple files
- [ ] Test tar.gz creation with each compression level (0-9)
- [ ] Test tar.bz2 creation with each compression level (0-9)
- [ ] Test tar.xz creation with each compression level (0-9)
- [ ] Test zip creation with each compression level (0-9)
- [ ] Test 7z creation with each compression level (0-9)
- [ ] Test compression when required CLI is not available (should fail gracefully)
- [ ] Test symlink preservation in archives
- [ ] Test file permission preservation
- [ ] Test timestamp preservation
- [ ] Test empty directory handling
- [ ] Test large file handling (> 1 GB, use mock)

**Command Availability:**
- [ ] Test CheckCommand for existing command
- [ ] Test CheckCommand for non-existing command
- [ ] Test GetAvailableFormats returns correct formats
- [ ] Test format availability based on installed commands

**Extraction:**
- [ ] Test tar extraction
- [ ] Test tar.gz extraction
- [ ] Test tar.bz2 extraction
- [ ] Test tar.xz extraction
- [ ] Test zip extraction
- [ ] Test 7z extraction
- [ ] Test smart extraction: single root directory
- [ ] Test smart extraction: multiple root items
- [ ] Test symlink restoration
- [ ] Test permission restoration
- [ ] Test timestamp restoration

**Format Detection:**
- [ ] Test detection by extension (.tar, .tar.gz, .tar.bz2, .tar.xz, .zip, .7z)
- [ ] Test detection by magic number
- [ ] Test unsupported format rejection
- [ ] Test corrupted file detection
- [ ] Test CLI availability detection for all formats

**Security:**
- [ ] Test path traversal rejection (../)
- [ ] Test absolute path rejection
- [ ] Test compression ratio check (zip bomb)
- [ ] Test setuid bit stripping
- [ ] Test symlink target validation

**Error Handling:**
- [ ] Test source file not found
- [ ] Test destination not writable
- [ ] Test disk space insufficient
- [ ] Test permission denied on read
- [ ] Test permission denied on write
- [ ] Test corrupted archive extraction
- [ ] Test I/O error during operation
- [ ] Test cancellation during operation

### Integration Tests

- [ ] Test complete compression flow: menu → format → level → name → create
- [ ] Test complete extraction flow: menu → extract → verify
- [ ] Test overwrite dialog flow: create → conflict → overwrite
- [ ] Test rename dialog flow: create → conflict → rename
- [ ] Test cancel during compression
- [ ] Test cancel during extraction
- [ ] Test progress updates during long operation
- [ ] Test background processing (UI remains responsive)
- [ ] Test multi-file mark and compress
- [ ] Test archive creation in opposite pane
- [ ] Test extraction to opposite pane
- [ ] Test file list refresh after operation

### E2E Tests

**E2E Test 1: Compress Single Directory**
```bash
# Start duofm, navigate to testdir
start_duofm "$CURRENT_SESSION"
send_keys "$CURRENT_SESSION" "j" "j"  # Navigate to directory
send_keys "$CURRENT_SESSION" "@"      # Open context menu
sleep 0.3
assert_contains "$CURRENT_SESSION" "Compress" "Compress option visible"
send_keys "$CURRENT_SESSION" "4"      # Select Compress
sleep 0.3
assert_contains "$CURRENT_SESSION" "as tar.xz" "Format submenu visible"
send_keys "$CURRENT_SESSION" "2"      # Select tar.xz
sleep 0.3
assert_contains "$CURRENT_SESSION" "Compression Level" "Level dialog visible"
send_keys "$CURRENT_SESSION" "Enter" # Accept default level
sleep 0.3
assert_contains "$CURRENT_SESSION" "Archive Name" "Name dialog visible"
send_keys "$CURRENT_SESSION" "Enter" # Accept default name
sleep 2.0                            # Wait for compression
send_keys "$CURRENT_SESSION" "Tab"   # Switch to opposite pane
assert_contains "$CURRENT_SESSION" ".tar.xz" "Archive created"
stop_duofm "$CURRENT_SESSION"
```

**E2E Test 2: Extract Archive**
```bash
start_duofm "$CURRENT_SESSION"
send_keys "$CURRENT_SESSION" "j"     # Navigate to archive
send_keys "$CURRENT_SESSION" "@"     # Open context menu
sleep 0.3
assert_contains "$CURRENT_SESSION" "Extract archive" "Extract option visible"
send_keys "$CURRENT_SESSION" "Enter" # Select extract
sleep 2.0                            # Wait for extraction
send_keys "$CURRENT_SESSION" "Tab"   # Switch to opposite pane
assert_contains "$CURRENT_SESSION" "extracted_dir" "Files extracted"
stop_duofm "$CURRENT_SESSION"
```

**E2E Test 3: Multi-file Compression**
```bash
start_duofm "$CURRENT_SESSION"
send_keys "$CURRENT_SESSION" "Space" "j" "Space" "j" "Space"  # Mark 3 files
send_keys "$CURRENT_SESSION" "@"
sleep 0.3
assert_contains "$CURRENT_SESSION" "Compress 3 files" "Multi-file option"
send_keys "$CURRENT_SESSION" "4" "3"  # Compress as zip
sleep 0.3
send_keys "$CURRENT_SESSION" "Enter" "Enter"  # Accept level and name
sleep 2.0
send_keys "$CURRENT_SESSION" "Tab"
assert_contains "$CURRENT_SESSION" ".zip" "Archive created"
stop_duofm "$CURRENT_SESSION"
```

**E2E Test 4: Overwrite Handling**
```bash
start_duofm "$CURRENT_SESSION"
# Create first archive
send_keys "$CURRENT_SESSION" "@" "4" "1" "Enter"  # Compress as tar
sleep 1.0
# Try to create again
send_keys "$CURRENT_SESSION" "Tab" "@" "4" "1" "Enter"
sleep 0.3
assert_contains "$CURRENT_SESSION" "already exists" "Overwrite dialog"
assert_contains "$CURRENT_SESSION" "Overwrite" "Overwrite option"
send_keys "$CURRENT_SESSION" "2"  # Select Rename
sleep 0.3
send_keys "$CURRENT_SESSION" "Enter"  # Accept renamed file
sleep 1.0
assert_contains "$CURRENT_SESSION" "_1.tar" "Renamed archive created"
stop_duofm "$CURRENT_SESSION"
```

**E2E Test 5: Cancel Operation**
```bash
start_duofm "$CURRENT_SESSION"
send_keys "$CURRENT_SESSION" "@" "4" "2" "Enter" "Enter"  # Start compression
sleep 0.5  # Let it start
assert_contains "$CURRENT_SESSION" "Compressing" "Progress dialog visible"
send_keys "$CURRENT_SESSION" "Escape"  # Cancel
sleep 0.5
assert_contains "$CURRENT_SESSION" "Cancelled" "Cancellation notification"
send_keys "$CURRENT_SESSION" "Tab"
assert_not_contains "$CURRENT_SESSION" ".tar.xz" "Partial file deleted"
stop_duofm "$CURRENT_SESSION"
```

### Edge Cases

- [ ] Empty file (0 bytes) compression and extraction
- [ ] Very large file (> 1 GB) - with mocked I/O for speed
- [ ] Deep directory hierarchy (> 100 levels)
- [ ] Many files (> 10,000 files)
- [ ] Long file name (255 characters)
- [ ] File name with special characters (spaces, unicode)
- [ ] Archive containing only symlinks
- [ ] Archive with broken symlinks
- [ ] Circular directory symlinks
- [ ] Archive with no read permission on some files
- [ ] Destination directory becomes read-only during operation
- [ ] Disk fills up during compression/extraction

### Performance Tests

- [ ] Measure compression time for 100 MB data (target: < 10 seconds at level 6)
- [ ] Measure extraction time for 100 MB archive (target: < 5 seconds)
- [ ] Verify UI responsiveness: key input response < 100ms during operation
- [ ] Measure progress update frequency (should be ≤ 10 Hz)
- [ ] Measure memory usage during compression of 1000 files (target: < 100 MB)
- [ ] Test cancellation response time (target: < 1 second)

## Security Considerations

**Path Traversal Prevention:**
- Validate all extracted paths for ".." components
- Convert to absolute paths and verify they're within destination
- Reject any path that escapes destination directory

**Symlink Safety:**
- Store symlinks as-is, not following them during compression
- During extraction, ensure symlink targets are relative and within extraction directory
- Warn user if absolute symlinks are encountered
- Prevent symlink loops during extraction

**Compression Bomb Protection:**
- Calculate compression ratio: uncompressed_size / compressed_size
- Display warning dialog if ratio > 1:1000 with option to continue or cancel
- No fixed maximum extraction size limit
- Track cumulative extracted size during extraction

**Disk Space Protection:**
- Calculate total extracted size from archive metadata
- Compare with available disk space on destination
- Display warning dialog if insufficient space with option to continue or cancel

**Pre-extraction Metadata Commands:**
- tar/tar.gz/tar.bz2/tar.xz: `tar -tvf` / `tar -tzvf` / `tar -tjvf` / `tar -tJvf`
- zip: `unzip -l`
- 7z: `7z l`

**Warning Dialog UI:**
```
Warning: Large extraction ratio detected

Archive size: 1 MB
Extracted size: 2 GB (ratio: 1:2000)

This may indicate a zip bomb or highly compressed data.
Do you want to continue?

[Continue] [Cancel]
```

```
Warning: Insufficient disk space

Required: 1.2 GB
Available: 500 MB

Do you want to continue anyway?

[Continue] [Cancel]
```

**Permission Handling:**
- Strip setuid and setgid bits during extraction
- Apply user's umask to extracted files
- Never create world-writable files (maximum: 0666 for files, 0777 for dirs, minus umask)

**Input Validation:**
- Validate file names: reject control characters, NUL bytes
- Enforce path length limits per OS
- Validate compression level range (0-9)
- Sanitize all user input in dialogs

**Error Information Disclosure:**
- Avoid exposing sensitive path information in error messages
- Log detailed errors but show sanitized messages to user
- Don't reveal internal system structure

## Error Handling

### Error Codes

| Code | Description | HTTP Status Equivalent | User Message |
|------|-------------|------------------------|--------------|
| ERR_ARCHIVE_001 | Source file not found | 404 | "Source file or directory not found: {name}" |
| ERR_ARCHIVE_002 | Permission denied (read) | 403 | "Cannot read source: Permission denied" |
| ERR_ARCHIVE_003 | Permission denied (write) | 403 | "Cannot write to destination: Permission denied" |
| ERR_ARCHIVE_004 | Disk space insufficient | 507 | "Insufficient disk space. Required: {size}" |
| ERR_ARCHIVE_005 | Unsupported format | 415 | "Unsupported archive format: {format}" |
| ERR_ARCHIVE_006 | Corrupted archive | 422 | "Archive file is corrupted or invalid" |
| ERR_ARCHIVE_007 | Invalid archive name | 400 | "Invalid archive name: {reason}" |
| ERR_ARCHIVE_008 | Path traversal detected | 403 | "Security error: Invalid file path in archive" |
| ERR_ARCHIVE_009 | Compression bomb detected | 413 | "Archive rejected: Excessive compression ratio" |
| ERR_ARCHIVE_010 | Operation cancelled | N/A | "Operation cancelled by user" |
| ERR_ARCHIVE_011 | I/O error | 500 | "File system error: {details}" |
| ERR_ARCHIVE_012 | Internal error | 500 | "Unexpected error occurred. Check logs." |

### Error Flow

```
Operation Start
  → Input Validation
    └─ Validation Error → Display Error Dialog → Return to Normal State
  → Pre-flight Checks (disk space, permissions)
    └─ Check Failed → Display Error Dialog → Return to Normal State
  → Begin Operation
    ├─ Success → Complete Normally
    └─ Runtime Error
        ├─ Transient Error (I/O timeout, network)
        │   └─ Auto Retry (max 3 times)
        │       ├─ Success → Continue
        │       └─ All Retries Failed → Cleanup → Display Error → Return
        └─ Permanent Error
            └─ Cleanup (delete partial files) → Display Error → Return
```

### Cleanup Procedures

**On Error:**
1. Stop current operation immediately
2. Close all open file handles
3. Delete partially created archive file (if compression)
4. Delete partially extracted files (if extraction)
5. Log detailed error information
6. Display user-friendly error message

**On Cancellation:**
1. Set cancellation flag
2. Wait for operation to acknowledge (check at regular intervals)
3. Perform same cleanup as error case
4. Display cancellation confirmation

## Performance Optimization

### Performance Goals

- Response time: < 100ms for all UI interactions during background operations
- Small file compression (< 10 MB): < 3 seconds
- Large file compression (100 MB): < 30 seconds at level 6
- Extraction: < 5 seconds for 100 MB archive
- Memory usage: < 100 MB for typical operations (< 1000 files)
- Progress update overhead: < 5% of total operation time

### Optimization Strategies

**Streaming I/O:**
- Use io.Copy and bufio.Reader/Writer for efficient data transfer
- Avoid loading entire files into memory
- Process archives incrementally

**Buffering:**
- Use 64 KB buffers for file I/O (balance between syscalls and memory)
- Reuse buffers where possible (sync.Pool)
- Tune buffer sizes based on operation type

**Concurrency:**
- Run compression/extraction in separate goroutine
- Use channels for progress communication
- Single-threaded archiving (most libraries are not thread-safe)
- Consider parallel compression for multiple independent files (future enhancement)

**Progress Optimization:**
- Batch progress updates (max 10 per second)
- Skip progress for very small files (< 1 MB)
- Pre-calculate total file count and size for accurate progress

**Memory Management:**
- Stream processing for large files
- Release resources explicitly (close files ASAP)
- Use context for cancellation propagation
- Limit concurrent operations to 1 to prevent memory spike

### Caching Strategy

No caching is needed for archive operations. Each operation is one-time and files are processed sequentially.

### Profiling Points

- Total operation time (compression/extraction)
- Individual file processing time (for large archives)
- Memory allocation during operation
- Goroutine count and channel buffer usage
- UI render time during background operation

## Success Criteria

- [ ] All functional requirements (FR1-FR10) are implemented and tested
- [ ] All non-functional requirements (NFR1-NFR5) are met
- [ ] All test scenarios pass (unit, integration, E2E)
- [ ] Performance meets specified goals
- [ ] Security requirements are satisfied
- [ ] Error handling covers all identified error cases
- [ ] Code review is completed with no major issues
- [ ] Documentation (godoc) is complete for all public APIs
- [ ] E2E tests demonstrate typical user workflows
- [ ] No memory leaks detected in long-running tests

**Acceptance Criteria Checklist:**
- [ ] Can compress single file/directory to tar, tar.gz, tar.bz2, tar.xz, zip, 7z
- [ ] Can compress multiple marked files to archive
- [ ] Can extract tar, tar.gz, tar.bz2, tar.xz, zip, 7z archives
- [ ] Format menu items only appear when required CLI tools are available
- [ ] Graceful handling when CLI tools are missing
- [ ] Smart extraction works correctly (single dir vs. multiple items)
- [ ] Progress bar shows accurate percentage and file count
- [ ] Background processing keeps UI responsive
- [ ] Compression level selection works (0-9)
- [ ] Archive naming allows editing with sensible defaults
- [ ] Overwrite dialog offers Overwrite/Rename/Cancel options
- [ ] Cancellation stops operation and cleans up partial files
- [ ] All errors display clear, actionable messages
- [ ] Symlinks are preserved correctly
- [ ] File permissions and timestamps are preserved
- [ ] Path traversal attacks are prevented
- [ ] Compression bombs are detected and warning is displayed (user can choose to continue)

## Open Questions

- [ ] Should we support password-protected zip archives in the future?
- [ ] Should we add a setting for default compression format/level?
- [ ] Should we keep a history of recent archive operations?
- [ ] Should we support drag-and-drop for files to compress (if terminal supports it)?
- [ ] Should very large archives (> 10 GB) have additional confirmation?
- [ ] Should we support streaming extraction (extract while downloading)?

## Implementation Phases

### Phase 1: Core Infrastructure (High Priority)
**Goals:** Establish foundation for archive operations

**Deliverables:**
- Archive format detection and constants
- Command availability checker (exec.LookPath wrapper)
- Basic command executor interface
- Task manager for background operations
- Progress tracking system
- Unit tests for core functionality

**Estimated Effort:** 2-3 days

### Phase 2: CLI Integration (High Priority)
**Goals:** Implement CLI wrappers for all formats

**Deliverables:**
- tar command wrapper (tar, tar.gz, tar.bz2, tar.xz)
- zip/unzip command wrapper
- 7z command wrapper
- Compression level argument handling
- Format-specific unit tests

**Estimated Effort:** 2-3 days

### Phase 3: UI Integration (High Priority)
**Goals:** Integrate archive operations into UI

**Deliverables:**
- Update context menu with Compress/Extract items
- Compression level selection dialog
- Archive name input dialog
- Overwrite confirmation dialog
- Progress display dialog
- UI integration tests

**Estimated Effort:** 3-4 days

### Phase 4: Smart Features (Medium Priority)
**Goals:** Implement intelligent behavior

**Deliverables:**
- Smart extraction logic (single dir vs. multiple items)
- Compression level selection with descriptions
- Default name generation logic
- Conflict resolution (rename with sequential numbers)
- Integration tests for smart features

**Estimated Effort:** 2-3 days

### Phase 5: Security and Error Handling (High Priority)
**Goals:** Ensure robust and secure operation

**Deliverables:**
- Path traversal prevention
- Compression bomb detection
- Symlink safety checks
- Comprehensive error handling
- Cleanup on failure/cancellation
- Security test scenarios

**Estimated Effort:** 2-3 days

### Phase 6: E2E Testing and Polish (High Priority)
**Goals:** Ensure production-ready quality

**Deliverables:**
- E2E test scripts for all workflows
- Performance testing and optimization
- Memory leak testing
- Documentation (godoc comments)
- User-facing documentation updates

**Estimated Effort:** 2-3 days

**Total Estimated Effort:** 12-17 days (reduced due to simpler CLI-based implementation)

## References

- duofm CONTRIBUTING.md: `/home/sakura/cache/worktrees/feature-add-archive/doc/CONTRIBUTING.md`
- duofm existing operations: `/home/sakura/cache/worktrees/feature-add-archive/internal/fs/operations.go`
- duofm context menu: `/home/sakura/cache/worktrees/feature-add-archive/internal/ui/context_menu_dialog.go`
- GNU tar Manual: https://www.gnu.org/software/tar/manual/
- gzip Manual: https://www.gnu.org/software/gzip/manual/
- bzip2 Manual: https://sourceware.org/bzip2/
- XZ Utils: https://tukaani.org/xz/
- Info-ZIP: https://infozip.sourceforge.net/
- p7zip project: https://github.com/p7zip-project/p7zip
- POSIX tar spec: https://pubs.opengroup.org/onlinepubs/9699919799/utilities/pax.html

---

**Last Updated:** 2026-01-01
**Status:** Draft (Updated: External CLI tool dependency, Linux-only, tar.gz/tar.bz2 compression support)
