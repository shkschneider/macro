// Package core provides core functionality for the macro editor.
package internal

import (
	"sort"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SecondaryCursor represents an additional cursor position for multi-cursor editing.
// Line and Column are 0-indexed.
type SecondaryCursor struct {
	Line   int
	Column int
}

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

	// Secondary cursors for multi-cursor editing
	// The primary cursor is managed by textarea.Model
	secondaryCursors []SecondaryCursor

	// Style for secondary cursors (slightly different from primary)
	secondaryCursorStyle lipgloss.Style
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
		diffTracker:          NewDiffTracker(),
		secondaryCursors:     []SecondaryCursor{},
		secondaryCursorStyle: lipgloss.NewStyle().Background(lipgloss.Color("240")),
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

// ===== Multi-Cursor Methods =====

// AddSecondaryCursor adds a secondary cursor at the specified position (0-indexed).
// If a cursor already exists at this position, it is not added again.
func (s *Textarea) AddSecondaryCursor(line, column int) {
	// Don't add duplicate cursors
	for _, c := range s.secondaryCursors {
		if c.Line == line && c.Column == column {
			return
		}
	}
	// Don't add cursor at primary cursor position
	if line == s.textarea.Line() && column == s.textarea.LineInfo().ColumnOffset {
		return
	}
	s.secondaryCursors = append(s.secondaryCursors, SecondaryCursor{Line: line, Column: column})
	s.sortCursors()
}

// ClearSecondaryCursors removes all secondary cursors.
func (s *Textarea) ClearSecondaryCursors() {
	s.secondaryCursors = []SecondaryCursor{}
}

// HasSecondaryCursors returns true if there are any secondary cursors.
func (s *Textarea) HasSecondaryCursors() bool {
	return len(s.secondaryCursors) > 0
}

// SecondaryCursorCount returns the number of secondary cursors.
func (s *Textarea) SecondaryCursorCount() int {
	return len(s.secondaryCursors)
}

// GetAllCursorPositions returns all cursor positions (primary first, then secondary).
// All positions are 0-indexed.
func (s *Textarea) GetAllCursorPositions() []SecondaryCursor {
	result := []SecondaryCursor{{
		Line:   s.textarea.Line(),
		Column: s.textarea.LineInfo().ColumnOffset,
	}}
	result = append(result, s.secondaryCursors...)
	return result
}

// sortCursors sorts secondary cursors by line then column.
func (s *Textarea) sortCursors() {
	sort.Slice(s.secondaryCursors, func(i, j int) bool {
		if s.secondaryCursors[i].Line != s.secondaryCursors[j].Line {
			return s.secondaryCursors[i].Line < s.secondaryCursors[j].Line
		}
		return s.secondaryCursors[i].Column < s.secondaryCursors[j].Column
	})
}

// insertTextAtAllCursors inserts text at all cursor positions (primary and secondary).
// Cursors are processed from bottom-right to top-left to avoid position shifts affecting later insertions.
func (s *Textarea) insertTextAtAllCursors(text string) {
	// Get all cursors and sort them from bottom-right to top-left
	allCursors := s.GetAllCursorPositions()
	sort.Slice(allCursors, func(i, j int) bool {
		if allCursors[i].Line != allCursors[j].Line {
			return allCursors[i].Line > allCursors[j].Line
		}
		return allCursors[i].Column > allCursors[j].Column
	})

	content := s.textarea.Value()
	lines := strings.Split(content, "\n")

	// Process insertions from bottom-right to top-left
	for _, cursor := range allCursors {
		lines = s.insertTextAtPosition(lines, cursor.Line, cursor.Column, text)
	}

	// Update the content
	newContent := strings.Join(lines, "\n")
	s.textarea.SetValue(newContent)

	// Update cursor positions - all cursors move right by len(text) or handle newlines
	textRunes := []rune(text)
	textLen := len(textRunes)
	hasNewline := strings.Contains(text, "\n")

	// Update primary cursor position
	primaryLine := s.textarea.Line()
	primaryCol := s.textarea.LineInfo().ColumnOffset

	// For single-line insert, primary cursor just moves right
	// For multi-line insert, it moves to the new line position
	if hasNewline {
		newlineCount := strings.Count(text, "\n")
		lastNewlineIdx := strings.LastIndex(text, "\n")
		afterLastNewline := text[lastNewlineIdx+1:]
		s.SetCursorPosition(primaryLine+newlineCount, len([]rune(afterLastNewline)))
	} else {
		s.SetCursorPosition(primaryLine, primaryCol+textLen)
	}

	// Update secondary cursor positions
	for i := range s.secondaryCursors {
		if hasNewline {
			newlineCount := strings.Count(text, "\n")
			lastNewlineIdx := strings.LastIndex(text, "\n")
			afterLastNewline := text[lastNewlineIdx+1:]
			s.secondaryCursors[i].Line += newlineCount
			s.secondaryCursors[i].Column = len([]rune(afterLastNewline))
		} else {
			s.secondaryCursors[i].Column += textLen
		}
	}
}

// insertTextAtPosition inserts text at a specific line and column position.
func (s *Textarea) insertTextAtPosition(lines []string, line, col int, text string) []string {
	if line < 0 || line >= len(lines) {
		return lines
	}

	lineContent := lines[line]
	runes := []rune(lineContent)

	// Clamp column to valid range
	if col < 0 {
		col = 0
	}
	if col > len(runes) {
		col = len(runes)
	}

	before := string(runes[:col])
	after := string(runes[col:])

	// Handle text with newlines
	if strings.Contains(text, "\n") {
		parts := strings.Split(text, "\n")
		newLines := make([]string, 0, len(lines)+len(parts)-1)

		// Add lines before the insertion point
		newLines = append(newLines, lines[:line]...)

		// Add the first part combined with content before cursor
		newLines = append(newLines, before+parts[0])

		// Add middle parts (if any)
		for i := 1; i < len(parts)-1; i++ {
			newLines = append(newLines, parts[i])
		}

		// Add the last part combined with content after cursor
		newLines = append(newLines, parts[len(parts)-1]+after)

		// Add lines after the original line
		newLines = append(newLines, lines[line+1:]...)

		return newLines
	}

	// Simple insertion without newlines
	lines[line] = before + text + after
	return lines
}

// deleteAtAllCursors deletes characters at all cursor positions.
// If forward is true, deletes character after cursor (Delete key).
// If forward is false, deletes character before cursor (Backspace key).
func (s *Textarea) deleteAtAllCursors(forward bool) {
	// Get all cursors and sort them from bottom-right to top-left
	allCursors := s.GetAllCursorPositions()
	sort.Slice(allCursors, func(i, j int) bool {
		if allCursors[i].Line != allCursors[j].Line {
			return allCursors[i].Line > allCursors[j].Line
		}
		return allCursors[i].Column > allCursors[j].Column
	})

	content := s.textarea.Value()
	lines := strings.Split(content, "\n")

	// Count line merges to adjust cursor line positions
	lineMergeCount := 0

	// Process deletions from bottom-right to top-left
	for _, cursor := range allCursors {
		var lineMerged bool
		lines, lineMerged = s.deleteAtPosition(lines, cursor.Line-lineMergeCount, cursor.Column, forward)
		if lineMerged {
			lineMergeCount++
		}
	}

	// Update the content
	newContent := strings.Join(lines, "\n")
	s.textarea.SetValue(newContent)

	// Update cursor positions
	primaryLine := s.textarea.Line()
	primaryCol := s.textarea.LineInfo().ColumnOffset

	// For backspace, primary cursor moves left by 1 (or up if at start of line)
	if !forward {
		if primaryCol > 0 {
			s.SetCursorPosition(primaryLine, primaryCol-1)
		} else if primaryLine > 0 {
			// Move to end of previous line
			prevLineLen := len([]rune(lines[primaryLine-1]))
			s.SetCursorPosition(primaryLine-1, prevLineLen)
		}
	}
	// For delete, cursor doesn't move

	// Update secondary cursor positions
	for i := range s.secondaryCursors {
		if !forward {
			if s.secondaryCursors[i].Column > 0 {
				s.secondaryCursors[i].Column--
			} else if s.secondaryCursors[i].Line > 0 {
				// Line merge happened, cursor moves to end of previous line
				s.secondaryCursors[i].Line--
				if s.secondaryCursors[i].Line < len(lines) {
					s.secondaryCursors[i].Column = len([]rune(lines[s.secondaryCursors[i].Line]))
				}
			}
		}
	}

	// Remove duplicate cursors that might have ended up at the same position
	s.removeDuplicateCursors()
}

// deleteAtPosition deletes a character at a specific position.
// Returns the modified lines and whether a line merge occurred.
func (s *Textarea) deleteAtPosition(lines []string, line, col int, forward bool) ([]string, bool) {
	if line < 0 || line >= len(lines) {
		return lines, false
	}

	lineContent := lines[line]
	runes := []rune(lineContent)

	if forward {
		// Delete key - delete character at cursor position
		if col < len(runes) {
			// Delete character in current line
			lines[line] = string(runes[:col]) + string(runes[col+1:])
			return lines, false
		} else if line < len(lines)-1 {
			// At end of line, merge with next line
			lines[line] = lineContent + lines[line+1]
			lines = append(lines[:line+1], lines[line+2:]...)
			return lines, true
		}
	} else {
		// Backspace - delete character before cursor position
		if col > 0 {
			// Delete character before cursor
			lines[line] = string(runes[:col-1]) + string(runes[col:])
			return lines, false
		} else if line > 0 {
			// At start of line, merge with previous line
			prevLine := lines[line-1]
			lines[line-1] = prevLine + lineContent
			lines = append(lines[:line], lines[line+1:]...)
			return lines, true
		}
	}

	return lines, false
}

// removeDuplicateCursors removes secondary cursors that are at the same position as primary or other secondary cursors.
func (s *Textarea) removeDuplicateCursors() {
	primaryLine := s.textarea.Line()
	primaryCol := s.textarea.LineInfo().ColumnOffset

	seen := make(map[string]bool)
	// Mark primary cursor position as seen
	seen[cursorKey(primaryLine, primaryCol)] = true

	var filtered []SecondaryCursor
	for _, c := range s.secondaryCursors {
		key := cursorKey(c.Line, c.Column)
		if !seen[key] {
			seen[key] = true
			filtered = append(filtered, c)
		}
	}
	s.secondaryCursors = filtered
}

// cursorKey creates a unique key for a cursor position.
func cursorKey(line, col int) string {
	return strconv.Itoa(line) + ":" + strconv.Itoa(col)
}

// AddCursorAtNextOccurrence finds the next occurrence of the word under/near the cursor
// and adds a secondary cursor there. Returns true if a new cursor was added.
func (s *Textarea) AddCursorAtNextOccurrence() bool {
	content := s.textarea.Value()
	lines := strings.Split(content, "\n")
	
	// Get current cursor position
	curLine := s.textarea.Line()
	curCol := s.textarea.LineInfo().ColumnOffset
	
	// Get the word at cursor position
	word := s.getWordAtPosition(lines, curLine, curCol)
	if word == "" {
		return false
	}
	
	// Find all positions of all cursors (primary + secondary)
	allCursors := s.GetAllCursorPositions()
	
	// Find all occurrences of the word
	occurrences := s.findWordOccurrences(lines, word)
	if len(occurrences) == 0 {
		return false
	}
	
	// Find the next occurrence after the last cursor
	lastCursor := allCursors[len(allCursors)-1]
	for _, occ := range occurrences {
		// Skip if this occurrence is before or at the last cursor
		if occ.Line < lastCursor.Line || (occ.Line == lastCursor.Line && occ.Column <= lastCursor.Column) {
			continue
		}
		// Skip if a cursor already exists at this position
		cursorExists := false
		for _, c := range allCursors {
			if c.Line == occ.Line && c.Column == occ.Column {
				cursorExists = true
				break
			}
		}
		if !cursorExists {
			s.AddSecondaryCursor(occ.Line, occ.Column)
			return true
		}
	}
	
	// Wrap around to the beginning if no occurrence found after last cursor
	for _, occ := range occurrences {
		// Skip if a cursor already exists at this position
		cursorExists := false
		for _, c := range allCursors {
			if c.Line == occ.Line && c.Column == occ.Column {
				cursorExists = true
				break
			}
		}
		if !cursorExists {
			s.AddSecondaryCursor(occ.Line, occ.Column)
			return true
		}
	}
	
	return false
}

// getWordAtPosition extracts the word at the given position.
func (s *Textarea) getWordAtPosition(lines []string, line, col int) string {
	if line < 0 || line >= len(lines) {
		return ""
	}
	lineContent := lines[line]
	runes := []rune(lineContent)
	
	if col < 0 || col > len(runes) {
		return ""
	}
	
	// Find word boundaries
	start := col
	end := col
	
	// Move start backward to find word beginning
	for start > 0 && isWordChar(runes[start-1]) {
		start--
	}
	
	// Move end forward to find word end
	for end < len(runes) && isWordChar(runes[end]) {
		end++
	}
	
	if start == end {
		return ""
	}
	
	return string(runes[start:end])
}

// findWordOccurrences finds all occurrences of a word in the content.
// Returns positions as line/column (0-indexed).
func (s *Textarea) findWordOccurrences(lines []string, word string) []SecondaryCursor {
	var occurrences []SecondaryCursor
	
	for lineIdx, lineContent := range lines {
		runes := []rune(lineContent)
		wordRunes := []rune(word)
		
		for col := 0; col <= len(runes)-len(wordRunes); col++ {
			// Check if word matches at this position
			match := true
			for i, wr := range wordRunes {
				if runes[col+i] != wr {
					match = false
					break
				}
			}
			if !match {
				continue
			}
			
			// Check word boundaries - make sure it's a complete word
			if col > 0 && isWordChar(runes[col-1]) {
				continue
			}
			if col+len(wordRunes) < len(runes) && isWordChar(runes[col+len(wordRunes)]) {
				continue
			}
			
			occurrences = append(occurrences, SecondaryCursor{Line: lineIdx, Column: col})
		}
	}
	
	return occurrences
}

// isWordChar returns true if the rune is a word character (alphanumeric or underscore).
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// Update handles messages and updates the textarea state.
// When there are secondary cursors, text operations are applied to all cursor positions.
func (s *Textarea) Update(msg tea.Msg) (*Textarea, tea.Cmd) {
	// If no secondary cursors, just pass through to textarea
	if len(s.secondaryCursors) == 0 {
		var cmd tea.Cmd
		s.textarea, cmd = s.textarea.Update(msg)
		return s, cmd
	}

	// Handle key messages specially for multi-cursor editing
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		// Handle special keys that need multi-cursor processing
		switch keyMsg.Type {
		case tea.KeyRunes:
			// Insert characters at all cursor positions
			s.insertTextAtAllCursors(string(keyMsg.Runes))
			return s, nil

		case tea.KeySpace:
			s.insertTextAtAllCursors(" ")
			return s, nil

		case tea.KeyTab:
			s.insertTextAtAllCursors("\t")
			return s, nil

		case tea.KeyEnter:
			s.insertTextAtAllCursors("\n")
			return s, nil

		case tea.KeyBackspace:
			s.deleteAtAllCursors(false)
			return s, nil

		case tea.KeyDelete:
			s.deleteAtAllCursors(true)
			return s, nil
		}
	}

	// For other messages (like window resize), pass through normally
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

		// Collect all cursor columns on this line
		var cursorCols []int
		var isPrimary []bool // Track which cursors are primary
		if i == cursorLine && s.textarea.Focused() {
			cursorCols = append(cursorCols, cursorCol)
			isPrimary = append(isPrimary, true)
		}
		// Add secondary cursors on this line
		for _, sc := range s.secondaryCursors {
			if sc.Line == i && s.textarea.Focused() {
				cursorCols = append(cursorCols, sc.Column)
				isPrimary = append(isPrimary, false)
			}
		}

		// Insert all cursors into the line content
		if len(cursorCols) > 0 {
			lineContent = s.insertMultipleCursors(lines[i], highlightedLines[i], cursorCols, isPrimary)
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

// insertMultipleCursors inserts multiple visible cursors at the specified column positions.
// Primary cursor uses reverse video, secondary cursors use underline.
func (s *Textarea) insertMultipleCursors(plainLine, highlightedLine string, cols []int, isPrimary []bool) string {
	if len(cols) == 0 {
		return highlightedLine
	}

	plainRunes := []rune(plainLine)
	highlightedRunes := []rune(highlightedLine)

	// Build a map of visible column to cursor info
	type cursorInfo struct {
		isPrimary bool
	}
	cursorMap := make(map[int]cursorInfo)
	for i, col := range cols {
		cursorMap[col] = cursorInfo{isPrimary: isPrimary[i]}
	}

	// Map visible column positions to their rune positions in highlighted line
	type runePos struct {
		start     int
		end       int
		isPrimary bool
	}
	var cursorPositions []runePos

	visibleCol := 0
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

		// This is a visible character - check if cursor is here
		if info, exists := cursorMap[visibleCol]; exists {
			cursorPositions = append(cursorPositions, runePos{
				start:     i,
				end:       i + 1,
				isPrimary: info.isPrimary,
			})
		}
		visibleCol++
	}

	// Handle cursors at end of line (beyond visible content)
	for col, info := range cursorMap {
		if col >= len(plainRunes) {
			// This cursor is at end of line - mark it with a special position
			cursorPositions = append(cursorPositions, runePos{
				start:     -1, // Special marker for end of line
				end:       -1,
				isPrimary: info.isPrimary,
			})
		}
	}

	// If no cursor positions found in line content, just handle end-of-line cursors
	if len(cursorPositions) == 0 {
		// Add cursors at end of line
		var endCursors string
		for _, info := range cursorMap {
			if info.isPrimary {
				endCursors += "\x1b[7m \x1b[27m"
			} else {
				endCursors += "\x1b[4m \x1b[24m"
			}
		}
		return highlightedLine + endCursors
	}

	// Sort cursor positions by start position (descending) so we can insert from right to left
	sort.Slice(cursorPositions, func(i, j int) bool {
		return cursorPositions[i].start > cursorPositions[j].start
	})

	// Build result by inserting cursor codes from right to left
	result := highlightedRunes
	var endOfLineCursors []runePos
	for _, pos := range cursorPositions {
		if pos.start == -1 {
			endOfLineCursors = append(endOfLineCursors, pos)
			continue
		}

		// Insert cursor styling
		var cursorStart, cursorEnd string
		if pos.isPrimary {
			cursorStart = "\x1b[7m"  // Reverse video ON
			cursorEnd = "\x1b[27m"   // Reverse video OFF
		} else {
			cursorStart = "\x1b[4m"  // Underline ON
			cursorEnd = "\x1b[24m"   // Underline OFF
		}

		// Insert cursor codes around the character
		before := string(result[:pos.start])
		char := string(result[pos.start:pos.end])
		after := string(result[pos.end:])
		result = []rune(before + cursorStart + char + cursorEnd + after)
	}

	// Add end-of-line cursors
	resultStr := string(result)
	for _, pos := range endOfLineCursors {
		if pos.isPrimary {
			resultStr += "\x1b[7m \x1b[27m"
		} else {
			resultStr += "\x1b[4m \x1b[24m"
		}
	}

	return resultStr
}

// intToStr converts an integer to string using the standard library.
func intToStr(n int) string {
	return strconv.Itoa(n)
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
