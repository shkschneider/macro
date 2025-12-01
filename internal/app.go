package internal

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shkschneider/macro/api"
)

var (
	defaultMessage = "Macro v0.12.0 | Hit Ctrl-Space for command input."
	TermWidth      = 0 // Will be updated on WindowSizeMsg
	TermHeight     = 0 // Will be updated on WindowSizeMsg
)

// ReadOnlyMode defines the mode for read-only handling
type ReadOnlyMode int

const (
	// ReadOnlyAuto - detect from file permissions (default)
	ReadOnlyAuto ReadOnlyMode = iota
	// ReadOnlyForced - force read-only mode
	ReadOnlyForced
	// ReadWriteForced - force read-write mode (if file is writable)
	ReadWriteForced
)

// Global read-only mode setting
var globalReadOnlyMode = ReadOnlyAuto

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

func (m Model) Init() tea.Cmd {
	if m.ShowPicker {
		return m.Filepicker.Init()
	}
	return m.Textarea.Focus()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// If showing filepicker, handle filepicker messages
	if m.ShowPicker {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Check if quit command key is pressed
			if cmd := GetCommandByKey(msg); cmd != nil && cmd.Name == "quit" {
				// Save cursor state before quitting
				m.saveCurrentBufferState()
				if m.CursorState != nil {
					_ = m.CursorState.Save()
				}
				return m, tea.Quit
			}
		case tea.WindowSizeMsg:
			pickerHeight := msg.Height - 6
			if pickerHeight < 1 {
				pickerHeight = 1
			}
			m.Filepicker.SetHeight(pickerHeight)
			TermWidth = msg.Width
			TermHeight = msg.Height
		}

		m.Filepicker, cmd = m.Filepicker.Update(msg)

		if didSelect, path := m.Filepicker.DidSelectFile(msg); didSelect {
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
				m.Message = defaultMessage
				m.Err = nil
			} else {
				m.Message = fmt.Sprintf("Error loading file: %v", err)
				m.Err = err
			}
			m.ShowPicker = false

			if TermHeight > 0 {
				contentHeight := TermHeight - 2
				if contentHeight < 1 {
					contentHeight = 1
				}
				m.Textarea.SetWidth(TermWidth)
				m.Textarea.SetHeight(contentHeight)
				m.Viewport.Width = TermWidth
				m.Viewport.Height = contentHeight
			}

			m.Textarea.Focus()
			return m, m.Textarea.Focus()
		}

		return m, cmd
	}

	// If showing dialog, handle dialog messages
	if m.ActiveDialog != nil {
		m.ActiveDialog, cmd = m.ActiveDialog.Update(msg)

		// Check if dialog was closed
		if !m.ActiveDialog.IsVisible() {
			m.ActiveDialog = nil
		}

		// Return immediately to prevent the message from reaching the underlying buffer
		// This ensures that keyboard input (like arrow keys) only affects the dialog
		return m, cmd
	}

	// If command input is active, handle command input messages
	if m.CommandInput != nil && m.CommandInput.IsActive() {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEscape:
				// Escape deactivates command input and returns focus to editor
				m.CommandInput.Deactivate()
				return m, m.Textarea.Focus()
			case tea.KeyEnter:
				// Execute the command and deactivate input
				cmdText := m.CommandInput.Value()
				m.CommandInput.Deactivate()
				// Execute the command
				if cmdText != "" {
					return m, m.executeCommandInput(cmdText)
				}
				return m, m.Textarea.Focus()
			}
		}
		// Pass other messages to the command input
		m.CommandInput, cmd = m.CommandInput.Update(msg)
		return m, cmd
	}

	// Handle plugin messages using the PluginMsg interface
	// This allows plugins to define their own message types without
	// the main app needing to know about them in a switch statement
	if pluginMsg, ok := msg.(api.PluginMsg); ok {
		return m, pluginMsg.Handle(&m)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle key bindings by looking up registered commands
		if cmd := GetCommandByKey(msg); cmd != nil && cmd.Execute != nil {
			return m, cmd.Execute(&m)
		}

	case tea.WindowSizeMsg:
		contentHeight := msg.Height - 2
		if contentHeight < 1 {
			contentHeight = 1
		}

		m.Textarea.SetWidth(msg.Width)
		m.Textarea.SetHeight(contentHeight)
		m.Viewport.Width = msg.Width
		m.Viewport.Height = contentHeight

		TermWidth = msg.Width
		TermHeight = msg.Height

		return m, nil
	}

	// Update the appropriate component based on read-only state
	readOnly := m.isCurrentBufferReadOnly()
	if readOnly && m.Err == nil {
		m.Viewport, cmd = m.Viewport.Update(msg)
	} else if !readOnly && m.Err == nil {
		m.Textarea, cmd = m.Textarea.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	// If showing filepicker, keep the old layout for now
	if m.ShowPicker {
		return fmt.Sprintf("%s\n\n%s\n\n%s",
			lipgloss.NewStyle().Bold(true).Render("macro - Select a file"),
			m.Filepicker.View(),
			MessageStyle.Render("↑/↓: Navigate | Enter: Select | Ctrl-Q: Quit"))
	}

	// Content area - use viewport for read-only, syntaxTA for writable (with highlighting)
	var contentView string
	readOnly := m.isCurrentBufferReadOnly()
	if readOnly && m.Err == nil {
		contentView = m.Viewport.View()
	} else {
		contentView = m.Textarea.View()
	}

	// Build status bar with file info
	statusBar := m.BuildStatusBar()

	// Command input / message line - use CommandInput component for both modes
	var commandLine string
	if m.CommandInput != nil {
		// Sync the message with the command input component
		m.CommandInput.SetMessage(m.Message)
		// Determine styling flags
		isWarning := strings.Contains(m.Message, "WARNING") || strings.Contains(m.Message, "read-only")
		isSuccess := m.Message != defaultMessage && m.Err == nil && !isWarning
		commandLine = m.CommandInput.View(TermWidth, m.Err, isWarning, isSuccess)
	} else {
		// Fallback to old rendering if CommandInput is nil
		if m.Err != nil {
			commandLine = ErrorStyle.Render(m.Message)
		} else if strings.Contains(m.Message, "WARNING") || strings.Contains(m.Message, "read-only") {
			commandLine = WarningStyle.Render(m.Message)
		} else if m.Message != defaultMessage {
			commandLine = SuccessStyle.Render(m.Message)
		} else {
			commandLine = MessageStyle.Render(m.Message)
		}
	}

	baseView := fmt.Sprintf(
		"%s\n%s\n%s",
		contentView,
		statusBar,
		commandLine,
	)

	// If showing dialog, overlay it on top of the base view
	if m.ActiveDialog != nil && m.ActiveDialog.IsVisible() {
		dialog := m.ActiveDialog.View(TermWidth, TermHeight)
		return OverlayDialog(baseView, dialog, TermWidth, TermHeight)
	}

	return baseView
}

// executeCommandInput parses and executes a command from the command input line.
// It supports both direct command names and arguments.
func (m *Model) executeCommandInput(input string) tea.Cmd {
	// Trim whitespace
	input = strings.TrimSpace(input)
	if input == "" {
		return m.Textarea.Focus()
	}

	// Split the input into command and arguments
	parts := strings.Fields(input)
	cmdName := parts[0]

	// Try to find a command by name
	cmd := GetCommandByName(cmdName)
	if cmd != nil && cmd.Execute != nil {
		return cmd.Execute(m)
	}

	// If no exact match, try partial matching
	for _, cmd := range GetCommands() {
		if strings.Contains(strings.ToLower(cmd.Name), strings.ToLower(cmdName)) {
			if cmd.Execute != nil {
				return cmd.Execute(m)
			}
		}
	}

	// Command not found
	m.Message = fmt.Sprintf("Unknown command: %s", cmdName)
	return m.Textarea.Focus()
}
