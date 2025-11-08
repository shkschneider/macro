package feature

import (
	"fmt"
	"strconv"
	"strings"

	macro "github.com/shkschneider/macro/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// ====== Command Registration ======

// GotoCommand returns the command definition for goto line
func GotoCommand() macro.CommandDef {
	return macro.CommandDef{
		Name:        "goto-line",
		Key:         "Ctrl-G",
		Description: "Go to line number or list all lines",
	}
}

// ====== Message Types ======

// GotoLineMsg is sent when a line is selected or entered in the goto dialog
type GotoLineMsg struct {
	Line int // 1-based line number
	Col  int // 1-based column number (0 means start of line)
}

// ====== Internal Types ======

// lineItem is used internally by GotoDialog
type lineItem struct {
	lineNum int
	content string
}

// ====== Dialog Implementation ======

// GotoDialog implements the Dialog interface for goto line selection
type GotoDialog struct {
	filterInput   textinput.Model
	allLines      []lineItem
	filteredLines []lineItem
	selectedIdx   int
	visible       bool
	lastQuery     string // Track last query to avoid unnecessary resets
}

// NewGotoDialog creates a new goto dialog
func NewGotoDialog(content string) *GotoDialog {
	ti := textinput.New()
	ti.Placeholder = "Type line number (e.g., 10 or 10:5)..."
	ti.CharLimit = 20
	ti.Width = 50
	ti.Focus()

	// Split content into lines
	lines := strings.Split(content, "\n")
	var lineItems []lineItem
	for i, line := range lines {
		lineItems = append(lineItems, lineItem{
			lineNum: i + 1, // 1-based line numbers
			content: line,
		})
	}

	return &GotoDialog{
		filterInput:   ti,
		allLines:      lineItems,
		filteredLines: lineItems,
		selectedIdx:   0,
		visible:       true,
	}
}

func (d *GotoDialog) Init() tea.Cmd {
	return textinput.Blink
}

func (d *GotoDialog) Update(msg tea.Msg) (macro.Dialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c", "ctrl+g":
			d.visible = false
			return d, nil
		case "enter":
			// Try to parse the input as line[:col]
			input := strings.TrimSpace(d.filterInput.Value())
			if input != "" {
				line, col := d.parseLineCol(input)
				if line > 0 {
					d.visible = false
					return d, func() tea.Msg {
						return GotoLineMsg{Line: line, Col: col}
					}
				}
			}
			// If no input or invalid, use selected line from list
			if d.selectedIdx >= 0 && d.selectedIdx < len(d.filteredLines) {
				selectedLine := d.filteredLines[d.selectedIdx]
				d.visible = false
				return d, func() tea.Msg {
					return GotoLineMsg{Line: selectedLine.lineNum, Col: 0}
				}
			}
		case "up", "ctrl+k":
			if d.selectedIdx > 0 {
				d.selectedIdx--
			}
			return d, nil
		case "down", "ctrl+j":
			if d.selectedIdx < len(d.filteredLines)-1 {
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
		d.applyFilter()
	}
	return d, cmd
}

// parseLineCol parses input like "10" or "10:5" into line and column numbers
func (d *GotoDialog) parseLineCol(input string) (int, int) {
	parts := strings.Split(input, ":")
	
	// Parse line number
	line, err := strconv.Atoi(parts[0])
	if err != nil || line < 1 {
		return 0, 0
	}
	
	// Parse column number if provided
	col := 0
	if len(parts) > 1 {
		parsedCol, err := strconv.Atoi(parts[1])
		if err == nil && parsedCol > 0 {
			col = parsedCol
		}
	}
	
	return line, col
}

func (d *GotoDialog) applyFilter() {
	query := d.filterInput.Value()

	if query == "" {
		d.filteredLines = d.allLines
		// Only reset selection if query actually changed (not just cursor blink)
		if d.lastQuery != "" {
			d.selectedIdx = 0
		}
		d.lastQuery = query
		return
	}

	// Check if query is a line number or line:col format
	if d.isLineNumberQuery(query) {
		// Filter to show lines near the target line number
		line, _ := d.parseLineCol(query)
		if line > 0 {
			d.filteredLines = d.filterNearLine(line)
			d.selectedIdx = d.findLineIndex(line)
			d.lastQuery = query
			return
		}
	}

	// Otherwise, do fuzzy search on line content
	var targets []string
	for _, line := range d.allLines {
		targets = append(targets, line.content)
	}

	// Perform fuzzy search
	matches := fuzzy.Find(query, targets)

	// Build filtered list based on matches
	d.filteredLines = []lineItem{}
	for _, match := range matches {
		d.filteredLines = append(d.filteredLines, d.allLines[match.Index])
	}

	// Only reset selection when the query changes, not on every blink
	if d.lastQuery != query {
		d.selectedIdx = 0
	}
	d.lastQuery = query
}

// isLineNumberQuery checks if the query looks like a line number
func (d *GotoDialog) isLineNumberQuery(query string) bool {
	parts := strings.Split(query, ":")
	_, err := strconv.Atoi(parts[0])
	return err == nil
}

// filterNearLine returns lines near the target line (context window)
func (d *GotoDialog) filterNearLine(targetLine int) []lineItem {
	const contextLines = 5
	start := targetLine - contextLines - 1
	if start < 0 {
		start = 0
	}
	end := targetLine + contextLines
	if end > len(d.allLines) {
		end = len(d.allLines)
	}

	var result []lineItem
	for i := start; i < end; i++ {
		result = append(result, d.allLines[i])
	}
	return result
}

// findLineIndex finds the index of a line number in the filtered list
func (d *GotoDialog) findLineIndex(lineNum int) int {
	for i, line := range d.filteredLines {
		if line.lineNum == lineNum {
			return i
		}
	}
	return 0
}

func (d *GotoDialog) View(termWidth, termHeight int) string {
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

	var lineListView strings.Builder
	listHeight := dialogHeight - 6

	startIdx := 0
	endIdx := len(d.filteredLines)

	if endIdx > listHeight {
		if d.selectedIdx >= listHeight {
			startIdx = d.selectedIdx - listHeight + 1
		}
		endIdx = startIdx + listHeight
		if endIdx > len(d.filteredLines) {
			endIdx = len(d.filteredLines)
		}
	}

	for i := startIdx; i < endIdx; i++ {
		line := d.filteredLines[i]
		// Truncate line content if too long
		content := line.content
		maxContentWidth := dialogWidth - 15 // Reserve space for line number and padding
		if len(content) > maxContentWidth {
			content = content[:maxContentWidth-3] + "..."
		}
		lineText := fmt.Sprintf("%4d: %s", line.lineNum, content)
		lineView := ""
		if i == d.selectedIdx {
			lineView = macro.DialogHighlightedStyle.
				Width(dialogWidth - 4).
				Render("> " + lineText)
		} else {
			lineView = macro.DialogItemStyle.
				Width(dialogWidth - 4).
				Render("  " + lineText)
		}
		lineListView.WriteString(lineView + "\n")
	}

	linesRendered := endIdx - startIdx
	for i := linesRendered; i < listHeight; i++ {
		lineListView.WriteString(strings.Repeat(" ", dialogWidth-4) + "\n")
	}

	title := macro.DialogTitleStyle.Render("Go to Line")
	lineCount := fmt.Sprintf("(%d lines)", len(d.allLines))
	titleLine := lipgloss.NewStyle().
		Width(dialogWidth - 4).
		Render(title + " " + macro.DialogCountStyle.Render(lineCount))

	separator := macro.DialogSeparatorStyle.
		Render(strings.Repeat("─", dialogWidth-4))

	inputLabel := macro.DialogInputLabelStyle.
		Render("Line: ")

	inputView := inputLabel + d.filterInput.View()

	instructions := macro.DialogInstructionsStyle.
		Render("↑/↓: Navigate | Enter: Go to Line | Esc: Close")

	fullContent := fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		titleLine,
		separator,
		strings.TrimRight(lineListView.String(), "\n"),
		inputView,
		instructions,
	)

	return macro.DialogBoxStyle.Render(fullContent)
}

func (d *GotoDialog) IsVisible() bool {
	return d.visible
}
