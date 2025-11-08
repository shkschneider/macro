package main

import tea "github.com/charmbracelet/bubbletea"

// ====== Command Registration ======

func init() {
	registerCommand(Command{
		name:        "quit",
		key:         "Ctrl-Q",
		description: "Quit the editor",
		execute:     executeQuit,
	})
}

// ====== Command Implementation ======

// executeQuit quits the editor
func executeQuit(m *model) tea.Cmd {
	return tea.Quit
}
