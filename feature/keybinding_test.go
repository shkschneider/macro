package feature

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHelpDialogKeyMap(t *testing.T) {
	tests := []struct {
		name    string
		msg     tea.KeyMsg
		binding key.Binding
		want    bool
	}{
		{
			name:    "Esc closes dialog",
			msg:     tea.KeyMsg{Type: tea.KeyEsc},
			binding: DefaultHelpDialogKeyMap.Close,
			want:    true,
		},
		{
			name:    "Ctrl+C closes dialog",
			msg:     tea.KeyMsg{Type: tea.KeyCtrlC},
			binding: DefaultHelpDialogKeyMap.Close,
			want:    true,
		},
		{
			name:    "Up arrow navigates up",
			msg:     tea.KeyMsg{Type: tea.KeyUp},
			binding: DefaultHelpDialogKeyMap.Up,
			want:    true,
		},
		{
			name:    "Down arrow navigates down",
			msg:     tea.KeyMsg{Type: tea.KeyDown},
			binding: DefaultHelpDialogKeyMap.Down,
			want:    true,
		},
		{
			name:    "Enter confirms",
			msg:     tea.KeyMsg{Type: tea.KeyEnter},
			binding: DefaultHelpDialogKeyMap.Enter,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := key.Matches(tt.msg, tt.binding)
			if got != tt.want {
				t.Errorf("key.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileDialogKeyMap(t *testing.T) {
	tests := []struct {
		name    string
		msg     tea.KeyMsg
		binding key.Binding
		want    bool
	}{
		{
			name:    "Esc closes dialog",
			msg:     tea.KeyMsg{Type: tea.KeyEsc},
			binding: DefaultFileDialogKeyMap.Close,
			want:    true,
		},
		{
			name:    "Ctrl+P closes dialog",
			msg:     tea.KeyMsg{Type: tea.KeyCtrlP},
			binding: DefaultFileDialogKeyMap.Close,
			want:    true,
		},
		{
			name:    "Up arrow navigates up",
			msg:     tea.KeyMsg{Type: tea.KeyUp},
			binding: DefaultFileDialogKeyMap.Up,
			want:    true,
		},
		{
			name:    "Down arrow navigates down",
			msg:     tea.KeyMsg{Type: tea.KeyDown},
			binding: DefaultFileDialogKeyMap.Down,
			want:    true,
		},
		{
			name:    "Enter confirms",
			msg:     tea.KeyMsg{Type: tea.KeyEnter},
			binding: DefaultFileDialogKeyMap.Enter,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := key.Matches(tt.msg, tt.binding)
			if got != tt.want {
				t.Errorf("key.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBufferDialogKeyMap(t *testing.T) {
	tests := []struct {
		name    string
		msg     tea.KeyMsg
		binding key.Binding
		want    bool
	}{
		{
			name:    "Esc closes dialog",
			msg:     tea.KeyMsg{Type: tea.KeyEsc},
			binding: DefaultBufferDialogKeyMap.Close,
			want:    true,
		},
		{
			name:    "Ctrl+B closes dialog",
			msg:     tea.KeyMsg{Type: tea.KeyCtrlB},
			binding: DefaultBufferDialogKeyMap.Close,
			want:    true,
		},
		{
			name:    "Up arrow navigates up",
			msg:     tea.KeyMsg{Type: tea.KeyUp},
			binding: DefaultBufferDialogKeyMap.Up,
			want:    true,
		},
		{
			name:    "Down arrow navigates down",
			msg:     tea.KeyMsg{Type: tea.KeyDown},
			binding: DefaultBufferDialogKeyMap.Down,
			want:    true,
		},
		{
			name:    "Enter confirms",
			msg:     tea.KeyMsg{Type: tea.KeyEnter},
			binding: DefaultBufferDialogKeyMap.Enter,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := key.Matches(tt.msg, tt.binding)
			if got != tt.want {
				t.Errorf("key.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAllDialogKeyMapsHaveHelp(t *testing.T) {
	// Test HelpDialogKeyMap
	helpBindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Close", DefaultHelpDialogKeyMap.Close},
		{"Up", DefaultHelpDialogKeyMap.Up},
		{"Down", DefaultHelpDialogKeyMap.Down},
		{"Enter", DefaultHelpDialogKeyMap.Enter},
	}

	for _, b := range helpBindings {
		help := b.binding.Help()
		if help.Key == "" {
			t.Errorf("HelpDialog %s binding should have help key text", b.name)
		}
		if help.Desc == "" {
			t.Errorf("HelpDialog %s binding should have help description", b.name)
		}
	}

	// Test FileDialogKeyMap
	fileBindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Close", DefaultFileDialogKeyMap.Close},
		{"Up", DefaultFileDialogKeyMap.Up},
		{"Down", DefaultFileDialogKeyMap.Down},
		{"Enter", DefaultFileDialogKeyMap.Enter},
	}

	for _, b := range fileBindings {
		help := b.binding.Help()
		if help.Key == "" {
			t.Errorf("FileDialog %s binding should have help key text", b.name)
		}
		if help.Desc == "" {
			t.Errorf("FileDialog %s binding should have help description", b.name)
		}
	}

	// Test BufferDialogKeyMap
	bufferBindings := []struct {
		name    string
		binding key.Binding
	}{
		{"Close", DefaultBufferDialogKeyMap.Close},
		{"Up", DefaultBufferDialogKeyMap.Up},
		{"Down", DefaultBufferDialogKeyMap.Down},
		{"Enter", DefaultBufferDialogKeyMap.Enter},
	}

	for _, b := range bufferBindings {
		help := b.binding.Help()
		if help.Key == "" {
			t.Errorf("BufferDialog %s binding should have help key text", b.name)
		}
		if help.Desc == "" {
			t.Errorf("BufferDialog %s binding should have help description", b.name)
		}
	}
}
