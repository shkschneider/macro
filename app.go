package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dustin/go-humanize"
	core "github.com/shkschneider/macro/core"
	feature "github.com/shkschneider/macro/feature"
)

var (
	defaultMessage = "Macro v0.11.0 | Hit Ctrl-Space for Command Palette."
	termWidth      = 0 // Will be updated on WindowSizeMsg
	termHeight     = 0 // Will be updated on WindowSizeMsg
)

// KeyMap defines the key bindings for the application
type KeyMap struct {
	Quit           key.Binding
	Save           key.Binding
	CommandPalette key.Binding
	FileOpen       key.Binding
	BufferSwitch   key.Binding
}

// DefaultKeyMap returns the default key bindings
var DefaultKeyMap = KeyMap{
	Quit: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "quit editor"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save file"),
	),
	CommandPalette: key.NewBinding(
		key.WithKeys("ctrl+@", "ctrl+ "), // ctrl+@ is what ctrl+space sends
		key.WithHelp("ctrl+space", "open command palette"),
	),
	FileOpen: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "open file switcher"),
	),
	BufferSwitch: key.NewBinding(
		key.WithKeys("ctrl+b"),
		key.WithHelp("ctrl+b", "switch buffer"),
	),
}

type model struct {
	syntaxTA      *core.SyntaxTextarea
	viewport      viewport.Model
	filepicker    filepicker.Model
	buffers       []Buffer // All open buffers
	currentBuffer int      // Index of current buffer
	message       string   // Message line for errors/warnings/info
	err           error
	showPicker    bool
	activeDialog  core.Dialog         // Single active dialog (nil when closed)
	cursorState   *core.CursorState   // Persistent cursor position storage
}

func initialModel(filePath string) model {
	sta := core.NewSyntaxTextarea()
	sta.Focus()

	fp := filepicker.New()
	fp.DirAllowed = false
	fp.FileAllowed = true

	vp := viewport.New(80, 24)

	m := model{
		syntaxTA:      sta,
		viewport:      vp,
		filepicker:    fp,
		buffers:       []Buffer{},
		currentBuffer: -1, // No buffer open initially
		message:       defaultMessage,
		err:           nil,
		showPicker:    false,
		activeDialog:  nil,
		cursorState:   core.NewCursorState(),
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
			// Check if file is read-only based on permissions and CLI flags
			readOnly := determineReadOnly(info)

			// Create initial buffer with file size and original content tracking
			contentStr := string(content)
			buf := Buffer{
				filePath:        filePath,
				content:         contentStr,
				originalContent: contentStr,
				readOnly:        readOnly,
				fileSize:        info.Size(),
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
	return m.syntaxTA.Focus()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// If showing filepicker, handle filepicker messages
	if m.showPicker {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			if key.Matches(msg, DefaultKeyMap.Quit) {
				// Save cursor state before quitting
				m.saveCurrentBufferState()
				if m.cursorState != nil {
					_ = m.cursorState.Save()
				}
				return m, tea.Quit
			}
		case tea.WindowSizeMsg:
			pickerHeight := msg.Height - 6
			if pickerHeight < 1 {
				pickerHeight = 1
			}
			m.filepicker.SetHeight(pickerHeight)
			termWidth = msg.Width
			termHeight = msg.Height
		}

		m.filepicker, cmd = m.filepicker.Update(msg)

		if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
			content, err := os.ReadFile(path)
			if err == nil {
				info, statErr := os.Stat(path)
				readOnly := false
				var fileSize int64
				if statErr == nil {
					readOnly = determineReadOnly(info)
					fileSize = info.Size()
				}

				bufferIdx := m.addBuffer(path, string(content), readOnly, fileSize)
				m.loadBuffer(bufferIdx)
				m.message = defaultMessage
				m.err = nil
			} else {
				m.message = fmt.Sprintf("Error loading file: %v", err)
				m.err = err
			}
			m.showPicker = false

			if termHeight > 0 {
				contentHeight := termHeight - 2
				if contentHeight < 1 {
					contentHeight = 1
				}
				m.syntaxTA.SetWidth(termWidth)
				m.syntaxTA.SetHeight(contentHeight)
				m.viewport.Width = termWidth
				m.viewport.Height = contentHeight
			}

			m.syntaxTA.Focus()
			return m, m.syntaxTA.Focus()
		}

		return m, cmd
	}

	// If showing dialog, handle dialog messages
	if m.activeDialog != nil {
		m.activeDialog, cmd = m.activeDialog.Update(msg)

		// Check if dialog was closed
		if !m.activeDialog.IsVisible() {
			m.activeDialog = nil
		}

		// Return immediately to prevent the message from reaching the underlying buffer
		// This ensures that keyboard input (like arrow keys) only affects the dialog
		return m, cmd
	}

	// Handle custom dialog result messages
	switch msg := msg.(type) {
	case feature.FileSelectedMsg:
		// Save current buffer state before opening new file
		m.saveCurrentBufferState()
		// Load the selected file into a new buffer
		content, err := os.ReadFile(msg.Path)
		if err == nil {
			info, statErr := os.Stat(msg.Path)
			readOnly := false
			var fileSize int64
			if statErr == nil {
				readOnly = determineReadOnly(info)
				fileSize = info.Size()
			}

			bufferIdx := m.addBuffer(msg.Path, string(content), readOnly, fileSize)
			m.loadBuffer(bufferIdx)
			m.message = fmt.Sprintf("Opened %s", filepath.Base(msg.Path))
			m.err = nil
		} else {
			m.message = fmt.Sprintf("Error loading file: %v", err)
			m.err = err
		}
		return m, nil

	case feature.BufferSelectedMsg:
		// Save current buffer state before switching
		m.saveCurrentBufferState()
		// Switch to selected buffer
		m.loadBuffer(msg.Index)
		m.message = fmt.Sprintf("Switched to buffer")
		return m, nil

	case feature.CommandSelectedMsg:
		// Execute the selected command
		cmd := getCommandByName(msg.CommandName)
		if cmd != nil && cmd.Execute != nil {
			return m, cmd.Execute(&m)
		}
		return m, nil

	case tea.KeyMsg:
		// Handle key bindings using key.Matches
		if key.Matches(msg, DefaultKeyMap.Quit) {
			// Save cursor state before quitting
			m.saveCurrentBufferState()
			if m.cursorState != nil {
				_ = m.cursorState.Save()
			}
			return m, tea.Quit
		}
		if key.Matches(msg, DefaultKeyMap.Save) {
			cmd := getCommandByName("file-save")
			if cmd != nil && cmd.Execute != nil {
				return m, cmd.Execute(&m)
			}
			return m, nil
		}
		if key.Matches(msg, DefaultKeyMap.CommandPalette) {
			cmd := getCommandByName("help-show")
			if cmd != nil && cmd.Execute != nil {
				return m, cmd.Execute(&m)
			}
			return m, nil
		}
		if key.Matches(msg, DefaultKeyMap.FileOpen) {
			cmd := getCommandByName("file-open")
			if cmd != nil && cmd.Execute != nil {
				return m, cmd.Execute(&m)
			}
			return m, nil
		}
		if key.Matches(msg, DefaultKeyMap.BufferSwitch) {
			cmd := getCommandByName("buffer-switch")
			if cmd != nil && cmd.Execute != nil {
				return m, cmd.Execute(&m)
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		contentHeight := msg.Height - 2
		if contentHeight < 1 {
			contentHeight = 1
		}

		m.syntaxTA.SetWidth(msg.Width)
		m.syntaxTA.SetHeight(contentHeight)
		m.viewport.Width = msg.Width
		m.viewport.Height = contentHeight

		termWidth = msg.Width
		termHeight = msg.Height

		return m, nil
	}

	// Update the appropriate component based on read-only state
	readOnly := m.isCurrentBufferReadOnly()
	if readOnly && m.err == nil {
		m.viewport, cmd = m.viewport.Update(msg)
	} else if !readOnly && m.err == nil {
		m.syntaxTA, cmd = m.syntaxTA.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	// If showing filepicker, keep the old layout for now
	if m.showPicker {
		return fmt.Sprintf("%s\n\n%s\n\n%s",
			lipgloss.NewStyle().Bold(true).Render("macro - Select a file"),
			m.filepicker.View(),
			core.MessageStyle.Render("↑/↓: Navigate | Enter: Select | Ctrl-Q: Quit"))
	}

	// Content area - use viewport for read-only, syntaxTA for writable (with highlighting)
	var contentView string
	readOnly := m.isCurrentBufferReadOnly()
	if readOnly && m.err == nil {
		contentView = m.viewport.View()
	} else {
		contentView = m.syntaxTA.View()
	}

	// Build status bar with file info
	statusBar := m.buildStatusBar()

	// Message line for warnings/errors/info
	var messageLine string
	if m.err != nil {
		messageLine = core.ErrorStyle.Render(m.message)
	} else if strings.Contains(m.message, "WARNING") || strings.Contains(m.message, "read-only") {
		messageLine = core.WarningStyle.Render(m.message)
	} else if m.message != defaultMessage {
		messageLine = core.SuccessStyle.Render(m.message)
	} else {
		messageLine = core.MessageStyle.Render(m.message)
	}

	baseView := fmt.Sprintf(
		"%s\n%s\n%s",
		contentView,
		statusBar,
		messageLine,
	)

	// If showing dialog, overlay it on top of the base view
	if m.activeDialog != nil && m.activeDialog.IsVisible() {
		dialog := m.activeDialog.View(termWidth, termHeight)
		return core.OverlayDialog(baseView, dialog, termWidth, termHeight)
	}

	return baseView
}

// executeFileSave saves the current buffer to disk
func executeFileSave(m *model) tea.Cmd {
	readOnly := m.isCurrentBufferReadOnly()
	if readOnly {
		m.message = "WARNING: Cannot save - file is read-only"
		return nil
	}
	filePath := m.getCurrentFilePath()
	if filePath == "" {
		m.message = "Error: No filename specified. Usage: macro <filename>"
		m.err = fmt.Errorf("no filename")
	} else {
		// Save current buffer state first
		m.saveCurrentBufferState()
		content := m.syntaxTA.Value()
		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			m.message = fmt.Sprintf("Error saving: %v", err)
			m.err = err
		} else {
			m.message = fmt.Sprintf("Saved to %s", filePath)
			m.err = nil
			// Update original content and file size after successful save
			if buf := m.getCurrentBuffer(); buf != nil {
				buf.originalContent = content
				// Update file size from the newly written file
				if info, err := os.Stat(filePath); err == nil {
					buf.fileSize = info.Size()
				}
			}
		}
	}
	return nil
}

// executeQuit quits the editor
func executeQuit(m *model) tea.Cmd {
	// Save cursor state before quitting
	m.saveCurrentBufferState()
	if m.cursorState != nil {
		_ = m.cursorState.Save()
	}
	return tea.Quit
}

// executeFileSwitcher opens the file switcher dialog
func executeFileSwitcher(m *model) tea.Cmd {
	if m.getCurrentFilePath() != "" {
		m.activeDialog = feature.NewFileDialog(filepath.Dir(m.getCurrentFilePath()))
		return m.activeDialog.Init()
	}
	m.message = "No file open to determine directory"
	return nil
}

// executeBufferSwitcher opens the buffer switcher dialog
func executeBufferSwitcher(m *model) tea.Cmd {
	if len(m.buffers) > 0 {
		// Convert buffers to BufferInfo
		var bufferInfos []core.BufferInfo
		for _, buf := range m.buffers {
			bufferInfos = append(bufferInfos, core.BufferInfo{
				FilePath: buf.filePath,
				ReadOnly: buf.readOnly,
			})
		}
		m.activeDialog = feature.NewBufferDialog(bufferInfos, m.currentBuffer)
		return m.activeDialog.Init()
	}
	m.message = "No buffers open"
	return nil
}

// executeCommandPalette opens the command palette dialog
func executeCommandPalette(m *model) tea.Cmd {
	// Get all commands
	var commands []core.CommandDef
	for _, cmd := range getKeybindings() {
		commands = append(commands, core.CommandDef{
			Name:        cmd.Name,
			Key:         cmd.Key,
			Description: cmd.Description,
		})
	}
	m.activeDialog = feature.NewHelpDialog(commands)
	return m.activeDialog.Init()
}

// determineReadOnly determines the read-only state based on file info and CLI flags
func determineReadOnly(info os.FileInfo) bool {
	// Check file permissions
	fileIsWritable := info.Mode()&0200 != 0

	switch globalReadOnlyMode {
	case ReadOnlyForced:
		return true
	case ReadWriteForced:
		// Only allow read-write if file is actually writable
		return !fileIsWritable
	default: // ReadOnlyAuto
		return !fileIsWritable
	}
}

// buildStatusBar creates the formatted status bar with left and right sections
// Left: "filename.ext* [language] | human filesize"
// Right: "[RO] [fileencoding] charset [directory/path]"
func (m *model) buildStatusBar() string {
	buf := m.getCurrentBuffer()
	if buf == nil {
		return core.StatusBarStyle.Width(termWidth).Render("New File")
	}

	// Get file info
	fileName := filepath.Base(buf.filePath)
	lang := core.DetectLanguage(buf.filePath)
	dirPath := filepath.Dir(buf.filePath)
	modified := m.isCurrentBufferModified()
	readOnly := buf.readOnly

	// Build left side: "filename.ext* [language] | human filesize"
	leftParts := []string{}

	// Filename with modification indicator
	if modified {
		leftParts = append(leftParts, fileName+"*")
	} else {
		leftParts = append(leftParts, fileName)
	}

	// Language
	if lang != "" {
		leftParts = append(leftParts, "["+lang+"]")
	}

	// File size
	leftParts = append(leftParts, "|", humanize.Bytes(uint64(buf.fileSize)))

	leftSection := strings.Join(leftParts, " ")

	// Build right side: "[RO] [fileencoding] charset [directory/path]"
	rightParts := []string{}

	// Read-only indicator
	if readOnly {
		rightParts = append(rightParts, "[RO]")
	}

	// File encoding (assuming UTF-8 as default since we're reading text files)
	rightParts = append(rightParts, "[utf-8]")

	// Directory path
	rightParts = append(rightParts, "["+dirPath+"]")

	rightSection := strings.Join(rightParts, " ")

	// Calculate padding needed to align right section
	// Account for StatusBarStyle padding (1 on each side = 2 total)
	padding := 2
	contentWidth := termWidth - padding
	leftLen := len(leftSection)
	rightLen := len(rightSection)
	spacesNeeded := contentWidth - leftLen - rightLen
	if spacesNeeded < 1 {
		spacesNeeded = 1
	}

	fullStatusContent := leftSection + strings.Repeat(" ", spacesNeeded) + rightSection
	return core.StatusBarStyle.Width(termWidth).Render(fullStatusContent)
}
