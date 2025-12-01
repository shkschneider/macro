package internal

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/api"
)

func TestCmdPalette_Constant(t *testing.T) {
	if CmdPalette != "help-show" {
		t.Errorf("CmdPalette should be 'help-show', got '%s'", CmdPalette)
	}
}

func TestPaletteKeyBinding(t *testing.T) {
	// Test that ctrl+@ (ctrl+space) matches the binding
	msg := tea.KeyMsg{Type: tea.KeyCtrlAt}
	if !key.Matches(msg, PaletteKeyBinding) {
		t.Error("Ctrl+@ (Ctrl+Space) should match PaletteKeyBinding")
	}
}

func TestNewPaletteDialog(t *testing.T) {
	commands := []api.CommandDef{
		{Name: "cmd1", Key: "Ctrl-1", Description: "Command 1"},
		{Name: "cmd2", Key: "Ctrl-2", Description: "Command 2"},
	}

	dialog := NewPaletteDialog(commands)

	if dialog == nil {
		t.Fatal("NewPaletteDialog should not return nil")
	}
	if !dialog.visible {
		t.Error("New dialog should be visible")
	}
	if len(dialog.allCommands) != 2 {
		t.Errorf("Dialog should have 2 commands, got %d", len(dialog.allCommands))
	}
	if len(dialog.filteredCommands) != 2 {
		t.Errorf("Dialog should have 2 filtered commands initially, got %d", len(dialog.filteredCommands))
	}
	if dialog.selectedIdx != 0 {
		t.Errorf("Dialog should start with selectedIdx 0, got %d", dialog.selectedIdx)
	}
}

func TestPaletteDialog_IsVisible(t *testing.T) {
	dialog := NewPaletteDialog([]api.CommandDef{})

	if !dialog.IsVisible() {
		t.Error("New dialog should be visible")
	}

	dialog.visible = false
	if dialog.IsVisible() {
		t.Error("Dialog should not be visible after setting visible to false")
	}
}

func TestPaletteDialog_Init(t *testing.T) {
	dialog := NewPaletteDialog([]api.CommandDef{})
	cmd := dialog.Init()

	if cmd == nil {
		t.Error("Init should return a command (for textinput.Blink)")
	}
}

func TestPaletteDialog_View(t *testing.T) {
	commands := []api.CommandDef{
		{Name: "test-cmd", Key: "Ctrl-T", Description: "Test command"},
	}
	dialog := NewPaletteDialog(commands)

	view := dialog.View(80, 24)

	if view == "" {
		t.Error("View should return non-empty string when visible")
	}

	// Check that it contains expected elements
	if len(view) < 10 {
		t.Error("View should contain substantial content")
	}
}

func TestPaletteDialog_View_Hidden(t *testing.T) {
	dialog := NewPaletteDialog([]api.CommandDef{})
	dialog.visible = false

	view := dialog.View(80, 24)

	if view != "" {
		t.Error("View should return empty string when not visible")
	}
}

func TestPaletteDialog_Update_Close(t *testing.T) {
	dialog := NewPaletteDialog([]api.CommandDef{})

	// Test ESC key closes dialog
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	result, _ := dialog.Update(escMsg)
	resultDialog := result.(*PaletteDialog)

	if resultDialog.visible {
		t.Error("ESC should close dialog")
	}
}

func TestPaletteDialog_Update_Navigation(t *testing.T) {
	commands := []api.CommandDef{
		{Name: "cmd1", Key: "Ctrl-1", Description: "Command 1"},
		{Name: "cmd2", Key: "Ctrl-2", Description: "Command 2"},
		{Name: "cmd3", Key: "Ctrl-3", Description: "Command 3"},
	}
	dialog := NewPaletteDialog(commands)

	// Initial position should be 0
	if dialog.selectedIdx != 0 {
		t.Errorf("Initial selectedIdx should be 0, got %d", dialog.selectedIdx)
	}

	// Move down
	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	result, _ := dialog.Update(downMsg)
	resultDialog := result.(*PaletteDialog)
	if resultDialog.selectedIdx != 1 {
		t.Errorf("After down, selectedIdx should be 1, got %d", resultDialog.selectedIdx)
	}

	// Move up
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	result, _ = resultDialog.Update(upMsg)
	resultDialog = result.(*PaletteDialog)
	if resultDialog.selectedIdx != 0 {
		t.Errorf("After up, selectedIdx should be 0, got %d", resultDialog.selectedIdx)
	}

	// Try to move up past 0 (should stay at 0)
	result, _ = resultDialog.Update(upMsg)
	resultDialog = result.(*PaletteDialog)
	if resultDialog.selectedIdx != 0 {
		t.Errorf("Should not go below 0, got %d", resultDialog.selectedIdx)
	}
}

func TestPaletteDialog_ApplyFuzzyFilter_Empty(t *testing.T) {
	commands := []api.CommandDef{
		{Name: "save", Key: "Ctrl-S", Description: "Save file"},
		{Name: "quit", Key: "Ctrl-Q", Description: "Quit editor"},
	}
	dialog := NewPaletteDialog(commands)

	// Simulate setting empty query
	dialog.filterInput.SetValue("")
	dialog.applyFuzzyFilter()

	// All commands should be shown
	if len(dialog.filteredCommands) != 2 {
		t.Errorf("Empty filter should show all commands, got %d", len(dialog.filteredCommands))
	}
}

func TestCommandSelectedMsg_Handle(t *testing.T) {
	// Reset registry
	oldRegistry := CommandRegistry
	CommandRegistry = nil
	defer func() { CommandRegistry = oldRegistry }()

	// Register a test command
	executed := false
	RegisterCommand(Command{
		Name: "test-execute",
		Execute: func(m *Model) tea.Cmd {
			executed = true
			return nil
		},
	})

	// Create a mock model as context
	m := InitialModel("")

	msg := CommandSelectedMsg{CommandName: "test-execute"}
	msg.Handle(&m)

	if !executed {
		t.Error("CommandSelectedMsg.Handle should execute the command")
	}
}
