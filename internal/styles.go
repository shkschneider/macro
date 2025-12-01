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

// Dialog box styles
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

// Diff indicator styles for showing line changes
var (
	// DiffAddedStyle is used for added lines (green)
	DiffAddedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	// DiffDeletedStyle is used for deleted lines (red)
	DiffDeletedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	// DiffModifiedStyle is used for modified lines (yellow)
	DiffModifiedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)
