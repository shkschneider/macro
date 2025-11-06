package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	statusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))

	defaultStatus = "Ctrl-S: Save | Ctrl-Q: Quit"
)

// isTextFile checks if the file content is text (not binary)
func isTextFile(content []byte) bool {
	if len(content) == 0 {
		return true // Empty files are treated as text
	}

	// Use http.DetectContentType to detect the MIME type
	contentType := http.DetectContentType(content)

	// Check if it's a text type
	return strings.HasPrefix(contentType, "text/") ||
		contentType == "application/json" ||
		contentType == "application/xml" ||
		contentType == "application/javascript"
}

// isReadOnly checks if the file is read-only
func isReadOnly(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// Check if we have write permission
	// On Unix-like systems, check if the owner write bit is set
	mode := info.Mode()
	return mode&0200 == 0 // Owner write bit not set
}

type model struct {
	textarea   textarea.Model
	viewport   viewport.Model
	filepicker filepicker.Model
	filePath   string
	status     string
	err        error
	showPicker bool
	readOnly   bool
	isWarning  bool
}

func initialModel(filePath string) model {
	ta := textarea.New()
	ta.Focus()
	ta.ShowLineNumbers = false

	fp := filepicker.New()
	fp.DirAllowed = false
	fp.FileAllowed = true

	vp := viewport.New(80, 24)

	m := model{
		textarea:   ta,
		viewport:   vp,
		filepicker: fp,
		filePath:   filePath,
		status:     defaultStatus,
		err:        nil,
		showPicker: false,
		readOnly:   false,
		isWarning:  false,
	}

	// Check if filePath is a directory
	if filePath != "" {
		info, err := os.Stat(filePath)
		if err != nil {
			m.status = fmt.Sprintf("Error: Error: %v | Ctrl-Q: Quit", err)
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
				m.status = fmt.Sprintf("Error: Cannot read file: %v | Ctrl-Q: Quit", err)
				m.err = err
				return m
			}
			// Check if file is text
			if !isTextFile(content) {
				m.status = "Error: Cannot open binary file. Please use a binary editor."
				m.err = fmt.Errorf("binary file detected")
				return m
			}
			// Check if file is read-only
			if isReadOnly(filePath) {
				m.readOnly = true
				m.isWarning = true
				m.status = "WARNING: File is read-only. Editing disabled. | Ctrl-Q: Quit"
				// Use viewport for read-only files
				m.viewport.SetContent(string(content))
			} else {
				// Use textarea for writable files
				m.textarea.SetValue(string(content))
			}
		}
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
			if m.readOnly {
				m.status = "WARNING: Cannot save - file is read-only | Ctrl-Q: Quit"
				m.isWarning = true
				return m, nil
			}
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
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 3
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
		if m.readOnly {
			title += " [READ-ONLY]"
		}
	}

	// Status bar
	statusBar := m.status
	if m.err != nil {
		statusBar = errorStyle.Render(m.status)
	} else if m.isWarning {
		statusBar = warningStyle.Render(m.status)
	} else if m.status != defaultStatus {
		statusBar = successStyle.Render(m.status)
	} else {
		statusBar = statusStyle.Render(m.status)
	}

	// Content area - use viewport for read-only, textarea for writable
	var contentView string
	if m.readOnly && m.err == nil {
		contentView = m.viewport.View()
	} else {
		contentView = m.textarea.View()
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		titleStyle.Render(title),
		contentView,
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
