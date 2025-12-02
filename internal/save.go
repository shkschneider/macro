package internal

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/api"
)

// CmdSave is the command name constant for save
const CmdSave = "file-save"

// SaveKeyBinding is the key binding for the save command
var SaveKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+s"),
	key.WithHelp("ctrl+s", "save file"),
)

func init() {
	api.RegisterCommand(api.CommandRegistration{
		Name:          CmdSave,
		Key:           "Ctrl-S",
		Description:   "Save current buffer to disk",
		KeyBinding:    SaveKeyBinding,
		PluginExecute: nil, // Main app provides execute handler
	})
}

// ExecuteSave saves the current buffer to disk
func ExecuteSave(m *Model) tea.Cmd {
	if m.isCurrentBufferReadOnly() {
		m.Message = "WARNING: Cannot save - file is read-only"
		return nil
	}

	filePath := m.getCurrentFilePath()
	if filePath == "" {
		m.Message = "Error: No filename specified. Usage: macro <filename>"
		m.Err = fmt.Errorf("no filename")
		return nil
	}

	// Save current buffer state first
	m.saveCurrentBufferState()
	content := m.Textarea.Value()
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		m.Message = fmt.Sprintf("Error saving: %v", err)
		m.Err = err
		return nil
	}

	m.Message = fmt.Sprintf("Saved to %s", filePath)
	m.Err = nil

	// Update original content and file size after successful save
	if info, err := os.Stat(filePath); err == nil {
		if buf := m.getCurrentBuffer(); buf != nil {
			buf.OriginalContent = content
			buf.FileSize = info.Size()
		}
	}

	return nil
}

// SaveCommand returns the command definition for saving files
func SaveCommand() api.CommandDef {
	return api.CommandDef{
		Name:        CmdSave,
		Key:         "Ctrl-S",
		Description: "Save current buffer to disk",
		KeyBinding:  SaveKeyBinding,
	}
}
