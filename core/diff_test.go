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

	// No UpdateContent called, so all lines should be unchanged
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

func TestDiffTracker_AddLine(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline2")

	// Simulate adding a line at position 2 (cursor at line 2)
	tracker.UpdateContent("line1\nline2\nline3", 2)

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

func TestDiffTracker_ModifyLine(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\noriginal\nline3")

	// Simulate modifying line at position 1 (cursor at line 1)
	tracker.UpdateContent("line1\nmodified\nline3", 1)

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

func TestDiffTracker_DeleteLine(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline2\nline3")

	// Simulate deleting line at position 1 (cursor at line 1)
	tracker.UpdateContent("line1\nline3", 1)

	states, deletedAt := tracker.ComputeLineStatesWithDeletions("line1\nline3")

	if len(states) != 2 {
		t.Fatalf("Expected 2 states, got %d", len(states))
	}
	if states[0] != LineUnchanged {
		t.Errorf("line 0 should be unchanged, got %v", states[0])
	}

	// Position 1 should have a deletion marker
	if !deletedAt[1] {
		t.Error("Expected deletion marker at position 1")
	}
}

func TestDiffTracker_AddThenDelete(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1\nline2")

	// First add a line at position 2
	tracker.UpdateContent("line1\nline2\nline3", 2)

	// Then delete that line
	tracker.UpdateContent("line1\nline2", 2)

	states := tracker.ComputeLineStates("line1\nline2")

	if len(states) != 2 {
		t.Fatalf("Expected 2 states, got %d", len(states))
	}
	// After adding then deleting, we should have a deletion marker
	// and no more added marker
}

func TestDiffTracker_MultipleAdds(t *testing.T) {
	tracker := NewDiffTracker()
	tracker.SetOriginal("line1")

	// Add line at position 1
	tracker.UpdateContent("line1\nline2", 1)
	// Add another line at position 2
	tracker.UpdateContent("line1\nline2\nline3", 2)

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
	if states[2] != LineAdded {
		t.Errorf("line 2 should be added, got %v", states[2])
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
