package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ArchiveWarningType represents the type of archive warning
type ArchiveWarningType int

const (
	// ArchiveWarningCompressionBomb indicates potential zip bomb detection
	ArchiveWarningCompressionBomb ArchiveWarningType = iota
	// ArchiveWarningDiskSpace indicates insufficient disk space
	ArchiveWarningDiskSpace
)

// ArchiveWarningChoice represents the user's choice on a warning dialog
type ArchiveWarningChoice int

const (
	// ArchiveWarningContinue - user wants to continue despite warning
	ArchiveWarningContinue ArchiveWarningChoice = iota
	// ArchiveWarningCancel - user wants to cancel the operation
	ArchiveWarningCancel
)

// archiveWarningResultMsg is sent when the warning dialog is closed
type archiveWarningResultMsg struct {
	warningType ArchiveWarningType
	choice      ArchiveWarningChoice
	archivePath string
}

// ArchiveWarningDialog displays security warnings before extraction
type ArchiveWarningDialog struct {
	warningType   ArchiveWarningType
	archivePath   string
	archiveSize   int64
	extractedSize int64
	availableSize int64
	ratio         float64
	selectedIndex int // 0 = Continue, 1 = Cancel
	active        bool
	width         int
}

// NewCompressionBombWarningDialog creates a dialog for compression bomb warning
func NewCompressionBombWarningDialog(archivePath string, archiveSize, extractedSize int64, ratio float64) *ArchiveWarningDialog {
	return &ArchiveWarningDialog{
		warningType:   ArchiveWarningCompressionBomb,
		archivePath:   archivePath,
		archiveSize:   archiveSize,
		extractedSize: extractedSize,
		ratio:         ratio,
		selectedIndex: 1, // Default to Cancel for safety
		active:        true,
		width:         60,
	}
}

// NewDiskSpaceWarningDialog creates a dialog for disk space warning
func NewDiskSpaceWarningDialog(archivePath string, requiredSize, availableSize int64) *ArchiveWarningDialog {
	return &ArchiveWarningDialog{
		warningType:   ArchiveWarningDiskSpace,
		archivePath:   archivePath,
		extractedSize: requiredSize,
		availableSize: availableSize,
		selectedIndex: 1, // Default to Cancel for safety
		active:        true,
		width:         60,
	}
}

// Update handles input messages
func (d *ArchiveWarningDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			d.selectedIndex = 0
		case "right", "l":
			d.selectedIndex = 1
		case "tab":
			d.selectedIndex = (d.selectedIndex + 1) % 2
		case "shift+tab":
			d.selectedIndex = (d.selectedIndex + 1) % 2
		case "enter":
			d.active = false
			choice := ArchiveWarningCancel
			if d.selectedIndex == 0 {
				choice = ArchiveWarningContinue
			}
			return d, func() tea.Msg {
				return archiveWarningResultMsg{
					warningType: d.warningType,
					choice:      choice,
					archivePath: d.archivePath,
				}
			}
		case "esc", "n":
			d.active = false
			return d, func() tea.Msg {
				return archiveWarningResultMsg{
					warningType: d.warningType,
					choice:      ArchiveWarningCancel,
					archivePath: d.archivePath,
				}
			}
		case "y":
			d.active = false
			return d, func() tea.Msg {
				return archiveWarningResultMsg{
					warningType: d.warningType,
					choice:      ArchiveWarningContinue,
					archivePath: d.archivePath,
				}
			}
		}
	}

	return d, nil
}

// View renders the dialog
func (d *ArchiveWarningDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder

	// Warning icon and title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("208")). // Orange/warning color
		MarginBottom(1)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246"))

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("228")). // Yellow for emphasis
		Bold(true)

	warningTextStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Italic(true).
		MarginTop(1).
		MarginBottom(1)

	buttonStyle := lipgloss.NewStyle().
		Padding(0, 2).
		Background(lipgloss.Color("240")).
		Foreground(lipgloss.Color("252"))

	selectedButtonStyle := buttonStyle.
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("255")).
		Bold(true)

	if d.warningType == ArchiveWarningCompressionBomb {
		b.WriteString(titleStyle.Render("Warning: Large extraction ratio detected"))
		b.WriteString("\n\n")

		// Archive size
		b.WriteString(infoStyle.Render("Archive size: "))
		b.WriteString(highlightStyle.Render(FormatSize(d.archiveSize)))
		b.WriteString("\n")

		// Extracted size with ratio
		b.WriteString(infoStyle.Render("Extracted size: "))
		b.WriteString(highlightStyle.Render(FormatSize(d.extractedSize)))
		b.WriteString(infoStyle.Render(fmt.Sprintf(" (ratio: 1:%.0f)", d.ratio)))
		b.WriteString("\n")

		b.WriteString(warningTextStyle.Render(
			"This may indicate a zip bomb or highly compressed data.\nDo you want to continue?",
		))
	} else {
		b.WriteString(titleStyle.Render("Warning: Insufficient disk space"))
		b.WriteString("\n\n")

		// Required size
		b.WriteString(infoStyle.Render("Required: "))
		b.WriteString(highlightStyle.Render(FormatSize(d.extractedSize)))
		b.WriteString("\n")

		// Available size
		b.WriteString(infoStyle.Render("Available: "))
		b.WriteString(highlightStyle.Render(FormatSize(d.availableSize)))
		b.WriteString("\n")

		b.WriteString(warningTextStyle.Render(
			"Do you want to continue anyway?",
		))
	}

	b.WriteString("\n\n")

	// Buttons
	continueText := "Continue"
	cancelText := "Cancel"

	var continueBtn, cancelBtn string
	if d.selectedIndex == 0 {
		continueBtn = selectedButtonStyle.Render(continueText)
		cancelBtn = buttonStyle.Render(cancelText)
	} else {
		continueBtn = buttonStyle.Render(continueText)
		cancelBtn = selectedButtonStyle.Render(cancelText)
	}

	buttonLine := lipgloss.JoinHorizontal(lipgloss.Center, continueBtn, "  ", cancelBtn)
	b.WriteString(lipgloss.NewStyle().Width(d.width - 4).Align(lipgloss.Center).Render(buttonLine))

	b.WriteString("\n\n")

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
	b.WriteString(helpStyle.Render("[y] Continue  [n/Esc] Cancel  [Tab/Arrow] Switch"))

	// Box border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("208")). // Orange border for warning
		Padding(1, 2).
		Width(d.width)

	return boxStyle.Render(b.String())
}

// IsActive returns whether the dialog is active
func (d *ArchiveWarningDialog) IsActive() bool {
	return d.active
}

// DisplayType returns the display type for this dialog
func (d *ArchiveWarningDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}
