package internal

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CommandInput represents the command input line at the bottom of the editor.
// It can operate in two modes:
// - Message mode (read-only): displays messages, errors, warnings
// - Input mode (read-write): accepts command input from the user
type CommandInput struct {
	textInput textinput.Model
	isActive  bool   // true when in input mode, false for message mode
	message   string // current message to display in message mode
	prompt    string // prompt to show in input mode
}

// NewCommandInput creates a new command input component.
func NewCommandInput() *CommandInput {
	ti := textinput.New()
	ti.Prompt = ": "
	ti.CharLimit = 256
	ti.Width = 80 // Will be updated on resize

	return &CommandInput{
		textInput: ti,
		isActive:  false,
		message:   defaultMessage,
		prompt:    ": ",
	}
}

// SetWidth sets the width of the command input.
func (c *CommandInput) SetWidth(w int) {
	c.textInput.Width = w - len(c.prompt) - 2 // Account for prompt and some padding
}

// SetMessage sets the message to display in message mode.
func (c *CommandInput) SetMessage(msg string) {
	c.message = msg
}

// GetMessage returns the current message.
func (c *CommandInput) GetMessage() string {
	return c.message
}

// Activate switches to input mode and focuses the input.
func (c *CommandInput) Activate() tea.Cmd {
	c.isActive = true
	c.textInput.SetValue("")
	c.textInput.Focus()
	return textinput.Blink
}

// Deactivate switches back to message mode.
func (c *CommandInput) Deactivate() {
	c.isActive = false
	c.textInput.Blur()
	c.textInput.SetValue("")
}

// IsActive returns true if the command input is in input mode.
func (c *CommandInput) IsActive() bool {
	return c.isActive
}

// Value returns the current input value.
func (c *CommandInput) Value() string {
	return c.textInput.Value()
}

// Update handles messages for the command input.
func (c *CommandInput) Update(msg tea.Msg) (*CommandInput, tea.Cmd) {
	if !c.isActive {
		return c, nil
	}

	var cmd tea.Cmd
	c.textInput, cmd = c.textInput.Update(msg)
	return c, cmd
}

// View renders the command input.
func (c *CommandInput) View(termWidth int, errState error, isWarning bool, isSuccess bool) string {
	if c.isActive {
		// Input mode - show the text input with prompt
		inputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("235"))
		
		promptStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Bold(true).
			Background(lipgloss.Color("235"))
		
		// Render the prompt and input
		prompt := promptStyle.Render(c.prompt)
		input := c.textInput.View()
		
		// Fill remaining width with background
		contentWidth := lipgloss.Width(prompt) + lipgloss.Width(input)
		padding := ""
		if contentWidth < termWidth {
			padding = strings.Repeat(" ", termWidth-contentWidth)
		}
		
		return inputStyle.Render(prompt + input + padding)
	}

	// Message mode - show the message with appropriate styling
	if errState != nil {
		return ErrorStyle.Width(termWidth).Render(c.message)
	} else if isWarning {
		return WarningStyle.Width(termWidth).Render(c.message)
	} else if isSuccess {
		return SuccessStyle.Width(termWidth).Render(c.message)
	}
	return MessageStyle.Width(termWidth).Render(c.message)
}
