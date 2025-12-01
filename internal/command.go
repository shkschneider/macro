package internal

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/api"
)

func ExecuteCommandPalette(m *Model) tea.Cmd {
	// Get all commands
	var commands []api.CommandDef
	for _, cmd := range GetKeybindings() {
		commands = append(commands, api.CommandDef{
			Name:        cmd.Name,
			Key:         cmd.Key,
			Description: cmd.Description,
		})
	}
	m.ActiveDialog = NewPaletteDialog(commands)
	return m.ActiveDialog.Init()
}

// Command represents an editor command with its keybinding and execution logic
type Command struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
	Execute     func(*Model) tea.Cmd
}

var CommandRegistry []Command

// registerCommand adds a command to the global registry
func RegisterCommand(cmd Command) {
	CommandRegistry = append(CommandRegistry, cmd)
}

// getKeybindings returns all registered commands
func GetKeybindings() []Command {
	return CommandRegistry
}

// getCommandByName returns a command by its name
func GetCommandByName(name string) *Command {
	for i := range CommandRegistry {
		if CommandRegistry[i].Name == name {
			return &CommandRegistry[i]
		}
	}
	return nil
}

// getCommandByKey returns a command that matches the given key message
func GetCommandByKey(msg tea.KeyMsg) *Command {
	for i := range CommandRegistry {
		if key.Matches(msg, CommandRegistry[i].KeyBinding) {
			return &CommandRegistry[i]
		}
	}
	return nil
}
