package main

// Buffer represents an open file with its state
type Buffer struct {
	filePath string
	content  string
	readOnly bool
}

// moveCursorToTop moves the textarea cursor to position (0,0)
func (m *model) moveCursorToTop() {
	m.textarea.CursorStart()
	for m.textarea.Line() > 0 {
		m.textarea.CursorUp()
	}
}

// loadBuffer loads a buffer's content into the UI (textarea or viewport)
func (m *model) loadBuffer(idx int) {
	if idx < 0 || idx >= len(m.buffers) {
		return
	}

	buf := m.buffers[idx]

	if buf.readOnly {
		m.viewport.SetContent(buf.content)
		m.viewport.GotoTop()
		m.message = "WARNING: File is read-only. Editing disabled."
	} else {
		m.textarea.SetValue(buf.content)
		m.moveCursorToTop()
		m.message = defaultMessage
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
		buf.content = m.textarea.Value()
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
		filePath: filePath,
		content:  content,
		readOnly: readOnly,
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
