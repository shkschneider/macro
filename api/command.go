package api

import (
	"github.com/charmbracelet/bubbles/key"
)

// CommandDef defines a command without execution logic
type CommandDef struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
}
