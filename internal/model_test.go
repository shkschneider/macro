package internal

import (
	"testing"
)

func TestInitialModel_Empty(t *testing.T) {
	m := InitialModel("")

	if m.Textarea == nil {
		t.Error("InitialModel should initialize Textarea")
	}
	if m.CurrentBuffer != -1 {
		t.Errorf("InitialModel with empty path should have CurrentBuffer -1, got %d", m.CurrentBuffer)
	}
	if len(m.Buffers) != 0 {
		t.Errorf("InitialModel with empty path should have no buffers, got %d", len(m.Buffers))
	}
	if m.ShowPicker {
		t.Error("InitialModel with empty path should not show picker")
	}
	if m.CursorState == nil {
		t.Error("InitialModel should initialize CursorState")
	}
}

func TestInitialModel_NonExistentFile(t *testing.T) {
	m := InitialModel("/nonexistent/path/to/file.txt")

	if m.Err == nil {
		t.Error("InitialModel should set error for non-existent file")
	}
	if m.CurrentBuffer != -1 {
		t.Error("InitialModel should not load buffer for non-existent file")
	}
}

func TestModel_GetCurrentBuffer_NoBuffer(t *testing.T) {
	m := InitialModel("")

	buf := m.getCurrentBuffer()
	if buf != nil {
		t.Error("getCurrentBuffer should return nil when no buffer is loaded")
	}
}

func TestModel_IsCurrentBufferReadOnly_NoBuffer(t *testing.T) {
	m := InitialModel("")

	if m.isCurrentBufferReadOnly() {
		t.Error("isCurrentBufferReadOnly should return false when no buffer is loaded")
	}
}

func TestModel_IsCurrentBufferModified_NoBuffer(t *testing.T) {
	m := InitialModel("")

	if m.isCurrentBufferModified() {
		t.Error("isCurrentBufferModified should return false when no buffer is loaded")
	}
}

func TestModel_GetCurrentFilePath_NoBuffer(t *testing.T) {
	m := InitialModel("")

	path := m.getCurrentFilePath()
	if path != "" {
		t.Error("getCurrentFilePath should return empty string when no buffer is loaded")
	}
}

func TestModel_GetDirectoryPath_NoBuffer(t *testing.T) {
	m := InitialModel("")

	path := m.getDirectoryPath()
	if path != "" {
		t.Error("getDirectoryPath should return empty string when no buffer is loaded")
	}
}

func TestModel_AddBuffer(t *testing.T) {
	m := InitialModel("")

	// Add first buffer
	idx := m.addBuffer("/path/to/file1.txt", "content1", false, 8)
	if idx != 0 {
		t.Errorf("First buffer should be at index 0, got %d", idx)
	}
	if len(m.Buffers) != 1 {
		t.Errorf("Should have 1 buffer, got %d", len(m.Buffers))
	}

	// Add second buffer
	idx = m.addBuffer("/path/to/file2.txt", "content2", false, 8)
	if idx != 1 {
		t.Errorf("Second buffer should be at index 1, got %d", idx)
	}
	if len(m.Buffers) != 2 {
		t.Errorf("Should have 2 buffers, got %d", len(m.Buffers))
	}

	// Add duplicate buffer (should return existing index)
	idx = m.addBuffer("/path/to/file1.txt", "different content", false, 16)
	if idx != 0 {
		t.Errorf("Duplicate buffer should return existing index 0, got %d", idx)
	}
	if len(m.Buffers) != 2 {
		t.Errorf("Should still have 2 buffers after duplicate add, got %d", len(m.Buffers))
	}
}

func TestModel_EditorContext_SetMessage(t *testing.T) {
	m := InitialModel("")

	m.SetMessage("test message")
	if m.Message != "test message" {
		t.Errorf("SetMessage should set message, got '%s'", m.Message)
	}
}

func TestModel_EditorContext_SetError(t *testing.T) {
	m := InitialModel("")

	testErr := &testError{msg: "test error"}
	m.SetError(testErr)
	if m.Err != testErr {
		t.Error("SetError should set error")
	}

	m.SetError(nil)
	if m.Err != nil {
		t.Error("SetError(nil) should clear error")
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestModel_EditorContext_GetBuffers(t *testing.T) {
	m := InitialModel("")
	m.addBuffer("/path/to/file1.txt", "content1", false, 8)
	m.addBuffer("/path/to/file2.txt", "content2", true, 8)

	buffers := m.GetBuffers()
	if len(buffers) != 2 {
		t.Errorf("GetBuffers should return 2 buffers, got %d", len(buffers))
	}
	if buffers[0].FilePath != "/path/to/file1.txt" {
		t.Error("First buffer should have correct path")
	}
	if buffers[1].ReadOnly != true {
		t.Error("Second buffer should be read-only")
	}
}

func TestModel_EditorContext_GetCurrentBufferIndex(t *testing.T) {
	m := InitialModel("")

	if m.GetCurrentBufferIndex() != -1 {
		t.Error("GetCurrentBufferIndex should return -1 when no buffer is loaded")
	}

	m.addBuffer("/path/to/file.txt", "content", false, 7)
	m.CurrentBuffer = 0
	if m.GetCurrentBufferIndex() != 0 {
		t.Error("GetCurrentBufferIndex should return 0 after setting CurrentBuffer")
	}
}
