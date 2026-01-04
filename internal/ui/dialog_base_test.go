package ui

import (
	"testing"
)

func TestNewBaseDialog(t *testing.T) {
	t.Run("creates with DialogDisplayPane", func(t *testing.T) {
		base := NewBaseDialog(DialogDisplayPane)

		if !base.IsActive() {
			t.Error("BaseDialog should be active by default")
		}
		if base.Width() != 50 {
			t.Errorf("Width() = %d, want 50", base.Width())
		}
		if base.DisplayType() != DialogDisplayPane {
			t.Errorf("DisplayType() = %v, want DialogDisplayPane", base.DisplayType())
		}
	})

	t.Run("creates with DialogDisplayScreen", func(t *testing.T) {
		base := NewBaseDialog(DialogDisplayScreen)

		if base.DisplayType() != DialogDisplayScreen {
			t.Errorf("DisplayType() = %v, want DialogDisplayScreen", base.DisplayType())
		}
	})
}

func TestBaseDialog_SetActive(t *testing.T) {
	base := NewBaseDialog(DialogDisplayPane)

	base.SetActive(false)
	if base.IsActive() {
		t.Error("SetActive(false) should deactivate dialog")
	}

	base.SetActive(true)
	if !base.IsActive() {
		t.Error("SetActive(true) should activate dialog")
	}
}

func TestBaseDialog_Close(t *testing.T) {
	base := NewBaseDialog(DialogDisplayPane)

	base.Close()
	if base.IsActive() {
		t.Error("Close() should deactivate dialog")
	}
}

func TestBaseDialog_SetWidth(t *testing.T) {
	base := NewBaseDialog(DialogDisplayPane)

	base.SetWidth(80)
	if base.Width() != 80 {
		t.Errorf("Width() = %d, want 80", base.Width())
	}
}

func TestNewDialogStyles(t *testing.T) {
	styles := NewDialogStyles(60, ColorPrimary)

	// Check that styles are not nil/empty (basic sanity check)
	if styles.Title.GetWidth() != 56 { // 60 - 4
		t.Errorf("Title width = %d, want 56", styles.Title.GetWidth())
	}
	if styles.Box.GetWidth() != 60 {
		t.Errorf("Box width = %d, want 60", styles.Box.GetWidth())
	}
}

func TestDefaultDialogStyles(t *testing.T) {
	styles := DefaultDialogStyles(50)

	if styles.Box.GetWidth() != 50 {
		t.Errorf("Box width = %d, want 50", styles.Box.GetWidth())
	}
}

func TestErrorDialogStyles(t *testing.T) {
	styles := ErrorDialogStyles(50)

	if styles.Box.GetWidth() != 50 {
		t.Errorf("Box width = %d, want 50", styles.Box.GetWidth())
	}
	// Error styles should have danger color for title and border
}

func TestDialogColors(t *testing.T) {
	tests := []struct {
		color DialogColor
		want  string
	}{
		{ColorPrimary, "39"},
		{ColorDanger, "196"},
		{ColorMuted, "240"},
		{ColorHighlight, "15"},
		{ColorInputBg, "236"},
		{ColorBorder, "62"},
	}

	for _, tt := range tests {
		if string(tt.color) != tt.want {
			t.Errorf("Color %v = %q, want %q", tt.color, string(tt.color), tt.want)
		}
	}
}
