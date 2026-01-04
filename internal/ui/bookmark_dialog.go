package ui

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/sakura/duofm/internal/config"
)

// BookmarkDialog is a dialog for managing bookmarks.
type BookmarkDialog struct {
	BaseDialog
	bookmarks  []config.Bookmark
	cursor     int
	pathExists []bool // Cache for path existence checks
	styles     DialogStyles
}

// bookmarkJumpMsg is sent when user wants to jump to a bookmark.
type bookmarkJumpMsg struct {
	path string
}

// bookmarkDeleteMsg is sent when user wants to delete a bookmark.
type bookmarkDeleteMsg struct {
	index int
}

// bookmarkEditMsg is sent when user wants to edit a bookmark.
type bookmarkEditMsg struct {
	index    int
	bookmark config.Bookmark
}

// bookmarkCloseMsg is sent when dialog is closed without action.
type bookmarkCloseMsg struct{}

// NewBookmarkDialog creates a new bookmark dialog.
func NewBookmarkDialog(bookmarks []config.Bookmark) *BookmarkDialog {
	base := NewBaseDialog(DialogDisplayPane)
	base.SetWidth(60)
	d := &BookmarkDialog{
		BaseDialog: base,
		bookmarks:  bookmarks,
		cursor:     0,
		pathExists: make([]bool, len(bookmarks)),
		styles:     NewDialogStyles(60, ColorPrimary),
	}

	// Check path existence for each bookmark
	for i, b := range bookmarks {
		_, err := os.Stat(b.Path)
		d.pathExists[i] = err == nil
	}

	return d
}

// Update handles keyboard input.
func (d *BookmarkDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	if !d.IsActive() {
		return d, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			switch string(msg.Runes) {
			case "j":
				d.moveCursorDown()
				return d, nil
			case "k":
				d.moveCursorUp()
				return d, nil
			case "d":
				if len(d.bookmarks) > 0 {
					index := d.cursor
					d.Close()
					return d, func() tea.Msg {
						return bookmarkDeleteMsg{index: index}
					}
				}
				return d, nil
			case "e":
				if len(d.bookmarks) > 0 {
					index := d.cursor
					bookmark := d.bookmarks[index]
					d.Close()
					return d, func() tea.Msg {
						return bookmarkEditMsg{index: index, bookmark: bookmark}
					}
				}
				return d, nil
			}

		case tea.KeyDown:
			d.moveCursorDown()
			return d, nil

		case tea.KeyUp:
			d.moveCursorUp()
			return d, nil

		case tea.KeyEnter:
			if len(d.bookmarks) > 0 && d.cursor < len(d.pathExists) && d.pathExists[d.cursor] {
				path := d.bookmarks[d.cursor].Path
				d.Close()
				return d, func() tea.Msg {
					return bookmarkJumpMsg{path: path}
				}
			}
			return d, nil

		case tea.KeyEsc:
			d.Close()
			return d, func() tea.Msg {
				return bookmarkCloseMsg{}
			}
		}
	}

	return d, nil
}

func (d *BookmarkDialog) moveCursorDown() {
	if len(d.bookmarks) == 0 {
		return
	}
	d.cursor++
	if d.cursor >= len(d.bookmarks) {
		d.cursor = 0
	}
}

func (d *BookmarkDialog) moveCursorUp() {
	if len(d.bookmarks) == 0 {
		return
	}
	d.cursor--
	if d.cursor < 0 {
		d.cursor = len(d.bookmarks) - 1
	}
}

// View renders the dialog.
func (d *BookmarkDialog) View() string {
	if !d.IsActive() {
		return ""
	}

	var b strings.Builder
	width := d.Width()

	// Title
	b.WriteString(d.styles.Title.Render("Bookmarks"))
	b.WriteString("\n\n")

	// Bookmark list or empty message
	if len(d.bookmarks) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(string(ColorMuted)))
		b.WriteString(emptyStyle.Render("No bookmarks"))
		b.WriteString("\n")
	} else {
		for i, bookmark := range d.bookmarks {
			isSelected := i == d.cursor
			exists := i < len(d.pathExists) && d.pathExists[i]

			// Line 1: Alias with optional warning
			aliasLine := bookmark.Name
			if !exists {
				aliasLine = "\u26a0 " + aliasLine // Warning emoji
			}

			aliasStyle := lipgloss.NewStyle().Width(width - 6).Padding(0, 1)

			if isSelected {
				aliasStyle = aliasStyle.
					Background(lipgloss.Color(string(ColorPrimary))).
					Foreground(lipgloss.Color("0")).
					Bold(true)
			} else if !exists {
				aliasStyle = aliasStyle.Foreground(lipgloss.Color(string(ColorMuted)))
			}

			b.WriteString(aliasStyle.Render(aliasLine))
			b.WriteString("\n")

			// Line 2: Path (wrapped if needed)
			pathStyle := lipgloss.NewStyle().Width(width - 6).Padding(0, 1)

			if isSelected {
				pathStyle = pathStyle.
					Background(lipgloss.Color(string(ColorPrimary))).
					Foreground(lipgloss.Color("0"))
			} else {
				pathStyle = pathStyle.Foreground(lipgloss.Color("245"))
			}

			// Wrap path if too long
			pathDisplay := d.wrapPath(bookmark.Path, width-8)
			b.WriteString(pathStyle.Render(pathDisplay))
			b.WriteString("\n")

			// Add spacing between bookmarks
			if i < len(d.bookmarks)-1 {
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(d.styles.Footer.Render("j/k/↑/↓:Move  Enter:Jump  d:Delete  e:Edit  Esc:Close"))

	return d.styles.Box.Render(b.String())
}

// wrapPath wraps a path to fit within the specified width.
// Uses runewidth for proper handling of multibyte characters.
func (d *BookmarkDialog) wrapPath(path string, maxWidth int) string {
	if runewidth.StringWidth(path) <= maxWidth {
		return path
	}

	var lines []string
	var currentLine strings.Builder
	currentWidth := 0

	for _, r := range path {
		charWidth := runewidth.RuneWidth(r)
		if currentWidth+charWidth > maxWidth && currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentWidth = 0
		}
		currentLine.WriteRune(r)
		currentWidth += charWidth
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}
