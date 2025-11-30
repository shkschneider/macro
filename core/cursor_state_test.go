package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCursorState(t *testing.T) {
	cs := NewCursorState()
	if cs == nil {
		t.Fatal("NewCursorState returned nil")
	}
	if cs.Positions == nil {
		t.Error("Positions map should be initialized")
	}
}

func TestCursorState_SetAndGetPosition(t *testing.T) {
	cs := &CursorState{
		Positions: make(map[string]CursorPosition),
	}

	// Set position for a file
	cs.SetPosition("/tmp/test.go", 10, 5)

	// Get position back
	pos, ok := cs.GetPosition("/tmp/test.go")
	if !ok {
		t.Error("GetPosition should return true for existing file")
	}
	if pos.Line != 10 {
		t.Errorf("Expected line 10, got %d", pos.Line)
	}
	if pos.Column != 5 {
		t.Errorf("Expected column 5, got %d", pos.Column)
	}
}

func TestCursorState_GetPosition_NonExistent(t *testing.T) {
	cs := &CursorState{
		Positions: make(map[string]CursorPosition),
	}

	_, ok := cs.GetPosition("/nonexistent/file.go")
	if ok {
		t.Error("GetPosition should return false for non-existent file")
	}
}

func TestCursorState_SaveAndLoad(t *testing.T) {
	// Create a temp directory for the test
	tmpDir, err := os.MkdirTemp("", "cursor_state_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a cursor state with a custom file path
	cs := &CursorState{
		Positions: make(map[string]CursorPosition),
		filePath:  filepath.Join(tmpDir, "cursor_state.json"),
	}

	// Set some positions
	cs.SetPosition("/test/file1.go", 100, 25)
	cs.SetPosition("/test/file2.py", 50, 10)

	// Save to disk
	err = cs.Save()
	if err != nil {
		t.Fatalf("Failed to save cursor state: %v", err)
	}

	// Create a new cursor state and load from disk
	cs2 := &CursorState{
		Positions: make(map[string]CursorPosition),
		filePath:  filepath.Join(tmpDir, "cursor_state.json"),
	}
	cs2.load()

	// Verify loaded positions
	pos1, ok := cs2.GetPosition("/test/file1.go")
	if !ok {
		t.Error("Should find file1.go after loading")
	}
	if pos1.Line != 100 || pos1.Column != 25 {
		t.Errorf("file1.go: expected (100, 25), got (%d, %d)", pos1.Line, pos1.Column)
	}

	pos2, ok := cs2.GetPosition("/test/file2.py")
	if !ok {
		t.Error("Should find file2.py after loading")
	}
	if pos2.Line != 50 || pos2.Column != 10 {
		t.Errorf("file2.py: expected (50, 10), got (%d, %d)", pos2.Line, pos2.Column)
	}
}

func TestCursorState_LoadNonExistent(t *testing.T) {
	// Create a cursor state with a non-existent file path
	cs := &CursorState{
		Positions: make(map[string]CursorPosition),
		filePath:  "/nonexistent/path/cursor_state.json",
	}

	// Load should not panic and Positions should remain empty but initialized
	cs.load()

	if cs.Positions == nil {
		t.Error("Positions should be initialized after load even if file doesn't exist")
	}
}

func TestCursorState_SaveWithNoFilePath(t *testing.T) {
	cs := &CursorState{
		Positions: make(map[string]CursorPosition),
		filePath:  "",
	}

	// Save should return nil (no error) when no file path is set
	err := cs.Save()
	if err != nil {
		t.Errorf("Save should return nil when filePath is empty, got: %v", err)
	}
}

func TestCursorPosition_ZeroValues(t *testing.T) {
	cs := &CursorState{
		Positions: make(map[string]CursorPosition),
	}

	// Set position at line 0, column 0
	cs.SetPosition("/test/file.go", 0, 0)

	pos, ok := cs.GetPosition("/test/file.go")
	if !ok {
		t.Error("GetPosition should return true for file with zero position")
	}
	if pos.Line != 0 {
		t.Errorf("Expected line 0, got %d", pos.Line)
	}
	if pos.Column != 0 {
		t.Errorf("Expected column 0, got %d", pos.Column)
	}
}
