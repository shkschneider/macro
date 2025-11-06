# macro

A nano-like simple CLI text editor built with [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- Simple terminal-based text editor with Elm-inspired architecture
- Keyboard shortcuts similar to nano:
  - **Ctrl-S**: Save file
  - **Ctrl-Q**: Quit editor
- Clean, modern UI with syntax highlighting for status messages

## Installation

```bash
go build -o macro
```

## Usage

```bash
# Create or edit a new file
./macro filename.txt

# The editor will open in your terminal
# Use Ctrl-S to save changes
# Use Ctrl-Q to quit
```

## Requirements

- Go 1.23 or later
- Terminal with support for ANSI/VT sequences

## Architecture

### Overview

The application follows the **Elm Architecture** pattern (Model-View-Update) via Bubbletea, providing a functional, reactive approach to building terminal UIs.

```
┌─────────────────────────────────────────────────┐
│                   main()                        │
│  - Parse CLI args (filename)                    │
│  - Initialize Bubbletea program                 │
└──────────────┬──────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────┐
│            initialModel()                       │
│  - Create textarea component (Bubbles)          │
│  - Load file content if exists                  │
│  - Set initial state                            │
└──────────────┬──────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────┐
│         Tea.Program Runtime                     │
│                                                 │
│  ┌───────────────────────────────────────────┐ │
│  │            Model (State)                   │ │
│  │  - textarea: Bubbles TextArea component   │ │
│  │  - filePath: Current file path            │ │
│  │  - status: Status bar message             │ │
│  │  - err: Error state flag                  │ │
│  └───────────────────────────────────────────┘ │
│                                                 │
│  ┌───────────────────────────────────────────┐ │
│  │         Update (Message Handler)          │ │
│  │  - KeyMsg: Handle keyboard input          │ │
│  │    • Ctrl-Q → Quit program                │ │
│  │    • Ctrl-S → Save file to disk           │ │
│  │    • Other keys → Pass to textarea        │ │
│  │  - WindowSizeMsg: Adjust textarea size    │ │
│  └───────────────────────────────────────────┘ │
│                                                 │
│  ┌───────────────────────────────────────────┐ │
│  │          View (Renderer)                  │ │
│  │  1. Title (styled with Lipgloss)          │ │
│  │  2. TextArea content                      │ │
│  │  3. Status bar (color-coded):             │ │
│  │     - Yellow: Default shortcuts           │ │
│  │     - Green: Success messages             │ │
│  │     - Red: Error messages                 │ │
│  └───────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

### Component Breakdown

**1. Model (`type model struct`)**
- Holds all application state
- Immutable updates via Update function
- Contains:
  - `textarea`: Bubbles TextArea for editing
  - `filePath`: Path to file being edited
  - `status`: Current status message
  - `err`: Error tracking

**2. Update (`func (m model) Update`)**
- Pure function: `(Model, Message) → (Model, Command)`
- Handles three message types:
  - **KeyMsg**: Keyboard events (Ctrl-S, Ctrl-Q, typing)
  - **WindowSizeMsg**: Terminal resize events
  - Delegates to textarea for text editing
- Returns new model state and optional commands

**3. View (`func (m model) View`)**
- Pure function: `Model → String`
- Renders UI from current state
- Uses Lipgloss for styling:
  - Title in bold blue
  - Status bar with conditional colors
- Layout: Title + TextArea + Status

**4. Styling (Lipgloss)**
- `titleStyle`: Bold blue for title bar
- `statusStyle`: Yellow for default status
- `successStyle`: Green for save confirmations
- `errorStyle`: Red for errors

### Data Flow

```
User Input (Keyboard)
    ↓
Message (tea.KeyMsg)
    ↓
Update Function
    ↓
New Model State
    ↓
View Function
    ↓
Rendered UI (Terminal)
```

This architecture ensures:
- **Predictable state management**: All state changes go through Update
- **Testability**: Pure functions are easy to test
- **Maintainability**: Clear separation of concerns
- **Extensibility**: Easy to add new features via new message types
