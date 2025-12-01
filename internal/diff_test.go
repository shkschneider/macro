package internal

import (
	"testing"
)

func TestNewDiffTracker(t *testing.T) {
	tracker := NewDiffTracker()
	if tracker == nil {
		t.Fatal("NewDiffTracker() returned nil")
	}
	if tracker.HasOriginal() {
		t.Error("New tracker should not have original content")
	}
}

func TestDiffTracker_ClearOriginal(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.ClearOriginal()

	if tracker.HasOriginal() {
		t.Error("HasOriginal() should return false after ClearOriginal()")
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "empty",
			content:  "",
			expected: 0,
		},
		{
			name:     "single line no newline",
			content:  "hello",
			expected: 1,
		},
		{
			name:     "single line with newline",
			content:  "hello\n",
			expected: 2,
		},
		{
			name:     "multiple lines",
			content:  "line1\nline2\nline3",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.content)
			if len(result) != tt.expected {
				t.Errorf("splitLines(%q) returned %d lines, want %d", tt.content, len(result), tt.expected)
			}
		})
	}
}

func TestParseGitDiffOutput(t *testing.T) {
	// Test parsing of git diff output
	diffOutput := `diff --git a/test.txt b/test.txt
--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,4 @@
 line1
+added line
 line2
 line3
`
	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}
	parseGitDiffOutput(diffOutput, result)

	if !result.AddedLines[2] {
		t.Error("Expected line 2 to be marked as added")
	}
}

func TestParseGitDiffOutput_Deletion(t *testing.T) {
	// Test parsing of git diff output with deletion
	diffOutput := `diff --git a/test.txt b/test.txt
--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,2 @@
 line1
-deleted line
 line3
`
	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}
	parseGitDiffOutput(diffOutput, result)

	if !result.DeletedLines[2] {
		t.Error("Expected deletion marker at position 2")
	}
}

func TestParseGitDiffOutput_Modification(t *testing.T) {
	// Test parsing of git diff output with modification (delete + add at same position)
	diffOutput := `diff --git a/test.txt b/test.txt
--- a/test.txt
+++ b/test.txt
@@ -1,3 +1,3 @@
 line1
-old line
+new line
 line3
`
	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}
	parseGitDiffOutput(diffOutput, result)

	if !result.ModifiedLines[2] {
		t.Error("Expected line 2 to be marked as modified")
	}
	if result.AddedLines[2] {
		t.Error("Line 2 should not be marked as added (it's modified)")
	}
	if result.DeletedLines[2] {
		t.Error("Line 2 should not be marked as deleted (it's modified)")
	}
}
