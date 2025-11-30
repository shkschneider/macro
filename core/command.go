package core

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// CommandDef defines a command without execution logic
type CommandDef struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
}

// BufferInfo contains information about a buffer for dialogs
type BufferInfo struct {
	FilePath string
	ReadOnly bool
}

// EditorContext provides an interface for features to interact with the editor
// This allows features to define their own execution logic without depending on
// the main package's model type directly.
type EditorContext interface {
	// Buffer operations
	IsCurrentBufferReadOnly() bool
	GetCurrentFilePath() string
	GetCurrentContent() string
	SaveCurrentBufferState()
	UpdateBufferAfterSave(content string, fileSize int64)

	// Message operations
	SetMessage(msg string)
	SetError(err error)
}

// FeatureCommand represents a command with its execution logic defined by a feature
type FeatureCommand struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
	Execute     func(ctx EditorContext) tea.Cmd
}
