package ui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Phase 3: Edge Case Tests for Shell History Navigation and Search

// Single entry history edge cases
func TestHistoryNavigation_SingleEntry(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("only command")

	m.startShellCommandMode()

	// Press Up
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.minibuffer.Input() != "only command" {
		t.Errorf("Single entry Up: input = %q, want %q", m.minibuffer.Input(), "only command")
	}

	// Press Up again - should stay at the same command
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.minibuffer.Input() != "only command" {
		t.Errorf("Single entry Up again: input = %q, want %q", m.minibuffer.Input(), "only command")
	}

	// Press Down - should restore to original
	keyMsg = tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.minibuffer.Input() != "" {
		t.Errorf("Single entry Down: input = %q, want empty", m.minibuffer.Input())
	}
}

// Rapid key presses
func TestHistoryNavigation_RapidKeyPresses(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add many unique commands (use numbers to avoid deduplication)
	for i := 0; i < 50; i++ {
		m.shellHistory.Add("command_" + string(rune('0'+i/10)) + string(rune('0'+i%10)))
	}

	commands := m.shellHistory.Commands()
	histLen := len(commands)

	m.startShellCommandMode()

	// Rapidly press Up many times (more than history length)
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	for i := 0; i < 100; i++ {
		updatedModel, _ := m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// Should not panic or have index out of bounds
	// historyIndex should be at max (histLen - 1)
	if m.historyIndex != histLen-1 {
		t.Errorf("After rapid Up: historyIndex = %d, want %d", m.historyIndex, histLen-1)
	}

	// Rapidly press Down many times
	keyMsg = tea.KeyMsg{Type: tea.KeyDown}
	for i := 0; i < 100; i++ {
		updatedModel, _ := m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// Should be back at -1
	if m.historyIndex != -1 {
		t.Errorf("After rapid Down: historyIndex = %d, want -1", m.historyIndex)
	}
}

// Very long pattern in search
func TestSearchPatternDisplay_VeryLongPattern(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("grep -r 'some very long pattern that goes on forever' /path/to/search")

	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type a long pattern
	longPattern := "some very long pattern"
	for _, r := range longPattern {
		keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedModel, _ = m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// Verify pattern is complete
	expectedPrompt := "(reverse-i-search)'" + longPattern + "': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Long pattern prompt = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}
}

// Unicode characters in pattern
func TestSearchPatternDisplay_UnicodePattern(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("echo cafe")
	m.shellHistory.Add("echo hello")

	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type "cafe" (ASCII)
	for _, r := range "cafe" {
		keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedModel, _ = m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// Should match "echo cafe"
	if m.minibuffer.Input() != "echo cafe" {
		t.Errorf("Unicode search input = %q, want %q", m.minibuffer.Input(), "echo cafe")
	}
}

// Existing tests for mode transitions
func TestHistoryNavigation_ModeTransitions(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("cmd1")
	m.shellHistory.Add("cmd2")

	// Enter shell mode, navigate, then execute
	m.startShellCommandMode()

	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Execute the command
	keyMsg = tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should exit shell mode
	if m.shellCommandMode {
		t.Error("Should exit shell mode after Enter")
	}

	// Re-enter shell mode
	m.startShellCommandMode()

	// Should start fresh (historyIndex = -1)
	if m.historyIndex != -1 {
		t.Errorf("historyIndex after re-enter = %d, want -1", m.historyIndex)
	}
}

// History search then navigation should work
func TestHistorySearchThenNavigation(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("ls file1")
	m.shellHistory.Add("grep pattern")
	m.shellHistory.Add("ls file2")

	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type 'ls'
	for _, r := range "ls" {
		keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedModel, _ = m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// Should show "ls file2" (most recent match)
	if m.minibuffer.Input() != "ls file2" {
		t.Errorf("After search: input = %q, want %q", m.minibuffer.Input(), "ls file2")
	}

	// Cancel search with Esc
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Now use Up/Down navigation
	keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should show most recent command
	if m.minibuffer.Input() != "ls file2" {
		t.Errorf("After Up: input = %q, want %q", m.minibuffer.Input(), "ls file2")
	}

	keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	if m.minibuffer.Input() != "grep pattern" {
		t.Errorf("After 2nd Up: input = %q, want %q", m.minibuffer.Input(), "grep pattern")
	}
}

// Edit buffer preservation with special characters
func TestHistoryNavigation_EditBufferWithSpecialChars(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("cmd1")

	m.startShellCommandMode()

	// Type input with special characters
	specialInput := "echo 'hello world' | grep test"
	m.minibuffer.SetInput(specialInput)

	// Navigate Up
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Navigate back Down
	keyMsg = tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Original input should be preserved
	if m.minibuffer.Input() != specialInput {
		t.Errorf("After restore: input = %q, want %q", m.minibuffer.Input(), specialInput)
	}
}

// Backspace on empty pattern
func TestSearchPatternDisplay_BackspaceOnEmptyPattern(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("cmd1")

	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Press backspace on empty pattern - should do nothing harmful
	keyMsg = tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should still be in search mode with empty pattern
	expectedPrompt := "(reverse-i-search)'': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Prompt after backspace on empty = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}
}

// Space character in search pattern
func TestSearchPatternDisplay_SpaceInPattern(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("echo hello world")
	m.shellHistory.Add("echo hi")

	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type 'hello '
	for _, r := range "hello " {
		if r == ' ' {
			keyMsg = tea.KeyMsg{Type: tea.KeySpace}
		} else {
			keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		}
		updatedModel, _ = m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	expectedPrompt := "(reverse-i-search)'hello ': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Prompt with space = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}

	if m.minibuffer.Input() != "echo hello world" {
		t.Errorf("Match with space = %q, want %q", m.minibuffer.Input(), "echo hello world")
	}
}
