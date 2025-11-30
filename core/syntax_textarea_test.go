package core

import (
	"strings"
	"testing"
)

func TestNewSyntaxTextarea(t *testing.T) {
	sta := NewSyntaxTextarea()
	if sta == nil {
		t.Fatal("NewSyntaxTextarea() returned nil")
	}
	if !sta.Focused() {
		t.Error("SyntaxTextarea should be focused after creation")
	}
}

func TestSyntaxTextarea_SetFilename(t *testing.T) {
	sta := NewSyntaxTextarea()

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

func TestSyntaxTextarea_SetValue(t *testing.T) {
	sta := NewSyntaxTextarea()
	
	testContent := "package main\n\nfunc main() {}"
	sta.SetValue(testContent)
	
	if sta.Value() != testContent {
		t.Errorf("Value() = %q, want %q", sta.Value(), testContent)
	}
}

func TestSyntaxTextarea_SetDimensions(t *testing.T) {
	sta := NewSyntaxTextarea()
	
	sta.SetWidth(80)
	sta.SetHeight(24)
	
	// Should not panic
}

func TestSyntaxTextarea_View(t *testing.T) {
	sta := NewSyntaxTextarea()
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

func TestSyntaxTextarea_FocusBlur(t *testing.T) {
	sta := NewSyntaxTextarea()
	
	if !sta.Focused() {
		t.Error("SyntaxTextarea should be focused after creation")
	}
	
	sta.Blur()
	if sta.Focused() {
		t.Error("SyntaxTextarea should not be focused after Blur()")
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
