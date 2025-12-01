package vanilla

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/api"
)

// CmdAddCursor is the command name for adding cursor at next occurrence
const CmdAddCursor = "cursor-add-next"

// CmdClearCursors is the command name for clearing secondary cursors
const CmdClearCursors = "cursor-clear"

// AddCursorKeyBinding is the key binding for adding cursor at next occurrence
var AddCursorKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+d"),
	key.WithHelp("ctrl+d", "add cursor at next occurrence"),
)

// ClearCursorsKeyBinding is the key binding for clearing secondary cursors
// Note: We handle this specially - Escape only clears cursors if there are secondary cursors,
// otherwise it does nothing (or could close dialogs, etc.)
var ClearCursorsKeyBinding = key.NewBinding(
	key.WithKeys("esc"),
	key.WithHelp("esc", "clear secondary cursors"),
)

func init() {
	api.RegisterCommand(api.CommandRegistration{
		Name:          CmdAddCursor,
		Key:           "Ctrl-D",
		Description:   "Add cursor at next occurrence of word",
		KeyBinding:    AddCursorKeyBinding,
		PluginExecute: executeAddCursor,
	})

	api.RegisterCommand(api.CommandRegistration{
		Name:          CmdClearCursors,
		Key:           "Esc",
		Description:   "Clear all secondary cursors",
		KeyBinding:    ClearCursorsKeyBinding,
		PluginExecute: executeClearCursors,
	})
}

// executeAddCursor adds a cursor at the next occurrence of the word under the cursor
func executeAddCursor(ctx api.EditorContext) tea.Cmd {
	if ctx.IsCurrentBufferReadOnly() {
		return nil
	}

	added := ctx.AddCursorAtNextOccurrence()
	if added {
		ctx.SetMessage("Added cursor at next occurrence")
	} else {
		ctx.SetMessage("No more occurrences found")
	}
	return nil
}

// executeClearCursors clears all secondary cursors
func executeClearCursors(ctx api.EditorContext) tea.Cmd {
	if ctx.HasSecondaryCursors() {
		ctx.ClearSecondaryCursors()
		ctx.SetMessage("Cleared all secondary cursors")
	}
	return nil
}
