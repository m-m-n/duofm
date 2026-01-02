#!/bin/bash
# E2E Test Runner for duofm
#
# Description: Main test runner that executes all E2E tests or selected categories
#
# Usage:
#   ./run_all_tests.sh              # Run all tests
#   ./run_all_tests.sh basic        # Run basic tests only
#   ./run_all_tests.sh file-ops     # Run file operation tests only
#   ./run_all_tests.sh --list       # List available test categories
#
# Available categories:
#   basic, directory, file-ops, copy-move, cursor, sort,
#   shell, config, bookmark, mark, history

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ ! -f "${SCRIPT_DIR}/helpers.sh" ]; then
    echo "Error: helpers.sh not found at ${SCRIPT_DIR}/helpers.sh" >&2
    exit 1
fi

source "${SCRIPT_DIR}/helpers.sh"

# Test category mapping
declare -A TEST_FILES=(
    ["basic"]="basic_tests.sh"
    ["directory"]="directory_tests.sh"
    ["file-ops"]="file_operation_tests.sh"
    ["copy-move"]="copy_move_tests.sh"
    ["cursor"]="cursor_tests.sh"
    ["sort"]="sort_tests.sh"
    ["shell"]="shell_tests.sh"
    ["config"]="config_tests.sh"
    ["bookmark"]="bookmark_tests.sh"
    ["mark"]="mark_tests.sh"
    ["history"]="history_tests.sh"
    ["archive"]="archive_tests.sh"
)

# Show usage
show_usage() {
    echo "Usage: $0 [category|--list|--help]"
    echo ""
    echo "Run all E2E tests or a specific category."
    echo ""
    echo "Options:"
    echo "  --list, -l    List available test categories"
    echo "  --help, -h    Show this help message"
    echo ""
    echo "Categories:"
    for category in "${!TEST_FILES[@]}"; do
        echo "  $category"
    done | sort
}

# List available categories
list_categories() {
    echo "Available test categories:"
    echo ""
    for category in "${!TEST_FILES[@]}"; do
        local file="${TEST_FILES[$category]}"
        local count
        count=$(grep -c "^test_" "${SCRIPT_DIR}/tests/${file}" 2>/dev/null || echo "?")
        printf "  %-12s %s (%s tests)\n" "$category" "$file" "$count"
    done | sort
}

# Run tests for a specific category
run_category() {
    local category="$1"
    local file="${TEST_FILES[$category]}"

    if [ -z "$file" ]; then
        echo "Error: Unknown category '$category'"
        echo ""
        list_categories
        exit 1
    fi

    local test_file="${SCRIPT_DIR}/tests/${file}"
    if [ ! -f "$test_file" ]; then
        echo "Error: Test file not found: $test_file"
        exit 1
    fi

    echo "========================================"
    echo "Running: $category tests"
    echo "========================================"

    source "$test_file"

    # Get all test functions from the file
    local tests
    tests=$(grep -o "^test_[a-zA-Z0-9_]*" "$test_file" | sort -u)

    for test_name in $tests; do
        run_test "$test_name"
    done
}

# Run all tests
run_all() {
    echo "========================================"
    echo "duofm E2E Tests - All"
    echo "========================================"
    echo "Working directory: $(pwd)"
    echo ""

    # Source all test files
    for file in "${TEST_FILES[@]}"; do
        source "${SCRIPT_DIR}/tests/${file}"
    done

    # Basic tests
    echo ""
    echo "=== Basic Tests ==="
    run_test test_basic_startup
    run_test test_jk_navigation
    run_test test_enter_directory
    run_test test_parent_directory
    run_test test_pane_switching
    run_test test_help_dialog
    run_test test_search_filter
    run_test test_quit
    run_test test_ctrlc_quit

    # Directory tests
    echo ""
    echo "=== Directory Tests ==="
    run_test test_symlink_display
    run_test test_permission_denied_directory
    run_test test_f5_refresh
    run_test test_ctrlr_refresh
    run_test test_refresh_cursor_preservation
    run_test test_sync_pane
    run_test test_sync_preserves_settings
    run_test test_sync_right_to_left
    run_test test_right_pane_same_path_navigation
    run_test test_right_pane_home_navigation

    # File operation tests
    echo ""
    echo "=== File Operation Tests ==="
    run_test test_cannot_delete_root_file
    run_test test_can_delete_user_file
    run_test test_create_new_file
    run_test test_create_new_directory
    run_test test_rename_file
    run_test test_cancel_file_creation
    run_test test_empty_filename_error
    run_test test_navigation_after_file_creation
    run_test test_navigation_after_dir_creation
    run_test test_navigation_after_rename
    run_test test_rename_parent_dir_ignored

    # Copy/Move tests
    echo ""
    echo "=== Copy/Move Tests ==="
    run_test test_copy_overwrite_cancel
    run_test test_copy_overwrite_confirm
    run_test test_copy_overwrite_rename
    run_test test_move_overwrite
    run_test test_directory_conflict_error
    run_test test_overwrite_dialog_navigation
    run_test test_rename_dialog_validation
    run_test test_copy_no_conflict

    # Mark tests
    echo ""
    echo "=== Mark Tests ==="
    run_test test_mark_file
    run_test test_mark_cursor_movement
    run_test test_unmark_file
    run_test test_mark_parent_dir_ignored
    run_test test_mark_multiple_files
    run_test test_marks_cleared_on_directory_change
    run_test test_batch_delete_marked_files
    run_test test_context_menu_mark_count

    # Sort tests
    echo ""
    echo "=== Sort Tests ==="
    run_test test_sort_dialog_opens
    run_test test_sort_dialog_hl_navigation
    run_test test_sort_dialog_jk_navigation
    run_test test_sort_dialog_confirm
    run_test test_sort_dialog_cancel
    run_test test_sort_dialog_q_cancel
    run_test test_sort_by_size_desc
    run_test test_sort_persists_after_navigation
    run_test test_sort_independent_panes
    run_test test_sort_dialog_arrow_keys

    # Cursor tests
    echo ""
    echo "=== Cursor Tests ==="
    run_test test_cursor_preserved_after_view
    run_test test_cursor_preserved_after_enter_view
    run_test test_cursor_reset_when_file_deleted
    run_test test_both_panes_preserve_cursor
    run_test test_parent_nav_cursor_on_subdir_h_key
    run_test test_parent_nav_cursor_on_subdir_dotdot
    run_test test_parent_nav_cursor_on_subdir_l_key
    run_test test_parent_nav_independent_pane_memory

    # Shell tests
    echo ""
    echo "=== Shell Tests ==="
    run_test test_shell_command_mode_enter
    run_test test_shell_command_input
    run_test test_shell_command_empty_enter
    run_test test_shell_command_ignored_with_dialog
    run_test test_shell_command_ignored_during_search
    run_test test_help_shows_shell_command

    # Config tests
    echo ""
    echo "=== Config Tests ==="
    run_test test_config_auto_generated
    run_test test_config_has_keybindings
    run_test test_config_has_comments
    run_test test_default_keybindings_without_config
    run_test test_help_shows_pascalcase

    # Bookmark tests
    echo ""
    echo "=== Bookmark Tests ==="
    run_test test_bookmark_dialog_opens
    run_test test_add_bookmark_dialog
    run_test test_bookmark_empty_state

    # History tests
    echo ""
    echo "=== History Navigation Tests ==="
    run_test test_history_back
    run_test test_history_forward
    run_test test_history_multiple_levels
    run_test test_history_no_history
    run_test test_history_independent_from_previous
    run_test test_history_forward_cleared
    run_test test_history_parent_navigation
    run_test test_history_home_navigation

    # Archive tests
    echo ""
    echo "=== Archive Tests ==="
    run_test test_compress_format_dialog_opens
    run_test test_compress_format_navigation
    run_test test_compression_level_dialog
    run_test test_archive_name_dialog
    run_test test_archive_conflict_dialog
    run_test test_compress_cancel_workflow
    run_test test_compress_complete_workflow
    run_test test_extract_complete_workflow
    run_test test_multifile_compress
}

# Main entry point
main() {
    case "${1:-}" in
        --help|-h)
            show_usage
            exit 0
            ;;
        --list|-l)
            list_categories
            exit 0
            ;;
        "")
            run_all
            ;;
        *)
            run_category "$1"
            ;;
    esac

    print_summary
    exit $?
}

main "$@"
