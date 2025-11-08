package feature

import (
	macro "github.com/shkschneider/macro/core"
)

// QuitCommand returns the command definition for quitting
func QuitCommand() macro.CommandDef {
	return macro.CommandDef{
		Name:        "quit",
		Key:         "Ctrl-Q",
		Description: "Quit the editor",
	}
}
