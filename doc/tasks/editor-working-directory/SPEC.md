# Feature: Editor/Viewer Working Directory

## Overview

Set the working directory to the active pane's directory when opening files with external applications (editor/viewer) in duofm. Additionally, respect the $EDITOR and $PAGER environment variables for application selection.

## Domain Rules

- The working directory of external applications MUST be the directory containing the target file
- Environment variables ($EDITOR, $PAGER) MUST take precedence over default applications
- If environment variables are not set, fall back to vim (editor) and less (viewer)

## Objectives

- Set the working directory to the active pane's directory when launching external applications
- Use $EDITOR environment variable for the `e` key (edit) operation
- Use $PAGER environment variable for the `v` key (view) and `Enter` key operations
- Maintain consistency with the existing `!` key (shell command) behavior

## User Stories

- As a user, I want the editor to open with the working directory set to the file's directory, so that commands like `:!ls` in vim show files in the same directory as the file I'm editing
- As a user, I want my $EDITOR preference to be used when pressing `e`, so that my preferred editor opens instead of vim
- As a user, I want my $PAGER preference to be used when pressing `v`, so that my preferred pager opens instead of less

## Functional Requirements

- FR1.1: When `e` key is pressed on a file, set the external application's working directory to the active pane's directory
- FR1.2: When `v` key is pressed on a file, set the external application's working directory to the active pane's directory
- FR1.3: When `Enter` key is pressed on a file, set the external application's working directory to the active pane's directory
- FR2.1: When `e` key is pressed, use $EDITOR if set; otherwise use "vim"
- FR2.2: When `v` key is pressed, use $PAGER if set; otherwise use "less"
- FR2.3: When `Enter` key is pressed on a file, use $PAGER if set; otherwise use "less"

## Non-Functional Requirements

- NFR1.1: Existing file path handling and error handling behavior must remain unchanged
- NFR1.2: Performance impact must be negligible (only adds environment variable lookup)

## Interface Contract

### Input/Output Specification

| Key | Target | Application | Working Directory |
|-----|--------|-------------|-------------------|
| `v` | File | $PAGER or "less" | Active pane directory |
| `e` | File | $EDITOR or "vim" | Active pane directory |
| `Enter` | File | $PAGER or "less" | Active pane directory |

### Preconditions

- Target is a file (not a directory)
- Target is not the parent directory entry (`..`)
- File must have read permission

### Postconditions

- External application is launched with:
  - Argument: absolute file path
  - Working directory: directory containing the file (active pane's path)
- Both panes are reloaded after the external application exits

### Error Conditions

| Error | Behavior |
|-------|----------|
| File not readable | Display "Cannot read file: [error]" in status bar |
| External application not found | Shell displays "command not found" error |
| Invalid $EDITOR/$PAGER command | Shell displays appropriate error |

## Dependencies

- Existing `tea.ExecProcess` for TUI suspension
- Go standard library `os` package for environment variable access
- Go standard library `os/exec` for command execution

## Test Scenarios

### Working Directory

- [ ] `v` on file: viewer runs with working directory set to file's directory
- [ ] `e` on file: editor runs with working directory set to file's directory
- [ ] `Enter` on file: viewer runs with working directory set to file's directory
- [ ] Shell command in vim (`:!pwd`) shows file's directory
- [ ] netrw in vim (`:e .`) shows file's directory contents

### Environment Variables

- [ ] $EDITOR set to "nano": `e` key launches nano
- [ ] $PAGER set to "moar": `v` key launches moar
- [ ] $EDITOR not set: `e` key launches vim
- [ ] $PAGER not set: `v` key launches less
- [ ] $EDITOR set to empty string: `e` key launches vim
- [ ] $PAGER set to empty string: `v` key launches less

### Existing Behavior Preserved

- [ ] `v` on directory: no action
- [ ] `e` on directory: no action
- [ ] `v` on `..`: no action
- [ ] `e` on `..`: no action
- [ ] File without read permission: status bar error displayed
- [ ] Screen restoration after external app exits
- [ ] Both panes reload after external app exits

## Success Criteria

- [ ] Working directory is correctly set for all external application launches
- [ ] $EDITOR environment variable is respected for edit operations
- [ ] $PAGER environment variable is respected for view operations
- [ ] Fallback to vim/less works when environment variables are not set
- [ ] All existing tests pass
- [ ] New tests for working directory and environment variable handling pass

## Constraints

- Must maintain consistency with `!` key shell command execution behavior
- External application launch mechanism (`tea.ExecProcess`) must not change
- File path must still be passed as an absolute path

## Open Questions

None - all requirements have been clarified.
