package api

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Dialog interface defines the contract for all dialogs
type Dialog interface {
	Init() tea.Cmd
	Update(tea.Msg) (Dialog, tea.Cmd)
	View(termWidth, termHeight int) string
	IsVisible() bool
}

// PluginMsg is an interface for messages from plugins that can handle themselves.
// This allows plugins to define their own message types without the main app
// needing to know about them in a switch statement.
type PluginMsg interface {
	// Handle processes the message and returns any resulting tea.Cmd.
	// The EditorContext provides access to editor state and operations.
	Handle(ctx EditorContext) tea.Cmd
}

// Dialog box styles - used by plugins to render consistent dialogs
var (
	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Background(lipgloss.Color("235"))

	DialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230")).
				Padding(0, 1)

	DialogTitleLineStyle = lipgloss.NewStyle().
				Bold(true)
)

// Dialog item styles (for lists)
var (
	DialogItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			PaddingLeft(2)

	DialogHighlightedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("63")).
				Bold(true).
				PaddingLeft(2)
)

// Dialog UI element styles
var (
	DialogSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	DialogInputLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("63")).
				Bold(true)

	DialogCountStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))

	DialogInstructionsStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)
