package main

import (
	"os"
	"testing"
)

func TestIsTextFile(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "empty file",
			content:  []byte{},
			expected: true,
		},
		{
			name:     "text file",
			content:  []byte("This is a text file"),
			expected: true,
		},
		{
			name:     "json file",
			content:  []byte(`{"key": "value"}`),
			expected: true,
		},
		{
			name:     "xml file",
			content:  []byte(`<?xml version="1.0"?><root></root>`),
			expected: true,
		},
		{
			name:     "binary file",
			content:  []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTextFile(tt.content)
			if result != tt.expected {
				t.Errorf("isTextFile() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsReadOnly(t *testing.T) {
	// Create a temporary writable file
	tmpFile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Test writable file
	if isReadOnly(tmpFile.Name()) {
		t.Error("Expected writable file to not be read-only")
	}

	// Make file read-only
	err = os.Chmod(tmpFile.Name(), 0444)
	if err != nil {
		t.Fatal(err)
	}

	// Test read-only file
	if !isReadOnly(tmpFile.Name()) {
		t.Error("Expected read-only file to be read-only")
	}

	// Test non-existent file
	if isReadOnly("/nonexistent/file.txt") {
		t.Error("Expected non-existent file to not be read-only")
	}
}

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

	// Test with binary file
	binaryFile, err := os.CreateTemp("", "test_*.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(binaryFile.Name())
	binaryFile.Write([]byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD})
	binaryFile.Close()

	m = initialModel(binaryFile.Name())
	if m.err == nil {
		t.Error("Expected error for binary file")
	}

	// Test with non-existent file
	m = initialModel("/nonexistent/file/path.txt")
	if m.err == nil {
		t.Error("Expected error for non-existent file")
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
	if m.filePath != tmpDir {
		t.Errorf("Expected filePath to be %s, got %s", tmpDir, m.filePath)
	}
}
