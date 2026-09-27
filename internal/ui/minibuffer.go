package ui

import (
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rivo/uniseg"
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

// widthUnit is a grapheme cluster: a maximal run of runes that
// github.com/rivo/uniseg groups as a single user-perceived character (e.g.
// a ZWJ-joined emoji sequence, or a base rune followed by a combining mark
// or variation selector). start/end are rune offsets into the ORIGINAL rune
// slice the segment came from (see partitionUnits' offset parameter), so a
// unit's bounds can be sliced directly without further translation. width is
// the unit's own text measured with the width basis exactly once.
type widthUnit struct {
	start, end int
	width      int
}

// partitionUnits splits seg into grapheme clusters (width units) using
// github.com/rivo/uniseg's grapheme-cluster boundaries -- the same
// splitting the width basis (lipgloss, via charmbracelet/x/ansi) uses
// internally, so a unit boundary is always a boundary the width basis
// itself would place a cursor at; a joined sequence is therefore never
// split mid-cluster by the scroll logic below. offset is seg's starting
// position in the original rune slice the caller will index into, so the
// returned units' start/end are absolute indices there (sliceable directly,
// no further translation needed). Single pass over seg; each unit's own
// width is measured (via the width basis, on the unit's full text) exactly
// once -- combining sequences (e.g. a rune followed by U+FE0F, variation
// selector-16) can render wider than the sum of their individual runes'
// widths, so any estimate built from unit sums must still be verified (and
// shrunk if necessary) against the actually assembled string (see
// guardFit).
func partitionUnits(seg []rune, offset int) []widthUnit {
	if len(seg) == 0 {
		return nil
	}
	units := make([]widthUnit, 0, len(seg))
	str := string(seg)
	runePos := 0
	for len(str) > 0 {
		cluster, rest, _, _ := uniseg.FirstGraphemeClusterInString(str, -1)
		n := utf8.RuneCountInString(cluster)
		units = append(units, widthUnit{
			start: offset + runePos,
			end:   offset + runePos + n,
			width: lipgloss.Width(cluster),
		})
		runePos += n
		str = rest
	}
	return units
}

// unitIndexContaining returns the index into units of the unit whose rune
// range [start, end) contains pos. units must be partitionUnits' result for
// a rune slice that itself contains pos at some position -- i.e. units
// cover [0, N) contiguously and pos < N -- so the first unit whose end
// exceeds pos is guaranteed to also satisfy start <= pos. Single pass,
// bounded by len(units).
func unitIndexContaining(units []widthUnit, pos int) int {
	for i, u := range units {
		if pos < u.end {
			return i
		}
	}
	return len(units) - 1
}

// unitsTotalWidth sums every unit's estimated width. Single pass.
func unitsTotalWidth(units []widthUnit) int {
	total := 0
	for _, u := range units {
		total += u.width
	}
	return total
}

// unitsPrefixCount returns how many units, counted from the front of units,
// have estimated widths summing to at most budget: the kept sum stays
// within budget, and keeping the next unit would exceed it. Single pass.
func unitsPrefixCount(units []widthUnit, budget int) int {
	sum := 0
	n := 0
	for n < len(units) {
		next := sum + units[n].width
		if next > budget {
			break
		}
		sum = next
		n++
	}
	return n
}

// unitsSuffixCount is unitsPrefixCount's mirror, counting from the back.
func unitsSuffixCount(units []widthUnit, budget int) int {
	sum := 0
	n := 0
	total := len(units)
	for n < total {
		next := sum + units[total-1-n].width
		if next > budget {
			break
		}
		sum = next
		n++
	}
	return n
}

// trimUnitsFront removes whole units from the front of units, in one step,
// until the removed total is at least excess or units is exhausted.
func trimUnitsFront(units []widthUnit, excess int) []widthUnit {
	removed := 0
	i := 0
	for i < len(units) && removed < excess {
		removed += units[i].width
		i++
	}
	return units[i:]
}

// trimUnitsBack mirrors trimUnitsFront, trimming from the end.
func trimUnitsBack(units []widthUnit, excess int) []widthUnit {
	removed := 0
	i := len(units)
	for i > 0 && removed < excess {
		i--
		removed += units[i].width
	}
	return units[:i]
}

// guardFit is CD4's bounded "estimate, then verify" guard. It measures
// render()'s result with the width basis and, while that exceeds budget,
// calls shrink(excess) to remove whole units from the side away from the
// cursor (or the trailing side, for the prompt-truncation path); shrink
// reports whether it was able to remove anything more. At most 3
// re-measurements after the first (4 lipgloss.Width calls total) --
// independent of how long render()'s underlying content is, since shrink
// operates on whole units, not individual runes.
func guardFit(budget int, render func() string, shrink func(excess int) bool) (result string, fits bool) {
	for tries := 0; ; tries++ {
		result = render()
		w := lipgloss.Width(result)
		if w <= budget {
			return result, true
		}
		if tries >= 3 || !shrink(w-budget) {
			return result, false
		}
	}
}

// buildLine renders the given rune slice with the cursor (at curPos,
// relative to the start of dr) highlighted, including the trailing
// block cursor when curPos sits just past the last rune.
func buildLine(dr []rune, curPos int) string {
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

// View renders the minibuffer
func (m *Minibuffer) View() string {
	if !m.visible {
		return ""
	}

	// contentWidth is the content width of the minibuffer style below
	// (style Width m.width-2, minus 1 column of padding on each side).
	// Widths 0-4 collapse it to <= 0: only "no panic, no line break" is
	// guaranteed there (CD5), so no prompt or input content is rendered.
	contentWidth := m.width - 4

	runes := []rune(m.input)
	promptRunes := []rune(m.prompt)
	promptWidth := lipgloss.Width(m.prompt)

	var promptToRender string
	var displayStart, displayEnd int // rune range into `runes`
	cursorDisplayPos := -1           // rune-relative index into runes[displayStart:displayEnd]; -1 = no highlight

	switch {
	case contentWidth <= 0:
		// CD5: nothing meaningful to show; only avoid panic/line-break.
		promptToRender = ""
		displayStart, displayEnd = 0, 0

	case promptWidth >= contentWidth:
		// FR3: remaining width <= 0. Truncate the prompt to the content
		// width; input and the cursor block are not rendered. Partition the
		// prompt into width units (grapheme clusters) in one pass, keep the
		// longest unit prefix whose estimated width is at most contentWidth,
		// then apply the guard (CD4): the estimate can undershoot the
		// actual rendered width for sequences joined across positive-width
		// runes, so the kept prefix is re-measured and, if necessary,
		// shrunk by whole trailing units; it finally falls back to an
		// empty prompt.
		units := partitionUnits(promptRunes, 0)
		keep := unitsPrefixCount(units, contentWidth)

		render := func() string {
			end := 0
			if keep > 0 {
				end = units[keep-1].end
			}
			return string(promptRunes[:end])
		}
		shrink := func(excess int) bool {
			if keep == 0 {
				return false
			}
			removed := 0
			for keep > 0 && removed < excess {
				keep--
				removed += units[keep].width
			}
			return true
		}

		text, fits := guardFit(contentWidth, render, shrink)
		if !fits {
			text = ""
		}
		promptToRender = text
		displayStart, displayEnd = 0, 0

	default:
		// FR4: remaining width >= 1.
		promptToRender = m.prompt
		remaining := contentWidth - promptWidth
		cursorAtEnd := m.cursorPos >= len(runes)

		// Step 1: fit check. Assemble the full line (prompt + whole input
		// with the cursor highlight, or the cursor block when the cursor is
		// at the end) once and measure it once; render untruncated if it
		// fits, instead of falling back to the windowed paths below.
		fullLine := promptToRender + buildLine(runes, m.cursorPos)
		if lipgloss.Width(fullLine) <= contentWidth {
			displayStart, displayEnd = 0, len(runes)
			cursorDisplayPos = m.cursorPos
			break
		}

		// Step 2: partition the whole input into grapheme clusters (units)
		// in one pass. Every window below is expressed in terms of this
		// single list -- when the cursor sits mid-input, the before- and
		// after-cursor slices are cut FROM this list, never re-partitioned,
		// so every boundary picked below is one of these units' boundaries
		// (FR1, FR2).
		units := partitionUnits(runes, 0)

		switch {
		case cursorAtEnd:
			// Step 3: scroll so the cursor block (1 column) sits at the
			// right edge. Keep the longest unit suffix whose estimated
			// width is at most remaining-1.
			n := unitsSuffixCount(units, remaining-1)
			kept := units[len(units)-n:]

			displayEnd = len(runes)
			if len(kept) > 0 {
				displayStart = kept[0].start
			} else {
				displayStart = len(runes)
			}
			cursorDisplayPos = displayEnd - displayStart

			render := func() string {
				return promptToRender + buildLine(runes[displayStart:displayEnd], cursorDisplayPos)
			}
			shrink := func(excess int) bool {
				if len(kept) == 0 {
					return false
				}
				kept = trimUnitsFront(kept, excess)
				if len(kept) > 0 {
					displayStart = kept[0].start
				} else {
					displayStart = displayEnd
				}
				cursorDisplayPos = displayEnd - displayStart
				return true
			}

			if _, fits := guardFit(contentWidth, render, shrink); !fits {
				// CD4: even the cursor block alone overflows.
				displayStart, displayEnd = len(runes), len(runes)
				cursorDisplayPos = -1
			}

		default:
			// Steps 4-6: locate the cursor's own grapheme (unit) in the
			// shared units list and split that list around it -- the
			// cursor's grapheme is excluded from both the before- and
			// after-cursor slices, so it is never itself trimmed away by
			// guardFit below.
			cIdx := unitIndexContaining(units, m.cursorPos)
			cursorUnit := units[cIdx]
			cursorWidth := cursorUnit.width

			switch {
			case cursorWidth > remaining:
				// Step 4: the cursor's own grapheme does not fit within the
				// remaining width even alone; do not render it.
				displayStart, displayEnd = cursorUnit.start, cursorUnit.start

			default:
				beforeUnits := units[:cIdx]
				afterUnits := units[cIdx+1:]
				beforeWidth := unitsTotalWidth(beforeUnits)

				if beforeWidth+cursorWidth <= remaining {
					// Step 5: the whole before-cursor segment plus the
					// cursor's grapheme fit; extend right over the longest
					// unit prefix of the after-cursor segment that fits the
					// leftover width.
					leftover := remaining - beforeWidth - cursorWidth
					nAfter := unitsPrefixCount(afterUnits, leftover)
					afterKept := afterUnits[:nAfter]
					beforeKept := beforeUnits

					displayStart = 0
					cursorDisplayPos = m.cursorPos
					if nAfter > 0 {
						displayEnd = afterKept[nAfter-1].end
					} else {
						displayEnd = cursorUnit.end
					}

					render := func() string {
						return promptToRender + buildLine(runes[displayStart:displayEnd], cursorDisplayPos)
					}
					// CD4: the side away from the cursor is tried right
					// first (the greedily-extended after-cursor part), then
					// left.
					shrink := func(excess int) bool {
						if len(afterKept) > 0 {
							afterKept = trimUnitsBack(afterKept, excess)
							if len(afterKept) > 0 {
								displayEnd = afterKept[len(afterKept)-1].end
							} else {
								displayEnd = cursorUnit.end
							}
							return true
						}
						if len(beforeKept) > 0 {
							beforeKept = trimUnitsFront(beforeKept, excess)
							if len(beforeKept) > 0 {
								displayStart = beforeKept[0].start
							} else {
								displayStart = cursorUnit.start
							}
							cursorDisplayPos = m.cursorPos - displayStart
							return true
						}
						return false
					}

					if _, fits := guardFit(contentWidth, render, shrink); !fits {
						displayStart, displayEnd = cursorUnit.start, cursorUnit.start
						cursorDisplayPos = -1
					}
				} else {
					// Step 6: scroll so the cursor's grapheme sits at the
					// right edge of the window; start at the longest unit
					// suffix of the before-cursor segment that fits.
					n := unitsSuffixCount(beforeUnits, remaining-cursorWidth)
					kept := beforeUnits[len(beforeUnits)-n:]

					displayEnd = cursorUnit.end
					if len(kept) > 0 {
						displayStart = kept[0].start
					} else {
						displayStart = cursorUnit.start
					}
					cursorDisplayPos = m.cursorPos - displayStart

					render := func() string {
						return promptToRender + buildLine(runes[displayStart:displayEnd], cursorDisplayPos)
					}
					shrink := func(excess int) bool {
						if len(kept) == 0 {
							return false
						}
						kept = trimUnitsFront(kept, excess)
						if len(kept) > 0 {
							displayStart = kept[0].start
						} else {
							displayStart = cursorUnit.start
						}
						cursorDisplayPos = m.cursorPos - displayStart
						return true
					}

					if _, fits := guardFit(contentWidth, render, shrink); !fits {
						displayStart, displayEnd = cursorUnit.start, cursorUnit.start
						cursorDisplayPos = -1
					}
				}
			}
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
