package internal

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/api"
)

// Command represents an editor command with its keybinding and execution logic.
// The Execute function takes *Model directly for core commands that need full access.
type Command struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
	Execute     func(*Model) tea.Cmd
}

var CommandRegistry []Command

// RegisterCommand adds a command to the global registry.
func RegisterCommand(cmd Command) {
	CommandRegistry = append(CommandRegistry, cmd)
}

// RegisterFromAPI registers all commands from the api registry.
// This should be called once at startup to convert api commands to internal commands.
func RegisterFromAPI() {
	for _, cmd := range api.GetCommands() {
		execFunc := wrapPluginExecute(cmd)
		RegisterCommand(Command{
			Name:        cmd.Name,
			Key:         cmd.Key,
			Description: cmd.Description,
			KeyBinding:  cmd.KeyBinding,
			Execute:     execFunc,
		})
	}
}

// wrapPluginExecute wraps an api command's PluginExecute to work with *Model.
func wrapPluginExecute(cmd api.CommandRegistration) func(*Model) tea.Cmd {
	// Special case: Command palette needs access to CommandRegistry
	if cmd.Name == CmdPalette {
		return ExecuteCommandPalette
	}
	if cmd.PluginExecute == nil {
		return nil
	}
	return func(m *Model) tea.Cmd {
		return cmd.PluginExecute(m)
	}
}

// GetKeybindings returns all registered commands.
func GetKeybindings() []Command {
	return CommandRegistry
}

// GetCommandByName returns a command by its name.
func GetCommandByName(name string) *Command {
	for i := range CommandRegistry {
		if CommandRegistry[i].Name == name {
			return &CommandRegistry[i]
		}
	}
	return nil
}

// GetCommandByKey returns a command that matches the given key message.
func GetCommandByKey(msg tea.KeyMsg) *Command {
	for i := range CommandRegistry {
		if key.Matches(msg, CommandRegistry[i].KeyBinding) {
			return &CommandRegistry[i]
		}
	}
	return nil
}
