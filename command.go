package main

import tea "github.com/charmbracelet/bubbletea"

// Command represents an editor command with its keybinding and execution logic
type Command struct {
	Name        string
	Key         string
	Description string
	Execute     func(*model) tea.Cmd
}

var commandRegistry []Command

// registerCommand adds a command to the global registry
func registerCommand(cmd Command) {
	commandRegistry = append(commandRegistry, cmd)
}

// getKeybindings returns all registered commands
func getKeybindings() []Command {
	return commandRegistry
}

// getCommandByName returns a command by its name
func getCommandByName(name string) *Command {
	for i := range commandRegistry {
		if commandRegistry[i].Name == name {
			return &commandRegistry[i]
		}
	}
	return nil
}
