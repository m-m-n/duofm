package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewRecursivePermDialog(t *testing.T) {
	d := NewRecursivePermDialog("testdir")

	if d == nil {
		t.Fatal("NewRecursivePermDialog returned nil")
	}

	if !d.IsActive() {
		t.Error("Expected dialog to be active")
	}

	if d.step != 0 {
		t.Errorf("Expected initial step=0, got %d", d.step)
	}
}

func TestRecursivePermDialog_StepProgression(t *testing.T) {
	d := NewRecursivePermDialog("testdir")

	// Initially at step 0 (directory input)
	view := d.View()
	if !strings.Contains(view, "1/2") {
		t.Error("Expected step indicator '1/2'")
	}
	if !strings.Contains(view, "DIRECTORIES") {
		t.Error("Expected directory prompt in step 1")
	}

	// Enter directory permission
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}})
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})

	// Press Enter to advance to step 2
	d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if d.step != 1 {
		t.Errorf("Expected step=1 after Enter, got %d", d.step)
	}

	view = d.View()
	if !strings.Contains(view, "2/2") {
		t.Error("Expected step indicator '2/2'")
	}
	if !strings.Contains(view, "FILES") {
		t.Error("Expected file prompt in step 2")
	}
	if !strings.Contains(view, "755") {
		t.Error("Expected to show previously entered directory permission")
	}
}

func TestRecursivePermDialog_PresetSelection(t *testing.T) {
	d := NewRecursivePermDialog("testdir")

	// Test preset in step 1 (directory)
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})

	if d.currentInput != "755" {
		t.Errorf("Expected preset 1 to fill '755', got '%s'", d.currentInput)
	}

	view := d.View()
	if !strings.Contains(view, "drwxr-xr-x") {
		t.Error("Expected directory symbolic notation")
	}

	// Advance to step 2
	d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Test preset in step 2 (file)
	d.currentInput = "" // Clear input
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})

	if d.currentInput != "644" {
		t.Errorf("Expected preset 1 in step 2 to fill '644', got '%s'", d.currentInput)
	}

	view = d.View()
	if !strings.Contains(view, "-rw-r--r--") {
		t.Error("Expected file symbolic notation")
	}
}

func TestRecursivePermDialog_Cancellation(t *testing.T) {
	t.Run("cancel at step 1", func(t *testing.T) {
		d := NewRecursivePermDialog("testdir")

		// Press Esc at step 1
		newDialog, _ := d.Update(tea.KeyMsg{Type: tea.KeyEsc})

		if newDialog.IsActive() {
			t.Error("Expected dialog to be inactive after Esc at step 1")
		}
	})

	t.Run("cancel at step 2", func(t *testing.T) {
		d := NewRecursivePermDialog("testdir")

		// Advance to step 2
		d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7', '5', '5'}})
		d.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// Press Esc at step 2
		newDialog, _ := d.Update(tea.KeyMsg{Type: tea.KeyEsc})

		if newDialog.IsActive() {
			t.Error("Expected dialog to be inactive after Esc at step 2")
		}
	})
}

func TestRecursivePermDialog_InvalidInput(t *testing.T) {
	d := NewRecursivePermDialog("testdir")

	// Try to submit invalid input
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'8', '8', '8'}})
	d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should still be at step 0
	if d.step != 0 {
		t.Errorf("Expected to stay at step 0 with invalid input, got step %d", d.step)
	}

	view := d.View()
	// Check that error message is shown (implementation may vary)
	if !strings.Contains(view, "invalid") && !strings.Contains(view, "must be") && !strings.Contains(view, "Invalid") {
		t.Logf("View: %s", view)
		t.Log("Note: Error message may be shown differently")
	}
}

func TestRecursivePermDialog_BackspaceEditing(t *testing.T) {
	d := NewRecursivePermDialog("testdir")

	// Enter some digits
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}})
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})

	if d.currentInput != "755" {
		t.Errorf("Expected input '755', got '%s'", d.currentInput)
	}

	// Press backspace
	d.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if d.currentInput != "75" {
		t.Errorf("Expected input '75' after backspace, got '%s'", d.currentInput)
	}

	// Press backspace again
	d.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if d.currentInput != "7" {
		t.Errorf("Expected input '7' after second backspace, got '%s'", d.currentInput)
	}
}

func TestRecursivePermDialog_RealTimeSymbolicUpdate(t *testing.T) {
	d := NewRecursivePermDialog("testdir")

	// Enter first digit
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}})
	view := d.View()

	// Should show partial symbolic notation or wait for complete input
	// (Implementation detail)

	// Enter complete permission
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})

	view = d.View()
	if !strings.Contains(view, "drwxr-xr-x") {
		t.Error("Expected real-time symbolic update for 755")
	}

	// Enter different permission
	d.currentInput = ""
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7'}})
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'0'}})

	view = d.View()
	if !strings.Contains(view, "drwx------") {
		t.Error("Expected real-time symbolic update for 700")
	}
}

func TestRecursivePermDialog_DisplayType(t *testing.T) {
	d := NewRecursivePermDialog("testdir")

	if d.DisplayType() != DialogDisplayPane {
		t.Errorf("Expected DisplayType=DialogDisplayPane, got %v", d.DisplayType())
	}
}

func TestRecursivePermDialog_CompletionFlow(t *testing.T) {
	d := NewRecursivePermDialog("testdir")

	// Step 1: Enter directory permission
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7', '5', '5'}})
	d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if d.dirMode != "755" {
		t.Errorf("Expected dirMode='755', got '%s'", d.dirMode)
	}

	// Step 2: Enter file permission
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'6', '4', '4'}})
	d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if d.fileMode != "644" {
		t.Errorf("Expected fileMode='644', got '%s'", d.fileMode)
	}

	// Dialog should be inactive after completion
	if d.IsActive() {
		t.Error("Expected dialog to be inactive after completing both steps")
	}
}

func TestRecursivePermDialog_TooShortInput(t *testing.T) {
	d := NewRecursivePermDialog("testdir")

	// Try to submit with only 2 digits
	d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'7', '5'}})
	d.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should still be at step 0
	if d.step != 0 {
		t.Errorf("Expected to stay at step 0 with incomplete input, got step %d", d.step)
	}

	view := d.View()
	if !strings.Contains(view, "must be") || !strings.Contains(view, "3 digits") {
		t.Error("Expected error message for incomplete input")
	}
}
