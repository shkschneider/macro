// Package core provides core functionality for the macro editor.
package internal

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Textarea wraps a textarea with syntax highlighting overlay.
// The textarea handles all input, while a highlighted version is displayed.
type Textarea struct {
	textarea textarea.Model
	filename string
	language string
	width    int
	height   int

	// Style for line numbers
	lineNumberStyle lipgloss.Style
	// Style for cursor line highlight
	cursorLineStyle lipgloss.Style

	// Diff tracking for showing changes
	diffTracker *DiffTracker
}

// NewTextarea creates a new syntax-highlighted textarea.
func NewTextarea() *Textarea {
	ta := textarea.New()
	ta.Focus()
	ta.Prompt = ""
	ta.ShowLineNumbers = false // We'll render our own with syntax highlighting

	// Make textarea text style minimal - we'll overlay highlighted text
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Text = lipgloss.NewStyle()

	return &Textarea{
		textarea: ta,
		lineNumberStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),
		cursorLineStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("236")),
		diffTracker: NewDiffTracker(),
	}
}

// SetFilename sets the filename for language detection and diff tracking.
func (s *Textarea) SetFilename(filename string) {
	s.filename = filename
	s.language = DetectLanguage(filename)
	// Also set file path for git diff tracking
	s.diffTracker.SetFilePath(filename)
}

// SetOriginalContent stores the original file content for diff tracking.
// Call this when loading a file to enable change indicators.
// Note: This is now a no-op since we use git diff directly with the file path.
func (s *Textarea) SetOriginalContent(content string) {
	s.diffTracker.SetOriginal(content)
}

// ClearOriginalContent clears the original content (e.g., for new files).
func (s *Textarea) ClearOriginalContent() {
	s.diffTracker.ClearOriginal()
}

// SetLanguage explicitly sets the language for highlighting.
func (s *Textarea) SetLanguage(language string) {
	s.language = language
}

// SetValue sets the textarea content.
func (s *Textarea) SetValue(value string) {
	s.textarea.SetValue(value)
}

// Value returns the current content.
func (s *Textarea) Value() string {
	return s.textarea.Value()
}

// SetWidth sets the width.
func (s *Textarea) SetWidth(w int) {
	s.width = w
	// Account for line numbers (4 chars) + diff indicator (1 char) + space (1 char) = 6 chars
	s.textarea.SetWidth(w - 6)
}

// SetHeight sets the height.
func (s *Textarea) SetHeight(h int) {
	s.height = h
	s.textarea.SetHeight(h)
}

// Focus focuses the textarea.
func (s *Textarea) Focus() tea.Cmd {
	return s.textarea.Focus()
}

// Blur blurs the textarea.
func (s *Textarea) Blur() {
	s.textarea.Blur()
}

// Focused returns whether the textarea is focused.
func (s *Textarea) Focused() bool {
	return s.textarea.Focused()
}

// CursorStart moves cursor to start of current line.
func (s *Textarea) CursorStart() {
	s.textarea.CursorStart()
}

// CursorUp moves cursor up one line.
func (s *Textarea) CursorUp() {
	s.textarea.CursorUp()
}

// Line returns the current line number (0-indexed).
func (s *Textarea) Line() int {
	return s.textarea.Line()
}

// Column returns the current column position (0-indexed).
func (s *Textarea) Column() int {
	return s.textarea.LineInfo().CharOffset
}

// CursorDown moves cursor down one line.
func (s *Textarea) CursorDown() {
	s.textarea.CursorDown()
}

// SetCursor sets the cursor column position within the current line.
func (s *Textarea) SetCursor(col int) {
	s.textarea.SetCursor(col)
}

// LineCount returns the total number of lines.
func (s *Textarea) LineCount() int {
	return s.textarea.LineCount()
}

// SetCursorPosition moves the cursor to the specified line and column.
// It first moves to the top, then moves down to the target line, then sets the column.
func (s *Textarea) SetCursorPosition(line, column int) {
	// Move to start
	s.textarea.CursorStart()
	for s.textarea.Line() > 0 {
		s.textarea.CursorUp()
	}

	// Move to target line
	lineCount := s.textarea.LineCount()
	if line >= lineCount {
		line = lineCount - 1
	}
	if line < 0 {
		line = 0
	}

	for s.textarea.Line() < line {
		s.textarea.CursorDown()
	}

	// Set column position
	if column > 0 {
		s.textarea.SetCursor(column)
	}
}

// Update handles messages and updates the textarea state.
func (s *Textarea) Update(msg tea.Msg) (*Textarea, tea.Cmd) {
	var cmd tea.Cmd
	s.textarea, cmd = s.textarea.Update(msg)
	return s, cmd
}

// View renders the syntax-highlighted textarea.
func (s *Textarea) View() string {
	content := s.textarea.Value()
	lines := strings.Split(content, "\n")

	// Get highlighted lines
	highlightedContent := HighlightCode(content, s.filename, s.language)
	highlightedLines := strings.Split(highlightedContent, "\n")

	// Compute diff states for change indicators
	lineStates, deletedAt := s.diffTracker.ComputeLineStatesWithDeletions(content)

	// Get cursor position from textarea
	cursorLine := s.textarea.Line()
	lineInfo := s.textarea.LineInfo()
	cursorCol := lineInfo.ColumnOffset

	// Calculate which lines to display based on textarea's viewport
	// The textarea handles scrolling internally, we need to sync with it
	startLine := 0
	visibleLines := s.height
	if visibleLines <= 0 {
		visibleLines = len(lines)
	}

	// Try to keep cursor visible by centering around it
	if len(lines) > visibleLines {
		halfHeight := visibleLines / 2
		startLine = cursorLine - halfHeight
		if startLine < 0 {
			startLine = 0
		}
		if startLine+visibleLines > len(lines) {
			startLine = len(lines) - visibleLines
			if startLine < 0 {
				startLine = 0
			}
		}
	}

	endLine := startLine + visibleLines
	if endLine > len(lines) {
		endLine = len(lines)
	}

	var result strings.Builder
	for i := startLine; i < endLine; i++ {
		// Line number with manual padding (right-aligned in 4 chars)
		numStr := intToStr(i + 1)
		padding := ""
		if len(numStr) < 4 {
			padding = strings.Repeat(" ", 4-len(numStr))
		}
		lineNum := s.lineNumberStyle.Render(padding + numStr)

		// Get diff indicator for this line
		diffIndicator := s.getDiffIndicator(i, lineStates, deletedAt)

		// Get the highlighted line content
		var lineContent string
		if i < len(highlightedLines) {
			lineContent = highlightedLines[i]
		} else {
			lineContent = ""
		}

		// If this is the cursor line, show the cursor
		if i == cursorLine && s.textarea.Focused() {
			lineContent = s.insertCursor(lines[i], highlightedLines[i], cursorCol)
		}

		result.WriteString(lineNum)
		result.WriteString(diffIndicator)
		result.WriteString(" ")
		result.WriteString(lineContent)

		if i < endLine-1 {
			result.WriteString("\n")
		}
	}

	// Pad with empty lines if needed
	for i := endLine - startLine; i < visibleLines; i++ {
		if i > 0 {
			result.WriteString("\n")
		}
		lineNum := s.lineNumberStyle.Render("   ~")
		result.WriteString(lineNum)
		result.WriteString(" ") // Space for diff indicator column
		result.WriteString(" ")
	}

	return result.String()
}

// getDiffIndicator returns a colored "|" indicator based on the line's diff state.
// - Green "|" for added lines (new lines that didn't exist in original)
// - Yellow "|" for modified lines (content changed from original)
// - Red "|" for positions where lines were deleted from original
func (s *Textarea) getDiffIndicator(lineIdx int, lineStates []LineState, deletedAt []bool) string {
	// First check if there's a deleted line at this position
	if lineIdx < len(deletedAt) && deletedAt[lineIdx] {
		return DiffDeletedStyle.Render("|")
	}

	// Then check the state of the current line
	if lineIdx >= len(lineStates) {
		return " " // No diff info, return space
	}

	switch lineStates[lineIdx] {
	case LineAdded:
		return DiffAddedStyle.Render("|")
	case LineModified:
		return DiffModifiedStyle.Render("|")
	default:
		return " " // Unchanged, just a space
	}
}

// insertCursor inserts a visible cursor at the specified column position.
func (s *Textarea) insertCursor(plainLine, highlightedLine string, col int) string {
	// Get the character at cursor position or space if at end
	plainRunes := []rune(plainLine)

	// If cursor is at or past end of line, append a reverse-video space
	if col >= len(plainRunes) {
		return highlightedLine + "\x1b[7m \x1b[27m"
	}

	// We need to find the position in the highlighted string where the visible
	// character at 'col' starts, and extract the styled character with its ANSI codes.
	highlightedRunes := []rune(highlightedLine)
	visibleCol := 0
	charStartPos := -1
	charEndPos := -1
	inEscape := false

	for i := 0; i < len(highlightedRunes); i++ {
		r := highlightedRunes[i]

		// Track ANSI escape sequences
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		// This is a visible character
		if visibleCol == col {
			charStartPos = i
			charEndPos = i + 1
			break
		}
		visibleCol++
	}

	// If we didn't find the position, append cursor at end
	if charStartPos == -1 {
		return highlightedLine + "\x1b[7m \x1b[27m"
	}

	// Build result: before cursor char + reverse video on + char + reverse video off + rest of line
	// This preserves the color that was set before the character
	before := string(highlightedRunes[:charStartPos])
	cursorChar := string(highlightedRunes[charStartPos:charEndPos])
	after := ""
	if charEndPos < len(highlightedRunes) {
		after = string(highlightedRunes[charEndPos:])
	}

	// Use \x1b[7m to turn ON reverse video, \x1b[27m to turn it OFF
	// This preserves any foreground/background colors that were set
	return before + "\x1b[7m" + cursorChar + "\x1b[27m" + after
}

// intToStr converts an integer to string.
func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// GetLanguage returns the detected or set language.
func (s *Textarea) GetLanguage() string {
	return s.language
}

// CursorPosition returns the current cursor position as (line, column), both 1-indexed.
func (s *Textarea) CursorPosition() (int, int) {
	line := s.textarea.Line() + 1 // Convert from 0-indexed to 1-indexed
	lineInfo := s.textarea.LineInfo()
	col := lineInfo.ColumnOffset + 1 // Convert from 0-indexed to 1-indexed
	return line, col
}
