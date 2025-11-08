package feature

import (
	macro "github.com/shkschneider/macro/core"
)

// SaveCommand returns the command definition for saving files
func SaveCommand() macro.CommandDef {
	return macro.CommandDef{
		Name:        "file-save",
		Key:         "Ctrl-S",
		Description: "Save current buffer to disk",
	}
}
