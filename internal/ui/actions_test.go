package ui

import (
	"testing"

	"github.com/sakura/duofm/internal/config"
)

func TestAction_String(t *testing.T) {
	tests := []struct {
		action   Action
		expected string
	}{
		{ActionNone, "none"},
		{ActionMoveDown, "move_down"},
		{ActionMoveUp, "move_up"},
		{ActionQuit, "quit"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.action.String(); got != tt.expected {
				t.Errorf("Action.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestKeybindingMap_GetAction(t *testing.T) {
	cfg := &config.Config{
		Keybindings: map[string][]string{
			"move_down": {"J", "Down"},
			"move_up":   {"K", "Up"},
			"quit":      {"Q"},
			"help":      {"?"},
		},
	}

	km := NewKeybindingMap(cfg)

	tests := []struct {
		key      string
		expected Action
	}{
		{"j", ActionMoveDown},
		{"down", ActionMoveDown},
		{"k", ActionMoveUp},
		{"up", ActionMoveUp},
		{"q", ActionQuit},
		{"?", ActionHelp},
		{"x", ActionNone}, // Unmapped key
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := km.GetAction(tt.key); got != tt.expected {
				t.Errorf("GetAction(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestKeybindingMap_HasKey(t *testing.T) {
	cfg := &config.Config{
		Keybindings: map[string][]string{
			"move_down": {"J"},
			"quit":      {"Q"},
		},
	}

	km := NewKeybindingMap(cfg)

	if !km.HasKey("j") {
		t.Error("HasKey(j) = false, want true")
	}
	if !km.HasKey("q") {
		t.Error("HasKey(q) = false, want true")
	}
	if km.HasKey("x") {
		t.Error("HasKey(x) = true, want false")
	}
}

func TestKeybindingMap_EmptyArray(t *testing.T) {
	cfg := &config.Config{
		Keybindings: map[string][]string{
			"help": {}, // Disabled
			"quit": {"Q"},
		},
	}

	km := NewKeybindingMap(cfg)

	// help action should not be mapped to any key
	if km.GetAction("?") != ActionNone {
		t.Error("Disabled action should not be mapped")
	}

	// quit should still work
	if km.GetAction("q") != ActionQuit {
		t.Error("Quit action should be mapped")
	}
}

func TestDefaultKeybindingMap(t *testing.T) {
	km := DefaultKeybindingMap()

	// Check some default mappings
	tests := []struct {
		key      string
		expected Action
	}{
		{"j", ActionMoveDown},
		{"down", ActionMoveDown},
		{"k", ActionMoveUp},
		{"q", ActionQuit},
		{"?", ActionHelp},
		{"c", ActionCopy},
		{"ctrl+h", ActionToggleHidden},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := km.GetAction(tt.key); got != tt.expected {
				t.Errorf("GetAction(%q) = %v, want %v", tt.key, got, tt.expected)
			}
		})
	}
}

func TestAction_String_Unknown(t *testing.T) {
	// Test with an unknown action value
	unknownAction := Action(9999)
	result := unknownAction.String()
	if result != "unknown" {
		t.Errorf("Unknown action String() = %q, want %q", result, "unknown")
	}
}

func TestActionFromName(t *testing.T) {
	t.Run("valid action names", func(t *testing.T) {
		tests := []struct {
			name     string
			expected Action
		}{
			{"move_down", ActionMoveDown},
			{"move_up", ActionMoveUp},
			{"quit", ActionQuit},
			{"help", ActionHelp},
			{"copy", ActionCopy},
			{"toggle_hidden", ActionToggleHidden},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := ActionFromName(tt.name)
				if got != tt.expected {
					t.Errorf("ActionFromName(%q) = %v, want %v", tt.name, got, tt.expected)
				}
			})
		}
	})

	t.Run("invalid action name", func(t *testing.T) {
		got := ActionFromName("invalid_action_name")
		if got != ActionNone {
			t.Errorf("ActionFromName(invalid) = %v, want %v", got, ActionNone)
		}
	})

	t.Run("empty action name", func(t *testing.T) {
		got := ActionFromName("")
		if got != ActionNone {
			t.Errorf("ActionFromName(\"\") = %v, want %v", got, ActionNone)
		}
	})
}

func TestKeybindingMap_NilConfig(t *testing.T) {
	km := NewKeybindingMap(nil)

	if km == nil {
		t.Fatal("NewKeybindingMap(nil) should return a non-nil map")
	}

	// Should return ActionNone for any key
	if km.GetAction("j") != ActionNone {
		t.Error("GetAction should return ActionNone for nil config")
	}

	if km.HasKey("j") {
		t.Error("HasKey should return false for nil config")
	}
}

func TestKeybindingMap_NilKeybindings(t *testing.T) {
	cfg := &config.Config{
		Keybindings: nil,
	}
	km := NewKeybindingMap(cfg)

	if km == nil {
		t.Fatal("NewKeybindingMap should return a non-nil map")
	}

	if km.GetAction("j") != ActionNone {
		t.Error("GetAction should return ActionNone for nil keybindings")
	}
}

func TestKeybindingMap_NilReceiver(t *testing.T) {
	var km *KeybindingMap = nil

	if km.GetAction("j") != ActionNone {
		t.Error("GetAction on nil receiver should return ActionNone")
	}

	if km.HasKey("j") {
		t.Error("HasKey on nil receiver should return false")
	}
}

func TestKeybindingMap_UnknownActionInConfig(t *testing.T) {
	cfg := &config.Config{
		Keybindings: map[string][]string{
			"unknown_action": {"x"},
			"quit":           {"q"},
		},
	}
	km := NewKeybindingMap(cfg)

	// Unknown action should be skipped
	if km.GetAction("x") != ActionNone {
		t.Error("Unknown action should not be mapped")
	}

	// Known action should still work
	if km.GetAction("q") != ActionQuit {
		t.Error("Known action should be mapped correctly")
	}
}
