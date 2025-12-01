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
	if m.cursorState != nil {
		_ = m.cursorState.Save()
	}
	return tea.Quit
}

// executeFileSwitcher opens the file switcher dialog
func ExecuteFileSwitcher(m *Model) tea.Cmd {
	if m.getCurrentFilePath() != "" {
		m.activeDialog = vanilla.NewFileDialog(filepath.Dir(m.getCurrentFilePath()))
		return m.activeDialog.Init()
	}
	m.message = "No file open to determine directory"
	return nil
}

// executeBufferSwitcher opens the buffer switcher dialog
func ExecuteBufferSwitcher(m *Model) tea.Cmd {
	if len(m.buffers) > 0 {
		// Convert buffers to BufferInfo
		var bufferInfos []api.BufferInfo
		for _, buf := range m.buffers {
			bufferInfos = append(bufferInfos, api.BufferInfo{
				FilePath: buf.filePath,
				ReadOnly: buf.readOnly,
			})
		}
		m.activeDialog = vanilla.NewBufferDialog(bufferInfos, m.currentBuffer)
		return m.activeDialog.Init()
	}
	m.message = "No buffers open"
	return nil
}

// executeCommandPalette opens the command palette dialog
func ExecuteCommandPalette(m *Model) tea.Cmd {
	// Get all commands
	var commands []api.CommandDef
	for _, cmd := range GetKeybindings() {
		commands = append(commands, api.CommandDef{
			Name:        cmd.Name,
			Key:         cmd.Key,
			Description: cmd.Description,
		})
	}
	m.activeDialog = vanilla.NewHelpDialog(commands)
	return m.activeDialog.Init()
}
