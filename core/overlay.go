package core

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// OverlayDialog overlays the dialog centered on top of the base view
func OverlayDialog(baseView, dialog string, termWidth, termHeight int) string {
	if termWidth == 0 || termHeight == 0 {
		return baseView
	}

	baseLines := strings.Split(baseView, "\n")
	dialogLines := strings.Split(dialog, "\n")

	dialogHeight := len(dialogLines)
	dialogWidth := 0
	for _, line := range dialogLines {
		width := lipgloss.Width(line)
		if width > dialogWidth {
			dialogWidth = width
		}
	}

	startY := (termHeight - dialogHeight) / 2
	if startY < 0 {
		startY = 0
	}
	startX := (termWidth - dialogWidth) / 2
	if startX < 0 {
		startX = 0
	}

	for len(baseLines) < termHeight {
		baseLines = append(baseLines, "")
	}

	for i, dialogLine := range dialogLines {
		y := startY + i
		if y >= 0 && y < len(baseLines) {
			// Build the new line by rendering spaces and the dialog content
			var newLine strings.Builder

			// Add left padding (spaces before dialog)
			if startX > 0 {
				newLine.WriteString(strings.Repeat(" ", startX))
			}

			// Add the dialog line
			newLine.WriteString(dialogLine)

			// Add right padding to fill the terminal width
			dialogLineWidth := lipgloss.Width(dialogLine)
			endX := startX + dialogLineWidth
			if endX < termWidth {
				newLine.WriteString(strings.Repeat(" ", termWidth-endX))
			}

			baseLines[y] = newLine.String()
		}
	}

	return strings.Join(baseLines, "\n")
}
