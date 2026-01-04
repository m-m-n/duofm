package ui

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// OpenWithDialog represents a dialog for opening files with custom applications.
// It allows users to specify a custom application and options to open files with.
type OpenWithDialog struct {
	BaseDialog
	title            string
	applicationInput *TextInput
	fileList         []string
	filesDisplay     string
	workDir          string
	styles           DialogStyles
}

// NewOpenWithDialog creates a new open with dialog for the specified files.
// It attempts to detect the default application for single files using xdg-mime.
func NewOpenWithDialog(files []string, workDir string) *OpenWithDialog {
	base := NewBaseDialog(DialogDisplayPane)

	// Detect default application for single file
	var defaultApp string
	if len(files) == 1 {
		fullPath := filepath.Clean(filepath.Join(workDir, files[0]))
		defaultApp = getDefaultApplication(fullPath)
	}

	d := &OpenWithDialog{
		BaseDialog:       base,
		title:            "Open with Application",
		applicationInput: NewTextInput(defaultApp),
		fileList:         files,
		workDir:          workDir,
	}

	// Format file list for display
	d.filesDisplay = d.formatFileList(files)

	// Set dialog width
	width := 60
	d.SetWidth(width)
	d.styles = NewDialogStyles(width, ColorPrimary)

	return d
}

// getDefaultApplication detects the default application for a file using xdg-mime.
// It queries the MIME type and default application, returning the app name without .desktop extension.
// Returns empty string if detection fails or times out (500ms).
func getDefaultApplication(filePath string) string {
	// Query MIME type
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "xdg-mime", "query", "filetype", filePath)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	mimeType := strings.TrimSpace(string(output))
	if mimeType == "" {
		return ""
	}

	// Query default application
	ctx2, cancel2 := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel2()

	cmd2 := exec.CommandContext(ctx2, "xdg-mime", "query", "default", mimeType)
	output2, err2 := cmd2.Output()
	if err2 != nil {
		return ""
	}

	desktop := strings.TrimSpace(string(output2))
	if desktop == "" {
		return ""
	}

	// Remove .desktop extension
	return strings.TrimSuffix(desktop, ".desktop")
}

// formatFileList formats file list with quotes for display, with truncation for long lists.
// Handles single long filenames and multi-byte UTF-8 characters correctly.
func (d *OpenWithDialog) formatFileList(files []string) string {
	const maxWidth = 50 // Maximum width for file list display

	if len(files) == 0 {
		return ""
	}

	var b strings.Builder
	currentWidth := 0

	for i, file := range files {
		quoted := fmt.Sprintf(`"%s"`, file)
		quotedLen := len([]rune(quoted)) // Count runes, not bytes

		// Special case: first file is too long
		if i == 0 && quotedLen > maxWidth {
			// Truncate the first filename itself
			runes := []rune(quoted)
			truncated := string(runes[:maxWidth-5]) + "\"..." // Keep opening quote, truncate, add closing and ...
			return truncated
		}

		// Check if adding this file would exceed the width
		if i > 0 {
			if currentWidth+1+quotedLen > maxWidth {
				// Truncate with "... and N more"
				remaining := len(files) - i
				b.WriteString(fmt.Sprintf("... and %d more", remaining))
				break
			}
			b.WriteString(" ")
			currentWidth++
		}
		b.WriteString(quoted)
		currentWidth += quotedLen
	}

	return b.String()
}

// Update handles keyboard input for the dialog.
// It processes Enter to confirm, Esc to cancel, and forwards other keys to the text input.
func (d *OpenWithDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			// Cancel
			d.Close()
			return d, func() tea.Msg {
				return openWithDialogResultMsg{cancelled: true}
			}

		case "enter":
			// Submit if application is not empty
			app := d.applicationInput.Value
			if strings.TrimSpace(app) == "" {
				// Empty application - do nothing
				return d, nil
			}

			d.Close()
			return d, func() tea.Msg {
				return openWithDialogResultMsg{
					application: app,
					files:       d.fileList,
					workDir:     d.workDir,
					cancelled:   false,
				}
			}

		default:
			// Forward to text input
			d.applicationInput.HandleKey(msg)
			return d, nil
		}
	}

	return d, nil
}

// View renders the dialog UI with application input, file list, and footer.
func (d *OpenWithDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(d.styles.Title.Render(d.title))
	b.WriteString("\n\n")

	// Application field
	b.WriteString("Application: ")
	b.WriteString(d.applicationInput.RenderWithCursor(d.Width() - 15))
	b.WriteString("\n\n")

	// Files field (read-only)
	b.WriteString("Files: ")
	b.WriteString(d.filesDisplay)
	b.WriteString("\n\n")

	// Footer
	footer := "[Enter] Open  [Esc] Cancel"
	b.WriteString(d.styles.Footer.Render(footer))

	return d.styles.Box.Render(b.String())
}
