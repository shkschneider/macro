package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
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

	defaultMessage = "Ctrl-S: Save | Ctrl-Space: Files | Ctrl-B: Buffers | Ctrl-Q: Quit"
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

// Buffer represents an open file with its state
type Buffer struct {
	filePath string
	content  string
	readOnly bool
}

// bufferItem implements list.Item interface for the buffer dialog
type bufferItem struct {
	name  string
	index int
}

func (i bufferItem) FilterValue() string { return i.name }
func (i bufferItem) Title() string       { return i.name }
func (i bufferItem) Description() string { return "" }

type model struct {
	textarea       textarea.Model
	viewport       viewport.Model
	filepicker     filepicker.Model
	fileList       list.Model
	bufferList     list.Model
	filterInput    textinput.Model
	bufferFilterInput textinput.Model
	allFiles       []fileItem      // All files in directory
	filteredFiles  []fileItem      // Filtered files based on input
	allBuffers     []bufferItem    // All buffers for buffer dialog
	filteredBuffers []bufferItem   // Filtered buffers based on input
	selectedIdx    int             // Selected file index in filtered list
	bufferSelectedIdx int          // Selected buffer index in filtered list
	buffers        []Buffer        // All open buffers
	currentBuffer  int             // Index of current buffer
	message        string          // Message line for errors/warnings/info
	err            error
	showPicker     bool
	showDialog     bool
	showBufferDialog bool
	dialogType     string          // "file" or "buffer"
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

	// Initialize list for file dialog (filtering disabled, we handle it ourselves)
	delegate := list.NewDefaultDelegate()
	fileList := list.New([]list.Item{}, delegate, 0, 0)
	fileList.Title = "File Switcher"
	fileList.SetShowStatusBar(false)
	fileList.SetFilteringEnabled(false) // Disable built-in filtering
	fileList.Styles.Title = dialogTitleStyle

	// Initialize list for buffer dialog
	bufferList := list.New([]list.Item{}, delegate, 0, 0)
	bufferList.Title = "Buffer Switcher"
	bufferList.SetShowStatusBar(false)
	bufferList.SetFilteringEnabled(false)
	bufferList.Styles.Title = dialogTitleStyle

	// Initialize text input for fuzzy filtering
	ti := textinput.New()
	ti.Placeholder = "Type to filter files..."
	ti.CharLimit = 100
	ti.Width = 50

	// Initialize text input for buffer filtering
	bti := textinput.New()
	bti.Placeholder = "Type to filter buffers..."
	bti.CharLimit = 100
	bti.Width = 50

	m := model{
		textarea:          ta,
		viewport:          vp,
		filepicker:        fp,
		fileList:          fileList,
		bufferList:        bufferList,
		filterInput:       ti,
		bufferFilterInput: bti,
		allFiles:          []fileItem{},
		filteredFiles:     []fileItem{},
		allBuffers:        []bufferItem{},
		filteredBuffers:   []bufferItem{},
		selectedIdx:       0,
		bufferSelectedIdx: 0,
		buffers:           []Buffer{},
		currentBuffer:     -1, // No buffer open initially
		message:           defaultMessage,
		err:               nil,
		showPicker:        false,
		showDialog:        false,
		showBufferDialog:  false,
		dialogType:        "",
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
			// It's a file, load it into first buffer
			content, err := os.ReadFile(filePath)
			if err != nil {
				// Handle file read errors
				m.message = fmt.Sprintf("Error: Cannot read file: %v | Ctrl-Q: Quit", err)
				m.err = err
				return m
			}
			// Check if file is read-only
			readOnly := info.Mode()&0200 == 0
			
			// Create initial buffer
			buf := Buffer{
				filePath: filePath,
				content:  string(content),
				readOnly: readOnly,
			}
			m.buffers = append(m.buffers, buf)
			m.currentBuffer = 0
			
			// Load buffer into UI
			m.loadBuffer(0)
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

// loadBuffer loads a buffer's content into the UI (textarea or viewport)
func (m *model) loadBuffer(idx int) {
	if idx < 0 || idx >= len(m.buffers) {
		return
	}
	
	buf := m.buffers[idx]
	
	if buf.readOnly {
		m.viewport.SetContent(buf.content)
		m.viewport.GotoTop()
		m.message = "WARNING: File is read-only. Editing disabled."
	} else {
		m.textarea.SetValue(buf.content)
		m.moveCursorToTop()
		m.message = defaultMessage
	}
	m.currentBuffer = idx
}

// saveCurrentBufferState saves the current UI state to the current buffer
func (m *model) saveCurrentBufferState() {
	if m.currentBuffer < 0 || m.currentBuffer >= len(m.buffers) {
		return
	}
	
	buf := &m.buffers[m.currentBuffer]
	if !buf.readOnly {
		buf.content = m.textarea.Value()
	}
}

// addBuffer adds a new buffer or switches to existing one if file already open
func (m *model) addBuffer(filePath string, content string, readOnly bool) int {
	// Check if buffer already exists
	for i, buf := range m.buffers {
		if buf.filePath == filePath {
			return i // Return existing buffer index
		}
	}
	
	// Create new buffer
	buf := Buffer{
		filePath: filePath,
		content:  content,
		readOnly: readOnly,
	}
	m.buffers = append(m.buffers, buf)
	return len(m.buffers) - 1
}

// getCurrentFilePath returns the file path of the current buffer
func (m *model) getCurrentFilePath() string {
	if m.currentBuffer >= 0 && m.currentBuffer < len(m.buffers) {
		return m.buffers[m.currentBuffer].filePath
	}
	return ""
}

// isCurrentBufferReadOnly returns whether the current buffer is read-only
func (m *model) isCurrentBufferReadOnly() bool {
	if m.currentBuffer >= 0 && m.currentBuffer < len(m.buffers) {
		return m.buffers[m.currentBuffer].readOnly
	}
	return false
}

// getFilesInDirectory returns a list of files in the directory of the current file
func (m *model) getFilesInDirectory() []fileItem {
	filePath := m.getCurrentFilePath()
	if filePath == "" {
		return []fileItem{}
	}

	dir := filepath.Dir(filePath)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []fileItem{}
	}

	var items []fileItem
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

// applyFuzzyFilter filters files based on the input text
func (m *model) applyFuzzyFilter() {
	query := m.filterInput.Value()
	
	if query == "" {
		// No filter, show all files
		m.filteredFiles = m.allFiles
		m.selectedIdx = 0
		return
	}

	// Build list of file names for fuzzy matching
	var targets []string
	for _, file := range m.allFiles {
		targets = append(targets, file.name)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)
	
	// Build filtered list based on matches
	m.filteredFiles = []fileItem{}
	for _, match := range matches {
		m.filteredFiles = append(m.filteredFiles, m.allFiles[match.Index])
	}

	// Reset selection to first item
	m.selectedIdx = 0
}

// applyBufferFuzzyFilter filters buffers based on the input text
func (m *model) applyBufferFuzzyFilter() {
	query := m.bufferFilterInput.Value()
	
	if query == "" {
		// No filter, show all buffers
		m.filteredBuffers = m.allBuffers
		m.bufferSelectedIdx = 0
		return
	}

	// Build list of buffer names for fuzzy matching
	var targets []string
	for _, buffer := range m.allBuffers {
		targets = append(targets, buffer.name)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)
	
	// Build filtered list based on matches
	m.filteredBuffers = []bufferItem{}
	for _, match := range matches {
		m.filteredBuffers = append(m.filteredBuffers, m.allBuffers[match.Index])
	}

	// Reset selection to first item
	m.bufferSelectedIdx = 0
}

// openFileDialog opens the file switcher dialog
func (m *model) openFileDialog() {
	items := m.getFilesInDirectory()
	if len(items) == 0 {
		m.message = "No files found in current directory"
		return
	}
	m.allFiles = items
	m.filteredFiles = items
	m.selectedIdx = 0
	m.filterInput.SetValue("")
	m.filterInput.Focus()
	m.showDialog = true
	m.dialogType = "file"
}

// closeFileDialog closes the file switcher dialog
func (m *model) closeFileDialog() {
	m.showDialog = false
	m.filterInput.Blur()
	m.message = defaultMessage
}

// openBufferDialog opens the buffer switcher dialog
func (m *model) openBufferDialog() {
	if len(m.buffers) == 0 {
		m.message = "No buffers open"
		return
	}
	
	// Build buffer list
	m.allBuffers = []bufferItem{}
	for i, buf := range m.buffers {
		name := filepath.Base(buf.filePath)
		if buf.readOnly {
			name += " [RO]"
		}
		if i == m.currentBuffer {
			name = "* " + name
		}
		m.allBuffers = append(m.allBuffers, bufferItem{
			name:  name,
			index: i,
		})
	}
	m.filteredBuffers = m.allBuffers
	m.bufferSelectedIdx = 0
	m.bufferFilterInput.SetValue("")
	m.bufferFilterInput.Focus()
	m.showBufferDialog = true
	m.dialogType = "buffer"
}

// closeBufferDialog closes the buffer switcher dialog
func (m *model) closeBufferDialog() {
	m.showBufferDialog = false
	m.bufferFilterInput.Blur()
	m.message = defaultMessage
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
			// Load the selected file into a new buffer
			content, err := os.ReadFile(path)
			if err == nil {
				info, statErr := os.Stat(path)
				readOnly := false
				if statErr == nil {
					readOnly = info.Mode()&0200 == 0
				}
				
				// Add buffer and switch to it
				bufferIdx := m.addBuffer(path, string(content), readOnly)
				m.loadBuffer(bufferIdx)
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
	if m.showDialog || m.showBufferDialog {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+q":
				return m, tea.Quit
			case "esc", "ctrl+space", "ctrl+b", "ctrl+c":
				// Close dialog
				if m.showDialog {
					m.closeFileDialog()
				} else if m.showBufferDialog {
					m.closeBufferDialog()
				}
				return m, nil
			case "enter":
				if m.showDialog {
					// Select file from filtered list
					if m.selectedIdx >= 0 && m.selectedIdx < len(m.filteredFiles) {
						item := m.filteredFiles[m.selectedIdx]
						// Save current buffer state before switching
						m.saveCurrentBufferState()
						
						// Load the selected file into a new buffer
						content, err := os.ReadFile(item.path)
						if err == nil {
							// Check if file is read-only
							info, statErr := os.Stat(item.path)
							readOnly := false
							if statErr == nil {
								readOnly = info.Mode()&0200 == 0
							}
							
							// Add buffer and switch to it
							bufferIdx := m.addBuffer(item.path, string(content), readOnly)
							m.loadBuffer(bufferIdx)
							m.message = fmt.Sprintf("Opened %s", item.name)
							m.err = nil
						} else {
							m.message = fmt.Sprintf("Error loading file: %v", err)
							m.err = err
						}
						m.closeFileDialog()
						return m, nil
					}
				} else if m.showBufferDialog {
					// Select buffer from filtered list
					if m.bufferSelectedIdx >= 0 && m.bufferSelectedIdx < len(m.filteredBuffers) {
						item := m.filteredBuffers[m.bufferSelectedIdx]
						// Save current buffer state before switching
						m.saveCurrentBufferState()
						
						// Switch to selected buffer
						m.loadBuffer(item.index)
						m.message = fmt.Sprintf("Switched to %s", m.buffers[item.index].filePath)
						m.closeBufferDialog()
						return m, nil
					}
				}
			case "up", "ctrl+k":
				// Move selection up
				if m.showDialog {
					if m.selectedIdx > 0 {
						m.selectedIdx--
					}
				} else if m.showBufferDialog {
					if m.bufferSelectedIdx > 0 {
						m.bufferSelectedIdx--
					}
				}
				return m, nil
			case "down", "ctrl+j":
				// Move selection down
				if m.showDialog {
					if m.selectedIdx < len(m.filteredFiles)-1 {
						m.selectedIdx++
					}
				} else if m.showBufferDialog {
					if m.bufferSelectedIdx < len(m.filteredBuffers)-1 {
						m.bufferSelectedIdx++
					}
				}
				return m, nil
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
			m.filterInput.Width = dialogWidth - 4 // Account for padding
			m.bufferFilterInput.Width = dialogWidth - 4
			termWidth = msg.Width
			termHeight = msg.Height
			return m, nil
		}

		// Update the appropriate text input based on dialog type
		var cmd tea.Cmd
		if m.showDialog {
			m.filterInput, cmd = m.filterInput.Update(msg)
			// Apply fuzzy filter after input changes
			m.applyFuzzyFilter()
		} else if m.showBufferDialog {
			m.bufferFilterInput, cmd = m.bufferFilterInput.Update(msg)
			// Apply buffer fuzzy filter after input changes
			m.applyBufferFuzzyFilter()
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
			readOnly := m.isCurrentBufferReadOnly()
			if readOnly {
				m.message = "WARNING: Cannot save - file is read-only"
				return m, nil
			}
			filePath := m.getCurrentFilePath()
			if filePath == "" {
				m.message = "Error: No filename specified. Usage: macro <filename>"
				m.err = fmt.Errorf("no filename")
			} else {
				// Save current buffer state first
				m.saveCurrentBufferState()
				err := os.WriteFile(filePath, []byte(m.textarea.Value()), 0644)
				if err != nil {
					m.message = fmt.Sprintf("Error saving: %v", err)
					m.err = err
				} else {
					m.message = fmt.Sprintf("Saved to %s", filePath)
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
		// Check for Ctrl-B to open buffer dialog
		if msg.String() == "ctrl+b" {
			m.openBufferDialog()
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
	readOnly := m.isCurrentBufferReadOnly()
	if readOnly && m.err == nil {
		// Use viewport for read-only files (allows scrolling)
		m.viewport, cmd = m.viewport.Update(msg)
	} else if !readOnly && m.err == nil {
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
	readOnly := m.isCurrentBufferReadOnly()
	if readOnly && m.err == nil {
		contentView = m.viewport.View()
	} else {
		contentView = m.textarea.View()
	}

	// Build status bar with file info
	statusInfo := ""
	filePath := m.getCurrentFilePath()
	if filePath != "" {
		statusInfo = filePath
		if readOnly {
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
	} else if strings.Contains(m.message, "WARNING") || strings.Contains(m.message, "read-only") {
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
		dialog := m.renderFileDialog()
		return m.overlayDialog(baseView, dialog)
	}
	if m.showBufferDialog {
		dialog := m.renderBufferDialog()
		return m.overlayDialog(baseView, dialog)
	}

	return baseView
}

// renderFileDialog renders the file dialog with its border and title
func (m model) renderFileDialog() string {
	// Calculate dialog dimensions
	dialogWidth := termWidth / 2
	dialogHeight := termHeight / 2
	if dialogWidth < 40 {
		dialogWidth = 40
	}
	if dialogHeight < 10 {
		dialogHeight = 10
	}

	// Build file list view
	var fileListView strings.Builder
	listHeight := dialogHeight - 6 // Reserve space for title, input, and padding
	
	startIdx := 0
	endIdx := len(m.filteredFiles)
	
	// Calculate visible range (scroll if needed)
	if endIdx > listHeight {
		// Ensure selected item is visible
		if m.selectedIdx >= listHeight {
			startIdx = m.selectedIdx - listHeight + 1
		}
		endIdx = startIdx + listHeight
		if endIdx > len(m.filteredFiles) {
			endIdx = len(m.filteredFiles)
		}
	}

	// Render visible files
	for i := startIdx; i < endIdx; i++ {
		file := m.filteredFiles[i]
		line := ""
		if i == m.selectedIdx {
			// Selected item
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("63")).
				Width(dialogWidth - 4).
				Render("> " + file.name)
		} else {
			// Normal item
			line = lipgloss.NewStyle().
				Width(dialogWidth - 4).
				Render("  " + file.name)
		}
		fileListView.WriteString(line + "\n")
	}
	
	// Pad the list if needed
	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		fileListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	// Build the complete dialog content
	title := dialogTitleStyle.Render("File Switcher")
	fileCount := fmt.Sprintf("(%d/%d files)", len(m.filteredFiles), len(m.allFiles))
	titleLine := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Render(title + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(fileCount))
	
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(strings.Repeat("─", dialogWidth-4))
	
	inputLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Filter: ")
	
	inputView := inputLabel + m.filterInput.View()
	
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("↑/↓: Navigate | Enter: Open | Esc: Close")

	fullContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		titleLine,
		separator,
		strings.TrimRight(fileListView.String(), "\n"),
		inputView,
		instructions,
	)
	
	return dialogBoxStyle.Render(fullContent)
}

// renderBufferDialog renders the buffer dialog with its border and title
func (m model) renderBufferDialog() string {
	// Calculate dialog dimensions
	dialogWidth := termWidth / 2
	dialogHeight := termHeight / 2
	if dialogWidth < 40 {
		dialogWidth = 40
	}
	if dialogHeight < 10 {
		dialogHeight = 10
	}

	// Build buffer list view
	var bufferListView strings.Builder
	listHeight := dialogHeight - 6 // Reserve space for title, input, and padding
	
	startIdx := 0
	endIdx := len(m.filteredBuffers)
	
	// Calculate visible range (scroll if needed)
	if endIdx > listHeight {
		// Ensure selected item is visible
		if m.bufferSelectedIdx >= listHeight {
			startIdx = m.bufferSelectedIdx - listHeight + 1
		}
		endIdx = startIdx + listHeight
		if endIdx > len(m.filteredBuffers) {
			endIdx = len(m.filteredBuffers)
		}
	}

	// Render visible buffers
	for i := startIdx; i < endIdx; i++ {
		buffer := m.filteredBuffers[i]
		line := ""
		if i == m.bufferSelectedIdx {
			// Selected item
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("63")).
				Width(dialogWidth - 4).
				Render("> " + buffer.name)
		} else {
			// Normal item
			line = lipgloss.NewStyle().
				Width(dialogWidth - 4).
				Render("  " + buffer.name)
		}
		bufferListView.WriteString(line + "\n")
	}
	
	// Pad the list if needed
	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		bufferListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	// Build the complete dialog content
	title := dialogTitleStyle.Render("Buffer Switcher")
	bufferCount := fmt.Sprintf("(%d/%d buffers)", len(m.filteredBuffers), len(m.allBuffers))
	titleLine := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Render(title + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(bufferCount))
	
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(strings.Repeat("─", dialogWidth-4))
	
	inputLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Filter: ")
	
	inputView := inputLabel + m.bufferFilterInput.View()
	
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("↑/↓: Navigate | Enter: Switch | Esc: Close")

	fullContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		titleLine,
		separator,
		strings.TrimRight(bufferListView.String(), "\n"),
		inputView,
		instructions,
	)
	
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
