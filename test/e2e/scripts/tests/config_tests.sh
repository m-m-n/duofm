#!/bin/bash
# Configuration Tests for duofm
#
# Description: Tests for configuration file handling including
#              auto-generation, structure, and default keybindings
# Tests: 5

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/../helpers.sh"

# Test: Config file auto-generated on first run
test_config_auto_generated() {
    # Remove any existing config
    rm -f ~/.config/duofm/config.toml 2>/dev/null || true
    rm -rf ~/.config/duofm 2>/dev/null || true

    start_duofm "$CURRENT_SESSION"
    sleep 0.5

    # Check if config file was created
    if [ -f ~/.config/duofm/config.toml ]; then
        echo -e "${GREEN}✓${NC} Config file auto-generated on first run"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Config file should be auto-generated"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi

    stop_duofm "$CURRENT_SESSION"
}

# Test: Config file contains keybindings section
test_config_has_keybindings() {
    # Ensure config exists
    if [ ! -f ~/.config/duofm/config.toml ]; then
        start_duofm "$CURRENT_SESSION"
        sleep 0.5
        stop_duofm "$CURRENT_SESSION"
    fi

    # Check for [keybindings] section
    if grep -q "\[keybindings\]" ~/.config/duofm/config.toml; then
        echo -e "${GREEN}✓${NC} Config file has keybindings section"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Config file should have keybindings section"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Test: Config file has comments
test_config_has_comments() {
    if [ ! -f ~/.config/duofm/config.toml ]; then
        start_duofm "$CURRENT_SESSION"
        sleep 0.5
        stop_duofm "$CURRENT_SESSION"
    fi

    # Check for comments
    if grep -q "^#" ~/.config/duofm/config.toml; then
        echo -e "${GREEN}✓${NC} Config file has comments"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC} Config file should have comments"
        TESTS_RUN=$((TESTS_RUN + 1))
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

# Test: Default keybindings work without config
test_default_keybindings_without_config() {
    # Remove config file
    rm -f ~/.config/duofm/config.toml 2>/dev/null || true

    start_duofm "$CURRENT_SESSION"
    sleep 0.3

    # Test j/k navigation (default keybindings)
    send_keys "$CURRENT_SESSION" "j" "j"
    sleep 0.3

    assert_cursor_position "$CURRENT_SESSION" "3" \
        "Default j key moves cursor down"

    stop_duofm "$CURRENT_SESSION"
}

# Test: Help dialog shows PascalCase keys
test_help_shows_pascalcase() {
    start_duofm "$CURRENT_SESSION"

    # Open help
    send_keys "$CURRENT_SESSION" "?"
    sleep 0.3

    # Should show PascalCase keys
    assert_contains "$CURRENT_SESSION" "J/K" \
        "Help shows J/K in PascalCase"

    # Close help
    send_keys "$CURRENT_SESSION" "Escape"

    stop_duofm "$CURRENT_SESSION"
}

# Execute tests when run directly
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    echo "========================================"
    echo "duofm E2E Tests - Config"
    echo "========================================"

    run_test test_config_auto_generated
    run_test test_config_has_keybindings
    run_test test_config_has_comments
    run_test test_default_keybindings_without_config
    run_test test_help_shows_pascalcase

    print_summary
    exit $?
fi
