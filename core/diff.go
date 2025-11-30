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
	// LineDeleted is a marker to track where deletions occurred (not rendered as a line).
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
// compared to the original content using a simple line-by-line diff algorithm.
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

	// Use LCS-based diff algorithm
	return computeDiffStates(d.originalLines, currentLines)
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

// computeDiffStates uses a simplified LCS-based algorithm to compute line states.
func computeDiffStates(original, current []string) []LineState {
	// Build a map of original lines to their indices for quick lookup
	originalMap := make(map[string][]int)
	for i, line := range original {
		originalMap[line] = append(originalMap[line], i)
	}

	states := make([]LineState, len(current))

	// Track which original lines have been "matched"
	matchedOriginal := make([]bool, len(original))

	// First pass: find exact matches in order
	origIdx := 0
	for i, line := range current {
		// Look for a match in remaining original lines
		found := false
		for j := origIdx; j < len(original); j++ {
			if !matchedOriginal[j] && original[j] == line {
				matchedOriginal[j] = true
				states[i] = LineUnchanged
				found = true
				origIdx = j + 1
				break
			}
		}
		if !found {
			// Check if this line exists anywhere in original (possibly moved/duplicated)
			if indices, exists := originalMap[line]; exists {
				for _, idx := range indices {
					if !matchedOriginal[idx] {
						matchedOriginal[idx] = true
						states[i] = LineUnchanged
						found = true
						break
					}
				}
			}
		}
		if !found {
			// Line doesn't exist in original - could be added or modified
			states[i] = LineAdded
		}
	}

	// Second pass: Check for modifications
	// If we have more lines in current than original, extra lines are "added"
	// If lines at same position are different, they're "modified"
	for i := range states {
		if states[i] == LineAdded {
			// Check if this position had a different line in original
			if i < len(original) && !matchedOriginal[i] {
				// There was a line at this position that wasn't matched elsewhere
				states[i] = LineModified
				matchedOriginal[i] = true
			}
		}
	}

	return states
}
