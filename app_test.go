package main

import (
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
