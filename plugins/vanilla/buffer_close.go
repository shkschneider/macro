package vanilla

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/api"
)

// CmdBufferClose is the command name constant for buffer close
const CmdBufferClose = "buffer-close"

// BufferCloseKeyBinding is the key binding for the buffer close command
var BufferCloseKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+w"),
	key.WithHelp("ctrl+w", "close buffer"),
)

func init() {
	api.RegisterCommand(api.CommandRegistration{
		Name:          CmdBufferClose,
		Key:           "Ctrl-W",
		Description:   "Close current buffer",
		KeyBinding:    BufferCloseKeyBinding,
		PluginExecute: ExecuteBufferClose,
	})
}

// ExecuteBufferClose closes the current buffer, showing confirmation if there are unsaved changes
func ExecuteBufferClose(ctx api.EditorContext) tea.Cmd {
	// Check if current buffer has unsaved changes
	if ctx.IsCurrentBufferModified() {
		// Get directory for potential file picker
		dir := filepath.Dir(ctx.GetCurrentFilePath())
		dialog := NewBufferCloseConfirmDialog(dir)
		return ctx.SetActiveDialog(dialog)
	}

	// No unsaved changes, close immediately
	return closeBufferNow(ctx)
}

// closeBufferNow performs the actual buffer close operation
func closeBufferNow(ctx api.EditorContext) tea.Cmd {
	// Get directory before closing (for file picker if needed)
	dir := filepath.Dir(ctx.GetCurrentFilePath())

	// Close the buffer
	wasLastBuffer := ctx.CloseCurrentBuffer()

	if wasLastBuffer {
		// Show file picker in the directory of the closed file
		ctx.SetMessage("Last buffer closed. Select a file to open.")
		return ctx.ShowFilePicker(dir)
	}
	
	ctx.SetMessage("Buffer closed")
	return nil
}

// ====== Message Types ======

// BufferCloseConfirmedMsg is sent when the user confirms closing the buffer
type BufferCloseConfirmedMsg struct {
	Directory string
}

// Handle implements api.PluginMsg - performs the actual buffer close
func (msg BufferCloseConfirmedMsg) Handle(ctx api.EditorContext) tea.Cmd {
	return closeBufferNow(ctx)
}

// ====== Key Bindings ======

// BufferCloseConfirmDialogKeyMap defines the key bindings for the buffer close confirm dialog
type BufferCloseConfirmDialogKeyMap struct {
	Yes   key.Binding
	No    key.Binding
	Close key.Binding
}

// DefaultBufferCloseConfirmDialogKeyMap returns the default key bindings
var DefaultBufferCloseConfirmDialogKeyMap = BufferCloseConfirmDialogKeyMap{
	Yes: key.NewBinding(
		key.WithKeys("y", "Y"),
		key.WithHelp("y", "yes, close"),
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

// BufferCloseConfirmDialog implements the Dialog interface for buffer close confirmation
type BufferCloseConfirmDialog struct {
	visible   bool
	directory string
}

// NewBufferCloseConfirmDialog creates a new buffer close confirmation dialog
func NewBufferCloseConfirmDialog(directory string) *BufferCloseConfirmDialog {
	return &BufferCloseConfirmDialog{
		visible:   true,
		directory: directory,
	}
}

func (d *BufferCloseConfirmDialog) Init() tea.Cmd {
	return nil
}

func (d *BufferCloseConfirmDialog) Update(msg tea.Msg) (api.Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, DefaultBufferCloseConfirmDialogKeyMap.Yes) {
			d.visible = false
			return d, func() tea.Msg {
				return BufferCloseConfirmedMsg{Directory: d.directory}
			}
		}
		if key.Matches(msg, DefaultBufferCloseConfirmDialogKeyMap.No) ||
			key.Matches(msg, DefaultBufferCloseConfirmDialogKeyMap.Close) {
			d.visible = false
			return d, nil
		}
	}
	return d, nil
}

func (d *BufferCloseConfirmDialog) View(termWidth, termHeight int) string {
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
		Render("This buffer has unsaved changes. Close anyway?")

	instructions := api.DialogInstructionsStyle.
		Render(fmt.Sprintf("%-20s %s", "[Y] Yes, close", "[N/Esc] No, cancel"))

	fullContent := fmt.Sprintf("%s\n%s\n\n%s\n\n%s",
		titleLine,
		separator,
		message,
		instructions,
	)

	return api.DialogBoxStyle.Render(fullContent)
}

func (d *BufferCloseConfirmDialog) IsVisible() bool {
	return d.visible
}
