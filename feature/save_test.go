package feature

import (
	"os"
	"path/filepath"
	"testing"

	macro "github.com/shkschneider/macro/core"
)

// mockEditorContext implements core.EditorContext for testing
type mockEditorContext struct {
	readOnly    bool
	filePath    string
	content     string
	message     string
	err         error
	savedState  bool
	updatedContent string
	updatedSize    int64
}

func (m *mockEditorContext) IsCurrentBufferReadOnly() bool {
	return m.readOnly
}

func (m *mockEditorContext) GetCurrentFilePath() string {
	return m.filePath
}

func (m *mockEditorContext) GetCurrentContent() string {
	return m.content
}

func (m *mockEditorContext) SaveCurrentBufferState() {
	m.savedState = true
}

func (m *mockEditorContext) UpdateBufferAfterSave(content string, fileSize int64) {
	m.updatedContent = content
	m.updatedSize = fileSize
}

func (m *mockEditorContext) SetMessage(msg string) {
	m.message = msg
}

func (m *mockEditorContext) SetError(err error) {
	m.err = err
}

// Verify mockEditorContext implements EditorContext
var _ macro.EditorContext = (*mockEditorContext)(nil)

func TestSaveCommand_ReturnsFeatureCommand(t *testing.T) {
	cmd := SaveCommand()

	if cmd.Name != "file-save" {
		t.Errorf("expected Name to be 'file-save', got '%s'", cmd.Name)
	}
	if cmd.Key != "Ctrl-S" {
		t.Errorf("expected Key to be 'Ctrl-S', got '%s'", cmd.Key)
	}
	if cmd.Description == "" {
		t.Error("expected Description to be non-empty")
	}
	if cmd.Execute == nil {
		t.Error("expected Execute to be non-nil")
	}
}

func TestExecuteSave_ReadOnlyBuffer(t *testing.T) {
	ctx := &mockEditorContext{
		readOnly: true,
		filePath: "/some/file.txt",
		content:  "test content",
	}

	cmd := SaveCommand()
	result := cmd.Execute(ctx)

	if result != nil {
		t.Error("expected nil command for read-only buffer")
	}
	if ctx.message != "WARNING: Cannot save - file is read-only" {
		t.Errorf("unexpected message: %s", ctx.message)
	}
	if ctx.savedState {
		t.Error("should not save state for read-only buffer")
	}
}

func TestExecuteSave_NoFilePath(t *testing.T) {
	ctx := &mockEditorContext{
		readOnly: false,
		filePath: "",
		content:  "test content",
	}

	cmd := SaveCommand()
	result := cmd.Execute(ctx)

	if result != nil {
		t.Error("expected nil command when no file path")
	}
	if ctx.err == nil {
		t.Error("expected error to be set")
	}
	if ctx.message != "Error: No filename specified. Usage: macro <filename>" {
		t.Errorf("unexpected message: %s", ctx.message)
	}
}

func TestExecuteSave_SuccessfulSave(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	ctx := &mockEditorContext{
		readOnly: false,
		filePath: tmpFile,
		content:  "test content to save",
	}

	cmd := SaveCommand()
	result := cmd.Execute(ctx)

	if result != nil {
		t.Error("expected nil command for successful save")
	}
	if ctx.err != nil {
		t.Errorf("unexpected error: %v", ctx.err)
	}
	if !ctx.savedState {
		t.Error("expected state to be saved")
	}
	if ctx.updatedContent != "test content to save" {
		t.Errorf("expected updated content to match, got '%s'", ctx.updatedContent)
	}
	if ctx.updatedSize == 0 {
		t.Error("expected updated size to be non-zero")
	}

	// Verify file was actually written
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Errorf("failed to read saved file: %v", err)
	}
	if string(content) != "test content to save" {
		t.Errorf("file content mismatch: got '%s'", string(content))
	}
}

func TestExecuteSave_WriteError(t *testing.T) {
	// Use a path that cannot be written to
	ctx := &mockEditorContext{
		readOnly: false,
		filePath: "/nonexistent/directory/file.txt",
		content:  "test content",
	}

	cmd := SaveCommand()
	result := cmd.Execute(ctx)

	if result != nil {
		t.Error("expected nil command for write error")
	}
	if ctx.err == nil {
		t.Error("expected error to be set")
	}
	if ctx.savedState != true {
		t.Error("state should still be saved before write attempt")
	}
}
