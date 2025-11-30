// Package core provides core functionality for the macro editor.
package core

import "strings"

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
// It tracks specific line operations (add, delete, modify) rather than
// comparing entire content line-by-line.
type DiffTracker struct {
	originalLines []string
	hasOriginal   bool

	// Track specific changes by line index
	addedLines    map[int]bool // Lines that were added (green)
	modifiedLines map[int]bool // Lines that were modified (yellow)
	deletedLines  map[int]bool // Line positions where content was deleted (red)

	// Previous state for detecting changes
	prevLineCount int
	prevContent   string
}

// NewDiffTracker creates a new DiffTracker with the original content.
func NewDiffTracker() *DiffTracker {
	return &DiffTracker{
		originalLines: nil,
		hasOriginal:   false,
		addedLines:    make(map[int]bool),
		modifiedLines: make(map[int]bool),
		deletedLines:  make(map[int]bool),
		prevLineCount: 0,
		prevContent:   "",
	}
}

// SetOriginal stores the original content for comparison.
func (d *DiffTracker) SetOriginal(content string) {
	d.originalLines = splitLines(content)
	d.hasOriginal = true
	// Reset tracked changes
	d.addedLines = make(map[int]bool)
	d.modifiedLines = make(map[int]bool)
	d.deletedLines = make(map[int]bool)
	d.prevLineCount = len(d.originalLines)
	d.prevContent = content
}

// ClearOriginal clears the original content (e.g., for new files).
func (d *DiffTracker) ClearOriginal() {
	d.originalLines = nil
	d.hasOriginal = false
	d.addedLines = make(map[int]bool)
	d.modifiedLines = make(map[int]bool)
	d.deletedLines = make(map[int]bool)
	d.prevLineCount = 0
	d.prevContent = ""
}

// HasOriginal returns true if original content has been set.
func (d *DiffTracker) HasOriginal() bool {
	return d.hasOriginal
}

// UpdateContent should be called whenever the content changes.
// It detects what type of change occurred and updates the tracked line states.
func (d *DiffTracker) UpdateContent(currentContent string, cursorLine int) {
	if !d.hasOriginal {
		return
	}

	currentLines := splitLines(currentContent)
	currentLineCount := len(currentLines)

	// Detect type of change
	if currentLineCount > d.prevLineCount {
		// Lines were added - mark the cursor line as added
		d.addedLines[cursorLine] = true
		// Remove any deleted marker at this position
		delete(d.deletedLines, cursorLine)
		// Shift existing markers for lines after the insertion
		d.shiftMarkersDown(cursorLine, currentLineCount-d.prevLineCount)
	} else if currentLineCount < d.prevLineCount {
		// Lines were deleted - mark this position as deleted
		d.deletedLines[cursorLine] = true
		// Remove any added marker at this position
		delete(d.addedLines, cursorLine)
		// Shift existing markers for lines after the deletion
		d.shiftMarkersUp(cursorLine, d.prevLineCount-currentLineCount)
	} else {
		// Same line count - check if content on cursor line changed
		if d.prevContent != currentContent {
			// Content changed on the same number of lines
			// Mark the cursor line as modified (unless it's already added)
			if !d.addedLines[cursorLine] {
				d.modifiedLines[cursorLine] = true
			}
		}
	}

	d.prevLineCount = currentLineCount
	d.prevContent = currentContent
}

// shiftMarkersDown shifts all markers at or after position down by count positions
func (d *DiffTracker) shiftMarkersDown(position int, count int) {
	// Create new maps with shifted positions
	newAdded := make(map[int]bool)
	newModified := make(map[int]bool)
	newDeleted := make(map[int]bool)

	for line := range d.addedLines {
		if line > position {
			newAdded[line+count] = true
		} else if line < position {
			newAdded[line] = true
		}
		// Lines at exactly position are being replaced by the new added line
	}
	// Add the newly added line
	newAdded[position] = true

	for line := range d.modifiedLines {
		if line >= position {
			newModified[line+count] = true
		} else {
			newModified[line] = true
		}
	}

	for line := range d.deletedLines {
		if line >= position {
			newDeleted[line+count] = true
		} else {
			newDeleted[line] = true
		}
	}

	d.addedLines = newAdded
	d.modifiedLines = newModified
	d.deletedLines = newDeleted
}

// shiftMarkersUp shifts all markers after position up by count positions
func (d *DiffTracker) shiftMarkersUp(position int, count int) {
	// Create new maps with shifted positions
	newAdded := make(map[int]bool)
	newModified := make(map[int]bool)
	newDeleted := make(map[int]bool)

	for line := range d.addedLines {
		if line > position {
			newLine := line - count
			if newLine >= 0 {
				newAdded[newLine] = true
			}
		} else if line < position {
			newAdded[line] = true
		}
		// Lines at position are being deleted, don't keep them
	}

	for line := range d.modifiedLines {
		if line > position {
			newLine := line - count
			if newLine >= 0 {
				newModified[newLine] = true
			}
		} else if line < position {
			newModified[line] = true
		}
	}

	for line := range d.deletedLines {
		if line > position {
			newLine := line - count
			if newLine >= 0 {
				newDeleted[newLine] = true
			}
		} else if line < position {
			newDeleted[line] = true
		}
	}

	// Mark the deleted position
	newDeleted[position] = true

	d.addedLines = newAdded
	d.modifiedLines = newModified
	d.deletedLines = newDeleted
}

// GetLineState returns the state of a specific line.
func (d *DiffTracker) GetLineState(lineIdx int) LineState {
	if d.addedLines[lineIdx] {
		return LineAdded
	}
	if d.modifiedLines[lineIdx] {
		return LineModified
	}
	if d.deletedLines[lineIdx] {
		return LineDeleted
	}
	return LineUnchanged
}

// ComputeLineStates returns the state for each line in the current content.
func (d *DiffTracker) ComputeLineStates(currentContent string) []LineState {
	currentLines := splitLines(currentContent)
	states := make([]LineState, len(currentLines))

	for i := range currentLines {
		states[i] = d.GetLineState(i)
	}

	return states
}

// ComputeLineStatesWithDeletions returns line states plus deletion markers.
func (d *DiffTracker) ComputeLineStatesWithDeletions(currentContent string) ([]LineState, []bool) {
	currentLines := splitLines(currentContent)
	states := make([]LineState, len(currentLines))
	deletedAt := make([]bool, len(currentLines)+1)

	for i := range currentLines {
		states[i] = d.GetLineState(i)
	}

	// Copy deleted line markers
	for line := range d.deletedLines {
		if line < len(deletedAt) {
			deletedAt[line] = true
		}
	}

	return states, deletedAt
}

// splitLines splits content into lines, preserving empty lines.
func splitLines(content string) []string {
	if content == "" {
		return []string{}
	}

	return strings.Split(content, "\n")
}
