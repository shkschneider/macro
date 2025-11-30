package main

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestDefaultKeyMap_Quit(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyCtrlQ}
	if !key.Matches(msg, DefaultKeyMap.Quit) {
		t.Error("Ctrl+Q should match Quit binding")
	}
}

func TestDefaultKeyMap_Save(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyCtrlS}
	if !key.Matches(msg, DefaultKeyMap.Save) {
		t.Error("Ctrl+S should match Save binding")
	}
}

func TestDefaultKeyMap_FileOpen(t *testing.T) {
	// Test ctrl+p
	msg := tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'p'},
		Alt:   false,
	}
	msg.Type = tea.KeyCtrlP
	if !key.Matches(msg, DefaultKeyMap.FileOpen) {
		t.Error("Ctrl+P should match FileOpen binding")
	}
}

func TestDefaultKeyMap_BufferSwitch(t *testing.T) {
	// Test ctrl+b
	msg := tea.KeyMsg{
		Type: tea.KeyCtrlB,
	}
	if !key.Matches(msg, DefaultKeyMap.BufferSwitch) {
		t.Error("Ctrl+B should match BufferSwitch binding")
	}
}

func TestDefaultKeyMap_CommandPalette(t *testing.T) {
	// Test ctrl+@ (which is what ctrl+space sends)
	msg := tea.KeyMsg{Type: tea.KeyCtrlAt}
	if !key.Matches(msg, DefaultKeyMap.CommandPalette) {
		t.Error("Ctrl+@ (Ctrl+Space) should match CommandPalette binding")
	}
}

func TestKeyMap_AllBindingsHaveHelp(t *testing.T) {
	// Verify all bindings have help text
	bindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Quit", DefaultKeyMap.Quit},
		{"Save", DefaultKeyMap.Save},
		{"CommandPalette", DefaultKeyMap.CommandPalette},
		{"FileOpen", DefaultKeyMap.FileOpen},
		{"BufferSwitch", DefaultKeyMap.BufferSwitch},
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

func TestBuffer_IsModified(t *testing.T) {
	// Test unmodified buffer
	buf := Buffer{
		filePath:        "/path/to/test.go",
		content:         "original content",
		originalContent: "original content",
		readOnly:        false,
		fileSize:        16,
	}
	if buf.IsModified() {
		t.Error("Buffer should not be modified when content equals original")
	}

	// Test modified buffer
	buf.content = "modified content"
	if !buf.IsModified() {
		t.Error("Buffer should be modified when content differs from original")
	}
}

func TestBuildStatusBar_NewFile(t *testing.T) {
	m := initialModel("")
	// Set termWidth for test
	termWidth = 80

	statusBar := m.buildStatusBar()
	if !strings.Contains(statusBar, "New File") {
		t.Error("Status bar should show 'New File' when no buffer is loaded")
	}
}

func TestBuildStatusBar_WithFile(t *testing.T) {
	m := initialModel("")
	// Set termWidth for test
	termWidth = 120

	// Add a buffer
	m.buffers = []Buffer{
		{
			filePath:        "/path/to/test.go",
			content:         "package main",
			originalContent: "package main",
			readOnly:        false,
			fileSize:        100,
		},
	}
	m.currentBuffer = 0

	statusBar := m.buildStatusBar()

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
	m := initialModel("")
	// Set termWidth for test
	termWidth = 120

	// Add a buffer with modification tracking
	m.buffers = []Buffer{
		{
			filePath:        "/path/to/test.go",
			content:         "package modified",
			originalContent: "package main",
			readOnly:        false,
			fileSize:        100,
		},
	}
	m.currentBuffer = 0
	// Set the textarea value to simulate modification
	m.syntaxTA.SetValue("package modified")

	statusBar := m.buildStatusBar()

	// Modified file should have asterisk
	if !strings.Contains(statusBar, "test.go*") {
		t.Error("Status bar should show asterisk for modified file")
	}
}

func TestBuildStatusBar_ReadOnlyFile(t *testing.T) {
	m := initialModel("")
	// Set termWidth for test
	termWidth = 120

	// Add a read-only buffer
	m.buffers = []Buffer{
		{
			filePath:        "/path/to/readonly.txt",
			content:         "read only content",
			originalContent: "read only content",
			readOnly:        true,
			fileSize:        17,
		},
	}
	m.currentBuffer = 0

	statusBar := m.buildStatusBar()

	// Read-only file should have (read-only) indicator
	if !strings.Contains(statusBar, "(read-only)") {
		t.Error("Status bar should show [RO] for read-only file")
	}
}

func TestBuildStatusBar_CursorPosition(t *testing.T) {
	m := initialModel("")
	// Set termWidth for test
	termWidth = 120

	// Add a buffer
	m.buffers = []Buffer{
		{
			filePath:        "/path/to/test.go",
			content:         "package main",
			originalContent: "package main",
			readOnly:        false,
			fileSize:        100,
		},
	}
	m.currentBuffer = 0

	statusBar := m.buildStatusBar()

	// Status bar should contain cursor position in line:col format
	// Default position is 1:1 (first line, first column)
	if !strings.Contains(statusBar, "1:1") {
		t.Error("Status bar should contain cursor position (1:1 for default)")
	}
}
