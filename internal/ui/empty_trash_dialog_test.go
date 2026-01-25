package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewEmptyTrashDialog(t *testing.T) {
	d := NewEmptyTrashDialog(5)

	if d == nil {
		t.Fatal("NewEmptyTrashDialog() returned nil")
	}

	if d.itemCount != 5 {
		t.Errorf("itemCount = %d, want %d", d.itemCount, 5)
	}

	if !d.IsActive() {
		t.Error("dialog should be active after creation")
	}
}

func TestEmptyTrashDialogUpdate(t *testing.T) {
	tests := []struct {
		name              string
		key               string
		expectedConfirmed bool
	}{
		{
			name:              "confirm with y",
			key:               "y",
			expectedConfirmed: true,
		},
		{
			name:              "confirm with Y",
			key:               "Y",
			expectedConfirmed: true,
		},
		{
			name:              "cancel with n",
			key:               "n",
			expectedConfirmed: false,
		},
		{
			name:              "cancel with N",
			key:               "N",
			expectedConfirmed: false,
		},
		{
			name:              "cancel with esc",
			key:               "esc",
			expectedConfirmed: false,
		},
		{
			name:              "cancel with ctrl+c",
			key:               "ctrl+c",
			expectedConfirmed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewEmptyTrashDialog(5)

			keyMsg := tea.KeyMsg{Type: tea.KeyRunes}
			switch tt.key {
			case "esc":
				keyMsg = tea.KeyMsg{Type: tea.KeyEsc}
			case "ctrl+c":
				keyMsg = tea.KeyMsg{Type: tea.KeyCtrlC}
			default:
				keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			}

			_, cmd := d.Update(keyMsg)

			if cmd == nil {
				t.Fatal("Update() returned nil cmd")
			}

			// Execute the command
			msg := cmd()
			result, ok := msg.(emptyTrashResultMsg)
			if !ok {
				t.Fatalf("expected emptyTrashResultMsg, got %T", msg)
			}

			if result.confirmed != tt.expectedConfirmed {
				t.Errorf("confirmed = %v, want %v", result.confirmed, tt.expectedConfirmed)
			}

			if d.IsActive() {
				t.Error("dialog should be closed after choice")
			}
		})
	}
}

func TestEmptyTrashDialogUpdateIgnoresUnrelatedKeys(t *testing.T) {
	d := NewEmptyTrashDialog(5)

	// Send an unrelated key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	_, cmd := d.Update(keyMsg)

	if cmd != nil {
		t.Error("Update() should return nil cmd for unrelated keys")
	}

	if !d.IsActive() {
		t.Error("dialog should remain active after unrelated key")
	}
}

func TestEmptyTrashDialogUpdateInactive(t *testing.T) {
	d := NewEmptyTrashDialog(5)
	d.Close()

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	_, cmd := d.Update(keyMsg)

	if cmd != nil {
		t.Error("Update() should return nil cmd when dialog is inactive")
	}
}

func TestEmptyTrashDialogView(t *testing.T) {
	d := NewEmptyTrashDialog(5)

	view := d.View()

	// Check that the view contains expected elements
	if !strings.Contains(view, "Empty Trash") {
		t.Error("View should contain 'Empty Trash'")
	}

	if !strings.Contains(view, "5 item(s)") {
		t.Error("View should contain item count")
	}

	if !strings.Contains(view, "cannot be undone") {
		t.Error("View should contain warning about irreversibility")
	}

	if !strings.Contains(view, "[Y]es") {
		t.Error("View should contain '[Y]es' option")
	}

	if !strings.Contains(view, "[N]o") {
		t.Error("View should contain '[N]o' option")
	}
}

func TestEmptyTrashDialogViewWithDifferentCounts(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{1, "1 item(s)"},
		{10, "10 item(s)"},
		{100, "100 item(s)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			d := NewEmptyTrashDialog(tt.count)
			view := d.View()

			if !strings.Contains(view, tt.expected) {
				t.Errorf("View should contain %q", tt.expected)
			}
		})
	}
}

func TestEmptyTrashDialogViewInactive(t *testing.T) {
	d := NewEmptyTrashDialog(5)
	d.Close()

	view := d.View()

	if view != "" {
		t.Error("View() should return empty string when dialog is inactive")
	}
}
