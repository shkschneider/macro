package main

import (
	"os"
	"testing"

	"github.com/shkschneider/macro/core"
	"github.com/shkschneider/macro/feature"
)

func TestInitialModel(t *testing.T) {
	// Test with no file
	m := initialModel("")
	if m.getCurrentFilePath() != "" {
		t.Errorf("Expected empty filePath, got %s", m.getCurrentFilePath())
	}
	if m.err != nil {
		t.Errorf("Expected no error, got %v", m.err)
	}
	if len(m.buffers) != 0 {
		t.Error("Expected no buffers initially")
	}

	// Test with text file
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("Test content")
	tmpFile.Close()

	m = initialModel(tmpFile.Name())
	if m.err != nil {
		t.Errorf("Expected no error for text file, got %v", m.err)
	}
	if len(m.buffers) != 1 {
		t.Errorf("Expected 1 buffer, got %d", len(m.buffers))
	}
	if m.isCurrentBufferReadOnly() {
		t.Error("Expected writable file to not be read-only")
	}

	// Test with read-only file
	os.Chmod(tmpFile.Name(), 0444)
	m = initialModel(tmpFile.Name())
	if !m.isCurrentBufferReadOnly() {
		t.Error("Expected read-only file to set readOnly flag")
	}

	// Test with directory
	tmpDir, err := os.MkdirTemp("", "test_dir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m = initialModel(tmpDir)
	if m.err != nil {
		t.Errorf("Expected no error for directory, got %v", m.err)
	}
	if !m.showPicker {
		t.Error("Expected showPicker to be true for directory")
	}
}

func TestCursorPositionAtTop(t *testing.T) {
	// Test that cursor is at top (line 0) when opening a writable file
	tmpFile, err := os.CreateTemp("", "test_cursor_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	// Write multi-line content
	content := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5"
	tmpFile.WriteString(content)
	tmpFile.Close()

	m := initialModel(tmpFile.Name())
	if m.err != nil {
		t.Errorf("Expected no error for text file, got %v", m.err)
	}

	// Check cursor is at line 0 (top)
	if m.textarea.Line() != 0 {
		t.Errorf("Expected cursor at line 0, got line %d", m.textarea.Line())
	}

	// Test that viewport is at top for read-only file
	os.Chmod(tmpFile.Name(), 0444)
	m = initialModel(tmpFile.Name())
	if m.err != nil {
		t.Errorf("Expected no error for read-only file, got %v", m.err)
	}

	// Check viewport is at top (YOffset should be 0)
	if m.viewport.YOffset != 0 {
		t.Errorf("Expected viewport YOffset at 0, got %d", m.viewport.YOffset)
	}
}

func TestFileDialog(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir, err := os.MkdirTemp("", "test_dialog_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files in the directory
	testFiles := []string{"file1.txt", "file2.txt", "file3.go"}
	for _, name := range testFiles {
		filePath := tmpDir + "/" + name
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test feature.NewFileDialog
	dialog := feature.NewFileDialog(tmpDir)
	if dialog == nil {
		t.Fatal("Expected feature.NewFileDialog to return a dialog")
	}
	if !dialog.IsVisible() {
		t.Error("Expected dialog to be visible initially")
	}

	// Test with empty directory
	emptyDir, err := os.MkdirTemp("", "test_empty_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(emptyDir)

	dialog2 := feature.NewFileDialog(emptyDir)
	if dialog2 == nil {
		t.Fatal("Expected feature.NewFileDialog to return a dialog even for empty directory")
	}
}

func TestBufferManagement(t *testing.T) {
	// Create temporary files
	tmpDir, err := os.MkdirTemp("", "test_buffers_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	file1Path := tmpDir + "/file1.txt"
	file2Path := tmpDir + "/file2.txt"

	err = os.WriteFile(file1Path, []byte("content1"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(file2Path, []byte("content2"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize with first file
	m := initialModel(file1Path)
	if len(m.buffers) != 1 {
		t.Errorf("Expected 1 buffer, got %d", len(m.buffers))
	}
	if m.currentBuffer != 0 {
		t.Errorf("Expected currentBuffer to be 0, got %d", m.currentBuffer)
	}

	// Add second buffer
	idx := m.addBuffer(file2Path, "content2", false)
	if idx != 1 {
		t.Errorf("Expected new buffer index to be 1, got %d", idx)
	}
	if len(m.buffers) != 2 {
		t.Errorf("Expected 2 buffers, got %d", len(m.buffers))
	}

	// Try to add same file again (should return existing buffer index)
	idx2 := m.addBuffer(file1Path, "content1", false)
	if idx2 != 0 {
		t.Errorf("Expected existing buffer index to be 0, got %d", idx2)
	}
	if len(m.buffers) != 2 {
		t.Errorf("Expected still 2 buffers, got %d", len(m.buffers))
	}

	// Test switching buffers
	m.loadBuffer(1)
	if m.currentBuffer != 1 {
		t.Errorf("Expected currentBuffer to be 1, got %d", m.currentBuffer)
	}
	if m.getCurrentFilePath() != file2Path {
		t.Errorf("Expected current file path to be %s, got %s", file2Path, m.getCurrentFilePath())
	}

	// Test saveCurrentBufferState
	m.textarea.SetValue("modified content2")
	m.saveCurrentBufferState()
	if m.buffers[1].content != "modified content2" {
		t.Errorf("Expected buffer content to be updated, got %s", m.buffers[1].content)
	}
}

func TestBufferDialog(t *testing.T) {
	// Create temporary files
	tmpDir, err := os.MkdirTemp("", "test_buffer_dialog_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	file1Path := tmpDir + "/file1.txt"
	file2Path := tmpDir + "/file2.txt"

	err = os.WriteFile(file1Path, []byte("content1"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(file2Path, []byte("content2"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Convert to BufferInfo for dialog
	bufferInfos := []core.BufferInfo{
		{FilePath: file1Path, ReadOnly: false},
		{FilePath: file2Path, ReadOnly: false},
	}

	// Test feature.NewBufferDialog
	dialog := feature.NewBufferDialog(bufferInfos, 0)
	if dialog == nil {
		t.Fatal("Expected feature.NewBufferDialog to return a dialog")
	}
	if !dialog.IsVisible() {
		t.Error("Expected dialog to be visible initially")
	}

	// Test with no buffers
	dialog2 := feature.NewBufferDialog([]core.BufferInfo{}, -1)
	if dialog2 == nil {
		t.Fatal("Expected feature.NewBufferDialog to return a dialog even with no buffers")
	}
}

func TestHelpDialog(t *testing.T) {
	// Create some test commands
	testCommands := []core.CommandDef{
		{Name: "file-save", Key: "Ctrl-S", Description: "Save file"},
		{Name: "quit", Key: "Ctrl-Q", Description: "Quit editor"},
	}

	// Test feature.NewHelpDialog
	dialog := feature.NewHelpDialog(testCommands)
	if dialog == nil {
		t.Fatal("Expected feature.NewHelpDialog to return a dialog")
	}
	if !dialog.IsVisible() {
		t.Error("Expected dialog to be visible initially")
	}
}

func TestCommandSystem(t *testing.T) {
	// Register commands for testing
	registerCommand(Command{
		Name:        "file-save",
		Key:         "Ctrl-S",
		Description: "Save file",
		Execute:     nil,
	})

	// Test getCommandByName
	cmd := getCommandByName("file-save")
	if cmd == nil {
		t.Error("Expected to find file-save command")
	}
	if cmd.Name != "file-save" {
		t.Errorf("Expected command name to be 'file-save', got %s", cmd.Name)
	}

	// Test non-existent command
	cmd = getCommandByName("non-existent")
	if cmd != nil {
		t.Error("Expected nil for non-existent command")
	}

	// Test that all commands have required fields
	for _, cmd := range getKeybindings() {
		if cmd.Name == "" {
			t.Error("Command has empty name")
		}
		if cmd.Key == "" {
			t.Error("Command has empty key")
		}
		if cmd.Description == "" {
			t.Error("Command has empty description")
		}
		// Note: execute can be nil for commands handled directly in Update
	}
}
