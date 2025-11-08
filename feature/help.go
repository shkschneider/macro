package feature

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// ====== Command Registration ======

// HelpCommand returns the command definition for showing help
func HelpCommand() CommandDef {
	return CommandDef{
		Name:        "help-show",
		Key:         "Ctrl-H",
		Description: "Show this help dialog",
	}
}

// ====== Message Types ======

// CommandSelectedMsg is sent when a command is selected in the help dialog
type CommandSelectedMsg struct {
	CommandName string
}

// ====== Internal Types ======

// commandItem is used internally by HelpDialog
type commandItem struct {
	command CommandDef
}

// ====== Dialog Implementation ======

// HelpDialog implements the Dialog interface for help/command selection
type HelpDialog struct {
	filterInput      textinput.Model
	allCommands      []commandItem
	filteredCommands []commandItem
	selectedIdx      int
	visible          bool
	lastQuery        string // Track last query to avoid unnecessary resets
}

// NewHelpDialog creates a new help dialog
func NewHelpDialog(commands []CommandDef) *HelpDialog {
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

	return &HelpDialog{
		filterInput:      ti,
		allCommands:      commandItems,
		filteredCommands: commandItems,
		selectedIdx:      0,
		visible:          true,
	}
}

func (d *HelpDialog) Init() tea.Cmd {
	return textinput.Blink
}

func (d *HelpDialog) Update(msg tea.Msg) (Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "ctrl+h":
			d.visible = false
			return d, nil
		case "enter":
			if d.selectedIdx >= 0 && d.selectedIdx < len(d.filteredCommands) {
				selectedCommand := d.filteredCommands[d.selectedIdx]
				d.visible = false
				return d, func() tea.Msg {
					return CommandSelectedMsg{CommandName: selectedCommand.command.Name}
				}
			}
		case "up", "ctrl+k":
			if d.selectedIdx > 0 {
				d.selectedIdx--
			}
			return d, nil
		case "down", "ctrl+j":
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

func (d *HelpDialog) applyFuzzyFilter() {
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

func (d *HelpDialog) View(termWidth, termHeight int) string {
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
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("63")).
				Width(dialogWidth - 4).
				Render("> " + cmdText)
		} else {
			line = lipgloss.NewStyle().
				Width(dialogWidth - 4).
				Render("  " + cmdText)
		}
		helpListView.WriteString(line + "\n")
	}

	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		helpListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	title := dialogTitleStyle.Render("Help - Commands")
	cmdCount := fmt.Sprintf("(%d/%d commands)", len(d.filteredCommands), len(d.allCommands))
	titleLine := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Render(title + " " + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(cmdCount))

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(strings.Repeat("─", dialogWidth-4))

	inputLabel := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Filter: ")

	inputView := inputLabel + d.filterInput.View()

	instructions := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("↑/↓: Navigate | Enter: Run Command | Esc: Close")

	fullContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		titleLine,
		separator,
		strings.TrimRight(helpListView.String(), "\n"),
		inputView,
		instructions,
	)

	return dialogBoxStyle.Render(fullContent)
}

func (d *HelpDialog) IsVisible() bool {
	return d.visible
}
