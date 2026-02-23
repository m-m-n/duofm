package ui

import "unicode/utf8"

// OutputBuffer is a circular line buffer for storing command output.
// It stores at most maxLines entries, evicting the oldest when full.
type OutputBuffer struct {
	lines    []string
	maxLines int
	head     int // write position (next slot to overwrite)
	count    int // current number of stored lines
}

// NewOutputBuffer creates a new OutputBuffer with the given capacity.
func NewOutputBuffer(maxLines int) *OutputBuffer {
	if maxLines < 1 {
		maxLines = 1
	}
	return &OutputBuffer{
		lines:    make([]string, maxLines),
		maxLines: maxLines,
	}
}

// Append adds a line to the buffer. If the buffer is full,
// the oldest line is evicted. Non-UTF8 bytes are replaced with U+FFFD.
func (b *OutputBuffer) Append(line string) {
	if !utf8.ValidString(line) {
		line = sanitizeUTF8(line)
	}
	b.lines[b.head] = line
	b.head = (b.head + 1) % b.maxLines
	if b.count < b.maxLines {
		b.count++
	}
}

// Lines returns all buffered lines in insertion order.
func (b *OutputBuffer) Lines() []string {
	if b.count == 0 {
		return []string{}
	}
	result := make([]string, b.count)
	start := (b.head - b.count + b.maxLines) % b.maxLines
	for i := 0; i < b.count; i++ {
		result[i] = b.lines[(start+i)%b.maxLines]
	}
	return result
}

// Clear empties the buffer.
func (b *OutputBuffer) Clear() {
	b.head = 0
	b.count = 0
}

// LineCount returns the current number of stored lines.
func (b *OutputBuffer) LineCount() int {
	return b.count
}

// TailLines returns the last n lines in insertion order.
// If n exceeds the stored count, all lines are returned.
func (b *OutputBuffer) TailLines(n int) []string {
	if n <= 0 || b.count == 0 {
		return []string{}
	}
	if n > b.count {
		n = b.count
	}
	result := make([]string, n)
	start := (b.head - n + b.maxLines) % b.maxLines
	for i := 0; i < n; i++ {
		result[i] = b.lines[(start+i)%b.maxLines]
	}
	return result
}

// sanitizeUTF8 replaces invalid UTF-8 bytes with the replacement character U+FFFD.
func sanitizeUTF8(s string) string {
	var b []byte
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			b = append(b, []byte(string(utf8.RuneError))...)
		} else {
			b = append(b, s[i:i+size]...)
		}
		i += size
	}
	return string(b)
}
