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
#   ./run_all_tests.sh --check-list # Check the run list without running tests
#
# Available categories:
#   basic, directory, file-ops, copy-move, cursor, cursor-preserve,
#   sort, shell, config, bookmark, mark, history, archive, mouse
#
# --check-list validates the full-run list against the tests actually
# defined under tests/*.sh: it flags run-list entries with no matching
# test_ function (undefined entries) and defined test_ functions absent
# from the run list (missing entries). It never starts tmux or duofm.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# Save runner's SCRIPT_DIR since sourced test files may overwrite it
RUNNER_DIR="${SCRIPT_DIR}"

if [ ! -f "${RUNNER_DIR}/helpers.sh" ]; then
    echo "Error: helpers.sh not found at ${RUNNER_DIR}/helpers.sh" >&2
    exit 1
fi

source "${RUNNER_DIR}/helpers.sh"

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
    ["cursor-preserve"]="cursor_preserve_tests.sh"
    ["background"]="background_tests.sh"
    ["mouse"]="mouse_tests.sh"
)

# Full-run list: the single source of truth for the "no argument" run and
# for the run-list guard (--check-list and the full run's guard failures
# read this same array, so they cannot drift apart).
# Each entry is "<group heading>|<test_ function name>", grouped and
# ordered by group as in the full run's group headings, and within each
# group ordered as the functions appear in their tests/*.sh file.
FULL_RUN_LIST=(
    "Basic Tests|test_basic_startup"
    "Basic Tests|test_jk_navigation"
    "Basic Tests|test_enter_directory"
    "Basic Tests|test_parent_directory"
    "Basic Tests|test_pane_switching"
    "Basic Tests|test_help_dialog"
    "Basic Tests|test_search_filter"
    "Basic Tests|test_quit"
    "Basic Tests|test_ctrlc_quit"

    "Directory Tests|test_symlink_display"
    "Directory Tests|test_permission_denied_directory"
    "Directory Tests|test_f5_refresh"
    "Directory Tests|test_ctrlr_refresh"
    "Directory Tests|test_refresh_cursor_preservation"
    "Directory Tests|test_sync_pane"
    "Directory Tests|test_sync_preserves_settings"
    "Directory Tests|test_sync_right_to_left"
    "Directory Tests|test_right_pane_same_path_navigation"
    "Directory Tests|test_right_pane_home_navigation"

    "File Operation Tests|test_cannot_delete_root_file"
    "File Operation Tests|test_can_delete_user_file"
    "File Operation Tests|test_create_new_file"
    "File Operation Tests|test_create_new_directory"
    "File Operation Tests|test_rename_file"
    "File Operation Tests|test_cancel_file_creation"
    "File Operation Tests|test_empty_filename_error"
    "File Operation Tests|test_navigation_after_file_creation"
    "File Operation Tests|test_navigation_after_dir_creation"
    "File Operation Tests|test_navigation_after_rename"
    "File Operation Tests|test_rename_parent_dir_ignored"
    "File Operation Tests|test_delete_confirmation_enter_ignored"
    "File Operation Tests|test_delete_confirmation_y_key_works"

    "Copy/Move Tests|test_copy_overwrite_cancel"
    "Copy/Move Tests|test_copy_overwrite_confirm"
    "Copy/Move Tests|test_copy_overwrite_rename"
    "Copy/Move Tests|test_move_overwrite"
    "Copy/Move Tests|test_directory_conflict_error"
    "Copy/Move Tests|test_overwrite_dialog_navigation"
    "Copy/Move Tests|test_rename_dialog_validation"
    "Copy/Move Tests|test_copy_no_conflict"

    "Mark Tests|test_mark_file"
    "Mark Tests|test_mark_cursor_movement"
    "Mark Tests|test_unmark_file"
    "Mark Tests|test_mark_parent_dir_ignored"
    "Mark Tests|test_mark_multiple_files"
    "Mark Tests|test_marks_cleared_on_directory_change"
    "Mark Tests|test_batch_delete_marked_files"
    "Mark Tests|test_context_menu_mark_count"

    "Sort Tests|test_sort_dialog_opens"
    "Sort Tests|test_sort_dialog_tab_navigation"
    "Sort Tests|test_sort_dialog_dropdown_expansion"
    "Sort Tests|test_sort_dialog_dropdown_jk_navigation"
    "Sort Tests|test_sort_dialog_cancel"
    "Sort Tests|test_sort_dialog_q_cancel"
    "Sort Tests|test_sort_dialog_q_cancel_with_dropdown"
    "Sort Tests|test_sort_by_size_desc"
    "Sort Tests|test_sort_persists_after_navigation"
    "Sort Tests|test_sort_independent_panes"
    "Sort Tests|test_sort_dialog_arrow_keys"
    "Sort Tests|test_sort_dialog_jk_major_item_navigation"
    "Sort Tests|test_sort_dialog_ok_button_confirmation"
    "Sort Tests|test_sort_dialog_tab_three_items"

    "Cursor Tests|test_cursor_preserved_after_view"
    "Cursor Tests|test_cursor_preserved_after_enter_view"
    "Cursor Tests|test_cursor_reset_when_file_deleted"
    "Cursor Tests|test_both_panes_preserve_cursor"
    "Cursor Tests|test_parent_nav_cursor_on_subdir_h_key"
    "Cursor Tests|test_parent_nav_cursor_on_subdir_dotdot"
    "Cursor Tests|test_parent_nav_cursor_on_subdir_l_key"
    "Cursor Tests|test_parent_nav_independent_pane_memory"

    "Shell Tests|test_shell_command_mode_enter"
    "Shell Tests|test_shell_command_input"
    "Shell Tests|test_shell_command_empty_enter"
    "Shell Tests|test_shell_command_ignored_with_dialog"
    "Shell Tests|test_shell_command_ignored_during_search"
    "Shell Tests|test_help_shows_shell_command"

    "Config Tests|test_config_auto_generated"
    "Config Tests|test_config_has_keybindings"
    "Config Tests|test_config_has_comments"
    "Config Tests|test_default_keybindings_without_config"
    "Config Tests|test_help_shows_pascalcase"

    "Bookmark Tests|test_bookmark_dialog_opens"
    "Bookmark Tests|test_add_bookmark_dialog"
    "Bookmark Tests|test_bookmark_empty_state"

    "History Navigation Tests|test_history_back"
    "History Navigation Tests|test_history_forward"
    "History Navigation Tests|test_history_multiple_levels"
    "History Navigation Tests|test_history_no_history"
    "History Navigation Tests|test_history_independent_from_previous"
    "History Navigation Tests|test_history_forward_cleared"
    "History Navigation Tests|test_history_parent_navigation"
    "History Navigation Tests|test_history_home_navigation"

    "Archive Tests|test_compress_format_dialog_opens"
    "Archive Tests|test_compress_format_navigation"
    "Archive Tests|test_compression_level_dialog"
    "Archive Tests|test_archive_name_dialog"
    "Archive Tests|test_archive_conflict_dialog"
    "Archive Tests|test_compress_cancel_workflow"
    "Archive Tests|test_compress_complete_workflow"
    "Archive Tests|test_extract_complete_workflow"
    "Archive Tests|test_multifile_compress"

    "Background Command Tests|test_bg_command_execution"
    "Background Command Tests|test_bg_mode_prompt"
    "Background Command Tests|test_bg_cancel_with_ctrlc"
    "Background Command Tests|test_bg_file_ops_during_execution"
    "Background Command Tests|test_bg_blocked_during_execution"

    "Cursor Preserve Tests|test_single_move_cursor_preserved"
    "Cursor Preserve Tests|test_single_move_dest_cursor_preserved"
    "Cursor Preserve Tests|test_move_last_file_cursor"
    "Cursor Preserve Tests|test_copy_cursor_preserved"
    "Cursor Preserve Tests|test_copy_dest_cursor_preserved"
    "Cursor Preserve Tests|test_rename_cursor_preserved"
    "Cursor Preserve Tests|test_batch_move_cursor_up"
    "Cursor Preserve Tests|test_batch_move_cursor_down"
    "Cursor Preserve Tests|test_batch_move_all_cursor_zero"
    "Cursor Preserve Tests|test_batch_cancel_cursor"
    "Cursor Preserve Tests|test_single_file_dir_move_cursor"
    "Cursor Preserve Tests|test_delete_cursor_not_regressed"
    "Cursor Preserve Tests|test_batch_delete_cursor_not_regressed"
    "Cursor Preserve Tests|test_directory_nav_cursor_not_regressed"

    "Mouse Tests|test_mouse_click_moves_cursor"
    "Mouse Tests|test_mouse_drag_marks_range"
    "Mouse Tests|test_mouse_double_click_enters_directory"
)

# Show usage
show_usage() {
    echo "Usage: $0 [category|--list|--help|--check-list]"
    echo ""
    echo "Run all E2E tests or a specific category."
    echo ""
    echo "Options:"
    echo "  --list, -l    List available test categories"
    echo "  --help, -h    Show this help message"
    echo "  --check-list  Check the run list against defined tests without running them"
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
        count=$(grep -c "^test_" "${RUNNER_DIR}/tests/${file}" 2>/dev/null || echo "?")
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

    local test_file="${RUNNER_DIR}/tests/${file}"
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

# Build the definition set: every top-level test_ function name found by
# scanning tests/*.sh directly (not the category map), mapped to its file.
# Populates the global associative array DEFINED_TESTS[name]=file.
build_defined_tests() {
    declare -gA DEFINED_TESTS=()

    local file name
    for file in "${RUNNER_DIR}"/tests/*.sh; do
        [ -f "$file" ] || continue
        while IFS= read -r name; do
            [ -n "$name" ] || continue
            DEFINED_TESTS["$name"]="$(basename "$file")"
        done < <(grep -ho "^test_[a-zA-Z0-9_]*" "$file")
    done
}

# Run-list guard: compares FULL_RUN_LIST against the definition set built by
# build_defined_tests. Populates the global arrays GUARD_UNDEFINED (run-list
# names with no matching test_ function) and GUARD_MISSING (defined test_
# functions absent from the run list, formatted as "name (file)").
# Used by both the full run and --check-list so they cannot drift apart.
# Returns 0 when there are no problems, 1 otherwise; never aborts the
# caller under `set -e` because it is always invoked as an if/while condition.
run_guard_checks() {
    GUARD_UNDEFINED=()
    GUARD_MISSING=()

    build_defined_tests

    local -A in_list=()
    local entry name
    for entry in "${FULL_RUN_LIST[@]}"; do
        name="${entry#*|}"
        in_list["$name"]=1
        if [ -z "${DEFINED_TESTS[$name]+x}" ]; then
            GUARD_UNDEFINED+=("$name")
        fi
    done

    local defined_name
    for defined_name in $(printf '%s\n' "${!DEFINED_TESTS[@]}" | sort); do
        if [ -z "${in_list[$defined_name]+x}" ]; then
            GUARD_MISSING+=("$defined_name (${DEFINED_TESTS[$defined_name]})")
        fi
    done

    if [ ${#GUARD_UNDEFINED[@]} -eq 0 ] && [ ${#GUARD_MISSING[@]} -eq 0 ]; then
        return 0
    fi
    return 1
}

# Report a run-list entry with no matching test_ function. It is never
# invoked; it counts as exactly one failure.
report_undefined_entry() {
    local name="$1"
    TESTS_RUN=$((TESTS_RUN + 1))
    TESTS_FAILED=$((TESTS_FAILED + 1))
    echo ""
    echo "--- Running: $name ---"
    echo -e "${RED}✗${NC} Run-list entry has no matching test_ function: $name"
}

# Report every defined test_ function absent from the run list. Each one
# counts as exactly one failure.
report_missing_tests() {
    if [ ${#GUARD_MISSING[@]} -eq 0 ]; then
        return 0
    fi

    echo ""
    echo "=== Run List Guard ==="
    local item
    for item in "${GUARD_MISSING[@]}"; do
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
        echo -e "${RED}✗${NC} Defined test not in run list: $item"
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
        source "${RUNNER_DIR}/tests/${file}"
    done

    run_guard_checks || true

    local prev_group="" entry group name
    for entry in "${FULL_RUN_LIST[@]}"; do
        group="${entry%%|*}"
        name="${entry#*|}"

        if [ "$group" != "$prev_group" ]; then
            echo ""
            echo "=== $group ==="
            prev_group="$group"
        fi

        if [ -n "${DEFINED_TESTS[$name]+x}" ]; then
            run_test "$name"
        else
            report_undefined_entry "$name"
        fi
    done

    report_missing_tests
}

# Check the run list against the definition set without running any test,
# and without starting tmux or duofm.
check_list() {
    local file
    for file in "${RUNNER_DIR}"/tests/*.sh; do
        [ -f "$file" ] || continue
        source "$file"
    done

    if run_guard_checks; then
        echo "OK: run list matches defined tests (${#FULL_RUN_LIST[@]} entries)"
        return 0
    fi

    local name item
    for name in "${GUARD_UNDEFINED[@]}"; do
        echo "Undefined run-list entry: $name"
    done
    for item in "${GUARD_MISSING[@]}"; do
        echo "Missing from run list: $item"
    done
    return 1
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
        --check-list)
            if check_list; then
                exit 0
            else
                exit 1
            fi
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
