// Package core provides core functionality for the macro editor.
package internal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// IsGitTracked returns true if the file is tracked by git.
// Returns false for untracked files, new files, or files not in a git repository.
func IsGitTracked(filePath string) bool {
	// Get the directory of the file
	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)

	// Run git ls-files to check if file is tracked
	// Use "--" to separate options from filenames to prevent injection
	cmd := exec.Command("git", "ls-files", "--", filename)
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		// Not in a git repo or git not available
		return false
	}

	// If output is non-empty, file is tracked
	return strings.TrimSpace(string(output)) != ""
}

// GetGitFileContent returns the content of a file from the git HEAD.
// Returns empty string and false if file is not tracked or git is not available.
func GetGitFileContent(filePath string) (string, bool) {
	// Get the directory of the file
	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)

	// Run git show to get content from HEAD
	// Use "HEAD:./filename" format to be explicit about path
	cmd := exec.Command("git", "show", "HEAD:./"+filename)
	cmd.Dir = dir

	output, err := cmd.Output()
	if err != nil {
		// File not in git, not committed yet, or git not available
		return "", false
	}

	return string(output), true
}

// GitDiffResult contains the line states as determined by git diff.
type GitDiffResult struct {
	AddedLines    map[int]bool // Line numbers (1-indexed) that were added
	ModifiedLines map[int]bool // Line numbers (1-indexed) that were modified
	DeletedLines  map[int]bool // Line numbers (1-indexed) where deletions occurred
}

// GetGitDiff runs git diff on the current file content and returns the line states.
// It writes both the HEAD content and current content to temp files and diffs them.
func GetGitDiff(filePath string, currentContent string) (*GitDiffResult, error) {
	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)

	result := &GitDiffResult{
		AddedLines:    make(map[int]bool),
		ModifiedLines: make(map[int]bool),
		DeletedLines:  make(map[int]bool),
	}

	// Get HEAD content first
	headCmd := exec.Command("git", "show", "HEAD:./"+filename)
	headCmd.Dir = dir
	headContent, err := headCmd.Output()
	if err != nil {
		// No HEAD content (new file or git error)
		return result, nil
	}

	// Write HEAD content to a temp file
	headTmpFile, err := os.CreateTemp("", "macro-diff-head-*")
	if err != nil {
		return result, err
	}
	defer os.Remove(headTmpFile.Name())

	if _, err := headTmpFile.Write(headContent); err != nil {
		headTmpFile.Close()
		return result, err
	}
	headTmpFile.Close()

	// Write current content to a temp file
	currentTmpFile, err := os.CreateTemp("", "macro-diff-current-*")
	if err != nil {
		return result, err
	}
	defer os.Remove(currentTmpFile.Name())

	if _, err := currentTmpFile.WriteString(currentContent); err != nil {
		currentTmpFile.Close()
		return result, err
	}
	currentTmpFile.Close()

	// Run git diff with both temp files
	// -U0 gives us minimal context (just the changed lines)
	cmd := exec.Command("git", "diff", "--no-index", "-U0", "--", headTmpFile.Name(), currentTmpFile.Name())
	cmd.Dir = dir

	output, _ := cmd.Output()
	// Note: git diff returns exit code 1 if there are differences, so we ignore the error

	// Parse the diff output
	parseGitDiffOutput(string(output), result)

	return result, nil
}

// parseGitDiffOutput parses git diff unified format output.
// Format: @@ -start,count +start,count @@
// Lines starting with + are added, - are deleted
func parseGitDiffOutput(diffOutput string, result *GitDiffResult) {
	lines := strings.Split(diffOutput, "\n")

	var newLineNum int
	inHunk := false

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			// Parse hunk header: @@ -oldStart,oldCount +newStart,newCount @@
			inHunk = true
			// Extract the +newStart part
			parts := strings.Split(line, " ")
			for _, part := range parts {
				if strings.HasPrefix(part, "+") && !strings.HasPrefix(part, "+++") {
					// Parse +start,count or +start
					numPart := strings.TrimPrefix(part, "+")
					if commaIdx := strings.Index(numPart, ","); commaIdx != -1 {
						numPart = numPart[:commaIdx]
					}
					newLineNum, _ = strconv.Atoi(numPart)
					break
				}
			}
		} else if inHunk {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				// Added line
				result.AddedLines[newLineNum] = true
				newLineNum++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				// Deleted line - mark at current position
				result.DeletedLines[newLineNum] = true
				// Don't increment newLineNum for deleted lines
			} else if !strings.HasPrefix(line, "\\") {
				// Context line (unchanged) - shouldn't appear with -U0 but handle anyway
				newLineNum++
			}
		}
	}

	// Post-process: if a line is both added and deleted at same position, it's modified
	for lineNum := range result.AddedLines {
		if result.DeletedLines[lineNum] {
			result.ModifiedLines[lineNum] = true
			delete(result.AddedLines, lineNum)
			delete(result.DeletedLines, lineNum)
		}
	}
}
