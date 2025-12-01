package internal

import (
	"path/filepath"

)

// Buffer represents an open file with its state
type Buffer struct {
	FilePath        string
	Content         string
	OriginalContent string // Original content for detecting modifications
	FileSize        int64  // File size in bytes for display
	ReadOnly        bool
	CursorLine      int
	CursorCol       int
}

// IsModified returns true if the buffer content has been modified from the original
func (b *Buffer) IsModified() bool {
	return b.Content != b.OriginalContent
}

// moveCursorToTop moves the syntaxTA cursor to position (0,0)
func (m *Model) moveCursorToTop() {
	m.SyntaxTA.CursorStart()
	for m.SyntaxTA.Line() > 0 {
		m.SyntaxTA.CursorUp()
	}
}

// loadBuffer loads a buffer's content into the UI (syntaxTA or viewport)
func (m *Model) loadBuffer(idx int) {
	if idx < 0 || idx >= len(m.Buffers) {
		return
	}

	buf := m.Buffers[idx]

	if buf.ReadOnly {
		// Apply syntax highlighting for read-only files in viewport
		highlightedContent := HighlightCode(buf.Content, buf.FilePath, "")
		m.Viewport.SetContent(highlightedContent)
		m.Viewport.GotoTop()
		lang := DetectLanguage(buf.FilePath)
		if lang != "" {
			m.Message = "WARNING: File is read-only. Editing disabled. [" + lang + "]"
		} else {
			m.Message = "WARNING: File is read-only. Editing disabled."
		}
	} else {
		// Set filename for syntax highlighting and diff tracking
		// SetFilename also sets up git diff tracking for git-tracked files
		m.SyntaxTA.SetFilename(buf.FilePath)
		m.SyntaxTA.SetValue(buf.Content)

		// Restore cursor position: first check buffer state, then fall back to cursor state
		if buf.CursorLine > 0 || buf.CursorCol > 0 {
			// Use buffer's cached cursor position
			m.SyntaxTA.SetCursorPosition(buf.CursorLine, buf.CursorCol)
		} else if m.CursorState != nil {
			// Try to restore from persistent storage
			if pos, ok := m.CursorState.GetPosition(buf.FilePath); ok {
				m.SyntaxTA.SetCursorPosition(pos.Line, pos.Column)
			} else {
				m.moveCursorToTop()
			}
		} else {
			m.moveCursorToTop()
		}

		lang := DetectLanguage(buf.FilePath)
		if lang != "" {
			m.Message = defaultMessage + " [" + lang + "]"
		} else {
			m.Message = defaultMessage
		}
	}
	m.CurrentBuffer = idx
}

// saveCurrentBufferState saves the current UI state to the current buffer
func (m *Model) saveCurrentBufferState() {
	if m.CurrentBuffer < 0 || m.CurrentBuffer >= len(m.Buffers) {
		return
	}

	buf := &m.Buffers[m.CurrentBuffer]
	if !buf.ReadOnly {
		buf.Content = m.SyntaxTA.Value()
		buf.CursorLine = m.SyntaxTA.Line()
		buf.CursorCol = m.SyntaxTA.Column()

		// Persist to cursor state storage
		if m.CursorState != nil {
			m.CursorState.SetPosition(buf.FilePath, buf.CursorLine, buf.CursorCol)
		}
	}
}

// addBuffer adds a new buffer or switches to existing one if file already open
func (m *Model) addBuffer(filePath string, content string, readOnly bool, fileSize int64) int {
	// Check if buffer already exists
	for i, buf := range m.Buffers {
		if buf.FilePath == filePath {
			return i // Return existing buffer index
		}
	}

	// Create new buffer
	buf := Buffer{
		FilePath:        filePath,
		Content:         content,
		OriginalContent: content, // Store original for modification tracking
		ReadOnly:        readOnly,
		FileSize:        fileSize,
		CursorLine:      0,
		CursorCol:       0,
	}
	m.Buffers = append(m.Buffers, buf)
	return len(m.Buffers) - 1
}

// getCurrentFilePath returns the file path of the current buffer
func (m *Model) getCurrentFilePath() string {
	if m.CurrentBuffer >= 0 && m.CurrentBuffer < len(m.Buffers) {
		return m.Buffers[m.CurrentBuffer].FilePath
	}
	return ""
}

// isCurrentBufferReadOnly returns whether the current buffer is read-only
func (m *Model) isCurrentBufferReadOnly() bool {
	if m.CurrentBuffer >= 0 && m.CurrentBuffer < len(m.Buffers) {
		return m.Buffers[m.CurrentBuffer].ReadOnly
	}
	return false
}

// isCurrentBufferModified returns whether the current buffer has been modified
func (m *Model) isCurrentBufferModified() bool {
	if m.CurrentBuffer >= 0 && m.CurrentBuffer < len(m.Buffers) {
		buf := &m.Buffers[m.CurrentBuffer]
		// For editable buffers, check current textarea content against original
		if !buf.ReadOnly {
			return m.SyntaxTA.Value() != buf.OriginalContent
		}
		return buf.IsModified()
	}
	return false
}

// getCurrentBuffer returns the current buffer, or nil if none is selected
func (m *Model) getCurrentBuffer() *Buffer {
	if m.CurrentBuffer >= 0 && m.CurrentBuffer < len(m.Buffers) {
		return &m.Buffers[m.CurrentBuffer]
	}
	return nil
}

// getDirectoryPath returns the directory portion of the current file path
func (m *Model) getDirectoryPath() string {
	if m.CurrentBuffer >= 0 && m.CurrentBuffer < len(m.Buffers) {
		return filepath.Dir(m.Buffers[m.CurrentBuffer].FilePath)
	}
	return ""
}

// ===== EditorContext interface implementation =====
// These methods implement api.EditorContext to allow features to interact with the editor

// IsCurrentBufferReadOnly implements api.EditorContext
func (m *Model) IsCurrentBufferReadOnly() bool {
	return m.isCurrentBufferReadOnly()
}

// GetCurrentFilePath implements api.EditorContext
func (m *Model) GetCurrentFilePath() string {
	return m.getCurrentFilePath()
}

// GetCurrentContent implements api.EditorContext
func (m *Model) GetCurrentContent() string {
	return m.SyntaxTA.Value()
}

// SaveCurrentBufferState implements api.EditorContext
func (m *Model) SaveCurrentBufferState() {
	m.saveCurrentBufferState()
}

// UpdateBufferAfterSave implements api.EditorContext
func (m *Model) UpdateBufferAfterSave(content string, fileSize int64) {
	if buf := m.getCurrentBuffer(); buf != nil {
		buf.OriginalContent = content
		buf.FileSize = fileSize
	}
}

// SetMessage implements api.EditorContext
func (m *Model) SetMessage(msg string) {
	m.Message = msg
}

// SetError implements api.EditorContext
func (m *Model) SetError(err error) {
	m.Err = err
}
