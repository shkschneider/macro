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
	"github.com/shkschneider/macro/core"
)

type Model struct {
	syntaxTA      *core.SyntaxTextarea
	viewport      viewport.Model
	filepicker    filepicker.Model
	buffers       []Buffer // All open buffers
	currentBuffer int      // Index of current buffer
	message       string   // Message line for errors/warnings/info
	err           error
	showPicker    bool
	activeDialog  api.Dialog         // Single active dialog (nil when closed)
	cursorState   *core.CursorState   // Persistent cursor position storage
}

func InitialModel(filePath string) Model {
	sta := core.NewSyntaxTextarea()
	sta.Focus()

	fp := filepicker.New()
	fp.DirAllowed = false
	fp.FileAllowed = true

	vp := viewport.New(80, 24)

	m := Model{
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

// buildStatusBar creates the formatted status bar with left and right sections
// Left: "filename.ext* [language] | human filesize"
// Right: "line:col [RO] [fileencoding] [directory/path]"
func (m *Model) BuildStatusBar() string {
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
	leftParts = append(leftParts, humanize.Bytes(uint64(buf.fileSize)))

	if readOnly {
		leftParts = append(leftParts, "(read-only)")
	}

	leftSection := strings.Join(leftParts, " ")

	// Build right side: "line:col [RO] [fileencoding] [directory/path]"
	rightParts := []string{}

	// File encoding (assuming UTF-8 as default since we're reading text files)
	// rightParts = append(rightParts, "[utf-8]")

	// Cursor position (line:column)
	line, col := m.syntaxTA.CursorPosition()
	rightParts = append(rightParts, fmt.Sprintf("%d:%d", line, col))

	// Directory path
	rightParts = append(rightParts, "["+dirPath+"/]")

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
