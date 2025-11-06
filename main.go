package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

	defaultStatus = "Ctrl-S: Save | Ctrl-Q: Quit"
)

type model struct {
	textarea   textarea.Model
	filepicker filepicker.Model
	filePath   string
	status     string
	err        error
	showPicker bool
}

func initialModel(filePath string) model {
	ti := textarea.New()
	ti.Focus()
	ti.ShowLineNumbers = false

	fp := filepicker.New()
	fp.DirAllowed = false
	fp.FileAllowed = true

	m := model{
		textarea:   ti,
		filepicker: fp,
		filePath:   filePath,
		status:     defaultStatus,
		err:        nil,
		showPicker: false,
	}

	// Check if filePath is a directory
	if filePath != "" {
		info, err := os.Stat(filePath)
		if err == nil && info.IsDir() {
			// It's a directory, show filepicker
			m.showPicker = true
			m.filepicker.CurrentDirectory = filePath
		} else if err == nil {
			// It's a file, load it
			content, err := os.ReadFile(filePath)
			if err == nil {
				m.textarea.SetValue(string(content))
			}
		}
		// If error, just leave empty
	}

	return m
}

func (m model) Init() tea.Cmd {
	if m.showPicker {
		return m.filepicker.Init()
	}
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// If showing filepicker, handle filepicker messages
	if m.showPicker {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlQ:
				return m, tea.Quit
			}
		case tea.WindowSizeMsg:
			m.filepicker.SetHeight(msg.Height - 3)
		}

		// Update filepicker
		m.filepicker, cmd = m.filepicker.Update(msg)

		// Check if file was selected
		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			// Load the selected file
			m.filePath = path
			content, err := os.ReadFile(path)
			if err == nil {
				m.textarea.SetValue(string(content))
				m.status = defaultStatus
				m.err = nil
			} else {
				m.status = fmt.Sprintf("Error loading file: %v", err)
				m.err = err
			}
			m.showPicker = false
			m.textarea.Focus()
			return m, textarea.Blink
		}

		return m, cmd
	}

	// Normal editor mode
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
	// If showing filepicker, render it instead
	if m.showPicker {
		title := titleStyle.Render("macro - Select a file")
		instructions := statusStyle.Render("↑/↓: Navigate | Enter: Select | Ctrl-Q: Quit")
		return fmt.Sprintf("%s\n\n%s\n\n%s", title, m.filepicker.View(), instructions)
	}

	// Normal editor view
	// Title
	title := "macro - Simple Text Editor"
	if m.filePath != "" {
		title = fmt.Sprintf("macro - %s", m.filePath)
	}

	// Status bar
	statusBar := m.status
	if m.err != nil {
		statusBar = errorStyle.Render(m.status)
	} else if m.status != defaultStatus {
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
