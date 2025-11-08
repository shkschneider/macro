package feature

// SaveCommand returns the command definition for saving files
func SaveCommand() CommandDef {
	return CommandDef{
		Name:        "file-save",
		Key:         "Ctrl-S",
		Description: "Save current buffer to disk",
	}
}
