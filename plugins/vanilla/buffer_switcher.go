package vanilla

import (
	"fmt"
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

// CmdBufferSwitch is the command name constant for buffer switcher
const CmdBufferSwitch = "buffer-switch"

// BufferSwitcherKeyBinding is the key binding for the buffer switcher command
var BufferSwitcherKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+b"),
	key.WithHelp("ctrl+b", "switch buffer"),
)

func init() {
	plugin.RegisterCommand(plugin.CommandRegistration{
		Name:           CmdBufferSwitch,
		Key:            "Ctrl-B",
		Description:    "Open buffer switcher dialog",
		KeyBinding:     BufferSwitcherKeyBinding,
		PluginExecute: nil, // Main app provides execute handler
	})
}

// BufferSwitcherCommand returns the command definition for buffer switching
func BufferSwitcherCommand() api.CommandDef {
	return api.CommandDef{
		Name:        CmdBufferSwitch,
		Key:         "Ctrl-B",
		Description: "Open buffer switcher dialog",
		KeyBinding:  BufferSwitcherKeyBinding,
	}
}

// ====== Message Types ======

// BufferSelectedMsg is sent when a buffer is selected in the buffer dialog
type BufferSelectedMsg struct {
	Index int
}

// ====== Key Bindings ======

// BufferDialogKeyMap defines the key bindings for the buffer dialog
type BufferDialogKeyMap struct {
	Close key.Binding
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
}

// DefaultBufferDialogKeyMap returns the default key bindings for buffer dialog
var DefaultBufferDialogKeyMap = BufferDialogKeyMap{
	Close: key.NewBinding(
		key.WithKeys("esc", "ctrl+c", "ctrl+b"),
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
		key.WithHelp("enter", "switch buffer"),
	),
}

// ====== Internal Types ======

// bufferItem is used internally by BufferDialog
type bufferItem struct {
	name  string
	index int
}

// ====== Dialog Implementation ======

// BufferDialog implements the Dialog interface for buffer selection
type BufferDialog struct {
	filterInput     textinput.Model
	allBuffers      []bufferItem
	filteredBuffers []bufferItem
	selectedIdx     int
	visible         bool
	lastQuery       string // Track last query to avoid unnecessary resets
}

// NewBufferDialog creates a new buffer dialog
func NewBufferDialog(buffers []api.BufferInfo, currentBuffer int) *BufferDialog {
	ti := textinput.New()
	ti.Placeholder = "Type to filter buffers..."
	ti.CharLimit = 100
	ti.Width = 50
	ti.Focus()

	// Build buffer list
	var bufferItems []bufferItem
	for i, buf := range buffers {
		name := filepath.Base(buf.FilePath)
		if buf.ReadOnly {
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

func (d *BufferDialog) Update(msg tea.Msg) (api.Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, DefaultBufferDialogKeyMap.Close) {
			d.visible = false
			return d, nil
		}
		if key.Matches(msg, DefaultBufferDialogKeyMap.Enter) {
			if d.selectedIdx >= 0 && d.selectedIdx < len(d.filteredBuffers) {
				selectedBuffer := d.filteredBuffers[d.selectedIdx]
				d.visible = false
				return d, func() tea.Msg {
					return BufferSelectedMsg{Index: selectedBuffer.index}
				}
			}
		}
		if key.Matches(msg, DefaultBufferDialogKeyMap.Up) {
			if d.selectedIdx > 0 {
				d.selectedIdx--
			}
			return d, nil
		}
		if key.Matches(msg, DefaultBufferDialogKeyMap.Down) {
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
	oldValue := d.filterInput.Value()
	d.filterInput, cmd = d.filterInput.Update(msg)

	// Only apply filter if the query actually changed
	if d.filterInput.Value() != oldValue {
		d.applyFuzzyFilter()
	}
	return d, cmd
}

func (d *BufferDialog) applyFuzzyFilter() {
	query := d.filterInput.Value()

	if query == "" {
		d.filteredBuffers = d.allBuffers
		// Only reset selection if query actually changed (not just cursor blink)
		if d.lastQuery != "" {
			d.selectedIdx = 0
		}
		d.lastQuery = query
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

	// Only reset selection when the query changes, not on every blink
	if d.lastQuery != query {
		d.selectedIdx = 0
	}
	d.lastQuery = query
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
			line = api.DialogHighlightedStyle.
				Width(dialogWidth - 4).
				Render("> " + buffer.name)
		} else {
			line = api.DialogItemStyle.
				Width(dialogWidth - 4).
				Render("  " + buffer.name)
		}
		bufferListView.WriteString(line + "\n")
	}

	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		bufferListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	title := api.DialogTitleStyle.Render("Buffer Switcher")
	bufferCount := fmt.Sprintf("(%d/%d buffers)", len(d.filteredBuffers), len(d.allBuffers))
	titleLine := api.DialogTitleLineStyle.
		Width(dialogWidth - 4).
		Render(title + " " + api.DialogCountStyle.Render(bufferCount))

	separator := api.DialogSeparatorStyle.
		Render(strings.Repeat("─", dialogWidth-4))

	inputLabel := api.DialogInputLabelStyle.
		Render("Filter: ")

	inputView := inputLabel + d.filterInput.View()

	instructions := api.DialogInstructionsStyle.
		Render("↑/↓: Navigate | Enter: Switch | Esc: Close")

	fullContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		titleLine,
		separator,
		strings.TrimRight(bufferListView.String(), "\n"),
		inputView,
		instructions,
	)

	return api.DialogBoxStyle.Render(fullContent)
}

func (d *BufferDialog) IsVisible() bool {
	return d.visible
}
