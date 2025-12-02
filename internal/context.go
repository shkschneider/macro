package internal

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/api"
)

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
	return m.Textarea.Value()
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

// GetBuffers implements api.EditorContext
func (m *Model) GetBuffers() []api.BufferInfo {
	var bufferInfos []api.BufferInfo
	for _, buf := range m.Buffers {
		bufferInfos = append(bufferInfos, api.BufferInfo{
			FilePath: buf.FilePath,
			ReadOnly: buf.ReadOnly,
		})
	}
	return bufferInfos
}

// GetCurrentBufferIndex implements api.EditorContext
func (m *Model) GetCurrentBufferIndex() int {
	return m.CurrentBuffer
}

// SetActiveDialog implements api.EditorContext
func (m *Model) SetActiveDialog(dialog api.Dialog) tea.Cmd {
	m.ActiveDialog = dialog
	if dialog != nil {
		return dialog.Init()
	}
	return nil
}

// SaveCursorState implements api.EditorContext
func (m *Model) SaveCursorState() {
	if m.CursorState != nil {
		_ = m.CursorState.Save()
	}
}

// OpenFile implements api.EditorContext - opens a file into a new buffer
func (m *Model) OpenFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	info, statErr := os.Stat(path)
	readOnly := false
	var fileSize int64
	if statErr == nil {
		readOnly = determineReadOnly(info)
		fileSize = info.Size()
	}

	bufferIdx := m.addBuffer(path, string(content), readOnly, fileSize)
	m.loadBuffer(bufferIdx)
	m.Err = nil
	return nil
}

// SwitchToBuffer implements api.EditorContext - switches to a buffer by index
func (m *Model) SwitchToBuffer(index int) {
	m.loadBuffer(index)
}

// ExecuteCommand implements api.EditorContext - executes a command by name
func (m *Model) ExecuteCommand(name string) tea.Cmd {
	cmd := GetCommandByName(name)
	if cmd != nil && cmd.Execute != nil {
		return cmd.Execute(m)
	}
	return nil
}
