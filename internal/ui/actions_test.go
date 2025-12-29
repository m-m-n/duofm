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
