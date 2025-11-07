package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("15")). // White background
			Foreground(lipgloss.Color("0")).  // Black foreground
			Bold(true).
			Padding(0, 1) // Add horizontal padding
	messageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))

	defaultMessage = "Ctrl-S: Save | Ctrl-Q: Quit"
	termWidth      = 0 // Will be updated on WindowSizeMsg
	termHeight     = 0 // Will be updated on WindowSizeMsg
)

type model struct {
	textarea   textarea.Model
	viewport   viewport.Model
	filepicker filepicker.Model
	filePath   string
	message    string // Message line for errors/warnings/info
	err        error
	showPicker bool
	readOnly   bool
	isWarning  bool
}

func initialModel(filePath string) model {
	ta := textarea.New()
	ta.Focus()
	ta.Prompt = ""
	ta.ShowLineNumbers = true

	fp := filepicker.New()
	fp.DirAllowed = false
	fp.FileAllowed = true

	vp := viewport.New(80, 24)

	m := model{
		textarea:   ta,
		viewport:   vp,
		filepicker: fp,
		filePath:   filePath,
		message:    defaultMessage,
		err:        nil,
		showPicker: false,
		readOnly:   false,
		isWarning:  false,
	}

	if filePath != "" {
		info, err := os.Stat(filePath)
		if err != nil {
			m.message = fmt.Sprintf("Error: Error: %v | Ctrl-Q: Quit", err)
			m.err = err
			return m
		}
		if info.IsDir() {
			// It's a directory, show filepicker
			m.showPicker = true
			m.filepicker.CurrentDirectory = filePath
		} else {
			// It's a file, load it
			content, err := os.ReadFile(filePath)
			if err != nil {
				// Handle file read errors
				m.message = fmt.Sprintf("Error: Cannot read file: %v | Ctrl-Q: Quit", err)
				m.err = err
				return m
			}
			// Check if file is read-only
			if info.Mode()&0200 == 0 {
				m.readOnly = true
				m.isWarning = true
				m.message = "WARNING: File is read-only. Editing disabled. | Ctrl-Q: Quit"
				// Use viewport for read-only files
				m.viewport.SetContent(string(content))
				m.viewport.GotoTop()
			} else {
				// Use textarea for writable files
				m.textarea.SetValue(string(content))
				m.moveCursorToTop()
			}
		}
	}
	return m
}

// moveCursorToTop moves the textarea cursor to position (0,0)
func (m *model) moveCursorToTop() {
	m.textarea.CursorStart()
	for m.textarea.Line() > 0 {
		m.textarea.CursorUp()
	}
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
			// Title: 1 line + 2 newlines = 3 lines
			// Instructions: 2 newlines + 1 line = 3 lines
			// Total overhead: 6 lines
			pickerHeight := msg.Height - 6
			if pickerHeight < 1 {
				pickerHeight = 1
			}
			m.filepicker.SetHeight(pickerHeight)

			// Store terminal dimensions
			termWidth = msg.Width
			termHeight = msg.Height
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
				m.moveCursorToTop()
				m.message = defaultMessage
				m.err = nil
			} else {
				m.message = fmt.Sprintf("Error loading file: %v", err)
				m.err = err
			}
			m.showPicker = false

			// Resize textarea/viewport to editor dimensions (termHeight - 2 for status bar and message line)
			if termHeight > 0 {
				contentHeight := termHeight - 2
				if contentHeight < 1 {
					contentHeight = 1
				}
				m.textarea.SetWidth(termWidth)
				m.textarea.SetHeight(contentHeight)
				m.viewport.Width = termWidth
				m.viewport.Height = contentHeight
			}

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
			if m.readOnly {
				m.message = "WARNING: Cannot save - file is read-only | Ctrl-Q: Quit"
				m.isWarning = true
				return m, nil
			}
			if m.filePath == "" {
				m.message = "Error: No filename specified. Usage: macro <filename>"
				m.err = fmt.Errorf("no filename")
			} else {
				err := os.WriteFile(m.filePath, []byte(m.textarea.Value()), 0644)
				if err != nil {
					m.message = fmt.Sprintf("Error saving: %v", err)
					m.err = err
				} else {
					m.message = fmt.Sprintf("Saved to %s | Ctrl-S: Save | Ctrl-Q: Quit", m.filePath)
					m.err = nil
				}
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		// Calculate available height for content
		// New layout: content + status bar (1 line) + message line (1 line) = 2 lines overhead
		contentHeight := msg.Height - 2
		if contentHeight < 1 {
			contentHeight = 1
		}

		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(contentHeight)
		m.viewport.Width = msg.Width
		m.viewport.Height = contentHeight

		// Update terminal dimensions for status bar and future use
		termWidth = msg.Width
		termHeight = msg.Height

		return m, nil
	}

	// Update the appropriate component based on read-only state
	if m.readOnly && m.err == nil {
		// Use viewport for read-only files (allows scrolling)
		m.viewport, cmd = m.viewport.Update(msg)
	} else if !m.readOnly && m.err == nil {
		// Use textarea for writable files
		m.textarea, cmd = m.textarea.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	// If showing filepicker, keep the old layout for now
	if m.showPicker {
		return fmt.Sprintf("%s\n\n%s\n\n%s",
			lipgloss.NewStyle().Bold(true).Render("macro - Select a file"),
			m.filepicker.View(),
			messageStyle.Render("↑/↓: Navigate | Enter: Select | Ctrl-Q: Quit"))
	}

	// Content area - use viewport for read-only, textarea for writable
	var contentView string
	if m.readOnly && m.err == nil {
		contentView = m.viewport.View()
	} else {
		contentView = m.textarea.View()
	}

	// Build status bar with file info
	statusInfo := ""
	if m.filePath != "" {
		statusInfo = m.filePath
		if m.readOnly {
			statusInfo += " [READ-ONLY]"
		}
	} else {
		statusInfo = "New File"
	}

	// Apply width to fill the entire line with reverse video
	statusBar := statusBarStyle.Width(termWidth).Render(statusInfo)

	// Message line for warnings/errors/info
	var messageLine string
	if m.err != nil {
		messageLine = errorStyle.Render(m.message)
	} else if m.isWarning {
		messageLine = warningStyle.Render(m.message)
	} else if m.message != defaultMessage {
		messageLine = successStyle.Render(m.message)
	} else {
		messageLine = messageStyle.Render(m.message)
	}

	return fmt.Sprintf(
		"%s\n%s\n%s",
		contentView,
		statusBar,
		messageLine,
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
