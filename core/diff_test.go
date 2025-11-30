package core

import (
	"reflect"
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

func TestDiffTracker_SetOriginal(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline2\nline3")

	if !tracker.HasOriginal() {
		t.Error("HasOriginal() should return true after SetOriginal()")
	}
}

func TestDiffTracker_ClearOriginal(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline2")
	tracker.ClearOriginal()

	if tracker.HasOriginal() {
		t.Error("HasOriginal() should return false after ClearOriginal()")
	}
}

func TestDiffTracker_NoOriginal(t *testing.T) {
	tracker := NewDiffTracker()
	states := tracker.ComputeLineStates("line1\nline2")

	// All lines should be "added" when there's no original
	expected := []LineState{LineAdded, LineAdded}
	if !reflect.DeepEqual(states, expected) {
		t.Errorf("Expected all lines to be Added, got %v", states)
	}
}

func TestDiffTracker_NoChanges(t *testing.T) {
	tracker := NewDiffTracker()
	content := "line1\nline2\nline3"
	tracker.SetOriginal(content)
	states := tracker.ComputeLineStates(content)

	// All lines should be unchanged
	expected := []LineState{LineUnchanged, LineUnchanged, LineUnchanged}
	if !reflect.DeepEqual(states, expected) {
		t.Errorf("Expected all lines unchanged, got %v", states)
	}
}

func TestDiffTracker_AddedLines(t *testing.T) {
	tracker := NewDiffTracker()
	// Original: line1, line2
	// Current:  line1, line2, line3
	// Line 3 is added (new line at end)
	tracker.SetOriginal("line1\nline2")
	states := tracker.ComputeLineStates("line1\nline2\nline3")

	if len(states) != 3 {
		t.Fatalf("Expected 3 states, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line1 should be unchanged, got %v", states[0])
	}
	if states[1] != LineUnchanged {
		t.Errorf("line2 should be unchanged, got %v", states[1])
	}
	if states[2] != LineAdded {
		t.Errorf("line3 should be added, got %v", states[2])
	}
}

func TestDiffTracker_ModifiedLine(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\noriginal\nline3")
	states := tracker.ComputeLineStates("line1\nmodified\nline3")

	// line1: unchanged, modified: modified, line3: unchanged
	if len(states) != 3 {
		t.Fatalf("Expected 3 states, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line1 should be unchanged, got %v", states[0])
	}
	if states[1] != LineModified {
		t.Errorf("middle line should be modified, got %v", states[1])
	}
	if states[2] != LineUnchanged {
		t.Errorf("line3 should be unchanged, got %v", states[2])
	}
}

func TestDiffTracker_AllLinesModified(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("aaa\nbbb\nccc")
	states := tracker.ComputeLineStates("xxx\nyyy\nzzz")

	// All lines should be modified since same count but different content
	expected := []LineState{LineModified, LineModified, LineModified}
	if !reflect.DeepEqual(states, expected) {
		t.Errorf("Expected all lines modified, got %v", states)
	}
}

func TestDiffTracker_EmptyOriginal(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("")
	states := tracker.ComputeLineStates("line1\nline2")

	// All lines should be added
	expected := []LineState{LineAdded, LineAdded}
	if !reflect.DeepEqual(states, expected) {
		t.Errorf("Expected all lines added, got %v", states)
	}
}

func TestDiffTracker_EmptyCurrent(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline2")
	states := tracker.ComputeLineStates("")

	// Should return empty slice
	if len(states) != 0 {
		t.Errorf("Expected empty states for empty content, got %v", states)
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "empty",
			content:  "",
			expected: []string{},
		},
		{
			name:     "single line no newline",
			content:  "hello",
			expected: []string{"hello"},
		},
		{
			name:     "single line with newline",
			content:  "hello\n",
			expected: []string{"hello", ""},
		},
		{
			name:     "multiple lines",
			content:  "line1\nline2\nline3",
			expected: []string{"line1", "line2", "line3"},
		},
		{
			name:     "empty lines",
			content:  "line1\n\nline3",
			expected: []string{"line1", "", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitLines(tt.content)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("splitLines(%q) = %v, want %v", tt.content, result, tt.expected)
			}
		})
	}
}

func TestDiffTracker_DeletedLines(t *testing.T) {
	tracker := NewDiffTracker()
	// Original: line1, line2, line3
	// Current:  line1, line3
	// Line at position 2 was deleted (or modified to line3)
	tracker.SetOriginal("line1\nline2\nline3")
	states, deletedAt := tracker.ComputeLineStatesWithDeletions("line1\nline3")

	// With line-by-line comparison:
	// Position 0: line1 == line1 -> unchanged
	// Position 1: line3 != line2 -> modified
	if len(states) != 2 {
		t.Fatalf("Expected 2 states, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line1 should be unchanged, got %v", states[0])
	}
	if states[1] != LineModified {
		t.Errorf("line at position 1 should be modified (line3 replaced line2), got %v", states[1])
	}

	// Deletion marker at position 2 (original had line3 there)
	if len(deletedAt) < 3 {
		t.Fatalf("Expected at least 3 deletion markers, got %d", len(deletedAt))
	}
	if !deletedAt[2] {
		t.Errorf("Expected deletion marker at position 2")
	}
}

func TestDiffTracker_DeletionAtEnd(t *testing.T) {
	tracker := NewDiffTracker()
	// Original: line1, line2, line3
	// Current:  line1
	// Lines at positions 2 and 3 were deleted
	tracker.SetOriginal("line1\nline2\nline3")
	states, deletedAt := tracker.ComputeLineStatesWithDeletions("line1")

	// Only line1 remains
	if len(states) != 1 {
		t.Fatalf("Expected 1 state, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line1 should be unchanged, got %v", states[0])
	}

	// Deletion markers at positions 1 and 2
	if len(deletedAt) < 3 {
		t.Fatalf("Expected at least 3 deletion markers, got %d", len(deletedAt))
	}
	if !deletedAt[1] {
		t.Errorf("Expected deletion marker at position 1")
	}
	if !deletedAt[2] {
		t.Errorf("Expected deletion marker at position 2")
	}
}
