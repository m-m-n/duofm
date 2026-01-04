package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OverwriteChoice represents the user's choice in the overwrite dialog
type OverwriteChoice int

const (
	OverwriteChoiceOverwrite OverwriteChoice = iota
	OverwriteChoiceCancel
	OverwriteChoiceRename
)

// OverwriteFileInfo holds file metadata for display
type OverwriteFileInfo struct {
	Size    int64
	ModTime time.Time
}

// OverwriteDialog displays overwrite confirmation options
type OverwriteDialog struct {
	BaseDialog
	filename  string            // Name of the file being copied/moved
	destPath  string            // Destination directory path
	srcPath   string            // Source file full path
	srcInfo   OverwriteFileInfo // Source file information
	destInfo  OverwriteFileInfo // Destination file information
	cursor    int               // Current selection (0-2)
	operation string            // "copy" or "move"
	styles    DialogStyles
}

// overwriteDialogResultMsg is the message sent when the dialog is closed
type overwriteDialogResultMsg struct {
	choice    OverwriteChoice
	srcPath   string
	destPath  string
	filename  string
	operation string
}

// NewOverwriteDialog creates a new overwrite confirmation dialog
func NewOverwriteDialog(filename, destPath string, srcInfo, destInfo OverwriteFileInfo, operation, srcPath string) *OverwriteDialog {
	base := NewBaseDialog(DialogDisplayPane)
	base.SetWidth(50)
	return &OverwriteDialog{
		BaseDialog: base,
		filename:   filename,
		destPath:   destPath,
		srcPath:    srcPath,
		srcInfo:    srcInfo,
		destInfo:   destInfo,
		cursor:     0,
		operation:  operation,
		styles:     NewDialogStyles(50, ColorPrimary),
	}
}

// Update handles keyboard input
func (d *OverwriteDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			d.cursor++
			if d.cursor > 2 {
				d.cursor = 0
			}
			return d, nil

		case "k", "up":
			d.cursor--
			if d.cursor < 0 {
				d.cursor = 2
			}
			return d, nil

		case "1":
			d.Close()
			return d, d.createResultCmd(OverwriteChoiceOverwrite)

		case "2":
			d.Close()
			return d, d.createResultCmd(OverwriteChoiceCancel)

		case "3":
			d.Close()
			return d, d.createResultCmd(OverwriteChoiceRename)

		case "enter":
			d.Close()
			return d, d.createResultCmd(OverwriteChoice(d.cursor))

		case "esc", "ctrl+c":
			d.Close()
			return d, d.createResultCmd(OverwriteChoiceCancel)
		}
	}

	return d, nil
}

// createResultCmd creates a command that returns the dialog result
func (d *OverwriteDialog) createResultCmd(choice OverwriteChoice) tea.Cmd {
	return func() tea.Msg {
		return overwriteDialogResultMsg{
			choice:    choice,
			srcPath:   d.srcPath,
			destPath:  d.destPath,
			filename:  d.filename,
			operation: d.operation,
		}
	}
}

// View renders the dialog
func (d *OverwriteDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render("File already exists"))
	b.WriteString("\n\n")

	// Filename and path
	filenameStyle := lipgloss.NewStyle().Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorMuted)))

	b.WriteString(fmt.Sprintf("%s already exists in", filenameStyle.Render("\""+d.filename+"\"")))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render(truncatePath(d.destPath, width-8)))
	b.WriteString("\n\n")

	// File info comparison
	srcSizeStr := formatFileSize(d.srcInfo.Size)
	srcTimeStr := formatModTime(d.srcInfo.ModTime)
	destSizeStr := formatFileSize(d.destInfo.Size)
	destTimeStr := formatModTime(d.destInfo.ModTime)

	b.WriteString(d.styles.Body.Render(fmt.Sprintf("Source: %s    %s", srcSizeStr, mutedStyle.Render(srcTimeStr))))
	b.WriteString("\n")
	b.WriteString(d.styles.Body.Render(fmt.Sprintf("Dest:   %s    %s", destSizeStr, mutedStyle.Render(destTimeStr))))
	b.WriteString("\n\n")

	// Options
	options := []string{"Overwrite", "Cancel", "Rename"}
	for i, opt := range options {
		optStyle := lipgloss.NewStyle().Width(width-4).Padding(0, 2)

		prefix := fmt.Sprintf("%d. ", i+1)
		optText := prefix + opt

		if i == d.cursor {
			optStyle = optStyle.
				Background(lipgloss.Color(string(ColorPrimary))).
				Foreground(lipgloss.Color("0"))
		}

		b.WriteString(optStyle.Render(optText))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(d.styles.Footer.Render("1-3/j/k:select Enter:confirm"))

	return d.styles.Box.Render(b.String())
}

// formatFileSize formats bytes into human-readable size
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatModTime formats time for display
func formatModTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("2006-01-02 15:04")
}

// truncatePath truncates a path if it's too long
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= 3 {
		return "..."
	}
	return "..." + path[len(path)-(maxLen-3):]
}
