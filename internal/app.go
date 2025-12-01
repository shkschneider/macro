package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shkschneider/macro/core"
	vanilla "github.com/shkschneider/macro/plugins/vanilla"
)

var (
	defaultMessage = "Macro v0.11.0 | Hit Ctrl-Space for Command Palette."
	termWidth      = 0 // Will be updated on WindowSizeMsg
	termHeight     = 0 // Will be updated on WindowSizeMsg
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

func (m Model) Init() tea.Cmd {
	if m.showPicker {
		return m.filepicker.Init()
	}
	return m.syntaxTA.Focus()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// If showing filepicker, handle filepicker messages
	if m.showPicker {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			// Check if quit command key is pressed
			if cmd := GetCommandByKey(msg); cmd != nil && cmd.Name == "quit" {
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
	case vanilla.FileSelectedMsg:
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

	case vanilla.BufferSelectedMsg:
		// Save current buffer state before switching
		m.saveCurrentBufferState()
		// Switch to selected buffer
		m.loadBuffer(msg.Index)
		m.message = fmt.Sprintf("Switched to buffer")
		return m, nil

	case vanilla.CommandSelectedMsg:
		// Execute the selected command
		cmd := GetCommandByName(msg.CommandName)
		if cmd != nil && cmd.Execute != nil {
			return m, cmd.Execute(&m)
		}
		return m, nil

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

func (m Model) View() string {
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
	statusBar := m.BuildStatusBar()

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
