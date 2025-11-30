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
	tracker.SetOriginal("line1\nline3")
	states := tracker.ComputeLineStates("line1\nline2\nline3")

	// line1: unchanged, line2: added, line3: unchanged
	if len(states) != 3 {
		t.Fatalf("Expected 3 states, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line1 should be unchanged, got %v", states[0])
	}
	if states[1] != LineAdded {
		t.Errorf("line2 should be added, got %v", states[1])
	}
	if states[2] != LineUnchanged {
		t.Errorf("line3 should be unchanged, got %v", states[2])
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
	tracker.SetOriginal("line1\nline2\nline3")
	result := tracker.ComputeDiff("line1\nline3")

	// line1: unchanged, line3: unchanged
	if len(result.LineStates) != 2 {
		t.Fatalf("Expected 2 states, got %d", len(result.LineStates))
	}
	if result.LineStates[0] != LineUnchanged {
		t.Errorf("line1 should be unchanged, got %v", result.LineStates[0])
	}
	if result.LineStates[1] != LineUnchanged {
		t.Errorf("line3 should be unchanged, got %v", result.LineStates[1])
	}

	// Deletion marker should be above line3 (index 1)
	if len(result.DeletionsAbove) != 3 {
		t.Fatalf("Expected 3 deletion markers, got %d", len(result.DeletionsAbove))
	}
	if result.DeletionsAbove[1] != true {
		t.Errorf("Expected deletion marker above line 1 (line3)")
	}
}

func TestDiffTracker_DeletionAtEnd(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline2\nline3")
	result := tracker.ComputeDiff("line1")

	// Only line1 remains
	if len(result.LineStates) != 1 {
		t.Fatalf("Expected 1 state, got %d", len(result.LineStates))
	}
	if result.LineStates[0] != LineUnchanged {
		t.Errorf("line1 should be unchanged, got %v", result.LineStates[0])
	}

	// Deletion markers should be at end (index 1, which is after all current lines)
	if len(result.DeletionsAbove) != 2 {
		t.Fatalf("Expected 2 deletion markers, got %d", len(result.DeletionsAbove))
	}
}
