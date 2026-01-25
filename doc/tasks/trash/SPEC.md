# Feature: Trash (Recycle Bin)

## Overview

Implement a trash/recycle bin feature compliant with the freedesktop.org Trash Specification. Users can safely move files to trash instead of permanent deletion, and restore them when needed. **Trash contents are displayed in a dedicated dialog**, keeping trash operations isolated from normal file management.

## Objectives

- Protect files from accidental deletion
- Enable recovery of deleted files
- Maintain compatibility with desktop environments (GNOME, KDE, etc.)
- Provide intuitive keyboard-driven trash operations
- Keep trash operations in a separate dialog to avoid keybinding conflicts

## User Stories

### US1: Move File to Trash
As a user, I want to move files to trash using the Delete key, so that I can recover them if needed.

**Acceptance Criteria:**
- [ ] Delete key moves selected file(s) to trash
- [ ] File appears in `~/.local/share/Trash/files/`
- [ ] Corresponding `.trashinfo` file is created
- [ ] File list updates immediately

### US2: Open Trash Dialog
As a user, I want to quickly open the trash dialog, so that I can see what I've deleted.

**Acceptance Criteria:**
- [ ] `T` key opens the trash dialog at screen center
- [ ] Both panes are dimmed behind the dialog
- [ ] Trash contents displayed with columns: Name, Size, Deleted, Original Path

### US3: Restore File from Trash
As a user, I want to restore files from trash to their original location, so that I can recover deleted files.

**Acceptance Criteria:**
- [ ] `R` key in trash dialog restores selected file to original path
- [ ] Conflict resolution dialog appears if file exists at destination
- [ ] `.trashinfo` file is removed after successful restore
- [ ] `R` key in normal file list performs rename (no conflict)

### US4: Empty Trash
As a user, I want to empty the trash to reclaim disk space, so that deleted files don't consume storage indefinitely.

**Acceptance Criteria:**
- [ ] `Shift+E` in trash dialog prompts for confirmation
- [ ] Confirmation dialog prevents accidental data loss
- [ ] All files and `.trashinfo` files are permanently deleted

### US5: View Trash Metadata
As a user, I want to see the original path and deletion date of trashed files, so that I can identify which file to restore.

**Acceptance Criteria:**
- [ ] Name column shows file/directory name
- [ ] Size column shows file size
- [ ] Deleted column shows deletion date/time
- [ ] Original Path column shows where the file came from

## Domain Rules

- **Trash Location**: `~/.local/share/Trash/` with `files/` and `info/` subdirectories
- **Trashinfo Format**: INI-style file with `[Trash Info]` section, `Path` and `DeletionDate` keys
- **Name Collision**: When moving to trash, if same name exists, append counter (file.txt -> file.2.txt)
- **Restore Collision**: User must choose: overwrite, rename, or skip
- **Dialog Isolation**: Trash dialog is a separate modal; `R` key has different functions inside/outside the dialog
- **No Keybinding Conflict**: `R` = restore (in trash dialog), `R` = rename (in normal file list)

## Technical Requirements

### Functional Requirements

#### Phase 1 (MVP)

- **FR1.1**: `Delete` key moves selected file(s) to `~/.local/share/Trash/files/`
- **FR1.2**: Generate `.trashinfo` file in `~/.local/share/Trash/info/` with original path and deletion timestamp
- **FR1.3**: Handle name collisions by appending counter (file.txt -> file.2.txt, file.3.txt, ...)
- **FR1.4**: `T` key opens trash dialog (screen-centered, both panes dimmed)
- **FR1.5**: Trash dialog displays: Name, Size, Deleted, Original Path columns
- **FR1.6**: Same-filesystem moves use `os.Rename` for efficiency
- **FR1.7**: Cross-filesystem moves use copy+delete (leverage existing TaskManager)
- **FR1.8**: j/k navigation within trash dialog

#### Phase 2 (Restore and Management)

- **FR2.1**: `R` key in trash dialog restores file to original path (read from `.trashinfo`)
- **FR2.2**: Display conflict resolution dialog when restore destination exists (overwrite/rename/skip)
- **FR2.3**: `Shift+E` in trash dialog empties trash with confirmation dialog
- **FR2.4**: Space key marks/unmarks files in trash dialog for batch operations

#### Phase 3 (Extended)

- **FR3.1**: Support `.Trash-$UID` directory on external devices

### Non-Functional Requirements

- **NFR1.1 - Performance**: Same-filesystem trash operation < 100ms
- **NFR1.2 - Performance**: Trash dialog opening with 1000 files < 100ms
- **NFR1.3 - Compatibility**: Full compliance with freedesktop.org Trash Specification
- **NFR1.4 - Security**: Validate paths to prevent directory traversal attacks

## Key Bindings

### Global (Normal File List)

| Key | Action |
|-----|--------|
| `Delete` | Move selected file(s) to trash |
| `d` | Direct delete (existing, unchanged) |
| `T` (Shift+t) | Open trash dialog |
| `R` | Rename file (unchanged) |

### Trash Dialog Only

| Key | Action |
|-----|--------|
| `j` / `Down` | Move cursor down |
| `k` / `Up` | Move cursor up |
| `Space` | Mark/unmark file |
| `R` | Restore selected/marked file(s) |
| `Shift+E` | Empty trash (with confirmation) |
| `Esc` / `q` | Close dialog |

## Trash Directory Structure

```
~/.local/share/Trash/
├── files/           # Actual file contents
│   ├── document.txt
│   └── photo.jpg
└── info/            # Metadata files
    ├── document.txt.trashinfo
    └── photo.jpg.trashinfo
```

## Trashinfo File Format

```ini
[Trash Info]
Path=/home/user/Documents/important.txt
DeletionDate=2026-01-25T10:30:00
```

| Field | Description | Format |
|-------|-------------|--------|
| Path | Original absolute path | URL-encoded for special characters |
| DeletionDate | Deletion timestamp | ISO 8601 (YYYY-MM-DDTHH:MM:SS), local time without timezone suffix |

## State Machine

```mermaid
stateDiagram-v2
    [*] --> Normal

    Normal --> TrashMove: Delete key
    TrashMove --> Normal: Success
    TrashMove --> Error: Failure

    Normal --> TrashDialog: T key
    TrashDialog --> Normal: Esc/q

    TrashDialog --> RestorePrompt: R key (file selected)
    RestorePrompt --> TrashDialog: Success/Cancel
    RestorePrompt --> ConflictDialog: Destination exists
    ConflictDialog --> TrashDialog: User choice applied

    TrashDialog --> EmptyConfirm: Shift+E
    EmptyConfirm --> TrashDialog: Confirmed (empty)
    EmptyConfirm --> TrashDialog: Cancelled
```

## Interface Contract

### Trash Operations

#### MoveToTrash

**Input:**
- `paths`: []string - Absolute paths of files to trash

**Output:**
- `error`: Error if operation fails

**Preconditions:**
- Files exist at specified paths
- User has read permission on source files
- User has write permission on trash directory

**Postconditions:**
- Files moved to `~/.local/share/Trash/files/`
- `.trashinfo` files created in `~/.local/share/Trash/info/`
- Original files removed from source location

#### RestoreFromTrash

**Input:**
- `trashPath`: string - Path within trash/files/

**Output:**
- `error`: Error if operation fails

**Preconditions:**
- File exists in trash
- Corresponding `.trashinfo` exists
- User has write permission on original directory

**Postconditions:**
- File restored to original location
- `.trashinfo` file removed

### Trashinfo Parser

**Input:**
- `trashinfoPath`: string - Path to `.trashinfo` file

**Output:**
- `originalPath`: string - URL-decoded original path
- `deletionDate`: time.Time - Parsed deletion timestamp
- `error`: Error if parsing fails

## Error Handling

| Error Condition | Behavior | User Message |
|-----------------|----------|--------------|
| Trash directory not writable | Show error, abort operation (no fallback to direct delete) | "Cannot access trash: permission denied" |
| Trash directory unavailable | Show error, abort operation (no fallback to direct delete) | "Trash is not available" |
| Source file not found | Skip file, continue with others | "File not found: {filename}" |
| Cross-filesystem move failure | Fallback to copy+delete | (transparent to user) |
| Invalid .trashinfo format | Show error on restore | "Cannot restore: invalid trash metadata" |
| Restore destination not writable | Show error | "Cannot restore: permission denied on destination" |

**Important:** When trash operations fail (permission denied, trash unavailable, etc.), the operation is aborted with an error message. The system does NOT silently fall back to direct deletion. Users who want to permanently delete files must explicitly use the `d` key (direct delete) command.

## UI Layout

### Trash Dialog (Screen Center)

The trash dialog is displayed at the **screen center** with **both panes dimmed** behind it. This is a full-screen dialog similar to the Help dialog.

```
┌─────────────────────────────────────────────────────────────────┐
│ Trash                                                     [5]   │
├─────────────────────────────────────────────────────────────────┤
│ Name                Size    Deleted              Original Path  │
├─────────────────────────────────────────────────────────────────┤
│ document.txt        2.1K    2026-01-25 10:30    ~/Documents/    │
│ photo.jpg           1.2M    2026-01-24 15:45    ~/Pictures/     │
│ project/               -    2026-01-23 09:00    ~/src/          │
│ backup.tar.gz      45.2M    2026-01-22 14:20    ~/Downloads/    │
│ notes.md            512B    2026-01-21 09:15    ~/Documents/    │
├─────────────────────────────────────────────────────────────────┤
│ R:Restore  Shift+E:Empty  Space:Mark  Esc:Close                 │
└─────────────────────────────────────────────────────────────────┘
```

**Dialog Elements:**
- **Title**: "Trash" with item count in brackets [N]
- **Columns**: Name, Size, Deleted (date/time), Original Path
- **Footer**: Keybinding hints
- **Display Type**: `DialogDisplayScreen` (both panes dimmed)

### Conflict Resolution Dialog

```
┌─────────────────────────────────────────┐
│ File already exists                     │
│                                         │
│ /home/user/Documents/document.txt       │
│                                         │
│ [O]verwrite  [R]ename  [S]kip           │
└─────────────────────────────────────────┘
```

### Empty Trash Confirmation

```
┌─────────────────────────────────────────┐
│ Empty Trash                             │
│                                         │
│ Permanently delete all files in trash?  │
│ This cannot be undone.                  │
│                                         │
│        [Y]es         [N]o               │
└─────────────────────────────────────────┘
```

## Implementation Phases

### Phase 1: MVP
**Goals:** Basic trash functionality with dialog
**Deliverables:**
- `Delete` key moves files to trash
- `.trashinfo` file generation
- Name collision handling
- `T` key opens trash dialog (screen center, both panes dimmed)
- Trash dialog with Name, Size, Deleted, Original Path columns
- j/k navigation in dialog (FR1.8)

### Phase 2: Restore and Management
**Goals:** Full trash management
**Deliverables:**
- `R` key restore with conflict handling
- `Shift+E` empty trash with confirmation
- Space key for marking files
- Batch restore/delete for marked files

### Phase 3: Extended
**Goals:** External device support
**Deliverables:**
- `.Trash-$UID` directory support on external devices

## Test Scenarios

### Unit Tests

#### Trashinfo Generation
- [ ] Generate valid .trashinfo with correct format
- [ ] URL-encode special characters in path
- [ ] ISO 8601 timestamp format

#### Trashinfo Parsing
- [ ] Parse valid .trashinfo file
- [ ] Handle URL-encoded paths
- [ ] Parse ISO 8601 timestamps
- [ ] Error on malformed file

#### Name Collision
- [ ] No collision: use original name
- [ ] First collision: append ".2"
- [ ] Multiple collisions: increment counter
- [ ] Handle files with extensions correctly

### Integration Tests

#### Move to Trash
- [ ] Single file moves correctly
- [ ] Directory moves recursively
- [ ] Multiple selected files move
- [ ] Cross-filesystem move works
- [ ] Permission denied handled gracefully

#### Trash Dialog
- [ ] Dialog opens at screen center
- [ ] Both panes are dimmed
- [ ] All columns displayed correctly
- [ ] j/k navigation works
- [ ] Space marks/unmarks files

#### Restore from Trash
- [ ] Restore to original location
- [ ] Handle missing destination directory
- [ ] Conflict dialog appears when needed
- [ ] Overwrite option works
- [ ] Rename option works
- [ ] Skip option works

#### Empty Trash
- [ ] Confirmation prevents accidental deletion
- [ ] All files and trashinfo removed
- [ ] Empty trash on empty directory is no-op

### E2E Tests

- [ ] Delete key moves file to trash
- [ ] T key opens trash dialog
- [ ] R key in dialog restores file
- [ ] R key outside dialog renames file (no conflict)
- [ ] Shift+E empties trash with confirmation
- [ ] Esc closes trash dialog
- [ ] All columns visible in trash dialog

### Edge Cases

- [ ] Unicode file names
- [ ] Very long path names
- [ ] Files with special characters (spaces, quotes)
- [ ] Symlinks (move link, not target)
- [ ] Empty trash when already empty
- [ ] Restore when original parent directory deleted
- [ ] Open trash dialog when trash is empty

## Security Considerations

- **Path Validation**: Validate all paths to prevent directory traversal
- **Permission Checks**: Verify write permissions before operations
- **Symlink Handling**: Move symlink itself, not follow to target
- **URL Encoding**: Properly encode/decode special characters in paths

## Dependencies

- Go standard library `os`, `path/filepath`, `net/url`
- Existing TaskManager for cross-filesystem operations
- Existing dialog infrastructure (`DialogDisplayScreen` type)

## Success Criteria

- [ ] All Phase 1 functional requirements implemented
- [ ] All Phase 2 functional requirements implemented
- [ ] freedesktop.org Trash Specification compliance
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] All E2E tests pass
- [ ] Performance targets met
- [ ] Code review completed

## Open Questions

None - all requirements have been clarified.

## Out of Scope

- Automatic trash cleanup (time-based expiration)
- Trash size limits
- Network drive support
- Undo functionality (quick restore of last deletion)
