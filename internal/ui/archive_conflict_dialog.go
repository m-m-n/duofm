package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ArchiveConflictChoice represents the user's choice in the archive conflict dialog
type ArchiveConflictChoice int

const (
	ArchiveConflictOverwrite ArchiveConflictChoice = iota
	ArchiveConflictRename
	ArchiveConflictCancel
)

// archiveConflictResultMsg is sent when the archive conflict dialog is closed
type archiveConflictResultMsg struct {
	choice      ArchiveConflictChoice
	archivePath string // Original archive path
	newName     string // New name if rename was chosen
	cancelled   bool   // True if cancelled
}

// ArchiveConflictDialog displays conflict resolution options for archive creation
type ArchiveConflictDialog struct {
	archivePath  string    // Full path to the archive
	filename     string    // Archive filename
	destDir      string    // Destination directory
	existingMod  time.Time // Modification time of existing file
	existingSize int64     // Size of existing file
	cursor       int       // Current selection (0-2)
	active       bool      // Whether dialog is active
	width        int       // Dialog width
}

// NewArchiveConflictDialog creates a new archive conflict resolution dialog
func NewArchiveConflictDialog(archivePath string) *ArchiveConflictDialog {
	filename := filepath.Base(archivePath)
	destDir := filepath.Dir(archivePath)

	// Get existing file info
	var existingMod time.Time
	var existingSize int64
	if info, err := os.Stat(archivePath); err == nil {
		existingMod = info.ModTime()
		existingSize = info.Size()
	}

	return &ArchiveConflictDialog{
		archivePath:  archivePath,
		filename:     filename,
		destDir:      destDir,
		existingMod:  existingMod,
		existingSize: existingSize,
		cursor:       0,
		active:       true,
		width:        55,
	}
}

// Update handles keyboard input
func (d *ArchiveConflictDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
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
			d.active = false
			return d, d.createResultCmd(ArchiveConflictOverwrite)

		case "2":
			d.active = false
			return d, d.createResultCmd(ArchiveConflictRename)

		case "3":
			d.active = false
			return d, d.createResultCmd(ArchiveConflictCancel)

		case "enter":
			d.active = false
			return d, d.createResultCmd(ArchiveConflictChoice(d.cursor))

		case "esc", "ctrl+c":
			d.active = false
			return d, d.createResultCmd(ArchiveConflictCancel)
		}
	}

	return d, nil
}

// createResultCmd creates a command that returns the dialog result
func (d *ArchiveConflictDialog) createResultCmd(choice ArchiveConflictChoice) tea.Cmd {
	return func() tea.Msg {
		return archiveConflictResultMsg{
			choice:      choice,
			archivePath: d.archivePath,
			cancelled:   choice == ArchiveConflictCancel,
		}
	}
}

// View renders the dialog
func (d *ArchiveConflictDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder
	width := d.width

	// Title style
	titleStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 2).
		Bold(true).
		Foreground(lipgloss.Color("208")) // Orange for warning
	b.WriteString(titleStyle.Render("Archive already exists"))
	b.WriteString("\n\n")

	// Filename
	filenameStyle := lipgloss.NewStyle().Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	messageStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 2)

	b.WriteString(messageStyle.Render(
		fmt.Sprintf("File: %s", filenameStyle.Render(d.filename)),
	))
	b.WriteString("\n")
	b.WriteString(messageStyle.Render(mutedStyle.Render(truncatePathForDialog(d.destDir, width-12))))
	b.WriteString("\n\n")

	// Existing file info
	infoStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 2).
		Foreground(lipgloss.Color("246"))

	sizeStr := formatFileSizeForDialog(d.existingSize)
	timeStr := formatModTimeForDialog(d.existingMod)
	b.WriteString(infoStyle.Render(fmt.Sprintf("Existing: %s   %s", sizeStr, mutedStyle.Render(timeStr))))
	b.WriteString("\n\n")

	// Options
	options := []string{"Overwrite", "Rename", "Cancel"}
	for i, opt := range options {
		optStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 2)

		prefix := fmt.Sprintf("%d. ", i+1)
		optText := prefix + opt

		if i == d.cursor {
			optStyle = optStyle.
				Background(lipgloss.Color("208")).
				Foreground(lipgloss.Color("0"))
		}

		b.WriteString(optStyle.Render(optText))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Footer
	footerStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 2).
		Foreground(lipgloss.Color("240"))
	b.WriteString(footerStyle.Render("[j/k] Navigate  [1-3] Select  [Enter] Confirm  [Esc] Cancel"))

	// Border
	boxStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("208")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// IsActive returns whether the dialog is active
func (d *ArchiveConflictDialog) IsActive() bool {
	return d.active
}

// DisplayType returns the dialog display type
func (d *ArchiveConflictDialog) DisplayType() DialogDisplayType {
	return DialogDisplayScreen
}

// formatFileSizeForDialog formats bytes into human-readable size
func formatFileSizeForDialog(bytes int64) string {
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

// formatModTimeForDialog formats time for display
func formatModTimeForDialog(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format("2006-01-02 15:04")
}

// truncatePathForDialog truncates a path if it's too long
func truncatePathForDialog(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= 3 {
		return "..."
	}
	return "..." + path[len(path)-(maxLen-3):]
}

// GenerateUniqueArchiveName generates a unique name by appending a number
func GenerateUniqueArchiveName(originalPath string) string {
	dir := filepath.Dir(originalPath)
	base := filepath.Base(originalPath)

	// Find extension(s) - handle .tar.gz, .tar.bz2, etc.
	var ext, name string
	if strings.HasSuffix(base, ".tar.gz") {
		ext = ".tar.gz"
		name = strings.TrimSuffix(base, ext)
	} else if strings.HasSuffix(base, ".tar.bz2") {
		ext = ".tar.bz2"
		name = strings.TrimSuffix(base, ext)
	} else if strings.HasSuffix(base, ".tar.xz") {
		ext = ".tar.xz"
		name = strings.TrimSuffix(base, ext)
	} else {
		ext = filepath.Ext(base)
		name = strings.TrimSuffix(base, ext)
	}

	// Try appending numbers until we find a unique name
	for i := 1; i < 1000; i++ {
		newName := fmt.Sprintf("%s_%d%s", name, i, ext)
		newPath := filepath.Join(dir, newName)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	// Fallback with timestamp
	timestamp := time.Now().Format("20060102_150405")
	return filepath.Join(dir, fmt.Sprintf("%s_%s%s", name, timestamp, ext))
}
