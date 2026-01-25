package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// RestoreConflictChoice represents the user's choice when restoring a file that conflicts
type RestoreConflictChoice int

const (
	RestoreChoiceOverwrite RestoreConflictChoice = iota
	RestoreChoiceRename
	RestoreChoiceSkip
	RestoreChoiceCancelled
)

// restoreConflictResultMsg is sent when the user makes a choice in the restore conflict dialog
type restoreConflictResultMsg struct {
	choice       RestoreConflictChoice
	trashName    string
	originalPath string
}

// RestoreConflictDialog shows options when restoring a file that already exists at the destination
type RestoreConflictDialog struct {
	BaseDialog
	originalPath string
	trashName    string
	styles       DialogStyles
}

// NewRestoreConflictDialog creates a new restore conflict dialog
func NewRestoreConflictDialog(trashName, originalPath string) *RestoreConflictDialog {
	base := NewBaseDialog(DialogDisplayPane)
	return &RestoreConflictDialog{
		BaseDialog:   base,
		trashName:    trashName,
		originalPath: originalPath,
		styles:       DefaultDialogStyles(base.Width()),
	}
}

// Update handles key input
func (d *RestoreConflictDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "o", "O":
			d.Close()
			return d, func() tea.Msg {
				return restoreConflictResultMsg{
					choice:       RestoreChoiceOverwrite,
					trashName:    d.trashName,
					originalPath: d.originalPath,
				}
			}

		case "r", "R":
			d.Close()
			return d, func() tea.Msg {
				return restoreConflictResultMsg{
					choice:       RestoreChoiceRename,
					trashName:    d.trashName,
					originalPath: d.originalPath,
				}
			}

		case "s", "S":
			d.Close()
			return d, func() tea.Msg {
				return restoreConflictResultMsg{
					choice:       RestoreChoiceSkip,
					trashName:    d.trashName,
					originalPath: d.originalPath,
				}
			}

		case "esc", "ctrl+c":
			d.Close()
			return d, func() tea.Msg {
				return restoreConflictResultMsg{
					choice:       RestoreChoiceCancelled,
					trashName:    d.trashName,
					originalPath: d.originalPath,
				}
			}
		}
	}

	return d, nil
}

// View renders the dialog
func (d *RestoreConflictDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(d.styles.Title.Render("File already exists"))
	b.WriteString("\n\n")

	// Path
	b.WriteString(d.styles.Body.Render(d.originalPath))
	b.WriteString("\n\n")

	// Options
	b.WriteString(d.styles.Footer.Render("[O]verwrite  [R]ename  [S]kip"))

	return d.styles.Box.Render(b.String())
}
