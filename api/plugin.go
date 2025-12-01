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

// CommandRegistration represents a command to be registered with the main app
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

// Register calls the provided callback for each registered plugin command.
// This allows plugins to auto-register with the main app's command registry.
func Register(registerFunc func(cmd CommandRegistration)) {
	for _, cmd := range registeredCommands {
		registerFunc(cmd)
	}
}

// GetCommands returns a copy of all registered plugin commands.
func GetCommands() []CommandRegistration {
	// Return a copy to prevent external modification
	return append([]CommandRegistration(nil), registeredCommands...)
}
