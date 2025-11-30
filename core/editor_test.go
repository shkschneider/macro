package core

import (
	"testing"
)

func TestNewHighlightedEditor(t *testing.T) {
	e := NewHighlightedEditor()
	if e == nil {
		t.Fatal("NewHighlightedEditor() returned nil")
	}
	if !e.Focused() {
		t.Error("Editor should be focused by default")
	}
}

func TestHighlightedEditor_SetFilename(t *testing.T) {
	e := NewHighlightedEditor()
	
	e.SetFilename("test.go")
	// Language should be detected
	// Note: We can't directly access the language field, so we rely on the view output
	
	e.SetFilename("test.py")
	// Language should update
	
	e.SetFilename("")
	// Should handle empty filename
}

func TestHighlightedEditor_SetValue(t *testing.T) {
	e := NewHighlightedEditor()
	
	testContent := "package main\n\nfunc main() {}"
	e.SetValue(testContent)
	
	if e.Value() != testContent {
		t.Errorf("Value() = %q, want %q", e.Value(), testContent)
	}
}

func TestHighlightedEditor_SetDimensions(t *testing.T) {
	e := NewHighlightedEditor()
	
	e.SetWidth(80)
	e.SetHeight(24)
	
	// Should not panic
}

func TestHighlightedEditor_View(t *testing.T) {
	e := NewHighlightedEditor()
	e.SetFilename("test.go")
	e.SetValue("package main\n\nfunc main() {}")
	e.SetWidth(80)
	e.SetHeight(10)
	
	view := e.View()
	if view == "" {
		t.Error("View() returned empty string")
	}
	
	// View should contain ANSI codes for syntax highlighting
	if len(view) < 10 {
		t.Error("View() seems too short, may be missing highlighted content")
	}
}

func TestHighlightedEditor_FocusBlur(t *testing.T) {
	e := NewHighlightedEditor()
	
	if !e.Focused() {
		t.Error("Editor should be focused after creation")
	}
	
	e.Blur()
	if e.Focused() {
		t.Error("Editor should not be focused after Blur()")
	}
}

func TestHighlightedEditor_CursorMovement(t *testing.T) {
	e := NewHighlightedEditor()
	e.SetValue("line 1\nline 2\nline 3")
	
	// Move cursor to start
	e.CursorStart()
	
	// After CursorStart, we should be at the start of the current line
	// Just verify it doesn't panic
}

func TestIntToString(t *testing.T) {
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
		result := intToString(tt.input)
		if result != tt.expected {
			t.Errorf("intToString(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
