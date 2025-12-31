# Feature: Directory Bookmark

## Overview

Add directory bookmark functionality to duofm. Users can register frequently accessed directories as bookmarks and quickly jump to them. Bookmarks are persisted in the configuration file.

## Domain Rules

- Bookmarks are stored in the `[bookmarks]` section of the configuration file
- Each bookmark has a `name` (alias) and `path` (directory path)
- New bookmarks are added to the top of the list
- Duplicate paths are not allowed (same path cannot be bookmarked twice)
- Bookmarks to non-existent paths are displayed with a warning indicator and cannot be jumped to

## Objectives

- Enable quick navigation to frequently used directories
- Allow users to assign memorable aliases to directories
- Persist bookmarks across sessions via configuration file

## User Stories

- As a user, I want to bookmark directories I frequently access
- As a user, I want to quickly jump to bookmarked directories from a list
- As a user, I want to assign meaningful names (aliases) to my bookmarks
- As a user, I want to remove bookmarks I no longer need

## Functional Requirements

### FR-1: Add Bookmark (Shift+B)

- FR-1.1: `Shift+B` key adds the current active pane's directory to bookmarks
- FR-1.2: Display an alias input dialog
- FR-1.3: Default alias is the directory's basename (e.g., `/path/to/projects` → `projects`)
- FR-1.4: `Enter` confirms, `Esc` cancels
- FR-1.5: If the same path is already bookmarked, display "Already bookmarked" and do not add
- FR-1.6: New bookmarks are added to the top of the list

### FR-2: Bookmark Manager Dialog (B)

- FR-2.1: `b` key opens the bookmark manager dialog
- FR-2.2: Bookmarks are displayed in two-line format (line 1: alias, line 2: path)
- FR-2.3: Long paths wrap to display completely
- FR-2.4: `j`/`k` keys move cursor up/down
- FR-2.5: `Enter` jumps to the selected bookmark's directory
- FR-2.6: `d` key deletes the selected bookmark
- FR-2.7: `e` key edits the selected bookmark's alias
- FR-2.8: `Esc` key closes the dialog

### FR-3: Jump to Bookmark

- FR-3.1: After selection, the active pane navigates to the target directory
- FR-3.2: Jumping to non-existent paths is disabled
- FR-3.3: The bookmark manager dialog closes automatically after jumping

### FR-4: Non-existent Path Display

- FR-4.1: Bookmarks with non-existent paths display a warning indicator (emoji)
- FR-4.2: Bookmarks with warning indicators cannot be jumped to
- FR-4.3: Deletion and editing are still allowed

### FR-5: Configuration File Storage

- FR-5.1: Bookmarks are stored in the `[bookmarks]` section
- FR-5.2: Use TOML array format (`[[bookmarks]]`)
- FR-5.3: Each bookmark has `name` (alias) and `path` fields
- FR-5.4: Configuration file is updated when bookmarks are added, deleted, or edited

### FR-6: Bookmark Editing

- FR-6.1: Edit dialog allows changing the alias name
- FR-6.2: Current alias is pre-filled in the input field
- FR-6.3: `Enter` confirms, `Esc` cancels

### FR-7: Bookmark Deletion

- FR-7.1: `d` key deletes the selected bookmark
- FR-7.2: Delete immediately without confirmation dialog (lightweight operation)

### FR-8: Help Dialog

- FR-8.1: Help dialog includes bookmark keybindings in "Bookmarks" section
- FR-8.2: Display `b` key for opening bookmark manager
- FR-8.3: Display `Shift+B` key for adding current directory to bookmarks

## Non-Functional Requirements

- NFR-1.1: Bookmark manager dialog displays within 100ms
- NFR-1.2: Jump to bookmark completes within 200ms
- NFR-1.3: Configuration file save completes within 100ms
- NFR-2.1: No limit on number of bookmarks (expect practical use of dozens)
- NFR-2.2: No limit on alias name length

## Interface Contract

### Input/Output Specification

#### Configuration File Format

```toml
# Bookmarks (newest first)
[[bookmarks]]
name = "Projects"
path = "/path/to/projects"

[[bookmarks]]
name = "Downloads"
path = "/path/to/Downloads"

[[bookmarks]]
name = "Config"
path = "/path/to/.config"
```

#### Bookmark Data Structure

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Alias/display name for the bookmark |
| `path` | string | Absolute path to the directory |

### Preconditions/Postconditions

#### Add Bookmark

- **Precondition**: Active pane displays a valid directory
- **Postcondition**: New bookmark added to configuration file (if not duplicate)

#### Jump to Bookmark

- **Precondition**: Selected bookmark path exists on filesystem
- **Postcondition**: Active pane displays the bookmarked directory

### State Transitions

```
[Normal Mode]
    |
    |-- b --> [Bookmark Manager Dialog]
    |              |
    |              |-- j/k --> [Navigate list]
    |              |-- Enter --> [Jump & Close]
    |              |-- d --> [Delete bookmark]
    |              |-- e --> [Edit Dialog]
    |              |-- Esc --> [Close]
    |
    |-- Shift+B --> [Add Bookmark Dialog]
                        |
                        |-- Enter --> [Save & Close]
                        |-- Esc --> [Cancel & Close]
```

### Error Conditions

| Condition | Behavior |
|-----------|----------|
| Configuration file does not exist | Start with empty bookmark list |
| No bookmarks section in config | Start with empty bookmark list |
| Invalid bookmark entry | Warning displayed, entry skipped |
| Config file write error | Error message in status bar |
| Path does not exist | Warning indicator shown, jump disabled |
| Duplicate path registration | "Already bookmarked" message, no action |

## Dependencies

- `internal/config` package (configuration file read/write)
- Existing dialog components (text input, list selection)
- Existing keybinding mechanism

## Test Scenarios

### Unit Tests

- [ ] AddBookmark correctly adds a new bookmark to the list
- [ ] AddBookmark returns error for duplicate path
- [ ] RemoveBookmark correctly removes bookmark from list
- [ ] EditBookmark correctly updates alias name
- [ ] IsPathBookmarked correctly detects existing paths
- [ ] LoadBookmarks correctly parses TOML configuration
- [ ] SaveBookmarks correctly writes TOML configuration
- [ ] Default alias is correctly derived from path basename

### Integration Tests

- [ ] Bookmark added via Shift+B appears in bookmark list
- [ ] Bookmark deleted via D is removed from configuration file
- [ ] Edited alias is persisted to configuration file
- [ ] Non-existent paths are marked with warning indicator

### E2E Tests

- [ ] `Shift+B` displays add bookmark dialog
- [ ] Default alias is directory basename
- [ ] `b` displays bookmark manager dialog
- [ ] `j`/`k` navigates bookmark selection
- [ ] `Enter` jumps to selected bookmark
- [ ] `d` deletes selected bookmark
- [ ] `e` displays edit dialog
- [ ] Warning indicator shown for non-existent paths
- [ ] Cannot jump to non-existent path
- [ ] Bookmarks are saved to configuration file

## Success Criteria

- [ ] `Shift+B` adds current directory to bookmarks
- [ ] `b` opens bookmark manager dialog
- [ ] Can jump to bookmarked directories from list
- [ ] Can edit and delete bookmarks
- [ ] Warning shown for non-existent paths
- [ ] Bookmarks are persisted in configuration file
- [ ] All unit tests pass
- [ ] All E2E tests pass

## Constraints

- Bookmark reordering is done by manually editing the configuration file
- Path editing is not supported (only alias can be edited)
- Bookmark import/export functionality is not supported
