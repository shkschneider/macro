package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/dustin/go-humanize"
	"github.com/shkschneider/macro/api"
)

type Model struct {
	SyntaxTA      *SyntaxTextarea
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

func InitialModel(filePath string) Model {
	sta := NewSyntaxTextarea()
	sta.Focus()

	fp := filepicker.New()
	fp.DirAllowed = false
	fp.FileAllowed = true

	vp := viewport.New(80, 24)

	m := Model{
		SyntaxTA:      sta,
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

// buildStatusBar creates the formatted status bar with left and right sections
// Left: "filename.ext* [language] | human filesize"
// Right: "line:col [RO] [fileencoding] [directory/path]"
func (m *Model) BuildStatusBar() string {
	buf := m.getCurrentBuffer()
	if buf == nil {
		return StatusBarStyle.Width(TermWidth).Render("New File")
	}

	// Get file info
	fileName := filepath.Base(buf.FilePath)
	lang := DetectLanguage(buf.FilePath)
	dirPath := filepath.Dir(buf.FilePath)
	modified := m.isCurrentBufferModified()
	readOnly := buf.ReadOnly

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
	leftParts = append(leftParts, humanize.Bytes(uint64(buf.FileSize)))

	if readOnly {
		leftParts = append(leftParts, "(read-only)")
	}

	leftSection := strings.Join(leftParts, " ")

	// Build right side: "line:col [RO] [fileencoding] [directory/path]"
	rightParts := []string{}

	// File encoding (assuming UTF-8 as default since we're reading text files)
	// rightParts = append(rightParts, "[utf-8]")

	// Cursor position (line:column)
	line, col := m.SyntaxTA.CursorPosition()
	rightParts = append(rightParts, fmt.Sprintf("%d:%d", line, col))

	// Directory path
	rightParts = append(rightParts, "["+dirPath+"/]")

	rightSection := strings.Join(rightParts, " ")

	// Calculate padding needed to align right section
	// Account for StatusBarStyle padding (1 on each side = 2 total)
	padding := 2
	contentWidth := TermWidth - padding
	leftLen := len(leftSection)
	rightLen := len(rightSection)
	spacesNeeded := contentWidth - leftLen - rightLen
	if spacesNeeded < 1 {
		spacesNeeded = 1
	}

	fullStatusContent := leftSection + strings.Repeat(" ", spacesNeeded) + rightSection
	return StatusBarStyle.Width(TermWidth).Render(fullStatusContent)
}
