package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConfigErrorDialog_StartupFix(t *testing.T) {
	d := NewConfigErrorDialog("Syntax error at line 23", "Details here")

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	_, cmd := d.Update(keyMsg)

	if cmd == nil {
		t.Fatal("Expected command from 'f' key")
	}

	msg := cmd()
	result, ok := msg.(configErrorDialogResultMsg)
	if !ok {
		t.Fatalf("Expected configErrorDialogResultMsg, got %T", msg)
	}

	if result.choice != ConfigErrorChoiceFix {
		t.Errorf("Expected ConfigErrorChoiceFix, got %v", result.choice)
	}

	if !result.isStartup {
		t.Error("Expected isStartup=true for startup dialog")
	}
}

func TestConfigErrorDialog_StartupQuit(t *testing.T) {
	d := NewConfigErrorDialog("Error", "Details")

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := d.Update(keyMsg)

	if cmd == nil {
		t.Fatal("Expected command from 'q' key")
	}

	msg := cmd()
	result, ok := msg.(configErrorDialogResultMsg)
	if !ok {
		t.Fatalf("Expected configErrorDialogResultMsg, got %T", msg)
	}

	if result.choice != ConfigErrorChoiceQuit {
		t.Errorf("Expected ConfigErrorChoiceQuit, got %v", result.choice)
	}
}

func TestConfigErrorDialog_ReloadFix(t *testing.T) {
	d := NewConfigErrorDialogForReload("Error", "Details")

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	_, cmd := d.Update(keyMsg)

	if cmd == nil {
		t.Fatal("Expected command from 'f' key")
	}

	msg := cmd()
	result, ok := msg.(configErrorDialogResultMsg)
	if !ok {
		t.Fatalf("Expected configErrorDialogResultMsg, got %T", msg)
	}

	if result.choice != ConfigErrorChoiceFix {
		t.Errorf("Expected ConfigErrorChoiceFix, got %v", result.choice)
	}

	if result.isStartup {
		t.Error("Expected isStartup=false for reload dialog")
	}
}

func TestConfigErrorDialog_ReloadKeep(t *testing.T) {
	d := NewConfigErrorDialogForReload("Error", "Details")

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}}
	_, cmd := d.Update(keyMsg)

	if cmd == nil {
		t.Fatal("Expected command from 'k' key")
	}

	msg := cmd()
	result, ok := msg.(configErrorDialogResultMsg)
	if !ok {
		t.Fatalf("Expected configErrorDialogResultMsg, got %T", msg)
	}

	if result.choice != ConfigErrorChoiceKeep {
		t.Errorf("Expected ConfigErrorChoiceKeep, got %v", result.choice)
	}
}

func TestConfigErrorDialog_InactiveIgnoresInput(t *testing.T) {
	d := NewConfigErrorDialog("Error", "Details")
	d.Close()

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}
	_, cmd := d.Update(keyMsg)

	if cmd != nil {
		t.Error("Expected nil command when dialog is inactive")
	}
}

func TestConfigErrorDialog_StartupView(t *testing.T) {
	d := NewConfigErrorDialog("Syntax error at line 23", "Details here")
	view := d.View()

	if !strings.Contains(view, "Configuration Error") {
		t.Error("View should contain title")
	}
	if !strings.Contains(view, "Syntax error at line 23") {
		t.Error("View should contain error message")
	}
	if !strings.Contains(view, "[f]") {
		t.Error("View should contain fix option")
	}
	if !strings.Contains(view, "[q]") {
		t.Error("View should contain quit option")
	}
}

func TestConfigErrorDialog_ReloadView(t *testing.T) {
	d := NewConfigErrorDialogForReload("Error message", "Details")
	view := d.View()

	if !strings.Contains(view, "[f]") {
		t.Error("View should contain fix option")
	}
	if !strings.Contains(view, "[k]") {
		t.Error("View should contain keep option")
	}
	// Should NOT have quit option
	if strings.Contains(view, "[q] Quit") {
		t.Error("Reload view should not contain quit option")
	}
}
