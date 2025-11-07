package main

import (
	"os"
	"strings"
	"testing"
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

	// Test NewFileDialog
	dialog := NewFileDialog(tmpDir)
	if dialog == nil {
		t.Fatal("Expected NewFileDialog to return a dialog")
	}
	if !dialog.IsVisible() {
		t.Error("Expected dialog to be visible initially")
	}
	if len(dialog.allFiles) != 3 {
		t.Errorf("Expected 3 files, got %d", len(dialog.allFiles))
	}
	if len(dialog.filteredFiles) != 3 {
		t.Errorf("Expected 3 filtered files initially, got %d", len(dialog.filteredFiles))
	}

	// Test fuzzy filtering
	dialog.filterInput.SetValue("go")
	dialog.applyFuzzyFilter()
	if len(dialog.filteredFiles) != 1 {
		t.Errorf("Expected 1 file matching 'go', got %d", len(dialog.filteredFiles))
	}
	if len(dialog.filteredFiles) > 0 && dialog.filteredFiles[0].name != "file3.go" {
		t.Errorf("Expected 'file3.go' to match, got %s", dialog.filteredFiles[0].name)
	}

	// Test clearing filter
	dialog.filterInput.SetValue("")
	dialog.applyFuzzyFilter()
	if len(dialog.filteredFiles) != 3 {
		t.Errorf("Expected 3 files after clearing filter, got %d", len(dialog.filteredFiles))
	}

	// Test with empty directory
	emptyDir, err := os.MkdirTemp("", "test_empty_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(emptyDir)
	
	dialog2 := NewFileDialog(emptyDir)
	if len(dialog2.allFiles) != 0 {
		t.Errorf("Expected 0 files in empty directory, got %d", len(dialog2.allFiles))
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

	// Create buffers to pass to dialog
	buffers := []Buffer{
		{filePath: file1Path, content: "content1", readOnly: false},
		{filePath: file2Path, content: "content2", readOnly: false},
	}

	// Test NewBufferDialog
	dialog := NewBufferDialog(buffers, 0)
	if dialog == nil {
		t.Fatal("Expected NewBufferDialog to return a dialog")
	}
	if !dialog.IsVisible() {
		t.Error("Expected dialog to be visible initially")
	}
	if len(dialog.allBuffers) != 2 {
		t.Errorf("Expected 2 buffers in allBuffers, got %d", len(dialog.allBuffers))
	}
	if len(dialog.filteredBuffers) != 2 {
		t.Errorf("Expected 2 buffers in filteredBuffers initially, got %d", len(dialog.filteredBuffers))
	}

	// Test buffer fuzzy filtering
	dialog.filterInput.SetValue("file1")
	dialog.applyFuzzyFilter()
	if len(dialog.filteredBuffers) != 1 {
		t.Errorf("Expected 1 buffer matching 'file1', got %d", len(dialog.filteredBuffers))
	}

	// Test clearing filter
	dialog.filterInput.SetValue("")
	dialog.applyFuzzyFilter()
	if len(dialog.filteredBuffers) != 2 {
		t.Errorf("Expected 2 buffers after clearing filter, got %d", len(dialog.filteredBuffers))
	}

	// Test with no buffers
	dialog2 := NewBufferDialog([]Buffer{}, -1)
	if len(dialog2.allBuffers) != 0 {
		t.Errorf("Expected 0 buffers, got %d", len(dialog2.allBuffers))
	}
}

func TestHelpDialog(t *testing.T) {
	// Test NewHelpDialog
	dialog := NewHelpDialog()
	if dialog == nil {
		t.Fatal("Expected NewHelpDialog to return a dialog")
	}
	if !dialog.IsVisible() {
		t.Error("Expected dialog to be visible initially")
	}
	if len(dialog.allCommands) == 0 {
		t.Error("Expected commands to be populated")
	}
	if len(dialog.filteredCommands) == 0 {
		t.Error("Expected filtered commands to be populated initially")
	}

	// Test help fuzzy filtering
	dialog.filterInput.SetValue("file")
	dialog.applyFuzzyFilter()
	if len(dialog.filteredCommands) == 0 {
		t.Error("Expected at least one command matching 'file'")
	}
	
	// Verify file-related commands are in results
	foundFileCommand := false
	for _, cmd := range dialog.filteredCommands {
		if strings.Contains(cmd.command.name, "file") {
			foundFileCommand = true
			break
		}
	}
	if !foundFileCommand {
		t.Error("Expected to find file-related commands in filtered results")
	}

	// Test clearing filter
	dialog.filterInput.SetValue("")
	dialog.applyFuzzyFilter()
	if len(dialog.filteredCommands) != len(dialog.allCommands) {
		t.Errorf("Expected %d commands after clearing filter, got %d", len(dialog.allCommands), len(dialog.filteredCommands))
	}
}

func TestCommandSystem(t *testing.T) {
	// Test getCommandByName
	cmd := getCommandByName("file-save")
	if cmd == nil {
		t.Error("Expected to find file-save command")
	}
	if cmd.name != "file-save" {
		t.Errorf("Expected command name to be 'file-save', got %s", cmd.name)
	}

	// Test non-existent command
	cmd = getCommandByName("non-existent")
	if cmd != nil {
		t.Error("Expected nil for non-existent command")
	}

	// Test that all commands have required fields
	for _, cmd := range getKeybindings() {
		if cmd.name == "" {
			t.Error("Command has empty name")
		}
		if cmd.key == "" {
			t.Error("Command has empty key")
		}
		if cmd.description == "" {
			t.Error("Command has empty description")
		}
		// Note: execute can be nil for commands handled directly in Update
	}
}
