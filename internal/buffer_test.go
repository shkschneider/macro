package internal

import (
	"testing"
)

func TestBuffer_IsModified_Unmodified(t *testing.T) {
	buf := Buffer{
		FilePath:        "/path/to/file.go",
		Content:         "original content",
		OriginalContent: "original content",
		ReadOnly:        false,
		FileSize:        16,
	}

	if buf.IsModified() {
		t.Error("Buffer should not be modified when content equals original")
	}
}

func TestBuffer_IsModified_Modified(t *testing.T) {
	buf := Buffer{
		FilePath:        "/path/to/file.go",
		Content:         "modified content",
		OriginalContent: "original content",
		ReadOnly:        false,
		FileSize:        16,
	}

	if !buf.IsModified() {
		t.Error("Buffer should be modified when content differs from original")
	}
}

func TestBuffer_IsModified_Empty(t *testing.T) {
	buf := Buffer{
		FilePath:        "/path/to/file.go",
		Content:         "",
		OriginalContent: "",
		ReadOnly:        false,
		FileSize:        0,
	}

	if buf.IsModified() {
		t.Error("Empty buffer should not be modified when original is also empty")
	}
}

func TestBuffer_IsModified_EmptyVsContent(t *testing.T) {
	buf := Buffer{
		FilePath:        "/path/to/file.go",
		Content:         "new content",
		OriginalContent: "",
		ReadOnly:        false,
		FileSize:        0,
	}

	if !buf.IsModified() {
		t.Error("Buffer with content should be modified when original is empty")
	}
}

func TestBuffer_Fields(t *testing.T) {
	buf := Buffer{
		FilePath:        "/path/to/test.go",
		Content:         "package main",
		OriginalContent: "package main",
		FileSize:        12,
		ReadOnly:        true,
		CursorLine:      5,
		CursorCol:       10,
	}

	if buf.FilePath != "/path/to/test.go" {
		t.Error("FilePath should be set correctly")
	}
	if buf.Content != "package main" {
		t.Error("Content should be set correctly")
	}
	if buf.OriginalContent != "package main" {
		t.Error("OriginalContent should be set correctly")
	}
	if buf.FileSize != 12 {
		t.Error("FileSize should be set correctly")
	}
	if !buf.ReadOnly {
		t.Error("ReadOnly should be true")
	}
	if buf.CursorLine != 5 {
		t.Error("CursorLine should be set correctly")
	}
	if buf.CursorCol != 10 {
		t.Error("CursorCol should be set correctly")
	}
}
