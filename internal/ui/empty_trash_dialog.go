package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// emptyTrashResultMsg is sent when the user confirms or cancels emptying trash
type emptyTrashResultMsg struct {
	confirmed bool
}

// EmptyTrashDialog shows confirmation for emptying the trash
type EmptyTrashDialog struct {
	BaseDialog
	itemCount int
	styles    DialogStyles
}

// NewEmptyTrashDialog creates a new empty trash confirmation dialog
func NewEmptyTrashDialog(itemCount int) *EmptyTrashDialog {
	base := NewBaseDialog(DialogDisplayPane)
	return &EmptyTrashDialog{
		BaseDialog: base,
		itemCount:  itemCount,
		styles:     ErrorDialogStyles(base.Width()), // Use error/danger styling
	}
}

// Update handles key input
func (d *EmptyTrashDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			d.Close()
			return d, func() tea.Msg {
				return emptyTrashResultMsg{confirmed: true}
			}

		case "n", "N", "esc", "ctrl+c":
			d.Close()
			return d, func() tea.Msg {
				return emptyTrashResultMsg{confirmed: false}
			}
		}
	}

	return d, nil
}

// View renders the dialog
func (d *EmptyTrashDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(d.styles.Title.Render("Empty Trash"))
	b.WriteString("\n\n")

	// Warning message with item count
	msg := fmt.Sprintf("Permanently delete %d item(s) in trash?", d.itemCount)
	b.WriteString(d.styles.Body.Render(msg))
	b.WriteString("\n")
	b.WriteString(d.styles.Error.Render("This action cannot be undone."))
	b.WriteString("\n\n")

	// Buttons
	b.WriteString(d.styles.Footer.Render("[Y]es  [N]o"))

	return d.styles.Box.Render(b.String())
}
