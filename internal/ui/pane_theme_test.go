package ui

import (
	"testing"
)

func TestPaneSetTheme(t *testing.T) {
	tmpDir := t.TempDir()

	pane, err := NewPane(LeftPane, tmpDir, 60, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane failed: %v", err)
	}

	originalTheme := pane.theme
	newTheme := DefaultTheme()
	newTheme.CursorFg = "255"

	pane.SetTheme(newTheme)

	if pane.theme == originalTheme {
		t.Error("SetTheme should update the theme")
	}
	if pane.theme != newTheme {
		t.Error("SetTheme should set the new theme pointer")
	}
}

func TestPaneSetTheme_NilKeepsExisting(t *testing.T) {
	tmpDir := t.TempDir()

	pane, err := NewPane(LeftPane, tmpDir, 60, 20, true, nil)
	if err != nil {
		t.Fatalf("NewPane failed: %v", err)
	}

	originalTheme := pane.theme

	pane.SetTheme(nil)

	if pane.theme != originalTheme {
		t.Error("SetTheme(nil) should keep existing theme")
	}
}
