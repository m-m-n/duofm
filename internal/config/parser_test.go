package config

import (
	"testing"
)

func TestNormalizeKey_SingleAlphabet(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"J", "j", false},
		{"K", "k", false},
		{"N", "n", false},
		{"a", "a", false},
		{"Z", "z", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := NormalizeKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizeKey(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeKey(%q) returned error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeKey_Symbols(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"?", "?", false},
		{"@", "@", false},
		{"!", "!", false},
		{"~", "~", false},
		{"/", "/", false},
		{"-", "-", false},
		{"=", "=", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := NormalizeKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizeKey(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeKey(%q) returned error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeKey_SpecialKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"Enter", "enter", false},
		{"Esc", "esc", false},
		{"Space", " ", false},
		{"Tab", "tab", false},
		{"Backspace", "backspace", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := NormalizeKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizeKey(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeKey(%q) returned error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeKey_ArrowKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"Up", "up", false},
		{"Down", "down", false},
		{"Left", "left", false},
		{"Right", "right", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := NormalizeKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizeKey(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeKey(%q) returned error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeKey_FunctionKeys(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"F1", "f1", false},
		{"F5", "f5", false},
		{"F12", "f12", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := NormalizeKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizeKey(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeKey(%q) returned error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeKey_CtrlModifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"Ctrl+H", "ctrl+h", false},
		{"Ctrl+R", "ctrl+r", false},
		{"Ctrl+F", "ctrl+f", false},
		{"Ctrl+=", "ctrl+=", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := NormalizeKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizeKey(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeKey(%q) returned error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeKey_ShiftModifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		// Shift+N produces uppercase N in bubble tea
		{"Shift+N", "N", false},
		{"Shift+A", "A", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := NormalizeKey(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NormalizeKey(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("NormalizeKey(%q) returned error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("NormalizeKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeKey_InvalidFormat(t *testing.T) {
	tests := []string{
		"",
		"Ctrl++",
		"Invalid",
		"Ctrl+Shift+",
		"++",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := NormalizeKey(input)
			if err == nil {
				t.Errorf("NormalizeKey(%q) expected error, got nil", input)
			}
		})
	}
}

func TestValidateAction_Valid(t *testing.T) {
	validActions := AllActions()

	for _, action := range validActions {
		t.Run(action, func(t *testing.T) {
			if !ValidateAction(action) {
				t.Errorf("ValidateAction(%q) = false, want true", action)
			}
		})
	}
}

func TestValidateAction_Invalid(t *testing.T) {
	invalidActions := []string{
		"invalid_action",
		"",
		"MOVE_DOWN",
		"moveDown",
	}

	for _, action := range invalidActions {
		t.Run(action, func(t *testing.T) {
			if ValidateAction(action) {
				t.Errorf("ValidateAction(%q) = true, want false", action)
			}
		})
	}
}

func TestValidateKeybindings_DuplicateKey(t *testing.T) {
	cfg := &Config{
		Keybindings: map[string][]string{
			"move_down":   {"J", "Down"},
			"custom_move": {"J"}, // Duplicate "J"
		},
	}

	warnings := ValidateKeybindings(cfg)

	if len(warnings) == 0 {
		t.Error("ValidateKeybindings() returned no warnings for duplicate key")
	}

	// Check that warning mentions both actions and the duplicate key
	found := false
	for _, w := range warnings {
		if containsHelper(w, "J") && containsHelper(w, "move_down") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Warning does not mention duplicate key: %v", warnings)
	}
}

func TestValidateKeybindings_NoDuplicates(t *testing.T) {
	cfg := &Config{
		Keybindings: map[string][]string{
			"move_down": {"J", "Down"},
			"move_up":   {"K", "Up"},
			"quit":      {"Q"},
		},
	}

	warnings := ValidateKeybindings(cfg)

	if len(warnings) != 0 {
		t.Errorf("ValidateKeybindings() returned warnings for valid config: %v", warnings)
	}
}
