package feature

// QuitCommand returns the command definition for quitting
func QuitCommand() CommandDef {
	return CommandDef{
		Name:        "quit",
		Key:         "Ctrl-Q",
		Description: "Quit the editor",
	}
}
