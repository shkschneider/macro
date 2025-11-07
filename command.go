package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// Command represents an editor command with its keybinding and execution logic
type Command struct {
	name        string
	key         string
	description string
	execute     func(*model) tea.Cmd
}

// getKeybindings returns all available commands and their keybindings
func getKeybindings() []Command {
	return []Command{
		{
			name:        "file-save",
			key:         "Ctrl-S",
			description: "Save current buffer to disk",
			execute: func(m *model) tea.Cmd {
				return m.executeFileSave()
			},
		},
		{
			name:        "file-open",
			key:         "Ctrl-Space",
			description: "Open file switcher dialog",
			execute:     nil, // Handled directly in Update
		},
		{
			name:        "buffer-switch",
			key:         "Ctrl-B",
			description: "Open buffer switcher dialog",
			execute:     nil, // Handled directly in Update
		},
		{
			name:        "help-show",
			key:         "Ctrl-H",
			description: "Show this help dialog",
			execute:     nil, // Handled directly in Update
		},
		{
			name:        "quit",
			key:         "Ctrl-Q",
			description: "Quit the editor",
			execute: func(m *model) tea.Cmd {
				return tea.Quit
			},
		},
	}
}

// getCommandByName returns a command by its name
func getCommandByName(name string) *Command {
	keybindings := getKeybindings()
	for i := range keybindings {
		if keybindings[i].name == name {
			return &keybindings[i]
		}
	}
	return nil
}

// executeFileSave saves the current buffer to disk
func (m *model) executeFileSave() tea.Cmd {
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
