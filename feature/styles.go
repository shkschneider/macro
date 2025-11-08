package feature

import "github.com/charmbracelet/lipgloss"

var (
	// DialogBoxStyle is the style for dialog boxes
	DialogBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Background(lipgloss.Color("235"))

	// DialogTitleStyle is the style for dialog titles
	DialogTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

	dialogBoxStyle  = DialogBoxStyle
	dialogTitleStyle = DialogTitleStyle
)
