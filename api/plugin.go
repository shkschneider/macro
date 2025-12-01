package api

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// PluginCommand represents a command with its execution logic defined by a plugin
type PluginCommand struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
	Execute     func(ctx EditorContext) tea.Cmd
}
