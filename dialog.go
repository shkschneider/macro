package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	tea "github.com/charmbracelet/bubbletea"
)

// fileItem implements list.Item interface for the file dialog
type fileItem struct {
	name string
	path string
}

func (i fileItem) FilterValue() string { return i.name }
func (i fileItem) Title() string       { return i.name }
func (i fileItem) Description() string { return "" }

// bufferItem implements list.Item interface for the buffer dialog
type bufferItem struct {
	name  string
	index int
}

func (i bufferItem) FilterValue() string { return i.name }
func (i bufferItem) Title() string       { return i.name }
func (i bufferItem) Description() string { return "" }

// commandItem implements list.Item interface for the help dialog
type commandItem struct {
	command Command
}

func (i commandItem) FilterValue() string { return i.command.name + " " + i.command.description }
func (i commandItem) Title() string       { return i.command.name }
func (i commandItem) Description() string { return i.command.description }

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

// applyHelpFuzzyFilter filters commands based on the input text
func (m *model) applyHelpFuzzyFilter() {
	query := m.helpFilterInput.Value()
	
	if query == "" {
		// No filter, show all commands
		m.filteredCommands = m.allCommands
		m.helpSelectedIdx = 0
		return
	}

	// Build list of command names and descriptions for fuzzy matching
	var targets []string
	for _, cmd := range m.allCommands {
		targets = append(targets, cmd.command.name+" "+cmd.command.description)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)
	
	// Build filtered list based on matches
	m.filteredCommands = []commandItem{}
	for _, match := range matches {
		m.filteredCommands = append(m.filteredCommands, m.allCommands[match.Index])
	}

	// Reset selection to first item
	m.helpSelectedIdx = 0
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

// openHelpDialog opens the help dialog with all commands
func (m *model) openHelpDialog() {
	// Build command list
	m.allCommands = []commandItem{}
	for _, cmd := range getKeybindings() {
		m.allCommands = append(m.allCommands, commandItem{
			command: cmd,
		})
	}
	m.filteredCommands = m.allCommands
	m.helpSelectedIdx = 0
	m.helpFilterInput.SetValue("")
	m.helpFilterInput.Focus()
	m.showHelpDialog = true
	m.dialogType = "help"
}

// closeHelpDialog closes the help dialog
func (m *model) closeHelpDialog() {
	m.showHelpDialog = false
	m.helpFilterInput.Blur()
	m.message = defaultMessage
}

// closeCurrentDialog closes whichever dialog is currently open
func (m *model) closeCurrentDialog() {
	if m.showDialog {
		m.closeFileDialog()
	} else if m.showBufferDialog {
		m.closeBufferDialog()
	} else if m.showHelpDialog {
		m.closeHelpDialog()
	}
}

// navigateDialogUp moves selection up in the current dialog
func (m *model) navigateDialogUp() {
	if m.showDialog {
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
	} else if m.showBufferDialog {
		if m.bufferSelectedIdx > 0 {
			m.bufferSelectedIdx--
		}
	} else if m.showHelpDialog {
		if m.helpSelectedIdx > 0 {
			m.helpSelectedIdx--
		}
	}
}

// navigateDialogDown moves selection down in the current dialog
func (m *model) navigateDialogDown() {
	if m.showDialog {
		if m.selectedIdx < len(m.filteredFiles)-1 {
			m.selectedIdx++
		}
	} else if m.showBufferDialog {
		if m.bufferSelectedIdx < len(m.filteredBuffers)-1 {
			m.bufferSelectedIdx++
		}
	} else if m.showHelpDialog {
		if m.helpSelectedIdx < len(m.filteredCommands)-1 {
			m.helpSelectedIdx++
		}
	}
}

// executeDialogSelect executes the selection in the current dialog
func (m *model) executeDialogSelect() tea.Cmd {
	if m.showDialog {
		return m.executeFileSelect()
	} else if m.showBufferDialog {
		return m.executeBufferSelect()
	} else if m.showHelpDialog {
		return m.executeHelpSelect()
	}
	return nil
}

// executeFileSelect handles file selection in file dialog
func (m *model) executeFileSelect() tea.Cmd {
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
	}
	return nil
}

// executeBufferSelect handles buffer selection in buffer dialog
func (m *model) executeBufferSelect() tea.Cmd {
	if m.bufferSelectedIdx >= 0 && m.bufferSelectedIdx < len(m.filteredBuffers) {
		item := m.filteredBuffers[m.bufferSelectedIdx]
		// Save current buffer state before switching
		m.saveCurrentBufferState()
		
		// Switch to selected buffer
		m.loadBuffer(item.index)
		m.message = fmt.Sprintf("Switched to %s", m.buffers[item.index].filePath)
		m.closeBufferDialog()
	}
	return nil
}

// executeHelpSelect executes the selected command in help dialog
func (m *model) executeHelpSelect() tea.Cmd {
	if m.helpSelectedIdx >= 0 && m.helpSelectedIdx < len(m.filteredCommands) {
		item := m.filteredCommands[m.helpSelectedIdx]
		m.closeHelpDialog()
		// Execute the command
		if item.command.execute != nil {
			return item.command.execute(m)
		}
	}
	return nil
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

// renderHelpDialog renders the help dialog with all commands
func (m model) renderHelpDialog() string {
	// Calculate dialog dimensions
	dialogWidth := termWidth / 2
	dialogHeight := termHeight / 2
	if dialogWidth < 40 {
		dialogWidth = 40
	}
	if dialogHeight < 10 {
		dialogHeight = 10
	}

	// Build command list view
	var helpListView strings.Builder
	listHeight := dialogHeight - 6 // Reserve space for title, input, and padding
	
	startIdx := 0
	endIdx := len(m.filteredCommands)
	
	// Calculate visible range (scroll if needed)
	if endIdx > listHeight {
		// Ensure selected item is visible
		if m.helpSelectedIdx >= listHeight {
			startIdx = m.helpSelectedIdx - listHeight + 1
		}
		endIdx = startIdx + listHeight
		if endIdx > len(m.filteredCommands) {
			endIdx = len(m.filteredCommands)
		}
	}

	// Render visible commands
	for i := startIdx; i < endIdx; i++ {
		cmd := m.filteredCommands[i]
		// Format: command-name (Key) - description
		cmdText := fmt.Sprintf("%-20s %-12s %s", cmd.command.name, cmd.command.key, cmd.command.description)
		line := ""
		if i == m.helpSelectedIdx {
			// Selected item
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("63")).
				Width(dialogWidth - 4).
				Render("> " + cmdText)
		} else {
			// Normal item
			line = lipgloss.NewStyle().
				Width(dialogWidth - 4).
				Render("  " + cmdText)
		}
		helpListView.WriteString(line + "\n")
	}
	
	// Pad the list if needed
	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		helpListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	// Build the complete dialog content
	title := dialogTitleStyle.Render("Help - Commands")
	cmdCount := fmt.Sprintf("(%d/%d commands)", len(m.filteredCommands), len(m.allCommands))
	titleLine := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Render(title + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(cmdCount))
	
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(strings.Repeat("─", dialogWidth-4))
	
	inputLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Filter: ")
	
	inputView := inputLabel + m.helpFilterInput.View()
	
	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("↑/↓: Navigate | Enter: Run Command | Esc: Close")

	fullContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		titleLine,
		separator,
		strings.TrimRight(helpListView.String(), "\n"),
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
