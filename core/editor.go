// Package core provides core functionality for the macro editor.
package core

import (
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HighlightedEditor wraps a textarea and provides syntax highlighting
// while maintaining full editing capabilities.
type HighlightedEditor struct {
	textarea       textarea.Model
	filename       string
	language       string
	highlightCache string
	width          int
	height         int
	lineNumberStyle lipgloss.Style
	cursorStyle    lipgloss.Style
}

// NewHighlightedEditor creates a new syntax-highlighted editor.
func NewHighlightedEditor() *HighlightedEditor {
	ta := textarea.New()
	ta.Focus()
	ta.Prompt = ""
	ta.ShowLineNumbers = false // We'll handle line numbers ourselves

	return &HighlightedEditor{
		textarea: ta,
		lineNumberStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(4).
			Align(lipgloss.Right).
			PaddingRight(1),
		cursorStyle: lipgloss.NewStyle().
			Background(lipgloss.Color("255")).
			Foreground(lipgloss.Color("0")),
	}
}

// SetFilename sets the filename for language detection.
func (e *HighlightedEditor) SetFilename(filename string) {
	e.filename = filename
	e.language = DetectLanguage(filename)
	e.invalidateCache()
}

// SetLanguage explicitly sets the language for highlighting.
func (e *HighlightedEditor) SetLanguage(language string) {
	e.language = language
	e.invalidateCache()
}

// SetValue sets the editor content.
func (e *HighlightedEditor) SetValue(s string) {
	e.textarea.SetValue(s)
	e.invalidateCache()
}

// Value returns the current content.
func (e *HighlightedEditor) Value() string {
	return e.textarea.Value()
}

// SetWidth sets the editor width.
func (e *HighlightedEditor) SetWidth(w int) {
	e.width = w
	// Account for line numbers (4 chars + 1 space)
	e.textarea.SetWidth(w - 5)
}

// SetHeight sets the editor height.
func (e *HighlightedEditor) SetHeight(h int) {
	e.height = h
	e.textarea.SetHeight(h)
}

// Focus focuses the editor.
func (e *HighlightedEditor) Focus() tea.Cmd {
	return e.textarea.Focus()
}

// Blur blurs the editor.
func (e *HighlightedEditor) Blur() {
	e.textarea.Blur()
}

// Focused returns whether the editor is focused.
func (e *HighlightedEditor) Focused() bool {
	return e.textarea.Focused()
}

// CursorStart moves cursor to start of line.
func (e *HighlightedEditor) CursorStart() {
	e.textarea.CursorStart()
}

// CursorUp moves cursor up.
func (e *HighlightedEditor) CursorUp() {
	e.textarea.CursorUp()
}

// Line returns the current line number.
func (e *HighlightedEditor) Line() int {
	return e.textarea.Line()
}

// invalidateCache clears the highlight cache.
func (e *HighlightedEditor) invalidateCache() {
	e.highlightCache = ""
}

// Update handles messages and updates the editor state.
func (e *HighlightedEditor) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	
	// Check if content changed
	oldValue := e.textarea.Value()
	e.textarea, cmd = e.textarea.Update(msg)
	if e.textarea.Value() != oldValue {
		e.invalidateCache()
	}
	
	return cmd
}

// View renders the syntax-highlighted editor.
func (e *HighlightedEditor) View() string {
	content := e.textarea.Value()
	
	// Get highlighted content
	highlighted := HighlightCode(content, e.filename, e.language)
	
	// Split into lines
	highlightedLines := strings.Split(highlighted, "\n")
	contentLines := strings.Split(content, "\n")
	
	// Get cursor position
	cursorLine := e.textarea.Line()
	lineInfo := e.textarea.LineInfo()
	cursorCol := lineInfo.ColumnOffset
	
	// Calculate visible range based on height and cursor position
	visibleStart := 0
	visibleEnd := len(highlightedLines)
	
	if e.height > 0 && len(highlightedLines) > e.height {
		// Center the view around the cursor
		halfHeight := e.height / 2
		visibleStart = cursorLine - halfHeight
		if visibleStart < 0 {
			visibleStart = 0
		}
		visibleEnd = visibleStart + e.height
		if visibleEnd > len(highlightedLines) {
			visibleEnd = len(highlightedLines)
			visibleStart = visibleEnd - e.height
			if visibleStart < 0 {
				visibleStart = 0
			}
		}
	}
	
	// Build the view with line numbers and syntax highlighting
	var result strings.Builder
	for i := visibleStart; i < visibleEnd; i++ {
		// Line number - use intToString for all numbers
		lineNumStr := intToString(i + 1)
		// Pad to 4 characters for alignment
		padding := ""
		if len(lineNumStr) < 4 {
			padding = strings.Repeat(" ", 4-len(lineNumStr))
		}
		lineNum := e.lineNumberStyle.Render(padding + lineNumStr)
		
		// Get the highlighted line
		var line string
		if i < len(highlightedLines) {
			line = highlightedLines[i]
		}
		
		// If this is the cursor line and we're focused, show cursor
		if i == cursorLine && e.textarea.Focused() {
			// We need to insert cursor into the content line
			if i < len(contentLines) {
				line = e.insertCursor(contentLines[i], highlightedLines[i], cursorCol)
			}
		}
		
		result.WriteString(lineNum)
		result.WriteString(line)
		if i < visibleEnd-1 {
			result.WriteString("\n")
		}
	}
	
	return result.String()
}

// insertCursor inserts a visible cursor at the specified column position.
// This function carefully handles ANSI escape codes in the highlighted line.
func (e *HighlightedEditor) insertCursor(plainLine, highlightedLine string, col int) string {
	// Get the character at cursor position or space if at end
	runes := []rune(plainLine)
	var cursorChar rune
	if col < len(runes) {
		cursorChar = runes[col]
	} else {
		cursorChar = ' '
	}
	
	// Build the cursor-styled character
	cursorStr := e.cursorStyle.Render(string(cursorChar))
	
	// If cursor is at end of line, simply append
	if col >= len(runes) {
		return highlightedLine + cursorStr
	}
	
	// Find the position in the highlighted string that corresponds to col
	// We need to count visible characters while skipping ANSI escape codes
	highlightedRunes := []rune(highlightedLine)
	visibleCol := 0
	insertPos := 0
	skipLen := 0 // Length of the character we're replacing
	
	for i := 0; i < len(highlightedRunes); {
		// Check if we're at an ANSI escape sequence
		if highlightedRunes[i] == '\x1b' && i+1 < len(highlightedRunes) && highlightedRunes[i+1] == '[' {
			// Skip the escape sequence
			j := i + 2
			for j < len(highlightedRunes) && highlightedRunes[j] != 'm' {
				j++
			}
			if j < len(highlightedRunes) {
				i = j + 1 // Skip past 'm'
				continue
			}
		}
		
		if visibleCol == col {
			insertPos = i
			skipLen = 1
			break
		}
		
		visibleCol++
		i++
	}
	
	// If we didn't find the position, append at end
	if visibleCol < col {
		return highlightedLine + cursorStr
	}
	
	// Build the result: before cursor + styled cursor char + after cursor
	before := string(highlightedRunes[:insertPos])
	after := ""
	if insertPos+skipLen < len(highlightedRunes) {
		after = string(highlightedRunes[insertPos+skipLen:])
	}
	
	// Reset style before cursor char and restore after
	return before + "\x1b[0m" + cursorStr + after
}

// intToString converts an integer to string (simple helper).
func intToString(n int) string {
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

// GetTextarea returns the underlying textarea for direct access if needed.
func (e *HighlightedEditor) GetTextarea() *textarea.Model {
	return &e.textarea
}

// KeyMap returns the key bindings for the editor.
func (e *HighlightedEditor) KeyMap() textarea.KeyMap {
	return e.textarea.KeyMap
}

// SetKeyMap sets the key bindings for the editor.
func (e *HighlightedEditor) SetKeyMap(km textarea.KeyMap) {
	e.textarea.KeyMap = km
}

// HighlightedEditorKeyMap provides key bindings that work with the highlighted editor.
type HighlightedEditorKeyMap struct {
	CharacterForward  key.Binding
	CharacterBackward key.Binding
	DeleteCharacterBackward key.Binding
	DeleteCharacterForward  key.Binding
	LineNext     key.Binding
	LinePrevious key.Binding
	LineStart    key.Binding
	LineEnd      key.Binding
	Paste        key.Binding
	DeleteAfterCursor  key.Binding
	DeleteBeforeCursor key.Binding
	DeleteWordBackward key.Binding
	DeleteWordForward  key.Binding
}

// DefaultHighlightedEditorKeyMap returns default key bindings.
var DefaultHighlightedEditorKeyMap = HighlightedEditorKeyMap{
	CharacterForward: key.NewBinding(key.WithKeys("right", "ctrl+f")),
	CharacterBackward: key.NewBinding(key.WithKeys("left", "ctrl+b")),
	DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h")),
	DeleteCharacterForward: key.NewBinding(key.WithKeys("delete", "ctrl+d")),
	LineNext: key.NewBinding(key.WithKeys("down", "ctrl+n")),
	LinePrevious: key.NewBinding(key.WithKeys("up", "ctrl+p")),
	LineStart: key.NewBinding(key.WithKeys("home", "ctrl+a")),
	LineEnd: key.NewBinding(key.WithKeys("end", "ctrl+e")),
	Paste: key.NewBinding(key.WithKeys("ctrl+v")),
	DeleteAfterCursor: key.NewBinding(key.WithKeys("ctrl+k")),
	DeleteBeforeCursor: key.NewBinding(key.WithKeys("ctrl+u")),
	DeleteWordBackward: key.NewBinding(key.WithKeys("ctrl+w")),
	DeleteWordForward: key.NewBinding(key.WithKeys("alt+d")),
}

// Cursor returns the cursor model.
func (e *HighlightedEditor) Cursor() cursor.Model {
	return e.textarea.Cursor
}

// SetCursor sets the cursor column position.
func (e *HighlightedEditor) SetCursor(col int) {
	e.textarea.SetCursor(col)
}
