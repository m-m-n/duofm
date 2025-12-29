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

# Color Theme Configuration
# Colors are specified as ANSI 256-color codes (0-255)
# Use ? key in duofm to see the color palette reference
[colors]
# Pane - Cursor (current row)
# cursor_fg = 15              # white
# cursor_bg = 39              # blue (active pane)
# cursor_bg_inactive = 240    # gray (inactive pane)

# Pane - Marked rows
# mark_fg = 0                 # black (active pane)
# mark_fg_inactive = 15       # white (inactive pane)
# mark_bg = 136               # dark yellow (active pane)
# mark_bg_inactive = 94       # darker yellow (inactive pane)

# Pane - Cursor + Marked rows
# cursor_mark_fg = 15         # white
# cursor_mark_bg = 30         # cyan (active pane)
# cursor_mark_bg_inactive = 23 # dark cyan (inactive pane)

# Pane - Path display
# path_fg = 39                # blue (active pane)
# path_fg_inactive = 240      # gray (inactive pane)

# Pane - Header (column names)
# header_fg = 245             # light gray (active pane)
# header_fg_inactive = 240    # gray (inactive pane)

# Pane - Structure
# border_fg = 240             # gray
# dimmed_bg = 236             # dark gray
# dimmed_fg = 243             # medium gray

# File Types
# directory_fg = 39           # blue
# symlink_fg = 14             # cyan
# executable_fg = 9           # red

# Dialog
# dialog_title_fg = 39        # blue
# dialog_border_fg = 39       # blue
# dialog_selected_fg = 0      # black
# dialog_selected_bg = 39     # blue
# dialog_footer_fg = 240      # gray

# Input Fields
# input_fg = 15               # white
# input_bg = 236              # dark gray
# input_border_fg = 240       # gray

# Minibuffer
# minibuffer_fg = 15          # white
# minibuffer_bg = 236         # dark gray

# Error and Warning
# error_fg = 196              # red
# error_border_fg = 196       # red
# warning_fg = 240            # gray

# Status Bar
# status_fg = 15              # white
# status_bg = 240             # gray
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
