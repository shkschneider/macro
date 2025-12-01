package vanilla

import (
	"github.com/charmbracelet/bubbles/key"
	api "github.com/shkschneider/macro/api"
	plugin "github.com/shkschneider/macro/plugins"
)

// CmdQuit is the command name constant for quit
const CmdQuit = "quit"

// QuitKeyBinding is the key binding for the quit command
var QuitKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+q"),
	key.WithHelp("ctrl+q", "quit editor"),
)

func init() {
	plugin.RegisterCommand(plugin.CommandRegistration{
		Name:           CmdQuit,
		Key:            "Ctrl-Q",
		Description:    "Quit the editor",
		KeyBinding:     QuitKeyBinding,
		PluginExecute: nil, // Main app provides execute handler
	})
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
