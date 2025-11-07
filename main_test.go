package main

import (
	"os"
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

	// Open one of the files
	testFile := tmpDir + "/file1.txt"
	m := initialModel(testFile)
	if m.err != nil {
		t.Errorf("Expected no error, got %v", m.err)
	}

	// Test getFilesInDirectory
	files := m.getFilesInDirectory()
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}

	// Test opening dialog
	m.openFileDialog()
	if !m.showDialog {
		t.Error("Expected showDialog to be true after openFileDialog")
	}
	if len(m.allFiles) != 3 {
		t.Errorf("Expected 3 files in allFiles, got %d", len(m.allFiles))
	}
	if len(m.filteredFiles) != 3 {
		t.Errorf("Expected 3 files in filteredFiles initially, got %d", len(m.filteredFiles))
	}

	// Test fuzzy filtering
	m.filterInput.SetValue("go")
	m.applyFuzzyFilter()
	if len(m.filteredFiles) != 1 {
		t.Errorf("Expected 1 file matching 'go', got %d", len(m.filteredFiles))
	}
	if len(m.filteredFiles) > 0 && m.filteredFiles[0].name != "file3.go" {
		t.Errorf("Expected 'file3.go' to match, got %s", m.filteredFiles[0].name)
	}

	// Test clearing filter
	m.filterInput.SetValue("")
	m.applyFuzzyFilter()
	if len(m.filteredFiles) != 3 {
		t.Errorf("Expected 3 files after clearing filter, got %d", len(m.filteredFiles))
	}

	// Test closing dialog
	m.closeFileDialog()
	if m.showDialog {
		t.Error("Expected showDialog to be false after closeFileDialog")
	}

	// Test with no file path
	m2 := initialModel("")
	files2 := m2.getFilesInDirectory()
	if len(files2) != 0 {
		t.Errorf("Expected 0 files when no file path, got %d", len(files2))
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

	// Initialize with first file
	m := initialModel(file1Path)
	
	// Add second buffer
	m.addBuffer(file2Path, "content2", false)

	// Test opening buffer dialog
	m.openBufferDialog()
	if !m.showBufferDialog {
		t.Error("Expected showBufferDialog to be true")
	}
	if len(m.allBuffers) != 2 {
		t.Errorf("Expected 2 buffers in allBuffers, got %d", len(m.allBuffers))
	}
	if len(m.filteredBuffers) != 2 {
		t.Errorf("Expected 2 buffers in filteredBuffers initially, got %d", len(m.filteredBuffers))
	}

	// Test buffer fuzzy filtering
	m.bufferFilterInput.SetValue("file1")
	m.applyBufferFuzzyFilter()
	if len(m.filteredBuffers) != 1 {
		t.Errorf("Expected 1 buffer matching 'file1', got %d", len(m.filteredBuffers))
	}

	// Test clearing filter
	m.bufferFilterInput.SetValue("")
	m.applyBufferFuzzyFilter()
	if len(m.filteredBuffers) != 2 {
		t.Errorf("Expected 2 buffers after clearing filter, got %d", len(m.filteredBuffers))
	}

	// Test closing buffer dialog
	m.closeBufferDialog()
	if m.showBufferDialog {
		t.Error("Expected showBufferDialog to be false after closing")
	}

	// Test with no buffers
	m2 := initialModel("")
	m2.openBufferDialog()
	if m2.showBufferDialog {
		t.Error("Expected showBufferDialog to remain false when no buffers")
	}
}
