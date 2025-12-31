package ui

// Action represents a user action that can be triggered by a keybinding.
type Action int

// Action constants for all 28 actions plus ActionNone.
const (
	ActionNone Action = iota
	// Navigation
	ActionMoveDown
	ActionMoveUp
	ActionMoveLeft
	ActionMoveRight
	ActionEnter
	// File operations
	ActionCopy
	ActionMove
	ActionDelete
	ActionRename
	ActionNewFile
	ActionNewDirectory
	ActionMark
	// Display
	ActionToggleInfo
	ActionToggleHidden
	ActionSort
	ActionHelp
	// Navigation extended
	ActionHome
	ActionPrevDir
	ActionRefresh
	ActionSyncPane
	// Search
	ActionSearch
	ActionRegexSearch
	// External applications
	ActionView
	ActionEdit
	ActionShellCommand
	ActionContextMenu
	// Application
	ActionQuit
	ActionEscape
	// Bookmarks
	ActionBookmark
	ActionAddBookmark
)

// actionNames maps Action values to their string names.
var actionNames = map[Action]string{
	ActionNone:         "none",
	ActionMoveDown:     "move_down",
	ActionMoveUp:       "move_up",
	ActionMoveLeft:     "move_left",
	ActionMoveRight:    "move_right",
	ActionEnter:        "enter",
	ActionCopy:         "copy",
	ActionMove:         "move",
	ActionDelete:       "delete",
	ActionRename:       "rename",
	ActionNewFile:      "new_file",
	ActionNewDirectory: "new_directory",
	ActionMark:         "mark",
	ActionToggleInfo:   "toggle_info",
	ActionToggleHidden: "toggle_hidden",
	ActionSort:         "sort",
	ActionHelp:         "help",
	ActionHome:         "home",
	ActionPrevDir:      "prev_dir",
	ActionRefresh:      "refresh",
	ActionSyncPane:     "sync_pane",
	ActionSearch:       "search",
	ActionRegexSearch:  "regex_search",
	ActionView:         "view",
	ActionEdit:         "edit",
	ActionShellCommand: "shell_command",
	ActionContextMenu:  "context_menu",
	ActionQuit:         "quit",
	ActionEscape:       "escape",
	ActionBookmark:     "bookmark",
	ActionAddBookmark:  "add_bookmark",
}

// nameToAction maps string names to Action values.
var nameToAction = map[string]Action{
	"move_down":     ActionMoveDown,
	"move_up":       ActionMoveUp,
	"move_left":     ActionMoveLeft,
	"move_right":    ActionMoveRight,
	"enter":         ActionEnter,
	"copy":          ActionCopy,
	"move":          ActionMove,
	"delete":        ActionDelete,
	"rename":        ActionRename,
	"new_file":      ActionNewFile,
	"new_directory": ActionNewDirectory,
	"mark":          ActionMark,
	"toggle_info":   ActionToggleInfo,
	"toggle_hidden": ActionToggleHidden,
	"sort":          ActionSort,
	"help":          ActionHelp,
	"home":          ActionHome,
	"prev_dir":      ActionPrevDir,
	"refresh":       ActionRefresh,
	"sync_pane":     ActionSyncPane,
	"search":        ActionSearch,
	"regex_search":  ActionRegexSearch,
	"view":          ActionView,
	"edit":          ActionEdit,
	"shell_command": ActionShellCommand,
	"context_menu":  ActionContextMenu,
	"quit":          ActionQuit,
	"escape":        ActionEscape,
	"bookmark":      ActionBookmark,
	"add_bookmark":  ActionAddBookmark,
}

// String returns the string name of the action.
func (a Action) String() string {
	if name, ok := actionNames[a]; ok {
		return name
	}
	return "unknown"
}

// ActionFromName returns the Action for the given name.
func ActionFromName(name string) Action {
	if action, ok := nameToAction[name]; ok {
		return action
	}
	return ActionNone
}
