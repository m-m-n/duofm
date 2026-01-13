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
		{ActionPageDown, "page_down"},
		{ActionPageUp, "page_up"},
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
			{"page_down", ActionPageDown},
			{"page_up", ActionPageUp},
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

func TestActionPathJump(t *testing.T) {
	t.Run("ActionFromName returns ActionPathJump", func(t *testing.T) {
		got := ActionFromName("path_jump")
		if got != ActionPathJump {
			t.Errorf("ActionFromName(\"path_jump\") = %v, want %v", got, ActionPathJump)
		}
	})

	t.Run("ActionPathJump.String returns path_jump", func(t *testing.T) {
		got := ActionPathJump.String()
		if got != "path_jump" {
			t.Errorf("ActionPathJump.String() = %q, want %q", got, "path_jump")
		}
	})

	t.Run("DefaultKeybindingMap maps Ctrl+J to ActionPathJump", func(t *testing.T) {
		km := DefaultKeybindingMap()
		got := km.GetAction("ctrl+j")
		if got != ActionPathJump {
			t.Errorf("GetAction(\"ctrl+j\") = %v, want %v", got, ActionPathJump)
		}
	})
}

func TestActionRenameFullName(t *testing.T) {
	t.Run("ActionFromName returns ActionRenameFullName", func(t *testing.T) {
		got := ActionFromName("rename_full_name")
		if got != ActionRenameFullName {
			t.Errorf("ActionFromName(\"rename_full_name\") = %v, want %v", got, ActionRenameFullName)
		}
	})

	t.Run("ActionRenameFullName.String returns rename_full_name", func(t *testing.T) {
		got := ActionRenameFullName.String()
		if got != "rename_full_name" {
			t.Errorf("ActionRenameFullName.String() = %q, want %q", got, "rename_full_name")
		}
	})

	t.Run("DefaultKeybindingMap maps Shift+R to ActionRenameFullName", func(t *testing.T) {
		km := DefaultKeybindingMap()
		// Shift+R is normalized to uppercase "R"
		got := km.GetAction("R")
		if got != ActionRenameFullName {
			t.Errorf("GetAction(\"R\") = %v, want %v", got, ActionRenameFullName)
		}
	})
}

func TestPageScrollActions(t *testing.T) {
	t.Run("page_down and page_up in DefaultKeybindingMap", func(t *testing.T) {
		km := DefaultKeybindingMap()

		tests := []struct {
			key      string
			expected Action
		}{
			{"ctrl+d", ActionPageDown},
			{"pgdown", ActionPageDown},
			{"ctrl+u", ActionPageUp},
			{"pgup", ActionPageUp},
		}

		for _, tt := range tests {
			t.Run(tt.key, func(t *testing.T) {
				got := km.GetAction(tt.key)
				if got != tt.expected {
					t.Errorf("GetAction(%q) = %v, want %v", tt.key, got, tt.expected)
				}
			})
		}
	})

	t.Run("page_down and page_up action conversion", func(t *testing.T) {
		if ActionPageDown.String() != "page_down" {
			t.Errorf("ActionPageDown.String() = %q, want %q", ActionPageDown.String(), "page_down")
		}
		if ActionPageUp.String() != "page_up" {
			t.Errorf("ActionPageUp.String() = %q, want %q", ActionPageUp.String(), "page_up")
		}

		if ActionFromName("page_down") != ActionPageDown {
			t.Errorf("ActionFromName(page_down) = %v, want %v", ActionFromName("page_down"), ActionPageDown)
		}
		if ActionFromName("page_up") != ActionPageUp {
			t.Errorf("ActionFromName(page_up) = %v, want %v", ActionFromName("page_up"), ActionPageUp)
		}
	})
}
