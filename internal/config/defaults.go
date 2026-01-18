package config

// DefaultKeybindings returns the default keybindings map.
// All 32 actions are defined with their default key assignments.
func DefaultKeybindings() map[string][]string {
	return map[string][]string{
		// Navigation
		"move_down":  {"J", "Down"},
		"move_up":    {"K", "Up"},
		"move_left":  {"H", "Left"},
		"move_right": {"L", "Right"},
		"enter":      {"Enter"},
		"page_down":  {"Ctrl+D", "PageDown"},
		"page_up":    {"Ctrl+U", "PageUp"},

		// File operations
		"copy":             {"C"},
		"move":             {"M"},
		"delete":           {"D"},
		"rename":           {"R"},
		"rename_full_name": {"Shift+R"},
		"new_file":         {"N"},
		"new_directory":    {"Shift+N"},
		"mark":             {"Space"},

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
		"sql_filter":   {"Ctrl+G"},

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

		// Permission edit
		"permission": {"P", "Shift+P"},

		// Path navigation
		"path_jump": {"Ctrl+J"},
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
	}
}
