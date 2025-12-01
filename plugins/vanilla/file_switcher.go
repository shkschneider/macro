package vanilla

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
	api "github.com/shkschneider/macro/api"
	plugin "github.com/shkschneider/macro/plugins"
)

// ====== Command Registration ======

// CmdFileOpen is the command name constant for file switcher
const CmdFileOpen = "file-open"

// FileSwitcherKeyBinding is the key binding for the file switcher command
var FileSwitcherKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+p"),
	key.WithHelp("ctrl+p", "open file switcher"),
)

func init() {
	plugin.RegisterCommand(plugin.CommandRegistration{
		Name:           CmdFileOpen,
		Key:            "Ctrl-P",
		Description:    "Open file switcher",
		KeyBinding:     FileSwitcherKeyBinding,
		PluginExecute: nil, // Main app provides execute handler
	})
}

// FileSwitcherCommand returns the command definition for file switching
func FileSwitcherCommand() api.CommandDef {
	return api.CommandDef{
		Name:        CmdFileOpen,
		Key:         "Ctrl-P",
		Description: "Open file switcher",
		KeyBinding:  FileSwitcherKeyBinding,
	}
}

// ====== Message Types ======

// FileSelectedMsg is sent when a file is selected in the file dialog
type FileSelectedMsg struct {
	Path string
}

// ====== Key Bindings ======

// FileDialogKeyMap defines the key bindings for the file dialog
type FileDialogKeyMap struct {
	Close key.Binding
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
}

// DefaultFileDialogKeyMap returns the default key bindings for file dialog
var DefaultFileDialogKeyMap = FileDialogKeyMap{
	Close: key.NewBinding(
		key.WithKeys("esc", "ctrl+c", "ctrl+p"),
		key.WithHelp("esc", "close dialog"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "ctrl+k"),
		key.WithHelp("↑/ctrl+k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "ctrl+j"),
		key.WithHelp("↓/ctrl+j", "move down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open file"),
	),
}

// ====== Internal Types ======

// fileItem is used internally by FileDialog
type fileItem struct {
	name string
	path string
}

// ====== Dialog Implementation ======

// FileDialog implements the Dialog interface for file selection
type FileDialog struct {
	filterInput   textinput.Model
	allFiles      []fileItem
	filteredFiles []fileItem
	selectedIdx   int
	visible       bool
	currentDir    string
	lastQuery     string // Track last query to avoid unnecessary resets
}

// NewFileDialog creates a new file dialog
func NewFileDialog(currentDir string) *FileDialog {
	ti := textinput.New()
	ti.Placeholder = "Type to filter files..."
	ti.CharLimit = 100
	ti.Width = 50
	ti.Focus()

	// Get files from directory
	var files []fileItem
	entries, err := os.ReadDir(currentDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				fullPath := filepath.Join(currentDir, entry.Name())
				files = append(files, fileItem{
					name: entry.Name(),
					path: fullPath,
				})
			}
		}
	}

	return &FileDialog{
		filterInput:   ti,
		allFiles:      files,
		filteredFiles: files,
		selectedIdx:   0,
		visible:       true,
		currentDir:    currentDir,
	}
}

func (d *FileDialog) Init() tea.Cmd {
	return textinput.Blink
}

func (d *FileDialog) Update(msg tea.Msg) (api.Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, DefaultFileDialogKeyMap.Close) {
			d.visible = false
			return d, nil
		}
		if key.Matches(msg, DefaultFileDialogKeyMap.Enter) {
			if d.selectedIdx >= 0 && d.selectedIdx < len(d.filteredFiles) {
				selectedFile := d.filteredFiles[d.selectedIdx]
				d.visible = false
				return d, func() tea.Msg {
					return FileSelectedMsg{Path: selectedFile.path}
				}
			}
		}
		if key.Matches(msg, DefaultFileDialogKeyMap.Up) {
			if d.selectedIdx > 0 {
				d.selectedIdx--
			}
			return d, nil
		}
		if key.Matches(msg, DefaultFileDialogKeyMap.Down) {
			if d.selectedIdx < len(d.filteredFiles)-1 {
				d.selectedIdx++
			}
			return d, nil
		}
	case tea.WindowSizeMsg:
		dialogWidth := msg.Width / 2
		if dialogWidth < 40 {
			dialogWidth = 40
		}
		d.filterInput.Width = dialogWidth - 4
		return d, nil
	}

	// Update filter input
	var cmd tea.Cmd
	oldValue := d.filterInput.Value()
	d.filterInput, cmd = d.filterInput.Update(msg)

	// Only apply filter if the query actually changed
	if d.filterInput.Value() != oldValue {
		d.applyFuzzyFilter()
	}
	return d, cmd
}

func (d *FileDialog) applyFuzzyFilter() {
	query := d.filterInput.Value()

	if query == "" {
		d.filteredFiles = d.allFiles
		// Only reset selection if query actually changed (not just cursor blink)
		if d.lastQuery != "" {
			d.selectedIdx = 0
		}
		d.lastQuery = query
		return
	}

	// Build list of file names for fuzzy matching
	var targets []string
	for _, file := range d.allFiles {
		targets = append(targets, file.name)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)

	// Build filtered list based on matches
	d.filteredFiles = []fileItem{}
	for _, match := range matches {
		d.filteredFiles = append(d.filteredFiles, d.allFiles[match.Index])
	}

	// Only reset selection when the query changes, not on every blink
	if d.lastQuery != query {
		d.selectedIdx = 0
	}
	d.lastQuery = query
}

func (d *FileDialog) View(termWidth, termHeight int) string {
	if !d.visible {
		return ""
	}

	dialogWidth := termWidth / 2
	dialogHeight := termHeight / 2
	if dialogWidth < 40 {
		dialogWidth = 40
	}
	if dialogHeight < 10 {
		dialogHeight = 10
	}

	var fileListView strings.Builder
	listHeight := dialogHeight - 6

	startIdx := 0
	endIdx := len(d.filteredFiles)

	if endIdx > listHeight {
		if d.selectedIdx >= listHeight {
			startIdx = d.selectedIdx - listHeight + 1
		}
		endIdx = startIdx + listHeight
		if endIdx > len(d.filteredFiles) {
			endIdx = len(d.filteredFiles)
		}
	}

	for i := startIdx; i < endIdx; i++ {
		file := d.filteredFiles[i]
		line := ""
		if i == d.selectedIdx {
			line = api.DialogHighlightedStyle.
				Width(dialogWidth - 4).
				Render("> " + file.name)
		} else {
			line = api.DialogItemStyle.
				Width(dialogWidth - 4).
				Render("  " + file.name)
		}
		fileListView.WriteString(line + "\n")
	}

	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		fileListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	title := api.DialogTitleStyle.Render("File Switcher")
	fileCount := fmt.Sprintf("(%d/%d files)", len(d.filteredFiles), len(d.allFiles))
	titleLine := api.DialogTitleLineStyle.
		Width(dialogWidth - 4).
		Render(title + " " + api.DialogCountStyle.Render(fileCount))

	separator := api.DialogSeparatorStyle.
		Render(strings.Repeat("─", dialogWidth-4))

	inputLabel := api.DialogInputLabelStyle.
		Render("Filter: ")

	inputView := inputLabel + d.filterInput.View()

	instructions := api.DialogInstructionsStyle.
		Render("↑/↓: Navigate | Enter: Open | Esc: Close")

	fullContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		titleLine,
		separator,
		strings.TrimRight(fileListView.String(), "\n"),
		inputView,
		instructions,
	)

	return api.DialogBoxStyle.Render(fullContent)
}

func (d *FileDialog) IsVisible() bool {
	return d.visible
}
