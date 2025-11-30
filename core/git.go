// Package core provides core functionality for the macro editor.
package core

import (
	"os/exec"
	"path/filepath"
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
