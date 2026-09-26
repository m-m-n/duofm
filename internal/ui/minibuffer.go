package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Minibuffer is a single-line text input area displayed at the bottom of the pane
type Minibuffer struct {
	prompt    string
	input     string
	cursorPos int
	visible   bool
	width     int
}

// NewMinibuffer creates a new minibuffer instance
func NewMinibuffer() *Minibuffer {
	return &Minibuffer{
		prompt:    "",
		input:     "",
		cursorPos: 0,
		visible:   false,
		width:     80,
	}
}

// SetPrompt sets the prompt text (e.g., "/: " or "(search): ")
func (m *Minibuffer) SetPrompt(prompt string) {
	m.prompt = prompt
}

// Prompt returns the current prompt text
func (m *Minibuffer) Prompt() string {
	return m.prompt
}

// SetWidth sets the display width
func (m *Minibuffer) SetWidth(width int) {
	m.width = width
}

// Clear resets the input and cursor position
func (m *Minibuffer) Clear() {
	m.input = ""
	m.cursorPos = 0
}

// Show makes the minibuffer visible
func (m *Minibuffer) Show() {
	m.visible = true
}

// Hide makes the minibuffer invisible
func (m *Minibuffer) Hide() {
	m.visible = false
}

// Input returns the current input text
func (m *Minibuffer) Input() string {
	return m.input
}

// SetInput sets the input text from external source and moves cursor to end.
// This is used for history search to display selected commands.
func (m *Minibuffer) SetInput(input string) {
	m.input = input
	m.cursorPos = len([]rune(input))
}

// IsVisible returns whether the minibuffer is visible
func (m *Minibuffer) IsVisible() bool {
	return m.visible
}

// CursorPos returns the current cursor position
func (m *Minibuffer) CursorPos() int {
	return m.cursorPos
}

// SetCursorPos sets the cursor position within valid bounds
func (m *Minibuffer) SetCursorPos(pos int) {
	runes := []rune(m.input)
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}
	m.cursorPos = pos
}

// HandleKey processes a key message and returns true if handled
func (m *Minibuffer) HandleKey(msg tea.KeyMsg) bool {
	if !m.visible {
		return false
	}

	switch msg.Type {
	case tea.KeyRunes:
		// Insert character at cursor position
		runes := []rune(m.input)
		newRunes := make([]rune, 0, len(runes)+len(msg.Runes))
		newRunes = append(newRunes, runes[:m.cursorPos]...)
		newRunes = append(newRunes, msg.Runes...)
		newRunes = append(newRunes, runes[m.cursorPos:]...)
		m.input = string(newRunes)
		m.cursorPos += len(msg.Runes)
		return true

	case tea.KeySpace:
		// Insert space at cursor position
		runes := []rune(m.input)
		newRunes := make([]rune, 0, len(runes)+1)
		newRunes = append(newRunes, runes[:m.cursorPos]...)
		newRunes = append(newRunes, ' ')
		newRunes = append(newRunes, runes[m.cursorPos:]...)
		m.input = string(newRunes)
		m.cursorPos++
		return true

	case tea.KeyBackspace:
		if m.cursorPos > 0 {
			runes := []rune(m.input)
			newRunes := make([]rune, 0, len(runes)-1)
			newRunes = append(newRunes, runes[:m.cursorPos-1]...)
			newRunes = append(newRunes, runes[m.cursorPos:]...)
			m.input = string(newRunes)
			m.cursorPos--
		}
		return true

	case tea.KeyDelete:
		runes := []rune(m.input)
		if m.cursorPos < len(runes) {
			newRunes := make([]rune, 0, len(runes)-1)
			newRunes = append(newRunes, runes[:m.cursorPos]...)
			newRunes = append(newRunes, runes[m.cursorPos+1:]...)
			m.input = string(newRunes)
		}
		return true

	case tea.KeyLeft:
		if m.cursorPos > 0 {
			m.cursorPos--
		}
		return true

	case tea.KeyRight:
		if m.cursorPos < len([]rune(m.input)) {
			m.cursorPos++
		}
		return true

	case tea.KeyCtrlA:
		// Move to beginning
		m.cursorPos = 0
		return true

	case tea.KeyCtrlE:
		// Move to end
		m.cursorPos = len([]rune(m.input))
		return true

	case tea.KeyCtrlK:
		// Kill to end of line
		runes := []rune(m.input)
		m.input = string(runes[:m.cursorPos])
		return true

	case tea.KeyCtrlU:
		// Kill to beginning of line
		runes := []rune(m.input)
		m.input = string(runes[m.cursorPos:])
		m.cursorPos = 0
		return true

	case tea.KeyCtrlB:
		// Move backward (same as left)
		if m.cursorPos > 0 {
			m.cursorPos--
		}
		return true

	case tea.KeyCtrlF:
		// Move forward (same as right)
		// Note: This conflicts with regex search key, so it's only active when minibuffer is visible
		if m.cursorPos < len([]rune(m.input)) {
			m.cursorPos++
		}
		return true
	}

	return false
}

// runeDisplayWidths returns, for each rune in runes, that rune's own display
// width as measured by lipgloss. This is used only to pick an initial
// candidate window; it is NOT a reliable upper bound on the rendered width
// of the resulting grapheme clusters. Some combining sequences (e.g. a
// rune followed by U+FE0F, variation selector-16) render wider than the sum
// of their individual rune widths, so any window chosen from this per-rune
// budget must still be verified (and shrunk if necessary) by measuring the
// actually assembled string with lipgloss.Width (see the shrink loop in
// View).
func runeDisplayWidths(runes []rune) []int {
	widths := make([]int, len(runes))
	for i, r := range runes {
		widths[i] = lipgloss.Width(string(r))
	}
	return widths
}

// widthPrefixSums returns cumulative sums of widths, with a leading 0, so
// that prefix[i] is the display width of the first i runes.
func widthPrefixSums(widths []int) []int {
	prefix := make([]int, len(widths)+1)
	for i, w := range widths {
		prefix[i+1] = prefix[i] + w
	}
	return prefix
}

// maxRunesWithinWidth returns the largest n such that the first n runes
// (per prefix, a widthPrefixSums result) fit within budget. A rune that
// would straddle the boundary (its addition would push the cumulative width
// past budget) is excluded entirely, never partially rendered (D3).
func maxRunesWithinWidth(prefix []int, budget int) int {
	n := 0
	for n < len(prefix)-1 && prefix[n+1] <= budget {
		n++
	}
	return n
}

// minStartWithinWidth returns the smallest start such that the runes in
// [start, end) fit within budget, given prefix (a widthPrefixSums result).
// Like maxRunesWithinWidth, a straddling rune at the left edge is excluded
// entirely (D3).
func minStartWithinWidth(prefix []int, end, budget int) int {
	target := prefix[end] - budget
	start := 0
	for start < end && prefix[start] < target {
		start++
	}
	return start
}

// View renders the minibuffer
func (m *Minibuffer) View() string {
	if !m.visible {
		return ""
	}

	// contentWidth is the content width of the minibuffer style below
	// (style Width m.width-2, minus 1 column of padding on each side).
	// Widths 0-4 collapse it to <= 0: only "no panic, no line break" is
	// guaranteed there (D4), so no prompt or input content is rendered.
	contentWidth := m.width - 4

	runes := []rune(m.input)
	runeWidths := runeDisplayWidths(runes)
	prefix := widthPrefixSums(runeWidths)

	promptRunes := []rune(m.prompt)
	promptWidth := lipgloss.Width(m.prompt)

	var promptToRender string
	var displayStart, displayEnd int // rune range into `runes`
	cursorDisplayPos := -1           // rune-relative index into runes[displayStart:displayEnd]; -1 = no highlight

	switch {
	case contentWidth <= 0:
		// D4: nothing meaningful to show; only avoid panic/line-break.
		promptToRender = ""
		displayStart, displayEnd = 0, 0

	case promptWidth >= contentWidth:
		// FR3: remaining width <= 0. Truncate the prompt to the content
		// width; input and the cursor block are not rendered. The
		// per-rune budget only picks a starting candidate: combining
		// sequences (e.g. a rune followed by VS16) can render wider than
		// the sum of their individual rune widths, so shrink further, one
		// trailing rune at a time, using the actually rendered width.
		promptRuneWidths := runeDisplayWidths(promptRunes)
		promptPrefix := widthPrefixSums(promptRuneWidths)
		end := maxRunesWithinWidth(promptPrefix, contentWidth)
		for end > 0 && lipgloss.Width(string(promptRunes[:end])) > contentWidth {
			end--
		}
		promptToRender = string(promptRunes[:end])
		displayStart, displayEnd = 0, 0

	default:
		// FR4: remaining width >= 1.
		promptToRender = m.prompt
		remaining := contentWidth - promptWidth

		cursorAtEnd := m.cursorPos >= len(runes)
		totalWidth := prefix[len(runes)]
		totalWithCursor := totalWidth
		if cursorAtEnd {
			totalWithCursor++
		}

		switch {
		case totalWithCursor <= remaining:
			// FR8: input and cursor block both fit; show untruncated.
			displayStart, displayEnd = 0, len(runes)
			cursorDisplayPos = m.cursorPos

		case cursorAtEnd:
			// Scroll so the cursor block (1 column) sits at the right
			// edge of the window.
			start := minStartWithinWidth(prefix, len(runes), remaining-1)
			displayStart, displayEnd = start, len(runes)
			cursorDisplayPos = len(runes) - start

		case runeWidths[m.cursorPos] > remaining:
			// D3: the cursor's own character does not fit within the
			// remaining width even alone (e.g. remaining=1, a full-width
			// character at the cursor); do not render it.
			displayStart, displayEnd = m.cursorPos, m.cursorPos

		case prefix[m.cursorPos+1] <= remaining:
			// Cursor fits in the window starting from position 0; show
			// the maximal prefix that fits (may extend past the cursor).
			end := maxRunesWithinWidth(prefix, remaining)
			displayStart, displayEnd = 0, end
			cursorDisplayPos = m.cursorPos

		default:
			// Scroll so the cursor's character sits at the right edge of
			// the window.
			start := minStartWithinWidth(prefix, m.cursorPos+1, remaining)
			displayStart, displayEnd = start, m.cursorPos+1
			cursorDisplayPos = m.cursorPos - start
		}
	}

	// buildLine renders the given rune slice with the cursor (at curPos,
	// relative to the start of dr) highlighted, including the trailing
	// block cursor when curPos sits just past the last rune.
	buildLine := func(dr []rune, curPos int) string {
		var b strings.Builder
		for i, r := range dr {
			if i == curPos {
				b.WriteString(lipgloss.NewStyle().Reverse(true).Render(string(r)))
			} else {
				b.WriteRune(r)
			}
		}
		if curPos >= 0 && curPos >= len(dr) {
			b.WriteString(lipgloss.NewStyle().Reverse(true).Render(" "))
		}
		return b.String()
	}

	// D2: the per-rune budget above only picks a starting candidate
	// window; it is not a reliable bound on the rendered width, since
	// combining sequences (e.g. a rune followed by VS16) can render wider
	// than the sum of their individual rune widths. Shrink the window,
	// one whole rune at a time from the side farther from the cursor (so
	// the cursor stays visible), using the actually rendered width, until
	// the assembled line fits contentWidth.
	for contentWidth > 0 && displayEnd > displayStart {
		dr := runes[displayStart:displayEnd]
		line := promptToRender + buildLine(dr, cursorDisplayPos)
		if lipgloss.Width(line) <= contentWidth {
			break
		}
		leftDist := cursorDisplayPos
		rightDist := (displayEnd - displayStart) - cursorDisplayPos
		if rightDist >= leftDist {
			displayEnd--
		} else {
			displayStart++
			cursorDisplayPos--
		}
	}

	displayRunes := runes[displayStart:displayEnd]
	contentLine := promptToRender + buildLine(displayRunes, cursorDisplayPos)

	// Style the whole minibuffer line
	style := lipgloss.NewStyle().
		Width(m.width-2).
		Padding(0, 1).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("236"))

	return style.Render(contentLine)
}
