package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// fileItem is used internally by FileDialog
type fileItem struct {
	name string
	path string
}

// FileDialog implements the Dialog interface for file selection
type FileDialog struct {
	filterInput   textinput.Model
	allFiles      []fileItem
	filteredFiles []fileItem
	selectedIdx   int
	visible       bool
	currentDir    string
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

func (d *FileDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "ctrl+space":
			d.visible = false
			return d, nil
		case "enter":
			if d.selectedIdx >= 0 && d.selectedIdx < len(d.filteredFiles) {
				selectedFile := d.filteredFiles[d.selectedIdx]
				d.visible = false
				return d, func() tea.Msg {
					return FileSelectedMsg{Path: selectedFile.path}
				}
			}
		case "up", "ctrl+k":
			if d.selectedIdx > 0 {
				d.selectedIdx--
			}
			return d, nil
		case "down", "ctrl+j":
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
	d.filterInput, cmd = d.filterInput.Update(msg)
	d.applyFuzzyFilter()
	return d, cmd
}

func (d *FileDialog) applyFuzzyFilter() {
	query := d.filterInput.Value()

	if query == "" {
		d.filteredFiles = d.allFiles
		d.selectedIdx = 0
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

	d.selectedIdx = 0
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
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("63")).
				Width(dialogWidth - 4).
				Render("> " + file.name)
		} else {
			line = lipgloss.NewStyle().
				Width(dialogWidth - 4).
				Render("  " + file.name)
		}
		fileListView.WriteString(line + "\n")
	}

	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		fileListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	title := dialogTitleStyle.Render("File Switcher")
	fileCount := fmt.Sprintf("(%d/%d files)", len(d.filteredFiles), len(d.allFiles))
	titleLine := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Render(title + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(fileCount))

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(strings.Repeat("─", dialogWidth-4))

	inputLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Filter: ")

	inputView := inputLabel + d.filterInput.View()

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

func (d *FileDialog) IsVisible() bool {
	return d.visible
}
