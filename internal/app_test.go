package internal

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/plugins/vanilla"
)

func TestFeatureKeyBinding_Quit(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyCtrlQ}
	if !key.Matches(msg, vanilla.QuitKeyBinding) {
		t.Error("Ctrl+Q should match Quit binding")
	}
}

func TestFeatureKeyBinding_Save(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	if !key.Matches(msg, vanilla.SaveKeyBinding) {
		t.Error("Ctrl+S should match Save binding")
	}
}

func TestFeatureKeyBinding_FileOpen(t *testing.T) {
	// Test ctrl+p
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'p'},
		Alt:   false,
	}
	msg.Type = tea.KeyCtrlP
	if !key.Matches(msg, vanilla.FileSwitcherKeyBinding) {
		t.Error("Ctrl+P should match FileOpen binding")
	}
}

func TestFeatureKeyBinding_BufferSwitch(t *testing.T) {
	// Test ctrl+b
	msg := tea.KeyMsg{
		Type: tea.KeyCtrlB,
	}
	if !key.Matches(msg, vanilla.BufferSwitcherKeyBinding) {
		t.Error("Ctrl+B should match BufferSwitch binding")
	}
}

func TestFeatureKeyBinding_CommandInput(t *testing.T) {
	// Test ctrl+@ (which is what ctrl+space sends)
	msg := tea.KeyMsg{Type: tea.KeyCtrlAt}
	if !key.Matches(msg, CommandInputKeyBinding) {
		t.Error("Ctrl+@ (Ctrl+Space) should match CommandInput binding")
	}
}

func TestFeatureKeyBindings_AllHaveHelp(t *testing.T) {
	// Verify all feature bindings have help text
	bindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Quit", vanilla.QuitKeyBinding},
		{"Save", vanilla.SaveKeyBinding},
		{"CommandInput", CommandInputKeyBinding},
		{"FileOpen", vanilla.FileSwitcherKeyBinding},
		{"BufferSwitch", vanilla.BufferSwitcherKeyBinding},
	}

	for _, b := range bindings {
		help := b.binding.Help()
		if help.Key == "" {
			t.Errorf("%s binding should have help key text", b.name)
		}
		if help.Desc == "" {
			t.Errorf("%s binding should have help description", b.name)
		}
	}
}

func TestGetCommandByKey_ReturnsCorrectCommand(t *testing.T) {
	// Register commands for testing
	CommandRegistry = nil // Reset registry
	RegisterCommand(Command{
		Name:        "test-quit",
		Key:         "Ctrl-Q",
		Description: "Test quit",
		KeyBinding:  vanilla.QuitKeyBinding,
		Execute:     nil,
	})
	RegisterCommand(Command{
		Name:        "test-save",
		Key:         "Ctrl-S",
		Description: "Test save",
		KeyBinding:  vanilla.SaveKeyBinding,
		Execute:     nil,
	})

	// Test that Ctrl+Q matches quit command
	quitMsg := tea.KeyMsg{Type: tea.KeyCtrlQ}
	cmd := GetCommandByKey(quitMsg)
	if cmd == nil {
		t.Error("GetCommandByKey should return command for Ctrl+Q")
	} else if cmd.Name != "test-quit" {
		t.Errorf("GetCommandByKey returned wrong command, got %s, want test-quit", cmd.Name)
	}

	// Test that Ctrl+S matches save command
	saveMsg := tea.KeyMsg{Type: tea.KeyCtrlS}
	cmd = GetCommandByKey(saveMsg)
	if cmd == nil {
		t.Error("GetCommandByKey should return command for Ctrl+S")
	} else if cmd.Name != "test-save" {
		t.Errorf("GetCommandByKey returned wrong command, got %s, want test-save", cmd.Name)
	}

	// Test that unregistered key returns nil
	unknownMsg := tea.KeyMsg{Type: tea.KeyCtrlX}
	cmd = GetCommandByKey(unknownMsg)
	if cmd != nil {
		t.Error("GetCommandByKey should return nil for unregistered key")
	}

	// Reset registry
	CommandRegistry = nil
}

func TestBuffer_IsModified(t *testing.T) {
	// Test unmodified buffer
	buf := Buffer{
		FilePath:        "/path/to/test.go",
		Content:         "original content",
		OriginalContent: "original content",
		ReadOnly:        false,
		FileSize:        16,
	}
	if buf.IsModified() {
		t.Error("Buffer should not be modified when content equals original")
	}

	// Test modified buffer
	buf.Content = "modified content"
	if !buf.IsModified() {
		t.Error("Buffer should be modified when content differs from original")
	}
}

func TestBuildStatusBar_NewFile(t *testing.T) {
	m := InitialModel("")
	// Set termWidth for test
	TermWidth = 80

	statusBar := m.BuildStatusBar()
	if !strings.Contains(statusBar, "New File") {
		t.Error("Status bar should show 'New File' when no buffer is loaded")
	}
}

func TestBuildStatusBar_WithFile(t *testing.T) {
	m := InitialModel("")
	// Set termWidth for test
	TermWidth = 120

	// Add a buffer
	m.Buffers = []Buffer{
		{
			FilePath:        "/path/to/test.go",
			Content:         "package main",
			OriginalContent: "package main",
			ReadOnly:        false,
			FileSize:        100,
		},
	}
	m.CurrentBuffer = 0

	statusBar := m.BuildStatusBar()

	// Check left side content
	if !strings.Contains(statusBar, "test.go") {
		t.Error("Status bar should contain filename")
	}
	if !strings.Contains(statusBar, "[Go]") {
		t.Error("Status bar should contain language in brackets")
	}
	if !strings.Contains(statusBar, "100 B") {
		t.Error("Status bar should contain human-readable file size")
	}

	// Check right side content
	if !strings.Contains(statusBar, "[/path/to/]") {
		t.Error("Status bar should contain directory path")
	}
}

func TestBuildStatusBar_ModifiedFile(t *testing.T) {
	m := InitialModel("")
	// Set termWidth for test
	TermWidth = 120

	// Add a buffer with modification tracking
	m.Buffers = []Buffer{
		{
			FilePath:        "/path/to/test.go",
			Content:         "package modified",
			OriginalContent: "package main",
			ReadOnly:        false,
			FileSize:        100,
		},
	}
	m.CurrentBuffer = 0
	// Set the textarea value to simulate modification
	m.Textarea.SetValue("package modified")

	statusBar := m.BuildStatusBar()

	// Modified file should have asterisk
	if !strings.Contains(statusBar, "test.go*") {
		t.Error("Status bar should show asterisk for modified file")
	}
}

func TestBuildStatusBar_ReadOnlyFile(t *testing.T) {
	m := InitialModel("")
	// Set termWidth for test
	TermWidth = 120

	// Add a read-only buffer
	m.Buffers = []Buffer{
		{
			FilePath:        "/path/to/readonly.txt",
			Content:         "read only content",
			OriginalContent: "read only content",
			ReadOnly:        true,
			FileSize:        17,
		},
	}
	m.CurrentBuffer = 0

	statusBar := m.BuildStatusBar()

	// Read-only file should have (read-only) indicator
	if !strings.Contains(statusBar, "(read-only)") {
		t.Error("Status bar should show [RO] for read-only file")
	}
}

func TestBuildStatusBar_CursorPosition(t *testing.T) {
	m := InitialModel("")
	// Set termWidth for test
	TermWidth = 120

	// Add a buffer
	m.Buffers = []Buffer{
		{
			FilePath:        "/path/to/test.go",
			Content:         "package main",
			OriginalContent: "package main",
			ReadOnly:        false,
			FileSize:        100,
		},
	}
	m.CurrentBuffer = 0

	statusBar := m.BuildStatusBar()

	// Status bar should contain cursor position in line:col format
	// Default position is 1:1 (first line, first column)
	if !strings.Contains(statusBar, "1:1") {
		t.Error("Status bar should contain cursor position (1:1 for default)")
	}
}
