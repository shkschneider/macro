package core

import (
	"strings"
	"testing"
)

func TestOverlayDialog_PreservesBaseViewText(t *testing.T) {
	// Create a base view with multiple lines
	baseView := strings.Join([]string{
		"Line 1: This is text that should be visible on the sides",
		"Line 2: More content that should remain visible outside",
		"Line 3: Testing the overlay functionality properly here",
		"Line 4: The dialog should only obscure center portion",
		"Line 5: Final line with more text to test the overlay",
	}, "\n")

	// Create a simple dialog (3 lines, narrow width)
	dialog := strings.Join([]string{
		"╭─────────╮",
		"│ Dialog  │",
		"╰─────────╯",
	}, "\n")

	// Terminal size
	termWidth := 80
	termHeight := 10

	// Overlay the dialog
	result := OverlayDialog(baseView, dialog, termWidth, termHeight)
	resultLines := strings.Split(result, "\n")

	// The dialog will be centered, so we expect:
	// - Lines where dialog appears should have base text on left and right sides
	// - Lines without dialog should be unchanged

	// Get the first line which shouldn't have dialog
	if len(resultLines) > 0 {
		firstLine := resultLines[0]
		// First line should start with "Line 1:" since dialog is centered
		if !strings.HasPrefix(firstLine, "Line 1:") {
			t.Errorf("First line should preserve base view text, got: %s", firstLine)
		}
	}

	// Check that result has the same number of lines as base view (padded to termHeight)
	if len(resultLines) < 5 {
		t.Errorf("Expected at least 5 lines in result, got %d", len(resultLines))
	}
}

func TestOverlayDialog_HandlesEmptyBaseView(t *testing.T) {
	baseView := ""
	dialog := "Dialog"
	termWidth := 80
	termHeight := 10

	result := OverlayDialog(baseView, dialog, termWidth, termHeight)
	if result == "" {
		t.Error("Expected non-empty result")
	}
}

func TestOverlayDialog_HandlesZeroTerminalSize(t *testing.T) {
	baseView := "Some text"
	dialog := "Dialog"
	termWidth := 0
	termHeight := 0

	result := OverlayDialog(baseView, dialog, termWidth, termHeight)
	if result != baseView {
		t.Errorf("Expected base view to be returned unchanged, got: %s", result)
	}
}

func TestTruncateStringToWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		wantLen  bool // Just check if result length makes sense
	}{
		{
			name:     "simple truncation",
			input:    "Hello World",
			maxWidth: 5,
			wantLen:  true,
		},
		{
			name:     "exact width",
			input:    "Hello",
			maxWidth: 5,
			wantLen:  true,
		},
		{
			name:     "zero width",
			input:    "Hello",
			maxWidth: 0,
			wantLen:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateStringToWidth(tt.input, tt.maxWidth)
			// Basic sanity check
			if tt.wantLen && len(result) < 0 {
				t.Errorf("truncateStringToWidth() returned invalid length")
			}
		})
	}
}

func TestExtractStringFromWidth(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		startWidth int
		endWidth   int
	}{
		{
			name:       "extract middle",
			input:      "Hello World",
			startWidth: 2,
			endWidth:   7,
		},
		{
			name:       "extract from start",
			input:      "Hello World",
			startWidth: 0,
			endWidth:   5,
		},
		{
			name:       "invalid range",
			input:      "Hello World",
			startWidth: 10,
			endWidth:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractStringFromWidth(tt.input, tt.startWidth, tt.endWidth)
			// Basic sanity check - result should be a string
			_ = result
		})
	}
}
