package internal

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// CursorPosition represents the cursor position in a file
type CursorPosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// CursorState manages cursor positions for files
type CursorState struct {
	Positions map[string]CursorPosition `json:"positions"`
	filePath  string
}

// getStateFilePath returns the path to the cursor state file
func getStateFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		homeDir, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", homeErr
		}
		configDir = homeDir
	}

	macroDir := filepath.Join(configDir, "macro")
	if err := os.MkdirAll(macroDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(macroDir, "cursor_state.json"), nil
}

// NewCursorState creates a new CursorState and loads existing data
func NewCursorState() *CursorState {
	cs := &CursorState{
		Positions: make(map[string]CursorPosition),
	}

	stateFile, err := getStateFilePath()
	if err != nil {
		return cs
	}
	cs.filePath = stateFile

	cs.load()
	return cs
}

// load reads the cursor state from disk
func (cs *CursorState) load() {
	if cs.filePath == "" {
		return
	}

	data, err := os.ReadFile(cs.filePath)
	if err != nil {
		// File doesn't exist or can't be read - that's fine
		return
	}

	// Parse JSON - ignore errors, just use empty state
	_ = json.Unmarshal(data, cs)
	if cs.Positions == nil {
		cs.Positions = make(map[string]CursorPosition)
	}
}

// Save writes the cursor state to disk
func (cs *CursorState) Save() error {
	if cs.filePath == "" {
		return nil
	}

	data, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cs.filePath, data, 0644)
}

// GetPosition returns the saved cursor position for a file
func (cs *CursorState) GetPosition(filePath string) (CursorPosition, bool) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	pos, ok := cs.Positions[absPath]
	return pos, ok
}

// SetPosition saves the cursor position for a file
func (cs *CursorState) SetPosition(filePath string, line, column int) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}

	cs.Positions[absPath] = CursorPosition{
		Line:   line,
		Column: column,
	}
}
