package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// ====== Command Registration ======

func init() {
	registerCommand(Command{
		name:        "file-save",
		key:         "Ctrl-S",
		description: "Save current buffer to disk",
		execute:     executeFileSave,
	})
}

// ====== Command Implementation ======

// executeFileSave saves the current buffer to disk
func executeFileSave(m *model) tea.Cmd {
	readOnly := m.isCurrentBufferReadOnly()
	if readOnly {
		m.message = "WARNING: Cannot save - file is read-only"
		return nil
	}
	filePath := m.getCurrentFilePath()
	if filePath == "" {
		m.message = "Error: No filename specified. Usage: macro <filename>"
		m.err = fmt.Errorf("no filename")
	} else {
		// Save current buffer state first
		m.saveCurrentBufferState()
		err := os.WriteFile(filePath, []byte(m.textarea.Value()), 0644)
		if err != nil {
			m.message = fmt.Sprintf("Error saving: %v", err)
			m.err = err
		} else {
			m.message = fmt.Sprintf("Saved to %s", filePath)
			m.err = nil
		}
	}
	return nil
}
