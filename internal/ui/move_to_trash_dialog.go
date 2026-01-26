package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// trashConfirmResultMsg is sent when the user confirms or cancels the move to trash
type trashConfirmResultMsg struct {
	confirmed bool
	paths     []string
}

// MoveToTrashDialog shows confirmation for moving files to trash
type MoveToTrashDialog struct {
	BaseDialog
	paths  []string
	styles DialogStyles
}

// NewMoveToTrashDialog creates a new move to trash confirmation dialog
func NewMoveToTrashDialog(paths []string) *MoveToTrashDialog {
	base := NewBaseDialog(DialogDisplayPane)
	return &MoveToTrashDialog{
		BaseDialog: base,
		paths:      paths,
		styles:     DefaultDialogStyles(base.Width()),
	}
}

// Update handles key input
func (d *MoveToTrashDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			d.Close()
			paths := d.paths
			return d, func() tea.Msg {
				return trashConfirmResultMsg{confirmed: true, paths: paths}
			}

		case "n", "N", "esc", "ctrl+c":
			d.Close()
			return d, func() tea.Msg {
				return trashConfirmResultMsg{confirmed: false, paths: nil}
			}
		}
	}

	return d, nil
}

// View renders the dialog
func (d *MoveToTrashDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(d.styles.Title.Render("Move to Trash"))
	b.WriteString("\n\n")

	// Message based on file count
	var msg string
	if len(d.paths) == 1 {
		filename := filepath.Base(d.paths[0])
		msg = fmt.Sprintf("Move '%s' to trash?", filename)
	} else {
		msg = fmt.Sprintf("Move %d items to trash?", len(d.paths))
	}
	b.WriteString(d.styles.Body.Render(msg))
	b.WriteString("\n\n")

	// Warning message about disk space (use appropriate singular/plural form)
	var warning1 string
	if len(d.paths) == 1 {
		warning1 = "File will not be permanently deleted."
	} else {
		warning1 = "Files will not be permanently deleted."
	}
	warning2 := "Disk space will not be freed until trash is emptied."
	b.WriteString(d.styles.Footer.Render(warning1))
	b.WriteString("\n")
	b.WriteString(d.styles.Footer.Render(warning2))
	b.WriteString("\n\n")

	// Buttons
	b.WriteString(d.styles.Footer.Render("[Y]es  [N]o"))

	return d.styles.Box.Render(b.String())
}
