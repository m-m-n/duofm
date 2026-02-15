package config

import (
	"testing"
)

func TestDefaultKeybindings(t *testing.T) {
	defaults := DefaultKeybindings()

	if defaults == nil {
		t.Fatal("DefaultKeybindings() returned nil")
	}

	// Verify number of actions
	expectedActions := 45 // Includes sql_filter, shell_log, trash, open_trash, restore, empty_trash, goto_top, goto_bottom
	if len(defaults) != expectedActions {
		t.Errorf("DefaultKeybindings() length = %d, want %d", len(defaults), expectedActions)
	}

	// Test navigation keybindings
	navigationTests := []struct {
		action   string
		expected []string
	}{
		{"move_down", []string{"J", "Down"}},
		{"move_up", []string{"K", "Up"}},
		{"move_left", []string{"H", "Left"}},
		{"move_right", []string{"L", "Right"}},
		{"enter", []string{"Enter"}},
		{"page_down", []string{"Ctrl+D", "PageDown"}},
		{"page_up", []string{"Ctrl+U", "PageUp"}},
	}

	for _, tt := range navigationTests {
		t.Run("navigation_"+tt.action, func(t *testing.T) {
			keys, ok := defaults[tt.action]
			if !ok {
				t.Errorf("DefaultKeybindings() missing action %s", tt.action)
				return
			}

			if len(keys) != len(tt.expected) {
				t.Errorf("DefaultKeybindings()[%s] length = %d, want %d", tt.action, len(keys), len(tt.expected))
				return
			}

			for i, key := range keys {
				if key != tt.expected[i] {
					t.Errorf("DefaultKeybindings()[%s][%d] = %s, want %s", tt.action, i, key, tt.expected[i])
				}
			}
		})
	}

	// Test file operations keybindings
	fileOpTests := []struct {
		action   string
		expected []string
	}{
		{"copy", []string{"C"}},
		{"move", []string{"M"}},
		{"delete", []string{"D"}},
		{"rename", []string{"R"}},
		{"new_file", []string{"N"}},
		{"new_directory", []string{"Shift+N"}},
		{"mark", []string{"Space"}},
	}

	for _, tt := range fileOpTests {
		t.Run("file_op_"+tt.action, func(t *testing.T) {
			keys, ok := defaults[tt.action]
			if !ok {
				t.Errorf("DefaultKeybindings() missing action %s", tt.action)
				return
			}

			if len(keys) != len(tt.expected) {
				t.Errorf("DefaultKeybindings()[%s] length = %d, want %d", tt.action, len(keys), len(tt.expected))
			}
		})
	}

	// Test display keybindings
	displayTests := []struct {
		action   string
		expected []string
	}{
		{"toggle_info", []string{"I"}},
		{"toggle_hidden", []string{"Ctrl+H"}},
		{"sort", []string{"S"}},
		{"help", []string{"?"}},
	}

	for _, tt := range displayTests {
		t.Run("display_"+tt.action, func(t *testing.T) {
			keys, ok := defaults[tt.action]
			if !ok {
				t.Errorf("DefaultKeybindings() missing action %s", tt.action)
				return
			}

			if len(keys) != len(tt.expected) {
				t.Errorf("DefaultKeybindings()[%s] length = %d, want %d", tt.action, len(keys), len(tt.expected))
			}
		})
	}

	// Test extended navigation keybindings
	extNavTests := []struct {
		action   string
		expected []string
	}{
		{"home", []string{"~"}},
		{"prev_dir", []string{"-"}},
		{"history_back", []string{"Alt+Left", "["}},
		{"history_forward", []string{"Alt+Right", "]"}},
		{"refresh", []string{"F5", "Ctrl+R"}},
		{"sync_pane", []string{"="}},
	}

	for _, tt := range extNavTests {
		t.Run("ext_nav_"+tt.action, func(t *testing.T) {
			keys, ok := defaults[tt.action]
			if !ok {
				t.Errorf("DefaultKeybindings() missing action %s", tt.action)
				return
			}

			if len(keys) != len(tt.expected) {
				t.Errorf("DefaultKeybindings()[%s] length = %d, want %d", tt.action, len(keys), len(tt.expected))
			}
		})
	}

	// Test search keybindings
	searchTests := []struct {
		action   string
		expected []string
	}{
		{"search", []string{"/"}},
		{"regex_search", []string{"Ctrl+F"}},
	}

	for _, tt := range searchTests {
		t.Run("search_"+tt.action, func(t *testing.T) {
			keys, ok := defaults[tt.action]
			if !ok {
				t.Errorf("DefaultKeybindings() missing action %s", tt.action)
				return
			}

			if len(keys) != len(tt.expected) {
				t.Errorf("DefaultKeybindings()[%s] length = %d, want %d", tt.action, len(keys), len(tt.expected))
			}
		})
	}

	// Test external application keybindings
	extAppTests := []struct {
		action   string
		expected []string
	}{
		{"view", []string{"V"}},
		{"edit", []string{"E"}},
		{"shell_command", []string{"!"}},
		{"context_menu", []string{"@"}},
	}

	for _, tt := range extAppTests {
		t.Run("ext_app_"+tt.action, func(t *testing.T) {
			keys, ok := defaults[tt.action]
			if !ok {
				t.Errorf("DefaultKeybindings() missing action %s", tt.action)
				return
			}

			if len(keys) != len(tt.expected) {
				t.Errorf("DefaultKeybindings()[%s] length = %d, want %d", tt.action, len(keys), len(tt.expected))
			}
		})
	}

	// Test application keybindings
	appTests := []struct {
		action   string
		expected []string
	}{
		{"quit", []string{"Q"}},
		{"escape", []string{"Esc"}},
	}

	for _, tt := range appTests {
		t.Run("app_"+tt.action, func(t *testing.T) {
			keys, ok := defaults[tt.action]
			if !ok {
				t.Errorf("DefaultKeybindings() missing action %s", tt.action)
				return
			}

			if len(keys) != len(tt.expected) {
				t.Errorf("DefaultKeybindings()[%s] length = %d, want %d", tt.action, len(keys), len(tt.expected))
			}
		})
	}

	// Test bookmark keybindings
	bookmarkTests := []struct {
		action   string
		expected []string
	}{
		{"bookmark", []string{"B"}},
		{"add_bookmark", []string{"Shift+B"}},
	}

	for _, tt := range bookmarkTests {
		t.Run("bookmark_"+tt.action, func(t *testing.T) {
			keys, ok := defaults[tt.action]
			if !ok {
				t.Errorf("DefaultKeybindings() missing action %s", tt.action)
				return
			}

			if len(keys) != len(tt.expected) {
				t.Errorf("DefaultKeybindings()[%s] length = %d, want %d", tt.action, len(keys), len(tt.expected))
			}
		})
	}

	// Test permission keybinding
	t.Run("permission", func(t *testing.T) {
		keys, ok := defaults["permission"]
		if !ok {
			t.Error("DefaultKeybindings() missing action permission")
			return
		}

		expected := []string{"P", "Shift+P"}
		if len(keys) != len(expected) {
			t.Errorf("DefaultKeybindings()[permission] length = %d, want %d", len(keys), len(expected))
		}
	})

	// Test path_jump keybinding
	t.Run("path_jump", func(t *testing.T) {
		keys, ok := defaults["path_jump"]
		if !ok {
			t.Error("DefaultKeybindings() missing action path_jump")
			return
		}

		expected := []string{"Ctrl+J"}
		if len(keys) != len(expected) {
			t.Errorf("DefaultKeybindings()[path_jump] length = %d, want %d", len(keys), len(expected))
		}
	})
}

func TestAllActions(t *testing.T) {
	actions := AllActions()

	if actions == nil {
		t.Fatal("AllActions() returned nil")
	}

	// Verify number of actions
	expectedCount := 45 // Includes sql_filter, shell_log, trash, open_trash, restore, empty_trash, goto_top, goto_bottom
	if len(actions) != expectedCount {
		t.Errorf("AllActions() length = %d, want %d", len(actions), expectedCount)
	}

	// Verify all expected actions are present
	expectedActions := []string{
		"move_down",
		"move_up",
		"move_left",
		"move_right",
		"enter",
		"page_down",
		"page_up",
		"copy",
		"move",
		"delete",
		"rename",
		"rename_full_name",
		"new_file",
		"new_directory",
		"mark",
		"toggle_info",
		"toggle_hidden",
		"sort",
		"help",
		"home",
		"prev_dir",
		"history_back",
		"history_forward",
		"refresh",
		"sync_pane",
		"search",
		"regex_search",
		"sql_filter",
		"view",
		"edit",
		"shell_command",
		"context_menu",
		"quit",
		"escape",
		"bookmark",
		"add_bookmark",
		"permission",
		"path_jump",
		"trash",
		"open_trash",
		"restore",
		"empty_trash",
		"goto_top",
		"goto_bottom",
	}

	actionSet := make(map[string]bool)
	for _, action := range actions {
		actionSet[action] = true
	}

	for _, expected := range expectedActions {
		if !actionSet[expected] {
			t.Errorf("AllActions() missing action %s", expected)
		}
	}
}

func TestAllActions_MatchDefaultKeybindings(t *testing.T) {
	actions := AllActions()
	defaults := DefaultKeybindings()

	// Every action should have a corresponding entry in DefaultKeybindings
	for _, action := range actions {
		if _, ok := defaults[action]; !ok {
			t.Errorf("Action %s exists in AllActions() but not in DefaultKeybindings()", action)
		}
	}

	// Every key in DefaultKeybindings should be in AllActions
	actionSet := make(map[string]bool)
	for _, action := range actions {
		actionSet[action] = true
	}

	for action := range defaults {
		if !actionSet[action] {
			t.Errorf("Action %s exists in DefaultKeybindings() but not in AllActions()", action)
		}
	}
}

func TestDefaultKeybindings_NonEmpty(t *testing.T) {
	defaults := DefaultKeybindings()

	// Actions that intentionally have no default keybinding
	// (e.g., context-dependent actions handled by other keys)
	emptyAllowed := map[string]bool{
		"restore": true, // R key is context-dependent: restore in trash, rename outside
	}

	for action, keys := range defaults {
		if len(keys) == 0 && !emptyAllowed[action] {
			t.Errorf("DefaultKeybindings()[%s] is empty, but all default actions should have at least one key", action)
		}

		for i, key := range keys {
			if key == "" {
				t.Errorf("DefaultKeybindings()[%s][%d] is empty string", action, i)
			}
		}
	}
}
