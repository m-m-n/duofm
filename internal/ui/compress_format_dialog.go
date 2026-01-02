// Package ui provides archive format selection dialog functionality for duofm.
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/archive"
)

// CompressFormatDialog allows users to select an archive format for compression.
type CompressFormatDialog struct {
	formats []archive.ArchiveFormat // Available formats
	cursor  int                     // Current cursor position
	active  bool                    // Whether dialog is active
	width   int                     // Dialog width
}

// compressFormatResultMsg is sent when a format is selected or dialog is cancelled.
type compressFormatResultMsg struct {
	format    archive.ArchiveFormat // Selected format
	cancelled bool                  // True if cancelled
}

// NewCompressFormatDialog creates a new format selection dialog.
func NewCompressFormatDialog() *CompressFormatDialog {
	formats := archive.GetAvailableFormats()

	return &CompressFormatDialog{
		formats: formats,
		cursor:  0,
		active:  true,
		width:   50,
	}
}

// Update handles keyboard input for format selection.
func (d *CompressFormatDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			d.cursor++
			if d.cursor >= len(d.formats) {
				d.cursor = 0
			}
			return d, nil

		case "k", "up":
			d.cursor--
			if d.cursor < 0 {
				d.cursor = len(d.formats) - 1
			}
			return d, nil

		case "esc", "ctrl+c":
			d.active = false
			return d, func() tea.Msg {
				return compressFormatResultMsg{cancelled: true}
			}

		case "enter":
			if d.cursor >= 0 && d.cursor < len(d.formats) {
				selectedFormat := d.formats[d.cursor]
				d.active = false
				return d, func() tea.Msg {
					return compressFormatResultMsg{format: selectedFormat}
				}
			}
			return d, nil

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			num := int(msg.String()[0]-'0') - 1
			if num >= 0 && num < len(d.formats) {
				selectedFormat := d.formats[num]
				d.active = false
				return d, func() tea.Msg {
					return compressFormatResultMsg{format: selectedFormat}
				}
			}
			return d, nil
		}
	}

	return d, nil
}

// View renders the format selection dialog.
func (d *CompressFormatDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Width(d.width-4).
		Padding(0, 2).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	b.WriteString(titleStyle.Render("Select Archive Format"))
	b.WriteString("\n\n")

	// Format items
	formatLabels := map[archive.ArchiveFormat]string{
		archive.FormatTar:    "tar (no compression)",
		archive.FormatTarGz:  "tar.gz (gzip compression)",
		archive.FormatTarBz2: "tar.bz2 (bzip2 compression)",
		archive.FormatTarXz:  "tar.xz (LZMA compression)",
		archive.FormatZip:    "zip (deflate compression)",
		archive.Format7z:     "7z (LZMA2 compression)",
	}

	for i, format := range d.formats {
		itemNumber := i + 1
		label := formatLabels[format]
		if label == "" {
			label = format.String()
		}

		itemText := lipgloss.JoinHorizontal(
			lipgloss.Left,
			lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(fmt.Sprintf("%d. ", itemNumber)),
			label,
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

		b.WriteString(itemStyle.Render(itemText))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Footer
	footerStyle := lipgloss.NewStyle().
		Width(d.width-4).
		Padding(0, 2).
		Foreground(lipgloss.Color("240"))
	b.WriteString(footerStyle.Render("[j/k] Navigate  [1-9] Select  [Enter] Confirm  [Esc] Cancel"))

	// Border
	boxStyle := lipgloss.NewStyle().
		Width(d.width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// IsActive returns whether the dialog is active.
func (d *CompressFormatDialog) IsActive() bool {
	return d.active
}

// DisplayType returns the dialog display type.
func (d *CompressFormatDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}
