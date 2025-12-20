// Package ui provides context menu dialog functionality for duofm.
// The context menu allows users to discover and execute file operations
// through an intuitive menu interface accessible via the @ key.
package ui

import (
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/fs"
)

// ContextMenuDialog represents a context menu that displays available file operations.
// It implements the Dialog interface and manages menu items, cursor position, and pagination.
type ContextMenuDialog struct {
	items        []MenuItem // Available menu items
	cursor       int        // Current cursor position
	currentPage  int        // Current page number (for pagination)
	itemsPerPage int        // Maximum items per page (default: 9)
	active       bool       // Whether dialog is active
	width        int        // Calculated dialog width
	minWidth     int        // Minimum dialog width
	maxWidth     int        // Maximum dialog width
	pane         *Pane      // Reference to active pane for symlink navigation
}

// MenuItem represents a single menu item with an action closure.
// Each menu item has a unique ID, display label, action function, and enabled state.
type MenuItem struct {
	ID      string       // Unique identifier (e.g., "copy", "move", "delete")
	Label   string       // Display text shown to user
	Action  func() error // Action to execute when selected
	Enabled bool         // Whether item is selectable (disabled items are dimmed)
}

// contextMenuResultMsg is sent when the context menu is closed.
// It contains either an action to execute or a cancellation flag.
type contextMenuResultMsg struct {
	action    func() error // Action to execute (nil if cancelled)
	actionID  string       // ID of the selected action (e.g., "delete" for confirmation)
	cancelled bool         // True if menu was cancelled with Esc
}

// NewContextMenuDialog creates a new context menu dialog for the given file entry.
// The menu items are generated based on the file type and properties.
//
// Parameters:
//   - entry: The file entry to create menu for
//   - sourcePath: Current directory path
//   - destPath: Opposite pane directory path (for copy/move operations)
//
// Returns a new ContextMenuDialog instance.
func NewContextMenuDialog(entry *fs.FileEntry, sourcePath, destPath string) *ContextMenuDialog {
	return NewContextMenuDialogWithPane(entry, sourcePath, destPath, nil)
}

// NewContextMenuDialogWithPane creates a new context menu dialog with pane reference.
// This variant allows symlink navigation by providing access to the active pane.
//
// Parameters:
//   - entry: The file entry to create menu for
//   - sourcePath: Current directory path
//   - destPath: Opposite pane directory path
//   - pane: Reference to active pane (for symlink navigation)
//
// Returns a new ContextMenuDialog instance.
func NewContextMenuDialogWithPane(entry *fs.FileEntry, sourcePath, destPath string, pane *Pane) *ContextMenuDialog {
	d := &ContextMenuDialog{
		cursor:       0,
		currentPage:  0,
		itemsPerPage: 9,
		active:       true,
		minWidth:     40,
		maxWidth:     60,
		pane:         pane,
	}

	d.items = d.buildMenuItems(entry, sourcePath, destPath)
	d.calculateWidth()

	return d
}

// buildMenuItems generates menu items based on file type
func (d *ContextMenuDialog) buildMenuItems(entry *fs.FileEntry, sourcePath, destPath string) []MenuItem {
	items := []MenuItem{}

	// Basic operations available for all file types
	items = append(items, MenuItem{
		ID:    "copy",
		Label: "Copy to other pane",
		Action: func() error {
			fullPath := filepath.Join(sourcePath, entry.Name)
			return fs.Copy(fullPath, destPath)
		},
		Enabled: true,
	})

	items = append(items, MenuItem{
		ID:    "move",
		Label: "Move to other pane",
		Action: func() error {
			fullPath := filepath.Join(sourcePath, entry.Name)
			return fs.MoveFile(fullPath, destPath)
		},
		Enabled: true,
	})

	items = append(items, MenuItem{
		ID:    "delete",
		Label: "Delete",
		Action: func() error {
			fullPath := filepath.Join(sourcePath, entry.Name)
			return fs.Delete(fullPath)
		},
		Enabled: true,
	})

	// Symlink-specific operations
	if entry.IsSymlink && entry.IsDir && !entry.LinkBroken {
		items = append(items, MenuItem{
			ID:    "enter_logical",
			Label: "Enter as directory (logical path)",
			Action: func() error {
				// Navigate to the symlink directory (following the symlink)
				if d.pane != nil {
					fullPath := filepath.Join(sourcePath, entry.Name)
					return d.pane.ChangeDirectory(fullPath)
				}
				return nil
			},
			Enabled: true,
		})

		items = append(items, MenuItem{
			ID:    "enter_physical",
			Label: "Open link target (physical path)",
			Action: func() error {
				// Navigate to the parent directory of the actual target
				if d.pane != nil {
					targetPath := entry.LinkTarget
					var targetDir string

					if filepath.IsAbs(targetPath) {
						targetDir = filepath.Dir(targetPath)
					} else {
						absTarget := filepath.Join(sourcePath, targetPath)
						targetDir = filepath.Dir(absTarget)
					}

					return d.pane.ChangeDirectory(targetDir)
				}
				return nil
			},
			Enabled: !entry.LinkBroken,
		})
	}

	return items
}

// Update handles keyboard input
func (d *ContextMenuDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			// Move cursor down with boundary check
			d.cursor++
			visibleItems := len(d.getCurrentPageItems())
			if d.cursor >= visibleItems {
				d.cursor = 0
			}
			return d, nil

		case "k", "up":
			// Move cursor up with boundary check
			d.cursor--
			if d.cursor < 0 {
				visibleItems := len(d.getCurrentPageItems())
				d.cursor = visibleItems - 1
			}
			return d, nil

		case "esc":
			// Cancel and close
			d.active = false
			return d, func() tea.Msg {
				return contextMenuResultMsg{cancelled: true}
			}

		case "enter":
			// Execute selected action
			items := d.getCurrentPageItems()
			if d.cursor >= 0 && d.cursor < len(items) {
				selectedItem := items[d.cursor]
				if selectedItem.Enabled {
					d.active = false
					return d, func() tea.Msg {
						return contextMenuResultMsg{
							action:   selectedItem.Action,
							actionID: selectedItem.ID,
						}
					}
				}
			}
			return d, nil

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Numeric key direct selection
			num := int(msg.String()[0]-'0') - 1
			items := d.getCurrentPageItems()
			if num >= 0 && num < len(items) {
				selectedItem := items[num]
				if selectedItem.Enabled {
					d.active = false
					return d, func() tea.Msg {
						return contextMenuResultMsg{
							action:   selectedItem.Action,
							actionID: selectedItem.ID,
						}
					}
				}
			}
			return d, nil
		}
	}

	return d, nil
}

// View renders the context menu
func (d *ContextMenuDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder

	// Title with page indicator
	titleText := "Context Menu"
	totalPages := d.getTotalPages()
	if totalPages > 1 {
		titleText = lipgloss.JoinHorizontal(
			lipgloss.Center,
			"Context Menu ",
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(
				"("+string(rune(d.currentPage+1))+"/"+string(rune(totalPages))+")",
			),
		)
	}

	titleStyle := lipgloss.NewStyle().
		Width(d.width-4).
		Padding(0, 2).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	b.WriteString(titleStyle.Render(titleText))
	b.WriteString("\n\n")

	// Menu items
	items := d.getCurrentPageItems()
	for i, item := range items {
		itemNumber := i + 1
		itemText := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(string(rune('0'+itemNumber))+". "),
			item.Label,
		)

		itemStyle := lipgloss.NewStyle().
			Width(d.width-4).
			Padding(0, 2)

		// Highlight selected item
		if i == d.cursor {
			itemStyle = itemStyle.
				Background(lipgloss.Color("39")).
				Foreground(lipgloss.Color("0"))
		}

		// Dim disabled items
		if !item.Enabled {
			itemStyle = itemStyle.Foreground(lipgloss.Color("240"))
		}

		b.WriteString(itemStyle.Render(itemText))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Footer with keyboard hints
	footerText := "[j/k] Navigate  [1-9] Select  [Enter] Execute  [Esc] Cancel"
	if totalPages > 1 {
		footerText = "[h/l] Page  " + footerText
	}

	footerStyle := lipgloss.NewStyle().
		Width(d.width-4).
		Padding(0, 2).
		Foreground(lipgloss.Color("240"))
	b.WriteString(footerStyle.Render(footerText))

	// Border
	boxStyle := lipgloss.NewStyle().
		Width(d.width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// IsActive returns whether the dialog is active
func (d *ContextMenuDialog) IsActive() bool {
	return d.active
}

// getCurrentPageItems returns items for the current page
func (d *ContextMenuDialog) getCurrentPageItems() []MenuItem {
	start := d.currentPage * d.itemsPerPage
	end := start + d.itemsPerPage
	if end > len(d.items) {
		end = len(d.items)
	}
	return d.items[start:end]
}

// getTotalPages returns the total number of pages
func (d *ContextMenuDialog) getTotalPages() int {
	return (len(d.items) + d.itemsPerPage - 1) / d.itemsPerPage
}

// calculateWidth calculates the optimal width for the menu
func (d *ContextMenuDialog) calculateWidth() {
	maxLabelWidth := 0
	for _, item := range d.items {
		labelWidth := len(item.Label) + 3 // "1. " prefix
		if labelWidth > maxLabelWidth {
			maxLabelWidth = labelWidth
		}
	}

	d.width = maxLabelWidth + 8 // Padding
	if d.width < d.minWidth {
		d.width = d.minWidth
	}
	if d.width > d.maxWidth {
		d.width = d.maxWidth
	}
}
