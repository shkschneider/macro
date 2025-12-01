package internal

import (
	"strings"
	"testing"
)

func TestOverlayDialog_EmptyDimensions(t *testing.T) {
	baseView := "base content"
	dialog := "dialog"

	// Test with zero width
	result := OverlayDialog(baseView, dialog, 0, 10)
	if result != baseView {
		t.Error("OverlayDialog should return baseView when termWidth is 0")
	}

	// Test with zero height
	result = OverlayDialog(baseView, dialog, 10, 0)
	if result != baseView {
		t.Error("OverlayDialog should return baseView when termHeight is 0")
	}
}

func TestOverlayDialog_BasicOverlay(t *testing.T) {
	baseView := "line1\nline2\nline3\nline4\nline5"
	dialog := "DLG"

	result := OverlayDialog(baseView, dialog, 20, 5)

	// Result should contain the dialog content
	if !strings.Contains(result, "DLG") {
		t.Error("OverlayDialog result should contain dialog content")
	}
}

func TestTruncateStringToWidth_Empty(t *testing.T) {
	result := truncateStringToWidth("hello", 0)
	if result != "" {
		t.Error("truncateStringToWidth should return empty string for width 0")
	}

	result = truncateStringToWidth("hello", -1)
	if result != "" {
		t.Error("truncateStringToWidth should return empty string for negative width")
	}
}

func TestTruncateStringToWidth_Normal(t *testing.T) {
	// Test truncation
	result := truncateStringToWidth("hello world", 5)
	// Should be 5 characters (may have padding)
	if len(result) < 5 {
		t.Error("truncateStringToWidth should return at least maxWidth characters")
	}

	// Test no truncation needed
	result = truncateStringToWidth("hi", 10)
	// Should be padded to 10 characters
	if len(result) != 10 {
		t.Errorf("truncateStringToWidth should pad to maxWidth, got %d", len(result))
	}
}

func TestTruncateStringToWidth_WithANSI(t *testing.T) {
	// ANSI escape codes should not count towards width
	ansiString := "\x1b[31mred\x1b[0m"
	result := truncateStringToWidth(ansiString, 10)

	// Result should contain the ANSI codes
	if !strings.Contains(result, "\x1b[") {
		t.Error("truncateStringToWidth should preserve ANSI codes")
	}
}

func TestExtractStringFromWidth_Empty(t *testing.T) {
	result := extractStringFromWidth("hello", 5, 3)
	if result != "" {
		t.Error("extractStringFromWidth should return empty when startWidth >= endWidth")
	}

	result = extractStringFromWidth("hello", 5, 5)
	if result != "" {
		t.Error("extractStringFromWidth should return empty when startWidth == endWidth")
	}
}

func TestExtractStringFromWidth_Normal(t *testing.T) {
	result := extractStringFromWidth("hello world", 6, 11)
	if result != "world" {
		t.Errorf("extractStringFromWidth should extract 'world', got '%s'", result)
	}
}
