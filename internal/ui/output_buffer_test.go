package ui

import (
	"testing"
)

func TestNewOutputBuffer(t *testing.T) {
	buf := NewOutputBuffer(100)
	if buf == nil {
		t.Fatal("NewOutputBuffer returned nil")
	}
	if len(buf.Lines()) != 0 {
		t.Errorf("new buffer should be empty, got %d lines", len(buf.Lines()))
	}
}

func TestOutputBuffer_Append(t *testing.T) {
	buf := NewOutputBuffer(10)
	buf.Append("line1")
	buf.Append("line2")

	lines := buf.Lines()
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" {
		t.Errorf("expected [line1, line2], got %v", lines)
	}
}

func TestOutputBuffer_CircularEviction(t *testing.T) {
	buf := NewOutputBuffer(3)
	buf.Append("a")
	buf.Append("b")
	buf.Append("c")
	buf.Append("d") // evicts "a"

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "b" || lines[1] != "c" || lines[2] != "d" {
		t.Errorf("expected [b, c, d], got %v", lines)
	}
}

func TestOutputBuffer_CircularEviction_WrapAround(t *testing.T) {
	buf := NewOutputBuffer(3)
	for i := 0; i < 10; i++ {
		buf.Append(string(rune('a' + i)))
	}

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	// Last 3 should be h, i, j
	if lines[0] != "h" || lines[1] != "i" || lines[2] != "j" {
		t.Errorf("expected [h, i, j], got %v", lines)
	}
}

func TestOutputBuffer_Clear(t *testing.T) {
	buf := NewOutputBuffer(10)
	buf.Append("line1")
	buf.Append("line2")
	buf.Clear()

	lines := buf.Lines()
	if len(lines) != 0 {
		t.Errorf("expected empty after clear, got %d lines", len(lines))
	}

	// Should be able to append after clear
	buf.Append("new")
	lines = buf.Lines()
	if len(lines) != 1 || lines[0] != "new" {
		t.Errorf("expected [new] after clear+append, got %v", lines)
	}
}

func TestOutputBuffer_EmptyLines(t *testing.T) {
	buf := NewOutputBuffer(10)
	lines := buf.Lines()
	if lines == nil {
		t.Error("Lines() should return non-nil empty slice")
	}
	if len(lines) != 0 {
		t.Errorf("expected 0 lines, got %d", len(lines))
	}
}

func TestOutputBuffer_Unicode(t *testing.T) {
	buf := NewOutputBuffer(10)
	buf.Append("日本語テスト")
	buf.Append("emoji: 🎉")
	buf.Append("中文测试")

	lines := buf.Lines()
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "日本語テスト" {
		t.Errorf("unicode not preserved: got %q", lines[0])
	}
	if lines[1] != "emoji: 🎉" {
		t.Errorf("emoji not preserved: got %q", lines[1])
	}
}

func TestOutputBuffer_SingleCapacity(t *testing.T) {
	buf := NewOutputBuffer(1)
	buf.Append("first")
	buf.Append("second")

	lines := buf.Lines()
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	if lines[0] != "second" {
		t.Errorf("expected 'second', got %q", lines[0])
	}
}

func TestOutputBuffer_BinaryOutputSanitized(t *testing.T) {
	buf := NewOutputBuffer(10)
	// Binary data with invalid UTF-8 bytes
	buf.Append("hello\x80\x81world")
	lines := buf.Lines()
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	expected := "hello\uFFFD\uFFFDworld"
	if lines[0] != expected {
		t.Errorf("expected sanitized output %q, got %q", expected, lines[0])
	}
}

func TestOutputBuffer_ValidUTF8Unmodified(t *testing.T) {
	buf := NewOutputBuffer(10)
	input := "valid utf8: あいう"
	buf.Append(input)
	lines := buf.Lines()
	if lines[0] != input {
		t.Errorf("valid UTF-8 should be unmodified, got %q", lines[0])
	}
}

func TestOutputBuffer_LineCount(t *testing.T) {
	buf := NewOutputBuffer(10)
	if buf.LineCount() != 0 {
		t.Errorf("expected 0 lines, got %d", buf.LineCount())
	}
	buf.Append("a")
	buf.Append("b")
	if buf.LineCount() != 2 {
		t.Errorf("expected 2 lines, got %d", buf.LineCount())
	}

	// After overflow
	buf2 := NewOutputBuffer(2)
	buf2.Append("a")
	buf2.Append("b")
	buf2.Append("c")
	if buf2.LineCount() != 2 {
		t.Errorf("expected 2 lines after overflow, got %d", buf2.LineCount())
	}
}
