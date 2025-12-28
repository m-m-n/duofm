package ui

import (
	"strings"
	"testing"
)

func TestHelpDialogContainsShellCommand(t *testing.T) {
	dialog := NewHelpDialog()
	view := dialog.View()

	// Check that the help dialog contains the shell command key binding
	if !strings.Contains(view, "!") {
		t.Error("Help dialog should contain '!' key binding for shell command")
	}
}
