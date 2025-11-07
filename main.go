package main

import (
"fmt"
"os"
"strings"

"github.com/charmbracelet/bubbles/filepicker"
"github.com/charmbracelet/bubbles/list"
"github.com/charmbracelet/bubbles/textarea"
"github.com/charmbracelet/bubbles/textinput"
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

defaultMessage = "Macro v0.7.0 | Hit Ctrl-H for Help."
termWidth      = 0 // Will be updated on WindowSizeMsg
termHeight     = 0 // Will be updated on WindowSizeMsg
)

type model struct {
textarea          textarea.Model
viewport          viewport.Model
filepicker        filepicker.Model
fileList          list.Model
bufferList        list.Model
helpList          list.Model
filterInput       textinput.Model
bufferFilterInput textinput.Model
helpFilterInput   textinput.Model
allFiles          []fileItem      // All files in directory
filteredFiles     []fileItem      // Filtered files based on input
allBuffers        []bufferItem    // All buffers for buffer dialog
filteredBuffers   []bufferItem    // Filtered buffers based on input
allCommands       []commandItem   // All commands for help dialog
filteredCommands  []commandItem   // Filtered commands based on input
selectedIdx       int             // Selected file index in filtered list
bufferSelectedIdx int             // Selected buffer index in filtered list
helpSelectedIdx   int             // Selected command index in filtered list
buffers           []Buffer        // All open buffers
currentBuffer     int             // Index of current buffer
message           string          // Message line for errors/warnings/info
err               error
showPicker        bool
showDialog        bool
showBufferDialog  bool
showHelpDialog    bool
dialogType        string          // "file", "buffer", or "help"
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

// Initialize list for help dialog
helpList := list.New([]list.Item{}, delegate, 0, 0)
helpList.Title = "Help"
helpList.SetShowStatusBar(false)
helpList.SetFilteringEnabled(false)
helpList.Styles.Title = dialogTitleStyle

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

// Initialize text input for help filtering
hti := textinput.New()
hti.Placeholder = "Type to filter commands..."
hti.CharLimit = 100
hti.Width = 50

m := model{
textarea:          ta,
viewport:          vp,
filepicker:        fp,
fileList:          fileList,
bufferList:        bufferList,
helpList:          helpList,
filterInput:       ti,
bufferFilterInput: bti,
helpFilterInput:   hti,
allFiles:          []fileItem{},
filteredFiles:     []fileItem{},
allBuffers:        []bufferItem{},
filteredBuffers:   []bufferItem{},
allCommands:       []commandItem{},
filteredCommands:  []commandItem{},
selectedIdx:       0,
bufferSelectedIdx: 0,
helpSelectedIdx:   0,
buffers:           []Buffer{},
currentBuffer:     -1, // No buffer open initially
message:           defaultMessage,
err:               nil,
showPicker:        false,
showDialog:        false,
showBufferDialog:  false,
showHelpDialog:    false,
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
if m.showDialog || m.showBufferDialog || m.showHelpDialog {
switch msg := msg.(type) {
case tea.KeyMsg:
switch msg.String() {
case "ctrl+q":
cmd := getCommandByName("quit")
if cmd != nil && cmd.execute != nil {
return m, cmd.execute(&m)
}
return m, tea.Quit
case "esc", "ctrl+space", "ctrl+b", "ctrl+h", "ctrl+c":
// Close dialog using command
cmd := getCommandByName("dialog-close")
if cmd != nil && cmd.execute != nil {
cmd.execute(&m)
}
return m, nil
case "enter":
// Execute dialog selection using command
cmd := getCommandByName("dialog-select")
if cmd != nil && cmd.execute != nil {
return m, cmd.execute(&m)
}
return m, nil
case "up", "ctrl+k":
// Navigate up using command
cmd := getCommandByName("dialog-navigate-up")
if cmd != nil && cmd.execute != nil {
cmd.execute(&m)
}
return m, nil
case "down", "ctrl+j":
// Navigate down using command
cmd := getCommandByName("dialog-navigate-down")
if cmd != nil && cmd.execute != nil {
cmd.execute(&m)
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
m.helpFilterInput.Width = dialogWidth - 4
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
} else if m.showHelpDialog {
m.helpFilterInput, cmd = m.helpFilterInput.Update(msg)
// Apply help fuzzy filter after input changes
m.applyHelpFuzzyFilter()
}

return m, cmd
}

// Normal editor mode
switch msg := msg.(type) {
case tea.KeyMsg:
switch msg.Type {
case tea.KeyCtrlQ:
cmd := getCommandByName("quit")
if cmd != nil && cmd.execute != nil {
return m, cmd.execute(&m)
}
return m, tea.Quit
case tea.KeyCtrlS:
cmd := getCommandByName("file-save")
if cmd != nil && cmd.execute != nil {
return m, cmd.execute(&m)
}
return m, nil
}
// Check for Ctrl-Space using string matching
if msg.String() == "ctrl+ " {
cmd := getCommandByName("file-open")
if cmd != nil && cmd.execute != nil {
cmd.execute(&m)
}
return m, nil
}
// Check for Ctrl-B to open buffer dialog
if msg.String() == "ctrl+b" {
cmd := getCommandByName("buffer-switch")
if cmd != nil && cmd.execute != nil {
cmd.execute(&m)
}
return m, nil
}
// Check for Ctrl-H to open help dialog
if msg.String() == "ctrl+h" {
cmd := getCommandByName("help-show")
if cmd != nil && cmd.execute != nil {
cmd.execute(&m)
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
if m.showHelpDialog {
dialog := m.renderHelpDialog()
return m.overlayDialog(baseView, dialog)
}

return baseView
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
