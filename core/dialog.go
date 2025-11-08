package core

import tea "github.com/charmbracelet/bubbletea"

// Dialog interface defines the contract for all dialogs
type Dialog interface {
	Init() tea.Cmd
	Update(tea.Msg) (Dialog, tea.Cmd)
	View(termWidth, termHeight int) string
	IsVisible() bool
}
