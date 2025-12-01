package feature

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	macro "github.com/shkschneider/macro/core"
)

// CommandRegistration represents a command to be registered with the main app
type CommandRegistration struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
	// Execute is set for commands that use EditorContext (like save)
	Execute func(ctx macro.EditorContext) tea.Cmd
}

// Register calls the provided callback for each feature command.
// This allows features to auto-register with the main app's command registry.
func Register(registerFunc func(cmd CommandRegistration)) {
	for _, cmd := range GetCommands() {
		registerFunc(cmd)
	}
}

// GetCommands returns all feature commands for registration.
// Commands with Execute set can be used directly with EditorContext.
// Commands without Execute need the main app to provide an execute handler.
func GetCommands() []CommandRegistration {
	saveCmd := SaveCommand()

	return []CommandRegistration{
		{
			Name:        QuitCommand().Name,
			Key:         QuitCommand().Key,
			Description: QuitCommand().Description,
			KeyBinding:  QuitCommand().KeyBinding,
			Execute:     nil, // Main app provides execute handler
		},
		{
			Name:        saveCmd.Name,
			Key:         saveCmd.Key,
			Description: saveCmd.Description,
			KeyBinding:  saveCmd.KeyBinding,
			Execute:     saveCmd.Execute, // Feature provides execute handler
		},
		{
			Name:        HelpCommand().Name,
			Key:         HelpCommand().Key,
			Description: HelpCommand().Description,
			KeyBinding:  HelpCommand().KeyBinding,
			Execute:     nil, // Main app provides execute handler
		},
		{
			Name:        FileSwitcherCommand().Name,
			Key:         FileSwitcherCommand().Key,
			Description: FileSwitcherCommand().Description,
			KeyBinding:  FileSwitcherCommand().KeyBinding,
			Execute:     nil, // Main app provides execute handler
		},
		{
			Name:        BufferSwitcherCommand().Name,
			Key:         BufferSwitcherCommand().Key,
			Description: BufferSwitcherCommand().Description,
			KeyBinding:  BufferSwitcherCommand().KeyBinding,
			Execute:     nil, // Main app provides execute handler
		},
	}
}
