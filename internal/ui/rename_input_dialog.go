package ui

import (
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/fs"
)

// RenameInputDialog is an input dialog with real-time validation for rename operations
type RenameInputDialog struct {
	BaseDialog
	title         string
	textInput     *TextInput      // reusable text input component
	destPath      string          // Destination directory
	srcPath       string          // Source file path
	operation     string          // "copy" or "move"
	existingFiles map[string]bool // Cached filenames in dest
	hasError      bool
	errorMessage  string
	suggestedName string
	styles        DialogStyles
}

// renameInputResultMsg is sent when the rename dialog is confirmed or cancelled
type renameInputResultMsg struct {
	newName   string
	srcPath   string
	destPath  string
	operation string
	cancelled bool // True if cancelled with Esc
}

// NewRenameInputDialog creates a new rename input dialog
func NewRenameInputDialog(destPath, srcPath, operation string) *RenameInputDialog {
	// Read destination directory
	existingFiles := loadExistingFiles(destPath)

	// Generate suggested name
	filename := filepath.Base(srcPath)
	suggested := suggestRename(filename, existingFiles)

	base := NewBaseDialog(DialogDisplayPane)
	base.SetWidth(50)
	return &RenameInputDialog{
		BaseDialog:    base,
		title:         "New name:",
		textInput:     NewTextInput(suggested),
		destPath:      destPath,
		srcPath:       srcPath,
		operation:     operation,
		existingFiles: existingFiles,
		hasError:      false,
		suggestedName: suggested,
		styles:        NewDialogStyles(50, ColorPrimary),
	}
}

// loadExistingFiles loads filenames from a directory into a set
func loadExistingFiles(dirPath string) map[string]bool {
	result := make(map[string]bool)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		result[entry.Name()] = true
	}

	return result
}

// suggestRename generates a suggested rename for a file
func suggestRename(filename string, existing map[string]bool) string {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	// Handle hidden files (starting with .) that have no real extension
	// e.g., ".gitignore" -> base="", ext=".gitignore"
	if base == "" && strings.HasPrefix(ext, ".") {
		base = ext
		ext = ""
	}

	// Try "name_copy.ext"
	candidate := base + "_copy" + ext
	if !existing[candidate] {
		return candidate
	}

	// Try "name_copy_2.ext", "name_copy_3.ext", etc.
	for i := 2; i <= 100; i++ {
		candidate = base + "_copy_" + itoa(i) + ext
		if !existing[candidate] {
			return candidate
		}
	}

	return filename
}

// itoa converts an integer to string (simple implementation)
func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}

// validateInput validates the current input
func (d *RenameInputDialog) validateInput() {
	if d.textInput.IsEmpty() {
		d.hasError = true
		d.errorMessage = "File name cannot be empty"
		return
	}

	if d.existingFiles[d.textInput.Value] {
		d.hasError = true
		d.errorMessage = "File already exists"
		return
	}

	if err := fs.ValidateFilename(d.textInput.Value); err != nil {
		d.hasError = true
		d.errorMessage = err.Error()
		return
	}

	d.hasError = false
	d.errorMessage = ""
}

// Update handles keyboard input
func (d *RenameInputDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if d.hasError {
				return d, nil // Do nothing if error
			}
			d.Close()
			return d, func() tea.Msg {
				return renameInputResultMsg{
					newName:   d.textInput.Value,
					srcPath:   d.srcPath,
					destPath:  d.destPath,
					operation: d.operation,
				}
			}

		case tea.KeyEsc, tea.KeyCtrlC:
			d.Close()
			return d, func() tea.Msg {
				return renameInputResultMsg{
					cancelled: true,
				}
			}

		default:
			// Delegate text editing to TextInput
			oldValue := d.textInput.Value
			if d.textInput.HandleKey(msg) {
				// Re-validate if input changed
				if d.textInput.Value != oldValue {
					d.validateInput()
				}
				return d, nil
			}
		}
	}

	return d, nil
}

// View renders the dialog
func (d *RenameInputDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render(d.title))
	b.WriteString("\n\n")

	// Input field
	inputWidth := width - 8
	b.WriteString(d.renderInputField(inputWidth))
	b.WriteString("\n")

	// Error message (if any)
	if d.hasError {
		b.WriteString("\n")
		b.WriteString(d.styles.Error.Render(d.errorMessage))
	}

	b.WriteString("\n")

	// Footer
	if d.hasError {
		b.WriteString(d.styles.Footer.Render("Esc: Cancel"))
	} else {
		b.WriteString(d.styles.Footer.Render("Enter: Confirm  Esc: Cancel"))
	}

	return d.styles.Box.Render(b.String())
}

// renderInputField renders the input field
func (d *RenameInputDialog) renderInputField(width int) string {
	// Input field style
	fieldStyle := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	return fieldStyle.Render(d.textInput.RenderWithCursor(width - 2))
}

// Input returns the current input value.
func (d *RenameInputDialog) Input() string {
	return d.textInput.Value
}

// SetInput sets the input value and positions cursor at the end.
func (d *RenameInputDialog) SetInput(value string) {
	d.textInput.Value = value
	d.textInput.CursorPos = len([]rune(value))
	d.validateInput()
}

// CursorPos returns the current cursor position.
func (d *RenameInputDialog) CursorPos() int {
	return d.textInput.CursorPos
}
