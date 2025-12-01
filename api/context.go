package api

import tea "github.com/charmbracelet/bubbletea"

// BufferInfo contains information about a buffer for dialogs
type BufferInfo struct {
	FilePath string
	ReadOnly bool
}

// EditorContext provides an interface for plugins to interact with the editor.
// This allows plugins to define their own execution logic without depending on
// the main package's model type directly.
type EditorContext interface {
	// Buffer operations
	IsCurrentBufferReadOnly() bool
	GetCurrentFilePath() string
	GetCurrentContent() string
	SaveCurrentBufferState()
	UpdateBufferAfterSave(content string, fileSize int64)
	GetBuffers() []BufferInfo
	GetCurrentBufferIndex() int
	HasUnsavedChanges() bool
	IsCurrentBufferModified() bool
	CloseCurrentBuffer() bool // Returns true if this was the last buffer

	// Buffer management
	OpenFile(path string) error
	SwitchToBuffer(index int)

	// Message operations
	SetMessage(msg string)
	SetError(err error)

	// Dialog operations
	SetActiveDialog(dialog Dialog) tea.Cmd

	// Cursor state operations
	SaveCursorState()

	// Command operations
	ExecuteCommand(name string) tea.Cmd

	// File picker operations
	ShowFilePicker(directory string)
}
