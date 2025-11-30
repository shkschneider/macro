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

// DiffResult contains both per-line states and inter-line deletion markers.
type DiffResult struct {
	// LineStates contains the state for each current line (Added, Modified, Unchanged)
	LineStates []LineState
	// DeletionsAbove[i] is true if there are deleted lines above current line i
	// DeletionsAbove[len(lines)] is true if there are deleted lines at the end
	DeletionsAbove []bool
}

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
	result := d.ComputeDiff(currentContent)
	return result.LineStates
}

// ComputeDiff computes both line states and deletion markers.
func (d *DiffTracker) ComputeDiff(currentContent string) DiffResult {
	currentLines := splitLines(currentContent)

	// If no original content, all lines are considered "added" (new file)
	if !d.hasOriginal {
		states := make([]LineState, len(currentLines))
		for i := range states {
			states[i] = LineAdded
		}
		return DiffResult{
			LineStates:     states,
			DeletionsAbove: make([]bool, len(currentLines)+1),
		}
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
func computeDiffStates(original, current []string) DiffResult {
	// Build a map of original lines to their indices for quick lookup
	originalMap := make(map[string][]int)
	for i, line := range original {
		originalMap[line] = append(originalMap[line], i)
	}

	states := make([]LineState, len(current))
	deletionsAbove := make([]bool, len(current)+1)

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

	// Third pass: Track deletions - find unmatched original lines
	// For each unmatched original line, mark deletion above the corresponding current position
	for i, matched := range matchedOriginal {
		if !matched {
			// This original line was deleted
			// Find where it would appear in the current content
			// (above the next matched line, or at the end)
			insertPos := len(current) // default to end
			for j := i + 1; j < len(original); j++ {
				if matchedOriginal[j] {
					// Find this matched line's position in current
					for k, currLine := range current {
						if currLine == original[j] && states[k] == LineUnchanged {
							insertPos = k
							break
						}
					}
					break
				}
			}
			// Mark deletion above this position
			if insertPos <= len(current) {
				deletionsAbove[insertPos] = true
			}
		}
	}

	return DiffResult{
		LineStates:     states,
		DeletionsAbove: deletionsAbove,
	}
}
