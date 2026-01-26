package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/sakura/duofm/internal/fs"
)

// TrashItem represents a single item in the trash
type TrashItem struct {
	Name         string
	Size         int64
	DeletionTime string
	OriginalPath string
	IsDir        bool
	marked       bool
}

// TrashDialog is a dialog for viewing and managing trash items
type TrashDialog struct {
	BaseDialog
	items         []TrashItem
	cursor        int
	scrollOffset  int
	visibleHeight int
	styles        DialogStyles
}

// trashDialogCloseMsg is sent when the dialog is closed
type trashDialogCloseMsg struct{}

// trashDialogRestoreMsg is sent when user wants to restore items
type trashDialogRestoreMsg struct {
	items []TrashItem
}

// trashDialogEmptyMsg is sent when user wants to empty the trash
type trashDialogEmptyMsg struct{}

// NewTrashDialog creates a new trash dialog
func NewTrashDialog(items []TrashItem) *TrashDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	base.SetWidth(80)

	d := &TrashDialog{
		BaseDialog:    base,
		items:         items,
		cursor:        0,
		scrollOffset:  0,
		visibleHeight: 15,
		styles:        NewDialogStyles(80, ColorPrimary),
	}

	return d
}

// Update handles keyboard input
func (d *TrashDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "j":
				d.moveCursorDown()
				return d, nil
			case "k":
				d.moveCursorUp()
				return d, nil
			case "q":
				d.Close()
				return d, func() tea.Msg {
					return trashDialogCloseMsg{}
				}
			case "r", "R":
				// Restore (accept both lowercase and uppercase for consistency)
				return d.handleRestore()
			case "e", "E":
				// Empty trash (accept both lowercase and uppercase for consistency)
				return d.handleEmptyTrash()
			}

		case tea.KeyDown:
			d.moveCursorDown()
			return d, nil

		case tea.KeyUp:
			d.moveCursorUp()
			return d, nil

		case tea.KeySpace:
			d.toggleMark()
			return d, nil

		case tea.KeyEsc:
			d.Close()
			return d, func() tea.Msg {
				return trashDialogCloseMsg{}
			}
		}
	}

	return d, nil
}

// moveCursorDown moves the cursor down
func (d *TrashDialog) moveCursorDown() {
	if len(d.items) == 0 {
		return
	}
	if d.cursor < len(d.items)-1 {
		d.cursor++
		d.adjustScroll()
	}
}

// moveCursorUp moves the cursor up
func (d *TrashDialog) moveCursorUp() {
	if len(d.items) == 0 {
		return
	}
	if d.cursor > 0 {
		d.cursor--
		d.adjustScroll()
	}
}

// toggleMark toggles the mark on the current item and moves cursor down
func (d *TrashDialog) toggleMark() {
	if len(d.items) == 0 {
		return
	}
	if d.cursor >= 0 && d.cursor < len(d.items) {
		d.items[d.cursor].marked = !d.items[d.cursor].marked
		// Move cursor down after marking (like in file pane)
		if d.cursor < len(d.items)-1 {
			d.cursor++
			d.adjustScroll()
		}
	}
}

// adjustScroll ensures the cursor is visible
func (d *TrashDialog) adjustScroll() {
	// If cursor is above visible area
	if d.cursor < d.scrollOffset {
		d.scrollOffset = d.cursor
	}
	// If cursor is below visible area
	if d.cursor >= d.scrollOffset+d.visibleHeight {
		d.scrollOffset = d.cursor - d.visibleHeight + 1
	}
}

// View renders the dialog
func (d *TrashDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Column widths
	const (
		markWidth = 2
		nameWidth = 20
		sizeWidth = 10
		dateWidth = 17
	)
	pathWidth := width - markWidth - nameWidth - sizeWidth - dateWidth - 12 // 12 for padding and borders

	// Title with item count
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(string(ColorPrimary)))
	countStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorMuted)))

	b.WriteString(titleStyle.Render("Trash"))
	b.WriteString(countStyle.Render(fmt.Sprintf(" [%d]", len(d.items))))
	b.WriteString("\n\n")

	// Empty trash message
	if len(d.items) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorMuted)))
		b.WriteString(emptyStyle.Render("Trash is empty"))
		b.WriteString("\n\n")
		b.WriteString(d.styles.Footer.Render("[Esc/q: close]"))
		return d.styles.Box.Render(b.String())
	}

	// Header row
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(string(ColorMuted)))
	header := fmt.Sprintf("%-*s %-*s %*s %-*s %-*s",
		markWidth, "",
		nameWidth, "Name",
		sizeWidth, "Size",
		dateWidth, "Deleted",
		pathWidth, "Original Path",
	)
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Separator
	b.WriteString(strings.Repeat("-", width-6))
	b.WriteString("\n")

	// Calculate visible range
	endIdx := d.scrollOffset + d.visibleHeight
	if endIdx > len(d.items) {
		endIdx = len(d.items)
	}

	// Render visible items
	for i := d.scrollOffset; i < endIdx; i++ {
		item := d.items[i]
		isSelected := i == d.cursor

		// Mark indicator
		mark := "  "
		if item.marked {
			mark = "* "
		}

		// Name (truncate if needed)
		name := truncateString(item.Name, nameWidth)

		// Size (show "-" for directories)
		var sizeStr string
		if item.IsDir {
			sizeStr = "-"
		} else {
			sizeStr = formatSize(item.Size)
		}

		// Deletion time
		dateStr := item.DeletionTime

		// Original path (truncate from left if needed)
		path := truncatePathFromLeft(item.OriginalPath, pathWidth)

		// Build line
		line := fmt.Sprintf("%s%-*s %*s %-*s %-*s",
			mark,
			nameWidth, name,
			sizeWidth, sizeStr,
			dateWidth, dateStr,
			pathWidth, path,
		)

		// Apply style
		lineStyle := lipgloss.NewStyle()
		if isSelected {
			lineStyle = lineStyle.
				Background(lipgloss.Color(string(ColorPrimary))).
				Foreground(lipgloss.Color("0"))
		}

		b.WriteString(lineStyle.Render(line))
		b.WriteString("\n")
	}

	// Padding if fewer items than visible height
	for i := endIdx - d.scrollOffset; i < d.visibleHeight; i++ {
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(d.items) > d.visibleHeight {
		scrollInfo := fmt.Sprintf("[%d-%d/%d]", d.scrollOffset+1, endIdx, len(d.items))
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorMuted))).Render(scrollInfo))
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(d.styles.Footer.Render("[j/k: move] [Space: mark] [r: restore] [e: empty] [Esc/q: close]"))

	return d.styles.Box.Render(b.String())
}

// SelectedItem returns the currently selected item, or nil if none
func (d *TrashDialog) SelectedItem() *TrashItem {
	if len(d.items) == 0 || d.cursor < 0 || d.cursor >= len(d.items) {
		return nil
	}
	return &d.items[d.cursor]
}

// GetMarkedItems returns all marked items, or the selected item if none are marked
func (d *TrashDialog) GetMarkedItems() []TrashItem {
	var marked []TrashItem
	for _, item := range d.items {
		if item.marked {
			marked = append(marked, item)
		}
	}
	return marked
}

// handleRestore handles the R key press for restoring items
func (d *TrashDialog) handleRestore() (Dialog, tea.Cmd) {
	if len(d.items) == 0 {
		return d, nil
	}

	// Get items to restore: marked items or current item
	var itemsToRestore []TrashItem
	marked := d.GetMarkedItems()
	if len(marked) > 0 {
		itemsToRestore = marked
	} else if d.cursor >= 0 && d.cursor < len(d.items) {
		itemsToRestore = []TrashItem{d.items[d.cursor]}
	}

	if len(itemsToRestore) == 0 {
		return d, nil
	}

	return d, func() tea.Msg {
		return trashDialogRestoreMsg{items: itemsToRestore}
	}
}

// handleEmptyTrash handles the e/E key press for emptying trash
func (d *TrashDialog) handleEmptyTrash() (Dialog, tea.Cmd) {
	if len(d.items) == 0 {
		return d, nil
	}

	return d, func() tea.Msg {
		return trashDialogEmptyMsg{}
	}
}

// loadTrashItems loads all items from the trash directory
func loadTrashItems() ([]TrashItem, error) {
	filesDir := fs.TrashFilesDir()
	infoDir := fs.TrashInfoDir()

	// Read files directory
	entries, err := os.ReadDir(filesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read trash files: %w", err)
	}

	var items []TrashItem
	for _, entry := range entries {
		name := entry.Name()

		// Read trashinfo file
		trashinfoPath := filepath.Join(infoDir, name+".trashinfo")
		info, err := fs.ParseTrashinfo(trashinfoPath)
		if err != nil {
			// Skip items without valid trashinfo
			continue
		}

		// Get file info
		itemPath := filepath.Join(filesDir, name)
		fileInfo, err := os.Lstat(itemPath)
		if err != nil {
			continue
		}

		item := TrashItem{
			Name:         name,
			Size:         fileInfo.Size(),
			DeletionTime: info.DeletionDate.Format("2006-01-02 15:04"),
			OriginalPath: info.OriginalPath,
			IsDir:        fileInfo.IsDir(),
			marked:       false,
		}

		items = append(items, item)
	}

	return items, nil
}

// truncateString truncates a string to fit within maxWidth
func truncateString(s string, maxWidth int) string {
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}

	result := ""
	width := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if width+rw+1 > maxWidth { // +1 for ellipsis
			break
		}
		result += string(r)
		width += rw
	}
	return result + "~"
}

// truncatePathFromLeft truncates a path from the left to fit within maxWidth
func truncatePathFromLeft(path string, maxWidth int) string {
	if runewidth.StringWidth(path) <= maxWidth {
		return path
	}

	// Start with ellipsis
	prefix := "~"
	available := maxWidth - runewidth.StringWidth(prefix)

	// Take from the right side
	runes := []rune(path)
	result := ""
	width := 0

	for i := len(runes) - 1; i >= 0; i-- {
		rw := runewidth.RuneWidth(runes[i])
		if width+rw > available {
			break
		}
		result = string(runes[i]) + result
		width += rw
	}

	return prefix + result
}

// formatSize formats a file size for display
func formatSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1fG", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1fM", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1fK", float64(size)/KB)
	default:
		return fmt.Sprintf("%dB", size)
	}
}
