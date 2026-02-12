package ui

// Action represents a user action that can be triggered by a keybinding.
type Action int

// Action constants for all 32 actions plus ActionNone.
const (
	ActionNone Action = iota
	// Navigation
	ActionMoveDown
	ActionMoveUp
	ActionMoveLeft
	ActionMoveRight
	ActionEnter
	ActionPageDown
	ActionPageUp
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
	ActionHistoryBack
	ActionHistoryForward
	ActionRefresh
	ActionSyncPane
	// Search
	ActionSearch
	ActionRegexSearch
	ActionSQLFilter
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
	// Permission edit
	ActionPermission
	// Path navigation
	ActionPathJump
	// Rename with full name
	ActionRenameFullName
	// Trash operations
	ActionTrash
	ActionOpenTrash
	ActionRestore
	ActionEmptyTrash
	// Shell log
	ActionShellLog
)

// actionNames maps Action values to their string names.
var actionNames = map[Action]string{
	ActionNone:           "none",
	ActionMoveDown:       "move_down",
	ActionMoveUp:         "move_up",
	ActionMoveLeft:       "move_left",
	ActionMoveRight:      "move_right",
	ActionEnter:          "enter",
	ActionPageDown:       "page_down",
	ActionPageUp:         "page_up",
	ActionCopy:           "copy",
	ActionMove:           "move",
	ActionDelete:         "delete",
	ActionRename:         "rename",
	ActionNewFile:        "new_file",
	ActionNewDirectory:   "new_directory",
	ActionMark:           "mark",
	ActionToggleInfo:     "toggle_info",
	ActionToggleHidden:   "toggle_hidden",
	ActionSort:           "sort",
	ActionHelp:           "help",
	ActionHome:           "home",
	ActionPrevDir:        "prev_dir",
	ActionHistoryBack:    "history_back",
	ActionHistoryForward: "history_forward",
	ActionRefresh:        "refresh",
	ActionSyncPane:       "sync_pane",
	ActionSearch:         "search",
	ActionRegexSearch:    "regex_search",
	ActionSQLFilter:      "sql_filter",
	ActionView:           "view",
	ActionEdit:           "edit",
	ActionShellCommand:   "shell_command",
	ActionContextMenu:    "context_menu",
	ActionQuit:           "quit",
	ActionEscape:         "escape",
	ActionBookmark:       "bookmark",
	ActionAddBookmark:    "add_bookmark",
	ActionPermission:     "permission",
	ActionPathJump:       "path_jump",
	ActionRenameFullName: "rename_full_name",
	ActionTrash:          "trash",
	ActionOpenTrash:      "open_trash",
	ActionRestore:        "restore",
	ActionEmptyTrash:     "empty_trash",
	ActionShellLog:       "shell_log",
}

// nameToAction maps string names to Action values.
var nameToAction = map[string]Action{
	"move_down":        ActionMoveDown,
	"move_up":          ActionMoveUp,
	"move_left":        ActionMoveLeft,
	"move_right":       ActionMoveRight,
	"enter":            ActionEnter,
	"page_down":        ActionPageDown,
	"page_up":          ActionPageUp,
	"copy":             ActionCopy,
	"move":             ActionMove,
	"delete":           ActionDelete,
	"rename":           ActionRename,
	"new_file":         ActionNewFile,
	"new_directory":    ActionNewDirectory,
	"mark":             ActionMark,
	"toggle_info":      ActionToggleInfo,
	"toggle_hidden":    ActionToggleHidden,
	"sort":             ActionSort,
	"help":             ActionHelp,
	"home":             ActionHome,
	"prev_dir":         ActionPrevDir,
	"history_back":     ActionHistoryBack,
	"history_forward":  ActionHistoryForward,
	"refresh":          ActionRefresh,
	"sync_pane":        ActionSyncPane,
	"search":           ActionSearch,
	"regex_search":     ActionRegexSearch,
	"sql_filter":       ActionSQLFilter,
	"view":             ActionView,
	"edit":             ActionEdit,
	"shell_command":    ActionShellCommand,
	"context_menu":     ActionContextMenu,
	"quit":             ActionQuit,
	"escape":           ActionEscape,
	"bookmark":         ActionBookmark,
	"add_bookmark":     ActionAddBookmark,
	"permission":       ActionPermission,
	"path_jump":        ActionPathJump,
	"rename_full_name": ActionRenameFullName,
	"trash":            ActionTrash,
	"open_trash":       ActionOpenTrash,
	"restore":          ActionRestore,
	"empty_trash":      ActionEmptyTrash,
	"shell_log":        ActionShellLog,
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
