package vanilla

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	macro "github.com/shkschneider/macro/core"
)

// Command name constants for use in execute handler mapping
const (
	CmdQuit           = "quit"
	CmdSave           = "file-save"
	CmdHelp           = "help-show"
	CmdFileOpen       = "file-open"
	CmdBufferSwitch   = "buffer-switch"
)

// CommandRegistration represents a command to be registered with the main app
type CommandRegistration struct {
	Name        string
	Key         string
	Description string
	KeyBinding  key.Binding
	// FeatureExecute is set for commands that use EditorContext (like save).
	// Commands without FeatureExecute need the main app to provide an execute handler.
	FeatureExecute func(ctx macro.EditorContext) tea.Cmd
}

// Cached command definitions (created once)
var cachedCommands []CommandRegistration

func init() {
	saveCmd := SaveCommand()

	cachedCommands = []CommandRegistration{
		{
			Name:           CmdQuit,
			Key:            QuitCommand().Key,
			Description:    QuitCommand().Description,
			KeyBinding:     QuitCommand().KeyBinding,
			FeatureExecute: nil, // Main app provides execute handler
		},
		{
			Name:           CmdSave,
			Key:            saveCmd.Key,
			Description:    saveCmd.Description,
			KeyBinding:     saveCmd.KeyBinding,
			FeatureExecute: saveCmd.Execute, // Feature provides execute handler
		},
		{
			Name:           CmdHelp,
			Key:            HelpCommand().Key,
			Description:    HelpCommand().Description,
			KeyBinding:     HelpCommand().KeyBinding,
			FeatureExecute: nil, // Main app provides execute handler
		},
		{
			Name:           CmdFileOpen,
			Key:            FileSwitcherCommand().Key,
			Description:    FileSwitcherCommand().Description,
			KeyBinding:     FileSwitcherCommand().KeyBinding,
			FeatureExecute: nil, // Main app provides execute handler
		},
		{
			Name:           CmdBufferSwitch,
			Key:            BufferSwitcherCommand().Key,
			Description:    BufferSwitcherCommand().Description,
			KeyBinding:     BufferSwitcherCommand().KeyBinding,
			FeatureExecute: nil, // Main app provides execute handler
		},
	}
}

// Register calls the provided callback for each feature command.
// This allows features to auto-register with the main app's command registry.
func Register(registerFunc func(cmd CommandRegistration)) {
	for _, cmd := range cachedCommands {
		registerFunc(cmd)
	}
}

// GetCommands returns all feature commands for registration.
// Commands with FeatureExecute set can be used directly with EditorContext.
// Commands without FeatureExecute need the main app to provide an execute handler.
func GetCommands() []CommandRegistration {
	return cachedCommands
}
