package main

import (
	"path/filepath"

	core "github.com/shkschneider/macro/core"
)

// Buffer represents an open file with its state
type Buffer struct {
	filePath        string
	content         string
	originalContent string // Original content for detecting modifications
	fileSize        int64 // File size in bytes for display
	readOnly        bool
	cursorLine      int
	cursorCol       int
}

// IsModified returns true if the buffer content has been modified from the original
func (b *Buffer) IsModified() bool {
	return b.content != b.originalContent
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

		// Set original content for diff tracking only for git-tracked files
		// This shows colored indicators for changes compared to the git HEAD version
		if core.IsGitTracked(buf.filePath) {
			if gitContent, ok := core.GetGitFileContent(buf.filePath); ok {
				m.syntaxTA.SetOriginalContent(gitContent)
			} else {
				// File is tracked but no HEAD content (new file staged)
				m.syntaxTA.ClearOriginalContent()
			}
		} else {
			// Untracked file - no diff indicators
			m.syntaxTA.ClearOriginalContent()
		}

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
func (m *model) addBuffer(filePath string, content string, readOnly bool, fileSize int64) int {
	// Check if buffer already exists
	for i, buf := range m.buffers {
		if buf.filePath == filePath {
			return i // Return existing buffer index
		}
	}

	// Create new buffer
	buf := Buffer{
		filePath:        filePath,
		content:         content,
		originalContent: content, // Store original for modification tracking
		readOnly:        readOnly,
		fileSize:        fileSize,
		cursorLine:      0,
		cursorCol:       0,
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

// isCurrentBufferModified returns whether the current buffer has been modified
func (m *model) isCurrentBufferModified() bool {
	if m.currentBuffer >= 0 && m.currentBuffer < len(m.buffers) {
		buf := &m.buffers[m.currentBuffer]
		// For editable buffers, check current textarea content against original
		if !buf.readOnly {
			return m.syntaxTA.Value() != buf.originalContent
		}
		return buf.IsModified()
	}
	return false
}

// getCurrentBuffer returns the current buffer, or nil if none is selected
func (m *model) getCurrentBuffer() *Buffer {
	if m.currentBuffer >= 0 && m.currentBuffer < len(m.buffers) {
		return &m.buffers[m.currentBuffer]
	}
	return nil
}

// getDirectoryPath returns the directory portion of the current file path
func (m *model) getDirectoryPath() string {
	if m.currentBuffer >= 0 && m.currentBuffer < len(m.buffers) {
		return filepath.Dir(m.buffers[m.currentBuffer].filePath)
	}
	return ""
}

// ===== EditorContext interface implementation =====
// These methods implement core.EditorContext to allow features to interact with the editor

// IsCurrentBufferReadOnly implements core.EditorContext
func (m *model) IsCurrentBufferReadOnly() bool {
	return m.isCurrentBufferReadOnly()
}

// GetCurrentFilePath implements core.EditorContext
func (m *model) GetCurrentFilePath() string {
	return m.getCurrentFilePath()
}

// GetCurrentContent implements core.EditorContext
func (m *model) GetCurrentContent() string {
	return m.syntaxTA.Value()
}

// SaveCurrentBufferState implements core.EditorContext
func (m *model) SaveCurrentBufferState() {
	m.saveCurrentBufferState()
}

// UpdateBufferAfterSave implements core.EditorContext
func (m *model) UpdateBufferAfterSave(content string, fileSize int64) {
	if buf := m.getCurrentBuffer(); buf != nil {
		buf.originalContent = content
		buf.fileSize = fileSize
	}
}

// SetMessage implements core.EditorContext
func (m *model) SetMessage(msg string) {
	m.message = msg
}

// SetError implements core.EditorContext
func (m *model) SetError(err error) {
	m.err = err
}
