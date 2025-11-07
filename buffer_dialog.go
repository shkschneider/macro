package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// BufferDialog implements the Dialog interface for buffer selection
type BufferDialog struct {
	filterInput     textinput.Model
	allBuffers      []bufferItem
	filteredBuffers []bufferItem
	selectedIdx     int
	visible         bool
}

// NewBufferDialog creates a new buffer dialog
func NewBufferDialog(buffers []Buffer, currentBuffer int) *BufferDialog {
	ti := textinput.New()
	ti.Placeholder = "Type to filter buffers..."
	ti.CharLimit = 100
	ti.Width = 50
	ti.Focus()

	// Build buffer list
	var bufferItems []bufferItem
	for i, buf := range buffers {
		name := filepath.Base(buf.filePath)
		if buf.readOnly {
			name += " [RO]"
		}
		if i == currentBuffer {
			name = "* " + name
		}
		bufferItems = append(bufferItems, bufferItem{
			name:  name,
			index: i,
		})
	}

	return &BufferDialog{
		filterInput:     ti,
		allBuffers:      bufferItems,
		filteredBuffers: bufferItems,
		selectedIdx:     0,
		visible:         true,
	}
}

func (d *BufferDialog) Init() tea.Cmd {
	return textinput.Blink
}

func (d *BufferDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "ctrl+b":
			d.visible = false
			return d, nil
		case "enter":
			if d.selectedIdx >= 0 && d.selectedIdx < len(d.filteredBuffers) {
				selectedBuffer := d.filteredBuffers[d.selectedIdx]
				d.visible = false
				return d, func() tea.Msg {
					return BufferSelectedMsg{Index: selectedBuffer.index}
				}
			}
		case "up", "ctrl+k":
			if d.selectedIdx > 0 {
				d.selectedIdx--
			}
			return d, nil
		case "down", "ctrl+j":
			if d.selectedIdx < len(d.filteredBuffers)-1 {
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

func (d *BufferDialog) applyFuzzyFilter() {
	query := d.filterInput.Value()

	if query == "" {
		d.filteredBuffers = d.allBuffers
		d.selectedIdx = 0
		return
	}

	// Build list of buffer names for fuzzy matching
	var targets []string
	for _, buffer := range d.allBuffers {
		targets = append(targets, buffer.name)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)

	// Build filtered list based on matches
	d.filteredBuffers = []bufferItem{}
	for _, match := range matches {
		d.filteredBuffers = append(d.filteredBuffers, d.allBuffers[match.Index])
	}

	d.selectedIdx = 0
}

func (d *BufferDialog) View(termWidth, termHeight int) string {
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

	var bufferListView strings.Builder
	listHeight := dialogHeight - 6

	startIdx := 0
	endIdx := len(d.filteredBuffers)

	if endIdx > listHeight {
		if d.selectedIdx >= listHeight {
			startIdx = d.selectedIdx - listHeight + 1
		}
		endIdx = startIdx + listHeight
		if endIdx > len(d.filteredBuffers) {
			endIdx = len(d.filteredBuffers)
		}
	}

	for i := startIdx; i < endIdx; i++ {
		buffer := d.filteredBuffers[i]
		line := ""
		if i == d.selectedIdx {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("63")).
				Width(dialogWidth - 4).
				Render("> " + buffer.name)
		} else {
			line = lipgloss.NewStyle().
				Width(dialogWidth - 4).
				Render("  " + buffer.name)
		}
		bufferListView.WriteString(line + "\n")
	}

	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		bufferListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	title := dialogTitleStyle.Render("Buffer Switcher")
	bufferCount := fmt.Sprintf("(%d/%d buffers)", len(d.filteredBuffers), len(d.allBuffers))
	titleLine := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Render(title + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(bufferCount))

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(strings.Repeat("─", dialogWidth-4))

	inputLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Filter: ")

	inputView := inputLabel + d.filterInput.View()

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("↑/↓: Navigate | Enter: Switch | Esc: Close")

	fullContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		titleLine,
		separator,
		strings.TrimRight(bufferListView.String(), "\n"),
		inputView,
		instructions,
	)

	return dialogBoxStyle.Render(fullContent)
}

func (d *BufferDialog) IsVisible() bool {
	return d.visible
}
