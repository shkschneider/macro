package api

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

	// Message operations
	SetMessage(msg string)
	SetError(err error)
}
