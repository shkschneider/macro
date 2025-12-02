package internal

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/shkschneider/macro/api"
)

type Model struct {
	Textarea      *Textarea
	Viewport      viewport.Model
	Filepicker    filepicker.Model
	Buffers       []Buffer // All open buffers
	CurrentBuffer int      // Index of current buffer
	Message       string   // Message line for errors/warnings/info
	Err           error
	ShowPicker    bool
	ActiveDialog  api.Dialog        // Single active dialog (nil when closed)
	CursorState   *CursorState // Persistent cursor position storage
}

func NewModel(filePath string) Model {
	ta := NewTextarea()
	ta.Focus()

	fp := filepicker.New()
	fp.DirAllowed = false
	fp.FileAllowed = true

	vp := viewport.New(80, 24)

	m := Model{
		Textarea:      ta,
		Viewport:      vp,
		Filepicker:    fp,
		Buffers:       []Buffer{},
		CurrentBuffer: -1, // No buffer open initially
		Message:       defaultMessage,
		Err:           nil,
		ShowPicker:    false,
		ActiveDialog:  nil,
		CursorState:   NewCursorState(),
	}

	if filePath != "" {
		info, err := os.Stat(filePath)
		if err != nil {
			m.Message = fmt.Sprintf("Error: Error: %v | Ctrl-Q: Quit", err)
			m.Err = err
			return m
		}
		if info.IsDir() {
			// It's a directory, show filepicker
			m.ShowPicker = true
			m.Filepicker.CurrentDirectory = filePath
		} else {
			// It's a file, load it into first buffer
			content, err := os.ReadFile(filePath)
			if err != nil {
				// Handle file read errors
				m.Message = fmt.Sprintf("Error: Cannot read file: %v | Ctrl-Q: Quit", err)
				m.Err = err
				return m
			}
			// Check if file is read-only based on permissions and CLI flags
			readOnly := determineReadOnly(info)

			// Create initial buffer with file size and original content tracking
			contentStr := string(content)
			buf := Buffer{
				FilePath:        filePath,
				Content:         contentStr,
				OriginalContent: contentStr,
				ReadOnly:        readOnly,
				FileSize:        info.Size(),
			}
			m.Buffers = append(m.Buffers, buf)
			m.CurrentBuffer = 0

			// Load buffer into UI
			m.loadBuffer(0)
		}
	}
	return m
}

