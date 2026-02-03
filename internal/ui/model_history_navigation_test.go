package ui

import (
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sakura/duofm/internal/config"
)

// Phase 1: Up/Down History Navigation Tests

func TestHistoryNavigation_UpOnEmptyHistory(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	// Create model with history enabled
	m := createModelWithHistory(t, historyFile, 100)

	// Enter shell command mode
	m.startShellCommandMode()

	// Press Up key - should do nothing with empty history
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify minibuffer input is still empty
	if m.minibuffer.Input() != "" {
		t.Errorf("Up on empty history: input = %q, want empty", m.minibuffer.Input())
	}

	// Verify historyIndex is still -1
	if m.historyIndex != -1 {
		t.Errorf("historyIndex = %d, want -1", m.historyIndex)
	}
}

func TestHistoryNavigation_UpShowsMostRecentCommand(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add commands to history
	m.shellHistory.Add("old command")
	m.shellHistory.Add("recent command")

	// Enter shell command mode
	m.startShellCommandMode()

	// Press Up key
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should show most recent command
	if m.minibuffer.Input() != "recent command" {
		t.Errorf("Up first press: input = %q, want %q", m.minibuffer.Input(), "recent command")
	}

	// historyIndex should be 0
	if m.historyIndex != 0 {
		t.Errorf("historyIndex = %d, want 0", m.historyIndex)
	}
}

func TestHistoryNavigation_UpShowsOlderCommands(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add commands to history
	m.shellHistory.Add("cmd1")
	m.shellHistory.Add("cmd2")
	m.shellHistory.Add("cmd3")

	// Enter shell command mode
	m.startShellCommandMode()

	// Press Up key 3 times
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}

	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)
	if m.minibuffer.Input() != "cmd3" {
		t.Errorf("After 1st Up: input = %q, want %q", m.minibuffer.Input(), "cmd3")
	}

	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)
	if m.minibuffer.Input() != "cmd2" {
		t.Errorf("After 2nd Up: input = %q, want %q", m.minibuffer.Input(), "cmd2")
	}

	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)
	if m.minibuffer.Input() != "cmd1" {
		t.Errorf("After 3rd Up: input = %q, want %q", m.minibuffer.Input(), "cmd1")
	}
}

func TestHistoryNavigation_UpAtOldestDoesNotAdvance(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add only 2 commands
	m.shellHistory.Add("cmd1")
	m.shellHistory.Add("cmd2")

	// Enter shell command mode
	m.startShellCommandMode()

	// Press Up 3 times (more than history length)
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}

	for i := 0; i < 3; i++ {
		updatedModel, _ := m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// Should still show oldest command
	if m.minibuffer.Input() != "cmd1" {
		t.Errorf("After exceeding history: input = %q, want %q", m.minibuffer.Input(), "cmd1")
	}

	// historyIndex should not exceed bounds
	if m.historyIndex != 1 {
		t.Errorf("historyIndex = %d, want 1 (history length - 1)", m.historyIndex)
	}
}

func TestHistoryNavigation_DownShowsNewerCommand(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add commands to history
	m.shellHistory.Add("cmd1")
	m.shellHistory.Add("cmd2")
	m.shellHistory.Add("cmd3")

	// Enter shell command mode
	m.startShellCommandMode()

	// Navigate to oldest command
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	for i := 0; i < 3; i++ {
		updatedModel, _ := m.Update(keyMsg)
		m = updatedModel.(Model)
	}

	// Now press Down
	keyMsg = tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should show newer command
	if m.minibuffer.Input() != "cmd2" {
		t.Errorf("After Down: input = %q, want %q", m.minibuffer.Input(), "cmd2")
	}
}

func TestHistoryNavigation_DownAtNewestRestoresOriginalInput(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add commands to history
	m.shellHistory.Add("cmd1")
	m.shellHistory.Add("cmd2")

	// Enter shell command mode and type something
	m.startShellCommandMode()
	m.minibuffer.SetInput("my partial input")

	// Navigate up
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify we're looking at history
	if m.minibuffer.Input() != "cmd2" {
		t.Errorf("After Up: input = %q, want %q", m.minibuffer.Input(), "cmd2")
	}

	// Navigate back down
	keyMsg = tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Should restore original input
	if m.minibuffer.Input() != "my partial input" {
		t.Errorf("After Down to -1: input = %q, want %q", m.minibuffer.Input(), "my partial input")
	}

	// historyIndex should be -1
	if m.historyIndex != -1 {
		t.Errorf("historyIndex = %d, want -1", m.historyIndex)
	}
}

func TestHistoryNavigation_EditBufferSavedOnlyOnFirstUp(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add commands to history
	m.shellHistory.Add("cmd1")
	m.shellHistory.Add("cmd2")

	// Enter shell command mode and type something
	m.startShellCommandMode()
	m.minibuffer.SetInput("original")

	// Press Up first time
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Edit buffer should be saved
	if m.historyEditBuf != "original" {
		t.Errorf("historyEditBuf = %q, want %q", m.historyEditBuf, "original")
	}

	// Press Up again
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Edit buffer should still be "original" (not overwritten)
	if m.historyEditBuf != "original" {
		t.Errorf("historyEditBuf after 2nd Up = %q, want %q", m.historyEditBuf, "original")
	}
}

func TestHistoryNavigation_EnteringShellModeResetsIndex(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add commands to history
	m.shellHistory.Add("cmd1")

	// Enter shell command mode
	m.startShellCommandMode()

	// Navigate up
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Exit shell mode with Esc
	keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// Re-enter shell mode
	m.startShellCommandMode()

	// historyIndex should be reset to -1
	if m.historyIndex != -1 {
		t.Errorf("historyIndex after re-enter = %d, want -1", m.historyIndex)
	}

	// historyEditBuf should be cleared
	if m.historyEditBuf != "" {
		t.Errorf("historyEditBuf after re-enter = %q, want empty", m.historyEditBuf)
	}
}

func TestHistoryNavigation_DownWithNoNavigationDoesNothing(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add commands to history
	m.shellHistory.Add("cmd1")

	// Enter shell command mode
	m.startShellCommandMode()
	m.minibuffer.SetInput("my input")

	// Press Down without having navigated Up first
	keyMsg := tea.KeyMsg{Type: tea.KeyDown}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Input should remain unchanged
	if m.minibuffer.Input() != "my input" {
		t.Errorf("Down without prior Up: input = %q, want %q", m.minibuffer.Input(), "my input")
	}

	// historyIndex should still be -1
	if m.historyIndex != -1 {
		t.Errorf("historyIndex = %d, want -1", m.historyIndex)
	}
}

func TestHistoryNavigation_UpDownDuringHistorySearchDoesNothing(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add commands to history
	m.shellHistory.Add("cmd1")
	m.shellHistory.Add("cmd2")

	// Enter shell command mode
	m.startShellCommandMode()

	// Start history search (Ctrl+R)
	keyMsg := tea.KeyMsg{Type: tea.KeyCtrlR}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Verify we're in history search mode
	if !m.historySearching {
		t.Fatal("Expected to be in history search mode")
	}

	// Press Up - should not change historyIndex
	keyMsg = tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(Model)

	// historyIndex should still be -1 (navigation is disabled during search)
	if m.historyIndex != -1 {
		t.Errorf("historyIndex during search = %d, want -1", m.historyIndex)
	}
}

func TestHistoryNavigation_DisabledHistory(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	// Create model with history disabled (limit = 0)
	m := createModelWithHistory(t, historyFile, 0)

	// Enter shell command mode
	m.startShellCommandMode()
	m.minibuffer.SetInput("my input")

	// Press Up - should do nothing
	keyMsg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.Update(keyMsg)
	m = updatedModel.(Model)

	// Input should remain unchanged
	if m.minibuffer.Input() != "my input" {
		t.Errorf("Up with disabled history: input = %q, want %q", m.minibuffer.Input(), "my input")
	}
}

func TestHistoryNavigation_Integration_UpDownSequence(t *testing.T) {
	tmpDir := t.TempDir()
	historyFile := filepath.Join(tmpDir, "history")

	m := createModelWithHistory(t, historyFile, 100)

	// Add commands
	m.shellHistory.Add("alpha")
	m.shellHistory.Add("beta")
	m.shellHistory.Add("gamma")

	// Enter shell mode and type
	m.startShellCommandMode()
	m.minibuffer.SetInput("delta")

	keyUp := tea.KeyMsg{Type: tea.KeyUp}
	keyDown := tea.KeyMsg{Type: tea.KeyDown}

	// Up x3
	for i := 0; i < 3; i++ {
		updatedModel, _ := m.Update(keyUp)
		m = updatedModel.(Model)
	}

	// Should be at "alpha"
	if m.minibuffer.Input() != "alpha" {
		t.Errorf("After 3 Ups: input = %q, want %q", m.minibuffer.Input(), "alpha")
	}

	// Down x2
	for i := 0; i < 2; i++ {
		updatedModel, _ := m.Update(keyDown)
		m = updatedModel.(Model)
	}

	// Should be at "gamma"
	if m.minibuffer.Input() != "gamma" {
		t.Errorf("After 2 Downs: input = %q, want %q", m.minibuffer.Input(), "gamma")
	}

	// Down x1 more to restore original
	updatedModel, _ := m.Update(keyDown)
	m = updatedModel.(Model)

	if m.minibuffer.Input() != "delta" {
		t.Errorf("After final Down: input = %q, want %q", m.minibuffer.Input(), "delta")
	}
}

// Helper function to create a model with shell history enabled
func createModelWithHistory(t *testing.T, historyFile string, limit int) Model {
	t.Helper()

	model := NewModelWithConfig(nil, nil, nil, limit, config.DefaultEnterBehavior(), config.MIMEBehaviorConfig{}, "")
	if limit > 0 {
		model.shellHistory = NewShellHistory(historyFile, limit)
	}
	model.leftPath = t.TempDir()
	model.rightPath = t.TempDir()

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := model.Update(msg)
	return updatedModel.(Model)
}
