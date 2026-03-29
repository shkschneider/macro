package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/go-errors/errors"
	"github.com/micro-editor/micro/v2/internal/shell"
)

// fuzzyFindFile uses fzf via micro's shell functions for proper terminal control
func fuzzyFindFile(directory string) (string, error) {
	// First check if fzf is available
	_, err := exec.LookPath("fzf")
	if err != nil {
		return "", errors.New("fzf not found - please install fzf for directory fuzzy finding")
	}

	// Create a simple shell script approach to avoid command parsing issues
	// Build the command as a proper shell script
	script := fmt.Sprintf(`#!/bin/bash
find "%s" -type f -not -path '*/.*' -not -path '*/.git/*' | fzf \
	--prompt='' \
	--border \
	--ansi \
	--layout=default
`, directory)

	// Create temporary script file
	tmpScript, err := os.CreateTemp("", "macro_fzf_*.sh")
	if err != nil {
		return "", fmt.Errorf("failed to create temp script: %w", err)
	}
	defer os.Remove(tmpScript.Name())

	// Write script to file
	if _, err := tmpScript.WriteString(script); err != nil {
		tmpScript.Close()
		return "", fmt.Errorf("failed to write script: %w", err)
	}
	tmpScript.Close()

	// Make script executable
	if err := os.Chmod(tmpScript.Name(), 0755); err != nil {
		return "", fmt.Errorf("failed to make script executable: %w", err)
	}

	// Use micro's RunInteractiveShell for proper terminal control
	output, err := shell.RunInteractiveShell(tmpScript.Name(), false, true)
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			switch exitError.ExitCode() {
			case 130: // Ctrl-C
				return "", errors.New("Cancelled")
			case 1: // No selection or Esc
				return "", errors.New("Cancelled")
			}
		}
		return "", fmt.Errorf("fzf failed: %w", err)
	}

	selectedFile := strings.TrimSpace(output)
	if selectedFile == "" {
		return "", errors.New("No file selected")
	}

	return selectedFile, nil
}
