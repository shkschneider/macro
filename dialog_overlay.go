package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// overlayDialog overlays the dialog centered on top of the base view
func overlayDialog(baseView, dialog string, termWidth, termHeight int) string {
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
			baseLine := baseLines[y]
			baseWidth := lipgloss.Width(baseLine)

			if baseWidth < termWidth {
				baseLine += strings.Repeat(" ", termWidth-baseWidth)
			}

			dialogLineWidth := lipgloss.Width(dialogLine)

			var newLine strings.Builder

			if startX > 0 {
				leftPart := baseLine
				if len(leftPart) > startX {
					leftPart = leftPart[:startX]
				}
				newLine.WriteString(leftPart)
			}

			newLine.WriteString(dialogLine)

			endX := startX + dialogLineWidth
			if endX < baseWidth {
				rightPart := baseLine[endX:]
				newLine.WriteString(rightPart)
			}

			baseLines[y] = newLine.String()
		}
	}

	return strings.Join(baseLines, "\n")
}
