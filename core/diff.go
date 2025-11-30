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
// It stores the file path to use git diff for comparison.
type DiffTracker struct {
	filePath    string
	hasOriginal bool
}

// NewDiffTracker creates a new DiffTracker.
func NewDiffTracker() *DiffTracker {
	return &DiffTracker{
		filePath:    "",
		hasOriginal: false,
	}
}

// SetFilePath sets the file path for git diff comparison.
func (d *DiffTracker) SetFilePath(filePath string) {
	d.filePath = filePath
	d.hasOriginal = IsGitTracked(filePath)
}

// SetOriginal is kept for compatibility but now just checks if file is git-tracked.
func (d *DiffTracker) SetOriginal(content string) {
	// This is now a no-op since we use git diff directly
	// The hasOriginal flag is set by SetFilePath
}

// ClearOriginal clears the original content (e.g., for new files).
func (d *DiffTracker) ClearOriginal() {
	d.filePath = ""
	d.hasOriginal = false
}

// HasOriginal returns true if original content has been set.
func (d *DiffTracker) HasOriginal() bool {
	return d.hasOriginal
}

// ComputeLineStates computes the diff between original and current content using git diff.
func (d *DiffTracker) ComputeLineStates(currentContent string) []LineState {
	currentLines := splitLines(currentContent)
	states := make([]LineState, len(currentLines))

	// If no file path or not git-tracked, all lines are unchanged
	if !d.hasOriginal || d.filePath == "" {
		return states
	}

	// Use git diff to get the actual line states
	gitDiff, err := GetGitDiff(d.filePath, currentContent)
	if err != nil {
		return states
	}

	// Apply git diff results to line states
	for i := range states {
		lineNum := i + 1 // git uses 1-indexed line numbers
		if gitDiff.AddedLines[lineNum] {
			states[i] = LineAdded
		} else if gitDiff.ModifiedLines[lineNum] {
			states[i] = LineModified
		}
		// Deleted lines don't show on any current line
	}

	return states
}

// ComputeLineStatesWithDeletions returns line states plus deletion markers.
func (d *DiffTracker) ComputeLineStatesWithDeletions(currentContent string) ([]LineState, []bool) {
	currentLines := splitLines(currentContent)
	states := make([]LineState, len(currentLines))
	deletedAt := make([]bool, len(currentLines)+1)

	// If no file path or not git-tracked, all lines are unchanged
	if !d.hasOriginal || d.filePath == "" {
		return states, deletedAt
	}

	// Use git diff to get the actual line states
	gitDiff, err := GetGitDiff(d.filePath, currentContent)
	if err != nil {
		return states, deletedAt
	}

	// Apply git diff results to line states
	for i := range states {
		lineNum := i + 1 // git uses 1-indexed line numbers
		if gitDiff.AddedLines[lineNum] {
			states[i] = LineAdded
		} else if gitDiff.ModifiedLines[lineNum] {
			states[i] = LineModified
		}
	}

	// Mark deletion positions (keep 1-indexed since that's what we get from git)
	for lineNum := range gitDiff.DeletedLines {
		// lineNum is 1-indexed, deletedAt array is 0-indexed
		// A deletion at line N means content was removed at that position
		if lineNum > 0 && lineNum <= len(deletedAt) {
			deletedAt[lineNum-1] = true
		}
	}

	return states, deletedAt
}

// splitLines splits content into lines.
func splitLines(content string) []string {
	if content == "" {
		return []string{}
	}
	return strings.Split(content, "\n")
}
