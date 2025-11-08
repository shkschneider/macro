package feature

// CommandDef defines a command without execution logic
type CommandDef struct {
	Name        string
	Key         string
	Description string
}

// BufferInfo contains information about a buffer for dialogs
type BufferInfo struct {
	FilePath string
	ReadOnly bool
}
