package main

import core "github.com/shkschneider/macro/core"

// Buffer represents an open file with its state
type Buffer struct {
	filePath   string
	content    string
	readOnly   bool
	cursorLine int
	cursorCol  int
}

// moveCursorToTop moves the syntaxTA cursor to position (0,0)
func (m *model) moveCursorToTop() {
	m.syntaxTA.CursorStart()
	for m.syntaxTA.Line() > 0 {
		m.syntaxTA.CursorUp()
	}
}

// loadBuffer loads a buffer's content into the UI (syntaxTA or viewport)
func (m *model) loadBuffer(idx int) {
	if idx < 0 || idx >= len(m.buffers) {
		return
	}

	buf := m.buffers[idx]

	if buf.readOnly {
		// Apply syntax highlighting for read-only files in viewport
		highlightedContent := core.HighlightCode(buf.content, buf.filePath, "")
		m.viewport.SetContent(highlightedContent)
		m.viewport.GotoTop()
		lang := core.DetectLanguage(buf.filePath)
		if lang != "" {
			m.message = "WARNING: File is read-only. Editing disabled. [" + lang + "]"
		} else {
			m.message = "WARNING: File is read-only. Editing disabled."
		}
	} else {
		// Set filename for syntax highlighting, then set content
		m.syntaxTA.SetFilename(buf.filePath)
		m.syntaxTA.SetValue(buf.content)

		// Restore cursor position: first check buffer state, then fall back to cursor state
		if buf.cursorLine > 0 || buf.cursorCol > 0 {
			// Use buffer's cached cursor position
			m.syntaxTA.SetCursorPosition(buf.cursorLine, buf.cursorCol)
		} else if m.cursorState != nil {
			// Try to restore from persistent storage
			if pos, ok := m.cursorState.GetPosition(buf.filePath); ok {
				m.syntaxTA.SetCursorPosition(pos.Line, pos.Column)
			} else {
				m.moveCursorToTop()
			}
		} else {
			m.moveCursorToTop()
		}

		lang := core.DetectLanguage(buf.filePath)
		if lang != "" {
			m.message = defaultMessage + " [" + lang + "]"
		} else {
			m.message = defaultMessage
		}
	}
	m.currentBuffer = idx
}

// saveCurrentBufferState saves the current UI state to the current buffer
func (m *model) saveCurrentBufferState() {
	if m.currentBuffer < 0 || m.currentBuffer >= len(m.buffers) {
		return
	}

	buf := &m.buffers[m.currentBuffer]
	if !buf.readOnly {
		buf.content = m.syntaxTA.Value()
		buf.cursorLine = m.syntaxTA.Line()
		buf.cursorCol = m.syntaxTA.Column()

		// Persist to cursor state storage
		if m.cursorState != nil {
			m.cursorState.SetPosition(buf.filePath, buf.cursorLine, buf.cursorCol)
		}
	}
}

// addBuffer adds a new buffer or switches to existing one if file already open
func (m *model) addBuffer(filePath string, content string, readOnly bool) int {
	// Check if buffer already exists
	for i, buf := range m.buffers {
		if buf.filePath == filePath {
			return i // Return existing buffer index
		}
	}

	// Create new buffer
	buf := Buffer{
		filePath:   filePath,
		content:    content,
		readOnly:   readOnly,
		cursorLine: 0,
		cursorCol:  0,
	}
	m.buffers = append(m.buffers, buf)
	return len(m.buffers) - 1
}

// getCurrentFilePath returns the file path of the current buffer
func (m *model) getCurrentFilePath() string {
	if m.currentBuffer >= 0 && m.currentBuffer < len(m.buffers) {
		return m.buffers[m.currentBuffer].filePath
	}
	return ""
}

// isCurrentBufferReadOnly returns whether the current buffer is read-only
func (m *model) isCurrentBufferReadOnly() bool {
	if m.currentBuffer >= 0 && m.currentBuffer < len(m.buffers) {
		return m.buffers[m.currentBuffer].readOnly
	}
	return false
}
