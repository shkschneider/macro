package internal

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/api"
)

// CmdQuit is the command name constant for quit
const CmdQuit = "quit"

// QuitKeyBinding is the key binding for the quit command
var QuitKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+q"),
	key.WithHelp("ctrl+q", "quit editor"),
)

func init() {
	api.RegisterCommand(api.CommandRegistration{
		Name:          CmdQuit,
		Key:           "Ctrl-Q",
		Description:   "Quit the editor",
		KeyBinding:    QuitKeyBinding,
		PluginExecute: nil, // Main app provides execute handler
	})
}

// ExecuteQuit quits the editor
func ExecuteQuit(m *Model) tea.Cmd {
	// Save cursor state before quitting
	m.saveCurrentBufferState()
	if m.CursorState != nil {
		_ = m.CursorState.Save()
	}
	return tea.Quit
}

// QuitCommand returns the command definition for quitting
func QuitCommand() api.CommandDef {
	return api.CommandDef{
		Name:        CmdQuit,
		Key:         "Ctrl-Q",
		Description: "Quit the editor",
		KeyBinding:  QuitKeyBinding,
	}
}
