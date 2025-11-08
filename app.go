package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	core "github.com/shkschneider/macro/core"
	feature "github.com/shkschneider/macro/feature"
)

var (
	defaultMessage = "Macro v0.8.0 | Hit Ctrl-H for Help."
	termWidth      = 0 // Will be updated on WindowSizeMsg
	termHeight     = 0 // Will be updated on WindowSizeMsg
)

type model struct {
	textarea      textarea.Model
	viewport      viewport.Model
	filepicker    filepicker.Model
	buffers       []Buffer // All open buffers
	currentBuffer int      // Index of current buffer
	message       string   // Message line for errors/warnings/info
	err           error
	showPicker    bool
	activeDialog  core.Dialog // Single active dialog (nil when closed)
}

func initialModel(filePath string) model {
	ta := textarea.New()
	ta.Focus()
	ta.Prompt = ""            // Remove default border on the left
	ta.ShowLineNumbers = true // Enable line numbers for better navigation

	fp := filepicker.New()
	fp.DirAllowed = false
	fp.FileAllowed = true

	vp := viewport.New(80, 24)

	m := model{
		textarea:      ta,
		viewport:      vp,
		filepicker:    fp,
		buffers:       []Buffer{},
		currentBuffer: -1, // No buffer open initially
		message:       defaultMessage,
		err:           nil,
		showPicker:    false,
		activeDialog:  nil,
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
				if statErr == nil {
					readOnly = info.Mode()&0200 == 0
				}

				bufferIdx := m.addBuffer(path, string(content), readOnly)
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
	if m.activeDialog != nil {
		m.activeDialog, cmd = m.activeDialog.Update(msg)

		// Check if dialog was closed
		if !m.activeDialog.IsVisible() {
			m.activeDialog = nil
		}

		// Dialog may have returned a command, let it propagate
		if cmd != nil {
			return m, cmd
		}
	}

	// Handle custom dialog result messages
	switch msg := msg.(type) {
	case feature.FileSelectedMsg:
		// Load the selected file into a new buffer
		content, err := os.ReadFile(msg.Path)
		if err == nil {
			info, statErr := os.Stat(msg.Path)
			readOnly := false
			if statErr == nil {
				readOnly = info.Mode()&0200 == 0
			}

			bufferIdx := m.addBuffer(msg.Path, string(content), readOnly)
			m.loadBuffer(bufferIdx)
			m.message = fmt.Sprintf("Opened %s", filepath.Base(msg.Path))
			m.err = nil
		} else {
			m.message = fmt.Sprintf("Error loading file: %v", err)
			m.err = err
		}
		return m, nil

	case feature.BufferSelectedMsg:
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
		switch msg.Type {
		case tea.KeyCtrlQ:
			return m, tea.Quit
		case tea.KeyCtrlS:
			cmd := getCommandByName("file-save")
			if cmd != nil && cmd.Execute != nil {
				return m, cmd.Execute(&m)
			}
			return m, nil
		}
		// Check for Ctrl-Space using string matching
		if msg.String() == "ctrl+ " {
			if m.getCurrentFilePath() != "" {
				m.activeDialog = feature.NewFileDialog(filepath.Dir(m.getCurrentFilePath()))
				return m, m.activeDialog.Init()
			}
			return m, nil
		}
		// Check for Ctrl-B to open buffer dialog
		if msg.String() == "ctrl+b" {
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
				return m, m.activeDialog.Init()
			} else {
				m.message = "No buffers open"
			}
			return m, nil
		}
		// Check for Ctrl-H to open help dialog
		if msg.String() == "ctrl+h" {
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
			return m, m.activeDialog.Init()
		}

	case tea.WindowSizeMsg:
		contentHeight := msg.Height - 2
		if contentHeight < 1 {
			contentHeight = 1
		}

		m.textarea.SetWidth(msg.Width)
		m.textarea.SetHeight(contentHeight)
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
			core.MessageStyle.Render("↑/↓: Navigate | Enter: Select | Ctrl-Q: Quit"))
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
	statusBar := core.StatusBarStyle.Width(termWidth).Render(statusInfo)

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
		err := os.WriteFile(filePath, []byte(m.textarea.Value()), 0644)
		if err != nil {
			m.message = fmt.Sprintf("Error saving: %v", err)
			m.err = err
		} else {
			m.message = fmt.Sprintf("Saved to %s", filePath)
			m.err = nil
		}
	}
	return nil
}

// executeQuit quits the editor
func executeQuit(m *model) tea.Cmd {
	return tea.Quit
}
