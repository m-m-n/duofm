package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sakura/duofm/internal/fs"
)

// ExtensionRenameDialog is a rename dialog that preserves file extension.
// The input field contains only the base name, while the extension is
// displayed separately and cannot be edited.
type ExtensionRenameDialog struct {
	BaseDialog
	title         string
	textInput     *TextInput
	extension     string          // Fixed extension (e.g., ".txt")
	dirPath       string          // Directory containing the file
	originalName  string          // Original filename (full name with extension)
	existingFiles map[string]bool // Cached filenames for duplicate check
	hasError      bool
	errorMessage  string
	styles        DialogStyles
}

// extensionRenameResultMsg is sent when the dialog is confirmed or cancelled
type extensionRenameResultMsg struct {
	newName   string // Full new name (base + extension)
	oldName   string // Original filename
	dirPath   string // Directory path
	cancelled bool
}

// NewExtensionRenameDialog creates a new extension-preserving rename dialog.
// Parameters:
//   - dirPath: directory containing the file
//   - originalName: original filename (full name with extension)
//   - baseName: editable part of the filename
//   - extension: fixed extension (e.g., ".txt")
func NewExtensionRenameDialog(dirPath, originalName, baseName, extension string) *ExtensionRenameDialog {
	// Load existing files for validation
	existingFiles := loadExistingFiles(dirPath)

	base := NewBaseDialog(DialogDisplayPane)
	base.SetWidth(50)

	title := "Rename"
	if extension != "" {
		title = "Rename (extension: " + extension + "):"
	}

	d := &ExtensionRenameDialog{
		BaseDialog:    base,
		title:         title,
		textInput:     NewTextInput(baseName),
		extension:     extension,
		dirPath:       dirPath,
		originalName:  originalName,
		existingFiles: existingFiles,
		hasError:      false,
		errorMessage:  "",
		styles:        NewDialogStyles(50, ColorPrimary),
	}

	// Validate initial input (may be invalid if empty)
	d.validateInput()

	return d
}

// validateInput validates the current input.
func (d *ExtensionRenameDialog) validateInput() {
	// Empty check
	if d.textInput.IsEmpty() {
		d.hasError = true
		d.errorMessage = "File name cannot be empty"
		return
	}

	// Build full filename
	fullName := d.textInput.Value + d.extension

	// Duplicate check (excluding the original filename)
	if fullName != d.originalName && d.existingFiles[fullName] {
		d.hasError = true
		d.errorMessage = "File already exists"
		return
	}

	// Invalid characters check (validate full name)
	if err := fs.ValidateFilename(fullName); err != nil {
		d.hasError = true
		d.errorMessage = err.Error()
		return
	}

	d.hasError = false
	d.errorMessage = ""
}

// Update handles keyboard input.
func (d *ExtensionRenameDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if d.hasError {
				return d, nil // Do nothing if there's an error
			}
			d.Close()
			return d, d.createResultCmd(false)

		case tea.KeyEsc:
			d.Close()
			return d, d.createResultCmd(true)

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

// createResultCmd creates a command that returns the dialog result.
func (d *ExtensionRenameDialog) createResultCmd(cancelled bool) tea.Cmd {
	return func() tea.Msg {
		fullName := d.textInput.Value + d.extension
		return extensionRenameResultMsg{
			newName:   fullName,
			oldName:   d.originalName,
			dirPath:   d.dirPath,
			cancelled: cancelled,
		}
	}
}

// View renders the dialog.
func (d *ExtensionRenameDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render(d.title))
	b.WriteString("\n\n")

	// Input field with extension
	inputWidth := width - 8 - len(d.extension) - 2 // Account for extension display
	if inputWidth < 20 {
		inputWidth = 20 // Minimum usable width
	}
	b.WriteString(d.renderInputFieldWithExtension(inputWidth))
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

// renderInputFieldWithExtension renders the input field with fixed extension display.
func (d *ExtensionRenameDialog) renderInputFieldWithExtension(inputWidth int) string {
	// Input field style
	fieldStyle := lipgloss.NewStyle().
		Width(inputWidth).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	// Extension style (muted)
	extStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Bold(true)

	inputField := fieldStyle.Render(d.textInput.RenderWithCursor(inputWidth - 2))

	// Combine input field and extension
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		inputField,
		" ",
		extStyle.Render(d.extension),
	)
}

// Input returns the current input value (base name only).
func (d *ExtensionRenameDialog) Input() string {
	return d.textInput.Value
}

// SetInput sets the input value and positions cursor at the end.
func (d *ExtensionRenameDialog) SetInput(value string) {
	d.textInput.Value = value
	d.textInput.CursorPos = len([]rune(value))
	d.validateInput()
}

// CursorPos returns the current cursor position.
func (d *ExtensionRenameDialog) CursorPos() int {
	return d.textInput.CursorPos
}

// Extension returns the fixed extension.
func (d *ExtensionRenameDialog) Extension() string {
	return d.extension
}
