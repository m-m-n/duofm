package ui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestNewMinibuffer(t *testing.T) {
	mb := NewMinibuffer()
	if mb == nil {
		t.Fatal("NewMinibuffer() returned nil")
	}
	if mb.visible {
		t.Error("New minibuffer should not be visible")
	}
	if mb.input != "" {
		t.Error("New minibuffer should have empty input")
	}
	if mb.cursorPos != 0 {
		t.Error("New minibuffer should have cursor at 0")
	}
}

func TestMinibufferShowHide(t *testing.T) {
	mb := NewMinibuffer()

	mb.Show()
	if !mb.IsVisible() {
		t.Error("Minibuffer should be visible after Show()")
	}

	mb.Hide()
	if mb.IsVisible() {
		t.Error("Minibuffer should not be visible after Hide()")
	}
}

func TestMinibufferSetPrompt(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("/: ")
	if mb.prompt != "/: " {
		t.Errorf("prompt = %q, want %q", mb.prompt, "/: ")
	}
}

func TestMinibufferClear(t *testing.T) {
	mb := NewMinibuffer()
	mb.input = "test"
	mb.cursorPos = 4

	mb.Clear()

	if mb.input != "" {
		t.Error("Clear() should empty input")
	}
	if mb.cursorPos != 0 {
		t.Error("Clear() should reset cursor to 0")
	}
}

func TestMinibufferInput(t *testing.T) {
	mb := NewMinibuffer()
	mb.input = "test input"
	if mb.Input() != "test input" {
		t.Errorf("Input() = %q, want %q", mb.Input(), "test input")
	}
}

func TestMinibufferHandleKeyCharacterInput(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()

	// Type "abc"
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if mb.input != "abc" {
		t.Errorf("input = %q, want %q", mb.input, "abc")
	}
	if mb.cursorPos != 3 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 3)
	}
}

func TestMinibufferHandleKeyBackspace(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 3

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})

	if mb.input != "ab" {
		t.Errorf("input = %q, want %q", mb.input, "ab")
	}
	if mb.cursorPos != 2 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 2)
	}

	// Backspace at beginning should do nothing
	mb.cursorPos = 0
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})
	if mb.input != "ab" {
		t.Errorf("input should not change when backspace at beginning")
	}
}

func TestMinibufferHandleKeyDelete(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 1

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyDelete})

	if mb.input != "ac" {
		t.Errorf("input = %q, want %q", mb.input, "ac")
	}
	if mb.cursorPos != 1 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 1)
	}

	// Delete at end should do nothing
	mb.cursorPos = 2
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyDelete})
	if mb.input != "ac" {
		t.Errorf("input should not change when delete at end")
	}
}

func TestMinibufferHandleKeyCursorMovement(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 1

	// Left arrow
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if mb.cursorPos != 0 {
		t.Errorf("cursorPos = %d, want %d after left", mb.cursorPos, 0)
	}

	// Left at beginning should stay at 0
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if mb.cursorPos != 0 {
		t.Errorf("cursorPos = %d, want %d after left at beginning", mb.cursorPos, 0)
	}

	// Right arrow
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRight})
	if mb.cursorPos != 1 {
		t.Errorf("cursorPos = %d, want %d after right", mb.cursorPos, 1)
	}

	// Right at end should stay at end
	mb.cursorPos = 3
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRight})
	if mb.cursorPos != 3 {
		t.Errorf("cursorPos = %d, want %d after right at end", mb.cursorPos, 3)
	}
}

func TestMinibufferHandleKeyCtrlA(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 3

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlA})
	if mb.cursorPos != 0 {
		t.Errorf("cursorPos = %d, want %d after Ctrl+A", mb.cursorPos, 0)
	}
}

func TestMinibufferHandleKeyCtrlE(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abc"
	mb.cursorPos = 0

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlE})
	if mb.cursorPos != 3 {
		t.Errorf("cursorPos = %d, want %d after Ctrl+E", mb.cursorPos, 3)
	}
}

func TestMinibufferHandleKeyCtrlK(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abcdef"
	mb.cursorPos = 3

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlK})
	if mb.input != "abc" {
		t.Errorf("input = %q, want %q after Ctrl+K", mb.input, "abc")
	}
	if mb.cursorPos != 3 {
		t.Errorf("cursorPos = %d, want %d after Ctrl+K", mb.cursorPos, 3)
	}
}

func TestMinibufferHandleKeyCtrlU(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "abcdef"
	mb.cursorPos = 3

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlU})
	if mb.input != "def" {
		t.Errorf("input = %q, want %q after Ctrl+U", mb.input, "def")
	}
	if mb.cursorPos != 0 {
		t.Errorf("cursorPos = %d, want %d after Ctrl+U", mb.cursorPos, 0)
	}
}

func TestMinibufferHandleKeyInsertAtMiddle(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "ac"
	mb.cursorPos = 1

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})

	if mb.input != "abc" {
		t.Errorf("input = %q, want %q", mb.input, "abc")
	}
	if mb.cursorPos != 2 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 2)
	}
}

func TestMinibufferHandleKeySpace(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()

	// Type "ls" then space then "-la"
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	mb.HandleKey(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if mb.input != "ls -la" {
		t.Errorf("input = %q, want %q", mb.input, "ls -la")
	}
	if mb.cursorPos != 6 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 6)
	}
}

func TestMinibufferHandleKeySpaceAtMiddle(t *testing.T) {
	mb := NewMinibuffer()
	mb.Show()
	mb.input = "lsla"
	mb.cursorPos = 2

	// Insert space in the middle
	mb.HandleKey(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune{' '}})

	if mb.input != "ls la" {
		t.Errorf("input = %q, want %q", mb.input, "ls la")
	}
	if mb.cursorPos != 3 {
		t.Errorf("cursorPos = %d, want %d", mb.cursorPos, 3)
	}
}

func TestMinibufferView(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("/: ")
	mb.SetWidth(40)
	mb.input = "test"
	mb.cursorPos = 4
	mb.Show()

	view := mb.View()

	// View should contain prompt and input
	if !strings.Contains(view, "/: ") {
		t.Error("View should contain prompt")
	}
	if !strings.Contains(view, "test") {
		t.Error("View should contain input")
	}
}

func TestMinibufferViewHidden(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("/: ")
	mb.input = "test"

	view := mb.View()
	if view != "" {
		t.Errorf("Hidden minibuffer View() = %q, want empty string", view)
	}
}

// --- AC-4 (FR6, FR5): Minibuffer.View at narrow widths (input display width <= 0) ---

func TestMinibufferView_NarrowWidths_NoPanicNoLineBreak(t *testing.T) {
	widths := []int{0, 1, 2, 3, 4}
	prompts := []string{"/: ", "!: "}
	inputs := []string{"", "a", "abc"}

	for _, width := range widths {
		for _, prompt := range prompts {
			for _, input := range inputs {
				runeCount := len([]rune(input))
				cursors := []struct {
					label string
					pos   int
				}{
					{"start", 0},
					{"half", runeCount / 2},
					{"end", runeCount},
				}

				for _, c := range cursors {
					name := fmt.Sprintf("width=%d/prompt=%q/input=%q/cursor=%s", width, prompt, input, c.label)
					t.Run(name, func(t *testing.T) {
						mb := NewMinibuffer()
						mb.SetPrompt(prompt)
						mb.SetWidth(width)
						mb.SetInput(input)
						mb.SetCursorPos(c.pos)
						mb.Show()

						result := mustNotPanic(t, "Minibuffer.View", mb.View)
						if strings.ContainsAny(result, "\n\r") {
							t.Errorf("result contains a line break: %q", result)
						}
					})
				}
			}
		}
	}
}

func TestMinibufferView_ReverseISearchPrompt_NarrowWidth_NoPanicNoLineBreak(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("(reverse-i-search)'': ")
	mb.SetWidth(24)
	mb.SetInput("a")
	mb.SetCursorPos(1)
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	if strings.ContainsAny(result, "\n\r") {
		t.Errorf("result contains a line break: %q", result)
	}
}

// --- task0001 shared helpers: width-based View() assertions ---

// assertBoundedSingleLine fails the test unless result contains no line
// break and its display width (measured via lipgloss, the single basis for
// width per Conventions/D1) is at most width-2, the minibuffer style's
// content width. width is m.width, not the content width itself.
func assertBoundedSingleLine(t *testing.T, width int, result string) {
	t.Helper()
	if strings.ContainsAny(result, "\n\r") {
		t.Errorf("width=%d: result contains a line break: %q", width, result)
	}
	if w := lipgloss.Width(result); w > width-2 {
		t.Errorf("width=%d: display width = %d, want <= %d (result: %q)", width, w, width-2, result)
	}
}

// firstFullWidthRuneIndex returns the rune index of the first rune whose own
// display width is 2 or more, and whether one was found.
func firstFullWidthRuneIndex(runes []rune) (int, bool) {
	for i, r := range runes {
		if lipgloss.Width(string(r)) >= 2 {
			return i, true
		}
	}
	return 0, false
}

// cursorPositionsForInput returns the cursor positions the task plan's
// AC-1/AC-2/AC-5 tables require: start, half (rune count / 2), the first
// full-width rune (if any), and end. An empty input collapses all of these
// to a single "start" position, per the task plan.
func cursorPositionsForInput(input string) []struct {
	label string
	pos   int
} {
	runes := []rune(input)
	n := len(runes)
	if n == 0 {
		return []struct {
			label string
			pos   int
		}{{"start", 0}}
	}
	positions := []struct {
		label string
		pos   int
	}{
		{"start", 0},
		{"half", n / 2},
		{"end", n},
	}
	if idx, ok := firstFullWidthRuneIndex(runes); ok {
		positions = append(positions, struct {
			label string
			pos   int
		}{"full-width", idx})
	}
	return positions
}

// --- AC-1, AC-2 (FR1, FR2, FR3, FR4, FR5, TS-1): width x input x cursor
// table. Whatever the width (24-32: prompt-only truncation range; 36-37:
// prompt fits and the input window scrolls) and whatever the input/cursor
// combination, View() must never wrap and must stay within the style's
// content width.

func TestMinibufferView_WidthInputCursorTable_BoundedSingleLine(t *testing.T) {
	prompt := "(reverse-i-search)'日本語': "
	longASCII := strings.Repeat("x", 32) // ASCII, 30+ runes, no full-width rune

	inputs := []string{"", "echo 日本語", longASCII}
	widths := []int{24, 25, 26, 27, 28, 29, 30, 31, 32, 36, 37}

	for _, width := range widths {
		for _, input := range inputs {
			for _, cur := range cursorPositionsForInput(input) {
				name := fmt.Sprintf("width=%d/input=%q/cursor=%s", width, input, cur.label)
				t.Run(name, func(t *testing.T) {
					mb := NewMinibuffer()
					mb.SetPrompt(prompt)
					mb.SetWidth(width)
					mb.SetInput(input)
					mb.SetCursorPos(cur.pos)
					mb.Show()

					result := mustNotPanic(t, "Minibuffer.View", mb.View)
					assertBoundedSingleLine(t, width, result)
				})
			}
		}
	}
}

// --- AC-3 (FR3, FR5, TS-2): boundary widths where the prompt's display
// width sits exactly at, just under, or just over the content width, and
// widths where a full-width character straddles the truncation boundary.

func TestMinibufferView_BoundaryWidths_TruncationAndBound(t *testing.T) {
	prompt := "(reverse-i-search)'日本語': "

	tests := []struct {
		width           int
		wantContains    string // "" = no content assertion for this width
		wantNotContains string
	}{
		{24, "(reverse-i-search)'", "日"},
		{25, "", ""},
		{26, "(reverse-i-search)'日", "本"},
		{31, "", ""},
		{32, "", ""},
		{33, "", ""},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("width=%d", tt.width), func(t *testing.T) {
			mb := NewMinibuffer()
			mb.SetPrompt(prompt)
			mb.SetWidth(tt.width)
			mb.Show()

			result := mustNotPanic(t, "Minibuffer.View", mb.View)
			assertBoundedSingleLine(t, tt.width, result)

			if tt.wantContains != "" && !strings.Contains(result, tt.wantContains) {
				t.Errorf("expected result to contain %q, got %q", tt.wantContains, result)
			}
			if tt.wantNotContains != "" && strings.Contains(result, tt.wantNotContains) {
				t.Errorf("expected result NOT to contain %q, got %q", tt.wantNotContains, result)
			}
		})
	}
}

func TestMinibufferView_Width33_FullWidthInputAtStart_BoundedSingleLine(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("(reverse-i-search)'日本語': ")
	mb.SetWidth(33)
	mb.SetInput("日本語")
	mb.SetCursorPos(0)
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, 33, result)
}

// --- AC-4 (FR4, TS-7): once the remaining width is >= 1 but the input does
// not fit, the character at the cursor position stays visible whenever its
// own display width allows it (and likewise the input's last rune when the
// cursor sits at the end).

func TestMinibufferView_ScrolledInput_CursorCharacterVisible(t *testing.T) {
	// All-distinct characters (ASCII and full-width) so that a rune's
	// presence in the result unambiguously identifies which part of the
	// input was kept in view.
	const input = "a1Bz9Y日本語花鳥"
	const width = 36 // contentWidth 32, promptWidth 28, remaining 4
	prompt := "(reverse-i-search)'日本語': "
	runes := []rune(input)

	if got := lipgloss.Width(prompt); got != 28 {
		t.Fatalf("test fixture assumption broken: prompt width = %d, want 28", got)
	}
	if remaining := (width - 4) - lipgloss.Width(prompt); lipgloss.Width(input)+1 <= remaining {
		t.Fatalf("test fixture assumption broken: input fits without scrolling (remaining=%d)", remaining)
	}

	for pos, r := range runes {
		t.Run(fmt.Sprintf("cursor_rune_index_%d", pos), func(t *testing.T) {
			mb := NewMinibuffer()
			mb.SetPrompt(prompt)
			mb.SetWidth(width)
			mb.SetInput(input)
			mb.SetCursorPos(pos)
			mb.Show()

			result := mustNotPanic(t, "Minibuffer.View", mb.View)
			assertBoundedSingleLine(t, width, result)

			if lipgloss.Width(string(r)) <= 4 {
				if !strings.Contains(result, string(r)) {
					t.Errorf("expected result to contain cursor rune %q, got %q", r, result)
				}
			}
		})
	}

	t.Run("cursor_at_end", func(t *testing.T) {
		mb := NewMinibuffer()
		mb.SetPrompt(prompt)
		mb.SetWidth(width)
		mb.SetInput(input)
		mb.SetCursorPos(len(runes))
		mb.Show()

		result := mustNotPanic(t, "Minibuffer.View", mb.View)
		assertBoundedSingleLine(t, width, result)

		last := runes[len(runes)-1]
		if lipgloss.Width(string(last))+1 <= 4 {
			if !strings.Contains(result, string(last)) {
				t.Errorf("expected result to contain last rune %q, got %q", last, result)
			}
		}
	})
}

// --- AC-5 (FR1, FR2, TS-5): grapheme clusters -- a combining character in
// the prompt, and a ZWJ-joined emoji sequence in the input -- never cause a
// line break or a width overflow, for any width in range and any cursor
// position, including positions inside the middle of a cluster.

func TestMinibufferView_GraphemeClusters_BoundedSingleLine(t *testing.T) {
	const combiningPrompt = "éclair: " // "e" + combining acute accent
	const zwjInput = "👨‍👩‍👧‍👦ls"        // ZWJ-joined family emoji, then ASCII

	widths := []int{24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37}
	runes := []rune(zwjInput)

	for _, width := range widths {
		for pos := 0; pos <= len(runes); pos++ {
			t.Run(fmt.Sprintf("width=%d/cursor=%d", width, pos), func(t *testing.T) {
				mb := NewMinibuffer()
				mb.SetPrompt(combiningPrompt)
				mb.SetWidth(width)
				mb.SetInput(zwjInput)
				mb.SetCursorPos(pos)
				mb.Show()

				result := mustNotPanic(t, "Minibuffer.View", mb.View)
				assertBoundedSingleLine(t, width, result)
			})
		}
	}
}

// --- AC-7 (FR6, FR8, TS-4/TS-6/TS-8): a width wide enough for both prompt
// and input renders them both without truncation.

func TestMinibufferView_Width40_PromptAndInputFitUntruncated(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("(reverse-i-search)'日本語': ")
	mb.SetWidth(40)
	mb.SetInput("ls")
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, 40, result)

	if !strings.Contains(result, "(reverse-i-search)'日本語': ") {
		t.Errorf("expected result to contain the full prompt, got %q", result)
	}
	if !strings.Contains(result, "ls") {
		t.Errorf("expected result to contain the full input, got %q", result)
	}
}

// --- task0001 (AC-1..AC-4 / SPEC AC1-AC5): the "fits untruncated" branch of
// Minibuffer.View's remaining-width >= 1 path must judge fit by the input's
// actual (grapheme-cluster) display width, not the sum of each rune's own
// width, so ZWJ-joined sequences that actually fit are shown whole.

// familyEmojiLsInput is a ZWJ-joined family emoji (U+1F468 U+200D U+1F469
// U+200D U+1F467 U+200D U+1F466), 7 runes, followed by "ls" (2 runes): 9
// runes total. Its actual display width is 4; the sum of each rune's own
// width is 10 -- the gap this task's fix exploits.
const familyEmojiLsInput = "👨‍👩‍👧‍👦ls"

// searchPrompt is the task plan's fixture prompt, with an actual display
// width of 10.
const searchPrompt = "(search): "

// checkFamilyEmojiFixture fails the test immediately if the numbers the
// task plan's test data relies on -- prompt width 10, input actual width 4,
// input rune-width sum 10 -- have drifted, so a later assertion failure
// can't be mistaken for one of these instead. Same shape as
// TestMinibufferView_ScrolledInput_CursorCharacterVisible's fixture check.
func checkFamilyEmojiFixture(t *testing.T) {
	t.Helper()
	if got := lipgloss.Width(searchPrompt); got != 10 {
		t.Fatalf("test fixture assumption broken: prompt width = %d, want 10", got)
	}
	if got := lipgloss.Width(familyEmojiLsInput); got != 4 {
		t.Fatalf("test fixture assumption broken: input actual width = %d, want 4", got)
	}
	runes := []rune(familyEmojiLsInput)
	if len(runes) != 9 {
		t.Fatalf("test fixture assumption broken: input rune count = %d, want 9", len(runes))
	}
	sum := 0
	for _, r := range runes {
		sum += lipgloss.Width(string(r))
	}
	if sum != 10 {
		t.Fatalf("test fixture assumption broken: input rune-width sum = %d, want 10", sum)
	}
}

// TestMinibufferView_ZWJInputCursorAtEnd_ShowsUntruncated is AC-1 (FR1, FR2
// / SPEC AC1, AC5): with the cursor at the end, m.width 19-24 (remaining
// width 5-10: actual width + 1 = 5 fits, but the rune-width sum + 1 = 11
// does not), the visible text must contain the prompt directly followed by
// the whole input -- the leading U+1F468 must not be dropped.
func TestMinibufferView_ZWJInputCursorAtEnd_ShowsUntruncated(t *testing.T) {
	checkFamilyEmojiFixture(t)

	want := searchPrompt + familyEmojiLsInput
	for width := 19; width <= 24; width++ {
		t.Run(fmt.Sprintf("width=%d", width), func(t *testing.T) {
			mb := NewMinibuffer()
			mb.SetPrompt(searchPrompt)
			mb.SetWidth(width)
			mb.SetInput(familyEmojiLsInput)
			mb.SetCursorPos(len([]rune(familyEmojiLsInput)))
			mb.Show()

			result := mustNotPanic(t, "Minibuffer.View", mb.View)
			assertBoundedSingleLine(t, width, result)

			visible := stripANSI(result)
			if !strings.Contains(visible, want) {
				t.Errorf("width=%d: expected visible text to contain %q, got %q", width, want, visible)
			}
		})
	}
}

// TestMinibufferView_ZWJInputCursorMidSequence_ShowsUntruncated is AC-2
// (FR1, FR3 / SPEC AC2): same prompt/input/width range as AC-1, but with
// the cursor on 'l' (rune index 7) or 's' (rune index 8) instead of at the
// end. The family emoji must not be split, and 's' must still be visible.
func TestMinibufferView_ZWJInputCursorMidSequence_ShowsUntruncated(t *testing.T) {
	checkFamilyEmojiFixture(t)

	cursors := []struct {
		label string
		pos   int
	}{
		{"l", 7},
		{"s", 8},
	}
	want := searchPrompt + familyEmojiLsInput

	for width := 19; width <= 24; width++ {
		for _, c := range cursors {
			t.Run(fmt.Sprintf("width=%d/cursor=%s", width, c.label), func(t *testing.T) {
				mb := NewMinibuffer()
				mb.SetPrompt(searchPrompt)
				mb.SetWidth(width)
				mb.SetInput(familyEmojiLsInput)
				mb.SetCursorPos(c.pos)
				mb.Show()

				result := mustNotPanic(t, "Minibuffer.View", mb.View)
				assertBoundedSingleLine(t, width, result)

				visible := stripANSI(result)
				if !strings.Contains(visible, want) {
					t.Errorf("width=%d/cursor=%s: expected visible text to contain %q, got %q", width, c.label, want, visible)
				}
			})
		}
	}
}

// TestMinibufferView_ZWJInputAllWidthsAndCursors_BoundedSingleLine is AC-3
// (FR4 / SPEC AC3): for every m.width 18-30 and every cursor position 0-9
// (including positions inside the ZWJ sequence), View's result must never
// contain a line break and its display width must stay within m.width-2.
func TestMinibufferView_ZWJInputAllWidthsAndCursors_BoundedSingleLine(t *testing.T) {
	checkFamilyEmojiFixture(t)

	for width := 18; width <= 30; width++ {
		for pos := 0; pos <= 9; pos++ {
			t.Run(fmt.Sprintf("width=%d/cursor=%d", width, pos), func(t *testing.T) {
				mb := NewMinibuffer()
				mb.SetPrompt(searchPrompt)
				mb.SetWidth(width)
				mb.SetInput(familyEmojiLsInput)
				mb.SetCursorPos(pos)
				mb.Show()

				result := mustNotPanic(t, "Minibuffer.View", mb.View)
				assertBoundedSingleLine(t, width, result)
			})
		}
	}
}

// TestMinibufferView_ZWJInputWidth18CursorAtEnd_ScrollsButShowsTail is AC-4
// (FR1 / SPEC a2): with the cursor at the end and m.width 18 (remaining
// width 4: even actual width + 1 = 5 does not fit), the existing scroll
// path is used; the visible text must still contain "ls" and satisfy the
// same single-line bound as AC-3.
func TestMinibufferView_ZWJInputWidth18CursorAtEnd_ScrollsButShowsTail(t *testing.T) {
	checkFamilyEmojiFixture(t)

	mb := NewMinibuffer()
	mb.SetPrompt(searchPrompt)
	mb.SetWidth(18)
	mb.SetInput(familyEmojiLsInput)
	mb.SetCursorPos(len([]rune(familyEmojiLsInput)))
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, 18, result)

	visible := stripANSI(result)
	if !strings.Contains(visible, "ls") {
		t.Errorf("expected visible text to contain %q, got %q", "ls", visible)
	}
}

func TestMinibufferViewTruncation(t *testing.T) {
	mb := NewMinibuffer()
	mb.SetPrompt("/: ")
	mb.SetWidth(10)
	mb.input = "this is a very long input that should be truncated"
	mb.cursorPos = len(mb.input)
	mb.Show()

	view := mb.View()

	// View should not exceed width
	// Note: actual rendering may include ANSI codes, so we check the visible content
	if len(view) > 100 { // generous limit accounting for ANSI codes
		// Just verify it renders without panic
	}
}

// --- task0001 TS-1..TS-4 (AC-1..AC-4, FR1, FR3, FR4, NFR2): time-bound
// regression tests for Minibuffer.View's linear-time range selection. Each
// fixture reproduces a shape (a very long run of zero-width combining marks
// next to a positive-width sequence whose per-rune width sum understates its
// rendered width) that made the pre-change per-rune shrink loop's initial
// range estimate swallow the whole input, and then shrink it back down one
// rune at a time -- runViewTimed fails the test rather than let a hang block
// the suite.

const linearTimeLimit = 10 * time.Second

// runViewTimed runs mb.View() in its own goroutine, logs the measured
// duration (AC-4, read by TS-6 in verbose output), and fails the test if
// View panics or does not return within linearTimeLimit. A panic inside the
// goroutine is reported as a test failure, not a crashed test binary; a
// goroutine that times out is abandoned (Test Notes: expected for the
// pre-change red-phase run).
func runViewTimed(t *testing.T, mb *Minibuffer) string {
	t.Helper()

	type viewResult struct {
		out   string
		panic any
	}
	done := make(chan viewResult, 1)
	start := time.Now()

	go func() {
		var res viewResult
		defer func() {
			res.panic = recover()
			done <- res
		}()
		res.out = mb.View()
	}()

	select {
	case res := <-done:
		t.Logf("Minibuffer.View took %s", time.Since(start))
		if res.panic != nil {
			t.Fatalf("Minibuffer.View panicked: %v", res.panic)
		}
		return res.out
	case <-time.After(linearTimeLimit):
		t.Fatalf("Minibuffer.View did not return within %s", linearTimeLimit)
		return ""
	}
}

// TestMinibufferView_LinearTime_TS1_CursorAtEndWideWidth is TS-1 (AC-1,
// AC-2): m.width 80, cursor at the end, a leading run of 100000 zero-width
// combining marks (U+0301) followed by 71 'a's and a copyright-with-VS16
// pair (U+00A9 U+FE0F) whose per-rune width sum (1) understates its
// rendered width (2).
func TestMinibufferView_LinearTime_TS1_CursorAtEndWideWidth(t *testing.T) {
	const width = 80
	input := strings.Repeat("́", 100000) + strings.Repeat("a", 71) + "©️"

	mb := NewMinibuffer()
	mb.SetPrompt("!: ")
	mb.SetWidth(width)
	mb.SetInput(input)
	mb.SetCursorPos(len([]rune(input)))
	mb.Show()

	result := runViewTimed(t, mb)
	if strings.ContainsAny(result, "\n\r") {
		t.Errorf("result contains a line break: %q", result)
	}
	if w := lipgloss.Width(result); w > width-2 {
		t.Errorf("display width = %d, want <= %d", w, width-2)
	}

	visible := stripANSI(result)
	if !strings.Contains(visible, "©️") {
		t.Errorf("expected visible result to contain the input's final U+00A9 U+FE0F, got %q", visible)
	}
}

// TestMinibufferView_LinearTime_TS2_CursorAtEndNarrowWidth is TS-2, the same
// shape as TS-1 at m.width 40 (fewer 'a's so the fixture still exercises the
// scroll path at a narrower content width).
func TestMinibufferView_LinearTime_TS2_CursorAtEndNarrowWidth(t *testing.T) {
	const width = 40
	input := strings.Repeat("́", 100000) + strings.Repeat("a", 31) + "©️"

	mb := NewMinibuffer()
	mb.SetPrompt("!: ")
	mb.SetWidth(width)
	mb.SetInput(input)
	mb.SetCursorPos(len([]rune(input)))
	mb.Show()

	result := runViewTimed(t, mb)
	if strings.ContainsAny(result, "\n\r") {
		t.Errorf("result contains a line break: %q", result)
	}
	if w := lipgloss.Width(result); w > width-2 {
		t.Errorf("display width = %d, want <= %d", w, width-2)
	}

	visible := stripANSI(result)
	if !strings.Contains(visible, "©️") {
		t.Errorf("expected visible result to contain the input's final U+00A9 U+FE0F, got %q", visible)
	}
}

// TestMinibufferView_LinearTime_TS3_CursorNearStartWideWidth is TS-3: the
// copyright-with-VS16 pair and the 'a's now sit at the FRONT of the input,
// with the cursor at rune index 2 (the first 'a', right after the pair) and
// the 100000 zero-width marks trailing off the end.
func TestMinibufferView_LinearTime_TS3_CursorNearStartWideWidth(t *testing.T) {
	const width = 40
	input := "©️" + strings.Repeat("a", 32) + strings.Repeat("́", 100000)

	mb := NewMinibuffer()
	mb.SetPrompt("!: ")
	mb.SetWidth(width)
	mb.SetInput(input)
	mb.SetCursorPos(2)
	mb.Show()

	result := runViewTimed(t, mb)
	if strings.ContainsAny(result, "\n\r") {
		t.Errorf("result contains a line break: %q", result)
	}
	if w := lipgloss.Width(result); w > width-2 {
		t.Errorf("display width = %d, want <= %d", w, width-2)
	}

	visible := stripANSI(result)
	if !strings.Contains(visible, "©️a") {
		t.Errorf("expected visible result to contain U+00A9 U+FE0F immediately followed by 'a', got %q", visible)
	}
}

// TestMinibufferView_LinearTime_TS4_PromptTruncationWideWidth is TS-4: the
// pathological run lives in the PROMPT instead of the input (empty input),
// exercising the prompt-truncation path (promptWidth >= contentWidth).
func TestMinibufferView_LinearTime_TS4_PromptTruncationWideWidth(t *testing.T) {
	const width = 40
	prompt := "(reverse-i-search)'" + "©️" + strings.Repeat("a", 16) + strings.Repeat("́", 100000) + "': "

	mb := NewMinibuffer()
	mb.SetPrompt(prompt)
	mb.SetWidth(width)
	mb.Show()

	result := runViewTimed(t, mb)
	if strings.ContainsAny(result, "\n\r") {
		t.Errorf("result contains a line break: %q", result)
	}
	if w := lipgloss.Width(result); w > width-2 {
		t.Errorf("display width = %d, want <= %d", w, width-2)
	}

	visible := stripANSI(result)
	if !strings.Contains(visible, "(reverse-i-search)'©️") {
		t.Errorf("expected visible result to contain the prompt head immediately followed by U+00A9 U+FE0F, got %q", visible)
	}
}

// TestMinibufferView_TS7_JoinedSequenceGuard is TS-7 (AC-5): a guard-path
// regression where per-rune width estimates undershoot the rendered width of
// a joined sequence (U+0600 ARABIC NUMBER SIGN followed by U+1F600 U+FE0E).
// The fixture assumption is asserted first so a later failure in this test
// cannot be mistaken for this one having drifted.
func TestMinibufferView_TS7_JoinedSequenceGuard(t *testing.T) {
	const width = 40
	const triple = "؀\U0001F600︎"

	individualSum := lipgloss.Width("؀") + lipgloss.Width("\U0001F600︎")
	joined := lipgloss.Width(triple)
	if individualSum >= joined {
		t.Fatalf("test fixture assumption broken: width(U+0600) + width(U+1F600 U+FE0E) = %d, want < width of the three together = %d", individualSum, joined)
	}

	input := strings.Repeat(triple, 30)

	mb := NewMinibuffer()
	mb.SetPrompt("!: ")
	mb.SetWidth(width)
	mb.SetInput(input)
	mb.SetCursorPos(len([]rune(input)))
	mb.Show()

	result := runViewTimed(t, mb)
	if strings.ContainsAny(result, "\n\r") {
		t.Errorf("result contains a line break: %q", result)
	}
	if w := lipgloss.Width(result); w > width-2 {
		t.Errorf("display width = %d, want <= %d", w, width-2)
	}

	visible := stripANSI(result)
	if !strings.Contains(visible, "\U0001F600") {
		t.Errorf("expected visible result to contain U+1F600, got %q", visible)
	}
}

// --- task0001 (AC-1..AC-7, FR1-FR5, NFR1-NFR3): VS16 (U+FE0F) emoji width
// regression tests. These reproduce, on both the input side and the prompt
// side, the condition under which the minibuffer previously underestimated
// the display width of a VS16-suffixed emoji pair and wrapped to a second
// line. internal/ui/minibuffer.go is not modified by this task (NFR1).

// vs16PromptJA is the task plan's P-JA fixture: display width 28.
const vs16PromptJA = "(reverse-i-search)'日本語': "

// vs16Pair is the task plan's PAIR fixture: U+2764 HEAVY BLACK HEART
// immediately followed by U+FE0F VARIATION SELECTOR-16. Display width of the
// pair is 2; the sum of the two runes' own display widths is 1 (verified by
// checkVS16Premises).
const vs16Pair = "❤️"

// vs16Input4Pairs is the task plan's IN-4 fixture: vs16Pair repeated 4
// times, 8 runes.
var vs16Input4Pairs = strings.Repeat(vs16Pair, 4)

// vs16PromptVS is the task plan's P-VS fixture: display width 30.
var vs16PromptVS = "(reverse-i-search)'" + strings.Repeat(vs16Pair, 4) + "': "

// checkVS16Premises is the task plan's Premise check (Design; AC-1, FR3): it
// verifies the fixture values every VS16 test below relies on before any
// other assertion, and stops the test immediately (fatal), naming the
// premise, the expected value and the observed value, when one has drifted.
// checkPVS additionally requires P-VS's display width to be 30 -- only the
// P-VS tests (TS-3, TS-4) need that extra premise.
func checkVS16Premises(t *testing.T, checkPVS bool) {
	t.Helper()
	if got := lipgloss.Width(vs16PromptJA); got != 28 {
		t.Fatalf("test fixture assumption broken: P-JA display width = %d, want 28", got)
	}
	if got := lipgloss.Width(vs16Pair); got != 2 {
		t.Fatalf("test fixture assumption broken: PAIR display width = %d, want 2", got)
	}
	pairRunes := []rune(vs16Pair)
	sum := 0
	for _, r := range pairRunes {
		sum += lipgloss.Width(string(r))
	}
	if sum != 1 {
		t.Fatalf("test fixture assumption broken: PAIR per-rune width sum = %d, want 1", sum)
	}
	if checkPVS {
		if got := lipgloss.Width(vs16PromptVS); got != 30 {
			t.Fatalf("test fixture assumption broken: P-VS display width = %d, want 30", got)
		}
	}
}

// checkVS16PairingIntact is the task plan's Pairing check (Design): visible
// must already be ANSI-stripped. It reports a non-fatal test error when any
// U+2764 is not immediately followed by U+FE0F, or any U+FE0F is not
// immediately preceded by U+2764 -- i.e. it detects a split VS16 pair. It
// reports nothing when visible contains neither rune.
func checkVS16PairingIntact(t *testing.T, visible string) {
	t.Helper()
	runes := []rune(visible)
	for i, r := range runes {
		switch r {
		case '❤':
			if i+1 >= len(runes) || runes[i+1] != '️' {
				t.Errorf("split VS16 pair: U+2764 at rune index %d is not immediately followed by U+FE0F in %q", i, visible)
			}
		case '️':
			if i == 0 || runes[i-1] != '❤' {
				t.Errorf("split VS16 pair: U+FE0F at rune index %d is not immediately preceded by U+2764 in %q", i, visible)
			}
		}
	}
}

// TestMinibufferView_VS16Input_Width37CursorAtEnd is TS-1 (AC-2, AC-3; FR1,
// FR2): the reproduction condition on the input side -- P-JA, m.width 37,
// input IN-4, cursor at the end -- must render as one line with both VS16
// pairs intact.
func TestMinibufferView_VS16Input_Width37CursorAtEnd(t *testing.T) {
	checkVS16Premises(t, false)

	mb := NewMinibuffer()
	mb.SetPrompt(vs16PromptJA)
	mb.SetWidth(37)
	mb.SetInput(vs16Input4Pairs)
	mb.SetCursorPos(len([]rune(vs16Input4Pairs)))
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, 37, result)

	visible := stripANSI(result)
	want := vs16PromptJA + vs16Pair + vs16Pair
	if !strings.Contains(visible, want) {
		t.Errorf("expected visible text to contain %q, got %q", want, visible)
	}
	checkVS16PairingIntact(t, visible)
}

// TestMinibufferView_VS16Input_WidthCursorTable_BoundedSingleLine is TS-2
// (AC-4; FR4): P-JA, input IN-4, every m.width 24..41 and every cursor
// position 0..8 (162 combinations, including positions on a U+FE0F rune)
// must render without panic, without a line break, within m.width-2.
func TestMinibufferView_VS16Input_WidthCursorTable_BoundedSingleLine(t *testing.T) {
	checkVS16Premises(t, false)

	for width := 24; width <= 41; width++ {
		for pos := 0; pos <= 8; pos++ {
			t.Run(fmt.Sprintf("width=%d/cursor=%d", width, pos), func(t *testing.T) {
				mb := NewMinibuffer()
				mb.SetPrompt(vs16PromptJA)
				mb.SetWidth(width)
				mb.SetInput(vs16Input4Pairs)
				mb.SetCursorPos(pos)
				mb.Show()

				result := mustNotPanic(t, "Minibuffer.View", mb.View)
				assertBoundedSingleLine(t, width, result)
			})
		}
	}
}

// TestMinibufferView_VS16Prompt_TruncationWidths_BoundedSingleLine is TS-3
// (AC-5; FR5): P-VS, empty input, every m.width 24..34 (the prompt
// truncation range) must render without panic, without a line break, within
// m.width-2.
func TestMinibufferView_VS16Prompt_TruncationWidths_BoundedSingleLine(t *testing.T) {
	checkVS16Premises(t, true)

	for width := 24; width <= 34; width++ {
		t.Run(fmt.Sprintf("width=%d", width), func(t *testing.T) {
			mb := NewMinibuffer()
			mb.SetPrompt(vs16PromptVS)
			mb.SetWidth(width)
			mb.Show()

			result := mustNotPanic(t, "Minibuffer.View", mb.View)
			assertBoundedSingleLine(t, width, result)
		})
	}
}

// TestMinibufferView_VS16Prompt_Width29_KeepsThreePairs is TS-4 (AC-5; FR5):
// P-VS, empty input, m.width 29 -- the prompt-truncation path keeps
// "(reverse-i-search)'" (19 columns) plus three PAIRs (6 columns).
func TestMinibufferView_VS16Prompt_Width29_KeepsThreePairs(t *testing.T) {
	checkVS16Premises(t, true)

	mb := NewMinibuffer()
	mb.SetPrompt(vs16PromptVS)
	mb.SetWidth(29)
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, 29, result)

	visible := stripANSI(result)
	want := "(reverse-i-search)'" + vs16Pair + vs16Pair + vs16Pair
	if !strings.Contains(visible, want) {
		t.Errorf("expected visible text to contain %q, got %q", want, visible)
	}
	checkVS16PairingIntact(t, visible)
}

// --- task0001 (AC-1..AC-8): grapheme-wise cursor movement and whole-cluster
// cursor highlight. Test Notes' "256-color profile helper" and
// "reversed-run extraction helper".

// fix256ColorProfile fixes the default lipgloss renderer's color profile to
// 256 colors for the duration of the calling test, restoring the previous
// profile at cleanup (Test Notes: 256-color profile helper). The profile is
// process-global, so callers of this helper must not run in parallel
// (no test in this package calls t.Parallel()). It first asserts, fatally,
// that rendering a sample character with the reverse attribute actually
// emits an escape sequence under the fixed profile -- otherwise a
// non-terminal test run would make AC-3, AC-4 and AC-5 pass vacuously
// (Test Notes): the split highlight would neither inflate the measured
// width nor be visible as a separate reversed run.
func fix256ColorProfile(t *testing.T) {
	t.Helper()
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.ANSI256)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(previous)
	})

	sample := lipgloss.NewStyle().Reverse(true).Render("x")
	if sample == "x" {
		t.Fatalf("fix256ColorProfile: reverse attribute produced no escape sequence under the fixed 256-color profile (got %q); AC-3/AC-4/AC-5 would pass vacuously", sample)
	}
}

// sgrParamsRegex matches one CSI SGR escape sequence, capturing its
// semicolon-separated parameter list (possibly empty, which is itself a
// reset in the SGR spec).
var sgrParamsRegex = regexp.MustCompile(`\x1b\[([0-9;]*)m`)

// reversedRuns returns, in order, the plain-text runs rendered while the
// SGR reverse attribute (parameter "7") is switched on in an ANSI-styled
// View() result (Test Notes: reversed-run extraction helper). Each CSI SGR
// sequence's parameter list is split on ";" and checked for an exact "7"
// (switches reverse on) or an exact "0" or empty list (a reset, switches
// reverse off and closes the current run). The outer minibuffer style's own
// color sequences -- e.g. "48;5;236" or "97;48;5;236" -- never contain "7"
// or "0" as a whole parameter, so they are never mistaken for the reverse
// attribute being toggled.
func reversedRuns(result string) []string {
	var runs []string
	var current strings.Builder
	reverseOn := false
	lastEnd := 0

	flushText := func(text string) {
		if reverseOn {
			current.WriteString(text)
		}
	}
	closeRun := func() {
		if reverseOn {
			runs = append(runs, current.String())
			current.Reset()
		}
	}

	for _, m := range sgrParamsRegex.FindAllStringSubmatchIndex(result, -1) {
		segStart, segEnd := m[0], m[1]
		paramStart, paramEnd := m[2], m[3]
		flushText(result[lastEnd:segStart])

		params := result[paramStart:paramEnd]
		isReverseOn := false
		isReset := params == ""
		for _, p := range strings.Split(params, ";") {
			switch p {
			case "7":
				isReverseOn = true
			case "0":
				isReset = true
			}
		}
		switch {
		case isReverseOn:
			reverseOn = true
		case isReset:
			closeRun()
			reverseOn = false
		}
		lastEnd = segEnd
	}
	flushText(result[lastEnd:])
	closeRun()

	return runs
}

// --- AC-1 (FR1; SPEC AC1, AC2; TS1, TS2): with familyEmojiLsInput, chained
// Left/Ctrl+B presses from the end and chained Right/Ctrl+F presses from
// the start move the cursor across exactly one grapheme cluster per press,
// and a further press at either boundary is a no-op.

func TestMinibufferHandleKey_GraphemeMovement_FamilyEmoji(t *testing.T) {
	checkFamilyEmojiFixture(t)

	runLeftLike := func(t *testing.T, key tea.KeyType) {
		mb := NewMinibuffer()
		mb.Show()
		mb.SetInput(familyEmojiLsInput)
		mb.SetCursorPos(9)

		for _, want := range []int{8, 7, 0} {
			mb.HandleKey(tea.KeyMsg{Type: key})
			if got := mb.CursorPos(); got != want {
				t.Fatalf("key=%v: cursorPos = %d, want %d", key, got, want)
			}
		}
		// A further press at 0 is a no-op.
		mb.HandleKey(tea.KeyMsg{Type: key})
		if got := mb.CursorPos(); got != 0 {
			t.Errorf("key=%v at 0: cursorPos = %d, want 0", key, got)
		}
	}
	runRightLike := func(t *testing.T, key tea.KeyType) {
		mb := NewMinibuffer()
		mb.Show()
		mb.SetInput(familyEmojiLsInput)
		mb.SetCursorPos(0)

		for _, want := range []int{7, 8, 9} {
			mb.HandleKey(tea.KeyMsg{Type: key})
			if got := mb.CursorPos(); got != want {
				t.Fatalf("key=%v: cursorPos = %d, want %d", key, got, want)
			}
		}
		// A further press at the end is a no-op.
		mb.HandleKey(tea.KeyMsg{Type: key})
		if got := mb.CursorPos(); got != 9 {
			t.Errorf("key=%v at 9: cursorPos = %d, want 9", key, got)
		}
	}

	t.Run("Left", func(t *testing.T) { runLeftLike(t, tea.KeyLeft) })
	t.Run("CtrlB", func(t *testing.T) { runLeftLike(t, tea.KeyCtrlB) })
	t.Run("Right", func(t *testing.T) { runRightLike(t, tea.KeyRight) })
	t.Run("CtrlF", func(t *testing.T) { runRightLike(t, tea.KeyCtrlF) })
}

// --- AC-2 (FR1; SPEC AC3, AC4; TS3, TS4): SetCursorPos inside the family
// emoji cluster (the tab-completion path), a table-driven walk over five
// non-emoji cluster shapes (combining accent, VS16 heart pair, regional-
// indicator flag pair, leading isolated combining mark, leading isolated
// ZWJ), and an empty input.

func TestMinibufferHandleKey_GraphemeMovement_SetCursorInsideCluster(t *testing.T) {
	checkFamilyEmojiFixture(t)

	for k := 1; k <= 6; k++ {
		for _, key := range []tea.KeyType{tea.KeyLeft, tea.KeyCtrlB} {
			t.Run(fmt.Sprintf("k=%d/key=%v", k, key), func(t *testing.T) {
				mb := NewMinibuffer()
				mb.Show()
				mb.SetInput(familyEmojiLsInput)
				mb.SetCursorPos(k)
				mb.HandleKey(tea.KeyMsg{Type: key})
				if got := mb.CursorPos(); got != 0 {
					t.Errorf("k=%d key=%v: cursorPos = %d, want 0", k, key, got)
				}
			})
		}
		for _, key := range []tea.KeyType{tea.KeyRight, tea.KeyCtrlF} {
			t.Run(fmt.Sprintf("k=%d/key=%v", k, key), func(t *testing.T) {
				mb := NewMinibuffer()
				mb.Show()
				mb.SetInput(familyEmojiLsInput)
				mb.SetCursorPos(k)
				mb.HandleKey(tea.KeyMsg{Type: key})
				if got := mb.CursorPos(); got != 7 {
					t.Errorf("k=%d key=%v: cursorPos = %d, want 7", k, key, got)
				}
			})
		}
	}
}

// graphemeMoveCase is one row of the AC-2 table: from cursor position
// `from`, pressing `key` must leave the cursor at exactly `want`.
type graphemeMoveCase struct {
	from int
	key  tea.KeyType
	want int
}

func assertGraphemeMoves(t *testing.T, name, input string, cases []graphemeMoveCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(fmt.Sprintf("%s/from=%d/key=%v/want=%d", name, c.from, c.key, c.want), func(t *testing.T) {
			mb := NewMinibuffer()
			mb.Show()
			mb.SetInput(input)
			mb.SetCursorPos(c.from)
			mb.HandleKey(tea.KeyMsg{Type: c.key})
			if got := mb.CursorPos(); got != c.want {
				t.Errorf("input=%q from=%d key=%v: cursorPos = %d, want %d", input, c.from, c.key, got, c.want)
			}
		})
	}
}

func TestMinibufferHandleKey_GraphemeMovement_TableDriven(t *testing.T) {
	// "e" + U+0301 (combining acute) + "x": clusters [0,2) "e´", [2,3) "x".
	const combiningInput = "éx"
	assertGraphemeMoves(t, "combining", combiningInput, []graphemeMoveCase{
		{from: 3, key: tea.KeyLeft, want: 2},
		{from: 2, key: tea.KeyLeft, want: 0},
		{from: 1, key: tea.KeyLeft, want: 0}, // inside the cluster -> its start
		{from: 0, key: tea.KeyLeft, want: 0}, // no-op at start
		{from: 0, key: tea.KeyRight, want: 2},
		{from: 2, key: tea.KeyRight, want: 3},
		{from: 1, key: tea.KeyRight, want: 2}, // inside the cluster -> its end
		{from: 3, key: tea.KeyRight, want: 3}, // no-op at end
		{from: 3, key: tea.KeyCtrlB, want: 2},
		{from: 0, key: tea.KeyCtrlF, want: 2},
	})

	// U+2764 U+FE0F (heart + VS16) between two ASCII letters: clusters
	// [0,1) "a", [1,3) heart, [3,4) "b".
	const heartInput = "a" + vs16Pair + "b"
	assertGraphemeMoves(t, "heart_vs16", heartInput, []graphemeMoveCase{
		{from: 4, key: tea.KeyLeft, want: 3},
		{from: 3, key: tea.KeyLeft, want: 1},
		{from: 2, key: tea.KeyLeft, want: 1}, // inside the cluster -> its start
		{from: 1, key: tea.KeyLeft, want: 0},
		{from: 0, key: tea.KeyLeft, want: 0}, // no-op at start
		{from: 0, key: tea.KeyRight, want: 1},
		{from: 1, key: tea.KeyRight, want: 3},
		{from: 2, key: tea.KeyRight, want: 3}, // inside the cluster -> its end
		{from: 3, key: tea.KeyRight, want: 4},
		{from: 4, key: tea.KeyRight, want: 4}, // no-op at end
		{from: 4, key: tea.KeyCtrlB, want: 3},
		{from: 0, key: tea.KeyCtrlF, want: 1},
	})

	// U+1F1EF U+1F1F5 (regional-indicator flag pair) between two ASCII
	// letters: clusters [0,1) "a", [1,3) flag, [3,4) "b".
	const flagInput = "a\U0001F1EF\U0001F1F5b"
	assertGraphemeMoves(t, "flag_pair", flagInput, []graphemeMoveCase{
		{from: 4, key: tea.KeyLeft, want: 3},
		{from: 3, key: tea.KeyLeft, want: 1},
		{from: 2, key: tea.KeyLeft, want: 1}, // inside the cluster -> its start
		{from: 1, key: tea.KeyLeft, want: 0},
		{from: 0, key: tea.KeyLeft, want: 0}, // no-op at start
		{from: 0, key: tea.KeyRight, want: 1},
		{from: 1, key: tea.KeyRight, want: 3},
		{from: 2, key: tea.KeyRight, want: 3}, // inside the cluster -> its end
		{from: 3, key: tea.KeyRight, want: 4},
		{from: 4, key: tea.KeyRight, want: 4}, // no-op at end
		{from: 4, key: tea.KeyCtrlB, want: 3},
		{from: 0, key: tea.KeyCtrlF, want: 1},
	})

	// A leading isolated combining mark (no base to attach to) followed by
	// ASCII letters: clusters [0,1) U+0301 alone, [1,2) "a", [2,3) "b".
	const leadingCombiningInput = "́ab"
	assertGraphemeMoves(t, "leading_combining", leadingCombiningInput, []graphemeMoveCase{
		{from: 3, key: tea.KeyLeft, want: 2},
		{from: 2, key: tea.KeyLeft, want: 1},
		{from: 1, key: tea.KeyLeft, want: 0},
		{from: 0, key: tea.KeyLeft, want: 0}, // no-op at start
		{from: 0, key: tea.KeyRight, want: 1},
		{from: 1, key: tea.KeyRight, want: 2},
		{from: 2, key: tea.KeyRight, want: 3},
		{from: 3, key: tea.KeyRight, want: 3}, // no-op at end
		{from: 3, key: tea.KeyCtrlB, want: 2},
		{from: 0, key: tea.KeyCtrlF, want: 1},
	})

	// A leading isolated ZWJ (no preceding or following pictographic to
	// join) followed by ASCII letters: same cluster shape as above.
	const leadingZWJInput = "‍ab"
	assertGraphemeMoves(t, "leading_zwj", leadingZWJInput, []graphemeMoveCase{
		{from: 3, key: tea.KeyLeft, want: 2},
		{from: 2, key: tea.KeyLeft, want: 1},
		{from: 1, key: tea.KeyLeft, want: 0},
		{from: 0, key: tea.KeyLeft, want: 0}, // no-op at start
		{from: 0, key: tea.KeyRight, want: 1},
		{from: 1, key: tea.KeyRight, want: 2},
		{from: 2, key: tea.KeyRight, want: 3},
		{from: 3, key: tea.KeyRight, want: 3}, // no-op at end
		{from: 3, key: tea.KeyCtrlB, want: 2},
		{from: 0, key: tea.KeyCtrlF, want: 1},
	})
}

func TestMinibufferHandleKey_GraphemeMovement_EmptyInput(t *testing.T) {
	for _, key := range []tea.KeyType{tea.KeyLeft, tea.KeyCtrlB, tea.KeyRight, tea.KeyCtrlF} {
		t.Run(fmt.Sprintf("key=%v", key), func(t *testing.T) {
			mb := NewMinibuffer()
			mb.Show()
			mb.HandleKey(tea.KeyMsg{Type: key})
			if got := mb.CursorPos(); got != 0 {
				t.Errorf("key=%v on empty input: cursorPos = %d, want 0", key, got)
			}
		})
	}
}

// --- AC-3 (FR2; SPEC AC5, AC6; TS5): with the 256-color profile fixed and
// a width at which the whole line fits (m.width 19), View's output contains
// exactly one reversed run for every cursor 0..9, and that run is exactly
// the expected cluster ("l" at 7, "s" at 8, the block cursor at 9, and the
// whole 7-rune family emoji for every cursor 0..6 -- the emoji is a single
// cluster regardless of which of its runes the cursor sits on).

func TestMinibufferView_FamilyEmoji_SingleReversedRunWholeCluster(t *testing.T) {
	fix256ColorProfile(t)
	checkFamilyEmojiFixture(t)

	const width = 19
	runes := []rune(familyEmojiLsInput)
	familyEmojiText := string(runes[0:7])

	for pos := 0; pos <= 9; pos++ {
		t.Run(fmt.Sprintf("cursor=%d", pos), func(t *testing.T) {
			mb := NewMinibuffer()
			mb.SetPrompt(searchPrompt)
			mb.SetWidth(width)
			mb.SetInput(familyEmojiLsInput)
			mb.SetCursorPos(pos)
			mb.Show()

			result := mustNotPanic(t, "Minibuffer.View", mb.View)
			assertBoundedSingleLine(t, width, result)
			runs := reversedRuns(result)

			var want string
			switch {
			case pos <= 6:
				want = familyEmojiText
			case pos == 7:
				want = "l"
			case pos == 8:
				want = "s"
			default: // pos == 9
				want = " "
			}

			if len(runs) != 1 || runs[0] != want {
				t.Errorf("cursor=%d: reversed runs = %#v, want exactly [%q]", pos, runs, want)
			}

			if pos <= 6 {
				visible := stripANSI(result)
				if !strings.Contains(visible, familyEmojiText) {
					t.Errorf("cursor=%d: visible text does not contain the whole family emoji: %q", pos, visible)
				}
			}
		})
	}
}

// --- AC-4 (FR3; SPEC AC7; TS6): with the 256-color profile fixed, for
// every m.width 19..24 and every cursor 1..6, the whole line renders
// untruncated (prompt immediately followed by the whole input) -- the
// cluster highlight no longer inflates the measured width.

func TestMinibufferView_FamilyEmoji_FitsUntruncated_WidthCursorTable(t *testing.T) {
	fix256ColorProfile(t)
	checkFamilyEmojiFixture(t)

	want := searchPrompt + familyEmojiLsInput
	for width := 19; width <= 24; width++ {
		for pos := 1; pos <= 6; pos++ {
			t.Run(fmt.Sprintf("width=%d/cursor=%d", width, pos), func(t *testing.T) {
				mb := NewMinibuffer()
				mb.SetPrompt(searchPrompt)
				mb.SetWidth(width)
				mb.SetInput(familyEmojiLsInput)
				mb.SetCursorPos(pos)
				mb.Show()

				result := mustNotPanic(t, "Minibuffer.View", mb.View)
				assertBoundedSingleLine(t, width, result)

				visible := stripANSI(result)
				if !strings.Contains(visible, want) {
					t.Errorf("width=%d/cursor=%d: expected visible text to contain %q, got %q", width, pos, want, visible)
				}
			})
		}
	}
}

// --- AC-5 (FR4; SPEC AC8; TS7): with the 256-color profile fixed, for
// every m.width 14..18 and every cursor 0..6, the family emoji either
// appears whole inside a single reversed run or none of its runes appear.
// At m.width 15 specifically, no input rune at all is rendered.

func TestMinibufferView_FamilyEmoji_NarrowWidths_WholeOrNone(t *testing.T) {
	fix256ColorProfile(t)
	checkFamilyEmojiFixture(t)

	runes := []rune(familyEmojiLsInput)
	familyEmojiText := string(runes[0:7])
	emojiRunes := runes[0:7]

	for width := 14; width <= 18; width++ {
		for pos := 0; pos <= 6; pos++ {
			t.Run(fmt.Sprintf("width=%d/cursor=%d", width, pos), func(t *testing.T) {
				mb := NewMinibuffer()
				mb.SetPrompt(searchPrompt)
				mb.SetWidth(width)
				mb.SetInput(familyEmojiLsInput)
				mb.SetCursorPos(pos)
				mb.Show()

				result := mustNotPanic(t, "Minibuffer.View", mb.View)
				assertBoundedSingleLine(t, width, result)

				visible := stripANSI(result)
				containsWhole := strings.Contains(visible, familyEmojiText)
				anyEmojiRune := false
				for _, r := range emojiRunes {
					if strings.ContainsRune(visible, r) {
						anyEmojiRune = true
						break
					}
				}

				switch {
				case containsWhole:
					runs := reversedRuns(result)
					found := false
					for _, run := range runs {
						if run == familyEmojiText {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("width=%d/cursor=%d: family emoji present but not as a single reversed run; runs=%#v, visible=%q", width, pos, runs, visible)
					}
				case anyEmojiRune:
					t.Errorf("width=%d/cursor=%d: partial family emoji runes rendered (neither fully present nor fully absent): visible=%q", width, pos, visible)
				}

				if width == 15 {
					// AC-5: the emoji's width exceeds the remaining width;
					// no input rune is rendered. "search" in the prompt
					// itself contains 's', so only the text after the
					// prompt is checked.
					remainder := strings.Replace(visible, searchPrompt, "", 1)
					if strings.ContainsAny(remainder, "ls") {
						t.Errorf("width=15/cursor=%d: expected no input rune after the prompt, got remainder=%q (visible=%q)", pos, remainder, visible)
					}
				}
			})
		}
	}
}

// --- AC-6 (NFR1; SPEC AC9; TS8): new bounded-single-line tables covering a
// flag pair, an e + U+0301 sequence, and a leading isolated zero-width rune
// (both a combining mark and a ZWJ), for every m.width 10..30 and every
// cursor 0..n, under both the default profile and the fixed 256-color
// profile. All existing bounded-single-line tables (unmodified) still pass.

func TestMinibufferView_NewGraphemeFixtures_BoundedSingleLine(t *testing.T) {
	fixtures := []string{
		"éx",                           // e + combining acute + x
		"a" + vs16Pair + vs16Pair + "b", // two VS16 heart pairs
		"a\U0001F1EF\U0001F1F5b",        // regional-indicator flag pair
		"́ab",                           // leading isolated combining mark
		"‍ab",                           // leading isolated ZWJ
	}

	run := func(t *testing.T) {
		for _, input := range fixtures {
			n := len([]rune(input))
			for width := 10; width <= 30; width++ {
				for pos := 0; pos <= n; pos++ {
					t.Run(fmt.Sprintf("input=%q/width=%d/cursor=%d", input, width, pos), func(t *testing.T) {
						mb := NewMinibuffer()
						mb.SetPrompt(searchPrompt)
						mb.SetWidth(width)
						mb.SetInput(input)
						mb.SetCursorPos(pos)
						mb.Show()

						result := mustNotPanic(t, "Minibuffer.View", mb.View)
						assertBoundedSingleLine(t, width, result)
					})
				}
			}
		}
	}

	t.Run("default_profile", func(t *testing.T) { run(t) })
	t.Run("fixed_256_profile", func(t *testing.T) {
		fix256ColorProfile(t)
		run(t)
	})
}

// --- AC-7 (FR2, FR5; SPEC AC10; TS9): editing operations on
// familyEmojiLsInput with cursor 3 stay rune-wise (GD4); an insertion that
// leaves the cursor inside a cluster highlights the whole cluster on the
// next View.

func TestMinibufferHandleKey_EditingOpsStayRuneWise_FamilyEmoji(t *testing.T) {
	checkFamilyEmojiFixture(t)
	runes := []rune(familyEmojiLsInput)

	t.Run("Backspace", func(t *testing.T) {
		mb := NewMinibuffer()
		mb.Show()
		mb.SetInput(familyEmojiLsInput)
		mb.SetCursorPos(3)
		mb.HandleKey(tea.KeyMsg{Type: tea.KeyBackspace})

		want := string(append(append([]rune{}, runes[:2]...), runes[3:]...))
		if mb.Input() != want {
			t.Errorf("input = %q, want %q", mb.Input(), want)
		}
		if got := mb.CursorPos(); got != 2 {
			t.Errorf("cursorPos = %d, want 2", got)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		mb := NewMinibuffer()
		mb.Show()
		mb.SetInput(familyEmojiLsInput)
		mb.SetCursorPos(3)
		mb.HandleKey(tea.KeyMsg{Type: tea.KeyDelete})

		want := string(append(append([]rune{}, runes[:3]...), runes[4:]...))
		if mb.Input() != want {
			t.Errorf("input = %q, want %q", mb.Input(), want)
		}
		if got := mb.CursorPos(); got != 3 {
			t.Errorf("cursorPos = %d, want 3", got)
		}
	})

	t.Run("InsertRune", func(t *testing.T) {
		mb := NewMinibuffer()
		mb.Show()
		mb.SetInput(familyEmojiLsInput)
		mb.SetCursorPos(3)
		mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Z'}})

		want := string(append(append(append([]rune{}, runes[:3]...), 'Z'), runes[3:]...))
		if mb.Input() != want {
			t.Errorf("input = %q, want %q", mb.Input(), want)
		}
		if got := mb.CursorPos(); got != 4 {
			t.Errorf("cursorPos = %d, want 4", got)
		}
	})

	t.Run("CtrlK", func(t *testing.T) {
		mb := NewMinibuffer()
		mb.Show()
		mb.SetInput(familyEmojiLsInput)
		mb.SetCursorPos(3)
		mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlK})

		want := string(runes[:3])
		if mb.Input() != want {
			t.Errorf("input = %q, want %q", mb.Input(), want)
		}
		if got := mb.CursorPos(); got != 3 {
			t.Errorf("cursorPos = %d, want 3", got)
		}
	})

	t.Run("CtrlU", func(t *testing.T) {
		mb := NewMinibuffer()
		mb.Show()
		mb.SetInput(familyEmojiLsInput)
		mb.SetCursorPos(3)
		mb.HandleKey(tea.KeyMsg{Type: tea.KeyCtrlU})

		want := string(runes[3:])
		if mb.Input() != want {
			t.Errorf("input = %q, want %q", mb.Input(), want)
		}
		if got := mb.CursorPos(); got != 0 {
			t.Errorf("cursorPos = %d, want 0", got)
		}
	})
}

func TestMinibufferView_InsertionLeavesCursorInsideCluster_ReversesWholeCluster(t *testing.T) {
	fix256ColorProfile(t)

	mb := NewMinibuffer()
	mb.SetPrompt("!: ")
	mb.SetWidth(40)
	mb.SetInput("é") // "e" + combining acute accent
	mb.SetCursorPos(1)
	mb.Show()

	mb.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	if got, want := mb.Input(), "eá"; got != want {
		t.Fatalf("input = %q, want %q", got, want)
	}
	if got := mb.CursorPos(); got != 2 {
		t.Fatalf("cursorPos = %d, want 2", got)
	}

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	runs := reversedRuns(result)
	want := "á"
	if len(runs) != 1 || runs[0] != want {
		t.Errorf("reversed runs = %#v, want exactly [%q]", runs, want)
	}
}

// --- task0001 (minibuffer-grapheme-scroll-boundary, AC-1..AC-7): the
// remaining-width>=1 scroll path must place displayStart/displayEnd on
// grapheme-cluster boundaries, never mid-cluster, so a ZWJ-joined sequence
// is always shown whole or not at all.

// familyEmoji is the ZWJ-joined family emoji sequence (U+1F468 U+200D
// U+1F469 U+200D U+1F467 U+200D U+1F466): 7 runes, display width 2.
const familyEmoji = "\U0001F468‍\U0001F469‍\U0001F467‍\U0001F466"

// scrollBoundaryInput is the task plan's fixture: "aaaaa" + familyEmoji +
// "bbbbb". 17 runes; familyEmoji sits at rune positions 5-11 (inclusive) and
// has display width 2; the input's actual display width is 12.
const scrollBoundaryInput = "aaaaa" + familyEmoji + "bbbbb"

// checkScrollBoundaryFixture is the task plan's premise check (AC-4): it
// verifies, before any other assertion in the AC-1/AC-2/AC-3/AC-6 tests
// below, that searchPrompt's display width is 10, scrollBoundaryInput has
// 17 runes, its actual display width is 12, and familyEmoji's display width
// is 2. It stops the test immediately (fatal), naming the item, the
// expected value and the observed value, on the first one that has
// drifted -- same shape as checkFamilyEmojiFixture/checkVS16Premises above.
func checkScrollBoundaryFixture(t *testing.T) {
	t.Helper()
	if got := lipgloss.Width(searchPrompt); got != 10 {
		t.Fatalf("test fixture assumption broken: prompt display width = %d, want 10", got)
	}
	runes := []rune(scrollBoundaryInput)
	if got := len(runes); got != 17 {
		t.Fatalf("test fixture assumption broken: input rune count = %d, want 17", got)
	}
	if got := lipgloss.Width(scrollBoundaryInput); got != 12 {
		t.Fatalf("test fixture assumption broken: input actual display width = %d, want 12", got)
	}
	if got := lipgloss.Width(familyEmoji); got != 2 {
		t.Fatalf("test fixture assumption broken: family emoji display width = %d, want 2", got)
	}
}

// assertWholeOrNone is the task plan's "whole or none" check (AC-3, AC-5):
// visible (an already ANSI-stripped display string) must either contain seq
// in full, or contain none of seq's runes at all -- anything in between
// means seq was split across the display window's boundary. Reports a
// non-fatal failure (so a table test's other subtests still run) when
// neither holds.
func assertWholeOrNone(t *testing.T, visible, seq string) {
	t.Helper()
	if strings.Contains(visible, seq) {
		return
	}
	for _, r := range seq {
		if strings.ContainsRune(visible, r) {
			t.Errorf("sequence %q split at display boundary: got %q", seq, visible)
			return
		}
	}
}

// TestMinibufferView_ScrollBoundary_CursorAtEnd_KeepsWholeFamilyEmoji is
// AC-1 (FR1, FR4 / SPEC AC1, TS-1): m.width 24, cursor at the end -- the
// visible text must contain familyEmoji immediately followed by "bbbbb",
// and the result must satisfy the single-line condition.
func TestMinibufferView_ScrollBoundary_CursorAtEnd_KeepsWholeFamilyEmoji(t *testing.T) {
	checkScrollBoundaryFixture(t)

	mb := NewMinibuffer()
	mb.SetPrompt(searchPrompt)
	mb.SetWidth(24)
	mb.SetInput(scrollBoundaryInput)
	mb.SetCursorPos(len([]rune(scrollBoundaryInput)))
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, 24, result)

	visible := stripANSI(result)
	want := familyEmoji + "bbbbb"
	if !strings.Contains(visible, want) {
		t.Errorf("expected visible text to contain %q, got %q", want, visible)
	}
}

// TestMinibufferView_ScrollBoundary_CursorAtStart_KeepsWholeFamilyEmoji is
// AC-2 (FR2, FR4 / SPEC AC2, TS-2): m.width 24, cursor at position 0 -- the
// visible text must contain "aaaaa" immediately followed by familyEmoji,
// and the result must satisfy the single-line condition.
func TestMinibufferView_ScrollBoundary_CursorAtStart_KeepsWholeFamilyEmoji(t *testing.T) {
	checkScrollBoundaryFixture(t)

	mb := NewMinibuffer()
	mb.SetPrompt(searchPrompt)
	mb.SetWidth(24)
	mb.SetInput(scrollBoundaryInput)
	mb.SetCursorPos(0)
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, 24, result)

	visible := stripANSI(result)
	want := "aaaaa" + familyEmoji
	if !strings.Contains(visible, want) {
		t.Errorf("expected visible text to contain %q, got %q", want, visible)
	}
}

// TestMinibufferView_ScrollBoundary_AllWidthsAndCursors_WholeOrNone is AC-3
// (FR1, FR2, FR3, FR4 / SPEC AC3, TS-3): for every m.width 15-27 and every
// cursor position 0-17 (including positions inside familyEmoji, 5-11), the
// visible text must either contain familyEmoji whole or not at all, and the
// result must satisfy the single-line condition.
func TestMinibufferView_ScrollBoundary_AllWidthsAndCursors_WholeOrNone(t *testing.T) {
	checkScrollBoundaryFixture(t)

	runes := []rune(scrollBoundaryInput)
	for width := 15; width <= 27; width++ {
		for pos := 0; pos <= len(runes); pos++ {
			t.Run(fmt.Sprintf("width=%d/cursor=%d", width, pos), func(t *testing.T) {
				mb := NewMinibuffer()
				mb.SetPrompt(searchPrompt)
				mb.SetWidth(width)
				mb.SetInput(scrollBoundaryInput)
				mb.SetCursorPos(pos)
				mb.Show()

				result := mustNotPanic(t, "Minibuffer.View", mb.View)
				assertBoundedSingleLine(t, width, result)

				visible := stripANSI(result)
				assertWholeOrNone(t, visible, familyEmoji)
			})
		}
	}
}

// TestMinibufferView_ScrollBoundary_OtherJoinedGraphemes_WholeOrNone is AC-5
// (FR1, FR2 / SPEC Edge Cases): the same whole-or-none property, for a
// regional-indicator flag (U+1F1EF U+1F1F5, the JP flag) and a
// skin-tone-modified emoji (U+1F44D U+1F3FD), each embedded as "aaaaa" +
// seq + "bbbbb" with prompt searchPrompt, across m.width 15-27 and every
// cursor position 0-12.
//
// Deviation from the task plan's literal premise (recorded in the task
// report): the plan requires both sequences' per-rune width sum to exceed
// their joined width. Measured in this codebase's width basis
// (lipgloss.Width, backed by github.com/rivo/uniseg): flag sum=4 > joined=2
// (holds), but skin-tone sum=2 == joined=2 (does not hold) -- U+1F3FD
// (EMOJI MODIFIER FITZPATRICK TYPE-4) has grapheme property Extend, which
// uniseg's width algorithm always measures as width 0, both standalone and
// joined to a base emoji, so no single-base-plus-single-modifier pair can
// ever satisfy "sum > joined" here. sumExceedsJoined below records the
// per-sequence expectation actually observed; where false, the mismatch is
// logged (not asserted fatal), and the whole-or-none grid still runs as a
// regression guard.
func TestMinibufferView_ScrollBoundary_OtherJoinedGraphemes_WholeOrNone(t *testing.T) {
	sequences := []struct {
		name             string
		seq              string
		sumExceedsJoined bool
	}{
		{"flag", "\U0001F1EF\U0001F1F5", true},       // regional indicators J, P -> JP flag
		{"skin-tone", "\U0001F44D\U0001F3FD", false}, // thumbs up + medium skin tone modifier
	}

	for _, s := range sequences {
		t.Run(s.name, func(t *testing.T) {
			joined := lipgloss.Width(s.seq)
			if joined != 2 {
				t.Fatalf("test fixture assumption broken: %s joined display width = %d, want 2", s.name, joined)
			}
			sum := 0
			for _, r := range s.seq {
				sum += lipgloss.Width(string(r))
			}
			if s.sumExceedsJoined {
				if sum <= joined {
					t.Fatalf("test fixture assumption broken: %s per-rune width sum = %d, want > %d", s.name, sum, joined)
				}
			} else {
				t.Logf("%s: per-rune width sum = %d, joined width = %d (no understatement gap for this pair in this codebase's width basis; see deviation note on this test)", s.name, sum, joined)
			}

			input := "aaaaa" + s.seq + "bbbbb"
			runes := []rune(input)

			for width := 15; width <= 27; width++ {
				for pos := 0; pos <= len(runes); pos++ {
					t.Run(fmt.Sprintf("width=%d/cursor=%d", width, pos), func(t *testing.T) {
						mb := NewMinibuffer()
						mb.SetPrompt(searchPrompt)
						mb.SetWidth(width)
						mb.SetInput(input)
						mb.SetCursorPos(pos)
						mb.Show()

						result := mustNotPanic(t, "Minibuffer.View", mb.View)
						assertBoundedSingleLine(t, width, result)

						visible := stripANSI(result)
						assertWholeOrNone(t, visible, s.seq)
					})
				}
			}
		})
	}
}

// TestMinibufferView_ScrollBoundary_CursorGraphemeWiderThanRemaining_HidesInput
// is AC-6 (FR3): m.width 15 (remaining width 1), cursor at each rune
// position inside familyEmoji (5-11) -- the part of the visible text after
// the prompt must contain no non-space character (the input is not shown
// at all), and the result must satisfy the single-line condition. The
// check looks only at the part after the prompt because the prompt itself
// ("(search): ") contains an 'a'.
func TestMinibufferView_ScrollBoundary_CursorGraphemeWiderThanRemaining_HidesInput(t *testing.T) {
	checkScrollBoundaryFixture(t)

	for pos := 5; pos <= 11; pos++ {
		t.Run(fmt.Sprintf("cursor=%d", pos), func(t *testing.T) {
			mb := NewMinibuffer()
			mb.SetPrompt(searchPrompt)
			mb.SetWidth(15)
			mb.SetInput(scrollBoundaryInput)
			mb.SetCursorPos(pos)
			mb.Show()

			result := mustNotPanic(t, "Minibuffer.View", mb.View)
			assertBoundedSingleLine(t, 15, result)

			visible := stripANSI(result)
			idx := strings.Index(visible, searchPrompt)
			if idx < 0 {
				t.Fatalf("prompt not found in visible text: %q", visible)
			}
			after := visible[idx+len(searchPrompt):]
			for _, r := range after {
				if r != ' ' {
					t.Errorf("cursor=%d: expected nothing but spaces after the prompt, got %q (full: %q)", pos, after, visible)
					break
				}
			}
		})
	}
}

// --- task0001 (minibuffer-guardfit-cursor-cap, AC-1..AC-4): once guardFit's
// re-measurement cap is reached (or shrink is exhausted) while the window
// still holds more than the cursor alone, View narrows the window to the
// cursor alone and shows it whenever that line fits the content width;
// only when the cursor alone does not fit either does it hide input and
// cursor. Shared fixture per the task plan's Acceptance Criteria: prompt
// U+0600 (display width 1), m.width 10 (content width 6, width left for
// input 5).

// guardCapPrompt/guardCapWidth/guardCapContentWidth are the task plan's
// shared fixture (Acceptance Criteria).
const guardCapPrompt = "؀"
const guardCapWidth = 10
const guardCapContentWidth = guardCapWidth - 4

// guardCapG1..guardCapG4 are the task plan's shared fixture graphemes: g1 is
// U+1F600 U+FE0E; g2/g3/g4 each chain one more "U+2764 U+200D" (heart + ZWJ)
// pair in front of the previous grapheme, so each of g1..g4 is a single
// ZWJ-joined grapheme cluster of increasing rune count (2, 4, 6, 8).
const guardCapG1 = "\U0001F600︎"
const guardCapG2 = "❤‍" + guardCapG1
const guardCapG3 = "❤‍" + guardCapG2
const guardCapG4 = "❤‍" + guardCapG3

// checkGuardCapPremises is this task's Premise check (Test Notes), run
// fatally before any other assertion in AC-1..AC-4 so fixture drift is
// never mistaken for a regression. It verifies, under the color profile
// already active when it's called (so it reflects exactly what the calling
// test's View() call will measure), the two facts Test Notes calls out:
// the prompt plus the whole input -- assembled and highlighted exactly as
// View's Step 1 fit-check assembles it -- does not fit the content width,
// and the prompt plus the cursor-alone line -- assembled exactly as View
// assembles its final line (Design "Contracts": Measured equals displayed)
// -- fits the content width iff wantCursorAloneFits says so.
func checkGuardCapPremises(t *testing.T, input string, cursorPos int, wantCursorAloneFits bool) {
	t.Helper()
	if got := lipgloss.Width(guardCapPrompt); got != 1 {
		t.Fatalf("test fixture assumption broken: prompt display width = %d, want 1", got)
	}

	runes := []rune(input)
	gs, ge := len(runes), len(runes)
	if cursorPos < len(runes) {
		gs, ge = graphemeClusterBounds(runes, cursorPos)
	}

	fullLine := guardCapPrompt + buildLine(runes, gs, ge)
	if w := lipgloss.Width(fullLine); w <= guardCapContentWidth {
		t.Fatalf("test fixture assumption broken: prompt + whole input fits the content width (got %d, content width %d); this input never reaches the failure-handling path", w, guardCapContentWidth)
	}

	var cursorAloneLine string
	if gs == len(runes) {
		cursorAloneLine = guardCapPrompt + buildLine(nil, 0, 0)
	} else {
		cursorAloneLine = guardCapPrompt + buildLine(runes[gs:ge], 0, ge-gs)
	}
	fits := lipgloss.Width(cursorAloneLine) <= guardCapContentWidth
	if fits != wantCursorAloneFits {
		t.Fatalf("test fixture assumption broken: prompt + cursor-alone line fits the content width = %v, want %v (measured width %d, content width %d)", fits, wantCursorAloneFits, lipgloss.Width(cursorAloneLine), guardCapContentWidth)
	}
}

// TestMinibufferView_GuardCapCursorAtEnd_ShowsBlockCursor is AC-1 (FR1, FR3,
// FR4; SPEC AC1): P1, the cursor-at-end path. Input g1+g2+g3+g4, cursor at
// the end (rune 20). guardFit's re-measurement cap is reached while the
// window still holds more than the trailing block cursor; View must narrow
// the window to the block cursor alone and show it, instead of hiding the
// cursor entirely (the defect this task fixes).
//
// Red first (Test Notes): against the unmodified View, this fails -- zero
// reversed runs, because the cursor disappears.
func TestMinibufferView_GuardCapCursorAtEnd_ShowsBlockCursor(t *testing.T) {
	fix256ColorProfile(t)

	input := guardCapG1 + guardCapG2 + guardCapG3 + guardCapG4
	cursorPos := len([]rune(input))
	checkGuardCapPremises(t, input, cursorPos, true)

	mb := NewMinibuffer()
	mb.SetPrompt(guardCapPrompt)
	mb.SetWidth(guardCapWidth)
	mb.SetInput(input)
	mb.SetCursorPos(cursorPos)
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, guardCapWidth, result)

	runs := reversedRuns(result)
	if len(runs) != 1 || runs[0] != " " {
		t.Fatalf("reversed runs = %#v, want exactly [%q]", runs, " ")
	}

	visible := stripANSI(result)
	if !strings.ContainsRune(visible, '؀') {
		t.Errorf("expected visible text to contain the prompt U+0600, got %q", visible)
	}
	if strings.ContainsRune(visible, '\U0001F600') || strings.ContainsRune(visible, '❤') {
		t.Errorf("expected visible text to contain neither U+1F600 nor U+2764, got %q", visible)
	}
}

// TestMinibufferView_GuardCapAfterThenBeforeTrim_ShowsCursorChar is AC-2
// (FR1, FR3, FR4; SPEC AC2): P2, the path that trims after-cursor units
// from the back before before-cursor units from the front. Input
// g2+g3+g4+"xyz", cursor 18 (on "x"). guardFit's cap is reached while the
// window still holds more than the cursor's own grapheme; View must narrow
// to the cursor's grapheme alone and show it as one reversed run.
//
// Red first (Test Notes): against the unmodified View, this fails -- zero
// reversed runs.
func TestMinibufferView_GuardCapAfterThenBeforeTrim_ShowsCursorChar(t *testing.T) {
	fix256ColorProfile(t)

	input := guardCapG2 + guardCapG3 + guardCapG4 + "xyz"
	const cursorPos = 18
	checkGuardCapPremises(t, input, cursorPos, true)

	mb := NewMinibuffer()
	mb.SetPrompt(guardCapPrompt)
	mb.SetWidth(guardCapWidth)
	mb.SetInput(input)
	mb.SetCursorPos(cursorPos)
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, guardCapWidth, result)

	runs := reversedRuns(result)
	if len(runs) != 1 || runs[0] != "x" {
		t.Fatalf("reversed runs = %#v, want exactly [%q]", runs, "x")
	}

	visible := stripANSI(result)
	if !strings.ContainsRune(visible, '؀') {
		t.Errorf("expected visible text to contain the prompt U+0600, got %q", visible)
	}
	if strings.ContainsRune(visible, '\U0001F600') || strings.ContainsRune(visible, '❤') {
		t.Errorf("expected visible text to contain neither U+1F600 nor U+2764, got %q", visible)
	}
}

// TestMinibufferView_GuardCapScrollLeft_ShowsCursorChar is AC-3 (FR1, FR3,
// FR4; SPEC AC3): P3, the path that scrolls left with the cursor's
// grapheme pinned at the right edge. Input "z"+g1+g2+g3+g4+"x", cursor 21
// (on "x"). Same assertions as AC-2.
//
// Red first (Test Notes): against the unmodified View, this fails -- zero
// reversed runs.
func TestMinibufferView_GuardCapScrollLeft_ShowsCursorChar(t *testing.T) {
	fix256ColorProfile(t)

	input := "z" + guardCapG1 + guardCapG2 + guardCapG3 + guardCapG4 + "x"
	const cursorPos = 21
	checkGuardCapPremises(t, input, cursorPos, true)

	mb := NewMinibuffer()
	mb.SetPrompt(guardCapPrompt)
	mb.SetWidth(guardCapWidth)
	mb.SetInput(input)
	mb.SetCursorPos(cursorPos)
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, guardCapWidth, result)

	runs := reversedRuns(result)
	if len(runs) != 1 || runs[0] != "x" {
		t.Fatalf("reversed runs = %#v, want exactly [%q]", runs, "x")
	}

	visible := stripANSI(result)
	if !strings.ContainsRune(visible, '؀') {
		t.Errorf("expected visible text to contain the prompt U+0600, got %q", visible)
	}
	if strings.ContainsRune(visible, '\U0001F600') || strings.ContainsRune(visible, '❤') {
		t.Errorf("expected visible text to contain neither U+1F600 nor U+2764, got %q", visible)
	}
}

// fixNoStyleColorProfile fixes the default lipgloss renderer's color
// profile to Ascii (emits no escape sequences at all, including the
// reverse attribute) for the duration of the calling test, restoring the
// previous profile at cleanup -- mirrors fix256ColorProfile above. AC-4's
// fixture relies on the boundary-joining effect between the prompt's last
// rune (U+0600, a Prepend grapheme-cluster-break character that joins with
// whatever text immediately follows it) and the cursor's grapheme: under a
// styled profile, a reverse-attribute escape sequence sits between the
// prompt and the highlighted cursor text, blocking that join; under Ascii,
// there is no escape sequence, so the join happens exactly as it would in
// the two texts assembled directly. The profile is process-global, so
// callers of this helper must not run in parallel (no test in this package
// calls t.Parallel()).
func fixNoStyleColorProfile(t *testing.T) {
	t.Helper()
	previous := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.Ascii)
	t.Cleanup(func() {
		lipgloss.SetColorProfile(previous)
	})

	sample := lipgloss.NewStyle().Reverse(true).Render("x")
	if sample != "x" {
		t.Fatalf("fixNoStyleColorProfile: reverse attribute produced an escape sequence under the Ascii profile (got %q); AC-4 needs plain text so the prompt and cursor grapheme sit adjacent", sample)
	}
}

// TestMinibufferView_GuardCapCursorAloneAlsoDoesNotFit_HidesInput is AC-4
// (FR2, FR3, FR4, TM-2; SPEC AC4): when the prompt plus the cursor-alone
// line is still wider than the content width, View renders neither input
// nor cursor. The fixture is guardCapG4 with one more "heart + ZWJ" pair
// prepended (making the whole input a single grapheme cluster) at cursor
// 0, run under the no-style color profile (Test Notes: under this profile
// the prompt's trailing U+0600 joins with the cursor's grapheme, so their
// combined measured width exceeds their separately-measured widths -- SPEC
// Edge Cases' "prompt boundary joins with the input side").
//
// This input's cursor-alone line already does not fit at guardFit's very
// first measurement (there is nothing else in the input to trim), so
// guardFit itself reports "does not fit" without this task's new
// cursor-alone fallback ever running its own extra measurement --
// situation (b), not (a). The outcome (hidden) is unchanged from before
// this task; see the task report's unconfirmed_reds for why this
// criterion has no observed red.
func TestMinibufferView_GuardCapCursorAloneAlsoDoesNotFit_HidesInput(t *testing.T) {
	fixNoStyleColorProfile(t)

	input := "❤‍" + guardCapG4
	const cursorPos = 0
	checkGuardCapPremises(t, input, cursorPos, false)

	mb := NewMinibuffer()
	mb.SetPrompt(guardCapPrompt)
	mb.SetWidth(guardCapWidth)
	mb.SetInput(input)
	mb.SetCursorPos(cursorPos)
	mb.Show()

	result := mustNotPanic(t, "Minibuffer.View", mb.View)
	assertBoundedSingleLine(t, guardCapWidth, result)

	if len(reversedRuns(result)) != 0 {
		t.Errorf("expected no reversed run, got %#v", reversedRuns(result))
	}

	visible := stripANSI(result)
	if strings.ContainsRune(visible, '\U0001F600') || strings.ContainsRune(visible, '❤') {
		t.Errorf("expected no input rune in the visible output, got %q", visible)
	}
}
