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
	title         string
	input         string
	cursorPos     int
	active        bool
	width         int
	destPath      string          // Destination directory
	srcPath       string          // Source file path
	operation     string          // "copy" or "move"
	existingFiles map[string]bool // Cached filenames in dest
	hasError      bool
	errorMessage  string
	suggestedName string
}

// renameInputResultMsg is sent when the rename dialog is confirmed
type renameInputResultMsg struct {
	newName   string
	srcPath   string
	destPath  string
	operation string
}

// NewRenameInputDialog creates a new rename input dialog
func NewRenameInputDialog(destPath, srcPath, operation string) *RenameInputDialog {
	// Read destination directory
	existingFiles := loadExistingFiles(destPath)

	// Generate suggested name
	filename := filepath.Base(srcPath)
	suggested := suggestRename(filename, existingFiles)

	return &RenameInputDialog{
		title:         "New name:",
		input:         suggested,
		cursorPos:     len(suggested),
		active:        true,
		width:         50,
		destPath:      destPath,
		srcPath:       srcPath,
		operation:     operation,
		existingFiles: existingFiles,
		hasError:      false,
		suggestedName: suggested,
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
	if d.input == "" {
		d.hasError = true
		d.errorMessage = "File name cannot be empty"
		return
	}

	if d.existingFiles[d.input] {
		d.hasError = true
		d.errorMessage = "File already exists"
		return
	}

	if err := fs.ValidateFilename(d.input); err != nil {
		d.hasError = true
		d.errorMessage = err.Error()
		return
	}

	d.hasError = false
	d.errorMessage = ""
}

// Update handles keyboard input
func (d *RenameInputDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.active {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if d.hasError {
				return d, nil // Do nothing if error
			}
			d.active = false
			return d, func() tea.Msg {
				return renameInputResultMsg{
					newName:   d.input,
					srcPath:   d.srcPath,
					destPath:  d.destPath,
					operation: d.operation,
				}
			}

		case tea.KeyEsc:
			d.active = false
			return d, nil

		case tea.KeyRunes:
			// 文字入力
			runes := []rune(d.input)
			newRunes := make([]rune, 0, len(runes)+len(msg.Runes))
			newRunes = append(newRunes, runes[:d.cursorPos]...)
			newRunes = append(newRunes, msg.Runes...)
			newRunes = append(newRunes, runes[d.cursorPos:]...)
			d.input = string(newRunes)
			d.cursorPos += len(msg.Runes)
			d.validateInput()
			return d, nil

		case tea.KeyBackspace:
			if d.cursorPos > 0 {
				runes := []rune(d.input)
				newRunes := make([]rune, 0, len(runes)-1)
				newRunes = append(newRunes, runes[:d.cursorPos-1]...)
				newRunes = append(newRunes, runes[d.cursorPos:]...)
				d.input = string(newRunes)
				d.cursorPos--
				d.validateInput()
			}
			return d, nil

		case tea.KeyDelete:
			runes := []rune(d.input)
			if d.cursorPos < len(runes) {
				newRunes := make([]rune, 0, len(runes)-1)
				newRunes = append(newRunes, runes[:d.cursorPos]...)
				newRunes = append(newRunes, runes[d.cursorPos+1:]...)
				d.input = string(newRunes)
				d.validateInput()
			}
			return d, nil

		case tea.KeyLeft:
			if d.cursorPos > 0 {
				d.cursorPos--
			}
			return d, nil

		case tea.KeyRight:
			if d.cursorPos < len([]rune(d.input)) {
				d.cursorPos++
			}
			return d, nil

		case tea.KeyCtrlA:
			d.cursorPos = 0
			return d, nil

		case tea.KeyCtrlE:
			d.cursorPos = len([]rune(d.input))
			return d, nil

		case tea.KeyCtrlU:
			runes := []rune(d.input)
			d.input = string(runes[d.cursorPos:])
			d.cursorPos = 0
			d.validateInput()
			return d, nil

		case tea.KeyCtrlK:
			runes := []rune(d.input)
			d.input = string(runes[:d.cursorPos])
			d.validateInput()
			return d, nil
		}
	}

	return d, nil
}

// View renders the dialog
func (d *RenameInputDialog) View() string {
	if !d.active {
		return ""
	}

	var b strings.Builder
	width := d.width

	// タイトル
	titleStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1).
		Bold(true).
		Foreground(lipgloss.Color("39"))
	b.WriteString(titleStyle.Render(d.title))
	b.WriteString("\n\n")

	// 入力フィールド
	inputWidth := width - 8
	b.WriteString(d.renderInputField(inputWidth))
	b.WriteString("\n")

	// エラーメッセージ（あれば）
	if d.hasError {
		errorStyle := lipgloss.NewStyle().
			Width(width-4).
			Padding(0, 1).
			Foreground(lipgloss.Color("196"))
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(d.errorMessage))
	}

	b.WriteString("\n")

	// フッター
	footerStyle := lipgloss.NewStyle().
		Width(width-4).
		Padding(0, 1).
		Foreground(lipgloss.Color("240"))

	if d.hasError {
		b.WriteString(footerStyle.Render("Esc: Cancel"))
	} else {
		b.WriteString(footerStyle.Render("Enter: Confirm  Esc: Cancel"))
	}

	// ボーダーで囲む
	boxStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	return boxStyle.Render(b.String())
}

// renderInputField は入力フィールドをレンダリング
func (d *RenameInputDialog) renderInputField(width int) string {
	runes := []rune(d.input)
	displayInput := d.input

	// 表示可能な範囲を計算
	cursorDisplayPos := d.cursorPos
	startPos := 0

	if len(runes) > width-2 {
		if d.cursorPos > width-3 {
			startPos = d.cursorPos - width + 3
		}
		endPos := startPos + width - 2
		if endPos > len(runes) {
			endPos = len(runes)
		}
		displayInput = string(runes[startPos:endPos])
		cursorDisplayPos = d.cursorPos - startPos
	}

	// カーソル付きで表示文字列を構築
	displayRunes := []rune(displayInput)
	var result strings.Builder
	for i, r := range displayRunes {
		if i == cursorDisplayPos {
			result.WriteString(lipgloss.NewStyle().Reverse(true).Render(string(r)))
		} else {
			result.WriteRune(r)
		}
	}
	if cursorDisplayPos >= len(displayRunes) {
		result.WriteString(lipgloss.NewStyle().Reverse(true).Render(" "))
	}

	// 入力フィールドのスタイル
	fieldStyle := lipgloss.NewStyle().
		Width(width).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	return fieldStyle.Render(result.String())
}

// IsActive returns whether the dialog is active
func (d *RenameInputDialog) IsActive() bool {
	return d.active
}

// DisplayType returns the dialog display type
func (d *RenameInputDialog) DisplayType() DialogDisplayType {
	return DialogDisplayPane
}
