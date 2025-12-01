package internal

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestRegisterCommand(t *testing.T) {
	// Reset registry before test
	oldRegistry := CommandRegistry
	CommandRegistry = nil
	defer func() { CommandRegistry = oldRegistry }()

	cmd := Command{
		Name:        "test-cmd",
		Key:         "Ctrl-T",
		Description: "Test command",
		KeyBinding: key.NewBinding(
			key.WithKeys("ctrl+t"),
		),
		Execute: nil,
	}

	RegisterCommand(cmd)

	if len(CommandRegistry) != 1 {
		t.Errorf("Expected 1 command in registry, got %d", len(CommandRegistry))
	}
	if CommandRegistry[0].Name != "test-cmd" {
		t.Errorf("Expected command name 'test-cmd', got '%s'", CommandRegistry[0].Name)
	}
}

func TestGetCommands(t *testing.T) {
	// Reset registry before test
	oldRegistry := CommandRegistry
	CommandRegistry = nil
	defer func() { CommandRegistry = oldRegistry }()

	// Register some commands
	RegisterCommand(Command{Name: "cmd1", Key: "Ctrl-1"})
	RegisterCommand(Command{Name: "cmd2", Key: "Ctrl-2"})

	commands := GetCommands()

	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
}

func TestGetCommandByName(t *testing.T) {
	// Reset registry before test
	oldRegistry := CommandRegistry
	CommandRegistry = nil
	defer func() { CommandRegistry = oldRegistry }()

	RegisterCommand(Command{Name: "find-me", Key: "Ctrl-F", Description: "Find command"})
	RegisterCommand(Command{Name: "other", Key: "Ctrl-O", Description: "Other command"})

	// Test finding existing command
	cmd := GetCommandByName("find-me")
	if cmd == nil {
		t.Error("GetCommandByName should find 'find-me' command")
	} else if cmd.Name != "find-me" {
		t.Errorf("Expected command name 'find-me', got '%s'", cmd.Name)
	}

	// Test finding non-existent command
	cmd = GetCommandByName("not-exists")
	if cmd != nil {
		t.Error("GetCommandByName should return nil for non-existent command")
	}
}

func TestGetCommandByKey(t *testing.T) {
	// Reset registry before test
	oldRegistry := CommandRegistry
	CommandRegistry = nil
	defer func() { CommandRegistry = oldRegistry }()

	testBinding := key.NewBinding(key.WithKeys("ctrl+t"))
	RegisterCommand(Command{
		Name:       "test-by-key",
		Key:        "Ctrl-T",
		KeyBinding: testBinding,
	})

	// Test finding by key
	msg := tea.KeyMsg{Type: tea.KeyCtrlT}
	cmd := GetCommandByKey(msg)
	if cmd == nil {
		t.Error("GetCommandByKey should find command for Ctrl+T")
	} else if cmd.Name != "test-by-key" {
		t.Errorf("Expected command name 'test-by-key', got '%s'", cmd.Name)
	}

	// Test with non-matching key
	unknownMsg := tea.KeyMsg{Type: tea.KeyCtrlX}
	cmd = GetCommandByKey(unknownMsg)
	if cmd != nil {
		t.Error("GetCommandByKey should return nil for unregistered key")
	}
}
