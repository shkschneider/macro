package internal

import "github.com/charmbracelet/lipgloss"

// StatusBar styles
var (
	StatusBarStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("15")). // White background
		Foreground(lipgloss.Color("0")).  // Black foreground
		Bold(true).
		Padding(0, 1) // Add horizontal padding
)

// Message styles
var (
	MessageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	WarningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
)

// Dialog styles are defined in api package - use api.DialogBoxStyle, etc.

// Diff indicator styles for showing line changes
var (
	// DiffAddedStyle is used for added lines (green)
	DiffAddedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	// DiffDeletedStyle is used for deleted lines (red)
	DiffDeletedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	// DiffModifiedStyle is used for modified lines (yellow)
	DiffModifiedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)
