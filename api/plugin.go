package api

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandRegistration represents a command to be registered with the main app.
// Plugins define commands with their keybindings and optional execution logic.
type CommandRegistration struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
	// PluginExecute is set for commands that use EditorContext (like save).
	// Commands without PluginExecute need the main app to provide an execute handler.
	PluginExecute func(ctx EditorContext) tea.Cmd
}

// Global registry of all plugin commands
var registeredCommands []CommandRegistration

// RegisterCommand adds a command to the global plugin registry.
// Plugins should call this in their init() function to self-register.
func RegisterCommand(cmd CommandRegistration) {
	registeredCommands = append(registeredCommands, cmd)
}

// GetCommands returns a copy of all registered plugin commands.
func GetCommands() []CommandRegistration {
	// Return a copy to prevent external modification
	return append([]CommandRegistration(nil), registeredCommands...)
}
