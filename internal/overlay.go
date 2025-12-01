package internal

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
			baseLine := baseLines[y]
			baseLineWidth := lipgloss.Width(baseLine)
			dialogLineWidth := lipgloss.Width(dialogLine)
			endX := startX + dialogLineWidth

			// Build the new line by preserving base view content outside dialog bounds
			var newLine strings.Builder

			// Preserve left portion of base line (before dialog)
			if startX > 0 {
				if baseLineWidth > startX {
					// Extract the visible part of the base line before dialog
					leftPart := truncateStringToWidth(baseLine, startX)
					newLine.WriteString(leftPart)
				} else {
					// Base line is shorter, pad with spaces
					newLine.WriteString(baseLine)
					if baseLineWidth < startX {
						newLine.WriteString(strings.Repeat(" ", startX-baseLineWidth))
					}
				}
			}

			// Add the dialog line
			newLine.WriteString(dialogLine)

			// Preserve right portion of base line (after dialog)
			if endX < termWidth && baseLineWidth > endX {
				// Extract the visible part of the base line after dialog
				rightPart := extractStringFromWidth(baseLine, endX, baseLineWidth)
				newLine.WriteString(rightPart)
			} else if endX < termWidth {
				// Base line doesn't extend beyond dialog, pad with spaces
				newLine.WriteString(strings.Repeat(" ", termWidth-endX))
			}

			baseLines[y] = newLine.String()
		}
	}

	return strings.Join(baseLines, "\n")
}

// truncateStringToWidth truncates a string to fit within the specified visible width,
// accounting for ANSI escape sequences and multi-byte characters
func truncateStringToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	var result strings.Builder
	currentWidth := 0
	inEscape := false

	for _, r := range s {
		// Handle ANSI escape sequences (don't count towards width)
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}
		if inEscape {
			result.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		// Calculate width of the character (1 for most, 2 for wide chars)
		charWidth := lipgloss.Width(string(r))
		if currentWidth+charWidth > maxWidth {
			break
		}

		result.WriteRune(r)
		currentWidth += charWidth
	}

	// Pad with spaces if needed
	if currentWidth < maxWidth {
		result.WriteString(strings.Repeat(" ", maxWidth-currentWidth))
	}

	return result.String()
}

// extractStringFromWidth extracts a substring starting from a given visible width position,
// accounting for ANSI escape sequences and multi-byte characters
func extractStringFromWidth(s string, startWidth, endWidth int) string {
	if startWidth >= endWidth {
		return ""
	}

	var result strings.Builder
	currentWidth := 0
	inEscape := false
	capturing := false
	var pendingEscapes strings.Builder

	for _, r := range s {
		// Handle ANSI escape sequences (don't count towards width)
		if r == '\x1b' {
			inEscape = true
			if capturing {
				result.WriteRune(r)
			} else {
				pendingEscapes.WriteRune(r)
			}
			continue
		}
		if inEscape {
			if capturing {
				result.WriteRune(r)
			} else {
				pendingEscapes.WriteRune(r)
			}
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		// Calculate width of the character
		charWidth := lipgloss.Width(string(r))

		// Check if we've reached the start position
		if !capturing && currentWidth >= startWidth {
			capturing = true
			// Add any pending escape sequences from before start
			if pendingEscapes.Len() > 0 {
				result.WriteString(pendingEscapes.String())
				pendingEscapes.Reset()
			}
		}

		// Add character if we're in the capture range
		if capturing {
			if currentWidth >= endWidth {
				break
			}
			result.WriteRune(r)
		}

		currentWidth += charWidth

		// Clear pending escapes if we haven't started capturing yet
		if !capturing && !inEscape {
			pendingEscapes.Reset()
		}
	}

	return result.String()
}
