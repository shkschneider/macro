// Package core provides core functionality for the macro editor.
package core

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SyntaxTextarea wraps a textarea with syntax highlighting overlay.
// The textarea handles all input, while a highlighted version is displayed.
type SyntaxTextarea struct {
	textarea textarea.Model
	filename string
	language string
	width    int
	height   int

	// Style for line numbers
	lineNumberStyle lipgloss.Style
	// Style for cursor line highlight
	cursorLineStyle lipgloss.Style
}

// NewSyntaxTextarea creates a new syntax-highlighted textarea.
func NewSyntaxTextarea() *SyntaxTextarea {
	ta := textarea.New()
	ta.Focus()
	ta.Prompt = ""
	ta.ShowLineNumbers = false // We'll render our own with syntax highlighting

	// Make textarea text style minimal - we'll overlay highlighted text
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle()
	ta.BlurredStyle.Text = lipgloss.NewStyle()

	return &SyntaxTextarea{
		textarea: ta,
		lineNumberStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(5).
			Align(lipgloss.Right).
			MarginRight(1),
		cursorLineStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("236")),
	}
}

// SetFilename sets the filename for language detection.
func (s *SyntaxTextarea) SetFilename(filename string) {
	s.filename = filename
	s.language = DetectLanguage(filename)
}

// SetLanguage explicitly sets the language for highlighting.
func (s *SyntaxTextarea) SetLanguage(language string) {
	s.language = language
}

// SetValue sets the textarea content.
func (s *SyntaxTextarea) SetValue(value string) {
	s.textarea.SetValue(value)
}

// Value returns the current content.
func (s *SyntaxTextarea) Value() string {
	return s.textarea.Value()
}

// SetWidth sets the width.
func (s *SyntaxTextarea) SetWidth(w int) {
	s.width = w
	// Account for line numbers (5 chars + 1 margin)
	s.textarea.SetWidth(w - 6)
}

// SetHeight sets the height.
func (s *SyntaxTextarea) SetHeight(h int) {
	s.height = h
	s.textarea.SetHeight(h)
}

// Focus focuses the textarea.
func (s *SyntaxTextarea) Focus() tea.Cmd {
	return s.textarea.Focus()
}

// Blur blurs the textarea.
func (s *SyntaxTextarea) Blur() {
	s.textarea.Blur()
}

// Focused returns whether the textarea is focused.
func (s *SyntaxTextarea) Focused() bool {
	return s.textarea.Focused()
}

// CursorStart moves cursor to start of current line.
func (s *SyntaxTextarea) CursorStart() {
	s.textarea.CursorStart()
}

// CursorUp moves cursor up one line.
func (s *SyntaxTextarea) CursorUp() {
	s.textarea.CursorUp()
}

// Line returns the current line number (0-indexed).
func (s *SyntaxTextarea) Line() int {
	return s.textarea.Line()
}

// Update handles messages and updates the textarea state.
func (s *SyntaxTextarea) Update(msg tea.Msg) (*SyntaxTextarea, tea.Cmd) {
	var cmd tea.Cmd
	s.textarea, cmd = s.textarea.Update(msg)
	return s, cmd
}

// View renders the syntax-highlighted textarea.
func (s *SyntaxTextarea) View() string {
	content := s.textarea.Value()
	lines := strings.Split(content, "\n")

	// Get highlighted lines
	highlightedContent := HighlightCode(content, s.filename, s.language)
	highlightedLines := strings.Split(highlightedContent, "\n")

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
		// Line number
		lineNum := s.lineNumberStyle.Render(intToStr(i + 1))

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
			// Optional: highlight the entire cursor line background
			// lineContent = s.cursorLineStyle.Render(lineContent)
		}

		result.WriteString(lineNum)
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
		lineNum := s.lineNumberStyle.Render("~")
		result.WriteString(lineNum)
	}

	return result.String()
}

// insertCursor inserts a visible cursor at the specified column position.
func (s *SyntaxTextarea) insertCursor(plainLine, highlightedLine string, col int) string {
	// Get the character at cursor position or space if at end
	plainRunes := []rune(plainLine)
	var cursorChar string
	if col < len(plainRunes) {
		cursorChar = string(plainRunes[col])
	} else {
		cursorChar = " "
	}

	// Style the cursor character with reverse video
	cursorStyle := lipgloss.NewStyle().
		Reverse(true)
	styledCursor := cursorStyle.Render(cursorChar)

	// If cursor is at or past end of line, append cursor
	if col >= len(plainRunes) {
		return highlightedLine + styledCursor
	}

	// We need to find the position in the highlighted string
	// This is tricky because of ANSI codes. We'll rebuild the line.
	highlightedRunes := []rune(highlightedLine)
	visibleCol := 0
	insertPos := 0
	inEscape := false

	for i := 0; i < len(highlightedRunes); i++ {
		r := highlightedRunes[i]

		// Track ANSI escape sequences
		if r == '\x1b' {
			inEscape = true
			if visibleCol == col {
				insertPos = i
				break
			}
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}

		if visibleCol == col {
			insertPos = i
			break
		}
		visibleCol++
	}

	// Build result: before cursor + styled cursor + after cursor (skip original char)
	before := string(highlightedRunes[:insertPos])

	// Find end of the character we're replacing (skip ANSI codes after it)
	skipTo := insertPos + 1
	for skipTo < len(highlightedRunes) {
		r := highlightedRunes[skipTo]
		if r == '\x1b' {
			// Keep ANSI codes that come after
			break
		}
		break
	}

	after := ""
	if skipTo < len(highlightedRunes) {
		after = string(highlightedRunes[skipTo:])
	}

	// Reset any styling before cursor and continue after
	return before + "\x1b[0m" + styledCursor + after
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
func (s *SyntaxTextarea) GetLanguage() string {
	return s.language
}
