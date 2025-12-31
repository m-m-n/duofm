# Feature: Refactor E2E Test Scripts

## Overview

The current E2E test suite for duofm is contained in a single monolithic file (`test/e2e/scripts/run_tests.sh`) that has grown to 2,867 lines and 83KB. This file exceeds Claude's token limit (25,000 tokens at 27,170 tokens), making AI-assisted development and maintenance difficult.

This refactoring splits the monolithic test file into logical, category-based test files while maintaining full test coverage and functionality.

## Objectives

1. **Primary**: Reduce file size to enable Claude to read and correctly analyze test files
2. **Secondary**: Improve test organization and maintainability through logical grouping
3. **Tertiary**: Enable selective test execution by category

## Domain Rules

### Test Independence
- Each test function must be executable independently
- Tests must not have execution order dependencies
- Each test runs in its own isolated tmux session

### Test Preservation
- All 78 existing test functions must be preserved
- Test behavior and assertions must remain unchanged
- No test deletion or merging allowed

### No Backward Compatibility Required
- No CI/CD usage exists for the current `run_tests.sh`
- The existing file can be safely replaced or removed

## Functional Requirements

### FR1: Split Tests by Category

**FR1.1**: Create category-based test files under `test/e2e/scripts/tests/` directory

| File | Category | Example Tests | Count |
|------|----------|---------------|-------|
| `basic_tests.sh` | Basic functionality | startup, quit, j/k navigation | 9 |
| `directory_tests.sh` | Directory operations | enter dir, parent dir, pane sync | 10 |
| `file_operation_tests.sh` | File operations | create, delete, rename | 11 |
| `copy_move_tests.sh` | Copy/Move operations | copy, move, overwrite handling | 8 |
| `cursor_tests.sh` | Cursor position memory | cursor preservation after navigation | 8 |
| `sort_tests.sh` | Sort functionality | sort dialog, sort behavior | 10 |
| `shell_tests.sh` | Shell command mode | shell command execution | 6 |
| `config_tests.sh` | Configuration | config files, keybindings | 5 |
| `bookmark_tests.sh` | Bookmarks | bookmark add, display | 3 |
| `mark_tests.sh` | File marking | mark files, batch operations | 8 |

**FR1.2**: Each test file must be independently executable

- Each file sources `helpers.sh`
- When executed directly, runs only its own tests
- No dependencies on other test files

**FR1.3**: File size prioritizes logical coherence over strict line limits

- No hard line count restriction
- Related tests grouped in same file
- Target: 200-500 lines per file (comfortably within Claude's limits)

### FR2: Main Test Runner

**FR2.1**: Create `run_all_tests.sh` to execute all tests

Features:
- Sources and runs all test files
- Displays consolidated test summary
- Returns non-zero exit code if any tests fail

**FR2.2**: Support selective test execution by category

```bash
./run_all_tests.sh              # Run all tests
./run_all_tests.sh basic        # Run basic_tests.sh only
./run_all_tests.sh file-ops     # Run file_operation_tests.sh only
./run_all_tests.sh copy-move    # Run copy_move_tests.sh only
```

### FR3: Directory Structure

```
test/e2e/scripts/
├── helpers.sh                    # Existing (unchanged)
├── run_all_tests.sh             # New main runner
├── tests/                       # New test directory
│   ├── basic_tests.sh
│   ├── directory_tests.sh
│   ├── file_operation_tests.sh
│   ├── copy_move_tests.sh
│   ├── cursor_tests.sh
│   ├── sort_tests.sh
│   ├── shell_tests.sh
│   ├── config_tests.sh
│   ├── bookmark_tests.sh
│   └── mark_tests.sh
├── test_open_file.sh            # Existing (unchanged)
└── interactive.sh               # Existing (unchanged)
```

### FR4: Test File Structure

**FR4.1**: Each test file follows this template:

```bash
#!/bin/bash
# [Category] Tests for duofm
#
# Description: [Overview of tests in this file]
# Tests: [Number of tests]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# ===========================================
# Test: [Test Name]
# ===========================================
test_function_name() {
    # Test implementation
}

# Additional test functions...

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    run_test test_function_name
    # Additional test executions...
    print_summary
    exit $?
fi
```

**FR4.2**: Migrate test functions without modification

- Test function code remains unchanged
- Test function names remain unchanged
- Test behavior is preserved

## Non-Functional Requirements

### NFR1: Performance

**NFR1.1**: Total test execution time must be equal to or better than current implementation

- File splitting overhead minimized
- Selective execution enables faster development cycles

### NFR2: Maintainability

**NFR2.1**: All test files must fit within Claude's token limit (25,000 tokens)

- Target: 5,000-10,000 tokens per file
- Maximum: 20,000 tokens per file

**NFR2.2**: Tests must be easy to add and modify

- New tests added to appropriate category file
- Related tests co-located for easy understanding

### NFR3: Compatibility

**NFR3.1**: No changes required to `helpers.sh`

- All existing helper functions remain available
- New helpers added only as needed

**NFR3.2**: Preserve tmux-based test execution methodology

- Existing test approach unchanged
- Functions like `start_duofm`, `send_keys`, `assert_contains` continue to work

## Test Migration Mapping

### basic_tests.sh (9 tests)
- test_basic_startup
- test_jk_navigation
- test_enter_directory
- test_parent_directory
- test_pane_switching
- test_help_dialog
- test_search_filter
- test_quit
- test_ctrlc_quit

### directory_tests.sh (10 tests)
- test_symlink_display
- test_permission_denied_directory
- test_f5_refresh
- test_ctrlr_refresh
- test_refresh_cursor_preservation
- test_sync_pane
- test_sync_preserves_settings
- test_sync_right_to_left
- test_right_pane_same_path_navigation
- test_right_pane_home_navigation

### file_operation_tests.sh (11 tests)
- test_cannot_delete_root_file
- test_can_delete_user_file
- test_create_new_file
- test_create_new_directory
- test_rename_file
- test_cancel_file_creation
- test_empty_filename_error
- test_navigation_after_file_creation
- test_navigation_after_dir_creation
- test_navigation_after_rename
- test_rename_parent_dir_ignored

### copy_move_tests.sh (8 tests)
- test_copy_overwrite_cancel
- test_copy_overwrite_confirm
- test_copy_overwrite_rename
- test_move_overwrite
- test_directory_conflict_error
- test_overwrite_dialog_navigation
- test_rename_dialog_validation
- test_copy_no_conflict

### cursor_tests.sh (8 tests)
- test_cursor_preserved_after_view
- test_cursor_preserved_after_enter_view
- test_cursor_reset_when_file_deleted
- test_both_panes_preserve_cursor
- test_parent_nav_cursor_on_subdir_h_key
- test_parent_nav_cursor_on_subdir_dotdot
- test_parent_nav_cursor_on_subdir_l_key
- test_parent_nav_independent_pane_memory

### sort_tests.sh (10 tests)
- test_sort_dialog_opens
- test_sort_dialog_hl_navigation
- test_sort_dialog_jk_navigation
- test_sort_dialog_confirm
- test_sort_dialog_cancel
- test_sort_dialog_q_cancel
- test_sort_by_size_desc
- test_sort_persists_after_navigation
- test_sort_independent_panes
- test_sort_dialog_arrow_keys

### shell_tests.sh (6 tests)
- test_shell_command_mode_enter
- test_shell_command_input
- test_shell_command_empty_enter
- test_shell_command_ignored_with_dialog
- test_shell_command_ignored_during_search
- test_help_shows_shell_command

### config_tests.sh (5 tests)
- test_config_auto_generated
- test_config_has_keybindings
- test_config_has_comments
- test_default_keybindings_without_config
- test_help_shows_pascalcase

### bookmark_tests.sh (3 tests)
- test_bookmark_dialog_opens
- test_add_bookmark_dialog
- test_bookmark_empty_state

### mark_tests.sh (8 tests)
- test_mark_file
- test_mark_cursor_movement
- test_unmark_file
- test_mark_parent_dir_ignored
- test_mark_multiple_files
- test_marks_cleared_on_directory_change
- test_batch_delete_marked_files
- test_context_menu_mark_count

## Implementation Approach

### Phase 1: Setup (30 minutes)

1. Create `test/e2e/scripts/tests/` directory
2. Create template for test files with proper header structure
3. Verify `helpers.sh` can be sourced from new location

### Phase 2: Test Migration (3-4 hours)

For each category:

1. Create new test file with header
2. Copy test functions from `run_tests.sh`
3. Add test execution block for standalone mode
4. Verify file is under 20,000 tokens
5. Test standalone execution

**Migration order** (by dependency/simplicity):
1. basic_tests.sh (simplest, fewest dependencies)
2. directory_tests.sh
3. file_operation_tests.sh
4. cursor_tests.sh
5. copy_move_tests.sh
6. sort_tests.sh
7. shell_tests.sh
8. config_tests.sh
9. mark_tests.sh
10. bookmark_tests.sh

### Phase 3: Main Runner (1 hour)

1. Create `run_all_tests.sh`
2. Implement argument parsing for selective execution
3. Source all test files
4. Execute tests based on arguments
5. Collect and display summary

**Argument mapping**:
```bash
basic       -> basic_tests.sh
directory   -> directory_tests.sh
file-ops    -> file_operation_tests.sh
copy-move   -> copy_move_tests.sh
cursor      -> cursor_tests.sh
sort        -> sort_tests.sh
shell       -> shell_tests.sh
config      -> config_tests.sh
bookmark    -> bookmark_tests.sh
mark        -> mark_tests.sh
```

### Phase 4: Verification (1 hour)

1. Run all tests via `run_all_tests.sh`
2. Verify all 78 tests execute
3. Compare results with original `run_tests.sh`
4. Test selective execution for each category
5. Test standalone execution for each file
6. Check token counts for all files

### Phase 5: Cleanup (30 minutes)

1. Archive or remove original `run_tests.sh`
2. Update any documentation referencing old structure
3. Final validation

## Test Scenarios

### Standalone Execution
- [ ] Each test file executes successfully when run directly
- [ ] Each test file shows correct pass/fail count
- [ ] Each test file sources helpers.sh correctly from new location

### Main Runner - Full Execution
- [ ] `./run_all_tests.sh` runs all 78 tests
- [ ] Summary shows correct total count (78)
- [ ] Exit code is 0 if all pass, non-zero if any fail
- [ ] Test output matches original run_tests.sh behavior

### Main Runner - Selective Execution
- [ ] `./run_all_tests.sh basic` runs only basic tests
- [ ] `./run_all_tests.sh file-ops` runs only file operation tests
- [ ] Invalid category shows helpful error message
- [ ] Summary shows only executed tests

### Token Limits
- [ ] All test files are under 25,000 tokens
- [ ] Largest file is under 20,000 tokens
- [ ] Average file size is 5,000-10,000 tokens
- [ ] Claude can successfully read and analyze each file

### Test Preservation
- [ ] All 78 original tests present in new structure
- [ ] Test function names unchanged
- [ ] Test behavior unchanged (same assertions)
- [ ] No tests accidentally duplicated or lost

### Helper Functions
- [ ] All existing helper functions work from new test files
- [ ] `start_duofm`, `send_keys`, `assert_contains` function correctly
- [ ] Test counters work correctly across all files
- [ ] `print_summary` displays accurate results

## Success Criteria

### File Size
- [ ] All split files fit within Claude's 25,000 token limit
- [ ] Largest file is under 20,000 tokens

### Test Completeness
- [ ] All 78 test functions migrated to new structure
- [ ] All tests execute and pass
- [ ] Test behavior unchanged from original

### Execution
- [ ] `./run_all_tests.sh` executes all tests
- [ ] `./run_all_tests.sh [category]` executes category tests
- [ ] `./tests/[file].sh` executes individual file tests
- [ ] Test summary displays correctly

### Maintainability
- [ ] Each test file has descriptive header comment
- [ ] Tests easy to locate by category
- [ ] Claude can read and understand each file

## Constraints

### Technical Constraints
- Bash scripting language
- tmux-based E2E test methodology
- Dependency on existing helpers.sh

### Resource Constraints
- Use existing test data (/testdata)
- No new external tool dependencies

### Operational Constraints
- No CI/CD backward compatibility required
- Must work in development environment

## Dependencies

- Existing `helpers.sh` (no modifications)
- tmux for test execution
- duofm binary for testing
- Test data directory (/testdata)

## Open Questions

None (all clarified through user feedback)
