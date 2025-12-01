package internal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sahilm/fuzzy"
	"github.com/shkschneider/macro/api"
)

// ====== Command Registration ======

// CmdCommandInput is the command name constant for focusing command input
const CmdCommandInput = "command-input"

// CmdPalette is the command name constant for help/command palette
const CmdPalette = "help-show"

// CommandInputKeyBinding is the key binding for the command input line
var CommandInputKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+@", "ctrl+ "), // ctrl+@ is what ctrl+space sends
	key.WithHelp("ctrl+space", "focus command input"),
)

// PaletteKeyBinding is the key binding for the command palette (can be invoked from command input)
var PaletteKeyBinding = key.NewBinding(
	key.WithKeys("ctrl+shift+p"),
	key.WithHelp("ctrl+shift+p", "open command palette"),
)

// ExecuteCommandInput focuses the command input line for typing commands.
func ExecuteCommandInput(m *Model) tea.Cmd {
	if m.CommandInput != nil {
		m.CommandInput.SetWidth(TermWidth)
		return m.CommandInput.Activate()
	}
	return nil
}

func ExecuteCommandPalette(m *Model) tea.Cmd {
	// Get all commands
	var commands []api.CommandRegistration
	for _, cmd := range GetCommands() {
		commands = append(commands, api.CommandRegistration{
			Name:        cmd.Name,
			Key:         cmd.Key,
			Description: cmd.Description,
		})
	}
	m.ActiveDialog = NewPaletteDialog(commands)
	return m.ActiveDialog.Init()
}

func init() {
	// Register command input activation (Ctrl-Space)
	api.RegisterCommand(api.CommandRegistration{
		Name:          CmdCommandInput,
		Key:           "Ctrl-Space",
		Description:   "Focus command input",
		KeyBinding:    CommandInputKeyBinding,
		PluginExecute: nil, // Main app provides execute handler
	})
	
	// Register command palette (accessible from command input)
	api.RegisterCommand(api.CommandRegistration{
		Name:          CmdPalette,
		Key:           "Ctrl-Shift-P",
		Description:   "Show command palette",
		KeyBinding:    PaletteKeyBinding,
		PluginExecute: nil, // Main app provides execute handler
	})
}


// ====== Message Types ======

// CommandSelectedMsg is sent when a command is selected in the help dialog
type CommandSelectedMsg struct {
	CommandName string
}

// Handle implements api.PluginMsg - executes the selected command
func (msg CommandSelectedMsg) Handle(ctx api.EditorContext) tea.Cmd {
	cmd := GetCommandByName(msg.CommandName)
	if cmd != nil && cmd.Execute != nil {
		// We need to cast ctx back to *Model to call Execute
		if m, ok := ctx.(*Model); ok {
			return cmd.Execute(m)
		}
	}
	return nil
}

// ====== Key Bindings ======

// PaletteDialogKeyMap defines the key bindings for the help dialog
type PaletteDialogKeyMap struct {
	Close key.Binding
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
}

// DefaultPaletteDialogKeyMap returns the default key bindings for help dialog
var DefaultPaletteDialogKeyMap = PaletteDialogKeyMap{
	Close: key.NewBinding(
		key.WithKeys("esc", "ctrl+c", "ctrl+@", "ctrl+ "),
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
		key.WithHelp("enter", "run command"),
	),
}

// ====== Internal Types ======

// commandItem is used internally by PaletteDialog
type commandItem struct {
	command api.CommandRegistration
}

// ====== Dialog Implementation ======

// PaletteDialog implements the Dialog interface for help/command selection
type PaletteDialog struct {
	filterInput      textinput.Model
	allCommands      []commandItem
	filteredCommands []commandItem
	selectedIdx      int
	visible          bool
	lastQuery        string // Track last query to avoid unnecessary resets
}

// NewPaletteDialog creates a new help dialog
func NewPaletteDialog(commands []api.CommandRegistration) *PaletteDialog {
	ti := textinput.New()
	ti.Placeholder = "Type to filter commands..."
	ti.CharLimit = 100
	ti.Width = 50
	ti.Focus()

	// Build command list
	var commandItems []commandItem
	for _, cmd := range commands {
		commandItems = append(commandItems, commandItem{
			command: cmd,
		})
	}

	return &PaletteDialog{
		filterInput:      ti,
		allCommands:      commandItems,
		filteredCommands: commandItems,
		selectedIdx:      0,
		visible:          true,
	}
}

func (d *PaletteDialog) Init() tea.Cmd {
	return textinput.Blink
}

func (d *PaletteDialog) Update(msg tea.Msg) (api.Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, DefaultPaletteDialogKeyMap.Close) {
			d.visible = false
			return d, nil
		}
		if key.Matches(msg, DefaultPaletteDialogKeyMap.Enter) {
			if d.selectedIdx >= 0 && d.selectedIdx < len(d.filteredCommands) {
				selectedCommand := d.filteredCommands[d.selectedIdx]
				d.visible = false
				return d, func() tea.Msg {
					return CommandSelectedMsg{CommandName: selectedCommand.command.Name}
				}
			}
		}
		if key.Matches(msg, DefaultPaletteDialogKeyMap.Up) {
			if d.selectedIdx > 0 {
				d.selectedIdx--
			}
			return d, nil
		}
		if key.Matches(msg, DefaultPaletteDialogKeyMap.Down) {
			if d.selectedIdx < len(d.filteredCommands)-1 {
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

func (d *PaletteDialog) applyFuzzyFilter() {
	query := d.filterInput.Value()

	if query == "" {
		d.filteredCommands = d.allCommands
		// Only reset selection if query actually changed (not just cursor blink)
		if d.lastQuery != "" {
			d.selectedIdx = 0
		}
		d.lastQuery = query
		return
	}

	// Build list of command names and descriptions for fuzzy matching
	var targets []string
	for _, cmd := range d.allCommands {
		targets = append(targets, cmd.command.Name+" "+cmd.command.Description)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)

	// Build filtered list based on matches
	d.filteredCommands = []commandItem{}
	for _, match := range matches {
		d.filteredCommands = append(d.filteredCommands, d.allCommands[match.Index])
	}

	// Only reset selection when the query changes, not on every blink
	if d.lastQuery != query {
		d.selectedIdx = 0
	}
	d.lastQuery = query
}

func (d *PaletteDialog) View(termWidth, termHeight int) string {
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

	var helpListView strings.Builder
	listHeight := dialogHeight - 6

	startIdx := 0
	endIdx := len(d.filteredCommands)

	if endIdx > listHeight {
		if d.selectedIdx >= listHeight {
			startIdx = d.selectedIdx - listHeight + 1
		}
		endIdx = startIdx + listHeight
		if endIdx > len(d.filteredCommands) {
			endIdx = len(d.filteredCommands)
		}
	}

	for i := startIdx; i < endIdx; i++ {
		cmd := d.filteredCommands[i]
		cmdText := fmt.Sprintf("%-20s %-12s %s", cmd.command.Name, cmd.command.Key, cmd.command.Description)
		line := ""
		if i == d.selectedIdx {
			line = api.DialogHighlightedStyle.
				Width(dialogWidth - 4).
				Render("> " + cmdText)
		} else {
			line = api.DialogItemStyle.
				Width(dialogWidth - 4).
				Render("  " + cmdText)
		}
		helpListView.WriteString(line + "\n")
	}

	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		helpListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	title := api.DialogTitleStyle.Render("Command Palette")
	cmdCount := fmt.Sprintf("(%d/%d commands)", len(d.filteredCommands), len(d.allCommands))
	titleLine := api.DialogTitleLineStyle.
		Width(dialogWidth - 4).
		Render(title + " " + api.DialogCountStyle.Render(cmdCount))

	separator := api.DialogSeparatorStyle.
		Render(strings.Repeat("─", dialogWidth-4))

	inputLabel := api.DialogInputLabelStyle.
		Render("Filter: ")

	inputView := inputLabel + d.filterInput.View()

	instructions := api.DialogInstructionsStyle.
		Render("↑/↓: Navigate | Enter: Run Command | Esc: Close")

	fullContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		titleLine,
		separator,
		strings.TrimRight(helpListView.String(), "\n"),
		inputView,
		instructions,
	)

	return api.DialogBoxStyle.Render(fullContent)
}

func (d *PaletteDialog) IsVisible() bool {
	return d.visible
}
