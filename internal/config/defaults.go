package config

// DefaultKeybindings returns the default keybindings map.
// All 28 actions are defined with their default key assignments.
func DefaultKeybindings() map[string][]string {
	return map[string][]string{
		// Navigation
		"move_down":  {"J", "Down"},
		"move_up":    {"K", "Up"},
		"move_left":  {"H", "Left"},
		"move_right": {"L", "Right"},
		"enter":      {"Enter"},

		// File operations
		"copy":          {"C"},
		"move":          {"M"},
		"delete":        {"D"},
		"rename":        {"R"},
		"new_file":      {"N"},
		"new_directory": {"Shift+N"},
		"mark":          {"Space"},

		// Display
		"toggle_info":   {"I"},
		"toggle_hidden": {"Ctrl+H"},
		"sort":          {"S"},
		"help":          {"?"},

		// Navigation extended
		"home":            {"~"},
		"prev_dir":        {"-"},
		"history_back":    {"Alt+Left", "["},
		"history_forward": {"Alt+Right", "]"},
		"refresh":         {"F5", "Ctrl+R"},
		"sync_pane":       {"="},

		// Search
		"search":       {"/"},
		"regex_search": {"Ctrl+F"},

		// External applications
		"view":          {"V"},
		"edit":          {"E"},
		"shell_command": {"!"},
		"context_menu":  {"@"},

		// Application
		"quit":   {"Q"},
		"escape": {"Esc"},

		// Bookmarks
		"bookmark":     {"B"},
		"add_bookmark": {"Shift+B"},
	}
}

// AllActions returns the list of all valid action names.
func AllActions() []string {
	return []string{
		"move_down",
		"move_up",
		"move_left",
		"move_right",
		"enter",
		"copy",
		"move",
		"delete",
		"rename",
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
		"view",
		"edit",
		"shell_command",
		"context_menu",
		"quit",
		"escape",
		"bookmark",
		"add_bookmark",
	}
}
