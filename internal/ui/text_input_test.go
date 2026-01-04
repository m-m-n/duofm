package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewTextInput(t *testing.T) {
	tests := []struct {
		name         string
		initialValue string
		wantValue    string
		wantCursor   int
	}{
		{
			name:         "empty initial value",
			initialValue: "",
			wantValue:    "",
			wantCursor:   0,
		},
		{
			name:         "non-empty initial value",
			initialValue: "hello",
			wantValue:    "hello",
			wantCursor:   5,
		},
		{
			name:         "unicode initial value",
			initialValue: "こんにちは",
			wantValue:    "こんにちは",
			wantCursor:   5, // 5 runes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := NewTextInput(tt.initialValue)
			if ti.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", ti.Value, tt.wantValue)
			}
			if ti.CursorPos != tt.wantCursor {
				t.Errorf("CursorPos = %d, want %d", ti.CursorPos, tt.wantCursor)
			}
		})
	}
}

func TestTextInput_InsertRunes(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursorPos  int
		insert     []rune
		wantValue  string
		wantCursor int
	}{
		{
			name:       "insert at end",
			initial:    "hello",
			cursorPos:  5,
			insert:     []rune("!"),
			wantValue:  "hello!",
			wantCursor: 6,
		},
		{
			name:       "insert at beginning",
			initial:    "world",
			cursorPos:  0,
			insert:     []rune("hello "),
			wantValue:  "hello world",
			wantCursor: 6,
		},
		{
			name:       "insert in middle",
			initial:    "hllo",
			cursorPos:  1,
			insert:     []rune("e"),
			wantValue:  "hello",
			wantCursor: 2,
		},
		{
			name:       "insert unicode",
			initial:    "hello",
			cursorPos:  5,
			insert:     []rune("世界"),
			wantValue:  "hello世界",
			wantCursor: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &TextInput{Value: tt.initial, CursorPos: tt.cursorPos}
			ti.InsertRunes(tt.insert)
			if ti.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", ti.Value, tt.wantValue)
			}
			if ti.CursorPos != tt.wantCursor {
				t.Errorf("CursorPos = %d, want %d", ti.CursorPos, tt.wantCursor)
			}
		})
	}
}

func TestTextInput_DeleteBackward(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursorPos  int
		wantValue  string
		wantCursor int
	}{
		{
			name:       "delete at end",
			initial:    "hello",
			cursorPos:  5,
			wantValue:  "hell",
			wantCursor: 4,
		},
		{
			name:       "delete in middle",
			initial:    "hello",
			cursorPos:  3,
			wantValue:  "helo",
			wantCursor: 2,
		},
		{
			name:       "delete at beginning - no change",
			initial:    "hello",
			cursorPos:  0,
			wantValue:  "hello",
			wantCursor: 0,
		},
		{
			name:       "delete unicode",
			initial:    "こんにちは",
			cursorPos:  3,
			wantValue:  "こんちは",
			wantCursor: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &TextInput{Value: tt.initial, CursorPos: tt.cursorPos}
			ti.DeleteBackward()
			if ti.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", ti.Value, tt.wantValue)
			}
			if ti.CursorPos != tt.wantCursor {
				t.Errorf("CursorPos = %d, want %d", ti.CursorPos, tt.wantCursor)
			}
		})
	}
}

func TestTextInput_DeleteForward(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursorPos  int
		wantValue  string
		wantCursor int
	}{
		{
			name:       "delete at beginning",
			initial:    "hello",
			cursorPos:  0,
			wantValue:  "ello",
			wantCursor: 0,
		},
		{
			name:       "delete in middle",
			initial:    "hello",
			cursorPos:  2,
			wantValue:  "helo",
			wantCursor: 2,
		},
		{
			name:       "delete at end - no change",
			initial:    "hello",
			cursorPos:  5,
			wantValue:  "hello",
			wantCursor: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &TextInput{Value: tt.initial, CursorPos: tt.cursorPos}
			ti.DeleteForward()
			if ti.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", ti.Value, tt.wantValue)
			}
			if ti.CursorPos != tt.wantCursor {
				t.Errorf("CursorPos = %d, want %d", ti.CursorPos, tt.wantCursor)
			}
		})
	}
}

func TestTextInput_CursorMovement(t *testing.T) {
	ti := &TextInput{Value: "hello", CursorPos: 2}

	// Move left
	ti.MoveCursorLeft()
	if ti.CursorPos != 1 {
		t.Errorf("after MoveCursorLeft: CursorPos = %d, want 1", ti.CursorPos)
	}

	// Move left at beginning
	ti.CursorPos = 0
	ti.MoveCursorLeft()
	if ti.CursorPos != 0 {
		t.Errorf("MoveCursorLeft at beginning: CursorPos = %d, want 0", ti.CursorPos)
	}

	// Move right
	ti.CursorPos = 2
	ti.MoveCursorRight()
	if ti.CursorPos != 3 {
		t.Errorf("after MoveCursorRight: CursorPos = %d, want 3", ti.CursorPos)
	}

	// Move right at end
	ti.CursorPos = 5
	ti.MoveCursorRight()
	if ti.CursorPos != 5 {
		t.Errorf("MoveCursorRight at end: CursorPos = %d, want 5", ti.CursorPos)
	}

	// Move to start
	ti.CursorPos = 3
	ti.MoveCursorToStart()
	if ti.CursorPos != 0 {
		t.Errorf("after MoveCursorToStart: CursorPos = %d, want 0", ti.CursorPos)
	}

	// Move to end
	ti.MoveCursorToEnd()
	if ti.CursorPos != 5 {
		t.Errorf("after MoveCursorToEnd: CursorPos = %d, want 5", ti.CursorPos)
	}
}

func TestTextInput_DeleteToStart(t *testing.T) {
	ti := &TextInput{Value: "hello world", CursorPos: 6}
	ti.DeleteToStart()
	if ti.Value != "world" {
		t.Errorf("Value = %q, want %q", ti.Value, "world")
	}
	if ti.CursorPos != 0 {
		t.Errorf("CursorPos = %d, want 0", ti.CursorPos)
	}
}

func TestTextInput_DeleteToEnd(t *testing.T) {
	ti := &TextInput{Value: "hello world", CursorPos: 5}
	ti.DeleteToEnd()
	if ti.Value != "hello" {
		t.Errorf("Value = %q, want %q", ti.Value, "hello")
	}
	if ti.CursorPos != 5 {
		t.Errorf("CursorPos = %d, want 5", ti.CursorPos)
	}
}

func TestTextInput_SetValue(t *testing.T) {
	ti := &TextInput{Value: "hello", CursorPos: 5}

	// Set shorter value - cursor should adjust
	ti.SetValue("hi")
	if ti.Value != "hi" {
		t.Errorf("Value = %q, want %q", ti.Value, "hi")
	}
	if ti.CursorPos != 2 {
		t.Errorf("CursorPos = %d, want 2", ti.CursorPos)
	}

	// Set longer value - cursor should not change
	ti.CursorPos = 1
	ti.SetValue("hello world")
	if ti.CursorPos != 1 {
		t.Errorf("CursorPos = %d, want 1", ti.CursorPos)
	}
}

func TestTextInput_Clear(t *testing.T) {
	ti := &TextInput{Value: "hello", CursorPos: 3}
	ti.Clear()
	if ti.Value != "" {
		t.Errorf("Value = %q, want empty", ti.Value)
	}
	if ti.CursorPos != 0 {
		t.Errorf("CursorPos = %d, want 0", ti.CursorPos)
	}
}

func TestTextInput_IsEmpty(t *testing.T) {
	ti := NewTextInput("")
	if !ti.IsEmpty() {
		t.Error("IsEmpty() = false for empty input, want true")
	}

	ti.Value = "x"
	if ti.IsEmpty() {
		t.Error("IsEmpty() = true for non-empty input, want false")
	}
}

func TestTextInput_Len(t *testing.T) {
	tests := []struct {
		value   string
		wantLen int
	}{
		{"", 0},
		{"hello", 5},
		{"こんにちは", 5}, // 5 runes
	}

	for _, tt := range tests {
		ti := &TextInput{Value: tt.value}
		if got := ti.Len(); got != tt.wantLen {
			t.Errorf("Len() for %q = %d, want %d", tt.value, got, tt.wantLen)
		}
	}
}

func TestTextInput_HandleKey(t *testing.T) {
	tests := []struct {
		name       string
		initial    string
		cursorPos  int
		keyMsg     tea.KeyMsg
		wantValue  string
		wantCursor int
		wantHandle bool
	}{
		{
			name:       "KeyRunes - insert character",
			initial:    "hello",
			cursorPos:  5,
			keyMsg:     tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("!")},
			wantValue:  "hello!",
			wantCursor: 6,
			wantHandle: true,
		},
		{
			name:       "KeyBackspace",
			initial:    "hello",
			cursorPos:  5,
			keyMsg:     tea.KeyMsg{Type: tea.KeyBackspace},
			wantValue:  "hell",
			wantCursor: 4,
			wantHandle: true,
		},
		{
			name:       "KeyDelete",
			initial:    "hello",
			cursorPos:  0,
			keyMsg:     tea.KeyMsg{Type: tea.KeyDelete},
			wantValue:  "ello",
			wantCursor: 0,
			wantHandle: true,
		},
		{
			name:       "KeyLeft",
			initial:    "hello",
			cursorPos:  3,
			keyMsg:     tea.KeyMsg{Type: tea.KeyLeft},
			wantValue:  "hello",
			wantCursor: 2,
			wantHandle: true,
		},
		{
			name:       "KeyRight",
			initial:    "hello",
			cursorPos:  3,
			keyMsg:     tea.KeyMsg{Type: tea.KeyRight},
			wantValue:  "hello",
			wantCursor: 4,
			wantHandle: true,
		},
		{
			name:       "KeyHome",
			initial:    "hello",
			cursorPos:  3,
			keyMsg:     tea.KeyMsg{Type: tea.KeyHome},
			wantValue:  "hello",
			wantCursor: 0,
			wantHandle: true,
		},
		{
			name:       "KeyEnd",
			initial:    "hello",
			cursorPos:  0,
			keyMsg:     tea.KeyMsg{Type: tea.KeyEnd},
			wantValue:  "hello",
			wantCursor: 5,
			wantHandle: true,
		},
		{
			name:       "KeyCtrlA",
			initial:    "hello",
			cursorPos:  3,
			keyMsg:     tea.KeyMsg{Type: tea.KeyCtrlA},
			wantValue:  "hello",
			wantCursor: 0,
			wantHandle: true,
		},
		{
			name:       "KeyCtrlE",
			initial:    "hello",
			cursorPos:  0,
			keyMsg:     tea.KeyMsg{Type: tea.KeyCtrlE},
			wantValue:  "hello",
			wantCursor: 5,
			wantHandle: true,
		},
		{
			name:       "KeyCtrlU",
			initial:    "hello world",
			cursorPos:  6,
			keyMsg:     tea.KeyMsg{Type: tea.KeyCtrlU},
			wantValue:  "world",
			wantCursor: 0,
			wantHandle: true,
		},
		{
			name:       "KeyCtrlK",
			initial:    "hello world",
			cursorPos:  5,
			keyMsg:     tea.KeyMsg{Type: tea.KeyCtrlK},
			wantValue:  "hello",
			wantCursor: 5,
			wantHandle: true,
		},
		{
			name:       "KeyEnter - not handled",
			initial:    "hello",
			cursorPos:  5,
			keyMsg:     tea.KeyMsg{Type: tea.KeyEnter},
			wantValue:  "hello",
			wantCursor: 5,
			wantHandle: false,
		},
		{
			name:       "KeyEsc - not handled",
			initial:    "hello",
			cursorPos:  5,
			keyMsg:     tea.KeyMsg{Type: tea.KeyEsc},
			wantValue:  "hello",
			wantCursor: 5,
			wantHandle: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &TextInput{Value: tt.initial, CursorPos: tt.cursorPos}
			handled := ti.HandleKey(tt.keyMsg)
			if handled != tt.wantHandle {
				t.Errorf("HandleKey() = %v, want %v", handled, tt.wantHandle)
			}
			if ti.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", ti.Value, tt.wantValue)
			}
			if ti.CursorPos != tt.wantCursor {
				t.Errorf("CursorPos = %d, want %d", ti.CursorPos, tt.wantCursor)
			}
		})
	}
}

func TestTextInput_RenderWithCursor(t *testing.T) {
	ti := NewTextInput("hello")

	// Cursor at end
	ti.CursorPos = 5
	result := ti.RenderWithCursor(0)
	if result == "" {
		t.Error("RenderWithCursor returned empty string")
	}
	// Should contain the block cursor at the end

	// Cursor in middle
	ti.CursorPos = 2
	result = ti.RenderWithCursor(0)
	if result == "" {
		t.Error("RenderWithCursor returned empty string")
	}
}
