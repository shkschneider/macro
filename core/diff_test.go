package core

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

func TestDiffTracker_NoChanges(t *testing.T) {
	tracker := NewDiffTracker()
	content := "line1\nline2\nline3"
	tracker.SetOriginal(content)

	states := tracker.ComputeLineStates(content)

	if len(states) != 3 {
		t.Fatalf("Expected 3 states, got %d", len(states))
	}
	for i, state := range states {
		if state != LineUnchanged {
			t.Errorf("line %d should be unchanged, got %v", i, state)
		}
	}
}

func TestDiffTracker_AddedLine(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline2")

	// Add a line at the end
	states := tracker.ComputeLineStates("line1\nline2\nline3")

	if len(states) != 3 {
		t.Fatalf("Expected 3 states, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line 0 should be unchanged, got %v", states[0])
	}
	if states[1] != LineUnchanged {
		t.Errorf("line 1 should be unchanged, got %v", states[1])
	}
	if states[2] != LineAdded {
		t.Errorf("line 2 should be added, got %v", states[2])
	}
}

func TestDiffTracker_ModifiedLine(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\noriginal\nline3")

	// Modify the middle line
	states := tracker.ComputeLineStates("line1\nmodified\nline3")

	if len(states) != 3 {
		t.Fatalf("Expected 3 states, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line 0 should be unchanged, got %v", states[0])
	}
	if states[1] != LineModified {
		t.Errorf("line 1 should be modified, got %v", states[1])
	}
	if states[2] != LineUnchanged {
		t.Errorf("line 2 should be unchanged, got %v", states[2])
	}
}

func TestDiffTracker_InsertedLine(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline3")

	// Insert a line in the middle
	states := tracker.ComputeLineStates("line1\nline2\nline3")

	if len(states) != 3 {
		t.Fatalf("Expected 3 states, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line 0 should be unchanged, got %v", states[0])
	}
	if states[1] != LineAdded {
		t.Errorf("line 1 should be added, got %v", states[1])
	}
	if states[2] != LineUnchanged {
		t.Errorf("line 2 should be unchanged, got %v", states[2])
	}
}

func TestDiffTracker_DeletedLine(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline2\nline3")

	// Delete the middle line
	states, deletedAt := tracker.ComputeLineStatesWithDeletions("line1\nline3")

	if len(states) != 2 {
		t.Fatalf("Expected 2 states, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line 0 should be unchanged, got %v", states[0])
	}
	if states[1] != LineUnchanged {
		t.Errorf("line 1 should be unchanged, got %v", states[1])
	}

	// Check for deletion marker
	hasDeleted := false
	for _, d := range deletedAt {
		if d {
			hasDeleted = true
			break
		}
	}
	if !hasDeleted {
		t.Error("Expected a deletion marker somewhere")
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

func TestComputeLCS(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected int
	}{
		{
			name:     "identical",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: 3,
		},
		{
			name:     "one added",
			a:        []string{"a", "c"},
			b:        []string{"a", "b", "c"},
			expected: 2,
		},
		{
			name:     "one removed",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "c"},
			expected: 2,
		},
		{
			name:     "empty",
			a:        []string{},
			b:        []string{"a"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeLCS(tt.a, tt.b)
			if len(result) != tt.expected {
				t.Errorf("computeLCS() returned %d items, want %d", len(result), tt.expected)
			}
		})
	}
}
