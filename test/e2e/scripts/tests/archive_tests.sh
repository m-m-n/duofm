#!/bin/bash
# Archive Tests for duofm
#
# Description: Tests for archive (compress/extract) operations including
#              format selection, compression level, and conflict handling
# Tests: 6

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# ===========================================
# Test: Compress format dialog opens
# ===========================================
test_compress_format_dialog_opens() {
    # Pre-cleanup
    rm -f /testdata/user_owned/archive_test.txt 2>/dev/null || true

    # Create a test file
    echo "archive content" > /testdata/user_owned/archive_test.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned directory
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Clear filter
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to archive_test.txt
    send_keys "$CURRENT_SESSION" "/" "a" "r" "c" "h" "i" "v" "e" "_" "t" "e" "s" "t" "Enter"
    sleep 0.3

    # Open context menu with '@'
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    # Should show context menu with numbered Compress item
    assert_contains "$CURRENT_SESSION" "8. Compress" \
        "Context menu shows numbered Compress option"

    # Select Compress by its number
    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    # Should show format selection dialog
    assert_contains "$CURRENT_SESSION" "Select Archive Format" \
        "Format selection dialog appears"

    # Verify format options are shown
    assert_contains "$CURRENT_SESSION" "tar.gz" \
        "tar.gz format option is shown"

    # Cancel dialog
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    # Cleanup
    rm -f /testdata/user_owned/archive_test.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Compress format dialog navigation
# ===========================================
test_compress_format_navigation() {
    rm -f /testdata/user_owned/navtest_arch.txt 2>/dev/null || true
    echo "nav test" > /testdata/user_owned/navtest_arch.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned and file
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "/" "n" "a" "v" "t" "e" "s" "t" "_" "a" "r" "c" "h" "Enter"
    sleep 0.3

    # Open context menu with '@'
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "8. Compress" \
        "Context menu shows numbered Compress option"

    # Select Compress by its number
    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    # Navigate with j/k
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "k"
    sleep 0.2

    # Format dialog should still be open
    assert_contains "$CURRENT_SESSION" "Select Archive Format" \
        "Format dialog stays open during navigation"

    # Assert the numbered tar.gz line before selecting it
    assert_contains "$CURRENT_SESSION" "2. tar.gz" \
        "Format dialog shows numbered tar.gz option"

    # Test number selection (press 2 for tar.gz)
    send_keys "$CURRENT_SESSION" "2"
    sleep 0.5

    # Format dialog should close (format selected)
    assert_not_contains "$CURRENT_SESSION" "Select Archive Format" \
        "Format dialog closes after selection"

    # tar.gz always opens the compression level dialog next
    assert_contains "$CURRENT_SESSION" "Compression Level" \
        "Compression level dialog appears after format selection"

    # Cancel
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    rm -f /testdata/user_owned/navtest_arch.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Compression level dialog for gzip format
# ===========================================
test_compression_level_dialog() {
    rm -f /testdata/user_owned/level_test.txt 2>/dev/null || true
    echo "level test" > /testdata/user_owned/level_test.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned and file
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "/" "l" "e" "v" "e" "l" "_" "t" "e" "s" "t" "Enter"
    sleep 0.3

    # Open context menu and select Compress
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "8. Compress" \
        "Context menu shows numbered Compress option"

    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "2. tar.gz" \
        "Format dialog shows numbered tar.gz option"

    # Select tar.gz format (should have compression level)
    send_keys "$CURRENT_SESSION" "2"  # tar.gz
    sleep 0.5

    # Should show compression level dialog
    assert_contains "$CURRENT_SESSION" "Compression Level" \
        "Compression level dialog appears for tar.gz"

    assert_contains "$CURRENT_SESSION" "Level 6" \
        "Default compression level 6 is shown"

    # Navigate levels with j/k
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "j"
    sleep 0.2

    # Test direct selection with number key
    send_keys "$CURRENT_SESSION" "9"
    sleep 0.2

    assert_contains "$CURRENT_SESSION" "Level 9" \
        "Direct number selection works in compression level dialog"

    # Cancel
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    rm -f /testdata/user_owned/level_test.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Archive name dialog
# ===========================================
test_archive_name_dialog() {
    rm -f /testdata/user_owned/name_test.txt 2>/dev/null || true
    echo "name test" > /testdata/user_owned/name_test.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned and file
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "/" "n" "a" "m" "e" "_" "t" "e" "s" "t" "Enter"
    sleep 0.3

    # Open context menu and select Compress
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "8. Compress" \
        "Context menu shows numbered Compress option"

    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "1. tar" \
        "Format dialog shows numbered tar option"

    # Select tar format (no compression level dialog)
    send_keys "$CURRENT_SESSION" "1"  # tar
    sleep 0.5

    # Should show archive name dialog directly (tar skips compression level)
    assert_contains "$CURRENT_SESSION" "Archive Name" \
        "Archive name dialog appears for tar format"

    # Default name should include source filename
    assert_contains "$CURRENT_SESSION" "name_test" \
        "Default name is based on source filename"

    # Cancel
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    rm -f /testdata/user_owned/name_test.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Archive conflict dialog
# ===========================================
test_archive_conflict_dialog() {
    rm -f /testdata/user_owned/conflict_src.txt 2>/dev/null || true
    rm -f /testdata/user_owned/conflictdir/conflict_src.tar 2>/dev/null || true
    rm -rf /testdata/user_owned/conflictdir 2>/dev/null || true

    # Create source file
    echo "conflict source" > /testdata/user_owned/conflict_src.txt

    # Create destination directory with existing archive
    mkdir -p /testdata/user_owned/conflictdir
    echo "existing archive" > /testdata/user_owned/conflictdir/conflict_src.tar

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned and file
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Sync right pane (inactive pane's destination)
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Navigate right pane to conflictdir
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "c" "o" "n" "f" "l" "i" "c" "t" "d" "i" "r" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to conflict_src.txt
    send_keys "$CURRENT_SESSION" "/" "c" "o" "n" "f" "l" "i" "c" "t" "_" "s" "r" "c" "." "t" "x" "t" "Enter"
    sleep 0.3

    # Open context menu and select Compress
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "8. Compress" \
        "Context menu shows numbered Compress option"

    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "1. tar" \
        "Format dialog shows numbered tar option"

    # Select tar format
    send_keys "$CURRENT_SESSION" "1"
    sleep 0.5

    # Accept default archive name
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Should show conflict dialog (the destination already held this exact archive)
    assert_contains "$CURRENT_SESSION" "already exists" \
        "Archive conflict dialog appears"

    # Test options are shown
    assert_contains "$CURRENT_SESSION" "Overwrite" \
        "Overwrite option is shown"

    assert_contains "$CURRENT_SESSION" "Rename" \
        "Rename option is shown"

    assert_contains "$CURRENT_SESSION" "Cancel" \
        "Cancel option is shown"

    # Cancel with Esc
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    rm -f /testdata/user_owned/conflict_src.txt
    rm -rf /testdata/user_owned/conflictdir

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Cancel compress workflow at any stage
# ===========================================
test_compress_cancel_workflow() {
    rm -f /testdata/user_owned/cancel_test.txt 2>/dev/null || true
    echo "cancel test" > /testdata/user_owned/cancel_test.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned and file
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2
    send_keys "$CURRENT_SESSION" "/" "c" "a" "n" "c" "e" "l" "_" "t" "e" "s" "t" "Enter"
    sleep 0.3

    # --- Test cancel at format selection ---
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "8. Compress" \
        "Context menu shows numbered Compress option"

    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "Select Archive Format" \
        "Format dialog opens"

    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    assert_not_contains "$CURRENT_SESSION" "Select Archive Format" \
        "Format dialog closes on Escape"

    # --- Test cancel at compression level ---
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "8. Compress" \
        "Context menu shows numbered Compress option"

    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "2. tar.gz" \
        "Format dialog shows numbered tar.gz option"

    send_keys "$CURRENT_SESSION" "2"  # tar.gz
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "Compression Level" \
        "Compression level dialog opens"

    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    assert_not_contains "$CURRENT_SESSION" "Compression Level" \
        "Compression level dialog closes on Escape"

    # --- Test cancel at archive name ---
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "8. Compress" \
        "Context menu shows numbered Compress option"

    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "1. tar" \
        "Format dialog shows numbered tar option"

    send_keys "$CURRENT_SESSION" "1"  # tar
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "Archive Name" \
        "Archive name dialog opens"

    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.3

    assert_not_contains "$CURRENT_SESSION" "Archive Name" \
        "Archive name dialog closes on Escape"

    rm -f /testdata/user_owned/cancel_test.txt

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Complete compress workflow (single file)
# ===========================================
test_compress_complete_workflow() {
    rm -f /testdata/user_owned/compress_test.txt 2>/dev/null || true
    rm -f /testdata/user_owned/compress_test.tar.gz 2>/dev/null || true

    # Create source file
    echo "compress workflow test content" > /testdata/user_owned/compress_test.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Point the inactive (right) pane at the destination before the operation
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Navigate to the source file
    send_keys "$CURRENT_SESSION" "/" "c" "o" "m" "p" "r" "e" "s" "s" "_" "t" "e" "s" "t" "." "t" "x" "t" "Enter"
    sleep 0.3

    # Open context menu and select Compress
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "8. Compress" \
        "Context menu shows numbered Compress option"

    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "2. tar.gz" \
        "Format dialog shows numbered tar.gz option"

    # Select tar.gz format
    send_keys "$CURRENT_SESSION" "2"
    sleep 0.5

    # Select compression level (use default 6)
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Accept default archive name
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 1.5

    # Verify archive was created at the destination (inactive pane's directory)
    if [ -f "/testdata/user_owned/compress_test.tar.gz" ]; then
        echo -e "${GREEN}✓${NC} Archive was created successfully"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Archive was not created"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Verify archive contains expected content
    local content
    content=$(tar -tzf /testdata/user_owned/compress_test.tar.gz 2>/dev/null | grep "compress_test.txt" || true)
    if [ -n "$content" ]; then
        echo -e "${GREEN}✓${NC} Archive contains expected file"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Archive does not contain expected file"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    rm -f /testdata/user_owned/compress_test.txt
    rm -f /testdata/user_owned/compress_test.tar.gz

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Complete extract workflow
# ===========================================
test_extract_complete_workflow() {
    rm -f /testdata/user_owned/extract_test.txt 2>/dev/null || true
    rm -f /testdata/user_owned/extract_test.tar.gz 2>/dev/null || true
    rm -rf /testdata/user_owned/extract_dest 2>/dev/null || true

    # Create test archive
    echo "extract workflow test content" > /testdata/user_owned/extract_test.txt
    tar -czf /testdata/user_owned/extract_test.tar.gz -C /testdata/user_owned extract_test.txt
    rm -f /testdata/user_owned/extract_test.txt

    # Create destination directory
    mkdir -p /testdata/user_owned/extract_dest

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Sync right pane and navigate to extract_dest (the destination)
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3
    send_keys "$CURRENT_SESSION" "l"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "/" "e" "x" "t" "r" "a" "c" "t" "_" "d" "e" "s" "t" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3

    # Switch back to left pane
    send_keys "$CURRENT_SESSION" "h"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Navigate to archive file
    send_keys "$CURRENT_SESSION" "/" "e" "x" "t" "r" "a" "c" "t" "_" "t" "e" "s" "t" "." "t" "a" "r" "." "g" "z" "Enter"
    sleep 0.3

    # Open context menu and select Extract
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "9. Extract archive" \
        "Context menu shows numbered Extract archive option"

    send_keys "$CURRENT_SESSION" "9"
    sleep 1.5

    # Verify file was extracted at the destination (inactive pane's directory)
    if [ -f "/testdata/user_owned/extract_dest/extract_test.txt" ]; then
        echo -e "${GREEN}✓${NC} File was extracted successfully"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} File was not extracted"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Verify content
    local content
    content=$(cat /testdata/user_owned/extract_dest/extract_test.txt 2>/dev/null || true)
    if [ "$content" = "extract workflow test content" ]; then
        echo -e "${GREEN}✓${NC} Extracted file has correct content"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Extracted file has incorrect content"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    rm -f /testdata/user_owned/extract_test.tar.gz
    rm -rf /testdata/user_owned/extract_dest

    stop_duofm "$CURRENT_SESSION"
}

# ===========================================
# Test: Multi-file compression
# ===========================================
test_multifile_compress() {
    rm -f /testdata/user_owned/multi1.txt 2>/dev/null || true
    rm -f /testdata/user_owned/multi2.txt 2>/dev/null || true
    rm -f /testdata/user_owned/multi3.txt 2>/dev/null || true
    rm -f /testdata/user_owned/user_owned.tar.gz 2>/dev/null || true

    # Create multiple test files
    echo "multi file 1" > /testdata/user_owned/multi1.txt
    echo "multi file 2" > /testdata/user_owned/multi2.txt
    echo "multi file 3" > /testdata/user_owned/multi3.txt

    start_duofm "$CURRENT_SESSION"

    # Navigate to user_owned
    send_keys "$CURRENT_SESSION" "/" "u" "s" "e" "r" "_" "o" "w" "n" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Escape"
    sleep 0.2

    # Point the inactive (right) pane at the destination before the operation
    send_keys "$CURRENT_SESSION" "="
    sleep 0.3

    # Mark multiple files
    send_keys "$CURRENT_SESSION" "/" "m" "u" "l" "t" "i" "1" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Space"  # Mark first file
    sleep 0.2

    send_keys "$CURRENT_SESSION" "/" "m" "u" "l" "t" "i" "2" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Space"  # Mark second file
    sleep 0.2

    send_keys "$CURRENT_SESSION" "/" "m" "u" "l" "t" "i" "3" "Enter"
    sleep 0.3
    send_keys "$CURRENT_SESSION" "Space"  # Mark third file
    sleep 0.2

    # Open context menu - should show numbered "Compress 3 files"
    send_keys "$CURRENT_SESSION" "@"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "8. Compress 3 files" \
        "Context menu shows numbered Compress 3 files option"

    # Select Compress by its number
    send_keys "$CURRENT_SESSION" "8"
    sleep 0.5

    assert_contains "$CURRENT_SESSION" "2. tar.gz" \
        "Format dialog shows numbered tar.gz option"

    # Select tar.gz format
    send_keys "$CURRENT_SESSION" "2"
    sleep 0.5

    # Select compression level (use default)
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 0.5

    # Accept default archive name (parent directory name: user_owned)
    send_keys "$CURRENT_SESSION" "Enter"
    sleep 1.5

    # Verify archive was created at the destination, named after the parent directory
    if [ -f "/testdata/user_owned/user_owned.tar.gz" ]; then
        echo -e "${GREEN}✓${NC} Multi-file archive was created"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Multi-file archive was not created"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    # Verify archive contains all 3 marked files
    local filecount
    filecount=$(tar -tzf /testdata/user_owned/user_owned.tar.gz 2>/dev/null | grep -c "multi.*\.txt" || true)
    if [ "$filecount" -ge 3 ]; then
        echo -e "${GREEN}✓${NC} Archive contains all 3 files"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Archive does not contain all files"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    rm -f /testdata/user_owned/multi1.txt
    rm -f /testdata/user_owned/multi2.txt
    rm -f /testdata/user_owned/multi3.txt
    rm -f /testdata/user_owned/user_owned.tar.gz

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Archive"
    echo "========================================"

    run_test test_compress_format_dialog_opens
    run_test test_compress_format_navigation
    run_test test_compression_level_dialog
    run_test test_archive_name_dialog
    run_test test_archive_conflict_dialog
    run_test test_compress_cancel_workflow
    run_test test_compress_complete_workflow
    run_test test_extract_complete_workflow
    run_test test_multifile_compress

    print_summary
    exit $?
fi
