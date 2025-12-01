package internal

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewCommandInput(t *testing.T) {
	ci := NewCommandInput()
	if ci == nil {
		t.Fatal("NewCommandInput should not return nil")
	}
	if ci.IsActive() {
		t.Error("New command input should not be active")
	}
	if ci.GetMessage() != defaultMessage {
		t.Errorf("New command input should have default message, got %q", ci.GetMessage())
	}
}

func TestCommandInput_SetMessage(t *testing.T) {
	ci := NewCommandInput()
	testMsg := "Test message"
	ci.SetMessage(testMsg)
	if ci.GetMessage() != testMsg {
		t.Errorf("GetMessage should return %q, got %q", testMsg, ci.GetMessage())
	}
}

func TestCommandInput_Activate(t *testing.T) {
	ci := NewCommandInput()
	ci.Activate()
	if !ci.IsActive() {
		t.Error("Command input should be active after Activate()")
	}
}

func TestCommandInput_Deactivate(t *testing.T) {
	ci := NewCommandInput()
	ci.Activate()
	ci.Deactivate()
	if ci.IsActive() {
		t.Error("Command input should not be active after Deactivate()")
	}
}

func TestCommandInput_SetWidth(t *testing.T) {
	ci := NewCommandInput()
	// Just ensure it doesn't panic
	ci.SetWidth(80)
	ci.SetWidth(120)
	ci.SetWidth(40)
}

func TestCommandInput_View_MessageMode(t *testing.T) {
	ci := NewCommandInput()
	ci.SetMessage("Test message")
	view := ci.View(80, nil, false, false)
	if !strings.Contains(view, "Test message") {
		t.Error("View in message mode should contain the message")
	}
}

func TestCommandInput_View_InputMode(t *testing.T) {
	ci := NewCommandInput()
	ci.Activate()
	ci.SetWidth(80)
	view := ci.View(80, nil, false, false)
	// In input mode, the view should contain the prompt
	if !strings.Contains(view, ":") {
		t.Error("View in input mode should contain the prompt")
	}
}

func TestCommandInput_Value(t *testing.T) {
	ci := NewCommandInput()
	ci.Activate()
	// Value should be empty initially
	if ci.Value() != "" {
		t.Errorf("Value should be empty initially, got %q", ci.Value())
	}
}

func TestCommandInput_Update(t *testing.T) {
	ci := NewCommandInput()
	// Update when not active should return nil command
	result, cmd := ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd != nil {
		t.Error("Update when not active should return nil command")
	}
	if result == nil {
		t.Error("Update should return the command input")
	}
}

func TestCommandInput_Update_Active(t *testing.T) {
	ci := NewCommandInput()
	ci.Activate()
	// Update when active should process the message
	result, _ := ci.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if result == nil {
		t.Error("Update should return the command input")
	}
}
