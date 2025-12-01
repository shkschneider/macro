package internal

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/api"
	"github.com/shkschneider/macro/plugins/vanilla"
)

// executeQuit quits the editor
func ExecuteQuit(m *Model) tea.Cmd {
	// Save cursor state before quitting
	m.saveCurrentBufferState()
	if m.CursorState != nil {
		_ = m.CursorState.Save()
	}
	return tea.Quit
}

// executeFileSwitcher opens the file switcher dialog
func ExecuteFileSwitcher(m *Model) tea.Cmd {
	if m.getCurrentFilePath() != "" {
		m.ActiveDialog = vanilla.NewFileDialog(filepath.Dir(m.getCurrentFilePath()))
		return m.ActiveDialog.Init()
	}
	m.Message = "No file open to determine directory"
	return nil
}

// executeBufferSwitcher opens the buffer switcher dialog
func ExecuteBufferSwitcher(m *Model) tea.Cmd {
	if len(m.Buffers) > 0 {
		// Convert buffers to BufferInfo
		var bufferInfos []api.BufferInfo
		for _, buf := range m.Buffers {
			bufferInfos = append(bufferInfos, api.BufferInfo{
				FilePath: buf.FilePath,
				ReadOnly: buf.ReadOnly,
			})
		}
		m.ActiveDialog = vanilla.NewBufferDialog(bufferInfos, m.CurrentBuffer)
		return m.ActiveDialog.Init()
	}
	m.Message = "No buffers open"
	return nil
}
