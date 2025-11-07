package main

import tea "github.com/charmbracelet/bubbletea"

// Dialog interface defines the contract for all dialogs
type Dialog interface {
	Init() tea.Cmd
	Update(tea.Msg) (Dialog, tea.Cmd)
	View(termWidth, termHeight int) string
	IsVisible() bool
}

// Custom message types for dialog results

// FileSelectedMsg is sent when a file is selected in the file dialog
type FileSelectedMsg struct {
	Path string
}

// BufferSelectedMsg is sent when a buffer is selected in the buffer dialog
type BufferSelectedMsg struct {
	Index int
}

// CommandSelectedMsg is sent when a command is selected in the help dialog
type CommandSelectedMsg struct {
	CommandName string
}
