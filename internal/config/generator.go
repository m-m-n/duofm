package config

import (
	"os"
	"path/filepath"
)

// defaultConfigTemplate is the template for the default configuration file.
// It includes comments explaining each section and action.
const defaultConfigTemplate = `# duofm configuration file
# Customize keybindings by modifying the values below
# Key format: Uppercase letters, symbols as-is, PascalCase for special keys
# Example: "J", "?", "Enter", "Ctrl+H"

[keybindings]

# Navigation
move_down = ["J", "Down"]
move_up = ["K", "Up"]
move_left = ["H", "Left"]
move_right = ["L", "Right"]
enter = ["Enter"]

# File operations
copy = ["C"]
move = ["M"]
delete = ["D"]
rename = ["R"]
new_file = ["N"]
new_directory = ["Shift+N"]
mark = ["Space"]

# Display
toggle_info = ["I"]
toggle_hidden = ["Ctrl+H"]
sort = ["S"]
help = ["?"]

# Navigation extended
home = ["~"]
prev_dir = ["-"]
refresh = ["F5", "Ctrl+R"]
sync_pane = ["="]

# Search
search = ["/"]
regex_search = ["Ctrl+F"]

# External applications
view = ["V"]
edit = ["E"]
shell_command = ["!"]
context_menu = ["@"]

# Application
quit = ["Q"]
escape = ["Esc"]
`

// GenerateDefaultConfig generates a default configuration file at the specified path.
// It creates the parent directory if it does not exist.
func GenerateDefaultConfig(path string) error {
	// Create parent directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write the default config
	return os.WriteFile(path, []byte(defaultConfigTemplate), 0644)
}
