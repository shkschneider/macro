# macro

A nano-like simple CLI text editor built with [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- Simple terminal-based text editor with Elm-inspired architecture
  - Main app only provides quit and command palette.
  - All other features are modularized in plugins (default: vanilla).
- Keyboard shortcuts similar to nano:
  - **Ctrl-S**: Save file
  - **Ctrl-Space**: Open fuzzy command palette
  - **Ctrl-Q**: Quit editor
  - ...
- **Syntax highlighting** powered by [Chroma](https://github.com/alecthomas/chroma)

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

### CLI Options

| Flag | Description |
|------|-------------|
| `-ro` | Force read-only mode (uses viewport with syntax highlighting) |
| `-rw` | Force read-write mode (uses textarea for editing, if file is writable) |

## Requirements

- Go 1.23 or later
- Terminal with support for ANSI/VT sequences

## Architecture

### Overview

The application follows the **Elm Architecture** pattern (Model-View-Update) via Bubbletea, providing a functional, reactive approach to building terminal UIs.

```
main() → InitialModel() → Tea.Program Runtime
    Model (State)
    Update (Message Handler)
    View (Renderer)
```

- Features (self-contained): quit.go, save.go, palette.go
- Core data: buffer.go (Buffer struct), model.go (Model struct)
- API interface: context.go (EditorContext implementation), command.go (Command registry)
- UI components: textarea.go, status.go, styles.go, dialog.go, overlay.go
- Git integration: git.go, diff.go
- Utilities: cursor.go, highlight.go
- App lifecycle: app.go

### Component Breakdown

**1. Model (`type model struct`)**
- Holds all application state
- Immutable updates via Update function

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

### Data Flow

```
User Input (Keyboard)
→ Message (tea.KeyMsg)
→ Update Function
→ New Model State
→ View Function
→ Rendered UI (Terminal)
```

This architecture ensures:
- **Predictable state management**: All state changes go through Update
- **Testability**: Pure functions are easy to test
- **Maintainability**: Clear separation of concerns
- **Extensibility**: Easy to add new features via new message types

## Syntax Highlighting

Macro uses [Chroma](https://github.com/alecthomas/chroma), a general purpose syntax highlighter in pure Go, to provide syntax highlighting capabilities.

### Supported Languages

Over 300 programming and markup languages are supported out of the box, including:
- Go, Python, JavaScript, TypeScript, Java, C, C++, Rust
- HTML, CSS, JSON, YAML, XML, Markdown
- SQL, Shell/Bash, PowerShell
- And many more...

Language detection is automatic based on file extension.

### Color Themes

Multiple color themes are available. The default theme is Monokai. Other popular themes include:
- Dracula
- GitHub
- Solarized (Dark/Light)
- Nord
- Gruvbox

### API Reference

...
