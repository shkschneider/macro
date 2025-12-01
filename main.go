package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	vanilla "github.com/shkschneider/macro/features/vanilla"
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
	vanilla.Register(func(cmd vanilla.CommandRegistration) {
		var execFunc func(*model) tea.Cmd

		// Provide execute handlers for commands that need *model access
		switch cmd.Name {
		case vanilla.CmdQuit:
			execFunc = executeQuit
		case vanilla.CmdHelp:
			execFunc = executeCommandPalette
		case vanilla.CmdFileOpen:
			execFunc = executeFileSwitcher
		case vanilla.CmdBufferSwitch:
			execFunc = executeBufferSwitcher
		default:
			// For commands with EditorContext execute (like save)
			if cmd.FeatureExecute != nil {
				execFunc = func(m *model) tea.Cmd {
					return cmd.FeatureExecute(m)
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
