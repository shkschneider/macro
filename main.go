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

	// Register feature commands
	registerCommand(Command{
		Name:        feature.FileSwitcherCommand().Name,
		Key:         feature.FileSwitcherCommand().Key,
		Description: feature.FileSwitcherCommand().Description,
		KeyBinding:  feature.FileSwitcherCommand().KeyBinding,
		Execute:     executeFileSwitcher,
	})
	registerCommand(Command{
		Name:        feature.BufferSwitcherCommand().Name,
		Key:         feature.BufferSwitcherCommand().Key,
		Description: feature.BufferSwitcherCommand().Description,
		KeyBinding:  feature.BufferSwitcherCommand().KeyBinding,
		Execute:     executeBufferSwitcher,
	})
	registerCommand(Command{
		Name:        feature.HelpCommand().Name,
		Key:         feature.HelpCommand().Key,
		Description: feature.HelpCommand().Description,
		KeyBinding:  feature.HelpCommand().KeyBinding,
		Execute:     executeCommandPalette,
	})

	// Register save command - uses feature's execution logic via EditorContext
	saveCmd := feature.SaveCommand()
	registerCommand(Command{
		Name:        saveCmd.Name,
		Key:         saveCmd.Key,
		Description: saveCmd.Description,
		KeyBinding:  saveCmd.KeyBinding,
		Execute: func(m *model) tea.Cmd {
			return saveCmd.Execute(m)
		},
	})

	registerCommand(Command{
		Name:        feature.QuitCommand().Name,
		Key:         feature.QuitCommand().Key,
		Description: feature.QuitCommand().Description,
		KeyBinding:  feature.QuitCommand().KeyBinding,
		Execute:     executeQuit,
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
