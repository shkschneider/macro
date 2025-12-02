package internal

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dustin/go-humanize"
)

// buildStatusBar creates the formatted status bar with left and right sections
// Left: "filename.ext* [language] | human filesize"
// Right: "line:col [RO] [fileencoding] [directory/path]"
func (m *Model) BuildStatusBar() string {
	buf := m.getCurrentBuffer()
	if buf == nil {
		return StatusBarStyle.Width(TermWidth).Render("New File")
	}

	// Get file info
	fileName := filepath.Base(buf.FilePath)
	lang := DetectLanguage(buf.FilePath)
	dirPath := filepath.Dir(buf.FilePath)
	modified := m.isCurrentBufferModified()
	readOnly := buf.ReadOnly

	// Build left side: "filename.ext* [language] | human filesize"
	leftParts := []string{}

	// Filename with modification indicator
	if modified {
		leftParts = append(leftParts, fileName+"*")
	} else {
		leftParts = append(leftParts, fileName)
	}

	// Language
	if lang != "" {
		leftParts = append(leftParts, "["+lang+"]")
	}

	// File size
	leftParts = append(leftParts, humanize.Bytes(uint64(buf.FileSize)))

	if readOnly {
		leftParts = append(leftParts, "(read-only)")
	}

	leftSection := strings.Join(leftParts, " ")

	// Build right side: "line:col [RO] [fileencoding] [directory/path]"
	rightParts := []string{}

	// File encoding (assuming UTF-8 as default since we're reading text files)
	// rightParts = append(rightParts, "[utf-8]")

	// Cursor position (line:column)
	line, col := m.Textarea.CursorPosition()
	rightParts = append(rightParts, fmt.Sprintf("%d:%d", line, col))

	// Directory path
	rightParts = append(rightParts, "["+dirPath+"/]")

	rightSection := strings.Join(rightParts, " ")

	// Calculate padding needed to align right section
	// Account for StatusBarStyle padding (1 on each side = 2 total)
	padding := 2
	contentWidth := TermWidth - padding
	leftLen := len(leftSection)
	rightLen := len(rightSection)
	spacesNeeded := contentWidth - leftLen - rightLen
	if spacesNeeded < 1 {
		spacesNeeded = 1
	}

	fullStatusContent := leftSection + strings.Repeat(" ", spacesNeeded) + rightSection
	return StatusBarStyle.Width(TermWidth).Render(fullStatusContent)
}
