// Package api provides the public API for macro plugins.
// This package contains only the types and interfaces that plugins need to implement.
package api

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommandDef defines a command without execution logic
type CommandDef struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
}

// BufferInfo contains information about a buffer for dialogs
type BufferInfo struct {
	FilePath string
	ReadOnly bool
}

// EditorContext provides an interface for plugins to interact with the editor.
// This allows plugins to define their own execution logic without depending on
// the main package's model type directly.
type EditorContext interface {
	// Buffer operations
	IsCurrentBufferReadOnly() bool
	GetCurrentFilePath() string
	GetCurrentContent() string
	SaveCurrentBufferState()
	UpdateBufferAfterSave(content string, fileSize int64)

	// Message operations
	SetMessage(msg string)
	SetError(err error)
}

// PluginCommand represents a command with its execution logic defined by a plugin
type PluginCommand struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
	Execute     func(ctx EditorContext) tea.Cmd
}

// Dialog interface defines the contract for all dialogs
type Dialog interface {
	Init() tea.Cmd
	Update(tea.Msg) (Dialog, tea.Cmd)
	View(termWidth, termHeight int) string
	IsVisible() bool
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
