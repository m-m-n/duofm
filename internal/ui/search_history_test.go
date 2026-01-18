package ui

import "testing"

func TestNewSearchHistory(t *testing.T) {
	tests := []struct {
		name    string
		maxSize int
	}{
		{
			name:    "default max size",
			maxSize: 50,
		},
		{
			name:    "small max size",
			maxSize: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewSearchHistory(tt.maxSize)
			if h == nil {
				t.Fatal("NewSearchHistory returned nil")
			}
			if h.maxSize != tt.maxSize {
				t.Errorf("maxSize = %d, want %d", h.maxSize, tt.maxSize)
			}
			if len(h.patterns) != 0 {
				t.Errorf("patterns should be empty, got %d entries", len(h.patterns))
			}
			if h.index != -1 {
				t.Errorf("index = %d, want -1", h.index)
			}
		})
	}
}

func TestSearchHistory_Add(t *testing.T) {
	tests := []struct {
		name         string
		maxSize      int
		addPatterns  []string
		wantPatterns []string
		wantLen      int
	}{
		{
			name:         "add to empty history",
			maxSize:      10,
			addPatterns:  []string{"pattern1"},
			wantPatterns: []string{"pattern1"},
			wantLen:      1,
		},
		{
			name:         "add multiple patterns",
			maxSize:      10,
			addPatterns:  []string{"first", "second", "third"},
			wantPatterns: []string{"third", "second", "first"}, // newest first
			wantLen:      3,
		},
		{
			name:         "duplicate pattern moves to front",
			maxSize:      10,
			addPatterns:  []string{"a", "b", "c", "a"},
			wantPatterns: []string{"a", "c", "b"}, // 'a' moved to front
			wantLen:      3,
		},
		{
			name:         "empty pattern is ignored",
			maxSize:      10,
			addPatterns:  []string{"a", "", "b"},
			wantPatterns: []string{"b", "a"},
			wantLen:      2,
		},
		{
			name:         "respects maxSize limit",
			maxSize:      3,
			addPatterns:  []string{"a", "b", "c", "d", "e"},
			wantPatterns: []string{"e", "d", "c"}, // oldest truncated
			wantLen:      3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewSearchHistory(tt.maxSize)
			for _, p := range tt.addPatterns {
				h.Add(p)
			}

			if len(h.patterns) != tt.wantLen {
				t.Errorf("len(patterns) = %d, want %d", len(h.patterns), tt.wantLen)
			}

			for i, want := range tt.wantPatterns {
				if i >= len(h.patterns) {
					t.Errorf("missing pattern at index %d", i)
					continue
				}
				if h.patterns[i] != want {
					t.Errorf("patterns[%d] = %q, want %q", i, h.patterns[i], want)
				}
			}
		})
	}
}

func TestSearchHistory_NavigateUp(t *testing.T) {
	tests := []struct {
		name          string
		patterns      []string
		currentInput  string
		navigateCount int
		wantResults   []string
	}{
		{
			name:          "empty history returns current input",
			patterns:      []string{},
			currentInput:  "typed",
			navigateCount: 1,
			wantResults:   []string{"typed"},
		},
		{
			name:          "navigate through history",
			patterns:      []string{"newest", "middle", "oldest"},
			currentInput:  "current",
			navigateCount: 3,
			wantResults:   []string{"newest", "middle", "oldest"},
		},
		{
			name:          "stay at oldest entry",
			patterns:      []string{"a", "b"},
			currentInput:  "input",
			navigateCount: 5,
			wantResults:   []string{"a", "b", "b", "b", "b"}, // stays at last
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewSearchHistory(50)
			// Add in reverse order so newest is at index 0
			for i := len(tt.patterns) - 1; i >= 0; i-- {
				h.Add(tt.patterns[i])
			}

			var results []string
			for i := 0; i < tt.navigateCount; i++ {
				result := h.NavigateUp(tt.currentInput)
				results = append(results, result)
			}

			for i, want := range tt.wantResults {
				if i >= len(results) {
					t.Errorf("missing result at index %d", i)
					continue
				}
				if results[i] != want {
					t.Errorf("NavigateUp[%d] = %q, want %q", i, results[i], want)
				}
			}
		})
	}
}

func TestSearchHistory_NavigateDown(t *testing.T) {
	tests := []struct {
		name         string
		patterns     []string
		currentInput string
		navigateUp   int
		navigateDown int
		wantResults  []string
	}{
		{
			name:         "navigate down returns to input",
			patterns:     []string{"a", "b", "c"},
			currentInput: "myinput",
			navigateUp:   3,
			navigateDown: 3,
			wantResults:  []string{"b", "a", "myinput"},
		},
		{
			name:         "stay at input position",
			patterns:     []string{"a"},
			currentInput: "input",
			navigateUp:   1,
			navigateDown: 3,
			wantResults:  []string{"input", "input", "input"},
		},
		{
			name:         "navigate down without up first",
			patterns:     []string{"a", "b"},
			currentInput: "test",
			navigateUp:   0,
			navigateDown: 2,
			wantResults:  []string{"", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewSearchHistory(50)
			for i := len(tt.patterns) - 1; i >= 0; i-- {
				h.Add(tt.patterns[i])
			}

			// Navigate up first
			for i := 0; i < tt.navigateUp; i++ {
				h.NavigateUp(tt.currentInput)
			}

			// Then navigate down
			var results []string
			for i := 0; i < tt.navigateDown; i++ {
				result := h.NavigateDown()
				results = append(results, result)
			}

			for i, want := range tt.wantResults {
				if i >= len(results) {
					t.Errorf("missing result at index %d", i)
					continue
				}
				if results[i] != want {
					t.Errorf("NavigateDown[%d] = %q, want %q", i, results[i], want)
				}
			}
		})
	}
}

func TestSearchHistory_Reset(t *testing.T) {
	h := NewSearchHistory(50)
	h.Add("a")
	h.Add("b")
	h.Add("c")

	// Navigate to set state
	h.NavigateUp("current")
	h.NavigateUp("current")

	if h.index == -1 {
		t.Error("index should not be -1 before reset")
	}

	// Reset
	h.Reset()

	if h.index != -1 {
		t.Errorf("index = %d after reset, want -1", h.index)
	}
	if h.editBuf != "" {
		t.Errorf("editBuf = %q after reset, want empty", h.editBuf)
	}
	// Patterns should remain
	if len(h.patterns) != 3 {
		t.Errorf("len(patterns) = %d after reset, want 3", len(h.patterns))
	}
}

func TestSearchHistory_PreserveOriginalInput(t *testing.T) {
	h := NewSearchHistory(50)
	h.Add("history1")
	h.Add("history2")

	// Start with some typed input
	originalInput := "my typed text"

	// Navigate up
	h.NavigateUp(originalInput)
	h.NavigateUp(originalInput)

	// Navigate back down to original
	h.NavigateDown()
	result := h.NavigateDown()

	if result != originalInput {
		t.Errorf("got %q, want original input %q", result, originalInput)
	}
}
