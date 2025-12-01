package internal

import (
	"strings"
	"testing"
)

func TestNewTextarea(t *testing.T) {
	sta := NewTextarea()
	if sta == nil {
		t.Fatal("NewTextarea() returned nil")
	}
	if !sta.Focused() {
		t.Error("Textarea should be focused after creation")
	}
}

func TestTextarea_SetFilename(t *testing.T) {
	sta := NewTextarea()

	sta.SetFilename("test.go")
	if sta.GetLanguage() != "Go" {
		t.Errorf("Expected language 'Go', got '%s'", sta.GetLanguage())
	}

	sta.SetFilename("test.py")
	if sta.GetLanguage() != "Python" {
		t.Errorf("Expected language 'Python', got '%s'", sta.GetLanguage())
	}

	sta.SetFilename("test.unknown")
	// Unknown extension should return empty
}

func TestTextarea_SetValue(t *testing.T) {
	sta := NewTextarea()
	
	testContent := "package main\n\nfunc main() {}"
	sta.SetValue(testContent)
	
	if sta.Value() != testContent {
		t.Errorf("Value() = %q, want %q", sta.Value(), testContent)
	}
}

func TestTextarea_SetDimensions(t *testing.T) {
	sta := NewTextarea()
	
	sta.SetWidth(80)
	sta.SetHeight(24)
	
	// Should not panic
}

func TestTextarea_View(t *testing.T) {
	sta := NewTextarea()
	sta.SetFilename("test.go")
	sta.SetValue("package main\n\nfunc main() {}")
	sta.SetWidth(80)
	sta.SetHeight(10)
	
	view := sta.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
	
	// View should contain line numbers
	if !strings.Contains(view, "1") {
		t.Error("View() should contain line number 1")
	}
}

func TestTextarea_FocusBlur(t *testing.T) {
	sta := NewTextarea()
	
	if !sta.Focused() {
		t.Error("Textarea should be focused after creation")
	}
	
	sta.Blur()
	if sta.Focused() {
		t.Error("Textarea should not be focused after Blur()")
	}
}

func TestIntToStr(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{123, "123"},
		{9999, "9999"},
	}
	
	for _, tt := range tests {
		result := intToStr(tt.input)
		if result != tt.expected {
			t.Errorf("intToStr(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestTextarea_CursorPosition(t *testing.T) {
	sta := NewTextarea()
	sta.SetValue("line 1\nline 2\nline 3\nline 4\nline 5")
	
	// Move cursor to start (line 0)
	sta.SetCursorPosition(0, 0)
	
	if sta.Line() != 0 {
		t.Errorf("After SetCursorPosition(0, 0), line should be 0, got %d", sta.Line())
	}
	
	// Move cursor to line 2, column 3
	sta.SetCursorPosition(2, 3)
	
	if sta.Line() != 2 {
		t.Errorf("Expected line 2, got %d", sta.Line())
	}
	
	// Test Column() method
	col := sta.Column()
	if col != 3 {
		t.Errorf("Expected column 3, got %d", col)
	}
}

func TestTextarea_CursorPositionBounds(t *testing.T) {
	sta := NewTextarea()
	sta.SetValue("line 1\nline 2\nline 3")
	
	// Test setting cursor beyond last line
	sta.SetCursorPosition(100, 0)
	if sta.Line() > 2 {
		t.Errorf("Line should not exceed max, got %d", sta.Line())
	}
	
	// Test setting negative line
	sta.SetCursorPosition(-5, 0)
	if sta.Line() < 0 {
		t.Errorf("Line should not be negative, got %d", sta.Line())
	}
}

func TestTextarea_LineCount(t *testing.T) {
	sta := NewTextarea()
	sta.SetValue("line 1\nline 2\nline 3\nline 4")
	
	if sta.LineCount() != 4 {
		t.Errorf("Expected 4 lines, got %d", sta.LineCount())
	}
}

func TestTextarea_CursorMovement(t *testing.T) {
	sta := NewTextarea()
	sta.SetValue("line 1\nline 2\nline 3")
	
	// First move cursor to top
	sta.SetCursorPosition(0, 0)
	
	// Move down
	sta.CursorDown()
	if sta.Line() != 1 {
		t.Errorf("After CursorDown, expected line 1, got %d", sta.Line())
	}
	
	// Move up
	sta.CursorUp()
	if sta.Line() != 0 {
		t.Errorf("After CursorUp, expected line 0, got %d", sta.Line())
	}
}

func TestTextarea_MultiCursor_AddSecondaryCursor(t *testing.T) {
	sta := NewTextarea()
	sta.SetValue("hello world\nhello again\nhello there")
	
	// Initially no secondary cursors
	if sta.HasSecondaryCursors() {
		t.Error("Should not have secondary cursors initially")
	}
	
	// Add a secondary cursor
	sta.AddSecondaryCursor(1, 0)
	
	if !sta.HasSecondaryCursors() {
		t.Error("Should have secondary cursors after adding one")
	}
	
	if sta.SecondaryCursorCount() != 1 {
		t.Errorf("Expected 1 secondary cursor, got %d", sta.SecondaryCursorCount())
	}
}

func TestTextarea_MultiCursor_ClearSecondaryCursors(t *testing.T) {
	sta := NewTextarea()
	sta.SetValue("hello world\nhello again")
	
	sta.AddSecondaryCursor(1, 0)
	sta.AddSecondaryCursor(0, 6)
	
	if sta.SecondaryCursorCount() != 2 {
		t.Errorf("Expected 2 secondary cursors, got %d", sta.SecondaryCursorCount())
	}
	
	sta.ClearSecondaryCursors()
	
	if sta.HasSecondaryCursors() {
		t.Error("Should not have secondary cursors after clearing")
	}
}

func TestTextarea_MultiCursor_NoDuplicates(t *testing.T) {
	sta := NewTextarea()
	sta.SetValue("hello world")
	
	// Add same cursor twice
	sta.AddSecondaryCursor(0, 6)
	sta.AddSecondaryCursor(0, 6)
	
	if sta.SecondaryCursorCount() != 1 {
		t.Errorf("Should not add duplicate cursors, got %d", sta.SecondaryCursorCount())
	}
}

func TestTextarea_MultiCursor_AddCursorAtNextOccurrence(t *testing.T) {
	sta := NewTextarea()
	sta.SetValue("hello world\nhello again\nhello there")
	
	// Move cursor to "hello" on first line
	sta.SetCursorPosition(0, 0)
	
	// Add cursor at next occurrence
	added := sta.AddCursorAtNextOccurrence()
	if !added {
		t.Error("Should have found next occurrence of 'hello'")
	}
	
	if sta.SecondaryCursorCount() != 1 {
		t.Errorf("Expected 1 secondary cursor, got %d", sta.SecondaryCursorCount())
	}
	
	// Add another
	added = sta.AddCursorAtNextOccurrence()
	if !added {
		t.Error("Should have found another occurrence of 'hello'")
	}
	
	if sta.SecondaryCursorCount() != 2 {
		t.Errorf("Expected 2 secondary cursors, got %d", sta.SecondaryCursorCount())
	}
}

func TestTextarea_MultiCursor_GetAllCursorPositions(t *testing.T) {
	sta := NewTextarea()
	sta.SetValue("hello world")
	
	// Primary cursor starts at (0, 0)
	sta.SetCursorPosition(0, 0)
	
	// Add secondary cursors
	sta.AddSecondaryCursor(0, 6)
	
	positions := sta.GetAllCursorPositions()
	
	if len(positions) != 2 {
		t.Errorf("Expected 2 cursor positions, got %d", len(positions))
	}
	
	// First should be primary
	if positions[0].Line != 0 || positions[0].Column != 0 {
		t.Errorf("Primary cursor should be at (0, 0), got (%d, %d)", positions[0].Line, positions[0].Column)
	}
	
	// Second should be secondary
	if positions[1].Line != 0 || positions[1].Column != 6 {
		t.Errorf("Secondary cursor should be at (0, 6), got (%d, %d)", positions[1].Line, positions[1].Column)
	}
}

func TestIsWordChar(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{'a', true},
		{'z', true},
		{'A', true},
		{'Z', true},
		{'0', true},
		{'9', true},
		{'_', true},
		{' ', false},
		{'.', false},
		{'-', false},
		{'(', false},
		{')', false},
	}
	
	for _, tt := range tests {
		result := isWordChar(tt.char)
		if result != tt.expected {
			t.Errorf("isWordChar(%q) = %v, want %v", tt.char, result, tt.expected)
		}
	}
}
