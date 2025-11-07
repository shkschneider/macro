package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
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

	// Dialog styles
	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			Background(lipgloss.Color("235"))
	dialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230")).
				Padding(0, 1)

	defaultMessage = "Ctrl-S: Save | Ctrl-Space: File Switcher | Ctrl-Q: Quit"
	termWidth      = 0 // Will be updated on WindowSizeMsg
	termHeight     = 0 // Will be updated on WindowSizeMsg
)

// fileItem implements list.Item interface for the file dialog
type fileItem struct {
	name string
	path string
}

func (i fileItem) FilterValue() string { return i.name }
func (i fileItem) Title() string       { return i.name }
func (i fileItem) Description() string { return "" }

type model struct {
	textarea   textarea.Model
	viewport   viewport.Model
	filepicker filepicker.Model
	fileList   list.Model
	filePath   string
	message    string // Message line for errors/warnings/info
	err        error
	showPicker bool
	showDialog bool
	readOnly   bool
	isWarning  bool
}

func initialModel(filePath string) model {
	ta := textarea.New()
	ta.Focus()
	ta.Prompt = ""              // Remove default border on the left
	ta.ShowLineNumbers = true   // Enable line numbers for better navigation

	fp := filepicker.New()
	fp.DirAllowed = false
	fp.FileAllowed = true

	vp := viewport.New(80, 24)

	// Initialize list for file dialog
	delegate := list.NewDefaultDelegate()
	fileList := list.New([]list.Item{}, delegate, 0, 0)
	fileList.Title = "File Switcher"
	fileList.SetShowStatusBar(false)
	fileList.SetFilteringEnabled(true)
	fileList.Styles.Title = dialogTitleStyle

	m := model{
		textarea:   ta,
		viewport:   vp,
		filepicker: fp,
		fileList:   fileList,
		filePath:   filePath,
		message:    defaultMessage,
		err:        nil,
		showPicker: false,
		showDialog: false,
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

// getFilesInDirectory returns a list of files in the directory of the current file
func (m *model) getFilesInDirectory() []list.Item {
	if m.filePath == "" {
		return []list.Item{}
	}

	dir := filepath.Dir(m.filePath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []list.Item{}
	}

	var items []list.Item
	for _, entry := range entries {
		if !entry.IsDir() {
			fullPath := filepath.Join(dir, entry.Name())
			items = append(items, fileItem{
				name: entry.Name(),
				path: fullPath,
			})
		}
	}
	return items
}

// openFileDialog opens the file switcher dialog
func (m *model) openFileDialog() {
	items := m.getFilesInDirectory()
	if len(items) == 0 {
		m.message = "No files found in current directory"
		m.isWarning = true
		return
	}
	m.fileList.SetItems(items)
	m.showDialog = true
}

// closeFileDialog closes the file switcher dialog
func (m *model) closeFileDialog() {
	m.showDialog = false
	m.message = defaultMessage
	m.isWarning = false
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

	// If showing dialog, handle dialog messages
	if m.showDialog {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+q":
				return m, tea.Quit
			case "esc", "ctrl+space":
				// Close dialog
				m.closeFileDialog()
				return m, nil
			case "enter":
				// Select file from list
				if item, ok := m.fileList.SelectedItem().(fileItem); ok {
					// Load the selected file
					content, err := os.ReadFile(item.path)
					if err == nil {
						m.filePath = item.path
						
						// Check if file is read-only
						info, statErr := os.Stat(item.path)
						if statErr == nil && info.Mode()&0200 == 0 {
							m.readOnly = true
							m.isWarning = true
							m.message = "WARNING: File is read-only. Editing disabled. | Ctrl-Q: Quit"
							m.viewport.SetContent(string(content))
							m.viewport.GotoTop()
						} else {
							m.readOnly = false
							m.isWarning = false
							m.textarea.SetValue(string(content))
							m.moveCursorToTop()
							m.message = fmt.Sprintf("Opened %s", item.name)
						}
						m.err = nil
					} else {
						m.message = fmt.Sprintf("Error loading file: %v", err)
						m.err = err
					}
					m.closeFileDialog()
					return m, nil
				}
			}
		case tea.WindowSizeMsg:
			// Update dialog size - fixed to 50% width and height
			dialogWidth := msg.Width / 2
			dialogHeight := msg.Height / 2
			if dialogWidth < 40 {
				dialogWidth = 40
			}
			if dialogHeight < 10 {
				dialogHeight = 10
			}
			m.fileList.SetSize(dialogWidth, dialogHeight)
			termWidth = msg.Width
			termHeight = msg.Height
			return m, nil
		}

		// Update the list
		m.fileList, cmd = m.fileList.Update(msg)
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
		// Check for Ctrl-Space using string matching
		if msg.String() == "ctrl+ " {
			m.openFileDialog()
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

	baseView := fmt.Sprintf(
		"%s\n%s\n%s",
		contentView,
		statusBar,
		messageLine,
	)

	// If showing dialog, overlay it on top of the base view
	if m.showDialog {
		dialog := m.renderDialog()
		return m.overlayDialog(baseView, dialog)
	}

	return baseView
}

// renderDialog renders the file dialog with its border and title
func (m model) renderDialog() string {
	dialogContent := m.fileList.View()
	
	// Add instructions at the bottom
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("↑/↓: Navigate | /: Filter | Enter: Open | Esc: Close")
	
	fullContent := dialogContent + "\n" + instructions
	
	return dialogBoxStyle.Render(fullContent)
}

// overlayDialog overlays the dialog centered on top of the base view
func (m model) overlayDialog(baseView, dialog string) string {
	if termWidth == 0 || termHeight == 0 {
		return baseView
	}

	// Split both into lines
	baseLines := strings.Split(baseView, "\n")
	dialogLines := strings.Split(dialog, "\n")

	// Calculate dialog dimensions
	dialogHeight := len(dialogLines)
	dialogWidth := 0
	for _, line := range dialogLines {
		// Strip ANSI codes for accurate width calculation
		width := lipgloss.Width(line)
		if width > dialogWidth {
			dialogWidth = width
		}
	}

	// Calculate centering position
	startY := (termHeight - dialogHeight) / 2
	if startY < 0 {
		startY = 0
	}
	startX := (termWidth - dialogWidth) / 2
	if startX < 0 {
		startX = 0
	}

	// Ensure we have enough base lines
	for len(baseLines) < termHeight {
		baseLines = append(baseLines, "")
	}

	// Overlay dialog lines onto base lines
	for i, dialogLine := range dialogLines {
		y := startY + i
		if y >= 0 && y < len(baseLines) {
			baseLine := baseLines[y]
			baseWidth := lipgloss.Width(baseLine)
			
			// Pad base line to terminal width if needed
			if baseWidth < termWidth {
				baseLine += strings.Repeat(" ", termWidth-baseWidth)
			}
			
			// Calculate where to place dialog line
			dialogLineWidth := lipgloss.Width(dialogLine)
			
			// Build the new line with dialog overlaid
			var newLine strings.Builder
			
			// Add left part of base (before dialog)
			if startX > 0 {
				leftPart := baseLine
				if len(leftPart) > startX {
					leftPart = leftPart[:startX]
				}
				newLine.WriteString(leftPart)
			}
			
			// Add dialog content
			newLine.WriteString(dialogLine)
			
			// Add right part of base (after dialog)
			endX := startX + dialogLineWidth
			if endX < baseWidth {
				rightPart := baseLine[endX:]
				newLine.WriteString(rightPart)
			}
			
			baseLines[y] = newLine.String()
		}
	}

	return strings.Join(baseLines, "\n")
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
