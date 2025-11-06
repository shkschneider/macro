package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app      *tview.Application
	textArea *tview.TextArea
	filePath string
	status   *tview.TextView
)

func main() {
	// Get filename from command line args
	if len(os.Args) > 1 {
		filePath = os.Args[1]
	}

	app = tview.NewApplication()

	// Create text area
	textArea = tview.NewTextArea()
	textArea.SetBorder(true)
	textArea.SetTitle("macro - Simple Text Editor")

	// Load file if specified
	if filePath != "" {
		content, err := os.ReadFile(filePath)
		if err == nil {
			textArea.SetText(string(content), true)
		}
		textArea.SetTitle(fmt.Sprintf("macro - %s", filePath))
	}

	// Create status bar
	status = tview.NewTextView()
	status.SetDynamicColors(true)
	status.SetText("[yellow]Ctrl-S[white]: Save | [yellow]Ctrl-Q[white]: Quit")

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textArea, 0, 1, true).
		AddItem(status, 1, 0, false)

	// Set input capture for keyboard shortcuts
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			// Quit
			app.Stop()
			return nil
		case tcell.KeyCtrlS:
			// Save
			if filePath == "" {
				status.SetText("[red]Error: No filename specified. Usage: macro <filename>")
			} else {
				err := os.WriteFile(filePath, []byte(textArea.GetText()), 0644)
				if err != nil {
					status.SetText(fmt.Sprintf("[red]Error saving: %v", err))
				} else {
					status.SetText(fmt.Sprintf("[green]Saved to %s | [yellow]Ctrl-S[white]: Save | [yellow]Ctrl-Q[white]: Quit", filePath))
				}
			}
			return nil
		}
		return event
	})

	// Run the application
	if err := app.SetRoot(flex, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
