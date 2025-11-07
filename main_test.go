package main

import (
	"os"
	"testing"
)

func TestInitialModel(t *testing.T) {
	// Test with no file
	m := initialModel("")
	if m.filePath != "" {
		t.Errorf("Expected empty filePath, got %s", m.filePath)
	}
	if m.err != nil {
		t.Errorf("Expected no error, got %v", m.err)
	}
	if m.readOnly {
		t.Error("Expected readOnly to be false")
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
	if m.readOnly {
		t.Error("Expected writable file to not be read-only")
	}

	// Test with read-only file
	os.Chmod(tmpFile.Name(), 0444)
	m = initialModel(tmpFile.Name())
	if !m.readOnly {
		t.Error("Expected read-only file to set readOnly flag")
	}
	if !m.isWarning {
		t.Error("Expected isWarning to be true for read-only file")
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
