package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/shkschneider/macro/internal"
	"github.com/shkschneider/macro/plugins"
)

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

	// Register all feature commands using auto-registration from features registry
	plugins.Register(func(cmd plugins.CommandRegistration) {
		var execFunc func(*internal.Model) tea.Cmd

		// Special case: Help command needs access to CommandRegistry
		if cmd.Name == internal.CmdPalette {
			execFunc = internal.ExecuteCommandPalette
		} else if cmd.PluginExecute != nil {
			// Use the plugin's execute function via EditorContext
			execFunc = func(m *internal.Model) tea.Cmd {
				return cmd.PluginExecute(m)
			}
		}

		internal.RegisterCommand(internal.Command{
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

	p := tea.NewProgram(internal.InitialModel(filePath), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
