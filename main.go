package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	feature "github.com/shkschneider/macro/feature"
)

// ReadOnlyMode defines the mode for read-only handling
type ReadOnlyMode int

const (
	// ReadOnlyAuto - detect from file permissions (default)
	ReadOnlyAuto ReadOnlyMode = iota
	// ReadOnlyForced - force read-only mode
	ReadOnlyForced
	// ReadWriteForced - force read-write mode (if file is writable)
	ReadWriteForced
)

// Global read-only mode setting
var globalReadOnlyMode = ReadOnlyAuto

func main() {
	// Parse command line flags
	forceRO := flag.Bool("ro", false, "Force read-only mode")
	forceRW := flag.Bool("rw", false, "Force read-write mode (if file is writable)")
	flag.Parse()

	// Determine read-only mode
	if *forceRO && *forceRW {
		fmt.Println("Error: Cannot use both -ro and -rw flags")
		os.Exit(1)
	}
	if *forceRO {
		globalReadOnlyMode = ReadOnlyForced
	} else if *forceRW {
		globalReadOnlyMode = ReadWriteForced
	}

	// Register feature commands using auto-registration
	feature.Register(func(cmd feature.CommandRegistration) {
		var execFunc func(*model) tea.Cmd

		// Provide execute handlers for commands that need *model access
		switch cmd.Name {
		case "quit":
			execFunc = executeQuit
		case "help-show":
			execFunc = executeCommandPalette
		case "file-open":
			execFunc = executeFileSwitcher
		case "buffer-switch":
			execFunc = executeBufferSwitcher
		default:
			// For commands with EditorContext execute (like save)
			if cmd.Execute != nil {
				execFunc = func(m *model) tea.Cmd {
					return cmd.Execute(m)
				}
			}
		}

		registerCommand(Command{
			Name:        cmd.Name,
			Key:         cmd.Key,
			Description: cmd.Description,
			KeyBinding:  cmd.KeyBinding,
			Execute:     execFunc,
		})
	})

	// Get filename from remaining command line args
	args := flag.Args()
	filePath := ""
	if len(args) > 0 {
		filePath = args[0]
	}

	p := tea.NewProgram(initialModel(filePath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
