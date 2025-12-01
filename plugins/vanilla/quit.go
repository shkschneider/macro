package vanilla

import (
	"fmt"
	"strings"

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
		PluginExecute: ExecuteQuit,
	})
}

// ExecuteQuit quits the editor, showing confirmation if there are unsaved changes
func ExecuteQuit(ctx api.EditorContext) tea.Cmd {
	// Check if any buffer has unsaved changes
	if ctx.HasUnsavedChanges() {
		dialog := NewQuitConfirmDialog()
		return ctx.SetActiveDialog(dialog)
	}
	
	// No unsaved changes, quit immediately
	ctx.SaveCurrentBufferState()
	ctx.SaveCursorState()
	return tea.Quit
}

// ====== Message Types ======

// QuitConfirmedMsg is sent when the user confirms quitting
type QuitConfirmedMsg struct{}

// Handle implements api.PluginMsg - performs the actual quit
func (msg QuitConfirmedMsg) Handle(ctx api.EditorContext) tea.Cmd {
	ctx.SaveCurrentBufferState()
	ctx.SaveCursorState()
	return tea.Quit
}

// ====== Key Bindings ======

// QuitConfirmDialogKeyMap defines the key bindings for the quit confirm dialog
type QuitConfirmDialogKeyMap struct {
	Yes   key.Binding
	No    key.Binding
	Close key.Binding
}

// DefaultQuitConfirmDialogKeyMap returns the default key bindings
var DefaultQuitConfirmDialogKeyMap = QuitConfirmDialogKeyMap{
	Yes: key.NewBinding(
		key.WithKeys("y", "Y"),
		key.WithHelp("y", "yes, quit"),
	),
	No: key.NewBinding(
		key.WithKeys("n", "N", "esc"),
		key.WithHelp("n/esc", "no, cancel"),
	),
	Close: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "cancel"),
	),
}

// ====== Dialog Implementation ======

// QuitConfirmDialog implements the Dialog interface for quit confirmation
type QuitConfirmDialog struct {
	visible bool
}

// NewQuitConfirmDialog creates a new quit confirmation dialog
func NewQuitConfirmDialog() *QuitConfirmDialog {
	return &QuitConfirmDialog{
		visible: true,
	}
}

func (d *QuitConfirmDialog) Init() tea.Cmd {
	return nil
}

func (d *QuitConfirmDialog) Update(msg tea.Msg) (api.Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, DefaultQuitConfirmDialogKeyMap.Yes) {
			d.visible = false
			return d, func() tea.Msg {
				return QuitConfirmedMsg{}
			}
		}
		if key.Matches(msg, DefaultQuitConfirmDialogKeyMap.No) ||
			key.Matches(msg, DefaultQuitConfirmDialogKeyMap.Close) {
			d.visible = false
			return d, nil
		}
	}
	return d, nil
}

func (d *QuitConfirmDialog) View(termWidth, termHeight int) string {
	if !d.visible {
		return ""
	}

	dialogWidth := 50
	if termWidth < dialogWidth+4 {
		dialogWidth = termWidth - 4
	}

	title := api.DialogTitleStyle.Render("Unsaved Changes")
	titleLine := api.DialogTitleLineStyle.
		Width(dialogWidth - 4).
		Render(title)

	separator := api.DialogSeparatorStyle.
		Render(strings.Repeat("─", dialogWidth-4))

	message := api.DialogItemStyle.
		Width(dialogWidth - 4).
		Render("You have unsaved changes. Quit anyway?")

	instructions := api.DialogInstructionsStyle.
		Render(fmt.Sprintf("%-20s %s", "[Y] Yes, quit", "[N/Esc] No, cancel"))

	fullContent := fmt.Sprintf("%s\n%s\n\n%s\n\n%s",
		titleLine,
		separator,
		message,
		instructions,
	)

	return api.DialogBoxStyle.Render(fullContent)
}

func (d *QuitConfirmDialog) IsVisible() bool {
	return d.visible
}
