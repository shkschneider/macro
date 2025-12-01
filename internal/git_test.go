package internal

import (
	"testing"
)

func TestGitDiffResult_Initialization(t *testing.T) {
	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}

	if result.AddedLines == nil {
		t.Error("AddedLines should be initialized")
	}
	if result.ModifiedLines == nil {
		t.Error("ModifiedLines should be initialized")
	}
	if result.DeletedLines == nil {
		t.Error("DeletedLines should be initialized")
	}
}

func TestParseGitDiffOutput_Empty(t *testing.T) {
	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}

	parseGitDiffOutput("", result)

	if len(result.AddedLines) != 0 {
		t.Error("Empty diff should have no added lines")
	}
	if len(result.ModifiedLines) != 0 {
		t.Error("Empty diff should have no modified lines")
	}
	if len(result.DeletedLines) != 0 {
		t.Error("Empty diff should have no deleted lines")
	}
}

func TestParseGitDiffOutput_AddedLine(t *testing.T) {
	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}

	// Simulated git diff output for an added line
	diffOutput := `@@ -0,0 +1,1 @@
+new line added`

	parseGitDiffOutput(diffOutput, result)

	if !result.AddedLines[1] {
		t.Error("Line 1 should be marked as added")
	}
}

func TestParseGitDiffOutput_DeletedLine(t *testing.T) {
	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}

	// Simulated git diff output for a deleted line
	diffOutput := `@@ -1,1 +1,0 @@
-deleted line`

	parseGitDiffOutput(diffOutput, result)

	if !result.DeletedLines[1] {
		t.Error("Line 1 should be marked as deleted")
	}
}

func TestParseGitDiffOutput_ModifiedLine(t *testing.T) {
	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}

	// Simulated git diff output for a modified line (delete + add at same position)
	diffOutput := `@@ -1,1 +1,1 @@
-old line
+new line`

	parseGitDiffOutput(diffOutput, result)

	// After post-processing, both add and delete at same position become modified
	if !result.ModifiedLines[1] {
		t.Error("Line 1 should be marked as modified")
	}
	if result.AddedLines[1] {
		t.Error("Line 1 should not be in AddedLines after modification detection")
	}
	if result.DeletedLines[1] {
		t.Error("Line 1 should not be in DeletedLines after modification detection")
	}
}

func TestParseGitDiffOutput_MultipleHunks(t *testing.T) {
	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}

	// Simulated git diff with multiple hunks
	diffOutput := `@@ -1,1 +1,1 @@
-old first
+new first
@@ -5,0 +5,1 @@
+added at line 5`

	parseGitDiffOutput(diffOutput, result)

	if !result.ModifiedLines[1] {
		t.Error("Line 1 should be marked as modified")
	}
	if !result.AddedLines[5] {
		t.Error("Line 5 should be marked as added")
	}
}

func TestIsGitTracked_NonExistentFile(t *testing.T) {
	// This should return false for a non-existent file
	result := IsGitTracked("/nonexistent/path/to/file.txt")
	if result {
		t.Error("IsGitTracked should return false for non-existent file")
	}
}

func TestGetGitFileContent_NonExistentFile(t *testing.T) {
	content, ok := GetGitFileContent("/nonexistent/path/to/file.txt")
	if ok {
		t.Error("GetGitFileContent should return false for non-existent file")
	}
	if content != "" {
		t.Error("GetGitFileContent should return empty string for non-existent file")
	}
}
