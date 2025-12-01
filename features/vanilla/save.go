package vanilla

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	macro "github.com/shkschneider/macro/core"
)

// SaveKeyBinding is the key binding for the save command
var SaveKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+s"),
	key.WithHelp("ctrl+s", "save file"),
)

// SaveCommand returns the command definition for saving files with execution logic
func SaveCommand() macro.FeatureCommand {
	return macro.FeatureCommand{
		Name:        "file-save",
		Key:         "Ctrl-S",
		Description: "Save current buffer to disk",
		KeyBinding:  SaveKeyBinding,
		Execute:     executeSave,
	}
}

// executeSave saves the current buffer to disk
func executeSave(ctx macro.EditorContext) tea.Cmd {
	if ctx.IsCurrentBufferReadOnly() {
		ctx.SetMessage("WARNING: Cannot save - file is read-only")
		return nil
	}

	filePath := ctx.GetCurrentFilePath()
	if filePath == "" {
		ctx.SetMessage("Error: No filename specified. Usage: macro <filename>")
		ctx.SetError(fmt.Errorf("no filename"))
		return nil
	}

	// Save current buffer state first
	ctx.SaveCurrentBufferState()
	content := ctx.GetCurrentContent()
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		ctx.SetMessage(fmt.Sprintf("Error saving: %v", err))
		ctx.SetError(err)
		return nil
	}

	ctx.SetMessage(fmt.Sprintf("Saved to %s", filePath))
	ctx.SetError(nil)

	// Update original content and file size after successful save
	if info, err := os.Stat(filePath); err == nil {
		ctx.UpdateBufferAfterSave(content, info.Size())
	}

	return nil
}
