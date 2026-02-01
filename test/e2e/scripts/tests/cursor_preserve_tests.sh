#!/bin/bash
# Cursor Preserve After File Operation Tests for duofm
#
# Description: Tests for cursor position preservation after file operations
#              (move, copy, rename, batch move, delete) verifying that the
#              cursor stays at an appropriate position instead of resetting to 0.
# Tests: 14

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# Helper: Setup test directory with numbered files for predictable ordering
# Creates files: 01_aaa.txt through 05_eee.txt
setup_test_dir() {
    local dir="$1"
    rm -rf "$dir" 2>/dev/null || true
    mkdir -p "$dir"
    echo "a" > "$dir/01_aaa.txt"
    echo "b" > "$dir/02_bbb.txt"
    echo "c" > "$dir/03_ccc.txt"
    echo "d" > "$dir/04_ddd.txt"
    echo "e" > "$dir/05_eee.txt"
}

# Helper: Assert file exists on filesystem
assert_file_exists() {
    local filepath="$1"
    local description="$2"

    TESTS_RUN=$((TESTS_RUN + 1))
    if [ -f "$filepath" ]; then
        echo -e "${GREEN}✓${NC} $description"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $description"
        echo -e "  ${YELLOW}File not found:${NC} $filepath"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Helper: Assert file does NOT exist on filesystem
assert_file_not_exists() {
    local filepath="$1"
    local description="$2"

    TESTS_RUN=$((TESTS_RUN + 1))
    if [ ! -f "$filepath" ]; then
        echo -e "${GREEN}✓${NC} $description"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗${NC} $description"
        echo -e "  ${YELLOW}File should not exist:${NC} $filepath"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Helper: Navigate both panes - right pane to dst, left pane to src
# Both directories must be under the starting directory (e.g., /testdata/user_owned)
# Syncs right pane first since it starts at ~ (home directory)
# Usage: setup_dual_pane <session> <src_dirname> <dst_dirname>
setup_dual_pane() {
    local session="$1"
    local src="$2"
    local dst="$3"

    # Sync right pane to left pane's directory (right pane starts at ~)
    send_keys "$session" "="
    sleep 0.3

    # Right pane: navigate to destination directory
    send_keys "$session" "l"
    sleep 0.2
    send_keys "$session" "/" "$dst" "Enter"
    sleep 0.3
    send_keys "$session" "Enter"
    sleep 0.3

    # Left pane: navigate to source directory
    send_keys "$session" "h"
    sleep 0.2
    send_keys "$session" "Escape"
    sleep 0.2
    send_keys "$session" "/" "$src" "Enter"
    sleep 0.3
    send_keys "$session" "Enter"
    sleep 0.3
    send_keys "$session" "Escape"
    sleep 0.2
}

# ===========================================
# Test 1: Single file move - source pane cursor stays at same index
# ===========================================
test_single_move_cursor_preserved() {
    local srcdir="/testdata/user_owned/cpmv_src"
    local dstdir="/testdata/user_owned/cpmv_dst"
    setup_test_dir "$srcdir"
    mkdir -p "$dstdir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "cpmv_src" "cpmv_dst"

    # Move cursor to 03_ccc.txt (j j j: .. → 01 → 02 → 03)
    send_keys "$CURRENT_SESSION" "j" "j" "j"
    sleep 0.3

    # Verify cursor is on 03_ccc.txt
    assert_contains "$CURRENT_SESSION" "03_ccc" \
        "Cursor on 03_ccc before move"

    # Move file with 'm'
    send_keys "$CURRENT_SESSION" "m"
    sleep 0.8

    # After move, cursor should stay at same index (now 04_ddd.txt)
    assert_contains "$CURRENT_SESSION" "04_ddd" \
        "Cursor on next file (04_ddd) after moving 03_ccc"

    # Verify via filesystem
    assert_file_not_exists "$srcdir/03_ccc.txt" \
        "03_ccc.txt removed from source directory"
    assert_file_exists "$dstdir/03_ccc.txt" \
        "03_ccc.txt present in destination directory"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 2: Single file move - destination pane cursor preserved
# ===========================================
test_single_move_dest_cursor_preserved() {
    local srcdir="/testdata/user_owned/dest_src"
    local dstdir="/testdata/user_owned/dest_dst"
    # Create a non-conflicting file in source
    rm -rf "$srcdir" "$dstdir" 2>/dev/null || true
    mkdir -p "$srcdir" "$dstdir"
    echo "extra" > "$srcdir/00_extra.txt"
    setup_test_dir "$dstdir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "dest_src" "dest_dst"

    # Move file with 'm' (cursor is on .. so move to file first)
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "m"
    sleep 0.8

    # Verify the move happened via filesystem
    assert_file_exists "$dstdir/00_extra.txt" \
        "File moved to destination"
    assert_file_not_exists "$srcdir/00_extra.txt" \
        "File removed from source"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 3: Move last file - cursor moves to previous entry
# ===========================================
test_move_last_file_cursor() {
    local srcdir="/testdata/user_owned/last_src"
    local dstdir="/testdata/user_owned/last_dst"
    setup_test_dir "$srcdir"
    mkdir -p "$dstdir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "last_src" "last_dst"

    # Go to last file with G
    send_keys "$CURRENT_SESSION" "G"
    sleep 0.3

    # Verify on last file
    assert_contains "$CURRENT_SESSION" "05_eee" \
        "Cursor on last file before move"

    # Move last file
    send_keys "$CURRENT_SESSION" "m"
    sleep 0.8

    # Cursor should now be on 04_ddd (the new last file)
    assert_contains "$CURRENT_SESSION" "04_ddd" \
        "Cursor on new last file after moving last file"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 4: Copy - source pane cursor preserved
# ===========================================
test_copy_cursor_preserved() {
    local srcdir="/testdata/user_owned/copy_src"
    local dstdir="/testdata/user_owned/copy_dst"
    setup_test_dir "$srcdir"
    mkdir -p "$dstdir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "copy_src" "copy_dst"

    # Move cursor to 03_ccc.txt
    send_keys "$CURRENT_SESSION" "j" "j" "j"
    sleep 0.3

    assert_contains "$CURRENT_SESSION" "03_ccc" \
        "Cursor on 03_ccc before copy"

    # Copy file with 'c'
    send_keys "$CURRENT_SESSION" "c"
    sleep 0.8

    # Source file should still exist and cursor should be on same file
    assert_contains "$CURRENT_SESSION" "03_ccc" \
        "Source file still visible after copy"

    # Verify copy via filesystem
    assert_file_exists "$dstdir/03_ccc.txt" \
        "File copied to destination"
    assert_file_exists "$srcdir/03_ccc.txt" \
        "Source file preserved after copy"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 5: Copy - destination pane cursor preserved
# ===========================================
test_copy_dest_cursor_preserved() {
    local srcdir="/testdata/user_owned/cpd_src"
    local dstdir="/testdata/user_owned/cpd_dst"
    rm -rf "$srcdir" "$dstdir" 2>/dev/null || true
    mkdir -p "$srcdir" "$dstdir"
    echo "extra" > "$srcdir/00_extra.txt"
    setup_test_dir "$dstdir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "cpd_src" "cpd_dst"

    # Move to file and copy
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "c"
    sleep 0.8

    # File should be copied
    assert_file_exists "$dstdir/00_extra.txt" \
        "File copied to destination"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 6: Rename - cursor preserved by base reload
# ===========================================
test_rename_cursor_preserved() {
    local dir="/testdata/user_owned/ren_dir"
    setup_test_dir "$dir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"

    # Enter ren_dir
    send_keys "$CURRENT_SESSION" "/" "r" "e" "n" "_" "d" "i" "r" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Move to 03_ccc.txt
    send_keys "$CURRENT_SESSION" "j" "j" "j"
    sleep 0.3

    assert_contains "$CURRENT_SESSION" "03_ccc" \
        "Cursor on 03_ccc before rename"

    # Rename with R key (extension-preserving rename)
    send_keys "$CURRENT_SESSION" "R"
    sleep 0.5

    # Type new name
    send_keys "$CURRENT_SESSION" "C-u"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "0" "3" "_" "z" "z" "z"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Cursor should be on renamed file
    assert_contains "$CURRENT_SESSION" "03_zzz" \
        "Cursor on renamed file"

    rm -rf "$dir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 7: Batch move - cursor on nearest unmarked file (up direction)
# ===========================================
test_batch_move_cursor_up() {
    local srcdir="/testdata/user_owned/bmup_src"
    local dstdir="/testdata/user_owned/bmup_dst"
    setup_test_dir "$srcdir"
    mkdir -p "$dstdir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "bmup_src" "bmup_dst"

    # Mark 03_ccc and 04_ddd (skip 01, 02)
    # Move to 03_ccc (j j j: .. → 01 → 02 → 03)
    send_keys "$CURRENT_SESSION" "j" "j" "j"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 03, cursor moves to 04
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 04, cursor moves to 05
    sleep 0.2

    # Move cursor back to a marked file area
    send_keys "$CURRENT_SESSION" "k"
    sleep 0.2

    # Batch move with 'm'
    send_keys "$CURRENT_SESSION" "m"
    sleep 1.5

    # After batch move, cursor should be on 02_bbb (nearest unmarked up from marked)
    assert_contains "$CURRENT_SESSION" "02_bbb" \
        "Cursor on nearest unmarked file (02_bbb) after batch move"

    # Verify via filesystem
    assert_file_not_exists "$srcdir/03_ccc.txt" \
        "03_ccc.txt moved from source"
    assert_file_not_exists "$srcdir/04_ddd.txt" \
        "04_ddd.txt moved from source"
    assert_file_exists "$dstdir/03_ccc.txt" \
        "03_ccc.txt in destination"
    assert_file_exists "$dstdir/04_ddd.txt" \
        "04_ddd.txt in destination"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 8: Batch move - cursor searches down when up has no unmarked
# ===========================================
test_batch_move_cursor_down() {
    local srcdir="/testdata/user_owned/bmdn_src"
    local dstdir="/testdata/user_owned/bmdn_dst"
    setup_test_dir "$srcdir"
    mkdir -p "$dstdir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "bmdn_src" "bmdn_dst"

    # Mark 01_aaa and 02_bbb (first two files)
    send_keys "$CURRENT_SESSION" "j"  # Move to 01_aaa
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 01, cursor to 02
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 02, cursor to 03
    sleep 0.2

    # Move cursor up to marked area
    send_keys "$CURRENT_SESSION" "k" "k"
    sleep 0.2

    # Batch move
    send_keys "$CURRENT_SESSION" "m"
    sleep 1.5

    # Cursor should find 03_ccc (first unmarked going down since up is all marked)
    assert_contains "$CURRENT_SESSION" "03_ccc" \
        "Cursor on first unmarked file downward (03_ccc) after batch move"

    # Verify via filesystem
    assert_file_not_exists "$srcdir/01_aaa.txt" \
        "01_aaa.txt moved from source"
    assert_file_not_exists "$srcdir/02_bbb.txt" \
        "02_bbb.txt moved from source"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 9: Batch move all files - cursor at 0 (..)
# ===========================================
test_batch_move_all_cursor_zero() {
    local srcdir="/testdata/user_owned/bmall_src"
    local dstdir="/testdata/user_owned/bmall_dst"
    rm -rf "$srcdir" "$dstdir" 2>/dev/null || true
    mkdir -p "$srcdir" "$dstdir"
    echo "a" > "$srcdir/01_aaa.txt"
    echo "b" > "$srcdir/02_bbb.txt"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "bmall_src" "bmall_dst"

    # Mark all files
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 01
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 02
    sleep 0.2

    # Batch move
    send_keys "$CURRENT_SESSION" "m"
    sleep 1.5

    # Verify files moved via filesystem
    assert_file_not_exists "$srcdir/01_aaa.txt" \
        "01_aaa.txt moved from source"
    assert_file_not_exists "$srcdir/02_bbb.txt" \
        "02_bbb.txt moved from source"
    assert_file_exists "$dstdir/01_aaa.txt" \
        "01_aaa.txt in destination"
    assert_file_exists "$dstdir/02_bbb.txt" \
        "02_bbb.txt in destination"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 10: Batch cancel - cursor at appropriate position
# ===========================================
test_batch_cancel_cursor() {
    local srcdir="/testdata/user_owned/bcan_src"
    local dstdir="/testdata/user_owned/bcan_dst"
    setup_test_dir "$srcdir"
    mkdir -p "$dstdir"
    # Create conflicting file for second file so we get a dialog to cancel
    echo "existing" > "$dstdir/02_bbb.txt"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "bcan_src" "bcan_dst"

    # Mark first two files
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 01
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 02
    sleep 0.2

    # Start batch move
    send_keys "$CURRENT_SESSION" "m"
    sleep 1.0

    # First file (01_aaa) should move successfully
    # Second file (02_bbb) should show overwrite dialog
    # Cancel on overwrite dialog
    send_keys "$CURRENT_SESSION" "2"
    sleep 0.5

    # Application should still be functional
    assert_contains "$CURRENT_SESSION" "duofm" \
        "Application still running after batch cancel"

    # 01_aaa should have been moved (first in batch)
    assert_file_exists "$dstdir/01_aaa.txt" \
        "First file was moved before cancel"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 11: Single file + 1 directory - move single file, cursor stays
# ===========================================
test_single_file_dir_move_cursor() {
    local srcdir="/testdata/user_owned/sfdir_src"
    local dstdir="/testdata/user_owned/sfdir_dst"
    rm -rf "$srcdir" "$dstdir" 2>/dev/null || true
    mkdir -p "$srcdir" "$dstdir"
    echo "only" > "$srcdir/only_file.txt"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"
    setup_dual_pane "$CURRENT_SESSION" "sfdir_src" "sfdir_dst"

    # Move to the only file
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2

    # Move the file
    send_keys "$CURRENT_SESSION" "m"
    sleep 0.8

    # Verify file moved via filesystem
    assert_file_not_exists "$srcdir/only_file.txt" \
        "File removed from source"
    assert_file_exists "$dstdir/only_file.txt" \
        "File moved to destination"

    rm -rf "$srcdir" "$dstdir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 12: Single delete - cursor not affected by this change
# ===========================================
test_delete_cursor_not_regressed() {
    local dir="/testdata/user_owned/del_test"
    setup_test_dir "$dir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"

    # Enter del_test
    send_keys "$CURRENT_SESSION" "/" "d" "e" "l" "_" "t" "e" "s" "t" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Move to 03_ccc (j j j: .. → 01 → 02 → 03)
    send_keys "$CURRENT_SESSION" "j" "j" "j"
    sleep 0.3

    assert_contains "$CURRENT_SESSION" "03_ccc" \
        "Cursor on 03_ccc before delete"

    # Delete with 'd' and confirm with 'y'
    send_keys "$CURRENT_SESSION" "d"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "y"
    sleep 0.5

    # Cursor should be on 04_ddd (next file at same index)
    assert_contains "$CURRENT_SESSION" "04_ddd" \
        "Cursor on next file after delete (not reset to top)"

    rm -rf "$dir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 13: Batch delete - cursor not affected
# ===========================================
test_batch_delete_cursor_not_regressed() {
    local dir="/testdata/user_owned/bdel_test"
    setup_test_dir "$dir"

    start_duofm "$CURRENT_SESSION" "/testdata/user_owned"

    # Enter bdel_test
    send_keys "$CURRENT_SESSION" "/" "b" "d" "e" "l" "_" "t" "e" "s" "t" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Mark 02_bbb and 03_ccc
    send_keys "$CURRENT_SESSION" "j" "j"  # Move to 02_bbb
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 02
    sleep 0.2
    send_keys "$CURRENT_SESSION" "Space"  # Mark 03
    sleep 0.2

    # Delete marked files
    send_keys "$CURRENT_SESSION" "d"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "y"
    sleep 0.5

    # Verify deleted via filesystem
    assert_file_not_exists "$dir/02_bbb.txt" \
        "02_bbb.txt deleted"
    assert_file_not_exists "$dir/03_ccc.txt" \
        "03_ccc.txt deleted"

    # Remaining files should still be visible
    assert_contains "$CURRENT_SESSION" "01_aaa" \
        "01_aaa still present"
    assert_contains "$CURRENT_SESSION" "04_ddd" \
        "04_ddd still present"

    rm -rf "$dir"
    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test 14: Directory navigation - cursor not affected
# ===========================================
test_directory_nav_cursor_not_regressed() {
    start_duofm "$CURRENT_SESSION"

    # Navigate into dir1
    send_keys "$CURRENT_SESSION" "/" "d" "i" "r" "1" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Go back with h
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.5

    # Cursor should be on dir1 (remembered from navigation)
    assert_contains "$CURRENT_SESSION" "dir1" \
        "Cursor on dir1 after navigating back"

    # Navigate into dir2
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "/" "d" "i" "r" "2" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Go back
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.5

    # Cursor should be on dir2
    assert_contains "$CURRENT_SESSION" "dir2" \
        "Cursor on dir2 after navigating back"

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Cursor Preserve After File Operations"
    echo "========================================"

    run_test test_single_move_cursor_preserved
    run_test test_single_move_dest_cursor_preserved
    run_test test_move_last_file_cursor
    run_test test_copy_cursor_preserved
    run_test test_copy_dest_cursor_preserved
    run_test test_rename_cursor_preserved
    run_test test_batch_move_cursor_up
    run_test test_batch_move_cursor_down
    run_test test_batch_move_all_cursor_zero
    run_test test_batch_cancel_cursor
    run_test test_single_file_dir_move_cursor
    run_test test_delete_cursor_not_regressed
    run_test test_batch_delete_cursor_not_regressed
    run_test test_directory_nav_cursor_not_regressed

    print_summary
    exit $?
fi
