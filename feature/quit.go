package feature

import (
	"github.com/charmbracelet/bubbles/key"
	macro "github.com/shkschneider/macro/core"
)

// QuitKeyBinding is the key binding for the quit command
var QuitKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+q"),
	key.WithHelp("ctrl+q", "quit editor"),
)

// QuitCommand returns the command definition for quitting
func QuitCommand() macro.CommandDef {
	return macro.CommandDef{
		Name:        "quit",
		Key:         "Ctrl-Q",
		Description: "Quit the editor",
		KeyBinding:  QuitKeyBinding,
	}
}
