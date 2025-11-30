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
// It uses the original content to compute diffs on-demand.
type DiffTracker struct {
	originalLines []string
	hasOriginal   bool
}

// NewDiffTracker creates a new DiffTracker.
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

// ComputeLineStates computes the diff between original and current content.
// Uses a simple LCS-based algorithm to properly detect added, modified, and deleted lines.
func (d *DiffTracker) ComputeLineStates(currentContent string) []LineState {
	currentLines := splitLines(currentContent)

	// If no original content, all lines are unchanged (new file)
	if !d.hasOriginal {
		return make([]LineState, len(currentLines))
	}

	return computeLCSDiff(d.originalLines, currentLines)
}

// ComputeLineStatesWithDeletions returns line states plus deletion markers.
func (d *DiffTracker) ComputeLineStatesWithDeletions(currentContent string) ([]LineState, []bool) {
	states := d.ComputeLineStates(currentContent)
	currentLines := splitLines(currentContent)
	deletedAt := make([]bool, len(currentLines)+1)

	// Compute where deletions occurred
	if d.hasOriginal && len(d.originalLines) > len(currentLines) {
		// Original had more lines - find where they were deleted
		lcs := computeLCS(d.originalLines, currentLines)
		
		// Track positions
		lcsIdx := 0
		currIdx := 0
		
		for origIdx := 0; origIdx < len(d.originalLines); origIdx++ {
			if lcsIdx < len(lcs) && d.originalLines[origIdx] == lcs[lcsIdx] {
				// This line is in LCS - advance current index to match
				for currIdx < len(currentLines) && currentLines[currIdx] != lcs[lcsIdx] {
					currIdx++
				}
				if currIdx < len(currentLines) {
					currIdx++
				}
				lcsIdx++
			} else {
				// This original line is not in LCS
				// Check if it was deleted (not present in current at all)
				found := false
				for _, c := range currentLines {
					if c == d.originalLines[origIdx] {
						found = true
						break
					}
				}
				if !found {
					// Line was deleted - mark at the current position
					if currIdx <= len(currentLines) {
						deletedAt[currIdx] = true
					}
				}
			}
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

// computeLCSDiff computes line states using LCS algorithm.
// Lines in LCS are unchanged, lines only in current are added,
// lines only in original are deleted, and lines at same position with different content are modified.
func computeLCSDiff(original, current []string) []LineState {
	states := make([]LineState, len(current))

	// Compute LCS
	lcs := computeLCS(original, current)
	lcsSet := make(map[string]int) // line content -> count in LCS
	for _, line := range lcs {
		lcsSet[line]++
	}

	// Track which original lines have been matched
	originalMatched := make([]bool, len(original))
	currentMatched := make([]bool, len(current))

	// First pass: match lines that are in LCS
	origIdx := 0
	for i, line := range current {
		// Find this line in remaining original
		for j := origIdx; j < len(original); j++ {
			if !originalMatched[j] && original[j] == line && lcsSet[line] > 0 {
				originalMatched[j] = true
				currentMatched[i] = true
				lcsSet[line]--
				states[i] = LineUnchanged
				origIdx = j + 1
				break
			}
		}
	}

	// Second pass: classify unmatched lines
	for i := range current {
		if currentMatched[i] {
			continue // Already matched as unchanged
		}

		// Check if this position had an original line that was modified
		if i < len(original) && !originalMatched[i] {
			// Same position, different content = modified
			states[i] = LineModified
			originalMatched[i] = true
		} else {
			// Line doesn't correspond to any original line = added
			states[i] = LineAdded
		}
	}

	return states
}

// computeLCS computes the Longest Common Subsequence of two string slices.
func computeLCS(a, b []string) []string {
	m, n := len(a), len(b)
	if m == 0 || n == 0 {
		return []string{}
	}

	// Build LCS length table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	// Backtrack to find LCS
	lcs := make([]string, 0, dp[m][n])
	i, j := m, n
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append([]string{a[i-1]}, lcs...)
			i--
			j--
		} else if dp[i-1][j] > dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	return lcs
}
