// Package core provides core functionality for the macro editor.
package core

// LineState represents the state of a line compared to the original file.
type LineState int

const (
	// LineUnchanged indicates the line has not been modified.
	LineUnchanged LineState = iota
	// LineAdded indicates the line is new (not in original).
	LineAdded
	// LineModified indicates the line content has changed.
	LineModified
	// LineDeleted indicates this position had a line that was deleted.
	LineDeleted
)

// DiffTracker tracks changes between original and current content.
type DiffTracker struct {
	originalLines []string
	hasOriginal   bool
}

// NewDiffTracker creates a new DiffTracker with the original content.
func NewDiffTracker() *DiffTracker {
	return &DiffTracker{
		originalLines: nil,
		hasOriginal:   false,
	}
}

// SetOriginal stores the original content for comparison.
func (d *DiffTracker) SetOriginal(content string) {
	d.originalLines = splitLines(content)
	d.hasOriginal = true
}

// ClearOriginal clears the original content (e.g., for new files).
func (d *DiffTracker) ClearOriginal() {
	d.originalLines = nil
	d.hasOriginal = false
}

// HasOriginal returns true if original content has been set.
func (d *DiffTracker) HasOriginal() bool {
	return d.hasOriginal
}

// ComputeLineStates computes the state of each line in the current content
// compared to the original content using a line-by-line comparison.
// Returns states for each current line position.
func (d *DiffTracker) ComputeLineStates(currentContent string) []LineState {
	currentLines := splitLines(currentContent)

	// If no original content, all lines are considered "added" (new file)
	if !d.hasOriginal {
		states := make([]LineState, len(currentLines))
		for i := range states {
			states[i] = LineAdded
		}
		return states
	}

	return computeLineByLineDiff(d.originalLines, currentLines)
}

// splitLines splits content into lines, preserving empty lines.
func splitLines(content string) []string {
	if content == "" {
		return []string{}
	}

	var lines []string
	start := 0
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, content[start:i])
			start = i + 1
		}
	}
	// Add the last line (might not end with \n)
	if start <= len(content) {
		lines = append(lines, content[start:])
	}
	return lines
}

// computeLineByLineDiff compares original and current line by line.
// For each current line position:
// - If the line content matches the original at that position: Unchanged
// - If the line content differs from original at that position: Modified
// - If this is a new position beyond original length: Added
func computeLineByLineDiff(original, current []string) []LineState {
	states := make([]LineState, len(current))

	// Compare line by line
	for i := 0; i < len(current); i++ {
		if i < len(original) {
			// Both original and current have this line
			if current[i] == original[i] {
				states[i] = LineUnchanged
			} else {
				states[i] = LineModified
			}
		} else {
			// Current has more lines than original - these are added
			states[i] = LineAdded
		}
	}

	return states
}

// ComputeLineStatesWithDeletions returns line states plus deletion markers.
// deletedAt[i] is true if there was a deletion at position i (even if i >= len(current))
func (d *DiffTracker) ComputeLineStatesWithDeletions(currentContent string) ([]LineState, []bool) {
	currentLines := splitLines(currentContent)

	// If no original content, all lines are considered "added" (new file)
	if !d.hasOriginal {
		states := make([]LineState, len(currentLines))
		for i := range states {
			states[i] = LineAdded
		}
		return states, make([]bool, len(currentLines)+1)
	}

	return computeLineByLineDiffWithDeletions(d.originalLines, currentLines)
}

// computeLineByLineDiffWithDeletions returns both line states and deletion markers.
func computeLineByLineDiffWithDeletions(original, current []string) ([]LineState, []bool) {
	states := make([]LineState, len(current))
	// deletedAt[i] means there's a deleted line that was at position i in original
	maxPos := len(current)
	if len(original) > maxPos {
		maxPos = len(original)
	}
	deletedAt := make([]bool, maxPos+1)

	// Compare line by line
	for i := 0; i < len(current); i++ {
		if i < len(original) {
			// Both original and current have this line
			if current[i] == original[i] {
				states[i] = LineUnchanged
			} else {
				states[i] = LineModified
			}
		} else {
			// Current has more lines than original - these are added
			states[i] = LineAdded
		}
	}

	// Mark positions where original had lines but current doesn't
	for i := len(current); i < len(original); i++ {
		deletedAt[i] = true
	}

	return states, deletedAt
}
