package ui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Phase 2: Search Pattern Display Tests

func TestHistorySearcher_Pattern(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	sh := NewShellHistory(historyFile, 100)
	defer sh.Close()

	sh.Add("ls -la")
	sh.Add("pwd")

	hs := NewHistorySearcher(sh)

	t.Run("Pattern returns empty string initially", func(t *testing.T) {
		if hs.Pattern() != "" {
			t.Errorf("Pattern() = %q, want empty", hs.Pattern())
		}
	})

	t.Run("Pattern returns current search pattern", func(t *testing.T) {
		hs.SetPattern("ls")
		if hs.Pattern() != "ls" {
			t.Errorf("Pattern() = %q, want %q", hs.Pattern(), "ls")
		}
	})

	t.Run("Pattern updates when SetPattern called", func(t *testing.T) {
		hs.SetPattern("pwd")
		if hs.Pattern() != "pwd" {
			t.Errorf("Pattern() = %q, want %q", hs.Pattern(), "pwd")
		}
	})

	t.Run("Pattern cleared after Reset", func(t *testing.T) {
		hs.SetPattern("test")
		hs.Reset()
		if hs.Pattern() != "" {
			t.Errorf("Pattern() after Reset = %q, want empty", hs.Pattern())
		}
	})
}

func TestSearchPatternDisplay_InitialPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("ls -la")

	// Enter shell command mode
	m.startShellCommandMode()

	// Start history search (Ctrl+R)
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify prompt shows empty pattern
	expectedPrompt := "(reverse-i-search)'': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Initial prompt = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}
}

func TestSearchPatternDisplay_TypingUpdatesPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("grep pattern file")
	m.shellHistory.Add("ls -la")

	// Enter shell command mode
	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type 'g'
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	expectedPrompt := "(reverse-i-search)'g': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Prompt after typing 'g' = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}

	// Type 'r'
	keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	expectedPrompt = "(reverse-i-search)'gr': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Prompt after typing 'gr' = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}
}

func TestSearchPatternDisplay_BackspaceRemovesCharacter(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("grep pattern")

	// Enter shell command mode
	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type 'grep'
	for _, r := range "grep" {
		keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedModel, _ = m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	expectedPrompt := "(reverse-i-search)'grep': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Prompt after typing 'grep' = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}

	// Press backspace
	keyMsg = tea.KeyMsg{Type: tea.KeyBackspace}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	expectedPrompt = "(reverse-i-search)'gre': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Prompt after backspace = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}
}

func TestSearchPatternDisplay_MatchedCommandDisplayed(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("grep pattern file.txt")
	m.shellHistory.Add("ls -la")

	// Enter shell command mode
	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type 'grep'
	for _, r := range "grep" {
		keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedModel, _ = m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// Matched command should be in minibuffer input
	if m.minibuffer.Input() != "grep pattern file.txt" {
		t.Errorf("Matched command = %q, want %q", m.minibuffer.Input(), "grep pattern file.txt")
	}
}

func TestSearchPatternDisplay_NoMatchShowsEmptyInput(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("ls -la")
	m.shellHistory.Add("pwd")

	// Enter shell command mode
	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type 'xyz' (no match)
	for _, r := range "xyz" {
		keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedModel, _ = m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// No match - input should be empty
	if m.minibuffer.Input() != "" {
		t.Errorf("Input with no match = %q, want empty", m.minibuffer.Input())
	}

	// But pattern should still be shown in prompt
	expectedPrompt := "(reverse-i-search)'xyz': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Prompt with no match = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}
}

func TestSearchPatternDisplay_PatternPersistsAcrossCtrlR(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("ls file1")
	m.shellHistory.Add("ls file2")
	m.shellHistory.Add("ls file3")

	// Enter shell command mode
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

	// First match
	if m.minibuffer.Input() != "ls file3" {
		t.Errorf("First match = %q, want %q", m.minibuffer.Input(), "ls file3")
	}

	// Press Ctrl+R again for next match
	keyMsg = tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Pattern should persist
	expectedPrompt := "(reverse-i-search)'ls': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Prompt after Ctrl+R = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}

	// Second match
	if m.minibuffer.Input() != "ls file2" {
		t.Errorf("Second match = %q, want %q", m.minibuffer.Input(), "ls file2")
	}
}

func TestSearchPatternDisplay_EscClearsPatternAndReturnsToShellMode(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("ls -la")

	// Enter shell command mode
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

	// Press Esc
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should be back in shell command mode (not history searching)
	if m.historySearching {
		t.Error("Should not be in history search mode after Esc")
	}

	// Should still be in shell command mode
	if !m.shellCommandMode {
		t.Error("Should still be in shell command mode after Esc")
	}

	// Prompt should be back to normal
	if m.minibuffer.Prompt() != "!: " {
		t.Errorf("Prompt after Esc = %q, want %q", m.minibuffer.Prompt(), "!: ")
	}
}

func TestSearchPatternDisplay_UnicodeCharacters(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)
	m.shellHistory.Add("echo nihao")
	m.shellHistory.Add("echo hello")

	// Enter shell command mode
	m.startShellCommandMode()

	// Start history search
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Type unicode characters
	for _, r := range "hao" {
		keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updatedModel, _ = m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	expectedPrompt := "(reverse-i-search)'hao': "
	if m.minibuffer.Prompt() != expectedPrompt {
		t.Errorf("Prompt with unicode = %q, want %q", m.minibuffer.Prompt(), expectedPrompt)
	}

	if m.minibuffer.Input() != "echo nihao" {
		t.Errorf("Matched command = %q, want %q", m.minibuffer.Input(), "echo nihao")
	}
}
