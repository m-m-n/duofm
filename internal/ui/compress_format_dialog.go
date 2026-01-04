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
	BaseDialog
	formats []archive.ArchiveFormat // Available formats
	cursor  int                     // Current cursor position
	styles  DialogStyles
}

// compressFormatResultMsg is sent when a format is selected or dialog is cancelled.
type compressFormatResultMsg struct {
	format    archive.ArchiveFormat // Selected format
	cancelled bool                  // True if cancelled
}

// NewCompressFormatDialog creates a new format selection dialog.
func NewCompressFormatDialog() *CompressFormatDialog {
	base := NewBaseDialog(DialogDisplayScreen)
	formats := archive.GetAvailableFormats()

	return &CompressFormatDialog{
		BaseDialog: base,
		formats:    formats,
		cursor:     0,
		styles:     DefaultDialogStyles(base.Width()),
	}
}

// Update handles keyboard input for format selection.
func (d *CompressFormatDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
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
			d.Close()
			return d, func() tea.Msg {
				return compressFormatResultMsg{cancelled: true}
			}

		case "enter":
			if d.cursor >= 0 && d.cursor < len(d.formats) {
				selectedFormat := d.formats[d.cursor]
				d.Close()
				return d, func() tea.Msg {
					return compressFormatResultMsg{format: selectedFormat}
				}
			}
			return d, nil

		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			num := int(msg.String()[0]-'0') - 1
			if num >= 0 && num < len(d.formats) {
				selectedFormat := d.formats[num]
				d.Close()
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
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render("Select Archive Format"))
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
			lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorMuted))).Render(fmt.Sprintf("%d. ", itemNumber)),
			label,
		)

		itemStyle := lipgloss.NewStyle().
			Width(width - 4).
			Padding(0, 2)

		// Highlight selected item
		if i == d.cursor {
			itemStyle = itemStyle.
				Background(lipgloss.Color(string(ColorPrimary))).
				Foreground(lipgloss.Color("0"))
		}

		b.WriteString(itemStyle.Render(itemText))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(d.styles.Footer.Render("[j/k] Navigate  [1-9] Select  [Enter] Confirm  [Esc] Cancel"))

	return d.styles.Box.Render(b.String())
}
