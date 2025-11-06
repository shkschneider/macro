package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

type model struct {
	textarea textarea.Model
	filePath string
	status   string
	err      error
}

func initialModel(filePath string) model {
	ti := textarea.New()
	ti.Focus()
	ti.ShowLineNumbers = false

	// Load file if specified
	if filePath != "" {
		content, err := os.ReadFile(filePath)
		if err == nil {
			ti.SetValue(string(content))
		}
	}

	return model{
		textarea: ti,
		filePath: filePath,
		status:   "Ctrl-S: Save | Ctrl-Q: Quit",
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlQ:
			return m, tea.Quit
		case tea.KeyCtrlS:
			if m.filePath == "" {
				m.status = "Error: No filename specified. Usage: macro <filename>"
				m.err = fmt.Errorf("no filename")
			} else {
				err := os.WriteFile(m.filePath, []byte(m.textarea.Value()), 0644)
				if err != nil {
					m.status = fmt.Sprintf("Error saving: %v", err)
					m.err = err
				} else {
					m.status = fmt.Sprintf("Saved to %s | Ctrl-S: Save | Ctrl-Q: Quit", m.filePath)
					m.err = nil
				}
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(msg.Height - 3) // Reserve space for title and status
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m model) View() string {
	// Title
	title := "macro - Simple Text Editor"
	if m.filePath != "" {
		title = fmt.Sprintf("macro - %s", m.filePath)
	}

	// Status bar
	statusBar := m.status
	if m.err != nil {
		statusBar = errorStyle.Render(m.status)
	} else if m.status != "Ctrl-S: Save | Ctrl-Q: Quit" {
		statusBar = successStyle.Render(m.status)
	} else {
		statusBar = statusStyle.Render(m.status)
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render(title),
		m.textarea.View(),
		statusBar,
	)
}

func main() {
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
