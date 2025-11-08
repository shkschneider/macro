package main

import (
"fmt"
"os"

tea "github.com/charmbracelet/bubbletea"
feature "github.com/shkschneider/macro/feature"
)

func main() {
// Register feature commands
registerCommand(Command{
Name:        feature.FileSwitcherCommand().Name,
Key:         feature.FileSwitcherCommand().Key,
Description: feature.FileSwitcherCommand().Description,
Execute:     nil,
})
registerCommand(Command{
Name:        feature.BufferSwitcherCommand().Name,
Key:         feature.BufferSwitcherCommand().Key,
Description: feature.BufferSwitcherCommand().Description,
Execute:     nil,
})
registerCommand(Command{
Name:        feature.HelpCommand().Name,
Key:         feature.HelpCommand().Key,
Description: feature.HelpCommand().Description,
Execute:     nil,
})
registerCommand(Command{
Name:        feature.SaveCommand().Name,
Key:         feature.SaveCommand().Key,
Description: feature.SaveCommand().Description,
Execute:     executeFileSave,
})
registerCommand(Command{
Name:        feature.QuitCommand().Name,
Key:         feature.QuitCommand().Key,
Description: feature.QuitCommand().Description,
Execute:     executeQuit,
})

// Get filename from command line args
filePath := ""
if len(os.Args) > 1 {
filePath = os.Args[1]
}

p := tea.NewProgram(initialModel(filePath), tea.WithAltScreen())
if _, err := p.Run(); err != nil {
fmt.Printf("Error: %v\n", err)
os.Exit(1)
}
}
