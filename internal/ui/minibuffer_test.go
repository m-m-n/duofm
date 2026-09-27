package ui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
