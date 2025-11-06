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
